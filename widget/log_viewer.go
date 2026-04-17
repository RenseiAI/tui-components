package widget

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// defaultMaxLines is the ring-buffer cap applied when no explicit
// WithMaxLines option is provided.
const defaultMaxLines = 10_000

// footerRows is the number of terminal rows reserved at the bottom of
// the widget for the FOLLOW / PAUSED indicator.
const footerRows = 1

// Compile-time assertion that *LogViewer satisfies component.Component.
var _ component.Component = (*LogViewer)(nil)

// LogViewer is a Bubble Tea widget that renders a stream of log lines
// inside a scrollable viewport. Lines are retained in a ring buffer
// (bounded by WithMaxLines) and rendered with ANSI SGR styling via the
// parser in widget/ansi.go.
//
// LogViewer implements the follow / scroll-lock state machine: while
// Following() is true, each [LogViewer.Append] auto-scrolls the
// viewport to the tail; otherwise the viewport offset is preserved so
// the user can read history without being yanked forward by new
// output. Explicit scroll-away keys ([up], [k], [pgup], [home], [g])
// and mouse-wheel-up events flip follow off; [end] / [G] re-engages
// and jumps to the tail; [f] toggles.
//
// The footer row below the viewport shows a short indicator rendered
// via [theme.LogFollow] / [theme.LogPaused]. The viewport itself is
// sized to height-1 so the footer always has a row.
//
// LogViewer is not safe for concurrent use. Drive it from a single
// goroutine (typically the tea.Program loop) using Append or AppendMsg.
type LogViewer struct {
	viewport viewport.Model
	parser   *sgrParser

	lines    []string
	maxLines int
	wrap     bool
	follow   bool
	focused  bool

	width  int
	height int
}

// Option configures a LogViewer at construction time.
type Option func(*LogViewer)

// WithMaxLines caps the number of retained lines. Once the buffer is
// full, the oldest line is discarded on each append. Passing 0 disables
// the cap (unbounded retention). Negative values are ignored with a
// warning logged to stderr and the default (10_000) is retained.
func WithMaxLines(n int) Option {
	return func(m *LogViewer) {
		if n < 0 {
			log.Warn("log viewer: negative WithMaxLines ignored", "requested", n, "fallback", defaultMaxLines)
			return
		}
		m.maxLines = n
	}
}

// WithWrap toggles soft-wrap rendering. When true (default), lines
// longer than the viewport width are wrapped to subsequent rows. When
// false, lines overflow horizontally and the viewport exposes
// horizontal scrolling.
func WithWrap(enabled bool) Option {
	return func(m *LogViewer) {
		m.wrap = enabled
	}
}

// WithFollow sets the initial follow state. When follow is on
// (default), each append auto-scrolls the viewport to the bottom.
func WithFollow(enabled bool) Option {
	return func(m *LogViewer) {
		m.follow = enabled
	}
}

// New constructs a LogViewer with the given options applied on top of
// the defaults (maxLines=10_000, wrap=true, follow=true).
func New(opts ...Option) *LogViewer {
	m := &LogViewer{
		parser:   newSGRParser(),
		maxLines: defaultMaxLines,
		wrap:     true,
		follow:   true,
	}
	for _, opt := range opts {
		opt(m)
	}

	// Pre-allocate the backing slice so steady-state appends re-use
	// existing capacity once the ring is warm.
	preallocCap := m.maxLines
	if preallocCap == 0 {
		preallocCap = defaultMaxLines
	}
	m.lines = make([]string, 0, preallocCap)

	vp := viewport.New()
	vp.SoftWrap = m.wrap
	m.viewport = vp

	return m
}

// AppendMsg is a tea.Msg form of Append. Dispatching this message
// through the program loop lets background producers push log lines
// without holding a reference to the widget.
type AppendMsg struct {
	// Lines is the batch of raw log lines (may contain ANSI escapes)
	// to append to the buffer.
	Lines []string
}

// Append adds one or more lines to the buffer. Overflow drops from the
// head. If Following() is true, the viewport is scrolled to the
// bottom after the append and re-render.
func (m *LogViewer) Append(lines ...string) {
	if len(lines) == 0 {
		return
	}

	for _, line := range lines {
		m.appendOne(line)
	}
	m.refresh()
}

// appendOne inserts a single line respecting the ring-buffer cap.
func (m *LogViewer) appendOne(line string) {
	if m.maxLines == 0 {
		m.lines = append(m.lines, line)
		return
	}
	if len(m.lines) < m.maxLines {
		m.lines = append(m.lines, line)
		return
	}
	// Full: drop oldest. copy reuses existing backing array.
	copy(m.lines, m.lines[1:])
	m.lines[len(m.lines)-1] = line
}

// Clear drops all retained lines and resets the SGR parser state.
func (m *LogViewer) Clear() {
	m.lines = m.lines[:0]
	m.parser.Reset()
	m.refresh()
}

// Following reports whether the viewport is currently tailing the log.
func (m *LogViewer) Following() bool {
	return m.follow
}

// SetFollowing toggles follow mode. Turning follow on re-scrolls the
// viewport to the bottom immediately so the caller doesn't need to
// issue a separate Append to catch up.
func (m *LogViewer) SetFollowing(enabled bool) {
	m.follow = enabled
	if enabled {
		m.viewport.GotoBottom()
	}
}

// Init satisfies tea.Model. LogViewer produces no initial commands.
func (m *LogViewer) Init() tea.Cmd {
	return nil
}

// scrollAwayKeys returns the set of key strings that, when received,
// flip follow off and scroll the viewport by one step. REN-997 will
// replace this hard-coded set with a configurable KeyMap; until then
// it is kept internal so the public surface can evolve without a
// breaking change.
func scrollAwayKeys() []string {
	return []string{"up", "k", "pgup", "home", "g"}
}

// scrollToTailKeys returns the set of key strings that re-engage
// follow and jump to the tail. Like scrollAwayKeys, this is a
// placeholder to be folded into REN-997's KeyMap.
func scrollToTailKeys() []string {
	return []string{"end", "G"}
}

// Update handles AppendMsg, WindowSizeMsg, and the internal
// scroll-away / scroll-to-tail / toggle key set. All other messages
// are dropped. REN-997 will replace the hard-coded key matching with
// a KeyMap and gate dispatch on focus.
func (m *LogViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AppendMsg:
		m.Append(msg.Lines...)
		return m, nil
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	}
	return m, nil
}

// handleKey dispatches a key press through the state machine. Any key
// that lies in scrollAwayKeys pauses follow and scrolls the viewport;
// scrollToTailKeys re-engages follow and jumps to the tail; "f"
// toggles. After the handler returns, if the viewport is no longer at
// the bottom, follow is flipped off (covers keys that the viewport's
// own KeyMap handles, e.g. half-page up).
func (m *LogViewer) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	s := msg.String()

	switch s {
	case "f":
		m.SetFollowing(!m.follow)
		return m, nil
	}

	for _, k := range scrollToTailKeys() {
		if s == k {
			m.SetFollowing(true)
			return m, nil
		}
	}

	for _, k := range scrollAwayKeys() {
		if s == k {
			m.follow = false
			m.viewport.ScrollUp(1)
			return m, nil
		}
	}

	// Delegate any other key to the viewport (e.g. page-down, down).
	// If the viewport moves off the tail we flip follow off; if it
	// lands on the tail we do NOT flip follow on — only the explicit
	// scrollToTailKeys and SetFollowing(true) re-engage tailing.
	vp, cmd := m.viewport.Update(msg)
	m.viewport = vp
	if !m.viewport.AtBottom() {
		m.follow = false
	}
	return m, cmd
}

// handleMouseWheel dispatches wheel events through the viewport and
// flips follow off if the wheel lifted the viewport off the tail.
func (m *LogViewer) handleMouseWheel(msg tea.MouseWheelMsg) (tea.Model, tea.Cmd) {
	vp, cmd := m.viewport.Update(msg)
	m.viewport = vp
	if !m.viewport.AtBottom() {
		m.follow = false
	}
	return m, cmd
}

// View renders the viewport plus the footer indicator as a tea.View.
// The viewport occupies height-footerRows rows; the footer takes the
// last row.
func (m *LogViewer) View() tea.View {
	vpView := m.viewport.View()

	var footer string
	if m.follow {
		footer = theme.LogFollow().Render("FOLLOW")
	} else {
		footer = theme.LogPaused().Render("PAUSED")
	}

	// Pad the footer row to the full width so composite layouts don't
	// see ragged edges.
	if m.width > 0 {
		footer = lipgloss.NewStyle().Width(m.width).Render(footer)
	}

	var b strings.Builder
	b.WriteString(vpView)
	if vpView != "" && !strings.HasSuffix(vpView, "\n") {
		b.WriteByte('\n')
	}
	b.WriteString(footer)
	return tea.NewView(b.String())
}

// SetSize propagates new dimensions to the embedded viewport and
// re-renders (wrap width changed). The viewport receives
// height-footerRows rows so the footer fits on the last row.
//
// Follow state is preserved across a resize: if follow is on the
// viewport is pinned to the bottom; if follow is off, the viewport
// keeps its prior Y-offset (clamped against the new content height,
// which the viewport handles internally). This is important so a
// paused reader is not yanked back to the tail on a terminal resize.
func (m *LogViewer) SetSize(width, height int) {
	m.width = width
	m.height = height

	vpHeight := height - footerRows
	if vpHeight < 0 {
		vpHeight = 0
	}
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(vpHeight)

	// Capture the offset before the content re-render so we can
	// restore it when paused. The viewport clamps to the new
	// maxYOffset on SetYOffset, so this is safe even when the content
	// height shrinks.
	prevOffset := m.viewport.YOffset()
	m.refresh()
	if !m.follow {
		m.viewport.SetYOffset(prevOffset)
	}
}

// Focus flips the internal focused flag. The full focus-gated dispatch
// lands in REN-997; until then Focus is a no-op stub.
func (m *LogViewer) Focus() {
	m.focused = true
}

// Blur flips the internal focused flag. See Focus for context.
func (m *LogViewer) Blur() {
	m.focused = false
}

// refresh re-renders the retained lines into the viewport. Called on
// every mutation that affects on-screen content (Append, Clear,
// SetSize).
func (m *LogViewer) refresh() {
	content := m.render()
	m.viewport.SetContent(content)
	if m.follow {
		m.viewport.GotoBottom()
	}
}

// render produces the single-string content fed to the viewport. Each
// retained line is SGR-parsed into styled runs and rendered via
// lipgloss; when wrap is enabled, each fully-styled line is then
// width-clamped to the viewport.
//
// The SGR parser is shared across lines deliberately: a multi-line
// escape sequence in the source stream preserves its colour across
// line boundaries. The parser is only reset on Clear.
func (m *LogViewer) render() string {
	if len(m.lines) == 0 {
		return ""
	}

	// Reset parser state each render so we replay from a clean slate
	// through the retained buffer. Without this, re-renders (e.g.
	// triggered by SetSize) would accumulate colour state from prior
	// passes.
	m.parser.Reset()

	var (
		out       strings.Builder
		lineBuf   strings.Builder
		wrapStyle = lipgloss.NewStyle()
	)
	if m.wrap && m.width > 0 {
		wrapStyle = wrapStyle.Width(m.width)
	}

	for i, raw := range m.lines {
		if i > 0 {
			out.WriteByte('\n')
		}
		lineBuf.Reset()
		runs := m.parser.Parse(raw)
		for _, r := range runs {
			if r.Text == "" {
				continue
			}
			lineBuf.WriteString(r.Style.Render(r.Text))
		}
		styled := lineBuf.String()
		if m.wrap && m.width > 0 {
			styled = wrapStyle.Render(styled)
		}
		out.WriteString(styled)
	}
	return out.String()
}

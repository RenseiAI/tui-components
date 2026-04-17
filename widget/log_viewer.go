package widget

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
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

// KeyMap defines the key bindings used by LogViewer. Consumers can
// rebind any action by constructing a new KeyMap and passing it via
// [WithKeyMap].
//
// Each binding carries its own help text (see [key.WithHelp]) so a
// future help-bar integration can surface the active bindings without
// needing a separate metadata table.
type KeyMap struct {
	// LineUp scrolls the viewport up by one line and pauses follow.
	LineUp key.Binding
	// LineDown scrolls the viewport down by one line.
	LineDown key.Binding
	// PageUp scrolls the viewport up by one page and pauses follow.
	PageUp key.Binding
	// PageDown scrolls the viewport down by one page.
	PageDown key.Binding
	// Home jumps to the top of the buffer and pauses follow.
	Home key.Binding
	// End jumps to the tail and re-engages follow.
	End key.Binding
	// ToggleFollow flips follow mode. Turning follow on jumps to the
	// tail.
	ToggleFollow key.Binding
	// ToggleWrap flips soft-wrap rendering.
	ToggleWrap key.Binding
	// Clear drops every retained line and resets the SGR parser.
	Clear key.Binding
}

// DefaultKeyMap returns the default key bindings for LogViewer:
//
//	LineUp:       up, k
//	LineDown:     down, j
//	PageUp:       pgup
//	PageDown:     pgdn
//	Home:         home, g
//	End:          end, G
//	ToggleFollow: f
//	ToggleWrap:   w
//	Clear:        c
func DefaultKeyMap() KeyMap {
	return KeyMap{
		LineUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		LineDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("home/g", "go to top"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("end/G", "go to tail (follow)"),
		),
		ToggleFollow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "toggle follow"),
		),
		ToggleWrap: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "toggle wrap"),
		),
		Clear: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clear buffer"),
		),
	}
}

// LogViewer is a Bubble Tea widget that renders a stream of log lines
// inside a scrollable viewport. Lines are retained in a ring buffer
// (bounded by WithMaxLines) and rendered with ANSI SGR styling via the
// parser in widget/ansi.go.
//
// LogViewer implements the follow / scroll-lock state machine: while
// Following() is true, each [LogViewer.Append] auto-scrolls the
// viewport to the tail; otherwise the viewport offset is preserved so
// the user can read history without being yanked forward by new
// output. The bindings in [KeyMap] drive the transitions (LineUp /
// PageUp / Home pause follow; End re-engages; ToggleFollow toggles);
// a mouse-wheel-up event also pauses follow.
//
// Key dispatch is gated on focus: a blurred LogViewer drops
// [tea.KeyPressMsg] messages but still processes [AppendMsg], mouse
// events, and [tea.WindowSizeMsg] so background producers keep
// delivering data regardless of which widget currently owns the focus
// ring.
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
	keys     KeyMap

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

// WithKeyMap overrides the default key bindings. Any unset fields on
// the supplied map disable the corresponding action (an empty
// [key.Binding] matches nothing).
func WithKeyMap(km KeyMap) Option {
	return func(m *LogViewer) {
		m.keys = km
	}
}

// New constructs a LogViewer with the given options applied on top of
// the defaults (maxLines=10_000, wrap=true, follow=true, KeyMap=DefaultKeyMap).
func New(opts ...Option) *LogViewer {
	m := &LogViewer{
		parser:   newSGRParser(),
		maxLines: defaultMaxLines,
		wrap:     true,
		follow:   true,
		keys:     DefaultKeyMap(),
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

// Wrap reports whether soft-wrap rendering is currently enabled.
func (m *LogViewer) Wrap() bool {
	return m.wrap
}

// SetWrap toggles soft-wrap rendering and re-renders so the new
// wrapping takes effect on the current buffer.
func (m *LogViewer) SetWrap(enabled bool) {
	m.wrap = enabled
	m.viewport.SoftWrap = enabled
	m.refresh()
}

// Init satisfies tea.Model. LogViewer produces no initial commands.
func (m *LogViewer) Init() tea.Cmd {
	return nil
}

// Update handles AppendMsg, WindowSizeMsg, focus-gated KeyPressMsg via
// the configured [KeyMap], and MouseWheelMsg. All other messages are
// dropped.
//
// Key dispatch is gated on focus: when the widget is blurred, a
// [tea.KeyPressMsg] is dropped without effect. Non-key messages
// (AppendMsg, WindowSizeMsg, MouseWheelMsg) are always processed so
// background producers and the program loop keep working regardless
// of which widget currently owns the focus ring.
func (m *LogViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AppendMsg:
		m.Append(msg.Lines...)
		return m, nil
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		return m.handleKey(msg)
	case tea.MouseWheelMsg:
		return m.handleMouseWheel(msg)
	}
	return m, nil
}

// handleKey matches the pressed key against the configured [KeyMap]
// and applies the corresponding action. LineUp / PageUp / Home pause
// follow and delegate to the viewport; LineDown / PageDown delegate
// without altering follow; End re-engages follow and jumps to the
// tail; ToggleFollow / ToggleWrap flip the respective fields; Clear
// empties the buffer.
//
// A key that doesn't match any binding falls through to the viewport
// so unmapped motions (e.g. the viewport's own half-page bindings)
// keep working. After such a delegation, if the viewport lifts off
// the tail, follow is flipped off; landing on the tail does NOT
// re-engage follow (only the explicit End / ToggleFollow actions do).
func (m *LogViewer) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.ToggleFollow):
		m.SetFollowing(!m.follow)
		return m, nil
	case key.Matches(msg, m.keys.ToggleWrap):
		m.SetWrap(!m.wrap)
		return m, nil
	case key.Matches(msg, m.keys.Clear):
		m.Clear()
		return m, nil
	case key.Matches(msg, m.keys.End):
		m.SetFollowing(true)
		return m, nil
	case key.Matches(msg, m.keys.Home):
		m.follow = false
		m.viewport.GotoTop()
		return m, nil
	case key.Matches(msg, m.keys.PageUp):
		m.follow = false
		m.viewport.PageUp()
		return m, nil
	case key.Matches(msg, m.keys.PageDown):
		m.viewport.PageDown()
		if !m.viewport.AtBottom() {
			m.follow = false
		}
		return m, nil
	case key.Matches(msg, m.keys.LineUp):
		m.follow = false
		m.viewport.ScrollUp(1)
		return m, nil
	case key.Matches(msg, m.keys.LineDown):
		m.viewport.ScrollDown(1)
		if !m.viewport.AtBottom() {
			m.follow = false
		}
		return m, nil
	}

	// Unmapped key: delegate to the viewport so its own KeyMap (e.g.
	// half-page up/down) still works. Pause follow if the viewport
	// moves off the tail; do NOT re-engage follow on tail landings.
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
		footer = theme.LogFooterRow().Width(m.width).Render(footer)
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

// Focus marks the widget as owning the focus ring. Subsequent
// [tea.KeyPressMsg] messages will be dispatched through the configured
// [KeyMap]; non-key messages are unaffected by focus.
func (m *LogViewer) Focus() {
	m.focused = true
}

// Blur clears the focus flag. Subsequent [tea.KeyPressMsg] messages
// are dropped until Focus is called again. Non-key messages (appends,
// resizes, mouse wheel) are still processed.
func (m *LogViewer) Blur() {
	m.focused = false
}

// Focused reports whether the widget currently owns the focus ring.
func (m *LogViewer) Focused() bool {
	return m.focused
}

// KeyMap returns a copy of the widget's active key bindings. Useful
// for help-bar integrations that want to enumerate what's currently
// bound.
func (m *LogViewer) KeyMap() KeyMap {
	return m.keys
}

// refresh re-renders the retained lines into the viewport. Called on
// every mutation that affects on-screen content (Append, Clear,
// SetSize, SetWrap).
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
		out     strings.Builder
		lineBuf strings.Builder
	)
	wrapStyle := theme.LogBody()
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

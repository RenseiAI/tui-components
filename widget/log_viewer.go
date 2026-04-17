package widget

import (
	"strings"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"

	"github.com/RenseiAI/tui-components/component"
)

// defaultMaxLines is the ring-buffer cap applied when no explicit
// WithMaxLines option is provided.
const defaultMaxLines = 10_000

// Compile-time assertion that *LogViewer satisfies component.Component.
var _ component.Component = (*LogViewer)(nil)

// LogViewer is a Bubble Tea widget that renders a stream of log lines
// inside a scrollable viewport. Lines are retained in a ring buffer
// (bounded by WithMaxLines) and rendered with ANSI SGR styling via the
// parser in widget/ansi.go.
//
// LogViewer is the substrate for a pausable / follow-tailing log pane.
// Follow/scroll-lock state transitions, key bindings, and focus-gated
// dispatch live in sibling tickets (REN-996 / REN-997); this type only
// exposes their getters/setters and stub Focus/Blur.
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

// Update handles the minimal set of messages for this issue: AppendMsg
// and tea.WindowSizeMsg. All other messages are dropped (no-op).
// Follow/scroll-lock transitions and key dispatch are wired in
// sibling tickets.
func (m *LogViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case AppendMsg:
		m.Append(msg.Lines...)
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

// View renders the viewport as a tea.View.
func (m *LogViewer) View() tea.View {
	return tea.NewView(m.viewport.View())
}

// SetSize propagates new dimensions to the embedded viewport and
// re-renders (wrap width changed). If follow is on, the viewport is
// pinned back to the bottom afterwards.
func (m *LogViewer) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(height)
	m.refresh()
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

package widget

import (
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that Spinner satisfies component.Component.
var _ component.Component = (*Spinner)(nil)

// Spinner is a themed animated spinner widget wrapping
// charm.land/bubbles/v2/spinner. It implements component.Component and
// renders the current frame using the accent color from the active theme.
// When a label is set, the output is "<frame> <label>".
type Spinner struct {
	inner   spinner.Model
	label   string
	focused bool
	t       theme.Theme
}

// SpinnerOption configures a Spinner during construction.
type SpinnerOption func(*Spinner)

// WithSpinnerStyle sets the Bubbles spinner animation style (e.g.
// spinner.Line, spinner.Dot, spinner.MiniDot, spinner.Jump, spinner.Pulse,
// spinner.Points, spinner.Globe, spinner.Moon, spinner.Monkey, spinner.Meter,
// spinner.Hamburger, spinner.Ellipsis).
func WithSpinnerStyle(s spinner.Spinner) SpinnerOption {
	return func(sp *Spinner) {
		sp.inner.Spinner = s
	}
}

// WithSpinnerLabel sets the label rendered next to the spinner frame. An
// empty label causes only the frame to be rendered.
func WithSpinnerLabel(label string) SpinnerOption {
	return func(sp *Spinner) {
		sp.label = label
	}
}

// WithSpinnerTheme sets the Theme used by the spinner for colors and styles.
// Calling this option mid-render updates the internal theme; the next View
// call uses the new theme.  The default is [theme.DefaultTheme].
func WithSpinnerTheme(t theme.Theme) SpinnerOption {
	return func(sp *Spinner) {
		sp.t = t
		sp.inner.Style = lipgloss.NewStyle().Foreground(t.Accent)
	}
}

// NewSpinner constructs a Spinner. By default the spinner uses
// spinner.Line, has no label, is focused (animating), and its frame is
// styled with the accent color from the active theme.
//
// Pass [WithSpinnerTheme] to override the theme, or [WithTheme] (the universal
// widget option) — both are accepted.
func NewSpinner(opts ...SpinnerOption) *Spinner {
	t := theme.DefaultTheme()
	inner := spinner.New(spinner.WithStyle(lipgloss.NewStyle().Foreground(t.Accent)))
	s := &Spinner{
		inner:   inner,
		focused: true,
		t:       t,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// SetLabel updates the label rendered next to the spinner frame.
func (s *Spinner) SetLabel(label string) {
	s.label = label
}

// SetStyle updates the Bubbles spinner animation style.
func (s *Spinner) SetStyle(style spinner.Spinner) {
	s.inner.Spinner = style
}

// SetTheme updates the theme used for rendering. The change takes effect on
// the next call to View.
func (s *Spinner) SetTheme(t theme.Theme) {
	s.t = t
	s.inner.Style = lipgloss.NewStyle().Foreground(t.Accent)
}

// Init returns the initial command that starts the spinner animation. It
// forwards to the inner spinner's tick command so that the spinner begins
// advancing frames as soon as the Bubble Tea program starts.
func (s *Spinner) Init() tea.Cmd {
	return s.inner.Tick
}

// Update forwards messages to the inner spinner model. When the Spinner is
// blurred, spinner.TickMsg messages are dropped so that the animation
// freezes on the current frame. All other messages pass through unchanged.
// Update always returns the receiver itself (as tea.Model) so that
// consumers can continue treating *Spinner as the component.
func (s *Spinner) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(spinner.TickMsg); ok && !s.focused {
		return s, nil
	}
	var cmd tea.Cmd
	s.inner, cmd = s.inner.Update(msg)
	return s, cmd
}

// View renders the spinner. The current frame is styled with the accent color
// from the active theme. When a label is set, the output is
// "<styled-frame> <styled-label>" where the label uses the theme's TextPrimary;
// otherwise only the styled frame is rendered.
func (s *Spinner) View() tea.View {
	frame := lipgloss.NewStyle().Foreground(s.t.Accent).Render(s.inner.View())
	if s.label == "" {
		return tea.NewView(frame)
	}
	label := lipgloss.NewStyle().Foreground(s.t.TextPrimary).Render(s.label)
	return tea.NewView(frame + " " + label)
}

// SetSize is a no-op. The Spinner renders a single short line whose width
// is determined by the current frame and optional label, so external size
// hints are ignored.
func (s *Spinner) SetSize(width, height int) {
	_ = width
	_ = height
}

// Focus resumes the spinner animation. While focused, spinner.TickMsg
// messages advance the frame. A Spinner is focused by default after
// construction.
func (s *Spinner) Focus() {
	s.focused = true
}

// Blur pauses the spinner animation. While blurred, the current frame is
// still rendered but spinner.TickMsg messages are ignored so the animation
// does not advance.
func (s *Spinner) Blur() {
	s.focused = false
}

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
// renders the current frame using the accent color from the theme package.
// When a label is set, the output is "<frame> <label>".
type Spinner struct {
	inner   spinner.Model
	label   string
	focused bool
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

// NewSpinner constructs a Spinner. By default the spinner uses
// spinner.Line, has no label, is focused (animating), and its frame is
// styled with theme.SpinnerStyle.
func NewSpinner(opts ...SpinnerOption) *Spinner {
	inner := spinner.New(spinner.WithStyle(theme.SpinnerStyle()))
	s := &Spinner{
		inner:   inner,
		focused: true,
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

// View renders the spinner. The current frame is styled with
// theme.SpinnerStyle. When a label is set, the output is
// "<styled-frame> <styled-label>" where the label uses theme.TextPrimary;
// otherwise only the styled frame is rendered.
func (s *Spinner) View() tea.View {
	frame := theme.SpinnerStyle().Render(s.inner.View())
	if s.label == "" {
		return tea.NewView(frame)
	}
	label := lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(s.label)
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

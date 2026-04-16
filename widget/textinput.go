package widget

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

var _ component.Component = (*TextInput)(nil)

// ValidateFunc reports an error for an invalid input value, or nil if valid.
type ValidateFunc func(string) error

// TextInputOption configures a TextInput at construction time.
type TextInputOption func(*TextInput)

// WithPlaceholder sets the placeholder text shown when the input is empty.
func WithPlaceholder(s string) TextInputOption {
	return func(t *TextInput) {
		t.model.Placeholder = s
	}
}

// WithValidate sets the validation callback invoked on every change.
// The most recent error is exposed via Err and rendered below the input.
func WithValidate(fn ValidateFunc) TextInputOption {
	return func(t *TextInput) {
		t.validate = fn
	}
}

// WithCharLimit caps the maximum number of characters accepted.
// A value of 0 or less disables the limit.
func WithCharLimit(n int) TextInputOption {
	return func(t *TextInput) {
		t.model.CharLimit = n
	}
}

// WithWidth sets the display width in cells. Panics on negative values.
func WithWidth(n int) TextInputOption {
	if n < 0 {
		panic("widget.WithWidth: width must be non-negative")
	}
	return func(t *TextInput) {
		t.width = n
		t.applyInnerWidth()
	}
}

// TextInput is a styled wrapper around bubbles/v2/textinput that implements
// component.Component and surfaces a validation error below the field.
//
// The zero value is not usable; construct with NewTextInput.
type TextInput struct {
	model    textinput.Model
	validate ValidateFunc
	err      error
	width    int
}

// NewTextInput constructs a TextInput. Apply options for placeholder, validation, etc.
func NewTextInput(opts ...TextInputOption) TextInput {
	t := TextInput{
		model: textinput.New(),
	}
	t.model.Prompt = ""
	for _, opt := range opts {
		opt(&t)
	}
	return t
}

// Init implements tea.Model. It returns no initial command.
func (t TextInput) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. It forwards the message to the embedded
// textinput and, on key presses, runs the configured ValidateFunc against
// the resulting value, caching the error for later retrieval via Err.
func (t TextInput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	t.model, cmd = t.model.Update(msg)
	if _, ok := msg.(tea.KeyPressMsg); ok && t.validate != nil {
		t.err = t.validate(t.model.Value())
	}
	return t, cmd
}

// View implements tea.Model. It renders the bordered input field and, when
// the value is non-empty and the validation callback has reported an error,
// an inline error message on the line below.
func (t TextInput) View() tea.View {
	field := fieldStyle().Render(t.model.View())
	var content string
	if t.err == nil || t.model.Value() == "" {
		content = field
	} else {
		content = lipgloss.JoinVertical(
			lipgloss.Left,
			field,
			theme.ErrorText().Render(t.err.Error()),
		)
	}
	return tea.NewView(content)
}

// SetSize sets the widget width in cells. The height argument is ignored
// because the input is always single-line.
func (t *TextInput) SetSize(width, _ int) {
	if width < 0 {
		width = 0
	}
	t.width = width
	t.applyInnerWidth()
}

// Focus activates the input so it receives keyboard events and shows a
// cursor. The cursor-blink command from the embedded model is intentionally
// discarded to keep the Component interface void-returning; consumers that
// want blinking can drive it via Init on the returned model.
func (t *TextInput) Focus() {
	_ = t.model.Focus()
}

// Blur deactivates the input so it stops receiving keyboard events and
// hides its cursor.
func (t *TextInput) Blur() {
	t.model.Blur()
}

// Value returns the current input value.
func (t TextInput) Value() string {
	return t.model.Value()
}

// SetValue replaces the current input value and re-runs validation.
func (t *TextInput) SetValue(s string) {
	t.model.SetValue(s)
	if t.validate != nil {
		t.err = t.validate(t.model.Value())
	}
}

// Reset clears the input value and any cached validation error.
func (t *TextInput) Reset() {
	t.model.Reset()
	t.err = nil
}

// Err returns the last validation error, or nil if the input is valid or
// no ValidateFunc was configured.
func (t TextInput) Err() error {
	return t.err
}

// fieldStyle is the slim bordered frame around the input. It mirrors
// theme.CardBorder but drops vertical padding so the widget stays
// single-line.
func fieldStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.SurfaceBorder).
		Padding(0, 1)
}

// fieldHorizontalOverhead is the number of cells the bordered field adds
// around the inner textinput (left border + left padding + right padding +
// right border).
const fieldHorizontalOverhead = 4

// applyInnerWidth sets the embedded textinput width so the rendered field
// fits within the widget width. A width of 0 disables the inner width cap.
func (t *TextInput) applyInnerWidth() {
	if t.width <= 0 {
		t.model.SetWidth(0)
		return
	}
	inner := t.width - fieldHorizontalOverhead
	if inner < 1 {
		inner = 1
	}
	t.model.SetWidth(inner)
}

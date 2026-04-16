package widget

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Result represents the outcome of a dialog interaction.
type Result int

const (
	// ResultNone indicates that the dialog has not produced a result yet.
	ResultNone Result = iota
	// ResultYes indicates that the user selected the affirmative button.
	ResultYes
	// ResultNo indicates that the user selected the negative button.
	ResultNo
	// ResultCancel indicates that the user cancelled the dialog.
	ResultCancel
)

// String returns a human-readable name for the Result.
func (r Result) String() string {
	switch r {
	case ResultYes:
		return "yes"
	case ResultNo:
		return "no"
	case ResultCancel:
		return "cancel"
	case ResultNone:
		fallthrough
	default:
		return "none"
	}
}

// Button describes a single selectable action in a Dialog.
type Button struct {
	// Label is the text shown on the button.
	Label string
	// Result is the Result emitted when this button is activated.
	Result Result
}

// DialogDoneMsg is emitted as a tea.Cmd payload when a Dialog is dismissed
// by activating a button or by pressing escape.
type DialogDoneMsg struct {
	// Result is the Result that was selected.
	Result Result
}

// Option configures a Dialog in the functional-options style.
type Option func(*Dialog)

// WithTitle sets the title shown at the top of the Dialog box.
func WithTitle(title string) Option {
	return func(d *Dialog) {
		d.title = title
	}
}

// WithBody sets the body text displayed inside the Dialog.
func WithBody(body string) Option {
	return func(d *Dialog) {
		d.body = body
	}
}

// WithYesLabel overrides the label of the default Yes button.
func WithYesLabel(label string) Option {
	return func(d *Dialog) {
		d.yesLabel = label
	}
}

// WithNoLabel overrides the label of the default No button.
func WithNoLabel(label string) Option {
	return func(d *Dialog) {
		d.noLabel = label
	}
}

// WithCancelLabel overrides the label of the default Cancel button.
func WithCancelLabel(label string) Option {
	return func(d *Dialog) {
		d.cancelLabel = label
	}
}

// WithButtons replaces the default Yes/No/Cancel button set with a custom
// sequence of buttons. The first button is focused initially.
func WithButtons(buttons ...Button) Option {
	return func(d *Dialog) {
		if len(buttons) == 0 {
			return
		}
		d.buttons = append(d.buttons[:0], buttons...)
		d.customButtons = true
	}
}

// Dialog is a modal widget presenting a title, body text, and a row of
// buttons. It implements [component.Component] and can be rendered standalone
// via View. Composition onto a parent view is handled separately by the
// overlay helper.
type Dialog struct {
	title  string
	body   string
	width  int
	height int

	buttons       []Button
	customButtons bool
	yesLabel      string
	noLabel       string
	cancelLabel   string

	focusIndex int
	focused    bool
	result     Result
}

// compile-time assertion that *Dialog satisfies component.Component.
var _ component.Component = (*Dialog)(nil)

// Default button labels used when none are supplied via options.
const (
	defaultYesLabel    = "Yes"
	defaultNoLabel     = "No"
	defaultCancelLabel = "Cancel"
)

// New constructs a Dialog with default Yes/No/Cancel buttons. Options may
// override labels, provide a custom button set, and set title/body text.
func New(opts ...Option) *Dialog {
	d := &Dialog{
		yesLabel:    defaultYesLabel,
		noLabel:     defaultNoLabel,
		cancelLabel: defaultCancelLabel,
		result:      ResultNone,
	}
	for _, opt := range opts {
		opt(d)
	}
	if !d.customButtons {
		d.buttons = []Button{
			{Label: d.yesLabel, Result: ResultYes},
			{Label: d.noLabel, Result: ResultNo},
			{Label: d.cancelLabel, Result: ResultCancel},
		}
	}
	return d
}

// Init implements tea.Model. Dialog has no startup work.
func (d *Dialog) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model. It handles focus navigation between buttons,
// activation via Enter, quick-activation via y/n, and cancellation via Esc.
func (d *Dialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return d, nil
	}
	switch key.String() {
	case "left", "shift+tab":
		d.focusPrev()
		return d, nil
	case "right", "tab":
		d.focusNext()
		return d, nil
	case "enter":
		return d, d.activateFocused()
	case "esc":
		d.result = ResultCancel
		return d, dialogDoneCmd(ResultCancel)
	case "y":
		if idx, found := d.buttonIndex(ResultYes); found {
			d.focusIndex = idx
			return d, d.activateFocused()
		}
		return d, nil
	case "n":
		if idx, found := d.buttonIndex(ResultNo); found {
			d.focusIndex = idx
			return d, d.activateFocused()
		}
		return d, nil
	default:
		return d, nil
	}
}

// View implements tea.Model. It renders a bordered, centered dialog box with
// a title, wrapped body text, and a row of buttons. An empty view is
// returned when the dialog has no content and no configured buttons.
func (d *Dialog) View() tea.View {
	return tea.NewView(d.Render())
}

// Render produces the dialog's content as a styled string without wrapping
// it in a tea.View. Useful for composition helpers such as Overlay.
func (d *Dialog) Render() string {
	box := d.renderBox()
	if box == "" {
		return ""
	}
	if d.width <= 0 || d.height <= 0 {
		return box
	}
	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, box)
}

// Overlay composes the dialog as a centered overlay on top of the given
// background string, with a dimmed backdrop. The background is expected to
// be a fully-rendered terminal view (may contain ANSI escape sequences).
//
// If SetSize has not been called (width or height is 0), Overlay falls back
// to returning the dialog box alone without a backdrop.
//
// Overlay is pure: it does not mutate any Dialog state.
func (d *Dialog) Overlay(background string) string {
	box := d.renderBox()
	if box == "" {
		return background
	}

	// Degrade gracefully when SetSize has not been called.
	if d.width <= 0 || d.height <= 0 {
		return box
	}

	// --- Dim the backdrop ------------------------------------------------
	dimStyle := theme.Dimmed()
	bgLines := strings.Split(background, "\n")

	// Ensure the backdrop fills the full viewport so the overlay has a
	// consistent canvas to composite onto.
	for len(bgLines) < d.height {
		bgLines = append(bgLines, "")
	}
	bgLines = bgLines[:d.height]

	for i, line := range bgLines {
		stripped := ansi.Strip(line)
		bgLines[i] = dimStyle.Render(stripped)
	}

	// --- Measure the box and compute centering offsets -------------------
	boxLines := strings.Split(box, "\n")
	boxH := len(boxLines)
	boxW := 0
	for _, l := range boxLines {
		if w := lipgloss.Width(l); w > boxW {
			boxW = w
		}
	}

	// Clamp box dimensions to viewport.
	if boxH > d.height {
		boxLines = boxLines[:d.height]
		boxH = d.height
	}

	top := (d.height - boxH) / 2
	if top < 0 {
		top = 0
	}

	// --- Line-level composition ------------------------------------------
	for i, bLine := range boxLines {
		row := top + i
		if row >= 0 && row < len(bgLines) {
			bgLines[row] = bLine
		}
	}

	return strings.Join(bgLines, "\n")
}

// renderBox builds the bordered dialog card without any viewport placement.
// It is shared by Render (which may wrap it with lipgloss.Place) and Overlay
// (which composites it manually onto a backdrop).
func (d *Dialog) renderBox() string {
	innerWidth := d.innerContentWidth()

	var sections []string
	if d.title != "" {
		sections = append(sections, theme.SectionTitle().Render(d.title))
	}
	if d.body != "" {
		body := d.body
		if innerWidth > 0 {
			body = lipgloss.Wrap(d.body, innerWidth, "")
		}
		sections = append(sections, body)
	}
	if len(d.buttons) > 0 {
		sections = append(sections, d.renderButtons())
	}

	if len(sections) == 0 {
		return ""
	}

	content := lipgloss.JoinVertical(lipgloss.Left, interleaveBlank(sections)...)
	return theme.CardBorder().Render(content)
}

// SetSize stores the outer dimensions used for centering and body wrapping.
// Re-wrapping happens lazily on the next View call.
func (d *Dialog) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	if height < 0 {
		height = 0
	}
	d.width = width
	d.height = height
}

// Focus marks the dialog as focused. The dialog always consumes its own keys
// while visible; focus state is primarily informational.
func (d *Dialog) Focus() {
	d.focused = true
}

// Blur marks the dialog as not focused.
func (d *Dialog) Blur() {
	d.focused = false
}

// Focused reports whether the dialog currently has focus.
func (d *Dialog) Focused() bool {
	return d.focused
}

// Result returns the Result currently set on the dialog. Before any button
// is activated the value is ResultNone.
func (d *Dialog) Result() Result {
	return d.result
}

// Reset clears the Result back to ResultNone and restores initial focus to
// the first button. Title, body, and button configuration are preserved.
func (d *Dialog) Reset() {
	d.result = ResultNone
	d.focusIndex = 0
}

// Buttons returns a copy of the dialog's configured buttons, in order.
func (d *Dialog) Buttons() []Button {
	out := make([]Button, len(d.buttons))
	copy(out, d.buttons)
	return out
}

// FocusedIndex returns the index of the currently focused button.
func (d *Dialog) FocusedIndex() int {
	return d.focusIndex
}

// focusNext advances the focused button index, wrapping around.
func (d *Dialog) focusNext() {
	if len(d.buttons) == 0 {
		return
	}
	d.focusIndex = (d.focusIndex + 1) % len(d.buttons)
}

// focusPrev moves the focused button index backwards, wrapping around.
func (d *Dialog) focusPrev() {
	if len(d.buttons) == 0 {
		return
	}
	d.focusIndex = (d.focusIndex - 1 + len(d.buttons)) % len(d.buttons)
}

// activateFocused records the Result of the currently focused button and
// returns a tea.Cmd that emits the corresponding DialogDoneMsg. Returns nil
// if there are no buttons configured.
func (d *Dialog) activateFocused() tea.Cmd {
	if len(d.buttons) == 0 {
		return nil
	}
	if d.focusIndex < 0 || d.focusIndex >= len(d.buttons) {
		d.focusIndex = 0
	}
	result := d.buttons[d.focusIndex].Result
	d.result = result
	return dialogDoneCmd(result)
}

// buttonIndex locates the first button whose Result matches r.
func (d *Dialog) buttonIndex(r Result) (int, bool) {
	for i, b := range d.buttons {
		if b.Result == r {
			return i, true
		}
	}
	return 0, false
}

// renderButtons lays buttons out horizontally with two spaces between each.
func (d *Dialog) renderButtons() string {
	rendered := make([]string, 0, len(d.buttons)*2)
	for i, btn := range d.buttons {
		if i > 0 {
			rendered = append(rendered, "  ")
		}
		rendered = append(rendered, d.renderButton(btn, i == d.focusIndex))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

// renderButton applies the focused or unfocused style to a single button.
// The label is padded with single-space gutters to visually separate the
// button from its neighbours.
func (d *Dialog) renderButton(btn Button, focused bool) string {
	label := " " + btn.Label + " "
	if focused {
		return theme.HelpKey().Reverse(true).Render(label)
	}
	return theme.Muted().Render(label)
}

// innerContentWidth returns the usable width for content inside the
// bordered card, or 0 when no outer width is configured.
func (d *Dialog) innerContentWidth() int {
	if d.width <= 0 {
		return 0
	}
	// Borders contribute 2 columns, CardBorder padding contributes 4 columns.
	const chrome = 2 + 4
	// Prefer a box that uses roughly two-thirds of the available width, but
	// never wider than the screen and never narrower than a sensible floor.
	preferred := (d.width * 2) / 3
	if preferred < 20 {
		preferred = 20
	}
	if preferred > d.width-chrome {
		preferred = d.width - chrome
	}
	if preferred < 1 {
		return 0
	}
	return preferred
}

// interleaveBlank inserts a blank line between non-empty sections so that
// title, body, and buttons are visually separated inside the box.
func interleaveBlank(sections []string) []string {
	out := make([]string, 0, len(sections)*2)
	for i, s := range sections {
		if i > 0 {
			out = append(out, "")
		}
		out = append(out, s)
	}
	return out
}

// dialogDoneCmd builds a tea.Cmd that emits a DialogDoneMsg carrying r.
func dialogDoneCmd(r Result) tea.Cmd {
	return func() tea.Msg {
		return DialogDoneMsg{Result: r}
	}
}

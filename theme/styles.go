package theme

import "charm.land/lipgloss/v2"

// Header returns the style for the top header bar.
func Header() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(Surface).
		Bold(true).
		Padding(0, 1)
}

// StatLabel returns the style for stat labels in the stats bar.
func StatLabel() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary)
}

// StatValue returns the style for stat values.
func StatValue() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Bold(true)
}

// StatValueAccent returns the style for highlighted stat values.
func StatValueAccent() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Accent).
		Bold(true)
}

// StatValueTeal returns the style for teal-colored stat values.
func StatValueTeal() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Teal).
		Bold(true)
}

// TableHeader returns the style for table column headers.
func TableHeader() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary).
		Bold(true)
}

// TableRow returns the base style for a table row.
func TableRow() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary)
}

// TableRowSelected returns the style for the selected table row.
func TableRowSelected() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(SurfaceRaised)
}

// Muted returns the style for muted/secondary text.
func Muted() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextSecondary)
}

// Dimmed returns the style for tertiary/dimmed text.
func Dimmed() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary)
}

// HelpBar returns the style for the bottom help bar.
func HelpBar() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary).
		Background(Surface).
		Padding(0, 1)
}

// HelpKey returns the style for a key binding label in the help bar.
func HelpKey() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextSecondary).
		Bold(true)
}

// HelpDesc returns the style for a key binding description in the help bar.
func HelpDesc() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary)
}

// CardBorder returns a bordered card style.
func CardBorder() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(SurfaceBorder).
		Padding(1, 2)
}

// SectionTitle returns the style for section titles.
func SectionTitle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Bold(true)
}

// SpinnerStyle returns the style used for animated spinner frames.
// The foreground color is the accent color (the natural "active/working"
// hue in the palette).
func SpinnerStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(Accent)
}

// ErrorText returns the style for inline error messages (e.g. validation
// feedback rendered beneath an input field).
func ErrorText() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(StatusError)
}

// TabActive returns the style for the currently-active tab in a tab bar.
func TabActive() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(SurfaceRaised).
		Bold(true).
		Padding(0, 1)
}

// TabInactive returns the style for inactive tabs in a tab bar.
func TabInactive() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextSecondary).
		Padding(0, 1)
}

// TabDisabled returns the style for disabled tabs in a tab bar.
// Disabled tabs are rendered with dimmer foreground text than inactive
// tabs to signal that they cannot be activated.
func TabDisabled() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(TextTertiary).
		Padding(0, 1)
}

// TabBar returns the background style that wraps an entire tab bar.
func TabBar() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(Surface)
}

// TabSeparator returns the style for the separator glyph rendered
// between adjacent tabs in a tab bar.
func TabSeparator() lipgloss.Style {
	return lipgloss.NewStyle().
		Foreground(SurfaceBorder)
}

// LogFollow returns the style used to render the LogViewer footer
// indicator when the viewport is actively tailing new output. It wraps
// [StatValueTeal] with single-column horizontal padding so the text
// reads as a recognisable badge (" FOLLOW ").
func LogFollow() lipgloss.Style {
	return StatValueTeal().Padding(0, 1)
}

// LogPaused returns the style used to render the LogViewer footer
// indicator when the viewport is scroll-locked (not following). It
// wraps [StatValueAccent] with single-column horizontal padding so the
// text reads as a recognisable badge (" PAUSED ").
func LogPaused() lipgloss.Style {
	return StatValueAccent().Padding(0, 1)
}

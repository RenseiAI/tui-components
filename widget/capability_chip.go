package widget

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *CapabilityChip satisfies component.Component.
var _ component.Component = (*CapabilityChip)(nil)

// CapabilityChip renders a typed capability flag alongside its human-readable
// label.  It is the canonical display primitive for any typed enum flag in
// the Rensei architecture (billing model, transport model, idle-cost model,
// etc.) as described in 002-provider-base-contract.md and
// 014-tui-operator-surfaces.md.
//
// The chip renders as:   ◆ <value>  <humanLabel>
// In no-color mode:       <symbol> <value>  <humanLabel>
//
// Example usage:
//
//	chip := widget.NewCapabilityChip(
//	    widget.WithCapabilityValue("active-cpu"),
//	    widget.WithCapabilityHumanLabel("Billed for active CPU only"),
//	)
//	fmt.Println(chip.ViewString())
type CapabilityChip struct {
	value       string
	humanLabel  string
	symbol      string
	accessLabel string
	t           theme.Theme
	noColor     bool
	width       int
}

// CapabilityChipOption configures a CapabilityChip during construction.
type CapabilityChipOption func(*CapabilityChip)

// WithCapabilityValue sets the raw capability flag value (e.g. "active-cpu",
// "wall-clock", "dial-in").
func WithCapabilityValue(v string) CapabilityChipOption {
	return func(c *CapabilityChip) { c.value = v }
}

// WithCapabilityHumanLabel sets the human-readable description rendered beside
// the flag value.
func WithCapabilityHumanLabel(label string) CapabilityChipOption {
	return func(c *CapabilityChip) { c.humanLabel = label }
}

// WithCapabilitySymbol overrides the default leading symbol ("◆").
func WithCapabilitySymbol(sym string) CapabilityChipOption {
	return func(c *CapabilityChip) { c.symbol = sym }
}

// WithCapabilityNoColor forces symbol-first, no-ANSI rendering.  Equivalent
// to honoring the NO_COLOR env var at the widget level.
func WithCapabilityNoColor(nc bool) CapabilityChipOption {
	return func(c *CapabilityChip) { c.noColor = nc }
}

// WithCapabilityTheme sets the Theme used for rendering colors.
func WithCapabilityTheme(t theme.Theme) CapabilityChipOption {
	return func(c *CapabilityChip) { c.t = t }
}

// WithCapabilityAccessibleLabel sets the accessible label used by
// screen-reader consumers (REN-1332 a11y opt-in).
func WithCapabilityAccessibleLabel(label string) CapabilityChipOption {
	return func(c *CapabilityChip) { c.accessLabel = label }
}

// NewCapabilityChip constructs a CapabilityChip with the default theme.
func NewCapabilityChip(opts ...CapabilityChipOption) *CapabilityChip {
	c := &CapabilityChip{
		symbol: "◆",
		t:      theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetTheme updates the theme used for rendering.
func (c *CapabilityChip) SetTheme(t theme.Theme) { c.t = t }

// AccessibleLabel returns the accessible label set on this chip.  If no
// explicit accessible label was set, it falls back to "<value>: <humanLabel>".
func (c *CapabilityChip) AccessibleLabel() string {
	if c.accessLabel != "" {
		return c.accessLabel
	}
	if c.humanLabel != "" {
		return c.value + ": " + c.humanLabel
	}
	return c.value
}

// ViewString renders the chip as a plain string (no tea.View wrapping).
func (c *CapabilityChip) ViewString() string {
	if c.noColor {
		if c.humanLabel != "" {
			return c.symbol + " " + c.value + "  " + c.humanLabel
		}
		return c.symbol + " " + c.value
	}
	sym := lipgloss.NewStyle().Foreground(c.t.Teal).Render(c.symbol)
	val := lipgloss.NewStyle().Foreground(c.t.TextPrimary).Bold(true).Render(c.value)
	if c.humanLabel == "" {
		return strings.Join([]string{sym, val}, " ")
	}
	lbl := lipgloss.NewStyle().Foreground(c.t.TextSecondary).Render(c.humanLabel)
	return strings.Join([]string{sym, val, " ", lbl}, " ")
}

// Init satisfies tea.Model; CapabilityChip is static — no commands needed.
func (c *CapabilityChip) Init() tea.Cmd { return nil }

// Update satisfies tea.Model; CapabilityChip has no interactive state.
func (c *CapabilityChip) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return c, nil }

// View renders the chip as a tea.View.
func (c *CapabilityChip) View() tea.View { return tea.NewView(c.ViewString()) }

// SetSize is a no-op for CapabilityChip; it renders its natural width.
func (c *CapabilityChip) SetSize(width, height int) { c.width = width }

// Focus is a no-op; CapabilityChip is a display-only primitive.
func (c *CapabilityChip) Focus() {}

// Blur is a no-op; CapabilityChip is a display-only primitive.
func (c *CapabilityChip) Blur() {}

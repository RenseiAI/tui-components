package widget

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *ProviderHealthDot satisfies component.Component.
var _ component.Component = (*ProviderHealthDot)(nil)

// ProviderHealth represents the operational health state of a provider
// (SandboxProvider, WorkareaProvider, AgentRuntimeProvider, etc.) as
// referenced in 002-provider-base-contract.md.
type ProviderHealth string

const (
	// ProviderHealthReady indicates the provider is operational.
	ProviderHealthReady ProviderHealth = "ready"
	// ProviderHealthDegraded indicates the provider is partially operational
	// or experiencing elevated latency/errors.
	ProviderHealthDegraded ProviderHealth = "degraded"
	// ProviderHealthUnhealthy indicates the provider is not operational.
	ProviderHealthUnhealthy ProviderHealth = "unhealthy"
)

// ProviderHealthDot renders a compact single-rune health indicator dot for a
// provider.  The dot pairs a Unicode symbol with a color so that the status
// is unambiguous in both color and symbol-only (NO_COLOR) modes.
//
//	● ready        (green)
//	◐ degraded     (yellow)
//	✗ unhealthy    (red)
type ProviderHealthDot struct {
	health      ProviderHealth
	showLabel   bool
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// ProviderHealthDotOption configures a ProviderHealthDot during construction.
type ProviderHealthDotOption func(*ProviderHealthDot)

// WithProviderHealth sets the health state.
func WithProviderHealth(h ProviderHealth) ProviderHealthDotOption {
	return func(d *ProviderHealthDot) { d.health = h }
}

// WithProviderHealthShowLabel includes the state label after the dot symbol.
func WithProviderHealthShowLabel(show bool) ProviderHealthDotOption {
	return func(d *ProviderHealthDot) { d.showLabel = show }
}

// WithProviderHealthTheme sets the Theme.
func WithProviderHealthTheme(t theme.Theme) ProviderHealthDotOption {
	return func(d *ProviderHealthDot) { d.t = t }
}

// WithProviderHealthNoColor forces symbol-only, no-ANSI rendering.
func WithProviderHealthNoColor(nc bool) ProviderHealthDotOption {
	return func(d *ProviderHealthDot) { d.noColor = nc }
}

// WithProviderHealthAccessibleLabel sets the accessible label.
func WithProviderHealthAccessibleLabel(label string) ProviderHealthDotOption {
	return func(d *ProviderHealthDot) { d.accessLabel = label }
}

// NewProviderHealthDot constructs a ProviderHealthDot with default theme.
func NewProviderHealthDot(opts ...ProviderHealthDotOption) *ProviderHealthDot {
	d := &ProviderHealthDot{
		health: ProviderHealthReady,
		t:      theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// SetTheme updates the theme.
func (d *ProviderHealthDot) SetTheme(t theme.Theme) { d.t = t }

// AccessibleLabel returns the accessible label or a fallback.
func (d *ProviderHealthDot) AccessibleLabel() string {
	if d.accessLabel != "" {
		return d.accessLabel
	}
	return "health: " + string(d.health)
}

type healthVisual struct {
	symbol string
	color  color.Color
	label  string
}

func (d *ProviderHealthDot) visual() healthVisual {
	switch d.health {
	case ProviderHealthDegraded:
		return healthVisual{"◐", d.t.StatusWarning, "degraded"}
	case ProviderHealthUnhealthy:
		return healthVisual{"✗", d.t.StatusError, "unhealthy"}
	default:
		return healthVisual{"●", d.t.StatusSuccess, "ready"}
	}
}

// ViewString renders the dot as a plain string.
func (d *ProviderHealthDot) ViewString() string {
	v := d.visual()
	if d.noColor {
		if d.showLabel {
			return v.symbol + " " + v.label
		}
		return v.symbol
	}
	sym := lipgloss.NewStyle().Foreground(v.color).Render(v.symbol)
	if d.showLabel {
		lbl := lipgloss.NewStyle().Foreground(d.t.TextSecondary).Render(v.label)
		return sym + " " + lbl
	}
	return sym
}

// Init satisfies tea.Model.
func (d *ProviderHealthDot) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (d *ProviderHealthDot) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return d, nil }

// View renders the dot as a tea.View.
func (d *ProviderHealthDot) View() tea.View { return tea.NewView(d.ViewString()) }

// SetSize is a no-op.
func (d *ProviderHealthDot) SetSize(width, height int) {}

// Focus is a no-op.
func (d *ProviderHealthDot) Focus() {}

// Blur is a no-op.
func (d *ProviderHealthDot) Blur() {}

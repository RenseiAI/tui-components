package widget

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *SandboxCapacityGauge satisfies component.Component.
var _ component.Component = (*SandboxCapacityGauge)(nil)

// SandboxCapacityGauge renders a SandboxProvider's concurrent / max capacity
// as a utilization bar.  When maxConcurrent is 0 (nil in the architecture),
// the gauge degenerates to "X / ∞" with no bar.  Per
// 004-sandbox-capability-matrix.md and 014-tui-operator-surfaces.md.
//
// Example renderings:
//
//	████████░░  5 / 8  (62%)
//	0 / ∞
type SandboxCapacityGauge struct {
	current     int
	max         int // 0 = unlimited (maxConcurrent: null)
	barWidth    int
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// SandboxCapacityGaugeOption configures a SandboxCapacityGauge.
type SandboxCapacityGaugeOption func(*SandboxCapacityGauge)

// WithGaugeCurrent sets the current concurrent session count.
func WithGaugeCurrent(n int) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.current = n }
}

// WithGaugeMax sets the maximum concurrent session count.  Pass 0 to
// indicate unlimited (maxConcurrent: null).
func WithGaugeMax(n int) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.max = n }
}

// WithGaugeBarWidth sets the pixel/cell width of the utilization bar.
// Default is 10.
func WithGaugeBarWidth(w int) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.barWidth = w }
}

// WithGaugeTheme sets the Theme.
func WithGaugeTheme(t theme.Theme) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.t = t }
}

// WithGaugeNoColor forces symbol-first, no-ANSI rendering.
func WithGaugeNoColor(nc bool) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.noColor = nc }
}

// WithGaugeAccessibleLabel sets the accessible label.
func WithGaugeAccessibleLabel(label string) SandboxCapacityGaugeOption {
	return func(g *SandboxCapacityGauge) { g.accessLabel = label }
}

// NewSandboxCapacityGauge constructs a SandboxCapacityGauge with default theme.
func NewSandboxCapacityGauge(opts ...SandboxCapacityGaugeOption) *SandboxCapacityGauge {
	g := &SandboxCapacityGauge{
		barWidth: 10,
		t:        theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// SetTheme updates the theme.
func (g *SandboxCapacityGauge) SetTheme(t theme.Theme) { g.t = t }

// AccessibleLabel returns the accessible label.
func (g *SandboxCapacityGauge) AccessibleLabel() string {
	if g.accessLabel != "" {
		return g.accessLabel
	}
	if g.max == 0 {
		return fmt.Sprintf("capacity: %d / unlimited", g.current)
	}
	return fmt.Sprintf("capacity: %d / %d", g.current, g.max)
}

// ratioString formats the numeric ratio text.
func (g *SandboxCapacityGauge) ratioString() string {
	if g.max == 0 {
		return fmt.Sprintf("%d / ∞", g.current)
	}
	return fmt.Sprintf("%d / %d", g.current, g.max)
}

// ViewString renders the gauge as a plain string.
func (g *SandboxCapacityGauge) ViewString() string {
	ratio := g.ratioString()

	// Unlimited: no bar.
	if g.max == 0 {
		if g.noColor {
			return ratio
		}
		return lipgloss.NewStyle().Foreground(g.t.TextSecondary).Render(ratio)
	}

	// Compute fraction safely.
	frac := 0.0
	if g.max > 0 {
		frac = float64(g.current) / float64(g.max)
		if frac > 1 {
			frac = 1
		}
	}

	pct := fmt.Sprintf("(%d%%)", int(frac*100))
	bar := buildCapacityBar(frac, g.barWidth)

	if g.noColor {
		return bar + "  " + ratio + "  " + pct
	}

	barColor := g.t.Teal
	if frac >= 0.9 {
		barColor = g.t.StatusError
	} else if frac >= 0.7 {
		barColor = g.t.StatusWarning
	}

	barStr := lipgloss.NewStyle().Foreground(barColor).Render(bar)
	ratioStr := lipgloss.NewStyle().Foreground(g.t.TextPrimary).Render(ratio)
	pctStr := lipgloss.NewStyle().Foreground(g.t.TextTertiary).Render(pct)
	return strings.Join([]string{barStr, ratioStr, pctStr}, "  ")
}

// buildCapacityBar builds a block-character bar of the given width.
func buildCapacityBar(frac float64, width int) string {
	filled := int(frac * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// Init satisfies tea.Model.
func (g *SandboxCapacityGauge) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (g *SandboxCapacityGauge) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return g, nil }

// View renders the gauge as a tea.View.
func (g *SandboxCapacityGauge) View() tea.View { return tea.NewView(g.ViewString()) }

// SetSize is a no-op.
func (g *SandboxCapacityGauge) SetSize(width, height int) {}

// Focus is a no-op.
func (g *SandboxCapacityGauge) Focus() {}

// Blur is a no-op.
func (g *SandboxCapacityGauge) Blur() {}

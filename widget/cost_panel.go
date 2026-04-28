package widget

import (
	"fmt"
	"image/color"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *CostPanel satisfies component.Component.
var _ component.Component = (*CostPanel)(nil)

// CostTrend indicates whether costs are rising, falling, or flat.
type CostTrend string

const (
	// CostTrendUp indicates costs are increasing.
	CostTrendUp CostTrend = "up"
	// CostTrendDown indicates costs are decreasing.
	CostTrendDown CostTrend = "down"
	// CostTrendFlat indicates costs are unchanged.
	CostTrendFlat CostTrend = "flat"
)

// CostBreakdown represents a single cost line item in the panel.
type CostBreakdown struct {
	// Label is the display name for this cost category.
	Label string
	// Amount is the cost in the smallest currency unit (e.g. USD cents).
	Amount float64
	// Currency is the currency code (e.g. "USD").
	Currency string
}

// CostPanel renders a per-session / per-issue / per-tenant cost breakdown
// with an optional trend indicator.  It taps idleCostModel and billingModel
// semantics from 006-cross-provider-interactions.md Seam 4.
//
// Example rendering:
//
//	Cost summary
//	  Session     $0.42  USD
//	  Issue       $1.87  USD  ↑
//	  Tenant      $42.00 USD
type CostPanel struct {
	title       string
	items       []CostBreakdown
	trend       CostTrend
	billingNote string // e.g. "active-cpu billing"
	accessLabel string
	t           theme.Theme
	noColor     bool
	width       int
}

// CostPanelOption configures a CostPanel during construction.
type CostPanelOption func(*CostPanel)

// WithCostTitle sets the panel title.
func WithCostTitle(title string) CostPanelOption {
	return func(p *CostPanel) { p.title = title }
}

// WithCostItems sets the cost line items.
func WithCostItems(items ...CostBreakdown) CostPanelOption {
	return func(p *CostPanel) { p.items = items }
}

// WithCostTrend sets the trend indicator shown on the last item.
func WithCostTrend(t CostTrend) CostPanelOption {
	return func(p *CostPanel) { p.trend = t }
}

// WithCostBillingNote sets a billing model annotation (e.g. "active-cpu billing").
func WithCostBillingNote(note string) CostPanelOption {
	return func(p *CostPanel) { p.billingNote = note }
}

// WithCostPanelTheme sets the Theme.
func WithCostPanelTheme(t theme.Theme) CostPanelOption {
	return func(p *CostPanel) { p.t = t }
}

// WithCostPanelNoColor forces symbol-first, no-ANSI rendering.
func WithCostPanelNoColor(nc bool) CostPanelOption {
	return func(p *CostPanel) { p.noColor = nc }
}

// WithCostPanelAccessibleLabel sets the accessible label.
func WithCostPanelAccessibleLabel(label string) CostPanelOption {
	return func(p *CostPanel) { p.accessLabel = label }
}

// NewCostPanel constructs a CostPanel with default theme.
func NewCostPanel(opts ...CostPanelOption) *CostPanel {
	p := &CostPanel{
		title: "Cost summary",
		trend: CostTrendFlat,
		t:     theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// SetTheme updates the theme.
func (p *CostPanel) SetTheme(t theme.Theme) { p.t = t }

// AccessibleLabel returns the accessible label.
func (p *CostPanel) AccessibleLabel() string {
	if p.accessLabel != "" {
		return p.accessLabel
	}
	return p.title
}

// trendSymbol returns the trend indicator symbol and its color.
func (p *CostPanel) trendSymbol() (string, color.Color) {
	switch p.trend {
	case CostTrendUp:
		return "↑", p.t.StatusError
	case CostTrendDown:
		return "↓", p.t.StatusSuccess
	default:
		return "→", p.t.TextTertiary
	}
}

// ViewString renders the panel as a plain string.
func (p *CostPanel) ViewString() string {
	var sb strings.Builder

	// Title
	if p.noColor {
		sb.WriteString(p.title + "\n")
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(p.t.TextPrimary).Bold(true).Render(p.title) + "\n")
	}

	trendSym, trendColor := p.trendSymbol()

	for i, item := range p.items {
		isLast := i == len(p.items)-1
		amount := fmt.Sprintf("$%.2f", item.Amount/100.0)
		currency := item.Currency
		if currency == "" {
			currency = "USD"
		}

		trend := ""
		if isLast && p.trend != CostTrendFlat {
			if p.noColor {
				trend = "  " + trendSym
			} else {
				trend = "  " + lipgloss.NewStyle().Foreground(trendColor).Render(trendSym)
			}
		}

		if p.noColor {
			fmt.Fprintf(&sb, "  %-12s  %8s  %s%s\n",
				item.Label, amount, currency, trend)
		} else {
			label := lipgloss.NewStyle().Foreground(p.t.TextSecondary).Render(fmt.Sprintf("%-12s", item.Label))
			amountStr := lipgloss.NewStyle().Foreground(p.t.TextPrimary).Bold(true).Render(fmt.Sprintf("%8s", amount))
			currStr := lipgloss.NewStyle().Foreground(p.t.TextTertiary).Render(currency)
			fmt.Fprintf(&sb, "  %s  %s  %s%s\n", label, amountStr, currStr, trend)
		}
	}

	if p.billingNote != "" {
		if p.noColor {
			sb.WriteString("  (" + p.billingNote + ")\n")
		} else {
			note := lipgloss.NewStyle().Foreground(p.t.TextTertiary).Render("  (" + p.billingNote + ")")
			sb.WriteString(note + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Init satisfies tea.Model.
func (p *CostPanel) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (p *CostPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }

// View renders the panel as a tea.View.
func (p *CostPanel) View() tea.View { return tea.NewView(p.ViewString()) }

// SetSize stores the width hint.
func (p *CostPanel) SetSize(width, height int) { p.width = width }

// Focus is a no-op.
func (p *CostPanel) Focus() {}

// Blur is a no-op.
func (p *CostPanel) Blur() {}

package widget

import (
	"strings"
	"testing"
)

func TestCostPanel_NoColor(t *testing.T) {
	t.Parallel()
	p := NewCostPanel(
		WithCostTitle("Cost summary"),
		WithCostItems(
			CostBreakdown{Label: "Session", Amount: 4200, Currency: "USD"},
			CostBreakdown{Label: "Issue", Amount: 18700, Currency: "USD"},
		),
		WithCostTrend(CostTrendUp),
		WithCostBillingNote("active-cpu billing"),
		WithCostPanelNoColor(true),
	)
	got := p.ViewString()
	if !strings.Contains(got, "Cost summary") {
		t.Errorf("ViewString() missing title: %q", got)
	}
	if !strings.Contains(got, "Session") {
		t.Errorf("ViewString() missing Session: %q", got)
	}
	if !strings.Contains(got, "$42.00") {
		t.Errorf("ViewString() missing amount $42.00: %q", got)
	}
	if !strings.Contains(got, "active-cpu billing") {
		t.Errorf("ViewString() missing billing note: %q", got)
	}
	if !strings.Contains(got, "↑") {
		t.Errorf("ViewString() missing trend symbol: %q", got)
	}
}

func TestCostPanel_TrendDown(t *testing.T) {
	t.Parallel()
	p := NewCostPanel(
		WithCostItems(CostBreakdown{Label: "Tenant", Amount: 100, Currency: "USD"}),
		WithCostTrend(CostTrendDown),
		WithCostPanelNoColor(true),
	)
	got := p.ViewString()
	if !strings.Contains(got, "↓") {
		t.Errorf("ViewString() missing down trend: %q", got)
	}
}

func TestCostPanel_Empty(t *testing.T) {
	t.Parallel()
	p := NewCostPanel(WithCostPanelNoColor(true))
	got := p.ViewString()
	if !strings.Contains(got, "Cost summary") {
		t.Errorf("ViewString() missing default title: %q", got)
	}
}

func TestCostPanel_AccessibleLabel(t *testing.T) {
	t.Parallel()
	p := NewCostPanel(WithCostTitle("My costs"))
	if got := p.AccessibleLabel(); got != "My costs" {
		t.Errorf("AccessibleLabel() = %q, want 'My costs'", got)
	}
}

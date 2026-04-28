package widget

import (
	"strings"
	"testing"
)

func TestSandboxCapacityGauge_Unlimited(t *testing.T) {
	t.Parallel()
	g := NewSandboxCapacityGauge(
		WithGaugeCurrent(3),
		WithGaugeMax(0), // unlimited
		WithGaugeNoColor(true),
	)
	got := g.ViewString()
	if got != "3 / ∞" {
		t.Errorf("ViewString() = %q, want '3 / ∞'", got)
	}
}

func TestSandboxCapacityGauge_Bounded(t *testing.T) {
	t.Parallel()
	g := NewSandboxCapacityGauge(
		WithGaugeCurrent(5),
		WithGaugeMax(8),
		WithGaugeBarWidth(8),
		WithGaugeNoColor(true),
	)
	got := g.ViewString()
	if !strings.Contains(got, "5 / 8") {
		t.Errorf("ViewString() missing ratio: %q", got)
	}
}

func TestSandboxCapacityGauge_AccessibleLabel(t *testing.T) {
	t.Parallel()
	g := NewSandboxCapacityGauge(
		WithGaugeCurrent(2),
		WithGaugeMax(4),
	)
	label := g.AccessibleLabel()
	if !strings.Contains(label, "2") || !strings.Contains(label, "4") {
		t.Errorf("AccessibleLabel() = %q, want 2 and 4", label)
	}
}

func TestBuildCapacityBar(t *testing.T) {
	t.Parallel()
	tests := []struct {
		frac  float64
		width int
		want  string
	}{
		{0.0, 4, "░░░░"},
		{1.0, 4, "████"},
		{0.5, 4, "██░░"},
		{0.25, 4, "█░░░"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := buildCapacityBar(tc.frac, tc.width)
			if got != tc.want {
				t.Errorf("buildCapacityBar(%v, %d) = %q, want %q", tc.frac, tc.width, got, tc.want)
			}
		})
	}
}

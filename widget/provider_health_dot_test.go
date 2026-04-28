package widget

import (
	"strings"
	"testing"
)

func TestProviderHealthDot_NoColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		health    ProviderHealth
		showLabel bool
		want      string
	}{
		{ProviderHealthReady, false, "●"},
		{ProviderHealthReady, true, "● ready"},
		{ProviderHealthDegraded, false, "◐"},
		{ProviderHealthDegraded, true, "◐ degraded"},
		{ProviderHealthUnhealthy, false, "✗"},
		{ProviderHealthUnhealthy, true, "✗ unhealthy"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(string(tc.health), func(t *testing.T) {
			t.Parallel()
			d := NewProviderHealthDot(
				WithProviderHealth(tc.health),
				WithProviderHealthShowLabel(tc.showLabel),
				WithProviderHealthNoColor(true),
			)
			got := d.ViewString()
			if got != tc.want {
				t.Errorf("ViewString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProviderHealthDot_AccessibleLabel(t *testing.T) {
	t.Parallel()
	d := NewProviderHealthDot(WithProviderHealth(ProviderHealthDegraded))
	label := d.AccessibleLabel()
	if !strings.Contains(label, "degraded") {
		t.Errorf("AccessibleLabel() = %q, want 'degraded'", label)
	}
}

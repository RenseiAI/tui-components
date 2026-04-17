package format

import (
	"math"
	"testing"
)

func TestDuration(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "0s"},
		{30, "30s"},
		{59, "59s"},
		{60, "1m"},
		{90, "1m 30s"},
		{300, "5m"},
		{2820, "47m"},
		{3600, "1h"},
		{3660, "1h 1m"},
		{4320, "1h 12m"},
		{7200, "2h"},
		{13500, "3h 45m"},
	}

	for _, tt := range tests {
		got := Duration(tt.seconds)
		if got != tt.want {
			t.Errorf("Duration(%d) = %q, want %q", tt.seconds, got, tt.want)
		}
	}
}

func TestCost(t *testing.T) {
	v := 3.42
	small := 0.005
	zero := 0.0

	tests := []struct {
		name string
		usd  *float64
		want string
	}{
		{"nil", nil, "--"},
		{"zero", &zero, "--"},
		{"normal", &v, "$3.42"},
		{"small", &small, "$0.0050"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cost(tt.usd)
			if got != tt.want {
				t.Errorf("Cost(%v) = %q, want %q", tt.usd, got, tt.want)
			}
		})
	}
}

// TestCostEdgeCases asserts the CURRENT behavior of Cost for edge-case inputs
// (REN-985 / TC-011.4). These tests document today's behavior so the follow-up
// roll-up sub-issue (TC-011.6) can revise them intentionally.
//
// TC-011.6 will revise the following cases:
//   - math.NaN()     → "--"  (currently "$NaN")
//   - math.Inf(+1)   → "--"  (currently "$+Inf")
//   - math.Inf(-1)   → "--"  (currently "$-Inf")
//   - negative (e.g. -3.42) → "--" (currently passthrough "$-3.42")
//   - 1_000_000.0    → possibly "$1.0M" (TBD in .6; currently "$1000000.00")
func TestCostEdgeCases(t *testing.T) {
	negZero := math.Copysign(0, -1)
	neg := -3.42
	nan := math.NaN()
	posInf := math.Inf(1)
	negInf := math.Inf(-1)
	million := 1_000_000.0
	tiny := 0.00001

	tests := []struct {
		name string
		usd  *float64
		want string
	}{
		{"nil", nil, "--"},
		{"positive zero", func() *float64 { z := 0.0; return &z }(), "--"},
		{"negative zero", &negZero, "--"},
		// Negative values are < 0.01, so they currently hit the four-decimal
		// branch and render with four decimal places.
		{"negative passthrough", &neg, "$-3.4200"},
		{"NaN passthrough (slated to become --)", &nan, "$NaN"},
		{"positive infinity passthrough (slated to become --)", &posInf, "$+Inf"},
		{"negative infinity passthrough (slated to become --)", &negInf, "$-Inf"},
		{"million (no abbreviation today)", &million, "$1000000.00"},
		{"below four-decimal threshold", &tiny, "$0.0000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Cost(tt.usd)
			if got != tt.want {
				t.Errorf("Cost(%v) = %q, want %q", tt.usd, got, tt.want)
			}
		})
	}
}

func TestProviderName(t *testing.T) {
	p := "anthropic"
	if got := ProviderName(&p); got != "anthropic" {
		t.Errorf("ProviderName(&%q) = %q, want %q", p, got, "anthropic")
	}
	if got := ProviderName(nil); got != "--" {
		t.Errorf("ProviderName(nil) = %q, want %q", got, "--")
	}
}

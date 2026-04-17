package format

import (
	"math"
	"testing"
	"time"
)

// fixedNow swaps the package-level now function for the duration of the test,
// restoring the original on cleanup. This keeps RelativeTime assertions
// deterministic without introducing a clock interface.
func fixedNow(t *testing.T, fixed time.Time) {
	t.Helper()
	orig := now
	now = func() time.Time { return fixed }
	t.Cleanup(func() { now = orig })
}

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

// TestDurationEdgeCases locks in the CURRENT behavior of Duration for
// negative inputs and day-scale values. These expectations intentionally
// mirror what the implementation produces today so the suite stays green.
//
// NOTE: sibling roll-up TC-011.6 (REN-126) is expected to change the
// implementation to:
//   - clamp negative inputs to "0s"
//   - render day-scale values with a day unit, e.g. "%dd %dh"
//
// When that roll-up lands it will need to update the "want" column for every
// case tagged below with a "TC-011.6" comment.
func TestDurationEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		// Negatives — all land in the `seconds < 60` branch and pass through
		// fmt.Sprintf("%ds", ...) unchanged today.
		// TC-011.6 will revise each of these to "0s".
		{"negative_one_second", -1, "-1s"},              // TC-011.6: "0s"
		{"negative_just_under_minute", -59, "-59s"},     // TC-011.6: "0s"
		{"negative_one_minute_boundary", -60, "-60s"},   // TC-011.6: "0s"
		{"negative_one_hour_boundary", -3600, "-3600s"}, // TC-011.6: "0s"

		// Day-scale — expressed as hours today because no day unit exists.
		// TC-011.6 will revise each of these to a "%dd %dh" (or "%dd") form.
		{"one_day", 86400, "24h"},                                   // TC-011.6: "1d"
		{"two_days", 172800, "48h"},                                 // TC-011.6: "2d"
		{"one_week", 604800, "168h"},                                // TC-011.6: "7d"
		{"week_plus_hour_minute_second", 7*86400 + 3723, "169h 2m"}, // TC-011.6: "7d 1h 2m"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Duration(tt.seconds)
			if got != tt.want {
				t.Errorf("Duration(%d) = %q, want %q", tt.seconds, got, tt.want)
			}
		})
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

func TestRelativeTime(t *testing.T) {
	// Fixed reference point for all relative assertions.
	ref := time.Date(2026, 4, 17, 12, 0, 0, 0, time.UTC)
	fixedNow(t, ref)

	// Helper to produce an RFC 3339 timestamp relative to ref.
	iso := func(d time.Duration) string {
		return ref.Add(-d).Format(time.RFC3339)
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Passthrough on parse failure.
		{"empty", "", ""},
		{"not-a-date", "not-a-date", "not-a-date"},
		{"partial-year-month", "2024-01", "2024-01"},
		{"missing-timezone", "2024-01-01T00:00:00", "2024-01-01T00:00:00"},

		// Future timestamp (diff < 0) collapses to "just now".
		{"future-5s", ref.Add(5 * time.Second).Format(time.RFC3339), "just now"},

		// Sub-minute boundary still renders as "just now".
		{"59s-ago", iso(59 * time.Second), "just now"},

		// Minute / hour / day boundaries.
		{"exactly-60s-ago", iso(60 * time.Second), "1m ago"},
		{"exactly-60m-ago", iso(60 * time.Minute), "1h ago"},
		{"exactly-24h-ago", iso(24 * time.Hour), "1d ago"},

		// Large day counts.
		{"365d-ago", iso(365 * 24 * time.Hour), "365d ago"},
		{"1000d-ago", iso(1000 * 24 * time.Hour), "1000d ago"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativeTime(tt.input)
			if got != tt.want {
				t.Errorf("RelativeTime(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestTimestamp(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", ""},
		{"not-a-date", "not-a-date", "not-a-date"},
		{"partial-year-month", "2024-01", "2024-01"},
		{"missing-timezone", "2024-01-01T00:00:00", "2024-01-01T00:00:00"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Timestamp(tt.input)
			if got != tt.want {
				t.Errorf("Timestamp(%q) = %q, want %q", tt.input, got, tt.want)
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

func TestTokens(t *testing.T) {
	ptr := func(n int) *int { return &n }

	tests := []struct {
		name  string
		count *int
		want  string
	}{
		{"nil", nil, "--"},
		{"zero", ptr(0), "0"},
		{"sub-thousand", ptr(999), "999"},
		{"exactly 1k", ptr(1000), "1.0k"},
		{"1.5k", ptr(1500), "1.5k"},
		{"exactly 1M", ptr(1_000_000), "1.0M"},
		{"1.5M", ptr(1_500_000), "1.5M"},
		{"25.5M", ptr(25_500_000), "25.5M"},
		{"negative small", ptr(-1), "--"},
		{"negative large", ptr(-100), "--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Tokens(tt.count)
			if got != tt.want {
				t.Errorf("Tokens(%v) = %q, want %q", tt.count, got, tt.want)
			}
		})
	}
}

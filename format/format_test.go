package format

import (
	"math"
	"strings"
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

// TestDurationEdgeCases locks in the ratified behavior of Duration for
// negative inputs and day-scale values:
//   - negatives clamp to "0s"
//   - day-scale (>= 86400s) renders as "%dd %dh", dropping minutes
func TestDurationEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    string
	}{
		// Negatives clamp to "0s".
		{"negative_one_second", -1, "0s"},
		{"negative_just_under_minute", -59, "0s"},
		{"negative_one_minute_boundary", -60, "0s"},
		{"negative_one_hour_boundary", -3600, "0s"},

		// Day-scale — "%dd %dh", minutes dropped.
		{"one_day", 86400, "1d 0h"},
		{"two_days", 172800, "2d 0h"},
		{"one_week", 604800, "7d 0h"},
		{"week_plus_hour_minute_second", 7*86400 + 3723, "7d 1h"},
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

// TestCostEdgeCases asserts the ratified behavior of Cost for edge-case
// inputs (REN-985 / TC-011.6):
//   - negative values, NaN, and ±Inf all render as "--"
//   - the million-dollar abbreviation was deferred; 1_000_000.0 still
//     renders with two decimal places as "$1000000.00"
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
		{"negative", &neg, "--"},
		{"NaN", &nan, "--"},
		{"positive infinity", &posInf, "--"},
		{"negative infinity", &negInf, "--"},
		{"million (no abbreviation)", &million, "$1000000.00"},
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
	// Ratified behavior (TC-011.6): nil, empty-string, and whitespace-only
	// pointer values all render as "--". Non-blank strings pass through
	// verbatim regardless of length.
	empty := ""
	anthropic := "anthropic"
	space := " "
	long := strings.Repeat("a", 256)

	tests := []struct {
		name     string
		provider *string
		want     string
	}{
		{"nil", nil, "--"},
		{"empty string pointer", &empty, "--"},
		{"anthropic", &anthropic, "anthropic"},
		{"single space", &space, "--"},
		{"long name 256 chars", &long, long},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ProviderName(tt.provider)
			if got != tt.want {
				t.Errorf("ProviderName(%v) = %q, want %q", tt.provider, got, tt.want)
			}
		})
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

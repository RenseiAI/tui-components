package format

import (
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

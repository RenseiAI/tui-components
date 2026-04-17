package format

import "testing"

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

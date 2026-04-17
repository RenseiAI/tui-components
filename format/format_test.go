package format

import (
	"strings"
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

func TestProviderName(t *testing.T) {
	// Cases marked "TC-011.6 will revise" document current passthrough
	// behavior that the follow-up sub-issue will change so that &"" and
	// whitespace-only pointers return "--" to match nil semantics.
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
		{"empty string pointer (TC-011.6 will revise to \"--\")", &empty, ""},
		{"anthropic", &anthropic, "anthropic"},
		{"single space (TC-011.6 will revise to \"--\")", &space, " "},
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

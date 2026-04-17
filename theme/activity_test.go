package theme

import (
	"image/color"
	"testing"
)

func TestActivityMapsAligned(t *testing.T) {
	if len(ActivityColors) != len(ActivityIcons) {
		t.Fatalf("len(ActivityColors)=%d != len(ActivityIcons)=%d", len(ActivityColors), len(ActivityIcons))
	}

	for key := range ActivityColors {
		if _, ok := ActivityIcons[key]; !ok {
			t.Errorf("key %q in ActivityColors is missing from ActivityIcons", key)
		}
	}
	for key := range ActivityIcons {
		if _, ok := ActivityColors[key]; !ok {
			t.Errorf("key %q in ActivityIcons is missing from ActivityColors", key)
		}
	}
}

func TestActivityRequiredKeys(t *testing.T) {
	required := []string{"thought", "action", "response", "error", "progress"}
	for _, key := range required {
		t.Run(key, func(t *testing.T) {
			if _, ok := ActivityColors[key]; !ok {
				t.Errorf("ActivityColors missing required key %q", key)
			}
			if _, ok := ActivityIcons[key]; !ok {
				t.Errorf("ActivityIcons missing required key %q", key)
			}
		})
	}
}

func TestActivitySpotChecks(t *testing.T) {
	tests := []struct {
		key      string
		wantIcon string
	}{
		{"thought", "\U0001f4ad"},
		{"action", "\u26a1"},
		{"response", "\U0001f4ac"},
		{"error", "\u2717"},
		{"progress", "\u2713"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			if got := ActivityIcons[tt.key]; got != tt.wantIcon {
				t.Errorf("ActivityIcons[%q] = %q, want %q", tt.key, got, tt.wantIcon)
			}
		})
	}

	colorChecks := []struct {
		key  string
		want color.Color
	}{
		{"thought", TextSecondary},
		{"action", Teal},
		{"response", TextPrimary},
		{"error", StatusError},
		{"progress", StatusSuccess},
	}
	for _, cc := range colorChecks {
		t.Run(cc.key+"-color", func(t *testing.T) {
			got := ActivityColors[cc.key]
			if !sameColor(got, cc.want) {
				t.Errorf("ActivityColors[%q] = %v, want %v", cc.key, got, cc.want)
			}
		})
	}
}

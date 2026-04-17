package theme

import (
	"testing"
)

func TestGetWorkTypeColor(t *testing.T) {
	for key, want := range workTypeColors {
		t.Run(key, func(t *testing.T) {
			got := GetWorkTypeColor(key)
			if !sameColor(got, want) {
				t.Errorf("GetWorkTypeColor(%q) = %v, want %v", key, got, want)
			}
		})
	}

	t.Run("case-insensitive", func(t *testing.T) {
		lower := GetWorkTypeColor("development")
		mixed := GetWorkTypeColor("Development")
		upper := GetWorkTypeColor("DEVELOPMENT")
		if !sameColor(lower, mixed) || !sameColor(lower, upper) {
			t.Errorf("case-insensitivity broken: lower=%v mixed=%v upper=%v", lower, mixed, upper)
		}
	})

	t.Run("unknown", func(t *testing.T) {
		got := GetWorkTypeColor("not-a-real-worktype")
		if !sameColor(got, TextSecondary) {
			t.Errorf("GetWorkTypeColor(unknown) = %v, want TextSecondary", got)
		}
	})
}

func TestGetWorkTypeLabel(t *testing.T) {
	for key, want := range workTypeLabels {
		t.Run(key, func(t *testing.T) {
			got := GetWorkTypeLabel(key)
			if got != want {
				t.Errorf("GetWorkTypeLabel(%q) = %q, want %q", key, got, want)
			}
		})
	}

	t.Run("case-insensitive", func(t *testing.T) {
		lower := GetWorkTypeLabel("development")
		mixed := GetWorkTypeLabel("Development")
		upper := GetWorkTypeLabel("DEVELOPMENT")
		if lower != mixed || lower != upper {
			t.Errorf("case-insensitivity broken: lower=%q mixed=%q upper=%q", lower, mixed, upper)
		}
	})

	t.Run("unknown-passthrough", func(t *testing.T) {
		input := "not-a-real-worktype"
		got := GetWorkTypeLabel(input)
		if got != input {
			t.Errorf("GetWorkTypeLabel(%q) = %q, want %q (passthrough)", input, got, input)
		}
	})
}

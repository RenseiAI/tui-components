package theme

import (
	"testing"
)

// builtinWorkTypes lists all work-type kinds that ship in core.
// When new kinds are added to registerBuiltinWorkTypes they must also
// appear here so the test coverage stays current.
var builtinWorkTypes = []struct {
	key   string
	label string
}{
	{"development", "Development"},
	{"bugfix", "Bug Fix"},
	{"feature", "Feature"},
	{"qa", "QA"},
	{"qa-coordination", "QA Coord"},
	{"acceptance", "Acceptance"},
	{"acceptance-coordination", "Accept Coord"},
	{"coordination", "Coordination"},
	{"research", "Research"},
	{"backlog-creation", "Backlog"},
	{"inflight", "Inflight"},
	{"refinement", "Refinement"},
	{"refinement-coordination", "Refine Coord"},
	{"refactor", "Refactor"},
	{"review", "Review"},
	{"docs", "Docs"},
}

func TestGetWorkTypeColor(t *testing.T) {
	for _, wt := range builtinWorkTypes {
		t.Run(wt.key, func(t *testing.T) {
			e, ok := GlobalRegistry.GetWorkType(wt.key)
			if !ok {
				t.Fatalf("GlobalRegistry.GetWorkType(%q): not registered", wt.key)
			}
			got := GetWorkTypeColor(wt.key)
			if !sameColor(got, e.Color) {
				t.Errorf("GetWorkTypeColor(%q) = %v, want %v", wt.key, got, e.Color)
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
		if !sameColor(got, pkg.TextSecondary) {
			t.Errorf("GetWorkTypeColor(unknown) = %v, want pkg.TextSecondary", got)
		}
	})
}

func TestGetWorkTypeLabel(t *testing.T) {
	for _, wt := range builtinWorkTypes {
		t.Run(wt.key, func(t *testing.T) {
			got := GetWorkTypeLabel(wt.key)
			if got != wt.label {
				t.Errorf("GetWorkTypeLabel(%q) = %q, want %q", wt.key, got, wt.label)
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

// TestNoClosedSwitches asserts that all 16 built-in work types are driven
// by the open registry — no closed switch — and that the registry-driven
// label and color lookups return the expected values.
func TestNoClosedSwitches(t *testing.T) {
	wts := GlobalRegistry.ListWorkTypes()
	if len(wts) < 16 {
		t.Errorf("expected at least 16 built-in work types in registry, got %d", len(wts))
	}

	for _, wt := range builtinWorkTypes {
		t.Run("registry-driven/"+wt.key, func(t *testing.T) {
			e, ok := GlobalRegistry.GetWorkType(wt.key)
			if !ok {
				t.Errorf("kind %q missing from GlobalRegistry", wt.key)
				return
			}
			if e.Label != wt.label {
				t.Errorf("registry label for %q = %q, want %q", wt.key, e.Label, wt.label)
			}
			if e.Color == nil {
				t.Errorf("registry color for %q is nil", wt.key)
			}
		})
	}
}

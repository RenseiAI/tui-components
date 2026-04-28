package theme

import (
	"sync"
	"testing"

	"charm.land/lipgloss/v2"
)

// TestRegistryBuiltinStatuses verifies that all six built-in status kinds
// are pre-registered in GlobalRegistry.
func TestRegistryBuiltinStatuses(t *testing.T) {
	kinds := []struct {
		kind   string
		label  string
		symbol string
	}{
		{"working", "Working", "●"},
		{"queued", "Queued", "◌"},
		{"parked", "Parked", "○"},
		{"completed", "Done", "✓"},
		{"failed", "Failed", "✗"},
		{"stopped", "Stopped", "■"},
	}
	for _, k := range kinds {
		t.Run(k.kind, func(t *testing.T) {
			e, ok := GlobalRegistry.GetStatus(k.kind)
			if !ok {
				t.Fatalf("GetStatus(%q): not registered", k.kind)
			}
			if e.Label != k.label {
				t.Errorf("Label = %q, want %q", e.Label, k.label)
			}
			if e.Symbol != k.symbol {
				t.Errorf("Symbol = %q, want %q", e.Symbol, k.symbol)
			}
			if e.Color == nil {
				t.Errorf("Color is nil")
			}
		})
	}
}

// TestRegistryBuiltinActivities verifies that all five built-in activity
// kinds are pre-registered in GlobalRegistry.
func TestRegistryBuiltinActivities(t *testing.T) {
	kinds := []struct {
		kind string
		icon string
	}{
		{"thought", "\U0001f4ad"},
		{"action", "⚡"},
		{"response", "\U0001f4ac"},
		{"error", "✗"},
		{"progress", "✓"},
	}
	for _, k := range kinds {
		t.Run(k.kind, func(t *testing.T) {
			e, ok := GlobalRegistry.GetActivity(k.kind)
			if !ok {
				t.Fatalf("GetActivity(%q): not registered", k.kind)
			}
			if e.Icon != k.icon {
				t.Errorf("Icon = %q, want %q", e.Icon, k.icon)
			}
			if e.Color == nil {
				t.Errorf("Color is nil")
			}
		})
	}
}

// TestRegistryRegisterStatus verifies that custom status entries can be
// registered and retrieved, and that re-registration overwrites the prior
// entry.
func TestRegistryRegisterStatus(t *testing.T) {
	r := newRegistry()

	e := StatusEntry{
		Kind:    "workarea-warming",
		Label:   "Warming pool",
		Symbol:  "↻",
		Color:   lipgloss.Color("#4B8BF5"),
		Animate: true,
	}
	r.RegisterStatus(e)

	got, ok := r.GetStatus("workarea-warming")
	if !ok {
		t.Fatal("GetStatus: not found after Register")
	}
	if got.Label != e.Label {
		t.Errorf("Label = %q, want %q", got.Label, e.Label)
	}
	if got.Symbol != e.Symbol {
		t.Errorf("Symbol = %q, want %q", got.Symbol, e.Symbol)
	}
	if !got.Animate {
		t.Error("Animate = false, want true")
	}

	// Overwrite.
	e2 := StatusEntry{Kind: "workarea-warming", Label: "Pool warm", Symbol: "✔", Color: lipgloss.Color("#22C55E")}
	r.RegisterStatus(e2)
	got2, _ := r.GetStatus("workarea-warming")
	if got2.Label != "Pool warm" {
		t.Errorf("overwrite: Label = %q, want %q", got2.Label, "Pool warm")
	}
}

// TestRegistryGetStatusCaseInsensitive verifies case-insensitive key lookup.
func TestRegistryGetStatusCaseInsensitive(t *testing.T) {
	r := newRegistry()
	r.RegisterStatus(StatusEntry{Kind: "Working", Label: "Working", Symbol: "●"})

	for _, key := range []string{"Working", "working", "WORKING"} {
		_, ok := r.GetStatus(key)
		if !ok {
			t.Errorf("GetStatus(%q): not found", key)
		}
	}
}

// TestRegistryGetStatusUnknown verifies the false-ok return for missing keys.
func TestRegistryGetStatusUnknown(t *testing.T) {
	r := newRegistry()
	_, ok := r.GetStatus("not-registered")
	if ok {
		t.Error("expected ok=false for unregistered kind")
	}
}

// TestRegistryListStatuses verifies that ListStatuses returns all registered
// entries.
func TestRegistryListStatuses(t *testing.T) {
	r := newRegistry()
	r.RegisterStatus(StatusEntry{Kind: "a", Label: "A", Symbol: "a"})
	r.RegisterStatus(StatusEntry{Kind: "b", Label: "B", Symbol: "b"})

	list := r.ListStatuses()
	if len(list) != 2 {
		t.Errorf("ListStatuses: len = %d, want 2", len(list))
	}
}

// TestRegistryRegisterWorkType verifies custom work-type registration.
func TestRegistryRegisterWorkType(t *testing.T) {
	r := newRegistry()
	e := WorkTypeEntry{Kind: "workarea-acquire", Label: "Acquire", Color: lipgloss.Color("#00D4AA")}
	r.RegisterWorkType(e)

	got, ok := r.GetWorkType("workarea-acquire")
	if !ok {
		t.Fatal("GetWorkType: not found after Register")
	}
	if got.Label != e.Label {
		t.Errorf("Label = %q, want %q", got.Label, e.Label)
	}
}

// TestRegistryRegisterActivity verifies custom activity registration.
func TestRegistryRegisterActivity(t *testing.T) {
	r := newRegistry()
	e := ActivityEntry{Kind: "tool-call", Icon: "⚙", Color: lipgloss.Color("#4B8BF5")}
	r.RegisterActivity(e)

	got, ok := r.GetActivity("tool-call")
	if !ok {
		t.Fatal("GetActivity: not found after Register")
	}
	if got.Icon != e.Icon {
		t.Errorf("Icon = %q, want %q", got.Icon, e.Icon)
	}
}

// TestRegistryGetActivityUnknown verifies the false-ok return for missing
// activity kinds.
func TestRegistryGetActivityUnknown(t *testing.T) {
	r := newRegistry()
	_, ok := r.GetActivity("not-registered")
	if ok {
		t.Error("expected ok=false for unregistered kind")
	}
}

// TestRegistryListWorkTypes verifies list returns all registered entries.
func TestRegistryListWorkTypes(t *testing.T) {
	r := newRegistry()
	r.RegisterWorkType(WorkTypeEntry{Kind: "a", Label: "A", Color: lipgloss.Color("#FFFFFF")})
	r.RegisterWorkType(WorkTypeEntry{Kind: "b", Label: "B", Color: lipgloss.Color("#000000")})

	list := r.ListWorkTypes()
	if len(list) != 2 {
		t.Errorf("ListWorkTypes: len = %d, want 2", len(list))
	}
}

// TestRegistryListActivities verifies list returns all registered entries.
func TestRegistryListActivities(t *testing.T) {
	r := newRegistry()
	r.RegisterActivity(ActivityEntry{Kind: "a", Icon: "A", Color: lipgloss.Color("#FFFFFF")})
	r.RegisterActivity(ActivityEntry{Kind: "b", Icon: "B", Color: lipgloss.Color("#000000")})

	list := r.ListActivities()
	if len(list) != 2 {
		t.Errorf("ListActivities: len = %d, want 2", len(list))
	}
}

// TestRegistryConcurrency verifies that concurrent Register + Get calls on
// the same registry do not race.
func TestRegistryConcurrency(t *testing.T) {
	r := newRegistry()
	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(2)
		kind := "kind-" + string(rune('a'+i%26))
		go func(k string) {
			defer wg.Done()
			r.RegisterStatus(StatusEntry{Kind: k, Label: k, Symbol: "?"})
		}(kind)
		go func(k string) {
			defer wg.Done()
			_, _ = r.GetStatus(k)
		}(kind)
	}
	wg.Wait()
}

// TestGetStatusStyleUnknownFallback verifies the "?" fallback for unknown
// status kinds.
func TestGetStatusStyleUnknownFallback(t *testing.T) {
	st := GetStatusStyle("completely-unknown-status-xyz")
	if st.Symbol != "?" {
		t.Errorf("fallback symbol = %q, want %q", st.Symbol, "?")
	}
	if st.Label != "Unknown" {
		t.Errorf("fallback label = %q, want %q", st.Label, "Unknown")
	}
	if st.Animate {
		t.Error("fallback animate should be false")
	}
}

// TestGetActivityIconFallback verifies the "?" fallback for unknown activity
// kinds.
func TestGetActivityIconFallback(t *testing.T) {
	got := GetActivityIcon("completely-unknown-activity-xyz")
	if got != "?" {
		t.Errorf("fallback icon = %q, want %q", got, "?")
	}
}

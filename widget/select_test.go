package widget

import (
	"io"
	"testing"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/bradleyjkemp/cupaloy/v2"
)

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// testItem implements SelectItem for testing purposes.
type testItem struct {
	id, title, desc, filterVal string
}

func (t testItem) ID() string          { return t.id }
func (t testItem) Title() string       { return t.title }
func (t testItem) Description() string { return t.desc }
func (t testItem) FilterValue() string { return t.filterVal }

// sampleItems returns a small deterministic slice of test items.
func sampleItems() []SelectItem {
	return []SelectItem{
		testItem{id: "1", title: "Alpha", desc: "First item", filterVal: "alpha"},
		testItem{id: "2", title: "Beta", desc: "Second item", filterVal: "beta"},
		testItem{id: "3", title: "Gamma", desc: "Third item", filterVal: "gamma"},
	}
}

// spaceKey returns a tea.KeyPressMsg for the space bar.
func spaceKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeySpace}
}

// downKey returns a tea.KeyPressMsg for the down arrow.
func downKey() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyDown}
}

// stubDelegate is a minimal ItemDelegate used to verify custom delegate
// override semantics.
type stubDelegate struct{}

func (stubDelegate) Render(io.Writer, list.Model, int, list.Item) {}
func (stubDelegate) Height() int                                  { return 1 }
func (stubDelegate) Spacing() int                                 { return 0 }
func (stubDelegate) Update(tea.Msg, *list.Model) tea.Cmd          { return nil }

// snapshotter returns a cupaloy instance writing snapshots to widget/.snapshots/.
var snapshotter = cupaloy.New(cupaloy.SnapshotSubdirectory(".snapshots"))

// ---------------------------------------------------------------------------
// Behavioral tests (table-driven, no snapshots)
// ---------------------------------------------------------------------------

func TestNewSelect(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())
	if s == nil {
		t.Fatal("NewSelect returned nil")
	}
	// Should have a non-empty view after sizing.
	s.SetSize(40, 10)
	s.Focus()
	v := s.View()
	if v.Content == "" {
		t.Error("expected non-empty View after SetSize and Focus")
	}
}

func TestSingleSelectSelected(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())
	s.SetSize(40, 10)
	s.Focus()

	sel := s.Selected()
	if len(sel) != 1 {
		t.Fatalf("expected 1 selected item, got %d", len(sel))
	}
	if sel[0].ID() != "1" {
		t.Errorf("expected selected ID %q, got %q", "1", sel[0].ID())
	}
}

func TestSingleSelectHighlighted(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())
	s.SetSize(40, 10)
	s.Focus()

	h := s.Highlighted()
	if h == nil {
		t.Fatal("Highlighted() returned nil")
	}
	if h.ID() != "1" {
		t.Errorf("expected highlighted ID %q, got %q", "1", h.ID())
	}
}

func TestEmptyListSelectedAndHighlighted(t *testing.T) {
	t.Parallel()

	s := NewSelect(nil)
	s.SetSize(40, 10)
	s.Focus()

	if sel := s.Selected(); sel != nil {
		t.Errorf("expected nil Selected on empty list, got %v", sel)
	}
	if h := s.Highlighted(); h != nil {
		t.Errorf("expected nil Highlighted on empty list, got %v", h)
	}
}

func TestMultiSelectToggle(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	// Initially nothing is toggled.
	if sel := s.Selected(); sel != nil {
		t.Fatalf("expected nil Selected before any toggle, got %d items", len(sel))
	}

	// Toggle first item (cursor is on item 1).
	s.Update(spaceKey())

	sel := s.Selected()
	if len(sel) != 1 {
		t.Fatalf("expected 1 selected after toggle, got %d", len(sel))
	}
	if sel[0].ID() != "1" {
		t.Errorf("expected ID %q, got %q", "1", sel[0].ID())
	}

	// Toggle again to deselect.
	s.Update(spaceKey())

	if sel := s.Selected(); sel != nil {
		t.Errorf("expected nil Selected after un-toggle, got %d items", len(sel))
	}
}

func TestMultiSelectInsertionOrder(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	// Toggle item 1.
	s.Update(spaceKey())

	// Move down, toggle item 2.
	s.Update(downKey())
	s.Update(spaceKey())

	// Move down, toggle item 3.
	s.Update(downKey())
	s.Update(spaceKey())

	sel := s.Selected()
	if len(sel) != 3 {
		t.Fatalf("expected 3 selected, got %d", len(sel))
	}

	expectedOrder := []string{"1", "2", "3"}
	for i, want := range expectedOrder {
		if sel[i].ID() != want {
			t.Errorf("Selected()[%d]: expected ID %q, got %q", i, want, sel[i].ID())
		}
	}
}

func TestSetMultiSelectFalseClearsSelection(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	// Toggle two items.
	s.Update(spaceKey())
	s.Update(downKey())
	s.Update(spaceKey())

	if len(s.Selected()) != 2 {
		t.Fatal("expected 2 selected items before clear")
	}

	s.SetMultiSelect(false)

	if s.IsSelected("1") || s.IsSelected("2") {
		t.Error("expected IsSelected to be false after SetMultiSelect(false)")
	}
}

func TestIsSelected(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	if s.IsSelected("1") {
		t.Error("expected IsSelected(\"1\") false before toggle")
	}

	s.Update(spaceKey())

	if !s.IsSelected("1") {
		t.Error("expected IsSelected(\"1\") true after toggle")
	}
	if s.IsSelected("2") {
		t.Error("expected IsSelected(\"2\") false")
	}
}

func TestSelectFocusBlur(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)

	// Blur: Update should be a no-op.
	s.Blur()
	s.Update(spaceKey())

	if s.IsSelected("1") {
		t.Error("expected no selection toggle when blurred")
	}

	// Focus: Update should work.
	s.Focus()
	s.Update(spaceKey())

	if !s.IsSelected("1") {
		t.Error("expected selection toggle when focused")
	}
}

func TestSetSizeDoesNotPanic(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())

	// Various sizes should not panic.
	s.SetSize(0, 0)
	s.SetSize(40, 10)
	s.SetSize(200, 50)

	s.Focus()
	v := s.View()
	if v.Content == "" {
		t.Error("expected non-empty View after SetSize")
	}
}

func TestWithFilterDisabled(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithFilter(false))
	if s.filterEnabled {
		t.Error("expected filterEnabled to be false")
	}
}

func TestWithMultiSelectOption(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	if !s.multiSelect {
		t.Error("expected multiSelect to be true")
	}
	if s.selected == nil {
		t.Error("expected selected map to be initialized")
	}
}

func TestWithDelegate(t *testing.T) {
	t.Parallel()

	d := stubDelegate{}
	s := NewSelect(sampleItems(), WithDelegate(d))
	if s.customDelegate == nil {
		t.Error("expected customDelegate to be set")
	}
}

// ---------------------------------------------------------------------------
// Golden snapshot tests
// ---------------------------------------------------------------------------

func TestSnapshot_SingleSelectFocused(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())
	s.SetSize(40, 10)
	s.Focus()

	snapshotter.SnapshotT(t, s.View().Content)
}

func TestSnapshot_SingleSelectBlurred(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems())
	s.SetSize(40, 10)
	s.Blur()

	snapshotter.SnapshotT(t, s.View().Content)
}

func TestSnapshot_SingleSelectEmpty(t *testing.T) {
	t.Parallel()

	s := NewSelect(nil)
	s.SetSize(40, 10)
	s.Focus()

	snapshotter.SnapshotT(t, s.View().Content)
}

func TestSnapshot_MultiSelectSomeSelected(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	// Toggle first item only.
	s.Update(spaceKey())
	// Move to second (leave it unselected).
	s.Update(downKey())

	snapshotter.SnapshotT(t, s.View().Content)
}

func TestSnapshot_MultiSelectAllSelected(t *testing.T) {
	t.Parallel()

	s := NewSelect(sampleItems(), WithMultiSelect(true))
	s.SetSize(40, 10)
	s.Focus()

	// Toggle all three items.
	s.Update(spaceKey())
	s.Update(downKey())
	s.Update(spaceKey())
	s.Update(downKey())
	s.Update(spaceKey())

	snapshotter.SnapshotT(t, s.View().Content)
}

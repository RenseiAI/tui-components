package widget

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bradleyjkemp/cupaloy/v2"
)

// tabKey builds a [tea.KeyPressMsg] whose String() representation
// matches the given keystroke. For single-character keystrokes (e.g.,
// "1", "a") it populates Text so the stringer shortcut returns the
// literal text. For named keys ("left", "right", "tab", "shift+tab")
// it sets the appropriate Code and modifier bits so Keystroke() returns
// the expected name.
func tabKey(t *testing.T, stroke string) tea.KeyPressMsg {
	t.Helper()
	switch stroke {
	case "left":
		return tea.KeyPressMsg{Code: tea.KeyLeft}
	case "right":
		return tea.KeyPressMsg{Code: tea.KeyRight}
	case "tab":
		return tea.KeyPressMsg{Code: tea.KeyTab}
	case "shift+tab":
		return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	default:
		// Treat as a literal printable character / string.
		// When Text is non-empty, Key.String() returns it directly.
		code := rune(0)
		for _, r := range stroke {
			code = r
			break
		}
		return tea.KeyPressMsg{Code: code, Text: stroke}
	}
}

// collectMsg runs cmd (if any) and returns the resulting message. nil
// commands produce a nil message — tests use that to assert no message
// was emitted.
func collectMsg(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// threeItems returns a small set of enabled tabs with distinct IDs and
// KeyHints for most tests.
func threeItems() []TabsItem {
	return []TabsItem{
		{ID: "a", Title: "Alpha", KeyHint: "1"},
		{ID: "b", Title: "Beta", KeyHint: "2"},
		{ID: "c", Title: "Gamma", KeyHint: "3"},
	}
}

// fiveItemsWithDisabled returns a list with two disabled tabs so
// disabled-skip behaviour is observable.
func fiveItemsWithDisabled() []TabsItem {
	return []TabsItem{
		{ID: "a", Title: "Alpha", KeyHint: "1"},
		{ID: "b", Title: "Beta", KeyHint: "2", Disabled: true},
		{ID: "c", Title: "Gamma", KeyHint: "3"},
		{ID: "d", Title: "Delta", KeyHint: "4", Disabled: true},
		{ID: "e", Title: "Epsilon", KeyHint: "5"},
	}
}

func TestTabsUpdate_SequentialNavigation(t *testing.T) {
	tests := []struct {
		name       string
		items      []TabsItem
		wraparound bool
		start      int
		keys       []string
		wantActive int
		// wantMsgs is the expected sequence of non-nil TabsSelectedMsg
		// indices produced by the key presses. Nil keys produce no
		// message and are filtered out.
		wantMsgs []int
	}{
		{
			name:       "right clamped at end",
			items:      threeItems(),
			start:      2,
			keys:       []string{"right"},
			wantActive: 2,
			wantMsgs:   nil,
		},
		{
			name:       "left clamped at start",
			items:      threeItems(),
			start:      0,
			keys:       []string{"left"},
			wantActive: 0,
			wantMsgs:   nil,
		},
		{
			name:       "right advances",
			items:      threeItems(),
			start:      0,
			keys:       []string{"right", "right"},
			wantActive: 2,
			wantMsgs:   []int{1, 2},
		},
		{
			name:       "left retreats",
			items:      threeItems(),
			start:      2,
			keys:       []string{"left", "left"},
			wantActive: 0,
			wantMsgs:   []int{1, 0},
		},
		{
			name:       "tab is right",
			items:      threeItems(),
			start:      0,
			keys:       []string{"tab"},
			wantActive: 1,
			wantMsgs:   []int{1},
		},
		{
			name:       "shift+tab is left",
			items:      threeItems(),
			start:      2,
			keys:       []string{"shift+tab"},
			wantActive: 1,
			wantMsgs:   []int{1},
		},
		{
			name:       "wraparound right wraps past end",
			items:      threeItems(),
			wraparound: true,
			start:      2,
			keys:       []string{"right"},
			wantActive: 0,
			wantMsgs:   []int{0},
		},
		{
			name:       "wraparound left wraps past start",
			items:      threeItems(),
			wraparound: true,
			start:      0,
			keys:       []string{"left"},
			wantActive: 2,
			wantMsgs:   []int{2},
		},
		{
			name:       "right skips disabled",
			items:      fiveItemsWithDisabled(),
			start:      0,
			keys:       []string{"right"},
			wantActive: 2,
			wantMsgs:   []int{2},
		},
		{
			name:       "left skips disabled",
			items:      fiveItemsWithDisabled(),
			start:      4,
			keys:       []string{"left"},
			wantActive: 2,
			wantMsgs:   []int{2},
		},
		{
			name:       "right clamped when all remaining disabled",
			items:      []TabsItem{{ID: "a", Title: "A"}, {ID: "b", Title: "B", Disabled: true}},
			start:      0,
			keys:       []string{"right"},
			wantActive: 0,
			wantMsgs:   nil,
		},
		{
			name:       "left clamped when all remaining disabled",
			items:      []TabsItem{{ID: "a", Title: "A", Disabled: true}, {ID: "b", Title: "B"}},
			start:      1,
			keys:       []string{"left"},
			wantActive: 1,
			wantMsgs:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []TabsOption{WithActive(tt.start)}
			if tt.wraparound {
				opts = append(opts, WithWraparound(true))
			}
			tabs := NewTabs(tt.items, opts...)
			tabs.Focus()

			var gotMsgs []int
			for _, k := range tt.keys {
				_, cmd := tabs.Update(tabKey(t, k))
				msg := collectMsg(cmd)
				if m, ok := msg.(TabsSelectedMsg); ok {
					gotMsgs = append(gotMsgs, m.Index)
					if m.ID != tt.items[m.Index].ID {
						t.Errorf("msg ID mismatch: got %q want %q", m.ID, tt.items[m.Index].ID)
					}
				}
			}

			if got := tabs.Active(); got != tt.wantActive {
				t.Errorf("Active() = %d, want %d", got, tt.wantActive)
			}
			if len(gotMsgs) != len(tt.wantMsgs) {
				t.Fatalf("emitted %d msgs (%v), want %d (%v)", len(gotMsgs), gotMsgs, len(tt.wantMsgs), tt.wantMsgs)
			}
			for i := range gotMsgs {
				if gotMsgs[i] != tt.wantMsgs[i] {
					t.Errorf("msg[%d] index = %d, want %d", i, gotMsgs[i], tt.wantMsgs[i])
				}
			}
		})
	}
}

func TestTabsUpdate_Shortcuts(t *testing.T) {
	tests := []struct {
		name       string
		items      []TabsItem
		start      int
		key        string
		wantActive int
		wantMsg    bool
	}{
		{
			name:       "digit activates tab",
			items:      threeItems(),
			start:      0,
			key:        "3",
			wantActive: 2,
			wantMsg:    true,
		},
		{
			name:       "digit for already-active is no-op",
			items:      threeItems(),
			start:      1,
			key:        "2",
			wantActive: 1,
			wantMsg:    false,
		},
		{
			name:       "digit for disabled tab is no-op",
			items:      fiveItemsWithDisabled(),
			start:      0,
			key:        "2",
			wantActive: 0,
			wantMsg:    false,
		},
		{
			name:       "unknown key is no-op",
			items:      threeItems(),
			start:      0,
			key:        "z",
			wantActive: 0,
			wantMsg:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabs := NewTabs(tt.items, WithActive(tt.start))
			tabs.Focus()

			_, cmd := tabs.Update(tabKey(t, tt.key))
			msg := collectMsg(cmd)
			gotMsg := false
			if m, ok := msg.(TabsSelectedMsg); ok {
				gotMsg = true
				if m.Index != tt.wantActive {
					t.Errorf("msg.Index = %d, want %d", m.Index, tt.wantActive)
				}
				if m.ID != tt.items[m.Index].ID {
					t.Errorf("msg.ID = %q, want %q", m.ID, tt.items[m.Index].ID)
				}
			}
			if gotMsg != tt.wantMsg {
				t.Errorf("msg emitted = %v, want %v", gotMsg, tt.wantMsg)
			}
			if got := tabs.Active(); got != tt.wantActive {
				t.Errorf("Active() = %d, want %d", got, tt.wantActive)
			}
		})
	}
}

func TestTabsUpdate_NoMsgOnNoop(t *testing.T) {
	// Pressing the already-active tab's shortcut must not emit a msg.
	t.Run("shortcut on already-active", func(t *testing.T) {
		tabs := NewTabs(threeItems(), WithActive(1))
		tabs.Focus()
		_, cmd := tabs.Update(tabKey(t, "2"))
		if cmd != nil {
			if _, ok := collectMsg(cmd).(TabsSelectedMsg); ok {
				t.Errorf("unexpected TabsSelectedMsg on already-active shortcut")
			}
		}
	})

	t.Run("left at index 0 clamped", func(t *testing.T) {
		tabs := NewTabs(threeItems(), WithActive(0))
		tabs.Focus()
		_, cmd := tabs.Update(tabKey(t, "left"))
		if cmd != nil {
			if _, ok := collectMsg(cmd).(TabsSelectedMsg); ok {
				t.Errorf("unexpected msg on clamped left at 0")
			}
		}
	})

	t.Run("right when all other disabled", func(t *testing.T) {
		items := []TabsItem{
			{ID: "a", Title: "A"},
			{ID: "b", Title: "B", Disabled: true},
			{ID: "c", Title: "C", Disabled: true},
		}
		tabs := NewTabs(items, WithActive(0))
		tabs.Focus()
		_, cmd := tabs.Update(tabKey(t, "right"))
		if cmd != nil {
			if _, ok := collectMsg(cmd).(TabsSelectedMsg); ok {
				t.Errorf("unexpected msg when all other tabs disabled")
			}
		}
	})
}

func TestTabsUpdate_FocusGate(t *testing.T) {
	tabs := NewTabs(threeItems(), WithActive(0))
	// Widget is blurred by default — verify input is ignored.
	_, cmd := tabs.Update(tabKey(t, "right"))
	if cmd != nil {
		t.Errorf("blurred widget produced a command")
	}
	if got := tabs.Active(); got != 0 {
		t.Errorf("blurred widget advanced active index to %d", got)
	}

	// After Focus, keys take effect.
	tabs.Focus()
	_, cmd = tabs.Update(tabKey(t, "right"))
	if cmd == nil {
		t.Errorf("focused widget did not emit a command on right")
	}
	if got := tabs.Active(); got != 1 {
		t.Errorf("focused Active() = %d, want 1", got)
	}

	// Blur again, and verify keys are again ignored.
	tabs.Blur()
	_, cmd = tabs.Update(tabKey(t, "right"))
	if cmd != nil {
		t.Errorf("re-blurred widget produced a command")
	}
	if got := tabs.Active(); got != 1 {
		t.Errorf("re-blurred Active() = %d, want 1", got)
	}
}

func TestTabsUpdate_IgnoresNonKeyMsg(t *testing.T) {
	tabs := NewTabs(threeItems(), WithActive(0))
	tabs.Focus()
	// WindowSizeMsg is a value type from bubbletea that is not a
	// KeyPressMsg; the widget should ignore it gracefully.
	_, cmd := tabs.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Errorf("unexpected command on non-key msg")
	}
	if got := tabs.Active(); got != 0 {
		t.Errorf("non-key msg altered Active() to %d", got)
	}
}

func TestTabsSetActive_Clamps(t *testing.T) {
	tests := []struct {
		name string
		in   int
		n    int
		want int
	}{
		{"negative", -5, 3, 0},
		{"above range", 99, 3, 2},
		{"in range", 1, 3, 1},
		{"empty items clamp to zero", 5, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var items []TabsItem
			if tt.n > 0 {
				items = threeItems()[:tt.n]
			}
			tabs := NewTabs(items)
			tabs.SetActive(tt.in)
			if got := tabs.Active(); got != tt.want {
				t.Errorf("Active() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestTabsWithActive_OutOfRangeClamped(t *testing.T) {
	// WithActive out-of-range at construction is also clamped.
	tabs := NewTabs(threeItems(), WithActive(99))
	if got := tabs.Active(); got != 2 {
		t.Errorf("Active() = %d after WithActive(99), want 2", got)
	}
	tabs = NewTabs(threeItems(), WithActive(-1))
	if got := tabs.Active(); got != 0 {
		t.Errorf("Active() = %d after WithActive(-1), want 0", got)
	}
}

func TestTabsSetItems_ClampsActive(t *testing.T) {
	tabs := NewTabs(threeItems(), WithActive(2))
	if got := tabs.Active(); got != 2 {
		t.Fatalf("initial Active() = %d, want 2", got)
	}

	// Shrink the list: active must clamp into the new range.
	tabs.SetItems([]TabsItem{{ID: "x", Title: "X"}})
	if got := tabs.Active(); got != 0 {
		t.Errorf("after SetItems(1 item), Active() = %d, want 0", got)
	}

	// Empty list: active falls back to 0.
	tabs.SetItems(nil)
	if got := tabs.Active(); got != 0 {
		t.Errorf("after SetItems(nil), Active() = %d, want 0", got)
	}

	// Growing the list preserves the currently-clamped active (0).
	tabs.SetItems(threeItems())
	if got := tabs.Active(); got != 0 {
		t.Errorf("after SetItems(3 items), Active() = %d, want 0", got)
	}
}

func TestTabsInit_NoCmd(t *testing.T) {
	// Init has no startup work; it must return nil.
	tabs := NewTabs(threeItems())
	if cmd := tabs.Init(); cmd != nil {
		t.Errorf("Init() = %v, want nil", cmd)
	}
}

func TestTabsSetSize_NegativeCoercedToZero(t *testing.T) {
	tabs := NewTabs(threeItems())
	tabs.SetSize(-5, -10)
	// Negative values coerce to zero. Width of zero behaves as
	// "unconstrained" — the bar renders at natural size with separators
	// intact, so the rendered view must be non-empty.
	view := tabs.View().Content
	if view == "" {
		t.Errorf("View() empty after SetSize(-5, -10); expected unconstrained render")
	}
}

func TestTabsKeyMap_Override(t *testing.T) {
	// Overriding the keymap disables the defaults.
	km := DefaultTabsKeyMap()
	tabs := NewTabs(threeItems(), WithActive(0), WithTabsKeyMap(km))
	tabs.Focus()
	_, cmd := tabs.Update(tabKey(t, "right"))
	if cmd == nil {
		t.Errorf("custom keymap did not honour right key")
	}
	if got := tabs.Active(); got != 1 {
		t.Errorf("Active() = %d, want 1", got)
	}
}

// --- Golden file tests for View() ---

// renderForGolden constructs a Tabs, applies width/height, and returns
// the rendered bar string exactly as View() would emit it.
func renderForGolden(items []TabsItem, active, width int) string {
	tabs := NewTabs(items, WithActive(active))
	tabs.SetSize(width, 1)
	return tabs.View().Content
}

func TestTabsView_Golden(t *testing.T) {
	// Use a deterministic set of tabs for most snapshots so horizontal
	// layout is stable across widths.
	items := []TabsItem{
		{ID: "a", Title: "Alpha", KeyHint: "1"},
		{ID: "b", Title: "Beta", KeyHint: "2"},
		{ID: "c", Title: "Gamma", KeyHint: "3"},
		{ID: "d", Title: "Delta", KeyHint: "4"},
	}

	withDisabled := []TabsItem{
		{ID: "a", Title: "Alpha", KeyHint: "1"},
		{ID: "b", Title: "Beta", KeyHint: "2", Disabled: true},
		{ID: "c", Title: "Gamma", KeyHint: "3"},
		{ID: "d", Title: "Delta", KeyHint: "4"},
	}

	// longTitles is used to guarantee the truncation path fires at
	// tight widths.
	longTitles := []TabsItem{
		{ID: "a", Title: "Dashboards"},
		{ID: "b", Title: "Integrations"},
		{ID: "c", Title: "Observability"},
		{ID: "d", Title: "Administration"},
	}

	cases := []struct {
		name   string
		items  []TabsItem
		active int
		width  int
	}{
		{"width40_active_first", items, 0, 40},
		{"width40_active_middle", items, 1, 40},
		{"width40_active_last", items, 3, 40},
		{"width80_active_first", items, 0, 80},
		{"width80_active_middle", items, 1, 80},
		{"width80_active_last", items, 3, 80},
		{"width120_active_first", items, 0, 120},
		{"width120_active_middle", items, 1, 120},
		{"width120_active_last", items, 3, 120},
		{"width80_with_disabled", withDisabled, 0, 80},
		{"width40_with_disabled", withDisabled, 2, 40},
		{"truncation_narrow", longTitles, 1, 30},
		{"truncation_very_narrow", longTitles, 2, 16},
		{"unconstrained", items, 1, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderForGolden(tc.items, tc.active, tc.width)
			cupaloy.SnapshotT(t, got)
		})
	}
}

func TestTabsView_EmptyItems(t *testing.T) {
	tabs := NewTabs(nil)
	tabs.SetSize(40, 1)
	cupaloy.SnapshotT(t, tabs.View().Content)
}

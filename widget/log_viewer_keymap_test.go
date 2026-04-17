package widget

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

// newFocusedForKeymap constructs a LogViewer that is focused, sized,
// and primed with enough lines to exercise scroll-away bindings.
func newFocusedForKeymap(t *testing.T, opts ...LogViewerOption) *LogViewer {
	t.Helper()
	m := NewLogViewer(opts...)
	m.Focus()
	m.SetSize(20, 6)
	for i := 0; i < 30; i++ {
		m.Append("line")
	}
	return m
}

func TestDefaultKeyMapBindings(t *testing.T) {
	km := DefaultKeyMap()
	tests := []struct {
		name    string
		binding key.Binding
		want    []string
	}{
		{"LineUp", km.LineUp, []string{"up", "k"}},
		{"LineDown", km.LineDown, []string{"down", "j"}},
		{"PageUp", km.PageUp, []string{"pgup"}},
		{"PageDown", km.PageDown, []string{"pgdown"}},
		{"Home", km.Home, []string{"home", "g"}},
		{"End", km.End, []string{"end", "G"}},
		{"ToggleFollow", km.ToggleFollow, []string{"f"}},
		{"ToggleWrap", km.ToggleWrap, []string{"w"}},
		{"Clear", km.Clear, []string{"c"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.binding.Keys()
			if len(got) != len(tc.want) {
				t.Fatalf("%s keys: want %v, got %v", tc.name, tc.want, got)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("%s keys[%d]: want %q, got %q", tc.name, i, tc.want[i], got[i])
				}
			}
			if tc.binding.Help().Key == "" {
				t.Errorf("%s: expected non-empty help label", tc.name)
			}
		})
	}
}

// TestKeyMapActionTable verifies that each default binding triggers
// the semantic action the widget promises.
func TestKeyMapActionTable(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		setup  func(*testing.T) *LogViewer
		verify func(*testing.T, *LogViewer)
	}{
		{
			name: "Clear empties buffer",
			key:  "c",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if len(m.lines) != 0 {
					t.Errorf("Clear: want 0 retained lines, got %d", len(m.lines))
				}
			},
		},
		{
			name: "ToggleFollow flips follow off",
			key:  "f",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("ToggleFollow: want follow off, still on")
				}
			},
		},
		{
			name: "ToggleWrap flips wrap off",
			key:  "w",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Wrap() {
					t.Error("ToggleWrap: want wrap off, still on")
				}
			},
		},
		{
			name: "End re-engages follow",
			key:  "end",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				m := newFocusedForKeymap(t)
				m.SetFollowing(false)
				m.viewport.GotoTop()
				return m
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("End: want follow on")
				}
				if !m.viewport.AtBottom() {
					t.Error("End: want viewport at tail")
				}
			},
		},
		{
			name: "Home pauses follow and goes to top",
			key:  "home",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("Home: want follow off")
				}
				if !m.viewport.AtTop() {
					t.Error("Home: want viewport at top")
				}
			},
		},
		{
			name: "PageUp pauses follow",
			key:  "pgup",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("PageUp: want follow off")
				}
				if m.viewport.AtBottom() {
					t.Error("PageUp: want viewport not at bottom")
				}
			},
		},
		{
			name: "PageDown at tail keeps follow unchanged",
			key:  "pgdown",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				// At tail before, at tail after: follow stays on.
				if !m.Following() {
					t.Error("PageDown at tail: want follow on")
				}
			},
		},
		{
			name: "LineUp pauses follow",
			key:  "up",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("LineUp: want follow off")
				}
				if m.viewport.AtBottom() {
					t.Error("LineUp: want viewport not at bottom")
				}
			},
		},
		{
			name: "LineUp via k also pauses follow",
			key:  "k",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				return newFocusedForKeymap(t)
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("k: want follow off")
				}
			},
		},
		{
			name: "LineDown when paused mid-buffer does not re-engage follow",
			key:  "down",
			setup: func(t *testing.T) *LogViewer {
				t.Helper()
				m := newFocusedForKeymap(t)
				m.SetFollowing(false)
				m.viewport.GotoTop()
				return m
			},
			verify: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("LineDown mid-buffer: follow should stay off")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t)
			_, _ = m.Update(logViewerKey(tc.key))
			tc.verify(t, m)
		})
	}
}

// TestWithKeyMapOverride swaps ToggleFollow to the uppercase 'F' key
// and verifies that (a) default 'f' is no longer bound and (b) 'F'
// now toggles follow.
func TestWithKeyMapOverride(t *testing.T) {
	custom := KeyMap{
		ToggleFollow: key.NewBinding(
			key.WithKeys("F"),
			key.WithHelp("F", "toggle follow"),
		),
	}
	m := NewLogViewer(WithLogViewerKeyMap(custom))
	m.Focus()
	m.SetSize(20, 6)
	for i := 0; i < 30; i++ {
		m.Append("line")
	}

	// Default 'f' should no longer toggle follow because every other
	// binding on the custom KeyMap is empty.
	before := m.Following()
	_, _ = m.Update(logViewerKey("f"))
	if m.Following() != before {
		t.Errorf("lowercase f should not toggle with override; before=%v after=%v", before, m.Following())
	}

	// Capital 'F' should.
	_, _ = m.Update(logViewerKey("F"))
	if m.Following() == before {
		t.Errorf("capital F should toggle follow; before=%v after=%v", before, m.Following())
	}
}

func TestFocusBlurDispatchMatrix(t *testing.T) {
	t.Run("focused key toggles wrap", func(t *testing.T) {
		m := NewLogViewer()
		m.Focus()
		m.SetSize(40, 10)
		before := m.Wrap()
		_, _ = m.Update(logViewerKey("w"))
		if m.Wrap() == before {
			t.Errorf("focused: wrap did not toggle (before=%v)", before)
		}
	})

	t.Run("blurred key dropped", func(t *testing.T) {
		m := NewLogViewer()
		// m.Focus() intentionally NOT called.
		m.SetSize(40, 10)
		before := m.Wrap()
		_, _ = m.Update(logViewerKey("w"))
		if m.Wrap() != before {
			t.Errorf("blurred: wrap toggled despite focus gate (before=%v after=%v)", before, m.Wrap())
		}
	})

	t.Run("blurred AppendMsg still grows buffer", func(t *testing.T) {
		m := NewLogViewer()
		m.SetSize(40, 10)
		_, _ = m.Update(AppendMsg{Lines: []string{"a", "b", "c"}})
		if len(m.lines) != 3 {
			t.Errorf("blurred AppendMsg: want 3 lines, got %d", len(m.lines))
		}
	})

	t.Run("blurred WindowSizeMsg still resizes", func(t *testing.T) {
		m := NewLogViewer()
		_, _ = m.Update(tea.WindowSizeMsg{Width: 77, Height: 11})
		if m.width != 77 || m.height != 11 {
			t.Errorf("blurred WindowSizeMsg: want 77x11, got %dx%d", m.width, m.height)
		}
	})

	t.Run("blurred MouseWheelMsg still scrolls viewport", func(t *testing.T) {
		m := NewLogViewer()
		m.SetSize(20, 6)
		for i := 0; i < 30; i++ {
			m.Append("line")
		}
		// Precondition: following at tail.
		if !m.Following() || !m.viewport.AtBottom() {
			t.Fatal("precondition: following at tail")
		}
		_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
		if m.Following() {
			t.Error("blurred MouseWheelMsg up: follow should pause")
		}
	})
}

func TestSetWrapTogglesField(t *testing.T) {
	m := NewLogViewer()
	if !m.Wrap() {
		t.Fatal("default wrap should be true")
	}
	m.SetWrap(false)
	if m.Wrap() {
		t.Error("SetWrap(false) did not flip field")
	}
	m.SetWrap(true)
	if !m.Wrap() {
		t.Error("SetWrap(true) did not flip field")
	}
}

func TestKeyMapGetter(t *testing.T) {
	m := NewLogViewer()
	km := m.KeyMap()
	want := DefaultKeyMap()
	if km.ToggleFollow.Keys()[0] != want.ToggleFollow.Keys()[0] {
		t.Errorf("KeyMap() ToggleFollow mismatch")
	}
}

func TestFocusedGetter(t *testing.T) {
	m := NewLogViewer()
	if m.Focused() {
		t.Error("default Focused: want false")
	}
	m.Focus()
	if !m.Focused() {
		t.Error("after Focus: want true")
	}
	m.Blur()
	if m.Focused() {
		t.Error("after Blur: want false")
	}
}

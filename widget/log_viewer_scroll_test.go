package widget

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

// logViewerKey constructs a KeyPressMsg whose String() matches the given
// textual representation. For printable single characters the Text
// field is used; for named keys (up, pgup, etc.) the Code field is
// set.
func logViewerKey(s string) tea.KeyPressMsg {
	switch s {
	case "up":
		return tea.KeyPressMsg{Code: tea.KeyUp}
	case "down":
		return tea.KeyPressMsg{Code: tea.KeyDown}
	case "pgup":
		return tea.KeyPressMsg{Code: tea.KeyPgUp}
	case "pgdown":
		return tea.KeyPressMsg{Code: tea.KeyPgDown}
	case "home":
		return tea.KeyPressMsg{Code: tea.KeyHome}
	case "end":
		return tea.KeyPressMsg{Code: tea.KeyEnd}
	default:
		// Single printable rune (e.g. "g", "G", "f", "k").
		r := []rune(s)
		return tea.KeyPressMsg{Code: r[0], Text: s}
	}
}

// fillFollowing primes a LogViewer such that it is following and has
// more lines than the viewport can show. The returned viewer has size
// 20x6 (5 content rows + 1 footer row), is focused so key dispatch is
// live, and has 30 lines appended.
func fillFollowing(t *testing.T) *LogViewer {
	t.Helper()
	m := NewLogViewer()
	m.Focus()
	m.SetSize(20, 6)
	for i := 0; i < 30; i++ {
		m.Append("line")
	}
	if !m.Following() {
		t.Fatal("precondition: Following should be true")
	}
	if !m.viewport.AtBottom() {
		t.Fatal("precondition: viewport should be at bottom")
	}
	return m
}

// fillPaused primes a LogViewer that is paused and scrolled to the
// top. Same dimensions as fillFollowing.
func fillPaused(t *testing.T) *LogViewer {
	t.Helper()
	m := fillFollowing(t)
	m.SetFollowing(false)
	m.viewport.GotoTop()
	if !m.viewport.AtTop() {
		t.Fatal("precondition: viewport should be at top")
	}
	return m
}

func TestScrollStateMachine(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(*testing.T) *LogViewer
		action func(*LogViewer)
		check  func(t *testing.T, m *LogViewer)
	}{
		{
			name:  "following then append stays following at tail",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				m.Append("more")
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("want Following true, got false")
				}
				if !m.viewport.AtBottom() {
					t.Error("want viewport at bottom")
				}
			},
		},
		{
			name:  "following then scroll-up key pauses and lifts off tail",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("up"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after scroll-up")
				}
				if m.viewport.AtBottom() {
					t.Error("want viewport NOT at bottom")
				}
			},
		},
		{
			name:  "paused then append does not jump viewport",
			setup: fillPaused,
			action: func(m *LogViewer) {
				m.Append("another")
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false")
				}
				if !m.viewport.AtTop() {
					t.Error("want viewport still at top")
				}
			},
		},
		{
			name:  "paused then G re-engages follow and jumps to tail",
			setup: fillPaused,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("G"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("want Following true after G")
				}
				if !m.viewport.AtBottom() {
					t.Error("want viewport at bottom after G")
				}
			},
		},
		{
			name:  "paused then end re-engages follow",
			setup: fillPaused,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("end"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("want Following true after end")
				}
				if !m.viewport.AtBottom() {
					t.Error("want viewport at bottom after end")
				}
			},
		},
		{
			name:  "f toggles follow off when following",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("f"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after f")
				}
			},
		},
		{
			name:  "f toggles follow on when paused",
			setup: fillPaused,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("f"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("want Following true after f")
				}
			},
		},
		{
			name:  "paused then SetSize preserves paused state",
			setup: fillPaused,
			action: func(m *LogViewer) {
				m.SetSize(60, 12)
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after SetSize while paused")
				}
				if m.viewport.AtBottom() {
					t.Error("want viewport NOT at bottom after SetSize while paused")
				}
			},
		},
		{
			name:  "following then SetSize keeps viewport at tail",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				m.SetSize(60, 12)
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if !m.Following() {
					t.Error("want Following true after SetSize while following")
				}
				if !m.viewport.AtBottom() {
					t.Error("want viewport at bottom after SetSize while following")
				}
			},
		},
		{
			name:  "following then k pauses",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("k"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after k")
				}
			},
		},
		{
			name:  "following then pgup pauses",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("pgup"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after pgup")
				}
			},
		},
		{
			name:  "following then home pauses",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("home"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after home")
				}
			},
		},
		{
			name:  "following then g pauses",
			setup: fillFollowing,
			action: func(m *LogViewer) {
				_, _ = m.Update(logViewerKey("g"))
			},
			check: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.Following() {
					t.Error("want Following false after g")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setup(t)
			tc.action(m)
			tc.check(t, m)
		})
	}
}

func TestMouseWheelUpPauses(t *testing.T) {
	m := fillFollowing(t)
	_, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelUp})
	if m.Following() {
		t.Error("want Following false after mouse wheel up")
	}
}

func TestGoldenFooterFollow(t *testing.T) {
	m := NewLogViewer()
	m.SetSize(80, 24)
	m.Append(
		"2026-04-17 10:00:00 INFO starting up",
		"2026-04-17 10:00:01 INFO ready",
		"2026-04-17 10:00:02 WARN retrying",
	)
	if !m.Following() {
		t.Fatal("precondition: follow should be on")
	}
	if err := logViewerSnapshotter().SnapshotMulti("footer_follow", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

func TestGoldenFooterPaused(t *testing.T) {
	m := NewLogViewer()
	m.SetSize(80, 24)
	m.Append(
		"2026-04-17 10:00:00 INFO starting up",
		"2026-04-17 10:00:01 INFO ready",
		"2026-04-17 10:00:02 WARN retrying",
	)
	m.SetFollowing(false)
	if m.Following() {
		t.Fatal("precondition: follow should be off")
	}
	if err := logViewerSnapshotter().SnapshotMulti("footer_paused", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

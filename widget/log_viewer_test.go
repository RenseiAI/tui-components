package widget

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/bradleyjkemp/cupaloy/v2"
)

// snapshotter returns a cupaloy config pinned at widget/.snapshots.
// Defaults preserve cupaloy's env-driven behaviour: when
// UPDATE_SNAPSHOTS=true, missing snapshots are created and existing
// ones overwritten; otherwise mismatches fail the test.
func snapshotter() *cupaloy.Config {
	return cupaloy.New(
		cupaloy.SnapshotSubdirectory(".snapshots"),
		cupaloy.EnvVariableName("UPDATE_SNAPSHOTS"),
	)
}

// viewContent extracts the plain-string content from a tea.View for
// golden comparison.
func viewContent(v tea.View) string {
	return v.Content
}

func TestNewDefaults(t *testing.T) {
	m := New()
	if m.maxLines != defaultMaxLines {
		t.Errorf("default maxLines: want %d, got %d", defaultMaxLines, m.maxLines)
	}
	if !m.wrap {
		t.Error("default wrap: want true, got false")
	}
	if !m.follow {
		t.Error("default follow: want true, got false")
	}
	if m.focused {
		t.Error("default focused: want false, got true")
	}
}

func TestOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want func(t *testing.T, m *LogViewer)
	}{
		{
			name: "with max lines",
			opts: []Option{WithMaxLines(50)},
			want: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.maxLines != 50 {
					t.Errorf("maxLines: want 50, got %d", m.maxLines)
				}
			},
		},
		{
			name: "with max lines zero means unbounded",
			opts: []Option{WithMaxLines(0)},
			want: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.maxLines != 0 {
					t.Errorf("maxLines: want 0, got %d", m.maxLines)
				}
			},
		},
		{
			name: "with max lines negative falls back to default",
			opts: []Option{WithMaxLines(-5)},
			want: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.maxLines != defaultMaxLines {
					t.Errorf("maxLines: want %d (default), got %d", defaultMaxLines, m.maxLines)
				}
			},
		},
		{
			name: "with wrap off",
			opts: []Option{WithWrap(false)},
			want: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.wrap {
					t.Error("wrap: want false, got true")
				}
			},
		},
		{
			name: "with follow off",
			opts: []Option{WithFollow(false)},
			want: func(t *testing.T, m *LogViewer) {
				t.Helper()
				if m.follow {
					t.Error("follow: want false, got true")
				}
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.opts...)
			tc.want(t, m)
		})
	}
}

func TestAppendThenClear(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append("one", "two", "three")
	if len(m.lines) != 3 {
		t.Fatalf("want 3 lines, got %d", len(m.lines))
	}
	m.Clear()
	if len(m.lines) != 0 {
		t.Errorf("want 0 lines after clear, got %d", len(m.lines))
	}
	// View should render empty content (no retained lines).
	if got := strings.TrimSpace(viewContent(m.View())); got != "" {
		t.Errorf("view after clear: want empty, got %q", got)
	}
}

func TestRingOverflow(t *testing.T) {
	m := New(WithMaxLines(3))
	m.SetSize(80, 24)
	m.Append("a", "b", "c", "d", "e")
	if len(m.lines) != 3 {
		t.Fatalf("want 3 lines retained, got %d", len(m.lines))
	}
	want := []string{"c", "d", "e"}
	for i, line := range want {
		if m.lines[i] != line {
			t.Errorf("line[%d]: want %q, got %q", i, line, m.lines[i])
		}
	}
}

func TestRingUnbounded(t *testing.T) {
	m := New(WithMaxLines(0))
	for i := 0; i < 20; i++ {
		m.Append("x")
	}
	if len(m.lines) != 20 {
		t.Errorf("unbounded: want 20 lines retained, got %d", len(m.lines))
	}
}

func TestAppendMsgViaUpdate(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	updated, cmd := m.Update(AppendMsg{Lines: []string{"alpha", "beta"}})
	if cmd != nil {
		t.Errorf("want nil cmd, got %v", cmd)
	}
	lv, ok := updated.(*LogViewer)
	if !ok {
		t.Fatalf("Update did not return *LogViewer; got %T", updated)
	}
	if len(lv.lines) != 2 {
		t.Errorf("want 2 lines after AppendMsg, got %d", len(lv.lines))
	}
}

func TestUpdateIgnoresOtherMessages(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append("before")
	before := len(m.lines)
	_, _ = m.Update(struct{}{})
	if len(m.lines) != before {
		t.Errorf("unknown msg mutated buffer: want %d, got %d", before, len(m.lines))
	}
}

func TestUpdateHandlesWindowSize(t *testing.T) {
	m := New()
	_, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	if m.width != 40 || m.height != 10 {
		t.Errorf("WindowSizeMsg not propagated: want 40x10, got %dx%d", m.width, m.height)
	}
}

func TestFollowingGetterSetter(t *testing.T) {
	m := New(WithFollow(false))
	if m.Following() {
		t.Error("Following(): want false after WithFollow(false)")
	}
	m.SetFollowing(true)
	if !m.Following() {
		t.Error("Following(): want true after SetFollowing(true)")
	}
}

func TestFollowScrollBehaviour(t *testing.T) {
	// Append enough lines to overflow a 5-row viewport, then verify:
	//  - with follow on, viewport ends at the bottom.
	//  - with follow off, viewport stays at its prior offset after a
	//    subsequent append.
	m := New()
	m.SetSize(20, 5)
	for i := 0; i < 30; i++ {
		m.Append("line")
	}
	if !m.viewport.AtBottom() {
		t.Error("follow on: viewport should be at bottom after appends")
	}

	// Pause follow and scroll to top manually to simulate user scroll.
	m.SetFollowing(false)
	m.viewport.GotoTop()
	if !m.viewport.AtTop() {
		t.Fatal("failed to seed AtTop state")
	}
	offsetBefore := m.viewport.YOffset()

	m.Append("another line")
	if m.viewport.YOffset() != offsetBefore {
		t.Errorf("follow off: offset changed after append: before=%d after=%d",
			offsetBefore, m.viewport.YOffset())
	}
	if m.viewport.AtBottom() && offsetBefore == 0 {
		// This can be true only if the buffer fits in the viewport,
		// which our fixture avoids.
		t.Error("follow off: viewport autoscrolled to bottom")
	}

	// Re-enable follow and append — should jump to bottom.
	m.SetFollowing(true)
	m.Append("last line")
	if !m.viewport.AtBottom() {
		t.Error("follow on after re-enable: viewport should be at bottom")
	}
}

func TestFocusBlurStubs(t *testing.T) {
	m := New()
	m.Focus()
	if !m.focused {
		t.Error("Focus() did not set internal flag")
	}
	m.Blur()
	if m.focused {
		t.Error("Blur() did not clear internal flag")
	}
}

func TestAppendNoOpOnEmpty(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append()
	if len(m.lines) != 0 {
		t.Errorf("want 0 lines, got %d", len(m.lines))
	}
}

// -- Golden-based rendering tests ------------------------------------

func TestGoldenEmpty(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	if err := snapshotter().SnapshotMulti("empty", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

func TestGoldenPlainShortLog(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append(
		"2026-04-17 10:00:00 INFO starting up",
		"2026-04-17 10:00:01 INFO ready",
		"2026-04-17 10:00:02 WARN retrying",
	)
	if err := snapshotter().SnapshotMulti("plain_short", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

func TestGoldenAnsiColouredLog(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append(
		"\x1b[31mred text\x1b[0m plain tail",
		"\x1b[1;32mbold green\x1b[0m",
		"\x1b[33myellow\x1b[39m default-again",
	)
	if err := snapshotter().SnapshotMulti("ansi_coloured", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

func TestGoldenLongLineWrapOn(t *testing.T) {
	m := New()
	m.SetSize(20, 24)
	long := strings.Repeat("abcdefghij", 6) // 60 chars, well over 20
	m.Append(long)
	if err := snapshotter().SnapshotMulti("long_line_wrap_on", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

func TestGoldenLongLineWrapOff(t *testing.T) {
	m := New(WithWrap(false))
	m.SetSize(20, 24)
	long := strings.Repeat("abcdefghij", 6)
	m.Append(long)
	if err := snapshotter().SnapshotMulti("long_line_wrap_off", viewContent(m.View())); err != nil {
		t.Fatalf("snapshot mismatch: %v", err)
	}
}

// Explicit regression: rendered output of an ANSI-styled line must
// contain some SGR escape sequence — we don't lose colour on the way
// through the pipeline.
func TestAnsiRenderingContainsStyledSequence(t *testing.T) {
	m := New()
	m.SetSize(80, 24)
	m.Append("\x1b[31mred\x1b[0m")
	got := viewContent(m.View())
	if !strings.Contains(got, "\x1b[") {
		t.Errorf("expected rendered output to contain an ANSI escape sequence, got %q", got)
	}
}

// Wrap-on renders a long line across multiple rows; wrap-off preserves
// it on a single row (horizontal overflow handled by viewport).
func TestWrapBehaviour(t *testing.T) {
	long := strings.Repeat("abcdefghij", 6) // 60 chars
	tests := []struct {
		name        string
		opts        []Option
		width       int
		wantWrapped bool
	}{
		{
			name:        "wrap on",
			opts:        []Option{WithWrap(true)},
			width:       20,
			wantWrapped: true,
		},
		{
			name:        "wrap off",
			opts:        []Option{WithWrap(false)},
			width:       20,
			wantWrapped: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := New(tc.opts...)
			m.SetSize(tc.width, 24)
			m.Append(long)
			got := viewContent(m.View())
			// Count non-empty output lines. Wrap=on should produce
			// multiple; wrap=off the single original.
			lines := 0
			for _, l := range strings.Split(got, "\n") {
				if strings.TrimSpace(l) != "" {
					lines++
				}
			}
			if tc.wantWrapped && lines < 2 {
				t.Errorf("wrap on: expected multiple non-empty rows, got %d (%q)", lines, got)
			}
			if !tc.wantWrapped && lines != 1 {
				t.Errorf("wrap off: expected 1 non-empty row, got %d (%q)", lines, got)
			}
		})
	}
}

func TestInitReturnsNil(t *testing.T) {
	m := New()
	if cmd := m.Init(); cmd != nil {
		t.Errorf("Init(): want nil cmd, got %v", cmd)
	}
}

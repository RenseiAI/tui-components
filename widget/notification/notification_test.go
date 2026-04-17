package notification_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/widget/notification"
)

// Runtime interface assertion (complements the compile-time one in
// notification.go).
var _ component.Component = (*notification.Model)(nil)

// asModel type-asserts a tea.Model back to notification.Model so tests
// can read package-level state on the returned value.
func asModel(t *testing.T, m tea.Model) notification.Model {
	t.Helper()
	got, ok := m.(notification.Model)
	if !ok {
		t.Fatalf("expected notification.Model, got %T", m)
	}
	return got
}

// dismissOf runs Init() and executes the resulting tick command,
// returning the dismissMsg-shaped message that the model would have
// received in production. The message type is private to the package,
// so callers route it back through Update via tea.Msg.
func dismissOf(t *testing.T, m notification.Model) tea.Msg {
	t.Helper()
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init returned nil cmd")
	}
	msg := cmd()
	if msg == nil {
		t.Fatal("Init cmd produced nil msg")
	}
	return msg
}

func TestNew_Defaults(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok")
	if got, want := m.Variant(), notification.VariantSuccess; got != want {
		t.Errorf("Variant = %q, want %q", got, want)
	}
	if got, want := m.Message(), "ok"; got != want {
		t.Errorf("Message = %q, want %q", got, want)
	}
	if m.Dismissed() {
		t.Error("Dismissed = true on fresh Model; want false")
	}
}

func TestNew_OptionsApplied(t *testing.T) {
	tests := []struct {
		name    string
		variant notification.Variant
		opts    []notification.Option
		check   func(t *testing.T, m notification.Model)
	}{
		{
			name:    "WithDuration positive",
			variant: notification.VariantSuccess,
			opts:    []notification.Option{notification.WithDuration(2 * time.Second)},
			check: func(t *testing.T, m notification.Model) {
				// Duration is internal; verify indirectly: SetSize+View
				// must still render with explicit width.
				m.SetSize(40, 0)
				if m.View().Content == "" {
					t.Error("View empty after SetSize(40)")
				}
			},
		},
		{
			name:    "WithDuration zero coerces to default",
			variant: notification.VariantSuccess,
			opts:    []notification.Option{notification.WithDuration(0)},
			check: func(t *testing.T, m notification.Model) {
				m.SetSize(40, 0)
				if m.View().Content == "" {
					t.Error("View empty after SetSize(40)")
				}
			},
		},
		{
			name:    "WithDuration negative coerces to default",
			variant: notification.VariantWarning,
			opts:    []notification.Option{notification.WithDuration(-5 * time.Second)},
			check: func(t *testing.T, m notification.Model) {
				m.SetSize(40, 0)
				if m.View().Content == "" {
					t.Error("View empty after SetSize(40)")
				}
			},
		},
		{
			name:    "WithWidth sets render width",
			variant: notification.VariantError,
			opts:    []notification.Option{notification.WithWidth(40)},
			check: func(t *testing.T, m notification.Model) {
				if m.View().Content == "" {
					t.Error("View empty when WithWidth(40) set")
				}
			},
		},
		{
			name:    "WithIcon overrides default glyph",
			variant: notification.VariantSuccess,
			opts: []notification.Option{
				notification.WithWidth(40),
				notification.WithIcon("!!"),
			},
			check: func(t *testing.T, m notification.Model) {
				if !strings.Contains(m.View().Content, "!!") {
					t.Errorf("View missing custom icon; got %q", m.View().Content)
				}
				if strings.Contains(m.View().Content, "\u2713") {
					t.Errorf("View still contains default success glyph; got %q", m.View().Content)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := notification.New(tt.variant, "msg", tt.opts...)
			tt.check(t, m)
		})
	}
}

func TestUpdate_DismissMatchingID_Flips(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok",
		notification.WithDuration(50*time.Millisecond),
		notification.WithWidth(40),
	)
	msg := dismissOf(t, m)

	next, cmd := m.Update(msg)
	got := asModel(t, next)
	if !got.Dismissed() {
		t.Error("Dismissed = false after matching dismissMsg; want true")
	}
	if cmd != nil {
		t.Error("Update returned non-nil cmd; expected nil")
	}
	if got.View().Content != "" {
		t.Errorf("dismissed View = %q, want empty", got.View().Content)
	}
}

func TestUpdate_DismissStaleID_NoOp(t *testing.T) {
	m1 := notification.New(notification.VariantSuccess, "first",
		notification.WithDuration(50*time.Millisecond),
		notification.WithWidth(40),
	)
	staleMsg := dismissOf(t, m1)

	// A second Model gets a fresh genID via the package-level counter.
	m2 := notification.New(notification.VariantSuccess, "second",
		notification.WithDuration(50*time.Millisecond),
		notification.WithWidth(40),
	)

	next, cmd := m2.Update(staleMsg)
	got := asModel(t, next)
	if got.Dismissed() {
		t.Error("Dismissed = true after stale dismissMsg; want false")
	}
	if cmd != nil {
		t.Error("Update returned non-nil cmd; expected nil")
	}
	if got.View().Content == "" {
		t.Error("View empty after stale tick; non-dismissed model should still render")
	}
}

func TestUpdate_UnrelatedMessage_NoOp(t *testing.T) {
	type customMsg struct{}
	m := notification.New(notification.VariantWarning, "hi",
		notification.WithWidth(40),
	)
	before := m.View().Content

	next, cmd := m.Update(customMsg{})
	got := asModel(t, next)
	if got.Dismissed() {
		t.Error("Dismissed flipped on unrelated message")
	}
	if cmd != nil {
		t.Error("Update returned non-nil cmd; expected nil")
	}
	if got.View().Content != before {
		t.Errorf("View changed after unrelated message: before=%q after=%q",
			before, got.View().Content)
	}
}

func TestView_ZeroWidth_Empty(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok")
	if got := m.View().Content; got != "" {
		t.Errorf("View() with width=0 = %q, want empty", got)
	}
}

func TestView_NegativeWidth_Empty(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok",
		notification.WithWidth(-10),
	)
	if got := m.View().Content; got != "" {
		t.Errorf("View() with width<0 = %q, want empty", got)
	}
}

func TestSetSize_UpdatesRenderedWidth(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok")
	if m.View().Content != "" {
		t.Fatal("precondition: View should be empty before SetSize")
	}
	m.SetSize(40, 1)
	if m.View().Content == "" {
		t.Error("View empty after SetSize(40, 1); expected rendered notification")
	}
}

func TestFocusBlur_NoOp(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "ok",
		notification.WithWidth(40),
	)
	before := m.View().Content
	m.Focus()
	m.Blur()
	if m.View().Content != before {
		t.Errorf("View changed across Focus/Blur: before=%q after=%q",
			before, m.View().Content)
	}
}

func TestInit_ReturnsTickCmd(t *testing.T) {
	m := notification.New(notification.VariantError, "boom",
		notification.WithDuration(10*time.Millisecond),
	)
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init returned nil cmd")
	}
	// Execute the cmd; it must produce a message that, when fed back
	// through Update, dismisses the same Model.
	msg := cmd()
	if msg == nil {
		t.Fatal("Init cmd produced nil msg")
	}
	next, _ := m.Update(msg)
	if !asModel(t, next).Dismissed() {
		t.Error("Init-produced tick did not dismiss the originating Model")
	}
}

// TestModel_Golden_VariantSuccess captures the rendered output of the
// success variant at a fixed width.
func TestModel_Golden_VariantSuccess(t *testing.T) {
	m := notification.New(notification.VariantSuccess, "Saved successfully",
		notification.WithWidth(40),
	)
	cupaloy.SnapshotT(t, m.View().Content)
}

// TestModel_Golden_VariantWarning captures the rendered output of the
// warning variant at a fixed width.
func TestModel_Golden_VariantWarning(t *testing.T) {
	m := notification.New(notification.VariantWarning, "Heads up",
		notification.WithWidth(40),
	)
	cupaloy.SnapshotT(t, m.View().Content)
}

// TestModel_Golden_VariantError captures the rendered output of the
// error variant at a fixed width.
func TestModel_Golden_VariantError(t *testing.T) {
	m := notification.New(notification.VariantError, "Operation failed",
		notification.WithWidth(40),
	)
	cupaloy.SnapshotT(t, m.View().Content)
}

// TestModel_Golden_UnknownFallback verifies that an unknown variant
// renders with the neutral fallback style and "?" icon. The package
// also emits a warn log on construction; that side-effect is not
// asserted here.
func TestModel_Golden_UnknownFallback(t *testing.T) {
	m := notification.New(notification.Variant("mystery"), "what is this",
		notification.WithWidth(40),
	)
	cupaloy.SnapshotT(t, m.View().Content)
}

// TestModel_Golden_LongMessageWrap verifies that a message longer than
// the configured width wraps inside the rendered border instead of
// overflowing.
func TestModel_Golden_LongMessageWrap(t *testing.T) {
	m := notification.New(notification.VariantSuccess,
		"This is an unusually long notification body that must wrap inside the rounded border at the configured width without overflowing.",
		notification.WithWidth(40),
	)
	cupaloy.SnapshotT(t, m.View().Content)
}

func TestNew_OptionsCombined(t *testing.T) {
	m := notification.New(notification.VariantWarning, "combined",
		notification.WithDuration(75*time.Millisecond),
		notification.WithIcon("[!]"),
		notification.WithWidth(50),
	)
	view := m.View().Content
	if !strings.Contains(view, "[!]") {
		t.Errorf("expected custom icon '[!]'; got %q", view)
	}
	if !strings.Contains(view, "combined") {
		t.Errorf("expected message text; got %q", view)
	}
	if got, want := lipgloss.Width(view), 50; got != want {
		t.Errorf("rendered width = %d, want %d", got, want)
	}
}

func TestSetSize_LongMessageWraps(t *testing.T) {
	m := notification.New(notification.VariantSuccess,
		strings.Repeat("word ", 30),
	)
	m.SetSize(30, 0)
	view := m.View().Content
	if !strings.Contains(view, "word") {
		t.Fatalf("expected wrapped content to contain 'word'; got %q", view)
	}
	lines := strings.Split(view, "\n")
	if len(lines) < 4 {
		t.Errorf("expected wrapped content to span multiple lines; got %d lines", len(lines))
	}
	for i, line := range lines {
		if w := lipgloss.Width(line); w > 30 {
			t.Errorf("line %d width = %d, exceeds configured width 30: %q", i, w, line)
		}
	}
}

func TestVariantStyle_DistinctColorsPerVariant(t *testing.T) {
	out := map[notification.Variant]string{}
	for _, v := range []notification.Variant{
		notification.VariantSuccess,
		notification.VariantWarning,
		notification.VariantError,
	} {
		m := notification.New(v, "x", notification.WithWidth(20))
		out[v] = m.View().Content
	}
	if out[notification.VariantSuccess] == out[notification.VariantWarning] ||
		out[notification.VariantSuccess] == out[notification.VariantError] ||
		out[notification.VariantWarning] == out[notification.VariantError] {
		t.Error("expected distinct ANSI rendering per variant; got duplicates")
	}
}

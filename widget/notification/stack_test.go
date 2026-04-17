package notification_test

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/widget/notification"
)

// Runtime interface assertions complementing the compile-time ones in
// stack.go.
var (
	_ tea.Model           = (*notification.Stack)(nil)
	_ component.Component = (*notification.Stack)(nil)
)

// asStack type-asserts a tea.Model back to notification.Stack.
func asStack(t *testing.T, m tea.Model) notification.Stack {
	t.Helper()
	s, ok := m.(notification.Stack)
	if !ok {
		t.Fatalf("expected notification.Stack, got %T", m)
	}
	return s
}

// pushAll builds a stack from the given models. It does not execute
// the auto-dismiss tick commands — tests that need a child's dismiss
// message must capture the cmd from Push directly (see helpers below)
// and ensure that child has a short [notification.WithDuration].
func pushAll(t *testing.T, models []notification.Model, opts ...notification.StackOption) notification.Stack {
	t.Helper()
	s := notification.NewStack(opts...)
	for _, m := range models {
		var cmd tea.Cmd
		s, cmd = s.Push(m)
		if cmd == nil {
			t.Fatalf("Push returned nil cmd")
		}
	}
	return s
}

// pushAndCaptureDismiss pushes a single model and synchronously fires
// its tick command to obtain the resulting dismissMsg. The provided
// model MUST have a short [notification.WithDuration]; otherwise this
// helper blocks for the default 4 s duration.
func pushAndCaptureDismiss(t *testing.T, s notification.Stack, m notification.Model) (notification.Stack, tea.Msg) {
	t.Helper()
	next, cmd := s.Push(m)
	if cmd == nil {
		t.Fatal("Push returned nil cmd")
	}
	return next, cmd()
}

func TestNewStack_EmptyDefaults(t *testing.T) {
	s := notification.NewStack()
	if got := s.Len(); got != 0 {
		t.Errorf("Len = %d, want 0", got)
	}
	if got := s.View().Content; got != "" {
		t.Errorf("View = %q, want empty", got)
	}
}

func TestStack_Push_AppendsAndReturnsTickCmd(t *testing.T) {
	s := notification.NewStack(notification.WithStackWidth(40))
	s, cmd := s.Push(notification.New(notification.VariantSuccess, "hi"))
	if cmd == nil {
		t.Fatal("Push returned nil cmd")
	}
	if got := s.Len(); got != 1 {
		t.Errorf("Len = %d, want 1", got)
	}
	if !strings.Contains(s.View().Content, "hi") {
		t.Errorf("View missing pushed message; got %q", s.View().Content)
	}
}

func TestStack_Update_PrunesDismissedChildren(t *testing.T) {
	short := notification.WithDuration(20 * time.Millisecond)
	s := notification.NewStack(notification.WithStackWidth(40))
	s, _ = s.Push(notification.New(notification.VariantSuccess, "a", short))
	s, dismissB := pushAndCaptureDismiss(t, s,
		notification.New(notification.VariantWarning, "b", short))
	s, _ = s.Push(notification.New(notification.VariantError, "c", short))

	if s.Len() != 3 {
		t.Fatalf("precondition: Len = %d, want 3", s.Len())
	}

	// Dismiss only the middle item.
	next, _ := s.Update(dismissB)
	s = asStack(t, next)

	if s.Len() != 2 {
		t.Errorf("after dismiss of middle, Len = %d, want 2", s.Len())
	}
	if !strings.Contains(s.View().Content, "a") || !strings.Contains(s.View().Content, "c") {
		t.Errorf("expected 'a' and 'c' to survive prune; got %q", s.View().Content)
	}
	if strings.Contains(s.View().Content, "b") {
		t.Errorf("dismissed 'b' still rendered; got %q", s.View().Content)
	}
}

func TestStack_Update_OrderPreserved(t *testing.T) {
	s := pushAll(t, []notification.Model{
		notification.New(notification.VariantSuccess, "first"),
		notification.New(notification.VariantWarning, "second"),
		notification.New(notification.VariantError, "third"),
	}, notification.WithStackWidth(40))

	view := s.View().Content
	posFirst := strings.Index(view, "first")
	posSecond := strings.Index(view, "second")
	posThird := strings.Index(view, "third")
	if posFirst < 0 || posSecond < 0 || posThird < 0 {
		t.Fatalf("missing label in view: %q", view)
	}
	if posFirst >= posSecond || posSecond >= posThird {
		t.Errorf("default order should be insertion order top→bottom; got positions first=%d second=%d third=%d",
			posFirst, posSecond, posThird)
	}
}

func TestStack_NewestFirst_InvertsOrder(t *testing.T) {
	s := pushAll(t, []notification.Model{
		notification.New(notification.VariantSuccess, "first"),
		notification.New(notification.VariantWarning, "second"),
		notification.New(notification.VariantError, "third"),
	}, notification.WithStackWidth(40), notification.WithNewestFirst())

	view := s.View().Content
	posFirst := strings.Index(view, "first")
	posThird := strings.Index(view, "third")
	if posThird >= posFirst {
		t.Errorf("WithNewestFirst should put 'third' before 'first'; got positions third=%d first=%d",
			posThird, posFirst)
	}
}

func TestStack_WithMax_EvictsOldest(t *testing.T) {
	s := notification.NewStack(
		notification.WithStackWidth(40),
		notification.WithMax(2),
	)
	for _, msg := range []string{"a", "b", "c"} {
		s, _ = s.Push(notification.New(notification.VariantSuccess, msg))
	}
	if got := s.Len(); got != 2 {
		t.Errorf("Len = %d, want 2 after eviction", got)
	}
	view := s.View().Content
	if strings.Contains(view, "a") {
		t.Errorf("oldest 'a' should have been evicted; got %q", view)
	}
	if !strings.Contains(view, "b") || !strings.Contains(view, "c") {
		t.Errorf("expected 'b' and 'c' to survive; got %q", view)
	}
}

func TestStack_WithMax_NonPositiveCoercesToDefault(t *testing.T) {
	s := notification.NewStack(notification.WithMax(-3))
	// Pushing more than DefaultStackMax (5) should evict.
	for i := 0; i < 7; i++ {
		s, _ = s.Push(notification.New(notification.VariantSuccess, "x",
			notification.WithWidth(40)))
	}
	if got, want := s.Len(), notification.DefaultStackMax; got != want {
		t.Errorf("Len = %d, want %d (DefaultStackMax)", got, want)
	}
}

func TestStack_Update_StaleTickIgnored(t *testing.T) {
	// Two pushes of the same payload — Push refreshes genID, so the
	// first tick must not dismiss the second instance after a
	// hypothetical replacement scenario.
	s := notification.NewStack(notification.WithStackWidth(40))

	short := notification.WithDuration(20 * time.Millisecond)
	first := notification.New(notification.VariantSuccess, "one", short)
	s2, cmd1 := s.Push(first)
	staleMsg := cmd1()

	second := notification.New(notification.VariantWarning, "two", short)
	s3, _ := s2.Push(second)

	// Feed the *first* tick — it should dismiss only "one".
	next, _ := s3.Update(staleMsg)
	s4 := asStack(t, next)

	if s4.Len() != 1 {
		t.Fatalf("after dismissing 'one', Len = %d, want 1", s4.Len())
	}
	if strings.Contains(s4.View().Content, "one") {
		t.Errorf("'one' should be pruned; got %q", s4.View().Content)
	}
	if !strings.Contains(s4.View().Content, "two") {
		t.Errorf("'two' should remain; got %q", s4.View().Content)
	}
}

func TestStack_Update_UnrelatedMessage_NoOp(t *testing.T) {
	type customMsg struct{}
	s := pushAll(t, []notification.Model{
		notification.New(notification.VariantSuccess, "x"),
		notification.New(notification.VariantWarning, "y"),
	}, notification.WithStackWidth(40))

	before := s.View().Content
	next, _ := s.Update(customMsg{})
	got := asStack(t, next)
	if got.Len() != 2 {
		t.Errorf("Len changed after unrelated message: got %d, want 2", got.Len())
	}
	if got.View().Content != before {
		t.Errorf("View changed after unrelated message: before=%q after=%q",
			before, got.View().Content)
	}
}

func TestStack_SetSize_PropagatesToChildren(t *testing.T) {
	s := notification.NewStack()
	s, _ = s.Push(notification.New(notification.VariantSuccess, "msg"))
	if s.View().Content != "" {
		t.Fatal("precondition: View should be empty before width is set")
	}
	s.SetSize(40, 0)
	if s.View().Content == "" {
		t.Error("View empty after SetSize; expected propagation to child")
	}
}

func TestStack_Init_Nil(t *testing.T) {
	s := notification.NewStack()
	if cmd := s.Init(); cmd != nil {
		t.Errorf("Init returned non-nil cmd; want nil")
	}
}

func TestStack_FocusBlur_NoOp(t *testing.T) {
	s := notification.NewStack(notification.WithStackWidth(40))
	s, _ = s.Push(notification.New(notification.VariantSuccess, "msg"))
	before := s.View().Content
	s.Focus()
	s.Blur()
	if s.View().Content != before {
		t.Error("Focus/Blur mutated stack View")
	}
}

func TestStack_Update_OnlyMatchingChildDismissed(t *testing.T) {
	s := notification.NewStack(notification.WithStackWidth(40))
	a := notification.New(notification.VariantSuccess, "a",
		notification.WithDuration(50*time.Millisecond))
	b := notification.New(notification.VariantWarning, "b",
		notification.WithDuration(50*time.Millisecond))

	s, cmdA := s.Push(a)
	s, _ = s.Push(b)
	dismissA := cmdA()

	next, _ := s.Update(dismissA)
	s = asStack(t, next)
	if !strings.Contains(s.View().Content, "b") {
		t.Errorf("'b' should be unaffected; got %q", s.View().Content)
	}
	if strings.Contains(s.View().Content, "a") {
		t.Errorf("'a' should be dismissed; got %q", s.View().Content)
	}
}

func TestStack_RapidPushes_AllSurviveUntilOwnTick(t *testing.T) {
	short := notification.WithDuration(20 * time.Millisecond)
	s := notification.NewStack(notification.WithStackWidth(40))
	for _, v := range []notification.Variant{
		notification.VariantSuccess,
		notification.VariantWarning,
		notification.VariantError,
		notification.VariantSuccess,
		notification.VariantWarning,
	} {
		s, _ = s.Push(notification.New(v, "msg", short))
	}
	if got, want := s.Len(), 5; got != want {
		t.Errorf("Len after rapid pushes = %d, want %d", got, want)
	}
}

// TestStack_Golden_ThreeVariants establishes a baseline render of a
// three-element stack at width 40.
func TestStack_Golden_ThreeVariants(t *testing.T) {
	s := pushAll(t, []notification.Model{
		notification.New(notification.VariantSuccess, "Saved"),
		notification.New(notification.VariantWarning, "Heads up"),
		notification.New(notification.VariantError, "Failed"),
	}, notification.WithStackWidth(40))
	cupaloy.SnapshotT(t, s.View().Content)
}

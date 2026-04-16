package widget_test

import (
	"strings"
	"testing"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/widget"
)

// Runtime interface assertion (complements the compile-time assertion in
// spinner.go).
var _ component.Component = (*widget.Spinner)(nil)

// asSpinner type-asserts a tea.Model back to *widget.Spinner. Tests use
// this to access the concrete Spinner after Update.
func asSpinner(t *testing.T, m tea.Model) *widget.Spinner {
	t.Helper()
	s, ok := m.(*widget.Spinner)
	if !ok {
		t.Fatalf("expected *widget.Spinner, got %T", m)
	}
	return s
}

// tickFromInit runs Init to get the initial tick command and executes it,
// returning a real spinner.TickMsg carrying the spinner's internal ID so
// the inner Bubbles spinner will accept it.
func tickFromInit(t *testing.T, s *widget.Spinner) spinner.TickMsg {
	t.Helper()
	cmd := s.Init()
	if cmd == nil {
		t.Fatalf("Init() returned nil cmd")
	}
	msg := cmd()
	tick, ok := msg.(spinner.TickMsg)
	if !ok {
		t.Fatalf("expected spinner.TickMsg from Init cmd, got %T", msg)
	}
	return tick
}

func TestNewSpinner_Defaults(t *testing.T) {
	s := widget.NewSpinner()
	if s == nil {
		t.Fatal("NewSpinner() returned nil")
	}
	v := s.View()
	if v.Content == "" {
		t.Fatal("View().Content is empty for default spinner")
	}
	// Default has no label, so a label wouldn't appear. We cannot assert
	// absence of an arbitrary string, but we can assert there is no
	// " " + <label> suffix when the label was never set. The Bubbles
	// spinner.Line first frame is "|"; the content should not end with
	// a space-separated trailing word (no label).
	if strings.Contains(v.Content, "loading") {
		t.Errorf("default spinner content unexpectedly contains a label: %q", v.Content)
	}
}

func TestNewSpinner_WithOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []widget.SpinnerOption
		wantSubs []string
	}{
		{
			name:     "label only",
			opts:     []widget.SpinnerOption{widget.WithSpinnerLabel("loading...")},
			wantSubs: []string{"loading..."},
		},
		{
			name: "style + label",
			opts: []widget.SpinnerOption{
				widget.WithSpinnerLabel("loading..."),
				widget.WithSpinnerStyle(spinner.Dot),
			},
			wantSubs: []string{"loading..."},
		},
		{
			name:     "style only",
			opts:     []widget.SpinnerOption{widget.WithSpinnerStyle(spinner.MiniDot)},
			wantSubs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := widget.NewSpinner(tt.opts...)
			content := s.View().Content
			if content == "" {
				t.Fatal("View().Content is empty")
			}
			for _, sub := range tt.wantSubs {
				if !strings.Contains(content, sub) {
					t.Errorf("View().Content = %q, want substring %q", content, sub)
				}
			}
		})
	}
}

func TestSpinner_SetLabel(t *testing.T) {
	s := widget.NewSpinner(widget.WithSpinnerLabel("old"))
	before := s.View().Content
	if !strings.Contains(before, "old") {
		t.Fatalf("initial View().Content = %q, want to contain %q", before, "old")
	}

	s.SetLabel("new")
	after := s.View().Content
	if !strings.Contains(after, "new") {
		t.Errorf("after SetLabel, View().Content = %q, want to contain %q", after, "new")
	}
	if strings.Contains(after, "old") {
		t.Errorf("after SetLabel, View().Content = %q, still contains previous label %q", after, "old")
	}
}

func TestSpinner_SetStyle(t *testing.T) {
	s := widget.NewSpinner()
	// Confirm render succeeds before and after style change. Bubbles'
	// glyph-level output is deliberately not asserted here.
	if s.View().Content == "" {
		t.Fatal("View().Content empty before SetStyle")
	}
	s.SetStyle(spinner.Dot)
	if s.View().Content == "" {
		t.Fatal("View().Content empty after SetStyle")
	}
}

func TestSpinner_UpdateTick_Focused_ReturnsCmd(t *testing.T) {
	s := widget.NewSpinner()
	tick := tickFromInit(t, s)

	m, cmd := s.Update(tick)
	got := asSpinner(t, m)
	if got != s {
		t.Errorf("Update returned different *Spinner; want same receiver")
	}
	if cmd == nil {
		t.Error("Update(TickMsg) while focused returned nil cmd; expected scheduled next tick")
	}
}

func TestSpinner_Update_UnrelatedMessage_NoOp(t *testing.T) {
	type customMsg struct{}

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"struct message", customMsg{}},
		{"string message", "hello"},
		{"nil message", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := widget.NewSpinner(widget.WithSpinnerLabel("stable"))
			before := s.View().Content

			// Must not panic.
			m, _ := s.Update(tt.msg)
			got := asSpinner(t, m)
			after := got.View().Content

			if before != after {
				t.Errorf("unrelated message mutated View: before=%q after=%q", before, after)
			}
			if !strings.Contains(after, "stable") {
				t.Errorf("label lost after unrelated message; View=%q", after)
			}
		})
	}
}

func TestSpinner_View_NoLabel_NoTrailingSpaceLabel(t *testing.T) {
	s := widget.NewSpinner()
	content := s.View().Content
	if content == "" {
		t.Fatal("View().Content empty")
	}
	// With no label the spinner should not emit the "<frame> <label>" form,
	// so the content should not end with a space+word pattern. The single
	// space used as a separator is only introduced when a label is set.
	if strings.HasSuffix(content, " ") {
		t.Errorf("View().Content ends with space; suggests label separator present: %q", content)
	}
}

func TestSpinner_View_WithLabel_ContainsSpaceAndLabel(t *testing.T) {
	s := widget.NewSpinner(widget.WithSpinnerLabel("hi"))
	content := s.View().Content
	if !strings.Contains(content, "hi") {
		t.Errorf("View().Content = %q, want to contain %q", content, "hi")
	}
	if !strings.Contains(content, " ") {
		t.Errorf("View().Content = %q, want to contain a space separator", content)
	}
}

func TestSpinner_Blur_DropsTickMsg_NoCmd(t *testing.T) {
	s := widget.NewSpinner()
	tick := tickFromInit(t, s)

	s.Blur()
	m, cmd := s.Update(tick)
	asSpinner(t, m)
	if cmd != nil {
		t.Errorf("blurred Update(TickMsg) returned non-nil cmd; expected nil")
	}
}

func TestSpinner_Focus_RestoresTickScheduling(t *testing.T) {
	s := widget.NewSpinner()
	tick := tickFromInit(t, s)

	s.Blur()
	_, cmd := s.Update(tick)
	if cmd != nil {
		t.Fatalf("precondition: blurred Update(TickMsg) should return nil cmd, got non-nil")
	}

	s.Focus()
	// Use a freshly generated tick carrying up-to-date tag; the inner
	// spinner's tag is still 0 since no Update incremented it.
	tick = tickFromInit(t, s)
	_, cmd = s.Update(tick)
	if cmd == nil {
		t.Errorf("focused Update(TickMsg) after Focus returned nil cmd; expected non-nil")
	}
}

func TestSpinner_SetSize_NoOp(t *testing.T) {
	s := widget.NewSpinner(widget.WithSpinnerLabel("sized"))
	before := s.View().Content

	s.SetSize(100, 50)
	after := s.View().Content

	if before != after {
		t.Errorf("SetSize mutated View; before=%q after=%q", before, after)
	}
}

func TestSpinner_Lifecycle_InitUpdateView(t *testing.T) {
	s := widget.NewSpinner(widget.WithSpinnerLabel("boot"))

	cmd := s.Init()
	if cmd == nil {
		t.Fatal("Init returned nil cmd")
	}
	msg := cmd()
	if msg == nil {
		t.Fatal("Init cmd produced nil msg")
	}

	m, nextCmd := s.Update(msg)
	asSpinner(t, m)
	// Focused by default → cmd should be non-nil.
	if nextCmd == nil {
		t.Error("Update after Init returned nil cmd; expected inner tick scheduler")
	}

	content := s.View().Content
	if content == "" {
		t.Fatal("View().Content empty after lifecycle")
	}
	if !strings.Contains(content, "boot") {
		t.Errorf("View().Content = %q, want to contain %q", content, "boot")
	}
}

// TestSpinner_Golden renders three deterministic variants and snapshots
// them via cupaloy. Freshly constructed spinners have frame index 0, so
// the rendered output is stable across runs.
func TestSpinner_Golden(t *testing.T) {
	tests := []struct {
		name string
		opts []widget.SpinnerOption
	}{
		{
			name: "default_no_label",
			opts: nil,
		},
		{
			name: "default_with_label",
			opts: []widget.SpinnerOption{widget.WithSpinnerLabel("Loading…")},
		},
		{
			name: "dot_with_label",
			opts: []widget.SpinnerOption{
				widget.WithSpinnerStyle(spinner.Dot),
				widget.WithSpinnerLabel("Working"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := widget.NewSpinner(tt.opts...)
			cupaloy.SnapshotT(t, s.View().Content)
		})
	}
}

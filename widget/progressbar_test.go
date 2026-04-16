package widget_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/bradleyjkemp/cupaloy/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
	"github.com/RenseiAI/tui-components/widget"
)

// twoPointClock returns a clock function that returns start on the first
// call (recorded as etaStart by SetPercent) and later on every subsequent
// call. View calls nowFn exactly once, so this lets the test pin the
// elapsed wall time precisely.
func twoPointClock(start, later time.Time) func() time.Time {
	calls := 0
	return func() time.Time {
		calls++
		if calls == 1 {
			return start
		}
		return later
	}
}

// Runtime interface assertion (complements the compile-time assertion in
// progressbar.go).
var _ component.Component = (*widget.Progressbar)(nil)

func TestNewProgressbar_Defaults(t *testing.T) {
	p := widget.NewProgressbar()
	if p == nil {
		t.Fatal("NewProgressbar() returned nil")
	}
	if got := p.Percent(); got != 0 {
		t.Errorf("default Percent() = %v, want 0", got)
	}
	if v := p.View(); v.Content == "" {
		t.Error("default View().Content is empty; want a rendered empty bar")
	}
}

func TestProgressbar_SetPercent_Clamps(t *testing.T) {
	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{"zero", 0, 0},
		{"mid", 0.5, 0.5},
		{"one", 1, 1},
		{"negative", -0.25, 0},
		{"above one", 1.5, 1},
		{"large negative", -1e9, 0},
		{"large positive", 1e9, 1},
		{"nan", math.NaN(), 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := widget.NewProgressbar()
			cmd := p.SetPercent(tt.in)
			if cmd != nil {
				t.Errorf("SetPercent(%v) returned non-nil cmd; deterministic mode reserves it for future use", tt.in)
			}
			if got := p.Percent(); got != tt.want {
				t.Errorf("Percent() after SetPercent(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestProgressbar_IncrBy(t *testing.T) {
	tests := []struct {
		name  string
		start float64
		incr  float64
		want  float64
	}{
		{"from zero", 0, 0.25, 0.25},
		{"accumulate", 0.25, 0.5, 0.75},
		{"clamps high", 0.9, 0.5, 1},
		{"clamps low", 0.1, -0.5, 0},
		{"no-op", 0.5, 0, 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := widget.NewProgressbar()
			p.SetPercent(tt.start)
			p.IncrBy(tt.incr)
			if got := p.Percent(); got != tt.want {
				t.Errorf("Percent() after start=%v IncrBy(%v) = %v, want %v",
					tt.start, tt.incr, got, tt.want)
			}
		})
	}
}

func TestProgressbar_SetSize_PropagatesWidth(t *testing.T) {
	p := widget.NewProgressbar(widget.WithProgressbarWidth(40))
	p.SetPercent(0.5)
	wide := p.View().Content

	p.SetSize(80, 1)
	wider := p.View().Content

	if lipgloss.Width(wide) >= lipgloss.Width(wider) {
		t.Errorf("expected wider rendering after SetSize(80,1); width40=%d width80=%d",
			lipgloss.Width(wide), lipgloss.Width(wider))
	}
}

func TestProgressbar_SetSize_NegativeIsClamped(t *testing.T) {
	p := widget.NewProgressbar()
	p.SetPercent(0.5)
	p.SetSize(-5, 1)
	if got := p.View().Content; got != "" {
		t.Errorf("View().Content with width<=0 = %q, want empty string", got)
	}
}

func TestProgressbar_View_ZeroWidth_Empty(t *testing.T) {
	p := widget.NewProgressbar(widget.WithProgressbarWidth(0))
	p.SetPercent(0.5)
	if got := p.View().Content; got != "" {
		t.Errorf("View().Content at width 0 = %q, want empty", got)
	}
}

func TestProgressbar_FocusBlur_NoOp(t *testing.T) {
	p := widget.NewProgressbar(widget.WithProgressbarWidth(40))
	p.SetPercent(0.5)
	before := p.View().Content

	p.Blur()
	if got := p.View().Content; got != before {
		t.Errorf("Blur changed View; before=%q after=%q", before, got)
	}
	p.Focus()
	if got := p.View().Content; got != before {
		t.Errorf("Focus changed View; before=%q after=%q", before, got)
	}
}

func TestProgressbar_Init_ReturnsNil(t *testing.T) {
	p := widget.NewProgressbar()
	if cmd := p.Init(); cmd != nil {
		t.Errorf("Init() = non-nil cmd; deterministic-mode bar should return nil")
	}
}

func TestProgressbar_Init_IndeterminateReturnsTickCmd(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(20),
		widget.WithProgressbarIndeterminate(true),
	)
	if cmd := p.Init(); cmd == nil {
		t.Errorf("Init() in indeterminate mode = nil cmd; want a tick cmd")
	}
}

func TestProgressbar_SetIndeterminate_ModeTransitions(t *testing.T) {
	tests := []struct {
		name      string
		startInd  bool
		toggleTo  bool
		wantCmd   bool
		startsAnd bool // expect render to look like sweep after transition
	}{
		{"det → ind", false, true, true, true},
		{"ind → det", true, false, false, false},
		{"det → det no-op", false, false, false, false},
		{"ind → ind no-op", true, true, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := []widget.ProgressbarOption{widget.WithProgressbarWidth(20)}
			if tt.startInd {
				opts = append(opts, widget.WithProgressbarIndeterminate(true))
			}
			p := widget.NewProgressbar(opts...)
			p.SetPercent(0.5) // give the deterministic-side state something to render
			cmd := p.SetIndeterminate(tt.toggleTo)
			if tt.wantCmd && cmd == nil {
				t.Errorf("SetIndeterminate(%v) returned nil cmd; want a tick cmd", tt.toggleTo)
			}
			if !tt.wantCmd && cmd != nil {
				t.Errorf("SetIndeterminate(%v) returned non-nil cmd; want nil", tt.toggleTo)
			}
			content := p.View().Content
			if content == "" {
				t.Fatal("View().Content empty after SetIndeterminate")
			}
		})
	}
}

func TestProgressbar_Indeterminate_ZeroWidth_Empty(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(0),
		widget.WithProgressbarIndeterminate(true),
	)
	if got := p.View().Content; got != "" {
		t.Errorf("indeterminate View().Content at width 0 = %q, want empty", got)
	}
}

func TestProgressbar_Indeterminate_SuppressesPercentAndETA(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(40),
		widget.WithProgressbarShowPercent(true),
		widget.WithProgressbarShowETA(true),
		widget.WithProgressbarClock(twoPointClock(start, start.Add(5*time.Second))),
		widget.WithProgressbarIndeterminate(true),
	)
	p.SetPercent(0.5)
	got := p.View().Content
	if strings.Contains(got, "%") {
		t.Errorf("indeterminate render = %q, expected no percent segment", got)
	}
	if strings.Contains(got, "~") {
		t.Errorf("indeterminate render = %q, expected no ETA segment", got)
	}
}

func TestProgressbar_Indeterminate_LabelStillRendered(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(40),
		widget.WithProgressbarLabel("Streaming"),
		widget.WithProgressbarIndeterminate(true),
	)
	got := p.View().Content
	if !strings.Contains(got, "Streaming") {
		t.Errorf("indeterminate render = %q, want to contain label %q", got, "Streaming")
	}
}

func TestProgressbar_Update_ReturnsReceiver(t *testing.T) {
	type customMsg struct{}
	p := widget.NewProgressbar(widget.WithProgressbarWidth(40))
	p.SetPercent(0.42)

	m, _ := p.Update(customMsg{})
	got, ok := m.(*widget.Progressbar)
	if !ok {
		t.Fatalf("Update returned %T, want *widget.Progressbar", m)
	}
	if got != p {
		t.Errorf("Update returned different *Progressbar; want same receiver")
	}
	if got.Percent() != 0.42 {
		t.Errorf("Percent() after Update = %v, want 0.42 (Update must not mutate)", got.Percent())
	}
}

func TestProgressbar_WithGradient_RendersDistinctly(t *testing.T) {
	defaultBar := widget.NewProgressbar(widget.WithProgressbarWidth(40))
	defaultBar.SetPercent(0.5)
	def := defaultBar.View().Content

	customBar := widget.NewProgressbar(
		widget.WithProgressbarWidth(40),
		widget.WithProgressbarGradient(theme.StatusSuccess, theme.StatusWarning),
	)
	customBar.SetPercent(0.5)
	custom := customBar.View().Content

	if def == custom {
		t.Errorf("custom gradient produced identical output to default; want different ANSI escapes")
	}
	if !strings.Contains(def, "\x1b[") {
		t.Errorf("expected ANSI escapes in default render, got %q", def)
	}
}

// TestProgressbar_Golden snapshots a 50%-filled bar at width 40 with the
// default theme gradient. Output is stable because View renders ViewAs at
// the most recent target percent (no spring animation).
func TestProgressbar_Golden(t *testing.T) {
	p := widget.NewProgressbar(widget.WithProgressbarWidth(40))
	p.SetPercent(0.5)
	cupaloy.SnapshotT(t, p.View().Content)
}

func TestProgressbar_WithLabel_RendersLabel(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(60),
		widget.WithProgressbarLabel("Uploading"),
	)
	p.SetPercent(0.5)
	got := p.View().Content
	if !strings.Contains(got, "Uploading") {
		t.Errorf("View().Content = %q, want to contain label %q", got, "Uploading")
	}
}

func TestProgressbar_WithShowPercent_RendersFormatted(t *testing.T) {
	tests := []struct {
		name    string
		percent float64
		want    string
	}{
		{"zero", 0, "  0.0%"},
		{"half", 0.5, " 50.0%"},
		{"full", 1, "100.0%"},
		{"odd", 0.234, " 23.4%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := widget.NewProgressbar(
				widget.WithProgressbarWidth(60),
				widget.WithProgressbarShowPercent(true),
			)
			p.SetPercent(tt.percent)
			got := p.View().Content
			if !strings.Contains(got, tt.want) {
				t.Errorf("View().Content = %q, want to contain %q", got, tt.want)
			}
		})
	}
}

func TestProgressbar_WithShowETA_PreStart_Empty(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(60),
		widget.WithProgressbarShowETA(true),
	)
	// No SetPercent yet → no etaStart recorded.
	got := p.View().Content
	if strings.Contains(got, "~") {
		t.Errorf("View().Content = %q, did not expect ETA marker before SetPercent", got)
	}
}

func TestProgressbar_WithShowETA_MidProgress_RendersDuration(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	later := start.Add(10 * time.Second) // 10s elapsed at 25% → 30s remaining
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(60),
		widget.WithProgressbarShowETA(true),
		widget.WithProgressbarClock(twoPointClock(start, later)),
	)
	p.SetPercent(0.25)
	got := p.View().Content
	if !strings.Contains(got, "~30s") {
		t.Errorf("View().Content = %q, want to contain %q", got, "~30s")
	}
}

func TestProgressbar_WithShowETA_AtComplete_Empty(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(60),
		widget.WithProgressbarShowETA(true),
		widget.WithProgressbarClock(twoPointClock(start, start.Add(10*time.Second))),
	)
	p.SetPercent(1)
	got := p.View().Content
	if strings.Contains(got, "~") {
		t.Errorf("View().Content = %q, did not expect ETA marker at 100%%", got)
	}
}

func TestProgressbar_AllSegments_Layout(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(60),
		widget.WithProgressbarLabel("Uploading"),
		widget.WithProgressbarShowPercent(true),
		widget.WithProgressbarShowETA(true),
		widget.WithProgressbarClock(twoPointClock(start, start.Add(5*time.Second))),
	)
	p.SetPercent(0.5)
	got := p.View().Content
	for _, want := range []string{"Uploading", " 50.0%", "~5s"} {
		if !strings.Contains(got, want) {
			t.Errorf("View().Content = %q, want to contain %q", got, want)
		}
	}
	// Total visible width should not exceed configured width.
	if w := lipgloss.Width(got); w > 60 {
		t.Errorf("rendered width = %d, want <= 60", w)
	}
}

func TestProgressbar_TightWidth_DropsLabelFirst(t *testing.T) {
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(20), // 8 (bar min) + 6 (percent) + 1 (sep) leaves no room for label
		widget.WithProgressbarLabel("Uploading"),
		widget.WithProgressbarShowPercent(true),
	)
	p.SetPercent(0.5)
	got := p.View().Content
	if strings.Contains(got, "Uploading") {
		t.Errorf("View().Content = %q, should have dropped label at tight width", got)
	}
	if !strings.Contains(got, " 50.0%") {
		t.Errorf("View().Content = %q, expected percent to be preserved over label", got)
	}
}

func TestProgressbar_WithProgressbarClock_Nil_RestoresDefault(t *testing.T) {
	// Should not panic and the bar still renders.
	p := widget.NewProgressbar(
		widget.WithProgressbarWidth(40),
		widget.WithProgressbarShowETA(true),
		widget.WithProgressbarClock(nil),
	)
	p.SetPercent(0.5)
	if v := p.View().Content; v == "" {
		t.Error("View().Content empty after passing nil clock")
	}
}

// TestProgressbar_DecorationsGolden snapshots the four decoration
// combinations called for in TC-005b at width 60. Each variant uses a
// pinned clock so the ETA segment is deterministic when present.
func TestProgressbar_DecorationsGolden(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)
	mid := start.Add(5 * time.Second) // at 50% percent → 5s remaining

	tests := []struct {
		name string
		opts []widget.ProgressbarOption
	}{
		{
			name: "label_only",
			opts: []widget.ProgressbarOption{
				widget.WithProgressbarWidth(60),
				widget.WithProgressbarLabel("Uploading"),
			},
		},
		{
			name: "percent_only",
			opts: []widget.ProgressbarOption{
				widget.WithProgressbarWidth(60),
				widget.WithProgressbarShowPercent(true),
			},
		},
		{
			name: "eta_only",
			opts: []widget.ProgressbarOption{
				widget.WithProgressbarWidth(60),
				widget.WithProgressbarShowETA(true),
				widget.WithProgressbarClock(twoPointClock(start, mid)),
			},
		},
		{
			name: "all_three",
			opts: []widget.ProgressbarOption{
				widget.WithProgressbarWidth(60),
				widget.WithProgressbarLabel("Uploading"),
				widget.WithProgressbarShowPercent(true),
				widget.WithProgressbarShowETA(true),
				widget.WithProgressbarClock(twoPointClock(start, mid)),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := widget.NewProgressbar(tt.opts...)
			p.SetPercent(0.5)
			cupaloy.SnapshotT(t, p.View().Content)
		})
	}
}

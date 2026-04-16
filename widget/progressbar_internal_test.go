package widget

import (
	"strings"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
)

// TestEtaString covers the table of edge cases the helper must handle:
// pre-start (zero start time), mid-progress with a known rate, near-done
// progress that should still report at least 1s, and the 100% terminal
// state which suppresses the segment.
func TestEtaString(t *testing.T) {
	start := time.Date(2026, 4, 16, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		start   time.Time
		now     time.Time
		percent float64
		want    string
	}{
		{
			name:    "pre-start zero start",
			start:   time.Time{},
			now:     start.Add(10 * time.Second),
			percent: 0.5,
			want:    "",
		},
		{
			name:    "non-positive percent",
			start:   start,
			now:     start.Add(10 * time.Second),
			percent: 0,
			want:    "",
		},
		{
			name:    "no elapsed time",
			start:   start,
			now:     start,
			percent: 0.5,
			want:    "",
		},
		{
			name:    "elapsed in past",
			start:   start,
			now:     start.Add(-1 * time.Second),
			percent: 0.5,
			want:    "",
		},
		{
			name:    "mid-progress 25%",
			start:   start,
			now:     start.Add(10 * time.Second),
			percent: 0.25, // 30s remaining
			want:    "~30s",
		},
		{
			name:    "near-done one-second remainder",
			start:   start,
			now:     start.Add(1 * time.Second),
			percent: 0.5, // 1s remaining (exact: 0.5/0.5 = 1.0)
			want:    "~1s",
		},
		{
			name:    "complete suppresses",
			start:   start,
			now:     start.Add(10 * time.Second),
			percent: 1,
			want:    "",
		},
		{
			name:    "minutes range",
			start:   start,
			now:     start.Add(60 * time.Second),
			percent: 0.25, // 180s -> "3m"
			want:    "~3m",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := etaString(tt.start, tt.now, tt.percent)
			if got != tt.want {
				t.Errorf("etaString(start=%v, now=%v, percent=%v) = %q, want %q",
					tt.start, tt.now, tt.percent, got, tt.want)
			}
		})
	}
}

// TestProgressbar_GenerationInvalidatesStaleTicks verifies that tickMsgs
// scheduled before SetIndeterminate flips the mode are dropped — the
// generation counter prevents a queued frame from advancing the sweep
// after the bar has paused.
func TestProgressbar_GenerationInvalidatesStaleTicks(t *testing.T) {
	p := NewProgressbar(
		WithProgressbarWidth(20),
		WithProgressbarIndeterminate(true),
	)
	staleTick := tickMsg{gen: p.generation}

	// Pause the animation. Generation advances; the queued tick is now stale.
	if cmd := p.SetIndeterminate(false); cmd != nil {
		t.Errorf("SetIndeterminate(false) returned non-nil cmd; expected nil")
	}

	prevStep := p.sweepStep
	_, nextCmd := p.Update(staleTick)
	if nextCmd != nil {
		t.Errorf("stale tickMsg should not produce a follow-up cmd; got %v", nextCmd)
	}
	if p.sweepStep != prevStep {
		t.Errorf("stale tickMsg advanced sweepStep: before=%d after=%d", prevStep, p.sweepStep)
	}
}

// TestProgressbar_DeterministicTickDropped verifies that a tickMsg
// arriving while the bar is in deterministic mode is silently ignored.
func TestProgressbar_DeterministicTickDropped(t *testing.T) {
	p := NewProgressbar(WithProgressbarWidth(20)) // deterministic
	prev := p.sweepStep
	_, cmd := p.Update(tickMsg{gen: p.generation})
	if cmd != nil {
		t.Errorf("tickMsg in deterministic mode should not return a cmd")
	}
	if p.sweepStep != prev {
		t.Errorf("tickMsg in deterministic mode mutated sweepStep")
	}
}

// TestProgressbar_LiveTickAdvancesAndReschedules verifies the happy path
// of indeterminate-mode ticking: a current-generation tick advances the
// sweep and returns a follow-up cmd.
func TestProgressbar_LiveTickAdvancesAndReschedules(t *testing.T) {
	p := NewProgressbar(
		WithProgressbarWidth(20),
		WithProgressbarIndeterminate(true),
	)
	prev := p.sweepStep
	_, cmd := p.Update(tickMsg{gen: p.generation})
	if cmd == nil {
		t.Fatal("live tickMsg should return a follow-up cmd")
	}
	if p.sweepStep != prev+1 {
		t.Errorf("sweepStep = %d, want %d", p.sweepStep, prev+1)
	}
}

// TestProgressbar_SetIndeterminate_NoOpReturnsNil ensures SetIndeterminate
// to the current mode does not increment the generation or emit a tick.
func TestProgressbar_SetIndeterminate_NoOpReturnsNil(t *testing.T) {
	p := NewProgressbar(WithProgressbarWidth(20))
	gen := p.generation
	if cmd := p.SetIndeterminate(false); cmd != nil {
		t.Errorf("no-op SetIndeterminate returned cmd; want nil")
	}
	if p.generation != gen {
		t.Errorf("no-op SetIndeterminate advanced generation: was %d now %d", gen, p.generation)
	}
}

// TestProgressbar_IndeterminateGolden snapshots three deterministic frames
// of the sweep at width 40 plus a wrap-around frame, by setting the
// internal frame counter directly so output is stable.
func TestProgressbar_IndeterminateGolden(t *testing.T) {
	tests := []struct {
		name string
		step int
	}{
		{name: "frame_0", step: 0},
		{name: "frame_5", step: 5},
		{name: "frame_10", step: 10},
		{name: "wrap", step: 35}, // width=40, sweep≈8 → wraps
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProgressbar(
				WithProgressbarWidth(40),
				WithProgressbarIndeterminate(true),
			)
			p.sweepStep = tt.step
			content := p.View().Content
			if content == "" {
				t.Fatal("indeterminate render produced empty content at positive width")
			}
			// Sanity check: every frame must include at least one lit cell.
			if !strings.ContainsRune(content, indeterminateLitRune) {
				t.Errorf("frame %d missing lit rune; content=%q", tt.step, content)
			}
			cupaloy.SnapshotT(t, content)
		})
	}
}

// TestTruncateLabel exercises the label-truncation helper across the
// boundary cases the layout logic relies on.
func TestTruncateLabel(t *testing.T) {
	tests := []struct {
		name string
		in   string
		max  int
		want string
	}{
		{"zero budget", "Uploading", 0, ""},
		{"negative budget", "Uploading", -3, ""},
		{"fits exactly", "Up", 2, "Up"},
		{"fits with room", "Up", 5, "Up"},
		{"truncate at 5", "Uploading", 5, "Uplo…"},
		{"truncate at 4", "Uploading", 4, "Upl…"},
		{"truncate to ellipsis", "Uploading", 1, "…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateLabel(tt.in, tt.max)
			if got != tt.want {
				t.Errorf("truncateLabel(%q, %d) = %q, want %q", tt.in, tt.max, got, tt.want)
			}
		})
	}
}

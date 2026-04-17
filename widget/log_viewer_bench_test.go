package widget

import (
	"strconv"
	"strings"
	"testing"
)

// benchLines pre-constructs a fixed-size batch of realistic log lines
// so the benchmark loop doesn't measure test-data allocation. Lines
// include a small ANSI colour run to exercise the SGR parser on the
// hot path.
func benchLines(n int) []string {
	lines := make([]string, n)
	for i := 0; i < n; i++ {
		lines[i] = "\x1b[32m2026-04-17 10:00:00\x1b[0m INFO event " +
			strconv.Itoa(i) + " " + strings.Repeat("payload ", 3)
	}
	return lines
}

// BenchmarkLogViewer_Append measures steady-state allocation cost of
// appending a single line once the ring buffer is warm (pre-filled to
// maxLines so the hot path is the ring-rotate, not the initial grow).
//
// Sub-benchmarks cover both wrap=on (the default, with per-line
// lipgloss width clamping) and wrap=off (no width clamp).
func BenchmarkLogViewer_Append(b *testing.B) {
	const maxLines = 10_000
	lines := benchLines(maxLines + 1) // one extra to append per-iter

	b.Run("WrapOn", func(b *testing.B) {
		m := New(WithMaxLines(maxLines), WithWrap(true))
		m.SetSize(80, 24)
		// Warm the ring: fill to capacity so every Append is a rotate.
		for i := 0; i < maxLines; i++ {
			m.appendOne(lines[i])
		}
		m.refresh()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Append(lines[maxLines])
		}
	})

	b.Run("WrapOff", func(b *testing.B) {
		m := New(WithMaxLines(maxLines), WithWrap(false))
		m.SetSize(80, 24)
		for i := 0; i < maxLines; i++ {
			m.appendOne(lines[i])
		}
		m.refresh()

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m.Append(lines[maxLines])
		}
	})
}

// TestLogViewer_AppendRingAllocs asserts that appendOne (the ring-slot
// write itself, excluding the refresh/render pipeline) performs zero
// heap allocations in steady state. The ring is pre-allocated to
// maxLines capacity in [New], so the rotating copy and final
// assignment must not grow the backing array.
//
// This is the canonical "zero allocations per append beyond the ring
// slot" assertion from the REN-123 acceptance criteria.
func TestLogViewer_AppendRingAllocs(t *testing.T) {
	const maxLines = 1000
	m := New(WithMaxLines(maxLines))
	m.SetSize(80, 24)

	lines := benchLines(maxLines + 1)
	for i := 0; i < maxLines; i++ {
		m.appendOne(lines[i])
	}

	allocs := testing.AllocsPerRun(100, func() {
		m.appendOne(lines[maxLines])
	})
	if allocs > 0 {
		t.Errorf("appendOne steady-state allocs/op: want 0, got %.2f", allocs)
	}
}

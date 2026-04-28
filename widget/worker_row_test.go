package widget

import (
	"strings"
	"testing"
)

func TestWorkerRow_NoColor(t *testing.T) {
	t.Parallel()
	row := NewWorkerRow(
		WithWorkerID("w-1"),
		WithWorkerStatus(WorkerStatusBusy),
		WithWorkerRegion("iad1"),
		WithWorkerLoadFraction(0.75),
		WithWorkerBillingModel("active-cpu"),
		WithWorkerNoColor(true),
	)
	got := row.ViewString()
	if !strings.Contains(got, "busy") {
		t.Errorf("ViewString() missing 'busy': %q", got)
	}
	if !strings.Contains(got, "iad1") {
		t.Errorf("ViewString() missing region: %q", got)
	}
	if !strings.Contains(got, "active-cpu") {
		t.Errorf("ViewString() missing billing model: %q", got)
	}
}

func TestWorkerRow_LoadFractionClamped(t *testing.T) {
	t.Parallel()
	row := NewWorkerRow(
		WithWorkerLoadFraction(2.5), // should clamp to 1.0
		WithWorkerNoColor(true),
	)
	got := row.ViewString()
	if !strings.Contains(got, "100%") {
		t.Errorf("ViewString() should show 100%%: %q", got)
	}
}

func TestWorkerRow_AccessibleLabel(t *testing.T) {
	t.Parallel()
	row := NewWorkerRow(
		WithWorkerID("w-1"),
		WithWorkerStatus(WorkerStatusIdle),
		WithWorkerRegion("sfo3"),
		WithWorkerLoadFraction(0),
	)
	label := row.AccessibleLabel()
	if !strings.Contains(label, "w-1") {
		t.Errorf("AccessibleLabel() missing worker ID: %q", label)
	}
}

func TestRenderLoadBar(t *testing.T) {
	t.Parallel()
	bar := renderLoadBar(0.5, 8)
	if bar != "████░░░░" {
		t.Errorf("renderLoadBar(0.5, 8) = %q, want ████░░░░", bar)
	}
	bar0 := renderLoadBar(0, 4)
	if bar0 != "░░░░" {
		t.Errorf("renderLoadBar(0, 4) = %q, want ░░░░", bar0)
	}
	bar1 := renderLoadBar(1, 4)
	if bar1 != "████" {
		t.Errorf("renderLoadBar(1, 4) = %q, want ████", bar1)
	}
}

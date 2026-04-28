package widget

import (
	"strings"
	"testing"
)

func TestFleetGrid_Empty(t *testing.T) {
	t.Parallel()
	g := NewFleetGrid(WithFleetNoColor(true))
	got := g.ViewString()
	if got != "(no workers)" {
		t.Errorf("ViewString() = %q, want '(no workers)'", got)
	}
}

func TestFleetGrid_GroupsByMachine(t *testing.T) {
	t.Parallel()
	g := NewFleetGrid(
		WithFleetWorkers(
			FleetWorker{ID: "w-1", MachineGroup: "m1", Status: WorkerStatusBusy, Region: "iad1", LoadFraction: 0.8},
			FleetWorker{ID: "w-2", MachineGroup: "m2", Status: WorkerStatusIdle, Region: "sfo3", LoadFraction: 0},
		),
		WithFleetNoColor(true),
	)
	got := g.ViewString()
	if !strings.Contains(got, "m1") {
		t.Errorf("ViewString() missing machine group m1: %q", got)
	}
	if !strings.Contains(got, "m2") {
		t.Errorf("ViewString() missing machine group m2: %q", got)
	}
	if !strings.Contains(got, "iad1") {
		t.Errorf("ViewString() missing region iad1: %q", got)
	}
}

func TestFleetGrid_SetWorkers(t *testing.T) {
	t.Parallel()
	g := NewFleetGrid(WithFleetNoColor(true))
	g.SetWorkers([]FleetWorker{{ID: "w-3", MachineGroup: "m3"}})
	got := g.ViewString()
	if !strings.Contains(got, "m3") {
		t.Errorf("after SetWorkers, missing m3: %q", got)
	}
}

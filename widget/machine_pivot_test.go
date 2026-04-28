package widget

import (
	"strings"
	"testing"
)

func TestMachinePivot_Empty(t *testing.T) {
	t.Parallel()
	p := NewMachinePivot(WithMachinePivotNoColor(true))
	got := p.ViewString()
	if got != "(no machines)" {
		t.Errorf("ViewString() = %q, want '(no machines)'", got)
	}
}

func TestMachinePivot_Render(t *testing.T) {
	t.Parallel()
	p := NewMachinePivot(
		WithMachines(
			MachineSummary{ID: "mac-01", Workers: 4, ActiveWorkers: 2, Region: "iad1", Health: ProviderHealthReady},
			MachineSummary{ID: "mac-02", Workers: 2, ActiveWorkers: 0, Region: "sfo3", Health: ProviderHealthDegraded},
		),
		WithMachinePivotNoColor(true),
	)
	got := p.ViewString()
	if !strings.Contains(got, "mac-01") {
		t.Errorf("ViewString() missing mac-01: %q", got)
	}
	if !strings.Contains(got, "mac-02") {
		t.Errorf("ViewString() missing mac-02: %q", got)
	}
	if !strings.Contains(got, "iad1") {
		t.Errorf("ViewString() missing iad1: %q", got)
	}
}

func TestMachinePivot_SetMachines(t *testing.T) {
	t.Parallel()
	p := NewMachinePivot(WithMachinePivotNoColor(true))
	p.SetMachines([]MachineSummary{{ID: "m-new", Region: "fra1", Health: ProviderHealthReady}})
	got := p.ViewString()
	if !strings.Contains(got, "m-new") {
		t.Errorf("after SetMachines, missing m-new: %q", got)
	}
}

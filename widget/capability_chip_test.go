package widget

import (
	"strings"
	"testing"

	"github.com/RenseiAI/tui-components/theme"
)

func TestCapabilityChip_NoColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		value      string
		humanLabel string
		want       string
	}{
		{
			name:       "value and human label",
			value:      "active-cpu",
			humanLabel: "Billed for active CPU only",
			want:       "◆ active-cpu  Billed for active CPU only",
		},
		{
			name:  "value only",
			value: "wall-clock",
			want:  "◆ wall-clock",
		},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			chip := NewCapabilityChip(
				WithCapabilityValue(tc.value),
				WithCapabilityHumanLabel(tc.humanLabel),
				WithCapabilityNoColor(true),
			)
			got := chip.ViewString()
			if got != tc.want {
				t.Errorf("ViewString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCapabilityChip_WithTheme(t *testing.T) {
	t.Parallel()
	chip := NewCapabilityChip(
		WithCapabilityValue("dial-in"),
		WithCapabilityHumanLabel("Dial-in transport"),
		WithCapabilityTheme(theme.DefaultTheme()),
	)
	got := chip.ViewString()
	// Color mode should contain the value and label
	if !strings.Contains(got, "dial-in") {
		t.Errorf("ViewString() missing value: %q", got)
	}
	if !strings.Contains(got, "Dial-in transport") {
		t.Errorf("ViewString() missing humanLabel: %q", got)
	}
}

func TestCapabilityChip_AccessibleLabel(t *testing.T) {
	t.Parallel()
	chip := NewCapabilityChip(
		WithCapabilityValue("wall-clock"),
		WithCapabilityHumanLabel("Billed by wall-clock time"),
	)
	want := "wall-clock: Billed by wall-clock time"
	if got := chip.AccessibleLabel(); got != want {
		t.Errorf("AccessibleLabel() = %q, want %q", got, want)
	}
}

func TestCapabilityChip_AccessibleLabelExplicit(t *testing.T) {
	t.Parallel()
	chip := NewCapabilityChip(
		WithCapabilityValue("x"),
		WithCapabilityAccessibleLabel("custom label"),
	)
	if got := chip.AccessibleLabel(); got != "custom label" {
		t.Errorf("AccessibleLabel() = %q, want %q", got, "custom label")
	}
}

func TestCapabilityChip_SetTheme(t *testing.T) {
	t.Parallel()
	chip := NewCapabilityChip(WithCapabilityValue("test"))
	chip.SetTheme(theme.DarkTheme())
	// Should not panic; just verify it runs
	_ = chip.ViewString()
}

func TestCapabilityChip_ComponentInterface(t *testing.T) {
	t.Parallel()
	chip := NewCapabilityChip(WithCapabilityValue("x"))
	chip.SetSize(80, 1)
	chip.Focus()
	chip.Blur()
	cmd := chip.Init()
	if cmd != nil {
		t.Error("Init() should return nil")
	}
	model, cmd2 := chip.Update(nil)
	if model != chip {
		t.Error("Update() should return same model")
	}
	if cmd2 != nil {
		t.Error("Update() should return nil cmd")
	}
}

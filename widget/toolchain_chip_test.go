package widget

import (
	"strings"
	"testing"
)

func TestToolchainSpec_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		spec ToolchainSpec
		want string
	}{
		{ToolchainSpec{"java", "17"}, "java=17"},
		{ToolchainSpec{"node", "20.x"}, "node=20.x"},
		{ToolchainSpec{"go", ""}, "go"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := tc.spec.String()
			if got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToolchainChip_NoColor(t *testing.T) {
	t.Parallel()
	chip := NewToolchainChip(
		WithToolchainSpecs(ToolchainSpec{"java", "17"}, ToolchainSpec{"node", "20.x"}),
		WithToolchainNoColor(true),
	)
	got := chip.ViewString()
	if !strings.Contains(got, "java=17") {
		t.Errorf("ViewString() missing java=17: %q", got)
	}
	if !strings.Contains(got, "node=20.x") {
		t.Errorf("ViewString() missing node=20.x: %q", got)
	}
}

func TestToolchainChip_Empty(t *testing.T) {
	t.Parallel()
	chip := NewToolchainChip(WithToolchainNoColor(true))
	got := chip.ViewString()
	if !strings.Contains(got, "(no toolchain)") {
		t.Errorf("ViewString() = %q, want '(no toolchain)'", got)
	}
}

func TestToolchainChip_AccessibleLabel(t *testing.T) {
	t.Parallel()
	chip := NewToolchainChip(
		WithToolchainSpecs(ToolchainSpec{"java", "17"}),
	)
	label := chip.AccessibleLabel()
	if !strings.Contains(label, "java=17") {
		t.Errorf("AccessibleLabel() = %q missing java=17", label)
	}
}

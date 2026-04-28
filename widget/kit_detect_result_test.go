package widget

import (
	"strings"
	"testing"
)

func TestKitDetectResult_NoKits(t *testing.T) {
	t.Parallel()
	r := NewKitDetectResult(WithKitDetectNoColor(true))
	got := r.ViewString()
	if got != "No kits detected" {
		t.Errorf("ViewString() = %q, want 'No kits detected'", got)
	}
}

func TestKitDetectResult_WithMatches(t *testing.T) {
	t.Parallel()
	r := NewKitDetectResult(
		WithKitMatches(
			KitMatch{Name: "spring-java", Version: "v1.2.0", Order: 1},
			KitMatch{Name: "nextjs", Version: "v0.9.1", Order: 2, Conflict: true, ConflictReason: "conflicting MCP servers"},
		),
		WithKitDetectNoColor(true),
	)
	got := r.ViewString()
	if !strings.Contains(got, "spring-java") {
		t.Errorf("ViewString() missing spring-java: %q", got)
	}
	if !strings.Contains(got, "nextjs") {
		t.Errorf("ViewString() missing nextjs: %q", got)
	}
	if !strings.Contains(got, "conflicting MCP servers") {
		t.Errorf("ViewString() missing conflict reason: %q", got)
	}
	if !strings.Contains(got, "Kits detected (2)") {
		t.Errorf("ViewString() missing header: %q", got)
	}
}

func TestKitDetectResult_AccessibleLabel(t *testing.T) {
	t.Parallel()
	r := NewKitDetectResult(
		WithKitMatches(KitMatch{Name: "go-module", Order: 1}),
	)
	label := r.AccessibleLabel()
	if !strings.Contains(label, "go-module") {
		t.Errorf("AccessibleLabel() missing kit name: %q", label)
	}
}

package widget

import (
	"strings"
	"testing"
)

func TestWorkareaPoolPanel_Empty(t *testing.T) {
	t.Parallel()
	p := NewWorkareaPoolPanel(WithPoolNoColor(true))
	got := p.ViewString()
	if got != "(empty pool)" {
		t.Errorf("ViewString() = %q, want '(empty pool)'", got)
	}
}

func TestWorkareaPoolPanel_Render(t *testing.T) {
	t.Parallel()
	p := NewWorkareaPoolPanel(
		WithPoolEntries(
			WorkareaPoolEntry{Repo: "github.com/org/api", Toolchain: "java=17", Warm: 3, Cold: 1, InUse: 2},
			WorkareaPoolEntry{Repo: "github.com/org/web", Toolchain: "node=20.x", Warm: 2, Cold: 0, InUse: 1},
		),
		WithPoolNoColor(true),
	)
	got := p.ViewString()
	if !strings.Contains(got, "github.com/org/api") {
		t.Errorf("ViewString() missing repo: %q", got)
	}
	if !strings.Contains(got, "java=17") {
		t.Errorf("ViewString() missing toolchain: %q", got)
	}
	if !strings.Contains(got, "3") {
		t.Errorf("ViewString() missing warm count: %q", got)
	}
}

func TestWorkareaPoolPanel_SetEntries(t *testing.T) {
	t.Parallel()
	p := NewWorkareaPoolPanel(WithPoolNoColor(true))
	p.SetEntries([]WorkareaPoolEntry{{Repo: "github.com/new", Toolchain: "go", Warm: 1}})
	got := p.ViewString()
	if !strings.Contains(got, "github.com/new") {
		t.Errorf("after SetEntries, missing repo: %q", got)
	}
}

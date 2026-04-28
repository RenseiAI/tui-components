package widget

import (
	"testing"

	"github.com/RenseiAI/tui-components/theme"
)

func TestScopePill_NoColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		scope Scope
		want  string
	}{
		{ScopeProject, "[project]"},
		{ScopeOrg, "[org]"},
		{ScopeTenant, "[tenant]"},
		{ScopeGlobal, "[global]"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(string(tc.scope), func(t *testing.T) {
			t.Parallel()
			p := NewScopePill(
				WithScopeValue(tc.scope),
				WithScopeNoColor(true),
			)
			got := p.ViewString()
			if got != tc.want {
				t.Errorf("ViewString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestScopePill_AccessibleLabel(t *testing.T) {
	t.Parallel()
	p := NewScopePill(WithScopeValue(ScopeOrg))
	want := "scope: org"
	if got := p.AccessibleLabel(); got != want {
		t.Errorf("AccessibleLabel() = %q, want %q", got, want)
	}
}

func TestScopePill_WithTheme(t *testing.T) {
	t.Parallel()
	p := NewScopePill(
		WithScopeValue(ScopeGlobal),
		WithScopeTheme(theme.HighContrastTheme()),
	)
	got := p.ViewString()
	if got == "" {
		t.Error("ViewString() returned empty string")
	}
}

func TestScopePill_SetTheme(t *testing.T) {
	t.Parallel()
	p := NewScopePill(WithScopeValue(ScopeTenant))
	p.SetTheme(theme.DarkTheme())
	_ = p.ViewString()
}

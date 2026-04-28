package theme

import (
	"testing"
)

// ---------------------------------------------------------------------------
// A11yMode constants
// ---------------------------------------------------------------------------

func TestA11yModeValues(t *testing.T) {
	// Guard that the zero value of A11yMode is A11yNone so that a default
	// Theme struct has the expected accessibility mode.
	var zero A11yMode
	if zero != A11yNone {
		t.Errorf("zero value of A11yMode = %d, want A11yNone (%d)", zero, A11yNone)
	}
	// Ensure the three modes are distinct.
	if A11yNone == A11yNoColor || A11yNoColor == A11yFull || A11yNone == A11yFull {
		t.Error("A11yMode constants must be distinct")
	}
}

// ---------------------------------------------------------------------------
// A11yModeFromEnv
// ---------------------------------------------------------------------------

func TestA11yModeFromEnv_None(t *testing.T) {
	t.Setenv("RENSEI_A11Y", "")
	t.Setenv("NO_COLOR", "")
	got := A11yModeFromEnv()
	if got != A11yNone {
		t.Errorf("A11yModeFromEnv() = %d, want A11yNone", got)
	}
}

func TestA11yModeFromEnv_NoColor(t *testing.T) {
	t.Setenv("RENSEI_A11Y", "")
	t.Setenv("NO_COLOR", "1")
	got := A11yModeFromEnv()
	if got != A11yNoColor {
		t.Errorf("A11yModeFromEnv() = %d, want A11yNoColor", got)
	}
}

func TestA11yModeFromEnv_Full(t *testing.T) {
	t.Setenv("RENSEI_A11Y", "true")
	t.Setenv("NO_COLOR", "")
	got := A11yModeFromEnv()
	if got != A11yFull {
		t.Errorf("A11yModeFromEnv() = %d, want A11yFull", got)
	}
}

func TestA11yModeFromEnv_FullPrecedesNoColor(t *testing.T) {
	// RENSEI_A11Y=true takes precedence over NO_COLOR.
	t.Setenv("RENSEI_A11Y", "true")
	t.Setenv("NO_COLOR", "1")
	got := A11yModeFromEnv()
	if got != A11yFull {
		t.Errorf("A11yModeFromEnv() = %d, want A11yFull when both env vars set", got)
	}
}

func TestA11yModeFromEnv_NoColor_AnyNonEmpty(t *testing.T) {
	// NO_COLOR spec: any non-empty string activates the mode.
	for _, val := range []string{"1", "true", "yes", " "} {
		t.Setenv("RENSEI_A11Y", "")
		t.Setenv("NO_COLOR", val)
		got := A11yModeFromEnv()
		if got != A11yNoColor {
			t.Errorf("A11yModeFromEnv() with NO_COLOR=%q = %d, want A11yNoColor", val, got)
		}
	}
}

// ---------------------------------------------------------------------------
// Theme.WithA11y
// ---------------------------------------------------------------------------

func TestWithA11y_None(t *testing.T) {
	base := DefaultTheme()
	got := base.WithA11y(A11yNone)
	if got.A11y != A11yNone {
		t.Errorf("WithA11y(A11yNone).A11y = %d, want A11yNone", got.A11y)
	}
	// Color tokens should be unchanged.
	if got.StatusError != base.StatusError {
		t.Error("WithA11y(A11yNone) must not change color tokens")
	}
}

func TestWithA11y_NoColor(t *testing.T) {
	base := DefaultTheme()
	got := base.WithA11y(A11yNoColor)
	if got.A11y != A11yNoColor {
		t.Errorf("WithA11y(A11yNoColor).A11y = %d, want A11yNoColor", got.A11y)
	}
	// Color tokens should be unchanged (only suppressed by renderers).
	if got.StatusError != base.StatusError {
		t.Error("WithA11y(A11yNoColor) must not change color tokens")
	}
}

func TestWithA11y_Full_SwitchesToHighContrastTheme(t *testing.T) {
	hc := HighContrastTheme()
	got := DefaultTheme().WithA11y(A11yFull)
	if got.A11y != A11yFull {
		t.Errorf("WithA11y(A11yFull).A11y = %d, want A11yFull", got.A11y)
	}
	// All color tokens should match HighContrastTheme.
	if got.StatusError != hc.StatusError {
		t.Errorf("WithA11y(A11yFull).StatusError = %v, want %v", got.StatusError, hc.StatusError)
	}
	if got.TextPrimary != hc.TextPrimary {
		t.Errorf("WithA11y(A11yFull).TextPrimary = %v, want %v", got.TextPrimary, hc.TextPrimary)
	}
	if got.BgPrimary != hc.BgPrimary {
		t.Errorf("WithA11y(A11yFull).BgPrimary = %v, want %v", got.BgPrimary, hc.BgPrimary)
	}
}

func TestWithA11y_Immutable(t *testing.T) {
	// WithA11y must return a new value; the original must be unchanged.
	base := DefaultTheme()
	origA11y := base.A11y
	origError := base.StatusError
	_ = base.WithA11y(A11yFull)
	if base.A11y != origA11y {
		t.Error("WithA11y mutated the receiver A11y field")
	}
	if base.StatusError != origError {
		t.Error("WithA11y mutated the receiver StatusError field")
	}
}

// ---------------------------------------------------------------------------
// Theme.RenderSymbol
// ---------------------------------------------------------------------------

func TestRenderSymbol(t *testing.T) {
	tests := []struct {
		name    string
		mode    A11yMode
		symbol  string
		verbose string
		want    string
	}{
		{"none_returns_symbol", A11yNone, "✓", "[OK]", "✓"},
		{"no_color_returns_symbol", A11yNoColor, "✗", "[ERROR]", "✗"},
		{"full_returns_verbose", A11yFull, "●", "[WORKING]", "[WORKING]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := DefaultTheme().WithA11y(tt.mode)
			got := th.RenderSymbol(tt.symbol, tt.verbose)
			if got != tt.want {
				t.Errorf("RenderSymbol(%q, %q) with A11y=%d = %q, want %q",
					tt.symbol, tt.verbose, tt.mode, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Theme.NoColor
// ---------------------------------------------------------------------------

func TestNoColor(t *testing.T) {
	tests := []struct {
		name string
		mode A11yMode
		want bool
	}{
		{"none_has_color", A11yNone, false},
		{"no_color_mode", A11yNoColor, true},
		{"full_suppresses_color", A11yFull, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := DefaultTheme().WithA11y(tt.mode)
			got := th.NoColor()
			if got != tt.want {
				t.Errorf("NoColor() with A11y=%d = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DefaultTheme / HighContrastTheme zero A11y field
// ---------------------------------------------------------------------------

func TestDefaultThemeA11yIsNone(t *testing.T) {
	if DefaultTheme().A11y != A11yNone {
		t.Error("DefaultTheme().A11y must be A11yNone")
	}
}

func TestHighContrastThemeA11yIsNone(t *testing.T) {
	// HighContrastTheme does not force A11yFull — callers must set the mode
	// explicitly via WithA11y.  This keeps the constructors orthogonal.
	if HighContrastTheme().A11y != A11yNone {
		t.Error("HighContrastTheme().A11y must be A11yNone (callers set A11yFull via WithA11y)")
	}
}

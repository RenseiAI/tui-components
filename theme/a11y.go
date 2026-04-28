package theme

import "os"

// A11yMode encodes the accessibility rendering mode for a Theme.
// Primitives that accept a Theme inspect A11yMode to determine whether to
// emit Unicode symbols with color, or to degrade to ASCII-only verbose labels.
//
// # Mode values
//
//   - [A11yNone]     — default rich rendering (Unicode symbols + color).
//   - [A11yNoColor]  — symbol-first rendering, color suppressed.
//     Activated when the NO_COLOR env var is set to any non-empty value.
//   - [A11yFull]     — high-contrast theme + symbol-first + verbose-label-only.
//     Activated when RENSEI_A11Y=true.
//
// The mode is stored on the Theme value so widgets and format helpers receive
// the a11y contract from the same source as the color contract — the Theme —
// rather than via a separate side channel.
type A11yMode int

const (
	// A11yNone is the default mode: full Unicode symbols and color.
	A11yNone A11yMode = 0

	// A11yNoColor suppresses color but preserves Unicode symbols.  Follows
	// the NO_COLOR spec (https://no-color.org/): set the NO_COLOR env var to
	// any non-empty string to activate.
	A11yNoColor A11yMode = 1

	// A11yFull enables high-contrast mode, symbol-first rendering, and
	// verbose label-only output (no icons).  Activated by RENSEI_A11Y=true.
	A11yFull A11yMode = 2
)

// A11yModeFromEnv detects the accessibility mode from the process environment.
//
// Detection order (highest precedence first):
//  1. RENSEI_A11Y=true  → [A11yFull]
//  2. NO_COLOR non-empty → [A11yNoColor]
//  3. otherwise          → [A11yNone]
//
// Call this once at startup and embed the result into your Theme; do not call
// it per-render as it performs an os.Getenv lookup each time.
func A11yModeFromEnv() A11yMode {
	if os.Getenv("RENSEI_A11Y") == "true" {
		return A11yFull
	}
	if os.Getenv("NO_COLOR") != "" {
		return A11yNoColor
	}
	return A11yNone
}

// WithA11y returns a copy of t with the given A11yMode applied.  If mode is
// [A11yFull], the color tokens are replaced with those of [HighContrastTheme]
// so the widget layer automatically renders at the correct contrast level
// without further intervention from the caller.
//
// Example:
//
//	t := theme.DefaultTheme().WithA11y(theme.A11yModeFromEnv())
func (t Theme) WithA11y(mode A11yMode) Theme {
	t.A11y = mode
	if mode == A11yFull {
		hc := HighContrastTheme()
		// Preserve the A11y field from t; overwrite all color tokens.
		t.BgPrimary = hc.BgPrimary
		t.BgSecondary = hc.BgSecondary
		t.BgTertiary = hc.BgTertiary
		t.Surface = hc.Surface
		t.SurfaceRaised = hc.SurfaceRaised
		t.SurfaceBorder = hc.SurfaceBorder
		t.SurfaceBorderBright = hc.SurfaceBorderBright
		t.Accent = hc.Accent
		t.AccentDim = hc.AccentDim
		t.Teal = hc.Teal
		t.TealDim = hc.TealDim
		t.Blue = hc.Blue
		t.StatusSuccess = hc.StatusSuccess
		t.StatusWarning = hc.StatusWarning
		t.StatusError = hc.StatusError
		t.StatusInfo = hc.StatusInfo
		t.TextPrimary = hc.TextPrimary
		t.TextSecondary = hc.TextSecondary
		t.TextTertiary = hc.TextTertiary
		t.A11y = A11yFull
	}
	return t
}

// RenderSymbol returns the appropriate representation for a status symbol
// given the theme's A11yMode.
//
// Parameters:
//   - symbol: the Unicode glyph used in [A11yNone] and [A11yNoColor] modes
//     (e.g. "✓", "✗", "●").
//   - verboseLabel: the ASCII-only label used in [A11yFull] mode
//     (e.g. "[OK]", "[ERROR]", "[WORKING]").
//
// Behavior by mode:
//   - [A11yNone]    → symbol (colored by the caller)
//   - [A11yNoColor] → symbol (no color; caller must suppress color application)
//   - [A11yFull]    → verboseLabel
//
// The symbol/verboseLabel distinction is intentionally kept here — in the
// theme package — so the format and widget packages share one rendering contract.
func (t Theme) RenderSymbol(symbol, verboseLabel string) string {
	if t.A11y == A11yFull {
		return verboseLabel
	}
	return symbol
}

// NoColor reports whether this theme suppresses color output.  Returns true
// for both [A11yNoColor] and [A11yFull] modes.
func (t Theme) NoColor() bool {
	return t.A11y == A11yNoColor || t.A11y == A11yFull
}

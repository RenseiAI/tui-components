// Package theme provides the color palette, Lipgloss styles, and Theme struct
// used by tui-components widgets.
//
// # Theme struct
//
// A [Theme] carries all color tokens and derived [charm.land/lipgloss/v2] styles
// in a single value that can be created, copied, and swapped at runtime.
// Three constructors are provided out of the box:
//
//   - [DefaultTheme] — the canonical dark-navy palette (mirrors the v0.1.0
//     package-level vars).
//   - [DarkTheme] — a deeper true-black dark variant.
//   - [HighContrastTheme] — high-contrast white-on-black for a11y environments.
//
// # Swap mechanism
//
// Widgets accept a [Theme] via their [widget.WithTheme] functional option.
// Calling WithTheme mid-render updates the widget's internal theme field; the
// next call to View uses the new theme.  There is no global mutable state —
// each widget owns its copy of the theme.
//
// For consumers that still need a package-level default (e.g. one-off style
// helpers), [Default] returns a mutable pointer to a package-level [Theme]
// that is initialised to [DefaultTheme].  Mutating [Default] affects all
// callers that reference it, so prefer passing an explicit Theme.
//
// # Backward compatibility
//
// In v0.1.0 the palette was exposed as package-level vars (BgPrimary, Accent,
// etc.).  Those vars are removed in v0.2.0.  See MIGRATION-v0.2.0.md for the
// one-line mechanical migration.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Theme holds all color tokens and derived Lipgloss styles for a single
// visual theme.  Pass a Theme to widgets via widget.WithTheme; create one
// with DefaultTheme, DarkTheme, or HighContrastTheme.
//
// Fields are grouped into five semantic sections:
//
//  1. Background hierarchy  (BgPrimary, BgSecondary, BgTertiary)
//  2. Surface hierarchy     (Surface, SurfaceRaised, SurfaceBorder, SurfaceBorderBright)
//  3. Accent palette        (Accent, AccentDim, Teal, TealDim, Blue)
//  4. Status semantics      (StatusSuccess, StatusWarning, StatusError, StatusInfo)
//  5. Text hierarchy        (TextPrimary, TextSecondary, TextTertiary)
type Theme struct {
	// --- Background hierarchy ------------------------------------------------

	// BgPrimary is the deepest background color — the canvas of the
	// application window.
	BgPrimary color.Color

	// BgSecondary is the second background level, used for panels or
	// sidebars that sit slightly above the primary canvas.
	BgSecondary color.Color

	// BgTertiary is the third background level, used for input areas or
	// nested containers.
	BgTertiary color.Color

	// --- Surface hierarchy ---------------------------------------------------

	// Surface is the base surface color for cards, modals, and similar
	// components that float above the background.
	Surface color.Color

	// SurfaceRaised is the elevated surface color for hovered, selected, or
	// focused elements within a surface.
	SurfaceRaised color.Color

	// SurfaceBorder is the standard border color separating surface elements.
	SurfaceBorder color.Color

	// SurfaceBorderBright is the brighter border variant used for focus rings
	// and prominent separators.
	SurfaceBorderBright color.Color

	// --- Accent palette ------------------------------------------------------

	// Accent is the primary brand color — used for highlights, active
	// indicators, and the spinner.
	Accent color.Color

	// AccentDim is a dimmer variant of Accent, used for pressed/active states.
	AccentDim color.Color

	// Teal is the secondary accent color used in progress bars, success chips,
	// and complementary highlights.
	Teal color.Color

	// TealDim is a dimmer variant of Teal.
	TealDim color.Color

	// Blue is the tertiary accent used for informational chips and links.
	Blue color.Color

	// --- Status semantics ----------------------------------------------------

	// StatusSuccess indicates a healthy, completed, or ready state.
	StatusSuccess color.Color

	// StatusWarning indicates a degraded or pending state that requires
	// attention but is not a failure.
	StatusWarning color.Color

	// StatusError indicates a failed or unhealthy state.
	StatusError color.Color

	// StatusInfo is a neutral informational color for status messages that
	// are neither success nor failure.
	StatusInfo color.Color

	// --- Text hierarchy ------------------------------------------------------

	// TextPrimary is the highest-contrast text color — body copy, labels.
	TextPrimary color.Color

	// TextSecondary is the medium-contrast text color — secondary labels,
	// help text.
	TextSecondary color.Color

	// TextTertiary is the lowest-contrast text color — hints, placeholders,
	// disabled text.
	TextTertiary color.Color

	// --- Accessibility -------------------------------------------------------

	// A11y controls the accessibility rendering mode for primitives that use
	// this theme.  The zero value ([A11yNone]) is the default: full Unicode
	// symbols and color.  Use [Theme.WithA11y] to apply a mode, or detect it
	// from the environment with [A11yModeFromEnv].
	//
	// Widgets and format helpers must read A11y from the Theme they receive —
	// never from os.Getenv directly — so that tests and server-side renderers
	// can override the mode without touching the process environment.
	A11y A11yMode
}

// DefaultTheme returns the canonical tui-components theme — the dark navy
// palette previously exposed as package-level vars in v0.1.0.
func DefaultTheme() Theme {
	return Theme{
		// Background hierarchy
		BgPrimary:   lipgloss.Color("#080C16"),
		BgSecondary: lipgloss.Color("#0D1220"),
		BgTertiary:  lipgloss.Color("#111828"),
		// Surface hierarchy
		Surface:             lipgloss.Color("#141B2D"),
		SurfaceRaised:       lipgloss.Color("#1A2236"),
		SurfaceBorder:       lipgloss.Color("#1E2740"),
		SurfaceBorderBright: lipgloss.Color("#283350"),
		// Accent palette
		Accent:    lipgloss.Color("#FF6B35"),
		AccentDim: lipgloss.Color("#CC5529"),
		Teal:      lipgloss.Color("#00D4AA"),
		TealDim:   lipgloss.Color("#00A886"),
		Blue:      lipgloss.Color("#4B8BF5"),
		// Status semantics
		StatusSuccess: lipgloss.Color("#22C55E"),
		StatusWarning: lipgloss.Color("#F59E0B"),
		StatusError:   lipgloss.Color("#EF4444"),
		StatusInfo:    lipgloss.Color("#4B8BF5"),
		// Text hierarchy
		TextPrimary:   lipgloss.Color("#F1F5F9"),
		TextSecondary: lipgloss.Color("#7C8DB5"),
		TextTertiary:  lipgloss.Color("#4B5B80"),
	}
}

// DarkTheme returns an explicit true-black dark variant suitable for OLED
// displays and terminal emulators that render pure black as transparent.
func DarkTheme() Theme {
	return Theme{
		// Background hierarchy — deeper blacks
		BgPrimary:   lipgloss.Color("#000000"),
		BgSecondary: lipgloss.Color("#050505"),
		BgTertiary:  lipgloss.Color("#0A0A0A"),
		// Surface hierarchy — slightly lighter than bg to retain contrast
		Surface:             lipgloss.Color("#0F0F0F"),
		SurfaceRaised:       lipgloss.Color("#1A1A1A"),
		SurfaceBorder:       lipgloss.Color("#222222"),
		SurfaceBorderBright: lipgloss.Color("#333333"),
		// Accent palette — same hue, slightly brightened for black bg
		Accent:    lipgloss.Color("#FF7A47"),
		AccentDim: lipgloss.Color("#CC6238"),
		Teal:      lipgloss.Color("#00E5B8"),
		TealDim:   lipgloss.Color("#00B892"),
		Blue:      lipgloss.Color("#5C9BFF"),
		// Status semantics
		StatusSuccess: lipgloss.Color("#2ECC71"),
		StatusWarning: lipgloss.Color("#FFC107"),
		StatusError:   lipgloss.Color("#FF5555"),
		StatusInfo:    lipgloss.Color("#5C9BFF"),
		// Text hierarchy — brighter on pure black
		TextPrimary:   lipgloss.Color("#FFFFFF"),
		TextSecondary: lipgloss.Color("#8899CC"),
		TextTertiary:  lipgloss.Color("#556688"),
	}
}

// HighContrastTheme returns a high-contrast theme that meets WCAG AA contrast
// requirements.  It is the preferred theme when RENSEI_A11Y=true or when the
// NO_COLOR env var is set with color-safe fallbacks.
func HighContrastTheme() Theme {
	return Theme{
		// Background hierarchy — solid blacks
		BgPrimary:   lipgloss.Color("#000000"),
		BgSecondary: lipgloss.Color("#000000"),
		BgTertiary:  lipgloss.Color("#111111"),
		// Surface hierarchy — high-contrast borders
		Surface:             lipgloss.Color("#000000"),
		SurfaceRaised:       lipgloss.Color("#1C1C1C"),
		SurfaceBorder:       lipgloss.Color("#FFFFFF"),
		SurfaceBorderBright: lipgloss.Color("#FFFFFF"),
		// Accent palette — pure yellows and cyans for a11y legibility
		Accent:    lipgloss.Color("#FFFF00"),
		AccentDim: lipgloss.Color("#CCCC00"),
		Teal:      lipgloss.Color("#00FFFF"),
		TealDim:   lipgloss.Color("#00CCCC"),
		Blue:      lipgloss.Color("#6699FF"),
		// Status semantics — high-contrast status colors
		StatusSuccess: lipgloss.Color("#00FF00"),
		StatusWarning: lipgloss.Color("#FFFF00"),
		StatusError:   lipgloss.Color("#FF0000"),
		StatusInfo:    lipgloss.Color("#6699FF"),
		// Text hierarchy — pure white on black
		TextPrimary:   lipgloss.Color("#FFFFFF"),
		TextSecondary: lipgloss.Color("#DDDDDD"),
		TextTertiary:  lipgloss.Color("#AAAAAA"),
	}
}

// pkg is the package-level default Theme used by the legacy style helpers
// in styles.go.  Callers that depend on the package-level helpers receive
// styles derived from this theme.  Replacing the pointer target propagates
// to all subsequent calls to those helpers.
//
// New code should use explicit Theme values rather than relying on this
// package-level default.
var pkg = DefaultTheme() //nolint:gochecknoglobals

// Default returns a pointer to the package-level default [Theme].  Mutating
// the returned value affects all callers of the legacy style helpers
// (Header, StatLabel, etc.) that do not accept an explicit Theme.  Prefer
// passing an explicit Theme to widgets via widget.WithTheme.
func Default() *Theme {
	return &pkg
}

package theme

import (
	"image/color"
)

// ActivityColors maps activity types to their display colors.
// Colors are derived from the package-level default theme at program startup.
// For theme-aware rendering, access colors from an explicit Theme value.
var ActivityColors = map[string]color.Color{
	"thought":  pkg.TextSecondary,
	"action":   pkg.Teal,
	"response": pkg.TextPrimary,
	"error":    pkg.StatusError,
	"progress": pkg.StatusSuccess,
}

// ActivityIcons maps activity types to their display icons.
var ActivityIcons = map[string]string{
	"thought":  "\U0001f4ad",
	"action":   "⚡",
	"response": "\U0001f4ac",
	"error":    "✗",
	"progress": "✓",
}

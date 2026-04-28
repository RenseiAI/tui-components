package theme

import "image/color"

// ActivityColors maps activity types to their display colors.
// This map is populated from [GlobalRegistry] at package init and reflects
// the built-in activity kinds.  For theme-aware or registry-aware rendering,
// use [GetActivityColor] or [Registry.GetActivity] directly.
//
// Deprecated: prefer [GetActivityColor] which queries [GlobalRegistry] and
// handles unknown kinds gracefully.
var ActivityColors = map[string]color.Color{
	"thought":  pkg.TextSecondary,
	"action":   pkg.Teal,
	"response": pkg.TextPrimary,
	"error":    pkg.StatusError,
	"progress": pkg.StatusSuccess,
}

// ActivityIcons maps activity types to their display icons.
// This map is populated from [GlobalRegistry] at package init and reflects
// the built-in activity kinds.  For registry-aware rendering, use
// [GetActivityIcon] or [Registry.GetActivity] directly.
//
// Deprecated: prefer [GetActivityIcon] which queries [GlobalRegistry] and
// handles unknown kinds gracefully.
var ActivityIcons = map[string]string{
	"thought":  "\U0001f4ad",
	"action":   "⚡",
	"response": "\U0001f4ac",
	"error":    "✗",
	"progress": "✓",
}

// GetActivityColor returns the display color for an activity kind.
// It queries [GlobalRegistry]; if the kind is not registered, it returns
// [Theme.TextSecondary] from the default theme.
//
// Built-in activity kinds are pre-registered at package init.  Plugins
// register additional kinds with [Registry.RegisterActivity].
func GetActivityColor(kind string) color.Color {
	if e, ok := GlobalRegistry.GetActivity(kind); ok {
		return e.Color
	}
	return pkg.TextSecondary
}

// GetActivityIcon returns the display icon for an activity kind.
// It queries [GlobalRegistry]; if the kind is not registered, it returns
// "?" as the fallback icon.
//
// Built-in activity kinds are pre-registered at package init.  Plugins
// register additional kinds with [Registry.RegisterActivity].
func GetActivityIcon(kind string) string {
	if e, ok := GlobalRegistry.GetActivity(kind); ok {
		return e.Icon
	}
	return "?"
}

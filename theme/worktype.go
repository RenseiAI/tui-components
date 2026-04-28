package theme

import (
	"image/color"
	"strings"
)

// GetWorkTypeColor returns the display color for a work type.
// It queries [GlobalRegistry] for the kind (case-insensitive); if the kind
// is not registered, it returns [Theme.TextSecondary] from the default theme.
//
// Built-in work types are pre-registered at package init.  Plugins register
// additional kinds with [Registry.RegisterWorkType].
func GetWorkTypeColor(workType string) color.Color {
	if e, ok := GlobalRegistry.GetWorkType(strings.ToLower(workType)); ok {
		return e.Color
	}
	return pkg.TextSecondary
}

// GetWorkTypeLabel returns the display label for a work type.
// It queries [GlobalRegistry] for the kind (case-insensitive); if the kind
// is not registered, the raw workType string is returned unchanged.
//
// Built-in work types are pre-registered at package init.  Plugins register
// additional kinds with [Registry.RegisterWorkType].
func GetWorkTypeLabel(workType string) string {
	if e, ok := GlobalRegistry.GetWorkType(strings.ToLower(workType)); ok {
		return e.Label
	}
	return workType
}

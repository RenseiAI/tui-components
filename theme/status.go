package theme

import "image/color"

// StatusStyle defines the visual representation of a session status.
// The canonical set of built-in status kinds is pre-registered in
// [GlobalRegistry] at package init.  Third-party callers can register
// additional kinds with [Registry.RegisterStatus].
type StatusStyle struct {
	Label   string
	Color   color.Color
	Symbol  string
	Animate bool
}

// GetStatusStyle returns the visual style for a session status string.
// It queries [GlobalRegistry] for the kind; if the kind is not registered,
// it returns a fallback style with a "?" symbol and the label "Unknown".
//
// Built-in statuses: "working", "queued", "parked", "completed", "failed",
// "stopped".  Plugins register additional kinds with
// [Registry.RegisterStatus].
func GetStatusStyle(status string) StatusStyle {
	if e, ok := GlobalRegistry.GetStatus(status); ok {
		return StatusStyle{
			Label:   e.Label,
			Color:   e.Color,
			Symbol:  e.Symbol,
			Animate: e.Animate,
		}
	}
	return StatusStyle{
		Label:   "Unknown",
		Color:   pkg.TextSecondary,
		Symbol:  "?",
		Animate: false,
	}
}

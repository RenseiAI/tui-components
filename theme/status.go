package theme

import (
	"image/color"
)

// StatusStyle defines the visual representation of a session status.
type StatusStyle struct {
	Label   string
	Color   color.Color
	Symbol  string
	Animate bool
}

// GetStatusStyle returns the visual style for a session status string.
// Known statuses: "working", "queued", "parked", "completed", "failed", "stopped".
func GetStatusStyle(status string) StatusStyle {
	switch status {
	case "working":
		return StatusStyle{"Working", pkg.StatusSuccess, "●", true} // ●
	case "queued":
		return StatusStyle{"Queued", pkg.StatusWarning, "◌", true} // ◌
	case "parked":
		return StatusStyle{"Parked", pkg.TextTertiary, "○", false} // ○
	case "completed":
		return StatusStyle{"Done", pkg.StatusSuccess, "✓", false} // ✓
	case "failed":
		return StatusStyle{"Failed", pkg.StatusError, "✗", false} // ✗
	case "stopped":
		return StatusStyle{"Stopped", pkg.TextTertiary, "■", false} // ■
	default:
		return StatusStyle{"Unknown", pkg.TextSecondary, "?", false}
	}
}

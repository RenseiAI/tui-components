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
		return StatusStyle{"Working", StatusSuccess, "\u25cf", true} // ●
	case "queued":
		return StatusStyle{"Queued", StatusWarning, "\u25cc", true} // ◌
	case "parked":
		return StatusStyle{"Parked", TextTertiary, "\u25cb", false} // ○
	case "completed":
		return StatusStyle{"Done", StatusSuccess, "\u2713", false} // ✓
	case "failed":
		return StatusStyle{"Failed", StatusError, "\u2717", false} // ✗
	case "stopped":
		return StatusStyle{"Stopped", TextTertiary, "\u25a0", false} // ■
	default:
		return StatusStyle{"Unknown", TextSecondary, "?", false}
	}
}

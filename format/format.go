package format

import (
	"fmt"
	"time"
)

// Duration formats seconds into a human-readable duration string.
//
// The input range accepted is any int, including negative values. Current
// behavior passes negatives through fmt.Sprintf unchanged, so Duration(-30)
// returns "-30s" and Duration(-3600) returns "-1h". Callers that cannot
// tolerate negative output should clamp the value before calling.
//
// Units scale from seconds to hours: values < 60 render as "%ds", values
// < 3600 render as "%dm" or "%dm %ds", and values >= 3600 render as "%dh"
// or "%dh %dm". Day-scale inputs are currently expressed in hours — for
// example, Duration(86400) returns "24h" and Duration(604800) returns "168h"
// — because no day unit is emitted.
//
// Future versions may clamp negative inputs to "0s" and render day-scale
// values with an explicit day unit (for example "%dd %dh"); callers should
// not rely on the current negative or day-scale output.
func Duration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		m := seconds / 60
		s := seconds % 60
		if s > 0 {
			return fmt.Sprintf("%dm %ds", m, s)
		}
		return fmt.Sprintf("%dm", m)
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	if m > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dh", h)
}

// Cost formats a USD cost value for display.
func Cost(usd *float64) string {
	if usd == nil || *usd == 0 {
		return "--"
	}
	if *usd < 0.01 {
		return fmt.Sprintf("$%.4f", *usd)
	}
	return fmt.Sprintf("$%.2f", *usd)
}

// RelativeTime formats an ISO 8601 timestamp as a relative time string.
func RelativeTime(isoString string) string {
	t, err := time.Parse(time.RFC3339, isoString)
	if err != nil {
		return isoString
	}
	diff := time.Since(t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	case diff < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	}
}

// Timestamp formats an ISO 8601 string to local time display.
func Timestamp(isoString string) string {
	t, err := time.Parse(time.RFC3339, isoString)
	if err != nil {
		return isoString
	}
	return t.Local().Format("3:04:05 PM")
}

// ProviderName returns a display name for a provider, or "--" if nil.
func ProviderName(provider *string) string {
	if provider == nil {
		return "--"
	}
	return *provider
}

// Tokens formats a token count for display.
func Tokens(count *int) string {
	if count == nil {
		return "--"
	}
	n := *count
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	return fmt.Sprintf("%.1fk", float64(n)/1000.0)
}

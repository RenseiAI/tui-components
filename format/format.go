package format

import (
	"fmt"
	"time"
)

// Duration formats seconds into a human-readable duration string.
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

// now is the clock source used by RelativeTime. It is a package variable
// solely so tests can substitute a fixed clock for deterministic assertions;
// production code should not reassign it.
var now = time.Now

// RelativeTime formats an ISO 8601 timestamp as a relative time string.
//
// If isoString cannot be parsed as RFC 3339, the original input is returned
// unchanged (passthrough-on-parse-failure). Timestamps in the future (diff < 0)
// render as "just now".
func RelativeTime(isoString string) string {
	t, err := time.Parse(time.RFC3339, isoString)
	if err != nil {
		return isoString
	}
	diff := now().Sub(t)
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
//
// If isoString cannot be parsed as RFC 3339, the original input is returned
// unchanged (passthrough-on-parse-failure).
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

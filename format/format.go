package format

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// Duration formats seconds into a human-readable duration string.
//
// Negative inputs are clamped to "0s" to match the zero sentinel; callers
// never see a negative duration in the output.
//
// Units scale from seconds up to days:
//   - values < 60 render as "%ds" (e.g. "30s")
//   - values < 3600 render as "%dm" or "%dm %ds" (e.g. "5m", "1m 30s")
//   - values < 86400 render as "%dh" or "%dh %dm" (e.g. "2h", "3h 45m")
//   - values >= 86400 render as "%dd %dh" (e.g. "1d 0h", "7d 1h"); minutes
//     are intentionally dropped at day scale.
func Duration(seconds int) string {
	if seconds < 0 {
		return "0s"
	}
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
	if seconds < 86400 {
		h := seconds / 3600
		m := (seconds % 3600) / 60
		if m > 0 {
			return fmt.Sprintf("%dh %dm", h, m)
		}
		return fmt.Sprintf("%dh", h)
	}
	d := seconds / 86400
	h := (seconds % 86400) / 3600
	return fmt.Sprintf("%dd %dh", d, h)
}

// Cost formats a USD cost value for display.
//
// Behavior:
//   - nil or zero (including -0.0) renders as "--" (treated as missing).
//   - Negative values render as "--" (costs are unsigned in this domain).
//   - NaN and ±Inf render as "--" (treated as invalid).
//   - Values < 0.01 render with four decimal places (e.g. "$0.0050").
//   - All other values render with two decimal places (e.g. "$3.42").
//   - Large values are not abbreviated (e.g. 1_000_000.0 → "$1000000.00").
func Cost(usd *float64) string {
	if usd == nil || *usd == 0 {
		return "--"
	}
	v := *usd
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return "--"
	}
	if v < 0 {
		return "--"
	}
	if v < 0.01 {
		return fmt.Sprintf("$%.4f", v)
	}
	return fmt.Sprintf("$%.2f", v)
}

// now is the clock source used by RelativeTime. It is a package variable
// solely so tests can substitute a fixed clock for deterministic assertions;
// production code should not reassign it.
var now = time.Now

// RelativeTime formats an ISO 8601 timestamp as a relative time string.
//
// If isoString cannot be parsed as RFC 3339, the original input is returned
// unchanged (passthrough-on-parse-failure). Timestamps in the future (where
// the parsed time is after now) collapse to "just now".
//
// Past timestamps bucket into minute/hour/day scales:
//   - diff < 1m  → "just now"
//   - diff < 1h  → "%dm ago"
//   - diff < 24h → "%dh ago"
//   - otherwise  → "%dd ago"
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

// ProviderName returns a display name for a provider.
//
// Behavior:
//   - nil pointer returns "--".
//   - A pointer to the empty string ("") returns "--".
//   - A pointer to a whitespace-only string (e.g. " ", "\t") returns "--".
//   - Any other non-nil pointer is dereferenced and returned verbatim
//     (including surrounding content — only the purely whitespace case is
//     collapsed to the missing sentinel).
func ProviderName(provider *string) string {
	if provider == nil {
		return "--"
	}
	if strings.TrimSpace(*provider) == "" {
		return "--"
	}
	return *provider
}

// Tokens formats a token count for display.
//
// Behavior:
//   - nil      → "--"
//   - negative → "--" (token counts are semantically unsigned)
//   - < 1000            → plain integer, e.g. 999 → "999"
//   - < 1_000_000       → k-scale with one decimal, e.g. 1500 → "1.5k"
//   - >= 1_000_000      → M-scale with one decimal, e.g. 1_500_000 → "1.5M"
func Tokens(count *int) string {
	if count == nil {
		return "--"
	}
	n := *count
	if n < 0 {
		return "--"
	}
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1_000_000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.1fM", float64(n)/1_000_000.0)
}

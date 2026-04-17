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
//
// Current behavior:
//   - nil or zero (including -0.0) renders as "--" (treated as missing).
//   - Values < 0.01 render with four decimal places (e.g. "$0.0050").
//   - All other values render with two decimal places (e.g. "$3.42").
//   - Negative values, NaN, and ±Inf are currently passed through to
//     fmt.Sprintf verbatim. Note that because any negative number is < 0.01,
//     negatives hit the four-decimal branch (e.g. -3.42 → "$-3.4200"); NaN
//     and ±Inf render as "$NaN", "$+Inf", "$-Inf".
//   - Large values are not abbreviated (e.g. 1_000_000.0 → "$1000000.00").
//
// Planned revision (TC-011.6): negative, NaN, and ±Inf will render as "--"
// (treated as invalid/missing), and large values may be abbreviated
// (e.g. "$1.0M"). Do not rely on the current passthrough behavior.
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

// ProviderName returns a display name for a provider.
//
// Current behavior:
//   - nil pointer returns "--".
//   - Any non-nil pointer is dereferenced and returned as-is, including
//     the empty string ("") and whitespace-only values (e.g. " ").
//
// Note: a future revision (tracked by TC-011.6) will treat empty and
// whitespace-only pointer values as "--" to match nil semantics, so callers
// should not rely on empty/whitespace passthrough.
func ProviderName(provider *string) string {
	if provider == nil {
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

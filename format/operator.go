package format

import (
	"fmt"
	"sort"
	"strings"
)

// CapacityRatio formats a current / max capacity pair as a human-readable
// string.
//
// Behavior:
//   - maxConcurrent == nil → "current / ∞" (unbounded capacity)
//   - maxConcurrent == 0   → "current / 0" (explicitly zero-bounded)
//   - otherwise            → "current / max"
//
// The current value is always rendered as-is; no clamping is applied.
//
// Examples:
//
//	CapacityRatio(5, intPtr(8))  → "5 / 8"
//	CapacityRatio(5, nil)        → "5 / ∞"
//	CapacityRatio(0, intPtr(0))  → "0 / 0"
func CapacityRatio(current int, maxConcurrent *int) string {
	if maxConcurrent == nil {
		return fmt.Sprintf("%d / ∞", current)
	}
	return fmt.Sprintf("%d / %d", current, *maxConcurrent)
}

// AttestationFingerprint formats a raw key fingerprint string into a
// human-readable truncated form.
//
// Format: "<algorithm>:<prefix>…<suffix>"
//
// Rules:
//   - If fingerprint is empty, "--" is returned.
//   - If fingerprint has no ":" separator, the raw value is returned unchanged.
//   - The hash portion is truncated: the first 7 characters followed by "…"
//     followed by the last 4 characters.  If the hash is 11 characters or
//     shorter, it is returned verbatim (no truncation applied).
//   - The algorithm prefix (everything before the first ":") is preserved as-is.
//
// Examples:
//
//	AttestationFingerprint("ed25519:abc1234abcdd4f2") → "ed25519:abc1234…d4f2"
//	AttestationFingerprint("")                        → "--"
//	AttestationFingerprint("nocolon")                 → "nocolon"
func AttestationFingerprint(fingerprint string) string {
	if fingerprint == "" {
		return "--"
	}
	idx := strings.Index(fingerprint, ":")
	if idx < 0 {
		return fingerprint
	}
	algo := fingerprint[:idx]
	hash := fingerprint[idx+1:]
	// Only truncate hashes longer than 11 characters (prefix 7 + suffix 4).
	if len(hash) <= 11 {
		return fingerprint
	}
	truncated := hash[:7] + "…" + hash[len(hash)-4:]
	return algo + ":" + truncated
}

// RegionList formats a slice of region identifiers as a compact
// human-readable string with optional overflow indicator.
//
// Behavior:
//   - nil or empty slice → "--"
//   - 1–maxVisible items → comma-separated list (e.g. "iad1, ord1")
//   - more than maxVisible items → first maxVisible joined, then " +N more"
//
// maxVisible must be >= 1; values < 1 are treated as 1.
//
// Examples:
//
//	RegionList([]string{"iad1", "ord1", "sea1", "dfw1"}, 1) → "iad1, +3 more"
//	RegionList([]string{"iad1", "ord1"}, 3)                  → "iad1, ord1"
//	RegionList(nil, 3)                                        → "--"
func RegionList(regions []string, maxVisible int) string {
	if len(regions) == 0 {
		return "--"
	}
	if maxVisible < 1 {
		maxVisible = 1
	}
	if len(regions) <= maxVisible {
		return strings.Join(regions, ", ")
	}
	overflow := len(regions) - maxVisible
	return strings.Join(regions[:maxVisible], ", ") + fmt.Sprintf(", +%d more", overflow)
}

// ToolchainSpec formats a map of toolchain name → version pairs into a
// compact human-readable string.
//
// The entries are sorted by name to produce a stable, deterministic output.
// Each entry is formatted as "name=version". Entries are joined with ", ".
//
// Behavior:
//   - nil or empty map → "--"
//   - otherwise → sorted "name=version" pairs joined by ", "
//
// Examples:
//
//	ToolchainSpec(map[string]string{"java": "17", "node": "20"}) → "java=17, node=20"
//	ToolchainSpec(nil)                                           → "--"
func ToolchainSpec(toolchains map[string]string) string {
	if len(toolchains) == 0 {
		return "--"
	}
	keys := make([]string, 0, len(toolchains))
	for k := range toolchains {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+toolchains[k])
	}
	return strings.Join(parts, ", ")
}

// HumanLabel returns the human-readable label for a typed flag value by
// performing a lookup in the provided labels map.
//
// T is constrained to comparable so that any typed enum (string alias, int
// alias, etc.) can be used as a map key.
//
// Behavior:
//   - If labels is nil or the key is not found, "--" is returned.
//   - Otherwise the mapped string label is returned.
//
// Examples:
//
//	type BillingModel string
//	labels := map[BillingModel]string{
//	    "wall-clock":  "Wall-clock time",
//	    "active-cpu":  "Active CPU only",
//	}
//	HumanLabel("wall-clock", labels)  → "Wall-clock time"
//	HumanLabel("invocation", labels)  → "--"
func HumanLabel[T comparable](value T, labels map[T]string) string {
	if labels == nil {
		return "--"
	}
	if label, ok := labels[value]; ok {
		return label
	}
	return "--"
}

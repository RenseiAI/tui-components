package format

import "testing"

// intPtr is a local helper that returns a pointer to an int literal.
func intPtr(n int) *int { return &n }

// ---------------------------------------------------------------------------
// CapacityRatio
// ---------------------------------------------------------------------------

func TestCapacityRatio(t *testing.T) {
	tests := []struct {
		name    string
		current int
		max     *int
		want    string
	}{
		{"bounded_5_of_8", 5, intPtr(8), "5 / 8"},
		{"bounded_0_of_0", 0, intPtr(0), "0 / 0"},
		{"bounded_0_of_1", 0, intPtr(1), "0 / 1"},
		{"bounded_max_equal_current", 8, intPtr(8), "8 / 8"},
		{"unbounded_nil_max", 5, nil, "5 / ∞"},
		{"unbounded_zero_current", 0, nil, "0 / ∞"},
		{"bounded_large", 1000, intPtr(9999), "1000 / 9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CapacityRatio(tt.current, tt.max)
			if got != tt.want {
				t.Errorf("CapacityRatio(%d, %v) = %q, want %q", tt.current, tt.max, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// AttestationFingerprint
// ---------------------------------------------------------------------------

func TestAttestationFingerprint(t *testing.T) {
	tests := []struct {
		name        string
		fingerprint string
		want        string
	}{
		// Empty → sentinel
		{"empty", "", "--"},

		// No colon → passthrough
		{"no_colon", "nocolon", "nocolon"},

		// Short hash (≤ 11 chars) → no truncation
		{"short_hash_exact_11", "ed25519:12345678901", "ed25519:12345678901"},
		{"short_hash_under_11", "rsa:abc123", "rsa:abc123"},

		// Long hash → truncated
		{"ed25519_long", "ed25519:abc1234abcdd4f2", "ed25519:abc1234…d4f2"},
		{"sha256_long", "sha256:abcdef1234567890abcdef1234567890abcdef12", "sha256:abcdef1…ef12"},

		// Empty hash after colon
		{"empty_hash", "ed25519:", "ed25519:"},

		// Algorithm only, colon at end
		{"algo_colon_short", "rsa:12345678901", "rsa:12345678901"},
		{"algo_colon_long", "rsa:12345678901x", "rsa:1234567…901x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AttestationFingerprint(tt.fingerprint)
			if got != tt.want {
				t.Errorf("AttestationFingerprint(%q) = %q, want %q", tt.fingerprint, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// RegionList
// ---------------------------------------------------------------------------

func TestRegionList(t *testing.T) {
	tests := []struct {
		name       string
		regions    []string
		maxVisible int
		want       string
	}{
		// Empty / nil → sentinel
		{"nil_regions", nil, 3, "--"},
		{"empty_slice", []string{}, 3, "--"},

		// Fits within maxVisible
		{"one_region", []string{"iad1"}, 3, "iad1"},
		{"two_regions_fits", []string{"iad1", "ord1"}, 3, "iad1, ord1"},
		{"three_regions_exact", []string{"iad1", "ord1", "sea1"}, 3, "iad1, ord1, sea1"},

		// Overflow
		{"four_regions_max1", []string{"iad1", "ord1", "sea1", "dfw1"}, 1, "iad1, +3 more"},
		{"four_regions_max3", []string{"iad1", "ord1", "sea1", "dfw1"}, 3, "iad1, ord1, sea1, +1 more"},
		{"five_regions_max2", []string{"iad1", "ord1", "sea1", "dfw1", "lax1"}, 2, "iad1, ord1, +3 more"},

		// maxVisible < 1 → treated as 1
		{"max_zero_treated_as_one", []string{"iad1", "ord1"}, 0, "iad1, +1 more"},
		{"max_negative_treated_as_one", []string{"iad1", "ord1"}, -5, "iad1, +1 more"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RegionList(tt.regions, tt.maxVisible)
			if got != tt.want {
				t.Errorf("RegionList(%v, %d) = %q, want %q", tt.regions, tt.maxVisible, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ToolchainSpec
// ---------------------------------------------------------------------------

func TestToolchainSpec(t *testing.T) {
	tests := []struct {
		name       string
		toolchains map[string]string
		want       string
	}{
		// Empty / nil → sentinel
		{"nil_map", nil, "--"},
		{"empty_map", map[string]string{}, "--"},

		// Single entry
		{"single_java", map[string]string{"java": "17"}, "java=17"},
		{"single_node", map[string]string{"node": "20"}, "node=20"},

		// Multiple entries — sorted by name
		{"java_node", map[string]string{"java": "17", "node": "20"}, "java=17, node=20"},
		{"node_java_order", map[string]string{"node": "20", "java": "17"}, "java=17, node=20"},
		{"three_tools", map[string]string{"rust": "1.78", "go": "1.22", "python": "3.12"}, "go=1.22, python=3.12, rust=1.78"},

		// Values with dots
		{"version_with_dots", map[string]string{"node": "20.x"}, "node=20.x"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToolchainSpec(tt.toolchains)
			if got != tt.want {
				t.Errorf("ToolchainSpec(%v) = %q, want %q", tt.toolchains, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// HumanLabel
// ---------------------------------------------------------------------------

// billingModel is a typed string alias used to exercise HumanLabel[T].
type billingModel string

func TestHumanLabel(t *testing.T) {
	labels := map[billingModel]string{
		"wall-clock": "Wall-clock time",
		"active-cpu": "Active CPU only",
		"invocation": "Per invocation",
	}

	tests := []struct {
		name  string
		value billingModel
		want  string
	}{
		{"known_wall_clock", "wall-clock", "Wall-clock time"},
		{"known_active_cpu", "active-cpu", "Active CPU only"},
		{"known_invocation", "invocation", "Per invocation"},
		{"unknown_returns_sentinel", "fixed", "--"},
		{"empty_string_unknown", "", "--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HumanLabel(tt.value, labels)
			if got != tt.want {
				t.Errorf("HumanLabel(%q, labels) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestHumanLabelNilMap verifies that a nil labels map returns the sentinel
// without panicking.
func TestHumanLabelNilMap(t *testing.T) {
	got := HumanLabel("wall-clock", (map[billingModel]string)(nil))
	if got != "--" {
		t.Errorf("HumanLabel(nil map) = %q, want %q", got, "--")
	}
}

// TestHumanLabelIntKey exercises HumanLabel with an integer key type to
// verify the generic constraint works beyond string aliases.
func TestHumanLabelIntKey(t *testing.T) {
	type priority int
	labels := map[priority]string{
		1: "Low",
		2: "Medium",
		3: "High",
	}
	got := HumanLabel(priority(2), labels)
	if got != "Medium" {
		t.Errorf("HumanLabel(2, labels) = %q, want %q", got, "Medium")
	}
	got = HumanLabel(priority(99), labels)
	if got != "--" {
		t.Errorf("HumanLabel(99, labels) = %q, want %q", got, "--")
	}
}

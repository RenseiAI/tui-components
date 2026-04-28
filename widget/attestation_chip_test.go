package widget

import (
	"strings"
	"testing"

	"github.com/RenseiAI/tui-components/theme"
)

func TestAttestationChip_NoColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		state       AttestationState
		fingerprint string
		wantPrefix  string
	}{
		{"verified with fp", AttestationVerified, "ed25519:abcd1234efgh5678", "✓ verified"},
		{"signed no fp", AttestationSigned, "", "~ signed"},
		{"unsigned", AttestationUnsigned, "", "✗ unsigned"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			chip := NewAttestationChip(
				WithAttestationState(tc.state),
				WithAttestationFingerprint(tc.fingerprint),
				WithAttestationNoColor(true),
			)
			got := chip.ViewString()
			if !strings.HasPrefix(got, tc.wantPrefix) {
				t.Errorf("ViewString() = %q, want prefix %q", got, tc.wantPrefix)
			}
		})
	}
}

func TestTruncateFingerprint(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{"ed25519:abcd1234efgh5678", "ed25519:abcd…5678"},
		{"short", "short"},
		{"ed25519:ab", "ed25519:ab"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.in, func(t *testing.T) {
			t.Parallel()
			got := truncateFingerprint(tc.in)
			if got != tc.want {
				t.Errorf("truncateFingerprint(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestAttestationChip_AccessibleLabel(t *testing.T) {
	t.Parallel()
	chip := NewAttestationChip(
		WithAttestationState(AttestationVerified),
		WithAttestationFingerprint("ed25519:abc"),
	)
	label := chip.AccessibleLabel()
	if !strings.Contains(label, "verified") {
		t.Errorf("AccessibleLabel() missing state: %q", label)
	}
}

func TestAttestationChip_WithTheme(t *testing.T) {
	t.Parallel()
	chip := NewAttestationChip(
		WithAttestationState(AttestationVerified),
		WithAttestationTheme(theme.DarkTheme()),
	)
	_ = chip.ViewString()
}

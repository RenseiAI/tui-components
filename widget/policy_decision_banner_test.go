package widget

import (
	"strings"
	"testing"
)

func TestPolicyDecisionBanner_NoColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		decision  PolicyDecision
		wantSym   string
		wantBadge string
	}{
		{PolicyAllowed, "✓", "ALLOWED"},
		{PolicyBlocked, "✗", "BLOCKED"},
		{PolicyNeedsApproval, "⚑", "APPROVAL"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(string(tc.decision), func(t *testing.T) {
			t.Parallel()
			b := NewPolicyDecisionBanner(
				WithPolicyDecision(tc.decision),
				WithPolicyBannerNoColor(true),
			)
			got := b.ViewString()
			if !strings.Contains(got, tc.wantSym) {
				t.Errorf("ViewString() missing symbol %q: %q", tc.wantSym, got)
			}
			if !strings.Contains(got, tc.wantBadge) {
				t.Errorf("ViewString() missing badge %q: %q", tc.wantBadge, got)
			}
		})
	}
}

func TestPolicyDecisionBanner_WithDescription(t *testing.T) {
	t.Parallel()
	b := NewPolicyDecisionBanner(
		WithPolicyDecision(PolicyBlocked),
		WithPolicyDescription("cannot write to /etc/"),
		WithPolicyReason("path-allowlist"),
		WithPolicyBannerNoColor(true),
	)
	got := b.ViewString()
	if !strings.Contains(got, "cannot write to /etc/") {
		t.Errorf("ViewString() missing description: %q", got)
	}
	if !strings.Contains(got, "path-allowlist") {
		t.Errorf("ViewString() missing reason: %q", got)
	}
}

func TestPolicyDecisionBanner_AccessibleLabel(t *testing.T) {
	t.Parallel()
	b := NewPolicyDecisionBanner(
		WithPolicyDecision(PolicyNeedsApproval),
		WithPolicyDescription("deploy to prod"),
	)
	label := b.AccessibleLabel()
	if !strings.Contains(label, "needs-approval") {
		t.Errorf("AccessibleLabel() missing decision: %q", label)
	}
	if !strings.Contains(label, "deploy to prod") {
		t.Errorf("AccessibleLabel() missing description: %q", label)
	}
}

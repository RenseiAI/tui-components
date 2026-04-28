package widget

import (
	"strings"
	"testing"
	"time"
)

func TestAuditChain_IntegrityOK(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	e1 := NewAuditEntry(WithAuditEventKind("session.start"), WithAuditTimestamp(ts), WithAuditNoColor(true))
	e2 := NewAuditEntry(WithAuditEventKind("session.end"), WithAuditTimestamp(ts.Add(time.Hour)), WithAuditNoColor(true))

	chain := NewAuditChain(
		WithChainEntries(e1, e2),
		WithChainIntegrity(ChainIntegrityOK),
		WithAuditChainNoColor(true),
	)
	got := chain.ViewString()
	if !strings.Contains(got, "Chain intact") {
		t.Errorf("ViewString() missing 'Chain intact': %q", got)
	}
	if !strings.Contains(got, "session.start") {
		t.Errorf("ViewString() missing first entry: %q", got)
	}
}

func TestAuditChain_IntegrityBroken(t *testing.T) {
	t.Parallel()
	chain := NewAuditChain(
		WithChainIntegrity(ChainIntegrityBroken),
		WithAuditChainNoColor(true),
	)
	got := chain.ViewString()
	if !strings.Contains(got, "Chain broken") {
		t.Errorf("ViewString() missing 'Chain broken': %q", got)
	}
}

func TestAuditChain_IntegrityUnverified(t *testing.T) {
	t.Parallel()
	chain := NewAuditChain(WithAuditChainNoColor(true))
	got := chain.ViewString()
	if !strings.Contains(got, "Chain unverified") {
		t.Errorf("ViewString() missing 'Chain unverified': %q", got)
	}
}

func TestSuffixN(t *testing.T) {
	t.Parallel()
	tests := []struct {
		n    int
		want string
	}{
		{1, "1 event"},
		{2, "2 events"},
		{10, "10 events"},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			got := suffixN(tc.n)
			if got != tc.want {
				t.Errorf("suffixN(%d) = %q, want %q", tc.n, got, tc.want)
			}
		})
	}
}

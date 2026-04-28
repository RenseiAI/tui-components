package widget

import (
	"strings"
	"testing"
	"time"
)

func TestAuditEntry_NoColor(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 28, 14, 23, 1, 0, time.UTC)
	e := NewAuditEntry(
		WithAuditEventKind("session.start"),
		WithAuditActor("worker-01"),
		WithAuditTimestamp(ts),
		WithAuditAttestation(AttestationVerified),
		WithAuditFingerprint("ed25519:abc1234567890def"),
		WithAuditNoColor(true),
	)
	got := e.ViewString()
	if !strings.Contains(got, "session.start") {
		t.Errorf("ViewString() missing event kind: %q", got)
	}
	if !strings.Contains(got, "worker-01") {
		t.Errorf("ViewString() missing actor: %q", got)
	}
	if !strings.Contains(got, "2026-04-28") {
		t.Errorf("ViewString() missing timestamp: %q", got)
	}
}

func TestAuditEntry_AccessibleLabel(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC)
	e := NewAuditEntry(
		WithAuditEventKind("commit.push"),
		WithAuditActor("agent-1"),
		WithAuditTimestamp(ts),
		WithAuditAttestation(AttestationSigned),
	)
	label := e.AccessibleLabel()
	if !strings.Contains(label, "commit.push") {
		t.Errorf("AccessibleLabel() missing event kind: %q", label)
	}
}

func TestAuditEntry_ZeroTimestamp(t *testing.T) {
	t.Parallel()
	e := NewAuditEntry(
		WithAuditEventKind("kit.detect"),
		WithAuditNoColor(true),
	)
	got := e.ViewString()
	if !strings.Contains(got, "—") {
		t.Errorf("ViewString() missing '—' for zero timestamp: %q", got)
	}
}

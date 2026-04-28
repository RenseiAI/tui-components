package widget

import (
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *AuditEntry satisfies component.Component.
var _ component.Component = (*AuditEntry)(nil)

// AuditEntry renders a single signed audit event row with attestation state
// and timestamp.  It is the building block for AuditChain.
// Per Layer 6 (Policy/Security/Observability) in 001 and
// 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	✓  2026-04-28 14:23:01  session.start   worker-01  ed25519:abc1…d4f2
type AuditEntry struct {
	// EventKind is the event type identifier (e.g. "session.start",
	// "commit.push", "kit.detect").
	EventKind string
	// Actor is the entity that triggered the event (worker ID, user ID, etc.).
	Actor string
	// Timestamp is when the event occurred.
	Timestamp time.Time
	// Attestation is the signing state of this event.
	Attestation AttestationState
	// Fingerprint is the optional key fingerprint for the attestation.
	Fingerprint string
	// AccessibleLabel provides a screen-reader-friendly description.
	AccessLabel string
	t           theme.Theme
	noColor     bool
}

// AuditEntryOption configures an AuditEntry during construction.
type AuditEntryOption func(*AuditEntry)

// WithAuditEventKind sets the event kind label.
func WithAuditEventKind(kind string) AuditEntryOption {
	return func(e *AuditEntry) { e.EventKind = kind }
}

// WithAuditActor sets the actor label.
func WithAuditActor(actor string) AuditEntryOption {
	return func(e *AuditEntry) { e.Actor = actor }
}

// WithAuditTimestamp sets the event timestamp.
func WithAuditTimestamp(t time.Time) AuditEntryOption {
	return func(e *AuditEntry) { e.Timestamp = t }
}

// WithAuditAttestation sets the attestation state.
func WithAuditAttestation(a AttestationState) AuditEntryOption {
	return func(e *AuditEntry) { e.Attestation = a }
}

// WithAuditFingerprint sets the key fingerprint.
func WithAuditFingerprint(fp string) AuditEntryOption {
	return func(e *AuditEntry) { e.Fingerprint = fp }
}

// WithAuditTheme sets the Theme.
func WithAuditTheme(t theme.Theme) AuditEntryOption {
	return func(e *AuditEntry) { e.t = t }
}

// WithAuditNoColor forces symbol-first, no-ANSI rendering.
func WithAuditNoColor(nc bool) AuditEntryOption {
	return func(e *AuditEntry) { e.noColor = nc }
}

// WithAuditAccessibleLabel sets the accessible label.
func WithAuditAccessibleLabel(label string) AuditEntryOption {
	return func(e *AuditEntry) { e.AccessLabel = label }
}

// NewAuditEntry constructs an AuditEntry with default theme.
func NewAuditEntry(opts ...AuditEntryOption) *AuditEntry {
	e := &AuditEntry{
		Attestation: AttestationUnsigned,
		t:           theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// SetTheme updates the theme.
func (e *AuditEntry) SetTheme(t theme.Theme) { e.t = t }

// AccessibleLabel returns the accessible label.
func (e *AuditEntry) AccessibleLabel() string {
	if e.AccessLabel != "" {
		return e.AccessLabel
	}
	ts := ""
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.UTC().Format(time.RFC3339)
	}
	return fmt.Sprintf("audit: %s by %s at %s (%s)", e.EventKind, e.Actor, ts, e.Attestation)
}

// ViewString renders the entry as a plain string.
func (e *AuditEntry) ViewString() string {
	chip := NewAttestationChip(
		WithAttestationState(e.Attestation),
		WithAttestationFingerprint(e.Fingerprint),
		WithAttestationTheme(e.t),
		WithAttestationNoColor(e.noColor),
	)
	attestStr := chip.ViewString()

	ts := "—"
	if !e.Timestamp.IsZero() {
		ts = e.Timestamp.UTC().Format("2006-01-02 15:04:05")
	}
	actor := e.Actor
	if actor == "" {
		actor = "—"
	}

	if e.noColor {
		return fmt.Sprintf("%-24s  %-20s  %-12s  %s",
			ts, e.EventKind, actor, attestStr)
	}

	tsStr := lipgloss.NewStyle().Foreground(e.t.TextTertiary).Render(fmt.Sprintf("%-24s", ts))
	kindStr := lipgloss.NewStyle().Foreground(e.t.TextPrimary).Bold(true).Render(fmt.Sprintf("%-20s", e.EventKind))
	actorStr := lipgloss.NewStyle().Foreground(e.t.TextSecondary).Render(fmt.Sprintf("%-12s", actor))
	return fmt.Sprintf("%s  %s  %s  %s", tsStr, kindStr, actorStr, attestStr)
}

// Init satisfies tea.Model.
func (e *AuditEntry) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (e *AuditEntry) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return e, nil }

// View renders the entry as a tea.View.
func (e *AuditEntry) View() tea.View { return tea.NewView(e.ViewString()) }

// SetSize is a no-op.
func (e *AuditEntry) SetSize(width, height int) {}

// Focus is a no-op.
func (e *AuditEntry) Focus() {}

// Blur is a no-op.
func (e *AuditEntry) Blur() {}

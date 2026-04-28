package widget

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *AuditChain satisfies component.Component.
var _ component.Component = (*AuditChain)(nil)

// ChainIntegrity represents whether the audit chain's hash chain is intact.
type ChainIntegrity string

const (
	// ChainIntegrityOK means all entries hash-chain correctly.
	ChainIntegrityOK ChainIntegrity = "ok"
	// ChainIntegrityBroken means at least one hash-chain link is broken.
	ChainIntegrityBroken ChainIntegrity = "broken"
	// ChainIntegrityUnverified means chain integrity has not been checked.
	ChainIntegrityUnverified ChainIntegrity = "unverified"
)

// AuditChain renders a composed list of AuditEntry rows along with a chain
// integrity indicator.  Per Layer 6 / 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	✓ Chain intact (3 events)
//	2026-04-28 14:23:01  session.start   worker-01  ✓ verified  ed25519:abc1…d4f2
//	2026-04-28 14:23:12  kit.detect      worker-01  ✓ verified  ed25519:abc1…d4f2
//	2026-04-28 14:45:30  session.end     worker-01  ✓ verified  ed25519:abc1…d4f2
type AuditChain struct {
	entries   []*AuditEntry
	integrity ChainIntegrity
	t         theme.Theme
	noColor   bool
	width     int
}

// AuditChainOption configures an AuditChain during construction.
type AuditChainOption func(*AuditChain)

// WithChainEntries sets the AuditEntry rows.
func WithChainEntries(entries ...*AuditEntry) AuditChainOption {
	return func(c *AuditChain) { c.entries = entries }
}

// WithChainIntegrity sets the chain integrity state.
func WithChainIntegrity(i ChainIntegrity) AuditChainOption {
	return func(c *AuditChain) { c.integrity = i }
}

// WithAuditChainTheme sets the Theme.
func WithAuditChainTheme(t theme.Theme) AuditChainOption {
	return func(c *AuditChain) {
		c.t = t
		for _, e := range c.entries {
			e.SetTheme(t)
		}
	}
}

// WithAuditChainNoColor forces symbol-first, no-ANSI rendering.
func WithAuditChainNoColor(nc bool) AuditChainOption {
	return func(c *AuditChain) {
		c.noColor = nc
		for _, e := range c.entries {
			e.noColor = nc
		}
	}
}

// NewAuditChain constructs an AuditChain with default theme.
func NewAuditChain(opts ...AuditChainOption) *AuditChain {
	c := &AuditChain{
		integrity: ChainIntegrityUnverified,
		t:         theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetTheme updates the theme for the chain and all entries.
func (c *AuditChain) SetTheme(t theme.Theme) {
	c.t = t
	for _, e := range c.entries {
		e.SetTheme(t)
	}
}

// integrityHeader renders the chain integrity badge.
func (c *AuditChain) integrityHeader() string {
	count := len(c.entries)
	switch c.integrity {
	case ChainIntegrityOK:
		msg := "Chain intact"
		if count > 0 {
			msg = "Chain intact"
		}
		suffix := ""
		if count > 0 {
			suffix = " (" + strings.TrimSpace(strings.TrimRight(suffixN(count), "\n")) + ")"
		}
		if c.noColor {
			return "✓ " + msg + suffix
		}
		return lipgloss.NewStyle().Foreground(c.t.StatusSuccess).Render("✓ "+msg) +
			lipgloss.NewStyle().Foreground(c.t.TextTertiary).Render(suffix)
	case ChainIntegrityBroken:
		msg := "Chain broken"
		suffix := ""
		if count > 0 {
			suffix = " (" + suffixN(count) + ")"
		}
		if c.noColor {
			return "✗ " + msg + suffix
		}
		return lipgloss.NewStyle().Foreground(c.t.StatusError).Render("✗ "+msg) +
			lipgloss.NewStyle().Foreground(c.t.TextTertiary).Render(suffix)
	default:
		msg := "Chain unverified"
		suffix := ""
		if count > 0 {
			suffix = " (" + suffixN(count) + ")"
		}
		if c.noColor {
			return "? " + msg + suffix
		}
		return lipgloss.NewStyle().Foreground(c.t.StatusWarning).Render("? "+msg) +
			lipgloss.NewStyle().Foreground(c.t.TextTertiary).Render(suffix)
	}
}

func suffixN(n int) string {
	if n == 1 {
		return "1 event"
	}
	return fmt.Sprintf("%d events", n)
}

// ViewString renders the chain as a plain string.
func (c *AuditChain) ViewString() string {
	var sb strings.Builder
	sb.WriteString(c.integrityHeader() + "\n")
	for _, e := range c.entries {
		sb.WriteString(e.ViewString() + "\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

// Init satisfies tea.Model.
func (c *AuditChain) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (c *AuditChain) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return c, nil }

// View renders the chain as a tea.View.
func (c *AuditChain) View() tea.View { return tea.NewView(c.ViewString()) }

// SetSize stores size hints.
func (c *AuditChain) SetSize(width, height int) { c.width = width }

// Focus is a no-op.
func (c *AuditChain) Focus() {}

// Blur is a no-op.
func (c *AuditChain) Blur() {}

package widget

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *PolicyDecisionBanner satisfies component.Component.
var _ component.Component = (*PolicyDecisionBanner)(nil)

// PolicyDecision represents the outcome of a policy evaluation.
// Per Layer 6 / 014-tui-operator-surfaces.md.
type PolicyDecision string

const (
	// PolicyAllowed indicates the action is permitted by policy.
	PolicyAllowed PolicyDecision = "allowed"
	// PolicyBlocked indicates the action is denied by policy.
	PolicyBlocked PolicyDecision = "blocked"
	// PolicyNeedsApproval indicates the action requires explicit approval.
	PolicyNeedsApproval PolicyDecision = "needs-approval"
)

// PolicyDecisionBanner renders a prominent banner conveying the outcome of a
// policy evaluation — allowed, blocked, or needs-approval — with an optional
// override actor and reason.
//
// Example renderings:
//
//	✓ ALLOWED    deploy to production
//	✗ BLOCKED    agent cannot write to /etc/  [policy: path-allowlist]
//	⚑ APPROVAL   deployment requires QA sign-off  [requested from: qa-team]
type PolicyDecisionBanner struct {
	decision    PolicyDecision
	description string
	actor       string // override actor or approver
	reason      string // policy name, override reason, etc.
	accessLabel string
	t           theme.Theme
	noColor     bool
	width       int
}

// PolicyDecisionBannerOption configures a PolicyDecisionBanner.
type PolicyDecisionBannerOption func(*PolicyDecisionBanner)

// WithPolicyDecision sets the decision outcome.
func WithPolicyDecision(d PolicyDecision) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.decision = d }
}

// WithPolicyDescription sets the human-readable action description.
func WithPolicyDescription(desc string) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.description = desc }
}

// WithPolicyActor sets the actor label (override actor, approver, etc.).
func WithPolicyActor(actor string) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.actor = actor }
}

// WithPolicyReason sets the policy rule or override reason.
func WithPolicyReason(reason string) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.reason = reason }
}

// WithPolicyBannerTheme sets the Theme.
func WithPolicyBannerTheme(t theme.Theme) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.t = t }
}

// WithPolicyBannerNoColor forces symbol-first, no-ANSI rendering.
func WithPolicyBannerNoColor(nc bool) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.noColor = nc }
}

// WithPolicyBannerAccessibleLabel sets the accessible label.
func WithPolicyBannerAccessibleLabel(label string) PolicyDecisionBannerOption {
	return func(b *PolicyDecisionBanner) { b.accessLabel = label }
}

// NewPolicyDecisionBanner constructs a PolicyDecisionBanner with default theme.
func NewPolicyDecisionBanner(opts ...PolicyDecisionBannerOption) *PolicyDecisionBanner {
	b := &PolicyDecisionBanner{
		decision: PolicyAllowed,
		t:        theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// SetTheme updates the theme.
func (b *PolicyDecisionBanner) SetTheme(t theme.Theme) { b.t = t }

// AccessibleLabel returns the accessible label.
func (b *PolicyDecisionBanner) AccessibleLabel() string {
	if b.accessLabel != "" {
		return b.accessLabel
	}
	label := "policy decision: " + string(b.decision)
	if b.description != "" {
		label += " — " + b.description
	}
	if b.reason != "" {
		label += " [" + b.reason + "]"
	}
	return label
}

type policyVisual struct {
	symbol string
	badge  string
	color  color.Color
}

func (b *PolicyDecisionBanner) visual() policyVisual {
	switch b.decision {
	case PolicyBlocked:
		return policyVisual{"✗", "BLOCKED", b.t.StatusError}
	case PolicyNeedsApproval:
		return policyVisual{"⚑", "APPROVAL", b.t.StatusWarning}
	default:
		return policyVisual{"✓", "ALLOWED", b.t.StatusSuccess}
	}
}

// ViewString renders the banner as a plain string.
func (b *PolicyDecisionBanner) ViewString() string {
	v := b.visual()
	meta := ""
	if b.actor != "" && b.reason != "" {
		meta = "  [" + b.reason + " / " + b.actor + "]"
	} else if b.reason != "" {
		meta = "  [" + b.reason + "]"
	} else if b.actor != "" {
		meta = "  [" + b.actor + "]"
	}

	desc := b.description
	if desc == "" {
		desc = ""
	}

	if b.noColor {
		line := v.symbol + " " + v.badge
		if desc != "" {
			line += "    " + desc
		}
		line += meta
		return line
	}

	badgeStyle := lipgloss.NewStyle().Foreground(v.color).Bold(true)
	badge := badgeStyle.Render(v.symbol + " " + v.badge)
	descStr := ""
	if desc != "" {
		descStr = "    " + lipgloss.NewStyle().Foreground(b.t.TextPrimary).Render(desc)
	}
	metaStr := ""
	if meta != "" {
		metaStr = lipgloss.NewStyle().Foreground(b.t.TextTertiary).Render(meta)
	}
	return badge + descStr + metaStr
}

// Init satisfies tea.Model.
func (b *PolicyDecisionBanner) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (b *PolicyDecisionBanner) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return b, nil }

// View renders the banner as a tea.View.
func (b *PolicyDecisionBanner) View() tea.View { return tea.NewView(b.ViewString()) }

// SetSize stores width hint.
func (b *PolicyDecisionBanner) SetSize(width, height int) { b.width = width }

// Focus is a no-op.
func (b *PolicyDecisionBanner) Focus() {}

// Blur is a no-op.
func (b *PolicyDecisionBanner) Blur() {}

package widget

import (
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *ScopePill satisfies component.Component.
var _ component.Component = (*ScopePill)(nil)

// Scope represents the resolution scope of a provider or plugin as defined
// in 002-provider-base-contract.md (project → org → tenant → global).
type Scope string

const (
	// ScopeProject is the narrowest scope — applies to a single project.
	ScopeProject Scope = "project"
	// ScopeOrg applies across all projects within an organization.
	ScopeOrg Scope = "org"
	// ScopeTenant applies across all orgs within a tenant.
	ScopeTenant Scope = "tenant"
	// ScopeGlobal is the widest scope — platform-wide.
	ScopeGlobal Scope = "global"
)

// ScopePill renders a provider or plugin resolution scope as a small colored
// pill badge.  The pill uses distinct colors for each scope level:
//   - project  → blue (narrowest)
//   - org      → teal
//   - tenant   → accent/orange
//   - global   → yellow (widest)
//
// In no-color mode the pill renders as plain bracketed text: [project].
type ScopePill struct {
	scope       Scope
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// ScopePillOption configures a ScopePill during construction.
type ScopePillOption func(*ScopePill)

// WithScopeValue sets the Scope value to display.
func WithScopeValue(s Scope) ScopePillOption {
	return func(p *ScopePill) { p.scope = s }
}

// WithScopeTheme sets the Theme used for rendering.
func WithScopeTheme(t theme.Theme) ScopePillOption {
	return func(p *ScopePill) { p.t = t }
}

// WithScopeNoColor forces symbol-first, no-ANSI rendering.
func WithScopeNoColor(nc bool) ScopePillOption {
	return func(p *ScopePill) { p.noColor = nc }
}

// WithScopeAccessibleLabel sets the accessible label for screen-reader
// consumers.
func WithScopeAccessibleLabel(label string) ScopePillOption {
	return func(p *ScopePill) { p.accessLabel = label }
}

// NewScopePill constructs a ScopePill with the default theme.
func NewScopePill(opts ...ScopePillOption) *ScopePill {
	p := &ScopePill{
		scope: ScopeProject,
		t:     theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// SetTheme updates the theme used for rendering.
func (p *ScopePill) SetTheme(t theme.Theme) { p.t = t }

// AccessibleLabel returns the accessible label.  Falls back to the scope
// string if no explicit label was set.
func (p *ScopePill) AccessibleLabel() string {
	if p.accessLabel != "" {
		return p.accessLabel
	}
	return "scope: " + string(p.scope)
}

// scopeColor returns the foreground color for the given scope.
func (p *ScopePill) scopeColor() color.Color {
	switch p.scope {
	case ScopeOrg:
		return p.t.Teal
	case ScopeTenant:
		return p.t.Accent
	case ScopeGlobal:
		return p.t.StatusWarning
	default: // project
		return p.t.Blue
	}
}

// ViewString renders the pill as a plain string.
func (p *ScopePill) ViewString() string {
	label := string(p.scope)
	if p.noColor {
		return "[" + label + "]"
	}
	style := lipgloss.NewStyle().
		Foreground(p.scopeColor()).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.scopeColor()).
		Padding(0, 1)
	return style.Render(label)
}

// Init satisfies tea.Model.
func (p *ScopePill) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (p *ScopePill) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }

// View renders the pill as a tea.View.
func (p *ScopePill) View() tea.View { return tea.NewView(p.ViewString()) }

// SetSize is a no-op.
func (p *ScopePill) SetSize(width, height int) {}

// Focus is a no-op.
func (p *ScopePill) Focus() {}

// Blur is a no-op.
func (p *ScopePill) Blur() {}

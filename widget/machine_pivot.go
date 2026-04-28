package widget

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *MachinePivot satisfies component.Component.
var _ component.Component = (*MachinePivot)(nil)

// MachineSummary represents an aggregated snapshot of a single machine (daemon)
// in a multi-machine SaaS fleet, per 013-orchestrator-and-governor.md and
// 014-tui-operator-surfaces.md.
type MachineSummary struct {
	// ID is the machine or daemon identifier.
	ID string
	// Workers is the total number of workers registered on this machine.
	Workers int
	// ActiveWorkers is the number currently executing a session.
	ActiveWorkers int
	// Region is the geographic region or datacenter label.
	Region string
	// Health is the aggregate health of this machine's workers.
	Health ProviderHealth
}

// MachinePivot renders a multi-machine breakdown suitable for the SaaS
// operator surface.  Each machine is shown as a summary row with worker
// counts, region, and aggregate health dot.
//
// Example rendering:
//
//	ID               Region  Workers  Active  Health
//	machine-01       iad1    4        3       ● ready
//	machine-02       sfo3    2        0       ◌ idle
type MachinePivot struct {
	machines []MachineSummary
	t        theme.Theme
	noColor  bool
	width    int
}

// MachinePivotOption configures a MachinePivot during construction.
type MachinePivotOption func(*MachinePivot)

// WithMachines sets the machines to display.
func WithMachines(machines ...MachineSummary) MachinePivotOption {
	return func(p *MachinePivot) { p.machines = machines }
}

// WithMachinePivotTheme sets the Theme.
func WithMachinePivotTheme(t theme.Theme) MachinePivotOption {
	return func(p *MachinePivot) { p.t = t }
}

// WithMachinePivotNoColor forces symbol-first, no-ANSI rendering.
func WithMachinePivotNoColor(nc bool) MachinePivotOption {
	return func(p *MachinePivot) { p.noColor = nc }
}

// NewMachinePivot constructs a MachinePivot with default theme.
func NewMachinePivot(opts ...MachinePivotOption) *MachinePivot {
	p := &MachinePivot{t: theme.DefaultTheme()}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// SetTheme updates the theme.
func (p *MachinePivot) SetTheme(t theme.Theme) { p.t = t }

// SetMachines replaces the current machine list.
func (p *MachinePivot) SetMachines(machines []MachineSummary) { p.machines = machines }

// ViewString renders the pivot table as a plain string.
func (p *MachinePivot) ViewString() string {
	if len(p.machines) == 0 {
		if p.noColor {
			return "(no machines)"
		}
		return lipgloss.NewStyle().Foreground(p.t.TextTertiary).Render("(no machines)")
	}

	var sb strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().Foreground(p.t.TextTertiary).Bold(true)
	if p.noColor {
		fmt.Fprintf(&sb, "%-20s %-8s %-8s %-8s %s\n",
			"ID", "Region", "Workers", "Active", "Health")
	} else {
		sb.WriteString(headerStyle.Render(fmt.Sprintf("%-20s %-8s %-8s %-8s %s",
			"ID", "Region", "Workers", "Active", "Health")) + "\n")
	}

	for _, m := range p.machines {
		dot := NewProviderHealthDot(
			WithProviderHealth(m.Health),
			WithProviderHealthShowLabel(true),
			WithProviderHealthTheme(p.t),
			WithProviderHealthNoColor(p.noColor),
		)
		healthStr := dot.ViewString()

		if p.noColor {
			fmt.Fprintf(&sb, "%-20s %-8s %-8d %-8d %s\n",
				m.ID, m.Region, m.Workers, m.ActiveWorkers, healthStr)
		} else {
			id := lipgloss.NewStyle().Foreground(p.t.TextPrimary).Render(fmt.Sprintf("%-20s", m.ID))
			region := lipgloss.NewStyle().Foreground(p.t.TextSecondary).Render(fmt.Sprintf("%-8s", m.Region))
			workers := lipgloss.NewStyle().Foreground(p.t.TextPrimary).Render(fmt.Sprintf("%-8d", m.Workers))
			active := lipgloss.NewStyle().Foreground(p.t.Accent).Render(fmt.Sprintf("%-8d", m.ActiveWorkers))
			fmt.Fprintf(&sb, "%s %s %s %s %s\n", id, region, workers, active, healthStr)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Init satisfies tea.Model.
func (p *MachinePivot) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (p *MachinePivot) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }

// View renders the pivot as a tea.View.
func (p *MachinePivot) View() tea.View { return tea.NewView(p.ViewString()) }

// SetSize stores the width hint.
func (p *MachinePivot) SetSize(width, height int) { p.width = width }

// Focus is a no-op.
func (p *MachinePivot) Focus() {}

// Blur is a no-op.
func (p *MachinePivot) Blur() {}

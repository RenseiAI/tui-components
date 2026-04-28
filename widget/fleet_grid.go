package widget

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *FleetGrid satisfies component.Component.
var _ component.Component = (*FleetGrid)(nil)

// FleetWorker is the data record for a single worker entry in the FleetGrid.
// It mirrors the fields of WorkerRow to allow callers to build a grid from
// data without constructing individual WorkerRow instances first.
type FleetWorker struct {
	ID           string
	MachineGroup string // machine/daemon grouping label
	Status       WorkerStatus
	Region       string
	LoadFraction float64
	BillingModel string
}

// FleetGrid renders a multi-worker fleet view grouped by machine/daemon as
// described in 013-orchestrator-and-governor.md and
// 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	machine-01
//	  ● busy   iad1   ████████  99%   active-cpu
//	  ◌ idle   iad1   ░░░░░░░░   0%   active-cpu
//	machine-02
//	  ◌ idle   sfo3   ░░░░░░░░   0%   wall-clock
type FleetGrid struct {
	workers []FleetWorker
	t       theme.Theme
	noColor bool
	width   int
	height  int
}

// FleetGridOption configures a FleetGrid during construction.
type FleetGridOption func(*FleetGrid)

// WithFleetWorkers sets the list of workers to display.
func WithFleetWorkers(workers ...FleetWorker) FleetGridOption {
	return func(g *FleetGrid) { g.workers = workers }
}

// WithFleetTheme sets the Theme.
func WithFleetTheme(t theme.Theme) FleetGridOption {
	return func(g *FleetGrid) { g.t = t }
}

// WithFleetNoColor forces symbol-first, no-ANSI rendering.
func WithFleetNoColor(nc bool) FleetGridOption {
	return func(g *FleetGrid) { g.noColor = nc }
}

// NewFleetGrid constructs a FleetGrid with default theme.
func NewFleetGrid(opts ...FleetGridOption) *FleetGrid {
	g := &FleetGrid{t: theme.DefaultTheme()}
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// SetTheme updates the theme.
func (g *FleetGrid) SetTheme(t theme.Theme) { g.t = t }

// SetWorkers replaces the current worker list.
func (g *FleetGrid) SetWorkers(workers []FleetWorker) { g.workers = workers }

// ViewString renders the grid as a plain string.
func (g *FleetGrid) ViewString() string {
	if len(g.workers) == 0 {
		if g.noColor {
			return "(no workers)"
		}
		return lipgloss.NewStyle().Foreground(g.t.TextTertiary).Render("(no workers)")
	}

	// Group workers by MachineGroup preserving insertion order.
	type group struct {
		name    string
		workers []FleetWorker
	}
	seen := map[string]bool{}
	groups := []group{}
	for _, w := range g.workers {
		mg := w.MachineGroup
		if mg == "" {
			mg = "default"
		}
		if !seen[mg] {
			seen[mg] = true
			groups = append(groups, group{name: mg})
		}
		for i := range groups {
			if groups[i].name == mg {
				groups[i].workers = append(groups[i].workers, w)
				break
			}
		}
	}

	var sb strings.Builder
	for gi, grp := range groups {
		if gi > 0 {
			sb.WriteByte('\n')
		}
		if g.noColor {
			sb.WriteString(grp.name + "\n")
		} else {
			header := lipgloss.NewStyle().
				Foreground(g.t.TextPrimary).
				Bold(true).
				Render(grp.name)
			sb.WriteString(header + "\n")
		}
		for _, w := range grp.workers {
			row := NewWorkerRow(
				WithWorkerID(w.ID),
				WithWorkerStatus(w.Status),
				WithWorkerRegion(w.Region),
				WithWorkerLoadFraction(w.LoadFraction),
				WithWorkerBillingModel(w.BillingModel),
				WithWorkerTheme(g.t),
				WithWorkerNoColor(g.noColor),
			)
			sb.WriteString("  " + row.ViewString() + "\n")
		}
	}
	// Trim trailing newline.
	result := sb.String()
	return strings.TrimRight(result, "\n")
}

// Init satisfies tea.Model.
func (g *FleetGrid) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (g *FleetGrid) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return g, nil }

// View renders the grid as a tea.View.
func (g *FleetGrid) View() tea.View { return tea.NewView(g.ViewString()) }

// SetSize stores size hints.
func (g *FleetGrid) SetSize(width, height int) {
	g.width = width
	g.height = height
}

// Focus is a no-op.
func (g *FleetGrid) Focus() {}

// Blur is a no-op.
func (g *FleetGrid) Blur() {}

package widget

import (
	"fmt"
	"image/color"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *WorkerRow satisfies component.Component.
var _ component.Component = (*WorkerRow)(nil)

// WorkerStatus represents the operational status of a worker process as
// described in 013-orchestrator-and-governor.md.
type WorkerStatus string

const (
	// WorkerStatusIdle indicates the worker is connected and waiting for work.
	WorkerStatusIdle WorkerStatus = "idle"
	// WorkerStatusBusy indicates the worker is executing a session.
	WorkerStatusBusy WorkerStatus = "busy"
	// WorkerStatusDraining indicates the worker is finishing its current
	// session and will not accept new work.
	WorkerStatusDraining WorkerStatus = "draining"
	// WorkerStatusOffline indicates the worker is not reachable.
	WorkerStatusOffline WorkerStatus = "offline"
)

// WorkerRow renders a single worker's status, region, load fraction, and
// billing model in a compact single-line format suitable for use inside a
// FleetGrid.  Per 013-orchestrator-and-governor.md and
// 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	● busy   iad1   ████░░░░  75%   active-cpu
type WorkerRow struct {
	id           string
	status       WorkerStatus
	region       string
	loadFraction float64 // 0.0–1.0
	billingModel string  // e.g. "active-cpu", "wall-clock", "invocation"
	accessLabel  string
	t            theme.Theme
	noColor      bool
	width        int
}

// WorkerRowOption configures a WorkerRow during construction.
type WorkerRowOption func(*WorkerRow)

// WithWorkerID sets the worker ID displayed at the start of the row.
func WithWorkerID(id string) WorkerRowOption {
	return func(r *WorkerRow) { r.id = id }
}

// WithWorkerStatus sets the worker operational status.
func WithWorkerStatus(s WorkerStatus) WorkerRowOption {
	return func(r *WorkerRow) { r.status = s }
}

// WithWorkerRegion sets the region label (e.g. "iad1", "sfo3").
func WithWorkerRegion(region string) WorkerRowOption {
	return func(r *WorkerRow) { r.region = region }
}

// WithWorkerLoadFraction sets the current load as a fraction [0.0, 1.0].
func WithWorkerLoadFraction(f float64) WorkerRowOption {
	return func(r *WorkerRow) {
		if f < 0 {
			f = 0
		}
		if f > 1 {
			f = 1
		}
		r.loadFraction = f
	}
}

// WithWorkerBillingModel sets the billing model label (e.g. "active-cpu").
func WithWorkerBillingModel(bm string) WorkerRowOption {
	return func(r *WorkerRow) { r.billingModel = bm }
}

// WithWorkerTheme sets the Theme.
func WithWorkerTheme(t theme.Theme) WorkerRowOption {
	return func(r *WorkerRow) { r.t = t }
}

// WithWorkerNoColor forces symbol-first, no-ANSI rendering.
func WithWorkerNoColor(nc bool) WorkerRowOption {
	return func(r *WorkerRow) { r.noColor = nc }
}

// WithWorkerAccessibleLabel sets the accessible label.
func WithWorkerAccessibleLabel(label string) WorkerRowOption {
	return func(r *WorkerRow) { r.accessLabel = label }
}

// NewWorkerRow constructs a WorkerRow with default theme.
func NewWorkerRow(opts ...WorkerRowOption) *WorkerRow {
	r := &WorkerRow{
		status: WorkerStatusIdle,
		t:      theme.DefaultTheme(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetTheme updates the theme.
func (r *WorkerRow) SetTheme(t theme.Theme) { r.t = t }

// AccessibleLabel returns the accessible label.
func (r *WorkerRow) AccessibleLabel() string {
	if r.accessLabel != "" {
		return r.accessLabel
	}
	return fmt.Sprintf("worker %s: %s in %s load %.0f%%",
		r.id, r.status, r.region, r.loadFraction*100)
}

type workerStatusVisual struct {
	symbol string
	color  color.Color
}

func (r *WorkerRow) statusVisual() workerStatusVisual {
	switch r.status {
	case WorkerStatusBusy:
		return workerStatusVisual{"●", r.t.StatusSuccess}
	case WorkerStatusDraining:
		return workerStatusVisual{"◐", r.t.StatusWarning}
	case WorkerStatusOffline:
		return workerStatusVisual{"○", r.t.StatusError}
	default: // idle
		return workerStatusVisual{"◌", r.t.TextTertiary}
	}
}

// renderLoadBar renders a small ASCII load bar.
func renderLoadBar(frac float64, width int) string {
	filled := int(frac * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

// ViewString renders the row as a plain string.
func (r *WorkerRow) ViewString() string {
	sv := r.statusVisual()
	loadPct := fmt.Sprintf("%3.0f%%", r.loadFraction*100)
	bar := renderLoadBar(r.loadFraction, 8)
	region := r.region
	if region == "" {
		region = "—"
	}
	bm := r.billingModel
	if bm == "" {
		bm = "—"
	}

	if r.noColor {
		return fmt.Sprintf("%s %-8s  %-6s  %s  %s  %s",
			sv.symbol, r.status, region, bar, loadPct, bm)
	}

	sym := lipgloss.NewStyle().Foreground(sv.color).Render(sv.symbol)
	statusStr := lipgloss.NewStyle().Foreground(sv.color).Render(fmt.Sprintf("%-8s", r.status))
	regionStr := lipgloss.NewStyle().Foreground(r.t.TextSecondary).Render(fmt.Sprintf("%-6s", region))
	barStr := lipgloss.NewStyle().Foreground(r.t.Teal).Render(bar)
	pctStr := lipgloss.NewStyle().Foreground(r.t.TextPrimary).Render(loadPct)
	bmStr := lipgloss.NewStyle().Foreground(r.t.TextTertiary).Render(bm)

	return fmt.Sprintf("%s %s  %s  %s  %s  %s", sym, statusStr, regionStr, barStr, pctStr, bmStr)
}

// Init satisfies tea.Model.
func (r *WorkerRow) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (r *WorkerRow) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return r, nil }

// View renders the row as a tea.View.
func (r *WorkerRow) View() tea.View { return tea.NewView(r.ViewString()) }

// SetSize stores the width hint.
func (r *WorkerRow) SetSize(width, height int) { r.width = width }

// Focus is a no-op.
func (r *WorkerRow) Focus() {}

// Blur is a no-op.
func (r *WorkerRow) Blur() {}

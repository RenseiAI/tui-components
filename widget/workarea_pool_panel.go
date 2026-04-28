package widget

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *WorkareaPoolPanel satisfies component.Component.
var _ component.Component = (*WorkareaPoolPanel)(nil)

// WorkareaPoolEntry represents a single pool slot entry keyed by repo and
// toolchain, per 003-workarea-provider.md and 014-tui-operator-surfaces.md.
type WorkareaPoolEntry struct {
	// Repo is the repository identifier (e.g. "github.com/org/repo").
	Repo string
	// Toolchain is the toolchain spec satisfied by this pool slot.
	Toolchain string
	// Warm is the number of pre-warmed, ready-to-use slots.
	Warm int
	// Cold is the number of un-warmed slots available.
	Cold int
	// InUse is the number of slots currently acquired by active sessions.
	InUse int
}

// WorkareaPoolPanel renders the warm / cold / in-use workarea pool breakdown
// per (repo, toolchain) key.  It is the primary display primitive for
// WorkareaProvider pool status as described in 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	Repo                         Toolchain    Warm  Cold  In-Use
//	github.com/org/api           java=17      3     1     2
//	github.com/org/web           node=20.x    2     0     1
type WorkareaPoolPanel struct {
	entries []WorkareaPoolEntry
	t       theme.Theme
	noColor bool
	width   int
}

// WorkareaPoolPanelOption configures a WorkareaPoolPanel during construction.
type WorkareaPoolPanelOption func(*WorkareaPoolPanel)

// WithPoolEntries sets the pool entries to display.
func WithPoolEntries(entries ...WorkareaPoolEntry) WorkareaPoolPanelOption {
	return func(p *WorkareaPoolPanel) { p.entries = entries }
}

// WithPoolTheme sets the Theme.
func WithPoolTheme(t theme.Theme) WorkareaPoolPanelOption {
	return func(p *WorkareaPoolPanel) { p.t = t }
}

// WithPoolNoColor forces symbol-first, no-ANSI rendering.
func WithPoolNoColor(nc bool) WorkareaPoolPanelOption {
	return func(p *WorkareaPoolPanel) { p.noColor = nc }
}

// NewWorkareaPoolPanel constructs a WorkareaPoolPanel with default theme.
func NewWorkareaPoolPanel(opts ...WorkareaPoolPanelOption) *WorkareaPoolPanel {
	p := &WorkareaPoolPanel{t: theme.DefaultTheme()}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// SetTheme updates the theme.
func (p *WorkareaPoolPanel) SetTheme(t theme.Theme) { p.t = t }

// SetEntries replaces the current pool entries.
func (p *WorkareaPoolPanel) SetEntries(entries []WorkareaPoolEntry) { p.entries = entries }

// ViewString renders the panel as a plain string.
func (p *WorkareaPoolPanel) ViewString() string {
	if len(p.entries) == 0 {
		if p.noColor {
			return "(empty pool)"
		}
		return lipgloss.NewStyle().Foreground(p.t.TextTertiary).Render("(empty pool)")
	}

	var sb strings.Builder
	// Header
	if p.noColor {
		fmt.Fprintf(&sb, "%-32s %-14s %-6s %-6s %s\n",
			"Repo", "Toolchain", "Warm", "Cold", "In-Use")
	} else {
		hdr := lipgloss.NewStyle().Foreground(p.t.TextTertiary).Bold(true)
		sb.WriteString(hdr.Render(fmt.Sprintf("%-32s %-14s %-6s %-6s %s",
			"Repo", "Toolchain", "Warm", "Cold", "In-Use")) + "\n")
	}

	for _, e := range p.entries {
		if p.noColor {
			fmt.Fprintf(&sb, "%-32s %-14s %-6d %-6d %d\n",
				e.Repo, e.Toolchain, e.Warm, e.Cold, e.InUse)
		} else {
			repo := lipgloss.NewStyle().Foreground(p.t.TextPrimary).Render(fmt.Sprintf("%-32s", e.Repo))
			tc := lipgloss.NewStyle().Foreground(p.t.Blue).Render(fmt.Sprintf("%-14s", e.Toolchain))
			warm := lipgloss.NewStyle().Foreground(p.t.StatusSuccess).Render(fmt.Sprintf("%-6d", e.Warm))
			cold := lipgloss.NewStyle().Foreground(p.t.TextSecondary).Render(fmt.Sprintf("%-6d", e.Cold))
			inUse := lipgloss.NewStyle().Foreground(p.t.Accent).Render(fmt.Sprintf("%d", e.InUse))
			fmt.Fprintf(&sb, "%s %s %s %s %s\n", repo, tc, warm, cold, inUse)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Init satisfies tea.Model.
func (p *WorkareaPoolPanel) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (p *WorkareaPoolPanel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return p, nil }

// View renders the panel as a tea.View.
func (p *WorkareaPoolPanel) View() tea.View { return tea.NewView(p.ViewString()) }

// SetSize stores the width hint.
func (p *WorkareaPoolPanel) SetSize(width, height int) { p.width = width }

// Focus is a no-op.
func (p *WorkareaPoolPanel) Focus() {}

// Blur is a no-op.
func (p *WorkareaPoolPanel) Blur() {}

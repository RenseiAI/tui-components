package widget

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *KitDetectResult satisfies component.Component.
var _ component.Component = (*KitDetectResult)(nil)

// KitMatch describes a single Kit that matched during the detection phase,
// per 005-kit-manifest-spec.md and 014-tui-operator-surfaces.md.
type KitMatch struct {
	// Name is the kit name (e.g. "spring-java", "nextjs", "go-module").
	Name string
	// Version is the matched kit version.
	Version string
	// Order is the composition order (lower = applied first).
	Order int
	// Conflict is true when this kit has a composition conflict with another.
	Conflict bool
	// ConflictReason describes the conflict if Conflict is true.
	ConflictReason string
}

// KitDetectResult renders the list of kits that matched detection for a given
// session, with ordering and conflict indicators.  Per 005-kit-manifest-spec.md
// and 014-tui-operator-surfaces.md.
//
// Example rendering:
//
//	Kits detected (2)
//	  1  spring-java  v1.2.0
//	  2  nextjs       v0.9.1  ⚠ conflicts with spring-java: conflicting MCP servers
type KitDetectResult struct {
	matches     []KitMatch
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// KitDetectResultOption configures a KitDetectResult during construction.
type KitDetectResultOption func(*KitDetectResult)

// WithKitMatches sets the list of detected kit matches.
func WithKitMatches(matches ...KitMatch) KitDetectResultOption {
	return func(r *KitDetectResult) { r.matches = matches }
}

// WithKitDetectTheme sets the Theme.
func WithKitDetectTheme(t theme.Theme) KitDetectResultOption {
	return func(r *KitDetectResult) { r.t = t }
}

// WithKitDetectNoColor forces symbol-first, no-ANSI rendering.
func WithKitDetectNoColor(nc bool) KitDetectResultOption {
	return func(r *KitDetectResult) { r.noColor = nc }
}

// WithKitDetectAccessibleLabel sets the accessible label.
func WithKitDetectAccessibleLabel(label string) KitDetectResultOption {
	return func(r *KitDetectResult) { r.accessLabel = label }
}

// NewKitDetectResult constructs a KitDetectResult with default theme.
func NewKitDetectResult(opts ...KitDetectResultOption) *KitDetectResult {
	r := &KitDetectResult{t: theme.DefaultTheme()}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// SetTheme updates the theme.
func (r *KitDetectResult) SetTheme(t theme.Theme) { r.t = t }

// AccessibleLabel returns the accessible label.
func (r *KitDetectResult) AccessibleLabel() string {
	if r.accessLabel != "" {
		return r.accessLabel
	}
	names := make([]string, len(r.matches))
	for i, m := range r.matches {
		names[i] = m.Name
	}
	return fmt.Sprintf("kits detected: %s", strings.Join(names, ", "))
}

// ViewString renders the result as a plain string.
func (r *KitDetectResult) ViewString() string {
	if len(r.matches) == 0 {
		if r.noColor {
			return "No kits detected"
		}
		return lipgloss.NewStyle().Foreground(r.t.TextTertiary).Render("No kits detected")
	}

	var sb strings.Builder

	// Header
	header := fmt.Sprintf("Kits detected (%d)", len(r.matches))
	if r.noColor {
		sb.WriteString(header + "\n")
	} else {
		sb.WriteString(lipgloss.NewStyle().Foreground(r.t.TextPrimary).Bold(true).Render(header) + "\n")
	}

	for _, m := range r.matches {
		conflict := ""
		if m.Conflict {
			reason := m.ConflictReason
			if reason == "" {
				reason = "conflict"
			}
			if r.noColor {
				conflict = "  ! " + reason
			} else {
				conflict = "  " + lipgloss.NewStyle().Foreground(r.t.StatusWarning).Render("⚠ "+reason)
			}
		}

		ver := m.Version
		if ver == "" {
			ver = "—"
		}

		if r.noColor {
			fmt.Fprintf(&sb, "  %d  %-20s %s%s\n", m.Order, m.Name, ver, conflict)
		} else {
			orderStr := lipgloss.NewStyle().Foreground(r.t.TextTertiary).Render(fmt.Sprintf("%d", m.Order))
			nameStr := lipgloss.NewStyle().Foreground(r.t.TextPrimary).Bold(true).Render(fmt.Sprintf("%-20s", m.Name))
			verStr := lipgloss.NewStyle().Foreground(r.t.TextSecondary).Render(ver)
			fmt.Fprintf(&sb, "  %s  %s %s%s\n", orderStr, nameStr, verStr, conflict)
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Init satisfies tea.Model.
func (r *KitDetectResult) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (r *KitDetectResult) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return r, nil }

// View renders the result as a tea.View.
func (r *KitDetectResult) View() tea.View { return tea.NewView(r.ViewString()) }

// SetSize is a no-op.
func (r *KitDetectResult) SetSize(width, height int) {}

// Focus is a no-op.
func (r *KitDetectResult) Focus() {}

// Blur is a no-op.
func (r *KitDetectResult) Blur() {}

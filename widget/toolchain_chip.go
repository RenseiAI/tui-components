package widget

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *ToolchainChip satisfies component.Component.
var _ component.Component = (*ToolchainChip)(nil)

// ToolchainSpec represents a single toolchain demand in "name=version" form
// as declared by a Kit or consumed by a WorkareaProvider.
// Per 004-sandbox-capability-matrix.md and 005-kit-manifest-spec.md.
type ToolchainSpec struct {
	// Name is the toolchain name (e.g. "java", "node", "go", "python").
	Name string
	// Version is the required version or range (e.g. "17", "20.x", ">=3.11").
	Version string
}

// String formats the spec as "name=version".
func (s ToolchainSpec) String() string {
	if s.Version == "" {
		return s.Name
	}
	return s.Name + "=" + s.Version
}

// ToolchainChip renders one or more ToolchainSpec values as a compact chip.
// It is the canonical display primitive for Kit toolchain demands and Workarea
// pool state per 014-tui-operator-surfaces.md.
//
// Single spec:   ⚙ java=17
// Multi spec:    ⚙ java=17, node=20.x
type ToolchainChip struct {
	specs       []ToolchainSpec
	accessLabel string
	t           theme.Theme
	noColor     bool
}

// ToolchainChipOption configures a ToolchainChip during construction.
type ToolchainChipOption func(*ToolchainChip)

// WithToolchainSpecs sets the toolchain specs to display.
func WithToolchainSpecs(specs ...ToolchainSpec) ToolchainChipOption {
	return func(c *ToolchainChip) { c.specs = specs }
}

// WithToolchainTheme sets the Theme.
func WithToolchainTheme(t theme.Theme) ToolchainChipOption {
	return func(c *ToolchainChip) { c.t = t }
}

// WithToolchainNoColor forces symbol-first, no-ANSI rendering.
func WithToolchainNoColor(nc bool) ToolchainChipOption {
	return func(c *ToolchainChip) { c.noColor = nc }
}

// WithToolchainAccessibleLabel sets the accessible label.
func WithToolchainAccessibleLabel(label string) ToolchainChipOption {
	return func(c *ToolchainChip) { c.accessLabel = label }
}

// NewToolchainChip constructs a ToolchainChip with the default theme.
func NewToolchainChip(opts ...ToolchainChipOption) *ToolchainChip {
	c := &ToolchainChip{t: theme.DefaultTheme()}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// SetTheme updates the theme.
func (c *ToolchainChip) SetTheme(t theme.Theme) { c.t = t }

// AccessibleLabel returns the accessible label.
func (c *ToolchainChip) AccessibleLabel() string {
	if c.accessLabel != "" {
		return c.accessLabel
	}
	parts := make([]string, len(c.specs))
	for i, s := range c.specs {
		parts[i] = s.String()
	}
	return "toolchain: " + strings.Join(parts, ", ")
}

// ViewString renders the chip as a plain string.
func (c *ToolchainChip) ViewString() string {
	parts := make([]string, len(c.specs))
	for i, s := range c.specs {
		parts[i] = s.String()
	}
	joined := strings.Join(parts, ", ")
	if joined == "" {
		joined = "(no toolchain)"
	}
	if c.noColor {
		return "⚙ " + joined
	}
	sym := lipgloss.NewStyle().Foreground(c.t.Blue).Render("⚙")
	val := lipgloss.NewStyle().Foreground(c.t.TextPrimary).Render(joined)
	return sym + " " + val
}

// Init satisfies tea.Model.
func (c *ToolchainChip) Init() tea.Cmd { return nil }

// Update satisfies tea.Model.
func (c *ToolchainChip) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return c, nil }

// View renders the chip as a tea.View.
func (c *ToolchainChip) View() tea.View { return tea.NewView(c.ViewString()) }

// SetSize is a no-op.
func (c *ToolchainChip) SetSize(width, height int) {}

// Focus is a no-op.
func (c *ToolchainChip) Focus() {}

// Blur is a no-op.
func (c *ToolchainChip) Blur() {}

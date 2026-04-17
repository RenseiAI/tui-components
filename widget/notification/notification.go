package notification

import (
	"image/color"
	"sync/atomic"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *Model satisfies component.Component.
var _ component.Component = (*Model)(nil)

// Variant is the status class of a notification. It is kept as a plain
// string to match the project-wide status-string convention (see
// theme/status.go) and to avoid an import cycle with the theme package.
type Variant string

// Known notification variants.
const (
	// VariantSuccess renders with the theme success color and a check
	// glyph by default.
	VariantSuccess Variant = "success"
	// VariantWarning renders with the theme warning color and a warning
	// glyph by default.
	VariantWarning Variant = "warning"
	// VariantError renders with the theme error color and a cross glyph
	// by default.
	VariantError Variant = "error"
)

// DefaultDuration is the auto-dismiss duration applied when no
// [WithDuration] option is supplied or when the supplied duration is
// non-positive.
const DefaultDuration = 4 * time.Second

const (
	iconSuccess = "\u2713" // ✓
	iconWarning = "\u26A0" // ⚠
	iconError   = "\u2717" // ✗
	iconUnknown = "?"
)

// genCounter produces monotonic ids used to invalidate stale
// auto-dismiss ticks. Shared across all Model instances in the package.
var genCounter atomic.Uint64

// nextGenID returns a fresh, never-zero generation id.
func nextGenID() uint64 {
	return genCounter.Add(1)
}

// dismissMsg is delivered by the auto-dismiss tick. The id field
// identifies the Model instance that scheduled the tick; ticks whose
// id no longer matches the current instance are ignored.
type dismissMsg struct {
	id uint64
}

// Model is a single transient toast notification. It implements
// [component.Component] and is intended to be embedded into a parent
// Bubble Tea model directly or composed via [Stack]. Model is a value
// type; mutating methods such as [Model.Update] return a fresh value.
//
// A Model carries a monotonic generation id assigned at construction.
// The auto-dismiss tick scheduled by [Model.Init] tags itself with that
// id, and [Model.Update] only flips the dismissed flag when the
// incoming tick's id matches. This prevents stale ticks (for example,
// from a Model replaced inside a [Stack]) from racing with newer state.
type Model struct {
	variant   Variant
	message   string
	icon      string
	duration  time.Duration
	width     int
	genID     uint64
	dismissed bool
}

// Option configures a [Model] during construction. Options are applied
// in order by [New].
type Option func(*Model)

// WithDuration overrides the auto-dismiss duration. Non-positive values
// are coerced to [DefaultDuration].
func WithDuration(d time.Duration) Option {
	return func(m *Model) {
		if d <= 0 {
			m.duration = DefaultDuration
			return
		}
		m.duration = d
	}
}

// WithWidth sets the rendered width of the notification box, in cells.
// A width of zero or less causes [Model.View] to return an empty string.
func WithWidth(w int) Option {
	return func(m *Model) {
		m.width = w
	}
}

// WithIcon overrides the default glyph rendered before the message.
// Pass an empty string to suppress the icon entirely.
func WithIcon(s string) Option {
	return func(m *Model) {
		m.icon = s
	}
}

// New constructs a [Model] for the given variant and message. Options
// are applied after the variant defaults are set, so [WithIcon] always
// wins over the variant's default glyph. The returned Model has a
// fresh generation id; calling [Model.Init] schedules the auto-dismiss
// tick tagged with that id.
func New(variant Variant, message string, opts ...Option) Model {
	if !knownVariant(variant) {
		log.Warn("notification: unknown variant", "variant", string(variant))
	}
	m := Model{
		variant:  variant,
		message:  message,
		icon:     defaultIcon(variant),
		duration: DefaultDuration,
		genID:    nextGenID(),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// knownVariant reports whether v is one of the package's defined
// variants.
func knownVariant(v Variant) bool {
	switch v {
	case VariantSuccess, VariantWarning, VariantError:
		return true
	default:
		return false
	}
}

// defaultIcon returns the default glyph for v.
func defaultIcon(v Variant) string {
	switch v {
	case VariantSuccess:
		return iconSuccess
	case VariantWarning:
		return iconWarning
	case VariantError:
		return iconError
	default:
		return iconUnknown
	}
}

// variantColor returns the foreground/border color for v. Unknown
// variants fall back to theme.TextSecondary.
func variantColor(v Variant) color.Color {
	switch v {
	case VariantSuccess:
		return theme.StatusSuccess
	case VariantWarning:
		return theme.StatusWarning
	case VariantError:
		return theme.StatusError
	default:
		return theme.TextSecondary
	}
}

// styleFor returns the lipgloss style used to render a notification of
// the given variant at the given outer width. The style draws a rounded
// border in the variant color and renders the body text in
// theme.TextPrimary. The outer width includes the border and padding;
// see [lipgloss.Style.Width] for exact semantics.
func styleFor(v Variant, outerWidth int) lipgloss.Style {
	c := variantColor(v)
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(c).
		Foreground(theme.TextPrimary).
		Padding(0, 1)
	if outerWidth > 0 {
		s = s.Width(outerWidth)
	}
	return s
}

// Init returns the auto-dismiss tick command. The command emits a
// [dismissMsg] tagged with the Model's current generation id after
// the configured duration has elapsed.
func (m Model) Init() tea.Cmd {
	id := m.genID
	d := m.duration
	return tea.Tick(d, func(time.Time) tea.Msg {
		return dismissMsg{id: id}
	})
}

// Update handles dismiss messages. A [dismissMsg] whose id matches the
// Model's generation id flips [Model.Dismissed] to true; ticks tagged
// with any other id are ignored. All other messages are passed through
// unchanged. Update never schedules additional commands.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if dm, ok := msg.(dismissMsg); ok {
		if dm.id == m.genID {
			m.dismissed = true
		}
	}
	return m, nil
}

// View renders the notification. When the configured width is zero or
// less, or when the notification has been dismissed, View returns an
// empty [tea.View]. Otherwise it returns a rounded-border box styled
// for the variant, containing the icon followed by the message.
//
// The total rendered width — borders and padding included — is the
// configured width. The interior content is wrapped to fit.
func (m Model) View() tea.View {
	if m.width <= 0 || m.dismissed {
		return tea.NewView("")
	}
	body := m.message
	if m.icon != "" {
		body = m.icon + " " + m.message
	}
	return tea.NewView(styleFor(m.variant, m.width).Render(body))
}

// SetSize sets the width used by [Model.View]. The height parameter is
// ignored — toasts always render at their natural height for the
// wrapped message. SetSize satisfies the [component.Component]
// interface; it is also valid to set width via [WithWidth].
func (m *Model) SetSize(width, height int) {
	m.width = width
	_ = height
}

// Focus is a no-op. It exists only to satisfy [component.Component]:
// notifications dismiss on their own timer and have no interactive
// focus state.
func (m *Model) Focus() {}

// Blur is a no-op. It exists only to satisfy [component.Component]:
// notifications dismiss on their own timer and have no interactive
// focus state.
func (m *Model) Blur() {}

// Dismissed reports whether the notification has been dismissed by an
// auto-dismiss tick whose id matched the Model's generation.
func (m Model) Dismissed() bool {
	return m.dismissed
}

// Variant returns the variant the Model was constructed with.
func (m Model) Variant() Variant {
	return m.variant
}

// Message returns the body text the Model was constructed with.
func (m Model) Message() string {
	return m.message
}

package widget

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/format"
	"github.com/RenseiAI/tui-components/theme"
)

// Compile-time assertion that *Progressbar satisfies component.Component.
var _ component.Component = (*Progressbar)(nil)

// defaultProgressbarWidth is the rendered width applied when no
// WithProgressbarWidth option and no SetSize call have been made.
const defaultProgressbarWidth = 40

// minBarSegmentWidth is the smallest bar width we will allocate before
// dropping decoration segments (label, ETA, percent) to make room. It
// guarantees the bar is never starved out of the layout entirely.
const minBarSegmentWidth = 8

// percentTextFormat right-pads the numeric percentage to a stable 6 cells
// (e.g. "  0.0%", " 50.0%", "100.0%") so the layout does not jitter as
// percent grows.
const percentTextFormat = "%5.1f%%"

// ellipsis is appended after a truncated label.
const ellipsis = "…"

// defaultTickInterval is the frame cadence used by the indeterminate-mode
// animation. 80ms is fast enough to feel alive without saturating the
// renderer; it is intentionally an unexported constant — add a
// WithProgressbarTickInterval option only if a real need surfaces.
const defaultTickInterval = 80 * time.Millisecond

// sweepWidthFraction is the proportion of the bar width occupied by the
// indeterminate-mode lit segment (~20%).
const sweepWidthFraction = 0.2

// indeterminateLitRune and indeterminateBgRune are the cells used by the
// indeterminate sweep render.
const (
	indeterminateLitRune = '▌'
	indeterminateBgRune  = '░'
)

// tickMsg is the internal message type that drives the indeterminate-mode
// sweep. The generation field lets the model discard ticks scheduled before
// the most recent SetIndeterminate transition so a paused-then-resumed bar
// does not double-tick.
type tickMsg struct {
	gen int
}

// Progressbar is a themed progress bar widget wrapping
// charm.land/bubbles/v2/progress.
//
// The widget renders the most recent percent set via SetPercent immediately;
// the smooth spring animation built into the inner Bubbles model is bypassed
// in favor of deterministic, snapshot-friendly output. Callers that need
// animated transitions should use the inner Bubbles progress model directly.
//
// Progressbar implements component.Component: it satisfies tea.Model and
// exposes SetSize, Focus, and Blur. Focus and Blur are intentionally inert
// because progress bars are non-interactive; they exist solely to satisfy
// the Component interface.
//
// The zero value is not usable; construct with NewProgressbar.
type Progressbar struct {
	inner progress.Model

	// target is the last value passed to SetPercent, clamped to [0, 1].
	target float64

	// width is the rendered width in cells. A non-positive width causes
	// View to render nothing.
	width  int
	height int

	// gradient endpoints used to build the inner Bubbles model.
	fromColor color.Color
	toColor   color.Color

	// label is rendered to the left of the bar when non-empty.
	label string

	// showPercent toggles the right-aligned numeric percentage segment.
	showPercent bool

	// showETA toggles the right-aligned estimated-time-remaining segment.
	showETA bool

	// etaStart is the wall-clock instant of the first SetPercent call that
	// produced a non-zero target. The zero value means progress has not
	// started yet, which suppresses the ETA segment.
	etaStart time.Time

	// nowFn returns the current time. Tests inject a fake clock via
	// WithProgressbarClock so ETA output is deterministic.
	nowFn func() time.Time

	// indeterminate switches rendering from a percent-driven fill to an
	// animated sweep, suitable for work whose total is unknown.
	indeterminate bool

	// sweepStep is the indeterminate animation frame counter. Each tick
	// advances it by one; the rendered lit-segment offset is sweepStep
	// modulo width, so animation wraps without ever overflowing.
	sweepStep int

	// generation is incremented on every SetIndeterminate mode change so
	// in-flight tick commands scheduled under a previous mode are detected
	// and discarded by Update.
	generation int

	focused bool
}

// ProgressbarOption configures a Progressbar at construction time.
type ProgressbarOption func(*Progressbar)

// WithProgressbarWidth sets the rendered width in cells. Negative values are
// clamped to zero, in which case the bar renders an empty string until
// SetSize or another option provides a positive width.
func WithProgressbarWidth(w int) ProgressbarOption {
	return func(p *Progressbar) {
		if w < 0 {
			w = 0
		}
		p.width = w
	}
}

// WithProgressbarGradient overrides the default gradient. The default is
// theme.Teal to theme.Accent; pass any two lipgloss-compatible colors to
// customize. The gradient is applied across the full bar width.
func WithProgressbarGradient(from, to color.Color) ProgressbarOption {
	return func(p *Progressbar) {
		p.fromColor = from
		p.toColor = to
	}
}

// WithProgressbarLabel sets a leading label rendered to the left of the bar
// (e.g. "Uploading"). Empty labels are not rendered. Long labels are
// truncated with an ellipsis when the combined render would exceed the
// widget width; the bar, percent, and ETA segments are preferred over
// label width.
func WithProgressbarLabel(s string) ProgressbarOption {
	return func(p *Progressbar) {
		p.label = s
	}
}

// WithProgressbarShowPercent enables the right-aligned numeric percentage
// segment formatted as "XX.X%" with stable 6-column width.
func WithProgressbarShowPercent(b bool) ProgressbarOption {
	return func(p *Progressbar) {
		p.showPercent = b
	}
}

// WithProgressbarShowETA enables the right-aligned estimated time remaining
// segment, e.g. "~12s". The estimate is derived from the wall-clock rate
// observed since the first non-zero SetPercent call, formatted via the
// format package. The segment is suppressed before progress begins, when
// the rate is undefined, and at 100%.
func WithProgressbarShowETA(b bool) ProgressbarOption {
	return func(p *Progressbar) {
		p.showETA = b
	}
}

// WithProgressbarClock injects a clock function used by the ETA estimator.
// Pass time.Now (the default) for production; tests use this to make ETA
// output deterministic. Passing nil restores the default.
func WithProgressbarClock(now func() time.Time) ProgressbarOption {
	return func(p *Progressbar) {
		if now == nil {
			now = time.Now
		}
		p.nowFn = now
	}
}

// WithProgressbarIndeterminate constructs the bar in indeterminate mode,
// where rendering is an animated sweep instead of a percent-driven fill.
// Use this when the total work is unknown (waiting on a stream or
// indeterminate network operation). When the total becomes known,
// SetIndeterminate(false) switches the bar back to deterministic mode.
//
// Indeterminate is preferred over a spinner when the operation occupies a
// progress slot in a layout (so the visual cell does not jump in width)
// and when emphasizing "work is happening, duration unknown" rather than
// "thinking, soon to return".
func WithProgressbarIndeterminate(b bool) ProgressbarOption {
	return func(p *Progressbar) {
		p.indeterminate = b
	}
}

// NewProgressbar constructs a Progressbar. By default the bar is 40 cells
// wide, fills with a theme.Teal -> theme.Accent gradient, starts at 0%
// progress, and renders no label, percentage, or ETA. Apply
// ProgressbarOptions to customize width, gradient, label, percentage, and
// ETA display.
func NewProgressbar(opts ...ProgressbarOption) *Progressbar {
	p := &Progressbar{
		fromColor: theme.Teal,
		toColor:   theme.Accent,
		width:     defaultProgressbarWidth,
		focused:   true,
		nowFn:     time.Now,
	}
	for _, opt := range opts {
		opt(p)
	}
	p.inner = progress.New(
		progress.WithColors(p.fromColor, p.toColor),
		progress.WithoutPercentage(),
		progress.WithWidth(p.width),
	)
	return p
}

// SetPercent sets the displayed progress to v, clamped to [0, 1]. NaN inputs
// are treated as 0. The first call that produces a non-zero target records
// the ETA start time, used by WithProgressbarShowETA to estimate remaining
// time. The returned tea.Cmd is reserved for future animated transitions
// and is currently always nil; callers may safely discard it.
func (p *Progressbar) SetPercent(v float64) tea.Cmd {
	p.target = clampPercent(v)
	if p.target > 0 && p.etaStart.IsZero() {
		p.etaStart = p.nowFn()
	}
	return nil
}

// IncrBy advances the displayed progress by d, clamped to [0, 1]. The
// returned tea.Cmd is reserved for future animated transitions and is
// currently always nil; callers may safely discard it.
func (p *Progressbar) IncrBy(d float64) tea.Cmd {
	return p.SetPercent(p.target + d)
}

// Percent returns the current progress in the range [0, 1].
func (p *Progressbar) Percent() float64 {
	return p.target
}

// Init implements tea.Model. In deterministic mode it returns nil. In
// indeterminate mode it returns the first tick command so the animation
// begins as soon as the Bubble Tea program starts.
func (p *Progressbar) Init() tea.Cmd {
	if p.indeterminate {
		return p.tick()
	}
	return nil
}

// Update implements tea.Model. Indeterminate-mode tickMsgs whose generation
// matches the current SetIndeterminate generation advance the sweep and
// schedule the next tick; stale tickMsgs are silently dropped. All other
// messages are forwarded to the inner Bubbles progress model.
func (p *Progressbar) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if tm, ok := msg.(tickMsg); ok {
		if !p.indeterminate || tm.gen != p.generation {
			return p, nil
		}
		p.sweepStep++
		return p, p.tick()
	}
	var cmd tea.Cmd
	p.inner, cmd = p.inner.Update(msg)
	return p, cmd
}

// SetIndeterminate switches the bar between deterministic and indeterminate
// rendering. When entering indeterminate mode, the returned tea.Cmd starts
// the sweep tick loop; pass it back through your Bubble Tea program. When
// leaving indeterminate mode, the returned cmd is nil. Switching modes
// increments an internal generation counter so any in-flight tickMsgs
// scheduled before the switch are discarded by Update — this prevents
// double-ticking when a paused bar resumes. A no-op call (mode unchanged)
// returns nil and does not increment the generation.
func (p *Progressbar) SetIndeterminate(b bool) tea.Cmd {
	if p.indeterminate == b {
		return nil
	}
	p.indeterminate = b
	p.generation++
	if b {
		return p.tick()
	}
	return nil
}

// tick returns a Bubble Tea command that emits a tickMsg tagged with the
// current generation after defaultTickInterval has elapsed. The generation
// snapshot ensures Update can recognise and discard ticks scheduled under
// a stale mode.
func (p *Progressbar) tick() tea.Cmd {
	gen := p.generation
	return tea.Tick(defaultTickInterval, func(time.Time) tea.Msg {
		return tickMsg{gen: gen}
	})
}

// View renders the bar at the current percent. When width is non-positive
// the view is empty.
func (p *Progressbar) View() tea.View {
	return tea.NewView(p.render())
}

// render returns the styled progress bar string with optional label,
// percentage, and ETA decorations laid out as
// "{label} {bar} {percent} {eta}". Segments are dropped to fit within the
// configured width with precedence (most-preserved first):
// bar > percent > eta > label. The bar is never starved below
// minBarSegmentWidth; if even an undecorated bar cannot fit the available
// width, the inner Bubbles model still receives whatever width remains.
//
// In indeterminate mode the percent and ETA segments are suppressed
// because they are meaningless when the total work is unknown; the label
// is still rendered.
func (p *Progressbar) render() string {
	if p.width <= 0 {
		return ""
	}

	percentText := ""
	if p.showPercent && !p.indeterminate {
		percentText = fmt.Sprintf(percentTextFormat, p.target*100)
	}
	etaText := ""
	if p.showETA && !p.indeterminate {
		etaText = etaString(p.etaStart, p.nowFn(), p.target)
	}

	// Decoration widths use the rendered text widths (lipgloss styles do
	// not add visible cells, only ANSI escapes). Compute against the raw
	// strings and then style after layout decisions.
	labelW := lipgloss.Width(p.label)
	percentW := lipgloss.Width(percentText)
	etaW := lipgloss.Width(etaText)

	// Decide which segments to keep. Drop in reverse precedence (label,
	// then eta, then percent) until the bar gets at least
	// minBarSegmentWidth. The bar is always present.
	keepLabel := p.label != ""
	keepETA := etaText != ""
	keepPercent := percentText != ""
	for {
		segs := 1 // bar
		w := 0
		if keepLabel {
			segs++
			w += labelW
		}
		if keepPercent {
			segs++
			w += percentW
		}
		if keepETA {
			segs++
			w += etaW
		}
		separators := segs - 1
		barWidth := p.width - w - separators
		if barWidth >= minBarSegmentWidth {
			return p.composeRow(keepLabel, keepPercent, keepETA, percentText, etaText, barWidth)
		}
		// Shrink the lowest-precedence present decoration.
		switch {
		case keepLabel:
			keepLabel = false
		case keepETA:
			keepETA = false
		case keepPercent:
			keepPercent = false
		default:
			// Bar alone still does not fit; render at whatever width is
			// available (clamped to zero).
			if barWidth < 0 {
				barWidth = 0
			}
			return p.composeRow(false, false, false, "", "", barWidth)
		}
	}
}

// composeRow renders the final ordered string given which segments to
// include and the bar width to use. In indeterminate mode the bar segment
// is replaced by the sweep render.
func (p *Progressbar) composeRow(keepLabel, keepPercent, keepETA bool, percentText, etaText string, barWidth int) string {
	var bar string
	if p.indeterminate {
		bar = p.renderSweep(barWidth)
	} else {
		p.inner.SetWidth(barWidth)
		bar = p.inner.ViewAs(p.target)
	}

	parts := make([]string, 0, 4)
	if keepLabel {
		label := truncateLabel(p.label, p.labelBudget(barWidth, percentText, etaText, keepPercent, keepETA))
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.TextPrimary).Render(label))
	}
	parts = append(parts, bar)
	if keepPercent {
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.TextSecondary).Render(percentText))
	}
	if keepETA {
		parts = append(parts, lipgloss.NewStyle().Foreground(theme.TextSecondary).Render(etaText))
	}
	return strings.Join(parts, " ")
}

// renderSweep produces the indeterminate-mode animation frame: a lit
// segment of approximately sweepWidthFraction of the bar width over a
// surface-border background. The lit segment position advances by one cell
// per tick and wraps to the start when it would run past the right edge.
// Returns "" for non-positive widths.
func (p *Progressbar) renderSweep(width int) string {
	if width <= 0 {
		return ""
	}
	sweepW := int(math.Round(float64(width) * sweepWidthFraction))
	if sweepW < 1 {
		sweepW = 1
	}
	if sweepW > width {
		sweepW = width
	}
	pos := 0
	if width > 0 {
		pos = ((p.sweepStep % width) + width) % width
	}

	litStyle := lipgloss.NewStyle().Foreground(theme.Teal)
	bgStyle := lipgloss.NewStyle().Foreground(theme.SurfaceBorder)

	var b strings.Builder
	b.Grow(width * 2)
	for i := 0; i < width; i++ {
		var lit bool
		if pos+sweepW <= width {
			lit = i >= pos && i < pos+sweepW
		} else {
			// Lit segment wraps around the end.
			lit = i >= pos || i < (pos+sweepW)-width
		}
		if lit {
			b.WriteString(litStyle.Render(string(indeterminateLitRune)))
		} else {
			b.WriteString(bgStyle.Render(string(indeterminateBgRune)))
		}
	}
	return b.String()
}

// labelBudget returns the maximum visible cells available for the label
// after accounting for the bar, the kept right-side segments, and their
// separators. Negative budgets are clamped to zero.
func (p *Progressbar) labelBudget(barWidth int, percentText, etaText string, keepPercent, keepETA bool) int {
	used := barWidth
	segs := 1
	if keepPercent {
		used += lipgloss.Width(percentText)
		segs++
	}
	if keepETA {
		used += lipgloss.Width(etaText)
		segs++
	}
	// Add one separator for the label itself plus separators between the
	// bar and right-side segments.
	separators := segs // bar-label + (bar-percent / percent-eta as applicable)
	budget := p.width - used - separators
	if budget < 0 {
		budget = 0
	}
	return budget
}

// truncateLabel shortens s to fit max visible cells, replacing the trailing
// rune(s) with an ellipsis when truncation occurs. Returns "" when max is
// non-positive.
func truncateLabel(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return ellipsis
	}
	// Drop runes from the end until the combined width fits.
	runes := []rune(s)
	for len(runes) > 0 {
		candidate := string(runes) + ellipsis
		if lipgloss.Width(candidate) <= max {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return ellipsis
}

// etaString returns a relative time-remaining string like "~12s". It
// returns "" when the percent is non-positive, the start time is zero, no
// time has elapsed, the rate is undefined, or the bar has reached 100%.
// The output uses format.Duration; sub-second remainders are rounded up to
// 1s so the segment is never blank-but-meaningful.
func etaString(start time.Time, now time.Time, percent float64) string {
	if start.IsZero() || percent <= 0 || percent >= 1 {
		return ""
	}
	elapsed := now.Sub(start)
	if elapsed <= 0 {
		return ""
	}
	rate := percent / elapsed.Seconds()
	if rate <= 0 {
		return ""
	}
	remaining := (1 - percent) / rate
	if remaining <= 0 {
		return ""
	}
	secs := int(math.Ceil(remaining))
	if secs < 1 {
		secs = 1
	}
	return "~" + format.Duration(secs)
}

// SetSize sets the rendered width and records the height. Negative width is
// clamped to zero. The height is stored but not used because the bar is
// always one line tall.
func (p *Progressbar) SetSize(width, height int) {
	if width < 0 {
		width = 0
	}
	p.width = width
	p.height = height
	p.inner.SetWidth(width)
}

// Focus marks the bar as focused. It is intentionally inert: the bar is
// non-interactive and ignores focus state for rendering. The method exists
// to satisfy component.Component.
func (p *Progressbar) Focus() {
	p.focused = true
}

// Blur marks the bar as blurred. It is intentionally inert: the bar is
// non-interactive and ignores focus state for rendering. The method exists
// to satisfy component.Component.
func (p *Progressbar) Blur() {
	p.focused = false
}

// clampPercent returns v clamped to [0, 1]; NaN becomes 0.
func clampPercent(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

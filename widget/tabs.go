package widget

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/component"
	"github.com/RenseiAI/tui-components/theme"
)

// tabSeparatorGlyph is the visible separator rendered between adjacent tabs.
const tabSeparatorGlyph = " │ "

// tabEllipsis is the single-rune ellipsis appended to truncated titles and
// used as the overflow marker when tabs are dropped from the rendered bar.
const tabEllipsis = "…"

// tabMinDisplayWidth is the lower bound on the display width of a
// truncated tab title (3 visible runes plus the ellipsis). Titles
// shorter than this are never shortened.
const tabMinDisplayWidth = 4

// TabsItem is a single entry in a [Tabs] widget.
type TabsItem struct {
	// ID is a stable identifier emitted with [TabsSelectedMsg] when the
	// tab becomes active. Callers use it to route view changes without
	// depending on the slice index.
	ID string

	// Title is the human-readable label rendered in the tab bar.
	Title string

	// KeyHint is an optional keystroke shown alongside the title and
	// used to activate the tab directly. It is matched against the
	// [tea.KeyPressMsg] string representation via exact equality.
	KeyHint string

	// Disabled marks a tab as non-interactive. Disabled tabs are
	// skipped during sequential navigation (left/right, tab/shift+tab)
	// and cannot be activated via [KeyHint] shortcuts.
	Disabled bool
}

// TabsKeyMap is the set of key bindings that drive [Tabs] navigation.
type TabsKeyMap struct {
	// Left moves the active tab one position toward the start of the list.
	Left key.Binding

	// Right moves the active tab one position toward the end of the list.
	Right key.Binding

	// Next moves the active tab one position toward the end of the list
	// (conventionally bound to the tab key).
	Next key.Binding

	// Prev moves the active tab one position toward the start of the list
	// (conventionally bound to shift+tab).
	Prev key.Binding
}

// DefaultTabsKeyMap returns the default [TabsKeyMap] binding left/right
// for horizontal navigation and tab/shift+tab as alternate shortcuts.
func DefaultTabsKeyMap() TabsKeyMap {
	return TabsKeyMap{
		Left: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "prev tab"),
		),
		Right: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next tab"),
		),
		Next: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		Prev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
	}
}

// TabsSelectedMsg is emitted by [Tabs] when the active tab changes.
// It is never emitted for no-op navigation that leaves the active index
// unchanged (for example, pressing left while already at index 0).
type TabsSelectedMsg struct {
	// Index is the new active tab index.
	Index int

	// ID is the [TabsItem.ID] of the newly-active tab.
	ID string
}

// TabsOption configures a [Tabs] widget via the functional-options pattern.
type TabsOption func(*Tabs)

// WithActive sets the initial active tab index. Out-of-range values are
// clamped to the valid range at construction time.
func WithActive(idx int) TabsOption {
	return func(t *Tabs) {
		t.activeIdx = idx
	}
}

// WithTabsKeyMap overrides the default [TabsKeyMap] used by the widget.
func WithTabsKeyMap(km TabsKeyMap) TabsOption {
	return func(t *Tabs) {
		t.keyMap = km
	}
}

// WithWraparound enables (or disables) wraparound navigation. When
// enabled, moving left from the first tab wraps to the last valid tab
// and moving right from the last tab wraps to the first. The default is
// false — sequential navigation clamps at the ends.
func WithWraparound(enabled bool) TabsOption {
	return func(t *Tabs) {
		t.wraparound = enabled
	}
}

// Tabs is a horizontal tab-bar widget. It implements
// [component.Component] and emits [TabsSelectedMsg] when the active tab
// changes via [Tabs.Update] key handling.
type Tabs struct {
	items      []TabsItem
	activeIdx  int
	width      int
	height     int
	focused    bool
	wraparound bool
	keyMap     TabsKeyMap
}

// Compile-time assertion that *Tabs satisfies component.Component.
var _ component.Component = (*Tabs)(nil)

// NewTabs constructs a [Tabs] widget with the given items and options.
// The active index is clamped to the valid range and defaults to 0.
func NewTabs(items []TabsItem, opts ...TabsOption) *Tabs {
	t := &Tabs{
		items:  items,
		keyMap: DefaultTabsKeyMap(),
	}
	for _, opt := range opts {
		opt(t)
	}
	t.activeIdx = clampIndex(t.activeIdx, len(t.items))
	return t
}

// Init satisfies [tea.Model]. The tab bar has no startup work.
func (t *Tabs) Init() tea.Cmd {
	return nil
}

// Update handles navigation key events. Input is ignored when the
// widget is not focused. When navigation changes the active index a
// [TabsSelectedMsg] command is returned.
//
// In addition to sequential navigation via the configured [TabsKeyMap],
// any [TabsItem.KeyHint] that matches the key event's string form
// activates that tab directly. Shortcut activation on a disabled tab or
// on the already-active tab is a no-op.
func (t *Tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !t.focused {
		return t, nil
	}

	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return t, nil
	}

	switch {
	case key.Matches(keyMsg, t.keyMap.Left), key.Matches(keyMsg, t.keyMap.Prev):
		return t, t.move(-1)
	case key.Matches(keyMsg, t.keyMap.Right), key.Matches(keyMsg, t.keyMap.Next):
		return t, t.move(+1)
	}

	if cmd := t.handleShortcut(keyMsg.String()); cmd != nil {
		return t, cmd
	}
	return t, nil
}

// handleShortcut activates the first tab whose [TabsItem.KeyHint]
// matches keyStr exactly. It returns nil when no tab matches, when the
// matched tab is disabled, or when the matched tab is already active.
func (t *Tabs) handleShortcut(keyStr string) tea.Cmd {
	if keyStr == "" {
		return nil
	}
	for i, item := range t.items {
		if item.KeyHint == "" || item.KeyHint != keyStr {
			continue
		}
		if item.Disabled {
			return nil
		}
		if i == t.activeIdx {
			return nil
		}
		t.activeIdx = i
		id := t.items[i].ID
		return func() tea.Msg {
			return TabsSelectedMsg{Index: i, ID: id}
		}
	}
	return nil
}

// View renders the tab bar as a single horizontal line. When a
// positive width has been recorded via [Tabs.SetSize] the bar is fit
// to that width by truncating titles, removing separators from the
// outer edges, or, as a last resort, dropping tabs and appending a
// trailing ellipsis marker while keeping the active tab visible.
func (t *Tabs) View() tea.View {
	return tea.NewView(t.render())
}

// render builds the tab bar string. It is separated from [Tabs.View] so
// tests and downstream consumers can inspect the rendered output as a
// plain string without materializing a [tea.View].
//
// When [Tabs.SetSize] has been called with a positive width the bar is
// fit to that budget through three cascading strategies: title
// truncation with a trailing ellipsis (down to a minimum display width
// of four), separator removal from the outside edges inward, and — as a
// last resort — dropping tabs entirely with a trailing ellipsis marker
// while keeping the active tab in view.
func (t *Tabs) render() string {
	if len(t.items) == 0 {
		return theme.TabBar().Render("")
	}

	titles := make([]string, len(t.items))
	for i, item := range t.items {
		if item.KeyHint != "" {
			titles[i] = item.Title + " [" + item.KeyHint + "]"
		} else {
			titles[i] = item.Title
		}
	}

	// Unconstrained width: render everything with full separators.
	if t.width <= 0 {
		return t.renderWithTitles(titles, fullSeparatorMask(len(titles)))
	}

	return t.fitTabs(titles, t.width)
}

// renderWithTitles joins the given (possibly truncated) titles using
// sepMask to decide, for each gap between adjacent tabs, whether a
// styled separator is rendered (true) or the tabs are joined without a
// separator (false). sepMask has length len(titles)-1 and is indexed so
// that sepMask[i] controls the gap between segment i and segment i+1.
func (t *Tabs) renderWithTitles(titles []string, sepMask []bool) string {
	if len(titles) == 0 {
		return theme.TabBar().Render("")
	}

	separator := theme.TabSeparator().Render(tabSeparatorGlyph)

	var b strings.Builder
	for i, title := range titles {
		if i > 0 {
			if i-1 < len(sepMask) && sepMask[i-1] {
				b.WriteString(separator)
			}
		}
		b.WriteString(t.styleForIndex(i).Render(title))
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, b.String())
	return theme.TabBar().Render(bar)
}

// styleForIndex returns the lipgloss style that should wrap the tab
// segment at index i: active wins over disabled, disabled wins over
// inactive.
func (t *Tabs) styleForIndex(i int) lipgloss.Style {
	switch {
	case i == t.activeIdx:
		return theme.TabActive()
	case i < len(t.items) && t.items[i].Disabled:
		return theme.TabDisabled()
	default:
		return theme.TabInactive()
	}
}

// fitTabs applies the cascading width-reduction strategy described on
// [Tabs.render] and returns the final rendered bar string constrained
// to width. Callers must pass width > 0.
func (t *Tabs) fitTabs(titles []string, width int) string {
	sepMask := fullSeparatorMask(len(titles))

	// Strategy 1 + 2: try the untruncated titles first, then shrink the
	// longest title one display unit at a time until the bar fits or
	// every title has hit the minimum display width.
	working := append([]string(nil), titles...)
	if out, ok := t.tryFit(working, sepMask, width); ok {
		return out
	}
	for shrinkLongestTitle(working) {
		if out, ok := t.tryFit(working, sepMask, width); ok {
			return out
		}
	}

	// Strategy 3: drop separators from the outside edges inward.
	for dropNextSeparator(sepMask) {
		if out, ok := t.tryFit(working, sepMask, width); ok {
			return out
		}
	}

	// Strategy 4: overflow render. Pick the widest contiguous window of
	// tabs that fits (with all separators between the surviving tabs
	// plus the trailing ellipsis marker), preferring windows that
	// include the active tab.
	return t.renderOverflow(working, width)
}

// tryFit renders the bar and returns (rendered, true) when it fits in
// width. Otherwise it returns ("", false) without allocating the final
// theme-wrapped string a second time.
func (t *Tabs) tryFit(titles []string, sepMask []bool, width int) (string, bool) {
	out := t.renderWithTitles(titles, sepMask)
	if lipgloss.Width(out) <= width {
		return out, true
	}
	return "", false
}

// fullSeparatorMask returns a slice of n-1 true values, or nil when
// fewer than two tabs are present.
func fullSeparatorMask(n int) []bool {
	if n < 2 {
		return nil
	}
	mask := make([]bool, n-1)
	for i := range mask {
		mask[i] = true
	}
	return mask
}

// shrinkLongestTitle finds the title with the greatest display width
// that is still above [tabMinDisplayWidth] and reduces it by one
// display unit (replacing the last visible character with the
// ellipsis). It returns true when a title was shrunk and false when
// every title is already at the minimum.
func shrinkLongestTitle(titles []string) bool {
	bestIdx := -1
	bestWidth := 0
	for i, title := range titles {
		w := lipgloss.Width(title)
		if w <= tabMinDisplayWidth {
			continue
		}
		if w > bestWidth {
			bestWidth = w
			bestIdx = i
		}
	}
	if bestIdx < 0 {
		return false
	}
	target := bestWidth - 1
	if target < tabMinDisplayWidth {
		target = tabMinDisplayWidth
	}
	titles[bestIdx] = truncateToWidth(titles[bestIdx], target)
	return true
}

// truncateToWidth returns s shortened so its display width is at most
// target, appending [tabEllipsis] when truncation occurs. It never
// splits a grapheme or wide rune: runes are appended whole, and the
// routine stops before any rune whose inclusion would push the
// ellipsis-terminated result past target. When s already fits, it is
// returned unchanged.
func truncateToWidth(s string, target int) string {
	if target <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= target {
		return s
	}
	ellipsisWidth := lipgloss.Width(tabEllipsis)
	if target <= ellipsisWidth {
		return tabEllipsis
	}
	var b strings.Builder
	budget := target - ellipsisWidth
	for _, r := range s {
		candidate := b.String() + string(r)
		if lipgloss.Width(candidate) > budget {
			break
		}
		b.WriteRune(r)
	}
	return b.String() + tabEllipsis
}

// dropNextSeparator flips the next true entry in sepMask to false
// following an alternating outside-in pattern: leftmost remaining,
// then rightmost remaining, and so on. It returns false when every
// separator is already suppressed.
func dropNextSeparator(sepMask []bool) bool {
	n := len(sepMask)
	if n == 0 {
		return false
	}
	// Count remaining separators to decide whether to drop from the
	// left or the right next. This mirrors the alternation rule:
	// strictly more visible gaps on the left means drop left; equal or
	// more on the right means drop right. We pick a single
	// deterministic ordering: always drop the outermost remaining
	// separator, alternating between left and right on successive
	// calls.
	left, right := 0, n-1
	for left <= right && !sepMask[left] {
		left++
	}
	for right >= 0 && !sepMask[right] {
		right--
	}
	if left > right {
		return false
	}
	// Count active separators on each side of the centre to pick a
	// deterministic alternation without persistent state: if the
	// left-remaining index is closer to its edge than the
	// right-remaining index is to its edge, drop the left one;
	// otherwise drop the right. Ties go to the left.
	leftDist := left
	rightDist := (n - 1) - right
	if leftDist <= rightDist {
		sepMask[left] = false
	} else {
		sepMask[right] = false
	}
	return true
}

// renderOverflow handles the case where the fully-shrunken,
// separator-free bar still exceeds width. It selects a contiguous
// window of tabs that fits alongside the trailing ellipsis marker,
// biased to keep the active tab visible. The returned string is
// already wrapped in the tab-bar style.
func (t *Tabs) renderOverflow(titles []string, width int) string {
	n := len(titles)
	if n == 0 || width <= 0 {
		return theme.TabBar().Render("")
	}
	// The marker is rendered with the unpadded separator style so it
	// stays exactly one display unit wide — padding on the regular
	// tab styles would push it to three and make tiny widths
	// unfriendly.
	marker := theme.TabSeparator().Render(tabEllipsis)
	markerWidth := lipgloss.Width(marker)

	// Degenerate budgets: just the marker (or as much of it as fits).
	if width <= markerWidth {
		return theme.TabBar().Render(marker)
	}

	active := clampIndex(t.activeIdx, n)
	budget := width - markerWidth

	// Try windows of decreasing size, preferring those that include
	// the active tab. A window of size n-1 is the largest meaningful
	// attempt because the caller only reaches this path after strategy
	// 3 already failed for size n.
	for size := n - 1; size >= 1; size-- {
		for _, start := range candidateStarts(n, size, active) {
			end := start + size
			sub := titles[start:end]
			subMask := fullSeparatorMask(len(sub))
			rendered := t.renderWindow(sub, subMask, start)
			if lipgloss.Width(rendered) <= budget {
				return theme.TabBar().Render(
					lipgloss.JoinHorizontal(lipgloss.Top, rendered, marker),
				)
			}
		}
	}

	// Nothing fit — emit only the marker.
	return theme.TabBar().Render(marker)
}

// renderWindow renders a slice of titles starting at global index
// offset within t.items, using sepMask for the inter-tab gaps. Styling
// honours the original item indices (so the active/disabled tabs keep
// their styles even inside a window).
func (t *Tabs) renderWindow(titles []string, sepMask []bool, offset int) string {
	separator := theme.TabSeparator().Render(tabSeparatorGlyph)
	var b strings.Builder
	for i, title := range titles {
		if i > 0 && i-1 < len(sepMask) && sepMask[i-1] {
			b.WriteString(separator)
		}
		b.WriteString(t.styleForIndex(offset + i).Render(title))
	}
	return b.String()
}

// candidateStarts returns the list of window start indices of the
// given size that are valid for an n-tab list, ordered so that windows
// containing active appear first. When multiple windows contain
// active, the one that centres active is preferred; ties break toward
// the left.
func candidateStarts(n, size, active int) []int {
	if size <= 0 || size > n {
		return nil
	}
	maxStart := n - size
	// Preferred start: centre active inside the window, clamped to the
	// valid range [0, maxStart].
	preferred := active - size/2
	if preferred < 0 {
		preferred = 0
	}
	if preferred > maxStart {
		preferred = maxStart
	}
	starts := []int{preferred}
	// Walk outward from preferred, alternating left/right, to build a
	// deterministic ordering.
	for delta := 1; delta <= maxStart; delta++ {
		if preferred-delta >= 0 {
			starts = append(starts, preferred-delta)
		}
		if preferred+delta <= maxStart {
			starts = append(starts, preferred+delta)
		}
	}
	return starts
}

// SetSize records the available width and height for the tab bar.
// Non-positive values are coerced to zero. A positive width causes
// [Tabs.View] to apply width-aware truncation; a zero width renders
// the bar at its natural size.
func (t *Tabs) SetSize(w, h int) {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	t.width = w
	t.height = h
}

// Focus enables key handling in [Tabs.Update].
func (t *Tabs) Focus() {
	t.focused = true
}

// Blur disables key handling in [Tabs.Update].
func (t *Tabs) Blur() {
	t.focused = false
}

// Active returns the currently-active tab index. When the item list is
// empty the returned value is 0 and has no associated tab.
func (t *Tabs) Active() int {
	return t.activeIdx
}

// SetActive sets the active tab index. Out-of-range values are clamped
// to the valid range. This method is for direct programmatic updates;
// it does not emit a [TabsSelectedMsg]. Messages are only emitted for
// navigation performed through [Tabs.Update].
func (t *Tabs) SetActive(idx int) {
	t.activeIdx = clampIndex(idx, len(t.items))
}

// SetItems replaces the tab items. The active index is clamped to the
// new range so it remains valid.
func (t *Tabs) SetItems(items []TabsItem) {
	t.items = items
	t.activeIdx = clampIndex(t.activeIdx, len(t.items))
}

// move walks the tab list in the given direction (delta must be ±1),
// skipping disabled items, and returns a command emitting
// [TabsSelectedMsg] if the active index actually changed. When the
// widget is configured with wraparound the walk wraps once around the
// list; otherwise it stops at the ends. The method returns nil when no
// non-disabled target exists in the requested direction.
func (t *Tabs) move(delta int) tea.Cmd {
	n := len(t.items)
	if n == 0 {
		return nil
	}
	if delta != 1 && delta != -1 {
		return nil
	}

	idx := t.activeIdx
	// Walk at most n-1 steps so we never visit the starting index
	// twice. This guarantees termination under both clamped and
	// wraparound modes even if every other item is disabled.
	for step := 0; step < n-1; step++ {
		next := idx + delta
		if next < 0 || next >= n {
			if !t.wraparound {
				return nil
			}
			// Wrap once. After wrapping, if we would land back on the
			// starting index the loop bound already prevents it.
			next = (next + n) % n
		}
		idx = next
		if !t.items[idx].Disabled {
			if idx == t.activeIdx {
				return nil
			}
			t.activeIdx = idx
			id := t.items[idx].ID
			return func() tea.Msg {
				return TabsSelectedMsg{Index: idx, ID: id}
			}
		}
	}
	return nil
}

// clampIndex returns idx bounded to the half-open interval [0, n).
// When n is zero the result is zero (no valid index exists).
func clampIndex(idx, n int) int {
	if n <= 0 {
		return 0
	}
	if idx < 0 {
		return 0
	}
	if idx >= n {
		return n - 1
	}
	return idx
}

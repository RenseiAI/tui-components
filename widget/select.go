// Package widget is defined in doc.go.

package widget

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/theme"
)

// SelectItem extends list.Item with a stable identity.
// Delegates receive a SelectItem and may type-assert to a richer concrete type
// for custom rendering. See [theme/worktype.go], [theme/status.go], and
// [theme/activity.go] for the canonical icon and color palettes that
// consumer-side delegates should use.
type SelectItem interface {
	list.Item
	// ID returns a stable, unique identifier for the item.
	ID() string
}

// SelectKeyMap defines key bindings for the Select widget.
type SelectKeyMap = list.KeyMap

// SelectToggleKeyMap holds the key binding used to toggle item selection
// in multi-select mode.
type SelectToggleKeyMap struct {
	// Toggle toggles the highlighted item's selection state.
	Toggle key.Binding
}

// DefaultSelectToggleKeyMap returns the default toggle key binding (space).
func DefaultSelectToggleKeyMap() SelectToggleKeyMap {
	return SelectToggleKeyMap{
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle selection"),
		),
	}
}

// SelectOption is a functional option for configuring a Select widget.
type SelectOption func(*Select)

// WithFilter enables or disables fuzzy filtering on the select list.
func WithFilter(enabled bool) SelectOption {
	return func(s *Select) {
		s.filterEnabled = enabled
	}
}

// WithKeyMap sets custom key bindings on the select list.
func WithKeyMap(km SelectKeyMap) SelectOption {
	return func(s *Select) {
		s.keyMap = &km
	}
}

// WithMultiSelect enables multi-select mode, allowing the user to toggle
// multiple items in the list via the space key.
func WithMultiSelect(enabled bool) SelectOption {
	return func(s *Select) {
		s.multiSelect = enabled
		if enabled {
			s.selected = make(map[string]SelectItem)
		}
	}
}

// WithDelegate sets a custom [list.ItemDelegate] that overrides the default
// themed delegate. The delegate's Render method receives items that satisfy
// [SelectItem]; callers may type-assert to a richer concrete type for custom
// rendering. When using a custom delegate, multi-select prefix indicators are
// the delegate's responsibility — see [Select.ShowMultiSelectIndicators].
func WithDelegate(d list.ItemDelegate) SelectOption {
	return func(s *Select) {
		s.customDelegate = d
	}
}

// Select is a list widget backed by bubbles/v2 list.Model that supports both
// single-select and multi-select modes. It implements the component.Component
// interface.
type Select struct {
	list           list.Model
	focused        bool
	multiSelect    bool
	filterEnabled  bool
	showIndicators bool
	keyMap         *SelectKeyMap
	toggleKeyMap   SelectToggleKeyMap
	customDelegate list.ItemDelegate
	selected       map[string]SelectItem
	selectedOrder  []string
}

// NewSelect creates a new Select widget from the given items and options.
// By default, filtering is enabled and the default key map is used.
func NewSelect(items []SelectItem, opts ...SelectOption) *Select {
	s := &Select{
		filterEnabled:  true,
		showIndicators: true,
		toggleKeyMap:   DefaultSelectToggleKeyMap(),
	}

	for _, opt := range opts {
		opt(s)
	}

	// Convert SelectItem slice to list.Item slice.
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	var delegate list.ItemDelegate
	if s.customDelegate != nil {
		delegate = s.customDelegate
	} else {
		delegate = newThemedSelectDelegate(s)
	}

	s.list = list.New(listItems, delegate, 0, 0)
	s.list.SetFilteringEnabled(s.filterEnabled)
	s.list.SetShowTitle(false)
	s.list.DisableQuitKeybindings()

	if s.keyMap != nil {
		s.list.KeyMap = *s.keyMap
	}

	applyListStyles(&s.list)

	return s
}

// Init implements tea.Model.
func (s *Select) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (s *Select) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !s.focused {
		return s, nil
	}

	// Handle space toggle in multi-select mode before passing to the list,
	// so the list does not consume the key press.
	if s.multiSelect {
		if msg, ok := msg.(tea.KeyPressMsg); ok {
			if key.Matches(msg, s.toggleKeyMap.Toggle) {
				s.toggleHighlighted()
				return s, nil
			}
		}
	}

	var cmd tea.Cmd
	s.list, cmd = s.list.Update(msg)
	return s, cmd
}

// View implements tea.Model.
func (s *Select) View() tea.View {
	return tea.NewView(s.list.View())
}

// SetSize implements component.Component.
func (s *Select) SetSize(width, height int) {
	s.list.SetSize(width, height)
}

// Focus implements component.Component.
func (s *Select) Focus() {
	s.focused = true
}

// Blur implements component.Component.
func (s *Select) Blur() {
	s.focused = false
}

// Selected returns the currently selected items.
// In multi-select mode it returns all toggled items in insertion order.
// In single-select mode it returns the highlighted item as a single-element
// slice. Returns nil if nothing is selected.
func (s *Select) Selected() []SelectItem {
	if !s.multiSelect {
		item := s.Highlighted()
		if item == nil {
			return nil
		}
		return []SelectItem{item}
	}

	if len(s.selectedOrder) == 0 {
		return nil
	}

	result := make([]SelectItem, 0, len(s.selectedOrder))
	for _, id := range s.selectedOrder {
		if item, ok := s.selected[id]; ok {
			result = append(result, item)
		}
	}
	return result
}

// Highlighted returns the item under the cursor, or nil if the list is empty.
func (s *Select) Highlighted() SelectItem {
	item := s.list.SelectedItem()
	if item == nil {
		return nil
	}
	si, ok := item.(SelectItem)
	if !ok {
		return nil
	}
	return si
}

// SetMultiSelect enables or disables multi-select mode.
// When switching to single-select mode the selection set is cleared.
// When enabling multi-select the internal map is initialized if needed.
func (s *Select) SetMultiSelect(v bool) {
	s.multiSelect = v
	if !v {
		s.selected = nil
		s.selectedOrder = nil
	} else if s.selected == nil {
		s.selected = make(map[string]SelectItem)
	}
}

// IsSelected reports whether the item with the given ID is in the selection set.
func (s *Select) IsSelected(id string) bool {
	if s.selected == nil {
		return false
	}
	_, ok := s.selected[id]
	return ok
}

// ShowMultiSelectIndicators controls whether the default themed delegate
// renders checkbox-style prefix indicators for multi-select mode. The default
// is true. When a custom delegate is provided via [WithDelegate], this flag
// has no effect — custom delegates are responsible for their own rendering.
func (s *Select) ShowMultiSelectIndicators(v bool) {
	s.showIndicators = v
}

// SetDelegate replaces the item delegate on the underlying list.
func (s *Select) SetDelegate(d list.ItemDelegate) {
	s.list.SetDelegate(d)
}

// toggleHighlighted adds or removes the highlighted item from the selection set.
func (s *Select) toggleHighlighted() {
	item := s.Highlighted()
	if item == nil {
		return
	}

	id := item.ID()
	if _, exists := s.selected[id]; exists {
		delete(s.selected, id)
		// Remove from order slice.
		for i, oid := range s.selectedOrder {
			if oid == id {
				s.selectedOrder = append(s.selectedOrder[:i], s.selectedOrder[i+1:]...)
				break
			}
		}
	} else {
		s.selected[id] = item
		s.selectedOrder = append(s.selectedOrder, id)
	}
}

// ---------------------------------------------------------------------------
// selectDelegate – custom delegate wrapping list.DefaultDelegate to add
// multi-select indicator prefixes. This lives behind the delegate interface
// so TC-004c can override it cleanly.
// ---------------------------------------------------------------------------

// selectDelegate wraps a [list.DefaultDelegate] to prepend multi-select
// indicator prefixes (✓ / ☐) when multi-select mode is active.
type selectDelegate struct {
	list.DefaultDelegate
	selectWidget *Select
}

// newThemedSelectDelegate creates a [selectDelegate] styled with theme colors.
func newThemedSelectDelegate(s *Select) *selectDelegate {
	d := list.NewDefaultDelegate()
	d.Styles = themedItemStyles()
	return &selectDelegate{
		DefaultDelegate: d,
		selectWidget:    s,
	}
}

// Render renders the item, prepending a selection indicator when multi-select
// mode is active. Selected items show ✓ in green; unselected items show ☐ in
// dim text.
func (d *selectDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	if !d.selectWidget.multiSelect || !d.selectWidget.showIndicators {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	si, ok := item.(SelectItem)
	if !ok {
		d.DefaultDelegate.Render(w, m, index, item)
		return
	}

	var prefix string
	if d.selectWidget.IsSelected(si.ID()) {
		prefix = lipgloss.NewStyle().
			Foreground(theme.StatusSuccess).
			Render("✓ ")
	} else {
		prefix = lipgloss.NewStyle().
			Foreground(theme.TextTertiary).
			Render("☐ ")
	}

	fmt.Fprint(w, prefix) //nolint:errcheck // list delegate writers don't fail
	d.DefaultDelegate.Render(w, m, index, item)
}

// themedItemStyles returns DefaultItemStyles wired to the theme palette.
func themedItemStyles() list.DefaultItemStyles {
	var s list.DefaultItemStyles

	s.NormalTitle = lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Padding(0, 0, 0, 2) //nolint:mnd // visual padding

	s.NormalDesc = lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Padding(0, 0, 0, 2) //nolint:mnd // visual padding

	s.SelectedTitle = lipgloss.NewStyle().
		Foreground(theme.TextPrimary).
		Background(theme.SurfaceRaised).
		Padding(0, 0, 0, 2). //nolint:mnd // visual padding
		Bold(true)

	s.SelectedDesc = lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Background(theme.SurfaceRaised).
		Padding(0, 0, 0, 2) //nolint:mnd // visual padding

	s.DimmedTitle = lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Padding(0, 0, 0, 2) //nolint:mnd // visual padding

	s.DimmedDesc = lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Padding(0, 0, 0, 2) //nolint:mnd // visual padding

	s.FilterMatch = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)

	return s
}

// applyListStyles applies theme-based styling to the list chrome.
func applyListStyles(m *list.Model) {
	styles := list.DefaultStyles(true)

	styles.TitleBar = lipgloss.NewStyle().
		Background(theme.Surface).
		Padding(0, 1) //nolint:mnd // visual padding

	styles.Title = theme.Header()

	styles.StatusBar = lipgloss.NewStyle().
		Foreground(theme.TextSecondary).
		Background(theme.Surface).
		Padding(0, 1) //nolint:mnd // visual padding

	styles.StatusEmpty = lipgloss.NewStyle().
		Foreground(theme.TextTertiary)

	styles.StatusBarActiveFilter = lipgloss.NewStyle().
		Foreground(theme.Accent)

	styles.StatusBarFilterCount = lipgloss.NewStyle().
		Foreground(theme.TextTertiary)

	styles.NoItems = lipgloss.NewStyle().
		Foreground(theme.TextTertiary)

	styles.HelpStyle = lipgloss.NewStyle().
		Foreground(theme.TextTertiary).
		Padding(1, 0, 0, 2) //nolint:mnd // visual padding

	styles.ActivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.Accent)

	styles.InactivePaginationDot = lipgloss.NewStyle().
		Foreground(theme.TextTertiary)

	styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true)

	m.Styles = styles
}

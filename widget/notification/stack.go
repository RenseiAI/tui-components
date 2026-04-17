package notification

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/log"

	"github.com/RenseiAI/tui-components/component"
)

// Compile-time assertions: *Stack satisfies both tea.Model and the
// project Component interface so callers can embed it directly into a
// parent Bubble Tea program.
var (
	_ tea.Model           = (*Stack)(nil)
	_ component.Component = (*Stack)(nil)
)

// DefaultStackMax is the cap applied when [WithMax] receives a
// non-positive argument. Without [WithMax], a Stack imposes no cap.
const DefaultStackMax = 5

// Stack composes multiple concurrent toasts into a single rendered
// block. Stack is a value type — its mutating methods ([Stack.Push],
// [Stack.Update]) return a fresh Stack — so embedding it in a parent
// Bubble Tea model and reassigning the result avoids aliasing bugs.
//
// Positioning is intentionally caller-owned: [Stack.View] returns the
// composed block as-is, with no padding, margin, or screen placement.
// Use [charm.land/lipgloss/v2.Place] or compose into a parent layout
// to position it.
type Stack struct {
	items       []Model
	max         int
	width       int
	newestFirst bool
}

// StackOption configures a [Stack] during construction.
type StackOption func(*Stack)

// WithMax caps the number of live notifications in the stack. When
// [Stack.Push] would exceed the cap, the oldest non-dismissed entry is
// evicted to make room. A non-positive cap coerces to [DefaultStackMax]
// and is logged at warn level.
func WithMax(n int) StackOption {
	return func(s *Stack) {
		if n <= 0 {
			log.Warn("notification: invalid Stack max; coercing to default",
				"given", n, "default", DefaultStackMax)
			s.max = DefaultStackMax
			return
		}
		s.max = n
	}
}

// WithNewestFirst inverts the render order so the most recently pushed
// notification appears at the top of the rendered block. Default is
// newest-last (most recent at the bottom).
func WithNewestFirst() StackOption {
	return func(s *Stack) {
		s.newestFirst = true
	}
}

// WithStackWidth sets the width applied to each child notification
// when it is added via [Stack.Push] or when [Stack.SetSize] is called
// later. Children with a non-zero own width still respect that width
// after [Stack.Push] applies the stack width.
func WithStackWidth(w int) StackOption {
	return func(s *Stack) {
		s.width = w
	}
}

// NewStack constructs an empty Stack and applies the given options.
func NewStack(opts ...StackOption) Stack {
	s := Stack{}
	for _, opt := range opts {
		opt(&s)
	}
	return s
}

// Push appends a notification to the stack and returns the resulting
// stack along with the auto-dismiss tick command for the new
// notification. When the configured cap is exceeded, the oldest
// non-dismissed entry is evicted to make room.
//
// Push assigns a fresh generation id to the incoming Model. The
// constructor already does so, so this is a defensive guard against a
// caller reusing a Model whose original tick is still outstanding —
// after Push, that prior tick is silently ignored.
func (s Stack) Push(m Model) (Stack, tea.Cmd) {
	m.refreshGenID()
	if s.width > 0 {
		m.SetSize(s.width, 0)
	}
	s.items = append(s.items, m)
	if s.max > 0 && len(s.items) > s.max {
		s.items = evictOldestNonDismissed(s.items, s.max)
	}
	return s, m.Init()
}

// Update fans the message to every child notification, prunes any that
// have flipped to dismissed, and batches the resulting commands.
//
// The returned tea.Model wraps a fresh Stack value; callers that hold
// a typed reference should type-assert it back: `s = next.(Stack)`.
func (s Stack) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if len(s.items) == 0 {
		return s, nil
	}
	cmds := make([]tea.Cmd, 0, len(s.items))
	out := make([]Model, 0, len(s.items))
	for _, child := range s.items {
		next, cmd := child.Update(msg)
		nm, ok := next.(Model)
		if !ok {
			// Should never happen — Model.Update always returns Model.
			out = append(out, child)
			continue
		}
		if nm.Dismissed() {
			continue
		}
		out = append(out, nm)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	s.items = out
	return s, tea.Batch(cmds...)
}

// View renders the stack as a single block. Items are joined
// vertically with [lipgloss.JoinVertical] using right alignment, which
// suits typical corner placements (top-right or bottom-right). Callers
// place the block on screen with [lipgloss.Place] or by composing it
// into a parent layout. An empty stack renders as an empty view.
func (s Stack) View() tea.View {
	if len(s.items) == 0 {
		return tea.NewView("")
	}
	rendered := make([]string, 0, len(s.items))
	if s.newestFirst {
		for i := len(s.items) - 1; i >= 0; i-- {
			rendered = append(rendered, s.items[i].View().Content)
		}
	} else {
		for i := range s.items {
			rendered = append(rendered, s.items[i].View().Content)
		}
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Right, rendered...))
}

// Init returns nil. New child notifications schedule their own tick
// commands when added via [Stack.Push]; the Stack itself has no
// initialization work.
func (s Stack) Init() tea.Cmd {
	return nil
}

// Len reports the number of live (non-dismissed) notifications.
// Dismissed entries are pruned during [Stack.Update], so the count
// reflects what [Stack.View] would render.
func (s Stack) Len() int {
	return len(s.items)
}

// SetSize records the width applied to subsequently pushed
// notifications and propagates it to existing children. The height
// argument is ignored — the stack's height is the sum of its
// children's heights.
func (s *Stack) SetSize(width, height int) {
	s.width = width
	_ = height
	for i := range s.items {
		(&s.items[i]).SetSize(width, height)
	}
}

// Focus is a no-op. It exists only to satisfy
// [component.Component]: notifications are inert.
func (s *Stack) Focus() {}

// Blur is a no-op. It exists only to satisfy [component.Component]:
// notifications are inert.
func (s *Stack) Blur() {}

// evictOldestNonDismissed shrinks items down to max by dropping the
// first non-dismissed entry. If every entry is dismissed (which would
// be unusual since Update prunes dismissed items), the oldest entry
// is dropped instead.
func evictOldestNonDismissed(items []Model, max int) []Model {
	for len(items) > max {
		idx := 0
		for i, it := range items {
			if !it.Dismissed() {
				idx = i
				break
			}
		}
		items = append(items[:idx], items[idx+1:]...)
	}
	return items
}

// refreshGenID is a private helper invoked by [Stack.Push] to assign a
// fresh generation id to the incoming Model. Defined on Model in this
// file to keep the cross-file coupling explicit.
func (m *Model) refreshGenID() {
	m.genID = nextGenID()
}

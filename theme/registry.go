package theme

import (
	"image/color"
	"strings"
	"sync"

	"charm.land/lipgloss/v2"
)

// StatusEntry describes the visual rendering for a single status kind.
// Register new status kinds with [Registry.RegisterStatus]; unknown kinds
// are returned by [GetStatusStyle] with a "?" symbol.
type StatusEntry struct {
	// Kind is the canonical string key for this status (e.g. "working",
	// "workarea-warming").  Keys are compared case-insensitively.
	Kind string

	// Label is the human-readable display label (e.g. "Working").
	Label string

	// Symbol is the Unicode character used as the non-color status indicator
	// for accessible / NO_COLOR rendering (e.g. "●").
	Symbol string

	// Color is the theme color token for this status.  Choose a field from
	// an existing [Theme] value (e.g. t.StatusSuccess) or supply a custom
	// [image/color.Color].
	Color color.Color

	// Animate signals that the symbol should be rendered with a spinner
	// animation rather than as a static glyph.
	Animate bool
}

// WorkTypeEntry describes the visual rendering for a single work-type kind.
// Register new work-type kinds with [Registry.RegisterWorkType]; unknown
// kinds fall back to [Theme.TextSecondary] with a passthrough label.
type WorkTypeEntry struct {
	// Kind is the canonical string key for this work type (e.g.
	// "development", "workarea-acquire").  Keys are compared
	// case-insensitively.
	Kind string

	// Label is the human-readable display label (e.g. "Development").
	Label string

	// Color is the display color for chips, badges, and row highlights.
	Color color.Color
}

// ActivityEntry describes the visual rendering for a single activity kind.
// Register new activity kinds with [Registry.RegisterActivity]; unknown
// kinds fall back to [Theme.TextSecondary] with a "?" icon.
type ActivityEntry struct {
	// Kind is the canonical string key for this activity (e.g. "thought",
	// "tool-call").  Keys are compared case-insensitively.
	Kind string

	// Icon is the Unicode character or emoji used to decorate the activity
	// in log and timeline views.
	Icon string

	// Color is the display color for the activity label and icon.
	Color color.Color
}

// Registry is a thread-safe, open registry of status, work-type, and
// activity entries.  Third-party plugins and kits add new kinds by calling
// [RegisterStatus], [RegisterWorkType], or [RegisterActivity] at activation
// time.  The core tui-components library pre-registers all built-in kinds
// during [init].
//
// Use the package-level [GlobalRegistry] value rather than constructing
// your own Registry.
type Registry struct {
	mu         sync.RWMutex
	statuses   map[string]StatusEntry
	workTypes  map[string]WorkTypeEntry
	activities map[string]ActivityEntry
}

// GlobalRegistry is the package-level open registry.  Built-in status,
// work-type, and activity kinds are pre-registered during package init.
// Third-party callers register additional kinds at plugin/kit activation
// time.
//
//nolint:gochecknoglobals
var GlobalRegistry = newRegistry()

func newRegistry() *Registry {
	return &Registry{
		statuses:   make(map[string]StatusEntry),
		workTypes:  make(map[string]WorkTypeEntry),
		activities: make(map[string]ActivityEntry),
	}
}

// RegisterStatus adds or replaces a [StatusEntry] in the registry.
// Existing entries with the same Kind (case-insensitive) are overwritten,
// allowing plugins to override built-in defaults.
func (r *Registry) RegisterStatus(e StatusEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.statuses[strings.ToLower(e.Kind)] = e
}

// GetStatus returns the [StatusEntry] registered for kind.
// The second return value is false when the kind is not registered.
// Prefer [GetStatusStyle] for rendering; use this only when the full
// entry is needed.
func (r *Registry) GetStatus(kind string) (StatusEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.statuses[strings.ToLower(kind)]
	return e, ok
}

// ListStatuses returns a snapshot of all registered [StatusEntry] values.
// The order is undefined.
func (r *Registry) ListStatuses() []StatusEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]StatusEntry, 0, len(r.statuses))
	for _, e := range r.statuses {
		out = append(out, e)
	}
	return out
}

// RegisterWorkType adds or replaces a [WorkTypeEntry] in the registry.
// Existing entries with the same Kind (case-insensitive) are overwritten.
func (r *Registry) RegisterWorkType(e WorkTypeEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.workTypes[strings.ToLower(e.Kind)] = e
}

// GetWorkType returns the [WorkTypeEntry] registered for kind.
// The second return value is false when the kind is not registered.
func (r *Registry) GetWorkType(kind string) (WorkTypeEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.workTypes[strings.ToLower(kind)]
	return e, ok
}

// ListWorkTypes returns a snapshot of all registered [WorkTypeEntry] values.
// The order is undefined.
func (r *Registry) ListWorkTypes() []WorkTypeEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]WorkTypeEntry, 0, len(r.workTypes))
	for _, e := range r.workTypes {
		out = append(out, e)
	}
	return out
}

// RegisterActivity adds or replaces an [ActivityEntry] in the registry.
// Existing entries with the same Kind (case-insensitive) are overwritten.
func (r *Registry) RegisterActivity(e ActivityEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activities[strings.ToLower(e.Kind)] = e
}

// GetActivity returns the [ActivityEntry] registered for kind.
// The second return value is false when the kind is not registered.
func (r *Registry) GetActivity(kind string) (ActivityEntry, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	e, ok := r.activities[strings.ToLower(kind)]
	return e, ok
}

// ListActivities returns a snapshot of all registered [ActivityEntry] values.
// The order is undefined.
func (r *Registry) ListActivities() []ActivityEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]ActivityEntry, 0, len(r.activities))
	for _, e := range r.activities {
		out = append(out, e)
	}
	return out
}

// init pre-registers all built-in status, work-type, and activity entries
// into GlobalRegistry so callers start with a fully populated registry
// without any additional setup.
func init() { //nolint:gochecknoinits
	registerBuiltinStatuses()
	registerBuiltinWorkTypes()
	registerBuiltinActivities()
}

func registerBuiltinStatuses() {
	entries := []StatusEntry{
		{Kind: "working", Label: "Working", Symbol: "●", Color: pkg.StatusSuccess, Animate: true},
		{Kind: "queued", Label: "Queued", Symbol: "◌", Color: pkg.StatusWarning, Animate: true},
		{Kind: "parked", Label: "Parked", Symbol: "○", Color: pkg.TextTertiary, Animate: false},
		{Kind: "completed", Label: "Done", Symbol: "✓", Color: pkg.StatusSuccess, Animate: false},
		{Kind: "failed", Label: "Failed", Symbol: "✗", Color: pkg.StatusError, Animate: false},
		{Kind: "stopped", Label: "Stopped", Symbol: "■", Color: pkg.TextTertiary, Animate: false},
	}
	for _, e := range entries {
		GlobalRegistry.RegisterStatus(e)
	}
}

func registerBuiltinWorkTypes() {
	entries := []WorkTypeEntry{
		{Kind: "development", Label: "Development", Color: lipgloss.Color("#60A5FA")},
		{Kind: "bugfix", Label: "Bug Fix", Color: lipgloss.Color("#F87171")},
		{Kind: "feature", Label: "Feature", Color: lipgloss.Color("#34D399")},
		{Kind: "qa", Label: "QA", Color: lipgloss.Color("#A78BFA")},
		{Kind: "qa-coordination", Label: "QA Coord", Color: lipgloss.Color("#C4B5FD")},
		{Kind: "acceptance", Label: "Acceptance", Color: lipgloss.Color("#F472B6")},
		{Kind: "acceptance-coordination", Label: "Accept Coord", Color: lipgloss.Color("#F9A8D4")},
		{Kind: "coordination", Label: "Coordination", Color: lipgloss.Color("#FB923C")},
		{Kind: "research", Label: "Research", Color: lipgloss.Color("#2DD4BF")},
		{Kind: "backlog-creation", Label: "Backlog", Color: lipgloss.Color("#94A3B8")},
		{Kind: "inflight", Label: "Inflight", Color: lipgloss.Color("#FACC15")},
		{Kind: "refinement", Label: "Refinement", Color: lipgloss.Color("#A3E635")},
		{Kind: "refinement-coordination", Label: "Refine Coord", Color: lipgloss.Color("#BEF264")},
		{Kind: "refactor", Label: "Refactor", Color: lipgloss.Color("#FBBF24")},
		{Kind: "review", Label: "Review", Color: lipgloss.Color("#22D3EE")},
		{Kind: "docs", Label: "Docs", Color: lipgloss.Color("#818CF8")},
	}
	for _, e := range entries {
		GlobalRegistry.RegisterWorkType(e)
	}
}

func registerBuiltinActivities() {
	entries := []ActivityEntry{
		{Kind: "thought", Icon: "\U0001f4ad", Color: pkg.TextSecondary},
		{Kind: "action", Icon: "⚡", Color: pkg.Teal},
		{Kind: "response", Icon: "\U0001f4ac", Color: pkg.TextPrimary},
		{Kind: "error", Icon: "✗", Color: pkg.StatusError},
		{Kind: "progress", Icon: "✓", Color: pkg.StatusSuccess},
	}
	for _, e := range entries {
		GlobalRegistry.RegisterActivity(e)
	}
}

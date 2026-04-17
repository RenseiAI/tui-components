// Package notification provides a transient toast widget for the
// AgentFactory and Rensei TUI applications.
//
// A [Model] renders a single auto-dismissing status message in one of
// three variants — [VariantSuccess], [VariantWarning], or [VariantError]
// — mapped to the project status palette in
// [github.com/RenseiAI/tui-components/theme]. Variant colors are read
// from theme/palette.go; the package contains no hardcoded colors.
//
// Auto-dismiss is driven by [charm.land/bubbletea/v2.Tick]. Each Model
// carries a monotonic generation id; a tick that arrives after the
// instance has been replaced or its id bumped is silently ignored, so
// stale timers can never flip [Model.Dismissed] on the wrong toast.
//
// Both [Model] and [Stack] satisfy the
// [github.com/RenseiAI/tui-components/component.Component] interface so
// they slot directly into a Bubble Tea program. [Model.Focus] and
// [Model.Blur] exist to satisfy that contract and are intentionally
// no-ops — toasts are inert and dismiss only on their own timer.
//
// Positioning is intentionally caller-owned. [Stack.View] returns a
// single block of text composed with
// [charm.land/lipgloss/v2.JoinVertical]; the caller places it on screen
// with [charm.land/lipgloss/v2.Place] or by composing it into a parent
// layout. The package never assumes corner, overlay, or z-order.
package notification

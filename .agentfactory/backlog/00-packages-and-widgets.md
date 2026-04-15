# tui-components — Issue Backlog

## Extracted Packages (v0.1.0 — done)

These are already shipped in v0.1.0:
- ~~theme/palette.go — Color constants~~
- ~~theme/styles.go — Lipgloss style functions~~
- ~~theme/status.go — Status style mappings (string-based)~~
- ~~theme/worktype.go — Work type color/label mappings~~
- ~~theme/activity.go — Activity type color/icon mappings~~
- ~~format/format.go — Duration, Cost, RelativeTime, Timestamp, ProviderName, Tokens~~
- ~~format/format_test.go — Table-driven tests~~
- ~~component/component.go — Component interface~~

## New Widget Issues

### TC-001: Implement spinner/loader widget
**Priority**: High
**Labels**: shared-component

Animated spinner for async operations. Configurable spinner style, label text. Implements Component interface. Used during API calls, process startup.

### TC-002: Implement modal/dialog widget
**Priority**: High
**Labels**: shared-component

Confirmation dialog overlay. Yes/No/Cancel actions. Customizable title, body, button labels. Used for destructive operations (stop agent, clear queue, etc.).

### TC-003: Implement text input widget
**Priority**: High
**Labels**: shared-component

Styled text input with placeholder, validation callback, error display. Used for chat messages, issue creation, search filters.

### TC-004: Implement select/list widget
**Priority**: High
**Labels**: shared-component

Selection list with fuzzy filtering (using sahilm/fuzzy). Single and multi-select modes. Customizable item rendering. Used for project selection, command palette.

### TC-005: Implement progress bar widget
**Priority**: Medium
**Labels**: shared-component

Styled progress indicator with percentage, label, and ETA. Deterministic (known total) and indeterminate modes.

### TC-006: Implement tab bar widget
**Priority**: Medium
**Labels**: shared-component

Horizontal tab navigation with keyboard shortcuts. Active tab highlighting. Used for switching between views within a panel.

### TC-007: Implement log viewer widget
**Priority**: Medium
**Labels**: shared-component

Scrollable log output with ANSI color support. Auto-scroll to bottom with manual scroll lock. Line wrapping. Used for activity streams, log analysis.

### TC-008: Implement notification/toast widget
**Priority**: Low
**Labels**: shared-component

Transient message overlay. Auto-dismiss with configurable duration. Success/warning/error variants using theme status colors.

## Testing & Quality

### TC-010: Add tests for theme package
**Priority**: Medium
**Labels**: testing

Test GetStatusStyle for all known statuses + unknown. Test GetWorkTypeColor and GetWorkTypeLabel for all types + unknown. Test ActivityColors/ActivityIcons map completeness.

### TC-011: Expand format tests with edge cases
**Priority**: Medium
**Labels**: testing

Add tests for: negative durations, very large token counts (millions), malformed ISO timestamps, nil/zero edge cases for all format functions.

### TC-012: Add godoc examples for all exported functions
**Priority**: Low
**Labels**: testing

Add Example* test functions that serve as both documentation and runnable examples. Required for pkg.go.dev rendering.

## Infrastructure

### TC-020: Set up automated releases with tags
**Priority**: Medium
**Labels**: infra

Create release workflow that publishes Go module on tag push. Consider adding a changelog generator.

### TC-021: Add Go module proxy cache warming
**Priority**: Low
**Labels**: infra

After tagging, hit `proxy.golang.org` to warm the module cache so downstream consumers get fast resolution.

#!/bin/bash
# Push tui-components backlog issues to Linear icebox
# Usage: export $(grep -v '^#' /path/to/.env.local | grep -v '^$' | xargs) && bash .agentfactory/backlog/push-to-linear.sh

set -uo pipefail

TEAM="Rensei"
PROJECT="tui-components"
STATE="Icebox"
COUNT=0
ERRORS=0

create_issue() {
  local title="$1"
  local description="$2"
  local labels="${3:-}"

  local args=(--title "$title" --team "$TEAM" --project "$PROJECT" --state "$STATE")
  if [ -n "$description" ]; then
    args+=(--description "$description")
  fi
  if [ -n "$labels" ]; then
    args+=(--labels "$labels")
  fi

  result=$(af-linear create-issue "${args[@]}" 2>&1) || {
    echo "  ERROR: $result"
    ERRORS=$((ERRORS + 1))
    return 1
  }

  id=$(echo "$result" | grep -o '"identifier":"[^"]*"' | head -1 | cut -d'"' -f4)
  echo "  Created $id: $title"
  COUNT=$((COUNT + 1))
  sleep 0.5
}

echo "=== New Widgets ==="
create_issue "TC-001: Implement spinner/loader widget" "Animated spinner for async operations. Configurable spinner style, label text. Implements Component interface. Used during API calls, process startup." "shared-component"
create_issue "TC-002: Implement modal/dialog widget" "Confirmation dialog overlay. Yes/No/Cancel actions. Customizable title, body, button labels. Used for destructive operations (stop agent, clear queue, etc.)." "shared-component"
create_issue "TC-003: Implement text input widget" "Styled text input with placeholder, validation callback, error display. Used for chat messages, issue creation, search filters." "shared-component"
create_issue "TC-004: Implement select/list widget" "Selection list with fuzzy filtering (using sahilm/fuzzy). Single and multi-select modes. Customizable item rendering. Used for project selection, command palette." "shared-component"
create_issue "TC-005: Implement progress bar widget" "Styled progress indicator with percentage, label, and ETA. Deterministic (known total) and indeterminate modes." "shared-component"
create_issue "TC-006: Implement tab bar widget" "Horizontal tab navigation with keyboard shortcuts. Active tab highlighting. Used for switching between views within a panel." "shared-component"
create_issue "TC-007: Implement log viewer widget" "Scrollable log output with ANSI color support. Auto-scroll to bottom with manual scroll lock. Line wrapping. Used for activity streams, log analysis." "shared-component"
create_issue "TC-008: Implement notification/toast widget" "Transient message overlay. Auto-dismiss with configurable duration. Success/warning/error variants using theme status colors." "shared-component"

echo ""
echo "=== Testing & Quality ==="
create_issue "TC-010: Add tests for theme package" "Test GetStatusStyle for all known statuses + unknown. Test GetWorkTypeColor and GetWorkTypeLabel for all types + unknown. Test ActivityColors/ActivityIcons map completeness." "testing"
create_issue "TC-011: Expand format tests with edge cases" "Add tests for: negative durations, very large token counts (millions), malformed ISO timestamps, nil/zero edge cases for all format functions." "testing"
create_issue "TC-012: Add godoc examples for all exported functions" "Add Example* test functions that serve as both documentation and runnable examples. Required for pkg.go.dev rendering." "testing"

echo ""
echo "=== Infrastructure ==="
create_issue "TC-020: Set up automated releases with tags" "Create release workflow that publishes Go module on tag push. Consider adding a changelog generator." "infra"
create_issue "TC-021: Add Go module proxy cache warming" "After tagging, hit proxy.golang.org to warm the module cache so downstream consumers get fast resolution." "infra"

echo ""
echo "=== Done ==="
echo "Created: $COUNT issues"
echo "Errors: $ERRORS"

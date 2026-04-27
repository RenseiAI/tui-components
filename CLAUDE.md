# tui-components

OSS shared TUI component library for the AgentFactory ecosystem.

**Module**: `github.com/RenseiAI/tui-components`

## Architecture

Authoritative architecture lives in `../rensei-architecture/`. Read in this order:

1. `001-layered-execution-model.md` — canonical synthesis. Always first.
2. The reference doc(s) for whichever layer you are working on (`002`–`008`, `011`, `013`–`016`).
3. Any open ADRs that touch your work (`ADR-*.md`).

If this project's docs conflict with `../rensei-architecture/`, the corpus wins. Either update this project's docs to align, or open an ADR to amend the corpus.

## Boundary

This is an open-source project. It must never contain or reference proprietary platform features, API details, or closed-source concepts. All components must be generic and reusable. Downstream consumers extend these components — this library must remain platform-agnostic.

## Dependency Stack

Charm v2 ecosystem only:
- `charm.land/bubbletea/v2` — TUI framework
- `charm.land/lipgloss/v2` — Terminal styling
- `charm.land/bubbles/v2` — Reusable UI components (list, textarea, viewport, textinput, filepicker)
- `github.com/charmbracelet/log` — Structured logging

No other direct dependencies without compelling justification.

## Packages

- `theme/` — Color palette, Lipgloss styles, status/worktype/activity visual mappings
- `format/` — Human-readable formatting (duration, cost, relative time, tokens)
- `component/` — Bubble Tea `Component` interface (tea.Model + SetSize + Focus + Blur)
- `widget/` — Shared TUI widgets wrapping/extending Bubbles v2 components

## Commands

```bash
make build      # go build ./...
make test       # go test -race ./...
make lint       # golangci-lint run
make fmt        # gofumpt -w .
make vuln       # govulncheck ./...
make coverage   # test with coverage report
```

## Conventions

- **Dependencies**: Charm v2 stack only. Every new dep must be justified.
- **Exports**: All exported functions and types must have godoc comments.
- **Testing**: stdlib `testing` + table-driven tests. No testify. Golden files with `cupaloy` for complex output. Target 90% for format/theme, 80% overall.
- **Errors**: Wrap with `fmt.Errorf("context: %w", err)`. Return errors to callers. Never panic.
- **Logging**: `charmbracelet/log` to stderr. No log.Fatal.
- **Naming**: Lowercase single-word packages. PascalCase exports.
- **Formatting**: `gofumpt` (stricter gofmt). Enforced by `golangci-lint`.
- **Linting**: `golangci-lint` with govet, staticcheck, gofumpt, errcheck, gosec, gocritic, revive.
- **Status strings**: Use plain strings (not typed enums) to avoid import cycles.
- **Widgets**: Wrap Bubbles v2 components where applicable. Accept Bubbles options for customization. Implement `Component` interface. Read colors from `theme/` — no hardcoded colors.
- **Examples**: Godoc `Example*` tests live in `example_test.go` (one per package), declared with `package <pkg>` — same package as the code under test so examples exercise intra-package usage. Use `// Output:` on deterministic examples and `// Unordered output:` where map iteration is involved. Non-deterministic examples (wall-clock, locale) are compile-only with no output comment. Lipgloss-rendered examples omit `// Output:` (ANSI bytes differ across terminals). Use `fmt.Println` only — no `charmbracelet/log` in examples. No hardcoded colors — read from `theme/`.
- **Stability**: No breaking changes within a minor version.

## Project Layout

```
theme/          Color palette, styles, status/worktype mappings
format/         Formatting utilities with tests
component/      Component interface definition
widget/         Shared Bubble Tea widget models (extend Bubbles v2)
```

## Worktrees

- `.claude/settings.json` registers a `SessionStart` hook running `scripts/refresh-worktree.sh` — auto-rebases and refreshes deps on linked worktrees only.

# tui-components

OSS shared TUI component library for the AgentFactory and Rensei ecosystem.

**Module**: `github.com/RenseiAI/tui-components`

## Packages

- `theme/` — Color palette, Lipgloss styles, status/worktype/activity visual mappings
- `format/` — Human-readable formatting (duration, cost, relative time, tokens)
- `component/` — Bubble Tea `Component` interface (tea.Model + SetSize + Focus + Blur)
- `widget/` — Shared TUI widgets (statsbar, table, card, helpbar)

## Commands

```bash
make build      # go build ./...
make test       # go test ./...
make lint       # go vet ./...
make fmt        # gofumpt -w .
make coverage   # test with coverage report
```

## Conventions

- **Dependencies**: Only Bubble Tea v2 and Lipgloss v2. No other direct dependencies.
- **Exports**: All exported functions and types must have godoc comments.
- **Testing**: All exported functions must have table-driven tests. Target 80% coverage.
- **Errors**: Wrap with `fmt.Errorf("context: %w", err)`. Return errors to callers.
- **Naming**: Lowercase single-word packages. PascalCase exports.
- **Formatting**: Use `gofumpt` (stricter gofmt).
- **Status strings**: Use plain strings (not typed enums) for status parameters to avoid import cycles with consuming repos.
- **Stability**: No breaking changes within a minor version.

## Project Layout

```
theme/          Color palette, styles, status/worktype mappings
format/         Formatting utilities with tests
component/      Component interface definition
widget/         Shared Bubble Tea widget models
```

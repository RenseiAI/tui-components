# tui-components — OSS shared TUI component library (OSS-public)

Go, module `github.com/RenseiAI/tui-components`. Four packages: `theme/`
(color palette, Lipgloss styles, status/worktype/activity visual mappings),
`format/` (duration, cost, relative time, tokens), `component/` (the Bubble Tea
`Component` interface: `tea.Model` + SetSize + Focus + Blur), `widget/`
(shared widgets wrapping/extending Bubbles v2, incl. `widget/notification/`).
Charm v2 stack. Consumed by the `donmai` binary and other downstream embedders.

## Operating context

- Governing corpus: `../donmai-architecture/` (public). Read order:
  `001-layered-execution-model.md` first → the layer doc(s) for your area
  (`002`–`008`, `011`, `013`–`016`) → open `ADR-*.md`. The corpus wins over
  code: align the code or open an ADR. Missing? `gh repo clone
  RenseiAI/donmai-architecture ../donmai-architecture` (from a worktree,
  siblings sit at `../../<repo>`).
- Code work happens in a sibling worktree: `scripts/create-worktree.sh <name>`
  → `../tui-components.wt/<name>`. A `SessionStart` hook
  (`.claude/settings.json`) runs `scripts/refresh-worktree.sh` on linked
  worktrees only.
- Releases are tags on `main`: `.github/workflows/release.yml` re-runs the CI
  gates, generates notes via `git-cliff` from conventional-commit prefixes,
  and fails if the module does not resolve on `proxy.golang.org`. Policy and
  the breaking-change definition live in `RELEASING.md`.

## Before you start — read in this order

| The moment you... | Read |
|---|---|
| start ANY task in this repo | this file, top to bottom (it is short) |
| are about to add, remove, or re-sign an exported symbol in `theme/` `format/` `component/` `widget/` | `RELEASING.md` (semver policy — what counts as breaking) |
| change a visual mapping, widget contract, or anything embedders render | the corpus read order above |
| are about to add a dependency to `go.mod` | §Iron rules (Charm v2 only) |
| are about to write "done"/"fixed" or open a PR | Gates below + `../donmai-architecture/agents/PROTOCOL.md` §V |
| hit a failing test or `-race` flake you did not predict | `../donmai-architecture/agents/PROTOCOL.md` §D |
| edit README, godoc, or anything shipped publicly | §Boundary below |

When a row matches, read that doc before your next edit and follow it literally.

## Gates — "done" means these passed

```bash
make build            # go build ./...           (compile gate)
make test             # go test -race ./...      (the race flag is mandatory)
make lint             # golangci-lint run
make check-examples   # TestExportedSymbolsHaveExamples — CI-required, easy to miss
make fmt              # gofumpt -w .
```

CI (`.github/workflows/ci.yml`) runs build, race tests + coverage summary,
golangci-lint (v2.11.4), `make check-examples`, and govulncheck
(continue-on-error). Also available: `make vuln`, `make coverage` (thresholds
below), `make clean`. Run the gates after your last edit and quote each result
line in your report.

## Iron rules

- Dependencies: Charm v2 only — `charm.land/{bubbletea,bubbles,lipgloss}/v2` +
  `github.com/charmbracelet/log`; no other direct dep without a stated
  compelling justification (every dep ships to all embedders).
- Every exported symbol carries a godoc comment AND an `Example*` test in the
  package's `example_test.go` (`make check-examples` fails CI otherwise).
- Examples use `fmt.Println` only — no `charmbracelet/log`; `// Output:` on
  deterministic ones, `// Unordered output:` over map iteration; Lipgloss-
  rendered and wall-clock/locale examples omit the output comment (ANSI bytes
  and clocks differ across environments).
- Tests: stdlib `testing`, table-driven, no testify; `cupaloy` golden files for
  complex rendered output.
- Coverage: 90% for `format/` and `theme/`, 80% overall (`make coverage`).
- Errors: `fmt.Errorf("context: %w", err)`, returned to callers; never `panic`
  (this is a library embedders link).
- Logging: `charmbracelet/log` to stderr; never `log.Fatal`.
- Widgets wrap Bubbles v2 components where applicable, accept Bubbles options
  for customization, and implement `component.Component`.
- Widgets read every color from `theme/` — zero hardcoded colors (breaks
  theme switching and a11y mode).
- Status strings stay plain strings, not typed enums (typed enums force
  import cycles across consumers).
- No breaking changes within a minor version; `RELEASING.md` defines
  "breaking" — prefer deprecate-then-remove across two minors.
- Formatting is `gofumpt` (enforced by lint); lowercase single-word packages,
  PascalCase exports.

## Boundary — this repo is public and platform-agnostic

- Never reference proprietary platform features, closed-source concepts,
  private tracker IDs, internal hostnames, or secrets — components stay
  generic and reusable; downstream embedders extend them.
- `RENSEI_A11Y=true` is the SHIPPED accessibility env toggle (`theme/a11y.go`;
  takes precedence over `NO_COLOR`). It is current public API and a known
  debrand candidate — do not rename it yourself; renaming needs an ADR plus a
  deprecation window.
- No automated leak guard exists in this repo — this section is the guard;
  re-read it before pushing anything that ships (README, godoc, examples).

## Gotchas

- `make check-examples` is a separate CI step that `make test` does NOT cover —
  a green test run can still fail CI on a missing `Example*`.
- `coverage.out` is tracked at the repo root; `make coverage` rewrites it and
  `make clean` deletes it — keep its churn out of unrelated commits.
- Release notes come from conventional-commit prefixes (`feat:`/`fix:`/`perf:`/
  `docs:`/`chore:` — see `RELEASING.md`); a mis-prefixed commit lands in the
  wrong changelog section.

## Hard stops

- NEVER remove/rename an exported symbol or change an exported signature in a
  minor release -> instead: deprecate, then remove across two minors per
  `RELEASING.md`.
- NEVER add a non-Charm direct dependency without stating the justification in
  the PR -> instead: propose it first.
- NEVER rename `RENSEI_A11Y` -> instead: open an ADR in
  `../donmai-architecture` proposing the migration.
- NEVER make a failing check pass by weakening it (skip, deleted test,
  loosened assert, lint-disable) -> instead: quote the failure and propose the
  change.
- NEVER run `git worktree remove/prune`, `git reset --hard`, `git clean -fd`,
  or checkout to another branch as a sub-agent -> instead: the orchestrator
  owns worktree lifecycle.

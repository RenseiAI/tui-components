# Changelog

All notable changes to `tui-components` are documented here.
Release notes are generated from conventional commits via `git-cliff` (see `cliff.toml`).
For the full release workflow, see `RELEASING.md`.

---

## [v0.2.0] — In progress

Architecture-aware primitives milestone. All items below land via dependent issues
(REN-1319, REN-1330, REN-1331, REN-1332) in a single coordinated breaking-change release.

### Breaking changes

- **Theme system overhaul (REN-1319) — LANDED:** `theme.BgPrimary`, `theme.Accent`, and all
  palette-level `var` declarations have been removed and replaced by a swappable `Theme`
  struct in `theme/theme.go`.  Every widget accepts a theme via per-widget `WithXxxTheme`
  options (e.g. `WithSpinnerTheme`, `WithProgressbarTheme`) and the universal
  `widget.WithTheme(t).Spinner()` / `.Progressbar()` / `.Dialog()` / `.Tabs()` helpers.
  `DefaultTheme()`, `DarkTheme()`, and `HighContrastTheme()` constructors are provided.
  Hot-swap is supported: call `SetTheme(t)` on any widget instance to update mid-render.
  `theme.Default()` returns a pointer to the package-level default Theme for legacy callers.
  See `MIGRATION-v0.2.0.md §1` for the mechanical migration steps.

- **Open capability registries (REN-1330) — LANDED:** `theme.GetStatusStyle`,
  `theme.GetWorkTypeColor`, and `theme.GetActivityIcon` previously used closed
  switch / map literals over a fixed set of string keys.  These are replaced by
  a thread-safe open registry API backed by `theme.GlobalRegistry` (`*Registry`).
  Plugins and kits register new kinds at activation time via
  `GlobalRegistry.RegisterStatus(StatusEntry{…})`,
  `GlobalRegistry.RegisterWorkType(WorkTypeEntry{…})`, and
  `GlobalRegistry.RegisterActivity(ActivityEntry{…})`.  All 16 built-in work
  types, 6 built-in status kinds, and 5 built-in activity kinds are
  pre-registered during `package init`.  Unknown kinds return a fallback style
  ("?" symbol, `TextSecondary` color, `"Unknown"` label) rather than a silent
  zero value.  `GetActivityColor(kind)` and `GetActivityIcon(kind)` are new
  registry-backed helpers; the legacy `ActivityColors` and `ActivityIcons` map
  vars are retained (deprecated) for backward compat.  Tests assert no closed
  switches remain (`TestNoClosedSwitches`).

### New primitives — architecture-concept layer (REN-1331) — LANDED

Primitives from `014-tui-operator-surfaces.md`.  All 13 widgets shipped in
`widget/` with tests and godoc examples.  Each accepts `WithXxxTheme(t)` for
theme swap and exposes `AccessibleLabel()` + `WithXxxNoColor(true)` for a11y:

| Primitive | File | Spec ref | Description |
|---|---|---|---|
| `CapabilityChip` | `capability_chip.go` | `002` | Typed flag + human label chip |
| `ScopePill` | `scope_pill.go` | `002` | `project \| org \| tenant \| global` scope indicator |
| `AttestationChip` | `attestation_chip.go` | `002` | Signed/unsigned/verified state with fingerprint |
| `ProviderHealthDot` | `provider_health_dot.go` | `002` | ready/degraded/unhealthy indicator dot |
| `WorkerRow` | `worker_row.go` | `013` | Single worker with status, region, load, billing model |
| `FleetGrid` | `fleet_grid.go` | `013` | Grid of WorkerRows grouped by machine/daemon |
| `MachinePivot` | `machine_pivot.go` | `013` | Multi-machine breakdown for SaaS aggregation |
| `SandboxCapacityGauge` | `sandbox_capacity_gauge.go` | `004` | Concurrent/max with utilization bar; degenerates to `∞` |
| `WorkareaPoolPanel` | `workarea_pool_panel.go` | `003`/`004` | Warm/cold/in-use slot breakdown per (repo, toolchain) |
| `KitDetectResult` | `kit_detect_result.go` | `005` | Kit match list with ordering and conflict indicators |
| `ToolchainChip` | `toolchain_chip.go` | `004`/`005` | `java=17`, `node=20.x` toolchain demand or workarea state |
| `AuditEntry` | `audit_entry.go` | Layer 6 | Signed event row with attestation + timestamp |
| `AuditChain` | `audit_chain.go` | Layer 6 | Composed AuditEntry list with chain integrity indicator |
| `PolicyDecisionBanner` | `policy_decision_banner.go` | Layer 6 | allowed/blocked/needs-approval banner |
| `CostPanel` | `cost_panel.go` | Layer 6 / `006` | Per-session/issue/tenant cost breakdown with trend |

### New format helpers (REN-1332) — LANDED

| Helper | Description |
|---|---|
| `format.CapacityRatio` | `"5 / 8"` or `"5 / ∞"` for null max-concurrent |
| `format.AttestationFingerprint` | `"ed25519:abc1234…d4f2"` truncated fingerprint |
| `format.RegionList` | `"iad1, +3 more"` compact list with overflow |
| `format.ToolchainSpec` | `"java=17, node=20"` multi-toolchain rendering |
| `format.HumanLabel[T]` | Generic typed-flag → human-readable string lookup |

### Accessibility opt-in (REN-1332) — LANDED

- `theme.A11yMode` type on the `Theme` struct; three values: `A11yNone` (default),
  `A11yNoColor` (color suppressed, Unicode symbols kept), `A11yFull` (verbose ASCII labels +
  high-contrast color tokens).
- `theme.A11yModeFromEnv()` detects `NO_COLOR` (any non-empty value → `A11yNoColor`) and
  `RENSEI_A11Y=true` (→ `A11yFull`) at startup.  Detection order: RENSEI_A11Y wins.
- `theme.Theme.WithA11y(mode)` returns an updated copy; `A11yFull` also replaces all color
  tokens with `HighContrastTheme()` values in the same call, so widgets automatically render
  at the correct contrast level without further intervention.
- `theme.Theme.RenderSymbol(symbol, verboseLabel)` selects between Unicode glyph and ASCII
  label based on the theme's `A11y` field — the single gate widgets must call.
- `theme.Theme.NoColor()` reports whether color application should be suppressed (`true` for
  both `A11yNoColor` and `A11yFull`).
- A11y is fully theme-driven: widgets read `theme.A11y` — never `os.Getenv` directly — so
  tests and server-side renderers can override the mode without touching the process environment.

---

## [v0.1.0] — 2026-04-27

Initial release. 53 issues completed.

### Packages shipped

- `theme/` — Color palette (`palette.go`), Lipgloss styles (`styles.go`),
  status/worktype/activity visual mappings.
- `format/` — Duration, cost, relative time, timestamp, provider name, token
  formatting with full edge-case test coverage.
- `component/` — `Component` interface (`tea.Model` + `SetSize` + `Focus`/`Blur`).
- `widget/` — Spinner, dialog, text input, select/list, progress bar, tab bar,
  log viewer, notification/toast.

### Infrastructure

- Tag-driven release workflow with `git-cliff` release notes.
- Go module proxy cache warming on release.
- golangci-lint v2 with gofumpt enforcement.
- Godoc examples for all exported symbols.

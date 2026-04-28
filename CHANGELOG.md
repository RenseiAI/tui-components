# Changelog

All notable changes to `tui-components` are documented here.
Release notes are generated from conventional commits via `git-cliff` (see `cliff.toml`).
For the full release workflow, see `RELEASING.md`.

---

## [v0.2.0] — In progress

Architecture-aware primitives milestone. All items below land via dependent issues
(REN-1319, REN-1330, REN-1331, REN-1332) in a single coordinated breaking-change release.

### Breaking changes

- **Theme system overhaul (REN-1319):** `theme.BgPrimary`, `theme.Accent`, and all
  palette-level `var` declarations move into a swappable `Theme` struct.
  Every widget that reads palette vars directly must be updated to accept a `Theme`
  via `widget.WithTheme(t)`.  A `DefaultTheme()` constructor preserves the current
  palette so existing consumers can migrate with a one-line change.

- **Open capability registries (REN-1330):** `theme.GetStatusStyle`,
  `theme.GetWorkTypeColor`, and `theme.GetActivityIcon` previously used closed
  switch statements over a fixed set of string keys.  These are replaced by a
  registry API (`themeRegistry.RegisterStatus`, `…RegisterWorkType`,
  `…RegisterActivity`).  Callers that switch on known keys continue to work;
  callers that relied on the exhaustive closed switch for compile-time coverage
  must migrate to registry lookups with the provided fallback rendering.

### New primitives — architecture-concept layer (REN-1331)

Primitives from `014-tui-operator-surfaces.md`:

| Primitive | Spec ref | Description |
|---|---|---|
| `CapabilityChip` | `002` | Typed flag + human label chip |
| `ScopePill` | `002` | `project | org | tenant | global` scope indicator |
| `AttestationChip` | `002` | Signed/unsigned/verified state with fingerprint |
| `WorkerRow` | `013` | Single worker with status, region, load, billing model |
| `FleetGrid` | `013` | Grid of WorkerRows grouped by machine/daemon |
| `MachinePivot` | `013` | Multi-machine breakdown for SaaS aggregation |
| `WorkareaPoolPanel` | `003`/`004` | Warm/cold/in-use slot breakdown per (repo, toolchain) |
| `KitDetectResult` | `005` | Kit match list with ordering and conflict indicators |
| `ToolchainChip` | `004`/`005` | `java=17`, `node=20.x` toolchain demand or workarea state |
| `AuditEntry` | Layer 6 | Signed event row with attestation + timestamp |
| `AuditChain` | Layer 6 | Composed AuditEntry list with chain integrity indicator |
| `PolicyDecisionBanner` | Layer 6 | allowed/blocked/needs-approval banner |
| `CostPanel` | Layer 6 / `006` | Per-session/issue/tenant cost breakdown with trend |

### New format helpers (REN-1332)

| Helper | Description |
|---|---|
| `format.CapacityRatio` | `"5 / 8"` or `"5 / ∞"` for null max-concurrent |
| `format.AttestationFingerprint` | `"ed25519:abc1234…d4f2"` truncated fingerprint |
| `format.RegionList` | `"iad1, +3 more"` compact list with overflow |
| `format.ToolchainSpec` | `"java=17, node=20"` multi-toolchain rendering |
| `format.HumanLabel[T]` | Generic typed-flag → human-readable string lookup |

### Accessibility opt-in (REN-1332)

- Honor `NO_COLOR` env var: force symbol-first rendering when set.
- `RENSEI_A11Y=true` env var or `--a11y` flag: high-contrast theme + verbose labels.
- Every new primitive declares an `accessibleLabel` field for screen-reader consumers.

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

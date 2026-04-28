# Migration guide — tui-components v0.1.0 → v0.2.0

v0.2.0 is a coordinated breaking-change release landing Theme struct refactoring,
open capability registries, architecture-concept primitives, format helpers, and
accessibility opt-in in a single push.  The sections below list every change that
requires consumer action and the corresponding mechanical migration step.

Consumers: `agentfactory-tui`, `rensei-tui`.

> **Status:** skeleton — details fill in as REN-1319, REN-1330, REN-1331, REN-1332 land.

---

## 1. Theme system (REN-1319)

### What changed

All top-level palette `var` declarations (`theme.BgPrimary`, `theme.Accent`, etc.)
move into a `Theme` struct with constructor functions.  Widgets no longer read
package-level vars; they accept a `Theme` via the `WithTheme` option.

### Migration steps

**Before**

```go
import "github.com/RenseiAI/tui-components/theme"

// Direct package-level color references
style := lipgloss.NewStyle().Background(theme.BgPrimary)
```

**After**

```go
import "github.com/RenseiAI/tui-components/theme"

t := theme.DefaultTheme()
style := lipgloss.NewStyle().Background(t.BgPrimary)
```

For widgets:

```go
// Before (implicit palette)
spinner := widget.NewSpinner()

// After (explicit theme)
spinner := widget.NewSpinner(widget.WithTheme(theme.DefaultTheme()))
```

**Available theme constructors**

| Constructor | Description |
|---|---|
| `theme.DefaultTheme()` | Current palette; zero migration cost |
| `theme.DarkTheme()` | Explicit dark variant |
| `theme.HighContrastTheme()` | A11y high-contrast |

### Scope

- All `agentfactory-tui` and `rensei-tui` widget instantiation sites.
- Custom `lipgloss.Style` definitions that reference `theme.*` package vars.

---

## 2. Open capability registries (REN-1330)

### What changed

`theme.GetStatusStyle(kind string)`, `theme.GetWorkTypeColor(kind string)`, and
`theme.GetActivityIcon(kind string)` previously used closed switch statements.
They now delegate to a `ThemeRegistry` that accepts `RegisterStatus`,
`RegisterWorkType`, and `RegisterActivity` calls.  Unknown kinds return a
fallback style with a `"?"` symbol and emit a one-time warning log.

### Migration steps

No import-path changes.  The call signatures for `GetStatusStyle`, etc., are
unchanged.  Migration is only required if you:

1. **Relied on compile-time exhaustiveness of the closed switch** — replace
   `switch kind { ... }` patterns with registry lookups.
2. **Need to register domain-specific states** — call `RegisterStatus` / etc.
   at `init()` or program startup.

**Example: registering a domain-specific status**

```go
func init() {
    theme.Registry.RegisterStatus(theme.StatusEntry{
        Kind:    "workarea-warming",
        Label:   "Warming pool",
        Symbol:  "↻",
        Color:   theme.DefaultTheme().StatusInfo,
        Animate: true,
    })
}
```

---

## 3. Architecture-concept primitives (REN-1331)

### What changed

New primitives added under `widget/`.  No existing symbols removed or renamed.
This section is informational for consumers adding new views.

See `docs/v0.2.0-scope.md` for the full primitive list and spec references.

### No migration required

No existing code breaks.  Adopt primitives as you build new panels.

---

## 4. Format helpers (REN-1332)

### What changed

New helpers added to `format/`:
`CapacityRatio`, `AttestationFingerprint`, `RegionList`, `ToolchainSpec`,
`HumanLabel[T]`.

No existing helpers are removed or changed.

### No migration required

No existing code breaks.  Replace any ad-hoc local formatting with the canonical
helpers for consistency.

---

## 5. Accessibility opt-in (REN-1332)

### What changed

- `NO_COLOR` env var is now honored: all widgets degrade to symbol-first
  rendering with no ANSI color codes.
- `RENSEI_A11Y=true` env var or `--a11y` CLI flag forces high-contrast theme +
  verbose label-only rendering.
- Every new primitive exposes an `AccessibleLabel` field on its model.

### Migration steps

No breaking changes to existing widgets.  If you render widgets in a context
where `NO_COLOR` was already set, visual output changes — this is intentional
and correct.

For new views, populate `AccessibleLabel` on all architecture-concept primitives.

---

## Upgrade checklist

- [ ] Bump `go.mod` to `github.com/RenseiAI/tui-components v0.2.0`.
- [ ] Replace all `theme.<PaletteVar>` references with `t.<PaletteVar>` from a
      `theme.DefaultTheme()` (or other) instance — **agentfactory-tui**.
- [ ] Replace all `theme.<PaletteVar>` references with `t.<PaletteVar>` from a
      `theme.DefaultTheme()` instance — **rensei-tui**.
- [ ] Add `widget.WithTheme(t)` to all widget constructors — both consumers.
- [ ] Audit any local `GetStatusStyle` / `GetWorkTypeColor` callers for switch
      exhaustiveness assumptions.
- [ ] Register any domain-specific statuses / work types / activity types with
      the new registry API.
- [ ] Set `AccessibleLabel` on all new primitives in both consumers.

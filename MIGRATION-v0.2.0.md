# Migration guide — tui-components v0.1.0 → v0.2.0

v0.2.0 is a coordinated breaking-change release landing Theme struct refactoring,
open capability registries, architecture-concept primitives, format helpers, and
accessibility opt-in in a single push.  The sections below list every change that
requires consumer action and the corresponding mechanical migration step.

Consumers: `agentfactory-tui`, `rensei-tui`.

> **Status:** REN-1319 (Theme struct + swap) landed. REN-1330, REN-1331, REN-1332 in progress.

---

## 1. Theme system (REN-1319) — LANDED

### What changed

All top-level palette `var` declarations (`theme.BgPrimary`, `theme.Accent`, etc.)
have been **removed** and replaced by a `Theme` struct in `theme/theme.go`.
Widgets no longer read package-level vars; each widget carries its own `Theme` value
and exposes a `WithXxxTheme` functional option and a `SetTheme` method for hot-swap.

The universal `widget.WithTheme(t)` helper constructs a `ThemeOption` value whose
per-widget converter methods (`.Spinner()`, `.Progressbar()`, `.Dialog()`, `.Tabs()`)
return the correct per-widget option type:

```go
opt := widget.WithTheme(theme.DarkTheme())
sp  := widget.NewSpinner(opt.Spinner())
bar := widget.NewProgressbar(opt.Progressbar())
```

### Migration steps

**Step 1 — Custom style definitions**

```go
// Before (v0.1.0) — package-level vars
style := lipgloss.NewStyle().Background(theme.BgPrimary)

// After (v0.2.0) — Theme field
t := theme.DefaultTheme()
style := lipgloss.NewStyle().Background(t.BgPrimary)
```

**Step 2 — Widget construction**

```go
// Before (v0.1.0) — implicit palette
spinner := widget.NewSpinner()

// After (v0.2.0) — explicit theme via per-widget option
spinner := widget.NewSpinner(widget.WithSpinnerTheme(theme.DefaultTheme()))

// Or via universal helper
spinner := widget.NewSpinner(widget.WithTheme(theme.DefaultTheme()).Spinner())
```

**Step 3 — Theme hot-swap (new in v0.2.0)**

```go
sp := widget.NewSpinner()
// ... later, when tenant theme changes ...
sp.SetTheme(tenantTheme) // next View() call uses tenantTheme
```

**Available theme constructors**

| Constructor | Description |
|---|---|
| `theme.DefaultTheme()` | Historic palette; zero migration cost for existing consumers |
| `theme.DarkTheme()` | True-black dark variant for OLED/pure-black terminals |
| `theme.HighContrastTheme()` | WCAG-AA high-contrast; also used with `RENSEI_A11Y=true` |

**Backward-compat bridge (`theme.Default()`)**

`theme.Default()` returns a `*theme.Theme` pointing at the package-level default
(initialised to `DefaultTheme()`).  Legacy code that cannot be migrated immediately
can read `theme.Default().BgPrimary` etc. without breaking; however, prefer the
explicit-theme pattern for all new code.

### Removed symbols

The following package-level `var` declarations are gone.  Replace each reference
with the corresponding `Theme` field:

| Removed | Replacement |
|---|---|
| `theme.BgPrimary` | `t.BgPrimary` |
| `theme.BgSecondary` | `t.BgSecondary` |
| `theme.BgTertiary` | `t.BgTertiary` |
| `theme.Surface` | `t.Surface` |
| `theme.SurfaceRaised` | `t.SurfaceRaised` |
| `theme.SurfaceBorder` | `t.SurfaceBorder` |
| `theme.SurfaceBorderBright` | `t.SurfaceBorderBright` |
| `theme.Accent` | `t.Accent` |
| `theme.AccentDim` | `t.AccentDim` |
| `theme.Teal` | `t.Teal` |
| `theme.TealDim` | `t.TealDim` |
| `theme.Blue` | `t.Blue` |
| `theme.StatusSuccess` | `t.StatusSuccess` |
| `theme.StatusWarning` | `t.StatusWarning` |
| `theme.StatusError` | `t.StatusError` |
| `theme.TextPrimary` | `t.TextPrimary` |
| `theme.TextSecondary` | `t.TextSecondary` |
| `theme.TextTertiary` | `t.TextTertiary` |

Where `t` is a value obtained from `theme.DefaultTheme()` (or whichever variant you
choose).

### New symbols for downstream consumers (REN-1330, REN-1331, REN-1332)

| Symbol | Package | Description |
|---|---|---|
| `theme.Theme` | `theme` | Value type carrying all color tokens |
| `theme.DefaultTheme()` | `theme` | Default palette constructor |
| `theme.DarkTheme()` | `theme` | Dark variant constructor |
| `theme.HighContrastTheme()` | `theme` | High-contrast constructor |
| `theme.Default()` | `theme` | Pointer to package-level default (bridge) |
| `widget.WithTheme(t)` | `widget` | Universal ThemeOption builder |
| `widget.WithSpinnerTheme(t)` | `widget` | SpinnerOption for theme |
| `widget.WithProgressbarTheme(t)` | `widget` | ProgressbarOption for theme |
| `widget.WithDialogTheme(t)` | `widget` | Dialog Option for theme |
| `widget.WithTabsTheme(t)` | `widget` | TabsOption for theme |

### Scope

- All `agentfactory-tui` and `rensei-tui` widget instantiation sites.
- Custom `lipgloss.Style` definitions that reference old `theme.*` package vars.

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

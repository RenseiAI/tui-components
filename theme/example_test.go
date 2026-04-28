package theme

import (
	"fmt"

	"charm.land/lipgloss/v2"
)

func ExampleGetStatusStyle() {
	// Iterate in a stable slice order so the verified Output: block is
	// deterministic (map iteration is not).
	statuses := []string{
		"working",
		"queued",
		"parked",
		"completed",
		"failed",
		"stopped",
		"mystery", // unknown → fallback
	}
	for _, s := range statuses {
		st := GetStatusStyle(s)
		fmt.Println(st.Label, st.Symbol)
	}
	// Output:
	// Working ●
	// Queued ◌
	// Parked ○
	// Done ✓
	// Failed ✗
	// Stopped ■
	// Unknown ?
}

func ExampleStatusStyle() {
	// Field-access pattern against the StatusStyle returned by
	// GetStatusStyle. Color is omitted because image/color.Color does
	// not format cleanly for verified output.
	st := GetStatusStyle("working")
	fmt.Println(st.Label)
	fmt.Println(st.Symbol)
	fmt.Println(st.Animate)
	// Output:
	// Working
	// ●
	// true
}

func ExampleGetWorkTypeColor() {
	// Case-insensitive lookup: "Bugfix" and "bugfix" resolve to the
	// same entry; an unknown key falls back to pkg.TextSecondary.
	// Compile-only because image/color.Color does not format cleanly.
	_ = GetWorkTypeColor("Bugfix")
	_ = GetWorkTypeColor("bugfix")
	c := GetWorkTypeColor("unknown-type") // → pkg.TextSecondary
	_ = c
}

func ExampleGetWorkTypeLabel() {
	fmt.Println(GetWorkTypeLabel("qa-coordination"))
	fmt.Println(GetWorkTypeLabel("something-unknown"))
	// Output:
	// QA Coord
	// something-unknown
}

func ExampleHeader() {
	// Representative example for the whole family of style constructors
	// in styles.go (Header, StatLabel, StatValue, TableHeader, HelpBar,
	// CardBorder, TabActive, SpinnerStyle, …). Compile-only because
	// lipgloss emits ANSI bytes that differ across terminals.
	_ = Header().Render("session title")
}

func ExampleTheme() {
	// Theme is a value type — create with a constructor and use fields directly.
	// Compile-only: lipgloss emits ANSI bytes that differ across terminals.
	t := DefaultTheme()
	style := lipgloss.NewStyle().
		Foreground(t.TextPrimary).
		Background(t.Surface)
	_ = style.Render("session title")
}

func ExampleDefault() {
	// Default returns a pointer to the package-level Theme. Use it when
	// transitioning legacy code that referenced the old package-level vars.
	// Prefer passing an explicit Theme to widgets via widget.WithTheme.
	t := Default()
	_ = t.Accent
}

func ExampleDefaultTheme() {
	// Construct a Theme and use it for styling — the canonical v0.2.0 pattern.
	t := DefaultTheme()
	style := lipgloss.NewStyle().
		Foreground(t.TextPrimary).
		Background(t.Surface)
	_ = style.Render("session title")
}

func ExampleDarkTheme() {
	// DarkTheme produces a true-black dark variant. Compile-only.
	t := DarkTheme()
	_ = t.BgPrimary
}

func ExampleHighContrastTheme() {
	// HighContrastTheme produces a WCAG-AA-compliant high-contrast variant.
	// Compile-only.
	t := HighContrastTheme()
	_ = t.BgPrimary
}

func ExampleA11yMode() {
	// A11yMode is embedded in a Theme via WithA11y. The zero value is A11yNone.
	t := DefaultTheme().WithA11y(A11yNoColor)
	// NoColor() reports whether color should be suppressed.
	_ = t.NoColor() // true for A11yNoColor and A11yFull
}

func ExampleA11yNone() {
	// A11yNone is the default: Unicode symbols + full color.
	t := DefaultTheme() // A11y defaults to A11yNone
	_ = t.A11y          // A11yNone (0)
}

func ExampleA11yNoColor() {
	// A11yNoColor suppresses color output, following the NO_COLOR spec.
	t := DefaultTheme().WithA11y(A11yNoColor)
	_ = t.NoColor() // true
}

func ExampleA11yFull() {
	// A11yFull enables high-contrast theme + verbose label-only output.
	t := DefaultTheme().WithA11y(A11yFull)
	label := t.RenderSymbol("✓", "[OK]") // returns "[OK]" in A11yFull mode
	_ = label
}

func ExampleA11yModeFromEnv() {
	// Detect the a11y mode from RENSEI_A11Y / NO_COLOR env vars and embed
	// into a Theme. Call once at startup; do not call per-render.
	t := DefaultTheme().WithA11y(A11yModeFromEnv())
	_ = t
}

func ExampleGetActivityColor() {
	// GetActivityColor queries GlobalRegistry and returns a fallback color
	// for unknown kinds. Compile-only: image/color.Color does not format
	// cleanly for verified output.
	_ = GetActivityColor("action")
	_ = GetActivityColor("unknown-activity") // → pkg.TextSecondary fallback
}

func ExampleGetActivityIcon() {
	fmt.Println(GetActivityIcon("error"))
	fmt.Println(GetActivityIcon("progress"))
	fmt.Println(GetActivityIcon("unknown-activity")) // fallback → "?"
	// Output:
	// ✗
	// ✓
	// ?
}

func Example() {
	// Intra-package composition: Theme struct, ActivityColors,
	// ActivityIcons, GlobalRegistry, and StatusStyle working together.
	// Sweeps: StatusEntry, WorkTypeEntry, ActivityEntry, Registry,
	// GlobalRegistry.
	t := DefaultTheme()

	// Register a custom status kind at activation time (plugin pattern).
	GlobalRegistry.RegisterStatus(StatusEntry{
		Kind:    "workarea-warming",
		Label:   "Warming pool",
		Symbol:  "↻",
		Color:   t.StatusInfo,
		Animate: true,
	})

	// Register a custom work-type kind.
	GlobalRegistry.RegisterWorkType(WorkTypeEntry{
		Kind:  "workarea-acquire",
		Label: "Acquire",
		Color: t.Teal,
	})

	// Register a custom activity kind.
	GlobalRegistry.RegisterActivity(ActivityEntry{
		Kind:  "tool-call",
		Icon:  "⚙",
		Color: t.Blue,
	})

	status := GetStatusStyle("working")
	titleStyle := Header()
	activityStyle := lipgloss.NewStyle().
		Foreground(t.TextPrimary).
		Background(ActivityColors["thought"])
	line := titleStyle.Render("session") +
		" " + status.Symbol + " " + status.Label +
		" " + activityStyle.Render(ActivityIcons["thought"]+" thinking")
	_ = line
}

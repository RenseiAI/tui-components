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

func Example() {
	// Intra-package composition: Theme struct, ActivityColors, and
	// ActivityIcons layered on top of a style constructor and a
	// StatusStyle lookup.
	t := DefaultTheme()
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

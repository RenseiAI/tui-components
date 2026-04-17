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
	// same entry; an unknown key falls back to TextSecondary. Compile-
	// only because image/color.Color does not format cleanly.
	_ = GetWorkTypeColor("Bugfix")
	_ = GetWorkTypeColor("bugfix")
	c := GetWorkTypeColor("unknown-type") // → TextSecondary
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

func Example() {
	// Intra-package composition that sweeps the remaining exported
	// surface: palette vars (TextPrimary), ActivityColors, and
	// ActivityIcons, layered on top of a style constructor and a
	// StatusStyle lookup.
	status := GetStatusStyle("working")
	titleStyle := Header()
	activityStyle := lipgloss.NewStyle().
		Foreground(TextPrimary).
		Background(ActivityColors["thought"])
	line := titleStyle.Render("session") +
		" " + status.Symbol + " " + status.Label +
		" " + activityStyle.Render(ActivityIcons["thought"]+" thinking")
	_ = line
}

package widget_test

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/widget"
)

// ExampleTabs demonstrates constructing a small [widget.Tabs] widget,
// fixing its size, and rendering the resulting tab bar. The example
// uses a narrow fixed width and ASCII-only titles so the expected
// output is stable across terminals.
func ExampleTabs() {
	tabs := widget.NewTabs(
		[]widget.TabsItem{
			{ID: "home", Title: "Home"},
			{ID: "work", Title: "Work"},
			{ID: "help", Title: "Help"},
		},
		widget.WithActive(1),
	)
	tabs.SetSize(24, 1)

	rendered := tabs.View().Content
	fmt.Println(lipgloss.Width(rendered))
	// Output: 24
}

package widget_test

import (
	"fmt"
	"strings"

	"github.com/RenseiAI/tui-components/widget"
)

// Example_logViewer demonstrates the minimal lifecycle of the
// LogViewer widget: construct, size, append, and render.
//
// The rendered frame contains ANSI SGR escape sequences whose exact
// byte layout is environment-dependent (lipgloss downsamples colours
// according to the active profile), so the example asserts only the
// row count, which is a stable function of the widget height.
func Example_logViewer() {
	lv := widget.New(widget.WithMaxLines(100))
	lv.SetSize(40, 6)
	lv.Append("hello", "world")

	view := lv.View()
	// Widget height is 6 rows (5 content + 1 footer), so the rendered
	// frame must contain exactly 6 rows — regardless of colour profile.
	rows := strings.Count(view.Content, "\n") + 1
	fmt.Printf("rows=%d\n", rows)

	// Output: rows=6
}

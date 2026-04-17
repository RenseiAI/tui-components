package notification_test

import (
	"fmt"

	"charm.land/lipgloss/v2"

	"github.com/RenseiAI/tui-components/widget/notification"
)

// ExampleModel constructs a single success toast at a fixed width and
// renders it. The example asserts the rendered width — borders included
// — matches the configured width, which is stable across terminals.
func ExampleModel() {
	n := notification.New(
		notification.VariantSuccess,
		"Saved successfully",
		notification.WithWidth(40),
	)
	rendered := n.View().Content
	fmt.Println(lipgloss.Width(rendered))
	// Output: 40
}

// ExampleStack pushes three toasts of different variants into a Stack
// at a fixed width and reports the live count along with the rendered
// width. Push returns the resulting stack and the auto-dismiss command
// for the new entry; the commands are intentionally discarded here so
// the example does not block on the default tick duration.
func ExampleStack() {
	s := notification.NewStack(notification.WithStackWidth(40))
	s, _ = s.Push(notification.New(notification.VariantSuccess, "Saved"))
	s, _ = s.Push(notification.New(notification.VariantWarning, "Heads up"))
	s, _ = s.Push(notification.New(notification.VariantError, "Failed"))

	fmt.Println("len:", s.Len())
	fmt.Println("width:", lipgloss.Width(s.View().Content))
	// Output:
	// len: 3
	// width: 40
}

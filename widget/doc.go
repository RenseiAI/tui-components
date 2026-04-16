// Package widget provides shared Bubble Tea UI components
// for the AgentFactory and Rensei TUI applications.
//
// # Widgets
//
//   - [Dialog] — a modal confirmation dialog with configurable buttons,
//     keyboard navigation, and overlay support.
//   - [Progressbar] — a themed progress bar wrapping
//     charm.land/bubbles/v2/progress with deterministic (percent-driven)
//     and indeterminate (animated sweep) modes, optional label, percent
//     text, and ETA estimator.
//
// # Progressbar example
//
// Construct a 60-cell-wide bar with a leading label, the right-aligned
// percentage, and an ETA estimator, then advance it as work completes:
//
//	bar := widget.NewProgressbar(
//	    widget.WithProgressbarWidth(60),
//	    widget.WithProgressbarLabel("Uploading"),
//	    widget.WithProgressbarShowPercent(true),
//	    widget.WithProgressbarShowETA(true),
//	)
//	bar.SetPercent(0.25)
//	// ... later ...
//	bar.IncrBy(0.25)
//	fmt.Println(bar.View().Content)
package widget

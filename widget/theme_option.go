// Package widget provides shared Bubble Tea UI components.
package widget

import "github.com/RenseiAI/tui-components/theme"

// ThemeOption is the universal theme-carrying option accepted by every
// widget constructor in this package.  Construct one with [WithTheme].
//
// Widget constructors that accept a variadic typed-option slice cannot
// directly accept a heterogeneous option type in idiomatic Go.  Instead,
// each widget exposes a thin helper — WithSpinnerTheme, WithProgressbarTheme,
// etc. — and the universal WithTheme helper constructs that per-widget
// option for you:
//
//	t := theme.DarkTheme()
//	sp  := widget.NewSpinner(widget.WithTheme(t).Spinner())
//	bar := widget.NewProgressbar(widget.WithTheme(t).Progressbar())
//
// Alternatively, use the per-widget helpers directly:
//
//	sp  := widget.NewSpinner(widget.WithSpinnerTheme(t))
//	bar := widget.NewProgressbar(widget.WithProgressbarTheme(t))
//
// Theme hot-swap: call SetTheme on any widget instance to update the theme
// mid-render; the next View call uses the new theme.
type ThemeOption struct {
	t theme.Theme
}

// WithTheme constructs a [ThemeOption] carrying t.  Use the receiver methods
// (.Spinner(), .Progressbar(), .Dialog(), .Tabs(), .Select(), .TextInput())
// to convert it to the appropriate widget-level option type.
func WithTheme(t theme.Theme) ThemeOption {
	return ThemeOption{t: t}
}

// Spinner returns a [SpinnerOption] that applies this theme to a Spinner.
func (o ThemeOption) Spinner() SpinnerOption {
	return WithSpinnerTheme(o.t)
}

// Progressbar returns a [ProgressbarOption] that applies this theme to a
// Progressbar.
func (o ThemeOption) Progressbar() ProgressbarOption {
	return WithProgressbarTheme(o.t)
}

// Dialog returns an [Option] (Dialog option) that applies this theme to a
// Dialog.
func (o ThemeOption) Dialog() Option {
	return WithDialogTheme(o.t)
}

// Tabs returns a [TabsOption] that applies this theme to a Tabs widget.
func (o ThemeOption) Tabs() TabsOption {
	return WithTabsTheme(o.t)
}

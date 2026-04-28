package widget_test

import (
	"testing"

	"github.com/RenseiAI/tui-components/theme"
	"github.com/RenseiAI/tui-components/widget"
)

// TestSpinnerWithTheme verifies that NewSpinner accepts WithSpinnerTheme and
// that SetTheme (mid-render hot-swap) does not panic.
func TestSpinnerWithTheme(t *testing.T) {
	themes := []struct {
		name string
		th   theme.Theme
	}{
		{"default", theme.DefaultTheme()},
		{"dark", theme.DarkTheme()},
		{"high-contrast", theme.HighContrastTheme()},
	}
	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			sp := widget.NewSpinner(widget.WithSpinnerTheme(tc.th))
			if sp == nil {
				t.Fatal("NewSpinner returned nil")
			}
			// View must not panic regardless of theme.
			_ = sp.View()
		})
	}
}

// TestSpinnerThemeHotSwap verifies that calling SetTheme mid-render updates
// the widget without panic or error.
func TestSpinnerThemeHotSwap(t *testing.T) {
	sp := widget.NewSpinner()
	v1 := sp.View()

	// Hot-swap to DarkTheme.
	sp.SetTheme(theme.DarkTheme())
	v2 := sp.View()

	// Both views must be non-empty (spinner renders at least one frame cell).
	if v1.Content == "" {
		t.Error("spinner view before swap is empty")
	}
	if v2.Content == "" {
		t.Error("spinner view after swap is empty")
	}
}

// TestProgressbarWithTheme verifies that NewProgressbar accepts
// WithProgressbarTheme and renders correctly for all built-in themes.
func TestProgressbarWithTheme(t *testing.T) {
	themes := []struct {
		name string
		th   theme.Theme
	}{
		{"default", theme.DefaultTheme()},
		{"dark", theme.DarkTheme()},
		{"high-contrast", theme.HighContrastTheme()},
	}
	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			bar := widget.NewProgressbar(
				widget.WithProgressbarTheme(tc.th),
				widget.WithProgressbarWidth(20),
			)
			_ = bar.SetPercent(0.5)
			view := bar.View()
			if view.Content == "" {
				t.Errorf("progressbar with %s theme rendered empty view", tc.name)
			}
		})
	}
}

// TestProgressbarThemeHotSwap verifies that SetTheme updates the bar's
// internal theme without panic.
func TestProgressbarThemeHotSwap(t *testing.T) {
	bar := widget.NewProgressbar(widget.WithProgressbarWidth(20))
	_ = bar.SetPercent(0.5)
	v1 := bar.View()

	bar.SetTheme(theme.HighContrastTheme())
	v2 := bar.View()

	if v1.Content == "" || v2.Content == "" {
		t.Error("progressbar view must be non-empty before and after theme swap")
	}
}

// TestDialogWithTheme verifies that New accepts WithDialogTheme and
// renders correctly for all built-in themes.
func TestDialogWithTheme(t *testing.T) {
	themes := []struct {
		name string
		th   theme.Theme
	}{
		{"default", theme.DefaultTheme()},
		{"dark", theme.DarkTheme()},
		{"high-contrast", theme.HighContrastTheme()},
	}
	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			d := widget.New(
				widget.WithDialogTheme(tc.th),
				widget.WithTitle("Title"),
				widget.WithBody("Body text"),
			)
			view := d.View()
			if view.Content == "" {
				t.Errorf("dialog with %s theme rendered empty view", tc.name)
			}
		})
	}
}

// TestTabsWithTheme verifies that NewTabs accepts WithTabsTheme and
// renders correctly for all built-in themes.
func TestTabsWithTheme(t *testing.T) {
	items := []widget.TabsItem{
		{ID: "a", Title: "Alpha"},
		{ID: "b", Title: "Beta"},
	}
	themes := []struct {
		name string
		th   theme.Theme
	}{
		{"default", theme.DefaultTheme()},
		{"dark", theme.DarkTheme()},
		{"high-contrast", theme.HighContrastTheme()},
	}
	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			tabs := widget.NewTabs(items, widget.WithTabsTheme(tc.th))
			view := tabs.View()
			if view.Content == "" {
				t.Errorf("tabs with %s theme rendered empty view", tc.name)
			}
		})
	}
}

// TestTabsThemeHotSwap verifies that SetTheme updates the Tabs widget's
// theme without panic.
func TestTabsThemeHotSwap(t *testing.T) {
	items := []widget.TabsItem{{ID: "a", Title: "Alpha"}, {ID: "b", Title: "Beta"}}
	tabs := widget.NewTabs(items)
	v1 := tabs.View()

	tabs.SetTheme(theme.DarkTheme())
	v2 := tabs.View()

	if v1.Content == "" || v2.Content == "" {
		t.Error("tabs view must be non-empty before and after theme swap")
	}
}

// TestWithThemeUniversalOption verifies that the universal WithTheme helper
// produces per-widget options that can be applied to each widget constructor.
func TestWithThemeUniversalOption(t *testing.T) {
	th := theme.DarkTheme()
	opt := widget.WithTheme(th)

	// Each converter must not panic and produce a widget that renders.
	t.Run("spinner", func(t *testing.T) {
		sp := widget.NewSpinner(opt.Spinner())
		if sp == nil {
			t.Fatal("NewSpinner returned nil")
		}
	})
	t.Run("progressbar", func(t *testing.T) {
		bar := widget.NewProgressbar(opt.Progressbar())
		if bar == nil {
			t.Fatal("NewProgressbar returned nil")
		}
	})
	t.Run("dialog", func(t *testing.T) {
		d := widget.New(opt.Dialog())
		if d == nil {
			t.Fatal("New(dialog) returned nil")
		}
	})
	t.Run("tabs", func(t *testing.T) {
		tabs := widget.NewTabs([]widget.TabsItem{{ID: "a", Title: "A"}}, opt.Tabs())
		if tabs == nil {
			t.Fatal("NewTabs returned nil")
		}
	})
}

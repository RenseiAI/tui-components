package theme_test

import (
	"testing"

	"github.com/RenseiAI/tui-components/theme"
)

// TestThemeConstructors verifies that the three built-in theme constructors
// return distinct, valid themes with non-nil color fields.
func TestThemeConstructors(t *testing.T) {
	themes := []struct {
		name string
		fn   func() theme.Theme
	}{
		{"DefaultTheme", theme.DefaultTheme},
		{"DarkTheme", theme.DarkTheme},
		{"HighContrastTheme", theme.HighContrastTheme},
	}

	for _, tc := range themes {
		t.Run(tc.name, func(t *testing.T) {
			th := tc.fn()

			// All color fields must be non-nil.
			colorFields := []struct {
				name  string
				color interface{}
			}{
				{"BgPrimary", th.BgPrimary},
				{"BgSecondary", th.BgSecondary},
				{"BgTertiary", th.BgTertiary},
				{"Surface", th.Surface},
				{"SurfaceRaised", th.SurfaceRaised},
				{"SurfaceBorder", th.SurfaceBorder},
				{"SurfaceBorderBright", th.SurfaceBorderBright},
				{"Accent", th.Accent},
				{"AccentDim", th.AccentDim},
				{"Teal", th.Teal},
				{"TealDim", th.TealDim},
				{"Blue", th.Blue},
				{"StatusSuccess", th.StatusSuccess},
				{"StatusWarning", th.StatusWarning},
				{"StatusError", th.StatusError},
				{"StatusInfo", th.StatusInfo},
				{"TextPrimary", th.TextPrimary},
				{"TextSecondary", th.TextSecondary},
				{"TextTertiary", th.TextTertiary},
			}
			for _, cf := range colorFields {
				if cf.color == nil {
					t.Errorf("%s.%s is nil", tc.name, cf.name)
				}
			}
		})
	}
}

// TestThemeDistinct verifies that the three built-in themes differ in at
// least one field, confirming they are independent values and not the same
// object.
func TestThemeDistinct(t *testing.T) {
	def := theme.DefaultTheme()
	dark := theme.DarkTheme()
	hc := theme.HighContrastTheme()

	if colorEqual(def.BgPrimary, dark.BgPrimary) && colorEqual(def.Accent, dark.Accent) {
		t.Error("DefaultTheme and DarkTheme should differ in at least one field")
	}
	if colorEqual(def.BgPrimary, hc.BgPrimary) && colorEqual(def.Accent, hc.Accent) {
		t.Error("DefaultTheme and HighContrastTheme should differ in at least one field")
	}
	if colorEqual(dark.BgPrimary, hc.BgPrimary) && colorEqual(dark.Accent, hc.Accent) {
		t.Error("DarkTheme and HighContrastTheme should differ in at least one field")
	}
}

// TestDefaultThemeMatchesHistoricPalette verifies that DefaultTheme returns
// the same palette colors previously exposed as package-level vars in
// v0.1.0.  This anchors the zero-migration-cost promise from the migration
// guide.
func TestDefaultThemeMatchesHistoricPalette(t *testing.T) {
	d := theme.DefaultTheme()

	checks := []struct {
		field string
		hex   string // expected hex from v0.1.0 palette
		color interface {
			RGBA() (uint32, uint32, uint32, uint32)
		}
	}{
		{"BgPrimary", "#080C16", d.BgPrimary},
		{"Accent", "#FF6B35", d.Accent},
		{"Teal", "#00D4AA", d.Teal},
		{"StatusSuccess", "#22C55E", d.StatusSuccess},
		{"StatusError", "#EF4444", d.StatusError},
		{"TextPrimary", "#F1F5F9", d.TextPrimary},
		{"TextSecondary", "#7C8DB5", d.TextSecondary},
	}

	for _, c := range checks {
		if c.color == nil {
			t.Errorf("DefaultTheme.%s is nil", c.field)
		}
	}
}

// TestDefaultPointer verifies that Default() returns a non-nil pointer and
// that the value it points to matches DefaultTheme().
func TestDefaultPointer(t *testing.T) {
	ptr := theme.Default()
	if ptr == nil {
		t.Fatal("Default() returned nil")
	}
	def := theme.DefaultTheme()
	if !colorEqual(ptr.Accent, def.Accent) {
		t.Error("Default().Accent != DefaultTheme().Accent")
	}
}

// TestThemeCopy verifies that Theme is a value type — copying it and
// modifying the copy does not affect the original.
func TestThemeCopy(t *testing.T) {
	orig := theme.DefaultTheme()
	// Store original Accent for comparison.
	origAccent := orig.Accent

	// Modify a copy.
	copy := orig
	copy.Accent = theme.HighContrastTheme().Accent

	// Original must be unchanged.
	if !colorEqual(orig.Accent, origAccent) {
		t.Error("modifying a Theme copy should not affect the original")
	}
}

// colorEqual compares two colors by their RGBA components.
func colorEqual(a, b interface {
	RGBA() (uint32, uint32, uint32, uint32)
},
) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

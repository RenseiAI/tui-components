package theme

import (
	"image/color"
	"testing"
)

func TestGetStatusStyle(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   StatusStyle
	}{
		{"working", "working", StatusStyle{Label: "Working", Color: pkg.StatusSuccess, Symbol: "●", Animate: true}},
		{"queued", "queued", StatusStyle{Label: "Queued", Color: pkg.StatusWarning, Symbol: "◌", Animate: true}},
		{"parked", "parked", StatusStyle{Label: "Parked", Color: pkg.TextTertiary, Symbol: "○", Animate: false}},
		{"completed", "completed", StatusStyle{Label: "Done", Color: pkg.StatusSuccess, Symbol: "✓", Animate: false}},
		{"failed", "failed", StatusStyle{Label: "Failed", Color: pkg.StatusError, Symbol: "✗", Animate: false}},
		{"stopped", "stopped", StatusStyle{Label: "Stopped", Color: pkg.TextTertiary, Symbol: "■", Animate: false}},
		{"unknown", "not-a-real-status", StatusStyle{Label: "Unknown", Color: pkg.TextSecondary, Symbol: "?", Animate: false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStatusStyle(tt.status)
			if got.Label != tt.want.Label {
				t.Errorf("Label = %q, want %q", got.Label, tt.want.Label)
			}
			if got.Symbol != tt.want.Symbol {
				t.Errorf("Symbol = %q, want %q", got.Symbol, tt.want.Symbol)
			}
			if got.Animate != tt.want.Animate {
				t.Errorf("Animate = %v, want %v", got.Animate, tt.want.Animate)
			}
			if !sameColor(got.Color, tt.want.Color) {
				t.Errorf("Color = %v, want %v", got.Color, tt.want.Color)
			}
		})
	}
}

func sameColor(a, b color.Color) bool {
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

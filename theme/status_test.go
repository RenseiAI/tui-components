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
		{"working", "working", StatusStyle{Label: "Working", Color: StatusSuccess, Symbol: "\u25cf", Animate: true}},
		{"queued", "queued", StatusStyle{Label: "Queued", Color: StatusWarning, Symbol: "\u25cc", Animate: true}},
		{"parked", "parked", StatusStyle{Label: "Parked", Color: TextTertiary, Symbol: "\u25cb", Animate: false}},
		{"completed", "completed", StatusStyle{Label: "Done", Color: StatusSuccess, Symbol: "\u2713", Animate: false}},
		{"failed", "failed", StatusStyle{Label: "Failed", Color: StatusError, Symbol: "\u2717", Animate: false}},
		{"stopped", "stopped", StatusStyle{Label: "Stopped", Color: TextTertiary, Symbol: "\u25a0", Animate: false}},
		{"unknown", "not-a-real-status", StatusStyle{Label: "Unknown", Color: TextSecondary, Symbol: "?", Animate: false}},
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

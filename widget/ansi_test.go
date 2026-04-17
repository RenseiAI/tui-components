package widget

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// collectText concatenates the text of all runs, so tests can verify the
// printable payload was preserved in order.
func collectText(runs []styledRun) string {
	var b strings.Builder
	for _, r := range runs {
		b.WriteString(r.Text)
	}
	return b.String()
}

// findRun returns the first run whose text exactly matches want. ok=false
// if none matches; the test can then dump runs for debugging.
func findRun(runs []styledRun, want string) (styledRun, bool) {
	for _, r := range runs {
		if r.Text == want {
			return r, true
		}
	}
	return styledRun{}, false
}

// colorEqual compares two color.Color values by their RGBA byte triplets.
// Direct equality fails for ANSIColor vs RGBA even when both resolve to
// the same hex.
func colorEqual(a, b color.Color) bool {
	if a == nil || b == nil {
		return a == b
	}
	ar, ag, ab, _ := a.RGBA()
	br, bg, bb, _ := b.RGBA()
	return ar == br && ag == bg && ab == bb
}

// isDefaultColor reports whether c represents "no color applied" — i.e.
// nil or lipgloss.NoColor{}.
func isDefaultColor(c color.Color) bool {
	if c == nil {
		return true
	}
	_, ok := c.(lipgloss.NoColor)
	return ok
}

func TestSGRParserPlainText(t *testing.T) {
	p := newSGRParser()
	runs := p.Parse("hello world")
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d: %#v", len(runs), runs)
	}
	if runs[0].Text != "hello world" {
		t.Errorf("text mismatch: got %q", runs[0].Text)
	}
	if _, isNo := runs[0].Style.GetForeground().(lipgloss.NoColor); !isNo {
		t.Errorf("expected no foreground, got %v", runs[0].Style.GetForeground())
	}
	if _, isNo := runs[0].Style.GetBackground().(lipgloss.NoColor); !isNo {
		t.Errorf("expected no background, got %v", runs[0].Style.GetBackground())
	}
}

func TestSGRParserFamilies(t *testing.T) {
	tests := []struct {
		name  string
		input string
		text  string
		check func(t *testing.T, s lipgloss.Style)
	}{
		{
			name:  "bold",
			input: "\x1b[1mB\x1b[0m",
			text:  "B",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetBold() {
					t.Error("expected bold")
				}
			},
		},
		{
			name:  "dim",
			input: "\x1b[2mD\x1b[0m",
			text:  "D",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetFaint() {
					t.Error("expected faint/dim")
				}
			},
		},
		{
			name:  "italic",
			input: "\x1b[3mI\x1b[0m",
			text:  "I",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetItalic() {
					t.Error("expected italic")
				}
			},
		},
		{
			name:  "underline",
			input: "\x1b[4mU\x1b[0m",
			text:  "U",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetUnderline() {
					t.Error("expected underline")
				}
			},
		},
		{
			name:  "reverse",
			input: "\x1b[7mR\x1b[0m",
			text:  "R",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetReverse() {
					t.Error("expected reverse")
				}
			},
		},
		{
			name:  "strikethrough",
			input: "\x1b[9mS\x1b[0m",
			text:  "S",
			check: func(t *testing.T, s lipgloss.Style) {
				if !s.GetStrikethrough() {
					t.Error("expected strikethrough")
				}
			},
		},
		{
			name:  "16-color fg",
			input: "\x1b[31mred\x1b[0m",
			text:  "red",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(1)
				if !colorEqual(s.GetForeground(), want) {
					t.Errorf("expected fg=ANSIColor(1), got %v", s.GetForeground())
				}
			},
		},
		{
			name:  "16-color bright fg",
			input: "\x1b[91mbright\x1b[0m",
			text:  "bright",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(9)
				if !colorEqual(s.GetForeground(), want) {
					t.Errorf("expected fg=ANSIColor(9), got %v", s.GetForeground())
				}
			},
		},
		{
			name:  "16-color bg",
			input: "\x1b[42mgreenbg\x1b[0m",
			text:  "greenbg",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(2)
				if !colorEqual(s.GetBackground(), want) {
					t.Errorf("expected bg=ANSIColor(2), got %v", s.GetBackground())
				}
			},
		},
		{
			name:  "16-color bright bg",
			input: "\x1b[105mbrightbg\x1b[0m",
			text:  "brightbg",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(13)
				if !colorEqual(s.GetBackground(), want) {
					t.Errorf("expected bg=ANSIColor(13), got %v", s.GetBackground())
				}
			},
		},
		{
			name:  "256-color fg",
			input: "\x1b[38;5;208msalmon\x1b[0m",
			text:  "salmon",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(208)
				if !colorEqual(s.GetForeground(), want) {
					t.Errorf("expected fg=ANSIColor(208), got %v", s.GetForeground())
				}
			},
		},
		{
			name:  "256-color bg",
			input: "\x1b[48;5;17mnavybg\x1b[0m",
			text:  "navybg",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.ANSIColor(17)
				if !colorEqual(s.GetBackground(), want) {
					t.Errorf("expected bg=ANSIColor(17), got %v", s.GetBackground())
				}
			},
		},
		{
			name:  "truecolor fg",
			input: "\x1b[38;2;255;107;53mbrand\x1b[0m",
			text:  "brand",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.Color("#ff6b35")
				if !colorEqual(s.GetForeground(), want) {
					t.Errorf("expected fg=#ff6b35, got %v", s.GetForeground())
				}
			},
		},
		{
			name:  "truecolor bg",
			input: "\x1b[48;2;0;212;170mtealbg\x1b[0m",
			text:  "tealbg",
			check: func(t *testing.T, s lipgloss.Style) {
				want := lipgloss.Color("#00d4aa")
				if !colorEqual(s.GetBackground(), want) {
					t.Errorf("expected bg=#00d4aa, got %v", s.GetBackground())
				}
			},
		},
		{
			name:  "default fg",
			input: "\x1b[31mred\x1b[39mdefault",
			text:  "default",
			check: func(t *testing.T, s lipgloss.Style) {
				if !isDefaultColor(s.GetForeground()) {
					t.Errorf("expected default/no fg, got %v", s.GetForeground())
				}
			},
		},
		{
			name:  "default bg",
			input: "\x1b[41mon\x1b[49moff",
			text:  "off",
			check: func(t *testing.T, s lipgloss.Style) {
				if !isDefaultColor(s.GetBackground()) {
					t.Errorf("expected default/no bg, got %v", s.GetBackground())
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newSGRParser()
			runs := p.Parse(tc.input)
			run, ok := findRun(runs, tc.text)
			if !ok {
				t.Fatalf("run %q not found in %#v", tc.text, runs)
			}
			tc.check(t, run.Style)
		})
	}
}

func TestSGRParserPartialSequenceCarryover(t *testing.T) {
	p := newSGRParser()

	first := p.Parse("\x1b[31")
	// The partial sequence must not emit any text.
	if got := collectText(first); got != "" {
		t.Errorf("first chunk should emit no text, got %q", got)
	}

	second := p.Parse("mhello\x1b[0m world")

	// "hello" should be red, " world" should be unstyled.
	hello, ok := findRun(second, "hello")
	if !ok {
		t.Fatalf("missing 'hello' run in %#v", second)
	}
	if !colorEqual(hello.Style.GetForeground(), lipgloss.ANSIColor(1)) {
		t.Errorf("expected 'hello' red, got fg=%v", hello.Style.GetForeground())
	}

	world, ok := findRun(second, " world")
	if !ok {
		t.Fatalf("missing ' world' run in %#v", second)
	}
	if !isDefaultColor(world.Style.GetForeground()) {
		t.Errorf("expected ' world' unstyled, got fg=%v", world.Style.GetForeground())
	}
}

func TestSGRParserResetCodeClearsState(t *testing.T) {
	p := newSGRParser()
	runs := p.Parse("\x1b[1;31mbold\x1b[0mplain")

	plain, ok := findRun(runs, "plain")
	if !ok {
		t.Fatalf("missing 'plain' run in %#v", runs)
	}
	if plain.Style.GetBold() {
		t.Error("expected 'plain' not bold after reset")
	}
	if !isDefaultColor(plain.Style.GetForeground()) {
		t.Errorf("expected 'plain' no fg, got %v", plain.Style.GetForeground())
	}

	// Empty SGR (\x1b[m) also resets.
	p2 := newSGRParser()
	runs2 := p2.Parse("\x1b[1mbold\x1b[mnormal")
	normal, ok := findRun(runs2, "normal")
	if !ok {
		t.Fatalf("missing 'normal' run in %#v", runs2)
	}
	if normal.Style.GetBold() {
		t.Error("expected 'normal' not bold after empty SGR reset")
	}
}

func TestSGRParserResetMethod(t *testing.T) {
	p := newSGRParser()
	_ = p.Parse("\x1b[31m") // set red, no text emitted yet
	p.Reset()
	runs := p.Parse("plain")
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if !isDefaultColor(runs[0].Style.GetForeground()) {
		t.Errorf("expected no fg after Reset, got %v", runs[0].Style.GetForeground())
	}
}

func TestSGRParserNonSGRCSIDropped(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"erase screen", "before\x1b[2Jafter", "beforeafter"},
		{"cursor position", "a\x1b[10;5Hb", "ab"},
		{"private mode set", "x\x1b[?25hy", "xy"},
		{"clear to EOL", "line\x1b[Ktrail", "linetrail"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := newSGRParser()
			runs := p.Parse(tc.input)
			if got := collectText(runs); got != tc.want {
				t.Errorf("got %q, want %q (runs=%#v)", got, tc.want, runs)
			}
		})
	}
}

func TestSGRParserNonSGRCSIPreservesActiveStyle(t *testing.T) {
	// A non-SGR CSI in the middle of a styled span must not reset the
	// active SGR state. Because the non-SGR sequence does not force a
	// run boundary, the surrounding red text may be emitted as a single
	// merged run — that's fine; what matters is the style.
	p := newSGRParser()
	runs := p.Parse("\x1b[31mred\x1b[2Jstill")
	got := collectText(runs)
	if got != "redstill" {
		t.Errorf("expected text 'redstill', got %q", got)
	}
	if len(runs) == 0 {
		t.Fatal("no runs emitted")
	}
	for _, r := range runs {
		if !colorEqual(r.Style.GetForeground(), lipgloss.ANSIColor(1)) {
			t.Errorf("run %q: expected red fg, got %v", r.Text, r.Style.GetForeground())
		}
	}
}

func TestSGRParserMalformedNoPanic(t *testing.T) {
	// Behaviour: a malformed sequence with no terminator at end-of-input
	// is retained as internal parser state and does NOT emit text or
	// crash. A subsequent Parse call can either terminate it or remain
	// stuck — either is acceptable; we just guarantee no panic.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("parser panicked: %v", r)
		}
	}()

	p := newSGRParser()
	runs := p.Parse("visible\x1b[999;")
	if got := collectText(runs); got != "visible" {
		t.Errorf("expected surrounding text preserved, got %q", got)
	}

	// Feed the terminator on a subsequent call — parser should recover.
	runs2 := p.Parse("mafter\x1b[0mtail")
	// "after" may or may not be styled depending on how the decoder
	// interpreted 999 — we only assert no panic and that "tail" exists.
	tail, ok := findRun(runs2, "tail")
	if !ok {
		// Depending on decoder state, "after" and "tail" could merge.
		// The important thing is that post-reset text is unstyled.
		if got := collectText(runs2); !strings.Contains(got, "tail") {
			t.Fatalf("expected 'tail' in output, got %q", got)
		}
		return
	}
	if !isDefaultColor(tail.Style.GetForeground()) {
		t.Errorf("expected 'tail' unstyled, got fg=%v", tail.Style.GetForeground())
	}
}

func TestSGRParserEmptyInput(t *testing.T) {
	p := newSGRParser()
	runs := p.Parse("")
	if runs != nil {
		t.Errorf("expected nil runs for empty input, got %#v", runs)
	}
}

func TestSGRParserMultipleAttributes(t *testing.T) {
	// Combined attributes in one sequence.
	p := newSGRParser()
	runs := p.Parse("\x1b[1;3;31mstyled\x1b[0m")
	styled, ok := findRun(runs, "styled")
	if !ok {
		t.Fatalf("missing 'styled' run in %#v", runs)
	}
	if !styled.Style.GetBold() {
		t.Error("expected bold")
	}
	if !styled.Style.GetItalic() {
		t.Error("expected italic")
	}
	if !colorEqual(styled.Style.GetForeground(), lipgloss.ANSIColor(1)) {
		t.Errorf("expected red fg, got %v", styled.Style.GetForeground())
	}
}

func TestSGRParserAttributeOff(t *testing.T) {
	// Partial resets: 22 turns off bold/dim, 23 italic, 24 underline,
	// 27 reverse, 29 strikethrough — without touching colours.
	p := newSGRParser()
	runs := p.Parse("\x1b[1;4;31mon\x1b[22;24moff\x1b[0m")

	off, ok := findRun(runs, "off")
	if !ok {
		t.Fatalf("missing 'off' run in %#v", runs)
	}
	if off.Style.GetBold() {
		t.Error("expected bold cleared by 22")
	}
	if off.Style.GetUnderline() {
		t.Error("expected underline cleared by 24")
	}
	// Colour should persist.
	if !colorEqual(off.Style.GetForeground(), lipgloss.ANSIColor(1)) {
		t.Errorf("expected red preserved, got %v", off.Style.GetForeground())
	}
}

func TestSGRParserMalformedExtendedColor(t *testing.T) {
	// 38 without the 2/5 subtype parameter — should not change colour.
	p := newSGRParser()
	runs := p.Parse("\x1b[38mtext\x1b[0m")
	text, ok := findRun(runs, "text")
	if !ok {
		t.Fatalf("missing 'text' run in %#v", runs)
	}
	if !isDefaultColor(text.Style.GetForeground()) {
		t.Errorf("expected no fg for malformed 38, got %v", text.Style.GetForeground())
	}

	// 38;5 with missing index — also dropped.
	p2 := newSGRParser()
	runs2 := p2.Parse("\x1b[38;5m\x1b[0mplain")
	plain, ok := findRun(runs2, "plain")
	if !ok {
		t.Fatalf("missing 'plain' in %#v", runs2)
	}
	if !isDefaultColor(plain.Style.GetForeground()) {
		t.Errorf("expected no fg, got %v", plain.Style.GetForeground())
	}
}

func TestSGRParserReverseAndStrikethroughOff(t *testing.T) {
	p := newSGRParser()
	runs := p.Parse("\x1b[7;9mon\x1b[27;29moff\x1b[0m")
	off, ok := findRun(runs, "off")
	if !ok {
		t.Fatalf("missing 'off' run in %#v", runs)
	}
	if off.Style.GetReverse() {
		t.Error("expected reverse cleared by 27")
	}
	if off.Style.GetStrikethrough() {
		t.Error("expected strikethrough cleared by 29")
	}
}

func TestSGRParserItalicOff(t *testing.T) {
	p := newSGRParser()
	runs := p.Parse("\x1b[3mon\x1b[23moff\x1b[0m")
	off, ok := findRun(runs, "off")
	if !ok {
		t.Fatalf("missing 'off' run in %#v", runs)
	}
	if off.Style.GetItalic() {
		t.Error("expected italic cleared by 23")
	}
}

func TestSGRParserTruecolorClampsOutOfRange(t *testing.T) {
	// Out-of-range RGB components clamp to 0-255 without panicking.
	p := newSGRParser()
	runs := p.Parse("\x1b[38;2;999;0;300mtext\x1b[0m")
	text, ok := findRun(runs, "text")
	if !ok {
		t.Fatalf("missing 'text' run in %#v", runs)
	}
	// 999 clamps to 255, 300 clamps to 255 → #ff00ff.
	want := lipgloss.Color("#ff00ff")
	if !colorEqual(text.Style.GetForeground(), want) {
		t.Errorf("expected clamped #ff00ff, got %v", text.Style.GetForeground())
	}
}

func TestSGRParser256ColorOutOfRange(t *testing.T) {
	// 256-colour index > 255 is dropped (no colour applied).
	p := newSGRParser()
	runs := p.Parse("\x1b[38;5;300mtext\x1b[0m")
	text, ok := findRun(runs, "text")
	if !ok {
		t.Fatalf("missing 'text' run in %#v", runs)
	}
	if !isDefaultColor(text.Style.GetForeground()) {
		t.Errorf("expected no fg for out-of-range 256, got %v", text.Style.GetForeground())
	}
}

func TestSGRParserOSCSequenceDropped(t *testing.T) {
	// Operating System Commands (hyperlinks, titles) are stripped.
	p := newSGRParser()
	runs := p.Parse("pre\x1b]0;window title\x07post")
	got := collectText(runs)
	if got != "prepost" {
		t.Errorf("expected OSC stripped, got %q", got)
	}
}

func TestSGRParserPreservesNewlines(t *testing.T) {
	// Log content with newlines must survive — we're not re-wrapping.
	p := newSGRParser()
	runs := p.Parse("line1\nline2\n\x1b[31mred\x1b[0m\n")
	got := collectText(runs)
	want := "line1\nline2\nred\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

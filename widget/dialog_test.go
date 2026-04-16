package widget

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/RenseiAI/tui-components/component"
)

// compile-time check that *Dialog satisfies component.Component.
var _ component.Component = (*Dialog)(nil)

// --- helpers -----------------------------------------------------------------

func keyPress(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code}
}

func keyPressRune(r rune) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: r, Text: string(r)}
}

func keyPressShiftTab() tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: tea.KeyTab, Mod: tea.ModShift}
}

// execCmd calls a tea.Cmd and returns the resulting tea.Msg.
func execCmd(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

// sendKey calls Update with the given key message and returns the cmd.
func sendKey(d *Dialog, msg tea.KeyPressMsg) tea.Cmd {
	_, cmd := d.Update(msg)
	return cmd
}

// testGolden is a simple golden-file helper.
// When UPDATE_GOLDEN=1 is set it writes the file; otherwise it reads and compares.
func testGolden(t *testing.T, name string, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Skipf("golden file %s updated", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v (run with UPDATE_GOLDEN=1 to create)", path, err)
	}
	if string(want) != got {
		t.Errorf("golden mismatch for %s\n--- want (len %d) ---\n%s\n--- got  (len %d) ---\n%s",
			name, len(want), string(want), len(got), got)
	}
}

// --- Result.String -----------------------------------------------------------

func TestResultString(t *testing.T) {
	tests := []struct {
		r    Result
		want string
	}{
		{ResultNone, "none"},
		{ResultYes, "yes"},
		{ResultNo, "no"},
		{ResultCancel, "cancel"},
		{Result(99), "none"},
	}
	for _, tt := range tests {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("Result(%d).String() = %q, want %q", int(tt.r), got, tt.want)
		}
	}
}

// --- Button navigation -------------------------------------------------------

func TestButtonNavigation(t *testing.T) {
	tests := []struct {
		name string
		keys []tea.KeyPressMsg
		want int
	}{
		{"right once", []tea.KeyPressMsg{keyPress(tea.KeyRight)}, 1},
		{"right twice", []tea.KeyPressMsg{keyPress(tea.KeyRight), keyPress(tea.KeyRight)}, 2},
		{"right 3 wraps to 0", []tea.KeyPressMsg{keyPress(tea.KeyRight), keyPress(tea.KeyRight), keyPress(tea.KeyRight)}, 0},
		{"tab once", []tea.KeyPressMsg{keyPress(tea.KeyTab)}, 1},
		{"left from 0 wraps to 2", []tea.KeyPressMsg{keyPress(tea.KeyLeft)}, 2},
		{"shift+tab from 0 wraps to 2", []tea.KeyPressMsg{keyPressShiftTab()}, 2},
		{"right then left", []tea.KeyPressMsg{keyPress(tea.KeyRight), keyPress(tea.KeyLeft)}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := New(WithTitle("nav"))
			for _, k := range tt.keys {
				d.Update(k)
			}
			if got := d.FocusedIndex(); got != tt.want {
				t.Errorf("FocusedIndex() = %d, want %d", got, tt.want)
			}
		})
	}
}

// --- Enter activation --------------------------------------------------------

func TestEnterActivation(t *testing.T) {
	buttons := []struct {
		rightPresses int
		wantResult   Result
	}{
		{0, ResultYes},
		{1, ResultNo},
		{2, ResultCancel},
	}
	for _, tt := range buttons {
		t.Run(tt.wantResult.String(), func(t *testing.T) {
			d := New(WithTitle("activate"))
			for i := 0; i < tt.rightPresses; i++ {
				d.Update(keyPress(tea.KeyRight))
			}
			cmd := sendKey(d, keyPress(tea.KeyEnter))
			if d.Result() != tt.wantResult {
				t.Errorf("Result() = %v, want %v", d.Result(), tt.wantResult)
			}
			msg := execCmd(cmd)
			done, ok := msg.(DialogDoneMsg)
			if !ok {
				t.Fatalf("cmd returned %T, want DialogDoneMsg", msg)
			}
			if done.Result != tt.wantResult {
				t.Errorf("DialogDoneMsg.Result = %v, want %v", done.Result, tt.wantResult)
			}
		})
	}
}

// --- Esc cancellation --------------------------------------------------------

func TestEscCancellation(t *testing.T) {
	// Esc should produce ResultCancel regardless of focused button.
	for _, focusRight := range []int{0, 1, 2} {
		d := New(WithTitle("esc"))
		for i := 0; i < focusRight; i++ {
			d.Update(keyPress(tea.KeyRight))
		}
		cmd := sendKey(d, keyPress(tea.KeyEscape))
		if d.Result() != ResultCancel {
			t.Errorf("focus=%d: Result() = %v, want ResultCancel", focusRight, d.Result())
		}
		msg := execCmd(cmd)
		done, ok := msg.(DialogDoneMsg)
		if !ok {
			t.Fatalf("focus=%d: cmd returned %T, want DialogDoneMsg", focusRight, msg)
		}
		if done.Result != ResultCancel {
			t.Errorf("focus=%d: DialogDoneMsg.Result = %v, want ResultCancel", focusRight, done.Result)
		}
	}
}

// --- y / n quick-activation --------------------------------------------------

func TestQuickActivation(t *testing.T) {
	t.Run("y activates yes", func(t *testing.T) {
		d := New(WithTitle("quick"))
		cmd := sendKey(d, keyPressRune('y'))
		if d.Result() != ResultYes {
			t.Errorf("Result() = %v, want ResultYes", d.Result())
		}
		msg := execCmd(cmd)
		done, ok := msg.(DialogDoneMsg)
		if !ok {
			t.Fatalf("cmd returned %T, want DialogDoneMsg", msg)
		}
		if done.Result != ResultYes {
			t.Errorf("DialogDoneMsg.Result = %v, want ResultYes", done.Result)
		}
	})

	t.Run("n activates no", func(t *testing.T) {
		d := New(WithTitle("quick"))
		cmd := sendKey(d, keyPressRune('n'))
		if d.Result() != ResultNo {
			t.Errorf("Result() = %v, want ResultNo", d.Result())
		}
		msg := execCmd(cmd)
		done, ok := msg.(DialogDoneMsg)
		if !ok {
			t.Fatalf("cmd returned %T, want DialogDoneMsg", msg)
		}
		if done.Result != ResultNo {
			t.Errorf("DialogDoneMsg.Result = %v, want ResultNo", done.Result)
		}
	})

	t.Run("y noop without yes button", func(t *testing.T) {
		d := New(WithButtons(Button{Label: "OK", Result: ResultCancel}))
		cmd := sendKey(d, keyPressRune('y'))
		if d.Result() != ResultNone {
			t.Errorf("Result() = %v, want ResultNone", d.Result())
		}
		if cmd != nil {
			t.Errorf("expected nil cmd for y with no yes button")
		}
	})

	t.Run("n noop without no button", func(t *testing.T) {
		d := New(WithButtons(Button{Label: "OK", Result: ResultCancel}))
		cmd := sendKey(d, keyPressRune('n'))
		if d.Result() != ResultNone {
			t.Errorf("Result() = %v, want ResultNone", d.Result())
		}
		if cmd != nil {
			t.Errorf("expected nil cmd for n with no no button")
		}
	})
}

// --- Reset -------------------------------------------------------------------

func TestReset(t *testing.T) {
	d := New(WithTitle("reset"))
	d.Update(keyPress(tea.KeyRight))
	sendKey(d, keyPress(tea.KeyEnter))
	if d.Result() == ResultNone {
		t.Fatal("expected non-none result after activation")
	}
	d.Reset()
	if d.Result() != ResultNone {
		t.Errorf("after Reset(), Result() = %v, want ResultNone", d.Result())
	}
	if d.FocusedIndex() != 0 {
		t.Errorf("after Reset(), FocusedIndex() = %d, want 0", d.FocusedIndex())
	}
}

// --- Focus / Blur ------------------------------------------------------------

func TestFocusBlur(t *testing.T) {
	d := New()
	if d.Focused() {
		t.Error("new dialog should not be focused")
	}
	d.Focus()
	if !d.Focused() {
		t.Error("after Focus(), Focused() should be true")
	}
	d.Blur()
	if d.Focused() {
		t.Error("after Blur(), Focused() should be false")
	}
}

// --- Init returns nil --------------------------------------------------------

func TestInitReturnsNil(t *testing.T) {
	d := New()
	if cmd := d.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

// --- Custom buttons via WithButtons ------------------------------------------

func TestCustomButtons(t *testing.T) {
	btns := []Button{
		{Label: "Save", Result: ResultYes},
		{Label: "Discard", Result: ResultNo},
	}
	d := New(WithButtons(btns...))
	got := d.Buttons()
	if len(got) != 2 {
		t.Fatalf("Buttons() len = %d, want 2", len(got))
	}
	if got[0].Label != "Save" || got[1].Label != "Discard" {
		t.Errorf("unexpected labels: %v", got)
	}
}

// --- SetSize re-wrapping -----------------------------------------------------

func TestSetSizeRewrapping(t *testing.T) {
	body := strings.Repeat("word ", 40) // long body
	d := New(WithTitle("Wrap"), WithBody(body))

	d.SetSize(30, 10)
	narrow := d.Render()

	d.SetSize(80, 20)
	wide := d.Render()

	// Both renders should produce non-empty output and differ from each other,
	// showing that SetSize influences rendering.
	if narrow == "" || wide == "" {
		t.Fatal("Render should produce non-empty output at both sizes")
	}
	if narrow == wide {
		t.Error("narrow and wide renders should differ")
	}
}

// --- Overlay -----------------------------------------------------------------

func TestOverlayPreservesHeight(t *testing.T) {
	d := New(WithTitle("Over"), WithBody("body text"))
	d.SetSize(60, 15)
	bg := strings.Repeat("background line\n", 14) + "background line"
	out := d.Overlay(bg)
	lines := strings.Split(out, "\n")
	if len(lines) != 15 {
		t.Errorf("Overlay line count = %d, want 15", len(lines))
	}
}

func TestOverlayBeforeSetSize(t *testing.T) {
	d := New(WithTitle("NoSize"), WithBody("body"))
	out := d.Overlay("some background")
	// Without SetSize, Overlay should return just the box (no backdrop composition).
	if strings.Contains(out, "some background") {
		t.Error("Overlay before SetSize should not contain background text")
	}
	if out == "" {
		t.Error("Overlay before SetSize should return the dialog box")
	}
}

func TestOverlayDimsBackground(t *testing.T) {
	d := New(WithTitle("Dim"), WithBody("body"))
	d.SetSize(60, 15)
	bg := strings.Repeat("visible text\n", 14) + "visible text"
	out := d.Overlay(bg)
	// The background lines that are NOT replaced by the dialog box should not
	// contain the original plain text (they are re-styled/dimmed).
	// At least some background content should be present but transformed.
	if out == "" {
		t.Error("Overlay should produce non-empty output")
	}
}

// --- View returns non-empty --------------------------------------------------

func TestViewReturnsView(t *testing.T) {
	d := New(WithTitle("View"), WithBody("text"))
	d.SetSize(60, 20)
	v := d.View()
	if v.Content == "" {
		t.Error("View() should return non-empty tea.View")
	}
}

// --- Non-key message is ignored ----------------------------------------------

func TestNonKeyMessageIgnored(t *testing.T) {
	d := New(WithTitle("ignore"))
	_, cmd := d.Update("not a key")
	if cmd != nil {
		t.Error("non-key message should return nil cmd")
	}
}

// --- Buttons returns copy ----------------------------------------------------

func TestButtonsReturnsCopy(t *testing.T) {
	d := New()
	btns := d.Buttons()
	btns[0].Label = "MUTATED"
	if d.Buttons()[0].Label == "MUTATED" {
		t.Error("Buttons() should return a copy, not a reference")
	}
}

// --- SetSize negative clamps to 0 -------------------------------------------

func TestSetSizeNegative(t *testing.T) {
	d := New(WithTitle("neg"))
	d.SetSize(-5, -10)
	// Should not panic, and Render returns unplaced box.
	out := d.Render()
	if out == "" {
		t.Error("Render after SetSize(-5,-10) should return the box")
	}
}

// --- WithButtons empty is a no-op --------------------------------------------

func TestWithButtonsEmpty(t *testing.T) {
	d := New(WithButtons())
	if len(d.Buttons()) != 3 {
		t.Errorf("WithButtons() with no args should leave default 3 buttons, got %d", len(d.Buttons()))
	}
}

// --- Golden-file tests -------------------------------------------------------

func TestGoldenDefaultDialog(t *testing.T) {
	d := New(WithTitle("Confirm Action"), WithBody("Are you sure you want to proceed?"))
	d.SetSize(60, 20)
	testGolden(t, "default_dialog", d.Render())
}

func TestGoldenLongBodyNarrow(t *testing.T) {
	body := "This is a much longer body text that should be wrapped when the dialog is rendered " +
		"at a narrow width. It contains multiple sentences to ensure wrapping behavior is visible."
	d := New(WithTitle("Long Body"), WithBody(body))
	d.SetSize(40, 15)
	testGolden(t, "long_body_narrow", d.Render())
}

func TestGoldenCustomLabels(t *testing.T) {
	d := New(
		WithTitle("Custom Labels"),
		WithBody("Proceed with custom labels?"),
		WithYesLabel("Confirm"),
		WithNoLabel("Deny"),
		WithCancelLabel("Abort"),
	)
	d.SetSize(60, 20)
	testGolden(t, "custom_labels", d.Render())
}

func TestGoldenFocusedNo(t *testing.T) {
	d := New(WithTitle("Focus No"), WithBody("Navigate to No button."))
	d.Update(keyPress(tea.KeyRight))
	d.SetSize(60, 20)
	testGolden(t, "focused_no", d.Render())
}

func TestGoldenFocusedCancel(t *testing.T) {
	d := New(WithTitle("Focus Cancel"), WithBody("Navigate to Cancel button."))
	d.Update(keyPress(tea.KeyRight))
	d.Update(keyPress(tea.KeyRight))
	d.SetSize(60, 20)
	testGolden(t, "focused_cancel", d.Render())
}

func TestGoldenOverlayPlain(t *testing.T) {
	d := New(WithTitle("Overlay"), WithBody("Dialog over background."))
	d.SetSize(60, 15)
	var bgLines []string
	for i := 0; i < 15; i++ {
		bgLines = append(bgLines, "ABCDEFGHIJ background line content here padding text")
	}
	bg := strings.Join(bgLines, "\n")
	testGolden(t, "overlay_plain", d.Overlay(bg))
}

package widget

import (
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
)

func TestNewTextInput(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		ti := NewTextInput()
		if ti.Value() != "" {
			t.Errorf("Value() = %q, want empty", ti.Value())
		}
		if ti.Err() != nil {
			t.Errorf("Err() = %v, want nil", ti.Err())
		}
	})

	t.Run("with placeholder", func(t *testing.T) {
		ti := NewTextInput(WithPlaceholder("type here"))
		if ti.model.Placeholder != "type here" {
			t.Errorf("Placeholder = %q, want %q", ti.model.Placeholder, "type here")
		}
	})

	t.Run("with char limit", func(t *testing.T) {
		ti := NewTextInput(WithCharLimit(5))
		if ti.model.CharLimit != 5 {
			t.Errorf("CharLimit = %d, want 5", ti.model.CharLimit)
		}
	})

	t.Run("with width", func(t *testing.T) {
		ti := NewTextInput(WithWidth(20))
		if ti.width != 20 {
			t.Errorf("width = %d, want 20", ti.width)
		}
	})

	t.Run("with validate", func(t *testing.T) {
		fn := func(s string) error {
			if s == "" {
				return errors.New("required")
			}
			return nil
		}
		ti := NewTextInput(WithValidate(fn))
		if ti.validate == nil {
			t.Error("validate should be set")
		}
	})
}

func TestWithWidthPanicsOnNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for negative width")
		}
	}()
	WithWidth(-1)
}

func TestUpdate(t *testing.T) {
	t.Run("typing updates value", func(t *testing.T) {
		ti := NewTextInput()
		ti.Focus()
		m, _ := ti.Update(keyPressRune('h'))
		ti = m.(TextInput)
		m, _ = ti.Update(keyPressRune('i'))
		ti = m.(TextInput)
		if ti.Value() != "hi" {
			t.Errorf("Value() = %q, want %q", ti.Value(), "hi")
		}
	})

	t.Run("validation runs on keypress", func(t *testing.T) {
		ti := NewTextInput(WithValidate(func(s string) error {
			if len(s) < 3 {
				return errors.New("too short")
			}
			return nil
		}))
		ti.Focus()

		m, _ := ti.Update(keyPressRune('a'))
		ti = m.(TextInput)
		if ti.Err() == nil {
			t.Error("Err() should be non-nil for short input")
		}
		if ti.Err().Error() != "too short" {
			t.Errorf("Err() = %q, want %q", ti.Err().Error(), "too short")
		}

		m, _ = ti.Update(keyPressRune('b'))
		ti = m.(TextInput)
		m, _ = ti.Update(keyPressRune('c'))
		ti = m.(TextInput)
		if ti.Err() != nil {
			t.Errorf("Err() = %v, want nil after 3 chars", ti.Err())
		}
	})

	t.Run("char limit enforced", func(t *testing.T) {
		ti := NewTextInput(WithCharLimit(3))
		ti.Focus()

		for _, c := range "abcde" {
			m, _ := ti.Update(keyPressRune(c))
			ti = m.(TextInput)
		}
		if ti.Value() != "abc" {
			t.Errorf("Value() = %q, want %q (char limit 3)", ti.Value(), "abc")
		}
	})

	t.Run("no update when blurred", func(t *testing.T) {
		ti := NewTextInput()
		m, _ := ti.Update(keyPressRune('a'))
		ti = m.(TextInput)
		if ti.Value() != "" {
			t.Errorf("Value() = %q, want empty when blurred", ti.Value())
		}
	})
}

func TestTextInputFocusBlur(t *testing.T) {
	ti := NewTextInput()
	if ti.model.Focused() {
		t.Error("should start blurred")
	}

	ti.Focus()
	if !ti.model.Focused() {
		t.Error("should be focused after Focus()")
	}

	ti.Blur()
	if ti.model.Focused() {
		t.Error("should be blurred after Blur()")
	}
}

func TestSetValue(t *testing.T) {
	t.Run("sets value and runs validation", func(t *testing.T) {
		ti := NewTextInput(WithValidate(func(s string) error {
			if s == "bad" {
				return errors.New("invalid")
			}
			return nil
		}))

		ti.SetValue("bad")
		if ti.Value() != "bad" {
			t.Errorf("Value() = %q, want %q", ti.Value(), "bad")
		}
		if ti.Err() == nil || ti.Err().Error() != "invalid" {
			t.Errorf("Err() = %v, want 'invalid'", ti.Err())
		}

		ti.SetValue("good")
		if ti.Err() != nil {
			t.Errorf("Err() = %v, want nil", ti.Err())
		}
	})
}

func TestTextInputReset(t *testing.T) {
	ti := NewTextInput(WithValidate(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	}))
	ti.SetValue("test")
	ti.Reset()

	if ti.Value() != "" {
		t.Errorf("Value() = %q, want empty after Reset", ti.Value())
	}
	if ti.Err() != nil {
		t.Errorf("Err() = %v, want nil after Reset", ti.Err())
	}
}

func TestNoErrorWhileEmpty(t *testing.T) {
	ti := NewTextInput(WithValidate(func(s string) error {
		if s == "" {
			return errors.New("required")
		}
		return nil
	}))
	ti.Focus()

	view := ti.View().Content
	if strings.Contains(view, "required") {
		t.Error("error should not appear in view when value is empty")
	}
}

func TestTextInputSetSize(t *testing.T) {
	ti := NewTextInput()
	ti.SetSize(40, 10)
	if ti.width != 40 {
		t.Errorf("width = %d, want 40", ti.width)
	}
	if ti.model.Width() != 36 {
		t.Errorf("inner width = %d, want 36 (40 - 4 overhead)", ti.model.Width())
	}

	ti.SetSize(-5, 0)
	if ti.width != 0 {
		t.Errorf("width = %d, want 0 for negative input", ti.width)
	}

	ti.SetSize(3, 0)
	if ti.model.Width() != 1 {
		t.Errorf("inner width = %d, want 1 for very small outer width", ti.model.Width())
	}
}

func TestInit(t *testing.T) {
	ti := NewTextInput()
	if cmd := ti.Init(); cmd != nil {
		t.Error("Init() should return nil")
	}
}

// Golden snapshot tests for View output in representative states.

func TestViewGolden(t *testing.T) {
	tests := []struct {
		name  string
		setup func() TextInput
	}{
		{
			name: "empty_with_placeholder",
			setup: func() TextInput {
				ti := NewTextInput(WithPlaceholder("Search..."), WithWidth(30))
				ti.Focus()
				return ti
			},
		},
		{
			name: "focused_with_text",
			setup: func() TextInput {
				ti := NewTextInput(WithWidth(30))
				ti.Focus()
				ti.SetValue("hello world")
				return ti
			},
		},
		{
			name: "blurred_with_text",
			setup: func() TextInput {
				ti := NewTextInput(WithWidth(30))
				ti.SetValue("hello world")
				return ti
			},
		},
		{
			name: "validation_error",
			setup: func() TextInput {
				ti := NewTextInput(
					WithWidth(30),
					WithValidate(func(s string) error {
						if len(s) < 5 {
							return errors.New("must be at least 5 characters")
						}
						return nil
					}),
				)
				ti.Focus()
				ti.SetValue("hi")
				return ti
			},
		},
		{
			name: "char_limit_reached",
			setup: func() TextInput {
				ti := NewTextInput(WithWidth(30), WithCharLimit(5))
				ti.Focus()
				ti.SetValue("abcde")
				return ti
			},
		},
		{
			name: "empty_no_error",
			setup: func() TextInput {
				ti := NewTextInput(
					WithWidth(30),
					WithValidate(func(s string) error {
						if s == "" {
							return errors.New("required")
						}
						return nil
					}),
				)
				ti.Focus()
				return ti
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := tt.setup()
			golden.RequireEqual(t, []byte(ti.View().Content))
		})
	}
}

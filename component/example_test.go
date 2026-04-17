package component

import tea "charm.land/bubbletea/v2"

// minimal is a compile-only sample Component used to illustrate the
// interface's surface in godoc. It is not intended for production use.
type minimal struct {
	width, height int
	focused       bool
}

func (m *minimal) Init() tea.Cmd { return nil }

func (m *minimal) Update(_ tea.Msg) (tea.Model, tea.Cmd) { return m, nil }

func (m *minimal) View() tea.View { return tea.NewView("minimal") }

func (m *minimal) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *minimal) Focus() { m.focused = true }

func (m *minimal) Blur() { m.focused = false }

// Example shows the minimal implementation of the Component interface:
// a tea.Model that also supports SetSize, Focus, and Blur.
func Example() {
	var c Component = &minimal{}
	c.SetSize(80, 24)
	c.Focus()
	c.Blur()
	_ = c.Init()
	_ = c.View()
}

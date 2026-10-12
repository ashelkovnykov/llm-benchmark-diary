package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
)

func (m model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.mode = modeSelect
		m.inputErr = ""
		return m, nil
	case "enter":
		name := m.input.Value()
		if err := db.ValidateNewName(name); err != nil {
			m.inputErr = err.Error()
			return m, nil
		}
		m.inputErr = ""
		return m, createCmd(name)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create a new database") + "\n\n")
	b.WriteString(m.input.View() + "\n\n")
	if m.inputErr != "" {
		b.WriteString(errorStyle.Render("Invalid name: "+m.inputErr) + "\n\n")
	}
	b.WriteString(mutedStyle.Render("enter: create · esc: back"))
	return b.String()
}

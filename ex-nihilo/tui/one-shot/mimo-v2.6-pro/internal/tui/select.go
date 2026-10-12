package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.names)-1 {
			m.cursor++
		}
	case "n":
		m.mode = modeCreate
		m.status = ""
		m.inputErr = ""
		m.input.SetValue("")
		m.input.Focus()
		return m, textinput.Blink
	case "enter":
		if len(m.names) == 0 {
			return m, nil
		}
		m.status = ""
		return m, openCmd(m.names[m.cursor])
	}
	return m, nil
}

func (m model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a database") + "\n\n")

	if len(m.names) == 0 {
		b.WriteString(mutedStyle.Render("No databases found in ./db/.") + "\n")
	} else {
		for i, name := range m.names {
			if i == m.cursor {
				b.WriteString(cursorStyle.Render("> "+name) + "\n")
			} else {
				b.WriteString("  " + name + "\n")
			}
		}
	}

	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(errorStyle.Render(m.status) + "\n")
	}
	b.WriteString(mutedStyle.Render("up/k down/j: move · enter: open · n: new database · q: quit"))
	return b.String()
}

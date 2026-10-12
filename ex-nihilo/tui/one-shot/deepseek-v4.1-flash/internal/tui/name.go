package tui

import (
	"strings"

	"dbnav/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// createDBMsg asks the root model to create and open a new database.
type createDBMsg struct{ name string }

// cancelMsg asks the root model to return to the chooser.
type cancelMsg struct{}

type nameModel struct {
	input    textinput.Model
	existing []string
	err      string
}

func newName(existing []string) nameModel {
	ti := textinput.New()
	ti.Placeholder = "my_database"
	ti.CharLimit = 64
	ti.Prompt = "> "

	return nameModel{input: ti, existing: existing}
}

func (n *nameModel) reset() tea.Cmd {
	n.input.SetValue("")
	n.err = ""
	return n.input.Focus()
}

func (n nameModel) Update(msg tea.Msg) (nameModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			return n, func() tea.Msg { return cancelMsg{} }
		case "enter":
			name := strings.TrimSpace(n.input.Value())
			if err := db.ValidateName(name, n.existing); err != nil {
				n.err = err.Error()
				return n, nil
			}
			return n, func() tea.Msg { return createDBMsg{name: name} }
		}
	}

	var cmd tea.Cmd
	n.input, cmd = n.input.Update(msg)
	n.err = ""
	return n, cmd
}

func (n nameModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create a new database"))
	b.WriteString("\n\n")
	b.WriteString("Name (letters and digits only):\n")
	b.WriteString(n.input.View())
	b.WriteString("\n\n")

	if n.err != "" {
		b.WriteString(errorStyle.Render("x " + n.err))
		b.WriteString("\n\n")
	}
	b.WriteString(subtleStyle.Render("enter: create  -  esc: back"))
	return b.String()
}

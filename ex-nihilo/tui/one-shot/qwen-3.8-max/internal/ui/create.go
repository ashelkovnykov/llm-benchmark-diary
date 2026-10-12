package ui

import (
	"fmt"
	"regexp"
	"strings"

	"dbnav/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

var validName = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type createModel struct {
	input textinput.Model
	err   string
}

func newCreateModel() createModel {
	ti := textinput.New()
	ti.Placeholder = "database name"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 32
	return createModel{input: ti}
}

func (m createModel) init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	c := &m.create

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.screen = screenSelect
			m.select_ = newSelectModel()
			return m, nil
		case "enter":
			name := strings.TrimSpace(c.input.Value())
			if reason := validateName(name, m.select_.dbs); reason != "" {
				c.err = reason
				return m, nil
			}
			if err := db.CreateDB(name); err != nil {
				c.err = err.Error()
				return m, nil
			}
			// Immediately load the new DB
			database, err := db.Open(name)
			if err != nil {
				c.err = err.Error()
				return m, nil
			}
			total, records, err := db.FetchRecords(database, 10)
			database.Close()
			if err != nil {
				c.err = err.Error()
				return m, nil
			}
			m.screen = screenView
			m.view = newViewModel(name, total, records)
			return m, nil
		}
	}

	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)
	return m, cmd
}

func validateName(name string, existing []string) string {
	if name == "" {
		return "Name cannot be empty."
	}
	if !validName.MatchString(name) {
		return "Name must contain only alphanumeric characters."
	}
	for _, e := range existing {
		if strings.EqualFold(name, e) {
			return fmt.Sprintf("Database %q already exists.", name)
		}
	}
	return ""
}

func (m createModel) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Create New Database"))
	b.WriteString("\n\n")
	b.WriteString(m.input.View())
	b.WriteString("\n")

	if m.err != "" {
		b.WriteString("\n")
		b.WriteString(errStyle.Render(m.err))
	}

	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter create • esc back • ctrl+c quit"))

	return b.String()
}

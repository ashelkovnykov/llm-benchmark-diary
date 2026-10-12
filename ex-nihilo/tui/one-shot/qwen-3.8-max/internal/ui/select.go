package ui

import (
	"fmt"
	"strings"

	"dbnav/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	titleStyle  = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type selectModel struct {
	dbs     []string
	cursor  int
	loading bool
	err     string
}

func newSelectModel() selectModel {
	dbs, _ := db.ListDBs()
	return selectModel{dbs: dbs}
}

func (m selectModel) init() tea.Cmd {
	return nil
}

func (m *Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	s := &m.select_

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if s.cursor > 0 {
				s.cursor--
			}
		case "down", "j":
			if s.cursor < len(s.dbs) { // last item is "Create New"
				s.cursor++
			}
		case "enter":
			if s.cursor == len(s.dbs) {
				// "Create New" selected
				m.screen = screenCreate
				m.create = newCreateModel()
				return m, m.create.init()
			}
			// Load selected DB
			name := s.dbs[s.cursor]
			database, err := db.Open(name)
			if err != nil {
				s.err = err.Error()
				return m, nil
			}
			total, records, err := db.FetchRecords(database, 10)
			database.Close()
			if err != nil {
				s.err = err.Error()
				return m, nil
			}
			m.screen = screenView
			m.view = newViewModel(name, total, records)
			return m, nil
		}
	}
	return m, nil
}

func (m selectModel) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Select a Database"))
	b.WriteString("\n")

	if len(m.dbs) == 0 {
		b.WriteString(dimStyle.Render("  No databases found in ./db/\n"))
	}

	for i, name := range m.dbs {
		cursor := "  "
		if i == m.cursor {
			cursor = cursorStyle.Render("▸ ")
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, name))
	}

	// "Create New" option
	cursor := "  "
	if m.cursor == len(m.dbs) {
		cursor = cursorStyle.Render("▸ ")
	}
	b.WriteString(fmt.Sprintf("%s%s\n", cursor, lipgloss.NewStyle().Italic(true).Render("+ Create New")))

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("↑/↓ navigate • enter select • q quit"))

	if m.err != "" {
		b.WriteString("\n\n")
		b.WriteString(errStyle.Render(m.err))
	}

	return b.String()
}

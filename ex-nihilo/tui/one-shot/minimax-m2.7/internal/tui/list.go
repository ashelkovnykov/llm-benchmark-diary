package tui

import (
	"dbviewer/internal/db"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type dbListModel struct {
	databases  []string
	selected   int
	newDBName  string
	newDBError string
	creating   bool
	onSelect   func(string)
	onCreate   func(string) error
	onCancel   func()
}

func NewDBList(databases []string, onSelect func(string), onCreate func(string) error) tea.Model {
	return dbListModel{
		databases: databases,
		selected:  0,
		onSelect:  onSelect,
		onCreate:  onCreate,
	}
}

func (m dbListModel) Init() tea.Cmd {
	return nil
}

func (m dbListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.creating {
			switch msg.Type {
			case tea.KeyRunes:
				runes := msg.Runes
				if len(runes) == 1 {
					c := runes[0]
					if c == 127 || c == 8 {
						if len(m.newDBName) > 0 {
							m.newDBName = m.newDBName[:len(m.newDBName)-1]
							m.newDBError = ""
						}
						return m, nil
					}
					if c >= 32 && c < 127 {
						m.newDBName += string(c)
						m.newDBError = ""
						return m, nil
					}
				}
			case tea.KeyEnter:
				ok, errMsg := db.ValidateName(m.newDBName, m.databases)
				if !ok {
					m.newDBError = errMsg
					return m, nil
				}
				if err := m.onCreate(m.newDBName); err != nil {
					m.newDBError = err.Error()
					return m, nil
				}
				m.onSelect(m.newDBName)
				return m, tea.Quit
			case tea.KeyEsc:
				m.creating = false
				m.newDBName = ""
				m.newDBError = ""
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.databases)-1 {
				m.selected++
			}
		case "enter":
			if len(m.databases) > 0 {
				name := strings.TrimSuffix(m.databases[m.selected], ".db")
				m.onSelect(name)
				return m, tea.Quit
			}
		case "n":
			m.creating = true
			m.newDBName = ""
			m.newDBError = ""
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m dbListModel) View() string {
	var s strings.Builder

	if m.creating {
		s.WriteString("Create New Database\n\n")
		s.WriteString("Name: ")
		s.WriteString(m.newDBName)
		s.WriteString("\n")

		if m.newDBError != "" {
			s.WriteString("Error: ")
			s.WriteString(m.newDBError)
			s.WriteString("\n")
		}

		s.WriteString("\nPress Enter to create, Esc to cancel.\n")
		s.WriteString("Type the database name above.\n")
		return s.String()
	}

	s.WriteString("Select a Database\n\n")

	if len(m.databases) == 0 {
		s.WriteString("No databases found.\n\n")
		s.WriteString("Press n to create a new database.\n")
		s.WriteString("Press q to quit.\n")
		return s.String()
	}

	s.WriteString("Available databases:\n\n")
	for i, dbName := range m.databases {
		prefix := "  "
		if i == m.selected {
			prefix = "> "
		}
		name := strings.TrimSuffix(dbName, ".db")
		s.WriteString(prefix)
		s.WriteString(name)
		s.WriteString("\n")
	}

	s.WriteString("\nPress Enter to select, n to create new, q to quit.\n")
	return s.String()
}

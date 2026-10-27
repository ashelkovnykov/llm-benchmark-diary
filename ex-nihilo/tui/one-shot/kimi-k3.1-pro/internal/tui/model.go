package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dbtui/internal/db"
)

type screen int

const (
	selectScreen screen = iota
	createScreen
	dataScreen
)

type Model struct {
	screen   screen
	names    []string
	cursor   int
	input    string
	err      string
	name     string
	database *db.DB
	snapshot db.Snapshot
}

func New() (Model, error) {
	names, err := db.List()
	if err != nil {
		return Model{}, err
	}
	return Model{names: names}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if key.String() == "ctrl+c" || (key.String() == "q" && m.screen != createScreen) {
		m.Close()
		return m, tea.Quit
	}

	switch m.screen {
	case selectScreen:
		return m.updateSelect(key)
	case createScreen:
		return m.updateCreate(key)
	}

	return m, nil
}

func (m Model) updateSelect(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.names)-1 {
			m.cursor++
		}
	case "n":
		m.screen = createScreen
		m.input = ""
		m.err = ""
	case "enter":
		if len(m.names) > 0 {
			m.load(m.names[m.cursor])
		}
	}
	return m, nil
}

func (m Model) updateCreate(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		m.screen = selectScreen
		m.input = ""
		m.err = ""
	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case "enter":
		if err := db.Create(m.input); err != nil {
			m.err = err.Error()
			return m, nil
		}

		name := m.input
		m.names, _ = db.List()
		m.input = ""
		m.err = ""
		m.load(name)
	default:
		switch key.Type {
		case tea.KeyRunes:
			m.input += string(key.Runes)
		case tea.KeySpace:
			m.input += " "
		}
	}

	return m, nil
}

func (m *Model) load(name string) {
	database, err := db.Open(name)
	if err != nil {
		m.err = err.Error()
		return
	}

	snapshot, err := database.Snapshot()
	if err != nil {
		database.Close()
		m.err = err.Error()
		return
	}

	m.Close()
	m.database = database
	m.snapshot = snapshot
	m.name = name
	m.screen = dataScreen
	m.err = ""
}

func (m Model) Close() {
	if m.database != nil {
		_ = m.database.Close()
	}
}

func (m Model) View() string {
	switch m.screen {
	case createScreen:
		return m.createView()
	case dataScreen:
		return m.dataView()
	default:
		return m.selectView()
	}
}

func (m Model) selectView() string {
	var view strings.Builder
	view.WriteString("SQLite Databases\n\n")

	if len(m.names) == 0 {
		view.WriteString("No databases found in ./db/.\n")
	} else {
		for i, name := range m.names {
			cursor := "  "
			if i == m.cursor {
				cursor = "> "
			}
			fmt.Fprintf(&view, "%s%s.db\n", cursor, name)
		}
	}

	if m.err != "" {
		fmt.Fprintf(&view, "\nError: %s\n", m.err)
	}

	view.WriteString("\n↑/↓ or j/k: navigate • enter: open • n: new database • q: quit\n")
	return view.String()
}

func (m Model) createView() string {
	var view strings.Builder
	view.WriteString("Create Database\n\n")
	fmt.Fprintf(&view, "Name: %s_\n", m.input)
	view.WriteString("Alphanumeric characters only. The .db extension is added automatically.\n")

	if m.err != "" {
		fmt.Fprintf(&view, "\nInvalid name: %s\n", m.err)
	}

	view.WriteString("\nenter: create • esc: back • ctrl+c: quit\n")
	return view.String()
}

func (m Model) dataView() string {
	var view strings.Builder
	fmt.Fprintf(&view, "Database: %s.db\n", m.name)
	fmt.Fprintf(&view, "Total records: %d\n\n", m.snapshot.Total)
	view.WriteString(strings.Join(m.snapshot.Columns, " | "))
	view.WriteString("\n")

	if len(m.snapshot.Rows) == 0 {
		view.WriteString("(no rows)\n")
	}
	for _, row := range m.snapshot.Rows {
		view.WriteString(strings.Join(row, " | "))
		view.WriteString("\n")
	}

	view.WriteString("\nShowing up to the first 10 rows • q: quit\n")
	return view.String()
}

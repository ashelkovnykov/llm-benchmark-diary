package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	dbstore "sqlnav/internal/db"
)

type screen int

const (
	chooseScreen screen = iota
	nameScreen
	dataScreen
)

type model struct {
	screen    screen
	databases []string
	cursor    int
	name      string
	message   string

	count   int
	columns []string
	rows    [][]string
}

func NewModel() tea.Model {
	databases, err := dbstore.List()
	m := &model{databases: databases}
	if err != nil {
		m.message = "Could not list databases: " + err.Error()
	}
	return m
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := message.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch key.String() {
	case "ctrl+c", "q":
		if m.screen != dataScreen {
			return m, tea.Quit
		}
	case "esc":
		if m.screen == nameScreen {
			m.screen = chooseScreen
			m.message = ""
		} else if m.screen == dataScreen {
			m.screen = chooseScreen
		}
		return m, nil
	}

	switch m.screen {
	case chooseScreen:
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.databases) {
				m.cursor++
			}
		case "n":
			m.screen, m.name, m.message = nameScreen, "", ""
		case "enter":
			if m.cursor < len(m.databases) {
				m.open(filepath.Join("./db", m.databases[m.cursor]+".db"))
			} else {
				m.screen, m.name, m.message = nameScreen, "", ""
			}
		}

	case nameScreen:
		switch key.String() {
		case "enter":
			m.create()
		case "backspace":
			if len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			}
		default:
			if len(key.Runes) > 0 {
				m.name += string(key.Runes)
			}
		}

	case dataScreen:
		// Press Esc to return to the database chooser.
	}

	return m, nil
}

func (m *model) create() {
	name := m.name
	if name == "" {
		m.message = "Invalid name: it cannot be empty."
		return
	}
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) || r > unicode.MaxASCII {
			m.message = "Invalid name: use alphanumeric characters only."
			return
		}
	}
	for _, existing := range m.databases {
		if existing == name {
			m.message = "Invalid name: a database with that name already exists."
			return
		}
	}

	path, err := dbstore.Create(name)
	if err != nil {
		m.message = "Could not create database: " + err.Error()
		return
	}
	m.open(path)
}

func (m *model) open(path string) {
	count, columns, rows, err := dbstore.Load(path)
	if err != nil {
		m.message = "Could not load database: " + err.Error()
		return
	}
	m.count, m.columns, m.rows = count, columns, rows
	m.screen, m.message = dataScreen, ""
}

func (m *model) View() string {
	var b strings.Builder

	switch m.screen {
	case chooseScreen:
		b.WriteString("Choose a database\n\n")
		for i, name := range m.databases {
			prefix := "  "
			if i == m.cursor {
				prefix = "> "
			}
			fmt.Fprintf(&b, "%s%s\n", prefix, name)
		}

		prefix := "  "
		if m.cursor == len(m.databases) {
			prefix = "> "
		}
		fmt.Fprintf(&b, "%sCreate a new database\n", prefix)
		if len(m.databases) == 0 {
			b.WriteString("\nNo databases found.\n")
		}
		b.WriteString("\n↑/↓ navigate • enter select • n new • q quit")

	case nameScreen:
		b.WriteString("Create a new database\n\n")
		fmt.Fprintf(&b, "Name: %s_\n", m.name)
		b.WriteString("\nUse letters and numbers only • enter create • esc back")
		if m.message != "" {
			fmt.Fprintf(&b, "\n\n%s", m.message)
		}

	case dataScreen:
		fmt.Fprintf(&b, "Records: %d\n\n", m.count)
		if len(m.rows) == 0 {
			b.WriteString("No records.\n")
		} else {
			b.WriteString(strings.Join(m.columns, " | ") + "\n")
			for _, row := range m.rows {
				b.WriteString(strings.Join(row, " | ") + "\n")
			}
		}
		b.WriteString("\nEsc back • q quit")
	}

	return b.String()
}

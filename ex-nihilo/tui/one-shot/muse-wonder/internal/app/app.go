package app

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"dbtui/internal/db"
)

type state int

const (
	stateSelect state = iota
	stateCreate
	stateLoaded
)

type model struct {
	state    state
	dbs      []string
	cursor   int
	input    string
	errMsg   string
	conn     *sql.DB
	dbName   string
	count    int
	rows     []string
	loadErr  error
	quitting bool
}

func Run() error {
	if err := db.EnsureDir(); err != nil {
		return err
	}
	names, err := db.List()
	if err != nil {
		return err
	}
	p := tea.NewProgram(model{state: stateSelect, dbs: names}, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

func (m model) Init() tea.Cmd { return nil }

func (m model) load(name string, create bool) (model, tea.Cmd) {
	var conn *sql.DB
	var err error
	if create {
		conn, err = db.Create(name)
	} else {
		conn, err = db.Open(name)
	}
	if err != nil {
		m.loadErr = err
		return m, tea.Quit
	}
	m.conn = conn
	m.dbName = name
	m.state = stateLoaded
	m.count, m.loadErr = db.Count(conn)
	if m.loadErr == nil {
		m.rows, m.loadErr = db.FirstRows(conn, 10)
	}
	if m.loadErr != nil {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch m.state {
	case stateSelect:
		switch key.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.dbs) { // last index is "Create new"
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.dbs) {
				m.state = stateCreate
				m.input, m.errMsg = "", ""
				return m, nil
			}
			return m.load(m.dbs[m.cursor], false)
		case "n":
			m.state = stateCreate
			m.input, m.errMsg = "", ""
		}

	case stateCreate:
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateSelect
			m.errMsg = ""
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		case "enter":
			if err := db.ValidateName(m.input, m.dbs); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			name := m.input
			m.dbs = append(m.dbs, name)
			return m.load(name, true)
		default:
			if key.Type == tea.KeyRunes {
				m.input += string(key.Runes)
			}
		}

	case stateLoaded:
		switch key.String() {
		case "ctrl+c", "q", "esc":
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.loadErr != nil {
		return fmt.Sprintf("Error: %v\n", m.loadErr)
	}

	switch m.state {
	case stateSelect:
		var b strings.Builder
		b.WriteString("Select a database (↑/↓, enter, n=new, q=quit):\n\n")
		if len(m.dbs) == 0 {
			b.WriteString("  (no databases found in ./db/)\n")
		}
		for i, name := range m.dbs {
			cursor := "  "
			if m.cursor == i {
				cursor = "> "
			}
			b.WriteString(cursor + name + db.Ext + "\n")
		}
		cursor := "  "
		if m.cursor == len(m.dbs) {
			cursor = "> "
		}
		b.WriteString(cursor + "[ Create new database ]\n")
		return b.String()

	case stateCreate:
		var b strings.Builder
		b.WriteString("Create new database\n\n")
		b.WriteString("Name (alphanumeric only): " + m.input + "█\n")
		if m.errMsg != "" {
			b.WriteString("\nError: " + m.errMsg + "\n")
		}
		b.WriteString("\nenter=create, esc=back\n")
		return b.String()

	case stateLoaded:
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Database: %s%s\n", m.dbName, db.Ext))
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.count))
		b.WriteString("First rows:\n")
		for _, r := range m.rows {
			b.WriteString(r + "\n")
		}
		b.WriteString("\nq=quit\n")
		return b.String()
	}
	return ""
}

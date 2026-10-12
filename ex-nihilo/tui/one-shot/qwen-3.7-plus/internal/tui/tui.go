package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"dbnav/internal/db"
	"github.com/charmbracelet/bubbletea"
)

type appState int

const (
	stateSelect appState = iota
	stateCreate
	stateView
)

type Model struct {
	state   appState
	dbs     []string
	cursor  int
	input   string
	errMsg  string
	dbConn  *sql.DB
	count   int
	records []db.Record
}

func InitialModel() Model {
	dbs, _ := db.ListDBs()
	return Model{
		state: stateSelect,
		dbs:   dbs,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateSelect:
			return m.updateSelect(msg)
		case stateCreate:
			return m.updateCreate(msg)
		case stateView:
			return m.updateView(msg)
		}
	}
	return m, nil
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.dbs)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.dbs) > 0 {
			conn, err := db.OpenDB(m.dbs[m.cursor])
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.dbConn = conn
			count, records, err := db.GetStats(conn)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.count = count
			m.records = records
			m.state = stateView
		}
	case "n":
		m.state = stateCreate
		m.input = ""
		m.errMsg = ""
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.state = stateSelect
		m.errMsg = ""
	case "enter":
		err := db.ValidateName(m.input, m.dbs)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		conn, err := db.CreateDB(m.input)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.dbConn = conn
		count, records, err := db.GetStats(conn)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.count = count
		m.records = records
		m.state = stateView
	case "backspace":
		runes := []rune(m.input)
		if len(runes) > 0 {
			m.input = string(runes[:len(runes)-1])
		}
	default:
		if len(msg.String()) == 1 {
			m.input += msg.String()
		}
	}
	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		if m.dbConn != nil {
			m.dbConn.Close()
		}
		return m, tea.Quit
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	switch m.state {
	case stateSelect:
		b.WriteString("Select a database to load (or press 'n' to create new):\n\n")
		if len(m.dbs) == 0 {
			b.WriteString("  (no databases found)\n")
		} else {
			for i, d := range m.dbs {
				cursor := "  "
				if m.cursor == i {
					cursor = "> "
				}
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, d))
			}
		}
	case stateCreate:
		b.WriteString("Enter new database name (alphanumeric only):\n\n")
		b.WriteString(fmt.Sprintf("> %s\n", m.input))
	case stateView:
		b.WriteString(fmt.Sprintf("Database loaded. Total records: %d\n\n", m.count))
		b.WriteString("First 10 rows:\n")
		b.WriteString(fmt.Sprintf("%-10s | %s\n", "rowid", "value"))
		b.WriteString(strings.Repeat("-", 30) + "\n")
		for _, r := range m.records {
			b.WriteString(fmt.Sprintf("%-10d | %s\n", r.RowID, r.Value))
		}
		b.WriteString("\nPress 'q' to quit.\n")
	}

	if m.errMsg != "" {
		b.WriteString(fmt.Sprintf("\nError: %s\n", m.errMsg))
	}

	return b.String()
}

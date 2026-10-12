package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"dbnav/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSelect state = iota
	stateCreate
	stateView
)

type model struct {
	state   state
	dbs     []string
	cursor  int
	input   textinput.Model
	errMsg  string
	db      *sql.DB
	dbName  string
	total   int
	records []db.Record
	err     error
}

func InitialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter database name"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30

	dbs, err := db.ListDBs()
	m := model{
		state: stateSelect,
		dbs:   dbs,
		input: ti,
	}
	if err != nil {
		m.errMsg = fmt.Sprintf("Error listing DBs: %v", err)
	}
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelect:
		return m.updateSelect(msg)
	case stateCreate:
		return m.updateCreate(msg)
	case stateView:
		return m.updateView(msg)
	}
	return m, nil
}

func (m model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.dbs) {
				m.cursor++
			}
		case "enter":
			if m.cursor == len(m.dbs) {
				// new
				m.state = stateCreate
				m.errMsg = ""
				m.input.SetValue("")
				return m, textinput.Blink
			}
			if m.cursor < len(m.dbs) {
				name := m.dbs[m.cursor]
				return m.loadDB(name)
			}
		case "n", "N":
			m.state = stateCreate
			m.errMsg = ""
			m.input.SetValue("")
			return m, textinput.Blink
		}
	}
	return m, nil
}

func (m model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateSelect
			m.errMsg = ""
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.input.Value())
			reason, ok := db.ValidateName(name, m.dbs)
			if !ok {
				m.errMsg = reason
				return m, nil
			}
			path, err := db.CreateDB(name)
			if err != nil {
				m.errMsg = fmt.Sprintf("Failed to create: %v", err)
				return m, nil
			}
			_ = path
			// reload list
			if dbs, err := db.ListDBs(); err == nil {
				m.dbs = dbs
			}
			return m.loadDB(name)
		}
	}
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.db != nil {
				m.db.Close()
			}
			return m, tea.Quit
		case "esc", "b":
			if m.db != nil {
				m.db.Close()
				m.db = nil
			}
			m.state = stateSelect
			return m, nil
		}
	}
	return m, nil
}

func (m model) loadDB(name string) (tea.Model, tea.Cmd) {
	dbConn, path, err := db.OpenDB(name)
	if err != nil {
		m.errMsg = fmt.Sprintf("Failed to open %s: %v", name, err)
		m.state = stateSelect
		return m, nil
	}
	_ = path
	total, recs, err := db.GetStats(dbConn)
	if err != nil {
		dbConn.Close()
		m.errMsg = fmt.Sprintf("Failed to query: %v", err)
		m.state = stateSelect
		return m, nil
	}
	m.db = dbConn
	m.dbName = name
	m.total = total
	m.records = recs
	m.state = stateView
	m.errMsg = ""
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateSelect:
		return m.viewSelect()
	case stateCreate:
		return m.viewCreate()
	case stateView:
		return m.viewView()
	}
	return ""
}

func (m model) viewSelect() string {
	s := "Select a database to load:\n\n"
	if len(m.dbs) == 0 {
		s += "  (no databases found)\n"
	}
	for i, name := range m.dbs {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, name)
	}
	cursor := " "
	if m.cursor == len(m.dbs) {
		cursor = ">"
	}
	s += fmt.Sprintf("%s [Create new database]\n", cursor)
	s += "\nUse ↑/↓ or j/k to navigate, Enter to select, n for new, q to quit.\n"
	if m.errMsg != "" {
		s += "\nError: " + m.errMsg + "\n"
	}
	return s
}

func (m model) viewCreate() string {
	s := "Create new database\n\n"
	s += m.input.View() + "\n\n"
	s += "Press Enter to create, Esc to cancel.\n"
	if m.errMsg != "" {
		s += "\nInvalid: " + m.errMsg + "\n"
	}
	return s
}

func (m model) viewView() string {
	s := fmt.Sprintf("Loaded database: %s\n\n", m.dbName)
	s += fmt.Sprintf("Total records: %d\n\n", m.total)
	s += "First 10 rows:\n"
	s += db.FormatRecords(m.records)
	s += "\n\nPress b or Esc to go back, q to quit.\n"
	return s
}

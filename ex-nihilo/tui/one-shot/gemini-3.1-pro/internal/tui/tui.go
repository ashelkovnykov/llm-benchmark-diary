package tui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/db-tui/internal/db"
)

type state int

const (
	stateList state = iota
	stateCreate
	stateView
)

type model struct {
	state     state
	dbs       []string
	cursor    int
	textInput textinput.Model
	err       string
	activeDB  string
	totalRows int
	records   []db.Record
}

func InitialModel() tea.Model {
	ti := textinput.New()
	ti.Placeholder = "Enter new database name..."
	ti.Focus()

	dbs, _ := db.GetDatabases()

	return model{
		state:     stateList,
		dbs:       dbs,
		textInput: ti,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.state != stateCreate {
				return m, tea.Quit
			}
		}
	}

	switch m.state {
	case stateList:
		return m.updateList(msg)
	case stateCreate:
		return m.updateCreate(msg)
	case stateView:
		return m.updateView(msg)
	}
	return m, nil
}

func (m model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
				m.state = stateCreate
				m.err = ""
				m.textInput.SetValue("")
				return m, nil
			}
			m.activeDB = m.dbs[m.cursor]
			return m.loadActiveDB()
		}
	}
	return m, nil
}

func (m model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateList
			m.err = ""
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.textInput.Value())
			if name == "" {
				m.err = "Name cannot be empty."
				return m, nil
			}
			isAlphanumeric := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
			if !isAlphanumeric(name) {
				m.err = "Name must be alphanumeric."
				return m, nil
			}
			for _, existing := range m.dbs {
				if existing == name+".db" || existing == name {
					m.err = "Database already exists."
					return m, nil
				}
			}

			err := db.CreateDatabase(name)
			if err != nil {
				m.err = fmt.Sprintf("Error creating DB: %v", err)
				return m, nil
			}

			m.activeDB = name + ".db"
			return m.loadActiveDB()
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.state = stateList
			var err error
			m.dbs, err = db.GetDatabases()
			if err != nil {
				m.err = err.Error()
			}
			m.cursor = 0
		}
	}
	return m, nil
}

func (m model) loadActiveDB() (tea.Model, tea.Cmd) {
	total, records, err := db.GetDatabaseStats(m.activeDB)
	if err != nil {
		m.err = fmt.Sprintf("Error loading DB: %v", err)
		m.state = stateList
		return m, nil
	}
	m.totalRows = total
	m.records = records
	m.state = stateView
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateList:
		s := "Select a database to load:\n\n"
		for i, dbName := range m.dbs {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, dbName)
		}
		cursor := " "
		if m.cursor == len(m.dbs) {
			cursor = ">"
		}
		s += fmt.Sprintf("%s [Create New Database]\n", cursor)
		if m.err != "" {
			s += fmt.Sprintf("\nError: %s\n", m.err)
		}
		s += "\nPress q or ctrl+c to quit."
		return s

	case stateCreate:
		s := "Enter new database name (alphanumeric only):\n\n"
		s += m.textInput.View() + "\n\n"
		if m.err != "" {
			s += fmt.Sprintf("Error: %s\n\n", m.err)
		}
		s += "(esc to go back)"
		return s

	case stateView:
		s := fmt.Sprintf("Database: %s\nTotal Records: %d\n\n", m.activeDB, m.totalRows)
		s += "First 10 Records:\n"
		s += "rowid\t| name\t\t\t| value\n"
		s += "----------------------------------------\n"
		for _, r := range m.records {
			s += fmt.Sprintf("%d\t| %s\t| %d\n", r.RowID, r.Name, r.Value)
		}
		s += "\n(esc to go back, q or ctrl+c to quit)"
		return s
	}
	return ""
}

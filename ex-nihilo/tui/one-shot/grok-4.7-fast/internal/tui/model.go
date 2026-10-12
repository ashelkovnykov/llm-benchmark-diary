package tui

import (
	"fmt"
	"strings"

	"dbtui/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateSelect state = iota
	stateCreate
	stateView
)

type Model struct {
	state     state
	dbs       []string
	cursor    int
	input     string
	errMsg    string
	currentDB *db.DB
	count     int
	headers   []string
	rows      [][]string
	dbDir     string
	schema    string
}

func New(dbDir, schema string) Model {
	dbs, _ := db.ListDatabases(dbDir)
	return Model{
		state:  stateSelect,
		dbs:    dbs,
		dbDir:  dbDir,
		schema: schema,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				// create new
				m.state = stateCreate
				m.input = ""
				m.errMsg = ""
				return m, nil
			}
			// load existing
			name := m.dbs[m.cursor]
			path := m.dbDir + "/" + name + ".db"
			d, err := db.Open(path)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.currentDB = d
			m.loadViewData()
			m.state = stateView
			return m, nil
		}
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			name := strings.TrimSpace(m.input)
			if err := db.ValidateName(name, m.dbs); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			d, err := db.CreateDatabase(m.dbDir, name, m.schema)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.dbs = append(m.dbs, name)
			m.currentDB = d
			m.loadViewData()
			m.state = stateView
			return m, nil
		case "backspace":
			if len(m.input) > 0 {
				m.input = m.input[:len(m.input)-1]
			}
		default:
			if len(msg.String()) == 1 {
				m.input += msg.String()
			}
		}
	}
	return m, nil
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.currentDB != nil {
				m.currentDB.Close()
				m.currentDB = nil
			}
			m.state = stateSelect
			m.errMsg = ""
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) loadViewData() {
	if m.currentDB == nil {
		return
	}
	count, err := m.currentDB.RecordCount()
	if err != nil {
		m.errMsg = err.Error()
		return
	}
	m.count = count
	rows, headers, err := m.currentDB.FirstNRecords(10)
	if err != nil {
		m.errMsg = err.Error()
		return
	}
	m.rows = rows
	m.headers = headers
}

func (m Model) View() string {
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

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	selectedStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")).Bold(true)
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m Model) viewSelect() string {
	s := titleStyle.Render("Select a database") + "\n\n"
	for i, name := range m.dbs {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
			s += selectedStyle.Render(cursor+" "+name+".db") + "\n"
		} else {
			s += itemStyle.Render(cursor+" "+name+".db") + "\n"
		}
	}
	// create option
	cursor := " "
	if m.cursor == len(m.dbs) {
		cursor = ">"
		s += selectedStyle.Render(cursor+" [Create new database]") + "\n"
	} else {
		s += itemStyle.Render(cursor+" [Create new database]") + "\n"
	}
	if m.errMsg != "" {
		s += "\n" + errorStyle.Render(m.errMsg)
	}
	s += "\n\n↑/↓ to navigate, enter to select, q to quit"
	return s
}

func (m Model) viewCreate() string {
	s := titleStyle.Render("Create new database") + "\n\n"
	s += "Name: " + m.input + "█\n"
	if m.errMsg != "" {
		s += "\n" + errorStyle.Render(m.errMsg) + "\n"
	}
	s += "\nEnter to create, esc to cancel"
	return s
}

func (m Model) viewView() string {
	name := "unknown"
	if m.currentDB != nil {
		name = strings.TrimSuffix(m.currentDB.Path(), ".db")
		// just basename
		parts := strings.Split(name, "/")
		name = parts[len(parts)-1]
	}
	s := titleStyle.Render(fmt.Sprintf("Database: %s", name)) + "\n\n"
	s += fmt.Sprintf("Total records: %d\n\n", m.count)
	if len(m.headers) > 0 {
		s += strings.Join(m.headers, " | ") + "\n"
		s += strings.Repeat("-", 40) + "\n"
		for _, row := range m.rows {
			s += strings.Join(row, " | ") + "\n"
		}
	}
	if m.errMsg != "" {
		s += "\n" + errorStyle.Render(m.errMsg)
	}
	s += "\n\nesc/q to go back"
	return s
}

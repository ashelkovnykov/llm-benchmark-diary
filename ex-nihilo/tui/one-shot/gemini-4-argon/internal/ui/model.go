package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sql-navigator/internal/db"
)

type sessionState int

const (
	stateSelectDB sessionState = iota
	stateCreateDB
	stateViewDB
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).MarginBottom(1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).MarginTop(1)
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).MarginTop(1)
	selStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("170")).Bold(true)
)

type Model struct {
	state      sessionState
	dbDir      string
	dbs        []string
	cursor     int
	textInput  textinput.Model
	errMsg     string
	loadedData *db.DBData
}

func NewModel(dbDir string) (*Model, error) {
	dbs, err := db.ListDBs(dbDir)
	if err != nil {
		return nil, err
	}

	ti := textinput.New()
	ti.Placeholder = "Enter alphanumeric DB name"
	ti.CharLimit = 64
	ti.Width = 30

	return &Model{
		state:     stateSelectDB,
		dbDir:     dbDir,
		dbs:       dbs,
		textInput: ti,
	}, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateSelectDB:
		return m.updateSelectDB(msg)
	case stateCreateDB:
		return m.updateCreateDB(msg)
	case stateViewDB:
		return m.updateViewDB(msg)
	}

	return m, nil
}

func (m *Model) updateSelectDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	totalOptions := len(m.dbs) + 1 // +1 for "Create new DB" option

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < totalOptions-1 {
				m.cursor++
			}
		case "enter":
			m.errMsg = ""
			if m.cursor == len(m.dbs) {
				m.state = stateCreateDB
				m.textInput.SetValue("")
				m.textInput.Focus()
				return m, textinput.Blink
			}
			selectedDB := m.dbs[m.cursor]
			data, err := db.LoadDBData(m.dbDir, selectedDB)
			if err != nil {
				m.errMsg = fmt.Sprintf("Failed to load DB: %v", err)
				return m, nil
			}
			m.loadedData = data
			m.state = stateViewDB
		}
	}
	return m, nil
}

func (m *Model) updateCreateDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.state = stateSelectDB
			m.errMsg = ""
			return m, nil
		case tea.KeyEnter:
			name := m.textInput.Value()
			if err := db.CreateDB(m.dbDir, name); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			data, err := db.LoadDBData(m.dbDir, name)
			if err != nil {
				m.errMsg = fmt.Sprintf("Created DB, but failed to load: %v", err)
				return m, nil
			}
			m.loadedData = data
			m.errMsg = ""
			m.state = stateViewDB
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *Model) updateViewDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "b", "esc":
			dbs, err := db.ListDBs(m.dbDir)
			if err != nil {
				m.errMsg = fmt.Sprintf("Error refreshing DB list: %v", err)
			} else {
				m.dbs = dbs
			}
			m.cursor = 0
			m.loadedData = nil
			m.state = stateSelectDB
		}
	}
	return m, nil
}

func (m *Model) View() string {
	var b strings.Builder

	switch m.state {
	case stateSelectDB:
		b.WriteString(titleStyle.Render("Select a Database"))
		b.WriteString("\n")

		if len(m.dbs) == 0 {
			b.WriteString("No existing databases found in ./db/\n\n")
		}

		for i, dbName := range m.dbs {
			cursor := "  "
			line := dbName
			if m.cursor == i {
				cursor = "> "
				line = selStyle.Render(line)
			}
			b.WriteString(fmt.Sprintf("%s%s\n", cursor, line))
		}

		createIdx := len(m.dbs)
		cursor := "  "
		createLine := "[+ Create new database]"
		if m.cursor == createIdx {
			cursor = "> "
			createLine = selStyle.Render(createLine)
		}
		b.WriteString(fmt.Sprintf("%s%s\n", cursor, createLine))

		if m.errMsg != "" {
			b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", m.errMsg)) + "\n")
		}
		b.WriteString(helpStyle.Render("\n↑/↓: navigate • enter: select • q: quit"))

	case stateCreateDB:
		b.WriteString(titleStyle.Render("Create New Database"))
		b.WriteString("\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n")

		if m.errMsg != "" {
			b.WriteString(errorStyle.Render(fmt.Sprintf("Invalid name: %s", m.errMsg)) + "\n")
		}
		b.WriteString(helpStyle.Render("\nenter: create • esc: cancel"))

	case stateViewDB:
		b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.loadedData.Name)))
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.loadedData.TotalCount))
		b.WriteString(fmt.Sprintf("%-8s | %-15s | %-20s\n", "ROWID", "NAME", "VALUE"))
		b.WriteString(strings.Repeat("-", 48) + "\n")

		for _, r := range m.loadedData.Records {
			b.WriteString(fmt.Sprintf("%-8d | %-15s | %-20s\n", r.RowID, r.Name, r.Value))
		}

		b.WriteString(helpStyle.Render("\nb/esc: back to DB list • q: quit"))
	}

	return "\n" + b.String() + "\n"
}

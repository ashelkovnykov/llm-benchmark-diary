package tui

import (
	"database/sql"
	"db-explorer/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateSelectDB state = iota
	stateCreateDB
	stateViewDB
)

type Model struct {
	state        state
	dbs          []string
	cursor       int
	selectedDB   string
	err          error
	textInput    textinput.Model
	database     *sql.DB
	totalRecords int
	records      []db.Record
	width        int
	height       int
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "Enter new DB name"
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 40

	dbs, _ := db.ListDBs()

	return Model{
		state:     stateSelectDB,
		dbs:       dbs,
		cursor:    0,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

type dbLoadedMsg struct {
	db    *sql.DB
	total int
	recs  []db.Record
}

type errMsg struct {
	err error
}

func loadDB(name string) tea.Cmd {
	return func() tea.Msg {
		database, err := db.OpenDB(name)
		if err != nil {
			return errMsg{err}
		}
		total, err := db.GetStats(database)
		if err != nil {
			return errMsg{err}
		}
		recs, err := db.GetRecords(database, 10)
		if err != nil {
			return errMsg{err}
		}
		return dbLoadedMsg{db: database, total: total, recs: recs}
	}
}

func createDB(name string) tea.Cmd {
	return func() tea.Msg {
		database, err := db.CreateDB(name)
		if err != nil {
			return errMsg{err}
		}
		total, err := db.GetStats(database)
		if err != nil {
			return errMsg{err}
		}
		recs, err := db.GetRecords(database, 10)
		if err != nil {
			return errMsg{err}
		}
		return dbLoadedMsg{db: database, total: total, recs: recs}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.state == stateViewDB {
				m.state = stateSelectDB
				if m.database != nil {
					m.database.Close()
					m.database = nil
				}
				return m, nil
			}
			return m, tea.Quit
		case "esc":
			if m.state == stateCreateDB {
				m.state = stateSelectDB
				m.textInput.SetValue("")
				m.err = nil
				return m, nil
			}
			if m.state == stateViewDB {
				m.state = stateSelectDB
				if m.database != nil {
					m.database.Close()
					m.database = nil
				}
				return m, nil
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case dbLoadedMsg:
		m.database = msg.db
		m.totalRecords = msg.total
		m.records = msg.recs
		m.state = stateViewDB
		m.err = nil
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
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

func (m Model) updateSelectDB(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.state = stateCreateDB
				m.textInput.Focus()
				return m, textinput.Blink
			}
			if len(m.dbs) > 0 {
				m.selectedDB = m.dbs[m.cursor]
				return m, loadDB(m.selectedDB)
			}
		}
	}
	return m, nil
}

func (m Model) updateCreateDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "enter" {
			name := m.textInput.Value()
			if err := db.ValidateName(name, m.dbs); err != nil {
				m.err = err
				return m, nil
			}
			m.selectedDB = name
			m.dbs = append(m.dbs, name)
			return m, createDB(name)
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) updateViewDB(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

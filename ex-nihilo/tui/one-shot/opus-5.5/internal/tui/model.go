package tui

import (
	"database/sql"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/example/sqlnav/internal/db"
)

const previewLimit = 10

type screen int

const (
	screenSelect screen = iota
	screenCreate
	screenView
)

type dbLoadedMsg struct {
	conn  *sql.DB
	name  string
	count int
	cols  []string
	rows  [][]string
}

type errMsg struct{ err error }

type Model struct {
	dir    string
	screen screen

	dbs    []string
	cursor int

	input    textinput.Model
	inputErr string

	conn   *sql.DB
	dbName string
	count  int
	cols   []string
	rows   [][]string

	err error
}

func New(dir string) (Model, error) {
	dbs, err := db.List(dir)
	if err != nil {
		return Model{}, err
	}
	ti := textinput.New()
	ti.Placeholder = "mydatabase"
	ti.CharLimit = 64
	ti.Width = 40
	return Model{dir: dir, dbs: dbs, input: ti}, nil
}

func (m Model) Init() tea.Cmd { return nil }

// Close releases the open database connection, if any.
func (m Model) Close() {
	if m.conn != nil {
		m.conn.Close()
	}
}

func load(conn *sql.DB, name string) tea.Cmd {
	return func() tea.Msg {
		count, err := db.Count(conn)
		if err != nil {
			conn.Close()
			return errMsg{err}
		}
		cols, rows, err := db.Preview(conn, previewLimit)
		if err != nil {
			conn.Close()
			return errMsg{err}
		}
		return dbLoadedMsg{conn: conn, name: name, count: count, cols: cols, rows: rows}
	}
}

func openCmd(dir, name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Open(dir, name)
		if err != nil {
			return errMsg{err}
		}
		return load(conn, name)()
	}
}

func createCmd(dir, name string, existing []string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Create(dir, name, existing)
		if err != nil {
			return errMsg{err}
		}
		return load(conn, name)()
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case dbLoadedMsg:
		m.conn, m.dbName = msg.conn, msg.name
		m.count, m.cols, m.rows = msg.count, msg.cols, msg.rows
		m.err = nil
		m.screen = screenView
		return m, nil
	case errMsg:
		m.err = msg.err
		return m, nil
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenCreate:
		return m.updateCreate(msg)
	case screenView:
		return m.updateView(msg)
	}
	return m, nil
}

// The final option in the select list is "create new database".
func (m Model) optionCount() int { return len(m.dbs) + 1 }

func (m Model) startCreate() (tea.Model, tea.Cmd) {
	m.screen = screenCreate
	m.err = nil
	m.inputErr = ""
	m.input.SetValue("")
	return m, m.input.Focus()
}

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < m.optionCount()-1 {
			m.cursor++
		}
	case "n":
		return m.startCreate()
	case "enter":
		if m.cursor == len(m.dbs) {
			return m.startCreate()
		}
		return m, openCmd(m.dir, m.dbs[m.cursor])
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.screen = screenSelect
			m.input.Blur()
			m.err = nil
			return m, nil
		case "enter":
			name := m.input.Value()
			if err := db.ValidateName(name, m.dbs); err != nil {
				m.inputErr = err.Error()
				return m, nil
			}
			m.inputErr = ""
			m.input.Blur()
			return m, createCmd(m.dir, name, m.dbs)
		}
	}
	var cmd tea.Cmd
	prev := m.input.Value()
	m.input, cmd = m.input.Update(msg)
	if m.input.Value() != prev {
		m.inputErr = ""
	}
	return m, cmd
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.Close()
		m.conn = nil
		dbs, err := db.List(m.dir)
		if err != nil {
			m.err = err
		} else {
			m.dbs = dbs
		}
		m.cursor = 0
		for i, n := range m.dbs {
			if n == m.dbName {
				m.cursor = i
			}
		}
		m.screen = screenSelect
	}
	return m, nil
}

package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sqlnav/internal/db"
)

const maxPreviewRows = 10

var (
	titleStyle  = lipgloss.NewStyle().Bold(true)
	headerStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type mode int

const (
	modeSelect mode = iota
	modeCreate
	modeViewer
)

// loadedMsg carries the result of opening or creating a database.
type loadedMsg struct {
	names []string
	conn  *db.DB
	total int
	table db.Table
	err   error
}

type model struct {
	mode   mode
	names  []string
	cursor int
	status string

	input    textinput.Model
	inputErr string

	conn  *db.DB
	total int
	table db.Table
}

// New builds the initial model.
func New() tea.Model {
	names, err := db.List()
	in := textinput.New()
	in.Placeholder = "mydatabase"
	in.CharLimit = 64

	m := model{mode: modeSelect, names: names, input: in}
	if err != nil {
		m.status = err.Error()
	}
	return m
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case loadedMsg:
		if msg.names != nil {
			m.names = msg.names
		}
		if msg.err != nil {
			if m.mode == modeCreate {
				m.inputErr = msg.err.Error()
				return m, nil
			}
			m.status = msg.err.Error()
			m.mode = modeSelect
			return m, nil
		}
		m.conn = msg.conn
		m.total = msg.total
		m.table = msg.table
		for i, n := range m.names {
			if n == m.conn.Name {
				m.cursor = i
			}
		}
		m.mode = modeViewer
		return m, nil

	case tea.KeyMsg:
		switch m.mode {
		case modeSelect:
			return m.updateSelect(msg)
		case modeCreate:
			return m.updateCreate(msg)
		case modeViewer:
			return m.updateViewer(msg)
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.mode {
	case modeSelect:
		return m.viewSelect()
	case modeCreate:
		return m.viewCreate()
	case modeViewer:
		return m.viewViewer()
	}
	return ""
}

func collect(conn *db.DB) loadedMsg {
	total, err := conn.TotalRecords()
	if err != nil {
		conn.Close()
		return loadedMsg{err: err}
	}
	table, err := conn.Records(maxPreviewRows)
	if err != nil {
		conn.Close()
		return loadedMsg{err: err}
	}
	return loadedMsg{conn: conn, total: total, table: table}
}

func openCmd(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Open(name)
		if err != nil {
			return loadedMsg{err: err}
		}
		return collect(conn)
	}
}

func createCmd(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Create(name)
		if err != nil {
			return loadedMsg{err: err}
		}
		msg := collect(conn)
		msg.names, _ = db.List()
		return msg
	}
}

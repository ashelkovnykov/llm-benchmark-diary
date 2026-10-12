package tui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
)

const (
	createLabel    = "+ Create a new database"
	maxPreviewRows = 10
)

type screen int

const (
	screenSelect screen = iota
	screenName
	screenView
)

// Model is the root bubbletea model.
type Model struct {
	screen screen
	width  int
	height int

	dbs    []db.Info
	cursor int

	name    textinput.Model
	nameErr string

	conn    *sql.DB
	dbName  string
	summary *db.Summary

	status string
}

// Run starts the TUI.
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

// New builds the initial model with the database list loaded.
func New() Model {
	ti := textinput.New()
	ti.Placeholder = "mydatabase"
	ti.CharLimit = 64
	ti.Width = 40

	m := Model{
		screen: screenSelect,
		width:  80,
		height: 24,
		name:   ti,
	}
	m.reload()
	return m
}

func (m *Model) reload() {
	list, err := db.List()
	if err != nil {
		m.status = errStyle.Render("could not read ./db: " + err.Error())
		return
	}
	m.dbs = list
	if m.cursor > len(m.dbs) {
		m.cursor = len(m.dbs)
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.name.Width = max(20, m.width-12)
		return m, nil

	case dbLoadedMsg:
		if msg.err != nil {
			m.status = errStyle.Render(msg.err.Error())
			return m, nil
		}
		if m.conn != nil {
			m.conn.Close()
		}
		m.conn, m.dbName, m.summary = msg.conn, msg.name, msg.summary
		m.status = ""
		m.screen = screenView
		return m, nil

	case dbCreatedMsg:
		if msg.err != nil {
			m.nameErr = msg.err.Error()
			return m, nil
		}
		if m.conn != nil {
			m.conn.Close()
		}
		m.reload()
		m.conn, m.dbName, m.summary = msg.conn, msg.name, msg.summary
		m.name.SetValue("")
		m.nameErr = ""
		m.status = ""
		m.screen = screenView
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.screen {
		case screenSelect:
			return m.updateSelect(msg)
		case screenName:
			return m.updateName(msg)
		case screenView:
			return m.updateView(msg)
		}
	}

	return m, nil
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.dbs) {
			m.cursor++
		}
	case "r":
		m.reload()
	case "enter":
		if m.cursor == len(m.dbs) {
			return m.openNameScreen()
		}
		info := m.dbs[m.cursor]
		m.status = subtleStyle.Render("opening " + info.Name + "…")
		return m, loadCmd(info.Name)
	}
	return m, nil
}

func (m Model) openNameScreen() (tea.Model, tea.Cmd) {
	m.screen = screenName
	m.nameErr = ""
	m.name.SetValue("")
	return m, m.name.Focus()
}

func (m Model) updateName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.name.Blur()
		m.nameErr = ""
		m.screen = screenSelect
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.name.Value())
		if err := db.ValidateName(name, m.dbs); err != nil {
			m.nameErr = err.Error()
			return m, nil
		}
		m.name.Blur()
		m.nameErr = ""
		return m, createCmd(name)
	}

	var cmd tea.Cmd
	m.name, cmd = m.name.Update(msg)
	return m, cmd
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "esc", "b":
		if m.conn != nil {
			m.conn.Close()
		}
		m.conn, m.dbName, m.summary = nil, "", nil
		m.screen = screenSelect
		m.reload()
		return m, nil
	case "r":
		if m.dbName != "" {
			return m, loadCmd(m.dbName)
		}
	}
	return m, nil
}

type dbLoadedMsg struct {
	name    string
	conn    *sql.DB
	summary *db.Summary
	err     error
}

func loadCmd(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Open(name)
		if err != nil {
			return dbLoadedMsg{name: name, err: err}
		}
		summary, err := db.Preview(conn, maxPreviewRows)
		if err != nil {
			conn.Close()
			return dbLoadedMsg{name: name, err: err}
		}
		return dbLoadedMsg{name: name, conn: conn, summary: summary}
	}
}

type dbCreatedMsg struct {
	name    string
	conn    *sql.DB
	summary *db.Summary
	err     error
}

func createCmd(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Create(name)
		if err != nil {
			return dbCreatedMsg{name: name, err: err}
		}
		summary, err := db.Preview(conn, maxPreviewRows)
		if err != nil {
			conn.Close()
			return dbCreatedMsg{name: name, err: err}
		}
		return dbCreatedMsg{name: name, conn: conn, summary: summary}
	}
}

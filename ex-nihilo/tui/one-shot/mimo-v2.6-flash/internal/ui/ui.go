// Package ui implements the bubbletea interface.
package ui

import (
	"database/sql"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
)

type screen int

const (
	screenPicker screen = iota
	screenCreate
	screenViewer
)

type model struct {
	screen  screen
	names   []string
	cursor  int
	input   textinput.Model
	message string
	isErr   bool
	dbName  string
	preview *db.Preview
	conn    *sql.DB
	width   int
	height  int
}

// New returns the root bubbletea program model.
func New() tea.Model {
	input := textinput.New()
	input.Placeholder = "e.g. mydb"
	input.Prompt = "> "
	input.CharLimit = 64

	m := &model{
		screen: screenPicker,
		input:  input,
	}
	m.refresh()
	return m
}

func (m *model) refresh() {
	names, err := db.List()
	if err != nil {
		m.setError(err.Error())
		return
	}
	m.names = names
	if m.cursor >= len(m.names) {
		m.cursor = len(m.names) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *model) setError(msg string) {
	m.message = msg
	m.isErr = true
}

func (m *model) setInfo(msg string) {
	m.message = msg
	m.isErr = false
}

func (m *model) clearMessage() { m.message = "" }

func (m *model) Init() tea.Cmd { return textinput.Blink }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case tea.KeyMsg:
		switch m.screen {
		case screenPicker:
			return m.updatePicker(msg)
		case screenCreate:
			return m.updateCreate(msg)
		case screenViewer:
			return m.updateViewer(msg)
		}
	}
	return m, nil
}

func (m *model) updatePicker(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "n":
		m.screen = screenCreate
		m.input.SetValue("")
		m.clearMessage()
		return m, m.input.Focus()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.names)-1 {
			m.cursor++
		}
	case "enter":
		if len(m.names) == 0 {
			m.setError("no databases available - press n to create one")
			return m, nil
		}
		return m.load(m.names[m.cursor])
	}
	return m, nil
}

func (m *model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.screen = screenPicker
		m.clearMessage()
		m.refresh()
		return m, nil
	case "enter":
		return m.create()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *model) create() (tea.Model, tea.Cmd) {
	name := strings.TrimSpace(m.input.Value())
	if err := db.ValidateName(name); err != nil {
		m.setError("invalid name: " + err.Error())
		return m, nil
	}
	if _, err := db.Create(name); err != nil {
		m.setError(err.Error())
		return m, nil
	}
	m.refresh()
	m.cursor = indexOf(m.names, name)
	// A freshly generated database is loaded immediately.
	return m.load(name)
}

func (m *model) updateViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc", "backspace":
		m.closeDB()
		m.screen = screenPicker
		m.clearMessage()
		m.refresh()
	}
	return m, nil
}

func (m *model) load(name string) (tea.Model, tea.Cmd) {
	conn, err := db.Open(name)
	if err != nil {
		m.setError(err.Error())
		return m, nil
	}
	preview, err := db.Preview(conn, db.PreviewLimit)
	if err != nil {
		conn.Close()
		m.setError(err.Error())
		return m, nil
	}
	m.closeDB()
	m.conn = conn
	m.dbName = name
	m.preview = preview
	m.screen = screenViewer
	m.clearMessage()
	return m, nil
}

func (m *model) closeDB() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
	m.preview = nil
	m.dbName = ""
}

func indexOf(items []string, v string) int {
	for i, it := range items {
		if it == v {
			return i
		}
	}
	return 0
}

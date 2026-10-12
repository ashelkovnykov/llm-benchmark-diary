package tui

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"sql-tui/internal/db"
)

var alphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}

		switch m.state {
		case stateBrowse:
			return m.updateBrowse(msg)
		case stateNewName:
			return m.updateNewName(msg)
		case stateView:
			// pressing any key returns to browse
			if key := msg.String(); key == "enter" || key == "esc" {
				return m.goBackToBrowse()
			}
			return m, nil
		}

	case dbLoadedMsg:
		m.state = stateView
		m.dbName = msg.dbName
		m.count = msg.count
		m.rows = msg.rows
		m.err = ""
		return m, nil

	case dbCreatedMsg:
		m.state = stateView
		m.dbName = msg.dbName
		m.count = msg.count
		m.rows = msg.rows
		m.err = ""
		m.textInput.Reset()
		m.inputErr = ""
		return m, nil

	case errMsg:
		m.err = msg.Error()
		return m, nil
	}

	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		selected := m.list.SelectedItem().(item)
		if selected.isCreate {
			m.state = stateNewName
			m.textInput.Focus()
			m.inputErr = ""
			return m, textinput.Blink
		}
		// load existing db
		filePath := filepath.Join(m.dbDir, selected.title+".db")
		return m, loadCmd(filePath)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) updateNewName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.goBackToBrowse()
	case "enter":
		name := strings.TrimSpace(m.textInput.Value())
		if name == "" {
			m.inputErr = "name cannot be empty"
			return m, nil
		}
		if !alphaNumeric.MatchString(name) {
			m.inputErr = "name must contain only alphanumeric characters"
			return m, nil
		}
		// check existing names (case-sensitive)
		for _, f := range m.dbFiles {
			if f == name {
				m.inputErr = "a database with this name already exists"
				return m, nil
			}
		}
		m.inputErr = ""
		filePath := filepath.Join(m.dbDir, name+".db")
		return m, createCmd(name, filePath)
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) goBackToBrowse() (tea.Model, tea.Cmd) {
	m.state = stateBrowse
	m.err = ""
	// refresh file list
	files, err := db.ListDBFiles(m.dbDir)
	if err != nil {
		m.err = err.Error()
	}
	m.dbFiles = files
	items := []list.Item{}
	for _, f := range files {
		items = append(items, item{title: f, desc: "existing database"})
	}
	items = append(items, item{
		title:    "Create new database",
		desc:     "generate a new SQLite database with dummy data",
		isCreate: true,
	})
	m.list.SetItems(items)
	return m, tea.ClearScreen
}

// commands
func loadCmd(filePath string) tea.Cmd {
	return func() tea.Msg {
		count, rows, err := db.LoadDB(filePath)
		if err != nil {
			return errMsg{error: err}
		}
		return dbLoadedMsg{
			dbName: dbName,
			count:  count,
			rows:   rows,
		}
	}
}

func createCmd(name, filePath string) tea.Cmd {
	return func() tea.Msg {
		if err := db.CreateDB(filePath, db.Schema()); err != nil {
			return errMsg{error: err}
		}
		count, rows, err := db.LoadDB(filePath)
		if err != nil {
			return errMsg{error: err}
		}
		return dbCreatedMsg{
			dbName: name,
			count:  count,
			rows:   rows,
		}
	}
}

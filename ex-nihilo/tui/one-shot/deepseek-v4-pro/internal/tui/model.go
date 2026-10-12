package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type state int

const (
	stateBrowse state = iota
	stateNewName
	stateView
)

type Model struct {
	state     state
	dbDir     string
	dbFiles   []string
	list      list.Model
	textInput textinput.Model
	inputErr  string
	dbName    string
	count     int
	rows      [][]string
	err       string
	width     int
	height    int
}

func NewModel(dbDir string, dbFiles []string) Model {
	items := []list.Item{}
	for _, f := range dbFiles {
		items = append(items, item{title: f, desc: "existing database"})
	}
	items = append(items, item{
		title:    "Create new database",
		desc:     "generate a new SQLite database with dummy data",
		isCreate: true,
	})

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Select a database"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))

	ti := textinput.New()
	ti.Placeholder = "Enter a name (alphanumeric only)"
	ti.CharLimit = 50
	ti.Width = 40

	return Model{
		state:     stateBrowse,
		dbDir:     dbDir,
		dbFiles:   dbFiles,
		list:      l,
		textInput: ti,
	}
}

func (m Model) Init() tea.Cmd {
	if m.state == stateNewName {
		return textinput.Blink
	}
	return nil
}

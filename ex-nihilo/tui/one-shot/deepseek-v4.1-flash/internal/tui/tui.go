// Package tui implements the Bubble Tea interface for dbnav.
package tui

import (
	"dbnav/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenChooser screen = iota
	screenName
	screenView
)

// Model is the root Bubble Tea model.
type Model struct {
	screen screen

	chooser chooserModel
	name    nameModel
	viewer  viewerModel

	store *db.Store
	err   error

	width  int
	height int
}

// New builds the root model, loading the list of available databases.
func New() (Model, error) {
	names, err := db.List()
	if err != nil {
		return Model{}, err
	}

	return Model{
		screen:  screenChooser,
		chooser: newChooser(names),
		name:    newName(names),
	}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.chooser.setSize(msg.Width, msg.Height)
		m.viewer.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case openMsg:
		return m.open(msg.name)

	case createMsg:
		m.err = nil
		m.screen = screenName
		return m, m.name.reset()

	case createDBMsg:
		store, err := db.Create(msg.name)
		if err != nil {
			m.err = err
			m.name.err = err.Error()
			return m, nil
		}
		return m.load(store)

	case cancelMsg:
		m.screen = screenChooser
		return m, nil
	}

	return m.updateScreen(msg)
}

func (m Model) open(name string) (tea.Model, tea.Cmd) {
	store, err := db.Open(name)
	if err != nil {
		m.err = err
		return m, nil
	}
	return m.load(store)
}

func (m Model) load(store *db.Store) (tea.Model, tea.Cmd) {
	m.store = store
	m.viewer = newViewer(store)
	m.viewer.setSize(m.width, m.height)
	m.screen = screenView
	return m, nil
}

func (m Model) updateScreen(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.screen {
	case screenChooser:
		m.chooser, cmd = m.chooser.Update(msg)
	case screenName:
		m.name, cmd = m.name.Update(msg)
	case screenView:
		m.viewer, cmd = m.viewer.Update(msg)
	}
	return m, cmd
}

func (m Model) View() string {
	switch m.screen {
	case screenName:
		return m.name.View()
	case screenView:
		return m.viewer.View()
	default:
		if m.err != nil {
			return errorStyle.Render("error: "+m.err.Error()) + "\n\n" + m.chooser.View()
		}
		return m.chooser.View()
	}
}

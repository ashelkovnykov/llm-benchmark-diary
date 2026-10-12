package ui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenSelect screen = iota
	screenCreate
	screenView
)

type Model struct {
	screen  screen
	select_ selectModel
	create  createModel
	view    viewModel
	err     error
}

func NewModel() Model {
	return Model{
		screen:  screenSelect,
		select_: newSelectModel(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.select_.init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m Model) View() string {
	switch m.screen {
	case screenSelect:
		return m.select_.view()
	case screenCreate:
		return m.create.view()
	case screenView:
		return m.view.view()
	}
	return ""
}

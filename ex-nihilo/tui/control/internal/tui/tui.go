package tui

import (
	tea "charm.land/bubbletea/v2"
)

func InitialModel() tea.Model {
	// Step 1: define all of the sub-components
	listModel := newLoadModel()
	exploreModel := modelExplore{}

	// Step 2: wire the sub-components up to each other to establish parent-child
	// relationships
	listModel.parent = &exploreModel
	exploreModel.loadDBMenu = listModel

	// Step 3: explicitly initialize all of the sub-components.
	//
	// This is necessary because some bubbles require explicit initalization to
	// work correctly, but `tea.Model.Init()` is only guaranteed to be called
	// EXACTLY ONCE and only for the model that's input DIRECTLY to
	// `tea.NewProgram()`.
	listModel.Update(listModel.Init()())

	// We begin in list model because we need to import a DB on boot
	return listModel
}

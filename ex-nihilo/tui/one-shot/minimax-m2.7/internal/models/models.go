package models

import "github.com/charmbracelet/bubbletea"

type Screen int

const (
	ScreenDBList Screen = iota
	ScreenNewDB
	ScreenViewer
)

type Model struct {
	Screen      Screen
	DBName      string
	Databases   []string
	SelectedDB  int
	NewDBName   string
	NewDBError  string
	RecordCount int64
	Records     []struct {
		RowID       int64
		Name        string
		Description string
		Value       int64
	}
	Quit bool
}

func InitialModel() tea.Model {
	return Model{
		Screen:     ScreenDBList,
		SelectedDB: 0,
	}
}

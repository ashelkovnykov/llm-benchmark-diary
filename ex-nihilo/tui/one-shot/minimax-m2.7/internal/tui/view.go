package tui

import (
	"dbviewer/internal/db"
	"fmt"

	"github.com/charmbracelet/bubbletea"
)

type dbViewerModel struct {
	dbName      string
	database    interface{}
	recordCount int64
	records     []db.Record
}

func NewDBViewer(dbName string) (tea.Model, error) {
	database, err := db.LoadDatabase(dbName)
	if err != nil {
		return nil, err
	}

	count, err := db.GetRecordCount(database)
	if err != nil {
		database.Close()
		return nil, err
	}

	records, err := db.GetRecords(database, 10)
	if err != nil {
		database.Close()
		return nil, err
	}

	return dbViewerModel{
		dbName:      dbName,
		database:    database,
		recordCount: count,
		records:     records,
	}, nil
}

func (m dbViewerModel) Init() tea.Cmd {
	return nil
}

func (m dbViewerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m dbViewerModel) View() string {
	s := fmt.Sprintf("Database: %s\n\n", m.dbName)
	s += fmt.Sprintf("Total records: %d\n\n", m.recordCount)

	if len(m.records) == 0 {
		s += "No records found.\n"
	} else {
		s += "First 10 records:\n\n"
		s += fmt.Sprintf("%-6s %-20s %-30s %s\n", "rowid", "name", "description", "value")
		s += fmt.Sprintf("%-6s %-20s %-30s %s\n", "-----", "----", "-----------", "-----")
		for _, r := range m.records {
			s += fmt.Sprintf("%-6d %-20s %-30s %d\n", r.RowID, r.Name, r.Description, r.Value)
		}
	}

	s += "\nPress q to quit.\n"
	return s
}

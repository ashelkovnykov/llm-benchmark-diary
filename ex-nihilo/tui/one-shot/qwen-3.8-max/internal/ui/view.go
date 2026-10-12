package ui

import (
	"fmt"
	"strings"

	"dbnav/internal/db"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().Bold(true).Underline(true)
	cellStyle   = lipgloss.NewStyle().Width(14)
)

type viewModel struct {
	dbName  string
	total   int
	records []db.Record
}

func newViewModel(name string, total int, records []db.Record) viewModel {
	return viewModel{dbName: name, total: total, records: records}
}

func (m *Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m viewModel) view() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.dbName)))
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.total))

	// Table header
	b.WriteString(headerStyle.Render(
		fmt.Sprintf("%-6s %-16s %-16s %-20s", "rowid", "name", "value", "created_at"),
	))
	b.WriteString("\n")
	b.WriteString(strings.Repeat("─", 60))
	b.WriteString("\n")

	for _, r := range m.records {
		b.WriteString(fmt.Sprintf("%-6d %-16s %-16s %-20s\n", r.RowID, r.Name, r.Value, r.CreatedAt))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("q quit"))

	return b.String()
}

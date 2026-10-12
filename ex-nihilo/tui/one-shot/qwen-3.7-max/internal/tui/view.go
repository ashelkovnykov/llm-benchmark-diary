package tui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	var b strings.Builder

	switch m.state {
	case stateSelectDB:
		b.WriteString("Select a database to load:\n\n")
		if len(m.dbs) == 0 {
			b.WriteString("  (No existing databases found)\n\n")
		} else {
			for i, db := range m.dbs {
				cursor := "  "
				if m.cursor == i {
					cursor = "> "
				}
				b.WriteString(fmt.Sprintf("%s%s\n", cursor, db))
			}
			b.WriteString("\n")
		}

		cursor := "  "
		if m.cursor == len(m.dbs) {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s[ Create New Database ]\n\n", cursor))

	case stateCreateDB:
		b.WriteString("Create a new database:\n\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		if m.err != nil {
			b.WriteString(fmt.Sprintf("Error: %s\n\n", m.err.Error()))
		}
		b.WriteString("(esc to go back)\n")

	case stateViewDB:
		b.WriteString(fmt.Sprintf("Database: %s\n", m.selectedDB))
		b.WriteString(fmt.Sprintf("Total Records: %d\n\n", m.totalRecords))

		b.WriteString("First 10 Records:\n")
		b.WriteString(fmt.Sprintf("%-6s | %-20s | %-8s | %-20s\n", "rowid", "name", "value", "created_at"))
		b.WriteString(strings.Repeat("-", 60) + "\n")
		for _, r := range m.records {
			b.WriteString(fmt.Sprintf("%-6d | %-20s | %-8.2f | %-20s\n", r.RowID, r.Name, r.Value, r.CreatedAt))
		}
		b.WriteString("\n(q or esc to go back)\n")
	}

	if m.err != nil && m.state == stateSelectDB {
		b.WriteString(fmt.Sprintf("\nError: %s\n", m.err.Error()))
	}

	return b.String()
}

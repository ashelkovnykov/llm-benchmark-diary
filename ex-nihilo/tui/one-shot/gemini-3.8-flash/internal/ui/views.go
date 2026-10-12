package ui

import (
	"fmt"
	"strings"
)

func (m Model) View() string {
	switch m.state {
	case viewSelectDB:
		return m.selectDBView()
	case viewCreateDB:
		return m.createDBView()
	case viewVisualizer:
		return m.visualizerView()
	default:
		return ""
	}
}

func (m Model) selectDBView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("SQL Navigator - Select Database") + "\n\n")

	if len(m.dbList) == 0 {
		b.WriteString(dimmedStyle.Render("No databases found in ./db/") + "\n\n")
	} else {
		for i, name := range m.dbList {
			if i == m.cursor {
				b.WriteString(selectedItemStyle.Render(fmt.Sprintf("> %s", name)) + "\n")
			} else {
				b.WriteString(fmt.Sprintf("  %s\n", name))
			}
		}
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()) + "\n")
	}

	b.WriteString(helpStyle.Render("↑/↓: navigate • enter: open • n: new DB • q: quit"))
	return b.String()
}

func (m Model) createDBView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create New Database") + "\n\n")
	b.WriteString("Enter DB name (alphanumeric, no extension):\n")
	b.WriteString(m.textInput.View() + "\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(m.err.Error()) + "\n\n")
	}

	b.WriteString(helpStyle.Render("enter: confirm • esc: cancel"))
	return b.String()
}

func (m Model) visualizerView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.loadedDB)) + "\n\n")
	b.WriteString(fmt.Sprintf("Total Records: %d\n\n", m.totalRecords))
	b.WriteString(m.table.View() + "\n\n")

	b.WriteString(helpStyle.Render("esc: back to DB list • q: quit"))
	return b.String()
}

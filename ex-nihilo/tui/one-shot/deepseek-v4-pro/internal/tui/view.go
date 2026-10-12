package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	docStyle  = lipgloss.NewStyle().Padding(1, 2)
	errStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	hintStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true)
)

func (m Model) View() string {
	switch m.state {
	case stateBrowse:
		return docStyle.Render(m.list.View() + "\n" + hintStyle.Render("↑/↓ navigate • enter select • q quit"))
	case stateNewName:
		var b strings.Builder
		b.WriteString("New database\n\n")
		b.WriteString(m.textInput.View())
		b.WriteString("\n\n")
		if m.inputErr != "" {
			b.WriteString(errStyle.Render("Error: " + m.inputErr))
		}
		b.WriteString("\n" + hintStyle.Render("enter confirm • esc back"))
		return docStyle.Render(b.String())
	case stateView:
		var b strings.Builder
		b.WriteString(fmt.Sprintf("Database: %s\n", m.dbName))
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.count))
		if m.count > 0 {
			b.WriteString("First records:\n")
			b.WriteString(fmt.Sprintf("%-8s %-15s %s\n", "rowid", "name", "amount"))
			b.WriteString(strings.Repeat("-", 40) + "\n")
			for _, row := range m.rows {
				b.WriteString(fmt.Sprintf("%-8s %-15s %s\n", row[0], row[1], row[2]))
			}
			if m.count > 10 {
				b.WriteString(fmt.Sprintf("\n... and %d more", m.count-10))
			}
		}
		if m.err != "" {
			b.WriteString("\n" + errStyle.Render("Error: "+m.err))
		}
		b.WriteString("\n\n" + hintStyle.Render("press any key to return to database list"))
		return docStyle.Render(b.String())
	default:
		return "unknown state"
	}
}

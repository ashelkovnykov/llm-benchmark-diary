package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sqlnav/internal/db"
)

const maxCellWidth = 28

func (m model) updateViewer(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		if m.conn != nil {
			m.conn.Close()
		}
		m.conn = nil
		m.mode = modeSelect
		return m, nil
	}
	return m, nil
}

func (m model) viewViewer() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Database: "+m.conn.Name) + "\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.total))
	b.WriteString(renderTable(m.table) + "\n\n")
	b.WriteString(mutedStyle.Render("esc: back to database list · q: quit"))
	return b.String()
}

func renderTable(t db.Table) string {
	if len(t.Columns) == 0 {
		return mutedStyle.Render("(no columns)")
	}

	widths := make([]int, len(t.Columns))
	for i, c := range t.Columns {
		widths[i] = lipgloss.Width(c)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if w := lipgloss.Width(cell); i < len(widths) && w > widths[i] {
				widths[i] = w
			}
		}
	}
	for i := range widths {
		if widths[i] > maxCellWidth {
			widths[i] = maxCellWidth
		}
	}

	var b strings.Builder
	header := make([]string, len(t.Columns))
	for i, c := range t.Columns {
		header[i] = pad(truncate(c, widths[i]), widths[i])
	}
	b.WriteString(headerStyle.Render(strings.Join(header, "  ")) + "\n")

	if len(t.Rows) == 0 {
		b.WriteString(mutedStyle.Render("(no rows)"))
		return b.String()
	}
	for _, row := range t.Rows {
		cells := make([]string, len(row))
		for i, cell := range row {
			cells[i] = pad(truncate(cell, widths[i]), widths[i])
		}
		b.WriteString(strings.Join(cells, "  ") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func pad(s string, width int) string {
	if n := width - lipgloss.Width(s); n > 0 {
		return s + strings.Repeat(" ", n)
	}
	return s
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max-1]) + "…"
}

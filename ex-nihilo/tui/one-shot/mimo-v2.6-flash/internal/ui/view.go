package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"sqlnav/internal/db"
)

const maxCell = 32

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81"))
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	infoStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

func (m *model) View() string {
	switch m.screen {
	case screenPicker:
		return m.viewPicker()
	case screenCreate:
		return m.viewCreate()
	case screenViewer:
		return m.viewViewer()
	}
	return ""
}

func (m *model) renderMessage() string {
	if m.message == "" {
		return ""
	}
	style := infoStyle
	if m.isErr {
		style = errStyle
	}
	return "\n" + style.Render(m.message) + "\n"
}

func (m *model) viewPicker() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("sqlnav") + "\n\n")
	b.WriteString(dimStyle.Render("select a database") + "\n\n")

	if len(m.names) == 0 {
		b.WriteString(dimStyle.Render("  (no databases in ./"+db.Dir+"/)") + "\n")
	}
	for i, name := range m.names {
		if i == m.cursor {
			b.WriteString(cursorStyle.Render("  > "+name) + "\n")
		} else {
			b.WriteString("    " + name + "\n")
		}
	}

	b.WriteString(m.renderMessage())
	b.WriteString("\n" + helpStyle.Render("[up/down move] [enter open] [n new db] [q quit]"))
	return b.String()
}

func (m *model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("create a new database") + "\n\n")
	b.WriteString("  " + m.input.View() + "\n")
	b.WriteString(m.renderMessage())
	b.WriteString("\n" + dimStyle.Render("alphanumeric only, e.g. mydb") + "\n")
	b.WriteString("\n" + helpStyle.Render("[enter create] [esc cancel]"))
	return b.String()
}

func (m *model) viewViewer() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("database: "+m.dbName) + "\n")
	b.WriteString(dimStyle.Render(fmt.Sprintf("records: %d", m.preview.Total)) + "\n\n")

	if m.preview.Total == 0 {
		b.WriteString(dimStyle.Render("  (the records table is empty)") + "\n")
	} else {
		b.WriteString(renderTable(m.preview))
		if int64(len(m.preview.Rows)) < m.preview.Total {
			b.WriteString(dimStyle.Render(fmt.Sprintf(
				"  showing %d of %d records", len(m.preview.Rows), m.preview.Total,
			)) + "\n")
		}
	}

	b.WriteString(m.renderMessage())
	b.WriteString("\n" + helpStyle.Render("[esc back] [q quit]"))
	return b.String()
}

func renderTable(p *db.Preview) string {
	if len(p.Columns) == 0 {
		return ""
	}
	widths := make([]int, len(p.Columns))
	for i, c := range p.Columns {
		widths[i] = runeLen(clip(c, maxCell))
	}
	for _, row := range p.Rows {
		for i, cell := range row {
			if i >= len(widths) {
				break
			}
			if w := runeLen(clip(cell, maxCell)); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder
	b.WriteString("  " + headerStyle.Render(formatCells(p.Columns, widths)) + "\n")
	rule := strings.Repeat("-", totalWidth(widths)+2*(len(widths)-1))
	b.WriteString(dimStyle.Render("  "+rule) + "\n")
	for _, row := range p.Rows {
		b.WriteString("  " + formatCells(row, widths) + "\n")
	}
	return b.String()
}

func formatCells(cells []string, widths []int) string {
	parts := make([]string, 0, len(cells))
	for i, cell := range cells {
		if i >= len(widths) {
			break
		}
		cell = clip(cell, widths[i])
		if pad := widths[i] - runeLen(cell); pad > 0 {
			cell += strings.Repeat(" ", pad)
		}
		parts = append(parts, cell)
	}
	return strings.TrimRight(strings.Join(parts, "  "), " ")
}

func clip(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	if w <= 1 {
		return "…"
	}
	return string(r[:w-1]) + "…"
}

func runeLen(s string) int { return len([]rune(s)) }

func totalWidth(widths []int) int {
	total := 0
	for _, w := range widths {
		total += w
	}
	return total
}

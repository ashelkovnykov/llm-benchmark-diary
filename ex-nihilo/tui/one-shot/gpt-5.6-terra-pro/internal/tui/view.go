package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxCellWidth = 28

func (m Model) View() string {
	switch m.screen {
	case screenName:
		return m.viewName()
	case screenView:
		return m.viewData()
	default:
		return m.viewSelect()
	}
}

func (m Model) viewSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("sqlnav"))
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("Select a database to open"))
	b.WriteString("\n\n")

	if len(m.dbs) == 0 {
		b.WriteString(subtleStyle.Render("  (no databases found in ./db/)"))
		b.WriteString("\n")
	}
	for i, info := range m.dbs {
		b.WriteString(m.renderItem(i, info.Name))
		b.WriteString("\n")
	}
	b.WriteString(m.renderItem(len(m.dbs), createLabel))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("up/down move • enter select • r refresh • q quit"))

	if m.status != "" {
		b.WriteString("\n")
		b.WriteString(m.status)
	}

	return appStyle.Render(b.String())
}

func (m Model) renderItem(i int, label string) string {
	if i == m.cursor {
		return cursorStyle.Render("> " + label)
	}
	return "  " + label
}

func (m Model) viewName() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("New database"))
	b.WriteString("\n\n")
	b.WriteString("Name: ")
	b.WriteString(m.name.View())
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render("letters and numbers only"))
	b.WriteString("\n\n")

	if m.nameErr != "" {
		b.WriteString(errStyle.Render("x " + m.nameErr))
		b.WriteString("\n\n")
	}

	b.WriteString(helpStyle.Render("enter create • esc back • ctrl+c quit"))

	return appStyle.Render(b.String())
}

func (m Model) viewData() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render(m.dbName))
	b.WriteString("\n\n")

	if m.summary == nil {
		b.WriteString(subtleStyle.Render("loading…"))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc back • q quit"))
		return appStyle.Render(b.String())
	}

	b.WriteString(labelStyle.Render("records: "))
	b.WriteString(fmt.Sprintf("%d", m.summary.Total))
	b.WriteString("\n\n")

	if len(m.summary.Rows) == 0 {
		b.WriteString(subtleStyle.Render("(no rows)"))
		b.WriteString("\n")
	} else {
		b.WriteString(renderTable(m.summary.Columns, m.summary.Rows))
		b.WriteString("\n")
	}

	if m.summary.Total > len(m.summary.Rows) {
		b.WriteString("\n")
		b.WriteString(subtleStyle.Render(fmt.Sprintf(
			"showing first %d of %d records", len(m.summary.Rows), m.summary.Total)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("r reload • esc back • q quit"))

	return appStyle.Render(b.String())
}

func renderTable(cols []string, rows [][]string) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = lipgloss.Width(c)
	}
	for _, row := range rows {
		for i := range cols {
			if i < len(row) {
				if w := lipgloss.Width(row[i]); w > widths[i] {
					widths[i] = w
				}
			}
		}
	}
	for i := range widths {
		if widths[i] > maxCellWidth {
			widths[i] = maxCellWidth
		}
	}

	render := func(cells []string) string {
		parts := make([]string, len(cols))
		for i := range cols {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			parts[i] = pad(truncate(cell, widths[i]), widths[i])
		}
		return strings.Join(parts, "  ")
	}

	header := render(cols)

	var b strings.Builder
	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n")
	b.WriteString(subtleStyle.Render(strings.Repeat("-", lipgloss.Width(header))))
	b.WriteString("\n")
	for _, row := range rows {
		b.WriteString(render(row))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n")
}

func truncate(s string, width int) string {
	if width <= 0 || lipgloss.Width(s) <= width {
		return s
	}
	if width == 1 {
		return "."
	}

	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > width-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "."
}

func pad(s string, width int) string {
	if gap := width - lipgloss.Width(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

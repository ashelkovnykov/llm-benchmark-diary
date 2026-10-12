package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const maxCellWidth = 24

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).MarginBottom(1)
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	docStyle      = lipgloss.NewStyle().Padding(1, 2)
)

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenSelect:
		m.viewSelect(&b)
	case screenCreate:
		m.viewCreate(&b)
	case screenView:
		m.viewData(&b)
	}
	if m.err != nil {
		b.WriteString("\n" + errStyle.Render("Error: "+m.err.Error()) + "\n")
	}
	return docStyle.Render(b.String())
}

func (m Model) viewSelect(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Select a database") + "\n")
	if len(m.dbs) == 0 {
		b.WriteString(dimStyle.Render(fmt.Sprintf("No databases found in %s", m.dir)) + "\n\n")
	}
	for i := 0; i < m.optionCount(); i++ {
		label := "+ Create new database"
		if i < len(m.dbs) {
			label = m.dbs[i]
		}
		if i == m.cursor {
			b.WriteString(selectedStyle.Render("> "+label) + "\n")
		} else {
			b.WriteString("  " + label + "\n")
		}
	}
	b.WriteString("\n" + dimStyle.Render("↑/↓: move • enter: select • n: new • q: quit") + "\n")
}

func (m Model) viewCreate(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Create a new database") + "\n")
	b.WriteString("Name:\n" + m.input.View() + "\n")
	if m.inputErr != "" {
		b.WriteString(errStyle.Render("Invalid name: "+m.inputErr) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter: create • esc: back • ctrl+c: quit") + "\n")
}

func (m Model) viewData(b *strings.Builder) {
	b.WriteString(titleStyle.Render("Database: "+m.dbName) + "\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n", m.count))
	b.WriteString(dimStyle.Render(fmt.Sprintf("Showing first %d row(s)", len(m.rows))) + "\n\n")
	if len(m.rows) == 0 {
		b.WriteString(dimStyle.Render("(no rows)") + "\n")
	} else {
		b.WriteString(renderTable(m.cols, m.rows))
	}
	b.WriteString("\n" + dimStyle.Render("esc: back • q: quit") + "\n")
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func renderTable(cols []string, rows [][]string) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len([]rune(truncate(c, maxCellWidth)))
	}
	for _, r := range rows {
		for i, v := range r {
			if w := len([]rune(truncate(v, maxCellWidth))); w > widths[i] {
				widths[i] = w
			}
		}
	}

	line := func(vals []string) string {
		cells := make([]string, len(vals))
		for i, v := range vals {
			v = truncate(v, maxCellWidth)
			cells[i] = v + strings.Repeat(" ", widths[i]-len([]rune(v)))
		}
		return strings.Join(cells, " │ ")
	}

	var b strings.Builder
	b.WriteString(headerStyle.Render(line(cols)) + "\n")
	seps := make([]string, len(cols))
	for i, w := range widths {
		seps[i] = strings.Repeat("─", w)
	}
	b.WriteString(dimStyle.Render(strings.Join(seps, "─┼─")) + "\n")
	for _, r := range rows {
		b.WriteString(line(r) + "\n")
	}
	return b.String()
}

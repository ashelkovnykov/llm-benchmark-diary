package tui

import (
	"fmt"
	"strings"

	"dbnav/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

// previewRows is how many rows of data are shown on the viewer screen.
const previewRows = 10

type viewerModel struct {
	store   *db.Store
	headers []string
	rows    [][]string
	count   int
	err     error
	width   int
	height  int
}

func newViewer(store *db.Store) viewerModel {
	v := viewerModel{store: store}

	count, err := store.Count()
	if err != nil {
		v.err = err
		return v
	}
	v.count = count

	headers, rows, err := store.Rows(previewRows)
	if err != nil {
		v.err = err
		return v
	}
	v.headers, v.rows = headers, rows
	return v
}

func (v *viewerModel) setSize(w, h int) {
	v.width, v.height = w, h
}

func (v viewerModel) Update(msg tea.Msg) (viewerModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "esc":
			return v, tea.Quit
		}
	}
	return v, nil
}

func (v viewerModel) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Database: " + v.store.Name))
	b.WriteString("\n\n")

	if v.err != nil {
		b.WriteString(errorStyle.Render("error: " + v.err.Error()))
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Total records: %d\n\n", v.count))

	shown := min(previewRows, len(v.rows))
	b.WriteString(headerStyle.Render(fmt.Sprintf("First %d row(s)", shown)))
	b.WriteString("\n\n")
	b.WriteString(v.table())
	b.WriteString("\n\n")
	b.WriteString(subtleStyle.Render("q: quit"))
	return b.String()
}

func (v viewerModel) table() string {
	if len(v.rows) == 0 {
		return subtleStyle.Render("(no rows)")
	}

	widths := make([]int, len(v.headers))
	for i, h := range v.headers {
		widths[i] = len(h)
	}
	for _, row := range v.rows {
		for i, c := range row {
			if i < len(widths) && len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}

	pad := func(s string, w int) string {
		if len(s) >= w {
			return s
		}
		return s + strings.Repeat(" ", w-len(s))
	}

	var b strings.Builder
	for i, h := range v.headers {
		b.WriteString(headerStyle.Render(pad(h, widths[i])))
		if i < len(v.headers)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	for _, row := range v.rows {
		for i, c := range row {
			b.WriteString(pad(c, widths[i]))
			if i < len(row)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// internal/tui/model.go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"sqltui/internal/db"
)

type screen int

const (
	screenSelect screen = iota
	screenCreate
	screenData
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("69"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
)

type Model struct {
	screen   screen
	dbs      []string
	cursor   int
	input    textinput.Model
	err      string
	loaded   *db.DB
	count    int
	cols     []string
	rows     [][]string
	quitting bool
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "new db name"
	ti.Focus()
	dbs, _ := db.List()
	return Model{screen: screenSelect, dbs: dbs, input: ti}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && (key.String() == "ctrl+c") {
		m.quitting = true
		return m, tea.Quit
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenCreate:
		return m.updateCreate(msg)
	case screenData:
		return m.updateData(msg)
	}
	return m, nil
}

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "esc":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.dbs)-1 {
			m.cursor++
		}
	case "n":
		m.screen = screenCreate
		m.err = ""
		m.input.SetValue("")
		return m, textinput.Blink
	case "enter":
		if len(m.dbs) == 0 {
			return m, nil
		}
		return m.loadDB(m.dbs[m.cursor])
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	switch key.String() {
	case "esc":
		m.screen = screenSelect
		m.err = ""
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())
		if err := db.ValidateName(name); err != nil {
			m.err = err.Error()
			return m, nil
		}
		d, err := db.Create(name)
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		return m.opened(d)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateData(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "q", "esc":
			if m.loaded != nil {
				m.loaded.Close()
			}
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) loadDB(name string) (tea.Model, tea.Cmd) {
	d, err := db.Open(name)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	return m.opened(d)
}

func (m Model) opened(d *db.DB) (tea.Model, tea.Cmd) {
	m.loaded = d
	count, err := d.RecordCount()
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	cols, rows, err := d.FirstRows(10)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}
	m.count, m.cols, m.rows = count, cols, rows
	m.err = ""
	m.screen = screenData
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenSelect:
		return m.viewSelect()
	case screenCreate:
		return m.viewCreate()
	case screenData:
		return m.viewData()
	}
	return ""
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a database") + "\n\n")
	if len(m.dbs) == 0 {
		b.WriteString(dimStyle.Render("  (no databases found)") + "\n")
	}
	for i, name := range m.dbs {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(cursor + name + "\n")
	}
	if m.err != "" {
		b.WriteString("\n" + errStyle.Render(m.err) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter: open • n: new db • q: quit"))
	return b.String()
}

func (m Model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create a new database") + "\n\n")
	b.WriteString(m.input.View() + "\n")
	if m.err != "" {
		b.WriteString("\n" + errStyle.Render(m.err) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter: create • esc: back"))
	return b.String()
}

func (m Model) viewData() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.loaded.Name)) + "\n\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.count))

	header := dimStyle.Render(strings.Join(m.cols, "  |  "))
	b.WriteString(header + "\n")
	for _, row := range m.rows {
		b.WriteString(strings.Join(row, "  |  ") + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("q: quit"))
	return b.String()
}

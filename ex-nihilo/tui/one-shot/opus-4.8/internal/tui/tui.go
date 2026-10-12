package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"dbnav/internal/db"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	screenList screen = iota
	screenCreate
	screenView
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("170"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle     = lipgloss.NewStyle().Faint(true)
	cellStyle     = lipgloss.NewStyle().Padding(0, 1)
)

type Model struct {
	screen screen

	dbs    []string
	cursor int

	input textinput.Model
	err   error

	conn   *sql.DB
	dbName string
	snap   db.Snapshot

	quitting bool
}

func New() Model {
	ti := textinput.New()
	ti.Placeholder = "new database name"
	ti.CharLimit = 64

	m := Model{screen: screenList, input: ti}
	m.refresh()
	return m
}

func (m *Model) refresh() {
	names, err := db.List()
	m.dbs = names
	m.err = err
	if m.cursor > len(m.dbs) {
		m.cursor = len(m.dbs)
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenList:
			return m.updateList(msg)
		case screenCreate:
			return m.updateCreate(msg)
		case screenView:
			return m.updateView(msg)
		}
	}
	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// cursor == len(m.dbs) is the "create new" entry.
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.dbs) {
			m.cursor++
		}
	case "enter":
		if m.cursor == len(m.dbs) {
			m.screen = screenCreate
			m.err = nil
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		}
		return m.load(m.dbs[m.cursor])
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "esc":
		m.screen = screenList
		m.err = nil
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())
		if err := db.ValidateName(name); err != nil {
			m.err = err
			return m, nil
		}
		conn, err := db.Create(name)
		if err != nil {
			m.err = err
			return m, nil
		}
		conn.Close()
		m.refresh()
		return m.load(name)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit
	case "esc", "backspace":
		if m.conn != nil {
			m.conn.Close()
			m.conn = nil
		}
		m.screen = screenList
		m.refresh()
		return m, nil
	}
	return m, nil
}

func (m Model) load(name string) (tea.Model, tea.Cmd) {
	conn, err := db.Open(name)
	if err != nil {
		m.err = err
		return m, nil
	}
	snap, err := db.Load(conn, 10)
	if err != nil {
		conn.Close()
		m.err = err
		return m, nil
	}
	m.conn = conn
	m.dbName = name
	m.snap = snap
	m.err = nil
	m.screen = screenView
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.screen {
	case screenList:
		return m.viewList()
	case screenCreate:
		return m.viewCreate()
	case screenView:
		return m.viewView()
	}
	return ""
}

func (m Model) viewList() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a database") + "\n\n")

	for i, name := range m.dbs {
		m.writeItem(&b, i, name)
	}
	m.writeItem(&b, len(m.dbs), "+ Create new database")

	if m.err != nil {
		b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • enter select • q quit"))
	return b.String()
}

func (m Model) writeItem(b *strings.Builder, i int, label string) {
	cursor := "  "
	if m.cursor == i {
		cursor = "> "
		label = selectedStyle.Render(label)
	}
	fmt.Fprintf(b, "%s%s\n", cursor, label)
}

func (m Model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create a new database") + "\n\n")
	b.WriteString(m.input.View() + "\n")
	if m.err != nil {
		b.WriteString("\n" + errorStyle.Render(m.err.Error()) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render("enter confirm • esc cancel"))
	return b.String()
}

func (m Model) viewView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.dbName)) + "\n\n")
	fmt.Fprintf(&b, "Total records: %d\n\n", m.snap.Count)

	if len(m.snap.Columns) > 0 {
		b.WriteString(renderRow(m.snap.Columns) + "\n")
		b.WriteString(strings.Repeat("─", lipgloss.Width(renderRow(m.snap.Columns))) + "\n")
		for _, r := range m.snap.Rows {
			b.WriteString(renderRow(r.Values) + "\n")
		}
	}

	b.WriteString("\n" + helpStyle.Render("esc back • q quit"))
	return b.String()
}

func renderRow(cells []string) string {
	rendered := make([]string, len(cells))
	for i, c := range cells {
		rendered[i] = cellStyle.Render(c)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
}

package tui

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dbtui/internal/db"
)

const newOptionLabel = "+ New database"

type screen int

const (
	screenSelect screen = iota
	screenCreate
	screenView
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	curStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	dimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	tableStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

// Model is the root bubbletea model.
type Model struct {
	screen screen

	names  []string
	cursor int

	input  string
	errMsg string

	conn   *sql.DB
	loaded string
	snap   *db.Snapshot
}

// New builds the initial model with the current database listing.
func New() (Model, error) {
	names, err := db.List()
	if err != nil {
		return Model{}, err
	}
	return Model{screen: screenSelect, names: names}, nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.screen {
		case screenSelect:
			return m.updateSelect(msg)
		case screenCreate:
			return m.updateCreate(msg)
		case screenView:
			return m.updateView(msg)
		}
	}
	return m, nil
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	last := len(m.names) // index of the "+ New database" entry

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < last {
			m.cursor++
		}

	case "n":
		m.screen = screenCreate
		m.input = ""
		m.errMsg = ""

	case "enter":
		if m.cursor == last {
			m.screen = screenCreate
			m.input = ""
			m.errMsg = ""
			return m, nil
		}
		if len(m.names) == 0 {
			return m, nil
		}
		return m.load(m.names[m.cursor], false)
	}
	return m, nil
}

// load opens the named database (optionally creating it first) and moves to
// the viewer screen.
func (m Model) load(name string, create bool) (tea.Model, tea.Cmd) {
	var (
		conn *sql.DB
		err  error
	)
	if create {
		conn, err = db.Create(name)
	} else {
		conn, err = db.Open(name)
	}
	if err != nil {
		m.errMsg = err.Error()
		return m, nil
	}

	snap, err := db.SnapshotOf(conn, 10)
	if err != nil {
		conn.Close()
		m.errMsg = err.Error()
		return m, nil
	}

	if m.conn != nil {
		m.conn.Close()
	}
	m.conn, m.snap, m.loaded = conn, snap, name
	m.errMsg = ""

	if names, err := db.List(); err == nil {
		m.names = names
	}

	m.screen = screenView
	return m, nil
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.screen = screenSelect
		m.input = ""
		m.errMsg = ""

	case "enter":
		name := strings.TrimSpace(m.input)
		if err := db.ValidateName(name, m.names); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		return m.load(name, true)

	case "backspace":
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}

	default:
		if len(msg.Runes) > 0 {
			m.input += string(msg.Runes)
		}
	}
	return m, nil
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.conn != nil {
			m.conn.Close()
		}
		return m, tea.Quit

	case "esc":
		if m.conn != nil {
			m.conn.Close()
		}
		m.conn, m.snap, m.loaded = nil, nil, ""
		if names, err := db.List(); err == nil {
			m.names = names
		}
		m.cursor = 0
		m.screen = screenSelect
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenSelect:
		return m.viewSelect()
	case screenCreate:
		return m.viewCreate()
	case screenView:
		return m.viewViewer()
	}
	return ""
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a database") + "\n\n")

	if len(m.names) == 0 {
		b.WriteString(dimStyle.Render("  no databases found in ./db") + "\n")
	}
	for i, n := range m.names {
		b.WriteString(m.optionLine(n, i) + "\n")
	}
	b.WriteString(m.optionLine(newOptionLabel, len(m.names)) + "\n")

	b.WriteString("\n" + dimStyle.Render("↑/↓ move • enter select • n new • q quit") + "\n")

	if m.errMsg != "" {
		b.WriteString("\n" + errStyle.Render("error: "+m.errMsg) + "\n")
	}
	return b.String()
}

func (m Model) optionLine(label string, i int) string {
	if m.cursor == i {
		return curStyle.Render("> " + label)
	}
	return "  " + label
}

func (m Model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("New database") + "\n\n")
	b.WriteString("Name: " + m.input + "_\n")

	if m.errMsg != "" {
		b.WriteString("\n" + errStyle.Render(m.errMsg) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter create • esc cancel") + "\n")
	return b.String()
}

func (m Model) viewViewer() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Database: "+m.loaded) + "\n\n")

	if m.snap == nil {
		return b.String()
	}

	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.snap.Count))
	b.WriteString(tableStyle.Render(renderTable(m.snap)) + "\n")
	b.WriteString("\n" + dimStyle.Render("esc back • q quit") + "\n")
	return b.String()
}

func renderTable(s *db.Snapshot) string {
	if len(s.Columns) == 0 {
		return "(empty table)"
	}

	widths := make([]int, len(s.Columns))
	for i, c := range s.Columns {
		widths[i] = len(c)
	}
	for _, row := range s.Rows {
		for i, v := range row {
			if i < len(widths) && len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	total := len(widths) - 1
	for _, w := range widths {
		total += w
	}

	var b strings.Builder
	b.WriteString(joinRow(s.Columns, widths) + "\n")
	b.WriteString(strings.Repeat("-", total) + "\n")
	for _, row := range s.Rows {
		b.WriteString(joinRow(row, widths) + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

func joinRow(vals []string, widths []int) string {
	parts := make([]string, len(widths))
	for i := range widths {
		v := ""
		if i < len(vals) {
			v = vals[i]
		}
		parts[i] = pad(v, widths[i])
	}
	return strings.Join(parts, "  ")
}

func pad(s string, w int) string {
	if len(s) >= w {
		return s
	}
	return s + strings.Repeat(" ", w-len(s))
}

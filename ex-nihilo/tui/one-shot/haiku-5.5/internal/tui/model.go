package tui

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/example/dbexplorer/internal/db"
)

type screen int

const (
	selectScreen screen = iota
	newScreen
	viewScreen
)

type Model struct {
	screen  screen
	dbs     []string
	cursor  int
	input   textinput.Model
	errMsg  string
	status  string
	current string
	conn    *sql.DB
	count   int
	cols    []string
	rows    [][]string
}

func New() (Model, error) {
	dbs, err := db.List()
	if err != nil {
		return Model{}, err
	}

	ti := textinput.New()
	ti.Placeholder = "mydb"
	ti.CharLimit = 64
	ti.Width = 30

	return Model{screen: selectScreen, dbs: dbs, input: ti}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, isKey := msg.(tea.KeyMsg)
	if !isKey {
		var cmd tea.Cmd
		if m.screen == newScreen {
			m.input, cmd = m.input.Update(msg)
		}
		return m, cmd
	}

	if k.Type == tea.KeyCtrlC {
		return m, tea.Quit
	}

	switch m.screen {
	case selectScreen:
		return m.updateSelect(k)
	case newScreen:
		return m.updateNew(k)
	default:
		return m.updateView(k)
	}
}

func (m Model) updateSelect(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := len(m.dbs) + 1

	switch k.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < options-1 {
			m.cursor++
		}
	case "q":
		return m, tea.Quit
	case "enter":
		if m.cursor == len(m.dbs) {
			m.screen = newScreen
			m.errMsg = ""
			m.input.SetValue("")
			return m, m.input.Focus()
		}
		return m.load(m.dbs[m.cursor]), nil
	}
	return m, nil
}

func (m Model) updateNew(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.Type {
	case tea.KeyEsc:
		m.screen = selectScreen
		m.input.Blur()
		return m, nil
	case tea.KeyEnter:
		name := m.input.Value()
		if err := db.Create(name); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		m.input.Blur()
		if dbs, err := db.List(); err == nil {
			m.dbs = dbs
		}
		return m.load(name), nil
	}

	var cmd tea.Cmd
	m.errMsg = ""
	m.input, cmd = m.input.Update(k)
	return m, cmd
}

func (m Model) updateView(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch k.String() {
	case "esc", "backspace":
		m.closeConn()
		m.screen = selectScreen
		m.status = ""
	case "q":
		m.closeConn()
		return m, tea.Quit
	}
	return m, nil
}

// load opens a database and switches to the view screen. On failure it
// stays on the current screen and records the error.
func (m Model) load(name string) Model {
	conn, err := db.Open(name)
	if err != nil {
		m.errMsg = err.Error()
		m.status = err.Error()
		return m
	}

	count, err := db.Count(conn)
	if err != nil {
		conn.Close()
		m.errMsg = err.Error()
		m.status = err.Error()
		return m
	}

	cols, rows, err := db.Preview(conn, db.PreviewRows)
	if err != nil {
		conn.Close()
		m.errMsg = err.Error()
		m.status = err.Error()
		return m
	}

	m.closeConn()
	m.conn = conn
	m.current = name
	m.count = count
	m.cols = cols
	m.rows = rows
	m.status = ""
	m.errMsg = ""
	m.screen = viewScreen
	return m
}

func (m *Model) closeConn() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
}

func (m Model) View() string {
	var b strings.Builder

	switch m.screen {
	case selectScreen:
		b.WriteString("Select a database\n\n")
		for i, name := range m.dbs {
			b.WriteString(cursorLine(i == m.cursor, name))
		}
		b.WriteString(cursorLine(m.cursor == len(m.dbs), "[+] Create new database"))
		if len(m.dbs) == 0 {
			b.WriteString("\nNo databases found in ./db/ yet.\n")
		}
		if m.status != "" {
			fmt.Fprintf(&b, "\nError: %s\n", m.status)
		}
		b.WriteString("\n↑/↓ move • enter select • q quit\n")

	case newScreen:
		b.WriteString("New database name:\n\n")
		b.WriteString(m.input.View() + "\n")
		if m.errMsg != "" {
			fmt.Fprintf(&b, "\nError: %s\n", m.errMsg)
		}
		b.WriteString("\nenter create • esc back\n")

	case viewScreen:
		fmt.Fprintf(&b, "Database: %s\n", m.current)
		fmt.Fprintf(&b, "Total records: %d\n\n", m.count)
		b.WriteString(renderTable(m.cols, m.rows))
		b.WriteString("\nesc back • q quit\n")
	}

	return b.String()
}

func cursorLine(selected bool, text string) string {
	if selected {
		return fmt.Sprintf("> %s\n", text)
	}
	return fmt.Sprintf("  %s\n", text)
}

func renderTable(cols []string, rows [][]string) string {
	if len(cols) == 0 {
		return "(no columns)\n"
	}

	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	for _, r := range rows {
		for i, v := range r {
			if i < len(widths) && len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	var b strings.Builder
	writeLine := func(vals []string) {
		for i, v := range vals {
			fmt.Fprintf(&b, "%-*s  ", widths[i], v)
		}
		b.WriteString("\n")
	}

	writeLine(cols)
	for i := range cols {
		b.WriteString(strings.Repeat("-", widths[i]) + "  ")
	}
	b.WriteString("\n")
	for _, r := range rows {
		writeLine(r)
	}
	if len(rows) == 0 {
		b.WriteString("(no rows)\n")
	}
	return b.String()
}

var _ = sort.Strings

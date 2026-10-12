package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"dbtui/internal/db"
)

type screen int

const (
	screenSelect screen = iota
	screenCreate
	screenView
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	errStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headerStyle = lipgloss.NewStyle().Bold(true).Underline(true)
)

const createLabel = "+ Create new database"

type Model struct {
	dir    string
	screen screen

	// select screen
	names  []string
	cursor int

	// create screen
	input  textinput.Model
	errMsg string

	// view screen
	conn   *sql.DB
	dbName string
	snap   db.Snapshot
}

func New(dir string) Model {
	ti := textinput.New()
	ti.Placeholder = "name"
	ti.CharLimit = 64
	ti.Prompt = "> "

	m := Model{dir: dir, input: ti}
	m.refreshList()
	return m
}

func (m *Model) refreshList() {
	names, err := db.List(m.dir)
	if err != nil {
		m.errMsg = err.Error()
	}
	m.names = names
	m.cursor = 0
}

func (m *Model) closeConn() {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	m.closeConn()
	return m, tea.Quit
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "ctrl+c" {
		return m.quit()
	}
	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenCreate:
		return m.updateCreate(msg)
	case screenView:
		return m.updateView(msg)
	}
	return m, nil
}

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	k, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	total := len(m.names) + 1 // + create option
	switch k.String() {
	case "q":
		return m.quit()
	case "up", "k":
		m.cursor = (m.cursor - 1 + total) % total
	case "down", "j":
		m.cursor = (m.cursor + 1) % total
	case "n":
		return m.startCreate()
	case "enter":
		if m.cursor == len(m.names) {
			return m.startCreate()
		}
		name := m.names[m.cursor]
		conn, err := db.Open(m.dir, name)
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		return m.enterView(conn, name)
	}
	return m, nil
}

func (m Model) startCreate() (tea.Model, tea.Cmd) {
	m.screen = screenCreate
	m.errMsg = ""
	m.input.SetValue("")
	m.input.Focus()
	return m, textinput.Blink
}

func (m Model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "esc":
			m.screen = screenSelect
			m.errMsg = ""
			m.input.Blur()
			m.refreshList()
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.input.Value())
			if err := db.ValidateName(m.dir, name); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			conn, err := db.Create(m.dir, name)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.input.Blur()
			return m.enterView(conn, name)
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) enterView(conn *sql.DB, name string) (tea.Model, tea.Cmd) {
	snap, err := db.Load(conn)
	if err != nil {
		conn.Close()
		m.screen = screenSelect
		m.errMsg = err.Error()
		m.refreshList()
		return m, nil
	}
	m.closeConn()
	m.conn, m.dbName, m.snap = conn, name, snap
	m.screen = screenView
	m.errMsg = ""
	return m, nil
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "q":
			return m.quit()
		case "esc", "backspace":
			m.closeConn()
			m.screen = screenSelect
			m.errMsg = ""
			m.refreshList()
		}
	}
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder
	switch m.screen {
	case screenSelect:
		b.WriteString(titleStyle.Render("Select a database") + "\n\n")
		if len(m.names) == 0 {
			b.WriteString(helpStyle.Render("No databases found in "+m.dir) + "\n\n")
		}
		items := append(append([]string{}, m.names...), createLabel)
		for i, it := range items {
			if i == m.cursor {
				b.WriteString(cursorStyle.Render("▸ "+it) + "\n")
			} else {
				b.WriteString("  " + it + "\n")
			}
		}
		b.WriteString("\n" + helpStyle.Render("↑/↓ move • enter select • n new • q quit"))
	case screenCreate:
		b.WriteString(titleStyle.Render("Create new database") + "\n\n")
		b.WriteString("Enter a name (letters and numbers only):\n\n")
		b.WriteString(m.input.View() + "\n")
		b.WriteString("\n" + helpStyle.Render("enter confirm • esc back"))
	case screenView:
		b.WriteString(titleStyle.Render("Database: "+m.dbName) + "\n\n")
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.snap.Total))
		b.WriteString(renderTable(m.snap.Columns, m.snap.Rows))
		b.WriteString("\n" + helpStyle.Render("esc back • q quit"))
	}
	if m.errMsg != "" {
		b.WriteString("\n\n" + errStyle.Render("Error: "+m.errMsg))
	}
	return b.String() + "\n"
}

func renderTable(cols []string, rows [][]string) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = lipgloss.Width(c)
	}
	for _, r := range rows {
		for i, v := range r {
			if w := lipgloss.Width(v); w > widths[i] {
				widths[i] = w
			}
		}
	}
	line := func(cells []string, style *lipgloss.Style) string {
		parts := make([]string, len(cells))
		for i, c := range cells {
			padded := c + strings.Repeat(" ", widths[i]-lipgloss.Width(c))
			if style != nil {
				padded = style.Render(padded)
			}
			parts[i] = padded
		}
		return strings.Join(parts, "  ") + "\n"
	}
	var b strings.Builder
	b.WriteString(line(cols, &headerStyle))
	for _, r := range rows {
		b.WriteString(line(r, nil))
	}
	return b.String()
}

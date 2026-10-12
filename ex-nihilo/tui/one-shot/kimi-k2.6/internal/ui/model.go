package ui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"sqlnav/internal/db"
)

type screen int

const (
	selectScreen screen = iota
	createScreen
	viewScreen
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA"))
	itemStyle  = lipgloss.NewStyle().PaddingLeft(2)
	selStyle   = lipgloss.NewStyle().PaddingLeft(0).Foreground(lipgloss.Color("#00FFFF"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000"))
)

type Model struct {
	screen      screen
	dbManager   *db.Manager
	dbs         []string
	cursor      int
	input       textinput.Model
	errMsg      string
	activeDB    *sql.DB
	dbName      string
	recordCount int
	records     []db.Record
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "database name"
	ti.CharLimit = 30
	ti.Width = 30

	mgr := db.NewManager("./db")
	dbs, _ := mgr.ListDBs()

	return Model{
		screen:    selectScreen,
		dbManager: mgr,
		dbs:       dbs,
		input:     ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m.close(), tea.Quit
		}

		switch m.screen {
		case selectScreen:
			return m.updateSelect(msg)
		case createScreen:
			return m.updateCreate(msg)
		case viewScreen:
			return m.updateView(msg)
		}

	case dbLoadedMsg:
		if msg.err != nil {
			m.errMsg = msg.err.Error()
			return m, nil
		}
		m.closeDB()
		m.activeDB = msg.db
		m.dbName = msg.name
		m.recordCount = msg.count
		m.records = msg.records
		m.screen = viewScreen
		m.errMsg = ""
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) closeDB() {
	if m.activeDB != nil {
		m.activeDB.Close()
		m.activeDB = nil
	}
}

func (m Model) close() Model {
	m.closeDB()
	return m
}

func (m Model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	items := m.menuItems()
	switch msg.String() {
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(items)-1 {
			m.cursor++
		}
	case "enter":
		choice := items[m.cursor]
		if choice == createLabel() {
			m.screen = createScreen
			m.input.SetValue("")
			m.errMsg = ""
			return m, m.input.Focus()
		}
		return m, openDBCmd(m.dbManager, choice)
	}
	return m, nil
}

func (m Model) updateCreate(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.screen = selectScreen
		m.errMsg = ""
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())
		if err := m.dbManager.ValidateName(name); err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		return m, createAndOpenDBCmd(m.dbManager, name)
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.closeDB()
		m.dbs, _ = m.dbManager.ListDBs()
		m.cursor = 0
		m.screen = selectScreen
		m.errMsg = ""
		return m, nil
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case selectScreen:
		return m.viewSelect()
	case createScreen:
		return m.viewCreate()
	case viewScreen:
		return m.viewView()
	}
	return ""
}

func createLabel() string {
	return "< create new database >"
}

func (m Model) menuItems() []string {
	items := make([]string, len(m.dbs))
	copy(items, m.dbs)
	return append(items, createLabel())
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Databases") + "\n\n")
	items := m.menuItems()
	for i, item := range items {
		if i == m.cursor {
			b.WriteString(selStyle.Render("> "+item) + "\n")
		} else {
			b.WriteString(itemStyle.Render(item) + "\n")
		}
	}
	b.WriteString("\n(j/k or ↑/↓ to move, enter to select, ctrl+c to quit)\n")
	if m.errMsg != "" {
		b.WriteString("\n" + errStyle.Render("Error: "+m.errMsg) + "\n")
	}
	return b.String()
}

func (m Model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create New Database") + "\n\n")
	b.WriteString("Enter a name (alphanumeric only):\n\n")
	b.WriteString(m.input.View() + "\n\n")
	if m.errMsg != "" {
		b.WriteString(errStyle.Render(m.errMsg) + "\n\n")
	}
	b.WriteString("(enter to confirm, esc to go back, ctrl+c to quit)\n")
	return b.String()
}

func (m Model) viewView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.dbName)) + "\n\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.recordCount))
	b.WriteString(fmt.Sprintf("%-8s %-15s %-15s\n", "rowid", "name", "value"))
	b.WriteString(strings.Repeat("-", 40) + "\n")
	for _, r := range m.records {
		b.WriteString(fmt.Sprintf("%-8d %-15s %-15s\n", r.RowID, r.Name, r.Value))
	}
	if len(m.records) == 0 {
		b.WriteString("(no records)\n")
	}
	b.WriteString("\n(esc to go back, ctrl+c to quit)\n")
	return b.String()
}

type dbLoadedMsg struct {
	name    string
	db      *sql.DB
	count   int
	records []db.Record
	err     error
}

func openDBCmd(mgr *db.Manager, name string) tea.Cmd {
	return func() tea.Msg {
		d, err := mgr.OpenDB(name)
		if err != nil {
			return dbLoadedMsg{err: err}
		}
		count, records, err := fetchData(d)
		return dbLoadedMsg{name: name, db: d, count: count, records: records, err: err}
	}
}

func createAndOpenDBCmd(mgr *db.Manager, name string) tea.Cmd {
	return func() tea.Msg {
		d, err := mgr.CreateDB(name)
		if err != nil {
			return dbLoadedMsg{err: err}
		}
		count, records, err := fetchData(d)
		return dbLoadedMsg{name: name, db: d, count: count, records: records, err: err}
	}
}

func fetchData(d *sql.DB) (int, []db.Record, error) {
	count, err := db.GetStats(d)
	if err != nil {
		return 0, nil, err
	}
	records, err := db.GetRecords(d, 10)
	return count, records, err
}

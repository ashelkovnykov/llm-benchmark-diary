package ui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/example/sqlnav/internal/db"
)

const (
	previewRows = 10
	newDBLabel  = "+ Create new database"
)

type screen int

const (
	screenSelect screen = iota
	screenName
	screenData
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	headStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
)

type item struct {
	name  string
	isNew bool
}

func (i item) Title() string {
	if i.isNew {
		return newDBLabel
	}
	return i.name
}

func (i item) Description() string {
	if i.isNew {
		return "generate a new SQLite database"
	}
	return db.Path(i.name)
}

func (i item) FilterValue() string { return i.Title() }

// Model is the root Bubble Tea model.
type Model struct {
	screen screen

	list  list.Model
	input textinput.Model

	dbName string
	conn   *sql.DB
	table  *db.Table

	err    string
	width  int
	height int
}

// New builds the initial model with the database selection screen.
func New() (Model, error) {
	items, err := loadItems()
	if err != nil {
		return Model{}, err
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a database"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)

	ti := textinput.New()
	ti.Placeholder = "database name"
	ti.CharLimit = 64

	return Model{
		screen: screenSelect,
		list:   l,
		input:  ti,
	}, nil
}

func loadItems() ([]list.Item, error) {
	names, err := db.List()
	if err != nil {
		return nil, err
	}

	items := make([]list.Item, 0, len(names)+1)
	for _, n := range names {
		items = append(items, item{name: n})
	}
	items = append(items, item{isNew: true})
	return items, nil
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if size, ok := msg.(tea.WindowSizeMsg); ok {
		m.width, m.height = size.Width, size.Height
		m.list.SetSize(size.Width, max(size.Height-2, 1))
		return m, nil
	}

	if key, ok := msg.(tea.KeyMsg); ok {
		switch m.screen {
		case screenSelect:
			return m.updateSelect(key)
		case screenName:
			return m.updateName(key)
		case screenData:
			return m.updateData(key)
		}
	}

	switch m.screen {
	case screenSelect:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	case screenName:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) updateSelect(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Let the list own all keys while its filter is active.
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(key)
		return m, cmd
	}

	switch key.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "enter":
		selected, ok := m.list.SelectedItem().(item)
		if !ok {
			return m, nil
		}
		if selected.isNew {
			m.err = ""
			m.screen = screenName
			m.input.SetValue("")
			m.input.Focus()
			return m, textinput.Blink
		}
		return m.loadDB(selected.name)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(key)
	return m, cmd
}

func (m Model) updateName(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.err = ""
		m.input.Blur()
		m.screen = screenSelect
		return m, nil
	case "enter":
		name := strings.TrimSpace(m.input.Value())

		existing, err := db.List()
		if err != nil {
			m.err = err.Error()
			return m, nil
		}
		if err := db.ValidateName(name, existing); err != nil {
			m.err = err.Error()
			return m, nil
		}
		if err := db.Create(name); err != nil {
			m.err = err.Error()
			return m, nil
		}

		m.input.Blur()
		return m.loadDB(name)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(key)
	return m, cmd
}

func (m Model) updateData(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		return m.closeDB()
	}
	return m, nil
}

func (m Model) loadDB(name string) (tea.Model, tea.Cmd) {
	conn, err := db.Open(name)
	if err != nil {
		m.err = err.Error()
		return m, nil
	}

	table, err := db.Load(conn, previewRows)
	if err != nil {
		conn.Close()
		m.err = err.Error()
		return m, nil
	}

	m.conn = conn
	m.dbName = name
	m.table = table
	m.err = ""
	m.screen = screenData
	return m, nil
}

func (m Model) closeDB() (tea.Model, tea.Cmd) {
	if m.conn != nil {
		m.conn.Close()
		m.conn = nil
	}
	m.table = nil
	m.dbName = ""
	m.err = ""

	items, err := loadItems()
	if err != nil {
		m.err = err.Error()
	} else {
		m.list.SetItems(items)
		m.list.ResetSelected()
		m.list.ResetFilter()
	}

	m.screen = screenSelect
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	switch m.screen {
	case screenSelect:
		b.WriteString(m.list.View())

	case screenName:
		b.WriteString(titleStyle.Render("Create a new database"))
		b.WriteString("\n")
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter: create • esc: back • ctrl+c: quit"))

	case screenData:
		b.WriteString(titleStyle.Render("Database: " + m.dbName))
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.table.Total))
		b.WriteString(m.renderTable())
		b.WriteString("\n")
		b.WriteString(helpStyle.Render("esc: back to database list • ctrl+c: quit"))
	}

	if m.err != "" {
		b.WriteString("\n\n")
		b.WriteString(errStyle.Render("error: " + m.err))
	}
	return b.String()
}

func (m Model) renderTable() string {
	t := m.table
	if t == nil || len(t.Columns) == 0 {
		return ""
	}

	widths := make([]int, len(t.Columns))
	for i, c := range t.Columns {
		widths[i] = len(c)
	}
	for _, row := range t.Rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
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
	for i, c := range t.Columns {
		b.WriteString(headStyle.Render(pad(c, widths[i])))
		if i < len(t.Columns)-1 {
			b.WriteString("  ")
		}
	}
	b.WriteString("\n")

	for _, row := range t.Rows {
		for i, cell := range row {
			b.WriteString(pad(cell, widths[i]))
			if i < len(row)-1 {
				b.WriteString("  ")
			}
		}
		b.WriteString("\n")
	}
	if len(t.Rows) == 0 {
		b.WriteString("(no records)\n")
	}
	return b.String()
}

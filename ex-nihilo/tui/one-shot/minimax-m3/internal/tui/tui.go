package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tuidb/internal/db"
)

type state int

const (
	stateSelect state = iota
	stateCreate
	stateView
)

const createNewLabel = "+ Create new database"

type item struct {
	name string
	desc string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return i.desc }
func (i item) FilterValue() string { return i.name }

type model struct {
	state  state
	width  int
	height int
	errMsg string

	dbList list.Model
	dbs    []string

	nameInput textinput.Model

	dbName  string
	count   int
	records []db.Record
}

func initialModel() (model, error) {
	dbs, err := db.ListDatabases()
	if err != nil {
		return model{}, err
	}

	items := make([]list.Item, 0, len(dbs)+1)
	for _, n := range dbs {
		items = append(items, item{name: n, desc: fmt.Sprintf("./db/%s.db", n)})
	}
	items = append(items, item{name: createNewLabel, desc: "Create a new SQLite database"})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a database"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)

	ti := textinput.New()
	ti.Placeholder = "my_database"
	ti.CharLimit = 64
	ti.Prompt = "Name: "
	ti.Focus()

	return model{
		state:     stateSelect,
		dbList:    l,
		dbs:       dbs,
		nameInput: ti,
	}, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.dbList.SetSize(msg.Width, msg.Height-4)
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	switch m.state {
	case stateSelect:
		return m.updateSelect(msg)
	case stateCreate:
		return m.updateCreate(msg)
	case stateView:
		return m.updateView(msg)
	}
	return m, nil
}

func (m model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && key.String() == "enter" {
		chosen, ok := m.dbList.SelectedItem().(item)
		if !ok {
			return m, nil
		}
		if chosen.name == createNewLabel {
			m.state = stateCreate
			m.errMsg = ""
			m.nameInput.SetValue("")
			m.nameInput.Focus()
			return m, nil
		}
		m.dbName = chosen.name
		m.state = stateView
		return m, loadDBCmd(chosen.name)
	}
	var cmd tea.Cmd
	m.dbList, cmd = m.dbList.Update(msg)
	return m, cmd
}

type dbLoadedMsg struct {
	dbName  string
	count   int
	records []db.Record
	err     error
}

func loadDBCmd(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.OpenDatabase(name)
		if err != nil {
			return dbLoadedMsg{err: err}
		}
		defer conn.Close()
		count, err := db.CountRecords(conn)
		if err != nil {
			return dbLoadedMsg{err: err}
		}
		records, err := db.GetRecords(conn, 10)
		if err != nil {
			return dbLoadedMsg{err: err}
		}
		return dbLoadedMsg{dbName: name, count: count, records: records}
	}
}

func (m model) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			m.state = stateSelect
			m.errMsg = ""
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.nameInput.Value())
			if err := db.ValidateName(name, m.dbs); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			if _, err := db.CreateDatabase(name); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			m.dbName = name
			m.state = stateView
			return m, loadDBCmd(name)
		}
	}
	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)
	return m, cmd
}

func (m model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case dbLoadedMsg:
		if msg.err != nil {
			m.errMsg = "Failed to load database: " + msg.err.Error()
			m.state = stateSelect
			dbs, lerr := db.ListDatabases()
			if lerr == nil {
				m.dbs = dbs
				items := make([]list.Item, 0, len(dbs)+1)
				for _, n := range dbs {
					items = append(items, item{name: n, desc: fmt.Sprintf("./db/%s.db", n)})
				}
				items = append(items, item{name: createNewLabel, desc: "Create a new SQLite database"})
				m.dbList.SetItems(items)
			}
			return m, nil
		}
		m.count = msg.count
		m.records = msg.records
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateSelect:
		return m.viewSelect()
	case stateCreate:
		return m.viewCreate()
	case stateView:
		return m.viewView()
	}
	return ""
}

func (m model) viewSelect() string {
	var b strings.Builder
	if m.errMsg != "" {
		b.WriteString(errorStyle.Render(m.errMsg))
		b.WriteString("\n\n")
	}
	b.WriteString(m.dbList.View())
	return b.String()
}

func (m model) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Create new database"))
	b.WriteString("\n\n")
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("alphanumeric only • cannot be empty • must be unique"))
	if m.errMsg != "" {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render("Error: " + m.errMsg))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("enter to confirm • esc to go back"))
	return b.String()
}

func (m model) viewView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(fmt.Sprintf("Database: %s", m.dbName)))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "Total records: %d\n\n", m.count)
	b.WriteString("First 10 records:\n")
	if len(m.records) == 0 {
		b.WriteString(dimStyle.Render("(no records)\n"))
	} else {
		for _, r := range m.records {
			fmt.Fprintf(&b, "  #%-3d %-15s | %-30s | %8.2f | %s\n",
				r.RowID,
				truncate(r.Title, 15),
				truncate(r.Description, 30),
				r.Amount,
				r.CreatedAt,
			)
		}
	}
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("q / esc to quit"))
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-3] + "..."
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63")).MarginBottom(1)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	dimStyle   = lipgloss.NewStyle().Faint(true)
)

// Run starts the TUI program.
func Run() error {
	m, err := initialModel()
	if err != nil {
		return err
	}
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

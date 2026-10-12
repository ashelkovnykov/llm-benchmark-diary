package tui

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/example/sql-tui/internal/config"
	"github.com/example/sql-tui/internal/db"
)

type state int

const (
	stateSelect state = iota
	stateName
	stateLoaded
)

type Model struct {
	cfg        config.Config
	state      state
	dbs        []db.Info
	selected   int
	nameInput  []rune
	err        string
	conn       *sql.DB
	loadedName string
	count      int64
	preview    db.Preview
}

func New(cfg config.Config) Model {
	return Model{
		cfg:   cfg,
		state: stateSelect,
	}
}

type listMsg struct {
	infos []db.Info
	err   error
}

type openMsg struct {
	conn *sql.DB
	name string
	err  error
}

type createMsg struct {
	conn *sql.DB
	name string
	err  error
}

type previewMsg struct {
	count   int64
	preview db.Preview
	err     error
}

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	descStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Bold(true)
	previewStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1)
)

func (m Model) Init() tea.Cmd {
	return m.listDatabases()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, nil

	case listMsg:
		m.dbs = msg.infos

		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.err = ""
		}

		m.clampSelected()
		return m, nil

	case openMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}

		m.conn = msg.conn
		m.loadedName = msg.name
		m.state = stateLoaded
		m.err = ""

		return m, m.refreshPreview()

	case createMsg:
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}

		m.conn = msg.conn
		m.loadedName = msg.name
		m.state = stateLoaded
		m.nameInput = nil
		m.err = ""

		return m, m.refreshPreview()

	case previewMsg:
		if m.state != stateLoaded {
			return m, nil
		}

		if msg.err != nil {
			m.err = msg.err.Error()
		} else {
			m.err = ""
		}

		m.count = msg.count
		m.preview = msg.preview

		return m, nil

	case tea.KeyMsg:
		return m.handleKeys(msg)

	default:
		return m, nil
	}
}

func (m Model) View() string {
	switch m.state {
	case stateSelect:
		return m.viewSelect()
	case stateName:
		return m.viewName()
	case stateLoaded:
		return m.viewLoaded()
	default:
		return ""
	}
}

func (m *Model) clampSelected() {
	if len(m.dbs) == 0 {
		m.selected = 0
		return
	}

	if m.selected < 0 {
		m.selected = 0
	}

	if m.selected >= len(m.dbs) {
		m.selected = len(m.dbs) - 1
	}
}

func (m Model) listDatabases() tea.Cmd {
	return func() tea.Msg {
		infos, err := db.List(m.cfg.DBDir)
		return listMsg{
			infos: infos,
			err:   err,
		}
	}
}

func (m Model) openDatabase(path, name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Open(path)
		return openMsg{
			conn: conn,
			name: name,
			err:  err,
		}
	}
}

func (m Model) createDatabase(name string) tea.Cmd {
	return func() tea.Msg {
		conn, err := db.Create(m.cfg.DBDir, name)
		return createMsg{
			conn: conn,
			name: name,
			err:  err,
		}
	}
}

func (m Model) refreshPreview() tea.Cmd {
	conn := m.conn

	if conn == nil {
		return func() tea.Msg {
			return previewMsg{err: fmt.Errorf("no database loaded")}
		}
	}

	return func() tea.Msg {
		count, err := db.CountRecords(conn)
		if err != nil {
			return previewMsg{err: err}
		}

		preview, err := db.QueryPreview(conn, 10)
		if err != nil {
			return previewMsg{count: count, err: err}
		}

		return previewMsg{count: count, preview: preview}
	}
}

func (m Model) handleKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC {
		return m.quit()
	}

	switch m.state {
	case stateSelect:
		return m.handleSelectKeys(msg)
	case stateName:
		return m.handleNameKeys(msg)
	case stateLoaded:
		return m.handleLoadedKeys(msg)
	default:
		return m, nil
	}
}

func (m Model) handleSelectKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.clampSelected()

	switch msg.String() {
	case "q", "esc":
		return m.quit()

	case "n":
		m.state = stateName
		m.nameInput = nil
		m.err = ""
		return m, nil

	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

		return m, nil

	case "down", "j":
		if m.selected < len(m.dbs)-1 {
			m.selected++
		}

		return m, nil

	case "enter":
		if len(m.dbs) == 0 {
			m.err = "no databases found; press n to create one"
			return m, nil
		}

		selected := m.dbs[m.selected]
		return m, m.openDatabase(selected.Path, selected.Name)

	default:
		return m, nil
	}
}

func (m Model) handleNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := string(m.nameInput)

		if err := db.ValidateName(m.cfg.DBDir, name); err != nil {
			m.err = err.Error()
			return m, nil
		}

		return m, m.createDatabase(name)

	case tea.KeyBackspace:
		if len(m.nameInput) > 0 {
			m.nameInput = m.nameInput[:len(m.nameInput)-1]
		}

		return m, nil

	case tea.KeyEscape:
		m.state = stateSelect
		m.nameInput = nil
		m.err = ""
		return m, nil

	case tea.KeySpace:
		m.nameInput = append(m.nameInput, ' ')
		return m, nil

	default:
		if len(msg.Runes) > 0 {
			m.nameInput = append(m.nameInput, msg.Runes...)
			return m, nil
		}

		return m, nil
	}
}

func (m Model) handleLoadedKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		return m.quit()

	case "r":
		return m, m.refreshPreview()

	case "b":
		return m.backToSelect()

	default:
		return m, nil
	}
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}

	return m, tea.Quit
}

func (m Model) backToSelect() (tea.Model, tea.Cmd) {
	if m.conn != nil {
		_ = m.conn.Close()
		m.conn = nil
	}

	m.state = stateSelect
	m.loadedName = ""
	m.count = 0
	m.preview = db.Preview{}
	m.nameInput = nil
	m.err = ""

	return m, m.listDatabases()
}

func (m Model) viewSelect() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("SQLite TUI") + "\n")
	b.WriteString(descStyle.Render("Database directory: "+m.cfg.DBDir) + "\n\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render("Error: "+m.err) + "\n\n")
	}

	if len(m.dbs) == 0 {
		b.WriteString(descStyle.Render("No .db files found.") + "\n")
	} else {
		for i, info := range m.dbs {
			if i == m.selected {
				b.WriteString("> " + cursorStyle.Render(info.Name) + "\n")
			} else {
				b.WriteString("  " + info.Name + "\n")
			}
		}
	}

	b.WriteString("\n" + descStyle.Render("up/down select · enter load · n new · q quit") + "\n")

	return b.String()
}

func (m Model) viewName() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Create new database") + "\n\n")

	if m.err != "" {
		b.WriteString(errorStyle.Render("Error: "+m.err) + "\n\n")
	}

	b.WriteString(descStyle.Render("Name (letters and numbers only):") + "\n")

	input := string(m.nameInput)

	if input == "" {
		b.WriteString("> " + descStyle.Render("(empty)") + "█" + "\n\n")
	} else {
		b.WriteString("> " + cursorStyle.Render(input) + "█" + "\n\n")
	}

	b.WriteString(descStyle.Render("enter create · esc back") + "\n")

	return b.String()
}

func (m Model) viewLoaded() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Database: "+m.loadedName) + "\n")
	b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.count))

	if m.err != "" {
		b.WriteString(errorStyle.Render("Error: "+m.err) + "\n\n")
	}

	if len(m.preview.Columns) == 0 {
		b.WriteString(descStyle.Render("No preview available.") + "\n")
	} else {
		b.WriteString(previewStyle.Render(previewTable(m.preview)) + "\n")
	}

	b.WriteString("\n" + descStyle.Render("r refresh · b back · q quit") + "\n")

	return b.String()
}

func previewTable(p db.Preview) string {
	var b strings.Builder

	b.WriteString(strings.Join(p.Columns, " | "))

	for _, row := range p.Rows {
		b.WriteString("\n" + row)
	}

	return b.String()
}

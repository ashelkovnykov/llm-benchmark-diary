package tui

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
	scrSelect screen = iota
	scrCreate
	scrView
)

// App is the root bubbletea model.
type App struct {
	conn   *sql.DB
	dbName string

	width  int
	height int

	screen screen

	// selection screen
	dbs    []string
	cursor int

	// create screen
	input     textinput.Model
	createErr string

	// view screen
	columns []string
	rows    [][]string
	total   int

	err error
}

// New builds the root model with the initial list of databases.
func New() (*App, error) {
	names, err := db.List()
	if err != nil {
		return nil, err
	}

	ti := textinput.New()
	ti.Placeholder = "database name"
	ti.CharLimit = 32
	ti.Width = 32
	ti.Prompt = "> "

	return &App{
		dbs:    names,
		screen: scrSelect,
		input:  ti,
	}, nil
}

func (a *App) Init() tea.Cmd { return nil }

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
		return a, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
	}

	switch a.screen {
	case scrSelect:
		return a.updateSelect(msg)
	case scrCreate:
		return a.updateCreate(msg)
	case scrView:
		return a.updateView(msg)
	}
	return a, nil
}

func (a *App) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}

	total := len(a.dbs) + 1 // +1 for the "create new" option
	switch key.String() {
	case "up", "k":
		a.cursor = (a.cursor - 1 + total) % total
	case "down", "j":
		a.cursor = (a.cursor + 1) % total
	case "enter":
		if a.cursor == len(a.dbs) {
			a.screen = scrCreate
			a.createErr = ""
			a.input.SetValue("")
			a.input.Focus()
			return a, textinput.Blink
		}
		return a.open(a.dbs[a.cursor])
	case "q":
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) updateCreate(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "esc":
			a.input.Blur()
			a.screen = scrSelect
			return a, nil
		case "enter":
			name := strings.TrimSpace(a.input.Value())
			if err := db.ValidateName(name); err != nil {
				a.createErr = err.Error()
				return a, nil
			}
			if err := db.Create(name); err != nil {
				a.createErr = err.Error()
				return a, nil
			}
			if names, err := db.List(); err == nil {
				a.dbs = names
			}
			a.input.Blur()
			return a.open(name)
		}
		a.createErr = ""
	}

	var cmd tea.Cmd
	a.input, cmd = a.input.Update(msg)
	return a, cmd
}

func (a *App) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}
	switch key.String() {
	case "esc", "q":
		if a.conn != nil {
			a.conn.Close()
			a.conn = nil
		}
		if names, err := db.List(); err == nil {
			a.dbs = names
			if a.cursor > len(a.dbs) {
				a.cursor = 0
			}
		}
		a.screen = scrSelect
	}
	return a, nil
}

// open loads a database and switches to the view screen.
func (a *App) open(name string) (tea.Model, tea.Cmd) {
	conn, err := db.Open(name)
	if err != nil {
		a.err = err
		return a, nil
	}

	cols, rows, err := db.Preview(conn, 10)
	if err != nil {
		conn.Close()
		a.err = err
		return a, nil
	}

	total, err := db.CountRecords(conn)
	if err != nil {
		conn.Close()
		a.err = err
		return a, nil
	}

	if a.conn != nil {
		a.conn.Close()
	}
	a.conn = conn
	a.dbName = name
	a.columns = cols
	a.rows = rows
	a.total = total
	a.screen = scrView
	return a, nil
}

func (a *App) View() string {
	if a.err != nil {
		return errorStyle.Render("error: "+a.err.Error()) + "\n\n" +
			helpStyle.Render("q quit")
	}

	switch a.screen {
	case scrSelect:
		return a.viewSelect()
	case scrCreate:
		return a.viewCreate()
	case scrView:
		return a.viewRecords()
	}
	return ""
}

func (a *App) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("SQL Navigator") + "\n")
	b.WriteString(subtleStyle.Render("Choose a database to open") + "\n\n")

	if len(a.dbs) == 0 {
		b.WriteString(subtleStyle.Render("  (no databases found)") + "\n")
	}
	for i, name := range a.dbs {
		if i == a.cursor {
			b.WriteString(cursorStyle.Render("> "+name) + "\n")
		} else {
			b.WriteString("  " + name + "\n")
		}
	}

	if a.cursor == len(a.dbs) {
		b.WriteString(cursorStyle.Render("> + Create new database") + "\n")
	} else {
		b.WriteString("  + Create new database\n")
	}

	b.WriteString("\n" + helpStyle.Render("↑/↓ move • enter select • q quit"))
	return b.String()
}

func (a *App) viewCreate() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("New database") + "\n")
	b.WriteString(subtleStyle.Render("Letters and numbers only") + "\n\n")
	b.WriteString(a.input.View() + "\n")

	if a.createErr != "" {
		b.WriteString("\n" + errorStyle.Render("✗ "+a.createErr) + "\n")
	}

	b.WriteString("\n" + helpStyle.Render("enter create • esc cancel"))
	return b.String()
}

func (a *App) viewRecords() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Database: "+a.dbName) + "\n")
	b.WriteString(fmt.Sprintf("Records: %d\n\n", a.total))

	if len(a.rows) == 0 {
		b.WriteString(subtleStyle.Render("no rows") + "\n")
	} else {
		b.WriteString(renderTable(a.columns, a.rows))
	}

	b.WriteString("\n" + helpStyle.Render("esc back • q quit"))
	return b.String()
}

func renderTable(cols []string, rows [][]string) string {
	widths := make([]int, len(cols))
	for i, c := range cols {
		widths[i] = len(c)
	}
	for _, r := range rows {
		for i, c := range r {
			if i < len(widths) && len(c) > widths[i] {
				widths[i] = len(c)
			}
		}
	}

	var b strings.Builder
	writeRow := func(cells []string, style lipgloss.Style) {
		parts := make([]string, len(cells))
		for i, c := range cells {
			parts[i] = fmt.Sprintf("%-*s", widths[i], c)
		}
		b.WriteString(style.Render(strings.Join(parts, "  ")) + "\n")
	}

	writeRow(cols, headerStyle)
	for _, r := range rows {
		writeRow(r, lipgloss.NewStyle())
	}
	return b.String()
}

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	subtleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	cursorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	errorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	headerStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
)

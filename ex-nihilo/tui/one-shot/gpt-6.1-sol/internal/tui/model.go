package tui

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"dbnav/internal/db"

	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	selectScreen screen = iota
	createScreen
	dataScreen
)

type Model struct {
	dir      string
	screen   screen
	names    []string
	cursor   int
	name     []rune
	message  string
	database *db.Database
	preview  db.Preview
}

func New(dir string) (*Model, error) {
	names, err := db.List(dir)
	if err != nil {
		return nil, err
	}
	return &Model{dir: dir, names: names}, nil
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Close() error {
	if m.database == nil {
		return nil
	}
	err := m.database.Close()
	m.database = nil
	return err
}

func (m *Model) load(database *db.Database) {
	preview, err := database.Preview()
	if err != nil {
		_ = database.Close()
		m.message = "Cannot load records: " + err.Error()
		return
	}

	m.database = database
	m.preview = preview
	m.message = ""
	m.screen = dataScreen
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case selectScreen:
		switch key.String() {
		case "q", "esc":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.names)-1 {
				m.cursor++
			}

		case "n":
			m.screen = createScreen
			m.name = nil
			m.message = ""

		case "enter":
			if len(m.names) == 0 {
				break
			}
			database, err := db.Open(m.dir, m.names[m.cursor])
			if err != nil {
				m.message = err.Error()
			} else {
				m.load(database)
			}

		case "r":
			names, err := db.List(m.dir)
			if err != nil {
				m.message = err.Error()
			} else {
				m.names = names
				m.cursor = 0
				m.message = ""
			}
		}

	case createScreen:
		switch key.Type {
		case tea.KeyEsc:
			m.screen = selectScreen
			m.message = ""

		case tea.KeyEnter:
			database, err := db.Create(m.dir, string(m.name))
			if err != nil {
				m.message = err.Error()
			} else {
				m.load(database)
			}

		case tea.KeyBackspace, tea.KeyDelete:
			if len(m.name) > 0 {
				m.name = m.name[:len(m.name)-1]
			}
			m.message = ""

		case tea.KeyRunes:
			m.name = append(m.name, key.Runes...)
			m.message = ""

		case tea.KeySpace:
			m.name = append(m.name, ' ')
			m.message = ""
		}

	case dataScreen:
		switch key.String() {
		case "q":
			return m, tea.Quit

		case "esc", "b":
			if err := m.Close(); err != nil {
				m.message = err.Error()
			} else {
				m.message = ""
			}
			m.screen = selectScreen
			m.cursor = 0
			names, err := db.List(m.dir)
			if err != nil {
				m.message = err.Error()
			} else {
				m.names = names
			}

		case "r":
			preview, err := m.database.Preview()
			if err != nil {
				m.message = err.Error()
			} else {
				m.preview = preview
				m.message = ""
			}
		}
	}

	return m, nil
}

// Quote database contents so control characters cannot affect the terminal.
func display(value string) string {
	return fmt.Sprintf("%q", value)
}

func (m *Model) View() string {
	var output strings.Builder

	switch m.screen {
	case selectScreen:
		output.WriteString("DB Navigator\n\nChoose a database:\n\n")
		if len(m.names) == 0 {
			output.WriteString("  No databases found. Press n to create one.\n")
		}
		for i, name := range m.names {
			marker := " "
			if i == m.cursor {
				marker = ">"
			}
			fmt.Fprintf(&output, "%s %s\n", marker, display(name))
		}
		output.WriteString("\n↑/↓: select • enter: load • n: new • r: refresh • q: quit\n")

	case createScreen:
		output.WriteString("Create database\n\n")
		output.WriteString("Name: " + display(string(m.name)) + " █\n")
		output.WriteString("\nUse only letters and numbers.\n")
		output.WriteString("enter: create and load • esc: cancel • ctrl+c: quit\n")

	case dataScreen:
		fmt.Fprintf(&output, "Database: %s\n\n", display(m.database.Name))
		fmt.Fprintf(&output, "Total records: %d\n", m.preview.Count)
		fmt.Fprintf(&output, "Showing %d rows, ordered by rowid:\n\n", len(m.preview.Rows))

		table := tabwriter.NewWriter(&output, 0, 4, 2, ' ', 0)
		headers := make([]string, len(m.preview.Columns))
		for i, column := range m.preview.Columns {
			headers[i] = display(column)
		}
		fmt.Fprintln(table, strings.Join(headers, "\t"))

		for _, row := range m.preview.Rows {
			cells := make([]string, len(row))
			for i, cell := range row {
				cells[i] = display(cell)
			}
			fmt.Fprintln(table, strings.Join(cells, "\t"))
		}
		_ = table.Flush()

		output.WriteString("\nr: refresh • b/esc: databases • q: quit\n")
	}

	if m.message != "" {
		output.WriteString("\nError: " + display(m.message) + "\n")
	}

	return output.String()
}

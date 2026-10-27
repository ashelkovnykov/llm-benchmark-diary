package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"example.com/dbtui/internal/db"
)

type screen int

const (
	chooseDB screen = iota
	createDB
	viewDB
)

type Model struct {
	screen   screen
	names    []string
	cursor   int
	input    string
	message  string
	snapshot db.Snapshot
}

func New() (Model, error) {
	names, err := db.List()
	if err != nil {
		return Model{}, err
	}
	return Model{names: names}, nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	if key.String() == "ctrl+c" {
		return m, tea.Quit
	}

	switch m.screen {
	case chooseDB:
		switch key.String() {
		case "q", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			// The final item is "Create a new database".
			if m.cursor < len(m.names) {
				m.cursor++
			}
		case "enter":
			m.message = ""
			if m.cursor == len(m.names) {
				m.screen = createDB
				m.input = ""
			} else {
				snapshot, err := db.Load(m.names[m.cursor])
				if err != nil {
					m.message = err.Error()
				} else {
					m.snapshot = snapshot
					m.screen = viewDB
				}
			}
		}

	case createDB:
		switch key.String() {
		case "esc":
			m.screen = chooseDB
			m.message = ""
		case "backspace", "ctrl+h":
			runes := []rune(m.input)
			if len(runes) > 0 {
				m.input = string(runes[:len(runes)-1])
			}
			m.message = ""
		case "enter":
			if err := db.Create(m.input); err != nil {
				m.message = err.Error()
				break
			}

			snapshot, err := db.Load(m.input)
			if err != nil {
				m.message = fmt.Sprintf("Database created, but loading failed: %v", err)
				m.names, _ = db.List()
				m.cursor = 0
				m.screen = chooseDB
				break
			}
			m.snapshot = snapshot
			m.message = ""
			m.screen = viewDB
		default:
			if key.Type == tea.KeyRunes {
				m.input += string(key.Runes)
				m.message = ""
			} else if key.Type == tea.KeySpace {
				m.input += " "
				m.message = ""
			}
		}

	case viewDB:
		switch key.String() {
		case "q":
			return m, tea.Quit
		case "esc", "b":
			names, err := db.List()
			if err != nil {
				m.message = err.Error()
				break
			}
			m.names = names
			m.cursor = 0
			m.message = ""
			m.screen = chooseDB
		}
	}

	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	switch m.screen {
	case chooseDB:
		b.WriteString("Choose a database\n\n")
		for i, name := range m.names {
			writeChoice(&b, i == m.cursor, name+".db")
		}
		writeChoice(&b, m.cursor == len(m.names), "Create a new database")
		b.WriteString("\n↑/↓ or j/k: navigate • enter: select • q: quit\n")

	case createDB:
		b.WriteString("New database name (letters and digits only)\n\n")
		fmt.Fprintf(&b, "> %s_\n", m.input)
		b.WriteString("\nenter: create • esc: back\n")

	case viewDB:
		fmt.Fprintf(&b, "Database: %s.db\n", m.snapshot.Name)
		fmt.Fprintf(&b, "Total records: %d\n\n", m.snapshot.Count)
		b.WriteString("First 10 rows:\n")
		fmt.Fprintln(&b, formatRow(m.snapshot.Columns))
		for _, row := range m.snapshot.Rows {
			fmt.Fprintln(&b, formatRow(row))
		}
		if len(m.snapshot.Rows) == 0 {
			b.WriteString("(no records)\n")
		}
		b.WriteString("\nb/esc: back • q: quit\n")
	}

	if m.message != "" {
		fmt.Fprintf(&b, "\nError: %s\n", m.message)
	}
	return b.String()
}

func writeChoice(b *strings.Builder, selected bool, label string) {
	marker := "  "
	if selected {
		marker = "> "
	}
	fmt.Fprintf(b, "%s%s\n", marker, label)
}

func formatRow(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		// Quote values so newlines and terminal control characters are visible.
		quoted[i] = fmt.Sprintf("%q", value)
	}
	return strings.Join(quoted, " | ")
}

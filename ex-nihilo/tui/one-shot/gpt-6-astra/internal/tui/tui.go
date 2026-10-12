package tui

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"dbnav/internal/db"
)

type screen int

const (
	selectScreen screen = iota
	createScreen
	dataScreen
)

type resultMsg struct {
	snapshot db.Snapshot
	err      error
}

type Model struct {
	directory string
	names     []string
	cursor    int
	screen    screen
	name      string
	err       string
	busy      bool
	snapshot  db.Snapshot
}

func New(directory string, names []string) Model {
	return Model{
		directory: directory,
		names:     names,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case resultMsg:
		m.busy = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.snapshot = msg.snapshot
		m.screen = dataScreen
		m.err = ""
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.busy {
			return m, nil
		}

		switch m.screen {
		case selectScreen:
			switch msg.String() {
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
				m.name = ""
				m.err = ""
			case "enter":
				if len(m.names) > 0 {
					m.busy = true
					m.err = ""
					directory, filename := m.directory, m.names[m.cursor]
					return m, func() tea.Msg {
						snapshot, err := db.Load(directory, filename)
						return resultMsg{snapshot: snapshot, err: err}
					}
				}
			}

		case createScreen:
			switch msg.String() {
			case "esc":
				m.screen = selectScreen
				m.err = ""
			case "enter":
				m.busy = true
				m.err = ""
				directory, name := m.directory, m.name
				return m, func() tea.Msg {
					snapshot, err := db.Create(directory, name)
					return resultMsg{snapshot: snapshot, err: err}
				}
			case "backspace", "ctrl+h":
				runes := []rune(m.name)
				if len(runes) > 0 {
					m.name = string(runes[:len(runes)-1])
				}
				m.err = ""
			case "ctrl+u":
				m.name = ""
				m.err = ""
			default:
				if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
					m.name += string(msg.Runes)
					m.err = ""
				}
			}

		case dataScreen:
			switch msg.String() {
			case "q", "esc":
				return m, tea.Quit
			case "b":
				names, err := db.List(m.directory)
				if err != nil {
					m.err = err.Error()
					return m, nil
				}
				m.names = names
				m.cursor = 0
				m.screen = selectScreen
				m.err = ""
			case "r":
				m.busy = true
				m.err = ""
				directory, filename := m.directory, m.snapshot.Name
				return m, func() tea.Msg {
					snapshot, err := db.Load(directory, filename)
					return resultMsg{snapshot: snapshot, err: err}
				}
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	var out strings.Builder

	switch m.screen {
	case selectScreen:
		out.WriteString("DB Navigator\n\nChoose a database:\n\n")
		if len(m.names) == 0 {
			out.WriteString("  No databases found in ./db/.\n")
		}
		for i, name := range m.names {
			prefix := "  "
			if i == m.cursor {
				prefix = "> "
			}
			fmt.Fprintf(&out, "%s%s\n", prefix, safe(name))
		}
		out.WriteString("\n↑/↓ select • enter load • n new • q quit\n")

	case createScreen:
		out.WriteString("Create database\n\n")
		out.WriteString("Name (letters and digits only; .db is added automatically):\n")
		fmt.Fprintf(&out, "> %s█\n", safe(m.name))
		out.WriteString("\nenter create • esc cancel • ctrl+u clear\n")

	case dataScreen:
		fmt.Fprintf(&out, "Database: %s\n", safe(m.snapshot.Name))
		fmt.Fprintf(&out, "Total records: %d\n", m.snapshot.Count)
		fmt.Fprintf(&out, "Showing %d rows, ordered by rowid:\n\n", len(m.snapshot.Rows))

		table := tabwriter.NewWriter(&out, 0, 4, 2, ' ', 0)
		writeRow(table, m.snapshot.Columns)
		for _, row := range m.snapshot.Rows {
			writeRow(table, row)
		}
		_ = table.Flush()

		if len(m.snapshot.Rows) == 0 {
			out.WriteString("(no records)\n")
		}
		out.WriteString("\nr refresh • b databases • q quit\n")
	}

	if m.busy {
		out.WriteString("\nWorking…\n")
	}
	if m.err != "" {
		fmt.Fprintf(&out, "\nError: %s\n", safe(m.err))
	}

	return out.String()
}

func writeRow(table *tabwriter.Writer, cells []string) {
	for i, cell := range cells {
		if i > 0 {
			fmt.Fprint(table, "\t")
		}
		runes := []rune(safe(cell))
		if len(runes) > 32 {
			runes = append(runes[:31], '…')
		}
		fmt.Fprint(table, string(runes))
	}
	fmt.Fprintln(table)
}

func safe(value string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
}

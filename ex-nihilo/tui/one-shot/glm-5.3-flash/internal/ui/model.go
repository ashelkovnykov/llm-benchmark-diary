package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
)

type screen int

const (
	screenChoose screen = iota
	screenNewName
	screenLoaded
)

type Model struct {
	store   *db.Store
	screen  screen
	dbs     []string
	cursor  int
	nameBuf string
	errMsg  string
}

func New(store *db.Store) Model {
	dbs, _ := store.List()
	return Model{store: store, screen: screenChoose, dbs: dbs}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

		switch m.screen {
		case screenChoose:
			switch key.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.dbs)-1 {
					m.cursor++
				}
			case "n":
				m.screen = screenNewName
				m.nameBuf, m.errMsg = "", ""
			case "enter":
				if len(m.dbs) > 0 {
					return m.load(m.dbs[m.cursor])
				}
			}
		case screenNewName:
			switch key.String() {
			case "esc":
				m.screen, m.nameBuf, m.errMsg = screenChoose, "", ""
			case "enter":
				if reason := m.store.ValidateName(m.nameBuf); reason != "" {
					m.errMsg = reason
				} else {
					if err := m.store.Create(m.nameBuf); err != nil {
						m.errMsg = fmt.Sprintf("failed to create database: %v", err)
					} else {
						return m.load(m.nameBuf)
					}
				}
			case "backspace":
				if len(m.nameBuf) > 0 {
					m.nameBuf = m.nameBuf[:len(m.nameBuf)-1]
				}
			default:
				if len(key.String()) == 1 {
					m.nameBuf += key.String()
				}
			}
		case screenLoaded:
			if key.String() == "q" {
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) load(name string) (tea.Model, tea.Cmd) {
	if err := m.store.Load(name); err != nil {
		m.errMsg = fmt.Sprintf("failed to load database: %v", err)
		return m, nil
	}
	m.errMsg = ""
	m.screen = screenLoaded
	return m, nil
}

func (m Model) View() string {
	var b strings.Builder

	switch m.screen {
	case screenChoose:
		b.WriteString("Select a database (up/down, enter) — n: new — ctrl+c: quit\n\n")
		if len(m.dbs) == 0 {
			b.WriteString("  (no databases found in ./db)\n")
		}
		for i, name := range m.dbs {
			marker := "  "
			if i == m.cursor {
				marker = "> "
			}
			b.WriteString(marker + name + "\n")
		}
		if m.errMsg != "" {
			b.WriteString("\nError: " + m.errMsg + "\n")
		}

	case screenNewName:
		b.WriteString("Name for new database (alphanumeric only)\n\n")
		b.WriteString("> " + m.nameBuf + "\n\n")
		b.WriteString("enter: create — esc: back\n")
		if m.errMsg != "" {
			b.WriteString("\nError: " + m.errMsg + "\n")
		}

	case screenLoaded:
		b.WriteString("Database loaded — q: quit\n\n")
		count, err := m.store.Count()
		if err != nil {
			b.WriteString("Error reading database: " + err.Error() + "\n")
			return b.String()
		}
		fmt.Fprintf(&b, "Total records: %d\n\n", count)
		recs, err := m.store.First(10)
		if err != nil {
			b.WriteString("Error reading rows: " + err.Error() + "\n")
			return b.String()
		}
		fmt.Fprintf(&b, "%-6s %-16s %-24s %s\n", "ROWID", "NAME", "VALUE", "CREATED_AT")
		for _, r := range recs {
			fmt.Fprintf(&b, "%-6d %-16s %-24s %s\n", r.RowID, r.Name, r.Value, r.CreatedAt)
		}
	}

	return b.String()
}

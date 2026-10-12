// ./internal/tui/model.go
package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	dbpkg "dbnav/internal/db"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	stateChoose state = iota
	stateCreate
	stateLoaded
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	errStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	helpStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
)

type Model struct {
	state    state
	dbs      []string
	cursor   int
	choices  []string // dbs + "<Create new DB>"
	input    textinput.Model
	errMsg   string
	stats    *dbpkg.Stats
	quitting bool
}

func New(dbs []string) Model {
	ti := textinput.New()
	ti.Placeholder = "alphanumeric name"
	ti.CharLimit = 64
	ti.Width = 30

	choices := append(append([]string{}, dbs...), "[Create new DB]")
	return Model{
		state:   stateChoose,
		dbs:     dbs,
		choices: choices,
		input:   ti,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateChoose:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				if m.cursor == len(m.choices)-1 {
					// create new
					m.state = stateCreate
					m.input.Focus()
					m.errMsg = ""
					return m, textinput.Blink
				}
				// load selected
				path := filepath.Join("./db", m.dbs[m.cursor])
				stats, err := dbpkg.LoadStats(path)
				if err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				m.stats = &stats
				m.state = stateLoaded
			}

		case stateCreate:
			switch msg.String() {
			case "ctrl+c", "esc":
				m.state = stateChoose
				m.errMsg = ""
				m.input.Blur()
				m.input.SetValue("")
				return m, nil
			case "enter":
				name := strings.TrimSpace(m.input.Value())
				if err := dbpkg.ValidateName(name, m.dbs); err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				path, err := dbpkg.CreateDB(name)
				if err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				stats, err := dbpkg.LoadStats(path)
				if err != nil {
					m.errMsg = err.Error()
					return m, nil
				}
				m.stats = &stats
				m.state = stateLoaded
				m.input.Blur()
				return m, nil
			default:
				var cmd tea.Cmd
				m.input, cmd = m.input.Update(msg)
				return m, cmd
			}

		case stateLoaded:
			switch msg.String() {
			case "ctrl+c", "q", "esc":
				m.quitting = true
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	switch m.state {
	case stateChoose:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Select a database") + "\n\n")
		if len(m.dbs) == 0 {
			b.WriteString(helpStyle.Render("No DBs found in ./db/") + "\n\n")
		}
		for i, c := range m.choices {
			cursor := "  "
			style := lipgloss.NewStyle()
			if m.cursor == i {
				cursor = "> "
				style = selStyle
			}
			b.WriteString(style.Render(cursor+c) + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("↑/↓ navigate • enter select • q quit"))
		if m.errMsg != "" {
			b.WriteString("\n" + errStyle.Render(m.errMsg))
		}
		return b.String()

	case stateCreate:
		var b strings.Builder
		b.WriteString(titleStyle.Render("Create new DB") + "\n\n")
		b.WriteString("Name (alphanumeric only):\n")
		b.WriteString(m.input.View() + "\n\n")
		if m.errMsg != "" {
			b.WriteString(errStyle.Render("Error: "+m.errMsg) + "\n\n")
		}
		b.WriteString(helpStyle.Render("enter confirm • esc back"))
		return b.String()

	case stateLoaded:
		var b strings.Builder
		b.WriteString(titleStyle.Render("DB: "+m.stats.Path) + "\n")
		b.WriteString(fmt.Sprintf("Total records: %d\n\n", m.stats.Count))
		// header
		b.WriteString(strings.Join(m.stats.Cols, " | ") + "\n")
		b.WriteString(strings.Repeat("-", 60) + "\n")
		for _, r := range m.stats.Rows {
			b.WriteString(strings.Join(r, " | ") + "\n")
		}
		if m.stats.Count > len(m.stats.Rows) {
			b.WriteString(helpStyle.Render(fmt.Sprintf("... showing 10 of %d", m.stats.Count)) + "\n")
		} else if m.stats.Count == 0 {
			b.WriteString(helpStyle.Render("(no records)") + "\n")
		}
		b.WriteString("\n" + helpStyle.Render("q/esc quit"))
		return b.String()
	}
	return ""
}

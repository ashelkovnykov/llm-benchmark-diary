package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"sqltui/internal/db"
)

type screen int

const (
	screenSelect screen = iota
	screenName
	screenView
)

const previewLimit = 10

type Model struct {
	screen screen
	dbs    []string
	cursor int // index into dbs; len(dbs) == "create new"
	input  textinput.Model
	errMsg string

	conn  *db.DB
	count int
	cols  []string
	rows  [][]string
}

func New() (Model, error) {
	ti := textinput.New()
	ti.Placeholder = "database name"
	ti.CharLimit = 64

	m := Model{input: ti}
	if err := m.refreshList(); err != nil {
		return m, err
	}
	return m, nil
}

func (m *Model) refreshList() error {
	dbs, err := db.List()
	if err != nil {
		return err
	}
	m.dbs = dbs
	if m.cursor > len(dbs) {
		m.cursor = len(dbs)
	}
	return nil
}

func (m *Model) load(d *db.DB) error {
	count, err := d.Count()
	if err != nil {
		return err
	}
	cols, rows, err := d.Preview(previewLimit)
	if err != nil {
		return err
	}
	if m.conn != nil {
		m.conn.Close()
	}
	m.conn, m.count, m.cols, m.rows = d, count, cols, rows
	m.screen = screenView
	m.errMsg = ""
	return nil
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if ok && key.Type == tea.KeyCtrlC {
		return m.quit()
	}

	switch m.screen {
	case screenSelect:
		return m.updateSelect(msg)
	case screenName:
		return m.updateName(msg)
	case screenView:
		return m.updateView(msg)
	}
	return m, nil
}

func (m Model) quit() (tea.Model, tea.Cmd) {
	if m.conn != nil {
		m.conn.Close()
	}
	return m, tea.Quit
}

func (m Model) updateSelect(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q", "esc":
		return m.quit()
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.dbs) {
			m.cursor++
		}
	case "n":
		return m.startNaming()
	case "enter":
		if m.cursor == len(m.dbs) {
			return m.startNaming()
		}
		d, err := db.Open(m.dbs[m.cursor])
		if err != nil {
			m.errMsg = err.Error()
			return m, nil
		}
		if err := m.load(d); err != nil {
			d.Close()
			m.errMsg = err.Error()
		}
	}
	return m, nil
}

func (m Model) startNaming() (tea.Model, tea.Cmd) {
	m.screen = screenName
	m.errMsg = ""
	m.input.SetValue("")
	return m, m.input.Focus()
}

func (m Model) updateName(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyEsc:
			m.screen = screenSelect
			m.errMsg = ""
			m.input.Blur()
			return m, nil
		case tea.KeyEnter:
			name := strings.TrimSpace(m.input.Value())
			if err := db.ValidateName(name, m.dbs); err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			d, err := db.Create(name)
			if err != nil {
				m.errMsg = err.Error()
				return m, nil
			}
			if err := m.load(d); err != nil {
				d.Close()
				m.errMsg = err.Error()
				return m, nil
			}
			m.input.Blur()
			_ = m.refreshList()
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) updateView(msg tea.Msg) (tea.Model, tea.Cmd) {
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	switch key.String() {
	case "q":
		return m.quit()
	case "esc", "b":
		if m.conn != nil {
			m.conn.Close()
			m.conn = nil
		}
		m.screen = screenSelect
		_ = m.refreshList()
	}
	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenSelect:
		return m.viewSelect()
	case screenName:
		return m.viewName()
	case screenView:
		return m.viewData()
	}
	return ""
}

func (m Model) viewSelect() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Select a database") + "\n\n")

	if len(m.dbs) == 0 {
		b.WriteString(dimStyle.Render("  (no databases found in "+db.Dir+")") + "\n")
	}
	for i, name := range m.dbs {
		b.WriteString(m.line(i, name))
	}
	b.WriteString(m.line(len(m.dbs), "+ Create new database"))

	if m.errMsg != "" {
		b.WriteString("\n" + errStyle.Render(m.errMsg) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("↑/↓ move • enter select • n new • q quit"))
	return b.String()
}

func (m Model) line(i int, text string) string {
	if i == m.cursor {
		return cursorStyle.Render("> "+text) + "\n"
	}
	return "  " + text + "\n"
}

func (m Model) viewName() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("New database") + "\n\n")
	b.WriteString("Name: " + m.input.View() + "\n")
	if m.errMsg != "" {
		b.WriteString("\n" + errStyle.Render(m.errMsg) + "\n")
	}
	b.WriteString("\n" + dimStyle.Render("enter create • esc cancel"))
	return b.String()
}

func (m Model) viewData() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Database: "+m.conn.Name) + "\n\n")
	fmt.Fprintf(&b, "Total records: %d\n\n", m.count)

	widths := make([]int, len(m.cols))
	for i, c := range m.cols {
		widths[i] = len(c)
	}
	for _, r := range m.rows {
		for i, v := range r {
			if len(v) > widths[i] {
				widths[i] = len(v)
			}
		}
	}

	var hdr []string
	for i, c := range m.cols {
		hdr = append(hdr, pad(c, widths[i]))
	}
	b.WriteString(headerStyle.Render(strings.Join(hdr, "  ")) + "\n")
	for _, r := range m.rows {
		var cells []string
		for i, v := range r {
			cells = append(cells, pad(v, widths[i]))
		}
		b.WriteString(strings.Join(cells, "  ") + "\n")
	}
	if len(m.rows) == 0 {
		b.WriteString(dimStyle.Render("(no rows)") + "\n")
	}

	fmt.Fprintf(&b, "\n%s\n", dimStyle.Render(fmt.Sprintf("showing %d of %d", len(m.rows), m.count)))
	b.WriteString(dimStyle.Render("esc/b back • q quit"))
	return b.String()
}

func pad(s string, w int) string {
	return s + strings.Repeat(" ", w-len(s))
}

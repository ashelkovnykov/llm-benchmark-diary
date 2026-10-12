package ui

import (
	"fmt"

	"sqlnav/internal/db"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	viewSelectDB sessionState = iota
	viewCreateDB
	viewVisualizer
)

type Model struct {
	state        sessionState
	dbList       []string
	cursor       int
	loadedDB     string
	totalRecords int
	textInput    textinput.Model
	table        table.Model
	err          error
}

func NewModel() Model {
	ti := textinput.New()
	ti.Placeholder = "mydb"
	ti.CharLimit = 32

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "rowid", Width: 8},
			{Title: "title", Width: 20},
			{Title: "value", Width: 10},
			{Title: "created_at", Width: 20},
		}),
		table.WithFocused(true),
		table.WithHeight(11),
	)

	m := Model{
		state:     viewSelectDB,
		textInput: ti,
		table:     t,
	}
	m.refreshDBList()
	return m
}

func (m *Model) refreshDBList() {
	dbs, err := db.ListDatabases()
	if err != nil {
		m.err = err
		return
	}
	m.dbList = dbs
	if m.cursor >= len(m.dbList) && len(m.dbList) > 0 {
		m.cursor = len(m.dbList) - 1
	}
}

func (m *Model) loadDatabase(filename string) error {
	total, records, err := db.LoadData(filename)
	if err != nil {
		return err
	}

	var rows []table.Row
	for _, r := range records {
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", r.RowID),
			r.Title,
			fmt.Sprintf("%.2f", r.Value),
			r.CreatedAt,
		})
	}

	m.loadedDB = filename
	m.totalRecords = total
	m.table.SetRows(rows)
	m.state = viewVisualizer
	m.err = nil
	return nil
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case viewSelectDB:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.dbList)-1 {
					m.cursor++
				}
			case "n":
				m.state = viewCreateDB
				m.textInput.Reset()
				m.textInput.Focus()
				m.err = nil
				return m, textinput.Blink
			case "enter":
				if len(m.dbList) > 0 {
					if err := m.loadDatabase(m.dbList[m.cursor]); err != nil {
						m.err = err
					}
				}
			}

		case viewCreateDB:
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = viewSelectDB
				m.err = nil
			case "enter":
				input := m.textInput.Value()
				if err := db.ValidateDBName(input, m.dbList); err != nil {
					m.err = err
					return m, nil
				}

				createdFile, err := db.CreateDatabase(input)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.refreshDBList()
				if err := m.loadDatabase(createdFile); err != nil {
					m.err = err
					m.state = viewSelectDB
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}

		case viewVisualizer:
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.state = viewSelectDB
				m.refreshDBList()
			}
		}
	}
	return m, nil
}

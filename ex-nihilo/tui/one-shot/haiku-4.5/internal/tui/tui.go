package tui

import (
	"database/sql"
	"strings"

	"dbui/internal/db"
	tea "github.com/charmbracelet/bubbletea"
)

type screen int

const (
	screenSelectDB screen = iota
	screenCreateDB
	screenViewDB
)

type Model struct {
	screen        screen
	databases     []string
	selectedIdx   int
	createInput   strings.Builder
	createError   string
	currentDB     *sql.DB
	currentDBName string
	recordCount   int
	firstRecords  []map[string]interface{}
}

func NewInitialModel() Model {
	databases, _ := db.ListDatabases()
	return Model{
		screen:      screenSelectDB,
		databases:   databases,
		selectedIdx: 0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentDB != nil {
				m.currentDB.Close()
			}
			return m, tea.Quit

		case "ctrl+n":
			if m.screen != screenCreateDB {
				m.screen = screenCreateDB
				m.createInput.Reset()
				m.createError = ""
			}

		case "esc":
			if m.screen == screenCreateDB {
				m.screen = screenSelectDB
				m.createInput.Reset()
				m.createError = ""
			}

		case "up":
			if m.screen == screenSelectDB && m.selectedIdx > 0 {
				m.selectedIdx--
			}

		case "down":
			if m.screen == screenSelectDB && m.selectedIdx < len(m.databases)-1 {
				m.selectedIdx++
			}

		case "enter":
			if m.screen == screenSelectDB {
				return m.selectDatabase()
			} else if m.screen == screenCreateDB {
				return m.createDatabase()
			}

		case "backspace":
			if m.screen == screenCreateDB {
				s := m.createInput.String()
				if len(s) > 0 {
					m.createInput.Reset()
					m.createInput.WriteString(s[:len(s)-1])
				}
				m.createError = ""
			}

		default:
			if m.screen == screenCreateDB && len(msg.String()) == 1 {
				m.createInput.WriteRune(rune(msg.String()[0]))
				m.createError = ""
			}
		}
	}

	return m, nil
}

func (m *Model) selectDatabase() (tea.Model, tea.Cmd) {
	dbName := m.databases[m.selectedIdx]
	dbConn, err := db.OpenDatabase(dbName)
	if err != nil {
		return m, nil
	}

	count, err := db.GetRecordCount(dbConn)
	if err != nil {
		dbConn.Close()
		return m, nil
	}

	records, err := db.GetFirstNRecords(dbConn, 10)
	if err != nil {
		dbConn.Close()
		return m, nil
	}

	m.currentDB = dbConn
	m.currentDBName = dbName
	m.recordCount = count
	m.firstRecords = records
	m.screen = screenViewDB

	return m, nil
}

func (m *Model) createDatabase() (tea.Model, tea.Cmd) {
	name := m.createInput.String()
	valid, errMsg := db.ValidateDatabaseName(name)
	if !valid {
		m.createError = errMsg
		return m, nil
	}

	dbConn, err := db.CreateDatabase(name)
	if err != nil {
		m.createError = "Failed to create database"
		return m, nil
	}

	m.currentDB = dbConn
	m.currentDBName = name
	m.recordCount = 15
	m.firstRecords, _ = db.GetFirstNRecords(dbConn, 10)
	m.screen = screenViewDB

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenSelectDB:
		return m.viewSelectDB()
	case screenCreateDB:
		return m.viewCreateDB()
	case screenViewDB:
		return m.viewViewDB()
	}
	return ""
}

func (m Model) viewSelectDB() string {
	var b strings.Builder
	b.WriteString("Select Database (Press Ctrl+N to create new)\n\n")

	if len(m.databases) == 0 {
		b.WriteString("No databases found. Press Ctrl+N to create a new one.\n")
	} else {
		for i, db := range m.databases {
			prefix := "  "
			if i == m.selectedIdx {
				prefix = "> "
			}
			b.WriteString(prefix + db + "\n")
		}
	}

	b.WriteString("\nPress q or Ctrl+C to quit")
	return b.String()
}

func (m Model) viewCreateDB() string {
	var b strings.Builder
	b.WriteString("Create New Database\n\n")
	b.WriteString("Name: " + m.createInput.String() + "\n\n")

	if m.createError != "" {
		b.WriteString("Error: " + m.createError + "\n\n")
	}

	b.WriteString("Press Enter to create, Esc to cancel")
	return b.String()
}

func (m Model) viewViewDB() string {
	var b strings.Builder
	b.WriteString("Database: " + m.currentDBName + "\n\n")
	b.WriteString("Total Records: ")
	b.WriteString(string(rune(m.recordCount + 48)))
	b.WriteString("\n\n")

	b.WriteString("First Records:\n")
	if len(m.firstRecords) == 0 {
		b.WriteString("(No records)\n")
	} else {
		for _, record := range m.firstRecords {
			b.WriteString("  ")
			first := true
			for k, v := range record {
				if !first {
					b.WriteString(" | ")
				}
				b.WriteString(k + ": " + toString(v))
				first = false
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\nPress q or Ctrl+C to quit")
	return b.String()
}

func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val + 48))
	case int64:
		return string(rune(val + 48))
	default:
		return "nil"
	}
}

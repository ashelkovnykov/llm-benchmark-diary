package tui

import (
	"database/sql"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"

	"control/internal/db"
)

type modelExplore struct {
	db *sql.DB

	loadDBMenu tea.Model
}

func (m *modelExplore) Init() tea.Cmd {
	return nil
}

func (m *modelExplore) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msgt := msg.(type) {
	case tea.WindowSizeMsg:
		_, _ = m.loadDBMenu.Update(msg)
		return m, nil

	case tea.KeyPressMsg:
		switch msgt.Code {
		case 'c':
			if msgt.Mod.Contains(tea.ModCtrl) {
				return m, tea.Quit
			}
		}

	case loadedDBMsg:
		if msgt.DB != nil {
			m.db = msgt.DB
		} else if m.db == nil {
			return m, tea.Quit
		}
	}

	// Default fall-through
	return m, nil
}

func (m *modelExplore) View() (v tea.View) {
	if m.db == nil {
		return tea.NewView("")
	}

	var records []db.Record
	n, err := db.CountRecords(m.db)
	if err == nil {
		records, err = db.FetchRecords(m.db, 10, 0)
	}
	if err != nil {
		v.SetContent(errorStyle.Render(err.Error()))
	}

	b := strings.Builder{}
	b.WriteString(textStyle.Render(fmt.Sprintf("Currently %d records in DB\n", n)))
	b.WriteByte('\n')

	if n == 0 {
		b.WriteString(dimStyle.Render("This database is empty"))
		b.WriteByte('\n')
	} else {
		b.WriteString(titleStyle.Render("| ID  | Count |"))
		b.WriteByte('\n')
		b.WriteString(titleStyle.Render("==============="))
		b.WriteByte('\n')

		for _, r := range records {
			b.WriteString(titleStyle.Render("| "))
			b.WriteString(cellStyle.Render(fmt.Sprintf("%d", r.ID)))
			b.WriteString(titleStyle.Render(" | "))
			b.WriteString(cellStyle.Render(fmt.Sprintf("%d", r.Count)))
			b.WriteString(titleStyle.Render(" |"))
			b.WriteByte('\n')
		}
	}

	v.SetContent(b.String())
	v.AltScreen = true

	return v
}

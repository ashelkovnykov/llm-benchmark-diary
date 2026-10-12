package tui

import (
	"database/sql"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"control/internal/db"
	"control/internal/fileutils"
)

//
// Constants
//

const (
	promptHeight = 1
)

var additionalHelpKeyBindings = []key.Binding{
	key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "new DB")),
}

//
// Style
//

type itemDelegate struct {
	item     lipgloss.Style
	selected lipgloss.Style
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	str := fmt.Sprintf("%d. %s", index+1, listItem)

	fn := d.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.selected.Render("> " + strings.Join(s, " "))
		}
	}

	_, _ = fmt.Fprint(w, fn(str))
}

func newItemDelegate() itemDelegate {
	return itemDelegate{
		item:     lipgloss.NewStyle().PaddingLeft(4),
		selected: lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170")),
	}
}

//
// Item helper
//

type item string

func (i item) FilterValue() string { return string(i) }

//
// Custom message types
//

type refreshDBMsg struct{}
type loadedDBMsg struct {
	DB *sql.DB
}

//
// CMD helpers
//

func refreshDB() tea.Msg {
	return refreshDBMsg{}
}

func loadDBCmd(db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		return loadedDBMsg{
			DB: db,
		}
	}
}

//
// Model
//

type modelLoad struct {
	list   list.Model
	prompt textinput.Model

	parent tea.Model
}

func newLoadModel() *modelLoad {
	//
	// Existing DBs list
	//

	listModel := list.New(make([]list.Item, 0), newItemDelegate(), 0, 0)

	listModel.Styles.Title = lipgloss.NewStyle().MarginLeft(2)
	listModel.Styles.PaginationStyle = listModel.Styles.PaginationStyle.PaddingLeft(4)
	listModel.Styles.HelpStyle = listModel.Styles.HelpStyle.PaddingLeft(4).PaddingBottom(1)

	listModel.Title = "Select DB to explore"
	listModel.InfiniteScrolling = false

	listModel.SetFilteringEnabled(false)
	listModel.SetShowFilter(false)
	listModel.SetShowStatusBar(false)

	listModel.KeyMap = list.KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "quit"),
		),
		ForceQuit: key.NewBinding(key.WithKeys("ctrl+c")),
	}
	listModel.AdditionalShortHelpKeys = func() []key.Binding {
		return additionalHelpKeyBindings
	}

	//
	// Make new DB prompt
	//

	promptModel := textinput.New()

	promptModel.Prompt = "New DB name > "

	promptModel.KeyMap = textinput.KeyMap{
		CharacterForward:        key.NewBinding(key.WithKeys("right")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left")),
		WordForward:             key.NewBinding(key.WithKeys("alt+right")),
		WordBackward:            key.NewBinding(key.WithKeys("alt+left")),
		DeleteWordBackward:      key.NewBinding(key.WithKeys("alt+backspace")),
		DeleteWordForward:       key.NewBinding(key.WithKeys("alt+delete")),
		DeleteCharacterBackward: key.NewBinding(key.WithKeys("backspace")),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),
	}

	//
	// Done
	//

	return &modelLoad{
		list:   listModel,
		prompt: promptModel,
	}
}

func (m *modelLoad) Init() tea.Cmd {
	return refreshDB
}

func (m *modelLoad) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msgt := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msgt.Width, msgt.Height-promptHeight)
		m.prompt.SetWidth(msgt.Width)
		return m, nil

	case tea.KeyPressMsg:
		if m.prompt.Focused() {
			switch msgt.Code {
			case tea.KeyEsc:
				m.prompt.Blur()
				return m, nil
			case tea.KeyEnter:
				if err := fileutils.ValidateName(m.prompt.Value()); err != nil {
					return m, m.list.NewStatusMessage(err.Error())
				}
				newDB, err := db.Create(fileutils.GetDBPath(m.prompt.Value()))
				if err != nil {
					return m, m.list.NewStatusMessage(err.Error())
				}

				m.prompt.Blur()
				return m.parent, loadDBCmd(newDB)
			}

			m.prompt, cmd = m.prompt.Update(msg)
		} else {
			switch msgt.Code {
			case tea.KeyEsc:
				return m.parent, loadDBCmd(nil)
			case tea.KeyEnter:
				loadDB, err := db.Open(fileutils.GetDBPath(m.list.SelectedItem().FilterValue()))
				if err != nil {
					return m, m.list.NewStatusMessage(err.Error())
				}
				return m.parent, loadDBCmd(loadDB)

			case 'n':
				m.prompt.Reset()
				m.prompt.Focus()
				return m, nil
			}

			m.list, cmd = m.list.Update(msg)
		}

	case refreshDBMsg:
		dbs, err := fileutils.ExistingDBs()
		if err != nil {
			return m, m.list.NewStatusMessage(fmt.Sprintf("error loading DBs: %s", err.Error()))
		}

		items := make([]list.Item, len(dbs))
		for i, n := range dbs {
			items[i] = item(n)
		}

		return m, m.list.SetItems(items)
	}

	return m, cmd
}

func (m *modelLoad) View() tea.View {
	var promptText string
	if m.prompt.Focused() {
		promptText = m.prompt.View()
	} else {
		promptText = ""
	}

	v := tea.NewView(lipgloss.JoinVertical(lipgloss.Top,
		m.list.View(),
		promptText))
	v.AltScreen = true

	return v
}

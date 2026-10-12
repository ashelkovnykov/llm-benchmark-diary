package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// openMsg asks the root model to open an existing database.
type openMsg struct{ name string }

// createMsg asks the root model to begin creating a new database.
type createMsg struct{}

// dbItem is a selectable existing database.
type dbItem string

func (d dbItem) Title() string       { return string(d) }
func (d dbItem) Description() string { return "open existing database" }
func (d dbItem) FilterValue() string { return string(d) }

// createItem is the sentinel entry that starts database creation.
type createItem struct{}

func (createItem) Title() string       { return "+ Create new database" }
func (createItem) Description() string { return "start a fresh database" }
func (createItem) FilterValue() string { return "create new database" }

type chooserModel struct {
	list list.Model
}

func newChooser(names []string) chooserModel {
	items := make([]list.Item, 0, len(names)+1)
	for _, n := range names {
		items = append(items, dbItem(n))
	}
	items = append(items, createItem{})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Select a database"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	return chooserModel{list: l}
}

func (c chooserModel) Init() tea.Cmd { return nil }

func (c chooserModel) Update(msg tea.Msg) (chooserModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok && c.list.FilterState() != list.Filtering {
		if key.String() == "enter" {
			switch it := c.list.SelectedItem().(type) {
			case createItem:
				return c, func() tea.Msg { return createMsg{} }
			case dbItem:
				return c, func() tea.Msg { return openMsg{name: string(it)} }
			}
		}
	}

	var cmd tea.Cmd
	c.list, cmd = c.list.Update(msg)
	return c, cmd
}

func (c *chooserModel) setSize(w, h int) {
	if h > 2 {
		h -= 2
	}
	c.list.SetSize(w, h)
}

func (c chooserModel) View() string {
	return c.list.View()
}

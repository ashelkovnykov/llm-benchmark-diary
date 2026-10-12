package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbletea"
)

type nameInputModel struct {
	prompt      string
	value       string
	cursorPos   int
	errorMsg    string
	validateFn  func(string) (bool, string)
	onSubmit    func(string)
	onCancel    func()
	readFromTTY bool
}

func NewNameInput(prompt string, validateFn func(string) (bool, string), onSubmit func(string), onCancel func()) tea.Model {
	return nameInputModel{
		prompt:     prompt,
		validateFn: validateFn,
		onSubmit:   onSubmit,
		onCancel:   onCancel,
	}
}

func (m nameInputModel) Init() tea.Cmd {
	return func() tea.Msg {
		if !m.readFromTTY {
			return nil
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Print(m.prompt)
		name, _ := reader.ReadString('\n')
		return nameInputMsg(strings.TrimSuffix(name, "\n"))
	}
}

type nameInputMsg string

func (m nameInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case nameInputMsg:
		m.value = string(msg)
		ok, errMsg := m.validateFn(m.value)
		if !ok {
			m.errorMsg = errMsg
			fmt.Printf("Error: %s\n", errMsg)
			fmt.Print(m.prompt)
			return m, func() tea.Msg {
				reader := bufio.NewReader(os.Stdin)
				name, _ := reader.ReadString('\n')
				return nameInputMsg(strings.TrimSuffix(name, "\n"))
			}
		}
		m.onSubmit(m.value)
		return m, tea.Quit
	}
	return m, nil
}

func (m nameInputModel) View() string {
	return ""
}

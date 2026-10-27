package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dbtui/internal/tui"
)

func main() {
	model, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	finalModel, err := tea.NewProgram(model, tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if m, ok := finalModel.(tui.Model); ok {
		m.Close()
	}
}

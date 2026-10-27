package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"example.com/dbtui/internal/tui"
)

func main() {
	model, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(model).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

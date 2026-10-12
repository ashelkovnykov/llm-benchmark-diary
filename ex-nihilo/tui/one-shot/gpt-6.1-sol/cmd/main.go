package main

import (
	"fmt"
	"os"

	"dbnav/internal/tui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	model, err := tui.New("./db")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	_, runErr := tea.NewProgram(model, tea.WithAltScreen()).Run()
	closeErr := model.Close()

	if runErr != nil {
		fmt.Fprintln(os.Stderr, runErr)
		os.Exit(1)
	}
	if closeErr != nil {
		fmt.Fprintln(os.Stderr, closeErr)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dbnav/internal/db"
	"dbnav/internal/tui"
)

func main() {
	const directory = "./db"

	names, err := db.List(directory)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(
		tui.New(directory, names),
		tea.WithAltScreen(),
	).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

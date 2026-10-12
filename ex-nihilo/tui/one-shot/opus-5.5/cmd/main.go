package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/example/sqlnav/internal/tui"
)

const dbDir = "./db"

func main() {
	m, err := tui.New(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	final, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if fm, ok := final.(tui.Model); ok {
		fm.Close()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/db"
	"sqlnav/internal/tui"
)

func main() {
	if err := db.EnsureDir(); err != nil {
		fmt.Fprintln(os.Stderr, "failed to prepare db directory:", err)
		os.Exit(1)
	}

	app, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

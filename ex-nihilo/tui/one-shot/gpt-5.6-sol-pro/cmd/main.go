package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dbtui/internal/tui"
)

func main() {
	m, err := tui.New()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

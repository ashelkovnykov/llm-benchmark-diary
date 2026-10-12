package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"dbtui/internal/db"
	"dbtui/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.New(db.Dir), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

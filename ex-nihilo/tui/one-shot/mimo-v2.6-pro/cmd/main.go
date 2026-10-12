package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.New(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

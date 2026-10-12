package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"sqlnav/internal/tui"
)

func main() {
	if err := os.MkdirAll("./db", 0755); err != nil {
		fmt.Fprintln(os.Stderr, "create db directory:", err)
		os.Exit(1)
	}

	if _, err := tea.NewProgram(tui.NewModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "run TUI:", err)
		os.Exit(1)
	}
}

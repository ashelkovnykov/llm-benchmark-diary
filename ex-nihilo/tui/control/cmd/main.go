package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"

	"control/internal/tui"
)

func main() {
	if _, err := tea.NewProgram(tui.InitialModel()).Run(); err != nil {
		log.Printf("error: %s", err.Error())
		os.Exit(1)
	}
}

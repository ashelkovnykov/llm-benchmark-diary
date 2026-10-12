package ui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FFA9")).
				Bold(true)

	dimmedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			MarginTop(1)
)

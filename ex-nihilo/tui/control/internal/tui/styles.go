package tui

import "charm.land/lipgloss/v2"

var (
	colorPrimary = lipgloss.Color("#00D170")
	colorAccent  = lipgloss.Color("#00FF88")
	colorDim     = lipgloss.Color("#004726")
	colorError   = lipgloss.Color("#004547")
)

var (
	textStyle = lipgloss.NewStyle().
			Foreground(colorPrimary)

	titleStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	cellStyle = lipgloss.NewStyle().
			Foreground(colorAccent).
			Width(3).
			MaxWidth(3)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorError).
			Bold(true)
)

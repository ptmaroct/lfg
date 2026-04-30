package tui

import "github.com/charmbracelet/lipgloss"

// Color palette — pulled to constants so a theme switch is a one-file change.
var (
	colPrimary   = lipgloss.Color("#7D56F4") // charm purple
	colAccent    = lipgloss.Color("#04B575") // success green
	colMuted     = lipgloss.Color("#626262")
	colWarn      = lipgloss.Color("#FFBB33")
	colError     = lipgloss.Color("#FF5555")
	colInstalled = lipgloss.Color("#04B575")
)

// Styles reused across screens.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colPrimary).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colMuted).
			Italic(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colPrimary).
			Padding(1, 2)

	footerStyle = lipgloss.NewStyle().
			Foreground(colMuted).
			MarginTop(1)

	installedStyle = lipgloss.NewStyle().Foreground(colInstalled)
	missingStyle   = lipgloss.NewStyle().Foreground(colMuted).Italic(true)
	warnStyle      = lipgloss.NewStyle().Foreground(colWarn).Bold(true)
	errorStyle     = lipgloss.NewStyle().Foreground(colError).Bold(true)
	accentStyle    = lipgloss.NewStyle().Foreground(colAccent).Bold(true)
)

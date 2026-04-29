package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorSuccess = lipgloss.Color("#2ECC71")
	ColorError   = lipgloss.Color("#E74C3C")
	ColorWarning = lipgloss.Color("#F1C40F")
	ColorInfo    = lipgloss.Color("#3498DB")
	ColorMuted   = lipgloss.Color("#95A5A6")
	ColorWhite   = lipgloss.Color("#FFFFFF")

	// Styles
	SuccessStyle = lipgloss.NewStyle().Foreground(ColorSuccess)
	ErrorStyle   = lipgloss.NewStyle().Foreground(ColorError)
	WarningStyle = lipgloss.NewStyle().Foreground(ColorWarning)
	InfoStyle    = lipgloss.NewStyle().Foreground(ColorInfo)
	MutedStyle   = lipgloss.NewStyle().Foreground(ColorMuted)
	BoldStyle    = lipgloss.NewStyle().Bold(true)
	TitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(ColorInfo)

	// Table styles
	TableHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(ColorWhite)
)

// Checkmark returns a green checkmark
func Checkmark() string {
	return SuccessStyle.Render("✓")
}

// Cross returns a red cross
func Cross() string {
	return ErrorStyle.Render("✗")
}

// Success renders text in green
func Success(text string) string {
	return SuccessStyle.Render(text)
}

// Error renders text in red
func Error(text string) string {
	return ErrorStyle.Render(text)
}

// Warning renders text in yellow
func Warning(text string) string {
	return WarningStyle.Render(text)
}

// Info renders text in blue
func Info(text string) string {
	return InfoStyle.Render(text)
}

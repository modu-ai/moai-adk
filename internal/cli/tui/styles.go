// Package tui provides Bubble Tea TUI components for MoAI-ADK.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Color constants for MoAI branding.
const (
	// ColorPrimary is the primary cyan color.
	ColorPrimary = "#0475FF"
	// ColorSuccess is the green success color.
	ColorSuccess = "#00F2EA"
	// ColorError is the red error color.
	ColorError = "#FF6B6B"
	// ColorWarning is the yellow warning color.
	ColorWarning = "#FDFF8C"
	// ColorMuted is the gray muted color.
	ColorMuted = "#626262"
	// ClaudeTerraCotta is the Claude terra cotta color.
	ClaudeTerraCotta = "#DA7756"
)

// StyleSet contains all lipgloss styles for the wizard TUI.
type StyleSet struct {
	// Title style for main titles
	title lipgloss.Style
	// Subtitle style for secondary text
	subtitle lipgloss.Style
	// Question style for prompts
	question lipgloss.Style
	// Input style for text input
	input lipgloss.Style
	// Instruction style for help text
	instruction lipgloss.Style
	// Error style for error messages
	error lipgloss.Style
	// Success title style
	successTitle lipgloss.Style
	// Border style for containers
	border lipgloss.Style
}

// styles is the global style set.
var styles = &StyleSet{
	title: lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Bold(true).
		MarginBottom(1),

	subtitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		MarginBottom(1),

	question: lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Bold(true).
		MarginBottom(1),

	input: lipgloss.NewStyle().
		Foreground(lipgloss.Color("white")).
		Background(lipgloss.Color(ColorMuted)).
		Padding(0, 1),

	instruction: lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Faint(true),

	error: lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorError)).
		Bold(true),

	successTitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSuccess)).
		Bold(true).
		MarginBottom(1),

	border: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorPrimary)).
		Padding(1, 2),
}

// Helper function to find minimum of two integers.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

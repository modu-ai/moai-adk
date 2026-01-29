package output

import (
	"github.com/charmbracelet/lipgloss"
)

// Color scheme constants
const (
	ColorPrimary    = "#7D56F4" // Purple
	ColorSuccess    = "#4ADE80" // Green
	ColorWarning    = "#FBBF24" // Yellow
	ColorError      = "#F87171" // Red
	ColorInfo       = "#60A5FA" // Blue
	ColorMuted      = "#9CA3AF" // Gray
	ColorBackground = "#1F2937" // Dark gray
)

// Style definitions
var (
	// HeaderStyle for main headers
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorPrimary)).
			MarginTop(1).
			MarginBottom(1)

	// SuccessStyle for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true)

	// WarningStyle for warning messages
	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorWarning)).
			Bold(true)

	// ErrorStyle for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true)

	// InfoStyle for informational messages
	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorInfo))

	// MutedStyle for subtle text
	MutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted))

	// CheckmarkStyle for checkmark indicators
	CheckmarkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)).
			Bold(true).
			PaddingRight(1)

	// CrossmarkStyle for cross indicators
	CrossmarkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true).
			PaddingRight(1)

	// BulletStyle for list items
	BulletStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			PaddingLeft(2)
)

// Styles for tables
var (
	// TableHeaderStyle for table headers
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color(ColorPrimary)).
				Padding(0, 1)

	// TableCellStyle for table cells
	TableCellStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// TableBorderStyle for table borders
	TableBorderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorMuted))
)

// Styles for sections
var (
	// SectionStyle for section headers
	SectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorInfo)).
			MarginTop(1).
			MarginBottom(0)

	// SubsectionStyle for subsection headers
	SubsectionStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorPrimary)).
			MarginLeft(2).
			MarginBottom(0)
)

// CodeStyle for code snippets
var (
	// CodeStyle for inline code
	CodeStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimary)).
		Background(lipgloss.Color(ColorBackground)).
		Padding(0, 1)
)

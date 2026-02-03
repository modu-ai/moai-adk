package wizard

import "github.com/charmbracelet/lipgloss"

// MoAI brand colors
const (
	// ColorPrimary is the MoAI brand orange color.
	ColorPrimary = "#DA7756"
	// ColorSecondary is a complementary color for highlights.
	ColorSecondary = "#7C3AED"
	// ColorSuccess is used for success messages.
	ColorSuccess = "#10B981"
	// ColorWarning is used for warning messages.
	ColorWarning = "#F59E0B"
	// ColorError is used for error messages.
	ColorError = "#EF4444"
	// ColorMuted is used for less important text.
	ColorMuted = "#6B7280"
	// ColorText is the default text color.
	ColorText = "#F9FAFB"
	// ColorBorder is used for borders and dividers.
	ColorBorder = "#4B5563"
)

// Styles holds all lipgloss styles used by the wizard.
type Styles struct {
	// Title style for question headers
	Title lipgloss.Style
	// Description style for question descriptions
	Description lipgloss.Style
	// Progress style for progress indicator (e.g., "[1/7]")
	Progress lipgloss.Style
	// Option style for unselected options
	Option lipgloss.Style
	// SelectedOption style for the currently selected option
	SelectedOption lipgloss.Style
	// Cursor style for the selection cursor
	Cursor lipgloss.Style
	// Input style for text input field
	Input lipgloss.Style
	// Placeholder style for input placeholders
	Placeholder lipgloss.Style
	// Error style for error messages
	Error lipgloss.Style
	// Success style for success messages
	Success lipgloss.Style
	// Muted style for less important text
	Muted lipgloss.Style
	// Help style for help text
	Help lipgloss.Style
	// Border style for containers
	Border lipgloss.Style
}

// NewStyles creates a new Styles instance with MoAI branding.
func NewStyles() *Styles {
	return &Styles{
		Title: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true).
			MarginBottom(1),

		Description: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Italic(true),

		Progress: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondary)).
			Bold(true),

		Option: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)),

		SelectedOption: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true),

		Cursor: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimary)).
			Bold(true),

		Input: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorText)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(0, 1),

		Placeholder: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			Italic(true),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)),

		Success: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSuccess)),

		Muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)),

		Help: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorMuted)).
			MarginTop(1),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ColorBorder)).
			Padding(1, 2),
	}
}

// NoColorStyles creates a Styles instance with no color formatting.
// Used when color output is disabled.
func NoColorStyles() *Styles {
	return &Styles{
		Title:          lipgloss.NewStyle().Bold(true),
		Description:    lipgloss.NewStyle(),
		Progress:       lipgloss.NewStyle().Bold(true),
		Option:         lipgloss.NewStyle(),
		SelectedOption: lipgloss.NewStyle().Bold(true),
		Cursor:         lipgloss.NewStyle().Bold(true),
		Input:          lipgloss.NewStyle(),
		Placeholder:    lipgloss.NewStyle(),
		Error:          lipgloss.NewStyle(),
		Success:        lipgloss.NewStyle(),
		Muted:          lipgloss.NewStyle(),
		Help:           lipgloss.NewStyle(),
		Border:         lipgloss.NewStyle(),
	}
}

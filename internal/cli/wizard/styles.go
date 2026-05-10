package wizard

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// wizardColors returns the adaptive brand colors derived from internal/tui tokens.
// All hex values originate from tui.LightTheme() / tui.DarkTheme() exclusively
// (AC-CLI-TUI-013: no hex literals outside internal/tui/).
func wizardColors() wizardColorSet {
	lt := tui.LightTheme()
	dt := tui.DarkTheme()
	return wizardColorSet{
		// Primary: deep teal accent (replaces terra cotta #DA7756 / #C45A3C)
		Primary: lipgloss.AdaptiveColor{Light: lt.Accent, Dark: dt.Accent},
		// Secondary: info color (replaces purple #7C3AED / #5B21B6)
		Secondary: lipgloss.AdaptiveColor{Light: lt.Info, Dark: dt.Info},
		// Success / error / muted / border from tui tokens
		Success: lipgloss.AdaptiveColor{Light: lt.Success, Dark: dt.Success},
		Warning: lipgloss.AdaptiveColor{Light: lt.Warning, Dark: dt.Warning},
		Error:   lipgloss.AdaptiveColor{Light: lt.Danger, Dark: dt.Danger},
		Muted:   lipgloss.AdaptiveColor{Light: lt.Dim, Dark: dt.Dim},
		Text:    lipgloss.AdaptiveColor{Light: lt.Fg, Dark: dt.Fg},
		Border:  lipgloss.AdaptiveColor{Light: lt.Rule, Dark: dt.Rule},
		// Button foreground (white text on filled button)
		ButtonFg: lipgloss.AdaptiveColor{Light: lt.Bg, Dark: dt.Fg},
		// Button blurred background
		ButtonBlurredBg: lipgloss.AdaptiveColor{Light: lt.Panel, Dark: dt.Panel},
	}
}

// wizardColorSet holds resolved adaptive colors for the wizard theme.
type wizardColorSet struct {
	Primary         lipgloss.AdaptiveColor
	Secondary       lipgloss.AdaptiveColor
	Success         lipgloss.AdaptiveColor
	Warning         lipgloss.AdaptiveColor
	Error           lipgloss.AdaptiveColor
	Muted           lipgloss.AdaptiveColor
	Text            lipgloss.AdaptiveColor
	Border          lipgloss.AdaptiveColor
	ButtonFg        lipgloss.AdaptiveColor
	ButtonBlurredBg lipgloss.AdaptiveColor
}

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
	c := wizardColors()
	return &Styles{
		Title: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true).
			MarginBottom(1),

		Description: lipgloss.NewStyle().
			Foreground(c.Muted).
			Italic(true),

		Progress: lipgloss.NewStyle().
			Foreground(c.Secondary).
			Bold(true),

		Option: lipgloss.NewStyle().
			Foreground(c.Text),

		SelectedOption: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true),

		Cursor: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true),

		Input: lipgloss.NewStyle().
			Foreground(c.Text).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(c.Border).
			Padding(0, 1),

		Placeholder: lipgloss.NewStyle().
			Foreground(c.Muted).
			Italic(true),

		Error: lipgloss.NewStyle().
			Foreground(c.Error),

		Success: lipgloss.NewStyle().
			Foreground(c.Success),

		Muted: lipgloss.NewStyle().
			Foreground(c.Muted),

		Help: lipgloss.NewStyle().
			Foreground(c.Muted).
			MarginTop(1),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(c.Border).
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

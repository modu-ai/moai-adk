// Package ui provides terminal UI components for MoAI-ADK Go Edition.
// It includes a theme system, interactive selectors, checkboxes, prompts,
// progress bars, and a multi-step wizard, all built on the Charmbracelet
// ecosystem (lipgloss for styling, bubbletea for Elm Architecture).
package ui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// ThemeConfig holds configuration for creating a Theme.
type ThemeConfig struct {
	// Mode is the color mode: "dark", "light", "auto", or "" (defaults to dark).
	Mode string
	// NoColor disables all color and styling when true.
	NoColor bool
}

// ColorPalette defines the color values used by theme rendering functions.
type ColorPalette struct {
	Primary   string
	Secondary string
	Success   string
	Warning   string
	Error     string
	Muted     string
	Text      string
	Border    string
}

// Theme provides consistent lipgloss-based styling for all UI components.
// All rendering goes through lipgloss; ANSI escape codes are never hardcoded.
type Theme struct {
	Colors  ColorPalette
	IsDark  bool
	NoColor bool

	// Pre-built lipgloss styles for rendering helpers.
	titleStyle     lipgloss.Style
	errorStyle     lipgloss.Style
	successStyle   lipgloss.Style
	warningStyle   lipgloss.Style
	mutedStyle     lipgloss.Style
	highlightStyle lipgloss.Style
}

// darkPalette returns the color palette optimized for dark backgrounds.
func darkPalette() ColorPalette {
	return ColorPalette{
		Primary:   "#7C3AED",
		Secondary: "#06B6D4",
		Success:   "#10B981",
		Warning:   "#F59E0B",
		Error:     "#EF4444",
		Muted:     "#6B7280",
		Text:      "#F9FAFB",
		Border:    "#4B5563",
	}
}

// lightPalette returns the color palette optimized for light backgrounds.
func lightPalette() ColorPalette {
	return ColorPalette{
		Primary:   "#5B21B6",
		Secondary: "#0891B2",
		Success:   "#059669",
		Warning:   "#D97706",
		Error:     "#DC2626",
		Muted:     "#9CA3AF",
		Text:      "#111827",
		Border:    "#D1D5DB",
	}
}

// NewTheme creates a Theme from the given configuration.
// It respects the MOAI_NO_COLOR environment variable and selects
// a dark or light palette based on the configured mode.
func NewTheme(cfg ThemeConfig) *Theme {
	noColor := cfg.NoColor
	if !noColor {
		if env := os.Getenv("MOAI_NO_COLOR"); env == "true" || env == "1" {
			noColor = true
		}
	}

	isDark := resolveDarkMode(cfg.Mode)

	var palette ColorPalette
	if isDark {
		palette = darkPalette()
	} else {
		palette = lightPalette()
	}

	t := &Theme{
		Colors:  palette,
		IsDark:  isDark,
		NoColor: noColor,
	}

	if noColor {
		// In NoColor mode, all styles are empty (no formatting).
		t.titleStyle = lipgloss.NewStyle()
		t.errorStyle = lipgloss.NewStyle()
		t.successStyle = lipgloss.NewStyle()
		t.warningStyle = lipgloss.NewStyle()
		t.mutedStyle = lipgloss.NewStyle()
		t.highlightStyle = lipgloss.NewStyle()
	} else {
		t.titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Primary)).
			Bold(true)
		t.errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Error))
		t.successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Success))
		t.warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Warning))
		t.mutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Muted))
		t.highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette.Secondary)).
			Bold(true)
	}

	return t
}

// resolveDarkMode determines whether dark mode should be used.
func resolveDarkMode(mode string) bool {
	switch mode {
	case "light":
		return false
	case "auto":
		// In non-TTY environments lipgloss.HasDarkBackground may not work
		// reliably; default to dark when detection is unavailable.
		return lipgloss.HasDarkBackground()
	default:
		// "dark" or empty string both default to dark mode.
		return true
	}
}

// RenderTitle renders text with the title style (primary color, bold).
func (t *Theme) RenderTitle(text string) string {
	if t.NoColor {
		return text
	}
	return t.titleStyle.Render(text)
}

// RenderError renders text with the error style.
func (t *Theme) RenderError(text string) string {
	if t.NoColor {
		return text
	}
	return t.errorStyle.Render(text)
}

// RenderSuccess renders text with the success style.
func (t *Theme) RenderSuccess(text string) string {
	if t.NoColor {
		return text
	}
	return t.successStyle.Render(text)
}

// RenderWarning renders text with the warning style.
func (t *Theme) RenderWarning(text string) string {
	if t.NoColor {
		return text
	}
	return t.warningStyle.Render(text)
}

// RenderMuted renders text with the muted style.
func (t *Theme) RenderMuted(text string) string {
	if t.NoColor {
		return text
	}
	return t.mutedStyle.Render(text)
}

// RenderHighlight renders text with the highlight style (secondary color, bold).
func (t *Theme) RenderHighlight(text string) string {
	if t.NoColor {
		return text
	}
	return t.highlightStyle.Render(text)
}

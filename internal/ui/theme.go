// Package ui provides terminal UI components for MoAI-ADK Go Edition.
// It includes a theme system, interactive selectors, checkboxes, prompts,
// progress bars, and a multi-step wizard, all built on the Charmbracelet
// ecosystem (lipgloss for styling, bubbletea for Elm Architecture).
package ui

import (
	"os"

	"github.com/charmbracelet/huh"
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

// NewMoAIHuhTheme creates a huh.Theme styled with MoAI branding.
// It uses AdaptiveColor for automatic light/dark mode support.
// Pass noColor=true to return a plain unstyled theme for headless mode.
func NewMoAIHuhTheme(noColor bool) *huh.Theme {
	if noColor {
		return huh.ThemeBase()
	}

	t := huh.ThemeBase()

	var (
		// MoAI brand orange for focused/selected elements.
		primary = lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: "#DA7756"}
		// Purple for highlights and selectors.
		secondary = lipgloss.AdaptiveColor{Light: "#5B21B6", Dark: "#7C3AED"}
		// Success green.
		green = lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"}
		// Error red.
		red = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}
		// Normal text.
		text = lipgloss.AdaptiveColor{Light: "#111827", Dark: "#F9FAFB"}
		// Muted/description text.
		muted = lipgloss.AdaptiveColor{Light: "#9CA3AF", Dark: "#6B7280"}
		// Border color.
		border = lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"}
		// Button background for blurred state.
		btnBg = lipgloss.AdaptiveColor{Light: "#E5E7EB", Dark: "#374151"}
	)

	// Focused field styles.
	t.Focused.Base = t.Focused.Base.BorderForeground(border)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(primary).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(primary).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(muted)
	t.Focused.Directory = t.Focused.Directory.Foreground(secondary)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(red)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(red)

	// Select styles.
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(primary).SetString("▸ ")
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(primary)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(primary)
	t.Focused.Option = t.Focused.Option.Foreground(text)

	// Multi-select styles.
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(primary)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(green)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(green).SetString("◆ ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(text)
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(muted).SetString("◇ ")

	// Text input styles.
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(primary)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(muted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(secondary)

	// Confirm button styles.
	t.Focused.FocusedButton = t.Focused.FocusedButton.
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(primary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.
		Foreground(text).Background(btnBg)
	t.Focused.Next = t.Focused.FocusedButton

	// Blurred styles inherit from Focused with hidden border.
	t.Blurred = t.Focused
	t.Blurred.Base = t.Focused.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	// Group styles.
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}

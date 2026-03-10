package statusline

import "github.com/charmbracelet/lipgloss"

// Theme defines the color palette for statusline rendering.
// Each implementation provides a consistent color scheme.
type Theme interface {
	// Primary returns the primary accent color.
	Primary() lipgloss.Color
	// Muted returns the muted/secondary text color.
	Muted() lipgloss.Color
	// Success returns the success/green color.
	Success() lipgloss.Color
	// Warning returns the warning/yellow color.
	Warning() lipgloss.Color
	// Danger returns the danger/red color.
	Danger() lipgloss.Color
	// Text returns the primary text color.
	Text() lipgloss.Color
	// BarGradient returns a color for the context bar based on usage percentage.
	// Stages: 0-25% (green), 26-50% (yellow), 51-75% (peach/orange), 76-100% (red).
	//
	// Deprecated: v3 uses continuous RGB interpolation via BuildGradientBar().
	// Retained for backward compatibility.
	BarGradient(percentage float64) lipgloss.Color
}

// NewTheme returns a Theme implementation for the given name.
// Falls back to catppuccinMocha for unknown names (REQ-SLE-012, REQ-NF-007).
func NewTheme(name string) Theme {
	switch name {
	case "catppuccin-latte":
		return &catppuccinLatte{}
	default:
		return &catppuccinMocha{}
	}
}

// catppuccinMocha implements the Catppuccin Mocha dark theme (REQ-SLE-013).
type catppuccinMocha struct{}

func (m *catppuccinMocha) Primary() lipgloss.Color { return lipgloss.Color("#CDD6F4") }
func (m *catppuccinMocha) Muted() lipgloss.Color   { return lipgloss.Color("#6C7086") }
func (m *catppuccinMocha) Success() lipgloss.Color { return lipgloss.Color("#A6E3A1") }
func (m *catppuccinMocha) Warning() lipgloss.Color { return lipgloss.Color("#F9E2AF") }
func (m *catppuccinMocha) Danger() lipgloss.Color  { return lipgloss.Color("#F38BA8") }
func (m *catppuccinMocha) Text() lipgloss.Color    { return lipgloss.Color("#CDD6F4") }

// BarGradient returns a 4-stage gradient using Catppuccin Mocha palette (REQ-SLE-015).
func (m *catppuccinMocha) BarGradient(pct float64) lipgloss.Color {
	switch {
	case pct <= 25:
		return lipgloss.Color("#A6E3A1") // Green
	case pct <= 50:
		return lipgloss.Color("#F9E2AF") // Yellow
	case pct <= 75:
		return lipgloss.Color("#FAB387") // Peach
	default:
		return lipgloss.Color("#F38BA8") // Red
	}
}

// catppuccinLatte implements the Catppuccin Latte light theme (REQ-SLE-014).
type catppuccinLatte struct{}

func (l *catppuccinLatte) Primary() lipgloss.Color { return lipgloss.Color("#4C4F69") }
func (l *catppuccinLatte) Muted() lipgloss.Color   { return lipgloss.Color("#9CA0B0") }
func (l *catppuccinLatte) Success() lipgloss.Color { return lipgloss.Color("#40A02B") }
func (l *catppuccinLatte) Warning() lipgloss.Color { return lipgloss.Color("#DF8E1D") }
func (l *catppuccinLatte) Danger() lipgloss.Color  { return lipgloss.Color("#D20F39") }
func (l *catppuccinLatte) Text() lipgloss.Color    { return lipgloss.Color("#4C4F69") }

// BarGradient returns a 4-stage gradient using Catppuccin Latte palette (REQ-SLE-015).
func (l *catppuccinLatte) BarGradient(pct float64) lipgloss.Color {
	switch {
	case pct <= 25:
		return lipgloss.Color("#40A02B") // Green
	case pct <= 50:
		return lipgloss.Color("#DF8E1D") // Yellow
	case pct <= 75:
		return lipgloss.Color("#FE640B") // Peach
	default:
		return lipgloss.Color("#D20F39") // Red
	}
}

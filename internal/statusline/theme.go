package statusline

import (
	"github.com/charmbracelet/lipgloss"
	tuipkg "github.com/modu-ai/moai-adk/internal/tui"
)

// Theme defines the color palette for statusline rendering.
// Each implementation provides a consistent color scheme.
//
// @MX:NOTE: [AUTO] M6-S4 R-07 thin wrapper — 자체 hex 상수 0건; 색상 값은 internal/tui/catppuccin.go에서 import
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
//
// @MX:ANCHOR: [AUTO] Theme 생성 진입점; fan_in >= 3 (renderer.go NewRenderer, theme_test.go, builder.go 경유)
// @MX:REASON: [AUTO] statusline 패키지 공개 API 경계; catppuccin 팔레트를 tui.Catppuccin* 상수로 thin-re-export
func NewTheme(name string) Theme {
	switch name {
	case "catppuccin-latte":
		return &catppuccinLatte{}
	default:
		return &catppuccinMocha{}
	}
}

// catppuccinMocha implements the Catppuccin Mocha dark theme (REQ-SLE-013).
// 색상 값은 internal/tui/catppuccin.go에서 thin-import (R-07 mitigation).
type catppuccinMocha struct{}

func (m *catppuccinMocha) Primary() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaPrimary)
}
func (m *catppuccinMocha) Muted() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaMuted)
}
func (m *catppuccinMocha) Success() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaSuccess)
}
func (m *catppuccinMocha) Warning() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaWarning)
}
func (m *catppuccinMocha) Danger() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaDanger)
}
func (m *catppuccinMocha) Text() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinMochaPrimary)
}

// BarGradient returns a 4-stage gradient using Catppuccin Mocha palette (REQ-SLE-015).
func (m *catppuccinMocha) BarGradient(pct float64) lipgloss.Color {
	switch {
	case pct <= 25:
		return lipgloss.Color(tuipkg.CatppuccinMochaSuccess) // Green
	case pct <= 50:
		return lipgloss.Color(tuipkg.CatppuccinMochaWarning) // Yellow
	case pct <= 75:
		return lipgloss.Color(tuipkg.CatppuccinMochaPeach) // Peach
	default:
		return lipgloss.Color(tuipkg.CatppuccinMochaDanger) // Red
	}
}

// catppuccinLatte implements the Catppuccin Latte light theme (REQ-SLE-014).
// 색상 값은 internal/tui/catppuccin.go에서 thin-import (R-07 mitigation).
type catppuccinLatte struct{}

func (l *catppuccinLatte) Primary() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLattePrimary)
}
func (l *catppuccinLatte) Muted() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLatteMuted)
}
func (l *catppuccinLatte) Success() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLatteSuccess)
}
func (l *catppuccinLatte) Warning() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLatteWarning)
}
func (l *catppuccinLatte) Danger() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLatteDanger)
}
func (l *catppuccinLatte) Text() lipgloss.Color {
	return lipgloss.Color(tuipkg.CatppuccinLattePrimary)
}

// BarGradient returns a 4-stage gradient using Catppuccin Latte palette (REQ-SLE-015).
func (l *catppuccinLatte) BarGradient(pct float64) lipgloss.Color {
	switch {
	case pct <= 25:
		return lipgloss.Color(tuipkg.CatppuccinLatteSuccess) // Green
	case pct <= 50:
		return lipgloss.Color(tuipkg.CatppuccinLatteWarning) // Yellow
	case pct <= 75:
		return lipgloss.Color(tuipkg.CatppuccinLattePeach) // Peach
	default:
		return lipgloss.Color(tuipkg.CatppuccinLatteDanger) // Red
	}
}

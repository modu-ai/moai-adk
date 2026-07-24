package uikit

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// CLI output styles sourced from tui.LightTheme/DarkTheme — single source of truth.
// AC-CLI-TUI-013: no hex literals outside internal/tui/. Uses AdaptiveColor with
// tui.LightTheme()/DarkTheme() values evaluated at package init.
//
// Moved from internal/cli/update.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). The render helpers depend on these styles; they
// are leaf (lipgloss + tui only), so they belong in the TUI kernel leaf.
var (
	SuccessStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Success, Dark: tui.DarkTheme().Success})
	WarnStyle    = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Warning, Dark: tui.DarkTheme().Warning})
	ErrorStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Danger, Dark: tui.DarkTheme().Danger})
	MutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Dim, Dark: tui.DarkTheme().Dim})
	PrimaryStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Accent, Dark: tui.DarkTheme().Accent})
	BorderStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Rule, Dark: tui.DarkTheme().Rule})
)

// SymSuccess returns a styled check mark for success indicators.
func SymSuccess() string { return SuccessStyle.Render(string(tui.GlyphDone)) }

// SymError returns a styled cross mark for error indicators.
func SymError() string { return ErrorStyle.Render(string(tui.GlyphErr)) }

// SymWarning returns a styled exclamation mark for warning indicators.
func SymWarning() string { return WarnStyle.Render(string(tui.GlyphWarn)) }

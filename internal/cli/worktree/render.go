package worktree

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// Worktree CLI styles use internal/tui design tokens (REQ-CLI-TUI-013).
// Theme is resolved at call-time via tui.ResolveOS() so it respects NO_COLOR
// and MOAI_THEME at runtime without a package-level init dependency.
func wtStyles() (primary, border, success lipgloss.Style) {
	th := tui.ResolveOS()
	primary = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: tui.LightTheme().Accent,
		Dark:  tui.DarkTheme().Accent,
	})
	_ = th // th informs future per-call style if needed
	border = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: tui.LightTheme().Rule,
		Dark:  tui.DarkTheme().Rule,
	})
	success = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: tui.LightTheme().Success,
		Dark:  tui.DarkTheme().Success,
	})
	return
}

// wtCardStyle returns a lipgloss style for a rounded-border card.
func wtCardStyle() lipgloss.Style {
	_, border, _ := wtStyles()
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border.GetForeground()).
		Padding(0, 2)
}

// wtCard renders content inside a rounded border box with a styled title.
func wtCard(title, content string) string {
	primary, _, _ := wtStyles()
	titleLine := primary.Bold(true).Render(title)
	body := titleLine + "\n\n" + content
	return wtCardStyle().Render(body)
}

// wtSuccessCard renders a success message inside a rounded border card.
func wtSuccessCard(title string, details ...string) string {
	_, _, success := wtStyles()
	titleLine := success.Render("\u2713") + " " + title
	var body strings.Builder
	body.WriteString(titleLine)
	if len(details) > 0 {
		body.WriteString("\n\n")
		for i, d := range details {
			if i > 0 {
				body.WriteString("\n")
			}
			body.WriteString(d)
		}
	}
	return wtCardStyle().Render(body.String())
}

// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Prompt renders a shell-style prompt line using design tokens.
// It is a pure function with no terminal I/O.
//
// Parameters:
//   - host: the user@hostname portion (e.g. "yuna@air")
//   - path: the working directory path (e.g. "~/work/my-app")
//   - branch: the git branch name (empty string to omit)
//   - dirty: indicates uncommitted changes (appends "*" to branch)
//   - cmd: the command to display after the prompt arrow (empty to omit)
//   - th: design theme; nil falls back to LightTheme
//
// Colours are sourced from Theme.PromptArrow and Theme.PromptPath tokens
// (REQ-CLI-TUI-013, no hex literals in this file).
//
// @MX:ANCHOR: [AUTO] Prompt is a core primitive; expected fan_in >= 3 in M4-M6
// @MX:REASON: doctor header, cc screen, and term chrome all display prompt lines
func Prompt(host, path, branch string, dirty bool, cmd string, th *Theme) string {
	t := LightTheme()
	if th != nil {
		t = *th
	}

	hostStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PromptArrow)).
		Bold(true)
	pathStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PromptPath))
	branchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Accent))
	arrowStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PromptArrow)).
		Bold(true)
	cmdStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Fg))

	// Build prompt: host path [branch[*]] > [cmd]
	result := hostStyle.Render(host) + " " + pathStyle.Render(path)

	if branch != "" {
		b := branch
		if dirty {
			b += "*"
		}
		result += " " + branchStyle.Render("("+b+")")
	}

	result += " " + arrowStyle.Render(">")

	if cmd != "" {
		result += " " + cmdStyle.Render(cmd)
	}

	return result
}

// Cursor renders a block cursor character styled with Theme.Cursor colour.
// It is a pure function with no terminal I/O.
// th may be nil; LightTheme is used in that case.
func Cursor(th *Theme) string {
	t := LightTheme()
	if th != nil {
		t = *th
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Cursor)).
		Bold(true)

	return style.Render("█")
}

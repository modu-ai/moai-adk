// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RadioOption represents a single radio button option.
type RadioOption struct {
	// Label is the display text for this option.
	Label string
	// Selected indicates this option is the currently selected one.
	Selected bool
}

// RadioRow renders a vertical list of radio button options.
// The selected item uses ◆ (U+25C6) as prefix; unselected items use ◇ (U+25C7).
// These are the only two radio prefixes used across the MoAI-ADK TUI (AC-CLI-TUI-004).
// th may be nil; LightTheme is used in that case.
//
// Returns an empty string when opts is nil or empty.
//
// @MX:ANCHOR: [AUTO] RadioRow is a core primitive; expected fan_in >= 3 in M5
// @MX:REASON: init wizard language/runtime/profile selection uses RadioRow per design source
func RadioRow(opts []RadioOption, th *Theme) string {
	if len(opts) == 0 {
		return ""
	}

	t := LightTheme()
	if th != nil {
		t = *th
	}

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Accent)).
		Bold(true)
	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Dim))
	selectedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Fg)).
		Bold(true)
	unselectedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Body))

	var lines []string
	for _, o := range opts {
		var row string
		if o.Selected {
			row = selectedStyle.Render("◆") + " " + selectedLabelStyle.Render(o.Label)
		} else {
			row = unselectedStyle.Render("◇") + " " + unselectedLabelStyle.Render(o.Label)
		}
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

// CheckOption represents a single checkbox option.
type CheckOption struct {
	// Label is the display text for this option.
	Label string
	// Checked indicates this option is currently checked.
	Checked bool
}

// CheckRow renders a vertical list of checkbox options.
// Checked items use ✓ (success colour); unchecked items use ○ (muted colour).
// th may be nil; LightTheme is used in that case.
//
// Returns an empty string when opts is nil or empty.
//
// @MX:ANCHOR: [AUTO] CheckRow is a core primitive; expected fan_in >= 3 in M5-M6
// @MX:REASON: init wizard optional features selection uses CheckRow per design source
func CheckRow(opts []CheckOption, th *Theme) string {
	if len(opts) == 0 {
		return ""
	}

	t := LightTheme()
	if th != nil {
		t = *th
	}

	checkedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Success)).
		Bold(true)
	uncheckedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Dim))
	checkedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Fg))
	uncheckedLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Body))

	var lines []string
	for _, o := range opts {
		var row string
		if o.Checked {
			row = checkedStyle.Render("✓") + " " + checkedLabelStyle.Render(o.Label)
		} else {
			row = uncheckedStyle.Render("○") + " " + uncheckedLabelStyle.Render(o.Label)
		}
		lines = append(lines, row)
	}

	return strings.Join(lines, "\n")
}

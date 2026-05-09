// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyHint represents a single keyboard shortcut hint for display in a HelpBar.
type KeyHint struct {
	// Key is the keyboard shortcut label (e.g. "enter", "↑↓", "ctrl+c").
	Key string
	// Label is the human-readable description (e.g. "선택", "이동", "quit").
	Label string
}

// HelpBar renders a horizontal row of keyboard shortcut hints.
// Each hint is formatted as: <key> <label>, separated by a dot divider.
//
// Returns an empty string when items is nil or empty.
// th may be nil; LightTheme is used in that case.
//
// Korean and mixed ko-en label text is handled correctly because lipgloss
// inherits terminal column width for ANSI-escaped strings.
//
// @MX:ANCHOR: [AUTO] HelpBar is a core primitive; expected fan_in >= 4 in M5-M6
// @MX:REASON: init wizard footer, loop screen, help command, and error surface all use HelpBar
func HelpBar(items []KeyHint, th *Theme) string {
	if len(items) == 0 {
		return ""
	}

	t := LightTheme()
	if th != nil {
		t = *th
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Accent)).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Dim))
	divStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Faint))

	var parts []string
	for _, h := range items {
		hint := keyStyle.Render(h.Key) + " " + labelStyle.Render(h.Label)
		parts = append(parts, hint)
	}

	div := divStyle.Render(" · ")
	return strings.Join(parts, div)
}

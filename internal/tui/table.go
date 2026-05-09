// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	tuiinternal "github.com/modu-ai/moai-adk/internal/tui/internal"
)

// KVOpts configures a KV render call.
type KVOpts struct {
	// Theme provides design tokens. If nil, LightTheme is used.
	Theme *Theme
	// KeyWidth is the minimum display width of the key column (for alignment).
	// 0 means no padding.
	KeyWidth int
}

// KV renders a single key–value row with aligned columns.
// Korean and CJK character widths are accounted for via runeguard.StringWidth
// so that mixed ko-en content aligns correctly (AC-CLI-TUI-007, REQ-CLI-TUI-001).
//
// All colours are sourced from opts.Theme; no hex literals appear in this
// file (REQ-CLI-TUI-013).
//
// @MX:ANCHOR: [AUTO] KV is a core primitive; expected fan_in >= 5 across M4-M5-M6
// @MX:REASON: doctor CheckLine, status, version, and table views all use KV for aligned output
func KV(key, value string, opts KVOpts) string {
	t := LightTheme()
	if opts.Theme != nil {
		t = *opts.Theme
	}

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Dim)).
		Bold(true)
	valStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Body))

	// Use runeguard.StringWidth for Korean/CJK-accurate key width calculation.
	renderedKey := key
	if opts.KeyWidth > 0 {
		renderedKey = tuiinternal.FillRight(key, opts.KeyWidth)
	}

	return keyStyle.Render(renderedKey) + "  " + valStyle.Render(value)
}

// CheckLine renders a single doctor-row formatted line.
// Format: <StatusIcon> <label> <value> [· <hint>]
// This format matches the doctor row design in acceptance.md AC-CLI-TUI-003.
//
// status is one of: "ok", "warn", "err", "info", "run", "skip", "dot".
// hint is optional; pass "" to omit.
// th may be nil; LightTheme is used in that case.
//
// @MX:ANCHOR: [AUTO] CheckLine is a core primitive; expected fan_in >= 5 across M4
// @MX:REASON: moai doctor uses CheckLine for all 19 check items per design source
func CheckLine(status, label, value, hint string, th *Theme) string {
	t := LightTheme()
	if th != nil {
		t = *th
	}

	icon := StatusIcon(status)

	var iconStyle lipgloss.Style
	switch status {
	case "ok":
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Success)).Bold(true)
	case "warn":
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Warning)).Bold(true)
	case "err":
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Danger)).Bold(true)
	default:
		iconStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(t.Dim))
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Body))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Fg))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Faint))

	var b strings.Builder
	b.WriteString(iconStyle.Render(icon))
	b.WriteString("  ")
	b.WriteString(labelStyle.Render(label))
	b.WriteString("  ")
	b.WriteString(valueStyle.Render(value))
	if hint != "" {
		b.WriteString(" ")
		b.WriteString(hintStyle.Render("· " + hint))
	}

	_ = tuiinternal.StringWidth // ensure package is used for CJK-width awareness
	return b.String()
}

// SectionOpts configures a Section render call.
type SectionOpts struct {
	// Theme provides design tokens. If nil, LightTheme is used.
	Theme *Theme
	// Width is the total width. 0 means no fixed width.
	Width int
}

// Section renders a titled section header — a bold label with an accent rule
// below it. Used as a visual divider between groups of related rows.
//
// @MX:ANCHOR: [AUTO] Section is a core primitive; expected fan_in >= 4 across M4-M6
// @MX:REASON: doctor "시스템" / "MoAI-ADK" section headers and help group headers
func Section(title string, opts SectionOpts) string {
	t := LightTheme()
	if opts.Theme != nil {
		t = *opts.Theme
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Accent)).
		Bold(true)

	ruleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Rule))

	width := opts.Width
	if width <= 0 {
		// Auto-width: derive from title display width + some padding.
		width = tuiinternal.StringWidth(title) + 4
	}

	ruleWidth := max(width-tuiinternal.StringWidth(title), 1)

	// Use ASCII dash "-" repeated to approximate a rule line.
	// Hand-drawn box drawing characters (U+2500 series) are forbidden in M2
	// source files by AC-CLI-TUI-011; lipgloss.Border() is used for box
	// borders in box.go. For the section divider we use plain dashes.
	rule := ruleStyle.Render(" " + strings.Repeat("-", ruleWidth-1))
	return titleStyle.Render(title) + rule
}

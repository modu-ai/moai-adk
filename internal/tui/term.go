// Package tui provides the MoAI-ADK terminal UI design system v2.
package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// TermOpts configures a Term render call.
type TermOpts struct {
	// Theme provides design tokens. If nil, LightTheme is used.
	Theme *Theme
}

// screenshotMode returns true when MOAI_SCREENSHOT is set to the exact value "1".
// Any other value (including "0", empty string) disables screenshot mode.
func screenshotMode() bool {
	return os.Getenv("MOAI_SCREENSHOT") == "1"
}

// Term renders a macOS-style window chrome panel (traffic-light buttons +
// title bar) wrapping the body content.
//
// Term returns an empty string unless MOAI_SCREENSHOT=1 is set in the
// environment. In live terminal mode (MOAI_SCREENSHOT unset or != "1"), the
// function is a no-op and returns "".
//
// This design allows CLI commands to include a Term call unconditionally and
// have it activate only in screenshot pipelines (e.g., automated CI captures).
//
// All colours are sourced from opts.Theme; no hex literals appear in this
// file (REQ-CLI-TUI-013). The chrome uses lipgloss.ThickBorder() via ThickBox
// to avoid hand-drawn box characters (AC-CLI-TUI-011).
//
// @MX:ANCHOR: [AUTO] Term is a screenshot-mode primitive; fan_in >= 2 in M4-M6
// @MX:REASON: automated screenshot pipelines for 14 CLI surfaces call Term
func Term(title string, width int, body string, opts TermOpts) string {
	if !screenshotMode() {
		return ""
	}

	t := LightTheme()
	if opts.Theme != nil {
		t = *opts.Theme
	}

	// Traffic-light buttons: close(●) minimise(●) maximise(●)
	// Colours approximate macOS colours using theme tokens.
	dangerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Danger))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Warning))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Success))
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Dim)).
		Bold(true)

	buttons := dangerStyle.Render("●") + " " +
		warnStyle.Render("●") + " " +
		successStyle.Render("●")

	titleBar := buttons + "  " + titleStyle.Render(title)

	// Render body inside the chrome frame using Chrome border colour.
	chromeBorder := t.ChromeBorder
	if chromeBorder == "" {
		chromeBorder = t.Rule
	}

	// Build the chrome frame using lipgloss.ThickBorder so we avoid
	// hand-drawn box characters (AC-CLI-TUI-011).
	border := lipgloss.ThickBorder()
	frameStyle := lipgloss.NewStyle().
		Border(border).
		BorderForeground(lipgloss.Color(chromeBorder)).
		Padding(1, 2).
		Foreground(lipgloss.Color(t.Fg))

	if width > 0 {
		innerW := max(width-2-4, 1)
		frameStyle = frameStyle.Width(innerW)
	}

	content := titleBar + "\n\n" + body
	return frameStyle.Render(content)
}

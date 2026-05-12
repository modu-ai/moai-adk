package cli

// @MX:NOTE: [AUTO] Rendering utilities for CLI output using lipgloss styling
// @MX:NOTE: [AUTO] Provides card, key-value, status, success/info card renderers

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// cardStyle returns a lipgloss style for a rounded-border card.
func cardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cliBorder.GetForeground()).
		Padding(0, 2)
}

// renderCard renders content inside a rounded border box with a styled title.
func renderCard(title, content string) string {
	titleLine := cliPrimary.Bold(true).Render(title)
	body := titleLine + "\n\n" + content
	return cardStyle().Render(body)
}

// renderKeyValue renders a key-value pair with the key right-padded to width.
func renderKeyValue(key, value string, keyWidth int) string {
	paddedKey := fmt.Sprintf("%-*s", keyWidth, key)
	return cliMuted.Render(paddedKey) + "  " + value
}

// renderKeyValueLines builds multiple key-value lines with uniform key width.
func renderKeyValueLines(pairs []kvPair) string {
	if len(pairs) == 0 {
		return ""
	}
	maxKey := 0
	for _, p := range pairs {
		if len(p.key) > maxKey {
			maxKey = len(p.key)
		}
	}
	var lines []string
	for _, p := range pairs {
		lines = append(lines, renderKeyValue(p.key, p.value, maxKey))
	}
	return strings.Join(lines, "\n")
}

// kvPair holds a key-value pair for rendering.
type kvPair struct {
	key   string
	value string
}

// renderStatusLine renders a status icon + label + message.
func renderStatusLine(status CheckStatus, label, message string, labelWidth int) string {
	icon := statusIcon(status)
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label)
	return fmt.Sprintf("%s %s  %s", icon, cliMuted.Render(paddedLabel), message)
}

// renderSuccessCard renders a success message inside a rounded border card.
func renderSuccessCard(title string, details ...string) string {
	titleLine := cliSuccess.Render("\u2713") + " " + title
	body := titleLine
	if len(details) > 0 {
		body += "\n\n" + strings.Join(details, "\n")
	}
	return cardStyle().Render(body)
}

// renderInfoCard renders an informational message inside a rounded border card.
func renderInfoCard(title string, details ...string) string {
	body := title
	if len(details) > 0 {
		body += "\n\n" + strings.Join(details, "\n")
	}
	return cardStyle().Render(body)
}

// renderSummaryLine renders a summary line with colored counts (e.g. "3 passed - 2 warnings - 0 failed").
func renderSummaryLine(ok, warn, fail int) string {
	return fmt.Sprintf("%s passed %s %s warnings %s %s failed",
		cliSuccess.Render(fmt.Sprintf("%d", ok)),
		cliMuted.Render("\u00b7"),
		cliWarn.Render(fmt.Sprintf("%d", warn)),
		cliMuted.Render("\u00b7"),
		cliError.Render(fmt.Sprintf("%d", fail)),
	)
}

// RenderError renders an error inside a ThickBox with a danger theme border
// and a StatusIcon("err") prefix, matching the ScreenError design
// (screens.jsx::ScreenError \u2014 ThickBox color=danger + StatusIcon "err").
//
// All colours are sourced from tui.LightTheme()/DarkTheme() via ThickBox;
// no hex literal appears in this function (AC-CLI-TUI-013).
//
// @MX:ANCHOR: [AUTO] RenderError is the global error render surface for M6-S6
// @MX:REASON: rootCmd error handler and key command RunE paths call this function
func RenderError(err error) string {
	th := resolveTheme()
	icon := tui.StatusIcon("err")
	// Compose title line: error icon + styled label using the danger token.
	dangerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: tui.LightTheme().Danger, Dark: tui.DarkTheme().Danger}).
		Bold(true)
	titleLine := icon + " " + dangerStyle.Render("Error")
	body := titleLine + "\n" + err.Error()

	return tui.ThickBox(tui.BoxOpts{
		Body:   body,
		Theme:  &th,
		Accent: false,
	})
}

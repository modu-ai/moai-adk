package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
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

package uikit

// @MX:NOTE: [AUTO] Rendering utilities for CLI output using lipgloss styling
// @MX:NOTE: [AUTO] Provides card, key-value, status, success/info card renderers
//
// Moved from internal/cli/render.go as part of the uikit kernel extraction
// (SPEC-CLI-UIKIT-KERNEL-001). All helpers exported; style refs resolved
// within package uikit (SuccessStyle/WarnStyle/etc.).

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/modu-ai/moai-adk/internal/tui"
)

// CardStyle returns a lipgloss style for a rounded-border card.
func CardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BorderStyle.GetForeground()).
		Padding(0, 2)
}

// RenderCard renders content inside a rounded border box with a styled title.
func RenderCard(title, content string) string {
	titleLine := PrimaryStyle.Bold(true).Render(title)
	body := titleLine + "\n\n" + content
	return CardStyle().Render(body)
}

// RenderKeyValue renders a key-value pair with the key right-padded to width.
func RenderKeyValue(key, value string, keyWidth int) string {
	paddedKey := fmt.Sprintf("%-*s", keyWidth, key)
	return MutedStyle.Render(paddedKey) + "  " + value
}

// RenderKeyValueLines builds multiple key-value lines with uniform key width.
func RenderKeyValueLines(pairs []KVPair) string {
	if len(pairs) == 0 {
		return ""
	}
	maxKey := 0
	for _, p := range pairs {
		if len(p.Key) > maxKey {
			maxKey = len(p.Key)
		}
	}
	var lines []string
	for _, p := range pairs {
		lines = append(lines, RenderKeyValue(p.Key, p.Value, maxKey))
	}
	return strings.Join(lines, "\n")
}

// KVPair holds a key-value pair for rendering.
type KVPair struct {
	Key   string
	Value string
}

// RenderStatusLine renders a status icon + label + message.
func RenderStatusLine(status CheckStatus, label, message string, labelWidth int) string {
	icon := StatusIcon(status)
	paddedLabel := fmt.Sprintf("%-*s", labelWidth, label)
	return fmt.Sprintf("%s %s  %s", icon, MutedStyle.Render(paddedLabel), message)
}

// RenderSuccessCard renders a success message inside a rounded border card.
func RenderSuccessCard(title string, details ...string) string {
	titleLine := SuccessStyle.Render(string(tui.GlyphDone)) + " " + title
	body := titleLine
	if len(details) > 0 {
		body += "\n\n" + strings.Join(details, "\n")
	}
	return CardStyle().Render(body)
}

// RenderInfoCard renders an informational message inside a rounded border card.
func RenderInfoCard(title string, details ...string) string {
	body := title
	if len(details) > 0 {
		body += "\n\n" + strings.Join(details, "\n")
	}
	return CardStyle().Render(body)
}

// RenderSummaryLine renders a summary line with colored counts.
func RenderSummaryLine(ok, warn, fail int) string {
	return fmt.Sprintf("%s passed %s %s warnings %s %s failed",
		SuccessStyle.Render(fmt.Sprintf("%d", ok)),
		MutedStyle.Render("·"),
		WarnStyle.Render(fmt.Sprintf("%d", warn)),
		MutedStyle.Render("·"),
		ErrorStyle.Render(fmt.Sprintf("%d", fail)),
	)
}

// RenderError renders an error inside a ThickBox with a danger theme border
// and a StatusIcon("err") prefix, matching the ScreenError design.
//
// All colours are sourced from tui.LightTheme()/DarkTheme() via ThickBox;
// no hex literal appears in this function (AC-CLI-TUI-013).
//
// @MX:ANCHOR: [AUTO] RenderError is the global error render surface for M6-S6
// @MX:REASON: rootCmd error handler and key command RunE paths call this function
func RenderError(err error) string {
	th := ResolveTheme()
	icon := tui.StatusIcon("err")
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

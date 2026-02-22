package worktree

import "strings"

import "github.com/charmbracelet/lipgloss"

// Worktree CLI styles matching the parent CLI package palette.
var (
	wtPrimary = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C45A3C", Dark: "#DA7756"})
	wtBorder  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D1D5DB", Dark: "#4B5563"})
	wtSuccess = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#059669", Dark: "#10B981"})
)

// wtCardStyle returns a lipgloss style for a rounded-border card.
func wtCardStyle() lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(wtBorder.GetForeground()).
		Padding(0, 2)
}

// wtCard renders content inside a rounded border box with a styled title.
func wtCard(title, content string) string {
	titleLine := wtPrimary.Bold(true).Render(title)
	body := titleLine + "\n\n" + content
	return wtCardStyle().Render(body)
}

// wtSuccessCard renders a success message inside a rounded border card.
func wtSuccessCard(title string, details ...string) string {
	titleLine := wtSuccess.Render("\u2713") + " " + title
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

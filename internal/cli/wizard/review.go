package wizard

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"

	"github.com/modu-ai/moai-adk/internal/tui"
)

// Review selection values returned by runReview.
const (
	reviewChoiceProceed = "proceed"
	reviewChoiceEdit    = "edit"
	reviewChoiceCancel  = "cancel"
)

// shortLabel returns a compact human-readable label for a question ID, used
// in the progress breadcrumb that precedes every wizard step.
func shortLabel(id string) string {
	switch id {
	case "project_name":
		return "Project"
	case "model_policy":
		return "Model"
	case "development_mode":
		return "Method"
	case "git_mode":
		return "Git mode"
	case "git_provider":
		return "Provider"
	case "github_username", "gitlab_username":
		return "Username"
	case "github_token", "gitlab_token":
		return "Token"
	case "gitlab_instance_url":
		return "GitLab URL"
	default:
		return id
	}
}

// renderProgressBreadcrumb returns a single-line breadcrumb that shows where
// currentQID sits inside the visible question sequence. Answered steps are
// marked with ✓, the current step with ▸, and pending steps with ·.
func renderProgressBreadcrumb(questions []Question, result *WizardResult, currentQID string) string {
	var pieces []string
	seenCurrent := false
	for i := range questions {
		q := &questions[i]
		if q.Condition != nil && !q.Condition(result) {
			continue
		}
		label := shortLabel(q.ID)
		switch {
		case q.ID == currentQID:
			pieces = append(pieces, "▸ "+label)
			seenCurrent = true
		case !seenCurrent:
			pieces = append(pieces, "✓ "+label)
		default:
			pieces = append(pieces, "· "+label)
		}
	}
	return "  " + strings.Join(pieces, "  ")
}

// reviewValue resolves a stored answer to a human-readable display string,
// preferring the Option label over its raw value and masking secret tokens.
func reviewValue(q *Question, result *WizardResult) string {
	raw := storedAnswer(q.ID, result)
	if raw == "" {
		return "(skipped)"
	}
	if q.Type == QuestionTypeSelect {
		for _, opt := range q.Options {
			if opt.Value == raw {
				return opt.Label
			}
		}
	}
	if q.ID == "github_token" || q.ID == "gitlab_token" {
		return "•••• (set)"
	}
	return raw
}

// storedAnswer reads an answer back out of the WizardResult by question ID.
func storedAnswer(id string, result *WizardResult) string {
	switch id {
	case "project_name":
		return result.ProjectName
	case "model_policy":
		return result.ModelPolicy
	case "development_mode":
		return result.DevelopmentMode
	case "git_mode":
		return result.GitMode
	case "git_provider":
		return result.GitProvider
	case "github_username":
		return result.GitHubUsername
	case "github_token":
		return result.GitHubToken
	case "gitlab_instance_url":
		return result.GitLabInstanceURL
	case "gitlab_username":
		return result.GitLabUsername
	case "gitlab_token":
		return result.GitLabToken
	}
	return ""
}

// renderReviewPanel composes the Review summary card that lists every visible
// answer before deployment.
func renderReviewPanel(questions []Question, result *WizardResult) string {
	maxLabel := 0
	type row struct{ label, value string }
	rows := make([]row, 0, len(questions))
	for i := range questions {
		q := &questions[i]
		if q.Condition != nil && !q.Condition(result) {
			continue
		}
		label := shortLabel(q.ID)
		if len(label) > maxLabel {
			maxLabel = len(label)
		}
		rows = append(rows, row{label, reviewValue(q, result)})
	}

	var lines []string
	for _, r := range rows {
		lines = append(lines, fmt.Sprintf("  %-*s   %s", maxLabel, r.label, r.value))
	}

	th := tui.ResolveOS()
	body := strings.Join(lines, "\n")
	return tui.Box(tui.BoxOpts{
		Title:  "Review",
		Body:   body,
		Width:  76,
		Theme:  &th,
		Accent: true,
	})
}

// runReview prints the Review panel and asks the user to Proceed / Edit / Cancel.
func runReview(questions []Question, result *WizardResult, locale *string, theme *huh.Theme) (string, error) {
	_ = locale
	fmt.Println()
	fmt.Println(renderReviewPanel(questions, result))
	fmt.Println()

	var choice string
	sel := huh.NewSelect[string]().
		Title("Proceed with these settings?").
		Description("Review your answers before the templates are deployed.").
		Options(
			huh.NewOption("Proceed — deploy templates", reviewChoiceProceed),
			huh.NewOption("Edit answers — restart the wizard", reviewChoiceEdit),
			huh.NewOption("Cancel", reviewChoiceCancel),
		).
		Value(&choice)

	form := huh.NewForm(huh.NewGroup(sel)).WithTheme(theme).WithAccessible(false)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return reviewChoiceCancel, nil
		}
		return "", fmt.Errorf("review step: %w", err)
	}
	return choice, nil
}

package wizard

import "path/filepath"

// DefaultQuestions returns the standard set of questions for project initialization.
// The questions follow this order:
// 1. Project name (required)
// 2. Development mode
// 3. Git mode
// 4. Git provider (conditional)
// 5. GitLab instance URL (conditional)
// 6. GitHub username (conditional)
// 7. GitHub token (conditional)
// 8. GitLab username (conditional)
// 9. GitLab token (conditional)
func DefaultQuestions(projectRoot string) []Question {
	// Use current directory name as default project name
	defaultProjectName := filepath.Base(projectRoot)
	if defaultProjectName == "." || defaultProjectName == "/" || defaultProjectName == `\` {
		defaultProjectName = "my-project"
	}

	return []Question{
		// 1. Project Name
		{
			ID:          "project_name",
			Type:        QuestionTypeInput,
			Title:       "Enter project name",
			Description: "The name of your project.",
			Default:     defaultProjectName,
			Required:    true,
		},
		// 2. Development Mode
		{
			ID:          "development_mode",
			Type:        QuestionTypeSelect,
			Title:       "Select development methodology",
			Description: "Controls the development workflow cycle used during implementation.",
			Options: []Option{
				{Label: "TDD (Recommended)", Value: "tdd", Desc: "Test-Driven Development: RED-GREEN-REFACTOR"},
				{Label: "DDD", Value: "ddd", Desc: "Domain-Driven Development: ANALYZE-PRESERVE-IMPROVE"},
			},
			Default:  "tdd",
			Required: true,
		},
		// 3. Git Mode
		{
			ID:          "git_mode",
			Type:        QuestionTypeSelect,
			Title:       "Select Git automation mode",
			Description: "Controls how much Git automation Claude can perform.",
			Options: []Option{
				{Label: "Manual", Value: "manual", Desc: "AI never commits or pushes"},
				{Label: "Personal", Value: "personal", Desc: "AI can create branches and commit"},
				{Label: "Team", Value: "team", Desc: "AI can create branches, commit, and open PRs"},
			},
			Default:  "manual",
			Required: true,
		},
		// 4. Git Provider (conditional - only for personal/team modes)
		{
			ID:          "git_provider",
			Type:        QuestionTypeSelect,
			Title:       "Select your Git provider",
			Description: "Choose the Git hosting platform for your project.",
			Options: []Option{
				{Label: "GitHub", Value: "github", Desc: "GitHub.com"},
				{Label: "GitLab", Value: "gitlab", Desc: "GitLab.com or self-hosted GitLab"},
			},
			Default:  "github",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "personal" || r.GitMode == "team"
			},
		},
		// 5. GitLab Instance URL (conditional - only for gitlab provider)
		{
			ID:          "gitlab_instance_url",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab instance URL",
			Description: "For GitLab.com use https://gitlab.com. For self-hosted, enter your instance URL.",
			Default:     "https://gitlab.com",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 6. GitHub Username (conditional - only for github provider)
		{
			ID:          "github_username",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitHub username",
			Description: "Required for Git automation features.",
			Default:     "",
			Required:    false, // Conditional requirement handled by wizard
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 7. GitHub Token (conditional - only for github provider)
		{
			ID:          "github_token",
			Type:        QuestionTypeInput,
			Title:       "Enter GitHub personal access token (optional)",
			Description: "Required for PR creation and pushing. Leave empty to skip or use gh CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "github"
			},
		},
		// 8. GitLab Username (conditional - only for gitlab provider)
		{
			ID:          "gitlab_username",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitLab username",
			Description: "Required for Git automation features with GitLab.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
		// 9. GitLab Token (conditional - only for gitlab provider)
		{
			ID:          "gitlab_token",
			Type:        QuestionTypeInput,
			Title:       "Enter GitLab personal access token (optional)",
			Description: "Required for MR creation and pushing. Leave empty to skip or use glab CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return (r.GitMode == "personal" || r.GitMode == "team") && r.GitProvider == "gitlab"
			},
		},
	}
}

// FilteredQuestions returns questions filtered by their conditions.
// Questions whose conditions return false are excluded.
func FilteredQuestions(questions []Question, result *WizardResult) []Question {
	filtered := make([]Question, 0, len(questions))
	for _, q := range questions {
		if q.Condition == nil || q.Condition(result) {
			filtered = append(filtered, q)
		}
	}
	return filtered
}

// TotalVisibleQuestions counts questions that would be visible given current state.
func TotalVisibleQuestions(questions []Question, result *WizardResult) int {
	count := 0
	for _, q := range questions {
		if q.Condition == nil || q.Condition(result) {
			count++
		}
	}
	return count
}

// QuestionByID finds a question by its ID.
func QuestionByID(questions []Question, id string) *Question {
	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}
	return nil
}

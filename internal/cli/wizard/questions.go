package wizard

import "path/filepath"

// DefaultQuestions returns the standard set of questions for project initialization.
// The questions follow this order:
// 1. Conversation language selection
// 2. User name (optional)
// 3. Project name (required)
// 4. Git mode
// 5. GitHub username (conditional)
// 6. Git commit language
// 7. Code comment language
// 8. Documentation language
// 9. Agent Teams mode
func DefaultQuestions(projectRoot string) []Question {
	// Use current directory name as default project name
	defaultProjectName := filepath.Base(projectRoot)
	if defaultProjectName == "." || defaultProjectName == "/" {
		defaultProjectName = "my-project"
	}

	return []Question{
		// 1. Conversation Language
		{
			ID:          "locale",
			Type:        QuestionTypeSelect,
			Title:       "Select conversation language",
			Description: "This determines the language Claude will use to communicate with you.",
			Options: []Option{
				{Label: "Korean (한국어)", Value: "ko", Desc: "Korean"},
				{Label: "English", Value: "en", Desc: "English"},
				{Label: "Japanese (日本語)", Value: "ja", Desc: "Japanese"},
				{Label: "Chinese (中文)", Value: "zh", Desc: "Chinese"},
			},
			Default:  "en",
			Required: true,
		},
		// 2. User Name
		{
			ID:          "user_name",
			Type:        QuestionTypeInput,
			Title:       "Enter your name",
			Description: "This will be used in configuration files. Press Enter to skip.",
			Default:     "",
			Required:    false,
		},
		// 3. Project Name
		{
			ID:          "project_name",
			Type:        QuestionTypeInput,
			Title:       "Enter project name",
			Description: "The name of your project.",
			Default:     defaultProjectName,
			Required:    true,
		},
		// 4. Git Mode
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
		// 5. GitHub Username (conditional)
		{
			ID:          "github_username",
			Type:        QuestionTypeInput,
			Title:       "Enter your GitHub username",
			Description: "Required for Git automation features.",
			Default:     "",
			Required:    false, // Conditional requirement handled by wizard
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "personal" || r.GitMode == "team"
			},
		},
		// 5b. GitHub Token (conditional, after github_username)
		{
			ID:          "github_token",
			Type:        QuestionTypeInput,
			Title:       "Enter GitHub personal access token (optional)",
			Description: "Required for PR creation and pushing. Leave empty to skip or use gh CLI.",
			Default:     "",
			Required:    false,
			Condition: func(r *WizardResult) bool {
				return r.GitMode == "personal" || r.GitMode == "team"
			},
		},
		// 6. Git Commit Language
		{
			ID:          "git_commit_lang",
			Type:        QuestionTypeSelect,
			Title:       "Select language for Git commits",
			Description: "Language used for commit messages.",
			Options: []Option{
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write commits in Korean"},
				{Label: "English", Value: "en", Desc: "Write commits in English"},
				{Label: "Japanese (日本語)", Value: "ja", Desc: "Write commits in Japanese"},
				{Label: "Chinese (中文)", Value: "zh", Desc: "Write commits in Chinese"},
			},
			Default:  "en",
			Required: true,
		},
		// 7. Code Comment Language
		{
			ID:          "code_comment_lang",
			Type:        QuestionTypeSelect,
			Title:       "Select language for code comments",
			Description: "Language used for comments in code.",
			Options: []Option{
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write comments in Korean"},
				{Label: "English", Value: "en", Desc: "Write comments in English"},
				{Label: "Japanese (日本語)", Value: "ja", Desc: "Write comments in Japanese"},
				{Label: "Chinese (中文)", Value: "zh", Desc: "Write comments in Chinese"},
			},
			Default:  "en",
			Required: true,
		},
		// 8. Documentation Language
		{
			ID:          "doc_lang",
			Type:        QuestionTypeSelect,
			Title:       "Select language for documentation",
			Description: "Language used for documentation files.",
			Options: []Option{
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write docs in Korean"},
				{Label: "English", Value: "en", Desc: "Write docs in English"},
				{Label: "Japanese (日本語)", Value: "ja", Desc: "Write docs in Japanese"},
				{Label: "Chinese (中文)", Value: "zh", Desc: "Write docs in Chinese"},
			},
			Default:  "en",
			Required: true,
		},
		// 9. Model Policy (Token Optimization)
		{
			ID:          "model_policy",
			Type:        QuestionTypeSelect,
			Title:       "Select agent model policy",
			Description: "Controls token consumption by assigning optimal models to each agent.",
			Options: []Option{
				{Label: "High (All inherit)", Value: "high", Desc: "Maximum quality, all agents use parent model"},
				{Label: "Medium (Recommended)", Value: "medium", Desc: "Smart optimization, critical agents use opus"},
				{Label: "Low (Maximum savings)", Value: "low", Desc: "Aggressive optimization, ~45% token savings"},
			},
			Default:  "high",
			Required: true,
		},
		// 10. Agent Teams Mode
		{
			ID:          "agent_teams_mode",
			Type:        QuestionTypeSelect,
			Title:       "Select Agent Teams execution mode",
			Description: "Controls whether MoAI uses Agent Teams (parallel) or sub-agents (sequential).",
			Options: []Option{
				{Label: "Auto (Recommended)", Value: "auto", Desc: "Intelligent selection based on task complexity"},
				{Label: "Sub-agent (Classic)", Value: "subagent", Desc: "Traditional single-agent mode"},
				{Label: "Team (Experimental)", Value: "team", Desc: "Parallel Agent Teams (requires experimental flag)"},
			},
			Default:  "auto",
			Required: true,
		},
		// 12. Max Teammates (conditional - only for team mode)
		{
			ID:          "max_teammates",
			Type:        QuestionTypeSelect,
			Title:       "Select maximum teammates",
			Description: "Maximum number of teammates in a team (2-10 recommended).",
			Options: []Option{
				{Label: "2", Value: "2", Desc: "Minimum for parallel work"},
				{Label: "3", Value: "3", Desc: "Small team"},
				{Label: "4", Value: "4", Desc: "Medium team"},
				{Label: "5", Value: "5", Desc: "Medium-large team"},
				{Label: "6", Value: "6", Desc: "Large team"},
				{Label: "7", Value: "7", Desc: "Large team"},
				{Label: "8", Value: "8", Desc: "Extra large team"},
				{Label: "9", Value: "9", Desc: "Extra large team"},
				{Label: "10", Value: "10", Desc: "Maximum team (default)"},
			},
			Default:  "10",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.AgentTeamsMode == "team"
			},
		},
		// 13. Default Model (conditional - only for team mode)
		{
			ID:          "default_model",
			Type:        QuestionTypeSelect,
			Title:       "Select default model for teammates",
			Description: "Default Claude model for Agent Teammates.",
			Options: []Option{
				{Label: "Haiku (Fast/Cheap)", Value: "haiku", Desc: "Fastest and lowest cost"},
				{Label: "Sonnet (Balanced)", Value: "sonnet", Desc: "Balanced performance and cost (default)"},
				{Label: "Opus (Powerful)", Value: "opus", Desc: "Most capable, higher cost"},
			},
			Default:  "sonnet",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.AgentTeamsMode == "team"
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

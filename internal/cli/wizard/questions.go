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
			// Default option must be first to avoid huh v0.8.0 viewport YOffset bug:
			// selectOption() sets viewport.YOffset = selected index unconditionally,
			// hiding options above the default when all items fit in the viewport.
			Options: []Option{
				{Label: "English", Value: "en", Desc: "English"},
				{Label: "Korean (한국어)", Value: "ko", Desc: "Korean"},
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
		// 4b. Git Provider (conditional - only for personal/team modes)
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
		// 4c. GitLab Instance URL (conditional - only for gitlab provider)
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
		// 5. GitHub Username (conditional - only for github provider)
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
		// 5b. GitHub Token (conditional - only for github provider)
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
		// 5c. GitLab Username (conditional - only for gitlab provider)
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
		// 5d. GitLab Token (conditional - only for gitlab provider)
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
		// 6. Git Commit Language
		{
			ID:          "git_commit_lang",
			Type:        QuestionTypeSelect,
			Title:       "Select language for Git commits",
			Description: "Language used for commit messages.",
			Options: []Option{
				{Label: "English", Value: "en", Desc: "Write commits in English"},
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write commits in Korean"},
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
				{Label: "English", Value: "en", Desc: "Write comments in English"},
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write comments in Korean"},
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
				{Label: "English", Value: "en", Desc: "Write docs in English"},
				{Label: "Korean (한국어)", Value: "ko", Desc: "Write docs in Korean"},
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
				{Label: "High (Max $200/mo)", Value: "high", Desc: "Best quality - opus 23, sonnet 1, haiku 4"},
				{Label: "Medium (Max $100/mo)", Value: "medium", Desc: "Balanced - opus 4, sonnet 19, haiku 5"},
				{Label: "Low (Plus $20/mo)", Value: "low", Desc: "No Opus - sonnet 12, haiku 16"},
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
			// Default option must be first to avoid huh v0.8.x YOffset bug.
			Options: []Option{
				{Label: "10", Value: "10", Desc: "Maximum team (default)"},
				{Label: "9", Value: "9", Desc: "Extra large team"},
				{Label: "8", Value: "8", Desc: "Extra large team"},
				{Label: "7", Value: "7", Desc: "Large team"},
				{Label: "6", Value: "6", Desc: "Large team"},
				{Label: "5", Value: "5", Desc: "Medium-large team"},
				{Label: "4", Value: "4", Desc: "Medium team"},
				{Label: "3", Value: "3", Desc: "Small team"},
				{Label: "2", Value: "2", Desc: "Minimum for parallel work"},
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
			// Default option must be first to avoid huh v0.8.x YOffset bug.
			Options: []Option{
				{Label: "Sonnet (Balanced)", Value: "sonnet", Desc: "Balanced performance and cost (default)"},
				{Label: "Haiku (Fast/Cheap)", Value: "haiku", Desc: "Fastest and lowest cost"},
				{Label: "Opus (Powerful)", Value: "opus", Desc: "Most capable, higher cost"},
			},
			Default:  "sonnet",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.AgentTeamsMode == "team"
			},
		},
		// 13b. Teammate Display Mode (conditional - only for team or auto mode)
		{
			ID:          "teammate_display",
			Type:        QuestionTypeSelect,
			Title:       "Select teammate display mode",
			Description: "Controls how Agent Teammates are displayed. Requires tmux for split panes.",
			// Default option must be first to avoid huh v0.8.x YOffset bug.
			Options: []Option{
				{Label: "Auto (Recommended)", Value: "auto", Desc: "Use tmux if available, else in-process (default)"},
				{Label: "In-Process", Value: "in-process", Desc: "Run in same terminal (works everywhere)"},
				{Label: "Tmux", Value: "tmux", Desc: "Split panes in tmux (requires tmux/iTerm2)"},
			},
			Default:  "auto",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.AgentTeamsMode == "team" || r.AgentTeamsMode == "auto"
			},
		},
		// 14. Statusline Preset
		{
			ID:          "statusline_preset",
			Type:        QuestionTypeSelect,
			Title:       "Select statusline display preset",
			Description: "Controls which segments are shown in the Claude Code statusline.",
			Options: []Option{
				{Label: "Full", Value: "full", Desc: "All 8 segments displayed"},
				{Label: "Compact", Value: "compact", Desc: "Model, context, git status, git branch"},
				{Label: "Minimal", Value: "minimal", Desc: "Model and context only"},
				{Label: "Custom", Value: "custom", Desc: "Choose individual segments"},
			},
			Default:  "full",
			Required: true,
		},
		// 14a. Statusline Segment: Model (conditional - only for custom preset)
		{
			ID:          "statusline_seg_model",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show model name",
			Description: "Display the current Claude model name in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show model segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide model segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14b. Statusline Segment: Context (conditional - only for custom preset)
		{
			ID:          "statusline_seg_context",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show context usage",
			Description: "Display the context window usage percentage in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show context segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide context segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14c. Statusline Segment: Output Style (conditional - only for custom preset)
		{
			ID:          "statusline_seg_output_style",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show output style",
			Description: "Display the active output style name in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show output style segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide output style segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14d. Statusline Segment: Directory (conditional - only for custom preset)
		{
			ID:          "statusline_seg_directory",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show directory name",
			Description: "Display the current working directory name in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show directory segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide directory segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14e. Statusline Segment: Git Status (conditional - only for custom preset)
		{
			ID:          "statusline_seg_git_status",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show git status",
			Description: "Display git status (staged, modified, untracked counts) in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show git status segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide git status segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14f. Statusline Segment: Claude Version (conditional - only for custom preset)
		{
			ID:          "statusline_seg_claude_version",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show Claude version",
			Description: "Display the Claude Code version in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show Claude version segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide Claude version segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14g. Statusline Segment: MoAI Version (conditional - only for custom preset)
		{
			ID:          "statusline_seg_moai_version",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show MoAI version",
			Description: "Display the MoAI-ADK version in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show MoAI version segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide MoAI version segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
			},
		},
		// 14h. Statusline Segment: Git Branch (conditional - only for custom preset)
		{
			ID:          "statusline_seg_git_branch",
			Type:        QuestionTypeSelect,
			Title:       "Statusline: Show git branch",
			Description: "Display the current git branch name in the statusline.",
			Options: []Option{
				{Label: "Enabled", Value: "true", Desc: "Show git branch segment"},
				{Label: "Disabled", Value: "false", Desc: "Hide git branch segment"},
			},
			Default:  "true",
			Required: true,
			Condition: func(r *WizardResult) bool {
				return r.StatuslinePreset == "custom"
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

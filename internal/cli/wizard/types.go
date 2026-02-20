// Package wizard provides an interactive Bubble Tea-based wizard
// for MoAI project initialization.
package wizard

import (
	"errors"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// WizardResult holds the user's selections from the init wizard.
type WizardResult struct {
	// Core settings
	ProjectName string // Project name (required)
	Locale      string // Conversation language code: ko, en, ja, zh
	UserName    string // User display name (optional)

	// Git settings
	GitMode           string // Git automation mode: manual, personal, team
	GitProvider       string // Git provider: "github", "gitlab"
	GitHubUsername    string // GitHub username (required for personal/team modes)
	GitHubToken       string // GitHub personal access token (optional)
	GitLabInstanceURL string // GitLab instance URL (for self-hosted, e.g. "https://gitlab.company.com")
	GitLabUsername    string // GitLab username (for personal/team modes with gitlab provider)
	GitLabToken       string // GitLab personal access token (optional)

	// Output language settings
	GitCommitLang   string // Language for git commit messages: en, ko, ja, zh
	CodeCommentLang string // Language for code comments: en, ko, ja, zh
	DocLang         string // Language for documentation: en, ko, ja, zh

	// Development methodology
	DevelopmentMode string // Development mode: ddd, tdd

	// Model policy
	ModelPolicy string // Token consumption tier: high, medium, low

	// Agent Teams settings
	AgentTeamsMode  string // Agent Teams execution mode: auto, subagent, team
	MaxTeammates    string // Maximum teammates (2-10, for team mode)
	DefaultModel    string // Default model for teammates: haiku, sonnet, opus
	TeammateDisplay string // Teammate display mode: auto, in-process, tmux

	// Statusline settings
	StatuslinePreset   string          // Statusline preset: full, compact, minimal, custom
	StatuslineSegments map[string]bool // Segment toggles (used when preset is "custom")
}

// QuestionType represents the type of wizard question.
type QuestionType int

const (
	// QuestionTypeSelect is a single-choice selection question.
	QuestionTypeSelect QuestionType = iota
	// QuestionTypeInput is a text input question.
	QuestionTypeInput
)

// Question defines a single wizard question.
type Question struct {
	ID          string                   // Unique identifier
	Type        QuestionType             // Select or Input
	Title       string                   // Question title
	Description string                   // Additional description
	Options     []Option                 // Options for select questions
	Default     string                   // Default value
	Required    bool                     // Whether the field is required
	Condition   func(*WizardResult) bool // Condition for showing this question
}

// Option represents a selectable option.
type Option struct {
	Label string // Display label
	Value string // Actual value stored
	Desc  string // Optional description
}

// State represents the current state of the wizard.
type State int

const (
	// StateRunning means the wizard is actively running.
	StateRunning State = iota
	// StateCompleted means the wizard finished successfully.
	StateCompleted
	// StateCancelled means the user cancelled the wizard.
	StateCancelled
)

// Error definitions for the wizard package.
var (
	// ErrCancelled is returned when the user cancels the wizard.
	ErrCancelled = errors.New("wizard cancelled by user")
	// ErrNoQuestions is returned when no questions are provided.
	ErrNoQuestions = errors.New("no questions provided")
	// ErrInvalidQuestion is returned when a question index is out of bounds.
	ErrInvalidQuestion = errors.New("invalid question index")
)

// LangNameMap is an alias to the canonical language map in pkg/models.
// Deprecated: Use models.LangNameMap directly.
var LangNameMap = models.LangNameMap

// GetLanguageName returns the full language name for a code.
// Returns "English" if the code is not found.
func GetLanguageName(code string) string {
	return models.GetLanguageName(code)
}

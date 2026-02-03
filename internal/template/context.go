package template

import "github.com/modu-ai/moai-adk-go/pkg/models"

// TemplateContext provides data for template rendering during project initialization.
// All fields are exported for use with Go's text/template package.
type TemplateContext struct {
	// Project
	ProjectName string
	ProjectRoot string

	// User
	UserName string

	// Language settings
	ConversationLanguage     string // e.g., "ko", "en"
	ConversationLanguageName string // e.g., "Korean (한국어)"
	AgentPromptLanguage      string // e.g., "en"
	GitCommitMessages        string // e.g., "en"
	CodeComments             string // e.g., "en"
	Documentation            string // e.g., "en"
	ErrorMessages            string // e.g., "en"

	// Git settings
	GitMode        string // "manual", "personal", "team"
	GitHubUsername string // GitHub username (for personal/team modes)

	// Development settings
	DevelopmentMode    string // "ddd", "tdd", "hybrid"
	EnforceQuality     bool   // true
	TestCoverageTarget int    // 85

	// Workflow settings
	AutoClear  bool // true
	PlanTokens int  // 30000
	RunTokens  int  // 180000
	SyncTokens int  // 40000

	// Meta
	Version  string // MoAI-ADK version
	Platform string // "darwin", "linux", "windows"
}

// ContextOption configures a TemplateContext.
type ContextOption func(*TemplateContext)

// NewTemplateContext creates a TemplateContext with sensible defaults,
// then applies any provided options.
func NewTemplateContext(opts ...ContextOption) *TemplateContext {
	ctx := &TemplateContext{
		// Defaults
		ConversationLanguage:     "en",
		ConversationLanguageName: "English",
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
		GitMode:                  "manual",
		GitHubUsername:           "",
		DevelopmentMode:          string(models.ModeDDD),
		EnforceQuality:           true,
		TestCoverageTarget:       85,
		AutoClear:                true,
		PlanTokens:               30000,
		RunTokens:                180000,
		SyncTokens:               40000,
	}

	for _, opt := range opts {
		opt(ctx)
	}

	// Resolve language name if only code was provided
	if ctx.ConversationLanguageName == "" || ctx.ConversationLanguageName == "English" {
		ctx.ConversationLanguageName = ResolveLanguageName(ctx.ConversationLanguage)
	}

	return ctx
}

// WithProject sets project-related fields.
func WithProject(name, root string) ContextOption {
	return func(c *TemplateContext) {
		c.ProjectName = name
		c.ProjectRoot = root
	}
}

// WithUser sets the user name.
func WithUser(name string) ContextOption {
	return func(c *TemplateContext) {
		c.UserName = name
	}
}

// WithLanguage sets the conversation language.
func WithLanguage(code string) ContextOption {
	return func(c *TemplateContext) {
		c.ConversationLanguage = code
		c.ConversationLanguageName = ResolveLanguageName(code)
	}
}

// WithDevelopmentMode sets the development mode.
func WithDevelopmentMode(mode string) ContextOption {
	return func(c *TemplateContext) {
		devMode := models.DevelopmentMode(mode)
		if devMode.IsValid() {
			c.DevelopmentMode = mode
		}
	}
}

// WithPlatform sets the target platform.
func WithPlatform(platform string) ContextOption {
	return func(c *TemplateContext) {
		c.Platform = platform
	}
}

// WithVersion sets the MoAI-ADK version.
func WithVersion(version string) ContextOption {
	return func(c *TemplateContext) {
		c.Version = version
	}
}

// WithGitMode sets the git mode (manual, personal, team).
func WithGitMode(mode string) ContextOption {
	return func(c *TemplateContext) {
		if mode == "manual" || mode == "personal" || mode == "team" {
			c.GitMode = mode
		}
	}
}

// WithGitHubUsername sets the GitHub username.
func WithGitHubUsername(username string) ContextOption {
	return func(c *TemplateContext) {
		c.GitHubUsername = username
	}
}

// WithOutputLanguages sets the output language settings.
func WithOutputLanguages(gitCommit, codeComment, documentation string) ContextOption {
	return func(c *TemplateContext) {
		if gitCommit != "" {
			c.GitCommitMessages = gitCommit
		}
		if codeComment != "" {
			c.CodeComments = codeComment
		}
		if documentation != "" {
			c.Documentation = documentation
		}
	}
}

// langNameMap maps language codes to full names with native script.
var langNameMap = map[string]string{
	"en": "English",
	"ko": "Korean (한국어)",
	"ja": "Japanese (日本語)",
	"zh": "Chinese (中文)",
	"es": "Spanish (Español)",
	"fr": "French (Français)",
	"de": "German (Deutsch)",
}

// ResolveLanguageName returns the full name for a language code.
func ResolveLanguageName(code string) string {
	if name, ok := langNameMap[code]; ok {
		return name
	}
	return "English"
}

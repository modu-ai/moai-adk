package template

import (
	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

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
	GitMode           string // "manual", "personal", "team"
	GitProvider       string // "github", "gitlab" (default: "github")
	GitHubUsername    string // GitHub username (for personal/team modes)
	GitLabInstanceURL string // GitLab instance URL (e.g., "https://gitlab.com")

	// Development settings
	DevelopmentMode    string // "ddd", "tdd"
	EnforceQuality     bool   // true
	TestCoverageTarget int    // 85

	// Workflow settings
	AutoClear  bool // true
	PlanTokens int  // 30000
	RunTokens  int  // 180000
	SyncTokens int  // 40000

	// Meta
	Version       string // MoAI-ADK version
	Platform      string // "darwin", "linux", "windows"
	InitializedAt string // ISO 8601 timestamp when project was initialized
	CreatedAt     string // ISO 8601 timestamp when project was created

	// Paths (detected during initialization)
	GoBinPath string // Detected Go binary installation path (e.g., "/home/user/go/bin")
	HomeDir   string // User's home directory
	SmartPATH string // Captured terminal PATH with essential dirs prepended

	// Model policy for agent model assignment
	ModelPolicy string // "high", "medium", "low" (default: "high")
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
		GitProvider:              "github",
		GitHubUsername:           "",
		GitLabInstanceURL:        "",
		DevelopmentMode:          string(models.ModeTDD),
		EnforceQuality:           true,
		TestCoverageTarget:       config.DefaultTestCoverageTarget,
		AutoClear:                true,
		PlanTokens:               config.DefaultPlanTokens,
		RunTokens:   config.DefaultRunTokens,
		SyncTokens:  config.DefaultSyncTokens,
		ModelPolicy: string(DefaultModelPolicy),
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

// WithInitializedAt sets the project initialization timestamp.
func WithInitializedAt(timestamp string) ContextOption {
	return func(c *TemplateContext) {
		c.InitializedAt = timestamp
	}
}

// WithCreatedAt sets the project creation timestamp.
func WithCreatedAt(timestamp string) ContextOption {
	return func(c *TemplateContext) {
		c.CreatedAt = timestamp
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

// WithGitProvider sets the git provider.
func WithGitProvider(provider string) ContextOption {
	return func(c *TemplateContext) {
		if provider == "github" || provider == "gitlab" {
			c.GitProvider = provider
		}
	}
}

// WithGitHubUsername sets the GitHub username.
func WithGitHubUsername(username string) ContextOption {
	return func(c *TemplateContext) {
		c.GitHubUsername = username
	}
}

// WithGitLabInstanceURL sets the GitLab instance URL.
func WithGitLabInstanceURL(url string) ContextOption {
	return func(c *TemplateContext) {
		c.GitLabInstanceURL = url
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

// WithGoBinPath sets the detected Go binary installation path.
func WithGoBinPath(path string) ContextOption {
	return func(c *TemplateContext) {
		c.GoBinPath = path
	}
}

// WithHomeDir sets the user's home directory.
func WithHomeDir(dir string) ContextOption {
	return func(c *TemplateContext) {
		c.HomeDir = dir
	}
}

// WithSmartPATH sets the captured terminal PATH with essential dirs prepended.
func WithSmartPATH(path string) ContextOption {
	return func(c *TemplateContext) {
		c.SmartPATH = path
	}
}

// WithModelPolicy sets the model policy for agent model assignment.
func WithModelPolicy(policy string) ContextOption {
	return func(c *TemplateContext) {
		if IsValidModelPolicy(policy) {
			c.ModelPolicy = policy
		}
	}
}

// ResolveLanguageName returns the full name for a language code.
// Uses the canonical language map from pkg/models.
func ResolveLanguageName(code string) string {
	return models.GetLanguageName(code)
}

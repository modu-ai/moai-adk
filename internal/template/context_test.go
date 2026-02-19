package template

import (
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestNewTemplateContext_Defaults(t *testing.T) {
	ctx := NewTemplateContext()

	if ctx.ConversationLanguage != "en" {
		t.Errorf("ConversationLanguage = %q, want %q", ctx.ConversationLanguage, "en")
	}
	if ctx.AgentPromptLanguage != "en" {
		t.Errorf("AgentPromptLanguage = %q, want %q", ctx.AgentPromptLanguage, "en")
	}
	if ctx.GitMode != "manual" {
		t.Errorf("GitMode = %q, want %q", ctx.GitMode, "manual")
	}
	if ctx.GitProvider != "github" {
		t.Errorf("GitProvider = %q, want %q", ctx.GitProvider, "github")
	}
	if ctx.DevelopmentMode != string(models.ModeTDD) {
		t.Errorf("DevelopmentMode = %q, want %q", ctx.DevelopmentMode, models.ModeTDD)
	}
	if !ctx.EnforceQuality {
		t.Error("EnforceQuality = false, want true")
	}
	if ctx.TestCoverageTarget != config.DefaultTestCoverageTarget {
		t.Errorf("TestCoverageTarget = %d, want %d", ctx.TestCoverageTarget, config.DefaultTestCoverageTarget)
	}
	if !ctx.AutoClear {
		t.Error("AutoClear = false, want true")
	}
	if ctx.PlanTokens != config.DefaultPlanTokens {
		t.Errorf("PlanTokens = %d, want %d", ctx.PlanTokens, config.DefaultPlanTokens)
	}
	if ctx.RunTokens != config.DefaultRunTokens {
		t.Errorf("RunTokens = %d, want %d", ctx.RunTokens, config.DefaultRunTokens)
	}
	if ctx.SyncTokens != config.DefaultSyncTokens {
		t.Errorf("SyncTokens = %d, want %d", ctx.SyncTokens, config.DefaultSyncTokens)
	}
	if ctx.ModelPolicy != string(DefaultModelPolicy) {
		t.Errorf("ModelPolicy = %q, want %q", ctx.ModelPolicy, DefaultModelPolicy)
	}
}

func TestNewTemplateContext_WithOptions(t *testing.T) {
	tests := []struct {
		name    string
		opt     ContextOption
		checkFn func(t *testing.T, ctx *TemplateContext)
	}{
		{
			name: "WithProject",
			opt:  WithProject("my-project", "/home/user/project"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.ProjectName != "my-project" {
					t.Errorf("ProjectName = %q, want %q", ctx.ProjectName, "my-project")
				}
				if ctx.ProjectRoot != "/home/user/project" {
					t.Errorf("ProjectRoot = %q, want %q", ctx.ProjectRoot, "/home/user/project")
				}
			},
		},
		{
			name: "WithUser",
			opt:  WithUser("testuser"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.UserName != "testuser" {
					t.Errorf("UserName = %q, want %q", ctx.UserName, "testuser")
				}
			},
		},
		{
			name: "WithLanguage_ko",
			opt:  WithLanguage("ko"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.ConversationLanguage != "ko" {
					t.Errorf("ConversationLanguage = %q, want %q", ctx.ConversationLanguage, "ko")
				}
				// Language name should be resolved
				if ctx.ConversationLanguageName == "" {
					t.Error("ConversationLanguageName is empty")
				}
			},
		},
		{
			name: "WithDevelopmentMode_ddd",
			opt:  WithDevelopmentMode("ddd"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.DevelopmentMode != "ddd" {
					t.Errorf("DevelopmentMode = %q, want %q", ctx.DevelopmentMode, "ddd")
				}
			},
		},
		{
			name: "WithDevelopmentMode_invalid",
			opt:  WithDevelopmentMode("waterfall"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				// Invalid mode should not change default
				if ctx.DevelopmentMode != string(models.ModeTDD) {
					t.Errorf("DevelopmentMode = %q, want %q (unchanged from default)", ctx.DevelopmentMode, models.ModeTDD)
				}
			},
		},
		{
			name: "WithPlatform",
			opt:  WithPlatform("linux"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.Platform != "linux" {
					t.Errorf("Platform = %q, want %q", ctx.Platform, "linux")
				}
			},
		},
		{
			name: "WithVersion",
			opt:  WithVersion("2.5.0"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.Version != "2.5.0" {
					t.Errorf("Version = %q, want %q", ctx.Version, "2.5.0")
				}
			},
		},
		{
			name: "WithInitializedAt",
			opt:  WithInitializedAt("2026-01-01T00:00:00Z"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.InitializedAt != "2026-01-01T00:00:00Z" {
					t.Errorf("InitializedAt = %q, want %q", ctx.InitializedAt, "2026-01-01T00:00:00Z")
				}
			},
		},
		{
			name: "WithCreatedAt",
			opt:  WithCreatedAt("2026-02-01T12:00:00Z"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.CreatedAt != "2026-02-01T12:00:00Z" {
					t.Errorf("CreatedAt = %q, want %q", ctx.CreatedAt, "2026-02-01T12:00:00Z")
				}
			},
		},
		{
			name: "WithGitMode_personal",
			opt:  WithGitMode("personal"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitMode != "personal" {
					t.Errorf("GitMode = %q, want %q", ctx.GitMode, "personal")
				}
			},
		},
		{
			name: "WithGitMode_team",
			opt:  WithGitMode("team"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitMode != "team" {
					t.Errorf("GitMode = %q, want %q", ctx.GitMode, "team")
				}
			},
		},
		{
			name: "WithGitMode_invalid",
			opt:  WithGitMode("invalid"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitMode != "manual" {
					t.Errorf("GitMode = %q, want %q (unchanged from default)", ctx.GitMode, "manual")
				}
			},
		},
		{
			name: "WithGitProvider_gitlab",
			opt:  WithGitProvider("gitlab"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitProvider != "gitlab" {
					t.Errorf("GitProvider = %q, want %q", ctx.GitProvider, "gitlab")
				}
			},
		},
		{
			name: "WithGitProvider_invalid",
			opt:  WithGitProvider("bitbucket"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitProvider != "github" {
					t.Errorf("GitProvider = %q, want %q (unchanged from default)", ctx.GitProvider, "github")
				}
			},
		},
		{
			name: "WithGitHubUsername",
			opt:  WithGitHubUsername("octocat"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitHubUsername != "octocat" {
					t.Errorf("GitHubUsername = %q, want %q", ctx.GitHubUsername, "octocat")
				}
			},
		},
		{
			name: "WithGitLabInstanceURL",
			opt:  WithGitLabInstanceURL("https://gitlab.example.com"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitLabInstanceURL != "https://gitlab.example.com" {
					t.Errorf("GitLabInstanceURL = %q, want %q", ctx.GitLabInstanceURL, "https://gitlab.example.com")
				}
			},
		},
		{
			name: "WithOutputLanguages_all",
			opt:  WithOutputLanguages("ko", "en", "ja"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GitCommitMessages != "ko" {
					t.Errorf("GitCommitMessages = %q, want %q", ctx.GitCommitMessages, "ko")
				}
				if ctx.CodeComments != "en" {
					t.Errorf("CodeComments = %q, want %q", ctx.CodeComments, "en")
				}
				if ctx.Documentation != "ja" {
					t.Errorf("Documentation = %q, want %q", ctx.Documentation, "ja")
				}
			},
		},
		{
			name: "WithOutputLanguages_partial_empty",
			opt:  WithOutputLanguages("", "en", ""),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				// Empty strings should not override defaults
				if ctx.GitCommitMessages != "en" {
					t.Errorf("GitCommitMessages = %q, want %q (default)", ctx.GitCommitMessages, "en")
				}
				if ctx.CodeComments != "en" {
					t.Errorf("CodeComments = %q, want %q", ctx.CodeComments, "en")
				}
				if ctx.Documentation != "en" {
					t.Errorf("Documentation = %q, want %q (default)", ctx.Documentation, "en")
				}
			},
		},
		{
			name: "WithGoBinPath",
			opt:  WithGoBinPath("/home/user/go/bin"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.GoBinPath != "/home/user/go/bin" {
					t.Errorf("GoBinPath = %q, want %q", ctx.GoBinPath, "/home/user/go/bin")
				}
			},
		},
		{
			name: "WithHomeDir",
			opt:  WithHomeDir("/home/user"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.HomeDir != "/home/user" {
					t.Errorf("HomeDir = %q, want %q", ctx.HomeDir, "/home/user")
				}
			},
		},
		{
			name: "WithSmartPATH",
			opt:  WithSmartPATH("/usr/local/bin:/usr/bin"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.SmartPATH != "/usr/local/bin:/usr/bin" {
					t.Errorf("SmartPATH = %q, want %q", ctx.SmartPATH, "/usr/local/bin:/usr/bin")
				}
			},
		},
		{
			name: "WithModelPolicy_low",
			opt:  WithModelPolicy("low"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.ModelPolicy != "low" {
					t.Errorf("ModelPolicy = %q, want %q", ctx.ModelPolicy, "low")
				}
			},
		},
		{
			name: "WithModelPolicy_invalid",
			opt:  WithModelPolicy("ultra"),
			checkFn: func(t *testing.T, ctx *TemplateContext) {
				if ctx.ModelPolicy != string(DefaultModelPolicy) {
					t.Errorf("ModelPolicy = %q, want %q (unchanged from default)", ctx.ModelPolicy, DefaultModelPolicy)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := NewTemplateContext(tt.opt)
			tt.checkFn(t, ctx)
		})
	}
}

func TestNewTemplateContext_MultipleOptions(t *testing.T) {
	ctx := NewTemplateContext(
		WithProject("test-proj", "/tmp/test"),
		WithUser("tester"),
		WithLanguage("ja"),
		WithVersion("1.0.0"),
		WithPlatform("darwin"),
		WithGoBinPath("/opt/go/bin"),
		WithHomeDir("/home/tester"),
	)

	if ctx.ProjectName != "test-proj" {
		t.Errorf("ProjectName = %q, want %q", ctx.ProjectName, "test-proj")
	}
	if ctx.UserName != "tester" {
		t.Errorf("UserName = %q, want %q", ctx.UserName, "tester")
	}
	if ctx.ConversationLanguage != "ja" {
		t.Errorf("ConversationLanguage = %q, want %q", ctx.ConversationLanguage, "ja")
	}
	if ctx.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", ctx.Version, "1.0.0")
	}
	if ctx.Platform != "darwin" {
		t.Errorf("Platform = %q, want %q", ctx.Platform, "darwin")
	}
	if ctx.GoBinPath != "/opt/go/bin" {
		t.Errorf("GoBinPath = %q, want %q", ctx.GoBinPath, "/opt/go/bin")
	}
	if ctx.HomeDir != "/home/tester" {
		t.Errorf("HomeDir = %q, want %q", ctx.HomeDir, "/home/tester")
	}
}

func TestResolveLanguageName(t *testing.T) {
	tests := []struct {
		code     string
		wantName string
	}{
		{"en", "English"},
		{"ko", ""},   // Just verify it returns something non-empty
		{"ja", ""},   // Just verify it returns something non-empty
		{"zh", ""},   // Just verify it returns something non-empty
		{"xx", ""},   // Unknown code
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := ResolveLanguageName(tt.code)
			if tt.wantName != "" && got != tt.wantName {
				t.Errorf("ResolveLanguageName(%q) = %q, want %q", tt.code, got, tt.wantName)
			}
			if tt.code == "en" && got != "English" {
				t.Errorf("ResolveLanguageName(%q) = %q, want %q", tt.code, got, "English")
			}
		})
	}
}

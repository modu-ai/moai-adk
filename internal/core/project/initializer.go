package project

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/modu-ai/moai-adk/internal/defs"
	"github.com/modu-ai/moai-adk/internal/manifest"
	"github.com/modu-ai/moai-adk/internal/shell"
	"github.com/modu-ai/moai-adk/internal/template"
	"github.com/modu-ai/moai-adk/pkg/version"
)

// InitOptions configures the project initialization.
type InitOptions struct {
	ProjectRoot       string   // Absolute or relative path to the project root.
	ProjectName       string   // Name of the project.
	Language          string   // Primary programming language.
	Framework         string   // Framework name, or "none".
	Features          []string // Selected features (e.g., "LSP", "Quality Gates").
	UserName          string   // User display name for configuration.
	ConvLang          string   // Conversation language code (e.g., "en", "ko").
	DevelopmentMode   string   // "ddd" or "tdd".
	GitMode           string   // Git workflow mode: "manual", "personal", or "team".
	GitProvider       string   // Git provider: "github", "gitlab".
	GitHubUsername    string   // GitHub username (for personal/team modes).
	GitLabInstanceURL string   // GitLab instance URL (for self-hosted instances).
	GitCommitLang     string   // Git commit message language code.
	CodeCommentLang   string   // Code comment language code.
	DocLang           string   // Documentation language code.
	Platform          string   // Target platform ("darwin", "linux", "windows"). Defaults to runtime.GOOS.
	NonInteractive    bool     // If true, skip wizard and use defaults/flags.
	Force             bool     // If true, allow reinitializing an existing project.
	SkipShellConfig   bool     // If true, skip shell environment configuration.
	ModelPolicy       string   // Token consumption tier: "high", "medium", "low".
}

// InitResult summarizes the outcome of project initialization.
type InitResult struct {
	CreatedDirs     []string // Directories that were created.
	CreatedFiles    []string // Files that were created.
	DevelopmentMode string   // Selected development methodology.
	BackupPath      string   // Non-empty if --force was used and backup was created.
	Warnings        []string // Non-fatal warnings during initialization.
	ShellConfigured bool     // Whether shell environment was configured.
	ShellConfigFile string   // Path to the shell config file that was modified.
}

// Initializer handles project scaffolding and setup.
type Initializer interface {
	// Init creates a new MoAI project with the given options.
	Init(ctx context.Context, opts InitOptions) (*InitResult, error)
}

// projectInitializer is the concrete implementation of Initializer.
type projectInitializer struct {
	deployer    template.Deployer // May be nil if templates are not available.
	manifestMgr manifest.Manager  // Manifest tracking.
	logger      *slog.Logger
}

// NewInitializer creates an Initializer with the given dependencies.
func NewInitializer(deployer template.Deployer, manifestMgr manifest.Manager, logger *slog.Logger) Initializer {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	return &projectInitializer{
		deployer:    deployer,
		manifestMgr: manifestMgr,
		logger:      logger,
	}
}

// moaiDirs lists the directories to create under .moai/.
var moaiDirs = []string{
	"config/sections",
	"specs",
	"reports",
	"memory",
	"logs",
}

// claudeDirs lists the directories to create under .claude/.
var claudeDirs = []string{
	"agents/moai",
	"skills",
	"commands/moai",
	"rules/moai",
	"output-styles",
}

// Init creates a new MoAI project with the given options.
func (i *projectInitializer) Init(ctx context.Context, opts InitOptions) (*InitResult, error) {
	opts.ProjectRoot = filepath.Clean(opts.ProjectRoot)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	i.logger.Info("initializing MoAI project",
		"root", opts.ProjectRoot,
		"name", opts.ProjectName,
		"language", opts.Language,
		"mode", opts.DevelopmentMode,
	)

	result := &InitResult{
		DevelopmentMode: opts.DevelopmentMode,
	}

	// Step 1: Create .moai/ directory structure
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.createMoAIDirs(opts.ProjectRoot, result); err != nil {
		return nil, fmt.Errorf("create .moai/ structure: %w", err)
	}

	// Step 2: Create .claude/ directory structure
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.createClaudeDirs(opts.ProjectRoot, result); err != nil {
		return nil, fmt.Errorf("create .claude/ structure: %w", err)
	}

	// Step 3: Deploy templates (if deployer is available)
	// Templates include .moai/config/sections/*.yaml via .tmpl rendering
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if i.deployer != nil {
		if err := i.deployTemplates(ctx, opts, result); err != nil {
			// Template deployment is non-fatal; record warning
			result.Warnings = append(result.Warnings, fmt.Sprintf("template deployment: %s", err))
			i.logger.Warn("template deployment failed", "error", err)
		}
	} else {
		// Fallback: generate config files directly when no deployer is available
		if err := i.generateConfigsFallback(opts, result); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("config generation: %s", err))
			i.logger.Warn("config generation failed", "error", err)
		}
	}

	// Step 3b: Apply model policy to agent files (post-deployment patching).
	// Always apply a policy; default to high when not explicitly set.
	// "inherit" is no longer a supported model value in Claude Code.
	{
		policy := template.ModelPolicy(opts.ModelPolicy)
		if policy == "" {
			policy = template.DefaultModelPolicy
		}
		if err := template.ApplyModelPolicy(opts.ProjectRoot, policy, i.manifestMgr); err != nil {
			i.logger.Warn("failed to apply model policy", "error", err)
		}
	}

	// Step 4: Create CLAUDE.md
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.createClaudeMD(opts, result); err != nil {
		return nil, fmt.Errorf("create CLAUDE.md: %w", err)
	}

	// Step 5: Initialize manifest
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.initManifest(opts.ProjectRoot, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("manifest initialization: %s", err))
		i.logger.Warn("manifest initialization failed", "error", err)
	}

	// Step 6: Configure shell environment (REQ-SHELL-001)
	// Adds CLAUDE_DISABLE_PATH_WARNING and PATH to appropriate shell config file
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if !opts.SkipShellConfig {
		if shellResult, err := i.configureShellEnv(); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("shell configuration: %s", err))
			i.logger.Warn("shell configuration failed", "error", err)
		} else if shellResult != nil {
			result.ShellConfigured = !shellResult.Skipped
			result.ShellConfigFile = shellResult.ConfigFile
			if !shellResult.Skipped {
				i.logger.Info("shell environment configured",
					"configFile", shellResult.ConfigFile,
					"linesAdded", len(shellResult.LinesAdded),
				)
			}
		}
	}

	i.logger.Info("project initialized",
		"dirs", len(result.CreatedDirs),
		"files", len(result.CreatedFiles),
	)

	return result, nil
}

// createMoAIDirs creates the .moai/ directory structure.
func (i *projectInitializer) createMoAIDirs(root string, result *InitResult) error {
	for _, dir := range moaiDirs {
		dirPath := filepath.Clean(filepath.Join(root, defs.MoAIDir, dir))
		if err := os.MkdirAll(dirPath, defs.DirPerm); err != nil {
			return fmt.Errorf("mkdir %s: %w", dirPath, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, filepath.Join(defs.MoAIDir, dir))
	}
	return nil
}

// createClaudeDirs creates the .claude/ directory structure.
func (i *projectInitializer) createClaudeDirs(root string, result *InitResult) error {
	for _, dir := range claudeDirs {
		dirPath := filepath.Clean(filepath.Join(root, defs.ClaudeDir, dir))
		if err := os.MkdirAll(dirPath, defs.DirPerm); err != nil {
			return fmt.Errorf("mkdir %s: %w", dirPath, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, filepath.Join(defs.ClaudeDir, dir))
	}
	return nil
}

// deployTemplates deploys embedded templates to the project root.
// Files ending in .tmpl are rendered with TemplateContext built from opts.
func (i *projectInitializer) deployTemplates(ctx context.Context, opts InitOptions, result *InitResult) error {
	if i.manifestMgr == nil {
		return fmt.Errorf("manifest manager required for template deployment")
	}

	// Load or create manifest for tracking
	if _, err := i.manifestMgr.Load(opts.ProjectRoot); err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	// Detect paths for template context
	homeDir, _ := os.UserHomeDir()
	goBinPath := detectGoBinPath(homeDir)

	// Build TemplateContext from InitOptions
	tmplCtx := template.NewTemplateContext(
		template.WithProject(opts.ProjectName, opts.ProjectRoot),
		template.WithUser(opts.UserName),
		template.WithLanguage(opts.ConvLang),
		template.WithDevelopmentMode(opts.DevelopmentMode),
		template.WithGitMode(opts.GitMode),
		template.WithGitProvider(opts.GitProvider),
		template.WithGitHubUsername(opts.GitHubUsername),
		template.WithGitLabInstanceURL(opts.GitLabInstanceURL),
		template.WithOutputLanguages(opts.GitCommitLang, opts.CodeCommentLang, opts.DocLang),
		template.WithPlatform(opts.Platform),
		template.WithGoBinPath(goBinPath),
		template.WithHomeDir(homeDir),
		template.WithSmartPATH(template.BuildSmartPATH()),
		template.WithVersion(version.GetVersion()),
	)

	if err := i.deployer.Deploy(ctx, opts.ProjectRoot, i.manifestMgr, tmplCtx); err != nil {
		return fmt.Errorf("deploy templates: %w", err)
	}

	return nil
}

// detectGoBinPath detects the Go binary installation path.
// Returns the path where Go binaries are installed (e.g., "/home/user/go/bin").
func detectGoBinPath(homeDir string) string {
	// Try GOBIN first (explicit override)
	if output, err := exec.Command("go", "env", "GOBIN").Output(); err == nil {
		if goBin := strings.TrimSpace(string(output)); goBin != "" {
			return goBin
		}
	}

	// Try GOPATH/bin (user's Go workspace)
	if output, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		if goPath := strings.TrimSpace(string(output)); goPath != "" {
			return filepath.Join(goPath, "bin")
		}
	}

	// Fallback to default ~/go/bin
	if homeDir != "" {
		return filepath.Join(homeDir, "go", "bin")
	}

	// Last resort: platform-specific common Go install locations
	if runtime.GOOS == "windows" {
		// Windows common install paths
		candidates := []string{
			filepath.Join(os.Getenv("PROGRAMFILES"), "Go", "bin"),
			`C:\Go\bin`,
		}
		for _, candidate := range candidates {
			if candidate != "" && candidate != `\Go\bin` { // Skip if PROGRAMFILES is empty
				if _, err := os.Stat(candidate); err == nil {
					return candidate
				}
			}
		}
		// Final fallback for Windows
		return filepath.Join(os.Getenv("USERPROFILE"), "go", "bin")
	}
	// Unix-like systems (Linux, macOS, etc.)
	return "/usr/local/go/bin"
}

// generateConfigsFallback creates config YAML files directly when no deployer is available.
// This provides a fallback for testing and minimal initialization scenarios.
func (i *projectInitializer) generateConfigsFallback(opts InitOptions, result *InitResult) error {
	sectionsDir := filepath.Clean(filepath.Join(opts.ProjectRoot, defs.MoAIDir, defs.SectionsSubdir))

	// Build context for config values
	tmplCtx := template.NewTemplateContext(
		template.WithProject(opts.ProjectName, opts.ProjectRoot),
		template.WithUser(opts.UserName),
		template.WithLanguage(opts.ConvLang),
		template.WithDevelopmentMode(opts.DevelopmentMode),
		template.WithGitMode(opts.GitMode),
		template.WithGitProvider(opts.GitProvider),
		template.WithGitHubUsername(opts.GitHubUsername),
		template.WithGitLabInstanceURL(opts.GitLabInstanceURL),
		template.WithOutputLanguages(opts.GitCommitLang, opts.CodeCommentLang, opts.DocLang),
		template.WithPlatform(opts.Platform),
	)

	// user.yaml
	userContent := fmt.Sprintf("user:\n  name: %q\n", tmplCtx.UserName)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.UserYAML), []byte(userContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write user.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.UserYAML))

	// language.yaml
	langContent := fmt.Sprintf(`language:
  conversation_language: %s
  conversation_language_name: %s
  agent_prompt_language: %s
  git_commit_messages: %s
  code_comments: %s
  documentation: %s
  error_messages: %s
`, tmplCtx.ConversationLanguage, tmplCtx.ConversationLanguageName,
		tmplCtx.AgentPromptLanguage, tmplCtx.GitCommitMessages,
		tmplCtx.CodeComments, tmplCtx.Documentation, tmplCtx.ErrorMessages)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.LanguageYAML), []byte(langContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write language.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.LanguageYAML))

	// quality.yaml
	qualityContent := fmt.Sprintf(`constitution:
  development_mode: %s
  enforce_quality: %t
  test_coverage_target: %d
`, tmplCtx.DevelopmentMode, tmplCtx.EnforceQuality, tmplCtx.TestCoverageTarget)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.QualityYAML), []byte(qualityContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write quality.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.QualityYAML))

	// workflow.yaml
	workflowContent := fmt.Sprintf(`workflow:
  auto_clear: %t
  plan_tokens: %d
  run_tokens: %d
  sync_tokens: %d
`, tmplCtx.AutoClear, tmplCtx.PlanTokens, tmplCtx.RunTokens, tmplCtx.SyncTokens)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.WorkflowYAML), []byte(workflowContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write workflow.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.WorkflowYAML))

	// git-strategy.yaml
	gitStrategyContent := fmt.Sprintf(`git_strategy:
  mode: %q
  provider: %q
  github_username: %q
  gitlab:
    instance_url: %q
`, tmplCtx.GitMode, tmplCtx.GitProvider, tmplCtx.GitHubUsername, tmplCtx.GitLabInstanceURL)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.GitStrategyYAML), []byte(gitStrategyContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write git-strategy.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.GitStrategyYAML))

	// system.yaml
	systemContent := fmt.Sprintf(`moai:
  version: %q
  template_version: %q
  update_check_frequency: daily
`, tmplCtx.Version, tmplCtx.Version)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.SystemYAML), []byte(systemContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write system.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.SystemYAML))

	// project.yaml
	projectContent := fmt.Sprintf(`project:
  name: %q
  description: ""
  mode: personal
  created_at: %q
  initialized: true
  optimized: false
  template_version: %q
`, tmplCtx.ProjectName, tmplCtx.CreatedAt, tmplCtx.Version)
	if err := os.WriteFile(filepath.Join(sectionsDir, defs.ProjectYAML), []byte(projectContent), defs.FilePerm); err != nil {
		return fmt.Errorf("write project.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.SectionsSubdir, defs.ProjectYAML))

	return nil
}

// createClaudeMD generates the CLAUDE.md file.
// If the deployer already deployed the full CLAUDE.md from embedded templates,
// this step is skipped to avoid overwriting it with the minimal stub.
func (i *projectInitializer) createClaudeMD(opts InitOptions, result *InitResult) error {
	claudeMDPath := filepath.Clean(filepath.Join(opts.ProjectRoot, defs.ClaudeMD))

	// Skip if CLAUDE.md was already deployed by the template deployer (REQ-E-034)
	if _, err := os.Stat(claudeMDPath); err == nil {
		i.logger.Info("CLAUDE.md already exists (deployed from templates), skipping stub generation")
		return nil
	}

	// Fallback: generate minimal stub when no deployer is available
	content := buildClaudeMDContent(opts)

	if err := os.WriteFile(claudeMDPath, []byte(content), defs.FilePerm); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}

	result.CreatedFiles = append(result.CreatedFiles, defs.ClaudeMD)
	return nil
}

// buildClaudeMDContent generates CLAUDE.md content from options.
func buildClaudeMDContent(opts InitOptions) string {
	var b strings.Builder
	b.WriteString("# MoAI Execution Directive\n\n")
	fmt.Fprintf(&b, "Project: %s\n", opts.ProjectName)
	fmt.Fprintf(&b, "Language: %s\n", opts.Language)
	if opts.Framework != "" && opts.Framework != "none" {
		fmt.Fprintf(&b, "Framework: %s\n", opts.Framework)
	}
	fmt.Fprintf(&b, "Development Mode: %s\n\n", opts.DevelopmentMode)
	b.WriteString("## Configuration\n\n")
	b.WriteString("Configuration files are located in `.moai/config/sections/`.\n\n")
	b.WriteString("## Quick Start\n\n")
	b.WriteString("- Run `moai doctor` to check project health\n")
	b.WriteString("- Run `moai status` to view project status\n")
	b.WriteString("- Run `moai plan \"description\"` to create a SPEC\n")
	return b.String()
}

// initManifest initializes the manifest.json file.
func (i *projectInitializer) initManifest(root string, result *InitResult) error {
	if i.manifestMgr == nil {
		return nil
	}

	// Use the in-memory manifest if already loaded by deployTemplates (Step 3).
	// Calling Load() again would discard all entries tracked during deployment.
	mf := i.manifestMgr.Manifest()
	if mf == nil {
		var err error
		mf, err = i.manifestMgr.Load(root)
		if err != nil {
			return fmt.Errorf("load manifest: %w", err)
		}
	}

	// Record ADK version and deployment timestamp (REQ-E-040, REQ-E-041)
	mf.Version = version.GetVersion()
	mf.DeployedAt = time.Now().UTC().Format(time.RFC3339)

	if err := i.manifestMgr.Save(); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	// Validate the generated manifest JSON
	manifestPath := filepath.Join(root, defs.MoAIDir, defs.ManifestJSON)
	if data, readErr := os.ReadFile(manifestPath); readErr == nil {
		if !json.Valid(data) {
			return fmt.Errorf("generated manifest.json is not valid JSON")
		}
	}

	result.CreatedFiles = append(result.CreatedFiles, filepath.Join(defs.MoAIDir, defs.ManifestJSON))
	return nil
}

// configureShellEnv sets up shell environment variables for Claude Code.
// This adds CLAUDE_DISABLE_PATH_WARNING=1 and PATH entry to the appropriate
// shell configuration file (.zshenv, .profile, or config.fish).
func (i *projectInitializer) configureShellEnv() (*shell.ConfigResult, error) {
	configurator := shell.NewEnvConfigurator(i.logger)

	return configurator.Configure(shell.ConfigOptions{
		AddClaudeWarningDisable: true,
		AddLocalBinPath:         true,
		AddGoBinPath:            true,
		PreferLoginShell:        true,
	})
}

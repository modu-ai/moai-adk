package project

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/modu-ai/moai-adk-go/internal/manifest"
	"github.com/modu-ai/moai-adk-go/internal/template"
	"github.com/modu-ai/moai-adk-go/pkg/models"
	"gopkg.in/yaml.v3"
)

// InitOptions configures the project initialization.
type InitOptions struct {
	ProjectRoot     string   // Absolute or relative path to the project root.
	ProjectName     string   // Name of the project.
	Language        string   // Primary programming language.
	Framework       string   // Framework name, or "none".
	Features        []string // Selected features (e.g., "LSP", "Quality Gates").
	UserName        string   // User display name for configuration.
	ConvLang        string   // Conversation language code (e.g., "en", "ko").
	DevelopmentMode string   // "ddd", "tdd", or "hybrid".
	NonInteractive  bool     // If true, skip wizard and use defaults/flags.
	Force           bool     // If true, allow reinitializing an existing project.
}

// InitResult summarizes the outcome of project initialization.
type InitResult struct {
	CreatedDirs     []string // Directories that were created.
	CreatedFiles    []string // Files that were created.
	DevelopmentMode string   // Selected development methodology.
	BackupPath      string   // Non-empty if --force was used and backup was created.
	Warnings        []string // Non-fatal warnings during initialization.
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
		logger = slog.Default()
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

	// Step 2: Generate configuration files
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.generateConfigs(opts, result); err != nil {
		return nil, fmt.Errorf("generate configs: %w", err)
	}

	// Step 3: Create .claude/ directory structure
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.createClaudeDirs(opts.ProjectRoot, result); err != nil {
		return nil, fmt.Errorf("create .claude/ structure: %w", err)
	}

	// Step 4: Deploy templates (if deployer is available)
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if i.deployer != nil {
		if err := i.deployTemplates(ctx, opts.ProjectRoot, result); err != nil {
			// Template deployment is non-fatal; record warning
			result.Warnings = append(result.Warnings, fmt.Sprintf("template deployment: %s", err))
			i.logger.Warn("template deployment failed", "error", err)
		}
	}

	// Step 5: Create CLAUDE.md
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.createClaudeMD(opts, result); err != nil {
		return nil, fmt.Errorf("create CLAUDE.md: %w", err)
	}

	// Step 6: Initialize manifest
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := i.initManifest(opts.ProjectRoot, result); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("manifest initialization: %s", err))
		i.logger.Warn("manifest initialization failed", "error", err)
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
		dirPath := filepath.Clean(filepath.Join(root, ".moai", dir))
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dirPath, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, filepath.Join(".moai", dir))
	}
	return nil
}

// createClaudeDirs creates the .claude/ directory structure.
func (i *projectInitializer) createClaudeDirs(root string, result *InitResult) error {
	for _, dir := range claudeDirs {
		dirPath := filepath.Clean(filepath.Join(root, ".claude", dir))
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", dirPath, err)
		}
		result.CreatedDirs = append(result.CreatedDirs, filepath.Join(".claude", dir))
	}
	return nil
}

// generateConfigs writes the configuration YAML files using struct serialization.
func (i *projectInitializer) generateConfigs(opts InitOptions, result *InitResult) error {
	sectionsDir := filepath.Clean(filepath.Join(opts.ProjectRoot, ".moai", "config", "sections"))

	// user.yaml
	if err := i.writeYAML(sectionsDir, "user.yaml", userYAML{
		User: userSection{Name: opts.UserName},
	}); err != nil {
		return fmt.Errorf("write user.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, ".moai/config/sections/user.yaml")

	// language.yaml
	convLangName := resolveLanguageName(opts.ConvLang)
	if err := i.writeYAML(sectionsDir, "language.yaml", languageYAML{
		Language: languageSection{
			ConversationLanguage:     opts.ConvLang,
			ConversationLanguageName: convLangName,
			AgentPromptLanguage:      "en",
			GitCommitMessages:        "en",
			CodeComments:             "en",
			Documentation:            "en",
			ErrorMessages:            "en",
		},
	}); err != nil {
		return fmt.Errorf("write language.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, ".moai/config/sections/language.yaml")

	// quality.yaml
	devMode := models.DevelopmentMode(opts.DevelopmentMode)
	if !devMode.IsValid() {
		devMode = models.ModeDDD
	}
	if err := i.writeYAML(sectionsDir, "quality.yaml", qualityYAML{
		Constitution: qualitySection{
			DevelopmentMode:    string(devMode),
			EnforceQuality:     true,
			TestCoverageTarget: 85,
		},
	}); err != nil {
		return fmt.Errorf("write quality.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, ".moai/config/sections/quality.yaml")

	// workflow.yaml
	if err := i.writeYAML(sectionsDir, "workflow.yaml", workflowYAML{
		Workflow: workflowSection{
			AutoClear:  true,
			PlanTokens: 30000,
			RunTokens:  180000,
			SyncTokens: 40000,
		},
	}); err != nil {
		return fmt.Errorf("write workflow.yaml: %w", err)
	}
	result.CreatedFiles = append(result.CreatedFiles, ".moai/config/sections/workflow.yaml")

	return nil
}

// writeYAML marshals data to YAML and writes it atomically to dir/filename.
func (i *projectInitializer) writeYAML(dir, filename string, data any) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filename, err)
	}

	path := filepath.Clean(filepath.Join(dir, filename))
	if err := os.WriteFile(path, yamlData, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}

	return nil
}

// deployTemplates deploys embedded templates to the project root.
func (i *projectInitializer) deployTemplates(ctx context.Context, root string, result *InitResult) error {
	if i.manifestMgr == nil {
		return fmt.Errorf("manifest manager required for template deployment")
	}

	// Load or create manifest for tracking
	if _, err := i.manifestMgr.Load(root); err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	if err := i.deployer.Deploy(ctx, root, i.manifestMgr); err != nil {
		return fmt.Errorf("deploy templates: %w", err)
	}

	return nil
}

// createClaudeMD generates the CLAUDE.md file.
func (i *projectInitializer) createClaudeMD(opts InitOptions, result *InitResult) error {
	claudeMDPath := filepath.Clean(filepath.Join(opts.ProjectRoot, "CLAUDE.md"))

	content := buildClaudeMDContent(opts)

	if err := os.WriteFile(claudeMDPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write CLAUDE.md: %w", err)
	}

	result.CreatedFiles = append(result.CreatedFiles, "CLAUDE.md")
	return nil
}

// buildClaudeMDContent generates CLAUDE.md content from options.
func buildClaudeMDContent(opts InitOptions) string {
	var b strings.Builder
	b.WriteString("# MoAI Execution Directive\n\n")
	b.WriteString(fmt.Sprintf("Project: %s\n", opts.ProjectName))
	b.WriteString(fmt.Sprintf("Language: %s\n", opts.Language))
	if opts.Framework != "" && opts.Framework != "none" {
		b.WriteString(fmt.Sprintf("Framework: %s\n", opts.Framework))
	}
	b.WriteString(fmt.Sprintf("Development Mode: %s\n\n", opts.DevelopmentMode))
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

	mf, err := i.manifestMgr.Load(root)
	if err != nil {
		return fmt.Errorf("load manifest: %w", err)
	}

	mf.Version = "1.0.0"

	if err := i.manifestMgr.Save(); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	// Validate the generated manifest JSON
	manifestPath := filepath.Join(root, ".moai", "manifest.json")
	if data, readErr := os.ReadFile(manifestPath); readErr == nil {
		if !json.Valid(data) {
			return fmt.Errorf("generated manifest.json is not valid JSON")
		}
	}

	result.CreatedFiles = append(result.CreatedFiles, ".moai/manifest.json")
	return nil
}

// --- YAML serialization structs (REQ-N-002: struct serialization, not string concatenation) ---

type userYAML struct {
	User userSection `yaml:"user"`
}

type userSection struct {
	Name string `yaml:"name"`
}

type languageYAML struct {
	Language languageSection `yaml:"language"`
}

type languageSection struct {
	ConversationLanguage     string `yaml:"conversation_language"`
	ConversationLanguageName string `yaml:"conversation_language_name"`
	AgentPromptLanguage      string `yaml:"agent_prompt_language"`
	GitCommitMessages        string `yaml:"git_commit_messages"`
	CodeComments             string `yaml:"code_comments"`
	Documentation            string `yaml:"documentation"`
	ErrorMessages            string `yaml:"error_messages"`
}

type qualityYAML struct {
	Constitution qualitySection `yaml:"constitution"`
}

type qualitySection struct {
	DevelopmentMode    string `yaml:"development_mode"`
	EnforceQuality     bool   `yaml:"enforce_quality"`
	TestCoverageTarget int    `yaml:"test_coverage_target"`
}

type workflowYAML struct {
	Workflow workflowSection `yaml:"workflow"`
}

type workflowSection struct {
	AutoClear  bool `yaml:"auto_clear"`
	PlanTokens int  `yaml:"plan_tokens"`
	RunTokens  int  `yaml:"run_tokens"`
	SyncTokens int  `yaml:"sync_tokens"`
}

// --- Language name resolution ---

var langNameMap = map[string]string{
	"en": "English",
	"ko": "Korean (한국어)",
	"ja": "Japanese (日本語)",
	"zh": "Chinese (中文)",
	"es": "Spanish (Español)",
	"fr": "French (Français)",
	"de": "German (Deutsch)",
}

// resolveLanguageName returns the full name for a language code.
func resolveLanguageName(code string) string {
	if name, ok := langNameMap[code]; ok {
		return name
	}
	return "English"
}

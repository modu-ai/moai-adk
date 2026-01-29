package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	globalConfig *Config
	configLoaded bool
)

// LoadConfig loads configuration from .moai/config/sections/ directory
// It reads all YAML section files and merges them into a single Config struct.
// Environment variables with MOAI_ prefix override config file values.
func LoadConfig(projectDir string) (*Config, error) {
	// Path to .moai/config/sections/
	configPath := filepath.Join(projectDir, ".moai", "config", "sections")

	// Check if config directory exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if directory doesn't exist
		config := getDefaultConfig()
		globalConfig = config
		configLoaded = true
		return config, nil
	}

	// Initialize empty config
	config := getDefaultConfig()

	// Section files to load (in order)
	sections := []string{
		"user",
		"language",
		"quality",
		"system",
		"git-strategy",
		"project",
		"llm",
		"pricing",
		"ralph",
		"workflow",
	}

	// Load each section file
	for _, section := range sections {
		filePath := filepath.Join(configPath, section+".yaml")

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			continue // Skip missing files
		}

		// Read file content
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s.yaml: %w", section, err)
		}

		// Parse YAML
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse %s.yaml: %w", section, err)
		}
	}

	// Apply environment variable overrides
	applyEnvVarOverrides(config)

	globalConfig = config
	configLoaded = true

	return config, nil
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *Config {
	return &Config{
		User:         UserConfig{Name: ""},
		Language:     getDefaultLanguageConfig(),
		Constitution: getDefaultConstitutionConfig(),
		System:       getDefaultSystemConfig(),
		GitStrategy:  getDefaultGitStrategyConfig(),
		Project:      ProjectConfig{},
		LLM:          getDefaultLLMConfig(),
		Service:      getDefaultServiceConfig(),
		Ralph:        getDefaultRalphConfig(),
		Workflow:     getDefaultWorkflowConfig(),
	}
}

// applyEnvVarOverrides manually applies environment variable overrides
func applyEnvVarOverrides(config *Config) {
	// User settings
	if val := os.Getenv("MOAI_USER_NAME"); val != "" {
		config.User.Name = val
	}

	// Language settings
	if val := os.Getenv("MOAI_CONVERSATION_LANG"); val != "" {
		config.Language.ConversationLanguage = val
	}
	if val := os.Getenv("MOAI_CODE_COMMENTS_LANG"); val != "" {
		config.Language.CodeComments = val
	}
	if val := os.Getenv("MOAI_GIT_COMMIT_MESSAGES_LANG"); val != "" {
		config.Language.GitCommitMessages = val
	}
	if val := os.Getenv("MOAI_DOCUMENTATION_LANG"); val != "" {
		config.Language.Documentation = val
	}
	if val := os.Getenv("MOAI_ERROR_MESSAGES_LANG"); val != "" {
		config.Language.ErrorMessages = val
	}
}

// GetConfig returns the globally loaded configuration
// Returns nil if LoadConfig has not been called
func GetConfig() *Config {
	return globalConfig
}

// IsConfigLoaded returns true if configuration has been loaded
func IsConfigLoaded() bool {
	return configLoaded
}

// LoadConfigFromEnv loads configuration with environment variable overrides
// This is a convenience function that loads config and applies env vars
func LoadConfigFromEnv(projectDir string) (*Config, error) {
	return LoadConfig(projectDir)
}

// GetConfigPath returns the path to the config sections directory
func GetConfigPath(projectDir string) string {
	return filepath.Join(projectDir, ".moai", "config", "sections")
}

// ConfigFileExists checks if a config file exists
func ConfigFileExists(projectDir, section string) bool {
	configPath := filepath.Join(projectDir, ".moai", "config", "sections", section+".yaml")
	_, err := os.Stat(configPath)
	return err == nil
}

// GetMoaiDir returns the path to the .moai directory
func GetMoaiDir(projectDir string) string {
	return filepath.Join(projectDir, ".moai")
}

// normalizeLanguageCode converts language codes to locale format
// e.g., "ko" -> "ko_KR", "en" -> "en_US"
func normalizeLanguageCode(lang string) string {
	langMap := map[string]string{
		"ko": "ko_KR",
		"en": "en_US",
		"ja": "ja_JP",
		"zh": "zh_CN",
		"es": "es_ES",
		"fr": "fr_FR",
		"de": "de_DE",
	}

	normalized := strings.ToLower(lang)
	if val, ok := langMap[normalized]; ok {
		return val
	}

	// Default to lang_US format if not in map
	if len(lang) == 2 {
		return lang + "_" + strings.ToUpper(lang)
	}

	return lang
}

// Default configuration providers

func getDefaultLanguageConfig() LanguageConfig {
	return LanguageConfig{
		ConversationLanguage:     "en",
		ConversationLanguageName: "English",
		AgentPromptLanguage:      "en",
		GitCommitMessages:        "en",
		CodeComments:             "en",
		Documentation:            "en",
		ErrorMessages:            "en",
	}
}

func getDefaultConstitutionConfig() ConstitutionConfig {
	return ConstitutionConfig{
		DevelopmentMode:    "ddd",
		EnforceQuality:     true,
		TestCoverageTarget: 85,
		DDDDSettings: DDDSettings{
			RequireExistingTests:  true,
			CharacterizationTests: true,
			BehaviorSnapshots:     true,
			MaxTransformationSize: "small",
		},
		CoverageExemptions: CoverageExemptions{
			Enabled:              false,
			RequireJustification: true,
			MaxExemptPercentage:  5,
		},
		TestQuality: TestQuality{
			SpecificationBased:     true,
			MeaningfulAssertions:   true,
			AvoidImplCoupling:      true,
			MutationTestingEnabled: false,
		},
		LSPPriorityGates: LSPPriorityGates{
			Enabled:         true,
			CacheTTLSeconds: 5,
			TimeoutSeconds:  3,
			Plan:            PhaseGates{RequireBaseline: true},
			Run:             RunGates{MaxErrors: 0, MaxTypeErrors: 0, MaxLintErrors: 0, AllowRegression: false},
			Sync:            SyncGates{MaxErrors: 0, MaxWarnings: 10, RequireCleanLSP: true},
		},
		LSPIntegration: LSPIntegration{
			TRUST5Integration: map[string][]string{
				"tested":    {"unit_tests_pass", "lsp_type_errors == 0", "lsp_errors == 0"},
				"readable":  {"naming_conventions_followed", "lsp_lint_errors == 0"},
				"secured":   {"security_scan_pass", "lsp_security_warnings == 0"},
				"trackable": {"logs_structured", "lsp_diagnostic_history_tracked"},
			},
			DiagnosticSources: []string{"typecheck", "lint", "security"},
			RegressionDetection: RegressionDetection{
				ErrorIncreaseThreshold:     0,
				WarningIncreaseThreshold:   10,
				TypeErrorIncreaseThreshold: 0,
			},
		},
		ReportGeneration: ReportGeneration{
			Enabled:    true,
			AutoCreate: false,
			WarnUser:   true,
			UserChoice: "Minimal",
		},
		LSPStateTracking: LSPStateTracking{
			Enabled:       true,
			CapturePoints: []string{"phase_start", "post_transformation", "pre_sync"},
			Comparison: struct {
				Baseline            string `yaml:"baseline"`
				RegressionThreshold int    `yaml:"regression_threshold"`
			}{
				Baseline:            "phase_start",
				RegressionThreshold: 0,
			},
			Logging: struct {
				LogLSPStateChanges     bool `yaml:"log_lsp_state_changes"`
				LogRegressionDetection bool `yaml:"log_regression_detection"`
				LogCompletionMarkers   bool `yaml:"log_completion_markers"`
				IncludeLSPInReports    bool `yaml:"include_lsp_in_reports"`
			}{
				LogLSPStateChanges:     true,
				LogRegressionDetection: true,
				LogCompletionMarkers:   true,
				IncludeLSPInReports:    true,
			},
		},
	}
}

func getDefaultSystemConfig() SystemConfig {
	return SystemConfig{
		MoAI: MoAIConfig{
			Version:              "dev",
			UpdateCheckFrequency: "daily",
			VersionCheck: VersionCheckConfig{
				Enabled:       true,
				CacheTTLHours: 24,
			},
		},
		GitHub: GitHubConfig{
			EnableTrust5:       true,
			AutoDeleteBranches: true,
			SpecGitWorkflow:    "main_direct",
		},
		DocumentManagement: DocumentManagementConfig{
			Enabled:            true,
			EnforceStructure:   true,
			BlockRootPollution: true,
			SeparationPolicy: SeparationPolicyConfig{
				StrictSeparation: true,
				Description:      "Clear separation between logs (runtime data) and docs (documentation content)",
				LogsScope:        "runtime sessions, errors, execution traces, agent transcripts",
				DocsScope:        "reports, analysis, validation, sync, inspection, documentation files",
			},
			Directories: DirectoriesConfig{
				Docs: DirectoryConfig{
					Base:          ".moai/docs/",
					RetentionDays: nil, // nil means keep indefinitely
				},
				Logs: DirectoryConfig{
					Base:          ".moai/logs/",
					RetentionDays: func() *int { i := 30; return &i }(),
					AutoCleanup:   false,
				},
				Temp: DirectoryConfig{
					Base:          ".moai/temp/",
					RetentionDays: func() *int { i := 7; return &i }(),
					AutoCleanup:   true,
				},
				Cache: DirectoryConfig{
					Base:          ".moai/cache/",
					RetentionDays: func() *int { i := 30; return &i }(),
					AutoCleanup:   true,
				},
			},
		},
	}
}

func getDefaultGitStrategyConfig() GitStrategyConfig {
	return GitStrategyConfig{
		Mode: "manual",
		Manual: GitModeConfig{
			Workflow:          "github-flow",
			Environment:       "local",
			GitHubIntegration: false,
			AutoCheckpoint:    "disabled",
			PushToRemote:      false,
			BranchCreation:    BranchCreationConfig{PromptAlways: true, AutoEnabled: false},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
		},
		Personal: GitModeConfig{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			BranchCreation:    BranchCreationConfig{PromptAlways: true, AutoEnabled: false},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
		},
		Team: GitTeamConfig{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			DraftPR:           true,
			RequiredReviews:   1,
			BranchProtection:  true,
			BranchCreation:    BranchCreationConfig{PromptAlways: false, AutoEnabled: true},
			Automation:        AutomationConfig{AutoBranch: true, AutoCommit: true, AutoPR: true, AutoPush: true},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: true},
		},
	}
}

func getDefaultLLMConfig() LLMConfig {
	return LLMConfig{
		Mode:      "claude-only",
		GLMEnvVar: "GLM_API_KEY",
		AutoWorktree: AutoWorktreeConfig{
			Enabled:       false,
			CopyGLMConfig: true,
		},
		GLM: GLMConfig{
			BaseURL: "https://api.z.ai/api/anthropic",
			Models: map[string]string{
				"haiku":  "glm-4.7-flashx",
				"sonnet": "glm-4.7",
				"opus":   "glm-4.7",
			},
		},
		Routing: RoutingConfig{
			AutoDetect:        true,
			ParallelThreshold: 10,
			ConfirmWorktree:   false,
		},
	}
}

func getDefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		Type:        "claude_subscription",
		PricingPlan: "pro",
		ModelAllocation: ModelAllocationConfig{
			Strategy: "auto",
			Custom:   nil,
		},
	}
}

func getDefaultRalphConfig() RalphConfig {
	return RalphConfig{
		Enabled: true,
		LSP: RalphLSPConfig{
			AutoStart:           true,
			TimeoutSeconds:      15,
			PollIntervalMs:      1000,
			GracefulDegradation: true,
		},
		ASTGrep: RalphASTGrepConfig{
			Enabled:      true,
			ConfigPath:   ".claude/skills/moai-tool-ast-grep/rules/sgconfig.yml",
			SecurityScan: true,
			QualityScan:  true,
			AutoFix:      false,
		},
		Loop: RalphLoopConfig{
			MaxIterations:       10,
			AutoFix:             false,
			RequireConfirmation: true,
			CooldownSeconds:     2,
			Completion: RalphCompletionConfig{
				ZeroErrors:        true,
				ZeroWarnings:      false,
				TestsPass:         true,
				CoverageThreshold: 85,
			},
		},
		Hooks: RalphHooksConfig{
			PostToolLSP: RalphPostToolLSPConfig{
				Enabled:           true,
				TriggerOn:         []string{"Write", "Edit"},
				SeverityThreshold: "error",
			},
			StopLoopController: RalphStopLoopConfig{
				Enabled:         true,
				CheckCompletion: true,
			},
		},
	}
}

func getDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		ExecutionMode: WorkflowExecutionConfig{
			Interactive: WorkflowModeConfig{
				UserApprovalRequired: true,
				EachStepExplicit:     true,
				ShowProgress:         true,
			},
			Autonomous: WorkflowModeConfig{
				UserApprovalRequired:   false,
				ContinuousLoop:         true,
				CompletionMarkerBased:  true,
				LSPFeedbackIntegration: true,
				ShowProgress:           true,
			},
		},
		CompletionMarkers: CompletionMarkersConfig{
			Plan: PhaseMarkersConfig{
				SpecDocumentCreated: true,
				LSPBaselineRecorded: true,
			},
			Run: PhaseMarkersConfig{
				TestsPassing:      true,
				BehaviorPreserved: true,
				LSPErrors:         0,
				TypeErrors:        0,
				CoverageMetTarget: true,
			},
			Sync: PhaseMarkersConfig{
				DocumentationGenerated: true,
				LSPClean:               true,
				QualityGatePassed:      true,
			},
		},
		LoopPrevention: LoopPreventionConfig{
			MaxIterations:       100,
			NoProgressThreshold: 5,
			StaleDetection: []string{
				"error_count_unchanged",
				"same_fix_attempted_twice",
				"alternative_strategy_required",
			},
		},
	}
}

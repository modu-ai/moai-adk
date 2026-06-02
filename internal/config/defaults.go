package config

import (
	"github.com/modu-ai/moai-adk/pkg/models"
)

// Default value constants to avoid magic numbers and strings.
const (
	DefaultConversationLanguage     = "en"
	DefaultConversationLanguageName = "English"
	DefaultAgentPromptLanguage      = "en"
	DefaultGitCommitMessages        = "en"
	DefaultCodeComments             = "en"
	DefaultDocumentation            = "en"
	DefaultErrorMessages            = "en"

	DefaultTestCoverageTarget    = 85
	DefaultMaxTransformationSize = "small"
	DefaultMinCoveragePerCommit  = 80
	DefaultMaxExemptPercentage   = 5

	DefaultLogLevel  = "info"
	DefaultLogFormat = "text"

	DefaultModel      = "sonnet"
	DefaultQualModel  = "opus"
	DefaultSpeedModel = "haiku"

	DefaultTokenBudget = 250000

	DefaultMaxIterations = 5

	DefaultPlanTokens = 30000
	DefaultRunTokens  = 180000
	DefaultSyncTokens = 40000

	DefaultBranchPrefix = "moai/"
	DefaultCommitStyle  = "conventional"

	DefaultGLMEnvVar  = "GLM_API_KEY"
	DefaultGLMBaseURL = "https://api.z.ai/api/anthropic"
	// GLM model tiers
	DefaultGLMHigh   = "glm-5.1"
	DefaultGLMMedium = "glm-4.7"
	DefaultGLMLow    = "glm-4.5-air"
	// Additional GLM models (available but not default-mapped)
	DefaultGLM45     = "glm-4.5"
	DefaultGLM46     = "glm-4.6"
	DefaultGLM5Turbo = "glm-5-turbo"
	// Legacy GLM model names (map to tiers)
	DefaultGLMHaiku  = "glm-4.5-air"
	DefaultGLMSonnet = "glm-4.7"
	DefaultGLMOpus   = "glm-5.1"
	// Default performance tier
	DefaultPerformanceTier = "medium"

	DefaultCacheTTLSeconds = 5
	DefaultTimeoutSeconds  = 3
	DefaultMaxWarnings     = 10

	DefaultGitConvention                    = "auto"
	DefaultGitConventionSampleSize          = 100
	DefaultGitConventionConfidenceThreshold = 0.5
	DefaultGitConventionFallback            = "conventional-commits"
	DefaultGitConventionMaxLength           = 100

	DefaultStateDir = ".moai/state"

	// DefaultStaleSeconds is the threshold (in seconds) at which a session checkpoint is considered stale.
	// SPEC-V3R2-RT-004 REQ-022: overridable via the stale_seconds key in ralph.yaml.
	DefaultStaleSeconds = 3600

	// Memory taxonomy defaults (SPEC-V3R2-EXT-001)
	DefaultMemoryStalenessHours          = 24  // files older than this are wrapped in staleness caveat
	DefaultMemoryIndexLineCap            = 200 // MEMORY.md lines beyond this trigger MEMORY_INDEX_OVERFLOW
	DefaultMemoryStaleAggregateThreshold = 10  // stale files >= this count emit one aggregated warning
)

// NewDefaultConfig returns a Config with all fields set to compiled defaults.
func NewDefaultConfig() *Config {
	return &Config{
		User:          NewDefaultUserConfig(),
		Language:      NewDefaultLanguageConfig(),
		Quality:       NewDefaultQualityConfig(),
		Project:       NewDefaultProjectConfig(),
		GitStrategy:   NewDefaultGitStrategyConfig(),
		GitConvention: NewDefaultGitConventionConfig(),
		System:        NewDefaultSystemConfig(),
		LLM:           NewDefaultLLMConfig(),
		Pricing:       NewDefaultPricingConfig(),
		Ralph:         NewDefaultRalphConfig(),
		Workflow:      NewDefaultWorkflowConfig(),
		State:         NewDefaultStateConfig(),
		Gate:          NewDefaultGateConfig(),
		Sunset:        NewDefaultSunsetConfig(),
		Research:      NewDefaultResearchConfig(),
		Session:       NewDefaultSessionConfig(),
		// MIG-003: 4 new section defaults (REQ-MIG003-004)
		Constitution:  defaultConstitutionConfig(),
		ContextSearch: defaultContextConfig(),
		Interview:     defaultInterviewConfig(),
		Design:        defaultDesignConfig(),
	}
}

// NewDefaultResearchConfig returns a ResearchConfig with safe defaults.
func NewDefaultResearchConfig() ResearchConfig {
	return ResearchConfig{
		Enabled: false,
		Passive: ResearchPassiveConfig{
			Enabled:                 true,
			CorrectionWindowSeconds: 60,
			PatternThresholds: ResearchPatternThresholds{
				Heuristic:      3,
				Rule:           5,
				HighConfidence: 10,
			},
		},
		Active: ResearchActiveConfig{
			RunsPerExperiment: 3,
			MaxExperiments:    20,
			PassThreshold:     0.80,
			TargetScore:       0.95,
			BudgetCapTokens:   500000,
		},
		Safety: ResearchSafetyConfig{
			WorktreeIsolation:         true,
			CanaryRegressionThreshold: 0.10,
			RateLimits: ResearchRateLimitConfig{
				MaxExperimentsPerSession: 20,
				MaxAcceptedPerSession:    5,
				MaxAutoResearchPerWeek:   3,
			},
		},
		Dashboard: ResearchDashboardConfig{
			DefaultMode:     "terminal",
			HTMLOpenBrowser: true,
		},
	}
}

// NewDefaultGateConfig returns a GateConfig with production-safe defaults.
func NewDefaultGateConfig() GateConfig {
	return GateConfig{
		Enabled:   true,
		SkipTests: false,
		Timeouts: GateTimeouts{
			Vet:  30,
			Lint: 60,
			Test: 120,
		},
	}
}

// NewDefaultUserConfig returns a UserConfig with default values.
// Note: Name is intentionally empty; it is populated from user.yaml.
func NewDefaultUserConfig() models.UserConfig {
	return models.UserConfig{}
}

// NewDefaultLanguageConfig returns a LanguageConfig with default values.
func NewDefaultLanguageConfig() models.LanguageConfig {
	return models.LanguageConfig{
		ConversationLanguage:     DefaultConversationLanguage,
		ConversationLanguageName: DefaultConversationLanguageName,
		AgentPromptLanguage:      DefaultAgentPromptLanguage,
		GitCommitMessages:        DefaultGitCommitMessages,
		CodeComments:             DefaultCodeComments,
		Documentation:            DefaultDocumentation,
		ErrorMessages:            DefaultErrorMessages,
	}
}

// NewDefaultQualityConfig returns a QualityConfig with default values.
func NewDefaultQualityConfig() models.QualityConfig {
	return models.QualityConfig{
		DevelopmentMode:    models.ModeTDD,
		EnforceQuality:     true,
		TestCoverageTarget: DefaultTestCoverageTarget,
		DDDSettings:        NewDefaultDDDSettings(),
		TDDSettings:        NewDefaultTDDSettings(),
		CoverageExemptions: NewDefaultCoverageExemptions(),
	}
}

// NewDefaultDDDSettings returns DDDSettings with default values.
func NewDefaultDDDSettings() models.DDDSettings {
	return models.DDDSettings{
		RequireExistingTests:  true,
		CharacterizationTests: true,
		BehaviorSnapshots:     true,
		MaxTransformationSize: DefaultMaxTransformationSize,
		PreserveBeforeImprove: true,
	}
}

// NewDefaultTDDSettings returns TDDSettings with default values.
func NewDefaultTDDSettings() models.TDDSettings {
	return models.TDDSettings{
		RedGreenRefactor:       true,
		TestFirstRequired:      true,
		MinCoveragePerCommit:   DefaultMinCoveragePerCommit,
		MutationTestingEnabled: false,
	}
}

// NewDefaultCoverageExemptions returns CoverageExemptions with default values.
func NewDefaultCoverageExemptions() models.CoverageExemptions {
	return models.CoverageExemptions{
		Enabled:              false,
		RequireJustification: true,
		MaxExemptPercentage:  DefaultMaxExemptPercentage,
	}
}

// NewDefaultProjectConfig returns a ProjectConfig with default values.
func NewDefaultProjectConfig() models.ProjectConfig {
	return models.ProjectConfig{}
}

// NewDefaultGitStrategyConfig returns a GitStrategyConfig with default values.
//
// Per SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-006: the three ModeProfile
// instances mirror the template-canonical defaults in
// internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
// (the schema SSOT). The deprecated FLAT fields retain their pre-existing
// default values for backward-compat (Option (c)).
func NewDefaultGitStrategyConfig() GitStrategyConfig {
	return GitStrategyConfig{
		// Top-level wire-through defaults.
		Mode:           "team",
		Provider:       "github",
		GitHubUsername: "",
		GitLab:         GitLabConfig{InstanceURL: ""},

		Manual: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "local",
			GitHubIntegration: false,
			PushToRemote:      false,
			AutoCheckpoint:    "disabled",
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},
		Personal: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: false},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: false},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},
		Team: ModeProfile{
			Workflow:          "github-flow",
			Environment:       "github",
			GitHubIntegration: true,
			PushToRemote:      true,
			BranchPrefix:      "feature/SPEC-",
			MainBranch:        "main",
			DraftPR:           true,
			RequiredReviews:   1,
			BranchProtection:  true,
			BranchCreation:    BranchCreationConfig{AutoEnabled: false, PromptAlways: true},
			Automation:        AutomationConfig{AutoBranch: false, AutoCommit: true, AutoPR: false, AutoPush: true},
			CommitStyle:       CommitStyleConfig{Format: "conventional", ScopeRequired: true},
			Hooks:             HooksConfig{PreCommit: "enforce", PrePush: "warn", CommitMsg: "warn"},
		},

		// Deprecated FLAT fields — preserve existing default values for backward-compat.
		AutoBranch:        false,
		BranchPrefix:      DefaultBranchPrefix,
		CommitStyle:       DefaultCommitStyle,
		WorktreeRoot:      "",
		GitLabInstanceURL: "",
	}
}

// NewDefaultSystemConfig returns a SystemConfig with default values.
func NewDefaultSystemConfig() SystemConfig {
	return SystemConfig{
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}
}

// @MX:ANCHOR: [AUTO] LLM configuration defaults factory. Single entry point for the full LLM config including model tiers, GLM settings, and performance policy.
// @MX:REASON: fan_in=6, referenced by many callers including config loader, CLI initialization, and test fixtures
// NewDefaultLLMConfig returns a LLMConfig with default values.
func NewDefaultLLMConfig() LLMConfig {
	return LLMConfig{
		GLMEnvVar:       DefaultGLMEnvVar,
		PerformanceTier: DefaultPerformanceTier,
		ClaudeModels: ClaudeTierModels{
			High:   "opus",
			Medium: "sonnet",
			Low:    "haiku",
		},
		DefaultModel: DefaultModel,
		QualityModel: DefaultQualModel,
		SpeedModel:   DefaultSpeedModel,
		GLM: GLMSettings{
			BaseURL: DefaultGLMBaseURL,
			Models: GLMModels{
				High:   DefaultGLMHigh,
				Medium: DefaultGLMMedium,
				Low:    DefaultGLMLow,
				// Legacy fields for backward compatibility
				Opus:   DefaultGLMOpus,
				Sonnet: DefaultGLMSonnet,
				Haiku:  DefaultGLMHaiku,
			},
		},
	}
}

// NewDefaultPricingConfig returns a PricingConfig with default values.
func NewDefaultPricingConfig() PricingConfig {
	return PricingConfig{
		TokenBudget: DefaultTokenBudget,
	}
}

// NewDefaultRalphConfig returns a RalphConfig with default values.
func NewDefaultRalphConfig() RalphConfig {
	return RalphConfig{
		MaxIterations:     DefaultMaxIterations,
		AutoConverge:      true,
		HumanReview:       true,
		LintAsInstruction: true,  // REQ-LAI-003: enabled by default
		WarnAsInstruction: false, // REQ-LAI-006: disabled by default
	}
}

// NewDefaultWorkflowConfig returns a WorkflowConfig with default values.
// The nested defaults mirror the template SSOT workflow.yaml exactly
// (internal/template/templates/.moai/config/sections/workflow.yaml).
func NewDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		AutoClear: AutoClearConfig{
			Enabled:        true,
			AfterPlan:      true,
			AfterRun:       false,
			TokenThreshold: 150000,
		},
		Completion: CompletionConfig{
			DetectInOutput: true,
			Markers: MarkersConfig{
				Complete: "<moai>COMPLETE</moai>",
				Done:     "<moai>DONE</moai>",
			},
		},
		DefaultMode:   "",
		ExecutionMode: "team",
		LoopPrevention: LoopPreventionConfig{
			FailurePatternDetection: true,
			MaxIterations:           100,
			MaxRetriesPerOperation:  3,
		},
		Memory: MemoryConfig{
			AuditEnabled:            true,
			IndexLineCap:            200,
			StaleAggregateThreshold: 10,
			StalenessThresholdHours: 24,
		},
		Team: TeamConfig{
			AutoSelection: TeamAutoSelectionConfig{
				MinDomainsForTeam:  3,
				MinFilesForTeam:    10,
				MinComplexityScore: 7,
			},
			Enabled:             true,
			MaxTeammates:        10,
			DefaultModel:        "sonnet",
			DelegateMode:        true,
			RequirePlanApproval: true,
			RoleProfileKeys:     []string{"implementer", "tester", "reviewer"},
			RoleProfiles: map[string]RoleProfileEntry{
				"researcher": {
					Mode:        "plan",
					Model:       "haiku",
					Isolation:   "none",
					Description: "Read-only codebase exploration and analysis",
				},
				"analyst": {
					Mode:        "plan",
					Model:       "sonnet",
					Isolation:   "none",
					Description: "Requirements analysis and validation",
				},
				"architect": {
					Mode:        "plan",
					Model:       "sonnet",
					Isolation:   "none",
					Description: "Solution design and architecture decisions",
				},
				"implementer": {
					Mode:        "acceptEdits",
					Model:       "sonnet",
					Isolation:   "worktree",
					Description: "Code implementation (backend, frontend, full-stack)",
				},
				"tester": {
					Mode:        "acceptEdits",
					Model:       "sonnet",
					Isolation:   "worktree",
					Description: "Test creation and coverage validation",
				},
				"designer": {
					Mode:        "acceptEdits",
					Model:       "sonnet",
					Isolation:   "worktree",
					Description: "UI/UX design with MCP design tools",
				},
				"reviewer": {
					Mode:        "plan",
					Model:       "haiku",
					Isolation:   "none",
					Description: "Code review and quality validation",
				},
			},
		},
		TokenBudget: TokenBudgetConfig{
			Plan: DefaultPlanTokens,
			Run:  DefaultRunTokens,
			Sync: DefaultSyncTokens,
		},
		Worktree: WorkflowWorktreeConfig{
			AutoCleanup:        true,
			AutoCreate:         false,
			AutoMerge:          true,
			SessionNamePattern: "moai-{ProjectName}-{SPEC-ID}",
			TmuxPreferred:      true,
		},
	}
}

// NewDefaultStateConfig returns a StateConfig with default values.
func NewDefaultStateConfig() StateConfig {
	return StateConfig{
		StateDir: DefaultStateDir,
	}
}

// NewDefaultSessionConfig returns a SessionConfig with default values.
// SPEC-V3R2-RT-004 REQ-022: default stale_seconds = 3600 (1 hour).
func NewDefaultSessionConfig() SessionConfig {
	return SessionConfig{
		StaleSeconds: DefaultStaleSeconds,
	}
}

// NewDefaultGitConventionConfig returns a GitConventionConfig with default values.
func NewDefaultGitConventionConfig() models.GitConventionConfig {
	return models.GitConventionConfig{
		Convention: DefaultGitConvention,
		AutoDetection: models.AutoDetectionConfig{
			Enabled:             true,
			SampleSize:          DefaultGitConventionSampleSize,
			ConfidenceThreshold: DefaultGitConventionConfidenceThreshold,
			Fallback:            DefaultGitConventionFallback,
		},
		Validation: models.ConventionValidationConfig{
			Enabled:         true,
			EnforceOnCommit: false,
			EnforceOnPush:   false,
			MaxLength:       DefaultGitConventionMaxLength,
		},
		Formatting: models.FormattingConfig{
			ShowExamples:    true,
			ShowSuggestions: true,
			Verbose:         false,
		},
	}
}

// NewDefaultSunsetConfig returns a SunsetConfig with default values.
func NewDefaultSunsetConfig() SunsetConfig {
	return SunsetConfig{
		Enabled:    false,
		Conditions: nil,
	}
}

// NewDefaultLSPQualityGates returns LSPQualityGates with default values.
func NewDefaultLSPQualityGates() LSPQualityGates {
	return LSPQualityGates{
		Enabled: true,
		Plan: PlanGate{
			RequireBaseline: true,
		},
		Run: RunGate{
			MaxErrors:       0,
			MaxTypeErrors:   0,
			MaxLintErrors:   0,
			AllowRegression: false,
		},
		Sync: SyncGate{
			MaxErrors:       0,
			MaxWarnings:     DefaultMaxWarnings,
			RequireCleanLSP: true,
		},
		CacheTTLSeconds: DefaultCacheTTLSeconds,
		TimeoutSeconds:  DefaultTimeoutSeconds,
	}
}

// defaultConstitutionConfig returns a ConstitutionConfig with defaults matching
// internal/template/templates/.moai/config/sections/constitution.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultConstitutionConfig() ConstitutionConfig {
	return ConstitutionConfig{
		ApprovedFrameworks: []string{"cobra", "viper"},
		ApprovedLanguages:  []string{"go"},
		Architecture: ConstitutionArchitecture{
			ForbiddenDependencies: []string{
				"circular imports",
				"direct template access from CLI handlers",
			},
			Patterns: []string{"clean-architecture", "repository-pattern"},
		},
		ForbiddenPatterns: []string{
			"global mutable state",
			"init() with side effects",
			"panic() in library code",
			"raw SQL without parameterized queries",
		},
		NamingConventions: ConstitutionNaming{
			Exported: "PascalCase",
			Files:    "snake_case.go",
			Packages: "lowercase, single word",
		},
		Performance: ConstitutionPerformance{},
		Security: ConstitutionSecurity{
			ForbiddenPractices: []string{
				"hardcoded credentials",
				"os.Exit in library code",
			},
			RequiredChecks: []string{"input-validation"},
		},
	}
}

// defaultContextConfig returns a ContextConfig with defaults matching
// internal/template/templates/.moai/config/sections/context.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultContextConfig() ContextConfig {
	return ContextConfig{
		AutoDetect: ContextAutoDetect{Enabled: true},
		Enabled:    true,
		MemoryIntegration: ContextMemoryIntegration{
			Enabled:            true,
			IncludeInContext:   true,
			PriorityOverSearch: true,
		},
		Performance: ContextPerformance{
			CacheTTLSeconds: 300,
			TimeoutSeconds:  10,
		},
		Search: ContextSearch{
			DateRangeDays:      30,
			MaxResults:         5,
			MaxTokensPerResult: 1000,
			ProjectScopeOnly:   true,
		},
		TokenBudget: ContextTokenBudget{
			MaxInjectionTokens: 5000,
			SkipIfUsageAbove:   150000,
		},
	}
}

// defaultInterviewConfig returns an InterviewConfig with defaults matching
// internal/template/templates/.moai/config/sections/interview.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
func defaultInterviewConfig() InterviewConfig {
	return InterviewConfig{
		ClarityThreshold: 4,
		Enabled:          true,
		Plan: InterviewMode{
			MaxRounds:         5,
			QuestionsPerRound: 3,
		},
		Project: InterviewMode{
			MaxRounds:         3,
			QuestionsPerRound: 3,
		},
		SkipConditions: []string{
			"resume_spec_id_present",
			"skip_interview_flag",
			"technical_keywords_gte_5",
		},
	}
}

// defaultDesignConfig returns a DesignConfig with defaults matching
// internal/template/templates/.moai/config/sections/design.yaml.
// REQ-MIG003-004: sensible defaults on absent file.
// Note: PassThreshold default 0.75 is above the FROZEN floor 0.60.
func defaultDesignConfig() DesignConfig {
	return DesignConfig{
		Adaptation: DesignAdaptation{
			ConfidenceThreshold: 0.7,
			Enabled:             true,
			IterationLimits: DesignIterationLimits{
				Builder:    3,
				Copywriter: 3,
				Designer:   2,
			},
			MinProjectsForAdaptation: 5,
		},
		BrandContext: DesignBrandContext{
			Dir:                 ".moai/project/brand",
			InterviewOnFirstRun: true,
		},
		ClaudeDesign: DesignClaudeDesign{
			Enabled:                 true,
			FallbackPath:            "code_based",
			SupportedBundleVersions: []string{"1.0"},
		},
		DefaultFramework: "next.js",
		DesignDocs: DesignDocs{
			AutoLoadOnDesignCommand: true,
			Dir:                     ".moai/design",
			Priority:                []string{"spec", "system", "research", "pencil-plan"},
			TokenBudget:             20000,
		},
		Enabled: true,
		Evaluator: DesignEvaluator{
			MemoryScope: "per_iteration",
		},
		Evolution: DesignEvolution{
			ArchiveAfterEvolve:  true,
			AutoEvolveThreshold: 3,
			CooldownHours:       24,
			GraduationCriteria: DesignGraduationCriteria{
				ConsistencyRatio:    0.8,
				MinimumConfidence:   0.8,
				MinimumObservations: 5,
				StalenessWindowDays: 30,
			},
			MaxActiveLearnings:      50,
			MaxEvolutionRatePerWeek: 3,
			RequireApproval:         true,
		},
		Figma: DesignFigma{Enabled: false},
		GanLoop: DesignGanLoop{
			EscalationAfter:      3,
			ImprovementThreshold: 0.05,
			MaxIterations:        5,
			PassThreshold:        0.75,
			SprintContract: DesignSprintContract{
				ArtifactDir:           ".moai/sprints",
				Enabled:               true,
				MaxNegotiationRounds:  2,
				OptionalHarnessLevels: []string{"standard"},
				RequiredHarnessLevels: []string{"thorough"},
			},
			StrictMode: false,
		},
		Version: "1.0.0",
	}
}

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
	DefaultGLMHigh   = "glm-5"
	DefaultGLMMedium = "glm-4.7"
	DefaultGLMLow    = "glm-4.7-flashx"
	// Legacy GLM model names (map to tiers)
	DefaultGLMHaiku  = "glm-4.7-flashx"
	DefaultGLMSonnet = "glm-4.7"
	DefaultGLMOpus   = "glm-5"
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
func NewDefaultGitStrategyConfig() GitStrategyConfig {
	return GitStrategyConfig{
		AutoBranch:   false,
		BranchPrefix: DefaultBranchPrefix,
		CommitStyle:  DefaultCommitStyle,
		Provider:     "github",
	}
}

// NewDefaultSystemConfig returns a SystemConfig with default values.
func NewDefaultSystemConfig() SystemConfig {
	return SystemConfig{
		LogLevel:  DefaultLogLevel,
		LogFormat: DefaultLogFormat,
	}
}

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
		MaxIterations: DefaultMaxIterations,
		AutoConverge:  true,
		HumanReview:   true,
	}
}

// NewDefaultWorkflowConfig returns a WorkflowConfig with default values.
func NewDefaultWorkflowConfig() WorkflowConfig {
	return WorkflowConfig{
		AutoClear:  true,
		PlanTokens: DefaultPlanTokens,
		RunTokens:  DefaultRunTokens,
		SyncTokens: DefaultSyncTokens,
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

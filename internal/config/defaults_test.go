package config

import (
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestNewDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if cfg == nil {
		t.Fatal("NewDefaultConfig() returned nil")
	}

	// Verify it returns a pointer
	cfg2 := NewDefaultConfig()
	if cfg == cfg2 {
		t.Error("NewDefaultConfig() returned the same pointer, expected distinct instances")
	}
}

func TestNewDefaultConfigContainsAllSections(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()

	// User section should have empty name (populated from file)
	if cfg.User.Name != "" {
		t.Errorf("User.Name: got %q, want empty", cfg.User.Name)
	}

	// Language section should have defaults
	if cfg.Language.ConversationLanguage != DefaultConversationLanguage {
		t.Errorf("Language.ConversationLanguage: got %q, want %q",
			cfg.Language.ConversationLanguage, DefaultConversationLanguage)
	}

	// Quality section should have defaults
	if cfg.Quality.DevelopmentMode != models.ModeTDD {
		t.Errorf("Quality.DevelopmentMode: got %q, want %q",
			cfg.Quality.DevelopmentMode, models.ModeTDD)
	}

	// GitStrategy should have defaults
	if cfg.GitStrategy.BranchPrefix != DefaultBranchPrefix {
		t.Errorf("GitStrategy.BranchPrefix: got %q, want %q",
			cfg.GitStrategy.BranchPrefix, DefaultBranchPrefix)
	}

	// System should have defaults
	if cfg.System.LogLevel != DefaultLogLevel {
		t.Errorf("System.LogLevel: got %q, want %q", cfg.System.LogLevel, DefaultLogLevel)
	}

	// LLM should have defaults
	if cfg.LLM.DefaultModel != DefaultModel {
		t.Errorf("LLM.DefaultModel: got %q, want %q", cfg.LLM.DefaultModel, DefaultModel)
	}

	// Pricing should have defaults
	if cfg.Pricing.TokenBudget != DefaultTokenBudget {
		t.Errorf("Pricing.TokenBudget: got %d, want %d",
			cfg.Pricing.TokenBudget, DefaultTokenBudget)
	}

	// Ralph should have defaults
	if cfg.Ralph.MaxIterations != DefaultMaxIterations {
		t.Errorf("Ralph.MaxIterations: got %d, want %d",
			cfg.Ralph.MaxIterations, DefaultMaxIterations)
	}

	// Workflow should have defaults
	if cfg.Workflow.PlanTokens != DefaultPlanTokens {
		t.Errorf("Workflow.PlanTokens: got %d, want %d",
			cfg.Workflow.PlanTokens, DefaultPlanTokens)
	}

	// GitConvention should have defaults
	if cfg.GitConvention.Convention != DefaultGitConvention {
		t.Errorf("GitConvention.Convention: got %q, want %q",
			cfg.GitConvention.Convention, DefaultGitConvention)
	}
}

func TestNewDefaultUserConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultUserConfig()
	if cfg.Name != "" {
		t.Errorf("Name: got %q, want empty (populated from file)", cfg.Name)
	}
}

func TestNewDefaultLanguageConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultLanguageConfig()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"ConversationLanguage", cfg.ConversationLanguage, DefaultConversationLanguage},
		{"ConversationLanguageName", cfg.ConversationLanguageName, DefaultConversationLanguageName},
		{"AgentPromptLanguage", cfg.AgentPromptLanguage, DefaultAgentPromptLanguage},
		{"GitCommitMessages", cfg.GitCommitMessages, DefaultGitCommitMessages},
		{"CodeComments", cfg.CodeComments, DefaultCodeComments},
		{"Documentation", cfg.Documentation, DefaultDocumentation},
		{"ErrorMessages", cfg.ErrorMessages, DefaultErrorMessages},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestNewDefaultQualityConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultQualityConfig()

	if cfg.DevelopmentMode != models.ModeTDD {
		t.Errorf("DevelopmentMode: got %q, want %q", cfg.DevelopmentMode, models.ModeTDD)
	}
	if !cfg.EnforceQuality {
		t.Error("EnforceQuality: expected true")
	}
	if cfg.TestCoverageTarget != DefaultTestCoverageTarget {
		t.Errorf("TestCoverageTarget: got %d, want %d",
			cfg.TestCoverageTarget, DefaultTestCoverageTarget)
	}

	// Verify nested settings are populated
	if !cfg.DDDSettings.RequireExistingTests {
		t.Error("DDDSettings.RequireExistingTests: expected true")
	}
	if !cfg.TDDSettings.RedGreenRefactor {
		t.Error("TDDSettings.RedGreenRefactor: expected true")
	}
	if cfg.CoverageExemptions.Enabled {
		t.Error("CoverageExemptions.Enabled: expected false")
	}
}

func TestNewDefaultDDDSettings(t *testing.T) {
	t.Parallel()

	s := NewDefaultDDDSettings()

	if !s.RequireExistingTests {
		t.Error("RequireExistingTests: expected true")
	}
	if !s.CharacterizationTests {
		t.Error("CharacterizationTests: expected true")
	}
	if !s.BehaviorSnapshots {
		t.Error("BehaviorSnapshots: expected true")
	}
	if s.MaxTransformationSize != DefaultMaxTransformationSize {
		t.Errorf("MaxTransformationSize: got %q, want %q",
			s.MaxTransformationSize, DefaultMaxTransformationSize)
	}
	if !s.PreserveBeforeImprove {
		t.Error("PreserveBeforeImprove: expected true")
	}
}

func TestNewDefaultTDDSettings(t *testing.T) {
	t.Parallel()

	s := NewDefaultTDDSettings()

	if !s.RedGreenRefactor {
		t.Error("RedGreenRefactor: expected true")
	}
	if !s.TestFirstRequired {
		t.Error("TestFirstRequired: expected true")
	}
	if s.MinCoveragePerCommit != DefaultMinCoveragePerCommit {
		t.Errorf("MinCoveragePerCommit: got %d, want %d",
			s.MinCoveragePerCommit, DefaultMinCoveragePerCommit)
	}
	if s.MutationTestingEnabled {
		t.Error("MutationTestingEnabled: expected false")
	}
}

func TestNewDefaultCoverageExemptions(t *testing.T) {
	t.Parallel()

	s := NewDefaultCoverageExemptions()

	if s.Enabled {
		t.Error("Enabled: expected false")
	}
	if !s.RequireJustification {
		t.Error("RequireJustification: expected true")
	}
	if s.MaxExemptPercentage != DefaultMaxExemptPercentage {
		t.Errorf("MaxExemptPercentage: got %d, want %d",
			s.MaxExemptPercentage, DefaultMaxExemptPercentage)
	}
}

func TestNewDefaultProjectConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultProjectConfig()
	// ProjectConfig defaults should be zero-valued
	if cfg.Name != "" {
		t.Errorf("Name: got %q, want empty", cfg.Name)
	}
}

func TestNewDefaultGitStrategyConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultGitStrategyConfig()

	if cfg.AutoBranch {
		t.Error("AutoBranch: expected false")
	}
	if cfg.BranchPrefix != DefaultBranchPrefix {
		t.Errorf("BranchPrefix: got %q, want %q", cfg.BranchPrefix, DefaultBranchPrefix)
	}
	if cfg.CommitStyle != DefaultCommitStyle {
		t.Errorf("CommitStyle: got %q, want %q", cfg.CommitStyle, DefaultCommitStyle)
	}
	if cfg.WorktreeRoot != "" {
		t.Errorf("WorktreeRoot: got %q, want empty", cfg.WorktreeRoot)
	}
}

func TestNewDefaultSystemConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultSystemConfig()

	if cfg.LogLevel != DefaultLogLevel {
		t.Errorf("LogLevel: got %q, want %q", cfg.LogLevel, DefaultLogLevel)
	}
	if cfg.LogFormat != DefaultLogFormat {
		t.Errorf("LogFormat: got %q, want %q", cfg.LogFormat, DefaultLogFormat)
	}
	if cfg.Version != "" {
		t.Errorf("Version: got %q, want empty", cfg.Version)
	}
	if cfg.NoColor {
		t.Error("NoColor: expected false")
	}
	if cfg.NonInteractive {
		t.Error("NonInteractive: expected false")
	}
}

func TestNewDefaultLLMConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultLLMConfig()

	if cfg.DefaultModel != DefaultModel {
		t.Errorf("DefaultModel: got %q, want %q", cfg.DefaultModel, DefaultModel)
	}
	if cfg.QualityModel != DefaultQualModel {
		t.Errorf("QualityModel: got %q, want %q", cfg.QualityModel, DefaultQualModel)
	}
	if cfg.SpeedModel != DefaultSpeedModel {
		t.Errorf("SpeedModel: got %q, want %q", cfg.SpeedModel, DefaultSpeedModel)
	}
}

func TestNewDefaultPricingConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultPricingConfig()

	if cfg.TokenBudget != DefaultTokenBudget {
		t.Errorf("TokenBudget: got %d, want %d", cfg.TokenBudget, DefaultTokenBudget)
	}
	if cfg.CostTracking {
		t.Error("CostTracking: expected false")
	}
}

func TestNewDefaultRalphConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultRalphConfig()

	if cfg.MaxIterations != DefaultMaxIterations {
		t.Errorf("MaxIterations: got %d, want %d", cfg.MaxIterations, DefaultMaxIterations)
	}
	if !cfg.AutoConverge {
		t.Error("AutoConverge: expected true")
	}
	if !cfg.HumanReview {
		t.Error("HumanReview: expected true")
	}
}

func TestNewDefaultWorkflowConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultWorkflowConfig()

	if !cfg.AutoClear {
		t.Error("AutoClear: expected true")
	}
	if cfg.PlanTokens != DefaultPlanTokens {
		t.Errorf("PlanTokens: got %d, want %d", cfg.PlanTokens, DefaultPlanTokens)
	}
	if cfg.RunTokens != DefaultRunTokens {
		t.Errorf("RunTokens: got %d, want %d", cfg.RunTokens, DefaultRunTokens)
	}
	if cfg.SyncTokens != DefaultSyncTokens {
		t.Errorf("SyncTokens: got %d, want %d", cfg.SyncTokens, DefaultSyncTokens)
	}
}

func TestNewDefaultLSPQualityGates(t *testing.T) {
	t.Parallel()

	gates := NewDefaultLSPQualityGates()

	if !gates.Enabled {
		t.Error("Enabled: expected true")
	}
	if !gates.Plan.RequireBaseline {
		t.Error("Plan.RequireBaseline: expected true")
	}
	if gates.Run.MaxErrors != 0 {
		t.Errorf("Run.MaxErrors: got %d, want 0", gates.Run.MaxErrors)
	}
	if gates.Run.MaxTypeErrors != 0 {
		t.Errorf("Run.MaxTypeErrors: got %d, want 0", gates.Run.MaxTypeErrors)
	}
	if gates.Run.MaxLintErrors != 0 {
		t.Errorf("Run.MaxLintErrors: got %d, want 0", gates.Run.MaxLintErrors)
	}
	if gates.Run.AllowRegression {
		t.Error("Run.AllowRegression: expected false")
	}
	if gates.Sync.MaxErrors != 0 {
		t.Errorf("Sync.MaxErrors: got %d, want 0", gates.Sync.MaxErrors)
	}
	if gates.Sync.MaxWarnings != DefaultMaxWarnings {
		t.Errorf("Sync.MaxWarnings: got %d, want %d", gates.Sync.MaxWarnings, DefaultMaxWarnings)
	}
	if !gates.Sync.RequireCleanLSP {
		t.Error("Sync.RequireCleanLSP: expected true")
	}
	if gates.CacheTTLSeconds != DefaultCacheTTLSeconds {
		t.Errorf("CacheTTLSeconds: got %d, want %d", gates.CacheTTLSeconds, DefaultCacheTTLSeconds)
	}
	if gates.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("TimeoutSeconds: got %d, want %d", gates.TimeoutSeconds, DefaultTimeoutSeconds)
	}
}

func TestDefaultConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"DefaultConversationLanguage", DefaultConversationLanguage, "en"},
		{"DefaultConversationLanguageName", DefaultConversationLanguageName, "English"},
		{"DefaultAgentPromptLanguage", DefaultAgentPromptLanguage, "en"},
		{"DefaultTestCoverageTarget", DefaultTestCoverageTarget, 85},
		{"DefaultMaxTransformationSize", DefaultMaxTransformationSize, "small"},
		{"DefaultLogLevel", DefaultLogLevel, "info"},
		{"DefaultLogFormat", DefaultLogFormat, "text"},
		{"DefaultModel", DefaultModel, "sonnet"},
		{"DefaultQualModel", DefaultQualModel, "opus"},
		{"DefaultSpeedModel", DefaultSpeedModel, "haiku"},
		{"DefaultTokenBudget", DefaultTokenBudget, 250000},
		{"DefaultMaxIterations", DefaultMaxIterations, 5},
		{"DefaultPlanTokens", DefaultPlanTokens, 30000},
		{"DefaultRunTokens", DefaultRunTokens, 180000},
		{"DefaultSyncTokens", DefaultSyncTokens, 40000},
		{"DefaultBranchPrefix", DefaultBranchPrefix, "moai/"},
		{"DefaultCommitStyle", DefaultCommitStyle, "conventional"},
		{"DefaultCacheTTLSeconds", DefaultCacheTTLSeconds, 5},
		{"DefaultTimeoutSeconds", DefaultTimeoutSeconds, 3},
		{"DefaultMaxWarnings", DefaultMaxWarnings, 10},
		{"DefaultGitConvention", DefaultGitConvention, "auto"},
		{"DefaultGitConventionSampleSize", DefaultGitConventionSampleSize, 100},
		{"DefaultGitConventionConfidenceThreshold", DefaultGitConventionConfidenceThreshold, 0.5},
		{"DefaultGitConventionFallback", DefaultGitConventionFallback, "conventional-commits"},
		{"DefaultGitConventionMaxLength", DefaultGitConventionMaxLength, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("%s: got %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestDefaultGitConventionConfidenceThreshold(t *testing.T) {
	t.Parallel()

	// Float comparison needs separate test (can't use any equality).
	if DefaultGitConventionConfidenceThreshold != 0.5 {
		t.Errorf("DefaultGitConventionConfidenceThreshold: got %f, want 0.5",
			DefaultGitConventionConfidenceThreshold)
	}
}

func TestNewDefaultGitConventionConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultGitConventionConfig()

	if cfg.Convention != DefaultGitConvention {
		t.Errorf("Convention: got %q, want %q", cfg.Convention, DefaultGitConvention)
	}

	// AutoDetection
	if !cfg.AutoDetection.Enabled {
		t.Error("AutoDetection.Enabled: expected true")
	}
	if cfg.AutoDetection.SampleSize != DefaultGitConventionSampleSize {
		t.Errorf("AutoDetection.SampleSize: got %d, want %d",
			cfg.AutoDetection.SampleSize, DefaultGitConventionSampleSize)
	}
	if cfg.AutoDetection.ConfidenceThreshold != DefaultGitConventionConfidenceThreshold {
		t.Errorf("AutoDetection.ConfidenceThreshold: got %f, want %f",
			cfg.AutoDetection.ConfidenceThreshold, DefaultGitConventionConfidenceThreshold)
	}
	if cfg.AutoDetection.Fallback != DefaultGitConventionFallback {
		t.Errorf("AutoDetection.Fallback: got %q, want %q",
			cfg.AutoDetection.Fallback, DefaultGitConventionFallback)
	}

	// Validation
	if !cfg.Validation.Enabled {
		t.Error("Validation.Enabled: expected true")
	}
	if cfg.Validation.EnforceOnCommit {
		t.Error("Validation.EnforceOnCommit: expected false")
	}
	if cfg.Validation.EnforceOnPush {
		t.Error("Validation.EnforceOnPush: expected false")
	}
	if cfg.Validation.MaxLength != DefaultGitConventionMaxLength {
		t.Errorf("Validation.MaxLength: got %d, want %d",
			cfg.Validation.MaxLength, DefaultGitConventionMaxLength)
	}

	// Formatting
	if !cfg.Formatting.ShowExamples {
		t.Error("Formatting.ShowExamples: expected true")
	}
	if !cfg.Formatting.ShowSuggestions {
		t.Error("Formatting.ShowSuggestions: expected true")
	}
	if cfg.Formatting.Verbose {
		t.Error("Formatting.Verbose: expected false")
	}
}

func TestNewDefaultConfigContainsGitConvention(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	if cfg.GitConvention.Convention != DefaultGitConvention {
		t.Errorf("GitConvention.Convention: got %q, want %q",
			cfg.GitConvention.Convention, DefaultGitConvention)
	}
}

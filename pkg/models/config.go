// @MX:NOTE: [AUTO] 개발 방법론 모드를 정의합니다. ModeDDD(기존 코드베이스)와 ModeTDD(신규 개발)를 지원합니다.
package models

// DevelopmentMode defines the development methodology mode.
type DevelopmentMode string

const (
	// ModeDDD uses Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE).
	// Best for: existing codebases with minimal test coverage.
	ModeDDD DevelopmentMode = "ddd"

	// ModeTDD uses Test-Driven Development (RED-GREEN-REFACTOR).
	// Best for: new projects, feature development, codebases with existing tests (default).
	ModeTDD DevelopmentMode = "tdd"
)

// ValidDevelopmentModes returns all valid development mode values.
func ValidDevelopmentModes() []DevelopmentMode {
	return []DevelopmentMode{ModeDDD, ModeTDD}
}

// IsValid checks if the development mode is a valid value.
func (m DevelopmentMode) IsValid() bool {
	switch m {
	case ModeDDD, ModeTDD:
		return true
	}
	return false
}

// UserConfig represents the user configuration section.
type UserConfig struct {
	Name string `yaml:"name"`
}

// LanguageConfig represents the language configuration section.
type LanguageConfig struct {
	ConversationLanguage     string `yaml:"conversation_language"`
	ConversationLanguageName string `yaml:"conversation_language_name"`
	AgentPromptLanguage      string `yaml:"agent_prompt_language"`
	GitCommitMessages        string `yaml:"git_commit_messages"`
	CodeComments             string `yaml:"code_comments"`
	Documentation            string `yaml:"documentation"`
	ErrorMessages            string `yaml:"error_messages"`
}

// @MX:NOTE: [AUTO] TRUST 5 품질 게이트 설정을 포함합니다. LSP 통합, 테스트 커버리지, 개발 방법론(DDD/TDD) 설정을 관리합니다.
// QualityConfig represents the quality configuration section.
type QualityConfig struct {
	DevelopmentMode    DevelopmentMode    `yaml:"development_mode"`
	EnforceQuality     bool               `yaml:"enforce_quality"`
	TestCoverageTarget int                `yaml:"test_coverage_target"`
	DDDSettings        DDDSettings        `yaml:"ddd_settings"`
	TDDSettings        TDDSettings        `yaml:"tdd_settings"`
	CoverageExemptions CoverageExemptions `yaml:"coverage_exemptions"`
	TestQuality        TestQuality        `yaml:"test_quality"`
	LSPQualityGates    LSPQualityGates    `yaml:"lsp_quality_gates"`
	Principles         Principles         `yaml:"principles"`
	LSPIntegration     LSPIntegration     `yaml:"lsp_integration"`
}

// TestQuality configures test quality requirements.
type TestQuality struct {
	SpecificationBased          bool `yaml:"specification_based"`
	MeaningfulAssertions        bool `yaml:"meaningful_assertions"`
	AvoidImplementationCoupling bool `yaml:"avoid_implementation_coupling"`
	MutationTestingEnabled      bool `yaml:"mutation_testing_enabled"`
}

// LSPQualityGates configures LSP-based quality validation.
type LSPQualityGates struct {
	Enabled         bool        `yaml:"enabled"`
	Plan            LSPPlanGate `yaml:"plan"`
	Run             LSPRunGate  `yaml:"run"`
	Sync            LSPSyncGate `yaml:"sync"`
	CacheTTLSeconds int         `yaml:"cache_ttl_seconds"`
	TimeoutSeconds  int         `yaml:"timeout_seconds"`
}

// LSPPlanGate configures Plan phase LSP requirements.
type LSPPlanGate struct {
	RequireBaseline bool `yaml:"require_baseline"`
}

// LSPRunGate configures Run phase LSP requirements.
type LSPRunGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxTypeErrors   int  `yaml:"max_type_errors"`
	MaxLintErrors   int  `yaml:"max_lint_errors"`
	AllowRegression bool `yaml:"allow_regression"`
}

// LSPSyncGate configures Sync phase LSP requirements.
type LSPSyncGate struct {
	MaxErrors       int  `yaml:"max_errors"`
	MaxWarnings     int  `yaml:"max_warnings"`
	RequireCleanLSP bool `yaml:"require_clean_lsp"`
}

// Principles configures quality principles.
type Principles struct {
	Simplicity SimplicityPrinciple `yaml:"simplicity"`
}

// SimplicityPrinciple configures simplicity constraints.
type SimplicityPrinciple struct {
	MaxParallelTasks int `yaml:"max_parallel_tasks"`
}

// LSPIntegration configures LSP integration with TRUST 5 framework.
type LSPIntegration struct {
	TRUST5Integration   TRUST5Integration   `yaml:"trust5_integration"`
	DiagnosticSources   []string            `yaml:"diagnostic_sources"`
	RegressionDetection RegressionDetection `yaml:"regression_detection"`
}

// TRUST5Integration maps TRUST 5 dimensions to LSP checks.
type TRUST5Integration struct {
	Tested         []string `yaml:"tested"`
	Readable       []string `yaml:"readable"`
	Understandable []string `yaml:"understandable"`
	Secured        []string `yaml:"secured"`
	Trackable      []string `yaml:"trackable"`
}

// RegressionDetection configures regression detection thresholds.
type RegressionDetection struct {
	ErrorIncreaseThreshold     int `yaml:"error_increase_threshold"`
	WarningIncreaseThreshold   int `yaml:"warning_increase_threshold"`
	TypeErrorIncreaseThreshold int `yaml:"type_error_increase_threshold"`
}

// DDDSettings configures Domain-Driven Development mode (ANALYZE-PRESERVE-IMPROVE).
type DDDSettings struct {
	RequireExistingTests  bool   `yaml:"require_existing_tests"`
	CharacterizationTests bool   `yaml:"characterization_tests"`
	BehaviorSnapshots     bool   `yaml:"behavior_snapshots"`
	MaxTransformationSize string `yaml:"max_transformation_size"`
	PreserveBeforeImprove bool   `yaml:"preserve_before_improve"`
}

// TDDSettings configures Test-Driven Development mode (RED-GREEN-REFACTOR).
type TDDSettings struct {
	RedGreenRefactor       bool `yaml:"red_green_refactor"`
	TestFirstRequired      bool `yaml:"test_first_required"`
	MinCoveragePerCommit   int  `yaml:"min_coverage_per_commit"`
	MutationTestingEnabled bool `yaml:"mutation_testing_enabled"`
}

// CoverageExemptions allows gradual coverage improvement for legacy code.
type CoverageExemptions struct {
	Enabled              bool `yaml:"enabled"`
	RequireJustification bool `yaml:"require_justification"`
	MaxExemptPercentage  int  `yaml:"max_exempt_percentage"`
}

// ReportGeneration configures quality report generation.
type ReportGeneration struct {
	Enabled    bool   `yaml:"enabled"`
	AutoCreate bool   `yaml:"auto_create"`
	WarnUser   bool   `yaml:"warn_user"`
	UserChoice string `yaml:"user_choice"`
}

// LSPStateTracking configures LSP state tracking across phases.
type LSPStateTracking struct {
	Enabled       bool               `yaml:"enabled"`
	CapturePoints []string           `yaml:"capture_points"`
	Comparison    LSPStateComparison `yaml:"comparison"`
	Logging       LSPStateLogging    `yaml:"logging"`
}

// LSPStateComparison configures LSP state comparison settings.
type LSPStateComparison struct {
	Baseline            string `yaml:"baseline"`
	RegressionThreshold int    `yaml:"regression_threshold"`
}

// LSPStateLogging configures LSP state logging settings.
type LSPStateLogging struct {
	LogLSPStateChanges     bool `yaml:"log_lsp_state_changes"`
	LogRegressionDetection bool `yaml:"log_regression_detection"`
	LogCompletionMarkers   bool `yaml:"log_completion_markers"`
	IncludeLSPInReports    bool `yaml:"include_lsp_in_reports"`
}

// FullQualityConfig represents the complete quality.yaml structure.
type FullQualityConfig struct {
	Constitution     QualityConfig    `yaml:"constitution"`
	ReportGeneration ReportGeneration `yaml:"report_generation"`
	LSPStateTracking LSPStateTracking `yaml:"lsp_state_tracking"`
}

// GitConventionConfig represents commit message convention settings.
type GitConventionConfig struct {
	// Convention name: auto, conventional-commits, angular, karma, custom
	Convention string `yaml:"convention"`

	// AutoDetection settings for convention discovery
	AutoDetection AutoDetectionConfig `yaml:"auto_detection"`

	// Validation settings for commit message checking
	Validation ConventionValidationConfig `yaml:"validation"`

	// Formatting settings for error display
	Formatting FormattingConfig `yaml:"formatting"`

	// Custom holds a user-defined convention configuration.
	Custom CustomConventionConfig `yaml:"custom"`
}

// AutoDetectionConfig configures convention auto-detection behavior.
type AutoDetectionConfig struct {
	Enabled             bool    `yaml:"enabled"`
	SampleSize          int     `yaml:"sample_size"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	Fallback            string  `yaml:"fallback"`
}

// ConventionValidationConfig configures when and how validation is enforced.
type ConventionValidationConfig struct {
	Enabled         bool `yaml:"enabled"`
	EnforceOnCommit bool `yaml:"enforce_on_commit"`
	EnforceOnPush   bool `yaml:"enforce_on_push"`
	MaxLength       int  `yaml:"max_length"`
}

// FormattingConfig configures error message formatting.
type FormattingConfig struct {
	ShowExamples    bool `yaml:"show_examples"`
	ShowSuggestions bool `yaml:"show_suggestions"`
	Verbose         bool `yaml:"verbose"`
}

// CustomConventionConfig holds a user-defined commit convention definition.
type CustomConventionConfig struct {
	Name      string   `yaml:"name"`
	Pattern   string   `yaml:"pattern"`
	Types     []string `yaml:"types"`
	Scopes    []string `yaml:"scopes"`
	MaxLength int      `yaml:"max_length"`
	Examples  []string `yaml:"examples"`
}

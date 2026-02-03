package models

// DevelopmentMode defines the development methodology mode.
type DevelopmentMode string

const (
	// ModeDDD uses Domain-Driven Development (ANALYZE-PRESERVE-IMPROVE).
	// Best for: legacy refactoring only (existing codebase, no new features).
	ModeDDD DevelopmentMode = "ddd"

	// ModeTDD uses Test-Driven Development (RED-GREEN-REFACTOR).
	// Best for: isolated new modules with no existing code dependencies (rare).
	ModeTDD DevelopmentMode = "tdd"

	// ModeHybrid uses TDD for new code and DDD for legacy code.
	// Best for: all development work (new projects, new features, ongoing development).
	ModeHybrid DevelopmentMode = "hybrid"
)

// ValidDevelopmentModes returns all valid development mode values.
func ValidDevelopmentModes() []DevelopmentMode {
	return []DevelopmentMode{ModeDDD, ModeTDD, ModeHybrid}
}

// IsValid checks if the development mode is a valid value.
func (m DevelopmentMode) IsValid() bool {
	switch m {
	case ModeDDD, ModeTDD, ModeHybrid:
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

// QualityConfig represents the quality configuration section.
type QualityConfig struct {
	DevelopmentMode    DevelopmentMode    `yaml:"development_mode"`
	EnforceQuality     bool               `yaml:"enforce_quality"`
	TestCoverageTarget int                `yaml:"test_coverage_target"`
	DDDSettings        DDDSettings        `yaml:"ddd_settings"`
	TDDSettings        TDDSettings        `yaml:"tdd_settings"`
	HybridSettings     HybridSettings     `yaml:"hybrid_settings"`
	CoverageExemptions CoverageExemptions `yaml:"coverage_exemptions"`
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

// HybridSettings configures Hybrid mode (TDD for new code, DDD for legacy).
type HybridSettings struct {
	NewFeatures         string `yaml:"new_features"`
	LegacyRefactoring   string `yaml:"legacy_refactoring"`
	MinCoverageNew      int    `yaml:"min_coverage_new"`
	MinCoverageLegacy   int    `yaml:"min_coverage_legacy"`
	PreserveRefactoring bool   `yaml:"preserve_refactoring"`
}

// CoverageExemptions allows gradual coverage improvement for legacy code.
type CoverageExemptions struct {
	Enabled              bool `yaml:"enabled"`
	RequireJustification bool `yaml:"require_justification"`
	MaxExemptPercentage  int  `yaml:"max_exempt_percentage"`
}

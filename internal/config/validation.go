package config

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/modu-ai/moai-adk/pkg/models"
)

// Dynamic token patterns that must not appear in configuration values.
// These indicate unexpanded template variables (ADR-011 compliance).
var dynamicTokenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\$\{[^}]+\}`),        // ${VAR}
	regexp.MustCompile(`\{\{[^}]+\}\}`),      // {{VAR}}
	regexp.MustCompile(`\$[A-Z_][A-Z0-9_]*`), // $VAR
}

// Validate checks the configuration for correctness.
// The loadedSections map indicates which sections were loaded from YAML files
// (as opposed to using defaults). Required field validation only applies
// to sections that were explicitly loaded.
func Validate(cfg *Config, loadedSections map[string]bool) error {
	var errs []ValidationError

	// Check required fields for loaded sections
	errs = append(errs, validateRequired(cfg, loadedSections)...)

	// Check development mode
	errs = append(errs, validateDevelopmentMode(cfg.Quality.DevelopmentMode)...)

	// Check quality config ranges
	errs = append(errs, validateQualityConfig(&cfg.Quality)...)

	// Check for unexpanded dynamic tokens
	errs = append(errs, validateDynamicTokens(cfg)...)

	if len(errs) > 0 {
		return &ValidationErrors{Errors: errs}
	}
	return nil
}

// validateRequired checks that required fields are populated for loaded sections.
func validateRequired(cfg *Config, loadedSections map[string]bool) []ValidationError {
	var errs []ValidationError

	if loadedSections["user"] && cfg.User.Name == "" {
		errs = append(errs, ValidationError{
			Field:   "user.name",
			Message: "required field is empty; set the user name in .moai/config/sections/user.yaml (example: name: YourName)",
			Wrapped: ErrInvalidConfig,
		})
	}

	return errs
}

// validateDevelopmentMode checks that the development mode is a valid value.
func validateDevelopmentMode(mode models.DevelopmentMode) []ValidationError {
	if mode == "" {
		return nil // empty is acceptable, defaults will be applied
	}
	if !mode.IsValid() {
		validModes := developmentModeStrings()
		return []ValidationError{
			{
				Field:   "quality.development_mode",
				Message: fmt.Sprintf("must be one of: %s", strings.Join(validModes, ", ")),
				Value:   string(mode),
				Wrapped: ErrInvalidDevelopmentMode,
			},
		}
	}
	return nil
}

// validateQualityConfig validates quality-specific configuration value ranges.
func validateQualityConfig(q *models.QualityConfig) []ValidationError {
	var errs []ValidationError

	if q.TestCoverageTarget < 0 || q.TestCoverageTarget > 100 {
		errs = append(errs, ValidationError{
			Field:   "quality.test_coverage_target",
			Message: "must be between 0 and 100",
			Value:   q.TestCoverageTarget,
			Wrapped: ErrInvalidConfig,
		})
	}

	if q.TDDSettings.MinCoveragePerCommit < 0 || q.TDDSettings.MinCoveragePerCommit > 100 {
		errs = append(errs, ValidationError{
			Field:   "quality.tdd_settings.min_coverage_per_commit",
			Message: "must be between 0 and 100",
			Value:   q.TDDSettings.MinCoveragePerCommit,
			Wrapped: ErrInvalidConfig,
		})
	}

	if q.HybridSettings.MinCoverageNew < 0 || q.HybridSettings.MinCoverageNew > 100 {
		errs = append(errs, ValidationError{
			Field:   "quality.hybrid_settings.min_coverage_new",
			Message: "must be between 0 and 100",
			Value:   q.HybridSettings.MinCoverageNew,
			Wrapped: ErrInvalidConfig,
		})
	}

	if q.HybridSettings.MinCoverageLegacy < 0 || q.HybridSettings.MinCoverageLegacy > 100 {
		errs = append(errs, ValidationError{
			Field:   "quality.hybrid_settings.min_coverage_legacy",
			Message: "must be between 0 and 100",
			Value:   q.HybridSettings.MinCoverageLegacy,
			Wrapped: ErrInvalidConfig,
		})
	}

	if q.CoverageExemptions.MaxExemptPercentage < 0 || q.CoverageExemptions.MaxExemptPercentage > 100 {
		errs = append(errs, ValidationError{
			Field:   "quality.coverage_exemptions.max_exempt_percentage",
			Message: "must be between 0 and 100",
			Value:   q.CoverageExemptions.MaxExemptPercentage,
			Wrapped: ErrInvalidConfig,
		})
	}

	return errs
}

// validateDynamicTokens checks all string fields for unexpanded dynamic tokens.
func validateDynamicTokens(cfg *Config) []ValidationError {
	var errs []ValidationError

	// User section
	errs = append(errs, checkStringField("user.name", cfg.User.Name)...)

	// Language section
	errs = append(errs, checkStringField("language.conversation_language", cfg.Language.ConversationLanguage)...)
	errs = append(errs, checkStringField("language.conversation_language_name", cfg.Language.ConversationLanguageName)...)
	errs = append(errs, checkStringField("language.agent_prompt_language", cfg.Language.AgentPromptLanguage)...)
	errs = append(errs, checkStringField("language.git_commit_messages", cfg.Language.GitCommitMessages)...)
	errs = append(errs, checkStringField("language.code_comments", cfg.Language.CodeComments)...)
	errs = append(errs, checkStringField("language.documentation", cfg.Language.Documentation)...)
	errs = append(errs, checkStringField("language.error_messages", cfg.Language.ErrorMessages)...)

	// Quality section
	errs = append(errs, checkStringField("quality.development_mode", string(cfg.Quality.DevelopmentMode))...)
	errs = append(errs, checkStringField("quality.ddd_settings.max_transformation_size", cfg.Quality.DDDSettings.MaxTransformationSize)...)
	errs = append(errs, checkStringField("quality.hybrid_settings.new_features", cfg.Quality.HybridSettings.NewFeatures)...)
	errs = append(errs, checkStringField("quality.hybrid_settings.legacy_refactoring", cfg.Quality.HybridSettings.LegacyRefactoring)...)

	// System section
	errs = append(errs, checkStringField("system.version", cfg.System.Version)...)
	errs = append(errs, checkStringField("system.log_level", cfg.System.LogLevel)...)
	errs = append(errs, checkStringField("system.log_format", cfg.System.LogFormat)...)

	// Git strategy section
	errs = append(errs, checkStringField("git_strategy.branch_prefix", cfg.GitStrategy.BranchPrefix)...)
	errs = append(errs, checkStringField("git_strategy.commit_style", cfg.GitStrategy.CommitStyle)...)

	// LLM section
	errs = append(errs, checkStringField("llm.default_model", cfg.LLM.DefaultModel)...)
	errs = append(errs, checkStringField("llm.quality_model", cfg.LLM.QualityModel)...)
	errs = append(errs, checkStringField("llm.speed_model", cfg.LLM.SpeedModel)...)

	return errs
}

// checkStringField checks a single string field for dynamic token patterns.
func checkStringField(field, value string) []ValidationError {
	if value == "" {
		return nil
	}
	for _, pattern := range dynamicTokenPatterns {
		if match := pattern.FindString(value); match != "" {
			return []ValidationError{
				{
					Field:   field,
					Message: fmt.Sprintf("contains unexpanded dynamic token: %s", match),
					Value:   value,
					Wrapped: ErrDynamicToken,
				},
			}
		}
	}
	return nil
}

// developmentModeStrings returns valid development mode values as strings.
func developmentModeStrings() []string {
	modes := models.ValidDevelopmentModes()
	strs := make([]string, len(modes))
	for i, m := range modes {
		strs[i] = string(m)
	}
	return strs
}

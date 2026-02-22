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

// @MX:ANCHOR: [AUTO] 모든 CLI 명령의 설정 유효성을 검사하는 핵심 진입점입니다.
// @MX:REASON: [AUTO] fan_in=37+, 모든 CLI 명령이 이 유효성 검사에 의존합니다
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

	// Check git convention config
	errs = append(errs, validateGitConventionConfig(&cfg.GitConvention)...)

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

// validGitConventionNames lists recognized convention names.
var validGitConventionNames = map[string]bool{
	"auto":                 true,
	"conventional-commits": true,
	"angular":              true,
	"karma":                true,
	"custom":               true,
}

// validateGitConventionConfig checks the git convention configuration.
func validateGitConventionConfig(gc *models.GitConventionConfig) []ValidationError {
	var errs []ValidationError

	if gc.Convention != "" && !validGitConventionNames[gc.Convention] {
		errs = append(errs, ValidationError{
			Field:   "git_convention.convention",
			Message: "must be one of: auto, conventional-commits, angular, karma, custom",
			Value:   gc.Convention,
			Wrapped: ErrInvalidConfig,
		})
	}

	if gc.AutoDetection.SampleSize < 0 {
		errs = append(errs, ValidationError{
			Field:   "git_convention.auto_detection.sample_size",
			Message: "must be non-negative",
			Value:   gc.AutoDetection.SampleSize,
			Wrapped: ErrInvalidConfig,
		})
	}

	if gc.AutoDetection.ConfidenceThreshold < 0 || gc.AutoDetection.ConfidenceThreshold > 1 {
		errs = append(errs, ValidationError{
			Field:   "git_convention.auto_detection.confidence_threshold",
			Message: "must be between 0.0 and 1.0",
			Value:   gc.AutoDetection.ConfidenceThreshold,
			Wrapped: ErrInvalidConfig,
		})
	}

	if gc.Validation.MaxLength < 0 {
		errs = append(errs, ValidationError{
			Field:   "git_convention.validation.max_length",
			Message: "must be non-negative",
			Value:   gc.Validation.MaxLength,
			Wrapped: ErrInvalidConfig,
		})
	}

	// When convention is "custom", pattern is required.
	if gc.Convention == "custom" && gc.Custom.Pattern == "" {
		errs = append(errs, ValidationError{
			Field:   "git_convention.custom.pattern",
			Message: "pattern is required when convention is 'custom'",
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

	// System section
	errs = append(errs, checkStringField("system.version", cfg.System.Version)...)
	errs = append(errs, checkStringField("system.log_level", cfg.System.LogLevel)...)
	errs = append(errs, checkStringField("system.log_format", cfg.System.LogFormat)...)

	// Git strategy section
	errs = append(errs, checkStringField("git_strategy.branch_prefix", cfg.GitStrategy.BranchPrefix)...)
	errs = append(errs, checkStringField("git_strategy.commit_style", cfg.GitStrategy.CommitStyle)...)

	// Git convention section
	errs = append(errs, checkStringField("git_convention.convention", cfg.GitConvention.Convention)...)
	errs = append(errs, checkStringField("git_convention.auto_detection.fallback", cfg.GitConvention.AutoDetection.Fallback)...)
	errs = append(errs, checkStringField("git_convention.custom.name", cfg.GitConvention.Custom.Name)...)
	errs = append(errs, checkStringField("git_convention.custom.pattern", cfg.GitConvention.Custom.Pattern)...)

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

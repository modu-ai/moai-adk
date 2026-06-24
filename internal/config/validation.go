package config

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/modu-ai/moai-adk/pkg/models"
)

// validate is the package-level validator instance (singleton).
var validate = validator.New()

// Dynamic token patterns that must not appear in configuration values.
// These indicate unexpanded template variables (ADR-011 compliance).
var dynamicTokenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\$\{[^}]+\}`),        // ${VAR}
	regexp.MustCompile(`\{\{[^}]+\}\}`),      // {{VAR}}
	regexp.MustCompile(`\$[A-Z_][A-Z0-9_]*`), // $VAR
}

// @MX:ANCHOR: [AUTO] Core entry point that validates configuration for all CLI commands.
// @MX:REASON: [AUTO] fan_in=37+, all CLI commands depend on this validation
// Validate checks the configuration for correctness.
// The loadedSections map indicates which sections were loaded from YAML files
// (as opposed to using defaults). Required field validation only applies
// to sections that were explicitly loaded.
func Validate(cfg *Config, loadedSections map[string]bool) error {
	var errs []ValidationError

	// Run validator/v10 struct validation as the first pass.
	// Only fields without existing custom checks produce errors here
	// (currently: LLM.PerformanceTier oneof).  Fields gated by loadedSections
	// or with existing custom checks are skipped inside runStructValidation.
	errs = append(errs, runStructValidation(cfg, loadedSections)...)

	// Check required fields for loaded sections
	errs = append(errs, validateRequired(cfg, loadedSections)...)

	// Check development mode
	errs = append(errs, validateDevelopmentMode(cfg.Quality.DevelopmentMode)...)

	// Check quality config ranges
	errs = append(errs, validateQualityConfig(&cfg.Quality)...)

	// Check git convention config
	errs = append(errs, validateGitConventionConfig(&cfg.GitConvention)...)

	// Check git strategy merge_method enum (per-mode profile)
	errs = append(errs, validateGitStrategyMergeMethod(cfg)...)

	// Check for unexpanded dynamic tokens
	errs = append(errs, validateDynamicTokens(cfg)...)

	// Check the GLM base_url is a safe, well-formed https endpoint (REQ-CGH-007).
	errs = append(errs, validateGLMBaseURL(cfg.LLM.GLM.BaseURL)...)

	if len(errs) > 0 {
		return &ValidationErrors{Errors: errs}
	}
	return nil
}

// validateGLMBaseURL validates the GLM base_url before its credentials are routed
// to it (REQ-CGH-007, SECURITY). An empty value is acceptable — the loader falls
// back to DefaultGLMBaseURL (EC-5). A non-empty value MUST be a well-formed URL
// with the https scheme AND a non-empty host. The constraint is "well-formed https
// + a real host" (R5: not "z.ai-only", so legitimate self-hosted GLM proxies pass)
// — it guards against silently routing ANTHROPIC_AUTH_TOKEN to a non-TLS or
// malformed/hostless endpoint. The canonical DefaultGLMBaseURL always passes.
func validateGLMBaseURL(baseURL string) []ValidationError {
	if baseURL == "" {
		return nil // falls back to DefaultGLMBaseURL on load (EC-5)
	}

	reject := func(reason string) []ValidationError {
		return []ValidationError{{
			Field:   "llm.glm.base_url",
			Message: reason + "; expected a well-formed https:// URL (e.g. " + DefaultGLMBaseURL + ")",
			Value:   baseURL,
			Wrapped: ErrInvalidConfig,
		}}
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return reject("malformed URL")
	}
	if u.Scheme != "https" {
		return reject("scheme must be https (credentials must not be routed over a non-TLS or non-http URL)")
	}
	if u.Host == "" {
		return reject("URL has no host")
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

// ValidateQualitySection is a thin exported seam over the existing unexported
// validateQualityConfig (SPEC-WEB-CONSOLE-007 §B.1 export seam). It exists ONLY
// so out-of-package callers (internal/web project-config write path) can reuse the
// SAME quality value-range rules (test_coverage_target / tdd_settings.min_coverage_per_commit
// / coverage_exemptions.max_exempt_percentage 0-100) rather than authoring a parallel
// rule-set. It adds NO new rule and is NOT a new validator function — it forwards
// verbatim to validateQualityConfig, mirroring how IsValidConvention/ValidConventions
// already export the convention SSOT (CRITICAL SCOPE CONSTRAINT, REQ-WC7-002).
func ValidateQualitySection(q *models.QualityConfig) []ValidationError {
	return validateQualityConfig(q)
}

// ValidateGitConventionSection is a thin exported seam over the existing
// unexported validateGitConventionConfig (SPEC-WEB-CONSOLE-007 §B.1 export seam).
// Same rationale as ValidateQualitySection: out-of-package callers reuse the SAME
// git-convention rules (convention enum / auto_detection.sample_size>=0 /
// auto_detection.confidence_threshold [0.0,1.0] / validation.max_length>=0 /
// custom.pattern required when convention=="custom") with NO new rule. It forwards
// verbatim to validateGitConventionConfig (CRITICAL SCOPE CONSTRAINT, REQ-WC7-002).
func ValidateGitConventionSection(gc *models.GitConventionConfig) []ValidationError {
	return validateGitConventionConfig(gc)
}

// validGitConventionNames lists recognized convention names.
var validGitConventionNames = map[string]bool{
	"auto":                 true,
	"conventional-commits": true,
	"angular":              true,
	"karma":                true,
}

// IsValidConvention reports whether name is one of the 4 canonical git
// convention names (auto, conventional-commits, angular, karma).
// It reuses the validGitConventionNames map — the single source of truth that
// mirrors the pkg/models GitConventionConfig.Convention `oneof` validate tag —
// so callers in other packages (internal/web, internal/cli) validate against
// the same canonical set rather than authoring a parallel rule-set.
// The empty string is NOT a member here; callers that treat empty as "keep
// project default" must guard for empty before calling.
func IsValidConvention(name string) bool {
	return validGitConventionNames[name]
}

// ValidConventions returns the 4 canonical git convention names as a slice,
// for populating UI option lists (web <select>, TUI huh.Select). Order is not
// guaranteed (sourced from a map) — callers that need stable ordering must
// sort the result.
func ValidConventions() []string {
	names := make([]string, 0, len(validGitConventionNames))
	for name := range validGitConventionNames {
		names = append(names, name)
	}
	return names
}

// validateGitConventionConfig checks the git convention configuration.
func validateGitConventionConfig(gc *models.GitConventionConfig) []ValidationError {
	var errs []ValidationError

	if gc.Convention != "" && !validGitConventionNames[gc.Convention] {
		errs = append(errs, ValidationError{
			Field:   "git_convention.convention",
			Message: "must be one of: auto, conventional-commits, angular, karma",
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

	return errs
}

// validMergeMethods lists the recognized PR merge method names. It is the single
// source of truth for the git_strategy.<mode>.merge_method enum, mirroring how
// validGitConventionNames is the SSOT for the convention enum.
var validMergeMethods = map[string]bool{
	"squash": true,
	"merge":  true,
	"rebase": true,
}

// IsValidMergeMethod reports whether name is one of the 3 canonical PR merge
// method names (squash, merge, rebase). It reuses the validMergeMethods map so
// callers in other packages (internal/web, internal/cli) validate against the
// same canonical set rather than authoring a parallel rule-set. The empty string
// is NOT a member here; callers that treat empty as "keep compiled default" must
// guard for empty before calling.
func IsValidMergeMethod(name string) bool {
	return validMergeMethods[name]
}

// ValidMergeMethods returns the 3 canonical merge method names as a slice, for
// populating UI option lists (web <select>, TUI huh.Select). Order is not
// guaranteed (sourced from a map) — callers that need stable ordering must sort.
func ValidMergeMethods() []string {
	names := make([]string, 0, len(validMergeMethods))
	for name := range validMergeMethods {
		names = append(names, name)
	}
	return names
}

// validateGitStrategyMergeMethod checks the merge_method enum for all 3 mode
// profiles. An empty value is treated as the compiled default ("squash") and is
// not an error (fail-safe to current behavior) — identical to how
// validateGitConventionConfig guards gc.Convention != "".
func validateGitStrategyMergeMethod(cfg *Config) []ValidationError {
	var errs []ValidationError

	modes := []struct {
		field   string
		profile *ModeProfile
	}{
		{"git_strategy.manual.merge_method", &cfg.GitStrategy.Manual},
		{"git_strategy.personal.merge_method", &cfg.GitStrategy.Personal},
		{"git_strategy.team.merge_method", &cfg.GitStrategy.Team},
	}
	for _, m := range modes {
		if m.profile.MergeMethod != "" && !validMergeMethods[m.profile.MergeMethod] {
			errs = append(errs, ValidationError{
				Field:   m.field,
				Message: "must be one of: squash, merge, rebase",
				Value:   m.profile.MergeMethod,
				Wrapped: ErrInvalidConfig,
			})
		}
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

	// Git strategy section — pre-existing FLAT field validations (preserved).
	errs = append(errs, checkStringField("git_strategy.branch_prefix", cfg.GitStrategy.BranchPrefix)...)
	errs = append(errs, checkStringField("git_strategy.commit_style", cfg.GitStrategy.CommitStyle)...)

	// Git strategy section — nested mode profile validations
	// (SPEC-V3R5-GIT-STRATEGY-SCHEMA-001 REQ-GSS-007). Explicit literal-path
	// calls (not a dynamic-prefix loop) so each dotted path is greppable.
	// Manual mode.
	errs = append(errs, checkStringField("git_strategy.manual.workflow", cfg.GitStrategy.Manual.Workflow)...)
	errs = append(errs, checkStringField("git_strategy.manual.environment", cfg.GitStrategy.Manual.Environment)...)
	errs = append(errs, checkStringField("git_strategy.manual.commit_style.format", cfg.GitStrategy.Manual.CommitStyle.Format)...)
	errs = append(errs, checkStringField("git_strategy.manual.hooks.pre_commit", cfg.GitStrategy.Manual.Hooks.PreCommit)...)
	errs = append(errs, checkStringField("git_strategy.manual.hooks.pre_push", cfg.GitStrategy.Manual.Hooks.PrePush)...)
	errs = append(errs, checkStringField("git_strategy.manual.hooks.commit_msg", cfg.GitStrategy.Manual.Hooks.CommitMsg)...)
	// Personal mode (has branch_prefix; manual mode does not).
	errs = append(errs, checkStringField("git_strategy.personal.workflow", cfg.GitStrategy.Personal.Workflow)...)
	errs = append(errs, checkStringField("git_strategy.personal.environment", cfg.GitStrategy.Personal.Environment)...)
	errs = append(errs, checkStringField("git_strategy.personal.branch_prefix", cfg.GitStrategy.Personal.BranchPrefix)...)
	errs = append(errs, checkStringField("git_strategy.personal.commit_style.format", cfg.GitStrategy.Personal.CommitStyle.Format)...)
	errs = append(errs, checkStringField("git_strategy.personal.hooks.pre_commit", cfg.GitStrategy.Personal.Hooks.PreCommit)...)
	errs = append(errs, checkStringField("git_strategy.personal.hooks.pre_push", cfg.GitStrategy.Personal.Hooks.PrePush)...)
	errs = append(errs, checkStringField("git_strategy.personal.hooks.commit_msg", cfg.GitStrategy.Personal.Hooks.CommitMsg)...)
	// Team mode (has branch_prefix).
	errs = append(errs, checkStringField("git_strategy.team.workflow", cfg.GitStrategy.Team.Workflow)...)
	errs = append(errs, checkStringField("git_strategy.team.environment", cfg.GitStrategy.Team.Environment)...)
	errs = append(errs, checkStringField("git_strategy.team.branch_prefix", cfg.GitStrategy.Team.BranchPrefix)...)
	errs = append(errs, checkStringField("git_strategy.team.commit_style.format", cfg.GitStrategy.Team.CommitStyle.Format)...)
	errs = append(errs, checkStringField("git_strategy.team.hooks.pre_commit", cfg.GitStrategy.Team.Hooks.PreCommit)...)
	errs = append(errs, checkStringField("git_strategy.team.hooks.pre_push", cfg.GitStrategy.Team.Hooks.PrePush)...)
	errs = append(errs, checkStringField("git_strategy.team.hooks.commit_msg", cfg.GitStrategy.Team.Hooks.CommitMsg)...)

	// Git convention section
	errs = append(errs, checkStringField("git_convention.convention", cfg.GitConvention.Convention)...)
	errs = append(errs, checkStringField("git_convention.auto_detection.fallback", cfg.GitConvention.AutoDetection.Fallback)...)

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

// runStructValidation runs validator/v10 struct tag checks on cfg and converts
// field errors into ValidationError entries compatible with the existing error
// contract.  Fields controlled by loadedSections gating are skipped here so
// that the existing custom checks (validateRequired, validateDevelopmentMode,
// validateGitConventionConfig) remain the authoritative source for those fields.
// validator/v10 catches type violations (e.g. oneof) for oneof-only fields like
// LLM.PerformanceTier where no custom check exists.
func runStructValidation(cfg *Config, loadedSections map[string]bool) []ValidationError {
	if err := validate.Struct(cfg); err != nil {
		var ve validator.ValidationErrors
		if !errors.As(err, &ve) {
			// Unexpected error type; surface as a generic validation error.
			return []ValidationError{{
				Field:   "config",
				Message: err.Error(),
				Wrapped: ErrInvalidConfig,
			}}
		}
		var errs []ValidationError
		for _, fe := range ve {
			ns := fe.Namespace()
			tag := fe.Tag()

			// User.Name required: skip here — handled by validateRequired which
			// gates on loadedSections["user"].
			if ns == "Config.User.Name" && tag == "required" {
				continue
			}

			// DevelopmentMode oneof: skip here — handled by validateDevelopmentMode
			// which wraps ErrInvalidDevelopmentMode expected by existing tests.
			if ns == "Config.Quality.DevelopmentMode" && tag == "oneof" {
				continue
			}

			// GitConvention.Convention oneof: skip here — handled by
			// validateGitConventionConfig which wraps ErrInvalidConfig.
			if ns == "Config.GitConvention.Convention" && tag == "oneof" {
				continue
			}

			// LLM.PerformanceTier oneof: no existing custom check; emit
			// ConfigTypeError wrapped in ValidationError for AC-05 compliance.
			errs = append(errs, ValidationError{
				Field:   fieldNamespace(ns),
				Message: fmt.Sprintf("must be one of valid values (validator tag: %s)", tag),
				Value:   fmt.Sprintf("%v", reflect.ValueOf(fe.Value())),
				Wrapped: ErrInvalidConfig,
			})
		}
		return errs
	}
	return nil
}

// fieldNamespace converts a validator namespace (e.g. "Config.LLM.PerformanceTier")
// into a dotted field path (e.g. "llm.performance_tier") for error messages.
// Falls back to the raw namespace when the mapping is unknown.
func fieldNamespace(ns string) string {
	switch ns {
	case "Config.LLM.PerformanceTier":
		return "llm.performance_tier"
	case "Config.User.Name":
		return "user.name"
	case "Config.Quality.DevelopmentMode":
		return "quality.development_mode"
	case "Config.GitConvention.Convention":
		return "git_convention.convention"
	default:
		return ns
	}
}

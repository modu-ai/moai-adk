package config

import (
	"fmt"
	"strings"
)

// ValidationError represents a configuration validation error with field path
type ValidationError struct {
	Field   string
	Message string
}

// Error returns the error message
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in %s: %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []error
}

// Error returns a combined error message
func (ve *ValidationErrors) Error() string {
	var msgs []string
	for _, err := range ve.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// ValidateConfig performs comprehensive validation of all configuration sections
// It returns nil if all validations pass, or an error if any validation fails
func ValidateConfig(cfg *Config) error {
	var errors []error

	// Validate each section
	if err := validateUserConfig(&cfg.User); err != nil {
		errors = append(errors, fmt.Errorf("user: %w", err))
	}

	if err := validateLanguageConfig(&cfg.Language); err != nil {
		errors = append(errors, fmt.Errorf("language: %w", err))
	}

	if err := validateConstitutionConfig(&cfg.Constitution); err != nil {
		errors = append(errors, fmt.Errorf("constitution: %w", err))
	}

	if err := validateSystemConfig(&cfg.System); err != nil {
		errors = append(errors, fmt.Errorf("system: %w", err))
	}

	if err := validateGitStrategyConfig(&cfg.GitStrategy); err != nil {
		errors = append(errors, fmt.Errorf("git_strategy: %w", err))
	}

	if err := validateProjectConfig(&cfg.Project); err != nil {
		errors = append(errors, fmt.Errorf("project: %w", err))
	}

	if err := validateLLMConfig(&cfg.LLM); err != nil {
		errors = append(errors, fmt.Errorf("llm: %w", err))
	}

	if err := validateServiceConfig(&cfg.Service); err != nil {
		errors = append(errors, fmt.Errorf("service: %w", err))
	}

	if err := validateRalphConfig(&cfg.Ralph); err != nil {
		errors = append(errors, fmt.Errorf("ralph: %w", err))
	}

	if err := validateWorkflowConfig(&cfg.Workflow); err != nil {
		errors = append(errors, fmt.Errorf("workflow: %w", err))
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return &ValidationErrors{Errors: errors}
	}

	return nil
}

// validateUserConfig validates user configuration
func validateUserConfig(cfg *UserConfig) error {
	// User name can be empty (will use default greeting)
	// No validation required
	return nil
}

// validateLanguageConfig validates language configuration
func validateLanguageConfig(cfg *LanguageConfig) error {
	validLangCodes := map[string]bool{
		"ko": true, "en": true, "ja": true, "zh": true,
		"es": true, "fr": true, "de": true,
	}

	// Validate conversation language
	if cfg.ConversationLanguage != "" {
		lang := strings.ToLower(cfg.ConversationLanguage)
		if !validLangCodes[lang] {
			return &ValidationError{
				Field:   "conversation_language",
				Message: fmt.Sprintf("invalid language code '%s', must be one of: ko, en, ja, zh, es, fr, de", cfg.ConversationLanguage),
			}
		}
	}

	return nil
}

// validateConstitutionConfig validates constitution configuration
func validateConstitutionConfig(cfg *ConstitutionConfig) error {
	// Validate development mode
	validModes := map[string]bool{
		"ddd": true,
		"tdd": true,
	}

	if cfg.DevelopmentMode != "" && !validModes[cfg.DevelopmentMode] {
		return &ValidationError{
			Field:   "development_mode",
			Message: fmt.Sprintf("invalid development mode '%s', must be one of: ddd, tdd", cfg.DevelopmentMode),
		}
	}

	// Validate test coverage target
	if cfg.TestCoverageTarget < 0 || cfg.TestCoverageTarget > 100 {
		return &ValidationError{
			Field:   "test_coverage_target",
			Message: fmt.Sprintf("invalid test coverage target '%d', must be between 0 and 100", cfg.TestCoverageTarget),
		}
	}

	// Validate max transformation size
	validSizes := map[string]bool{
		"small":  true,
		"medium": true,
		"large":  true,
	}

	if cfg.DDDDSettings.MaxTransformationSize != "" && !validSizes[cfg.DDDDSettings.MaxTransformationSize] {
		return &ValidationError{
			Field:   "ddd_settings.max_transformation_size",
			Message: fmt.Sprintf("invalid max transformation size '%s', must be one of: small, medium, large", cfg.DDDDSettings.MaxTransformationSize),
		}
	}

	return nil
}

// validateSystemConfig validates system configuration
func validateSystemConfig(cfg *SystemConfig) error {
	// Validate version check frequency
	validFreq := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"never":   true,
	}

	if cfg.MoAI.UpdateCheckFrequency != "" && !validFreq[cfg.MoAI.UpdateCheckFrequency] {
		return &ValidationError{
			Field:   "moai.update_check_frequency",
			Message: fmt.Sprintf("invalid update check frequency '%s', must be one of: daily, weekly, monthly, never", cfg.MoAI.UpdateCheckFrequency),
		}
	}

	// Validate spec git workflow
	validWorkflows := map[string]bool{
		"main_direct":    true,
		"main_feature":   true,
		"develop_direct": true,
		"feature_branch": true,
		"per_spec":       true,
	}

	if cfg.GitHub.SpecGitWorkflow != "" && !validWorkflows[cfg.GitHub.SpecGitWorkflow] {
		return &ValidationError{
			Field:   "github.spec_git_workflow",
			Message: fmt.Sprintf("invalid spec git workflow '%s', must be one of: main_direct, main_feature, develop_direct, feature_branch, per_spec", cfg.GitHub.SpecGitWorkflow),
		}
	}

	return nil
}

// validateGitStrategyConfig validates git strategy configuration
func validateGitStrategyConfig(cfg *GitStrategyConfig) error {
	validModes := map[string]bool{
		"manual":   true,
		"personal": true,
		"team":     true,
	}

	if cfg.Mode != "" && !validModes[cfg.Mode] {
		return &ValidationError{
			Field:   "mode",
			Message: fmt.Sprintf("invalid git strategy mode '%s', must be one of: manual, personal, team", cfg.Mode),
		}
	}

	validWorkflows := map[string]bool{
		"github-flow": true,
		"git-flow":    true,
	}

	// Validate manual mode
	if cfg.Manual.Workflow != "" && !validWorkflows[cfg.Manual.Workflow] {
		return &ValidationError{
			Field:   "manual.workflow",
			Message: fmt.Sprintf("invalid workflow '%s', must be one of: github-flow, git-flow", cfg.Manual.Workflow),
		}
	}

	// Validate personal mode
	if cfg.Personal.Workflow != "" && !validWorkflows[cfg.Personal.Workflow] {
		return &ValidationError{
			Field:   "personal.workflow",
			Message: fmt.Sprintf("invalid workflow '%s', must be one of: github-flow, git-flow", cfg.Personal.Workflow),
		}
	}

	// Validate team mode
	if cfg.Team.Workflow != "" && !validWorkflows[cfg.Team.Workflow] {
		return &ValidationError{
			Field:   "team.workflow",
			Message: fmt.Sprintf("invalid workflow '%s', must be one of: github-flow, git-flow", cfg.Team.Workflow),
		}
	}

	validCommitStyles := map[string]bool{
		"conventional": true,
		"simple":       true,
	}

	if cfg.Manual.CommitStyle.Format != "" && !validCommitStyles[cfg.Manual.CommitStyle.Format] {
		return &ValidationError{
			Field:   "manual.commit_style.format",
			Message: fmt.Sprintf("invalid commit style '%s', must be one of: conventional, simple", cfg.Manual.CommitStyle.Format),
		}
	}

	return nil
}

// validateProjectConfig validates project configuration
func validateProjectConfig(cfg *ProjectConfig) error {
	// All fields are optional, no strict validation required
	return nil
}

// validateLLMConfig validates LLM configuration
func validateLLMConfig(cfg *LLMConfig) error {
	validModes := map[string]bool{
		"claude-only": true,
		"mashup":      true,
		"glm-only":    true,
	}

	if cfg.Mode != "" && !validModes[cfg.Mode] {
		return &ValidationError{
			Field:   "mode",
			Message: fmt.Sprintf("invalid LLM mode '%s', must be one of: claude-only, mashup, glm-only", cfg.Mode),
		}
	}

	// Validate GLM models
	if cfg.GLM.Models != nil {
		validModels := map[string]bool{
			"haiku":  true,
			"sonnet": true,
			"opus":   true,
		}

		for model := range cfg.GLM.Models {
			if !validModels[model] {
				return &ValidationError{
					Field:   "glm.models",
					Message: fmt.Sprintf("invalid GLM model key '%s', must be one of: haiku, sonnet, opus", model),
				}
			}
		}
	}

	return nil
}

// validateServiceConfig validates service configuration
func validateServiceConfig(cfg *ServiceConfig) error {
	validTypes := map[string]bool{
		"claude_subscription": true,
		"claude_api":          true,
		"glm":                 true,
		"hybrid":              true,
	}

	if cfg.Type != "" && !validTypes[cfg.Type] {
		return &ValidationError{
			Field:   "type",
			Message: fmt.Sprintf("invalid service type '%s', must be one of: claude_subscription, claude_api, glm, hybrid", cfg.Type),
		}
	}

	validPlans := map[string]bool{
		"pro":     true,
		"max5":    true,
		"max20":   true,
		"basic":   true,
		"glm_pro": true,
	}

	if cfg.PricingPlan != "" && !validPlans[cfg.PricingPlan] {
		return &ValidationError{
			Field:   "pricing_plan",
			Message: fmt.Sprintf("invalid pricing plan '%s', must be one of: pro, max5, max20, basic, glm_pro", cfg.PricingPlan),
		}
	}

	validStrategies := map[string]bool{
		"auto":   true,
		"custom": true,
	}

	if cfg.ModelAllocation.Strategy != "" && !validStrategies[cfg.ModelAllocation.Strategy] {
		return &ValidationError{
			Field:   "model_allocation.strategy",
			Message: fmt.Sprintf("invalid allocation strategy '%s', must be one of: auto, custom", cfg.ModelAllocation.Strategy),
		}
	}

	return nil
}

// validateRalphConfig validates Ralph configuration
func validateRalphConfig(cfg *RalphConfig) error {
	// Validate LSP timeout
	if cfg.LSP.TimeoutSeconds < 1 || cfg.LSP.TimeoutSeconds > 300 {
		return &ValidationError{
			Field:   "lsp.timeout_seconds",
			Message: fmt.Sprintf("invalid timeout '%d', must be between 1 and 300 seconds", cfg.LSP.TimeoutSeconds),
		}
	}

	// Validate poll interval
	if cfg.LSP.PollIntervalMs < 100 || cfg.LSP.PollIntervalMs > 60000 {
		return &ValidationError{
			Field:   "lsp.poll_interval_ms",
			Message: fmt.Sprintf("invalid poll interval '%d', must be between 100 and 60000 ms", cfg.LSP.PollIntervalMs),
		}
	}

	// Validate loop max iterations
	if cfg.Loop.MaxIterations < 1 || cfg.Loop.MaxIterations > 1000 {
		return &ValidationError{
			Field:   "loop.max_iterations",
			Message: fmt.Sprintf("invalid max iterations '%d', must be between 1 and 1000", cfg.Loop.MaxIterations),
		}
	}

	// Validate coverage threshold
	if cfg.Loop.Completion.CoverageThreshold < 0 || cfg.Loop.Completion.CoverageThreshold > 100 {
		return &ValidationError{
			Field:   "loop.completion.coverage_threshold",
			Message: fmt.Sprintf("invalid coverage threshold '%d', must be between 0 and 100", cfg.Loop.Completion.CoverageThreshold),
		}
	}

	validSeverities := map[string]bool{
		"error":   true,
		"warning": true,
		"info":    true,
	}

	if cfg.Hooks.PostToolLSP.SeverityThreshold != "" && !validSeverities[cfg.Hooks.PostToolLSP.SeverityThreshold] {
		return &ValidationError{
			Field:   "hooks.post_tool_lsp.severity_threshold",
			Message: fmt.Sprintf("invalid severity threshold '%s', must be one of: error, warning, info", cfg.Hooks.PostToolLSP.SeverityThreshold),
		}
	}

	return nil
}

// validateWorkflowConfig validates workflow configuration
func validateWorkflowConfig(cfg *WorkflowConfig) error {
	// Validate loop prevention settings
	if cfg.LoopPrevention.MaxIterations < 1 || cfg.LoopPrevention.MaxIterations > 1000 {
		return &ValidationError{
			Field:   "loop_prevention.max_iterations",
			Message: fmt.Sprintf("invalid max iterations '%d', must be between 1 and 1000", cfg.LoopPrevention.MaxIterations),
		}
	}

	if cfg.LoopPrevention.NoProgressThreshold < 1 || cfg.LoopPrevention.NoProgressThreshold > 100 {
		return &ValidationError{
			Field:   "loop_prevention.no_progress_threshold",
			Message: fmt.Sprintf("invalid no progress threshold '%d', must be between 1 and 100", cfg.LoopPrevention.NoProgressThreshold),
		}
	}

	return nil
}

package config

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/pkg/models"
)

func TestValidateValidConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.User.Name = "TestUser"
	loaded := map[string]bool{"user": true}

	err := Validate(cfg, loaded)
	if err != nil {
		t.Errorf("Validate() expected no error for valid config, got: %v", err)
	}
}

func TestValidateDefaultConfigNoLoadedSections(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	loaded := map[string]bool{}

	// Default config with no loaded sections should pass
	err := Validate(cfg, loaded)
	if err != nil {
		t.Errorf("Validate() expected no error for defaults, got: %v", err)
	}
}

func TestValidateRequiredUserNameWhenLoaded(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	// User section loaded but name is empty
	loaded := map[string]bool{"user": true}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("Validate() expected error for empty user.name when user section loaded")
	}

	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if e.Field == "user.name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation error for field user.name")
	}
}

func TestValidateUserNameNotRequiredWhenNotLoaded(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	// User section NOT loaded, so empty name is acceptable
	loaded := map[string]bool{}

	err := Validate(cfg, loaded)
	if err != nil {
		t.Errorf("Validate() expected no error when user section not loaded, got: %v", err)
	}
}

func TestValidateDevelopmentMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mode    models.DevelopmentMode
		wantErr bool
	}{
		{"ddd is valid", models.ModeDDD, false},
		{"tdd is valid", models.ModeTDD, false},
		{"empty is valid (defaults applied)", "", false},
		{"waterfall is invalid", models.DevelopmentMode("waterfall"), true},
		{"agile is invalid", models.DevelopmentMode("agile"), true},
		{"DDD uppercase is invalid", models.DevelopmentMode("DDD"), true},
		{"random string is invalid", models.DevelopmentMode("foobar"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.User.Name = "TestUser"
			cfg.Quality.DevelopmentMode = tt.mode
			loaded := map[string]bool{"user": true}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error for mode %q, got nil", tt.mode)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() expected no error for mode %q, got: %v", tt.mode, err)
			}

			if tt.wantErr && err != nil {
				if !errors.Is(err, ErrInvalidDevelopmentMode) {
					t.Errorf("expected ErrInvalidDevelopmentMode, got: %v", err)
				}
			}
		})
	}
}

func TestValidateCoverageTargetBounds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		target  int
		wantErr bool
	}{
		{"0 is valid lower bound", 0, false},
		{"50 is valid", 50, false},
		{"85 is valid", 85, false},
		{"100 is valid upper bound", 100, false},
		{"-1 is invalid", -1, true},
		{"101 is invalid", 101, true},
		{"-100 is invalid", -100, true},
		{"200 is invalid", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.Quality.TestCoverageTarget = tt.target
			loaded := map[string]bool{}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error for coverage target %d, got nil", tt.target)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() expected no error for coverage target %d, got: %v", tt.target, err)
			}
		})
	}
}

func TestValidateTDDMinCoveragePerCommit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"0 is valid", 0, false},
		{"80 is valid", 80, false},
		{"100 is valid", 100, false},
		{"-1 is invalid", -1, true},
		{"101 is invalid", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.Quality.TDDSettings.MinCoveragePerCommit = tt.value
			loaded := map[string]bool{}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for MinCoveragePerCommit %d", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for MinCoveragePerCommit %d, got: %v", tt.value, err)
			}
		})
	}
}

func TestValidateMaxExemptPercentage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"0 is valid", 0, false},
		{"5 is valid", 5, false},
		{"100 is valid", 100, false},
		{"-1 is invalid", -1, true},
		{"101 is invalid", 101, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.Quality.CoverageExemptions.MaxExemptPercentage = tt.value
			loaded := map[string]bool{}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for MaxExemptPercentage %d", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for MaxExemptPercentage %d, got: %v", tt.value, err)
			}
		})
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.User.Name = "" // required when loaded
	cfg.Quality.DevelopmentMode = "invalid_mode"
	cfg.Quality.TestCoverageTarget = -1
	loaded := map[string]bool{"user": true}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}

	// Should have at least 3 errors: user.name, development_mode, test_coverage_target
	if len(ve.Errors) < 3 {
		t.Errorf("expected at least 3 validation errors, got %d: %v", len(ve.Errors), ve.Errors)
	}
}

func TestValidateDynamicTokens(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		field   string
		value   string
		wantErr bool
	}{
		{"no token", "user.name", "Alice", false},
		{"dollar brace token", "user.name", "${USER}", true},
		{"double brace token", "user.name", "{{USER}}", true},
		{"dollar var token", "user.name", "$HOME_DIR", true},
		{"empty value is ok", "user.name", "", false},
		{"partial match prefix", "user.name", "prefix_${VAR}_suffix", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs := checkStringField(tt.field, tt.value)
			if tt.wantErr && len(errs) == 0 {
				t.Errorf("expected error for value %q, got none", tt.value)
			}
			if !tt.wantErr && len(errs) > 0 {
				t.Errorf("expected no error for value %q, got: %v", tt.value, errs)
			}

			if tt.wantErr && len(errs) > 0 {
				if !errors.Is(&errs[0], ErrDynamicToken) {
					t.Errorf("expected ErrDynamicToken, got: %v", errs[0].Wrapped)
				}
			}
		})
	}
}

func TestValidateDynamicTokensInConfig(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.User.Name = "${MOAI_USER}"
	loaded := map[string]bool{"user": true}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected error for dynamic token in user.name")
	}
	if !errors.Is(err, ErrDynamicToken) {
		t.Errorf("expected ErrDynamicToken, got: %v", err)
	}
}

func TestValidateDynamicTokensInLanguageFields(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.Language.ConversationLanguage = "{{LANG}}"
	loaded := map[string]bool{}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected error for dynamic token in language field")
	}
	if !errors.Is(err, ErrDynamicToken) {
		t.Errorf("expected ErrDynamicToken, got: %v", err)
	}
}

func TestValidateDynamicTokensInSystemFields(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.System.LogLevel = "$LOG_LEVEL"
	loaded := map[string]bool{}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected error for dynamic token in system.log_level")
	}
	if !errors.Is(err, ErrDynamicToken) {
		t.Errorf("expected ErrDynamicToken, got: %v", err)
	}
}

func TestValidationErrorFormat(t *testing.T) {
	t.Parallel()

	t.Run("with value", func(t *testing.T) {
		t.Parallel()
		ve := ValidationError{
			Field:   "quality.development_mode",
			Message: "must be one of: ddd, tdd",
			Value:   "waterfall",
		}
		got := ve.Error()
		if got == "" {
			t.Error("Error() returned empty string")
		}
		// Should contain the field name and value
		for _, want := range []string{"quality.development_mode", "waterfall"} {
			if !containsSubstring(got, want) {
				t.Errorf("Error() output %q does not contain %q", got, want)
			}
		}
	})

	t.Run("without value", func(t *testing.T) {
		t.Parallel()
		ve := ValidationError{
			Field:   "user.name",
			Message: "required field is empty",
		}
		got := ve.Error()
		if got == "" {
			t.Error("Error() returned empty string")
		}
		if !containsSubstring(got, "user.name") {
			t.Errorf("Error() output %q does not contain field name", got)
		}
	})
}

func TestValidationErrorUnwrap(t *testing.T) {
	t.Parallel()

	ve := ValidationError{
		Field:   "test",
		Message: "test error",
		Wrapped: ErrInvalidConfig,
	}
	if !errors.Is(&ve, ErrInvalidConfig) {
		t.Error("expected Unwrap to return ErrInvalidConfig")
	}
}

func TestValidationErrorsErrorFormat(t *testing.T) {
	t.Parallel()

	t.Run("with errors", func(t *testing.T) {
		t.Parallel()
		ve := &ValidationErrors{
			Errors: []ValidationError{
				{Field: "a", Message: "error 1"},
				{Field: "b", Message: "error 2"},
			},
		}
		got := ve.Error()
		if !containsSubstring(got, "2 error(s)") {
			t.Errorf("Error() output %q does not contain error count", got)
		}
	})

	t.Run("empty errors", func(t *testing.T) {
		t.Parallel()
		ve := &ValidationErrors{}
		got := ve.Error()
		if !containsSubstring(got, "no errors") {
			t.Errorf("Error() output %q does not contain 'no errors'", got)
		}
	})
}

func TestValidationErrorsIs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		errors []ValidationError
		target error
		want   bool
	}{
		{
			name:   "matches ErrInvalidConfig always",
			errors: []ValidationError{},
			target: ErrInvalidConfig,
			want:   true,
		},
		{
			name: "matches wrapped error",
			errors: []ValidationError{
				{Field: "test", Wrapped: ErrDynamicToken},
			},
			target: ErrDynamicToken,
			want:   true,
		},
		{
			name: "matches ErrInvalidDevelopmentMode",
			errors: []ValidationError{
				{Field: "test", Wrapped: ErrInvalidDevelopmentMode},
			},
			target: ErrInvalidDevelopmentMode,
			want:   true,
		},
		{
			name: "does not match unrelated error",
			errors: []ValidationError{
				{Field: "test", Wrapped: ErrDynamicToken},
			},
			target: ErrNotInitialized,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ve := &ValidationErrors{Errors: tt.errors}
			if got := errors.Is(ve, tt.target); got != tt.want {
				t.Errorf("errors.Is() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateGitConventionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		conv    string
		wantErr bool
	}{
		{"auto is valid", "auto", false},
		{"conventional-commits is valid", "conventional-commits", false},
		{"angular is valid", "angular", false},
		{"karma is valid", "karma", false},
		{"custom is invalid (engine removed)", "custom", true},
		{"empty is valid (defaults applied)", "", false},
		{"invalid convention", "gitmoji", true},
		{"uppercase is invalid", "AUTO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.GitConvention.Convention = tt.conv
			loaded := map[string]bool{}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("Validate() expected error for convention %q, got nil", tt.conv)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("Validate() expected no error for convention %q, got: %v", tt.conv, err)
			}
		})
	}
}

func TestValidateGitConventionSampleSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		value   int
		wantErr bool
	}{
		{"0 is valid", 0, false},
		{"100 is valid", 100, false},
		{"-1 is invalid", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.GitConvention.AutoDetection.SampleSize = tt.value
			loaded := map[string]bool{}

			err := Validate(cfg, loaded)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for SampleSize %d", tt.value)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("expected no error for SampleSize %d, got: %v", tt.value, err)
			}
		})
	}
}

func TestValidateGitConventionDynamicTokens(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.GitConvention.Convention = "${GIT_CONV}"
	loaded := map[string]bool{}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected error for dynamic token in git_convention.convention")
	}
	if !errors.Is(err, ErrDynamicToken) {
		t.Errorf("expected ErrDynamicToken, got: %v", err)
	}
}

// findFieldMessage returns the Message of the first ValidationError whose Field
// matches, or "" when no such error exists.
func findFieldMessage(errs []ValidationError, field string) string {
	for _, e := range errs {
		if e.Field == field {
			return e.Message
		}
	}
	return ""
}

// TestValidateQualitySection covers SPEC-WEB-CONSOLE-007 AC-WC7-005: the exported
// ValidateQualitySection seam reuses the EXACT existing quality value-range rules
// (test_coverage_target / tdd_settings.min_coverage_per_commit 0-100) with the
// existing "must be between 0 and 100" message — proving the seam forwards to
// validateQualityConfig and adds no new rule (REQ-WC7-002, CRITICAL SCOPE CONSTRAINT).
func TestValidateQualitySection(t *testing.T) {
	t.Parallel()

	t.Run("valid in-range passes", func(t *testing.T) {
		t.Parallel()
		q := &models.QualityConfig{TestCoverageTarget: 85}
		q.TDDSettings.MinCoveragePerCommit = 80
		if errs := ValidateQualitySection(q); len(errs) != 0 {
			t.Errorf("in-range quality should pass, got %v", errs)
		}
	})

	t.Run("test_coverage_target out of range reuses existing message", func(t *testing.T) {
		t.Parallel()
		q := &models.QualityConfig{TestCoverageTarget: 150}
		msg := findFieldMessage(ValidateQualitySection(q), "quality.test_coverage_target")
		if msg != "must be between 0 and 100" {
			t.Errorf("test_coverage_target=150 message = %q, want existing %q", msg, "must be between 0 and 100")
		}
	})

	t.Run("tdd_settings.min_coverage_per_commit negative reuses existing message", func(t *testing.T) {
		t.Parallel()
		q := &models.QualityConfig{}
		q.TDDSettings.MinCoveragePerCommit = -5
		msg := findFieldMessage(ValidateQualitySection(q), "quality.tdd_settings.min_coverage_per_commit")
		if msg != "must be between 0 and 100" {
			t.Errorf("min_coverage_per_commit=-5 message = %q, want existing %q", msg, "must be between 0 and 100")
		}
	})
}

// TestValidateGitConventionSection covers the exported ValidateGitConventionSection
// seam: it reuses the EXACT existing git-convention rules (convention 4-value enum +
// confidence_threshold [0.0,1.0]) with the existing messages — proving the seam
// forwards to validateGitConventionConfig and adds no new rule (REQ-WC7-008 CRITICAL
// SCOPE CONSTRAINT, REQ-WC9-003 custom removal).
func TestValidateGitConventionSection(t *testing.T) {
	t.Parallel()

	t.Run("custom convention is rejected (engine removed)", func(t *testing.T) {
		t.Parallel()
		gc := &models.GitConventionConfig{Convention: "custom"}
		msg := findFieldMessage(ValidateGitConventionSection(gc), "git_convention.convention")
		if msg != "must be one of: auto, conventional-commits, angular, karma" {
			t.Errorf("custom convention message = %q, want enum message", msg)
		}
	})

	t.Run("confidence_threshold out of range reuses existing message", func(t *testing.T) {
		t.Parallel()
		gc := &models.GitConventionConfig{}
		gc.AutoDetection.ConfidenceThreshold = 1.5
		msg := findFieldMessage(ValidateGitConventionSection(gc), "git_convention.auto_detection.confidence_threshold")
		if msg != "must be between 0.0 and 1.0" {
			t.Errorf("confidence_threshold=1.5 message = %q, want existing %q", msg, "must be between 0.0 and 1.0")
		}
	})

	t.Run("valid 4-value convention passes", func(t *testing.T) {
		t.Parallel()
		gc := &models.GitConventionConfig{Convention: "angular"}
		gc.AutoDetection.ConfidenceThreshold = 0.75
		if errs := ValidateGitConventionSection(gc); len(errs) != 0 {
			t.Errorf("valid convention should pass, got %v", errs)
		}
	})
}

func TestDevelopmentModeStrings(t *testing.T) {
	t.Parallel()

	strs := developmentModeStrings()
	if len(strs) != 2 {
		t.Fatalf("expected 2 mode strings, got %d", len(strs))
	}

	expected := map[string]bool{"ddd": true, "tdd": true}
	for _, s := range strs {
		if !expected[s] {
			t.Errorf("unexpected mode string: %q", s)
		}
	}
}

// TestValidate_RequiredFieldMissing exercises the validator/v10 required path
// for User.Name when the user section is explicitly loaded.
// AC-05 hardening: validator now produces type errors for required-field-missing cases.
func TestValidate_RequiredFieldMissing(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	// User.Name is empty — the validate:"required" tag on UserConfig.Name
	// triggers the validator/v10 path (skipped by runStructValidation and
	// delegated to the existing validateRequired custom check).
	loaded := map[string]bool{"user": true}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("Validate() expected error when User.Name is empty and user section is loaded")
	}

	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T: %v", err, err)
	}

	found := false
	for _, e := range ve.Errors {
		if e.Field == "user.name" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation error for field user.name in: %v", ve.Errors)
	}
}

// TestValidate_OneofViolation exercises the validator/v10 oneof path for
// LLM.PerformanceTier.  Invalid values must produce an error that wraps
// ErrInvalidConfig (AC-05 ConfigTypeError alignment).
func TestValidate_OneofViolation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tier  string
		valid bool
	}{
		{"high is valid", "high", true},
		{"medium is valid", "medium", true},
		{"low is valid", "low", true},
		{"empty is valid (omitempty)", "", true},
		{"ultra is invalid", "ultra", false},
		{"HIGH uppercase is invalid", "HIGH", false},
		{"extreme is invalid", "extreme", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.User.Name = "TestUser"
			cfg.LLM.PerformanceTier = tt.tier
			loaded := map[string]bool{"user": true}

			err := Validate(cfg, loaded)
			if tt.valid && err != nil {
				t.Errorf("Validate() expected no error for tier %q, got: %v", tt.tier, err)
			}
			if !tt.valid && err == nil {
				t.Errorf("Validate() expected error for tier %q, got nil", tt.tier)
			}
			if !tt.valid && err != nil {
				// Must be identifiable as a config error.
				if !errors.Is(err, ErrInvalidConfig) {
					t.Errorf("expected ErrInvalidConfig for tier %q, got: %v", tt.tier, err)
				}
			}
		})
	}
}

// containsSubstring is a test helper that checks if s contains substr.
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

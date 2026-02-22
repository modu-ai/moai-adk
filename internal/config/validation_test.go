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
		{"custom is valid", "custom", false},
		{"empty is valid (defaults applied)", "", false},
		{"invalid convention", "gitmoji", true},
		{"uppercase is invalid", "AUTO", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := NewDefaultConfig()
			cfg.GitConvention.Convention = tt.conv
			// When convention is "custom", pattern is required.
			if tt.conv == "custom" {
				cfg.GitConvention.Custom.Pattern = `^.+$`
			}
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

func TestValidateGitConventionCustomRequiresPattern(t *testing.T) {
	t.Parallel()

	cfg := NewDefaultConfig()
	cfg.GitConvention.Convention = "custom"
	cfg.GitConvention.Custom.Pattern = "" // missing pattern
	loaded := map[string]bool{}

	err := Validate(cfg, loaded)
	if err == nil {
		t.Fatal("expected error for custom convention without pattern")
	}

	var ve *ValidationErrors
	if !errors.As(err, &ve) {
		t.Fatalf("expected *ValidationErrors, got %T", err)
	}

	found := false
	for _, e := range ve.Errors {
		if e.Field == "git_convention.custom.pattern" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected validation error for field git_convention.custom.pattern")
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

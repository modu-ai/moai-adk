// Package harness — HRN-003 M4 profile loader + malformed profile rejection tests.
// REQ-HRN-003-005: ParseRubricMarkdown.
// REQ-HRN-003-019: 5th dimension rejection.
// REQ-HRN-003-018: must-pass bypass rejection.
package harness

import (
	"errors"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestParseRubricMarkdown_RejectsFifthDimension verifies that a profile declaring
// the unknown dimension "Performance" is rejected at the Validate() layer.
// malformed-5dim.md declares Functionality/Security/Craft/Performance
// (Consistency missing + Performance unknown).
// The parser is lenient and skips unknown dimensions, leaving 3 canonical dims.
// Rubric.Validate() then returns ErrInvalidConfig from the 4-dimension count check.
// REQ-HRN-003-019, AC-HRN-003-08.
func TestParseRubricMarkdown_RejectsFifthDimension(t *testing.T) {
	rubric, err := ParseRubricMarkdown("testdata/profiles/malformed-5dim.md")
	if err != nil {
		// If the parser raises an error (e.g., ErrUnknownDimension) — correct behavior.
		if errors.Is(err, config.ErrUnknownDimension) || errors.Is(err, config.ErrInvalidConfig) {
			return
		}
		t.Fatalf("ParseRubricMarkdown(malformed-5dim.md) unexpected error = %v", err)
	}

	// If the parser returned a rubric, verify it via Validate().
	// In malformed-5dim.md, Performance is skipped, leaving 3 canonical dims.
	// Validate() returns ErrInvalidConfig due to the "exactly 4 dimensions" violation.
	if rubric == nil {
		t.Fatal("ParseRubricMarkdown returned nil rubric without error")
	}
	validateErr := rubric.Validate()
	if validateErr == nil {
		t.Fatal("Rubric.Validate() = nil, want error for profile with missing canonical dimension")
	}
	if !errors.Is(validateErr, config.ErrInvalidConfig) {
		t.Errorf("Rubric.Validate() error = %v, want ErrInvalidConfig", validateErr)
	}
}

// TestParseRubricMarkdown_MustPassBypassRejected verifies that a profile excluding
// Security from Must-Pass returns ErrMustPassBypassProhibited.
// REQ-HRN-003-018, AC-HRN-003-11.
func TestParseRubricMarkdown_MustPassBypassRejected(t *testing.T) {
	rubric, err := ParseRubricMarkdown("testdata/profiles/malformed-bypass.md")
	if err != nil {
		// Case where the parser rejects directly (ErrMustPassBypassProhibited wrapped).
		if errors.Is(err, config.ErrMustPassBypassProhibited) {
			return
		}
		t.Fatalf("ParseRubricMarkdown(malformed-bypass.md) unexpected error = %v", err)
	}

	if rubric == nil {
		t.Fatal("ParseRubricMarkdown returned nil rubric, want non-nil")
	}

	// Validate() must catch the MustPass floor violation.
	validateErr := rubric.Validate()
	if validateErr == nil {
		t.Fatal("Rubric.Validate() = nil, want ErrMustPassBypassProhibited for Security-excluded profile")
	}
	if !errors.Is(validateErr, config.ErrMustPassBypassProhibited) {
		t.Errorf("Rubric.Validate() error = %v, want ErrMustPassBypassProhibited", validateErr)
	}
}

// TestLoadHarnessConfigWithProfiles verifies the Profiles field added in HRN-003 M4.
// LoadHarnessConfig must return a Profiles map.
// AC-HRN-003-07.c.
func TestLoadHarnessConfigWithProfiles(t *testing.T) {
	cfg, err := loadEvaluatorConfig()
	if err != nil {
		t.Skipf("loadEvaluatorConfig() not yet implemented: %v", err)
	}
	if cfg == nil {
		t.Fatal("loadEvaluatorConfig() returned nil")
	}
	// The Profiles map must be present (at least the default profile included).
	if len(cfg.Profiles) == 0 {
		t.Error("EvaluatorConfig.Profiles is empty, want at least one profile")
	}
}

// TestEvaluatorConfig_DefaultsSet verifies the default values of EvaluatorConfig
// are set correctly.
// M4 T4.5: defaults for Aggregation + MustPassDimensions.
func TestEvaluatorConfig_DefaultsSet(t *testing.T) {
	cfg := newDefaultEvaluatorConfig()
	if cfg.Aggregation != "min" {
		t.Errorf("EvaluatorConfig.Aggregation = %q, want %q", cfg.Aggregation, "min")
	}
	if len(cfg.MustPassDimensions) == 0 {
		t.Error("EvaluatorConfig.MustPassDimensions is empty, want default [Functionality, Security]")
	}
}

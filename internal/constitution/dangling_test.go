package constitution_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestValidateRuleReferencesReturnsWarningForUnknownID verifies that a warning string
// is returned for a non-existent ID.
// Direct mapping to AC-CON-001-011.
func TestValidateRuleReferencesReturnsWarningForUnknownID(t *testing.T) {
	t.Parallel()

	// CONST-V3R2-999 is not in valid_registry.md
	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{"CONST-V3R2-999"})

	if len(warnings) < 1 {
		t.Fatalf("warning count = %d, must be >= 1", len(warnings))
	}

	first := warnings[0]
	if first == "" {
		t.Error("first warning must not be empty")
	}

	const wantSubstr = "CONST-V3R2-999"
	if !strings.Contains(first, wantSubstr) {
		t.Errorf("warning %q must contain %q", first, wantSubstr)
	}
}

// TestValidateRuleReferencesNoWarningForKnownID verifies that no warning is returned for an existing ID.
func TestValidateRuleReferencesNoWarningForKnownID(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{"CONST-V3R2-001"})

	if len(warnings) != 0 {
		t.Errorf("warning count for existing ID = %d, must be 0; warnings: %v", len(warnings), warnings)
	}
}

// TestValidateRuleReferencesEmptyRefs verifies that an empty result is returned for an empty refs slice.
func TestValidateRuleReferencesEmptyRefs(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{})

	if len(warnings) != 0 {
		t.Errorf("warning count for empty refs = %d, must be 0", len(warnings))
	}
}

// TestValidateRuleReferencesMixedRefs verifies behavior for mixed refs (existing and non-existing).
func TestValidateRuleReferencesMixedRefs(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry error: %v", err)
	}

	refs := []string{"CONST-V3R2-001", "CONST-V3R2-999", "CONST-V3R2-002", "CONST-V3R2-888"}
	warnings := constitution.ValidateRuleReferences(reg, refs)

	// CONST-V3R2-999 and CONST-V3R2-888 are absent, so 2 warnings expected
	if len(warnings) != 2 {
		t.Errorf("mixed refs warning count = %d, want 2; warnings: %v", len(warnings), warnings)
	}
}

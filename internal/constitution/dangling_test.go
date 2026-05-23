package constitution_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestValidateRuleReferencesReturnsWarningForUnknownID verifies that a
// warning string is returned for an ID that does not exist.
// Direct mapping to AC-CON-001-011.
func TestValidateRuleReferencesReturnsWarningForUnknownID(t *testing.T) {
	t.Parallel()

	// valid_registry.md does not contain CONST-V3R2-999.
	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{"CONST-V3R2-999"})

	if len(warnings) < 1 {
		t.Fatalf("경고 수 = %d, ≥1이어야 한다", len(warnings))
	}

	first := warnings[0]
	if first == "" {
		t.Error("첫 번째 경고가 비어 있어서는 안 된다")
	}

	const wantSubstr = "CONST-V3R2-999"
	if !strings.Contains(first, wantSubstr) {
		t.Errorf("경고 %q에 %q가 포함되어야 한다", first, wantSubstr)
	}
}

// TestValidateRuleReferencesNoWarningForKnownID verifies that no warning is returned for an existing ID.
func TestValidateRuleReferencesNoWarningForKnownID(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{"CONST-V3R2-001"})

	if len(warnings) != 0 {
		t.Errorf("존재하는 ID에 대한 경고 수 = %d, 0이어야 한다; 경고: %v", len(warnings), warnings)
	}
}

// TestValidateRuleReferencesEmptyRefs verifies that an empty refs slice returns an empty result.
func TestValidateRuleReferencesEmptyRefs(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	warnings := constitution.ValidateRuleReferences(reg, []string{})

	if len(warnings) != 0 {
		t.Errorf("빈 refs에 대한 경고 수 = %d, 0이어야 한다", len(warnings))
	}
}

// TestValidateRuleReferencesMixedRefs verifies behavior for mixed refs (existing and non-existing).
func TestValidateRuleReferencesMixedRefs(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	refs := []string{"CONST-V3R2-001", "CONST-V3R2-999", "CONST-V3R2-002", "CONST-V3R2-888"}
	warnings := constitution.ValidateRuleReferences(reg, refs)

	// CONST-V3R2-999 and CONST-V3R2-888 are missing, so two warnings are expected.
	if len(warnings) != 2 {
		t.Errorf("혼합 refs 경고 수 = %d, want 2; 경고: %v", len(warnings), warnings)
	}
}

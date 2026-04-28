package constitution_test

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestValidateRuleReferencesReturnsWarningForUnknownID는 존재하지 않는 ID에 대해
// 경고 문자열을 반환함을 검증한다.
// AC-CON-001-011 직접 매핑.
func TestValidateRuleReferencesReturnsWarningForUnknownID(t *testing.T) {
	t.Parallel()

	// valid_registry.md에는 CONST-V3R2-999가 없다
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

// TestValidateRuleReferencesNoWarningForKnownID는 존재하는 ID에 대해 경고를 반환하지 않음을 검증한다.
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

// TestValidateRuleReferencesEmptyRefs는 빈 refs 슬라이스에 대해 빈 결과를 반환함을 검증한다.
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

// TestValidateRuleReferencesMixedRefs는 혼합된 refs (존재/비존재)에 대한 동작을 검증한다.
func TestValidateRuleReferencesMixedRefs(t *testing.T) {
	t.Parallel()

	reg, err := constitution.LoadRegistry(testdataPath("valid_registry.md"), ".")
	if err != nil {
		t.Fatalf("LoadRegistry 오류: %v", err)
	}

	refs := []string{"CONST-V3R2-001", "CONST-V3R2-999", "CONST-V3R2-002", "CONST-V3R2-888"}
	warnings := constitution.ValidateRuleReferences(reg, refs)

	// CONST-V3R2-999와 CONST-V3R2-888이 없으므로 경고 2개
	if len(warnings) != 2 {
		t.Errorf("혼합 refs 경고 수 = %d, want 2; 경고: %v", len(warnings), warnings)
	}
}

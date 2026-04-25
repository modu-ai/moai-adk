package constitution_test

import (
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestRuleStructFieldsMatchRegistrySchema는 Rule 구조체가 정확히 6개 exported 필드를 가짐을 검증한다.
// AC-CON-001-004 직접 매핑.
func TestRuleStructFieldsMatchRegistrySchema(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeOf(constitution.Rule{})

	// exported 필드만 카운트
	var exportedFields []string
	for i := range rt.NumField() {
		f := rt.Field(i)
		if f.IsExported() {
			exportedFields = append(exportedFields, f.Name)
		}
	}

	const wantCount = 6
	if len(exportedFields) != wantCount {
		t.Errorf("Rule exported 필드 수 = %d, want %d; 필드: %v", len(exportedFields), wantCount, exportedFields)
	}

	// 필드명 순서 및 일치 검증
	wantFields := []string{"ID", "Zone", "File", "Anchor", "Clause", "CanaryGate"}
	if !reflect.DeepEqual(exportedFields, wantFields) {
		t.Errorf("Rule exported 필드 = %v, want %v", exportedFields, wantFields)
	}
}

// TestRuleStructYAMLTags는 Rule 구조체의 yaml 태그가 registry 스키마와 일치함을 검증한다.
// AC-CON-001-017 관련.
func TestRuleStructYAMLTags(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeOf(constitution.Rule{})

	wantTags := map[string]string{
		"ID":         "id",
		"Zone":       "zone",
		"File":       "file",
		"Anchor":     "anchor",
		"Clause":     "clause",
		"CanaryGate": "canary_gate",
	}

	for fieldName, wantTag := range wantTags {
		f, ok := rt.FieldByName(fieldName)
		if !ok {
			t.Errorf("필드 %q를 찾을 수 없다", fieldName)
			continue
		}
		gotTag := f.Tag.Get("yaml")
		if gotTag != wantTag {
			t.Errorf("Rule.%s yaml 태그 = %q, want %q", fieldName, gotTag, wantTag)
		}
	}
}

// TestRuleValidateValidRule은 유효한 Rule의 Validate()가 nil을 반환함을 검증한다.
func TestRuleValidateValidRule(t *testing.T) {
	t.Parallel()

	r := constitution.Rule{
		ID:         "CONST-V3R2-001",
		Zone:       constitution.ZoneFrozen,
		File:       ".claude/rules/moai/workflow/spec-workflow.md",
		Anchor:     "#phase-overview",
		Clause:     "SPEC+EARS format",
		CanaryGate: true,
	}

	if err := r.Validate(); err != nil {
		t.Errorf("유효한 Rule.Validate() = %v, nil이어야 한다", err)
	}
}

// TestRuleValidateEmptyID는 빈 ID의 Validate()가 오류를 반환함을 검증한다.
func TestRuleValidateEmptyID(t *testing.T) {
	t.Parallel()

	r := constitution.Rule{
		ID:     "",
		Zone:   constitution.ZoneFrozen,
		File:   ".claude/rules/moai/core/moai-constitution.md",
		Anchor: "#quality-gates",
		Clause: "TRUST 5",
	}

	if err := r.Validate(); err == nil {
		t.Error("빈 ID Rule.Validate()가 오류를 반환해야 한다")
	}
}

// TestRuleValidateInvalidIDFormat은 잘못된 ID 형식의 Validate()가 오류를 반환함을 검증한다.
func TestRuleValidateInvalidIDFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		id   string
	}{
		{"no-prefix", "001"},
		{"wrong-prefix", "RULE-001"},
		{"no-padding", "CONST-V3R2-1"},
		{"extra-chars", "CONST-V3R2-001-extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := constitution.Rule{
				ID:     tt.id,
				Zone:   constitution.ZoneFrozen,
				File:   ".claude/rules/moai/core/moai-constitution.md",
				Anchor: "#quality-gates",
				Clause: "TRUST 5",
			}
			if err := r.Validate(); err == nil {
				t.Errorf("잘못된 ID %q Rule.Validate()가 오류를 반환해야 한다", tt.id)
			}
		})
	}
}

// TestRuleValidateEmptyClause는 빈 Clause의 Validate()가 오류를 반환함을 검증한다.
func TestRuleValidateEmptyClause(t *testing.T) {
	t.Parallel()

	r := constitution.Rule{
		ID:     "CONST-V3R2-001",
		Zone:   constitution.ZoneFrozen,
		File:   ".claude/rules/moai/workflow/spec-workflow.md",
		Anchor: "#phase-overview",
		Clause: "",
	}

	if err := r.Validate(); err == nil {
		t.Error("빈 Clause Rule.Validate()가 오류를 반환해야 한다")
	}
}

// TestRuleValidateEmptyFile은 빈 File의 Validate()가 오류를 반환함을 검증한다.
func TestRuleValidateEmptyFile(t *testing.T) {
	t.Parallel()

	r := constitution.Rule{
		ID:     "CONST-V3R2-001",
		Zone:   constitution.ZoneFrozen,
		File:   "",
		Anchor: "#phase-overview",
		Clause: "SPEC+EARS format",
	}

	if err := r.Validate(); err == nil {
		t.Error("빈 File Rule.Validate()가 오류를 반환해야 한다")
	}
}

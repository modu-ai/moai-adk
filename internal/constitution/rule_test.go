package constitution_test

import (
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestRuleStructFieldsMatchRegistrySchema verifies that the Rule struct has exactly 7 exported fields.
// Direct mapping to AC-CON-001-004. 7 = 6 original fields + ZoneClass (SPEC-V3R5-CONSTITUTION-DUAL-001).
func TestRuleStructFieldsMatchRegistrySchema(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeOf(constitution.Rule{})

	// Count exported fields only.
	var exportedFields []string
	for i := range rt.NumField() {
		f := rt.Field(i)
		if f.IsExported() {
			exportedFields = append(exportedFields, f.Name)
		}
	}

	const wantCount = 7
	if len(exportedFields) != wantCount {
		t.Errorf("Rule exported 필드 수 = %d, want %d; 필드: %v", len(exportedFields), wantCount, exportedFields)
	}

	// Verify field-name order and consistency.
	wantFields := []string{"ID", "Zone", "File", "Anchor", "Clause", "CanaryGate", "ZoneClass"}
	if !reflect.DeepEqual(exportedFields, wantFields) {
		t.Errorf("Rule exported 필드 = %v, want %v", exportedFields, wantFields)
	}
}

// TestRuleStructYAMLTags verifies that the Rule struct's yaml tags match the registry schema.
// Related to AC-CON-001-017.
func TestRuleStructYAMLTags(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeOf(constitution.Rule{})

	wantTags := map[string]string{
		"ID":         "id",
		"Zone":       "zone",
		"ZoneClass":  "zone_class,omitempty",
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

// TestRuleValidateValidRule verifies that Validate() returns nil for a valid Rule.
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

// TestRuleValidateEmptyID verifies that Validate() returns an error for an empty ID.
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

// TestRuleValidateInvalidIDFormat verifies that Validate() returns an error for an invalid ID format.
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

// TestRuleValidateV3R5ID verifies that CONST-V3R5-NNN format IDs are valid.
// Parallel namespace support from SPEC-V3R5-CONSTITUTION-DUAL-001.
func TestRuleValidateV3R5ID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"v3r5-valid", "CONST-V3R5-001", false},
		{"v3r5-valid-large", "CONST-V3R5-039", false},
		{"v3r2-valid", "CONST-V3R2-150", false},
		{"v3r3-invalid", "CONST-V3R3-001", true},
		{"v3r4-invalid", "CONST-V3R4-001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := constitution.Rule{
				ID:     tt.id,
				Zone:   constitution.ZoneFrozen,
				File:   ".claude/rules/moai/core/moai-constitution.md",
				Anchor: "#quality-gates",
				Clause: "some clause",
			}
			err := r.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("ID %q: Validate() = nil, want error", tt.id)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ID %q: Validate() = %v, want nil", tt.id, err)
			}
		})
	}
}

// TestRuleValidateV3R6ID verifies that CONST-V3R6-NNN format IDs are valid.
// V3R6 namespace support from SPEC-V3R6-RULES-CONST-RULEID-001 (ruleIDPattern broadened [25] -> [256]).
func TestRuleValidateV3R6ID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"v3r6-valid", "CONST-V3R6-001", false},
		{"v3r6-valid-large", "CONST-V3R6-999", false},
		// Regression guards (REQ-RCR-002 additive-only): these MUST stay green.
		{"v3r2-regression", "CONST-V3R2-150", false},
		{"v3r5-regression", "CONST-V3R5-039", false},
		{"v3r3-still-invalid", "CONST-V3R3-001", true},
		{"v3r4-still-invalid", "CONST-V3R4-001", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := constitution.Rule{
				ID:     tt.id,
				Zone:   constitution.ZoneFrozen,
				File:   ".claude/rules/moai/workflow/runtime-recovery-doctrine.md",
				Anchor: "#4-anti-death-spiral-hook-carve-out-documentation-only-policy",
				Clause: "Recovery-Signal Carve-Out",
			}
			err := r.Validate()
			if tt.wantErr && err == nil {
				t.Errorf("ID %q: Validate() = nil, want error", tt.id)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("ID %q: Validate() = %v, want nil", tt.id, err)
			}
		})
	}
}

// TestRuleValidateEmptyClause verifies that Validate() returns an error for an empty Clause.
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

// TestRuleValidateEmptyFile verifies that Validate() returns an error for an empty File.
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

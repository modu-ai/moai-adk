package constitution_test

import (
	"reflect"
	"testing"

	"github.com/modu-ai/moai-adk/internal/constitution"
)

// TestRuleStructFieldsMatchRegistrySchema verifies that the Rule struct has exactly 6 exported fields.
// Direct mapping to AC-CON-001-004.
func TestRuleStructFieldsMatchRegistrySchema(t *testing.T) {
	t.Parallel()

	rt := reflect.TypeOf(constitution.Rule{})

	// Count exported fields only
	var exportedFields []string
	for i := range rt.NumField() {
		f := rt.Field(i)
		if f.IsExported() {
			exportedFields = append(exportedFields, f.Name)
		}
	}

	const wantCount = 6
	if len(exportedFields) != wantCount {
		t.Errorf("Rule exported field count = %d, want %d; fields: %v", len(exportedFields), wantCount, exportedFields)
	}

	// Verify field names and order
	wantFields := []string{"ID", "Zone", "File", "Anchor", "Clause", "CanaryGate"}
	if !reflect.DeepEqual(exportedFields, wantFields) {
		t.Errorf("Rule exported fields = %v, want %v", exportedFields, wantFields)
	}
}

// TestRuleStructYAMLTags verifies that the yaml tags on the Rule struct match the registry schema.
// Related to AC-CON-001-017.
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
			t.Errorf("field %q not found", fieldName)
			continue
		}
		gotTag := f.Tag.Get("yaml")
		if gotTag != wantTag {
			t.Errorf("Rule.%s yaml tag = %q, want %q", fieldName, gotTag, wantTag)
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
		t.Errorf("valid Rule.Validate() = %v, must be nil", err)
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
		t.Error("Rule.Validate() with empty ID must return an error")
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
				t.Errorf("Rule.Validate() with invalid ID %q must return an error", tt.id)
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
		t.Error("Rule.Validate() with empty Clause must return an error")
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
		t.Error("Rule.Validate() with empty File must return an error")
	}
}

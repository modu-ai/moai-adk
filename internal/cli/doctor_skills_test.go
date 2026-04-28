// SPEC-V3R3-HARNESS-001 / T-M3-01
// RED phase: classifySkill and runSkillsCheck do not exist yet — tests fail.
// GREEN phase: implement doctor_skills.go → tests pass.

package cli

import (
	"testing"
)

// TestClassifySkill verifies the classification logic for skill names.
// Table-driven, covering all 4 classification branches from REQ-HARNESS-003.
func TestClassifySkill(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		skillName string
		wantClass string
	}{
		{
			name:      "valid static core skill returns PASS",
			skillName: "moai-foundation-cc",
			wantClass: "PASS",
		},
		{
			name:      "unknown moai- prefixed skill returns WARN",
			skillName: "moai-custom-foo",
			wantClass: "WARN",
		},
		{
			name:      "user customization my-harness prefix returns INFO",
			skillName: "my-harness-test",
			wantClass: "INFO",
		},
		{
			name:      "empty string returns INFO (non-moai, no enforcement)",
			skillName: "",
			wantClass: "INFO",
		},
		{
			name:      "moai- prefix only (no name part) returns WARN",
			skillName: "moai-",
			wantClass: "WARN",
		},
		{
			name:      "valid static core skill moai-meta-harness returns PASS",
			skillName: "moai-meta-harness",
			wantClass: "PASS",
		},
		{
			name:      "valid static core skill moai-design-system returns PASS",
			skillName: "moai-design-system",
			wantClass: "PASS",
		},
		{
			name:      "third-party skill without moai- prefix returns INFO",
			skillName: "my-custom-thing",
			wantClass: "INFO",
		},
	}

	for _, tt := range tests {
		tt := tt // capture loop variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := classifySkill(tt.skillName)
			if got != tt.wantClass {
				t.Errorf("classifySkill(%q) = %q, want %q", tt.skillName, got, tt.wantClass)
			}
		})
	}
}

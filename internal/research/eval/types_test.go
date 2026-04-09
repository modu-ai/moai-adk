package eval

import "testing"

// TestCriterionWeight_Constants CriterionWeight 상수 값이 올바른지 검증한다.
func TestCriterionWeight_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  CriterionWeight
		want string
	}{
		{"MustPass 값", MustPass, "must_pass"},
		{"NiceToHave 값", NiceToHave, "nice_to_have"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("CriterionWeight = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestTargetSpec_Fields TargetSpec 필드가 올바르게 초기화되는지 검증한다.
func TestTargetSpec_Fields(t *testing.T) {
	t.Parallel()

	ts := TargetSpec{Path: "path/to/skill.md", Type: "skill"}
	if ts.Path != "path/to/skill.md" {
		t.Errorf("Path = %q, want %q", ts.Path, "path/to/skill.md")
	}
	if ts.Type != "skill" {
		t.Errorf("Type = %q, want %q", ts.Type, "skill")
	}
}

// TestEvalSettings_Defaults EvalSettings 제로 값이 올바른지 검증한다.
func TestEvalSettings_Defaults(t *testing.T) {
	t.Parallel()

	var s EvalSettings
	if s.RunsPerExperiment != 0 {
		t.Errorf("RunsPerExperiment zero value = %d, want 0", s.RunsPerExperiment)
	}
	if s.PassThreshold != 0.0 {
		t.Errorf("PassThreshold zero value = %f, want 0.0", s.PassThreshold)
	}
}

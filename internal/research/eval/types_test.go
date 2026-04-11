package eval

import "testing"

// TestCriterionWeight_Constants verifies that CriterionWeight constant values are correct.
func TestCriterionWeight_Constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  CriterionWeight
		want string
	}{
		{"MustPass value", MustPass, "must_pass"},
		{"NiceToHave value", NiceToHave, "nice_to_have"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.got) != tt.want {
				t.Errorf("CriterionWeight = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

// TestTargetSpec_Fields verifies that TargetSpec fields are initialized correctly.
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

// TestEvalSettings_Defaults verifies that EvalSettings zero values are correct.
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

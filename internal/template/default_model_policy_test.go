package template

// SPEC-CLI-WIZARD-RESTRUCTURE-001 M2 (C9) — the new-project default model
// policy moves High -> Medium (REQ-WIZ-009/011). ModelPolicyHigh remains a
// defined, valid tier; only the DEFAULT seed changes.

import "testing"

func TestDefaultModelPolicyIsMedium(t *testing.T) {
	if DefaultModelPolicy != ModelPolicyMedium {
		t.Errorf("DefaultModelPolicy = %q, want %q", DefaultModelPolicy, ModelPolicyMedium)
	}
}

func TestModelPolicyHighSurvivesAsOption(t *testing.T) {
	if !IsValidModelPolicy(string(ModelPolicyHigh)) {
		t.Error("ModelPolicyHigh must remain a valid model policy after the default moved to Medium")
	}
	valid := ValidModelPolicies()
	if len(valid) != 3 || valid[0] != "high" || valid[1] != "medium" || valid[2] != "low" {
		t.Errorf("ValidModelPolicies() = %v, want [high medium low]", valid)
	}
}

package v4manifest

import "testing"

// Specialist model and effort are optional (SPEC-AGENT-MODEL-INHERIT-001
// REQ-AMI-006, AC-AMI-006, design D11): a specialist that names neither
// inherits the main session's, and an existing manifest that still names a
// valid pair keeps parsing. A present value is still checked against the
// closed set.
func TestValidate_SpecialistModelAndEffortAreOptional(t *testing.T) {
	neither := validManifest()
	neither.Specialists[0].Model = ""
	neither.Specialists[0].Effort = ""
	if err := Validate(neither); err != nil {
		t.Errorf("a specialist without model/effort must validate, got: %v", err)
	}

	both := validManifest()
	both.Specialists[0].Model = ModelOpus
	both.Specialists[0].Effort = EffortHigh
	if err := Validate(both); err != nil {
		t.Errorf("an existing manifest with model/effort must still validate, got: %v", err)
	}

	bad := validManifest()
	bad.Specialists[0].Effort = "ultra"
	if err := Validate(bad); err == nil {
		t.Error("a declared out-of-set effort must still be rejected")
	}
}

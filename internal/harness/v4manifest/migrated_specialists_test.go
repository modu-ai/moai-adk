// Package v4manifest — migrated-specialists regression test (AC-HV4-013a).
//
// The 4 moai-adk maintainer-local specialists (cli-template, quality, workflow,
// hook-ci) were migrated to v4 manifest format at M6 (SPEC-V3R6-HARNESS-V4-001
// REQ-HV4-013). Each specialist's agent file carries an in-body "v4 Manifest
// Entry" section declaring the 5 manifest fields (role / primitive / isolation /
// effort / model).
//
// This test mechanically verifies those declared fields produce a schema-valid
// v4manifest.Specialist (and a schema-valid Manifest when assembled). It is the
// regression net: if a field is reverted or set to an invalid enum value, the
// test fails. The declared values are reproduced here verbatim from the
// in-body sections of:
//   - .claude/agents/harness/cli-template-specialist.md
//   - .claude/agents/harness/quality-specialist.md
//   - .claude/agents/harness/workflow-specialist.md
//   - .claude/agents/harness/hook-ci-specialist.md
//
// @MX:ANCHOR: [AUTO] migratedSpecialists is the v4 mapping SSOT for the 4 Layer B specialists.
// @MX:REASON: [AUTO] fan_in >= 3 candidate: Validate, Runner dispatch, future lifecycle tooling.
package v4manifest

import "testing"

// migratedSpecialists reproduces the 4 specialists' v4 manifest entries
// declared in their in-body "v4 Manifest Entry" sections (AC-HV4-013a).
// If an in-body section is edited, update this table to match verbatim.
var migratedSpecialists = []Specialist{
	{
		Role:      "cli-template-specialist",
		Primitive: PrimitiveSubAgent,
		Isolation: IsolationNone,
		Effort:    EffortHigh,
		Model:     ModelInherit,
	},
	{
		Role:      "quality-specialist",
		Primitive: PrimitiveSubAgent,
		Isolation: IsolationNone,
		Effort:    EffortHigh,
		Model:     ModelInherit,
	},
	{
		Role:      "workflow-specialist",
		Primitive: PrimitiveSubAgent,
		Isolation: IsolationNone,
		Effort:    EffortHigh,
		Model:     ModelInherit,
	},
	{
		Role:      "hook-ci-specialist",
		Primitive: PrimitiveSubAgent,
		Isolation: IsolationNone,
		Effort:    EffortHigh,
		Model:     ModelInherit,
	},
}

// TestMigratedSpecialists_AllFourHaveValidManifestEntries verifies each of the
// 4 migrated specialists declares all 5 v4 manifest sub-fields with valid enum
// values (AC-HV4-013a — "4 specialists migrated to v4 manifest format").
func TestMigratedSpecialists_AllFourHaveValidManifestEntries(t *testing.T) {
	if len(migratedSpecialists) != 4 {
		t.Fatalf("expected exactly 4 migrated specialists, got %d", len(migratedSpecialists))
	}
	expectedRoles := map[string]bool{
		"cli-template-specialist": false,
		"quality-specialist":      false,
		"workflow-specialist":     false,
		"hook-ci-specialist":      false,
	}
	for i, s := range migratedSpecialists {
		if s.Role == "" {
			t.Errorf("migratedSpecialists[%d].role is empty", i)
			continue
		}
		if _, ok := expectedRoles[s.Role]; !ok {
			t.Errorf("migratedSpecialists[%d].role %q is not one of the 4 expected specialist roles", i, s.Role)
			continue
		}
		expectedRoles[s.Role] = true
		if !IsValidPrimitive(s.Primitive) {
			t.Errorf("migratedSpecialists[%d] (%s).primitive %q is not one of the 5 primitives", i, s.Role, s.Primitive)
		}
		if !validIsolations[s.Isolation] {
			t.Errorf("migratedSpecialists[%d] (%s).isolation %q is not none|worktree", i, s.Role, s.Isolation)
		}
		if !validEfforts[s.Effort] {
			t.Errorf("migratedSpecialists[%d] (%s).effort %q is not low|medium|high|xhigh|max", i, s.Role, s.Effort)
		}
		if !validModels[s.Model] {
			t.Errorf("migratedSpecialists[%d] (%s).model %q is not inherit|haiku|sonnet|opus", i, s.Role, s.Model)
		}
	}
	for role, seen := range expectedRoles {
		if !seen {
			t.Errorf("expected specialist role %q was not declared in migratedSpecialists", role)
		}
	}
}

// TestMigratedSpecialists_AssembledManifestIsValid verifies the 4 migrated
// specialists, assembled into a complete Manifest with the other 7 top-level
// fields populated, pass Validate (AC-HV4-013a + AC-HV4-006a regression net).
func TestMigratedSpecialists_AssembledManifestIsValid(t *testing.T) {
	m := Manifest{
		Name:          "moai-adk-dev",
		Domain:        "moai-adk-go maintainer-local development",
		SourceRequest: "migrate the 4 Layer B specialists to v4 manifest format",
		Patterns:      []string{PatternPipeline, PatternExpertPool},
		Specialists:   migratedSpecialists,
		SprintContract: SprintContract{
			Dimensions: []string{"correctness", "template-neutrality"},
			Thresholds: map[string]interface{}{"correctness": 0.85, "template-neutrality": 1.0},
		},
		EntryCommand:   "/harness:moai-adk-dev",
		RunnerWorkflow: "harness-moai-adk-dev-run.js",
	}
	if err := Validate(m); err != nil {
		t.Fatalf("Validate(migratedSpecialists manifest) failed: %v", err)
	}
}

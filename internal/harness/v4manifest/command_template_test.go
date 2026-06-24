package v4manifest

import (
	"strings"
	"testing"
)

// validTestManifest is a minimal valid Manifest used across command-template
// tests. It satisfies Validate() so the tests can exercise the generator on a
// well-formed input.
func validTestManifest() Manifest {
	return Manifest{
		Name:          "dev",
		Domain:        "moai-adk CLI template development",
		SourceRequest: "build a harness for moai-adk CLI template development",
		Patterns:      []string{PatternPipeline},
		Specialists: []Specialist{
			{
				Role:      "template-neutrality-auditor",
				Primitive: PrimitiveSubAgent,
				Isolation: IsolationNone,
				Effort:    EffortHigh,
				Model:     ModelInherit,
			},
		},
		SprintContract: SprintContract{
			Dimensions: []string{"correctness"},
			Thresholds: map[string]interface{}{"correctness": 0.9},
		},
		EntryCommand:   "/harness:dev",
		RunnerWorkflow: "harness-dev-run.js",
	}
}

// TestGenerateCommand_ReferencesRunnerWorkflow verifies the generated thin
// wrapper command dispatches to the harness's Runner Workflow
// (harness-<name>-run.js) — AC-HV4-002b.
func TestGenerateCommand_ReferencesRunnerWorkflow(t *testing.T) {
	m := validTestManifest()
	out, err := GenerateCommand(m)
	if err != nil {
		t.Fatalf("GenerateCommand returned error: %v", err)
	}
	// The wrapper MUST reference the Runner Workflow filename verbatim.
	if !strings.Contains(out, m.RunnerWorkflow) {
		t.Fatalf("generated command does not reference Runner Workflow %q:\n%s", m.RunnerWorkflow, out)
	}
}

// TestGenerateCommand_ContainsHarnessName verifies the generated command
// carries the harness name so /harness:<name> dispatch is self-describing —
// AC-HV4-002a (the command is auto-generated for harness <name>).
func TestGenerateCommand_ContainsHarnessName(t *testing.T) {
	m := validTestManifest()
	out, err := GenerateCommand(m)
	if err != nil {
		t.Fatalf("GenerateCommand returned error: %v", err)
	}
	if !strings.Contains(out, m.Name) {
		t.Fatalf("generated command does not carry harness name %q:\n%s", m.Name, out)
	}
}

// TestGenerateCommand_TemplateNeutrality verifies the command template carries
// no internal-state markers (SPEC IDs, REQ/AC tokens, commit SHAs) — C-HV4-005.
func TestGenerateCommand_TemplateNeutrality(t *testing.T) {
	m := validTestManifest()
	out, err := GenerateCommand(m)
	if err != nil {
		t.Fatalf("GenerateCommand returned error: %v", err)
	}
	forbidden := []string{
		"SPEC-V3R6-HARNESS-V4-001",
		"REQ-HV4-",
		"AC-HV4-",
	}
	for _, f := range forbidden {
		if strings.Contains(out, f) {
			t.Fatalf("generated command contains forbidden marker %q (C-HV4-005 violation):\n%s", f, out)
		}
	}
}

// TestGenerateCommand_RejectsInvalidManifest verifies the generator validates
// its input — an invalid manifest produces an error, not a malformed wrapper.
func TestGenerateCommand_RejectsInvalidManifest(t *testing.T) {
	m := validTestManifest()
	m.Name = "" // invalid — required field empty
	_, err := GenerateCommand(m)
	if err == nil {
		t.Fatal("GenerateCommand accepted an invalid manifest (empty name); expected validation error")
	}
}

// TestGenerateCommand_NameSubstitutedPerHarness verifies the generator
// substitutes the per-harness name into the template (two different harnesses
// produce two different wrapper contents).
func TestGenerateCommand_NameSubstitutedPerHarness(t *testing.T) {
	a := validTestManifest() // name=dev, runner=harness-dev-run.js
	b := validTestManifest()
	b.Name = "research"
	b.RunnerWorkflow = "harness-research-run.js"
	b.EntryCommand = "/harness:research"

	outA, err := GenerateCommand(a)
	if err != nil {
		t.Fatalf("GenerateCommand(a) error: %v", err)
	}
	outB, err := GenerateCommand(b)
	if err != nil {
		t.Fatalf("GenerateCommand(b) error: %v", err)
	}
	if strings.Contains(outA, "harness-research-run.js") {
		t.Fatal("wrapper for harness 'dev' references harness 'research' Runner — name substitution broken")
	}
	if !strings.Contains(outB, "harness-research-run.js") {
		t.Fatal("wrapper for harness 'research' does not reference its own Runner — name substitution broken")
	}
}

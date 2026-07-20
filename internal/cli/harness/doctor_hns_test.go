// SPEC-HNS-PREFIX-RENAME-001 M2 — RED-phase specification tests for the
// doctor dual-pattern specialist regex (REQ-HPR-012).
//
// runnerSpecialistRE MUST match specialist references of BOTH generations
// (hns-<x>-specialist and harness-<x>-specialist). Runner resolution needs NO
// change: Doctor resolves the Runner from the manifest runner_workflow value
// via a prefix-agnostic path join, so a hns-<name>-run.js manifest value
// resolves as-is — the test below re-verifies that ground-truth anchor.
//
// @MX:NOTE: [AUTO] AC-HPR-007 specification — dual-pattern specialist regex +
// manifest-driven (prefix-agnostic) Runner resolution.
// @MX:SPEC: SPEC-HNS-PREFIX-RENAME-001 acceptance.md AC-HPR-007
package harness

import (
	"path/filepath"
	"testing"
)

// TestRunnerSpecialistRE_HNSDualPattern pins the regex against both prefix
// generations and against non-matching shapes.
func TestRunnerSpecialistRE_HNSDualPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"hns specialist matched", `agent("hns-x-specialist")`, []string{"hns-x-specialist"}},
		{"legacy specialist matched", `agent("harness-x-specialist")`, []string{"harness-x-specialist"}},
		{
			"both generations in one Runner",
			`agent("hns-acme-core-specialist"); agent("harness-acme-auditor-specialist")`,
			[]string{"hns-acme-core-specialist", "harness-acme-auditor-specialist"},
		},
		{"non-specialist token not matched", `const x = "hns-acme-run.js"`, nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := runnerSpecialistRE.FindAllString(tt.src, -1)
			if len(got) != len(tt.want) {
				t.Fatalf("FindAllString(%q) = %v, want %v", tt.src, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("match[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestDoctor_HNSHarness_Passes builds a fully-hns-prefixed harness fixture
// (manifest runner_workflow: hns-x-run.js + hns- specialist) and asserts the
// smoke gate reports ZERO errors — proving (a) Runner resolution is
// manifest-driven and prefix-agnostic (REQ-HPR-012 ground truth) and (b) the
// dual-pattern specialist regex resolves the hns- reference against disk.
func TestDoctor_HNSHarness_Passes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	const name = "hnsgood"
	manifestRel := ".claude/commands/harness/" + name + "/manifest.json"
	agentRef := "hns-" + name + "-core-specialist"

	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name+".md"),
		"# /harness:"+name+"\nThin wrapper.\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name, "manifest.json"),
		hnsManifestJSON(name))
	writeFile(t, filepath.Join(root, ".claude", "workflows", "hns-"+name+"-run.js"),
		doctorRunnerJS(manifestRel, agentRef))
	writeFile(t, filepath.Join(root, ".claude", "agents", "harness", agentRef+".md"),
		"# "+agentRef+"\nSpecialist.\n")

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	if report.ErrorCount != 0 {
		t.Errorf("ErrorCount = %d, want 0; findings=%+v", report.ErrorCount, report.Findings)
	}
}

// TestDoctor_HNSDanglingSpecialist_Detected asserts the dual-pattern regex is
// load-bearing: a Runner referencing a NON-existent hns- specialist MUST
// produce an agent-axis ERROR (before the dual-pattern extension the hns-
// reference is invisible to the regex and the defect passes silently).
func TestDoctor_HNSDanglingSpecialist_Detected(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	const name = "hnsbad"
	manifestRel := ".claude/commands/harness/" + name + "/manifest.json"

	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name+".md"), "# thin\n")
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", name, "manifest.json"),
		hnsManifestJSON(name))
	// Runner references an hns- specialist that does NOT exist on disk.
	writeFile(t, filepath.Join(root, ".claude", "workflows", "hns-"+name+"-run.js"),
		doctorRunnerJS(manifestRel, "hns-"+name+"-missing-specialist"))

	report, err := Doctor(root)
	if err != nil {
		t.Fatalf("Doctor: %v", err)
	}
	foundAgentError := false
	for _, f := range report.Findings {
		if f.Axis == "agent" && f.Severity == SeverityError {
			foundAgentError = true
		}
	}
	if !foundAgentError {
		t.Errorf("expected agent-axis ERROR for dangling hns- specialist; findings=%+v", report.Findings)
	}
}

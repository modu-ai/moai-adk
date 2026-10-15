package cli

// t1619_doctor_harness_coverage_test.go — card t1619, bundle B2: the unexecuted changed
// lines of doctor_harness.go (workflowRootHasRequiredFiles and the agent skill-reference
// check of checkLayer6AgentActivation). Family: TestDoctor.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDoctorHarnessWorkflowRootNeedsEveryRequiredFileWithImport — a workflow root is
// selected only when all required files exist and each carries the harness import.
func TestDoctorHarnessWorkflowRootNeedsEveryRequiredFileWithImport(t *testing.T) {
	dir := t.TempDir()
	for _, f := range workflowRequiredFiles {
		if err := os.WriteFile(filepath.Join(dir, f), []byte("@.moai/harness/ import\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if !workflowRootHasRequiredFiles(dir) {
		t.Fatal("a root holding every required file with the harness import must be selected")
	}
	if err := os.Remove(filepath.Join(dir, "sync.md")); err != nil {
		t.Fatal(err)
	}
	if workflowRootHasRequiredFiles(dir) {
		t.Fatal("a root missing one required file must not be selected")
	}
}

// TestDoctorHarnessAgentSkillReferenceMustResolve — an agent naming an hns-* skill is
// judged by whether that skill directory exists in the project or the user tree.
func TestDoctorHarnessAgentSkillReferenceMustResolve(t *testing.T) {
	root := t.TempDir()
	agentsDir := filepath.Join(root, "agents")
	skillsDir := filepath.Join(root, "skills")
	userSkillsDir := filepath.Join(root, "user-skills")
	for _, d := range []string{agentsDir, filepath.Join(skillsDir, "hns-present"), userSkillsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeAgent := func(skills string) {
		body := "---\ndescription: fixture agent\nskills: [" + skills + "]\n---\nbody\n"
		if err := os.WriteFile(filepath.Join(agentsDir, "fixture.md"), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	writeAgent("hns-present")
	if status, detail := checkLayer6AgentActivation(agentsDir, skillsDir, userSkillsDir); status != "PASS" {
		t.Fatalf("a resolvable skill reference must pass, got %s (%s)", status, detail)
	}

	writeAgent("hns-present, hns-absent")
	status, detail := checkLayer6AgentActivation(agentsDir, skillsDir, userSkillsDir)
	if status != "FAIL" || !strings.Contains(detail, "dangling skills: reference hns-absent") {
		t.Fatalf("an unresolvable skill reference must fail naming it, got %s (%s)", status, detail)
	}
}

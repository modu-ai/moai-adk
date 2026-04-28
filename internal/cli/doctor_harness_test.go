// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-03 tests
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// fullySetupHarnessLayout creates a complete fixture with all 5 Layers active.
func fullySetupHarnessLayout(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	// L5: .moai/harness/ baseline files
	harnessDir := filepath.Join(root, ".moai", "harness")
	files := []string{"main.md", "plan-extension.md", "run-extension.md", "sync-extension.md",
		"chaining-rules.yaml", "interview-results.md", "README.md"}
	for _, f := range files {
		writeFile(t, filepath.Join(harnessDir, f), "# "+f+"\n")
	}

	// L1: a my-harness-* skill with triggers
	skillBody := `---
name: my-harness-ios-patterns
description: foo
triggers:
  paths: ["**/*.swift"]
  keywords: ["ios"]
  agents: ["manager-tdd"]
  phases: ["plan", "run"]
---

# Body
`
	writeFile(t, filepath.Join(root, ".claude", "skills", "my-harness-ios-patterns", "SKILL.md"), skillBody)

	// L2: workflow.yaml with harness section
	writeFile(t, filepath.Join(root, ".moai", "config", "sections", "workflow.yaml"),
		"workflow:\n  team:\n    enabled: true\n  harness:\n    enabled: true\n")

	// L3: CLAUDE.md with marker block
	claudeMd := `# Project

## Project-Specific Configuration (Harness-Generated)
<!-- moai:harness-start id="S-1" generated="2026-04-27T00:00:00Z" -->
**Domain**: ios-mobile
<!-- moai:harness-end -->
`
	writeFile(t, filepath.Join(root, "CLAUDE.md"), claudeMd)

	// L4: 4 workflow files with @.moai/harness/ import
	wfDir := filepath.Join(root, ".claude", "skills", "moai", "workflows")
	for _, f := range []string{"plan.md", "run.md", "sync.md", "design.md"} {
		writeFile(t, filepath.Join(wfDir, f), "# "+f+"\n@.moai/harness/"+strings.TrimSuffix(f, ".md")+"-extension.md\n")
	}

	return root
}

func TestRunHarnessCheck_AllPass(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	check := runHarnessCheck(root)
	if check.Status != CheckOK {
		t.Errorf("status = %v, want OK; msg=%s; detail=%s", check.Status, check.Message, check.Detail)
	}
	for _, layer := range []string{"L1:PASS", "L2:PASS", "L3:PASS", "L4:PASS", "L5:PASS"} {
		if !strings.Contains(check.Message, layer) {
			t.Errorf("missing %s in message %q", layer, check.Message)
		}
	}
}

func TestRunHarnessCheck_NoHarnessDir(t *testing.T) {
	root := t.TempDir()
	check := runHarnessCheck(root)
	if check.Status != CheckOK {
		t.Errorf("expected OK when harness not configured, got %v", check.Status)
	}
}

func TestRunHarnessCheck_L5Missing(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Remove a required L5 file
	_ = os.Remove(filepath.Join(root, ".moai", "harness", "main.md"))
	check := runHarnessCheck(root)
	if check.Status != CheckFail {
		t.Errorf("expected FAIL, got %v", check.Status)
	}
	if !strings.Contains(check.Message, "L5:FAIL") {
		t.Errorf("L5 should be FAIL: %s", check.Message)
	}
}

func TestRunHarnessCheck_L3MarkerUnpaired(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Inject a duplicate marker
	claudeMd, _ := os.ReadFile(filepath.Join(root, "CLAUDE.md"))
	doubled := string(claudeMd) + "\n<!-- moai:harness-start id=\"X\" -->\n<!-- moai:harness-end -->\n"
	writeFile(t, filepath.Join(root, "CLAUDE.md"), doubled)
	check := runHarnessCheck(root)
	if check.Status != CheckFail {
		t.Errorf("expected FAIL for duplicate marker")
	}
}

func TestRunHarnessCheck_L4MissingImport(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Overwrite plan.md without import line
	writeFile(t, filepath.Join(root, ".claude", "skills", "moai", "workflows", "plan.md"), "# plan no import\n")
	check := runHarnessCheck(root)
	if check.Status != CheckFail {
		t.Errorf("expected FAIL when L4 import missing")
	}
}

func TestRunHarnessCheck_PrefixConflictAddsWarn(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Create a conflict
	writeFile(t, filepath.Join(root, ".claude", "skills", "moai-foundation-core", "SKILL.md"), "x")
	writeFile(t, filepath.Join(root, ".claude", "skills", "my-harness-foundation-core", "SKILL.md"), `---
name: my-harness-foundation-core
triggers:
  paths: []
  keywords: []
  agents: []
  phases: []
---
`)
	check := runHarnessCheck(root)
	if check.Status != CheckWarn {
		t.Errorf("expected WARN with prefix conflict, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Detail, "my-harness-foundation-core") {
		t.Errorf("conflict detail missing: %s", check.Detail)
	}
}

func TestCheckLayer1Triggers_NoSkillsDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing")
	status, _ := checkLayer1Triggers(dir)
	if status != "PASS" {
		t.Errorf("expected PASS when skills dir missing, got %s", status)
	}
}

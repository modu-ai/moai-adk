// SPEC-V3R3-PROJECT-HARNESS-001 / T-P4-03 tests
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/uikit"
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

	// L1: a harness-* skill with triggers
	skillBody := `---
name: harness-ios-patterns
description: foo
triggers:
  paths: ["**/*.swift"]
  keywords: ["ios"]
  agents: ["manager-tdd"]
  phases: ["plan", "run"]
---

# Body
`
	writeFile(t, filepath.Join(root, ".claude", "skills", "harness-ios-patterns", "SKILL.md"), skillBody)

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
	if check.Status != uikit.CheckOK {
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
	if check.Status != uikit.CheckOK {
		t.Errorf("expected OK when harness not configured, got %v", check.Status)
	}
}

func TestRunHarnessCheck_L5Missing(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Remove a required L5 file
	_ = os.Remove(filepath.Join(root, ".moai", "harness", "main.md"))
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckFail {
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
	if check.Status != uikit.CheckFail {
		t.Errorf("expected FAIL for duplicate marker")
	}
}

func TestRunHarnessCheck_L4MissingImport(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Overwrite plan.md without import line
	writeFile(t, filepath.Join(root, ".claude", "skills", "moai", "workflows", "plan.md"), "# plan no import\n")
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckFail {
		t.Errorf("expected FAIL when L4 import missing")
	}
}

func TestRunHarnessCheck_PrefixConflictAddsWarn(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Create a conflict
	writeFile(t, filepath.Join(root, ".claude", "skills", "moai-foundation-core", "SKILL.md"), "x")
	writeFile(t, filepath.Join(root, ".claude", "skills", "harness-foundation-core", "SKILL.md"), `---
name: harness-foundation-core
triggers:
  paths: []
  keywords: []
  agents: []
  phases: []
---
`)
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckWarn {
		t.Errorf("expected WARN with prefix conflict, got %v: %s", check.Status, check.Message)
	}
	if !strings.Contains(check.Detail, "harness-foundation-core") {
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

// --- SPEC-V3R6-DOCTOR-FALSE-SIGNAL-001 Defect A (#1087): telemetry-presence
// must NOT be mistaken for a configured harness. ---

// writeDFSTelemetry writes a runtime learning-telemetry line to
// .moai/harness/usage-log.jsonl (the artifact whose mere presence triggered the
// #1087 false FAIL). Reuses the shared writeUsageLog helper.
func writeDFSTelemetry(t *testing.T, root string) {
	t.Helper()
	writeUsageLog(t, root, []string{`{"event":"tool_failure","tool":"Bash","ts":"2026-07-16T00:00:00Z"}`})
}

// TestRunHarnessCheck_TelemetryOnlyReportsOK is AC-DFS-001 (RED before fix): a
// project that has NEVER configured a harness but whose learning hook recorded a
// tool_failure observation into .moai/harness/usage-log.jsonl must still report
// CheckOK "no harness configured" — NOT a red FAIL with L5 missing.
func TestRunHarnessCheck_TelemetryOnlyReportsOK(t *testing.T) {
	root := t.TempDir()
	writeDFSTelemetry(t, root) // creates .moai/harness/ with ONLY telemetry

	check := runHarnessCheck(root)

	if check.Status != uikit.CheckOK {
		t.Errorf("telemetry-only harness dir must be OK (no harness configured), got %v: msg=%s detail=%s",
			check.Status, check.Message, check.Detail)
	}
	if !strings.Contains(check.Message, "no harness configured") {
		t.Errorf("message should indicate no harness configured, got %q", check.Message)
	}
	if strings.Contains(check.Message, "L5") || strings.Contains(check.Detail, "L5 missing") {
		t.Errorf("telemetry-only dir must NOT run the L5 battery: msg=%q detail=%q", check.Message, check.Detail)
	}
}

// TestRunHarnessCheck_TelemetryDirsOnlyReportsOK verifies the learning
// subsystem's other runtime artifacts (learning-history/, proposals/) alongside
// usage-log.jsonl still count as "not configured" (REQ-DFS-002).
func TestRunHarnessCheck_TelemetryDirsOnlyReportsOK(t *testing.T) {
	root := t.TempDir()
	writeDFSTelemetry(t, root)
	writeFile(t, filepath.Join(root, ".moai", "harness", "learning-history", "archive", "old.jsonl"), "{}\n")
	writeFile(t, filepath.Join(root, ".moai", "harness", "proposals", "p1.md"), "# proposal\n")

	check := runHarnessCheck(root)
	if check.Status != uikit.CheckOK {
		t.Errorf("runtime-telemetry-only harness dir must be OK, got %v: %s", check.Status, check.Detail)
	}
}

// TestRunHarnessCheck_FullBaselineWithTelemetryRunsBattery is AC-DFS-002
// (preserve): a genuinely-configured harness (all 7 baseline files) with a
// usage-log.jsonl ALSO present must still run the full L1-L6 battery — the
// telemetry-exclusion fix must not suppress a real harness.
func TestRunHarnessCheck_FullBaselineWithTelemetryRunsBattery(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	writeDFSTelemetry(t, root) // telemetry present alongside a genuine harness

	check := runHarnessCheck(root)

	// Battery ran → message carries the per-layer L1: … L6: status string, and
	// it is NOT short-circuited to "no harness configured".
	if strings.Contains(check.Message, "no harness configured") {
		t.Errorf("genuine harness + telemetry must NOT short-circuit to 'not configured': %q", check.Message)
	}
	for _, layer := range []string{"L1:", "L5:", "L6:"} {
		if !strings.Contains(check.Message, layer) {
			t.Errorf("battery did not run — missing %s in %q", layer, check.Message)
		}
	}
	if check.Status != uikit.CheckOK {
		t.Errorf("full baseline + telemetry should be OK, got %v: %s", check.Status, check.Detail)
	}
}

// TestRunHarnessCheck_PartialBaselineWithTelemetryStillFails is AC-DFS-003
// (guard): a partially-configured harness (some but not all 7 baseline files)
// with telemetry present is a genuinely misconfigured harness — it MUST reach
// CheckFail with an L5 missing detail, NOT be masked as "not configured".
func TestRunHarnessCheck_PartialBaselineWithTelemetryStillFails(t *testing.T) {
	root := t.TempDir()
	// Only main.md (a baseline file) present → configured, but L5 incomplete.
	writeFile(t, filepath.Join(root, ".moai", "harness", "main.md"), "# main\n")
	writeDFSTelemetry(t, root)

	check := runHarnessCheck(root)

	if check.Status != uikit.CheckFail {
		t.Errorf("partial baseline must still FAIL (not masked), got %v: %s", check.Status, check.Detail)
	}
	if !strings.Contains(check.Message, "L5:FAIL") {
		t.Errorf("L5 should FAIL for missing baseline files: %s", check.Message)
	}
}

// TestRunHarnessCheck_NonTelemetryContentStillEvaluated is the D2 plan-audit
// guard: a .moai/harness/ directory containing a NON-baseline, NON-telemetry
// artifact (e.g. a stray custom file) alongside full L2-L4 markers but zero L5
// baseline files must be treated as "configured" (its content is not telemetry)
// and still reach CheckFail — proving the telemetry-exclusion predicate does not
// introduce a NEW false-negative by masking a marker-configured harness.
func TestRunHarnessCheck_NonTelemetryContentStillEvaluated(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// Remove all L5 baseline files, leaving the dir with a stray non-telemetry
	// file → dir is non-empty of real content but L5 is incomplete.
	harnessDir := filepath.Join(root, ".moai", "harness")
	for _, f := range []string{"main.md", "plan-extension.md", "run-extension.md",
		"sync-extension.md", "chaining-rules.yaml", "interview-results.md", "README.md"} {
		_ = os.Remove(filepath.Join(harnessDir, f))
	}
	writeFile(t, filepath.Join(harnessDir, "custom-note.md"), "# stray custom harness content\n")
	writeDFSTelemetry(t, root)

	check := runHarnessCheck(root)

	if check.Status == uikit.CheckOK {
		t.Errorf("marker-configured harness with non-telemetry content must NOT be masked as OK: %s / %s",
			check.Message, check.Detail)
	}
	if !strings.Contains(check.Message, "L5:FAIL") {
		t.Errorf("expected L5:FAIL for incomplete baseline: %s", check.Message)
	}
}

// SPEC-V3R6-HARNESS-ACTIVATION-WIRING-001 — Phase-6 smoke gate (REQ-HAW-010..014
// + 013b). The smoke gate extends runHarnessCheck with three agent-frontmatter
// checks atop L1-L5; the existing L1-L5 cases above (TestRunHarnessCheck_*) must
// remain green (AC-HAW-014).

// writeHarnessAgent creates a generated .claude/agents/harness/<name>.md fixture
// with the given frontmatter body. The body is written verbatim between `---`
// fences.
func writeHarnessAgent(t *testing.T, root, name, frontmatter string) {
	t.Helper()
	body := "---\n" + frontmatter + "\n---\n\n# " + name + "\n"
	writeFile(t, filepath.Join(root, ".claude", "agents", "harness", name+".md"), body)
}

// writeHarnessSkill creates a companion .claude/skills/<name>/SKILL.md so a
// generated agent's skills: reference resolves (non-dangling). The SKILL.md
// carries a valid triggers: section so it also satisfies the existing L1 check.
func writeHarnessSkill(t *testing.T, root, name string) {
	t.Helper()
	body := "---\nname: " + name + "\ndescription: domain skill\ntriggers:\n" +
		"  paths: [\"**/*.swift\"]\n  keywords: [\"ios\"]\n  agents: [\"manager-tdd\"]\n  phases: [\"run\"]\n---\n\n# Body\n"
	writeFile(t, filepath.Join(root, ".claude", "skills", name, "SKILL.md"), body)
}

// TestRunHarnessCheck_GoodAgentPasses verifies a generated agent with a
// non-empty description AND a resolvable skills: preload does NOT trip the
// smoke gate (the all-pass fixture remains OK).
func TestRunHarnessCheck_GoodAgentPasses(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	writeHarnessSkill(t, root, "harness-ios-patterns")
	writeHarnessAgent(t, root, "ios-architect",
		"name: ios-architect\ndescription: iOS 도메인 아키텍처 설계 시 활성\nskills:\n  - harness-ios-patterns")
	check := runHarnessCheck(root)
	if check.Status == uikit.CheckFail {
		t.Errorf("good agent should not FAIL: msg=%s detail=%s", check.Message, check.Detail)
	}
}

// TestRunHarnessCheck_EmptyAgentDescription verifies REQ-HAW-012 / AC-HAW-012:
// a generated agent with an empty description trips the smoke gate.
func TestRunHarnessCheck_EmptyAgentDescription(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	writeHarnessSkill(t, root, "harness-ios-patterns")
	writeHarnessAgent(t, root, "ios-architect",
		"name: ios-architect\ndescription:\nskills:\n  - harness-ios-patterns")
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckFail {
		t.Errorf("expected FAIL for empty description, got %v (%s)", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "description") {
		t.Errorf("detail should name the description problem: %s", check.Detail)
	}
	if !strings.Contains(check.Detail, "ios-architect") {
		t.Errorf("detail should name the offending agent: %s", check.Detail)
	}
}

// TestRunHarnessCheck_MissingSkillsKey verifies REQ-HAW-013b / AC-HAW-015: a
// generated agent that OMITS the skills: key entirely trips the smoke gate.
func TestRunHarnessCheck_MissingSkillsKey(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	writeHarnessAgent(t, root, "ios-engineer",
		"name: ios-engineer\ndescription: iOS 구현 시 활성")
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckFail {
		t.Errorf("expected FAIL for missing skills: key, got %v (%s)", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "skills") {
		t.Errorf("detail should mention the missing skills key: %s", check.Detail)
	}
	if !strings.Contains(check.Detail, "ios-engineer") {
		t.Errorf("detail should name the offending agent: %s", check.Detail)
	}
}

// TestRunHarnessCheck_DanglingSkillReference verifies REQ-HAW-013 / AC-HAW-013:
// a generated agent whose skills: entry points at a non-existent harness-*
// directory trips the smoke gate.
func TestRunHarnessCheck_DanglingSkillReference(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	// References harness-nonexistent which has NO skill dir on disk
	// (the fixture only creates harness-ios-patterns) → dangling.
	writeHarnessAgent(t, root, "ios-architect",
		"name: ios-architect\ndescription: iOS 설계 시 활성\nskills:\n  - harness-nonexistent")
	check := runHarnessCheck(root)
	if check.Status != uikit.CheckFail {
		t.Errorf("expected FAIL for dangling skill ref, got %v (%s)", check.Status, check.Detail)
	}
	if !strings.Contains(check.Detail, "harness-nonexistent") {
		t.Errorf("detail should name the dangling skill: %s", check.Detail)
	}
}

// TestRunHarnessCheck_TemplateSkillNotDangling verifies EC-4: a generated agent
// that references a template-distributed moai-* skill (not harness-*) is NOT
// treated as dangling even if the dir is absent — only harness-* references
// are resolved against disk.
func TestRunHarnessCheck_TemplateSkillNotDangling(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	writeHarnessAgent(t, root, "ios-architect",
		"name: ios-architect\ndescription: iOS 설계 시 활성\nskills:\n  - moai-domain-frontend")
	check := runHarnessCheck(root)
	if check.Status == uikit.CheckFail {
		t.Errorf("moai-* skill reference must NOT be treated as dangling (EC-4): %s", check.Detail)
	}
}

// TestRunHarnessCheck_NoGeneratedAgents verifies the smoke gate is a no-op when
// no .claude/agents/harness/ agents exist (the agent-frontmatter checks only
// apply when generated agents are present).
func TestRunHarnessCheck_NoGeneratedAgents(t *testing.T) {
	root := fullySetupHarnessLayout(t)
	check := runHarnessCheck(root)
	if check.Status == uikit.CheckFail {
		t.Errorf("no generated agents → agent-frontmatter checks must not FAIL: %s", check.Detail)
	}
}

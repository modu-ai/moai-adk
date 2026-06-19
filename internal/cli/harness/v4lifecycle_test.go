package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile is a test helper that creates a file with the given content,
// creating parent directories as needed.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

// validManifestJSON is a minimal valid harness-v4 manifest.json used as a test
// fixture. It satisfies v4manifest.Validate() (8 fields + 1 specialist +
// sprint_contract populated).
const validManifestJSON = `{
  "name": "dev",
  "domain": "moai-adk CLI template development",
  "source_request": "build a harness for moai-adk CLI template development",
  "patterns": ["Pipeline"],
  "specialists": [
    {
      "role": "template-neutrality-auditor",
      "primitive": "sub-agent",
      "isolation": "none",
      "effort": "high",
      "model": "inherit"
    }
  ],
  "sprint_contract": {
    "dimensions": ["correctness"],
    "thresholds": {"correctness": 0.9}
  },
  "entry_command": "/harness:dev",
  "runner_workflow": "harness-dev-run.js"
}`

// seedHarness builds a complete harness "dev" fixture under projectRoot so that
// List / Remove can operate on a fully-populated harness (command + workflow +
// specialists + skills + manifest). Returns the paths created.
func seedHarness(t *testing.T, projectRoot, name, manifestJSON string) {
	t.Helper()
	commandsDir := filepath.Join(projectRoot, ".claude", "commands", "harness")
	workflowsDir := filepath.Join(projectRoot, ".claude", "workflows")
	agentsDir := filepath.Join(projectRoot, ".claude", "agents", "harness")
	skillsDir := filepath.Join(projectRoot, ".claude", "skills", "harness-"+name)

	// command file (thin wrapper)
	writeFile(t, filepath.Join(commandsDir, name+".md"),
		"---\ndescription: harness "+name+"\n---\nRun harness "+name+"\n")

	// manifest.json co-located with the command (subdirectory layout per
	// design §B.2 — manifest lives at .claude/commands/harness/<name>/manifest.json)
	writeFile(t, filepath.Join(commandsDir, name, "manifest.json"), manifestJSON)

	// Runner workflow
	writeFile(t, filepath.Join(workflowsDir, "harness-"+name+"-run.js"),
		"// Runner for "+name+"\n")

	// specialist agent definition
	writeFile(t, filepath.Join(agentsDir, "harness-"+name+"-auditor-specialist.md"),
		"---\nname: harness-"+name+"-auditor-specialist\n---\nspecialist\n")

	// companion skill
	writeFile(t, filepath.Join(skillsDir, "SKILL.md"),
		"# harness-"+name+" skill\n")
}

// TestListHarnesses_EnumeratesAllHarnesses verifies /moai:harness list
// enumerates every harness present (AC-HV4-011a — two harnesses both shown).
func TestListHarnesses_EnumeratesAllHarnesses(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", strings.ReplaceAll(validManifestJSON, `"dev"`, `"dev"`))
	// second harness "research"
	researchManifest := strings.ReplaceAll(validManifestJSON, "dev", "research")
	researchManifest = strings.ReplaceAll(researchManifest, "harness-dev-run.js", "harness-research-run.js")
	researchManifest = strings.ReplaceAll(researchManifest, "/harness:dev", "/harness:research")
	seedHarness(t, root, "research", researchManifest)

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 harnesses, got %d: %+v", len(entries), entries)
	}
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name] = true
	}
	if !names["dev"] || !names["research"] {
		t.Fatalf("expected dev+research, got %v", names)
	}
}

// TestListHarnesses_JoinsManifestDomain verifies the list output joins each
// command with its manifest.json (domain + entry_command surfaced).
func TestListHarnesses_JoinsManifestDomain(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", validManifestJSON)

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 harness, got %d", len(entries))
	}
	e := entries[0]
	if e.Name != "dev" {
		t.Errorf("Name = %q, want dev", e.Name)
	}
	if e.Domain == "" {
		t.Error("Domain empty — manifest not joined")
	}
	if e.EntryCommand != "/harness:dev" {
		t.Errorf("EntryCommand = %q, want /harness:dev", e.EntryCommand)
	}
}

// TestEditHarness_LocatesManifestAndSpecialists verifies EditHarness surfaces
// the manifest (SSOT) + specialist + skill paths for the named harness.
func TestEditHarness_LocatesManifestAndSpecialists(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", validManifestJSON)

	paths, err := EditHarness(root, "dev")
	if err != nil {
		t.Fatalf("EditHarness error: %v", err)
	}
	if paths.Name != "dev" {
		t.Errorf("Name = %q, want dev", paths.Name)
	}
	if paths.ManifestPath == "" {
		t.Error("ManifestPath empty")
	}
	// manifest.json MUST exist at the surfaced path (SSOT).
	if _, err := os.Stat(paths.ManifestPath); err != nil {
		t.Errorf("surfaced ManifestPath does not exist: %s (%v)", paths.ManifestPath, err)
	}
	// seedHarness creates one specialist + one skill — both should be located.
	if len(paths.SpecialistPaths) != 1 {
		t.Errorf("expected 1 specialist path, got %d: %v", len(paths.SpecialistPaths), paths.SpecialistPaths)
	}
	if len(paths.SkillPaths) != 1 {
		t.Errorf("expected 1 skill path, got %d: %v", len(paths.SkillPaths), paths.SkillPaths)
	}
}

// TestEditHarness_MissingManifestReturnsError verifies EditHarness errors
// when the manifest (SSOT) is missing — the user must not be directed to edit
// a harness whose SSOT is gone.
func TestEditHarness_MissingManifestReturnsError(t *testing.T) {
	root := t.TempDir()
	// command file only, no manifest
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "orphan.md"),
		"---\ndescription: orphan\n---\nRun orphan\n")

	_, err := EditHarness(root, "orphan")
	if err == nil {
		t.Fatal("EditHarness succeeded with missing manifest; expected error")
	}
}

// TestListHarnesses_CommandWithoutManifestStillListed verifies a command file
// whose manifest.json is missing is still enumerated (with a manifest-missing
// flag) — list must not crash on partial state; remove handles the atomicity.
func TestListHarnesses_CommandWithoutManifestStillListed(t *testing.T) {
	root := t.TempDir()
	// create command file only, no manifest
	writeFile(t, filepath.Join(root, ".claude", "commands", "harness", "orphan.md"),
		"---\ndescription: orphan\n---\nRun orphan\n")

	entries, err := ListHarnesses(root)
	if err != nil {
		t.Fatalf("ListHarnesses error on orphan command: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry (orphan), got %d", len(entries))
	}
	if entries[0].ManifestMissing != true {
		t.Error("expected ManifestMissing=true for command without manifest")
	}
}

// TestRemoveHarness_RemovesAllFiveArtifactTypes verifies remove is atomic and
// removes command + workflow + specialists + skills + manifest (AC-HV4-011b).
func TestRemoveHarness_RemovesAllFiveArtifactTypes(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", validManifestJSON)

	cmdPath := filepath.Join(root, ".claude", "commands", "harness", "dev.md")
	manifestPath := filepath.Join(root, ".claude", "commands", "harness", "dev", "manifest.json")
	runnerPath := filepath.Join(root, ".claude", "workflows", "harness-dev-run.js")
	agentPath := filepath.Join(root, ".claude", "agents", "harness", "harness-dev-auditor-specialist.md")
	skillPath := filepath.Join(root, ".claude", "skills", "harness-dev", "SKILL.md")

	if err := RemoveHarness(root, "dev"); err != nil {
		t.Fatalf("RemoveHarness error: %v", err)
	}

	for _, p := range []string{cmdPath, manifestPath, runnerPath, agentPath, skillPath} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("artifact still exists after remove: %s (err=%v)", p, err)
		}
	}
}

// TestRemoveHarness_FailClosedOnMissingManifest verifies remove FAILS CLOSED
// when the manifest is missing (AC-HV4-011c — orphan prevention). The command
// file must remain unchanged (no partial state).
func TestRemoveHarness_FailClosedOnMissingManifest(t *testing.T) {
	root := t.TempDir()
	// Seed a complete harness, then delete ONLY the manifest to simulate
	// partial state.
	seedHarness(t, root, "dev", validManifestJSON)
	cmdPath := filepath.Join(root, ".claude", "commands", "harness", "dev.md")
	manifestPath := filepath.Join(root, ".claude", "commands", "harness", "dev", "manifest.json")
	if err := os.Remove(manifestPath); err != nil {
		t.Fatalf("setup: remove manifest: %v", err)
	}

	cmdContentBefore, _ := os.ReadFile(cmdPath)

	err := RemoveHarness(root, "dev")
	if err == nil {
		t.Fatal("RemoveHarness succeeded with missing manifest; expected fail-closed error")
	}
	// Error MUST name the missing artifact.
	if !strings.Contains(strings.ToLower(err.Error()), "manifest") {
		t.Errorf("error does not name the missing manifest artifact: %v", err)
	}

	// Command file MUST be unchanged (no partial removal).
	cmdContentAfter, _ := os.ReadFile(cmdPath)
	if string(cmdContentBefore) != string(cmdContentAfter) {
		t.Error("command file changed despite fail-closed — partial state left behind")
	}
}

// TestRemoveHarness_NonexistentHarnessReturnsError verifies remove on a
// harness that does not exist at all returns an error (fail-closed on the
// precheck), not a silent no-op.
func TestRemoveHarness_NonexistentHarnessReturnsError(t *testing.T) {
	root := t.TempDir()
	err := RemoveHarness(root, "ghost")
	if err == nil {
		t.Fatal("RemoveHarness on nonexistent harness returned nil; expected error")
	}
}

// TestRemoveHarness_PrecheckFailsOnMissingRunner verifies the atomicity
// precheck catches a missing Runner workflow too (not just manifest) — the
// command file stays intact.
func TestRemoveHarness_PrecheckFailsOnMissingRunner(t *testing.T) {
	root := t.TempDir()
	seedHarness(t, root, "dev", validManifestJSON)
	runnerPath := filepath.Join(root, ".claude", "workflows", "harness-dev-run.js")
	if err := os.Remove(runnerPath); err != nil {
		t.Fatalf("setup: remove runner: %v", err)
	}
	cmdPath := filepath.Join(root, ".claude", "commands", "harness", "dev.md")
	before, _ := os.ReadFile(cmdPath)

	err := RemoveHarness(root, "dev")
	if err == nil {
		t.Fatal("RemoveHarness succeeded with missing Runner; expected fail-closed error")
	}
	after, _ := os.ReadFile(cmdPath)
	if string(before) != string(after) {
		t.Error("command file changed despite fail-closed on missing Runner")
	}
}

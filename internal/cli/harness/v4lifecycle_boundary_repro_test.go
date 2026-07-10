package harness

// SPEC-CLIFIX-CRITICAL-001 M1 — Reproduction test (RED phase).
//
// Defect d — bare "harness-"+name prefix match deletes a sibling harness.
// Removing "release" also collects "release-update" artifacts because
// strings.HasPrefix("harness-release-update-...", "harness-release") is true.
// AC-CRIT-001-004

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/harness/v4manifest"
)

func TestRemoveHarnessBoundary_Repro(t *testing.T) {
	tmp := t.TempDir()

	// --- set up the "release" harness (the one being removed) ---
	cmdDir := filepath.Join(tmp, v4CommandsDir)
	if err := os.MkdirAll(filepath.Join(cmdDir, "release"), 0o755); err != nil {
		t.Fatalf("mkdir release cmd: %v", err)
	}
	if err := os.WriteFile(filepath.Join(cmdDir, "release.md"), []byte("# release\n"), 0o644); err != nil {
		t.Fatalf("write release.md: %v", err)
	}
	// release-update is a SIBLING harness — it must have its own command file so
	// listHarnessCommandNames can disambiguate "release" from "release-update".
	if err := os.WriteFile(filepath.Join(cmdDir, "release-update.md"), []byte("# release-update\n"), 0o644); err != nil {
		t.Fatalf("write release-update.md: %v", err)
	}
	m := v4manifest.Manifest{RunnerWorkflow: "harness-release-run.js"}
	mb, _ := json.Marshal(&m)
	if err := os.WriteFile(filepath.Join(cmdDir, "release", "manifest.json"), mb, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	wfDir := filepath.Join(tmp, v4WorkflowsDir)
	if err := os.MkdirAll(wfDir, 0o755); err != nil {
		t.Fatalf("mkdir wf: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wfDir, "harness-release-run.js"), []byte("// run\n"), 0o644); err != nil {
		t.Fatalf("write runner: %v", err)
	}

	// --- specialist files for BOTH release and release-update ---
	agentsDir := filepath.Join(tmp, v4AgentsDir)
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	releaseSpec := filepath.Join(agentsDir, "harness-release-auditor-specialist.md")
	releaseUpdateSpec := filepath.Join(agentsDir, "harness-release-update-auditor-specialist.md")
	if err := os.WriteFile(releaseSpec, []byte("# release auditor\n"), 0o644); err != nil {
		t.Fatalf("write release spec: %v", err)
	}
	if err := os.WriteFile(releaseUpdateSpec, []byte("# release-update auditor\n"), 0o644); err != nil {
		t.Fatalf("write release-update spec: %v", err)
	}

	// --- companion skill dirs for BOTH ---
	skillsDir := filepath.Join(tmp, v4SkillsDir)
	releaseSkill := filepath.Join(skillsDir, "harness-release")
	releaseUpdateSkill := filepath.Join(skillsDir, "harness-release-update")
	if err := os.MkdirAll(releaseSkill, 0o755); err != nil {
		t.Fatalf("mkdir release skill: %v", err)
	}
	if err := os.MkdirAll(releaseUpdateSkill, 0o755); err != nil {
		t.Fatalf("mkdir release-update skill: %v", err)
	}

	if err := RemoveHarness(tmp, "release"); err != nil {
		t.Fatalf("RemoveHarness(release): %v", err)
	}

	// release-update specialist MUST survive.
	if _, err := os.Stat(releaseUpdateSpec); err != nil {
		t.Errorf("release-update specialist was deleted by RemoveHarness(release) (defect d): %v", err)
	}
	// release-update skill dir MUST survive.
	if _, err := os.Stat(releaseUpdateSkill); err != nil {
		t.Errorf("release-update skill dir was deleted by RemoveHarness(release) (defect d): %v", err)
	}
	// release specialist MUST be removed (sanity).
	if _, err := os.Stat(releaseSpec); err == nil {
		t.Errorf("release specialist was NOT removed (test setup issue)")
	}
}

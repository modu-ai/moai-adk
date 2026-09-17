package cli

// SPEC-INIT-HARNESS-001 M4 — update coherence (REQ-IH-010, AC-IH-009 /
// AC-IH-015): `moai update` on a codex-only project re-deploys through the
// codex-only surface set (no .claude/ resurrection) and the llm.harness key
// survives the 3-way config merge.

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
)

// runCodexOnlyProjectThenUpdate runs a codex-only init, then an update
// template sync inside the project directory. HOME stays pinned to the temp
// dir for the whole flow (home-fingerprint discipline, REQ-IQW-014/016).
func runCodexOnlyProjectThenUpdate(t *testing.T) (projectDir string) {
	t.Helper()
	var homeDir string
	projectDir, homeDir = runInitForAutonomy(t, nil, map[string]string{"llm": "codex"})
	_ = homeDir

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}

	resetUpdateFlags(t, "force", "templates-only", "binary", "check", "yes")
	if err := updateCmd.Flags().Set("force", "true"); err != nil {
		t.Fatal(err)
	}
	if err := updateCmd.Flags().Set("yes", "true"); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	updateCmd.SetOut(buf)
	updateCmd.SetErr(buf)
	// Direct invocation skips cobra's Execute wiring — ValidateAll dereferences
	// cmd.Context(), so a nil context panics mid-update without this.
	updateCmd.SetContext(context.Background())

	if err := runTemplateSync(updateCmd); err != nil {
		t.Fatalf("runTemplateSync after codex-only init: %v (stderr: %s)", err, buf.String())
	}
	return projectDir
}

// resetUpdateFlags restores the update command flags a test touches, so the
// global command object does not leak state into sibling tests.
func resetUpdateFlags(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		if f := updateCmd.Flags().Lookup(name); f != nil {
			before := f.DefValue
			if err := updateCmd.Flags().Set(name, before); err != nil {
				t.Fatalf("reset --%s: %v", name, err)
			}
		}
	}
}

// TestUpdateCodexOnlyNoClaudeResurrection verifies AC-IH-009: after an update
// on a codex-only project, no .claude/** path was resurrected and the codex
// surfaces were refreshed.
func TestUpdateCodexOnlyNoClaudeResurrection(t *testing.T) {
	projectDir := runCodexOnlyProjectThenUpdate(t)

	if _, err := os.Stat(filepath.Join(projectDir, ".claude")); err == nil {
		t.Error(".claude/ resurrected by update on a codex-only project (REQ-IH-010 violation)")
	}
	if _, err := os.Stat(filepath.Join(projectDir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md resurrected by update on a codex-only project (REQ-IH-010 violation)")
	}
	// Codex surfaces still present after the re-deploy.
	for _, rel := range []string{"AGENTS.md", ".codex/config.toml", ".moai/config/sections/llm.yaml"} {
		if _, err := os.Stat(filepath.Join(projectDir, rel)); err != nil {
			t.Errorf("%s missing after codex-only update: %v", rel, err)
		}
	}
}

// TestUpdatePreservesHarnessKey verifies AC-IH-015: the llm.harness value init
// wrote survives the update's 3-way config merge.
func TestUpdatePreservesHarnessKey(t *testing.T) {
	projectDir := runCodexOnlyProjectThenUpdate(t)

	if got := readLLMHarness(t, projectDir); got != "codex" {
		t.Errorf("llm.harness after update = %q, want codex (3-way merge survival)", got)
	}
}

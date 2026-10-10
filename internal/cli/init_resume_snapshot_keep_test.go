package cli

// init_resume_snapshot_keep_test.go — card t1591 (SPEC-USERASSET-DEPLOY-GUARD-001),
// codex round-2 finding 3: a resumed init must keep the settings snapshot its
// interrupted first run staged, and promote it once the resume completes.
//
// The first run stages the template render in .claude/settings.json.pending right after
// the deploy, then stops at the user-asset ensure. Its autonomy answer rewrites
// settings.json after that staging point, so the live file no longer equals the staged
// render. The rerun's leftover judgement reads the difference as a revert and discards
// the copy before the resume runs, so the resume's promotion finds nothing to promote
// and the next update has no template baseline.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
)

// settingsWithDefaultMode returns the render with permissions.defaultMode set to mode,
// the key the autonomy tier writes into settings.json.
func settingsWithDefaultMode(t *testing.T, render []byte, mode string) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(render, &doc); err != nil {
		t.Fatalf("decode the staged render: %v", err)
	}
	perms, _ := doc["permissions"].(map[string]any)
	if perms == nil {
		perms = map[string]any{}
	}
	perms["defaultMode"] = mode
	doc["permissions"] = perms
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(out, '\n')
}

// TestInitResumeKeepsSettingsSnapshotForInterruptedInit reproduces finding 3. A settings
// change made after the staging point (the autonomy answer of the interrupted run) must
// not cost the resumed run its template baseline: the resume promotes the staged render,
// and the canonical baseline holds exactly those bytes afterwards. The control (a healthy
// project with no pending snapshot) must keep its baseline and its refusal path unchanged.
func TestInitResumeKeepsSettingsSnapshotForInterruptedInit(t *testing.T) {
	t.Run("resume_promotes_the_staged_render_after_a_settings_change", func(t *testing.T) {
		isolatedInitHome(t)
		projectDir := filepath.Join(t.TempDir(), "snapshot-keep-proj")
		stubFailingFirstEnsure(t)

		if err := initRunWithBundles(t, projectDir, "", ""); err == nil || !strings.Contains(err.Error(), "user-asset install failed") {
			t.Fatalf("run 1 did not fail at the user-asset ensure (the injection did not bite): %v", err)
		}
		render, err := os.ReadFile(backup.SettingsSnapshotPendingPath(projectDir))
		if err != nil {
			t.Fatalf("precondition: the interrupted first run staged no settings snapshot: %v", err)
		}

		// The interrupted run's autonomy answer changes the live file after the staging point.
		changed := settingsWithDefaultMode(t, render, "acceptEdits")
		if bytes.Equal(changed, render) {
			t.Fatal("precondition: the autonomy change did not alter the staged render")
		}
		if err := os.WriteFile(filepath.Join(projectDir, ".claude", "settings.json"), changed, 0o644); err != nil {
			t.Fatal(err)
		}

		if err := initRunWithBundles(t, projectDir, "", ""); err != nil {
			t.Fatalf("run 2 (the resume) failed: %v", err)
		}
		got, ok := backup.LoadSettingsSnapshot(projectDir)
		if !ok {
			t.Fatal("the resumed run left no canonical settings baseline: the staged render was discarded before the resume could promote it, so the next update has no template baseline")
		}
		if !bytes.Equal(got, render) {
			t.Errorf("the canonical settings baseline is not the staged template render:\n got: %s\nwant: %s", got, render)
		}
	})

	t.Run("control_healthy_project_without_pending_is_unchanged", func(t *testing.T) {
		isolatedInitHome(t)
		projectDir := filepath.Join(t.TempDir(), "snapshot-healthy-proj")

		if err := initRunWithBundles(t, projectDir, "", ""); err != nil {
			t.Fatalf("the healthy init failed: %v", err)
		}
		if _, err := os.Stat(backup.SettingsSnapshotPendingPath(projectDir)); err == nil {
			t.Fatal("precondition: a healthy init left a pending settings snapshot")
		}
		before, beforeOK := backup.LoadSettingsSnapshot(projectDir)

		err := initRunWithBundles(t, projectDir, "", "")
		if err == nil || !strings.Contains(err.Error(), "already initialized") {
			t.Fatalf("a healthy rerun was not refused with the update redirect: %v", err)
		}
		after, afterOK := backup.LoadSettingsSnapshot(projectDir)
		if afterOK != beforeOK || !bytes.Equal(after, before) {
			t.Errorf("a healthy rerun changed the settings baseline: before ok=%v, after ok=%v", beforeOK, afterOK)
		}
		if _, statErr := os.Stat(backup.SettingsSnapshotPendingPath(projectDir)); statErr == nil {
			t.Errorf("a healthy rerun created a pending settings snapshot")
		}
	})
}

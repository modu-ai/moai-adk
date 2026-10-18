package cli

// update_prior_manifest_tmpcleanup_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001
// (card t1865, card-review round 4, finding B): writePriorManifest reaches the
// final name only by renaming a complete temporary file onto it. When the rename
// failed it used to return with `manifest.prior.json.tmp-*` still in the run
// backup directory, and the reader ignoring that name was not enough: the legacy
// restore walk copies any file at the backup root into .moai/config, so the
// temporary file landed in the project (the restore side is pinned in
// internal/cli/update/backup/restore_prior_manifest_tmp_test.go).
//
// Expected behaviour, pinned here: after a failed rename the temporary file is
// gone (best effort), the failure is still reported with the same error, the
// final name does not exist, and the run still removes nothing. The temporary
// file's name is spelled out as a literal so the test pins what is really written.

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const priorTmpLiteralPrefix = "manifest.prior.json.tmp-"

func priorTmpLeftovers(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var left []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), priorTmpLiteralPrefix) {
			left = append(left, e.Name())
		}
	}
	return left
}

func TestFailedPriorManifestRenameLeavesNoTemporaryFile(t *testing.T) {
	data := []byte(`{"version":"v0","files":{}}` + "\n")
	injected := errors.New("injected rename failure")
	newRunDir := func(t *testing.T) string {
		t.Helper()
		dir := filepath.Join(t.TempDir(), ".moai-backups", "20260101_000001")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		return dir
	}

	t.Run("TheNameTheWriterUsesIsTheNameTheRestoreSkips", func(t *testing.T) {
		dir := newRunDir(t)
		tmp, err := writePriorManifestTemp(dir, data)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(filepath.Base(tmp), priorTmpLiteralPrefix) {
			t.Errorf("the temporary file is named %q, want the prefix %q the restore walk skips", filepath.Base(tmp), priorTmpLiteralPrefix)
		}
	})

	t.Run("InjectedRenameFailure", func(t *testing.T) {
		dir := newRunDir(t)
		swapPriorWriteHooks(t, nil, func(string, string) error { return injected })
		err := writePriorManifest(dir, data)
		if err == nil || !errors.Is(err, injected) || !strings.Contains(err.Error(), "rename into place") {
			t.Fatalf("the failure is not reported as before: %v", err)
		}
		if left := priorTmpLeftovers(t, dir); len(left) != 0 {
			t.Errorf("temporary files remain after the failed rename: %v", left)
		}
		if _, statErr := os.Lstat(filepath.Join(dir, priorManifestName)); !os.IsNotExist(statErr) {
			t.Errorf("the final name exists after the failed rename (%v)", statErr)
		}
	})

	t.Run("RealRenameFailureOntoAnOccupiedName", func(t *testing.T) {
		dir := newRunDir(t)
		occupant := filepath.Join(dir, priorManifestName, "keep")
		if err := os.MkdirAll(filepath.Dir(occupant), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(occupant, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := writePriorManifest(dir, data); err == nil {
			t.Fatalf("renaming onto a non-empty directory returned nil")
		}
		if left := priorTmpLeftovers(t, dir); len(left) != 0 {
			t.Errorf("temporary files remain after the failed rename: %v", left)
		}
		if got, err := os.ReadFile(occupant); err != nil || string(got) != "x" {
			t.Errorf("the occupant was disturbed: %v", err)
		}
	})

	// The Backup stage's hand-off: the failure is reported, the run directory is
	// clean, and the migration that follows removes nothing. A control run with
	// the same fixture and no injection removes the file, so the empty result is
	// the failure's doing.
	t.Run("TheRunStillRemovesNothing", func(t *testing.T) {
		removedWith := func(t *testing.T, fail bool) (removed []string, dir string, state priorManifestState, err error) {
			t.Helper()
			root, home := ownershipFixtureBase(t)
			runDir := filepath.Join(root, ".moai-backups", "20260101_000002")
			if mkErr := os.MkdirAll(runDir, 0o755); mkErr != nil {
				t.Fatal(mkErr)
			}
			if fail {
				swapPriorWriteHooks(t, nil, func(string, string) error { return injected })
			}
			dir, state, err = preservePriorManifest(root, runDir, data)
			rc := &migrationRunContext{userAssetsInstalled: true, runBackupDir: dir, priorManifest: state, priorManifestErr: err}
			removed, migErr := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}, rc)
			if migErr != nil {
				t.Fatalf("migration: %v", migErr)
			}
			return removed, dir, state, err
		}
		okRemoved, _, okState, okErr := removedWith(t, false)
		if okState != priorManifestWritten || okErr != nil || !slices.Equal(okRemoved, []string{migIdenticalSkill}) {
			t.Fatalf("control: state=%v err=%v removed=%v, want written/nil/[%s]", okState, okErr, okRemoved, migIdenticalSkill)
		}
		removed, dir, state, err := removedWith(t, true)
		if state != priorManifestFailed || err == nil || !errors.Is(err, injected) {
			t.Errorf("state=%v err=%v, want a failed state carrying the injected error", state, err)
		}
		if len(removed) != 0 {
			t.Errorf("the run removed %v although its prior manifest copy failed", removed)
		}
		if left := priorTmpLeftovers(t, dir); len(left) != 0 {
			t.Errorf("temporary files remain in the run backup: %v", left)
		}
	})
}

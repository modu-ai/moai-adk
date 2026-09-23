package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// M3 — Snapshot write-time hook wiring. The trigger sites of Decision D4 as
// amended by card t1139 — init right after its template deploy, and each update
// path right after its deploy — funnel through writeTemplateSnapshotBestEffort;
// runUpdateRestore deploys nothing and writes no snapshot. The helper tests
// below pin the funnel itself; the call-site ordering is pinned end to end in
// update_identity_preserve_test.go and update_identity_clean_install_test.go.

// TestInit_WritesSnapshot is AC-TBS-001 at the helper level. init.go calls
// writeTemplateSnapshotBestEffort from InitOptions.AfterTemplateDeploy, right
// after the template deploy and before any wizard patch; this test exercises
// that helper and asserts the snapshot appears on disk.
func TestInit_WritesSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	// Simulate init having deployed rendered sections.
	if err := os.MkdirAll(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir), defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}

	// The exact call init.go makes from its AfterTemplateDeploy hook.
	var buf bytes.Buffer
	writeTemplateSnapshotBestEffort(projectRoot, &buf)

	snapDir := filepath.Join(projectRoot, defs.MoAIDir, "cache", "template-snapshot", "sections")
	entries, err := os.ReadDir(snapDir)
	if err != nil {
		t.Fatalf("init did not write snapshot: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("init wrote an empty snapshot")
	}
	// Verify the snapshot is byte-equal to the on-disk rendered file.
	got, err := os.ReadFile(filepath.Join(snapDir, "system.yaml"))
	if err != nil {
		t.Fatalf("read snapshot system.yaml: %v", err)
	}
	if string(got) != "version: \"3.0.1\"\n" {
		t.Errorf("snapshot system.yaml = %q, want rendered version", got)
	}
}

// TestUpdateRestore_WritesSnapshot_TemplateSync exercises the helper the
// template-sync deploy site calls. Since card t1139 update_template_sync.go
// calls writeTemplateSnapshotBestEffort right after the Deploy Templates step,
// BEFORE Restore Settings runs RestoreMoaiConfig. This test reaches only the
// helper, never that call site; the call-site ordering is pinned end to end by
// TestUpdateForce_SnapshotIsTheDeployedRender.
func TestUpdateRestore_WritesSnapshot_TemplateSync(t *testing.T) {
	t.Parallel()
	testHelperWritesSnapshot(t)
}

// TestUpdateRestore_WritesSnapshot_CleanInstall exercises the helper the
// clean-reinstall deploy site calls. Since card t1139 update_clean_install.go
// calls writeTemplateSnapshotBestEffort right after the Step 5 deploy, BEFORE
// Step 5.5 runs RestoreMoaiConfig. This test reaches only the helper, never
// that call site; the call-site ordering and the identity render are pinned end
// to end by TestCleanReinstall_SnapshotIsTheDeployedRenderAndIdentitySurvives.
func TestUpdateRestore_WritesSnapshot_CleanInstall(t *testing.T) {
	t.Parallel()
	testHelperWritesSnapshot(t)
}

// TestRunUpdateRestore_LeavesSnapshotUntouched replaces the former
// AC-TBS-002 runUpdateRestore site test (card t1139). The lockout-escape
// restore deploys no template, so it has no render to record; writing the
// restored user config into the snapshot would make it the next merge BASE and
// the next update would drop every customization it carries. The snapshot the
// last deploy wrote must survive the restore byte-for-byte.
func TestRunUpdateRestore_LeavesSnapshotUntouched(t *testing.T) {
	t.Parallel()
	projectRoot, backupDir := newRestoreFixture(t)

	// The last deploy's render: user.name empty, as the template renders it.
	snapPath := filepath.Join(backup.SnapshotDir(projectRoot), "sections", "user.yaml")
	if err := os.MkdirAll(filepath.Dir(snapPath), defs.DirPerm); err != nil {
		t.Fatalf("mkdir snapshot: %v", err)
	}
	const rendered = "user:\n  name: \"\"\n"
	if err := os.WriteFile(snapPath, []byte(rendered), defs.FilePerm); err != nil {
		t.Fatalf("write snapshot: %v", err)
	}

	var buf bytes.Buffer
	if err := runUpdateRestore(projectRoot, backupDir, &buf); err != nil {
		t.Fatalf("runUpdateRestore: %v", err)
	}

	// Positive control: the restore did write the user's value to the tree.
	live, err := os.ReadFile(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "user.yaml"))
	if err != nil {
		t.Fatalf("read restored user.yaml: %v", err)
	}
	if !bytes.Contains(live, []byte("fixture-operator")) {
		t.Fatalf("restore did not write the backed-up user.yaml; got %q", live)
	}
	got, err := os.ReadFile(snapPath)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(got) != rendered {
		t.Errorf("runUpdateRestore rewrote the snapshot: got %q, want the last render %q", got, rendered)
	}
}

// testHelperWritesSnapshot is the shared body for the two update deploy-site
// helper tests: both sites call writeTemplateSnapshotBestEffort, so they share
// the same "the helper copies the section files" assertion.
func testHelperWritesSnapshot(t *testing.T) {
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir), defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "quality.yaml"),
		[]byte("test_coverage_target: 85\n"), defs.FilePerm); err != nil {
		t.Fatalf("write quality.yaml: %v", err)
	}

	var buf bytes.Buffer
	writeTemplateSnapshotBestEffort(projectRoot, &buf)

	snapDir := filepath.Join(projectRoot, defs.MoAIDir, "cache", "template-snapshot", "sections")
	if _, err := os.Stat(filepath.Join(snapDir, "quality.yaml")); err != nil {
		t.Fatalf("snapshot helper did not write the snapshot: %v", err)
	}
}

// TestWriteSnapshot_FailureDoesNotBlockInit is AC-TBS-015: a snapshot write
// failure does not propagate from writeTemplateSnapshotBestEffort (the helper
// init/update call). The helper returns no error and emits only a warning.
func TestWriteSnapshot_FailureDoesNotBlockInit(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir), defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}
	// Block the snapshot dir.
	cachePath := filepath.Join(projectRoot, defs.MoAIDir, "cache")
	if err := os.WriteFile(cachePath, []byte("blocker"), defs.FilePerm); err != nil {
		t.Fatalf("plant blocker: %v", err)
	}

	var buf bytes.Buffer
	// Must NOT panic / must NOT return an error (it returns nothing).
	writeTemplateSnapshotBestEffort(projectRoot, &buf)
	// A warning was emitted to the buffer.
	if buf.Len() == 0 {
		t.Errorf("expected a warning on failure, got empty output")
	}
}

func TestWriteSnapshot_FailureDoesNotBlockUpdate(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir), defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}
	cachePath := filepath.Join(projectRoot, defs.MoAIDir, "cache")
	if err := os.WriteFile(cachePath, []byte("blocker"), defs.FilePerm); err != nil {
		t.Fatalf("plant blocker: %v", err)
	}

	var buf bytes.Buffer
	writeTemplateSnapshotBestEffort(projectRoot, &buf)
	if buf.Len() == 0 {
		t.Errorf("expected a warning on failure, got empty output")
	}
}

// TestFirstUpdate_NoSnapshot_CompletesAndWritesSnapshot is AC-TBS-014: a
// pre-existing install with NO snapshot (first post-feature update) completes
// the snapshot write at the end and produces a snapshot for the next cycle.
func TestFirstUpdate_NoSnapshot_CompletesAndWritesSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	// Pre-existing install: sections present, NO snapshot yet.
	if err := os.MkdirAll(filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir), defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(
		filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write system.yaml: %v", err)
	}

	// Before: no snapshot (fallback path active).
	if backup.HasSnapshot(projectRoot) {
		t.Fatalf("pre-condition: snapshot must be absent for first-update test")
	}

	// The restore-completion site fires the snapshot write.
	var buf bytes.Buffer
	writeTemplateSnapshotBestEffort(projectRoot, &buf)

	// After: snapshot exists.
	if !backup.HasSnapshot(projectRoot) {
		t.Fatalf("first update did not write a snapshot for the next cycle")
	}
}

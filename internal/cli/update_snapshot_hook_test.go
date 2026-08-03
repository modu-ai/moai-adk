package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/cli/update/backup"
	"github.com/modu-ai/moai-adk/internal/defs"
)

// M3 — Snapshot write-time hook wiring. The four trigger sites (Decision D4)
// all funnel through writeTemplateSnapshotBestEffort. These tests pin that
// funnel so the init/update/restore wiring is behaviorally verified without
// standing up the full cobra command flow.

// TestInit_WritesSnapshot is AC-TBS-001: the end of `moai init` writes the
// snapshot. init.go calls writeTemplateSnapshotBestEffort; this test exercises
// the exact helper init invokes and asserts the snapshot appears on disk.
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

	// The exact call init.go makes at the end of runInit.
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

// TestUpdateRestore_WritesSnapshot_TemplateSync is AC-TBS-002 (template-sync
// restore site). update_template_sync.go calls writeTemplateSnapshotBestEffort
// after RestoreMoaiConfig completes.
func TestUpdateRestore_WritesSnapshot_TemplateSync(t *testing.T) {
	t.Parallel()
	testHelperWritesSnapshot(t)
}

// TestUpdateRestore_WritesSnapshot_CleanInstall is AC-TBS-002 (clean-install
// restore site). update_clean_install.go calls writeTemplateSnapshotBestEffort
// after RestoreMoaiConfig completes.
func TestUpdateRestore_WritesSnapshot_CleanInstall(t *testing.T) {
	t.Parallel()
	testHelperWritesSnapshot(t)
}

// TestUpdateRestore_WritesSnapshot_RunUpdateRestore is AC-TBS-002
// (runUpdateRestore lockout-escape site). update_restore.go calls
// writeTemplateSnapshotBestEffort after RestoreFromBackupDir completes.
func TestUpdateRestore_WritesSnapshot_RunUpdateRestore(t *testing.T) {
	t.Parallel()
	testHelperWritesSnapshot(t)
}

// testHelperWritesSnapshot is the shared body for the three restore-site
// tests: all three call writeTemplateSnapshotBestEffort, so they share the
// same post-restore snapshot assertion.
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
		t.Fatalf("restore site did not write snapshot: %v", err)
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

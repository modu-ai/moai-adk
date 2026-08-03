package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// M3 — Snapshot lifecycle. AC-TBS-004 (snapshot survives the clean step) and
// AC-TBS-015 (best-effort non-blocking).

// TestSnapshot_SurvivesCleanStep is AC-TBS-004. The update clean step walks
// and deletes .moai/config/ (research.md §A confirms it never touches
// .moai/cache/). This test simulates the clean step by deleting .moai/config/
// wholesale (the worst-case clean radius) and asserts the snapshot under
// .moai/cache/template-snapshot/ is preserved byte-identical.
func TestSnapshot_SurvivesCleanStep(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()

	// Populate on-disk sections + write the snapshot.
	sections := map[string]string{
		"system.yaml":  "version: \"3.0.1\"\n",
		"quality.yaml": "test_coverage_target: 80\n",
	}
	writeSections(t, projectRoot, sections)
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	// Read the snapshot bytes BEFORE the clean step.
	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	preBytes := map[string][]byte{}
	for name := range sections {
		b, err := os.ReadFile(filepath.Join(snapSections, name))
		if err != nil {
			t.Fatalf("read pre-snapshot %s: %v", name, err)
		}
		preBytes[name] = b
	}

	// Simulate the clean step: delete .moai/config/ wholesale (the clean radius
	// per update_cleanup.go / update_clean_install.go — it deletes .moai/config/
	// but never .moai/cache/, verified empirically by research.md §A).
	configDir := filepath.Join(projectRoot, defs.MoAIDir, defs.ConfigSubdir)
	if err := os.RemoveAll(configDir); err != nil {
		t.Fatalf("remove config dir (simulate clean): %v", err)
	}

	// Assert the snapshot survived.
	for name, pre := range preBytes {
		got, err := os.ReadFile(filepath.Join(snapSections, name))
		if err != nil {
			t.Fatalf("clean step destroyed snapshot %s: %v", name, err)
		}
		if !bytes.Equal(got, pre) {
			t.Errorf("snapshot %s not byte-identical after clean: got %q, want %q", name, got, pre)
		}
	}

	// The cache directory itself must still exist.
	if _, err := os.Stat(filepath.Join(projectRoot, defs.MoAIDir, "cache")); err != nil {
		t.Errorf("clean step deleted .moai/cache/: %v", err)
	}
}

// TestWriteSnapshot_FailureDoesNotBlock is AC-TBS-015 in the unit form:
// WriteSnapshot returns a non-nil error when the snapshot target is
// unwritable, but the error is structured (not a panic) so the caller's
// best-effort wrapper can swallow it. The full init/update non-blocking
// behavior is verified by the cli-package wiring test
// (TestWriteSnapshot_FailureDoesNotBlockInit / _Update).
func TestWriteSnapshot_FailureDoesNotBlock(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})

	// Make .moai/cache/ a read-only file (not a directory) so WriteSnapshot's
	// MkdirAll on .moai/cache/template-snapshot/sections/ fails.
	cachePath := filepath.Join(projectRoot, defs.MoAIDir, "cache")
	if err := os.WriteFile(cachePath, []byte("blocker"), defs.FilePerm); err != nil {
		t.Fatalf("plant blocker file: %v", err)
	}

	err := WriteSnapshot(projectRoot)
	if err == nil {
		t.Fatalf("WriteSnapshot must return a non-nil error when the snapshot dir is unwritable")
	}
	// No panic — the error is returnable for best-effort swallowing.
}

// TestSnapshot_ScopeLimitedToSections is AC-TBS-016: the snapshot writer's
// walk root references defs.SectionsSubdir only (no .claude/, no
// .moai/project/).
func TestSnapshot_ScopeLimitedToSections(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()

	// Plant a file under .moai/project/ (NOT in scope) and under sections (in scope).
	projectDir := filepath.Join(projectRoot, defs.MoAIDir, "project")
	if err := os.MkdirAll(projectDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir project: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "tech.md"), []byte("out of scope"), defs.FilePerm); err != nil {
		t.Fatalf("write tech.md: %v", err)
	}
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})

	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	// The snapshot must contain system.yaml but NOT project/tech.md.
	snapRoot := SnapshotDir(projectRoot)
	if _, err := os.Stat(filepath.Join(snapRoot, "sections", "system.yaml")); err != nil {
		t.Errorf("snapshot missing sections/system.yaml: %v", err)
	}
	if _, err := os.Stat(filepath.Join(snapRoot, "project", "tech.md")); err == nil {
		t.Errorf("snapshot MUST NOT contain .moai/project/ content (scope leak)")
	}
	if _, err := os.Stat(filepath.Join(snapRoot, "sections", "tech.md")); err == nil {
		t.Errorf("snapshot MUST NOT contain project/ files under sections/")
	}
}

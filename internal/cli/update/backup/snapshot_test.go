package backup

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// M1 — Snapshot package skeleton. These tests pin the data-model decision
// (Decision D1): the snapshot lives at .moai/cache/template-snapshot/sections/,
// is a verbatim byte-copy of the on-disk rendered .moai/config/sections/ tree,
// and is best-effort non-blocking (REQ-TBS-014).

// writeSections writes the given name->content map under
// <projectRoot>/.moai/config/sections/ and returns the sections dir.
func writeSections(t *testing.T, projectRoot string, files map[string]string) string {
	t.Helper()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	for name, content := range files {
		path := filepath.Join(sectionsDir, name)
		if err := os.MkdirAll(filepath.Dir(path), defs.DirPerm); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), defs.FilePerm); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return sectionsDir
}

// TestWriteSnapshot_CopiesRenderedSections asserts that WriteSnapshot copies
// every .yaml/.yml file under .moai/config/sections/ byte-for-byte into the
// snapshot at .moai/cache/template-snapshot/sections/ (AC-TBS-001, AC-TBS-003).
func TestWriteSnapshot_CopiesRenderedSections(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	rendered := map[string]string{
		"system.yaml":  "version: \"3.0.1\"\n",
		"quality.yaml": "test_coverage_target: 80\n",
		"nested/foo.yaml": "key: value\n",
	}
	writeSections(t, projectRoot, rendered)

	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	for rel, want := range rendered {
		got, err := os.ReadFile(filepath.Join(snapSections, rel))
		if err != nil {
			t.Fatalf("read snapshot %s: %v", rel, err)
		}
		if string(got) != want {
			t.Errorf("snapshot %s = %q, want %q (verbatim byte copy)", rel, got, want)
		}
	}
}

// TestSnapshot_CarriesRenderedValues_NotPlaceholders is AC-TBS-003: the
// snapshot must carry rendered values ("3.0.1"), not Go-template placeholders.
func TestSnapshot_CarriesRenderedValues_NotPlaceholders(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	rendered := map[string]string{
		"system.yaml": "version: \"3.0.1\"\ngoBinPath: \"/usr/local/go/bin\"\n",
	}
	writeSections(t, projectRoot, rendered)

	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(SnapshotDir(projectRoot), "sections", "system.yaml"))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if !contains(string(got), "3.0.1") {
		t.Errorf("snapshot must carry rendered version 3.0.1, got %q", got)
	}
	if contains(string(got), "{{.Version}}") || contains(string(got), "{{.GoBinPath}}") {
		t.Errorf("snapshot must NOT carry unresolved placeholders, got %q", got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// TestWriteSnapshot_NoConfigDirIsBestEffort asserts that a missing config dir
// returns a non-nil error (total failure), but does NOT panic (REQ-TBS-014).
func TestWriteSnapshot_NoConfigDirIsBestEffort(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir() // no .moai/config/sections/

	err := WriteSnapshot(projectRoot)
	if err == nil {
		t.Fatalf("WriteSnapshot on a project with no config dir must return a non-nil error")
	}
}

// TestHasSnapshot_True/False pins HasSnapshot: true iff the snapshot sections
// dir exists AND is non-empty.
func TestHasSnapshot_True(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})

	if HasSnapshot(projectRoot) {
		t.Fatalf("HasSnapshot must be false before WriteSnapshot runs")
	}
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	if !HasSnapshot(projectRoot) {
		t.Fatalf("HasSnapshot must be true after WriteSnapshot populated sections")
	}
}

func TestHasSnapshot_False(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir() // nothing

	if HasSnapshot(projectRoot) {
		t.Fatalf("HasSnapshot must be false on an empty project")
	}

	// Empty sections dir (created but no files) also reports false.
	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	if err := os.MkdirAll(snapSections, defs.DirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if HasSnapshot(projectRoot) {
		t.Fatalf("HasSnapshot must be false when snapshot sections dir is empty")
	}
}

// TestSnapshotDir pins the location (Decision D1).
func TestSnapshotDir(t *testing.T) {
	t.Parallel()
	got := SnapshotDir("/proj")
	want := filepath.Join("/proj", defs.MoAIDir, "cache", "template-snapshot")
	if got != want {
		t.Errorf("SnapshotDir = %q, want %q", got, want)
	}
}

// TestWriteSnapshot_SwallowsIndividualCopyErrors covers REQ-TBS-014's
// best-effort branches: a single bad section file does NOT abort the snapshot.
// Three injection techniques exercise the mkdir/read/write swallow paths.
func TestWriteSnapshot_SwallowsIndividualCopyErrors(t *testing.T) {
	t.Parallel()

	// Case A — mkdir swallow: a nested source file whose destination parent
	// collides with a pre-planted regular file.
	t.Run("mkdir_collision", func(t *testing.T) {
		t.Parallel()
		projectRoot := t.TempDir()
		writeSections(t, projectRoot, map[string]string{
			"sub/foo.yaml": "key: value\n",
			"system.yaml":  "version: \"3.0.1\"\n",
		})
		// Pre-plant a regular FILE at the nested destination parent, so
		// MkdirAll(snapshotDir/sections/sub) fails inside the walk callback.
		collision := filepath.Join(SnapshotDir(projectRoot), "sections", "sub")
		if err := os.MkdirAll(filepath.Dir(collision), defs.DirPerm); err != nil {
			t.Fatalf("mkdir collision parent: %v", err)
		}
		if err := os.WriteFile(collision, []byte("blocker"), defs.FilePerm); err != nil {
			t.Fatalf("plant collision: %v", err)
		}

		if err := WriteSnapshot(projectRoot); err != nil {
			t.Fatalf("WriteSnapshot must swallow the individual mkdir failure: %v", err)
		}
		// The non-colliding file was still snapshotted.
		got, err := os.ReadFile(filepath.Join(SnapshotDir(projectRoot), "sections", "system.yaml"))
		if err != nil {
			t.Fatalf("expected system.yaml snapshotted despite the sibling collision: %v", err)
		}
		if !contains(string(got), "3.0.1") {
			t.Errorf("system.yaml snapshot = %q, want rendered version", got)
		}
	})

	// Case B — write swallow: destination path is pre-occupied by a directory,
	// so os.WriteFile fails ("is a directory").
	t.Run("write_collision", func(t *testing.T) {
		t.Parallel()
		projectRoot := t.TempDir()
		writeSections(t, projectRoot, map[string]string{
			"system.yaml":  "version: \"3.0.1\"\n",
			"quality.yaml": "test_coverage_target: 80\n",
		})
		// Pre-plant a DIRECTORY at the system.yaml destination path.
		collision := filepath.Join(SnapshotDir(projectRoot), "sections", "system.yaml")
		if err := os.MkdirAll(filepath.Dir(collision), defs.DirPerm); err != nil {
			t.Fatalf("mkdir collision parent: %v", err)
		}
		if err := os.Mkdir(collision, defs.DirPerm); err != nil {
			t.Fatalf("plant collision dir: %v", err)
		}

		if err := WriteSnapshot(projectRoot); err != nil {
			t.Fatalf("WriteSnapshot must swallow the individual write failure: %v", err)
		}
		// The sibling file was still snapshotted.
		if _, err := os.Stat(filepath.Join(SnapshotDir(projectRoot), "sections", "quality.yaml")); err != nil {
			t.Errorf("expected quality.yaml snapshotted despite the sibling write collision: %v", err)
		}
	})

	// Case C — read swallow: the source file is unreadable (chmod 0000). On a
	// non-root run this triggers the os.ReadFile error path.
	t.Run("read_unreadable", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("chmod-based unreadability is ineffective when running as root")
		}
		t.Parallel()
		projectRoot := t.TempDir()
		sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
		if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
			t.Fatalf("mkdir sections: %v", err)
		}
		bad := filepath.Join(sectionsDir, "unreadable.yaml")
		if err := os.WriteFile(bad, []byte("x: 1\n"), defs.FilePerm); err != nil {
			t.Fatalf("write unreadable: %v", err)
		}
		if err := os.Chmod(bad, 0); err != nil {
			t.Fatalf("chmod unreadable: %v", err)
		}
		t.Cleanup(func() { _ = os.Chmod(bad, 0o644) })
		// A readable sibling so the walk does not bail entirely.
		if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"),
			[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
			t.Fatalf("write system.yaml: %v", err)
		}

		if err := WriteSnapshot(projectRoot); err != nil {
			t.Fatalf("WriteSnapshot must swallow the individual read failure: %v", err)
		}
		if _, err := os.Stat(filepath.Join(SnapshotDir(projectRoot), "sections", "system.yaml")); err != nil {
			t.Errorf("expected system.yaml snapshotted despite the sibling read failure: %v", err)
		}
	})
}

// TestWriteSnapshot_TopLevelMkdirFails covers the total-failure return path:
// when the top-level snapshot dir cannot be created (path occupied by a file),
// WriteSnapshot returns a non-nil error.
func TestWriteSnapshot_TopLevelMkdirFails(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})
	// Occupy the cache dir with a regular file so MkdirAll on cache/template-snapshot fails.
	cachePath := filepath.Join(projectRoot, defs.MoAIDir, "cache")
	if err := os.MkdirAll(filepath.Dir(cachePath), defs.DirPerm); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.WriteFile(cachePath, []byte("blocker"), defs.FilePerm); err != nil {
		t.Fatalf("plant blocker: %v", err)
	}
	if err := WriteSnapshot(projectRoot); err == nil {
		t.Fatalf("WriteSnapshot must return a non-nil error when the top-level snapshot dir cannot be created")
	}
}

// TestWriteSnapshot_ConfigSectionsIsFile covers the `!info.IsDir()` branch:
// the config sections path exists but is a regular file, not a directory.
func TestWriteSnapshot_ConfigSectionsIsFile(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	// Plant a regular FILE at the sections path (not a directory).
	sectionsPath := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(filepath.Dir(sectionsPath), defs.DirPerm); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	if err := os.WriteFile(sectionsPath, []byte("not a dir"), defs.FilePerm); err != nil {
		t.Fatalf("plant file: %v", err)
	}
	if err := WriteSnapshot(projectRoot); err == nil {
		t.Fatalf("WriteSnapshot must error when config sections is a file, not a directory")
	}
}

// TestHasSnapshot_ReadDirError covers the ReadDir error branch (the sections
// dir exists but is unreadable).
func TestHasSnapshot_ReadDirError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("chmod-based unreadability is ineffective when running as root")
	}
	t.Parallel()
	projectRoot := t.TempDir()
	// Create a snapshot sections dir with a file, then make it unreadable.
	snapDir := filepath.Join(SnapshotDir(projectRoot), "sections")
	if err := os.MkdirAll(snapDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir snap: %v", err)
	}
	if err := os.WriteFile(filepath.Join(snapDir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Chmod(snapDir, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(snapDir, 0o755) })

	// ReadDir fails → HasSnapshot returns false (degrades to fallback).
	if HasSnapshot(projectRoot) {
		t.Errorf("HasSnapshot must return false when the snapshot dir is unreadable")
	}
}

// TestWriteSnapshot_NonYamlSkipped covers the extension-filter branch: a
// .json or .txt file under sections/ is NOT snapshotted.
func TestWriteSnapshot_NonYamlSkipped(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	sectionsDir := filepath.Join(projectRoot, defs.MoAIDir, defs.SectionsSubdir)
	if err := os.MkdirAll(sectionsDir, defs.DirPerm); err != nil {
		t.Fatalf("mkdir sections: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "system.yaml"),
		[]byte("version: \"3.0.1\"\n"), defs.FilePerm); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sectionsDir, "notes.txt"),
		[]byte("skip me"), defs.FilePerm); err != nil {
		t.Fatalf("write txt: %v", err)
	}

	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	snapDir := filepath.Join(SnapshotDir(projectRoot), "sections")
	if _, err := os.Stat(filepath.Join(snapDir, "system.yaml")); err != nil {
		t.Errorf("yaml file should be snapshotted: %v", err)
	}
	if _, err := os.Stat(filepath.Join(snapDir, "notes.txt")); err == nil {
		t.Errorf("non-yaml file must NOT be snapshotted")
	}
}

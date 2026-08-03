package backup

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// M2 — SaveTemplateBase base-loader. Decision D2: introduce
// SaveTemplateBase(destDir, projectRoot) that prefers the rendered snapshot
// when present and falls back to SaveTemplateDefaults (embedded-raw) when
// absent. BackupMoaiConfig switches to SaveTemplateBase.

// TestSaveTemplateBase_PrefersSnapshot asserts that when a snapshot exists,
// SaveTemplateBase copies the snapshot bytes (rendered, e.g. version "3.0.1")
// into destDir/sections/, NOT the embedded-raw bytes ({{.Version}}).
func TestSaveTemplateBase_PrefersSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()

	// Rendered on-disk sections (what the snapshot copies from).
	writeSections(t, projectRoot, map[string]string{
		"system.yaml": "version: \"3.0.1\"\n",
	})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(destDir, "sections", "system.yaml"))
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	want := []byte("version: \"3.0.1\"\n")
	if !bytes.Equal(got, want) {
		t.Errorf("SaveTemplateBase wrote %q, want snapshot bytes %q", got, want)
	}
}

// TestSaveTemplateBase_FallsBackWhenSnapshotAbsent asserts that when NO
// snapshot exists, SaveTemplateBase delegates to SaveTemplateDefaults and the
// resulting bytes match what SaveTemplateDefaults would produce (AC-TBS-007).
func TestSaveTemplateBase_FallsBackWhenSnapshotAbsent(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir() // no .moai/config/sections, no snapshot

	destBase := t.TempDir()
	destSnapshotAbsent := filepath.Join(destBase, "no-snap")
	destDefaults := filepath.Join(destBase, "defaults")

	if err := SaveTemplateBase(destSnapshotAbsent, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (no snapshot): %v", err)
	}
	if err := SaveTemplateDefaults(destDefaults); err != nil {
		t.Fatalf("SaveTemplateDefaults: %v", err)
	}

	// The fallback path must produce the same sections/ tree as SaveTemplateDefaults.
	defaultsSections := filepath.Join(destDefaults, "sections")
	entries, err := os.ReadDir(defaultsSections)
	if err != nil {
		t.Fatalf("readdir defaults sections: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		want, err := os.ReadFile(filepath.Join(defaultsSections, e.Name()))
		if err != nil {
			t.Fatalf("read default %s: %v", e.Name(), err)
		}
		got, err := os.ReadFile(filepath.Join(destSnapshotAbsent, "sections", e.Name()))
		if err != nil {
			t.Fatalf("read fallback %s: %v", e.Name(), err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("fallback %s = %q, want SaveTemplateDefaults bytes %q", e.Name(), got, want)
		}
	}
}

// TestSaveTemplateBase_BytesMatchSnapshot is AC-TBS-006b: the base bytes that
// SaveTemplateBase writes must byte-equal the snapshot's section bytes.
func TestSaveTemplateBase_BytesMatchSnapshot(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{
		"system.yaml":  "version: \"3.0.1\"\n",
		"quality.yaml": "test_coverage_target: 80\n",
	})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase: %v", err)
	}

	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	for _, name := range []string{"system.yaml", "quality.yaml"} {
		snapBytes, err := os.ReadFile(filepath.Join(snapSections, name))
		if err != nil {
			t.Fatalf("read snapshot %s: %v", name, err)
		}
		destBytes, err := os.ReadFile(filepath.Join(destDir, "sections", name))
		if err != nil {
			t.Fatalf("read dest %s: %v", name, err)
		}
		if !bytes.Equal(destBytes, snapBytes) {
			t.Errorf("dest %s (%q) != snapshot %s (%q)", name, destBytes, name, snapBytes)
		}
	}
}

// TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaultsBytes is the
// explicit AC-TBS-007 name: a project with NO snapshot delegates to the
// embedded-raw path. Verified via a project that HAS a config dir but NO
// snapshot yet (simulating a pre-existing install on first post-feature update).
func TestSaveTemplateBase_FallbackMatchesSaveTemplateDefaultsBytes(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	// Pre-existing install: config sections present, but NO snapshot.
	writeSections(t, projectRoot, map[string]string{
		"system.yaml": "version: \"3.0.1\"\n",
	})
	// Deliberately do NOT write a snapshot — simulates the migration case.

	destBase := filepath.Join(projectRoot, ".moai-backups", "test")
	if err := os.MkdirAll(destBase, defs.DirPerm); err != nil {
		t.Fatalf("mkdir destBase: %v", err)
	}
	destFallback := filepath.Join(destBase, "base")
	destDefaults := filepath.Join(destBase, "defaults")
	if err := SaveTemplateBase(destFallback, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (fallback): %v", err)
	}
	if err := SaveTemplateDefaults(destDefaults); err != nil {
		t.Fatalf("SaveTemplateDefaults: %v", err)
	}

	defaultsSections := filepath.Join(destDefaults, "sections")
	entries, err := os.ReadDir(defaultsSections)
	if err != nil {
		t.Fatalf("readdir defaults: %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("SaveTemplateDefaults produced no sections — embedded FS read failed")
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		want, _ := os.ReadFile(filepath.Join(defaultsSections, e.Name()))
		got, err := os.ReadFile(filepath.Join(destFallback, "sections", e.Name()))
		if err != nil {
			t.Fatalf("fallback missing %s: %v", e.Name(), err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("fallback %s = %q, want %q", e.Name(), got, want)
		}
	}
}

// TestSaveTemplateBase_SnapshotMkdirFails covers SaveTemplateBase's
// dest-dir creation failure path.
func TestSaveTemplateBase_SnapshotMkdirFails(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	// destDir parent is a regular file → MkdirAll(destDir/sections) fails.
	destParent := t.TempDir()
	blocker := filepath.Join(destParent, "sections")
	if err := os.WriteFile(blocker, []byte("x"), defs.FilePerm); err != nil {
		t.Fatalf("plant blocker: %v", err)
	}
	if err := SaveTemplateBase(destParent, projectRoot); err == nil {
		t.Fatalf("SaveTemplateBase must error when dest sections dir cannot be created")
	}
}

// TestSaveTemplateBase_SnapshotCopyError covers SaveTemplateBase's
// per-file copy error path (a snapshot source file that is unreadable).
func TestSaveTemplateBase_SnapshotCopyError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("chmod-based unreadability is ineffective when running as root")
	}
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"system.yaml": "version: \"3.0.1\"\n"})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	// Make a snapshot source unreadable so SaveTemplateBase's os.ReadFile fails.
	bad := filepath.Join(SnapshotDir(projectRoot), "sections", "system.yaml")
	if err := os.Chmod(bad, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(bad, 0o644) })

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err == nil {
		t.Fatalf("SaveTemplateBase must error when a snapshot file cannot be read")
	}
}

// TestSaveTemplateBase_SnapshotDirMissing covers the degenerate case: the
// snapshot marker (sections dir) exists but the actual snapshot root path
// resolves to something unexpected.
func TestSaveTemplateBase_SnapshotDirMissing(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	// No config dir, no snapshot → HasSnapshot false → fallback to SaveTemplateDefaults.
	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (no snapshot, fallback): %v", err)
	}
	// Fallback wrote embedded sections.
	if _, err := os.Stat(filepath.Join(destDir, "sections")); err != nil {
		t.Errorf("fallback did not write sections/: %v", err)
	}
}

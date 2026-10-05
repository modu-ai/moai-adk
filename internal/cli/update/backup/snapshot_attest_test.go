package backup

// Card t1216: a merge BASE is only as good as the moment its snapshot was
// taken. Binaries before card t1139 wrote the snapshot AFTER the restore step,
// so the snapshot held the user's own values; the next merge then read every
// customization as "unchanged from BASE" and replaced it with the template
// default. Such a snapshot is still on disk in every project that ran an
// update with one of those binaries, and it looks exactly like a good one.
//
// SaveTemplateBase must therefore use a snapshot only when it can tell the
// snapshot was taken right after a deploy (before any user value was written
// over the render); any other snapshot falls back to the embedded defaults,
// which never mistake a user value for an unchanged one.

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
)

// readSectionsTree returns name -> bytes for every file under dir/sections.
func readSectionsTree(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	sections := filepath.Join(dir, "sections")
	entries, err := os.ReadDir(sections)
	if err != nil {
		t.Fatalf("readdir %s: %v", sections, err)
	}
	out := map[string][]byte{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(sections, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = data
	}
	return out
}

// assertEmbeddedDefaultsBase asserts destDir holds exactly the bytes
// SaveTemplateDefaults writes, i.e. the fallback BASE was taken.
func assertEmbeddedDefaultsBase(t *testing.T, destDir string) {
	t.Helper()
	want := t.TempDir()
	if err := SaveTemplateDefaults(want); err != nil {
		t.Fatalf("SaveTemplateDefaults: %v", err)
	}
	const name = "git-strategy.yaml"
	w, ok := readSectionsTree(t, want)[name]
	if !ok {
		t.Fatalf("premise: embedded defaults carry no %s", name)
	}
	if g := readSectionsTree(t, destDir)[name]; !bytes.Equal(g, w) {
		t.Errorf("BASE %s is not the embedded default: got %q", name, g)
	}
}

// TestSaveTemplateBase_UnattestedSnapshotFallsBack: a snapshot tree with no
// record of when it was taken is the shape every pre-t1139 update left behind,
// so it is not trusted as BASE.
func TestSaveTemplateBase_UnattestedSnapshotFallsBack(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	if err := os.MkdirAll(snapSections, defs.DirPerm); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	// The user's own value, as a post-restore snapshot would carry it.
	if err := os.WriteFile(filepath.Join(snapSections, "git-strategy.yaml"),
		[]byte("git_strategy:\n  worktree_base_branch: develop\n"), defs.FilePerm); err != nil {
		t.Fatalf("write: %v", err)
	}

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase: %v", err)
	}
	assertEmbeddedDefaultsBase(t, destDir)
}

// TestSaveTemplateBase_SnapshotRewrittenAfterWriteFallsBack: an older binary
// run after a newer one rewrites the snapshot files after its restore. The
// rewritten snapshot must not inherit the trust the newer write earned.
func TestSaveTemplateBase_SnapshotRewrittenAfterWriteFallsBack(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{
		"git-strategy.yaml": "git_strategy:\n  worktree_base_branch: \"\"\n",
	})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	// Premise: an untouched WriteSnapshot result is used as BASE.
	trusted := t.TempDir()
	if err := SaveTemplateBase(trusted, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (trusted): %v", err)
	}
	if got := readSectionsTree(t, trusted)["git-strategy.yaml"]; string(got) != "git_strategy:\n  worktree_base_branch: \"\"\n" {
		t.Fatalf("premise: WriteSnapshot result not used as BASE, got %q", got)
	}

	// The older binary's post-restore rewrite.
	snapFile := filepath.Join(SnapshotDir(projectRoot), "sections", "git-strategy.yaml")
	if err := os.WriteFile(snapFile, []byte("git_strategy:\n  worktree_base_branch: develop\n"), defs.FilePerm); err != nil {
		t.Fatalf("rewrite snapshot: %v", err)
	}

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase: %v", err)
	}
	assertEmbeddedDefaultsBase(t, destDir)
}

// TestSaveTemplateBase_UnreadableSnapshotFallsBack: a snapshot whose digest
// cannot be computed is not proven, so it must not be used as BASE. Using it
// half-way is worse than not at all: the copy stops at the unreadable file,
// the files copied before it (possibly rewritten with user values by an older
// binary) become a partial BASE, and the merge reads those user values as
// unchanged and drops them — the loss this card fixes (sync-audit F1).
func TestSaveTemplateBase_UnreadableSnapshotFallsBack(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod-based unreadability does not deny access on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("chmod-based unreadability is ineffective when running as root")
	}
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{
		"a-git-strategy.yaml": "git_strategy:\n  worktree_base_branch: \"\"\n",
		"z-system.yaml":       "version: \"3.0.1\"\n",
	})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	// An older binary's post-restore rewrite, then an unreadable file after it.
	snapSections := filepath.Join(SnapshotDir(projectRoot), "sections")
	if err := os.WriteFile(filepath.Join(snapSections, "a-git-strategy.yaml"),
		[]byte("git_strategy:\n  worktree_base_branch: develop\n"), defs.FilePerm); err != nil {
		t.Fatalf("rewrite snapshot: %v", err)
	}
	bad := filepath.Join(snapSections, "z-system.yaml")
	if err := os.Chmod(bad, 0); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(bad, 0o644) })

	destDir := t.TempDir()
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase: %v (an unproven snapshot must fall back, not fail half-copied)", err)
	}
	if got := readSectionsTree(t, destDir)["a-git-strategy.yaml"]; got != nil {
		t.Errorf("BASE carries the snapshot's user value %q", got)
	}
	assertEmbeddedDefaultsBase(t, destDir)
	if _, err := os.Stat(filepath.Join(destDir, UnattestedBaseMarker)); err != nil {
		t.Errorf("fallback BASE carries no marker: %v", err)
	}
}

// TestSaveTemplateBase_UnattestedMarkerFollowsTheBase: the fallback leaves the
// marker the restore reads to list kept values, and an attested BASE written
// into the same directory afterwards (two updates in one second share a backup
// directory) clears it.
func TestSaveTemplateBase_UnattestedMarkerFollowsTheBase(t *testing.T) {
	t.Parallel()
	projectRoot := t.TempDir()
	writeSections(t, projectRoot, map[string]string{"quality.yaml": "test_coverage_target: 80\n"})
	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	if err := os.Remove(snapshotAttestPath(projectRoot)); err != nil {
		t.Fatalf("drop attestation: %v", err)
	}
	destDir := t.TempDir()
	marker := filepath.Join(destDir, UnattestedBaseMarker)

	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (unattested): %v", err)
	}
	if _, err := os.Stat(marker); err != nil {
		t.Fatalf("fallback BASE carries no marker: %v", err)
	}

	if err := WriteSnapshot(projectRoot); err != nil {
		t.Fatalf("WriteSnapshot: %v", err)
	}
	if err := SaveTemplateBase(destDir, projectRoot); err != nil {
		t.Fatalf("SaveTemplateBase (attested): %v", err)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Errorf("attested BASE kept the stale marker (stat err = %v)", err)
	}
}

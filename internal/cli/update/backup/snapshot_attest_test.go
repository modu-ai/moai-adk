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

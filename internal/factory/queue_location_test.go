// queue_location_test.go — SPEC-MOAI-DB-PATHS-001 (card t1827) M1: the pure
// queue-location report. Every case builds its project under t.TempDir() and
// pins MOAI_HOME to an absolute path inside it (or clears it for the
// temporary-origin layout); nothing reaches the operator's real home.
package factory

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
	"github.com/modu-ai/moai-adk/internal/paths"
)

// snapshotTree lists every path below dir with its size and mtime, so a
// before/after comparison detects a created, moved, or rewritten file.
func snapshotTree(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		out[path] = info.ModTime().String() + "|" + filepath.Base(path)
		if !info.IsDir() {
			out[path] += "|" + strconv.FormatInt(info.Size(), 10)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return out
}

func sameSnapshot(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

func TestQueueLocationForRoot_ProjectLocal(t *testing.T) {
	root := t.TempDir()
	t.Setenv(paths.EnvHome, "")

	got := QueueLocationForRoot(root)

	wantDir := filepath.Join(root, ".moai", "state", "todo")
	if got.Dir != wantDir {
		t.Errorf("Dir = %q, want %q", got.Dir, wantDir)
	}
	if want := filepath.Join(wantDir, "backlog.db"); got.DBPath != want {
		t.Errorf("DBPath = %q, want %q", got.DBPath, want)
	}
	if got.Layout != QueueLayoutProjectLocal {
		t.Errorf("Layout = %q, want %q", got.Layout, QueueLayoutProjectLocal)
	}
	if got.LegacyDir != "" {
		t.Errorf("LegacyDir = %q, want empty", got.LegacyDir)
	}
}

func TestQueueLocationForRoot_Home(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "moai-home")
	t.Setenv(paths.EnvHome, home)

	got := QueueLocationForRoot(root)

	if want := StateDirForRoot(root); got.Dir != want {
		t.Errorf("Dir = %q, want the resolver value %q", got.Dir, want)
	}
	if want := filepath.Join(home, "db", homestate.ProjectKey(root), "todo"); got.Dir != want {
		t.Errorf("Dir = %q, want home layout %q", got.Dir, want)
	}
	if want := filepath.Join(got.Dir, "backlog.db"); got.DBPath != want {
		t.Errorf("DBPath = %q, want %q", got.DBPath, want)
	}
	if got.Layout != QueueLayoutHome {
		t.Errorf("Layout = %q, want %q", got.Layout, QueueLayoutHome)
	}
	if got.LegacyDir != "" {
		t.Errorf("LegacyDir = %q, want empty", got.LegacyDir)
	}
}

// TestQueueLocationForRoot_LegacyPreAdoption: the queue sits only in the
// project-local directory while the canonical home directory holds none. The
// canonical path is reported, the legacy directory is disclosed, and nothing
// is moved or created.
func TestQueueLocationForRoot_LegacyPreAdoption(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "moai-home")
	t.Setenv(paths.EnvHome, home)

	legacy := projectStateDirForRoot(root)
	if _, _, err := NewBacklogStore(filepath.Join(legacy, backlogFileName)).Add("legacy card"); err != nil {
		t.Fatalf("seed legacy queue: %v", err)
	}
	before := snapshotTree(t, root)

	got := QueueLocationForRoot(root)

	if want := StateDirForRoot(root); got.Dir != want {
		t.Errorf("Dir = %q, want canonical %q", got.Dir, want)
	}
	if got.LegacyDir != legacy {
		t.Errorf("LegacyDir = %q, want %q", got.LegacyDir, legacy)
	}
	if got.Layout != QueueLayoutHome {
		t.Errorf("Layout = %q, want %q", got.Layout, QueueLayoutHome)
	}
	if after := snapshotTree(t, root); !sameSnapshot(before, after) {
		t.Errorf("QueueLocationForRoot changed the project tree")
	}
	if _, err := os.Stat(got.Dir); !os.IsNotExist(err) {
		t.Errorf("canonical directory %q was created (stat err = %v)", got.Dir, err)
	}
}

// TestQueueLocationForRoot_PostAdoption: after the adopting open path has
// relocated the queue, the same call reports the canonical directory with no
// legacy disclosure and the same Dir it reported before the relocation.
func TestQueueLocationForRoot_PostAdoption(t *testing.T) {
	root := t.TempDir()
	home := filepath.Join(t.TempDir(), "moai-home")
	t.Setenv(paths.EnvHome, home)

	legacy := projectStateDirForRoot(root)
	if _, _, err := NewBacklogStore(filepath.Join(legacy, backlogFileName)).Add("legacy card"); err != nil {
		t.Fatalf("seed legacy queue: %v", err)
	}
	pre := QueueLocationForRoot(root)

	BacklogPathForRootAdopting(root)

	post := QueueLocationForRoot(root)
	if post.Dir != pre.Dir {
		t.Errorf("Dir changed across adoption: %q -> %q", pre.Dir, post.Dir)
	}
	if post.DBPath != pre.DBPath {
		t.Errorf("DBPath changed across adoption: %q -> %q", pre.DBPath, post.DBPath)
	}
	if post.LegacyDir != "" {
		t.Errorf("LegacyDir after adoption = %q, want empty", post.LegacyDir)
	}
	if _, err := os.Stat(post.DBPath); err != nil {
		t.Errorf("queue database %q absent after adoption: %v", post.DBPath, err)
	}
}

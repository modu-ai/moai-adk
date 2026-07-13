package verify

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a throwaway git repo under t.TempDir() with one
// committed file (tracked.txt) and returns its path. `.moai/` is gitignored,
// mirroring the real project layout — the snapshot store writes under
// .moai/state/, which MUST be ignored so recording a snapshot does not
// invalidate the very key it was recorded under (porcelain=v2 lists untracked
// non-ignored paths only).
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	gitRun(t, dir, "init", "-q")
	gitRun(t, dir, "config", "user.email", "test@example.com")
	gitRun(t, dir, "config", "user.name", "test")
	gitRun(t, dir, "config", "commit.gpgsign", "false")
	writeFile(t, dir, ".gitignore", ".moai/\n")
	writeFile(t, dir, "tracked.txt", "v1\n")
	gitRun(t, dir, "add", ".gitignore", "tracked.txt")
	gitRun(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// TestSnapshotKey exercises the 3-input key composition (HEAD SHA +
// porcelain-v2 digest + diff-HEAD content hash): every tree-state change —
// including the D13 boundary case of RE-EDITING an already-dirty tracked file,
// where porcelain-v2 output is byte-identical — must yield a distinct key.
func TestSnapshotKey(t *testing.T) {
	t.Parallel()
	dir := initTestRepo(t)
	ctx := context.Background()

	keys := map[string]string{}
	record := func(label string) {
		t.Helper()
		k, err := Key(ctx, dir)
		if err != nil {
			t.Fatalf("Key(%s): %v", label, err)
		}
		for prev, pk := range keys {
			if pk == k {
				t.Fatalf("key for %q equals key for %q: %s", label, prev, k)
			}
		}
		keys[label] = k
	}

	// Determinism first: same tree twice → same key.
	k1, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key(clean #1): %v", err)
	}
	k2, err := Key(ctx, dir)
	if err != nil {
		t.Fatalf("Key(clean #2): %v", err)
	}
	if k1 != k2 {
		t.Fatalf("same tree must yield same key: %s vs %s", k1, k2)
	}

	record("clean tree")

	// Dirty tracked edit (unstaged).
	writeFile(t, dir, "tracked.txt", "v2\n")
	record("dirty tracked edit")

	// D13 boundary: RE-EDIT the already-dirty tracked file. porcelain-v2 output
	// is byte-identical across successive edits to an already-dirty file; the
	// diff-HEAD content-hash leg must change the key.
	writeFile(t, dir, "tracked.txt", "v3\n")
	record("re-edit of already-dirty tracked file (D13)")

	// Staged edit (same content, now staged — index state changes).
	gitRun(t, dir, "add", "tracked.txt")
	record("staged edit")

	// Add untracked file.
	writeFile(t, dir, "untracked.txt", "u1\n")
	record("add untracked file")

	// HEAD advance (commit everything → clean tree at NEW head; distinct from
	// the original clean-tree key because HEAD SHA differs).
	gitRun(t, dir, "add", ".")
	gitRun(t, dir, "commit", "-q", "-m", "advance")
	record("HEAD advance")
}

// TestSnapshotKeyNonRepo asserts a non-git directory returns an error (callers
// fall back to re-execution — fail-open).
func TestSnapshotKeyNonRepo(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if _, err := Key(context.Background(), dir); err == nil {
		t.Fatal("Key on a non-repo must return an error")
	}
}

// TestSnapshotKeyContextCancel asserts a cancelled context aborts key
// computation with an error (the Advisory-Check time-box contract: on
// deadline exceed the caller skips the optimization and re-executes).
func TestSnapshotKeyContextCancel(t *testing.T) {
	t.Parallel()
	dir := initTestRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := Key(ctx, dir); err == nil {
		t.Fatal("Key with cancelled context must return an error")
	}
}

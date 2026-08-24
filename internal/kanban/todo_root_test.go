package kanban

// todo_root_test.go — SPEC-WEB-TODO-QUEUE-001 M1: the relocated queue-root
// resolution, split into a pure resolver (the console's entry point) and an
// adopt-then-resolve entry point (the `moai todo` command path).
//
// Every test here overrides the package's HomeDirFn seam, which is
// process-global, so none of them run in parallel.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// initTodoRootGitRepo turns dir into a committed repository so the
// git-resolvable branch treats it as a primary checkout.
func initTodoRootGitRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	runTodoRootGit(t, dir, "init", "-q")
	runTodoRootGit(t, dir, "config", "user.email", "t@example.com")
	runTodoRootGit(t, dir, "config", "user.name", "t")
	runTodoRootGit(t, dir, "commit", "--allow-empty", "-q", "-m", "init")
}

func runTodoRootGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", full, err, out)
	}
}

// sameTodoRootDir compares two path spellings, resolving symlinks (macOS
// /var -> /private/var) where both resolve.
func sameTodoRootDir(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}

// stubHome points the package's home seam at dir for the test's duration.
func stubHome(t *testing.T, dir string) {
	t.Helper()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return dir, nil }
	t.Cleanup(func() { HomeDirFn = orig })
}

// seedLocalQueue writes a project-local backlog file under base holding n
// queued items and returns its path.
func seedLocalQueue(t *testing.T, base string, n int) string {
	t.Helper()
	path := BacklogPathForRoot(base)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir local queue dir: %v", err)
	}
	rows := make([]string, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, fmt.Sprintf(
			`{"id":"t%d","text":"card %d","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"}`, i, i))
	}
	seed := fmt.Sprintf(`{"version":1,"last_seq":%d,"items":[%s]}`, n, strings.Join(rows, ","))
	if err := os.WriteFile(path, []byte(seed), 0o600); err != nil {
		t.Fatalf("seed local queue: %v", err)
	}
	return path
}

// TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary — AC-WTQ-005 producer
// half: from a linked worktree the pure resolver returns the primary checkout.
func TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary(t *testing.T) {
	primary := t.TempDir()
	initTodoRootGitRepo(t, primary)
	wt := filepath.Join(t.TempDir(), "wt")
	runTodoRootGit(t, primary, "worktree", "add", wt)

	if got := ResolveTodoQueueRoot(wt); !sameTodoRootDir(got, primary) {
		t.Fatalf("pure resolver from worktree = %q, want primary %q", got, primary)
	}
}

// TestResolveTodoQueueRoot_FallbackNoGit — the home-based fallback root, and
// the deterministic project key it is named for.
func TestResolveTodoQueueRoot_FallbackNoGit(t *testing.T) {
	dir := t.TempDir() // deliberately NOT a git repository
	home := t.TempDir()
	stubHome(t, home)

	got := ResolveTodoQueueRoot(dir)
	want := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))
	if got != want {
		t.Fatalf("fallback queue root = %q, want %q", got, want)
	}
	key := TodoQueueProjectKey(dir)
	base := filepath.Base(dir)
	if !strings.HasPrefix(key, base+"-") {
		t.Fatalf("project key %q lacks %q prefix", key, base+"-")
	}
	if suffix := strings.TrimPrefix(key, base+"-"); len(suffix) != 8 {
		t.Fatalf("project key digest %q is not 8 hex chars", suffix)
	}
	if TodoQueueProjectKey(dir) == TodoQueueProjectKey(t.TempDir()) {
		t.Fatalf("distinct directories share a project key")
	}
}

// TestResolveTodoQueueRoot_PureFallbackWritesNothing — AC-WTQ-006: under the
// exact preconditions adoption migrates on, the pure resolver leaves the disk
// untouched — the local file at its original path with its original mtime,
// and nothing created under the fallback root.
func TestResolveTodoQueueRoot_PureFallbackWritesNothing(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)

	local := seedLocalQueue(t, dir, 2)
	before, err := os.Stat(local)
	if err != nil {
		t.Fatalf("stat seeded local queue: %v", err)
	}
	fallbackRoot := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))

	time.Sleep(10 * time.Millisecond)
	_ = ResolveTodoQueueRoot(dir)

	after, err := os.Stat(local)
	if err != nil {
		t.Fatalf("local queue moved or removed by the pure resolver: %v", err)
	}
	if !after.ModTime().Equal(before.ModTime()) {
		t.Fatalf("local queue mtime changed: %v -> %v", before.ModTime(), after.ModTime())
	}
	if _, err := os.Stat(fallbackRoot); !os.IsNotExist(err) {
		t.Fatalf("pure resolver created the fallback root %q (stat err = %v)", fallbackRoot, err)
	}
}

// TestResolveTodoQueueRoot_ReadThroughToProjectLocal — AC-WTQ-007 / decision
// D-2: when the fallback root holds no queue file and a project-local one
// exists, the pure resolver resolves to the PROJECT-LOCAL root, so the console
// lists the same N items `moai todo` reports.
func TestResolveTodoQueueRoot_ReadThroughToProjectLocal(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	seedLocalQueue(t, dir, 3)

	got := ResolveTodoQueueRoot(dir)
	if got != dir {
		t.Fatalf("read-through root = %q, want project-local root %q", got, dir)
	}
	rec, err := NewBacklogStore(BacklogPathForRoot(got)).Load()
	if err != nil {
		t.Fatalf("load through resolved root: %v", err)
	}
	if len(rec.Items) != 3 {
		t.Fatalf("read-through load holds %d items, want 3", len(rec.Items))
	}
}

// TestResolveTodoQueueRoot_PopulatedFallbackWins — the read-through predicate
// mirrors adoption's early return: once the fallback root holds a queue file,
// resolution stays on it even though a local file lingers.
func TestResolveTodoQueueRoot_PopulatedFallbackWins(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	seedLocalQueue(t, dir, 2)

	fallbackRoot := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))
	if err := os.MkdirAll(fallbackRoot, 0o755); err != nil {
		t.Fatalf("mkdir fallback root: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fallbackRoot, "backlog.json"),
		[]byte(`{"version":1,"last_seq":0,"items":[]}`), 0o600); err != nil {
		t.Fatalf("seed fallback queue: %v", err)
	}

	if got := ResolveTodoQueueRoot(dir); got != fallbackRoot {
		t.Fatalf("resolved root = %q, want populated fallback %q", got, fallbackRoot)
	}
}

// TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing — the third branch
// (plan.md §G): no git AND no home. It returns the in-project root and, like
// every other branch of the pure resolver, writes nothing.
func TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing(t *testing.T) {
	dir := t.TempDir()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { HomeDirFn = orig })

	got := ResolveTodoQueueRoot(dir)
	want := filepath.Join(dir, ".moai", "state", "kanban")
	if got != want {
		t.Fatalf("home-unresolvable root = %q, want %q", got, want)
	}
	if _, err := os.Stat(want); !os.IsNotExist(err) {
		t.Fatalf("home-unresolvable branch created %q (stat err = %v)", want, err)
	}
}

// TestResolveTodoQueueRootAdopting_AdoptsLocalQueue — AC-WTQ-008: the command
// path still adopts. The local file leaves its original path, the fallback
// root holds the queue, and item count and states are unchanged.
func TestResolveTodoQueueRootAdopting_AdoptsLocalQueue(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	local := seedLocalQueue(t, dir, 3)

	root := ResolveTodoQueueRootAdopting(dir)
	want := filepath.Join(home, ".moai", "todo", TodoQueueProjectKey(dir))
	if root != want {
		t.Fatalf("adopting root = %q, want %q", root, want)
	}
	if _, err := os.Stat(local); !os.IsNotExist(err) {
		t.Fatalf("local queue still at %q after adoption (stat err = %v)", local, err)
	}
	rec, err := NewBacklogStore(filepath.Join(root, "backlog.json")).Load()
	if err != nil {
		t.Fatalf("load adopted queue: %v", err)
	}
	if len(rec.Items) != 3 {
		t.Fatalf("adopted queue holds %d items, want 3", len(rec.Items))
	}
	for _, it := range rec.Items {
		if string(it.State) != "queued" {
			t.Fatalf("adopted item %s state = %q, want queued", it.ID, it.State)
		}
	}
}

// TestResolveTodoQueueRootAdopting_GitBranchIsPure — the adopting entry point
// only adopts on the fallback branch; a git-resolvable context is untouched.
func TestResolveTodoQueueRootAdopting_GitBranchIsPure(t *testing.T) {
	primary := t.TempDir()
	initTodoRootGitRepo(t, primary)
	home := t.TempDir()
	stubHome(t, home)

	if got := ResolveTodoQueueRootAdopting(primary); !sameTodoRootDir(got, primary) {
		t.Fatalf("adopting root from primary = %q, want %q", got, primary)
	}
	if entries, err := os.ReadDir(home); err == nil && len(entries) != 0 {
		t.Fatalf("git branch touched the home root: %d entries", len(entries))
	}
}

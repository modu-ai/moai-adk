package cli

// todo_queue_root_test.go — queue-residence tests for t106: the backlog
// queue hangs from the PRIMARY checkout, so a linked worktree and the
// primary see one file (the delegation-channel invariant), and a launch
// context without git metadata falls back to the home-based key.
//
// Every test here is serial: they mutate CLAUDE_PROJECT_DIR (t.Setenv) and
// the userHomeDirFn seam, which are process-global.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/kanban"
)

// initGitRepo turns dir into a committed git repository so queue-root
// resolution — which resolves through git — treats it as a primary checkout.
// Skips the calling test when git is unavailable. User- and system-level
// git config are pointed at /dev/null so a host's global config cannot leak
// into (or break) the init.
func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	runGitIn(t, dir, "init", "-q")
	runGitIn(t, dir, "config", "user.email", "t@example.com")
	runGitIn(t, dir, "config", "user.name", "t")
	runGitIn(t, dir, "commit", "--allow-empty", "-q", "-m", "init")
}

// runGitIn runs `git -C dir <args...>` and fails the test on a non-zero exit.
func runGitIn(t *testing.T, dir string, args ...string) {
	t.Helper()
	full := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", full...)
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_SYSTEM=/dev/null")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", full, err, out)
	}
}

// addGitWorktree creates a linked worktree of the committed repository at
// primary and returns its path.
func addGitWorktree(t *testing.T, primary string) string {
	t.Helper()
	wt := filepath.Join(t.TempDir(), "wt")
	runGitIn(t, primary, "worktree", "add", wt)
	return wt
}

// sameDir reports whether two path spellings name the same directory,
// resolving symlinks on platforms where the temp dir path is a symlink
// (macOS /var -> /private/var) while git may report either spelling.
func sameDir(a, b string) bool {
	ra, errA := filepath.EvalSymlinks(a)
	rb, errB := filepath.EvalSymlinks(b)
	if errA != nil || errB != nil {
		return filepath.Clean(a) == filepath.Clean(b)
	}
	return ra == rb
}

// TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary is the t106 core
// assertion: invoked from a linked worktree, the queue root is the primary
// checkout — never the worktree itself.
func TestResolveTodoQueueRoot_WorktreeConvergesOnPrimary(t *testing.T) {
	primary := t.TempDir()
	initGitRepo(t, primary)
	wt := addGitWorktree(t, primary)

	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	if got := resolveTodoQueueRoot(); !sameDir(got, primary) {
		t.Fatalf("queue root from worktree = %q, want primary %q", got, primary)
	}
}

// TestResolveTodoQueueRoot_PrimaryIsItself pins the primary-side identity:
// invoked from the primary checkout, the queue root is that checkout.
func TestResolveTodoQueueRoot_PrimaryIsItself(t *testing.T) {
	primary := t.TempDir()
	initGitRepo(t, primary)

	t.Setenv("CLAUDE_PROJECT_DIR", primary)
	if got := resolveTodoQueueRoot(); !sameDir(got, primary) {
		t.Fatalf("queue root from primary = %q, want %q", got, primary)
	}
}

// TestResolveTodoQueueRoot_SubdirectoryResolvesToRepoRoot covers a launch
// context inside a repository subdirectory: the queue still hangs from the
// repository root, not the subdirectory.
func TestResolveTodoQueueRoot_SubdirectoryResolvesToRepoRoot(t *testing.T) {
	primary := t.TempDir()
	initGitRepo(t, primary)
	sub := filepath.Join(primary, "deep", "nested")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatalf("mkdir sub: %v", err)
	}

	t.Setenv("CLAUDE_PROJECT_DIR", sub)
	if got := resolveTodoQueueRoot(); !sameDir(got, primary) {
		t.Fatalf("queue root from subdir = %q, want repo root %q", got, primary)
	}
}

// TestResolveTodoQueueRoot_FallbackNoGit covers the home-based fallback: a
// launch context with no git metadata keeps one queue under
// ~/.moai/todo/<project-key>/, keyed deterministically from the directory.
func TestResolveTodoQueueRoot_FallbackNoGit(t *testing.T) {
	dir := t.TempDir() // deliberately NOT a git repository
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	home := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	got := resolveTodoQueueRoot()
	want := filepath.Join(home, ".moai", "todo", todoQueueProjectKey(dir))
	if got != want {
		t.Fatalf("fallback queue root = %q, want %q", got, want)
	}

	// The key is base name + 8 hex digest chars — readable and collision-safe.
	key := todoQueueProjectKey(dir)
	base := filepath.Base(dir)
	if !strings.HasPrefix(key, base+"-") {
		t.Fatalf("project key %q lacks base+%q prefix", key, base)
	}
	if suffix := strings.TrimPrefix(key, base+"-"); len(suffix) != 8 {
		t.Fatalf("project key digest %q is not 8 hex chars", suffix)
	}

	// Two distinct directories must not share a key.
	other := t.TempDir()
	if todoQueueProjectKey(dir) == todoQueueProjectKey(other) {
		t.Fatalf("distinct directories %q and %q share a project key", dir, other)
	}
}

// TestTodoQueue_FallbackAdoptsExistingLocalQueue is [HARD] verification 3 in
// code form: a project-local queue that predates the fallback cutover must be
// ADOPTED — same items, same states — never shadowed behind an empty
// home-based queue.
func TestTodoQueue_FallbackAdoptsExistingLocalQueue(t *testing.T) {
	dir := t.TempDir() // deliberately NOT a git repository
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	home := t.TempDir()
	orig := userHomeDirFn
	userHomeDirFn = func() (string, error) { return home, nil }
	t.Cleanup(func() { userHomeDirFn = orig })

	// Seed a pre-fallback local queue: 2 queued + 1 picked, ids from an
	// earlier high-water mark — the shape a v3.1.0-era project carries.
	spec := "SPEC-EXAMPLE-001"
	localDir := filepath.Join(dir, ".moai", "state", "kanban")
	if err := os.MkdirAll(localDir, 0o755); err != nil {
		t.Fatalf("mkdir local queue dir: %v", err)
	}
	seed := `{"version":1,"last_seq":7,"items":[` +
		`{"id":"t6","text":"queued one","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"},` +
		`{"id":"t7","text":"queued two","added_at":"2026-08-14T00:01:00Z","spec_id":null,"state":"queued"},` +
		`{"id":"t5","text":"picked one","added_at":"2026-08-14T00:02:00Z","spec_id":"` + spec + `","state":"picked"}]}`
	if err := os.WriteFile(filepath.Join(localDir, "backlog.json"), []byte(seed), 0o600); err != nil {
		t.Fatalf("seed local queue: %v", err)
	}

	// First fallback-resolution run: the root computation itself adopts.
	root := resolveTodoQueueRoot()
	want := filepath.Join(home, ".moai", "todo", todoQueueProjectKey(dir))
	if root != want {
		t.Fatalf("fallback queue root = %q, want %q", root, want)
	}

	rec, err := kanban.NewBacklogStore(filepath.Join(root, "backlog.json")).Load()
	if err != nil {
		t.Fatalf("load adopted queue: %v", err)
	}
	if len(rec.Items) != 3 {
		t.Fatalf("adopted queue holds %d items, want 3 (items: %+v)", len(rec.Items), rec.Items)
	}
	states := map[string]string{}
	for _, it := range rec.Items {
		states[it.ID] = string(it.State)
	}
	if states["t6"] != "queued" || states["t7"] != "queued" || states["t5"] != "picked" {
		t.Fatalf("adopted states altered: %+v", states)
	}
	if rec.LastSeq != 7 {
		t.Fatalf("adopted last_seq = %d, want 7", rec.LastSeq)
	}

	// A re-run must not duplicate or re-adopt: the populated fallback wins and
	// the local path (renamed away on the same volume, or an inert leftover
	// after a cross-volume copy) is never allowed to shadow it.
	if again := resolveTodoQueueRoot(); again != want {
		t.Fatalf("second fallback resolution = %q, want %q", again, want)
	}
	rec2, err := kanban.NewBacklogStore(filepath.Join(root, "backlog.json")).Load()
	if err != nil {
		t.Fatalf("reload adopted queue: %v", err)
	}
	if len(rec2.Items) != 3 {
		t.Fatalf("second adoption run changed item count: %d, want 3", len(rec2.Items))
	}
}
// TestTodoQueue_WorktreeSeesPrimaryQueue is the [HARD] acceptance pair in
// code form: (1) a list from the worktree reports the primary's item count,
// and (2) an add issued from the worktree lands in the primary's queue file.
func TestTodoQueue_WorktreeSeesPrimaryQueue(t *testing.T) {
	primary := t.TempDir()
	initGitRepo(t, primary)
	wt := addGitWorktree(t, primary)

	primaryStore := kanban.NewBacklogStore(todoBacklogPath(primary))
	for _, text := range []string{"seed one", "seed two", "seed three"} {
		if _, _, err := primaryStore.Add(text); err != nil {
			t.Fatalf("seed add: %v", err)
		}
	}

	// (1) list from the worktree context sees the primary's three items.
	t.Setenv("CLAUDE_PROJECT_DIR", wt)
	out, _, err := runTodo(t, "list")
	if err != nil {
		t.Fatalf("list from worktree: %v", err)
	}
	if got := strings.Count(out, "seed "); got != 3 {
		t.Fatalf("worktree list shows %d seeded items, want 3\nlist output:\n%s", got, out)
	}

	// (2) an add from the worktree lands in the primary's queue file.
	if _, _, err := runTodo(t, "add", "from worktree"); err != nil {
		t.Fatalf("add from worktree: %v", err)
	}
	rec, err := primaryStore.Load()
	if err != nil {
		t.Fatalf("load primary queue: %v", err)
	}
	found := false
	for _, it := range rec.Items {
		if it.Text == "from worktree" {
			found = true
		}
	}
	if !found {
		t.Fatalf("add from worktree did not land in the primary queue; items: %+v", rec.Items)
	}
}

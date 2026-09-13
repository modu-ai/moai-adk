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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/paths"
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
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
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

// TestResolveTodoQueueRoot_FallbackNoGit — the home-based fallback, and the
// deterministic project key it is named for.
//
// INTENTIONAL UPDATE (t621): the asserted value moves from the home fallback
// ROOT to the launch base, and the home property is asserted where it is now
// decided — on the queue PATH. The two are not the same claim, and the old one
// became false: a root under ~/.moai is re-keyed by the layer below, so
// naming one sent the queue to ~/.moai/db/<key>-<hash>/todo while every
// anchor-based surface read ~/.moai/db/<key>/todo. What this test exists to
// pin — a non-git project gets a HOME queue, keyed so two such projects cannot
// collide — is asserted here directly rather than through a root-value proxy.
func TestResolveTodoQueueRoot_FallbackNoGit(t *testing.T) {
	dir := t.TempDir() // deliberately NOT a git repository
	home := t.TempDir()
	stubHome(t, home)
	// SPEC-TODO-HOME-TEMP-GUARD-001 preservation transfer: t.TempDir() is
	// inside os.TempDir(), so the temporary-origin guard would otherwise
	// refuse this home queue and the assertion below could never be reached.
	// The fixture moves to a NON-temporary base through the temp-root seam.
	declareNonTemporary(t)

	got := ResolveTodoQueueRoot(dir)
	if got != dir {
		t.Fatalf("fallback queue root = %q, want the launch base %q", got, dir)
	}
	// The home property, on the value that carries it: the queue itself lands
	// under the home directory, not inside the project.
	queue := BacklogPathForRoot(got)
	if !strings.HasPrefix(queue, home+string(filepath.Separator)) {
		t.Fatalf("fallback queue = %q, want it under the home directory %q", queue, home)
	}
	if strings.HasPrefix(queue, dir+string(filepath.Separator)) {
		t.Fatalf("fallback queue = %q is project-local; the no-git fallback must be home-based", queue)
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

// TestResolveTodoQueueRoot_LegacyHomeFallbackQueueIsAdopted — the SUCCESSOR of
// TestResolveTodoQueueRoot_PopulatedFallbackWins (t621).
//
// The predecessor asserted that once the home fallback ROOT held a queue,
// resolution stayed on it. That root is no longer a resolution target, and an
// assertion about which of two locations wins cannot be carried over to a
// design with one. What MUST be carried over is the operator-visible half: a
// queue already sitting at the old fallback location is not stranded by the
// collapse. It is reached the same way every other legacy location is — the
// adoption resolveStateDir performs through legacyHomeStateDirsForRoot, which
// lists ~/.moai/todo/<key> and its nested state directories.
//
// MOAI_HOME is set rather than only stubbing HomeDirFn because the legacy scan
// resolves its home through paths.MoaiHome, which does not read this package's
// seam — without the override the scan would look in the operator's real home
// and the test would assert nothing about the fixture.
func TestResolveTodoQueueRoot_LegacyHomeFallbackQueueIsAdopted(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	moaiHome := filepath.Join(home, ".moai")
	stubHome(t, home)
	t.Setenv(paths.EnvHome, moaiHome)

	// The exact path the retired fallback wrote: BacklogPathForRoot of the
	// fallback root, back when that expanded to <root>/.moai/state/todo.
	legacyRoot := filepath.Join(moaiHome, "todo", TodoQueueProjectKey(dir))
	legacyQueue := filepath.Join(legacyRoot, ".moai", "state", stateDirName, backlogFileName)
	if err := os.MkdirAll(filepath.Dir(legacyQueue), 0o755); err != nil {
		t.Fatalf("mkdir legacy fallback root: %v", err)
	}
	if err := os.WriteFile(legacyQueue,
		[]byte(`{"version":1,"last_seq":2,"items":[`+
			`{"id":"t1","text":"card 1","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"},`+
			`{"id":"t2","text":"card 2","added_at":"2026-08-14T00:00:00Z","spec_id":null,"state":"queued"}]}`),
		0o600); err != nil {
		t.Fatalf("seed legacy fallback queue: %v", err)
	}
	// Precondition: the scan really does name this directory. Without it a
	// PASS below could come from an unrelated path.
	if !slices.Contains(legacyHomeStateDirsForRoot(dir), filepath.Dir(legacyQueue)) {
		t.Fatalf("precondition: %q is not among the legacy sources %v",
			filepath.Dir(legacyQueue), legacyHomeStateDirsForRoot(dir))
	}

	root := ResolveTodoQueueRootAdopting(dir)
	rec, err := NewBacklogStore(todoBacklogPathForTest(root)).Load()
	if err != nil {
		t.Fatalf("load through the resolved root: %v", err)
	}
	if len(rec.Items) != 2 {
		t.Fatalf("the legacy fallback queue was stranded: %d items through %q, want 2",
			len(rec.Items), todoBacklogPathForTest(root))
	}
}

// todoBacklogPathForTest is the command path's own join — the adopting form,
// which is what carries a legacy queue forward.
func todoBacklogPathForTest(root string) string { return BacklogPathForRootAdopting(root) }

// TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing — the third branch
// (plan.md §G): no git AND no home. It returns the in-project root and, like
// every other branch of the pure resolver, writes nothing.
//
// INTENTIONAL UPDATE (SPEC-TODO-HOME-TEMP-GUARD-001, REQ-THG-001): the base
// here is a t.TempDir(), so "no home" and "temporary origin" hold TOGETHER,
// and that requirement fixes the return on their intersection at the launch
// BASE. The asserted value therefore moves from dir/.moai/state/todo to dir.
// This is the implementation of a decision, not an accommodation of the
// implementation: the old value is one layer below the root every consumer
// extends with BacklogPathForRoot, and re-picking it here would have let
// predicate PLACEMENT decide the return. The temp-origin half of this
// intersection is covered by AC-THG-001 branch (c); this test keeps the
// no-home branch's own "writes nothing" half.
func TestResolveTodoQueueRoot_HomeUnresolvableWritesNothing(t *testing.T) {
	dir := t.TempDir()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { HomeDirFn = orig })

	got := ResolveTodoQueueRoot(dir)
	if got != dir {
		t.Fatalf("home-unresolvable root = %q, want the launch base %q", got, dir)
	}
	// Nothing was created on the way: neither the state directory the old
	// return named, nor anything else under the base.
	stateDir := filepath.Join(dir, ".moai", "state", "todo")
	if _, err := os.Stat(stateDir); !os.IsNotExist(err) {
		t.Fatalf("home-unresolvable branch created %q (stat err = %v)", stateDir, err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".moai")); !os.IsNotExist(err) {
		t.Fatalf("home-unresolvable branch created %q", filepath.Join(dir, ".moai"))
	}
}

// TestResolveTodoQueueRootAdopting_AdoptsLocalQueue — AC-WTQ-008: the command
// path still adopts. The local file leaves its original path, the fallback
// root holds the queue, and item count and states are unchanged.
func TestResolveTodoQueueRootAdopting_AdoptsLocalQueue(t *testing.T) {
	dir := t.TempDir() // no git
	home := t.TempDir()
	stubHome(t, home)
	// SPEC-TODO-HOME-TEMP-GUARD-001 preservation transfer: this is
	// SPEC-WEB-TODO-QUEUE-001's AC-WTQ-008 producer, so its adopt-not-shadow
	// assertion is preserved verbatim on a non-temporary base rather than
	// rewritten to the guarded behaviour — rewriting it would withdraw that
	// criterion silently.
	declareNonTemporary(t)
	seedLocalQueue(t, dir, 3)

	root := ResolveTodoQueueRootAdopting(dir)
	// INTENTIONAL UPDATE (t621): the criterion is adopt-NOT-SHADOW — the
	// command path must surface the project's existing cards rather than an
	// empty queue beside them. It was asserted through two proxies that have
	// since stopped tracking it: the root's VALUE (a home root is re-keyed by
	// the layer below, so naming one forks the queue away from every
	// anchor-based surface) and the local file's DISAPPEARANCE (seedLocalQueue
	// resolves through BacklogPathForRoot, which since the home-state migration
	// already places the file in the home database — so "it moved" now means it
	// was moved OFF the canonical path, which is the defect, not the
	// criterion). The criterion itself is asserted below, unweakened: the cards
	// are readable through the resolved root, in the same count and states.
	if root != dir {
		t.Fatalf("adopting root = %q, want the launch base %q", root, dir)
	}
	queue := BacklogPathForRoot(root)
	if !strings.HasPrefix(queue, home+string(filepath.Separator)) {
		t.Fatalf("adopted queue = %q, want it under the home directory %q", queue, home)
	}
	rec, err := NewBacklogStore(BacklogPathForRoot(root)).Load()
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

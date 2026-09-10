package kanban

// root_layer_repro_test.go — t549: homeTodoQueueRoot's no-home branch
// returned a STATE DIRECTORY where every caller expects a ROOT, while the
// consumers re-derive the state directory from whatever root they receive
// (BacklogPathForRoot → resolveStateDir). The pair read one layer too deep and
// rendered an empty queue, with no error, while a real one existed. These
// tests failed on the unrepaired tree and pin the repair.
//
// Every test here overrides process-global seams (HomeDirFn, TempRootsFn,
// MOAI_HOME), so none of them run in parallel.

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// stubHomeUnresolvable makes the package's home seam fail for the test's
// duration.
func stubHomeUnresolvable(t *testing.T) {
	t.Helper()
	orig := HomeDirFn
	HomeDirFn = func() (string, error) { return "", os.ErrNotExist }
	t.Cleanup(func() { HomeDirFn = orig })
}

// pureItemCount reads the queue at path the way the console does (LoadPure)
// and reports the item count, or -1 with the error when the read failed.
func pureItemCount(path string) (int, error) {
	rec, err := NewBacklogStore(path).LoadPure()
	if err != nil {
		return -1, err
	}
	if rec == nil {
		return 0, nil
	}
	return len(rec.Items), nil
}

// TestT549_BacklogPathForRootTreatsArgumentAsRoot pins the consumer-side
// contract the defect collided with: the argument is a ROOT, and a state
// directory passed in its place is extended a second time. Every resolver
// return value must therefore be a root.
func TestT549_BacklogPathForRootTreatsArgumentAsRoot(t *testing.T) {
	t.Setenv(paths.EnvHome, "")
	dir := t.TempDir()
	stubHomeUnresolvable(t)

	stateDir := filepath.Join(dir, ".moai", "state", stateDirName)
	fromRoot := BacklogPathForRoot(dir)
	fromStateDir := BacklogPathForRoot(stateDir)
	t.Logf("BacklogPathForRoot(root)     = %s", fromRoot)
	t.Logf("BacklogPathForRoot(stateDir) = %s", fromStateDir)

	if want := filepath.Join(stateDir, backlogFileName); fromRoot != want {
		t.Fatalf("BacklogPathForRoot(root) = %q, want %q", fromRoot, want)
	}
	if doubled := filepath.Join(stateDir, ".moai", "state", stateDirName, backlogFileName); fromStateDir != doubled {
		t.Fatalf("BacklogPathForRoot(stateDir) = %q, want the doubled %q", fromStateDir, doubled)
	}
}

// TestT549_NoHomeFallbackReadsTheExistingQueue goes end to end through both
// public resolvers, on the branch that reaches homeTodoQueueRoot's no-home
// return: no git, no absolute MOAI_HOME, a base inside os.TempDir() that the
// temp-origin discriminant does NOT classify temporary, and no resolvable home.
func TestT549_NoHomeFallbackReadsTheExistingQueue(t *testing.T) {
	t.Setenv(paths.EnvHome, "")
	dir := t.TempDir()
	stubHomeUnresolvable(t)
	declareNonTemporary(t)

	// Preconditions — each one is a way this test could pass without reaching
	// the branch under test.
	if _, ok := primaryCheckoutRoot(dir); ok {
		t.Fatalf("precondition: base %q must not resolve as a git checkout", dir)
	}
	if explicitMoaiHome() {
		t.Fatalf("precondition: MOAI_HOME must not be an absolute override")
	}
	if reason, isTemp := TempOriginReason(dir); isTemp {
		t.Fatalf("precondition: base %q still classifies temporary (reason %q)", dir, reason)
	}
	if !pathInsideTempDir(dir) {
		t.Fatalf("precondition: base %q is not inside os.TempDir(), fallback branch unreachable", dir)
	}
	if _, ok := homeTodoQueueRoot(dir); ok {
		t.Fatalf("precondition: home must be unresolvable")
	}

	local := seedLocalQueue(t, dir, 3)
	if want := filepath.Join(dir, ".moai", "state", stateDirName, backlogFileName); local != want {
		t.Fatalf("precondition: seeded queue at %q, want the project-local %q", local, want)
	}
	// Control: the fixture is readable at its canonical location. Without it a
	// zero below could be an unreadable fixture rather than a wrong path.
	if n, err := pureItemCount(BacklogPathForRoot(dir)); n != 3 {
		t.Fatalf("control: canonical read holds %d items (err %v), want 3", n, err)
	}

	for _, tc := range []struct {
		name    string
		resolve func(string) string
	}{
		{"pure", ResolveTodoQueueRoot},
		{"adopting", ResolveTodoQueueRootAdopting},
	} {
		root := tc.resolve(dir)
		path := BacklogPathForRoot(root)
		n, err := pureItemCount(path)
		t.Logf("%s: root=%s path=%s items=%d err=%v", tc.name, root, path, n, err)
		if root != dir {
			t.Errorf("%s resolver root = %q, want the launch base %q", tc.name, root, dir)
		}
		if n != 3 {
			t.Errorf("%s resolver read %d items through its root (err %v), want 3", tc.name, n, err)
		}
	}
}

// TestT549_SymlinkedTempBaseWithDefaultRootsReturnsBase shows the branch is
// reachable with the DEFAULT temp-root set (no TempRootsFn stub): a base
// lexically inside os.TempDir() that symlinks to a non-temporary, non-git
// directory passes pathInsideTempDir but not TempOriginReason. Read-only — the
// symlink target is only stat'ed.
func TestT549_SymlinkedTempBaseWithDefaultRootsReturnsBase(t *testing.T) {
	t.Setenv(paths.EnvHome, "")
	const target = "/usr"
	if info, err := os.Stat(target); err != nil || !info.IsDir() {
		t.Skipf("symlink target %q unavailable: %v", target, err)
	}
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	stubHomeUnresolvable(t)

	_, inGit := primaryCheckoutRoot(link)
	reason, isTemp := TempOriginReason(link)
	inside := pathInsideTempDir(link)
	if !inside || isTemp || inGit {
		t.Skipf("branch not reached on this host: inside=%v isTemp=%v (reason %q) git=%v", inside, isTemp, reason, inGit)
	}
	for _, tc := range []struct {
		name    string
		resolve func(string) string
	}{
		{"pure", ResolveTodoQueueRoot},
		{"adopting", ResolveTodoQueueRootAdopting},
	} {
		if root := tc.resolve(link); root != link {
			t.Errorf("%s resolver root = %q, want the launch base %q", tc.name, root, link)
		}
	}
}

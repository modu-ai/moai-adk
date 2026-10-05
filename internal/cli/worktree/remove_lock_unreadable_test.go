package worktree

// remove_lock_unreadable_test.go — the `remove` limb of the fail-closed lock
// read. When `git worktree list --porcelain` cannot be read, the lock state of
// the removal target is undetermined, and undetermined is never unlocked: the
// command refuses without --force (naming the unreadable source and the
// target) and proceeds with a warning under --force.

import (
	"errors"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/core/git"
)

// unreadableRemoveEnv installs a mock provider holding tree, isolates the
// session-registry lookup from the checkout the test runs in, and makes the
// lock listing fail. No registry anchors the tree, so the lock read is the
// only source that decides.
func unreadableRemoveEnv(t *testing.T, tree string) *mockWorktreeProvider {
	t.Helper()
	t.Setenv(config.EnvClaudeProjectDir, t.TempDir())

	origProvider := WorktreeProvider
	mock := &mockWorktreeProvider{worktrees: []git.Worktree{{Branch: "feature/lock-unreadable", Path: tree}}}
	WorktreeProvider = mock

	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "worktree" && args[1] == "list" {
			return "", errors.New("fatal: not a git repository")
		}
		return origGitCmd(args...)
	}

	origPrune := pruneLaunchLedgerFn
	pruneLaunchLedgerFn = func() ([]string, error) { return nil, nil }

	t.Cleanup(func() {
		WorktreeProvider = origProvider
		gitWorktreeCmd = origGitCmd
		pruneLaunchLedgerFn = origPrune
	})
	return mock
}

// TestRunRemove_UnreadableLockSourceRefusesWithoutForce: an unreadable lock
// source refuses the removal, names the cause and the target, and never
// reaches the provider's Remove.
func TestRunRemove_UnreadableLockSourceRefusesWithoutForce(t *testing.T) {
	tree := t.TempDir()
	mock := unreadableRemoveEnv(t, tree)

	_, err := runRemoveCmd(t, tree)
	if err == nil {
		t.Fatal("remove must refuse while the lock state of the target cannot be read")
	}
	msg := err.Error()
	for _, want := range []string{"ANCHORED_SESSIONS_PRESENT", "cause=" + causeLockSourceUnreadable, "source: lock", tree} {
		if !strings.Contains(msg, want) {
			t.Errorf("refusal must contain %q, got: %v", want, msg)
		}
	}
	if mock.removeCalled {
		t.Error("Remove() must not be called on refusal")
	}
}

// TestRunRemove_UnreadableLockSourceForceProceeds: --force removes the tree
// and warns that the anchor could not be determined.
func TestRunRemove_UnreadableLockSourceForceProceeds(t *testing.T) {
	tree := t.TempDir()
	mock := unreadableRemoveEnv(t, tree)

	combined, err := runRemoveCmd(t, "--force", tree)
	if err != nil {
		t.Fatalf("--force must remove despite the unreadable lock source, got error: %v", err)
	}
	if !mock.removeCalled {
		t.Error("Remove() must be called under --force")
	}
	if !strings.Contains(combined, "force removing") || !strings.Contains(combined, "cause="+causeLockSourceUnreadable) {
		t.Errorf("--force must warn naming the unreadable lock source, output: %q", combined)
	}
}

// stubReadableEmptyLockList makes `git worktree list --porcelain` read as an
// empty listing — readable, no tree locked — so a remove test decides on its
// own fixture instead of the git checkout (or its absence) the test binary
// happens to run in. Without it, a test run outside a git repository reads
// the lock source as unreadable and refuses the removal.
func stubReadableEmptyLockList(t *testing.T) {
	t.Helper()
	origGitCmd := gitWorktreeCmd
	gitWorktreeCmd = func(args ...string) (string, error) {
		if len(args) >= 2 && args[0] == "worktree" && args[1] == "list" {
			return "", nil
		}
		return origGitCmd(args...)
	}
	t.Cleanup(func() { gitWorktreeCmd = origGitCmd })
}

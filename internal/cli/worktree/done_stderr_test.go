package worktree

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/core/git"
)

// doneCmdHarness wires the done command against an injectable provider and
// returns the command plus its captured buffers.
func doneCmdHarness(t *testing.T, mock *mockWorktreeManager, args ...string) error {
	t.Helper()
	origProvider := WorktreeProvider
	WorktreeProvider = mock
	t.Cleanup(func() { WorktreeProvider = origProvider })

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs(args)
	return cmd.Execute()
}

// TestRunDone_RemoveFailureKeepsGitStderr (t41 a): when removal fails, the
// error the user sees must still contain git's own stderr text. The text
// survives the wrap chain; the silent-failure defect lived in the structural
// ExitCoder match (covered by internal/core/git exec_stderr_test.go), and
// this test pins the done-side wrap so the text cannot be dropped here.
func TestRunDone_RemoveFailureKeepsGitStderr(t *testing.T) {
	gitStderr := "fatal: cannot remove a locked working tree, lock reason: claude session (pid 4242)"

	mock := &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Branch: "feature/x", Path: "/tmp/wt-x"}}, nil
		},
		removeFunc: func(path string, force bool) error {
			return fmt.Errorf("remove worktree at %q: %w", path,
				&git.CommandError{Op: "worktree", Stderr: gitStderr, ExitStatus: 128})
		},
	}

	err := doneCmdHarness(t, mock, "feature/x")
	if err == nil {
		t.Fatal("done must fail when removal fails")
	}
	if !strings.Contains(err.Error(), "cannot remove a locked working tree") {
		t.Errorf("git stderr must survive the done wrap chain, got: %v", err)
	}
	if !strings.Contains(err.Error(), "claude session (pid 4242)") {
		t.Errorf("lock reason must stay visible, got: %v", err)
	}
}

// TestRunDone_LockedTreeGivesActionableGuidance (t41 b): when removal fails
// because the tree is locked, the error must carry actionable guidance
// (unlock hint + double-force escalation) naming the tree — and done must
// NOT silently escalate its own --force to -f -f: a lock usually means a
// live session still uses the tree.
func TestRunDone_LockedTreeGivesActionableGuidance(t *testing.T) {
	wtPath := "/tmp/wt-locked-guidance"
	forced := -1 // sentinel: not called

	mock := &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Branch: "feature/x", Path: wtPath}}, nil
		},
		removeFunc: func(path string, force bool) error {
			_ = path
			if force {
				forced = 1
			} else {
				forced = 0
			}
			// Shaped like the real worktreeManager.Remove lock mapping.
			return fmt.Errorf("remove worktree at %q: %w: git worktree: fatal: cannot remove a locked working tree, lock reason: claude session (pid 4242)",
				path, git.ErrWorktreeLocked)
		},
	}

	err := doneCmdHarness(t, mock, "feature/x")
	if err == nil {
		t.Fatal("done must fail on a locked tree")
	}
	msg := err.Error()
	for _, want := range []string{"git worktree unlock", "-f -f", wtPath} {
		if !strings.Contains(msg, want) {
			t.Errorf("lock guidance must mention %q, got: %v", want, msg)
		}
	}
	if forced != 0 {
		t.Errorf("done must pass the user's force flag through unchanged (no silent -f -f escalation), got forced=%d", forced)
	}
}

// TestRunDoneAuto_RemoveFailureIsNonZero (t41 c): --auto is output-silent,
// not exit-code-silent. A failed removal must surface as a command error so
// automation can detect it — measured behavior before the fix: exit 0 with
// no output, the error swallowed.
func TestRunDoneAuto_RemoveFailureIsNonZero(t *testing.T) {
	mock := &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Branch: "feature/x", Path: "/tmp/wt-x"}}, nil
		},
		removeFunc: func(path string, force bool) error {
			return fmt.Errorf("remove worktree at %q: %w", path,
				&git.CommandError{Op: "worktree", Stderr: "fatal: cannot remove a locked working tree", ExitStatus: 128})
		},
	}

	if err := doneCmdHarness(t, mock, "feature/x", "--auto"); err == nil {
		t.Fatal("--auto must not swallow a failed removal into exit 0")
	}
}

// TestRunDoneAuto_ListFailureIsNonZero (t41 c): same honesty for the listing
// step that locates the tree.
func TestRunDoneAuto_ListFailureIsNonZero(t *testing.T) {
	mock := &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return nil, fmt.Errorf("list worktrees: %w",
				&git.CommandError{Op: "worktree", Stderr: "fatal: not a git repository", ExitStatus: 128})
		},
	}

	if err := doneCmdHarness(t, mock, "feature/x", "--auto"); err == nil {
		t.Fatal("--auto must not swallow a failed listing into exit 0")
	}
}

// TestRunDoneAuto_NoWorktreeIsStillSuccess: guards the INTENDED grace of
// --auto (SPEC-WORKTREE-002 R2): a branch with no worktree left is a
// completed cleanup, not a failure. Honest exit codes must not regress this.
func TestRunDoneAuto_NoWorktreeIsStillSuccess(t *testing.T) {
	mock := &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Branch: "feature/other", Path: "/tmp/wt-other"}}, nil
		},
	}

	if err := doneCmdHarness(t, mock, "feature/x", "--auto"); err != nil {
		t.Fatalf("--auto with no matching worktree must stay a clean success, got: %v", err)
	}
}

// TestRunDoneAuto_SuccessStaysOutputSilent: --auto keeps its contract of
// not printing a success card — honesty applies to the exit code only.
func TestRunDoneAuto_SuccessStaysOutputSilent(t *testing.T) {
	origProvider := WorktreeProvider
	WorktreeProvider = &mockWorktreeManager{
		listFunc: func() ([]git.Worktree, error) {
			return []git.Worktree{{Branch: "feature/x", Path: "/tmp/wt-x"}}, nil
		},
	}
	t.Cleanup(func() { WorktreeProvider = origProvider })

	cmd := newDoneCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	cmd.SetArgs([]string{"feature/x", "--auto"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("--auto success path must stay error-free, got: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("--auto success must print nothing, got stdout: %q", out.String())
	}
}

// f3_f4_probe_test.go — RED probes for review findings F3 (detached-HEAD
// arm unreachable; git failures laundered into "no branch") and F4
// (readBranchStatus swallows every git-show failure as "no spec.md").
package kanban

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestReportedBranch_DetachedHEADReturnsEmptyCleanly — F3: a detached HEAD
// is `symbolic-ref --quiet` exiting 1 with EMPTY stderr (--quiet suppresses
// the message); reportedBranch must classify that as detachment and return
// ("", nil), reaching the genuine detached arm in ReadCardStatus — NOT the
// error arm, and NOT a laundered "no branch" from a tool failure.
func TestReportedBranch_DetachedHEADReturnsEmptyCleanly(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("detached-HEAD fixture uses posix git plumbing")
	}
	root := t.TempDir()
	runGitAt(t, root, "init", "-b", "main")
	runGitAt(t, root, "config", "user.email", "f3@example.com")
	runGitAt(t, root, "config", "user.name", "F3")
	if err := exec.Command("git", "-C", root, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
		t.Fatalf("empty commit: %v", err)
	}
	wt := filepath.Join(root, "detached")
	runGitAt(t, root, "worktree", "add", "--detach", wt, "main")
	t.Cleanup(func() { runGitAt(t, root, "worktree", "remove", "--force", wt) })

	branch, err := reportedBranch(wt)
	if err != nil {
		t.Fatalf("reportedBranch(detached) err = %v, want nil — a genuine detached HEAD must reach the empty-string arm cleanly, not the error arm", err)
	}
	if branch != "" {
		t.Fatalf("reportedBranch(detached) = %q, want %q (the detached-HEAD signal)", branch, "")
	}
}

// TestReportedBranch_GitToolFailureIsAnError — F3 (b): a genuine tool
// failure (missing git binary, simulated by swapping the branch probe to a
// command that does not exist) is propagated as an error, never laundered
// into the "" / detached signal. A missing git binary must not produce
// "worktree reports no branch".
func TestReportedBranch_GitToolFailureIsAnError(t *testing.T) {
	orig := branchProbe
	t.Cleanup(func() { branchProbe = orig })
	branchProbe = func(name string, args ...string) *exec.Cmd {
		return exec.Command("/this-binary-does-not-exist-anywhere")
	}
	if _, err := reportedBranch(t.TempDir()); err == nil {
		t.Fatal("reportedBranch(tool failure) err = nil, want non-nil — a tool failure must propagate, not become the detached signal")
	}
}

// TestReadBranchStatus_DeletedRefIsAnErrorNotNoFile — F4: a git-show failure
// for a reason OTHER than the genuine no-file case (here: a branch that does
// not exist / a deleted ref) must propagate as an error, not collapse into
// the affirmative SpecFilePresent=false "no spec.md on this branch" case —
// which pairingConsistent would then treat as legitimately pre-planning.
func TestReadBranchStatus_DeletedRefIsAnErrorNotNoFile(t *testing.T) {
	if runtimeIsWindows() {
		t.Skip("fixture uses posix git plumbing")
	}
	root := t.TempDir()
	runGitAt(t, root, "init", "-b", "main")
	runGitAt(t, root, "config", "user.email", "f4@example.com")
	runGitAt(t, root, "config", "user.name", "F4")
	if err := exec.Command("git", "-C", root, "commit", "--allow-empty", "-m", "init").Run(); err != nil {
		t.Fatalf("empty commit: %v", err)
	}
	// A branch that never existed: git show fails with exit 128, but NOT with
	// "path ... does not exist in ..." — the ref itself is missing.
	_, err := readBranchStatus(root, "SPEC-F4-001", "no-such-branch", "")
	if err == nil {
		t.Fatal("readBranchStatus(missing ref) err = nil, want non-nil — a ref failure must not collapse into the no-spec.md case")
	}
}

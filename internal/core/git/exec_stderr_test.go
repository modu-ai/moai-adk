package git

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// exitCodeCarrier mirrors the structural interface both cmd/moai/main.go
// (ExitCoder) and internal/cli/fang.go (exitCoder) match with errors.As:
// any error in the chain exposing `ExitCode() int`. *exec.ExitError
// satisfies it BY ACCIDENT, which is how a failed git subprocess used to
// leave `moai worktree done` SILENT while exiting with git's raw code
// (t41 incident a, 2026-08-15: rc=128 with 0 lines of stderr).
type exitCodeCarrier interface{ ExitCode() int }

// TestExecGit_FailureKeepsStderrVisible (t41 a): a failed git subprocess
// must keep its stderr in the returned error text, because the CLI surfaces
// whatever the error chain carries — an error without the stderr text gives
// the user nothing to act on.
func TestExecGit_FailureKeepsStderrVisible(t *testing.T) {
	// A directory with no .git makes `git status` exit 128 with a fatal line.
	dir := t.TempDir()

	_, err := execGit(context.Background(), dir, "status")
	if err == nil {
		t.Fatal("execGit on a non-repository must fail")
	}
	if !strings.Contains(err.Error(), "not a git repository") {
		t.Errorf("error must carry git's stderr verbatim, got: %v", err)
	}
}

// TestExecGit_FailureDoesNotSatisfyStructuralExitCoder (t41 a): the error
// returned for a failed git subprocess must NOT satisfy the structural
// ExitCoder interface — neither bare nor wrapped the way the worktree CLI
// wraps it. Matching it silences the error at both boundaries (fang's
// moaiErrorHandler stays quiet for ExitCoder carriers; main.go exits with
// the carrier's code), producing the measured "rc=128, 0 lines of stderr".
func TestExecGit_FailureDoesNotSatisfyStructuralExitCoder(t *testing.T) {
	dir := t.TempDir()

	_, err := execGit(context.Background(), dir, "status")
	if err == nil {
		t.Fatal("execGit on a non-repository must fail")
	}

	wrapped := fmt.Errorf("remove worktree: %w", err)
	for _, e := range []error{err, wrapped} {
		var ec exitCodeCarrier
		if errors.As(e, &ec) {
			t.Errorf("error chain must not satisfy the structural ExitCoder interface (it would print nothing and pass git's exit code through): %v", e)
		}
	}
}

// TestExecGit_DeadlineKillIsLegible (t41 b2): when a per-operation deadline
// ends a git invocation (the old 10s remove cap did this to large trees
// while the raw command, which has no deadline, succeeded), the error must
// NAME the deadline so a retry or the raw command is recognisable as the
// remedy — and must not satisfy the structural ExitCoder interface either.
// A pre-expired context exercises the never-started variant; a mid-execution
// kill renders "git killed: context deadline exceeded" and satisfies the
// same property.
func TestExecGit_DeadlineKillIsLegible(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 1)
	defer cancel()

	_, err := execGit(ctx, dir, "status")
	if err == nil {
		t.Fatal("execGit with an already-expired context must fail")
	}
	if !strings.Contains(err.Error(), "deadline exceeded") {
		t.Errorf("deadline failure must be named in the error, got: %v", err)
	}
	var ec exitCodeCarrier
	if errors.As(err, &ec) {
		t.Errorf("deadline-kill error must not satisfy the structural ExitCoder interface: %v", err)
	}
}

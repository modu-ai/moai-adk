package mergeq

// The typed errors of the merge queue, produced by real failures (a real merge conflict, a real gate command that
// exits non-zero, a real git step refused by a hook) and read the way a caller reads them: errors.Is and errors.As
// through Unwrap, and the text that reaches the ledger and the verb's output.

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// A merge that git refuses is a *MergeError: errors.Is(ErrMerge) holds, Unwrap reaches the *GitError and through it the
// *exec.ExitError of the real git process, the text names the failure and none of what git printed, and the output is
// kept in the error.
func TestMergeError_ARealConflictIsTypedAndUnwrapsToGitsExitStatus(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "x.txt", "from the card\n")
	advanceTrunk(t, dir, "x.txt", "from the trunk\n")
	g := GitCLI{Repo: dir, ScratchD: t.TempDir(), Who: Identity{Name: "mergeq", Email: "mergeq@example.test"}}

	_, err := g.PrepareCandidate(context.Background(), git(t, dir, "rev-parse", "trunk"), cand, "merge")
	if !errors.Is(err, ErrMerge) {
		t.Fatalf("%v, want errors.Is(ErrMerge)", err)
	}
	if errors.Is(err, ErrBadCandidate) {
		t.Errorf("%v: a merge conflict is not a bad candidate", err)
	}
	var me *MergeError
	if !errors.As(err, &me) {
		t.Fatalf("%T %v, want a *MergeError", err, err)
	}
	if !strings.Contains(string(me.Output), "CONFLICT") {
		t.Errorf("the error does not keep git's output: %q", me.Output)
	}
	if text := err.Error(); !strings.HasPrefix(text, ErrMerge.Error()) || strings.Contains(text, "CONFLICT") || strings.Contains(text, "x.txt") {
		t.Errorf("the text must name the failure and none of git's output: %q", text)
	}
	if me.Unwrap() != me.Cause {
		t.Errorf("Unwrap returned %v, want the cause %v", me.Unwrap(), me.Cause)
	}
	var ge *GitError
	if !errors.As(err, &ge) || ge.Sub != "merge" {
		t.Errorf("the chain holds no *GitError for the merge: %v", err)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() == 0 {
		t.Errorf("the chain does not reach git's *exec.ExitError: %v", err)
	}
}

// A gate that fails is a *GateError: its text names the gate id, the cause and how much it printed, and nothing it
// printed; Unwrap reaches the *exec.ExitError of the real command, so a caller can read its exit status.
func TestGateError_ARealFailingGateIsTypedAndUnwrapsToItsExitStatus(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the gate is an sh command")
	}
	verify := func(ctx context.Context, dir string) error {
		out := NewTailWriter(1024)
		cmd := exec.CommandContext(ctx, "sh", "-c", "echo gate-says-no; exit 3")
		cmd.Dir = dir
		cmd.Stdout, cmd.Stderr = out, out
		if err := cmd.Run(); err != nil {
			return &GateError{Gate: "gate-0123456789ab", Cause: err, Output: append([]byte(nil), out.Bytes()...), Total: out.Total()}
		}
		return nil
	}
	err := verify(context.Background(), t.TempDir())
	var ge *GateError
	if !errors.As(err, &ge) {
		t.Fatalf("%T %v, want a *GateError", err, err)
	}
	if want := `gate "gate-0123456789ab": exit status 3 (output 13 bytes)`; err.Error() != want {
		t.Errorf("text %q, want %q (no output in it)", err.Error(), want)
	}
	var ee *exec.ExitError
	if !errors.As(err, &ee) || ee.ExitCode() != 3 {
		t.Errorf("Unwrap does not reach the command's exit status 3: %v", err)
	}
	if ge.Unwrap() != ge.Cause {
		t.Errorf("Unwrap returned %v, want the cause %v", ge.Unwrap(), ge.Cause)
	}
	if string(ge.Output) != "gate-says-no\n" {
		t.Errorf("the output kept is %q", ge.Output)
	}

	// the same gate through Land: the failure is recorded and the reason carries the text of the error and no output
	cdir := repo(t)
	cand := candidate(t, cdir, "c1", "a.txt", "a\n")
	m := newMerger(t, cdir, &fakeQueue{}, verify)
	res, lerr := m.Land(context.Background(), "t1", cand)
	if lerr != nil || res.Outcome != OutcomeFailed {
		t.Fatalf("%+v %v", res, lerr)
	}
	if !strings.Contains(res.Reason, `gate "gate-0123456789ab": exit status 3`) || strings.Contains(res.Reason, "gate-says-no") {
		t.Errorf("reason %q", res.Reason)
	}
}

// A git step that is refused is a *GitError whose text holds no output; gitOutputOf reads the output from anywhere in
// a chain and says nothing for an error that holds no git failure.
func TestGitError_TheOutputIsReadFromAChainAndAnotherErrorHasNone(t *testing.T) {
	g := GitCLI{Repo: t.TempDir(), Who: Identity{Name: "t", Email: "t@example.test"}}
	_, err := g.RevParse(context.Background(), "refs/heads/nothing")
	if err == nil {
		t.Skip("an empty directory resolved a ref; git is configured oddly here")
	}
	wrapped := errors.Join(errors.New("context"), err)
	if out := gitOutputOf(wrapped); !strings.Contains(out, "not a git repository") {
		t.Errorf("the output of a git failure inside a chain: %q", out)
	}
	if out := gitOutputOf(errors.New("not a git failure")); out != "" {
		t.Errorf("an error with no git failure has output %q", out)
	}
	var ge *GitError
	if !errors.As(wrapped, &ge) || ge.Unwrap() != ge.Cause {
		t.Errorf("Unwrap of the git failure: %v", err)
	}
}

// PrepareCandidate refuses to start without a scratch directory, and when the scratch directory cannot be made.
func TestPrepareCandidate_ARefusedScratchDirectoryIsAnErrorBeforeAnyGitCall(t *testing.T) {
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	trunk := git(t, dir, "rev-parse", "trunk")
	who := Identity{Name: "mergeq", Email: "mergeq@example.test"}

	if _, err := (GitCLI{Repo: dir, Who: who}).PrepareCandidate(context.Background(), trunk, cand, "m"); err == nil {
		t.Error("a candidate was prepared with no scratch directory")
	}
	blocker := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := (GitCLI{Repo: dir, ScratchD: filepath.Join(blocker, "scratch"), Who: who}).PrepareCandidate(context.Background(), trunk, cand, "m"); err == nil {
		t.Error("a candidate was prepared with a scratch directory below a regular file")
	}
	if out := git(t, dir, "worktree", "list"); strings.Count(out, "\n") != 0 {
		t.Errorf("a worktree was registered by a refused call:\n%s", out)
	}
}

package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"charm.land/fang/v2"
	"github.com/modu-ai/moai-adk/internal/cli/worktree"
)

// The two guards in this file reproduce and pin the card-t130 defect class:
// a raw *exec.ExitError carried in a %w-wrapped error chain structurally
// satisfies the ExitCoder interface (`ExitCode() int`) that both CLI exit
// seams match with errors.As. That made cmd/moai/main.go adopt the
// subprocess's raw exit code (rc=128 from git) and internal/cli/fang.go's
// moaiErrorHandler suppress the styled error box — the silent-failure pair
// measured in the t41 (`moai worktree done`) and t129 (`moai hook
// worktree-create`) recurrences. Both guards assert on the errors.As chain,
// never on message strings.

// rawExitError runs a genuinely failing subprocess (the test binary itself,
// via the TestHelperProcess idiom) so the returned error is a real
// *exec.ExitError with ProcessState populated — not a hand-built struct,
// which would not exercise the same errors.As path.
func rawExitError(t *testing.T) *exec.ExitError {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestHelperProcess$")
	cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected *exec.ExitError from failing subprocess, got %T: %v", err, err)
	}
	return exitErr
}

// TestHelperProcess is not a real test; it is the child target rawExitError
// executes to produce a genuine non-zero process exit.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(3)
}

// exitCoderWrappingRaw is the adversarial shape: an intentional ExitCoder
// that %w-wraps a raw *exec.ExitError. The rejection must win over the
// intentional carrier — a producer that wraps a raw subprocess failure has
// not made a deliberate exit-code decision (no current producer does this).
type exitCoderWrappingRaw struct{ error }

func (e exitCoderWrappingRaw) ExitCode() int { return 7 }

// Unwrap makes errors.As descend into the wrapped chain, the way fmt.Errorf's
// %w wrapper does — without it the raw *exec.ExitError below is unreachable
// and the test would exercise nothing.
func (e exitCoderWrappingRaw) Unwrap() error { return e.error }

// TestRawExecExitErrorMustNotSteerExitCode: a wrapped raw *exec.ExitError
// reaching the main seam must map to the default exit 1, never adopt the
// subprocess's own code (3 here; 128 in the git incidents).
func TestRawExecExitErrorMustNotSteerExitCode(t *testing.T) {
	wrapped := fmt.Errorf("git worktree done: %w", rawExitError(t))
	if got := exitCodeForError(wrapped); got != 1 {
		t.Fatalf("exit code = %d, want 1 (raw *exec.ExitError in the chain must not steer the process exit code)", got)
	}
}

// TestMoaiErrorHandlerRendersWrappedExecExitError: the fang error handler
// must render the styled error box for a wrapped raw *exec.ExitError instead
// of staying silent (the 0-bytes-of-stderr face of the same defect).
func TestMoaiErrorHandlerRendersWrappedExecExitError(t *testing.T) {
	wrapped := fmt.Errorf("git worktree done: %w", rawExitError(t))
	var buf bytes.Buffer
	moaiErrorHandler(&buf, fang.Styles{}, wrapped)
	if buf.Len() == 0 {
		t.Fatal("moaiErrorHandler stayed silent for a wrapped raw *exec.ExitError — the silent-failure face of the card-t130 defect")
	}
}

// TestResolveExitCodeIntentionalStillResolves: the narrowing must not break
// the deliberate ExitCoder producers — direct, wrapped, and both the cli-root
// and worktree carriers still resolve their code.
func TestResolveExitCodeIntentionalStillResolves(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"cli_root_direct", &exitCodeError{code: 2, msg: "system error"}, 2},
		{"cli_root_wrapped", fmt.Errorf("spec audit: %w", &exitCodeError{code: 3, msg: "invalid args"}), 3},
		{"worktree_verify", &worktree.ExitCodeError{Code: 2}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, ok := ResolveExitCode(tc.err)
			if !ok || code != tc.wantCode {
				t.Fatalf("ResolveExitCode(%v) = (%d, %v); want (%d, true)", tc.err, code, ok, tc.wantCode)
			}
		})
	}
}

// TestResolveExitCodeRejectsRawInEveryShape: the raw type is refused bare,
// wrapped, and even alongside an intentional coder in the same chain.
func TestResolveExitCodeRejectsRawInEveryShape(t *testing.T) {
	raw := rawExitError(t)
	if _, ok := ResolveExitCode(raw); ok {
		t.Fatal("bare *exec.ExitError resolved as intentional")
	}
	if _, ok := ResolveExitCode(fmt.Errorf("wrap: %w", raw)); ok {
		t.Fatal("wrapped *exec.ExitError resolved as intentional")
	}
	both := exitCoderWrappingRaw{fmt.Errorf("inner: %w", raw)}
	if _, ok := ResolveExitCode(both); ok {
		t.Fatal("chain containing BOTH an intentional coder and a raw *exec.ExitError resolved as intentional")
	}
	if _, ok := ResolveExitCode(nil); ok {
		t.Fatal("nil resolved as intentional")
	}
	if code, ok := ResolveExitCode(errors.New("plain")); ok || code != 0 {
		t.Fatal("plain error resolved as intentional")
	}
}

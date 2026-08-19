package cli

import (
	"errors"
	"os/exec"
)

// ExitCoder lets a subcommand surface a non-default exit code through the
// cobra error chain. Used by `moai worktree verify` (0=clean, 1=divergence,
// 2=suspect, 3=both) and the cli-root exitCodeError family. It replaced the
// interface formerly declared in cmd/moai/main.go so the matching rule lives
// in one place.
type ExitCoder interface {
	ExitCode() int
}

// ResolveExitCode returns the intentional process exit code carried by err's
// chain: true and the code when a deliberate ExitCoder is reachable, false
// when there is none (the caller exits with its default, 1).
//
// It deliberately refuses raw *exec.ExitError FIRST. *exec.ExitError happens
// to satisfy the ExitCoder shape (`ExitCode() int`), so a %w-wrapped
// subprocess failure that reaches the CLI would otherwise be mistaken for an
// intentional exit code — cmd/moai/main.go would adopt the subprocess's raw
// code (git's 128) and fang's error handler would suppress the error text.
// Those are the two faces of the silent-failure defect that recurred as t41
// (`moai worktree done`, rc=128, 0 lines of stderr) and t129 (`moai hook
// worktree-create`, rc=128, 0 bytes of output). Only errors constructed on
// purpose — the exitCodeError family — may steer the process exit code; a
// wrapped subprocess failure is a genuine error and renders as one.
// The rejection is chain-wide: a chain carrying BOTH an intentional coder
// and a raw *exec.ExitError still resolves false, because a producer that
// wraps a raw subprocess failure has not made a deliberate exit-code
// decision. No current ExitCoder implementation wraps another error.
func ResolveExitCode(err error) (int, bool) {
	var raw *exec.ExitError
	if errors.As(err, &raw) {
		return 0, false
	}
	var ec ExitCoder
	if errors.As(err, &ec) {
		return ec.ExitCode(), true
	}
	return 0, false
}

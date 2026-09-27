//go:build !windows

package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// execOrSpawnClaude replaces the current process with the claude binary via
// syscall.Exec (execve(2)). On POSIX hosts this is the canonical launch path:
// the current shell process becomes claude, so no defer() runs after this call
// and the parent process identity is preserved.
//
// Because execve(2) replaces the process rather than forking one, this process's
// PID *is* the session's PID from the next instruction onward. That makes this
// the one place that knows the session PID outright, so it stamps
// MOAI_SESSION_PID into the launch environment: every hook subprocess the
// session later spawns inherits it, and the coordination registry records the
// live session PID without walking the process ancestry.
//
// REQ-CGH-001: syscall.Exec is POSIX-only. The Windows companion
// (launch_exec_windows.go) spawns a child and propagates its exit code instead,
// mirroring the reexecNewBinary pattern in update.go.
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	launchEnv := withSessionPID(env, os.Getpid())
	root := launchProjectRoot()
	pending, err := registerFactoryLaunchPending(context.Background(), root, launchEnv, os.Getpid(), homestate.CurrentProcessFingerprint())
	if err != nil {
		return err
	}
	if err := syscall.Exec(claudeBin, args, launchEnv); err != nil {
		rollbackErr := rollbackFactoryLaunchPending(context.Background(), root, pending)
		return fmt.Errorf("exec Claude: %w", errors.Join(err, rollbackErr))
	}
	return nil
}

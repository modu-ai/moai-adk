//go:build !windows

package cli

import (
	"os"
	"syscall"
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
	return syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))
}

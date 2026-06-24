//go:build !windows

package cli

import "syscall"

// execOrSpawnClaude replaces the current process with the claude binary via
// syscall.Exec (execve(2)). On POSIX hosts this is the canonical launch path:
// the current shell process becomes claude, so no defer() runs after this call
// and the parent process identity is preserved.
//
// REQ-CGH-001: syscall.Exec is POSIX-only. The Windows companion
// (launch_exec_windows.go) spawns a child and propagates its exit code instead,
// mirroring the reexecNewBinary pattern in update.go.
func execOrSpawnClaude(claudeBin string, args, env []string) error {
	return syscall.Exec(claudeBin, args, env)
}

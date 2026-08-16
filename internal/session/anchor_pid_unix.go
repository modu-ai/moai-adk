//go:build !windows

// anchor_pid_unix.go — process-liveness probe for the anchor guard, in the
// same shape as internal/cli/update_cleanup_unix.go (signal 0 semantics).
package session

import (
	"errors"

	"golang.org/x/sys/unix"
)

// isProcessAlive reports whether a process with the given PID exists.
// Signal 0 probes existence without delivering a signal: EPERM means the
// process exists under another owner, ESRCH means it is gone.
func isProcessAlive(pid int) bool {
	err := unix.Kill(pid, 0)
	if err == nil {
		return true
	}
	if errors.Is(err, unix.EPERM) {
		return true
	}
	return false
}

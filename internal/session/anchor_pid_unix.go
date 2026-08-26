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

// probeProcessLiveness is the two-valued liveness probe backing the shared
// anchor decision (SPEC-WORKTREE-REAPER-001 REQ-WR-008, design.md §B.5). It
// wraps the same signal-0 syscall as isProcessAlive but keeps "I do not know"
// distinguishable from "dead", which the bare-bool form collapses.
//
//	nil    → (true, true)   the process exists
//	EPERM  → (true, true)   it exists under another owner
//	ESRCH  → (false, true)  it is positively gone
//	other  → (false, false) UNDETERMINED — the caller must fail closed
func probeProcessLiveness(pid int) (alive bool, determined bool) {
	err := unix.Kill(pid, 0)
	switch {
	case err == nil:
		return true, true
	case errors.Is(err, unix.EPERM):
		return true, true
	case errors.Is(err, unix.ESRCH):
		return false, true
	default:
		return false, false
	}
}

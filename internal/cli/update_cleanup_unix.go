//go:build !windows

// Package cli — update_cleanup_unix.go
//
// Unix-specific (linux/darwin/freebsd/openbsd) process-liveness probe for
// SPEC-V3R3-UPDATE-CLEANUP-001 stale-lock detection.
//
// SPEC-V2.20.0-RC1 hotfix: extracted from update_cleanup.go because the
// previous single-file implementation referenced syscall.Kill behind a
// `runtime.GOOS == "windows"` runtime guard, which the Go compiler still
// rejects at compile time on the Windows target.

package cli

import (
	"errors"
	"syscall"
)

// isProcessAlive reports whether a process with the given PID is alive.
// Uses syscall.Kill(pid, 0) — signal 0 checks existence without killing.
// EPERM means the process exists but is owned by another user — still alive.
// ESRCH means the process does not exist.
func isProcessAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	if err == nil {
		return true
	}
	if errors.Is(err, syscall.EPERM) {
		return true
	}
	// syscall.ESRCH (or any other error): treat as stale.
	return false
}

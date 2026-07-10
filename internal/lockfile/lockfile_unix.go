//go:build !windows

// Package lockfile provides cross-platform advisory file-lock primitives.
//
// The Unix implementation uses syscall.Flock (kernel advisory locks, safe
// across processes). The Windows implementation falls back to in-process
// per-path mutexes — see lockfile_windows.go for the documented cross-process
// limitation. Migrated from internal/cli team_spawn_lock_{unix,windows}.go
// (SPEC-AGENT-TEAM-RETIRE-001 M0, REQ-ATR-001) with behavior preserved.
package lockfile

import (
	"os"
	"syscall"
)

// @MX:ANCHOR: [AUTO] Lock/Unlock exported advisory-lock API — invariant contract
// @MX:REASON: fan_in >= 3: internal/cli/taskledger.ClaimTask, internal/cli/settings.go,
// internal/cli/glm_tools.go all serialize concurrent file writes through this pair;
// advisory-lock semantics (blocking exclusive lock, unlock on same fd) must be preserved.

// Lock acquires an exclusive flock on the given file (Unix only).
func Lock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// Unlock releases the flock on the given file (Unix only).
func Unlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

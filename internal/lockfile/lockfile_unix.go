//go:build !windows

// Package lockfile provides cross-platform advisory file-lock primitives.
//
// The Unix implementation uses syscall.Flock (kernel advisory locks, safe
// across processes). Windows uses an OS-backed exclusive byte-range lock
// for the same blocking/nonblocking acquisition and descriptor lifetime.
// Migrated from internal/cli team_spawn_lock_{unix,windows}.go and extended
// for the cross-process writers of SPEC-WORKTREE-HOME-ROOT-001.
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

// TryLock acquires an exclusive flock on the given file without blocking,
// returning an error when another holder owns it (SPEC-WORKTREE-HOME-ROOT-001
// M5, design.md D-12: the state-file lock is try-and-bounded-wait, never
// blocking). Nothing else about the lockfile surface changes.
func TryLock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
}

// Unlock releases the flock on the given file (Unix only).
func Unlock(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

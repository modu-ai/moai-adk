//go:build !windows

// Unix advisory lock implementation for observation append race protection.
// Uses flock(2) — POSIX-standard advisory file lock.

package capture

import (
	"os"
	"syscall"
)

// acquireExclusiveLock takes an exclusive advisory lock on the file descriptor.
// Best-effort: failures are non-fatal (continue without lock).
func acquireExclusiveLock(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// releaseLock releases any advisory lock held on the file descriptor.
func releaseLock(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

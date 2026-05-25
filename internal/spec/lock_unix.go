//go:build !windows

// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Unix per-SPEC close lock (flock-based).
// Mirrors the pattern from internal/session/registry_lock_unix.go for consistency.
// Per CLAUDE.local.md §14: no naked syscall.Flock in spec.go body; abstraction lives here.
package spec

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// flockSpecLock holds an open file descriptor with flock(LOCK_EX|LOCK_NB) held.
type flockSpecLock struct {
	fd int
}

func (f *flockSpecLock) release() error {
	if f == nil || f.fd == 0 {
		return nil
	}
	// Close releases the flock atomically.
	err := unix.Close(f.fd)
	f.fd = 0
	return err
}

// acquireSpecCloseLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock. Returns ErrSpecCloseLockHeld on contention.
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
	}
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return nil, ErrSpecCloseLockHeld
	}
	return &flockSpecLock{fd: fd}, nil
}

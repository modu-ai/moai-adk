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
	// Unlock before closing rather than relying on close-time release
	// visibility to the next acquirer. Keep contention nonblocking: an
	// EWOULDBLOCK from acquire must report the holder seen at that attempt.
	var unlockErr error
	for {
		unlockErr = unix.Flock(f.fd, unix.LOCK_UN)
		if unlockErr != unix.EINTR {
			break
		}
	}
	closeErr := unix.Close(f.fd)
	f.fd = 0
	if unlockErr != nil {
		return fmt.Errorf("unlock spec-close lock: %w", unlockErr)
	}
	return closeErr
}

// flockAttempt is the seam the lock tests use to inject flock(2) outcomes
// (EINTR, foreign errnos) deterministically. It is unix.Flock in production.
var flockAttempt = unix.Flock

// acquireSpecCloseLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock. A signal-interrupted attempt is retried;
// EWOULDBLOCK reports the current holder immediately; any other errno is
// wrapped and returned, never mislabeled as a held lock (card t1291).
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
	}
	for {
		err := flockAttempt(fd, unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			return &flockSpecLock{fd: fd}, nil
		}
		switch err {
		case unix.EINTR:
			// The runtime's preemption signals reach raw syscalls as EINTR,
			// and heavy parallel test load widens that window. Retrying is
			// safe: LOCK_NB never blocks, so the retry re-runs the same
			// atomic grant check and cannot admit a second holder the first
			// attempt excluded.
			continue
		case unix.EWOULDBLOCK:
			_ = unix.Close(fd)
			return nil, ErrSpecCloseLockHeld
		default:
			_ = unix.Close(fd)
			return nil, fmt.Errorf("flock spec-close lock %s: %w", lockPath, err)
		}
	}
}

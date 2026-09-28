//go:build !windows

// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Unix per-SPEC close lock (flock-based).
// Mirrors the pattern from internal/session/registry_lock_unix.go for consistency.
// Per CLAUDE.local.md §14: no naked syscall.Flock in spec.go body; abstraction lives here.
package spec

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

const (
	// Transient recheck budget after EWOULDBLOCK: close()'s return can lead
	// the flock teardown's visibility by a sub-millisecond window, so an
	// immediately following acquire may observe EWOULDBLOCK with no live
	// holder (card t1291 probe: first=held, recheck-after-2ms=nil on a
	// contention-free path under parallel signal pressure). A genuine holder
	// keeps the lock for a whole close operation — far beyond this budget —
	// so the recheck cannot absorb real contention. Mirrors the transient
	// class in lock_windows.go (lockTransientRetries / lockTransientDelay).
	specLockTransientRetries = 10
	specLockTransientDelay   = 5 * time.Millisecond
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

// flockAttempt is the seam the lock tests use to inject flock(2) outcomes
// (EINTR, foreign errnos) deterministically. It is unix.Flock in production.
var flockAttempt = unix.Flock

// acquireSpecCloseLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock. Failure classes mirror the Windows twin
// (lock_windows.go): a signal-interrupted attempt is retried immediately;
// EWOULDBLOCK gets a bounded transient recheck (a just-released lock's
// teardown can lag its close()'s return) before being reported as
// ErrSpecCloseLockHeld; any other errno is wrapped and returned — never
// mislabeled as the lock being held, which would make the closer refuse a
// legitimate close (card t1291).
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
	}
	transientLeft := specLockTransientRetries
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
			if transientLeft > 0 {
				transientLeft--
				time.Sleep(specLockTransientDelay)
				continue
			}
			_ = unix.Close(fd)
			return nil, ErrSpecCloseLockHeld
		default:
			_ = unix.Close(fd)
			return nil, fmt.Errorf("flock spec-close lock %s: %w", lockPath, err)
		}
	}
}

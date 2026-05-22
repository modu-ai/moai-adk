//go:build !windows

package session

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

// fileLock is the cross-platform advisory-lock interface.
// SPEC-V3R2-RT-004 REQ-040: lock primitive that prevents concurrent checkpoint writes.
// @MX:ANCHOR: [AUTO] SPEC-V3R2-RT-004 REQ-040 cross-platform lock contract
// @MX:REASON: lock_unix.go and lock_windows.go are the two implementations. New platforms add a new file with the same interface.
type fileLock interface {
	acquire(path string, retries int, backoff time.Duration) error
	release() error
}

// newFileLock returns the platform-specific lock implementation.
// SPEC-V3R2-RT-004 plan.md §3.2: uses Unix flock (Windows is implemented separately).
func newFileLock() fileLock {
	return &flockLock{}
}

// flockLock is the Unix flock advisory-lock implementation.
// SPEC-V3R2-RT-004 REQ-040: uses syscall.Flock(LOCK_EX|LOCK_NB).
type flockLock struct {
	mu     sync.Mutex
	fd     int
	active bool
}

// acquire obtains an exclusive lock via flock(LOCK_EX|LOCK_NB).
// Lock companion file path: <checkpoint-path>.lock
func (l *flockLock) acquire(path string, retries int, backoff time.Duration) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Lock companion file path
	lockPath := path + ".lock"

	// Open the file (creating it if needed)
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0644)
	if err != nil {
		return fmt.Errorf("open lock file %s: %w", lockPath, err)
	}

	l.fd = fd

	// Try the non-blocking exclusive lock
	err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
	if err != nil {
		_ = unix.Close(fd)
		l.fd = 0
		return fmt.Errorf("flock %s: %w", lockPath, err)
	}

	l.active = true
	return nil
}

// release releases the flock and closes the lock file.
func (l *flockLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.active {
		return nil // already released
	}

	if l.fd != 0 {
		// Flock unlock (implicit on close)
		err := unix.Close(l.fd)
		l.fd = 0
		l.active = false
		return err
	}

	return nil
}

// SPEC-V3R2-RT-004: acquireWithRetry obtains the lock using a 3-retry /
// 10ms-backoff policy.
// REQ-040: returns ErrCheckpointConcurrent on repeated loss.
func acquireWithRetry(lock fileLock, path string, retries int, backoff time.Duration) error {
	var lastErr error
	for i := 0; i < retries; i++ {
		err := lock.acquire(path, retries, backoff)
		if err == nil {
			return nil // success
		}
		lastErr = err
		// Backoff wait
		time.Sleep(backoff)
	}
	return fmt.Errorf("%w: after %d retries: %v", ErrCheckpointConcurrent, retries, lastErr)
}

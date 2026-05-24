//go:build !windows

package session

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

// registryLock is the unix flock-based advisory lock used exclusively by
// the multi-session coordination registry (active-sessions.json.lock).
// Distinct from flockLock in lock.go which serves SPEC-V3R2-RT-004 checkpoints.
//
// SPEC-V3R6-MULTI-SESSION-COORD-001 REQ-COORD-008, REQ-COORD-022.
type registryLock struct {
	mu sync.Mutex
	fd int
}

// newRegistryLock returns a fresh registryLock instance.
func newRegistryLock() *registryLock {
	return &registryLock{}
}

// acquire opens the lock companion file at lockPath and applies a
// non-blocking exclusive flock. Returns an error on contention so the
// caller (withLock) can retry within the LockTimeout window.
func (l *registryLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return fmt.Errorf("registry lock open %s: %w", lockPath, err)
	}

	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return fmt.Errorf("registry lock flock %s: %w", lockPath, err)
	}

	l.fd = fd
	return nil
}

// release releases the flock and closes the underlying fd. Idempotent.
func (l *registryLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fd == 0 {
		return nil
	}
	err := unix.Close(l.fd) // close releases the flock
	l.fd = 0
	return err
}

// _ ensures os is referenced under conditional builds (some sub-toolchains
// otherwise drop the import). Compile-only sentinel.
var _ = os.Getpid

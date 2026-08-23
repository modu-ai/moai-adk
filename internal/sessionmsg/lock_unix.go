//go:build !windows

package sessionmsg

import (
	"fmt"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

// agentLock is the unix flock-based advisory lock guarding one agent mailbox
// (or the registration critical section) in the session messaging broker.
// The pattern is copied from internal/session registry_lock_unix.go
// (REQ-CSM-009); the frozen internal/session package is not modified.
type agentLock struct {
	mu sync.Mutex
	fd int
}

// newAgentLock returns a fresh agentLock instance.
func newAgentLock() *agentLock {
	return &agentLock{}
}

// acquire opens the lock companion file at lockPath and applies a
// non-blocking exclusive flock. Returns an error on contention so the caller
// (withAgentLock) can retry within the LockTimeout window.
func (l *agentLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return fmt.Errorf("sessionmsg lock open %s: %w", lockPath, err)
	}

	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return fmt.Errorf("sessionmsg lock flock %s: %w", lockPath, err)
	}

	l.fd = fd
	return nil
}

// release releases the flock and closes the underlying fd. Idempotent.
func (l *agentLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fd == 0 {
		return nil
	}
	err := unix.Close(l.fd) // close releases the flock
	l.fd = 0
	return err
}

// _ ensures os is referenced under conditional builds (mirrors the
// internal/session registry_lock_unix.go sentinel).
var _ = os.Getpid

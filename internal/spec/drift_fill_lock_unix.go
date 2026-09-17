//go:build !windows

package spec

import (
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
)

// driftFillLock is the unix flock-based advisory lock guarding the out-of-band
// drift-cache fill. The pattern is copied from internal/sessionmsg
// lock_unix.go (itself copied from internal/session registry_lock_unix.go);
// neither of those packages is modified.
//
// The flock is associated with the OPEN FILE DESCRIPTION, so the kernel
// releases it when the descriptor is closed — which includes the implicit close
// performed when the process dies, SIGKILL included. That is REQ-DCF-008's
// "released by process exit, including abnormal termination".
type driftFillLock struct {
	mu sync.Mutex
	fd int
}

// unacquiredDriftFillFD is the sentinel for "this lock holds no descriptor". It
// must NOT be 0: fd 0 is a valid descriptor the kernel hands out whenever stdin
// is closed — which is exactly the state of the detached fill child — and using
// it as the sentinel would make release() a silent no-op in that case.
const unacquiredDriftFillFD = -1

func newDriftFillLock() *driftFillLock {
	return &driftFillLock{fd: unacquiredDriftFillFD}
}

// acquire opens the lock file and applies a NON-BLOCKING exclusive flock.
// Contention returns an error; it is never retried.
func (l *driftFillLock) acquire(lockPath string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return fmt.Errorf("drift fill lock open %s: %w", lockPath, err)
	}

	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return fmt.Errorf("drift fill lock flock %s: %w", lockPath, err)
	}

	l.fd = fd
	return nil
}

// release closes the descriptor, which releases the flock. Idempotent.
func (l *driftFillLock) release() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.fd < 0 {
		return nil
	}
	err := unix.Close(l.fd) // close releases the flock
	l.fd = unacquiredDriftFillFD
	return err
}

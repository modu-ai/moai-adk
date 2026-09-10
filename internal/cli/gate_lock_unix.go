//go:build !windows

package cli

// gate_lock_unix.go — Unix substrate of the gate-run lock: flock(2) on an
// open descriptor, mirroring internal/kanban/board_lock_unix.go's pattern.
// The kernel releases the flock when the descriptor closes, which it does on
// process exit — so a killed holder leaves an artifact that blocks nothing on
// this platform; the stale-lock clear exists for the Windows substrate, where
// the artifact IS the lock.

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// flockGateLock holds an open descriptor with flock(LOCK_EX|LOCK_NB) held.
type flockGateLock struct {
	fd int
}

func (f *flockGateLock) release() error {
	if f == nil || f.fd == 0 {
		return nil
	}
	// Close releases the flock atomically.
	err := unix.Close(f.fd)
	f.fd = 0
	return err
}

// acquireGateLockImpl opens lockPath O_CREAT|O_RDWR and applies a
// non-blocking exclusive flock, then records this process's identity IN the
// artifact. Only the lock holder writes, so the recorded identity always
// names the current owner; a contender that fails the flock writes nothing
// and the previous owner's record stands.
func acquireGateLockImpl(lockPath string) (gateLockImpl, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open gate-run lock %s: %w", lockPath, err)
	}
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		return nil, ErrGateLockHeld
	}

	record := newGateLockOwnerRecord()
	if err := unix.Ftruncate(fd, 0); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("record gate-run lock owner (truncate) %s: %w", lockPath, err)
	}
	if _, err := unix.Write(fd, record); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("record gate-run lock owner %s: %w", lockPath, err)
	}
	return &flockGateLock{fd: fd}, nil
}

// ClearStaleGateLock on Unix is a no-op reporting the platform gate: the
// stale-clear window does not arise here. The lock artifact is inert and a
// subsequent AcquireGateLock succeeds by taking a fresh flock on the same
// path.
func ClearStaleGateLock(projectDir string) (*GateLockClearReport, error) {
	return &GateLockClearReport{
		Removed: false,
		Reason:  "stale-lock clear is gated to windows; the unix substrate releases flock on process exit and an orphaned artifact blocks nothing here",
	}, nil
}

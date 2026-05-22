//go:build !windows

package migration

import "golang.org/x/sys/unix"

// lockHandle holds the file descriptor of a Unix advisory lock.
type lockHandle struct{ fd int }

// acquireLock acquires an exclusive advisory lock on lockPath.
// Uses unix.Flock(LOCK_EX), which blocks until any other process holding
// the lock releases it (REQ-V3R2-RT-007-031).
func acquireLock(lockPath string) (*lockHandle, error) {
	fd, err := unix.Open(lockPath, unix.O_CREAT|unix.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}

	return &lockHandle{fd: fd}, nil
}

// releaseLock releases the advisory lock and closes the file descriptor.
func releaseLock(h *lockHandle) error {
	if err := unix.Flock(h.fd, unix.LOCK_UN); err != nil {
		_ = unix.Close(h.fd)
		return err
	}
	return unix.Close(h.fd)
}

//go:build !windows

package homestate

import "golang.org/x/sys/unix"

type admissionLockImpl interface{ release() error }
type unixAdmissionLock struct{ fd int }

func acquireAdmissionLock(path string) (admissionLockImpl, error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(fd, unix.LOCK_EX); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	return &unixAdmissionLock{fd: fd}, nil
}

// tryAcquireAdmissionLock is acquireAdmissionLock without the wait: ok is false
// (and nothing is held) when another descriptor holds the lock.
func tryAcquireAdmissionLock(path string) (impl admissionLockImpl, ok bool, err error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_RDWR|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, false, err
	}
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		if err == unix.EWOULDBLOCK {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &unixAdmissionLock{fd: fd}, true, nil
}

func (l *unixAdmissionLock) release() error {
	if l.fd < 0 {
		return nil
	}
	fd := l.fd
	l.fd = -1
	return unix.Close(fd)
}

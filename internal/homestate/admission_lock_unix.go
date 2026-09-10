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
func (l *unixAdmissionLock) release() error {
	if l.fd < 0 {
		return nil
	}
	fd := l.fd
	l.fd = -1
	return unix.Close(fd)
}

//go:build windows

package homestate

import (
	"golang.org/x/sys/windows"
	"os"
)

type admissionLockImpl interface{ release() error }
type windowsAdmissionLock struct {
	file       *os.File
	overlapped windows.Overlapped
}

func acquireAdmissionLock(path string) (admissionLockImpl, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	l := &windowsAdmissionLock{file: f}
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK, 0, 1, 0, &l.overlapped); err != nil {
		_ = f.Close()
		return nil, err
	}
	return l, nil
}

// tryAcquireAdmissionLock is acquireAdmissionLock without the wait: ok is false
// (and nothing is held) when another handle holds the lock.
func tryAcquireAdmissionLock(path string) (impl admissionLockImpl, ok bool, err error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, false, err
	}
	l := &windowsAdmissionLock{file: f}
	if err := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &l.overlapped); err != nil {
		_ = f.Close()
		if err == windows.ERROR_LOCK_VIOLATION {
			return nil, false, nil
		}
		return nil, false, err
	}
	return l, true, nil
}

func (l *windowsAdmissionLock) release() error {
	if l.file == nil {
		return nil
	}
	_ = windows.UnlockFileEx(windows.Handle(l.file.Fd()), 0, 1, 0, &l.overlapped)
	err := l.file.Close()
	l.file = nil
	return err
}

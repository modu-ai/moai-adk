//go:build !windows

package toolpolicy

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func openRecordGuard(path string) (*recordGuard, error) {
	fd, err := unix.Open(path, unix.O_CREAT|unix.O_EXCL|unix.O_RDWR|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0o600)
	if errors.Is(err, unix.EEXIST) {
		fd, err = unix.Open(path, unix.O_RDWR|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	}
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	id, err := inspectRecordGuard(f, path)
	if err != nil {
		return nil, errors.Join(err, f.Close())
	}
	return &recordGuard{file: f, path: path, id: id}, nil
}

func inspectRecordGuard(f *os.File, path string) (string, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(int(f.Fd()), &stat); err != nil {
		return "", err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Nlink != 1 || stat.Uid != uint32(os.Geteuid()) || stat.Mode&0o7777 != 0o600 {
		return "", errors.New("record guard must be a private current-owner regular single-link file")
	}
	if stat.Ino == 0 {
		return "", errors.New("record guard file identity is unavailable")
	}
	held, err := f.Stat()
	if err != nil {
		return "", err
	}
	current, err := os.Lstat(path)
	if err != nil {
		return "", err
	}
	if !os.SameFile(held, current) {
		return "", errors.New("persistent record guard path was replaced")
	}
	return fmt.Sprintf("%x:%x", stat.Dev, stat.Ino), nil
}

func tryRecordGuardLock(f *os.File) error {
	err := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return errRecordGuardBusy
	}
	return err
}

func unlockRecordGuard(f *os.File) error {
	return unix.Flock(int(f.Fd()), unix.LOCK_UN)
}

func openRecordObservation(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_CLOEXEC|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, err
	}
	f := os.NewFile(uintptr(fd), path)
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return nil, errors.Join(err, f.Close())
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Nlink != 1 {
		return nil, errors.Join(errors.New("record lock is not a regular single-link file"), f.Close())
	}
	return f, nil
}

func createRecordPrimary(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
}

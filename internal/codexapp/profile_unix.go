//go:build !windows

package codexapp

import (
	"os"
	"syscall"
)

func privateOwner(_ string, info os.FileInfo) bool {
	stat, ok := info.Sys().(*syscall.Stat_t)
	return ok && stat.Uid == uint32(os.Geteuid()) && info.Mode().Perm()&0077 == 0
}

func profileLease(path string) (*os.File, error) {
	fd, err := syscall.Open(path, syscall.O_RDWR|syscall.O_CREAT|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), path)
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || !privateOwner(path, info) {
		_ = file.Close() // lease rejected; the descriptor is being discarded
		return nil, os.ErrPermission
	}
	if err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close() // lock unavailable; the descriptor is being discarded
		return nil, err
	}
	return file, nil
}

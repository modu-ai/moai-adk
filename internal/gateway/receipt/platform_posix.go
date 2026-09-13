//go:build !windows

package receipt

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"syscall"
	"time"
)

func platformSupported() error { return nil }
func safeOpenFlags() int       { return unix.O_NOFOLLOW | unix.O_NONBLOCK }
func private(i os.FileInfo, dir bool) bool {
	if i == nil || i.Mode().Perm()&0077 != 0 {
		return false
	}
	s, ok := i.Sys().(*syscall.Stat_t)
	if !ok || s.Uid != uint32(os.Geteuid()) {
		return false
	}
	if dir {
		return i.IsDir()
	}
	return i.Mode().IsRegular() && s.Nlink == 1
}
func lock(ctx context.Context, f *os.File) error {
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		e := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if e == nil {
			return nil
		}
		if !errors.Is(e, unix.EWOULDBLOCK) && !errors.Is(e, unix.EAGAIN) {
			return ErrState
		}
		timer := time.NewTimer(10 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}
func unlock(f *os.File) { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }

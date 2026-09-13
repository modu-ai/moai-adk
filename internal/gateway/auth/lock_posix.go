//go:build !windows

package auth

import (
	"context"
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"time"
)

func lockFile(ctx context.Context, f *os.File) error {
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		e := unix.Flock(int(f.Fd()), unix.LOCK_EX|unix.LOCK_NB)
		if e == nil {
			return nil
		}
		if !errors.Is(e, unix.EWOULDBLOCK) && !errors.Is(e, unix.EAGAIN) {
			return ErrAuthState
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func unlockFile(f *os.File) { _ = unix.Flock(int(f.Fd()), unix.LOCK_UN) }
func syncDirectory(path string) error {
	f, e := os.Open(path)
	if e != nil {
		return e
	}
	defer f.Close()
	return f.Sync()
}

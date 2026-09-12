//go:build windows

package auth

import (
	"context"
	"errors"
	"golang.org/x/sys/windows"
	"os"
	"time"
)

func lockFile(ctx context.Context, f *os.File) error {
	for {
		if e := ctx.Err(); e != nil {
			return e
		}
		e := windows.LockFileEx(windows.Handle(f.Fd()), windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, &windows.Overlapped{})
		if e == nil {
			return nil
		}
		if !errors.Is(e, windows.ERROR_LOCK_VIOLATION) {
			return ErrAuthState
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
		}
	}
}
func unlockFile(f *os.File) {
	_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, &windows.Overlapped{})
}

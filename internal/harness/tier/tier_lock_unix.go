//go:build !windows

// Unix advisory lock implementation for tier state mutation protection.
// Uses flock(2) on the observations.yaml file descriptor.

package tier

import (
	"os"
	"syscall"
)

func acquireExclusiveLock(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

func releaseLock(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

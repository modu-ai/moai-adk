//go:build !windows

package cli

import (
	"os"
	"syscall"
)

// lockFile acquires an exclusive flock on the given file (Unix only).
func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX)
}

// unlockFile releases the flock on the given file (Unix only).
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}

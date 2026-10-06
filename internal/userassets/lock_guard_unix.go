//go:build !windows

package userassets

import (
	"os"
	"syscall"
)

// guardFd is the OS-level serialization guard for the whole
// reclaim-then-acquire sequence (review-fix round 2, F5/A3/B2): a stale
// reclaim via rename(2) is atomic for the reclaim itself, but two racing
// callers can still both proceed to create and believe they own the lock.
// The guard is a SECOND file held with flock(2) for the entire acquire
// attempt, so reclaim+create runs strictly serially per user.
var guardFd *os.File

func acquireGuard(path string) error {
	fd, err := os.OpenFile(path+".guard", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	if err := syscall.Flock(int(fd.Fd()), syscall.LOCK_EX); err != nil {
		_ = fd.Close()
		return err
	}
	guardFd = fd
	return nil
}

func releaseGuard() {
	if guardFd != nil {
		_ = syscall.Flock(int(guardFd.Fd()), syscall.LOCK_UN)
		_ = guardFd.Close()
		guardFd = nil
	}
}

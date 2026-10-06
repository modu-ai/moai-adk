//go:build windows

package hygiene

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// acquirePassLock takes the rotator's cross-process sidecar lock on
// Windows: a LockFileEx exclusive, fail-immediately lock on
// <logs>/.hygiene-rotation.lock. The in-process mutex is not cross-process
// (D17): rotation proceeds only when this sidecar lock is taken AND
// verified; contention reads skipped-locked, an unestablishable sidecar
// reads skipped-platform, and a chunk is never destroyed on a path whose
// exclusivity is unverified (REQ-HYG-002).
func (r *Rotator) acquirePassLock() (func(), lockResult) {
	if r.lockProbe != nil {
		return func() {}, r.lockProbe(r.LogDir)
	}
	if err := os.MkdirAll(r.LogDir, 0o755); err != nil {
		return func() {}, lockUnverifiable
	}
	f, err := os.OpenFile(filepath.Join(r.LogDir, PassLockName), os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return func() {}, lockUnverifiable
	}
	ol := new(windows.Overlapped)
	err = windows.LockFileEx(windows.Handle(f.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, ol)
	if err != nil {
		f.Close()
		if isLockViolation(err) {
			return func() {}, lockHeld
		}
		return func() {}, lockUnverifiable
	}
	release := func() {
		_ = windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, ol)
		_ = f.Close()
	}
	return release, lockAcquired
}

// isLockViolation reports whether err is the LockFileEx contention shape
// (ERROR_LOCK_VIOLATION, win32 error 33).
func isLockViolation(err error) bool {
	errno, ok := err.(windows.Errno)
	return ok && errno == 33 // ERROR_LOCK_VIOLATION; not exported by x/sys
}

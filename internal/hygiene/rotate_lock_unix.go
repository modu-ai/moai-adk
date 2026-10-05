//go:build !windows

package hygiene

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
)

// acquirePassLock takes the rotator's cross-process pass lock
// (<logs>/.hygiene-rotation.lock, advisory flock, non-blocking) and
// returns the release func plus the outcome. A held lock is contention
// (skipped-locked); an unopenable lock file leaves exclusion unverifiable
// (skipped-platform) — rotation never proceeds on unverified exclusivity
// (REQ-HYG-002).
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EACCES) {
			return func() {}, lockHeld
		}
		return func() {}, lockUnverifiable
	}
	release := func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}
	return release, lockAcquired
}

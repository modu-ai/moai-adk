//go:build unix

package jev

// cache_lock_unix.go — P2 concurrency (card t1968 codex review round 9,
// operator-approved repair): updateAliasMap's read-modify-write must
// serialize across PROCESSES, not just goroutines — the cache dir is shared
// by every jev client on the machine, and file-replace atomicity alone does
// not stop two updaters from reading the same prior mapping and one last
// writer erasing the other's Prior append. An flock on models.lock inside
// the cache dir holds for the whole read-modify-write.

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// lockAliasMap acquires an exclusive flock on models.lock inside dir and
// returns the open lock file — hold it until unlockAliasMap.
func lockAliasMap(dir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dir, "models.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("jev cache: alias-map lock failed: %w", err)
	}
	return f, nil
}

// unlockAliasMap releases the alias-map lock and closes the lock file. A
// failure on either is a best-effort cleanup the caller cannot act on (the
// update has already landed; the OS drops the flock at process exit anyway),
// so the error paths are deliberately discarded rather than escalated.
func unlockAliasMap(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}

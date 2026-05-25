//go:build windows

// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 — Windows per-SPEC close lock.
// Windows lacks fcntl-style advisory flock; we use atomic-create-file (O_CREATE|O_EXCL)
// per design.md §D.2 fallback. Stale lock detection (PID + timestamp embedded) is a
// post-MVP enhancement; M1 leaves stale-lock cleanup as a known-issue requiring
// manual `del .moai/state/spec-close-*.lock`.
package spec

import (
	"fmt"
	"os"
)

// atomicFileSpecLock holds an exclusively-created lock file. Releasing the
// lock means removing the file.
type atomicFileSpecLock struct {
	lockPath string
	file     *os.File
}

func (f *atomicFileSpecLock) release() error {
	if f == nil {
		return nil
	}
	if f.file != nil {
		_ = f.file.Close()
		f.file = nil
	}
	if f.lockPath != "" {
		err := os.Remove(f.lockPath)
		f.lockPath = ""
		if err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// acquireSpecCloseLockImpl creates lockPath with O_CREATE|O_EXCL — atomic on
// Windows NTFS. Returns ErrSpecCloseLockHeld when the file already exists.
func acquireSpecCloseLockImpl(lockPath string) (specCloseLockImpl, error) {
	file, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, ErrSpecCloseLockHeld
		}
		return nil, fmt.Errorf("open spec-close lock %s: %w", lockPath, err)
	}
	// Embed PID for future stale-lock detection (not yet acted upon in M1).
	_, _ = fmt.Fprintf(file, "pid=%d\n", os.Getpid())
	return &atomicFileSpecLock{lockPath: lockPath, file: file}, nil
}

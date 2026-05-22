//go:build windows

package migration

import (
	"errors"
	"io/fs"
	"os"
	"time"
)

// lockHandle holds the path of a Windows file mutex lock file.
// Windows has no equivalent to unix.Flock advisory lock, so we substitute a
// file mutex based on O_CREATE|O_EXCL (atomic create-or-fail).
// The REQ-V3R2-RT-007-031 advisory-lock semantics are simplified under the single-user assumption.
type lockHandle struct{ path string }

const (
	lockMaxRetries = 100 // max retry attempts (~1s @ 10ms/retry)
	lockRetryDelay = 10 * time.Millisecond
)

// acquireLock acquires the mutex by creating the lock file with O_CREATE|O_EXCL.
// If the file already exists, waits 10ms and retries, timing out after 100 attempts (~1s).
func acquireLock(lockPath string) (*lockHandle, error) {
	for i := 0; i < lockMaxRetries; i++ {
		f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY|os.O_TRUNC, 0600)
		if err == nil {
			_ = f.Close()
			return &lockHandle{path: lockPath}, nil
		}
		if errors.Is(err, fs.ErrExist) {
			time.Sleep(lockRetryDelay)
			continue
		}
		return nil, err
	}
	return nil, errors.New("lock acquire timeout (windows)")
}

// releaseLock releases the mutex by deleting the lock file.
// Treats a missing file as success (idempotent).
func releaseLock(h *lockHandle) error {
	err := os.Remove(h.path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

package escalation

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/modu-ai/moai-adk/internal/lockfile"
)

// cardLockTimeout bounds how long a hook waits for the per-card lock. A lock
// not taken in time is a REQ-AE-004 fault, never a tamper.
const cardLockTimeout = 3 * time.Second

// lockCard takes the per-card exclusive lock (design.md §C.6) and returns the
// release function.
//
// @MX:WARN: [AUTO] the blocking flock runs in a goroutine so the wait can be bounded; on timeout the goroutine is left to finish and releases the lock itself
// @MX:REASON: lockfile.Lock has no timeout; a hook must not outlive its budget waiting for another hook of the same card, and an abandoned acquisition must still unlock and close the file
func lockCard(path string, timeout time.Duration) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("escalation: create store: %w", err)
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("escalation: open card lock: %w", err)
	}
	// Unbuffered on purpose: a send completes only while the caller is still
	// waiting, so after a timeout the goroutine always sees abandoned instead.
	acquired := make(chan error)
	abandoned := make(chan struct{})
	go func() {
		err := lockfile.Lock(f)
		select {
		case <-abandoned:
			if err == nil {
				_ = lockfile.Unlock(f)
			}
			_ = f.Close()
		case acquired <- err:
		}
	}()
	select {
	case err := <-acquired:
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("escalation: lock card: %w", err)
		}
		return func() {
			_ = lockfile.Unlock(f)
			_ = f.Close()
		}, nil
	case <-time.After(timeout):
		close(abandoned)
		return nil, fmt.Errorf("escalation: card lock not taken within %s", timeout)
	}
}

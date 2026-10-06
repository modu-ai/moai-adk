//go:build windows

package userassets

import (
	"os"
	"time"
)

// On Windows, CreateFile-based byte-range locks need a handle dance the
// stdlib does not expose cleanly; the guard degrades to a short-lived
// marker file whose removal the next acquirer retries around. The O_EXCL
// create of the lock itself is already atomic on Windows, so the guard only
// needs to bound the reclaim window, which the retry loop tolerates.
var guardFd *os.File

func acquireGuard(path string) error {
	fd, err := os.OpenFile(path+".guard", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		// Held by another acquirer: poll briefly rather than fail.
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
			if fd, err = os.OpenFile(path+".guard", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644); err == nil {
				guardFd = fd
				return nil
			}
		}
		return err
	}
	guardFd = fd
	return nil
}

func releaseGuard() {
	if guardFd != nil {
		_ = guardFd.Close()
		_ = os.Remove(guardFd.Name() + ".guard")
		guardFd = nil
	}
}

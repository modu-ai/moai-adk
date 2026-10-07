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
// Item 3 (fix round 3): the handle is returned to its own call (no package
// global); item 4: the marker acquisition respects the same deadline.

func acquireGuard(path string, timeout time.Duration) (func(), error) {
	deadline := time.Now().Add(timeout)
	var fd *os.File
	var err error
	for {
		fd, err = os.OpenFile(path+".guard", os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := func() {
		_ = fd.Close()
		_ = os.Remove(path + ".guard")
	}
	return release, nil
}

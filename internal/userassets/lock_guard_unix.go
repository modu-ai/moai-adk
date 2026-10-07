//go:build !windows

package userassets

import (
	"os"
	"syscall"
	"time"
)

// The guard is a SECOND file held with flock(2) for the entire acquire
// attempt, so the reclaim-then-create sequence runs strictly serially per
// user (review-fix round 2 F5/A3/B2). Item 3 (fix round 3): the handle is
// returned to ITS OWN call — the former package-global was overwritten by
// concurrent calls from different HOMEs, releasing another call's guard and
// leaking the original (a real data race under -race).

func acquireGuard(path string, timeout time.Duration) (func(), error) {
	// Item 4: ONE deadline covers the whole acquire — the flock is
	// NON-BLOCKING with a retry loop bounded by the same deadline that
	// bounds the O_EXCL retries; a held guard can no longer overshoot it.
	deadline := time.Now().Add(timeout)
	fd, err := os.OpenFile(path+".guard", os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	for {
		err = syscall.Flock(int(fd.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			_ = fd.Close()
			return nil, err
		}
		time.Sleep(5 * time.Millisecond)
	}
	release := func() {
		_ = syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
		_ = fd.Close()
	}
	return release, nil
}

//go:build unix

package jev

// cache_lock_unix.go — P2 concurrency (card t1968 codex review round 9,
// operator-approved repair): updateAliasMap's read-modify-write must
// serialize across PROCESSES, not just goroutines — the cache dir is shared
// by every jev client on the machine, and file-replace atomicity alone does
// not stop two updaters from reading the same prior mapping and one last
// writer erasing the other's Prior append. An flock on models.lock inside
// the cache dir holds for the whole read-modify-write.
//
// iter16 F1 (operator-approved bounded-retry shape): the blocking LOCK_EX
// held an already-ANSWERED Ask hostage — the lock is reached only after the
// network answer is final, so a contending holder stalled the caller past
// its own deadline with nothing left to wait for. The acquire is now a
// LOCK_NB retry loop bounded by the caller's ctx and aliasLockMaxWait; at
// either bound it reports errAliasLockSkipped and updateAliasMap SKIPS the
// update — never worse than the pre-lock code, which lost this append
// silently in the same contended case. The max wait stays generous by
// design: the read-modify-write it guards is milliseconds, so a
// co-operating holder is always admitted within the bound and the
// Prior-append serialization the lock exists for is preserved (the barrier
// race test must keep passing — a skip there would be a lost append).

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// Alias-map lock acquire bounds: the retry loop polls LOCK_NB every
// aliasLockRetryInterval; either bound — the caller's ctx or
// aliasLockMaxWait — ends the wait with errAliasLockSkipped. The max wait
// must stay generous relative to the read-modify-write it protects (a few
// ms): it is two orders above it, so contention between two well-behaved
// updaters always resolves by acquisition, never by a skip.
const (
	aliasLockRetryInterval = 2 * time.Millisecond
	aliasLockMaxWait       = 500 * time.Millisecond
)

// lockAliasMap acquires an exclusive flock on models.lock inside dir and
// returns the open lock file — hold it until unlockAliasMap. The wait is
// bounded: when the caller's ctx is done, or aliasLockMaxWait passes before
// the lock is granted, it returns an error wrapping errAliasLockSkipped and
// the caller skips the update rather than blocking an answered call.
func lockAliasMap(ctx context.Context, dir string) (*os.File, error) {
	f, err := os.OpenFile(filepath.Join(dir, "models.lock"), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}
	deadline := time.Now().Add(aliasLockMaxWait)
	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return f, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			_ = f.Close()
			return nil, fmt.Errorf("jev cache: alias-map lock failed: %w", err)
		}
		remain := time.Until(deadline)
		if remain <= 0 {
			_ = f.Close()
			return nil, fmt.Errorf("%w: max wait %v exceeded", errAliasLockSkipped, aliasLockMaxWait)
		}
		wait := aliasLockRetryInterval
		if remain < wait {
			wait = remain
		}
		t := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			t.Stop()
			_ = f.Close()
			return nil, fmt.Errorf("%w: %v", errAliasLockSkipped, ctx.Err())
		case <-t.C:
		}
	}
}

// unlockAliasMap releases the alias-map lock and closes the lock file. A
// failure on either is a best-effort cleanup the caller cannot act on (the
// update has already landed; the OS drops the flock at process exit anyway),
// so the error paths are deliberately discarded rather than escalated.
func unlockAliasMap(f *os.File) {
	_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	_ = f.Close()
}

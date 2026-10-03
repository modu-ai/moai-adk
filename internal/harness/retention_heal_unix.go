//go:build !windows

package harness

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"time"
)

// acquireHealLock takes an exclusive flock on the heal-lock file at path and returns the function that
// releases it. The caller holds the lock only across the re-inspection and removal of a faulty
// state-path entry, so concurrent healers cannot remove each other's fresh state file.
//
// The entry is inspected without following links. An absent path is created exclusively with mode
// 0600; an existing regular file is opened read-write without create or truncate. Every open carries
// O_NOFOLLOW and O_NONBLOCK, and the opened file must be regular and still the inspected one. A
// symbolic link, a directory, a FIFO or any other non-regular entry is never opened, and the entry is
// never removed, replaced, truncated or chmodded. The owner check is asked about the path before the
// lock is requested. The lock is requested in exclusive mode only (a shared request would be compatible
// with another shared holder) and polled without blocking every pruneHealPoll until pruneHealWait has
// passed on the real clock.
//
// A failed create, open, inspection or lock call returns an error naming the path and wrapping the
// operating-system cause; a refused entry and a timeout return an error naming the path that wraps
// nothing.
//
// @MX:WARN: [AUTO] flock on a file that a hostile actor could replace; never opened through a link.
// @MX:REASON: [AUTO] The heal-lock path sits in a directory other processes write; O_NOFOLLOW, the
// regular-file check and os.SameFile keep the open from reaching a planted target or blocking on a FIFO.
func acquireHealLock(path string, ownerCheck func(string) bool) (func(), error) {
	var f *os.File
	for range maxStateInspections {
		if f != nil {
			break
		}
		fi, err := os.Lstat(path)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			cf, cerr := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0o600)
			if cerr == nil {
				f = cf
				continue
			}
			if errors.Is(cerr, fs.ErrExist) {
				continue // another process created it first: inspect again
			}
			return nil, healLockError(path, "cannot be created", cerr)
		case err != nil:
			return nil, healLockError(path, "cannot be inspected", err)
		case fi.Mode().IsRegular():
			of, oerr := os.OpenFile(path, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
			if oerr != nil {
				return nil, healLockError(path, "cannot be opened", oerr)
			}
			if st, serr := of.Stat(); serr != nil || !st.Mode().IsRegular() || !os.SameFile(fi, st) {
				_ = of.Close()
				continue // swapped between inspection and open: inspect again
			}
			f = of
		default:
			return nil, healLockError(path, "is not a regular file", nil)
		}
	}
	if f == nil {
		return nil, healLockError(path, "changed on every inspection", nil)
	}
	if !ownerCheck(path) {
		_ = f.Close()
		return nil, healLockError(path, "is not owned by the current user", nil)
	}
	deadline := time.Now().Add(pruneHealWait)
	for {
		ferr := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if ferr == nil {
			return func() {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
			}, nil
		}
		if !errors.Is(ferr, syscall.EWOULDBLOCK) && !errors.Is(ferr, syscall.EINTR) {
			_ = f.Close()
			return nil, healLockError(path, "cannot be locked", ferr)
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, healLockError(path, fmt.Sprintf("was not acquired within %s", pruneHealWait), nil)
		}
		time.Sleep(pruneHealPoll)
	}
}

// healLockError names the heal-lock path and, when there is one, wraps the operating-system cause.
func healLockError(path, what string, cause error) error {
	if cause != nil {
		return fmt.Errorf("retention: prune heal lock %s %s: %w", path, what, cause)
	}
	return fmt.Errorf("retention: prune heal lock %s %s", path, what)
}

//go:build !windows

package harness

// DRAFT measurement instrument for the SPEC-HARNESS-RETENTION-HARDEN-001 amendment 0.4.1 (card t1432).
// NOT the implementation and not in the tree: it is a minimal model of the heal-lock design of
// spec.md §B D4.c, injected with `go test -overlay` together with retention_proto.go (a copy of the
// tree's retention.go whose healStateEntry calls acquireHealLockProto). Its only purpose is to observe
// how the draft tests behave against a faithful model of the design (a test that a model cannot pass,
// or that a one-constant mutant of the model still passes, is shallow). The run phase writes its own
// implementation from the criteria, not from this file.
//
// Model: Lstat the heal-lock path (at most three inspections); create it exclusively with mode 0600
// when absent; open an existing regular file read-write without create or truncate, with O_NOFOLLOW and
// O_NONBLOCK; require the opened file to be regular and the inspected file; ask the owner check about
// the path; poll a non-blocking flock every 10 ms up to 2 s on the real clock. A symbolic link, a
// directory, a FIFO or another non-regular entry is never opened. Two package variables let a mutant
// overlay (a tiny init() file) change the lock mode or stop wrapping the cause.

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"syscall"
	"time"
)

const (
	protoHealSuffix = ".prune-heal"
	protoHealWait   = 2 * time.Second
	protoHealPoll   = 10 * time.Millisecond
)

var (
	protoLockMode = syscall.LOCK_EX // a mutant overlay sets LOCK_SH
	protoWrap     = true            // a mutant overlay sets false: the cause is formatted with %v

	// protoIgnoreRemoval is read by retention_proto.go: a mutant overlay sets true and the heal then
	// ignores a failed removal of the state-path entry.
	protoIgnoreRemoval = false

	// protoLockCommonIfExists is read by retention_proto.go: a mutant overlay sets true and the common
	// path then opens and locks a heal-lock file that already exists.
	protoLockCommonIfExists = false
)

func protoHealErr(path string, cause error, what string) error {
	switch {
	case cause != nil && protoWrap:
		return fmt.Errorf("retention: prune heal lock %s %s: %w", path, what, cause)
	case cause != nil:
		return fmt.Errorf("retention: prune heal lock %s %s: %v", path, what, cause)
	default:
		return fmt.Errorf("retention: prune heal lock %s %s", path, what)
	}
}

func acquireHealLockProto(path string, ownerCheck func(string) bool) (func(), error) {
	var f *os.File
	for i := 0; i < 3 && f == nil; i++ {
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
			return nil, protoHealErr(path, cerr, "cannot be created")
		case err != nil:
			return nil, protoHealErr(path, err, "cannot be inspected")
		case fi.Mode().IsRegular():
			of, oerr := os.OpenFile(path, os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
			if oerr != nil {
				return nil, protoHealErr(path, oerr, "cannot be opened")
			}
			if st, serr := of.Stat(); serr != nil || !st.Mode().IsRegular() || !os.SameFile(fi, st) {
				_ = of.Close()
				continue // swapped between inspection and open: inspect again
			}
			f = of
		default:
			return nil, protoHealErr(path, nil, "is not a regular file")
		}
	}
	if f == nil {
		return nil, protoHealErr(path, nil, "changed on every inspection")
	}
	if !ownerCheck(path) {
		_ = f.Close()
		return nil, protoHealErr(path, nil, "is not owned by the current user")
	}
	deadline := time.Now().Add(protoHealWait)
	for {
		ferr := syscall.Flock(int(f.Fd()), protoLockMode|syscall.LOCK_NB)
		if ferr == nil {
			return func() {
				_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				_ = f.Close()
			}, nil
		}
		if !errors.Is(ferr, syscall.EWOULDBLOCK) && !errors.Is(ferr, syscall.EINTR) {
			_ = f.Close()
			return nil, protoHealErr(path, ferr, "cannot be locked")
		}
		if time.Now().After(deadline) {
			_ = f.Close()
			return nil, protoHealErr(path, nil, fmt.Sprintf("was not acquired within %s", protoHealWait))
		}
		time.Sleep(protoHealPoll)
	}
}

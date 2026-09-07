package lockfile

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The tests in this file carry NO build constraint on purpose.
//
// Before card t364 the package's only test file was lockfile_unix_test.go,
// tagged //go:build !windows. That made it the sole test file on Unix and left
// windows with none at all, so `go test` reported
// "? internal/lockfile [no test files]" — an ok-shaped line indistinguishable
// from a package that ran and passed under rc-only reporting. The windows
// Lock/Unlock pair shipped with zero executed coverage (t358 census, run
// 33308057570, windows leg).
//
// lockfile_unix_test.go stays Unix-only: it asserts flock(2) kernel semantics
// through syscall directly, which have no Windows counterpart. What IS shared
// by both implementations is the exported contract — Lock/Unlock mutually
// exclude by file path within one process, and Unlock releases. That contract
// is what this file asserts, through the exported API only, so it compiles and
// runs on every GOOS.

// TestLockUnlockRoundTripReturnsNoError pins the plain success path of the
// exported pair on the current platform.
func TestLockUnlockRoundTripReturnsNoError(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "roundtrip.lock")

	f, err := os.Create(lockPath)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := Lock(f); err != nil {
		t.Fatalf("Lock: %v", err)
	}
	if err := Unlock(f); err != nil {
		t.Fatalf("Unlock: %v", err)
	}
}

// TestLockExcludesSecondHandleWhileHeld is the load-bearing assertion: while one
// handle holds the lock, a second handle on the same path must NOT acquire it,
// and must acquire it once the first releases.
//
// Non-vacuous in both directions. A no-op Lock fails the first select (the
// second handle acquires while the lock is held); a no-op Unlock fails the
// second (the second handle never acquires). Both were confirmed by deliberate
// mutation of lockfile_unix.go on darwin — see .moai/reports/t364/verdict.md.
//
// The second handle is closed only on paths where its Lock has demonstrably
// returned. Closing a handle whose Lock is still blocked in flight wedges the
// test binary until the go test timeout instead of failing fast — observed on
// darwin during the mutation run above, where an earlier draft of this test
// took 124s to report a failure it should have reported in 5s.
func TestLockExcludesSecondHandleWhileHeld(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "exclusion.lock")

	f1, err := os.Create(lockPath)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	defer func() { _ = f1.Close() }()

	// The path string must match f1.Name() exactly: the Windows implementation
	// keys its mutex map on that string.
	f2, err := os.OpenFile(lockPath, os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open second handle: %v", err)
	}

	if err := Lock(f1); err != nil {
		t.Fatalf("Lock first handle: %v", err)
	}

	started := make(chan struct{})
	acquired := make(chan error, 1)
	go func() {
		close(started)
		acquired <- Lock(f2)
	}()
	<-started

	// Negative window. This wait can only produce a false PASS (if the goroutine
	// has not reached Lock yet), never a false FAIL — a failure here requires the
	// lock to have actually been granted while held. The positive half below is
	// what keeps the test non-vacuous even when this window is racy.
	select {
	case err := <-acquired:
		_ = Unlock(f2)
		_ = f2.Close()
		t.Fatalf("second handle acquired the lock while the first held it (err=%v)", err)
	case <-time.After(250 * time.Millisecond):
	}

	if err := Unlock(f1); err != nil {
		t.Fatalf("Unlock first handle: %v", err)
	}

	select {
	case err := <-acquired:
		if err != nil {
			t.Errorf("Lock on second handle after release: %v", err)
		}
		if err := Unlock(f2); err != nil {
			t.Errorf("Unlock second handle: %v", err)
		}
		_ = f2.Close()
	case <-time.After(5 * time.Second):
		// f2 is deliberately left open: its Lock is still in flight.
		t.Error("second handle did not acquire the lock within 5s of release")
	}
}

// TestUnlockWithoutLockIsNoError pins the documented tolerance of releasing a
// file that was never locked (lockfile_windows.go returns nil for an unknown
// path; flock(LOCK_UN) on an unlocked descriptor likewise succeeds). Call sites
// unlock from deferred cleanup paths that can run before a successful Lock.
func TestUnlockWithoutLockIsNoError(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "never-locked.lock")

	f, err := os.Create(lockPath)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	defer func() { _ = f.Close() }()

	if err := Unlock(f); err != nil {
		t.Errorf("Unlock without a prior Lock should be a no-op, got: %v", err)
	}
}

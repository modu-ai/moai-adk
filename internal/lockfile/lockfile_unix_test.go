//go:build !windows

package lockfile

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestFlock tests filesystem lock behavior for concurrent access. Migrated from
// internal/cli/team_spawn_lock_unix_test.go (SPEC-AGENT-TEAM-RETIRE-001 M0,
// REQ-ATR-001), re-pointed at the exported Lock/Unlock API. It is the renamed
// form of the previously-misnamed TestFilesystemLock (the file was
// team_spawn_lock_test_unix.go — NOT _test.go-suffixed — so the test never
// compiled into the test binary and the testing package leaked into the
// production dependency graph). SPEC-CLIFIX-CONTRACT-001 M4 / REQ-CONT-001-007.
//
// Non-vacuous guard: this test has TWO load-bearing t.Error paths — (1) the
// goroutine acquires the lock while it is held (flock regression), and (2) the
// 2s timeout fires (the non-blocking attempt hung). A no-op Lock would
// trigger path (1), so the test genuinely fails if flock semantics regress.
// This is Unix-only because it validates flock(2) kernel semantics which are
// not available on Windows (where Lock is an in-process mutex by design).
func TestFlock(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping lock test in short mode")
	}

	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, "test.lock")

	// Create lock file
	f, err := os.Create(lockPath)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	defer func() {
		_ = f.Close()
		_ = os.Remove(lockPath)
	}()

	// Acquire exclusive lock via the exported API
	if err := Lock(f); err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	defer func() {
		_ = Unlock(f)
	}()

	// Try to acquire lock in another "process" (goroutine with timeout)
	lockAcquired := make(chan bool)
	go func() {
		f2, err := os.Open(lockPath)
		if err != nil {
			t.Logf("open lock file in goroutine: %v", err)
			return
		}
		defer func() { _ = f2.Close() }()

		// Try non-blocking lock
		err = syscall.Flock(int(f2.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			// Lock is held, as expected
			lockAcquired <- false
			return
		}

		// Lock acquired (should not happen)
		lockAcquired <- true
		_ = syscall.Flock(int(f2.Fd()), syscall.LOCK_UN)
	}()

	select {
	case acquired := <-lockAcquired:
		if acquired {
			t.Error("lock should not be acquired while held by another process")
		}
	case <-time.After(2 * time.Second):
		t.Error("lock test timeout")
	}
}

// TestLockUnlockRoundTrip verifies the exported Lock/Unlock pair releases the
// advisory lock so a subsequent non-blocking acquisition succeeds.
func TestLockUnlockRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	lockPath := filepath.Join(tempDir, "roundtrip.lock")

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

	// After Unlock, a second fd must be able to acquire the lock non-blocking.
	f2, err := os.Open(lockPath)
	if err != nil {
		t.Fatalf("open second fd: %v", err)
	}
	defer func() { _ = f2.Close() }()

	if err := syscall.Flock(int(f2.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("non-blocking lock after Unlock should succeed, got: %v", err)
	}
	_ = syscall.Flock(int(f2.Fd()), syscall.LOCK_UN)
}

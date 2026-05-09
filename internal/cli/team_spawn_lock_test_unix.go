//go:build !windows

package cli

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

// TestFilesystemLock tests filesystem lock behavior for concurrent access.
// This test is Unix-only because it validates flock(2) kernel semantics
// which are not available on Windows (where lockFile is a no-op by design).
func TestFilesystemLock(t *testing.T) {
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

	// Acquire exclusive lock
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	defer func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
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

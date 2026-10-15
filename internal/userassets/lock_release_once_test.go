// lock_release_once_test.go — the user lock handle is released at most once,
// and its removal runs only after the holder's token is verified on disk
// (card t1591, factory finding P1). init_resume.go releases the lock after the
// user-asset ensure and again from a deferred call at exit. The second call
// found the file already gone, skipped the token check, and removed whatever
// lock another owner had created at the path in between.
package userassets

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	releaseSuccessorRecord = "pid=1 token=successor-owner acquired=2026-10-10T00:00:00Z\n"
	releaseForeignRecord   = "pid=1 token=foreign-owner acquired=2026-10-10T00:00:00Z\n"
)

// createReleaseSuccessor creates a successor's lock at path with O_EXCL, the
// way AcquireUserLock does. It reports whether the successor now exists.
func createReleaseSuccessor(t *testing.T, path string) bool {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		t.Errorf("setup: the seam could not create the successor lock: %v", err)
		return false
	}
	_, _ = f.WriteString(releaseSuccessorRecord)
	_ = f.Close()
	return true
}

// writeReleaseFixture overwrites path with content, creating it when absent.
func writeReleaseFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// requireReleaseFixture asserts that path still holds content, byte for byte.
func requireReleaseFixture(t *testing.T, path, content string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(raw) != content {
		t.Fatalf("%s changed: %q, want %q", path, raw, content)
	}
}

// TestLockReleaseWithAbsentFileNeverRemovesSuccessor pins the removal guard:
// the lock file is already gone at the ownership read, the seam creates a
// successor's lock in the window before Release returns, and Release must leave
// that lock in place. Today the unguarded removal deletes it. Not parallel: it
// swaps the package-level seam.
func TestLockReleaseWithAbsentFileNeverRemovesSuccessor(t *testing.T) {
	home := t.TempDir()
	lock, err := AcquireUserLock(home, time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	lockPath := LockPath(home)
	if err := os.Remove(lockPath); err != nil {
		t.Fatalf("setup: remove the held lock file: %v", err)
	}

	reached, created := false, false
	prev := releaseWindowHook
	releaseWindowHook = func(path string) {
		reached = true
		created = createReleaseSuccessor(t, path)
	}
	t.Cleanup(func() { releaseWindowHook = prev })

	err = lock.Release()
	if !reached || !created {
		t.Fatalf("setup: the seam did not run with the successor lock created (reached=%v created=%v)", reached, created)
	}
	raw, readErr := os.ReadFile(lockPath)
	if readErr != nil {
		t.Fatalf("Release removed another owner's lock: the successor created while the lock file was absent is gone (%v)", readErr)
	}
	if string(raw) != releaseSuccessorRecord {
		t.Fatalf("the successor's lock was rewritten: %q", raw)
	}
	if err != nil {
		t.Fatalf("release of an absent lock returned %v, want nil", err)
	}
}

// TestLockSecondReleaseNeverRemovesSuccessorLock is the double-release shape
// init_resume.go produces: acquire, release, then a second Release on the same
// handle while another owner acquires the path. A released handle returns before
// it reads, so the seam, which sits after the read, must not fire at all. Not
// parallel: it swaps the package-level seam.
func TestLockSecondReleaseNeverRemovesSuccessorLock(t *testing.T) {
	home := t.TempDir()
	lock, err := AcquireUserLock(home, time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Fatalf("first release: %v", err)
	}
	lockPath := LockPath(home)

	reached, created := false, false
	prev := releaseWindowHook
	releaseWindowHook = func(path string) {
		reached = true
		created = createReleaseSuccessor(t, path)
	}
	t.Cleanup(func() { releaseWindowHook = prev })

	secondErr := lock.Release() // second call on the same handle
	raw, readErr := os.ReadFile(lockPath)
	if reached {
		if !created {
			t.Fatal("setup: the seam ran but the successor lock was not created")
		}
		if readErr != nil {
			t.Fatalf("the second Release removed another owner's lock: the successor created in its window is gone (%v)", readErr)
		}
		t.Fatalf("the second Release read the lock path after the first release (the seam ran); a released handle must not read: %q", raw)
	}
	// The released handle returned before reading: the path is untouched.
	if !os.IsNotExist(readErr) {
		t.Fatalf("the released handle changed the lock path: read %q (%v)", raw, readErr)
	}
	if secondErr != nil {
		t.Fatalf("second release returned %v, want nil (a no-op)", secondErr)
	}
}

// TestLockSecondReleaseIsNoOp pins the released-handle contract: a later Release
// returns nil and touches neither a successor's lock nor an unrelated file. The
// reclaimed subtest covers the other first-release outcome, where the first call
// refused a path another owner had taken.
func TestLockSecondReleaseIsNoOp(t *testing.T) {
	t.Run("successor lock and unrelated file are left alone", func(t *testing.T) {
		home := t.TempDir()
		lock, err := AcquireUserLock(home, time.Second)
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		if err := lock.Release(); err != nil {
			t.Fatalf("first release: %v", err)
		}
		lockPath := LockPath(home)
		unrelated := filepath.Join(filepath.Dir(lockPath), "unrelated.sentinel")
		writeReleaseFixture(t, lockPath, releaseSuccessorRecord)
		writeReleaseFixture(t, unrelated, "unrelated\n")

		if err := lock.Release(); err != nil {
			t.Fatalf("second release on a released handle returned %v, want nil (a no-op)", err)
		}
		requireReleaseFixture(t, lockPath, releaseSuccessorRecord)
		requireReleaseFixture(t, unrelated, "unrelated\n")
	})

	t.Run("reclaimed handle stays a no-op after its first release", func(t *testing.T) {
		home := t.TempDir()
		lock, err := AcquireUserLock(home, time.Second)
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		lockPath := LockPath(home)
		writeReleaseFixture(t, lockPath, releaseForeignRecord)

		if err := lock.Release(); err == nil || !strings.Contains(err.Error(), "reclaimed") {
			t.Fatalf("first release of a reclaimed lock returned %v, want the reclaimed-owner refusal", err)
		}
		if err := lock.Release(); err != nil {
			t.Fatalf("second release returned %v, want nil (a no-op)", err)
		}
		requireReleaseFixture(t, lockPath, releaseForeignRecord)
	})
}

// TestLockReleaseRefusesUnverifiableLockPath pins the fail-closed branch: when the
// lock path cannot be read for a reason other than absence, ownership cannot be
// verified, so Release returns an error and removes nothing. The path here is an
// empty directory, which the unguarded removal deletes while reporting success.
func TestLockReleaseRefusesUnverifiableLockPath(t *testing.T) {
	home := t.TempDir()
	lockPath := LockPath(home)
	if err := os.MkdirAll(lockPath, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	lock := &UserLock{path: lockPath, token: "mine"}

	if err := lock.Release(); err == nil {
		t.Error("release of an unverifiable lock path returned nil; ownership cannot be verified, so it must refuse")
	}
	if info, err := os.Stat(lockPath); err != nil || !info.IsDir() {
		t.Errorf("release removed the unverifiable lock path: stat = %v, %v", info, err)
	}
}

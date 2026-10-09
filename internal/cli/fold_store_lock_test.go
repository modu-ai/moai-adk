// fold_store_lock_test.go — the fold store lock's unit cells
// (SPEC-MEMORY-FOLD-RENAME-RACE-001, plan.md M2 step 1): acquire →
// contending acquire refused within the bound → release → acquire
// succeeds; a stale unlocked lock file acquires normally; two stores'
// locks interleave independently; release never removes the lock file.

package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestFoldStoreLockAcquireRefuseRelease is plan.md M2 step 1's core cell: a
// held lock refuses a contending non-blocking acquire immediately and a
// bounded acquire within the deadline; after release, acquisition
// succeeds again; release is idempotent and never removes the lock file
// (D-3).
func TestFoldStoreLockAcquireRefuseRelease(t *testing.T) {
	dir := t.TempDir()
	held := newFoldStoreLock()
	if err := held.acquire(dir); err != nil {
		t.Fatalf("first acquire: %v", err)
	}

	// A contending non-blocking acquire is refused immediately.
	try := newFoldStoreLock()
	if err := try.tryAcquire(dir); err == nil {
		_ = try.release()
		t.Fatal("the contending non-blocking acquire succeeded while the store lock was held")
	}

	// The bounded acquire refuses WITHIN the deadline — it never blocks
	// without a bound (REQ-MRR-003) — and the refusal names the contended
	// store.
	start := time.Now()
	boundedErr := newFoldStoreLock().acquire(dir)
	if boundedErr == nil {
		t.Fatal("the bounded acquire succeeded while the store lock was held")
	}
	if elapsed := time.Since(start); elapsed > foldStoreLockRetryDeadline+time.Second {
		t.Errorf("the bounded acquire waited %s, over the %s deadline", elapsed, foldStoreLockRetryDeadline)
	}
	if !strings.Contains(boundedErr.Error(), dir) {
		t.Errorf("the bounded refusal does not name the contended store %s: %v", dir, boundedErr)
	}

	// After release, acquisition succeeds again.
	if err := held.release(); err != nil {
		t.Fatalf("first release: %v", err)
	}
	again := newFoldStoreLock()
	if err := again.acquire(dir); err != nil {
		t.Fatalf("acquire after release: %v", err)
	}
	// release is idempotent.
	if err := again.release(); err != nil {
		t.Errorf("second release: %v", err)
	}
	if err := again.release(); err != nil {
		t.Errorf("third release (idempotence): %v", err)
	}

	// The lock file is never removed on release (D-3).
	if _, err := os.Stat(filepath.Join(dir, foldLockFileName)); err != nil {
		t.Errorf("the lock file was removed on release: %v", err)
	}
}

// TestFoldStoreLockStaleFileAcquiresNormally is §3 edge 1: a lock file left
// behind by a killed holder is unlocked — the lock's kernel lifetime means
// the file's presence alone never blocks, so the acquire succeeds.
func TestFoldStoreLockStaleFileAcquiresNormally(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, foldLockFileName), nil, 0o644); err != nil {
		t.Fatalf("seed the stale lock file: %v", err)
	}
	l := newFoldStoreLock()
	if err := l.acquire(dir); err != nil {
		t.Fatalf("acquire on an existing unlocked lock file: %v", err)
	}
	if err := l.release(); err != nil {
		t.Fatalf("release: %v", err)
	}
}

// TestFoldStoreLockTwoStoresAreIndependent is §3 edge 3: two folds on
// DIFFERENT stores take distinct lock files and hold their locks
// simultaneously, without interference.
func TestFoldStoreLockTwoStoresAreIndependent(t *testing.T) {
	dirA, dirB := t.TempDir(), t.TempDir()
	lockA := newFoldStoreLock()
	if err := lockA.acquire(dirA); err != nil {
		t.Fatalf("acquire store A: %v", err)
	}
	defer func() { _ = lockA.release() }()

	// Store B's lock is uncontended by store A's hold.
	lockB := newFoldStoreLock()
	if err := lockB.acquire(dirB); err != nil {
		t.Fatalf("acquire store B while store A's lock is held: %v", err)
	}
	if err := lockB.release(); err != nil {
		t.Fatalf("release store B: %v", err)
	}

	// The two stores carry two distinct lock files.
	for _, dir := range []string{dirA, dirB} {
		if _, err := os.Stat(filepath.Join(dir, foldLockFileName)); err != nil {
			t.Errorf("store %s carries no lock file after its acquire: %v", dir, err)
		}
	}
}

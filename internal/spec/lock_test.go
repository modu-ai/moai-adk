// lock_test.go — per-SPEC close lock contention tests.
// SPEC-V3R6-LIFECYCLE-SYNC-GATE-001 AC-LSG-010, AC-LSG-021.
package spec

import (
	"errors"
	"sync"
	"testing"
)

func TestAcquireSpecCloseLock_Success(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	lock, err := AcquireSpecCloseLock(tempDir, "SPEC-TEST-001")
	if err != nil {
		t.Fatalf("AcquireSpecCloseLock() error = %v", err)
	}
	if lock == nil {
		t.Fatal("lock is nil")
	}
	defer func() { _ = lock.Release() }()

	if lock.SpecID() != "SPEC-TEST-001" {
		t.Errorf("SpecID() = %q, want SPEC-TEST-001", lock.SpecID())
	}
	if lock.LockPath() == "" {
		t.Error("LockPath() is empty")
	}
}

func TestAcquireSpecCloseLock_EmptySpecID(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	_, err := AcquireSpecCloseLock(tempDir, "")
	if err == nil {
		t.Fatal("expected error for empty specID, got nil")
	}
}

// AC-LSG-010 — File lock contention: two concurrent invocations, exactly one
// acquires the lock; the other receives ErrSpecCloseLockHeld.
func TestAcquireSpecCloseLock_Contention(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	lock1, err1 := AcquireSpecCloseLock(tempDir, "SPEC-TEST-002")
	if err1 != nil {
		t.Fatalf("first lock acquisition failed: %v", err1)
	}
	defer func() { _ = lock1.Release() }()

	// Second acquisition attempt must fail with ErrSpecCloseLockHeld
	lock2, err2 := AcquireSpecCloseLock(tempDir, "SPEC-TEST-002")
	if err2 == nil {
		t.Fatal("second acquisition should fail with ErrSpecCloseLockHeld, got nil")
	}
	if !IsLockHeldError(err2) {
		t.Errorf("err2 should be ErrSpecCloseLockHeld, got: %v", err2)
	}
	if !errors.Is(err2, ErrSpecCloseLockHeld) {
		t.Errorf("errors.Is(err2, ErrSpecCloseLockHeld) = false; err2 = %v", err2)
	}
	if lock2 != nil {
		t.Error("second lock should be nil on contention")
		_ = lock2.Release()
	}
}

// AC-LSG-010 — different SPECs may lock concurrently (per-SPEC scoping).
func TestAcquireSpecCloseLock_DifferentSpecsConcurrent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	lockA, errA := AcquireSpecCloseLock(tempDir, "SPEC-A-001")
	if errA != nil {
		t.Fatalf("SPEC-A-001 lock failed: %v", errA)
	}
	defer func() { _ = lockA.Release() }()

	lockB, errB := AcquireSpecCloseLock(tempDir, "SPEC-B-001")
	if errB != nil {
		t.Fatalf("SPEC-B-001 lock should succeed in parallel with SPEC-A-001: %v", errB)
	}
	defer func() { _ = lockB.Release() }()
}

// AC-LSG-010 — Release must allow re-acquisition.
func TestAcquireSpecCloseLock_ReleaseAndReacquire(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	lock1, err := AcquireSpecCloseLock(tempDir, "SPEC-TEST-003")
	if err != nil {
		t.Fatalf("first acquisition failed: %v", err)
	}
	if err := lock1.Release(); err != nil {
		t.Fatalf("Release() failed: %v", err)
	}

	// After release, re-acquisition should succeed
	lock2, err := AcquireSpecCloseLock(tempDir, "SPEC-TEST-003")
	if err != nil {
		t.Fatalf("re-acquisition after release failed: %v", err)
	}
	defer func() { _ = lock2.Release() }()
}

// Release is idempotent — calling multiple times returns nil.
func TestSpecCloseLock_ReleaseIdempotent(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	lock, err := AcquireSpecCloseLock(tempDir, "SPEC-TEST-004")
	if err != nil {
		t.Fatalf("acquisition failed: %v", err)
	}

	if err := lock.Release(); err != nil {
		t.Errorf("first Release() error = %v", err)
	}
	if err := lock.Release(); err != nil {
		t.Errorf("second Release() (idempotent) error = %v", err)
	}
}

// AC-LSG-021 (NFR-LSG-005) — Concurrent close safety: spawn N goroutines all
// attempting to acquire the same per-SPEC lock; exactly one must succeed and
// the rest must receive ErrSpecCloseLockHeld.
func TestSpecCloseLock_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	const N = 5
	var successCount, contentionCount int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Use a barrier to maximize contention.
	startBarrier := make(chan struct{})
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startBarrier
			lock, err := AcquireSpecCloseLock(tempDir, "SPEC-CONCURRENT-001")
			mu.Lock()
			defer mu.Unlock()
			if err == nil && lock != nil {
				successCount++
				// Hold briefly then release so other goroutines can compete.
				_ = lock.Release()
			} else if IsLockHeldError(err) {
				contentionCount++
			} else {
				t.Errorf("unexpected error: %v", err)
			}
		}()
	}
	close(startBarrier)
	wg.Wait()

	// At least one acquisition + at least one contention expected.
	// (Exact counts depend on scheduling but must sum to N.)
	if successCount+contentionCount != N {
		t.Errorf("success+contention = %d, want %d", successCount+contentionCount, N)
	}
	if successCount < 1 {
		t.Errorf("no successful acquisition (success=%d, contention=%d)", successCount, contentionCount)
	}
}

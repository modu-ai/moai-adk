package factory

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestFactoryStepLockSerializesContenders — SPEC-FACTORY-ATOMIC-LEASE-001
// plan D3: the contenders of one root never hold the lock at once.
func TestFactoryStepLockSerializesContenders(t *testing.T) {
	root := t.TempDir()
	const contenders = 6
	var inside, maxInside atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < contenders; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			release, err := AcquireFactoryStepLock(root, 20*time.Second)
			if err != nil {
				t.Errorf("acquire: %v", err)
				return
			}
			n := inside.Add(1)
			for {
				m := maxInside.Load()
				if n <= m || maxInside.CompareAndSwap(m, n) {
					break
				}
			}
			time.Sleep(20 * time.Millisecond)
			inside.Add(-1)
			if err := release(); err != nil {
				t.Errorf("release: %v", err)
			}
		}()
	}
	wg.Wait()
	if got := maxInside.Load(); got != 1 {
		t.Fatalf("%d contenders held the step lock at once, want 1", got)
	}
}

// TestFactoryStepLockReleasesOnErrorAndPanic — a caller that defers the
// release leaves the lock free whether its step returns an error or panics.
func TestFactoryStepLockReleasesOnErrorAndPanic(t *testing.T) {
	root := t.TempDir()
	step := func(fail func()) {
		release, err := AcquireFactoryStepLock(root, time.Second)
		if err != nil {
			t.Fatalf("acquire: %v", err)
		}
		defer func() { _ = release() }()
		fail()
	}
	step(func() {})
	func() {
		defer func() { _ = recover() }()
		step(func() { panic("step failed") })
	}()
	release, err := AcquireFactoryStepLock(root, time.Second)
	if err != nil {
		t.Fatalf("the lock stayed held after a panicking step: %v", err)
	}
	if err := release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if err := release(); err != nil {
		t.Fatalf("a second release must be safe, got %v", err)
	}
}

// TestFactoryStepLockWaitIsBounded — a held lock fails a contender within its
// wait (plus one retry wait), names the artifact, wraps the held sentinel and
// still returns a callable release.
func TestFactoryStepLockWaitIsBounded(t *testing.T) {
	root := t.TempDir()
	held, err := AcquireFactoryStepLock(root, time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	t.Cleanup(func() { _ = held() })

	const wait = 200 * time.Millisecond
	start := time.Now()
	release, err := AcquireFactoryStepLock(root, wait)
	elapsed := time.Since(start)
	if err == nil {
		_ = release()
		t.Fatalf("a contender obtained a held step lock")
	}
	if !IsStateLockHeld(err) {
		t.Errorf("error %v does not wrap the held sentinel", err)
	}
	if elapsed < wait || elapsed > wait+500*time.Millisecond {
		t.Errorf("the wait returned after %s, want within [%s, %s]", elapsed, wait, wait+500*time.Millisecond)
	}
	if release == nil {
		t.Fatalf("release is nil on failure")
	}
	if err := release(); err != nil {
		t.Errorf("the failure's release must be a no-op, got %v", err)
	}
}

// TestFactoryStepLockIsItsOwnFile — the step lock is not the queue's lock:
// holding it does not block the queue lock.
func TestFactoryStepLockIsItsOwnFile(t *testing.T) {
	root := t.TempDir()
	release, err := AcquireFactoryStepLock(root, time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer func() { _ = release() }()
	if _, err := os.Stat(filepath.Join(root, ".moai", "state", "factory-worktree-step.lock")); err != nil {
		t.Fatalf("the step lock artifact is not at the documented path: %v", err)
	}
	queue := NewBacklogStore(filepath.Join(root, ".moai", "state", backlogFileName))
	held, err := queue.acquireLock()
	if err != nil {
		t.Fatalf("the queue lock was blocked by the step lock: %v", err)
	}
	_ = held.Release()
}

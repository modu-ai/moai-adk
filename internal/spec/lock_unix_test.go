//go:build !windows

// lock_unix_test.go — Unix-specific flock pressure repro (card t1291).
//
// The flock(2) attempt in acquireSpecCloseLockImpl is a raw syscall: the Go
// runtime delivers its preemption signal (SIGURG) without SA_RESTART, so a
// LOCK_NB|LOCK_EX attempt landing under signal delivery returns EINTR. Heavy
// parallel test load (go test's default -p across packages) widens that window;
// the failure was observed only there — solo and -p 1 runs stayed green
// (t1233 differential: remeasure-merged.txt vs remeasure-merged-serial.txt).
// The acquire path must retry an interrupted attempt, not report the lock held.
package spec

import (
	"fmt"
	"sync"
	"testing"

	"golang.org/x/sys/unix"
)

// TestAcquireSpecCloseLock_SignalPressureReleaseReacquire hammers the
// acquire/release cycle on private lock paths while a goroutine floods the
// process with SIGURG — the same signal class the runtime itself uses for
// preemption, amplified. Any iteration whose flock attempt is interrupted must
// still succeed; a misclassified EINTR surfaces as "spec-close lock held".
//
// The flooder is fully self-contained: it stops on channel close before the
// test returns, leaving no process or signal source behind.
func TestAcquireSpecCloseLock_SignalPressureReleaseReacquire(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()

	stop := make(chan struct{})
	floodDone := make(chan struct{})
	go func() {
		defer close(floodDone)
		for {
			select {
			case <-stop:
				return
			default:
				_ = unix.Kill(unix.Getpid(), unix.SIGURG)
			}
		}
	}()
	defer func() { <-floodDone }()
	defer close(stop)

	// Distinct specIDs keep the goroutines contention-free, so any
	// ErrSpecCloseLockHeld observed here is spurious by construction.
	const workers = 4
	const iters = 20000
	errCh := make(chan error, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			specID := "SPEC-REPRO-" + string(rune('A'+w))
			for i := 0; i < iters; i++ {
				lock, err := AcquireSpecCloseLock(tempDir, specID)
				if err != nil {
					errCh <- fmt.Errorf("iter %d: %w", i, err)
					return
				}
				if err := lock.Release(); err != nil {
					errCh <- fmt.Errorf("iter %d: release: %w", i, err)
					return
				}
			}
			errCh <- nil
		}(w)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatalf("signal-pressure acquire/release failed: %v", err)
		}
	}
}

// TestAcquireSpecCloseLock_RetriesInterruptedAttempt (card t1291): a
// signal-interrupted flock attempt (EINTR — raw syscalls observe the runtime's
// preemption signals; heavy parallel test load widens the window) must be
// retried, never reported as the lock being held.
//
// Deliberately NOT t.Parallel(): it stubs the package-level flockAttempt seam,
// and sequential tests run while the parallel ones above are paused, so the
// stub never races a real acquire.
func TestAcquireSpecCloseLock_RetriesInterruptedAttempt(t *testing.T) {
	tempDir := t.TempDir()

	calls := 0
	orig := flockAttempt
	defer func() { flockAttempt = orig }()
	flockAttempt = func(fd int, how int) error {
		calls++
		if calls == 1 {
			return unix.EINTR
		}
		return orig(fd, how)
	}

	lock, err := AcquireSpecCloseLock(tempDir, "SPEC-INTR-001")
	if err != nil {
		t.Fatalf("acquire after an interrupted attempt failed: %v", err)
	}
	defer func() { _ = lock.Release() }()
	if calls < 2 {
		t.Fatalf("flock attempted %d times, want a retry after EINTR", calls)
	}
}

// TestAcquireSpecCloseLock_ForeignErrnoNotHeld (card t1291): a non-EWOULDBLOCK
// flock failure must surface as its own error, never as ErrSpecCloseLockHeld —
// mislabeling it makes the closer refuse a legitimate close with "another
// close operation in progress". Same no-parallel rule as the test above.
func TestAcquireSpecCloseLock_ForeignErrnoNotHeld(t *testing.T) {
	tempDir := t.TempDir()

	orig := flockAttempt
	defer func() { flockAttempt = orig }()
	flockAttempt = func(fd int, how int) error { return unix.EIO }

	_, err := AcquireSpecCloseLock(tempDir, "SPEC-EIO-001")
	if err == nil {
		t.Fatal("expected an error from a failing flock, got nil")
	}
	if IsLockHeldError(err) {
		t.Fatalf("foreign errno mislabeled as lock held: %v", err)
	}
}

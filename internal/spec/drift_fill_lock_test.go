package spec

// drift_fill_lock_test.go — AC-DCF-005: single-flight under concurrency, and no
// wedge after abnormal exit.
//
// The lock is a build-tagged pair (drift_fill_lock_unix.go /
// drift_fill_lock_windows.go), so this file is written against the
// platform-neutral WithDriftFillLock interface and runs under BOTH build tags.
// The CI matrix's windows/amd64 job is where the Windows half is observed; a
// build-tag split that left the Windows side with no test would be an empty
// sweep and is explicitly not acceptable.

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestWithDriftFillLockSerialisesConcurrentFills is AC-DCF-005 clause (a): two
// fills released by a common barrier, one compute, and — the clause that
// separates a NON-BLOCKING lock from a blocking one — the loser returns before
// the winner has finished.
func TestWithDriftFillLockSerialisesConcurrentFills(t *testing.T) {
	dir := t.TempDir()

	var computes atomic.Int32
	const hold = 300 * time.Millisecond

	var barrier sync.WaitGroup
	barrier.Add(1)

	type outcome struct {
		ran        bool
		returnedAt time.Time
	}
	results := make([]outcome, 2)
	var winnerFinishedAt atomic.Int64

	var wg sync.WaitGroup
	for i := range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			barrier.Wait()
			ran, err := WithDriftFillLock(dir, func() {
				computes.Add(1)
				time.Sleep(hold)
				winnerFinishedAt.Store(time.Now().UnixNano())
			})
			if err != nil && ran {
				t.Errorf("WithDriftFillLock ran with error: %v", err)
			}
			results[i] = outcome{ran: ran, returnedAt: time.Now()}
		}()
	}
	barrier.Done()
	wg.Wait()

	if got := computes.Load(); got != 1 {
		t.Fatalf("compute counter = %d, want exactly 1", got)
	}
	ranCount := 0
	var loser outcome
	for _, r := range results {
		if r.ran {
			ranCount++
		} else {
			loser = r
		}
	}
	if ranCount != 1 {
		t.Fatalf("ran=true count = %d, want exactly 1", ranCount)
	}
	finished := time.Unix(0, winnerFinishedAt.Load())
	if !loser.returnedAt.Before(finished) {
		t.Fatalf("loser returned at %v, winner finished at %v — the loser WAITED, so the lock is blocking, not skip-on-contention",
			loser.returnedAt, finished)
	}
}

// helperDriftFillLockEnv is the channel between this test and its re-executed
// child: the value is the project dir whose fill lock the child holds.
const helperDriftFillLockEnv = "MOAI_TEST_DRIFT_FILL_LOCK_DIR"

// TestDriftFillLockHelperHold is not a test. Re-executed as a child by
// TestWithDriftFillLockSurvivesAbnormalHolderExit, it takes the fill lock,
// announces that it holds it, and waits.
//
// The child BOUNDS ITSELF: it exits on its own after helperHoldBound whatever
// the parent does, and -test.timeout caps it from outside. There is no trailing
// kill relied on for cleanup — the parent's SIGKILL is the SUBJECT of the test
// (abnormal termination), not its cleanup mechanism.
func TestDriftFillLockHelperHold(t *testing.T) {
	dir := os.Getenv(helperDriftFillLockEnv)
	if dir == "" {
		t.Skip("helper process only")
	}
	const helperHoldBound = 30 * time.Second
	ran, err := WithDriftFillLock(dir, func() {
		if werr := os.WriteFile(filepath.Join(dir, "held"), []byte("1"), 0o644); werr != nil {
			t.Errorf("write held marker: %v", werr)
		}
		time.Sleep(helperHoldBound)
	})
	if err != nil {
		t.Fatalf("helper WithDriftFillLock: %v", err)
	}
	if !ran {
		t.Fatal("helper could not acquire the fill lock")
	}
}

// TestWithDriftFillLockSurvivesAbnormalHolderExit is AC-DCF-005 clause (b), and
// it also supplies the CROSS-PROCESS evidence clause (a) cannot: a separate
// process holds the lock, this process is excluded by it, the holder is killed
// without any chance to clean up, and the next acquire succeeds.
//
// A pid-file lock passes nothing here: it survives SIGKILL and wedges every
// later fill. An OS-handle lock is released by process exit, including
// abnormal termination, which is the property REQ-DCF-008 states.
func TestWithDriftFillLockSurvivesAbnormalHolderExit(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}

	cmd := exec.Command(os.Args[0], "-test.run=^TestDriftFillLockHelperHold$", "-test.timeout=60s")
	cmd.Env = append(os.Environ(), helperDriftFillLockEnv+"="+dir)
	if err := cmd.Start(); err != nil {
		t.Fatalf("start helper: %v", err)
	}
	killed := false
	t.Cleanup(func() {
		if !killed {
			_ = cmd.Process.Kill()
		}
		_ = cmd.Wait()
	})

	// Wait for the child to announce it holds the lock.
	marker := filepath.Join(dir, "held")
	deadline := time.Now().Add(20 * time.Second)
	for {
		if _, err := os.Stat(marker); err == nil {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("helper never announced holding the fill lock")
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Cross-process exclusion: this process must be turned away, not blocked.
	start := time.Now()
	ran, err := WithDriftFillLock(dir, func() { t.Error("acquired a lock another PROCESS holds") })
	if err != nil {
		t.Fatalf("contended acquire returned an error: %v", err)
	}
	if ran {
		t.Fatal("acquired the fill lock while another process held it — the lock is not cross-process")
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("contended acquire took %v — it waited instead of skipping", elapsed)
	}

	// Abnormal termination: no unlock, no deferred cleanup, no exit handler.
	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("kill helper: %v", err)
	}
	killed = true
	_ = cmd.Wait()

	// The lock must be free. Retry briefly: the kernel releases the flock when
	// the process is reaped, which is not instantaneous.
	deadline = time.Now().Add(10 * time.Second)
	for {
		ran, err = WithDriftFillLock(dir, func() {})
		if err != nil {
			t.Fatalf("post-kill acquire returned an error: %v", err)
		}
		if ran {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("the fill lock stayed held after the holder was killed — a terminated fill wedges every later fill")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDriftFillLockIsHeldByAnOSHandle is the structural half of REQ-DCF-008: the
// lock is an OS handle released by process exit, on BOTH platforms. A pid-file
// or a claim-stamp file would satisfy the behavioural clauses above on the
// happy path and wedge on SIGKILL, so the mechanism is asserted directly.
func TestDriftFillLockIsHeldByAnOSHandle(t *testing.T) {
	cases := []struct {
		file  string
		wants []string
	}{
		{file: "drift_fill_lock_unix.go", wants: []string{"//go:build !windows", "unix.Flock", "LOCK_EX", "LOCK_NB"}},
		{file: "drift_fill_lock_windows.go", wants: []string{"//go:build windows", "LockFileEx", "LOCKFILE_EXCLUSIVE_LOCK", "LOCKFILE_FAIL_IMMEDIATELY"}},
	}
	for _, tc := range cases {
		raw, err := os.ReadFile(tc.file)
		if err != nil {
			t.Fatalf("read %s: %v — the lock pair must exist on both platforms", tc.file, err)
		}
		for _, want := range tc.wants {
			if !strings.Contains(string(raw), want) {
				t.Errorf("%s does not contain %q", tc.file, want)
			}
		}
	}
}

// TestDriftFillLockPathIsBesideTheCache pins the lock location so a rename
// cannot silently split two moai binaries onto two different locks.
func TestDriftFillLockPathIsBesideTheCache(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, ".moai", "state", "drift-cache.lock")
	if got := DriftFillLockPath(dir); got != want {
		t.Fatalf("DriftFillLockPath = %q, want %q", got, want)
	}
}

// TestWithDriftFillLockLeavesTheCacheReadable is a guard against the lock file
// and the cache file being confused for one another: they are distinct paths,
// and holding the lock never touches the cache payload.
func TestWithDriftFillLockLeavesTheCacheReadable(t *testing.T) {
	dir := t.TempDir()
	saveDriftCache(dir, "head-A", &DriftReport{Records: []DriftRecord{}, Count: 0})
	ran, err := WithDriftFillLock(dir, func() {})
	if err != nil || !ran {
		t.Fatalf("WithDriftFillLock ran=%v err=%v, want ran=true err=nil", ran, err)
	}
	raw, err := os.ReadFile(driftCachePath(dir))
	if err != nil {
		t.Fatalf("read cache: %v", err)
	}
	var payload driftCacheFile
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("cache no longer parses after the lock was taken: %v", err)
	}
}

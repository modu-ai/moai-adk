//go:build !windows

// The post-close gate finding (card t1595, SPEC-DISPATCH-INTEGRITY-001):
// an abandoned fold must not publish. Unix-only — the poll-gap fixture is
// a FIFO.
package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestReviewFindingAbandonedFoldDoesNotPublish pins the last-observable-
// moment guard: the bounded reads of the final byte comparison can complete
// INSIDE a poll interval and return normally after the caller's deadline
// has passed — without a re-check immediately before the rename, an
// abandoned fold's rename published. The fixture flips abandonment at
// t0+107ms and delivers the plan-time bytes through the FIFO at
// read-start+195ms: the poll at read-start+100ms still sees the step
// un-abandoned, the read completes before the next tick with matching
// bytes, and control reaches the rename with forbidden already true.
func TestReviewFindingAbandonedFoldDoesNotPublish(t *testing.T) {
	dir := t.TempDir()
	name := "MEMORY.md"
	old := []byte("original\n")
	if err := os.WriteFile(filepath.Join(dir, name), old, 0o644); err != nil {
		t.Fatal(err)
	}
	t0 := time.Now()
	// forbidden flips at t0+107ms: after the pre-comparison guard check
	// (which runs within ~1ms), inside the final comparison's read.
	forbidden := func() bool { return time.Since(t0) > 107*time.Millisecond }
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })
	t.Cleanup(func() { _ = os.Remove(filepath.Join(dir, name)) })
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage != "bytes-done" {
			return
		}
		// The pre-rename position of THIS write: swap the file for a FIFO
		// whose writer delivers the plan-time bytes at read-start+195ms —
		// inside the same poll gap (the poll at +100ms saw false; the next
		// would be at +200ms, after the read has completed).
		memoryFoldSeam = foldTestSeam{}
		fifoPath := filepath.Join(dir, name)
		if err := os.Remove(fifoPath); err != nil {
			panic(err)
		}
		if err := mkfifoForTest(fifoPath); err != nil {
			panic(err)
		}
		go func() {
			w, err := os.OpenFile(fifoPath, os.O_WRONLY, 0)
			if err != nil {
				return
			}
			defer func() { _ = w.Close() }()
			time.Sleep(195 * time.Millisecond)
			_, _ = w.Write(old)
		}()
	}
	err := atomicWriteFoldFile(dir, name, []byte("fold output\n"), old, nil, forbidden, nil)
	if err == nil {
		// The publication signature the gate's repro observed: the rename
		// landed under an already-abandoned step.
		got, rerr := os.ReadFile(filepath.Join(dir, name))
		if rerr != nil {
			t.Fatalf("the abandoned fold published (err=<nil>); reading back what it published also failed: %v", rerr)
		}
		t.Fatalf("the abandoned fold published: err=<nil> final=%q", got)
	}
	if !strings.Contains(err.Error(), "abandoned") {
		t.Errorf("the refusal is not the abandonment error: %v", err)
	}
}

// TestReviewFindingLockFileRefusesSymlink is the post-close gate's finding
// 4: the lock file must not follow symlinks. A `.moai-store-lock` that is a
// symlink pointed O_CREATE outside the store — the gate's repro locked
// another repository's `.git/index.lock` into existence. The acquisition
// must refuse, and the symlink target must stay untouched.
func TestReviewFindingLockFileRefusesSymlink(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "outside.lock")
	lockPath := filepath.Join(dir, foldLockName)
	if err := os.Symlink(outside, lockPath); err != nil {
		t.Fatal(err)
	}
	release, err := acquireFoldStoreLock(dir, nil)
	if err == nil {
		release()
		t.Fatal("the lock acquisition followed a symlinked lock file and locked the target outside the store")
	}
	t.Logf("symlinked lock refused: %v", err)
	if _, statErr := os.Lstat(outside); statErr == nil {
		t.Errorf("the symlink target was created outside the store: %v", statErr)
	}
	// A regular lock file still acquires: the refusal is the symlink's, not
	// the open path's.
	if err := os.Remove(lockPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(lockPath, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	release2, err := acquireFoldStoreLock(dir, nil)
	if err != nil {
		t.Fatalf("a regular lock file must still acquire: %v", err)
	}
	release2()
}

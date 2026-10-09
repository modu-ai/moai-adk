//go:build !windows

// The FIFO-based run-gate finding 2 characterization (SPEC-DISPATCH-
// INTEGRITY-001 M4): an abandoned bounded fold must release the store lock.
// Unix-only — the blocked-read fixture is a FIFO (AC-MFB-008 vii, plan B1).
package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
)

// TestReviewFindingAbandonedFoldReleasesLockOnApplyReads is the
// post-report gate's finding 1: the abandonment cancellation must cover
// EVERY read inside the lock hold, not just the initial snapshot. A fold
// whose in-apply validation read blocks (here: MEMORY.md becomes a
// blocking FIFO mid-apply, at the archive write's seam) must release the
// store lock when its caller abandons it — a worker parked in a later
// read pins the lock past the caller's timeout exactly like the initial
// read did.
func TestReviewFindingAbandonedFoldReleasesLockOnApplyReads(t *testing.T) {
	root, _ := todoFixture(t)
	cfgDir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	memDir := filepath.Join(cfgDir, "projects", memoryProjectSlug(root), "memory")
	if err := os.MkdirAll(memDir, 0o700); err != nil {
		t.Fatalf("memory dir: %v", err)
	}
	copyFixtureStore(t, memDir)
	prev := memoryFoldSeam
	t.Cleanup(func() { memoryFoldSeam = prev })
	// The fold worker is joined BEFORE the seam restore above runs: the
	// worker's probe writes touch memoryFoldSeam, and the caller's timeout
	// return creates no happens-before edge with the worker — without the
	// join, the cleanup restore raced the worker's accesses under -race
	// (gate-reproduced). foldOnDoneExit is closed by the worker's own
	// deferred signal when it exits.
	savedExit := foldOnDoneExit
	exit := make(chan struct{})
	foldOnDoneExit = exit
	t.Cleanup(func() { foldOnDoneExit = savedExit })
	var fifo *wireFIFO
	t.Cleanup(func() {
		if fifo != nil {
			fifo.release() // unblock the worker so it can exit
		}
		<-exit // the worker is gone: its seam accesses happened-before the restore
	})
	fifoCh := make(chan *wireFIFO, 1)
	memoryFoldSeam.orderProbe = func(stage string) {
		if stage != "bytes-done" {
			return
		}
		// The archive write's pre-rename position: swap MEMORY.md — the
		// file the apply's NEXT reads must open — for a blocking FIFO.
		memoryFoldSeam = foldTestSeam{}
		if fx, err := blockOnReadRaw(filepath.Join(memDir, "MEMORY.md")); err == nil {
			fifoCh <- fx
		}
	}
	savedBound := memoryFoldOnDoneBound
	memoryFoldOnDoneBound = 300 * time.Millisecond
	t.Cleanup(func() { memoryFoldOnDoneBound = savedBound })

	foldClosedCardMemory("t9001") // returns at the bound; the worker parks in the apply's read
	select {
	case fx := <-fifoCh:
		fifo = fx // released — and the worker joined — by the cleanup above
	default:
	}

	probe := make(chan error, 1)
	go func() {
		r, err := acquireFoldStoreLock(memDir, nil)
		if err == nil {
			defer r()
		}
		probe <- err
	}()
	select {
	case err := <-probe:
		if err != nil {
			t.Fatalf("the follow-up locker failed to acquire the abandoned store's lock: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the store lock stayed held after the bounded fold was abandoned — a blocking read inside the apply pins the lock past its caller's timeout")
	}
}

// TestReviewFindingAbandonedFoldReleasesStoreLock is the run-gate finding 2:
// when the bounded auto-fold worker blocks on the store read and its caller
// abandons it, the worker must RELEASE the store lock. A worker that keeps
// holding the lock while blocked pinned every follow-up fold on the store
// until the blocked read happened to complete — the caller's timeout
// released nothing.
func TestReviewFindingAbandonedFoldReleasesStoreLock(t *testing.T) {
	root, _ := todoFixture(t)
	cfgDir := t.TempDir()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("USERPROFILE", t.TempDir())
	t.Setenv(config.EnvClaudeConfigDir, cfgDir)
	t.Setenv(config.EnvMemoryFoldOnDone, "1")
	memDir := filepath.Join(cfgDir, "projects", memoryProjectSlug(root), "memory")
	if err := os.MkdirAll(memDir, 0o700); err != nil {
		t.Fatalf("memory dir: %v", err)
	}
	copyFixtureStore(t, memDir)
	// The index read blocks until a writer appears: the fold's worker
	// acquires the store lock and parks inside the snapshot read.
	fifo := blockOnRead(t, filepath.Join(memDir, "MEMORY.md"))
	savedBound := memoryFoldOnDoneBound
	memoryFoldOnDoneBound = 300 * time.Millisecond
	t.Cleanup(func() { memoryFoldOnDoneBound = savedBound })

	foldClosedCardMemory("t9001") // returns at the bound; the worker stays blocked in the read

	// The abandoned worker must not keep the store's lock: a follow-up
	// fold acquires it within a bounded wait.
	probe := make(chan error, 1)
	go func() {
		r, err := acquireFoldStoreLock(memDir, nil)
		if err == nil {
			defer r()
		}
		probe <- err
	}()
	select {
	case err := <-probe:
		if err != nil {
			t.Fatalf("the follow-up locker failed to acquire the abandoned store's lock: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("the store lock stayed held after the bounded fold was abandoned — the blocked worker pins the lock past its caller's timeout")
	}
	// The FIFO's release (t.Cleanup) unblocks the worker's read; the
	// abandoned step then refuses its write through the forbidden check and
	// the store keeps its contents.
	_ = fifo
}

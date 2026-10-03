// factory_lease_worktree_test.go — SPEC-FACTORY-ATOMIC-LEASE-001 (card t1458)
// AC-FAL-006, the step lock's bounded wait and its derivation (plan D3). The
// concurrency half of AC-FAL-006, which needs a `git` shim on PATH, lives in
// factory_lease_unix_test.go.
package cli

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// TestFactoryEnsureCardWorktreeStepLockBounded — with the worktree-step lock
// held by the test, a step run with factoryWorktreeStepWait shortened by the
// test returns the wait error within that wait plus 500 ms, creates no
// directory and leaves the card's record row unchanged.
func TestFactoryEnsureCardWorktreeStepLockBounded(t *testing.T) {
	root, store := nmBase(t, factory.BacklogStateQueued)
	nmSetText(t, store, "t1", "alpha step lock probe")
	far := "2099-01-01T00:00:00Z"
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: far, Stage: homestate.CardRun})
	before := fcCard(t, root, "t1")

	release, err := factory.AcquireFactoryStepLock(root, time.Second)
	if err != nil {
		t.Fatalf("the test could not take the worktree-step lock: %v", err)
	}
	t.Cleanup(func() { _ = release() })

	const wait = 300 * time.Millisecond
	prev := factoryWorktreeStepWait
	factoryWorktreeStepWait = wait
	t.Cleanup(func() { factoryWorktreeStepWait = prev })

	start := time.Now()
	_, created, stepErr := factoryEnsureCardWorktree(context.Background(), root, fcRun, before, "lane-1", io.Discard)
	elapsed := time.Since(start)
	t.Logf("step under a held step lock: elapsed=%s created=%v err=%v", elapsed, created, stepErr)
	if stepErr == nil {
		t.Errorf("the step succeeded while the step lock was held, want the wait error")
	}
	if elapsed > wait+flMargin {
		t.Errorf("the step returned after %s, want within %s (the wait plus 500 ms)", elapsed, wait+flMargin)
	}
	if _, statErr := os.Lstat(filepath.Join(root, sessionWorktreeSubdir, "t1")); statErr == nil {
		t.Errorf("the timed-out step created the card's worktree directory")
	}
	if after := fcCard(t, root, "t1"); after.Version != before.Version || after.WorktreePath != before.WorktreePath || after.State != before.State {
		t.Errorf("the record row changed: %+v -> %+v", before, after)
	}
}

// TestFactoryWorktreeStepWaitDerivation — the step lock's default wait is not
// smaller than lanes (10) x the worst observed step (2.9 s, ledger L7) x
// headroom (2), and the package variable the step reads starts at it.
func TestFactoryWorktreeStepWaitDerivation(t *testing.T) {
	const lanes, worstStep, headroom = 10, 2900 * time.Millisecond, 2
	if min := time.Duration(lanes*headroom) * worstStep; factoryWorktreeStepWaitDefault < min {
		t.Errorf("factoryWorktreeStepWaitDefault = %s < lanes x worst step x headroom = %s", factoryWorktreeStepWaitDefault, min)
	}
	if factoryWorktreeStepWait != factoryWorktreeStepWaitDefault {
		t.Errorf("factoryWorktreeStepWait = %s, want it initialised from factoryWorktreeStepWaitDefault (%s)", factoryWorktreeStepWait, factoryWorktreeStepWaitDefault)
	}
}

package cli

// factory_leaselost_t2013_test.go — card t2013, Git-AFF S2: the `lease-lost`
// stop of the release route (factory_card.go, after factoryFastCheckKeepingLease).
// The fast required check PASSED, but the card lease could not be renewed while
// it ran, so a merge would start on a lapsed lease. complete must hold the card,
// merge nothing, and release the window it acquired.
//
// The fixture is the held-open check of
// TestAFFS2CompleteKeepsCardLeaseAliveWhileTheCheckRuns: the card's test stays
// running until the test releases it. The lease is made unrenewable from inside
// the renewal itself (the factoryRenewBeforeWrite seam, the first renewal that
// runs after the card's test has started), so the loss happens while the check
// is open and no assertion depends on how long anything takes.

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestAFFS2CompleteHoldsLeaseLostWhenTheLeaseIsLostDuringTheCheck(t *testing.T) {
	f := affs2New(t, affs2Opts{slowTest: true})

	prev := factoryCheckLeaseRenewEvery
	factoryCheckLeaseRenewEvery = 100 * time.Millisecond
	t.Cleanup(func() { factoryCheckLeaseRenewEvery = prev })

	// The merge step must not be entered on this path.
	var mergeStepEntered atomic.Bool
	factoryCompleteMergeSeamsHook = func(*factory.MergeStepSeams) { mergeStepEntered.Store(true) }
	t.Cleanup(func() { factoryCompleteMergeSeamsHook = nil })

	// The first renewal that runs once the card's test is running lapses the
	// stored lease immediately before its own write, so that renewal is refused
	// with an expired lease and no later renewal can succeed. Runs on the renewal
	// goroutine, so it reports with Errorf, never Fatal.
	var lost atomic.Bool
	started := filepath.Join(f.gate, "started")
	prevHook := factoryRenewBeforeWrite
	factoryRenewBeforeWrite = func() {
		if !affs2Exists(started) || !lost.CompareAndSwap(false, true) {
			return
		}
		db, err := homestate.OpenFactory(f.root)
		if err != nil {
			t.Errorf("open the factory database to lapse the lease: %v", err)
			return
		}
		defer func() { _ = db.Close() }()
		past := time.Now().Add(-time.Hour).UTC().Format(time.RFC3339Nano)
		if _, err := db.DB.ExecContext(context.Background(), `UPDATE cards SET lease_expires_at=? WHERE run_id=? AND card_id=?`, past, fcRun, "t1"); err != nil {
			t.Errorf("lapse the card lease: %v", err)
		}
	}
	t.Cleanup(func() { factoryRenewBeforeWrite = prevHook })

	done := make(chan error, 1)
	go func() {
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		done <- err
	}()
	released := false
	release := func() {
		if released {
			return
		}
		released = true
		if err := os.WriteFile(filepath.Join(f.gate, "release"), []byte("1"), 0o644); err != nil {
			t.Errorf("release the check: %v", err)
		}
	}
	returned := false
	defer func() {
		release()
		if returned {
			return
		}
		select {
		case <-done:
		case <-time.After(affs2LeaseWait):
			t.Error("complete did not return after the check was released")
		}
	}()

	// The check is running, then the lease is lost while it is held open.
	if !affs2Eventually(affs2LeaseWait, func() bool { return lost.Load() }, done) {
		select {
		case err := <-done:
			returned = true
			t.Fatalf("complete returned before the lease was lost during the check: %v", err)
		default:
		}
		t.Fatal("the card lease was never lapsed within the wait bound")
	}

	release()
	var err error
	select {
	case err = <-done:
		returned = true
	case <-time.After(affs2LeaseWait):
		t.Fatal("complete did not return after the check was released")
	}

	// (1) complete does not succeed and names the lost lease.
	if err == nil {
		t.Fatal("complete succeeded although the card lease was lost while the check ran")
	}
	if !strings.Contains(err.Error(), "lease could not be kept alive") {
		t.Errorf("complete error = %v, want the lease-could-not-be-kept-alive refusal", err)
	}
	holds := affs2Holds(t, f.root)
	if len(holds) != 1 {
		t.Fatalf("hold records = %d, want 1: %v", len(holds), holds)
	}
	if h := holds[0]; h["reason"] != "lease-lost" || h["card_head"] != f.cardHead || h["check_key"] == nil || h["check_key"] == "" {
		t.Errorf("hold = %v, want reason lease-lost, card_head %s and the check_key of the check that passed", h, f.cardHead)
	}

	// (2) nothing merged: the merge step was never entered, and the release
	// branch, main, the remotes and the card branch kept their SHAs.
	if mergeStepEntered.Load() {
		t.Error("the merge step was entered although the card lease was lost")
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got != f.relBefore {
		t.Errorf("release branch moved %s -> %s", f.relBefore, got)
	}
	if got := affs2Git(t, f.cardWT, "rev-parse", "HEAD"); got != f.cardHead {
		t.Errorf("card branch moved %s -> %s", f.cardHead, got)
	}
	affs2Untouched(t, f, "after the lease was lost")

	// (3) the window complete acquired is released.
	affs2WindowReleased(t, f, "after the lease was lost")

	// (4) the ledger records a hold, not a transition: the card is still merge-ready at its version.
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMergeReady || c.Version != 1 {
		t.Errorf("card = %s v%d, want merge-ready v1 (a hold record, not a transition)", c.State, c.Version)
	}
}

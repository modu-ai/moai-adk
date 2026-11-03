package cli

// factory_aff_s2_lease_t2013_test.go — sync audit 6 finding C5: the fast
// required check can outlast the card lease, and a lapsed lease cannot be
// renewed afterwards, so the card lease must be kept alive WHILE the check runs.
// Real git, the real merge step, and the real `moai verify run` code path.
//
// The first form of this test set the card lease to lapse two seconds after the
// test began and ran a check that slept three. Everything complete does before
// the check starts (the step-1 gates, the window, the re-measure record, the
// working-copy probe, a handful of git processes) counted against those two
// seconds: about 0.9s on a quiet machine, and past 2s under load, so the lease
// had already lapsed before the first renewal could run and the test failed
// with "card lease expired" (or, one stage earlier, with the REQ-MWQ-019 step 1
// refusal). The fixture raced the wall clock, not the product.
//
// The check is now held open by the test instead of timed: the card's test stays
// running until the test releases it, and the test releases it only after it has
// observed renewals happen while the check is running and move the stored lease
// forward. No assertion depends on how long anything takes.

import (
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// affs2LeaseWait bounds each wait of the test; it only turns a hang into a
// failure and is never compared with how long the check takes.
const affs2LeaseWait = 90 * time.Second

func TestAFFS2CompleteKeepsCardLeaseAliveWhileTheCheckRuns(t *testing.T) {
	f := affs2New(t, affs2Opts{slowTest: true})

	prev := factoryCheckLeaseRenewEvery
	factoryCheckLeaseRenewEvery = 100 * time.Millisecond
	t.Cleanup(func() { factoryCheckLeaseRenewEvery = prev })

	// Every renewal write goes through factoryRenewBeforeWrite; the count lets
	// the test tell renewals that happen while the check is held open.
	var renewals atomic.Int64
	prevHook := factoryRenewBeforeWrite
	factoryRenewBeforeWrite = func() { renewals.Add(1) }
	t.Cleanup(func() { factoryRenewBeforeWrite = prevHook })

	done := make(chan error, 1)
	go func() {
		_, _, err := runFactory(t, "complete", "t1", "--run", fcRun)
		done <- err
	}()
	// Whatever happens below, the held-open check must be released so complete returns.
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
	returned := false // complete's result has been read from done
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

	// Wait until the card's test is running: the check is now in progress, and the
	// renewal loop has been running since the check began.
	if !affs2Eventually(affs2LeaseWait, func() bool { return affs2Exists(filepath.Join(f.gate, "started")) }, done) {
		select {
		case err := <-done:
			returned = true
			t.Fatalf("complete returned before the card's test started: %v", err)
		default:
		}
		t.Fatal("the card's test never started within the wait bound")
	}
	atStart := affs2LeaseExpiry(t, f.root)
	renewalsAtStart := renewals.Load()

	// Held open, the check must keep being renewed, and the renewals must move the
	// stored expiry forward. Three renewals after the test started prove the
	// renewal loop outlives the check's first moments; the third hook call can
	// only follow the second renewal's write, so the stored expiry has moved.
	if !affs2Eventually(affs2LeaseWait, func() bool { return renewals.Load() >= renewalsAtStart+3 }, done) {
		t.Fatalf("renewals while the check was held open = %d, want at least 3 (the lease is not renewed during the check)", renewals.Load()-renewalsAtStart)
	}
	if during := affs2LeaseExpiry(t, f.root); !during.After(atStart) {
		t.Errorf("lease expiry %s while the check was held open is not after %s seen when it started: renewals did not extend the lease", during, atStart)
	}

	release()
	select {
	case err := <-done:
		returned = true
		if err != nil {
			t.Fatalf("complete: %v (the card lease lapsed during the check, so the merge was refused)", err)
		}
	case <-time.After(affs2LeaseWait):
		t.Fatal("complete did not return after the check was released")
	}
	c := fcCard(t, f.root, "t1")
	if c.State != homestate.CardMergedLocal {
		t.Fatalf("card = %s, want merged-local", c.State)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got == f.relBefore || got != c.MergeSHA {
		t.Errorf("release HEAD = %s (before %s), want the card's merge %s", got, f.relBefore, c.MergeSHA)
	}
}

// affs2Exists reports whether path exists.
func affs2Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// affs2Eventually polls cond until it holds. It gives up (false) when the wait
// bound passes or when complete has already returned, since the condition can
// then no longer come true.
func affs2Eventually(wait time.Duration, cond func() bool, done chan error) bool {
	deadline := time.Now().Add(wait)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		select {
		case err := <-done:
			// complete returned early; put its result back (the channel holds one)
			// for the caller that reads it after the release.
			done <- err
			return cond()
		case <-time.After(20 * time.Millisecond):
		}
	}
	return false
}

// affs2LeaseExpiry reads card t1's stored lease expiry.
func affs2LeaseExpiry(t *testing.T, root string) time.Time {
	t.Helper()
	c := fcCard(t, root, "t1")
	expires, err := time.Parse(time.RFC3339Nano, c.LeaseExpiresAt)
	if err != nil {
		t.Fatalf("card t1 lease expiry %q: %v", c.LeaseExpiresAt, err)
	}
	return expires
}

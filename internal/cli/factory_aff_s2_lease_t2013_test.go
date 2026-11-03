package cli

// factory_aff_s2_lease_t2013_test.go — sync audit 6 finding C5: the fast
// required check can outlast the card lease, and a lapsed lease cannot be
// renewed afterwards, so the card lease must be kept alive WHILE the check runs.
// Real git, the real merge step, and the real `moai verify run` code path; the
// card lease is set to lapse before the check can finish.

import (
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestAFFS2CompleteKeepsCardLeaseAliveWhileTheCheckRuns(t *testing.T) {
	f := affs2New(t, affs2Opts{slowTest: true})

	// The card lease lapses two seconds from now; the check takes longer than
	// that (the card's test sleeps slowTestPause seconds).
	db := fcOpen(t, f.root)
	expires := time.Now().UTC().Add(2 * time.Second).Format(time.RFC3339Nano)
	if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at=? WHERE card_id='t1'`, expires); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()

	prev := factoryCheckLeaseRenewEvery
	factoryCheckLeaseRenewEvery = 100 * time.Millisecond
	t.Cleanup(func() { factoryCheckLeaseRenewEvery = prev })

	if _, _, err := runFactory(t, "complete", "t1", "--run", fcRun); err != nil {
		t.Fatalf("complete: %v (the card lease lapsed during the check, so the merge was refused)", err)
	}
	c := fcCard(t, f.root, "t1")
	if c.State != homestate.CardMergedLocal {
		t.Fatalf("card = %s, want merged-local", c.State)
	}
	if got := affs2Git(t, f.relWT, "rev-parse", "HEAD"); got == f.relBefore || got != c.MergeSHA {
		t.Errorf("release HEAD = %s (before %s), want the card's merge %s", got, f.relBefore, c.MergeSHA)
	}
}

// factory_next_expired_staged_t2116_test.go — card t2116: the registered owner
// of a card parked in a staged state (plan … merge-ready) whose lease EXPIRED
// re-leases it with `moai factory next --card <id>`, exactly as card t2013
// allows for a `leased` card: the card comes back `leased` to its owner with
// stage, worktree and evidence kept, and the verb never touches the worktree.
// Another lane, the owner before expiry, and an interrupted merge keep
// refusing. Fixtures are the real temp record, queue and git repository of
// factory_next_expired_owner_t2013_test.go; time is the factoryCardNow seam.

package cli

import (
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

var t2116StagedStates = []string{
	homestate.CardPlan, homestate.CardPlanAudit, homestate.CardRun,
	homestate.CardSync, homestate.CardSyncAudit, homestate.CardMergeReady,
}

// t2116Setup is reacqSetup with the card parked in state (its stage named like
// the state).
func t2116Setup(t *testing.T, state string, expiry time.Time) reacqFixture {
	t.Helper()
	root, store := nmBase(t, factory.BacklogStatePicked)
	wt := reacqWorktree(t)
	fcPlace(t, root, homestate.Card{
		CardID: "t1", State: state, OwnerLabel: "lane-1", LeaseHolder: "lane-1",
		Stage: state, Version: 3,
		LeaseExpiresAt: expiry.Format(time.RFC3339Nano), HeartbeatAt: expiry.Add(-15 * time.Minute).Format(time.RFC3339Nano),
		WorktreePath: wt, EvidenceSHA: reacqEvidenceSHA, EvidencePath: reacqEvidencePath,
	})
	return reacqFixture{root: root, store: store, wt: wt}
}

func t2116Expired(t *testing.T, state string) reacqFixture {
	return t2116Setup(t, state, fcNow.Add(-10*time.Minute))
}

// The owner's `next --card` on its expired staged card exits 0 and re-leases it
// in one step, keeping the stage.
func TestT2116OwnerNextReleasesExpiredStagedLease(t *testing.T) {
	for _, state := range t2116StagedStates {
		t.Run(state, func(t *testing.T) {
			fx := t2116Expired(t, state)
			nmLaneEnv(t, "lane-1", "")
			before := fcCard(t, fx.root, "t1")

			out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
			if err != nil {
				t.Fatalf("owner next --card t1 on its expired %s lease: %v (stderr %q), want exit 0", state, err, stderr)
			}
			if nmLeasedHead(out) == "" {
				t.Errorf("stdout = %q, want the leased card's head line", out)
			}
			reacqAssertReacquired(t, fx, before)
			if after := fcCard(t, fx.root, "t1"); after.Stage != state {
				t.Errorf("stage = %q after the re-lease, want %q kept", after.Stage, state)
			}
		})
	}
}

// Another lane is refused and nothing is written.
func TestT2116OtherLaneNextRefusedOnExpiredStagedLease(t *testing.T) {
	for _, state := range t2116StagedStates {
		t.Run(state, func(t *testing.T) {
			fx := t2116Expired(t, state)
			nmLaneEnv(t, "lane-2", "")
			before := nmSnapshot(t, fx.root, fx.store)
			rowsBefore := fcRows(t, fx.root)

			out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
			nmAssertRefused(t, out, stderr, err, "owned")
			if after := nmSnapshot(t, fx.root, fx.store); after != before {
				t.Errorf("a refused nomination changed the queue or the record:\nbefore:\n%s\nafter:\n%s", before, after)
			}
			if after := fcRows(t, fx.root); after != rowsBefore {
				t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
			}
		})
	}
}

// The owner BEFORE expiry is refused as in flight, nothing written.
func TestT2116OwnerNextRefusedBeforeExpiryStaged(t *testing.T) {
	for _, state := range t2116StagedStates {
		t.Run(state, func(t *testing.T) {
			fx := t2116Setup(t, state, fcNow.Add(10*time.Minute))
			nmLaneEnv(t, "lane-1", "")
			rowsBefore := fcRows(t, fx.root)

			out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
			nmAssertRefused(t, out, stderr, err, "owned")
			if !strings.Contains(stderr, "in flight") {
				t.Errorf("stderr = %q, want the in-flight detail", stderr)
			}
			if after := fcRows(t, fx.root); after != rowsBefore {
				t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
			}
		})
	}
}

// An expired lease on an interrupted merge is NOT re-leased by the owner.
func TestT2116OwnerNextRefusedOnExpiredMerging(t *testing.T) {
	fx := t2116Expired(t, homestate.CardMerging)
	nmLaneEnv(t, "lane-1", "")
	rowsBefore := fcRows(t, fx.root)

	out, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1")
	nmAssertRefused(t, out, stderr, err, "owned")
	if after := fcRows(t, fx.root); after != rowsBefore {
		t.Errorf("a refused nomination changed the rows or events:\nbefore:\n%s\nafter:\n%s", rowsBefore, after)
	}
}

// The verb never touches the card's worktree, uncommitted work included.
func TestT2116OwnerNextLeavesStagedWorktreeUntouched(t *testing.T) {
	fx := t2116Expired(t, homestate.CardSync)
	nmLaneEnv(t, "lane-1", "")
	before := reacqWorktreeSnapshot(t, fx.wt)

	if _, stderr, err := qasRunNext(t, "--run", fcRun, "--card", "t1"); err != nil {
		t.Fatalf("owner next --card t1: %v (stderr %q), want exit 0 before the worktree is compared", err, stderr)
	}
	if after := reacqWorktreeSnapshot(t, fx.wt); after != before {
		t.Errorf("the worktree changed:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

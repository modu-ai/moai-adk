// factory_t1533_test.go — card t1533: the four P1 factory defects reproduced
// during the t1454 card-review rounds (r2–r2f, preserved under
// .moai/reports/t1533/verdicts-lane25/). RED-first: each test failed against
// the pre-repair tree and pins the repaired behavior.
//
//	t1533-1  the bundle loader took over another lane's chained member
//	t1533-2  a hub-waiting picked row held the serial slot (r2 P1-2's sibling)
//	t1533-3  the nominated lease overwrote a generated after hint (t2↔t3 loop)
//	t1533-4  the hub skip waited on one predecessor where the candidate has several
package cli

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/factorylane"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Two hub paths from internal/homestate/hub_files.txt, so the hub chain
// forms without touching the embedded list.
const (
	t1533HubA = "internal/template/catalog.yaml"
	t1533HubB = "internal/config/defaults.go"
)

// TestFactoryBundleRefusesAnotherLanesMember — t1533-1 (r2f finding 1):
// lane-2 re-bundled lane-1's bundle follow-up card and took over its
// assignment — the loader checked only the queue state, and a member waiting
// at `picked` for its chain is exactly that state. A lane never mutates
// another lane's assignment or bundle: the load refuses a member whose
// record already carries another lane's chain, and the refused load changes
// nothing.
func TestFactoryBundleRefusesAnotherLanesMember(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t1", "t2")
	first := fcCard(t, root, "t2")
	if first.BundleID == "" || first.HintAfter != "t1" || first.State != homestate.CardPicked {
		t.Fatalf("fixture: t2 = %s bundle=%q after=%q, want lane-1's picked chained member", first.State, first.BundleID, first.HintAfter)
	}

	// The head reaches the local merge — the moment the follow-up member
	// becomes leasable, and exactly the state the review's takeover happened
	// in (an unmerged member is refused by the T2 guard instead).
	fcSetCardState(t, root, "t1", homestate.CardMergedLocal)

	// lane-2 bundles lane-1's follow-up member as its own head — the
	// assignment takeover the review reproduced.
	sdClearLaneEnv(t)
	_, _, err := runFactory(t, "bundle", "lane-2", "t2", "--run", fcRun)
	if err == nil {
		t.Fatal("lane-2's re-bundle of lane-1's bundle member was accepted — the assignment was taken over")
	}
	after := fcCard(t, root, "t2")
	if after.BundleID != first.BundleID || after.HintAfter != "t1" || after.State != homestate.CardPicked || strings.TrimSpace(after.OwnerLabel) != "" {
		t.Fatalf("t2 = %s owner=%q bundle=%q after=%q, want lane-1's untouched member (picked, chained after t1)", after.State, after.OwnerLabel, after.BundleID, after.HintAfter)
	}
	if c := fcCard(t, root, "t1"); c.State != homestate.CardMergedLocal || c.OwnerLabel != "lane-1" {
		t.Fatalf("t1 = %s owner=%q, want lane-1's untouched head", c.State, c.OwnerLabel)
	}
}

// TestFactoryNextHubChainPickedRowReleasesSerialSlot — t1533-2 (r2f finding
// 2, the sibling of the landed r2 P1-2 repair): a picked row carrying an
// after hint is chain-ordered work — it waits on its predecessor's merge and
// nobody is implementing it — yet its driven row held the serial slot, so
// the serial candidate behind the whole chain answered no card. The slot
// keeps its meaning: the predecessor in flight holds it (parallelizable
// here, so the test isolates the waiting row's contribution), and a driven
// picked row WITHOUT a chain identity still holds it
// (TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling).
func TestFactoryNextHubChainPickedRowReleasesSerialSlot(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-1", LeaseHolder: "lane-1", LeaseExpiresAt: live},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, OwnerLabel: "lane-2", HintAfter: "t1"},
	)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	sdLaneEnv(t, "lane-1", "")
	_, _, err := runFactory(t, "next", "--run", fcRun)
	if err != nil {
		t.Fatalf("lane-1 next: %v — the hub-waiting picked t2 must not hold the serial slot", err)
	}
	c := fcCard(t, root, "t3")
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t3 = %s holder=%q, want leased to lane-1 — t2's hub wait is chain-ordered work, not work in flight", c.State, c.LeaseHolder)
	}
}

// TestReviewNominatedBundlePreservesHint — t1533-3: the nominated lease
// recomputed the hub-chain hint on EVERY record write, and a recomputed tail
// differs from the stored one as soon as a later queue card shares the hub
// path — for a bundle member the recomputed hint overwrote the bundle hint,
// the member's chain order (and the T2 guard that enforces it) followed the
// wrong predecessor, and two members pointing at each other ordered in a
// circle. Hub chains generate after hints at record creation only: an
// explicit --after wins, and a generated hint already on the row survives
// every subsequent write.
func TestReviewNominatedBundlePreservesHint(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3", "t4"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	for _, id := range []string{"t2", "t3", "t4"} {
		fbSeedFiles(t, store, id, t1533HubA)
	}
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t2", "t3")
	if c := fcCard(t, root, "t3"); c.BundleID == "" || c.HintAfter != "t2" {
		t.Fatalf("fixture: t3 = bundle=%q after=%q, want the member chained after t2", c.BundleID, c.HintAfter)
	}

	// t4's record carries the generated hub hint t3 — the queue-later sharer
	// the tail rule named at its creation (placed directly; since the
	// nominated validation decides the wait pre-promotion, the
	// refused-nomination residue no longer mints rows).
	fcPlace(t, root, homestate.Card{CardID: "t4", State: homestate.CardPicked, HintAfter: "t3"})

	// The bundle head reaches the local merge; the member is next in its own
	// chain, and lane-1 nominates it.
	fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun); err != nil {
		t.Fatalf("next --card t3: %v — the member's bundle hint must survive the nominated lease", err)
	}
	c := fcCard(t, root, "t3")
	if c.HintAfter != "t2" {
		t.Fatalf("t3's after = %q, want t2 — a generated hint must never overwrite the stored one", c.HintAfter)
	}
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t3 = %s holder=%q, want leased to lane-1 (the head merged; the member follows it)", c.State, c.LeaseHolder)
	}
	if c := fcCard(t, root, "t4"); c.State != homestate.CardPicked || c.HintAfter != "t3" {
		t.Fatalf("t4 = %s after=%q, want untouched by t3's lease", c.State, c.HintAfter)
	}
}

// TestFactoryBundleHeadCarriesHubCondition — gate r2 finding (a): only the
// second and later bundle members carried an after condition; the head was
// recorded hint-less and was assigned on the spot, so a head whose files
// cross a hub path leased beside another lane's in-flight work on the same
// path with no conflict check at all. The head carries the same generated
// hub condition every record path applies (a stored hint still wins): an
// unmerged sharer refuses the load through the assignment's T2 guard, a
// merged one chains the head.
func TestFactoryBundleHeadCarriesHubCondition(t *testing.T) {
	t.Run("unmerged-sharer-refuses-the-load", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		for _, id := range []string{"t1", "t2"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubA)
		live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live})
		sdRegisterLane(t, root, "lane-1")
		t.Chdir(root)

		sdClearLaneEnv(t)
		_, _, err := runFactory(t, "bundle", "lane-1", "t2", "--run", fcRun)
		if err == nil {
			t.Fatal("the bundle head leased with no hub conflict check while t1 — a sharer of its hub path — is in flight")
		}
		if fcHasCard(t, root, "t2") {
			t.Fatalf("t2 = %s — the refused load left a record row behind", fcCard(t, root, "t2").State)
		}
	})

	t.Run("merged-sharer-chains-the-head", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		for _, id := range []string{"t1", "t2"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubA)
		fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardMergedLocal, OwnerLabel: "lane-2"})
		sdRegisterLane(t, root, "lane-1")
		t.Chdir(root)

		sdClearLaneEnv(t)
		if _, _, err := runFactory(t, "bundle", "lane-1", "t2", "--run", fcRun); err != nil {
			t.Fatalf("bundle with a merged hub sharer: %v", err)
		}
		c := fcCard(t, root, "t2")
		if c.HintAfter != "t1" {
			t.Fatalf("t2's after = %q, want t1 — the head carries the hub condition too", c.HintAfter)
		}
		if c.State != homestate.CardAssigned || c.OwnerLabel != "lane-1" {
			t.Fatalf("t2 = %s owner=%q, want assigned to lane-1", c.State, c.OwnerLabel)
		}
	})
}

// TestFactoryNextMergedPRPredecessorReleasesFollower — gate r2 finding (b):
// the selector's completion set omitted merged-pr, so under github-flow a
// predecessor that reached merged-pr never released its follower — the T2
// guard accepts merged-pr, the selection did not, and the follower answered
// no card forever.
func TestFactoryNextMergedPRPredecessorReleasesFollower(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardMergedPR, OwnerLabel: "lane-1"},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"},
	)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("lane-1 next: %v — a merged-pr predecessor must release its follower", err)
	}
	c := fcCard(t, root, "t2")
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t2 = %s holder=%q, want leased to lane-1 past the merged-pr predecessor", c.State, c.LeaseHolder)
	}
}

// TestReviewPickedHubSkip — t1533-4: the hub skip waited on the ONE
// predecessor the (generated or stored) hint names, but a candidate whose
// files cross several hub paths has a predecessor per hub path. With t2
// merged and t1 — the sharer of the OTHER hub path — still open, the
// candidate leased beside the in-flight t1 and both edited the same hub
// files; the skip must hold until EVERY hub predecessor has merged.
func TestReviewPickedHubSkip(t *testing.T) {
	t.Run("rowless-candidate-waits-behind-every-hub-predecessor", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
		for _, id := range []string{"t1", "t2", "t3"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubB)
		fbSeedFiles(t, store, "t3", t1533HubA, t1533HubB)
		for _, lane := range []string{"lane-1", "lane-2", "lane-3"} {
			sdRegisterLane(t, root, lane)
		}
		t.Chdir(root)

		if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
			t.Fatalf("lane-1's lease = %q, want t1", got)
		}
		if got := fbLeasedCard(t, root, "lane-2"); got != "t2" {
			t.Fatalf("lane-2's lease = %q, want t2", got)
		}
		// t2 merges; t1 — the sharer of the OTHER hub path — is still open.
		fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-3"); got != "" {
			t.Fatalf("lane-3 leased %s while hub predecessor t1 is unmerged — a multi-hub candidate waits behind every predecessor", got)
		}
		// The last predecessor merges; the candidate follows.
		fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-3"); got != "t3" {
			t.Fatalf("lane-3's lease = %q, want t3 once every hub predecessor merged", got)
		}
	})

	t.Run("recorded-candidate-keeps-waiting-on-the-other-hub-predecessor", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
		for _, id := range []string{"t1", "t2", "t3"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubB)
		fbSeedFiles(t, store, "t3", t1533HubA, t1533HubB)
		for _, lane := range []string{"lane-1", "lane-2", "lane-3"} {
			sdRegisterLane(t, root, lane)
		}
		t.Chdir(root)

		if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
			t.Fatalf("lane-1's lease = %q, want t1", got)
		}
		if got := fbLeasedCard(t, root, "lane-2"); got != "t2" {
			t.Fatalf("lane-2's lease = %q, want t2", got)
		}
		// t3's record carries its stored hub hint t2 — the queue-later sharer
		// the generation named as the tail at its creation (placed directly;
		// since the nominated validation decides the wait pre-promotion, the
		// refused-nomination residue no longer mints rows).
		fcPlace(t, root, homestate.Card{CardID: "t3", State: homestate.CardPicked, HintAfter: "t2"})
		// t2 merges; t1 — the sharer of the OTHER hub path — is still open.
		fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-3"); got != "" {
			t.Fatalf("lane-3 leased %s while hub predecessor t1 is unmerged — the recorded candidate waits behind every hub predecessor too", got)
		}
		// The last predecessor merges; the candidate follows.
		fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-3"); got != "t3" {
			t.Fatalf("lane-3's lease = %q, want t3 once every hub predecessor merged", got)
		}
	})
}

// TestFactoryCompleteMergingRetryRefusesForeignLane — gate r2 finding (c):
// the merge-ready entry takes the T14 lease-holder edge, but a card already
// at merging (a crashed delivery's retry shape) entered the mutating path
// with no owner check at all — the gate observed the foreign lane's push,
// pull-request create, and auto-merge request land BEFORE any refusal. A
// merging card is a delivery retry: only the recorded lease holder re-enters
// it, and the refusal precedes the push.
func TestFactoryCompleteMergingRetryRefusesForeignLane(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardMerging})
	d := newGHDouble(t, f)

	sdLaneEnv(t, "lane-2", "")
	if _, err := ghfComplete(t); err == nil {
		t.Fatal("a foreign lane's merging retry was accepted")
	}
	if d.created {
		t.Fatalf("the foreign retry created the pull request before any refusal; calls: %v", d.calls)
	}
	if tip := f.remoteBranchTip(t); tip != "" {
		t.Fatalf("the foreign retry pushed %s before any refusal", tip)
	}

	// The recorded lease holder's own retry still delivers.
	sdLaneEnv(t, ghfLane, "")
	if _, err := ghfComplete(t); err != nil {
		t.Fatalf("the lease holder's own retry: %v", err)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
		t.Fatalf("card = %s, want pr-open after the holder's retry", c.State)
	}
}

// TestFactoryCompleteMergingRetryRefusesExpiredLease — gate r5: the merging
// retry's owner check carried the holder NAME only, so an EXPIRED existing
// holder still ran the push, the pull-request create, and the auto-merge
// request before the record refused it. The retry requires a holder whose
// lease has not expired, refused before the first remote mutation.
func TestFactoryCompleteMergingRetryRefusesExpiredLease(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardMerging})
	d := newGHDouble(t, f)
	expired := fcNow.Add(-time.Minute).Format(time.RFC3339Nano)
	db := fcOpen(t, f.root)
	if _, err := db.DB.Exec(`UPDATE cards SET lease_expires_at = ? WHERE card_id = 't1'`, expired); err != nil {
		t.Fatalf("expire the merging lease: %v", err)
	}
	_ = db.Close()

	sdLaneEnv(t, ghfLane, "")
	if _, err := ghfComplete(t); err == nil {
		t.Fatal("an expired holder's merging retry was accepted")
	}
	if d.created {
		t.Fatalf("the expired holder created the pull request before any refusal; calls: %v", d.calls)
	}
	if tip := f.remoteBranchTip(t); tip != "" {
		t.Fatalf("the expired holder pushed %s before any refusal", tip)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMerging {
		t.Fatalf("card = %s, want still merging (the refusal changed nothing)", c.State)
	}
}

// TestReviewNominatedWaitsOnAllHubPredecessors — gate r5: the nominated
// lease bypassed the multi-hub wait the un-nominated selection applies — a
// candidate whose stored hint's predecessor merged but whose OTHER hub
// path's predecessor is still in flight leased straight through the direct
// path. The nominated validation runs the same wait before it promotes
// anything, refusing with the predecessor error the T2 guard would give.
func TestReviewNominatedWaitsOnAllHubPredecessors(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubB)
	fbSeedFiles(t, store, "t3", t1533HubA, t1533HubB)
	for _, lane := range []string{"lane-1", "lane-2", "lane-3"} {
		sdRegisterLane(t, root, lane)
	}
	t.Chdir(root)

	if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
		t.Fatalf("lane-1's lease = %q, want t1", got)
	}
	if got := fbLeasedCard(t, root, "lane-2"); got != "t2" {
		t.Fatalf("lane-2's lease = %q, want t2", got)
	}
	// t3's record carries its stored hub hint t2 — the queue-later sharer
	// the generation named as the tail at its creation.
	fcPlace(t, root, homestate.Card{CardID: "t3", State: homestate.CardPicked, HintAfter: "t2"})
	// The stored predecessor merges; the OTHER hub path's t1 is in flight.
	fcSetCardState(t, root, "t2", homestate.CardMergedLocal)

	sdLaneEnv(t, "lane-3", "")
	if _, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun); err == nil {
		t.Fatal("the nominated lease succeeded while hub predecessor t1 is unmerged — the direct path waits on every hub predecessor too")
	}
	if c := fcCard(t, root, "t3"); c.State != homestate.CardPicked {
		t.Fatalf("t3 = %s, want still picked (the refusal decided before any promotion)", c.State)
	}
	// The last hub predecessor merges; the direct lease follows.
	fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
	if _, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun); err != nil {
		t.Fatalf("next --card t3 after every hub predecessor merged: %v", err)
	}
	if c := fcCard(t, root, "t3"); c.State != homestate.CardLeased || c.LeaseHolder != "lane-3" {
		t.Fatalf("t3 = %s holder=%q, want leased to lane-3", c.State, c.LeaseHolder)
	}
}

// TestReviewNominatedLeaseDoesNotFillAnEmptyHint — gate r6: the generated
// hint skipped rows that HELD a hint but still filled rows whose hint was
// empty, and the fill minted a cycle in the record: t1's empty row gained
// after=t2 while t2 waits after=t1, the lease failed at the T2 guard, and
// the cycle stayed behind. Generated hints are record-CREATION inputs —
// never written into an existing row, empty or not.
func TestReviewNominatedLeaseDoesNotFillAnEmptyHint(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	// t1's row is created with no hint: no sharer is recorded yet.
	if _, _, err := runFactory(t, "assign", "t1", "--run", fcRun); err != nil {
		t.Fatalf("assign t1: %v", err)
	}
	if c := fcCard(t, root, "t1"); c.HintAfter != "" || c.State != homestate.CardPicked {
		t.Fatalf("fixture: t1 = %s after=%q, want the hint-less picked row", c.State, c.HintAfter)
	}
	// t2 waits on t1 — the row the cycle repro places.
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"})

	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--card", "t1", "--run", fcRun); err != nil {
		t.Fatalf("next --card t1: %v — an empty hint must not be filled on an existing row", err)
	}
	c := fcCard(t, root, "t1")
	if c.HintAfter != "" {
		t.Fatalf("t1's after = %q, want empty — the generated tail (t2) must never be written into the existing row", c.HintAfter)
	}
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t1 = %s holder=%q, want leased to lane-1 (no hint, no guard)", c.State, c.LeaseHolder)
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardPicked || c.HintAfter != "t1" {
		t.Fatalf("t2 = %s after=%q, want untouched", c.State, c.HintAfter)
	}
}

// TestReviewUnrecordedPickedHubWait — gate r6/r7: the b2 arm (queue-picked,
// no record row) recorded the card and attempted the lease even when a hub
// predecessor was unmerged, and the refused claim errored the whole `next`
// call while an unrelated ready card waited behind it. The no-record arms
// wait like every other path: b2 skips before recording, and selection
// reaches the ready card.
func TestReviewUnrecordedPickedHubWait(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStateQueued)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--run", fcRun); err != nil {
		t.Fatalf("lane-1 next: %v — the unrecorded hub-waiting card must be skipped, not errored on", err)
	}
	if fcHasCard(t, root, "t2") {
		t.Fatalf("t2 = %s — the waiting card must not be recorded before its predecessor merges", fcCard(t, root, "t2").State)
	}
	c := fcCard(t, root, "t3")
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t3 = %s holder=%q, want the unrelated ready card leased to lane-1", c.State, c.LeaseHolder)
	}
}

// TestFactoryCompletePRMergePinnedToCardTip — gate r7: the auto-merge
// request was not bound to the verified card commit — a pull request whose
// head moved off the card's tip (a concurrent push) still had auto-merge
// requested and pr-open recorded. The delivery verifies the PR head equals
// the card tip and pins the merge request with --match-head-commit, so a
// concurrent change cannot merge unverified code.
func TestFactoryCompletePRMergePinnedToCardTip(t *testing.T) {
	t.Run("moved-head-refuses-before-the-merge", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		d.headOid = "1111111111111111111111111111111111111111"
		if _, err := ghfComplete(t); err == nil {
			t.Fatal("the delivery merged a pull request whose head moved off the card tip")
		}
		if d.count("pr", "merge") != 0 {
			t.Fatalf("the merge request ran despite the moved head; calls: %v", d.calls)
		}
	})

	t.Run("matching-head-pins-the-merge", func(t *testing.T) {
		f := ghfNew(t, ghfOpts{syncStatus: "complete"})
		d := newGHDouble(t, f)
		if _, err := ghfComplete(t); err != nil {
			t.Fatalf("complete: %v", err)
		}
		want := "pr merge 7 --auto --squash --match-head-commit " + f.tip
		if got := strings.Join(d.last("pr", "merge"), " "); got != want {
			t.Fatalf("merge request = %q, want %q (pinned to the verified card tip)", got, want)
		}
		if c := fcCard(t, f.root, "t1"); c.State != homestate.CardPROpen {
			t.Fatalf("card = %s, want pr-open", c.State)
		}
	})
}

// TestFactoryBundleHeadIgnoresOwnMembers — gate r8: the bundle head's
// hub-candidate set included the bundle's OWN members, so loading t1 with
// recorded t2 (same hub) made the head wait on its own follower and refused
// the load — a dependency opposite to the explicit bundle order. Members
// are ordered by the bundle; only NON-member sharers chain the head.
func TestFactoryBundleHeadIgnoresOwnMembers(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "bundle", "lane-1", "t1", "t2", "--run", fcRun); err != nil {
		t.Fatalf("bundle: %v — the head must not wait on its own member", err)
	}
	head := fcCard(t, root, "t1")
	if head.State != homestate.CardAssigned || head.OwnerLabel != "lane-1" || head.HintAfter != "" {
		t.Fatalf("t1 = %s owner=%q after=%q, want the assigned hint-less head", head.State, head.OwnerLabel, head.HintAfter)
	}
	member := fcCard(t, root, "t2")
	if member.HintAfter != "t1" || member.BundleID == "" || member.BundleID != head.BundleID {
		t.Fatalf("t2 = after=%q bundle=%q, want the member chained after its head %q", member.HintAfter, member.BundleID, head.BundleID)
	}
}

// TestReviewReversedBundleLeasesInBundleOrder — gate r9: a bundle loaded in
// REVERSE queue order (`bundle lane-1 t2 t1`, queue t1→t2, same hub) loads
// fine — the head carries no hub hint — and then the selection deadlocked:
// the head t2 was skipped by the queue-order hub wait on its OWN follower
// t1, while t1 waited on t2 by the bundle rule, so not even the bundle's
// first card leased. Same-bundle members wait by the RECORDED BUNDLE ORDER,
// not the queue order — on every path that consults the wait, the
// un-nominated arms and the nominated validation alike.
func TestReviewReversedBundleLeasesInBundleOrder(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	// Queue order is t1→t2; the bundle loads t2 FIRST.
	fbBundle(t, root, "lane-1", "t2", "t1")
	if head := fcCard(t, root, "t2"); head.State != homestate.CardAssigned || head.OwnerLabel != "lane-1" || head.HintAfter != "" {
		t.Fatalf("fixture: head t2 = %s owner=%q after=%q, want the assigned hint-less head", head.State, head.OwnerLabel, head.HintAfter)
	}
	if member := fcCard(t, root, "t1"); member.HintAfter != "t2" || member.BundleID == "" {
		t.Fatalf("fixture: member t1 = after=%q bundle=%q, want the member chained after t2", member.HintAfter, member.BundleID)
	}

	// The head leads its own bundle: the queue-earlier follower must not
	// hold it back.
	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("lane-1's lease = %q, want t2 — the bundle head is not waited back on its own follower", got)
	}
	// The head merges; the follower follows.
	fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
	if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
		t.Fatalf("lane-1's lease = %q, want t1 (the follower follows its merged head)", got)
	}
}

// TestFactoryCompletePRMergePinnedToCheckTimeTip — gate r10: the merge
// request re-read the card tip at merge time, so a commit landing AFTER the
// readiness check was read as "the verified tip" and auto-merged unverified.
// The tip is captured AT the check; a tip or PR head that moved afterwards
// refuses, and the request pins to the captured SHA.
func TestFactoryCompletePRMergePinnedToCheckTimeTip(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete"})
	d := newGHDouble(t, f)
	// The pull request's head follows the real remote branch tip.
	d.headOidFn = func() string { return f.remoteBranchTip(t) }
	prev := factoryPRReadiness
	factoryPRReadiness = func(out io.Writer, root string, card homestate.Card, lane, cardBranch, target string) (factorylane.MergeCheckRun, error) {
		run, err := prev(out, root, card, lane, cardBranch, target)
		if err == nil {
			// A commit lands after the check, moving the card tip.
			if werr := os.WriteFile(filepath.Join(f.wt, "late.txt"), []byte("late\n"), 0o600); werr != nil {
				t.Fatal(werr)
			}
			fcGit(t, f.wt, "add", "-A")
			fcGit(t, f.wt, "commit", "-q", "-m", "post-check commit")
		}
		return run, err
	}
	t.Cleanup(func() { factoryPRReadiness = prev })

	if _, err := ghfComplete(t); err == nil {
		t.Fatal("the delivery auto-merged a commit that landed after the readiness check")
	}
	// The moved commit must not leave the machine either: no push, no
	// pull request (review-gate r17 — the push is pinned to the verified
	// tip like every other remote mutation).
	if tip := f.remoteBranchTip(t); tip != "" {
		t.Fatalf("the post-check commit was pushed to the remote before any refusal: %s", tip)
	}
	if d.created {
		t.Fatalf("the post-check commit opened a pull request before any refusal; calls: %v", d.calls)
	}
	if d.count("pr", "merge") != 0 {
		t.Fatalf("the merge request ran for the post-check commit; calls: %v", d.calls)
	}
	if c := fcCard(t, f.root, "t1"); c.State == homestate.CardPROpen {
		t.Fatal("pr-open was recorded for a post-check tip")
	}
}

// TestFactoryCompleteRechecksLeaseBeforeRemoteMutations — gate r13: the
// lease-expiry check guarded only the readiness stage, so a lease that
// expired DURING the delivery still ran the push, the pull-request create,
// and the auto-merge request before the state write refused (the gate's
// repro: merge calls=1). Owner, version, and expiry are re-verified
// immediately before every remote mutation.
func TestFactoryCompleteRechecksLeaseBeforeRemoteMutations(t *testing.T) {
	f := ghfNew(t, ghfOpts{syncStatus: "complete", state: homestate.CardMerging})
	d := newGHDouble(t, f)
	prev := factoryPRReadiness
	factoryPRReadiness = func(out io.Writer, root string, card homestate.Card, lane, cardBranch, target string) (factorylane.MergeCheckRun, error) {
		run, err := prev(out, root, card, lane, cardBranch, target)
		if err == nil {
			// The lease expires mid-delivery, after the readiness check.
			expired := fcNow.Add(-time.Minute).Format(time.RFC3339Nano)
			db := fcOpen(t, f.root)
			if _, uerr := db.DB.Exec(`UPDATE cards SET lease_expires_at = ? WHERE card_id = 't1'`, expired); uerr != nil {
				t.Fatalf("expire the lease mid-delivery: %v", uerr)
			}
			_ = db.Close()
		}
		return run, err
	}
	t.Cleanup(func() { factoryPRReadiness = prev })

	if _, err := ghfComplete(t); err == nil {
		t.Fatal("the delivery completed although the lease expired mid-flight")
	}
	if tip := f.remoteBranchTip(t); tip != "" {
		t.Fatalf("the expired delivery pushed %s before any refusal", tip)
	}
	if d.created {
		t.Fatalf("the expired delivery created the pull request before any refusal; calls: %v", d.calls)
	}
	if d.count("pr", "merge") != 0 {
		t.Fatalf("the expired delivery requested auto-merge before any refusal; calls: %v", d.calls)
	}
	if c := fcCard(t, f.root, "t1"); c.State != homestate.CardMerging {
		t.Fatalf("card = %s, want still merging (the refusal changed nothing)", c.State)
	}
}

// TestReviewHubWaitNeverClosesACycle — gate r14/r15: in an INDIRECT after
// chain (t1→t2→t3 waits, t1 and t3 sharing a hub) the queue-order hub wait
// added t3→t1 and closed a cycle into an existing ACYCLIC relation — no
// card ever leased again. A sharer ordered behind the candidate THROUGH THE
// CHAIN its hint opens never holds the candidate either: the wait follows
// the stored relations instead of contradicting them, directly or
// transitively.
func TestReviewHubWaitNeverClosesACycle(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	fbSeedFiles(t, store, "t3", t1533HubA)
	// The explicit acyclic chain: t1 waits on t2 waits on t3.
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardPicked, HintAfter: "t2"},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t3"},
		homestate.Card{CardID: "t3", State: homestate.CardPicked},
	)
	t.Chdir(root)

	queueRec, err := store.Load()
	if err != nil {
		t.Fatalf("load the queue: %v", err)
	}
	db := fcOpen(t, root)
	rows, err := db.ListCards(t.Context(), fcRun)
	if err != nil {
		t.Fatalf("list the records: %v", err)
	}
	_ = db.Close()
	merged := factoryMergedCards(rows)

	if blocker, wait := factoryHubWaitUnmerged(queueRec, rows, merged, "t3"); wait {
		t.Fatalf("wait(t3) = (%q, true), want false — t1 and t2 are ordered behind t3 by the chain; the wait must not close t3 back onto them", blocker)
	}
}

// TestReviewGenerationNeverReversesAnAfterChain — gate r16: the GENERATION
// side of the same rule. With t1.after=t2 recorded and t2 not yet recorded,
// the generated hint named t1 as t2's predecessor — the exact reversal of
// the stored relation — and the cycle went into the record with the row.
// Generation excludes the candidate's after-successors, direct and
// transitive, exactly like the wait does.
func TestReviewGenerationNeverReversesAnAfterChain(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	// t1's EXPLICIT after names t2, which has no record row yet.
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked, HintAfter: "t2"})
	t.Chdir(root)

	queueRec, err := store.Load()
	if err != nil {
		t.Fatalf("load the queue: %v", err)
	}
	db := fcOpen(t, root)
	rows, err := db.ListCards(t.Context(), fcRun)
	if err != nil {
		t.Fatalf("list the records: %v", err)
	}
	_ = db.Close()

	if hf := factoryHubChainFields(queueRec, rows, "t2", nil); hf.HintAfter != nil {
		t.Fatalf("generation for t2 = after %q, want none — t1 is ordered behind t2; naming t1 as t2's predecessor stores the reversal", *hf.HintAfter)
	}
}

// TestReviewNominatedCreationFollowsTheAfterRelation — gate r16, the CLI
// half: the nominated creation of the unrecorded successor leased it
// straight through, while the recorded predecessor waited on ITS merge —
// the pair ordered in a circle at the record. The creation's generated
// hint follows the stored relation (t1 waits on t2, so t2 leads with no
// hint), and the follower comes after the merge.
func TestReviewNominatedCreationFollowsTheAfterRelation(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA)
	fbSeedFiles(t, store, "t2", t1533HubA)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	// t1's EXPLICIT after names t2; t2 has no record row yet.
	if _, _, err := runFactory(t, "assign", "t1", "--after", "t2", "--run", fcRun); err != nil {
		t.Fatalf("assign t1: %v", err)
	}

	sdLaneEnv(t, "lane-1", "")
	if _, _, err := runFactory(t, "next", "--card", "t2", "--run", fcRun); err != nil {
		t.Fatalf("next --card t2: %v — the successor leads its own predecessor", err)
	}
	c := fcCard(t, root, "t2")
	if c.HintAfter != "" {
		t.Fatalf("t2's after = %q, want empty — the generated tail must not reverse the stored relation", c.HintAfter)
	}
	if c.State != homestate.CardLeased || c.LeaseHolder != "lane-1" {
		t.Fatalf("t2 = %s holder=%q, want leased to lane-1", c.State, c.LeaseHolder)
	}
	if c := fcCard(t, root, "t1"); c.HintAfter != "t2" || c.State != homestate.CardPicked {
		t.Fatalf("t1 = %s after=%q, want the untouched follower", c.State, c.HintAfter)
	}
	// The successor merges; the follower follows.
	fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
	if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
		t.Fatalf("lane-1's lease = %q, want t1 (the follower follows its merged predecessor)", got)
	}
}

// TestReviewHubWaitCoversLaterSharers — gate r10: the hub wait scanned only
// queue-EARLIER entries, so a queue-LATER card already working a shared hub
// did not hold the candidate and both lanes edited the same hub files. An
// active (lease-holding) sharer holds the candidate regardless of queue
// direction; a later card still waiting its turn holds nothing.
func TestReviewHubWaitCoversLaterSharers(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	fbSeedFiles(t, store, "t1", t1533HubA, t1533HubB)
	fbSeedFiles(t, store, "t2", t1533HubB)
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	// t1 holds the row its earlier creation left (hint-less); t2 jumped
	// ahead and is in flight on the shared hub.
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardPicked},
		homestate.Card{CardID: "t2", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live},
	)
	t.Chdir(root)

	queueRec, err := store.Load()
	if err != nil {
		t.Fatalf("load the queue: %v", err)
	}
	db := fcOpen(t, root)
	rows, err := db.ListCards(t.Context(), fcRun)
	if err != nil {
		t.Fatalf("list the records: %v", err)
	}
	_ = db.Close()
	merged := factoryMergedCards(rows)

	blocker, wait := factoryHubWaitUnmerged(queueRec, rows, merged, "t1")
	if !wait || blocker != "t2" {
		t.Fatalf("wait(t1) = (%q, %v), want (t2, true) — the later in-flight sharer of hub B holds the candidate", blocker, wait)
	}

	// The control: a later sharer still waiting its turn holds nothing.
	fcSetCardState(t, root, "t2", homestate.CardPicked)
	rows, err = fcOpen(t, root).ListCards(t.Context(), fcRun)
	if err != nil {
		t.Fatalf("re-list the records: %v", err)
	}
	merged = factoryMergedCards(rows)
	if blocker, wait = factoryHubWaitUnmerged(queueRec, rows, merged, "t1"); wait {
		t.Fatalf("wait(t1) = (%q, true), want false — a queue-later unstarted sharer must not invert the queue", blocker)
	}
}

// TestReviewWaitFollowsTheAfterRelation — gate r11/r12: the hub wait judged
// queue positions alone, and against a stored or generated AFTER relation it
// ordered the pair BACKWARDS — t1.after=t2 (explicit or generated) with both
// picked left t1 waiting on t2's merge while t2 was skipped by the
// queue-order wait on t1, and no card ever leased again. A sharer whose own
// hint names the candidate is ordered BEHIND it by that relation: the
// candidate leads, and the wait never flips it.
func TestReviewWaitFollowsTheAfterRelation(t *testing.T) {
	t.Run("explicit-after", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		for _, id := range []string{"t1", "t2"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubA)
		sdRegisterLane(t, root, "lane-1")
		t.Chdir(root)

		// t1's EXPLICIT after names t2; t2 carries no hint.
		if _, _, err := runFactory(t, "assign", "t1", "--after", "t2", "--run", fcRun); err != nil {
			t.Fatalf("assign t1: %v", err)
		}
		fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardPicked})

		sdLaneEnv(t, "lane-1", "")
		if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
			t.Fatalf("lane-1's lease = %q, want t2 — the after relation puts t2 first", got)
		}
		fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
			t.Fatalf("lane-1's lease = %q, want t1 (the follower follows its merged predecessor)", got)
		}
	})

	t.Run("generated-after", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		for _, id := range []string{"t1", "t2"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		fbSeedFiles(t, store, "t1", t1533HubA)
		fbSeedFiles(t, store, "t2", t1533HubA)
		sdRegisterLane(t, root, "lane-1")
		t.Chdir(root)

		// t2 is recorded first (no recorded sharer, no hint); t1's record
		// lands later and the generation chains it after t2.
		if _, _, err := runFactory(t, "assign", "t2", "--run", fcRun); err != nil {
			t.Fatalf("assign t2: %v", err)
		}
		if _, _, err := runFactory(t, "assign", "t1", "--run", fcRun); err != nil {
			t.Fatalf("assign t1: %v", err)
		}
		if c := fcCard(t, root, "t1"); c.HintAfter != "t2" {
			t.Fatalf("fixture: t1 = after=%q, want the generated t2", c.HintAfter)
		}

		sdLaneEnv(t, "lane-1", "")
		if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
			t.Fatalf("lane-1's lease = %q, want t2 — the generated relation puts t2 first", got)
		}
		fcSetCardState(t, root, "t2", homestate.CardMergedLocal)
		if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
			t.Fatalf("lane-1's lease = %q, want t1 (the follower follows its merged predecessor)", got)
		}
	})
}

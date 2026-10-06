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
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
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

	// t4's record lands as the residue of a refused nominated claim: created
	// with the hub hint t3, refused at the T2 guard, left picked and unowned
	// (spec § B.8).
	sdLaneEnv(t, "lane-2", "")
	if _, _, err := runFactory(t, "next", "--card", "t4", "--run", fcRun); err == nil {
		t.Fatal("fixture: t4 leased although its hub predecessor t3 is unmerged")
	}
	if c := fcCard(t, root, "t4"); c.State != homestate.CardPicked || c.HintAfter != "t3" {
		t.Fatalf("fixture: t4 = %s after=%q, want the picked residue chained after t3", c.State, c.HintAfter)
	}

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
		// t3's record lands as the refused nomination's residue (hint t2,
		// unmerged); the queue item is restored, the row stays.
		sdLaneEnv(t, "lane-3", "")
		if _, _, err := runFactory(t, "next", "--card", "t3", "--run", fcRun); err == nil {
			t.Fatal("fixture: t3 leased although its hub predecessor t2 is unmerged")
		}
		if c := fcCard(t, root, "t3"); c.State != homestate.CardPicked || c.HintAfter != "t2" {
			t.Fatalf("fixture: t3 = %s after=%q, want the picked residue chained after t2", c.State, c.HintAfter)
		}
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

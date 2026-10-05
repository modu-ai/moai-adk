// factory_bundle_test.go — SPEC-TODO-CARD-ISSUANCE-001 M4 acceptance tests
// for the bundle chain and the hub hint (REQ-TCI-018/-020, AC-TCI-017 and
// the cli half of AC-TCI-019): the bundle loads serially into its own lane,
// other lanes never receive a member, the serial slot keeps its meaning,
// the order guard refuses an out-of-band assignment, and the keep-set
// verdict reads no file overlap.
package cli

import (
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// fbSeedFiles sets one card's recorded expected files (the `add --files`
// write shape) without walking the CLI.
func fbSeedFiles(t *testing.T, store *factory.BacklogStore, cardID string, files ...string) {
	t.Helper()
	if err := store.Mutate(func(rec *factory.BacklogRecord) error {
		for i := range rec.Items {
			if rec.Items[i].ID != cardID {
				continue
			}
			if rec.Items[i].Issuance == nil {
				rec.Items[i].Issuance = &factory.BacklogIssuance{}
			}
			rec.Items[i].Issuance.Files = files
			return nil
		}
		return nil
	}); err != nil {
		t.Fatalf("seed files on %s: %v", cardID, err)
	}
}

// fbBundle loads the bundle the way the operator does: leader-side, queue
// items picked first.
func fbBundle(t *testing.T, root string, lane string, cards ...string) {
	t.Helper()
	sdClearLaneEnv(t)
	args := append([]string{"bundle", lane}, cards...)
	args = append(args, "--run", fcRun)
	if _, _, err := runFactory(t, args...); err != nil {
		t.Fatalf("bundle %v: %v", cards, err)
	}
}

// fbLeasedCard returns the card id `factory next` leased for the lane, or ""
// when the verb answered no-card. A failing verb fails the test.
func fbLeasedCard(t *testing.T, root, lane string) string {
	t.Helper()
	sdLaneEnv(t, lane, "")
	_, _, err := runFactory(t, "next", "--run", fcRun)
	if err != nil {
		code, ok := ResolveExitCode(err)
		if ok && code == 3 {
			return ""
		}
		t.Fatalf("lane %s next: %v", lane, err)
	}
	// The lease moved one card to leased for this lane; read it from the
	// record rather than parsing the render.
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	rows, err := db.ListCards(t.Context(), fcRun)
	if err != nil {
		t.Fatalf("list cards: %v", err)
	}
	for _, c := range rows {
		if c.State == homestate.CardLeased && strings.TrimSpace(c.LeaseHolder) == lane {
			return c.CardID
		}
	}
	return ""
}

// TestFactoryNextBundleSerialLane — AC-TCI-017 (a)(b)(c): the bundle loads
// with only the first member assigned, lane-1 never receives an unmerged
// member (the skip, not an error), lane-2 never receives a member, and after
// the first member merges, lane-1's next lease prefers the next member over
// an unowned queued card.
func TestFactoryNextBundleSerialLane(t *testing.T) {
	root, store := fcFixture(t)
	// t1 = u1 (queued, not in the bundle); t2..t4 = b1..b3 (queue-picked).
	// Everything runs parallelizable: the serial slot must not decide this
	// test — AC-TCI-017 (d) has its own test.
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3", "t4"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t2", "t3", "t4")

	b1 := fcCard(t, root, "t2")
	if b1.State != homestate.CardAssigned || b1.OwnerLabel != "lane-1" || b1.BundleID == "" || b1.BundleOrder != 0 {
		t.Fatalf("b1 = %s owner=%q bundle=%q order=%d, want lane-1's assigned first member", b1.State, b1.OwnerLabel, b1.BundleID, b1.BundleOrder)
	}
	b2 := fcCard(t, root, "t3")
	if b2.State != homestate.CardPicked || b2.BundleOrder != 1 || b2.HintAfter != "t2" || strings.TrimSpace(b2.OwnerLabel) != "" {
		t.Fatalf("b2 = %s after=%q order=%d owner=%q, want an unowned picked second member after b1", b2.State, b2.HintAfter, b2.BundleOrder, b2.OwnerLabel)
	}
	b3 := fcCard(t, root, "t4")
	if b3.BundleOrder != 2 || b3.HintAfter != "t3" {
		t.Fatalf("b3 = %s after=%q order=%d, want the third member after b2", b3.State, b3.HintAfter, b3.BundleOrder)
	}
	if b1.BundleID != b2.BundleID || b2.BundleID != b3.BundleID {
		t.Fatalf("bundle ids diverge: %q %q %q", b1.BundleID, b2.BundleID, b3.BundleID)
	}

	// (a) lane-1 leases b1 through the assignment edge.
	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("lane-1's first lease = %q, want t2 (b1)", got)
	}
	// Before b1 merges, lane-1's next lease never returns b2 — the skip, not
	// an error. u1 (queued) is the only other candidate, and taking it is
	// permitted; the assertion is that b2 stays put.
	if got := fbLeasedCard(t, root, "lane-1"); got == "t3" || got == "t4" {
		t.Fatalf("lane-1 received bundle member %q before its predecessor merged", got)
	}
	// (b) lane-2 never receives a member.
	if got := fbLeasedCard(t, root, "lane-2"); got == "t3" || got == "t4" {
		t.Fatalf("lane-2 received bundle member %q", got)
	}

	// (c) fresh scenario: after b1 reaches the local merge, lane-1's next
	// lease prefers b2 over u1 — in a fixture where u1 is still available.
	// The lane env the earlier leases stamped is cleared first: the queue
	// seeds below run leader-side.
	root2, store2 := fcFixture(t)
	sdClearLaneEnv(t)
	fcQueue(t, store2, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2", "t3"} {
		fcClassify(t, store2, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	sdRegisterLane(t, root2, "lane-1")
	t.Chdir(root2)
	fbBundle(t, root2, "lane-1", "t2", "t3")
	if got := fbLeasedCard(t, root2, "lane-1"); got != "t2" {
		t.Fatalf("scenario C: lane-1's first lease = %q, want t2", got)
	}
	fcSetCardState(t, root2, "t2", homestate.CardMergedLocal)
	if got := fbLeasedCard(t, root2, "lane-1"); got != "t3" {
		t.Fatalf("scenario C: lane-1's next lease = %q, want t3 (b2) preferred over u1", got)
	}
}

// TestFactoryBundleKeepsSerialSlot — AC-TCI-017 (d): the fleet-wide serial
// slot keeps its present meaning beside a bundle: the second serial card is
// skipped while the first is in flight, no bundle member leaks to the other
// lane, and the non-bundle selection shape matches the pre-change golden.
func TestFactoryBundleKeepsSerialSlot(t *testing.T) {
	root, store := fcFixture(t)
	// t1/t2 serial queued; t3/t4 the bundle members (queue-picked).
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t3", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t4", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t3", "t4")

	if got := fbLeasedCard(t, root, "lane-1"); got != "t3" {
		t.Fatalf("lane-1's first lease = %q, want t3 (b1)", got)
	}
	if got := fbLeasedCard(t, root, "lane-2"); got != "t1" {
		t.Fatalf("lane-2's lease = %q, want t1 (the serial card, slot free)", got)
	}
	// The serial slot is held by t1: lane-2 skips t2; t4 belongs to lane-1's
	// bundle. The only-serial-candidates shape ends on the no-card exit.
	if got := fbLeasedCard(t, root, "lane-2"); got != "" {
		t.Fatalf("lane-2's next lease = %q, want no card (serial slot held, t4 is lane-1's)", got)
	}
	// t2 never left the queue — it has no factory record at all, and its
	// queue state is still queued.
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load queue: %v", err)
	}
	if rec.Items[1].State != factory.BacklogStateQueued {
		t.Fatalf("t2 queue state = %s, want still queued (the serial slot held it back)", rec.Items[1].State)
	}
}

// TestFactoryAssignBundleOrderGuard — AC-TCI-017: the order guard stays an
// ERROR on the direct path: assigning a later member while its predecessor
// has not reached the local merge is refused and changes no state.
func TestFactoryAssignBundleOrderGuard(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t1", "t2")

	sdLaneEnv(t, "lane-2", "")
	if _, _, err := runFactory(t, "assign", "t2", "--to", "lane-2", "--run", fcRun); err == nil {
		t.Fatal("assigning the second member while b1 is unmerged was accepted")
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardPicked || strings.TrimSpace(c.OwnerLabel) != "" {
		t.Fatalf("t2 = %s owner=%q, want an unowned picked row (the refusal changed nothing)", c.State, c.OwnerLabel)
	}
}

// TestFactoryAssignBundleHubChain — AC-TCI-019 (a)(c), the cli half: a card
// whose recorded files cross the embedded hub list is recorded with its
// after hint naming the earlier open card that shares the hub path; two
// cards sharing only a non-hub path chain not at all.
func TestFactoryAssignBundleHubChain(t *testing.T) {
	root, store := fcFixture(t)
	// t1/t2 share the hub path catalog.yaml; t3/t4 share a non-hub path. All
	// four are queue-picked: the assign path's precondition (REQ-FR-022).
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fbSeedFiles(t, store, "t1", "internal/template/catalog.yaml")
	fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml")
	fbSeedFiles(t, store, "t3", "some/other.go")
	fbSeedFiles(t, store, "t4", "some/other.go")
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", fcRun); err != nil {
		t.Fatalf("assign t1: %v", err)
	}
	// The record lands WITH the hub hint; the assign's own assignment
	// transition is what the T2 guard refuses while t1 is unmerged — that
	// refusal is the (a) evidence on the direct path.
	if _, _, err := runFactory(t, "assign", "t2", "--to", "lane-2", "--run", fcRun); err == nil {
		t.Fatal("assigning the hub-chained card was accepted before its predecessor merged")
	}
	if c := fcCard(t, root, "t2"); c.HintAfter != "t1" {
		t.Fatalf("t2's after = %q, want t1 (the hub chain)", c.HintAfter)
	}
	if c := fcCard(t, root, "t2"); strings.TrimSpace(c.OwnerLabel) != "" {
		t.Fatalf("t2 owner = %q, want unowned (the assignment was refused)", c.OwnerLabel)
	}
	if c := fcCard(t, root, "t1"); c.HintAfter != "" {
		t.Fatalf("t1's after = %q, want empty (no earlier card shares the hub path)", c.HintAfter)
	}

	// (c) the non-hub pair chains not at all.
	if _, _, err := runFactory(t, "assign", "t3", "--to", "lane-1", "--run", fcRun); err != nil {
		t.Fatalf("assign t3: %v", err)
	}
	if _, _, err := runFactory(t, "assign", "t4", "--to", "lane-2", "--run", fcRun); err != nil {
		t.Fatalf("assign t4: %v", err)
	}
	if c := fcCard(t, root, "t4"); c.HintAfter != "" {
		t.Fatalf("t4's after = %q, want empty — a non-hub overlap is no chain", c.HintAfter)
	}

	// (a) the chained card is not leased before its predecessor merges: the
	// selection skips it, and the direct path still refuses.
	sdLaneEnv(t, "lane-2", "")
	if got := fbLeasedCard(t, root, "lane-2"); got == "t2" {
		t.Fatal("lane-2 leased the hub-chained card before its predecessor merged")
	}
	sdLaneEnv(t, "lane-2", "")
	if _, _, err := runFactory(t, "next", "--card", "t2", "--run", fcRun); err == nil {
		t.Fatal("the nominated lease of the hub-chained card was accepted before its predecessor merged")
	}
}

// TestFactorySerialBundleHeadLeasesDespitePickedMembers — card t1454
// card-review r2 P1-2: a SERIAL bundle's unowned pending members sit at
// `picked`, and their rows held the serial slot against the head itself —
// the very first `factory next` of the bundle lane answered no card. A
// picked row carrying a bundle identity is chain-ordered work, not
// independently picked work: the in-flight exclusion skips it, while a
// standalone picked row keeps holding the slot exactly as the t1407
// operator-ruling tests pin (factory_serial_slot_stale_test.go).
func TestFactorySerialBundleHeadLeasesDespitePickedMembers(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeSerial)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	fbBundle(t, root, "lane-1", "t1", "t2")

	if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
		t.Fatalf("lane-1's first lease = %q, want t1 (the bundle head) — the picked pending member held the serial slot", got)
	}
}

// TestFactoryBundleLoadAtomicWhenAMemberIsLeased — card t1454 card-review
// r2 finding 3: the bundle load is ONE act. A member whose record has
// already moved past `picked` aborts the whole load, and the members
// recorded before it leave no residue — the per-member calls committed the
// earlier records and stranded them.
func TestFactoryBundleLoadAtomicWhenAMemberIsLeased(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcPlace(t, root, homestate.Card{CardID: "t2", State: homestate.CardLeased})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "bundle", "lane-1", "t1", "t2", "--run", fcRun); err == nil {
		t.Fatal("the bundle load accepted a member whose record is already leased")
	}
	if fcHasCard(t, root, "t1") {
		t.Fatal("t1's record survived the failed load — the bundle load is not atomic")
	}
	if c := fcCard(t, root, "t2"); c.State != homestate.CardLeased {
		t.Fatalf("t2 = %s, want still leased (the refused load changed nothing)", c.State)
	}
}

// TestFactoryNextSkipsHubCandidateWhosePredecessorIsUnmerged — card t1454
// card-review r2 finding 4: a rowless queued candidate's hub hint is
// computed at promotion, so its predecessor condition is checked BEFORE the
// promotion. A candidate whose hint names an unmerged recorded predecessor
// was promoted and then failed the claim, erroring the whole `next` verb.
func TestFactoryNextSkipsHubCandidateWhosePredecessorIsUnmerged(t *testing.T) {
	root, store := fcFixture(t)
	// t1 is the chainable predecessor: recorded, open, queue-held (so no arm
	// can take it), and merged nowhere. t2 is the rowless queued candidate.
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStateQueued)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fbSeedFiles(t, store, "t1", "internal/template/catalog.yaml")
	fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml")
	fcPlace(t, root, homestate.Card{CardID: "t1", State: homestate.CardPicked})
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	if got := fbLeasedCard(t, root, "lane-1"); got != "" {
		t.Fatalf("lane-1 leased %q — the hub-chained candidate was promoted past its unmerged predecessor", got)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if rec.Items[1].State != factory.BacklogStateQueued {
		t.Fatalf("t2 queue state = %s, want still queued (the skip, not a failed claim)", rec.Items[1].State)
	}
	if fcHasCard(t, root, "t2") {
		t.Fatal("t2's record was created by the failed promotion — the candidate should have been skipped")
	}
}

// TestFactoryNextAfterGuardCountsOtherRunsMerges — card t1454 card-review
// r2 finding 5: the after guard's merged set reads every run, like
// predecessorMerged does. A predecessor merged under a previous run id
// gated its successor forever when only the current run's rows were read.
func TestFactoryNextAfterGuardCountsOtherRunsMerges(t *testing.T) {
	root, store := fcFixture(t)
	// t1 is queue-held with no row in the current run: its merge lives only
	// in run-prev, and no arm may take or re-record t1 itself.
	fcQueue(t, store, factory.BacklogStateHold, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcPlace(t, root,
		homestate.Card{RunID: "run-prev", CardID: "t1", State: homestate.CardMergedLocal},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"},
	)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)

	if got := fbLeasedCard(t, root, "lane-1"); got != "t2" {
		t.Fatalf("lane-1's lease = %q, want t2 — a predecessor merged under run-prev still gated the after guard", got)
	}
}

// TestFactoryHubChainRequiresSharedHubPath — card t1454 card-review r2
// finding 6: the hub chain orders cards that share A HUB PATH with the
// candidate. Two cards crossing DIFFERENT hub paths chain not at all.
func TestFactoryHubChainRequiresSharedHubPath(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	fbSeedFiles(t, store, "t1", "internal/config/defaults.go")
	fbSeedFiles(t, store, "t2", "internal/template/catalog.yaml")
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	sdClearLaneEnv(t)
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", fcRun); err != nil {
		t.Fatalf("assign t1: %v", err)
	}
	sdLaneEnv(t, "lane-2", "")
	_, _, _ = runFactory(t, "assign", "t2", "--to", "lane-2", "--run", fcRun)
	if c := fcCard(t, root, "t2"); c.HintAfter != "" {
		t.Fatalf("t2's after = %q, want empty — the cards cross different hub paths, so no chain", c.HintAfter)
	}
}

// TestFactoryKeepSetReadsNoFileOverlap — AC-TCI-019 (b): the keep-set
// verdict reads no file overlap — two cards whose recorded files share every
// path lease independently, and no refusal names a file.
func TestFactoryKeepSetReadsNoFileOverlap(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued)
	// Parallelizable classification: the serial slot must not be what
	// decides this test — the assertion is that no FILE OVERLAP refusal
	// exists anywhere in the keep-set path.
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fcClassify(t, store, "t2", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	fbSeedFiles(t, store, "t1", "shared/path.go")
	fbSeedFiles(t, store, "t2", "shared/path.go")
	sdRegisterLane(t, root, "lane-1")
	sdRegisterLane(t, root, "lane-2")
	t.Chdir(root)

	if got := fbLeasedCard(t, root, "lane-1"); got != "t1" {
		t.Fatalf("lane-1's lease = %q, want t1", got)
	}
	if got := fbLeasedCard(t, root, "lane-2"); got != "t2" {
		t.Fatalf("lane-2's lease = %q, want t2 — the keep-set verdict reads no file overlap", got)
	}
}

package mergeq

// issue-1948, follow-up of the final code audit of a36ea8c368 (the window left open by N66). Merger.LastCheck reads the
// ledger once with the landing lock held and the row in landing; a failing run recorded after that read and before the
// compare-and-swap of the trunk ref was landed all the same, because evidence writes take no landing lock. The ledger
// closes it from the other side: the evidence write refuses a failing green, gate or review run about the candidate of a
// row that is in landing. These tests hook the instant itself (LastCheck, which runs with the row in landing) with the
// real ledger and real git, and assert both outcomes of the race: a failure recorded just before the row is landing is
// seen by the check and stops the landing; one recorded while it is landing is refused, and the landing that stands is
// consistent (trunk moved, row landed) with the retry after it recorded and holding the card.

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
	"github.com/modu-ai/moai-adk/internal/agenticfactory/ledger"
)

// cliLastCheck is the check `moai af land` installs (internal/cli/af_land.go): the ledger's own read of the newest runs.
func cliLastCheck(s *ledger.Store) func(context.Context, string, string) error {
	return func(ctx context.Context, cardID, sha string) error {
		kind, err := s.FailingNewestRun(ctx, cardID, sha)
		if err != nil {
			return err
		}
		if kind != "" {
			return fmt.Errorf("the newest %s run of commit %s failed", kind, sha)
		}
		return nil
	}
}

func cardEpochOf(t *testing.T, s *ledger.Store, id string) int64 {
	t.Helper()
	cards, err := s.Cards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cards {
		if c.ID == id {
			return c.Epoch
		}
	}
	t.Fatalf("no card %s", id)
	return 0
}

func failures(t *testing.T, s *ledger.Store, id string) int {
	t.Helper()
	ev, err := s.EvidenceFor(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range ev {
		if e.Result == ledger.ResultFail {
			n++
		}
	}
	return n
}

// The race lost by the failing run: it arrives while the row is landing, and is refused. It is hooked at both ends of the
// check: just before the check reads the ledger (the old code let the check see it and stop the landing; the new code
// refuses it, so the landing stands) and just after the check read, before the compare-and-swap of the trunk ref (the
// window itself: the old code recorded it and landed the commit all the same). Either way the landing that stands is
// consistent (trunk moved, row landed, no failing run on record), and the retry after the landing is recorded and then
// holds the card.
func TestLandingWindow_AFailureRecordedWhileTheRowIsLandingIsRefusedAndTheLandingStandsConsistent(t *testing.T) {
	for _, afterRead := range []bool{false, true} {
		name := "just before the check reads the ledger"
		if afterRead {
			name = "after the check read it, before the trunk ref moves"
		}
		t.Run(name, func(t *testing.T) { landingWindowRace(t, afterRead) })
	}
}

func landingWindowRace(t *testing.T, afterRead bool) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	e := cardEpochOf(t, s, "t1")
	trunk := git(t, dir, "rev-parse", "refs/heads/trunk")
	m := realMerger(t, dir, s)
	check := cliLastCheck(s)
	var refusal error
	record := func(ctx context.Context) {
		if got := rowState(t, s, "t1"); got != ledger.QueueLanding {
			t.Errorf("the hook runs with the row %s, want landing", got)
		}
		refusal = s.AddEvidence(ctx, "t1", e, ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "tree-" + cand, SubjectSHA: cand, Result: ledger.ResultFail})
	}
	m.LastCheck = func(ctx context.Context, cardID, sha string) error {
		if !afterRead {
			record(ctx)
		}
		err := check(ctx, cardID, sha)
		if afterRead && err == nil {
			record(ctx)
		}
		return err
	}
	res, err := m.Land(ctx, "t1", cand)
	if err != nil || res.Outcome != OutcomeLanded {
		t.Fatalf("Land = %+v, %v; want landed", res, err)
	}
	if refusal == nil || !strings.Contains(refusal.Error(), "being landed") {
		t.Fatalf("the failing run recorded while the row was landing: %v, want the refusal", refusal)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got == trunk {
		t.Error("the trunk did not move")
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueLanded {
		t.Errorf("the row is %s, want landed", got)
	}
	if n := failures(t, s, "t1"); n != 0 {
		t.Errorf("%d failing runs on record after a landing that refused one: the trunk holds a commit the ledger has failing", n)
	}

	// the retry after the landing is recorded, and the rule after landing holds the card
	if err := s.AddEvidence(ctx, "t1", e, ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "tree-" + cand, SubjectSHA: cand, Result: ledger.ResultFail}); err != nil {
		t.Fatalf("the retry after the landing is refused: %v", err)
	}
	err = s.Move(ctx, "t1", e, core.Position{Stage: core.StageMerging, RemoteIntegrated: true}, "mergeq", "merge seen")
	if err == nil {
		err = s.Move(ctx, "t1", e, core.Position{Stage: core.StageDone}, "mergeq", "landed")
	}
	if !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("finishing the card after a failure recorded after the landing: %v, want the post-landing rule (ErrGuardMissing)", err)
	}
}

// The race won by the failing run: it commits just before the row is landing, and the check sees it.
func TestLandingWindow_AFailureRecordedBeforeTheRowIsLandingStopsTheLanding(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	e := cardEpochOf(t, s, "t1")
	trunk := git(t, dir, "rev-parse", "refs/heads/trunk")
	m := realMerger(t, dir, s)
	m.LastCheck = cliLastCheck(s)
	// recorded while the gates run: the row is verifying, which is not refused
	m.Verify = func(ctx context.Context, _ string) error {
		if got := rowState(t, s, "t1"); got != ledger.QueueVerifying {
			t.Errorf("the gates run with the row %s, want verifying", got)
		}
		return s.AddEvidence(ctx, "t1", e, ledger.Evidence{Kind: ledger.KindReview, TreeSHA: "tree-" + cand, SubjectSHA: cand, Result: ledger.ResultFail})
	}
	res, err := m.Land(ctx, "t1", cand)
	if err == nil || !strings.Contains(err.Error(), "review") {
		t.Fatalf("Land = %+v, %v; want the refusal of the check, naming the review run", res, err)
	}
	if got := git(t, dir, "rev-parse", "refs/heads/trunk"); got != trunk {
		t.Errorf("the trunk moved from %s to %s", trunk, got)
	}
	if got := rowState(t, s, "t1"); got != ledger.QueueWaiting {
		t.Errorf("the row is %s, want waiting", got)
	}
	if n := failures(t, s, "t1"); n != 1 {
		t.Errorf("%d failing runs on record, want the one recorded while the gates ran", n)
	}
}

// A row left in landing by a lander that died keeps refusing, and a reconcile frees it (here: the candidate is not in
// the trunk, so the row goes back to waiting and the failing run can be recorded and stops the next landing).
func TestLandingWindow_ReconcileFreesARowALanderThatDiedLeftInLanding(t *testing.T) {
	ctx := context.Background()
	dir := repo(t)
	cand := candidate(t, dir, "c1", "a.txt", "a\n")
	s := queuedCard(t, "t1", cand)
	e := leaveLanding(t, s, cand)
	fail := ledger.Evidence{Kind: ledger.KindGate, TreeSHA: "tree-" + cand, SubjectSHA: cand, Result: ledger.ResultFail}
	if err := s.AddEvidence(ctx, "t1", e, fail); err == nil || !strings.Contains(err.Error(), "being landed") {
		t.Fatalf("the failing run was recorded against a row a dead lander left in landing: %v", err)
	}
	m := realMerger(t, dir, s)
	acts, err := m.Reconcile(ctx, itemsOf(t, s))
	if err != nil {
		t.Fatalf("Reconcile: %+v %v", acts, err)
	}
	if got := rowState(t, s, "t1"); got == ledger.QueueLanding {
		t.Fatalf("the row is still landing after the reconcile: %+v", acts)
	}
	if err := s.AddEvidence(ctx, "t1", e, fail); err != nil {
		t.Fatalf("still refused after the reconcile settled the row (%s): %v", rowState(t, s, "t1"), err)
	}
}

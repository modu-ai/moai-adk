package cli

// Card t2104 (GitHub #1938): a card left behind in a dead run with an EXPIRED
// lease has a formal path into a live run — `moai factory runs --retire <dead>
// --rebind-to <live>` — and `assign --reassign` is deliberately scoped to the
// card's current dispatch run. These tests pin the route so it cannot regress
// into a gap: (a) the expired leased row moves intact, (b) a card that already
// has a live-run row (assign-first, or a second dead run whose lease was
// renewed at another time) is abandoned with a pointer instead of refused,
// because the source's EXPIRED lease is not live work, so the card keeps one
// live row and every dead run can be retired, (c) a divergence on any column
// other than the dead lease still refuses. Fixtures are isolated sandboxes.

import (
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

func expiredLeaseRebindFixture(t *testing.T, assignedInLive bool) string {
	t.Helper()
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "run-src", deadPID, deadStart)
	seedRun(t, root, "run-tgt", os.Getpid(), homestate.CurrentProcessFingerprint())
	fcPlace(t, root, homestate.Card{
		RunID: "run-src", CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-9",
		LeaseHolder: "lane-9", LeaseExpiresAt: "2026-09-20T01:00:00Z", Version: 3,
	})
	if assignedInLive {
		fcPlace(t, root, homestate.Card{RunID: "run-tgt", CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 1})
	}
	return root
}

func cardStateIn(t *testing.T, root, runID string) string {
	t.Helper()
	return lcQuery(t, root, `SELECT COALESCE(group_concat(state), '') FROM cards WHERE run_id=? AND card_id='t1'`, runID)
}

func TestFactoryRunsRebindMovesExpiredLeasedRow(t *testing.T) {
	root := expiredLeaseRebindFixture(t, false)
	out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
	if err != nil {
		t.Fatalf("rebind of an expired leased row: %v\n%s", err, out)
	}
	if !strings.Contains(out, "moved cards: t1") {
		t.Errorf("output does not report t1 as moved:\n%s", out)
	}
	if got := lcQuery(t, root, `SELECT group_concat(run_id||':'||state||':'||owner_label, ' ; ') FROM cards WHERE card_id='t1'`); got != "run-tgt:leased:lane-9" {
		t.Errorf("card rows = %q, want the single row run-tgt:leased:lane-9", got)
	}
	if got := runStatusOf(t, root, "run-src"); got != "retired" {
		t.Errorf("source status = %q, want retired", got)
	}
}

// Assign-first: the live run already holds an assigned t1 row. The source's
// expired lease is the only divergence, so the source row is abandoned with a
// pointer; the live row stays the card's one live row and the dead run retires.
func TestFactoryRunsRebindAbandonsExpiredRowDuplicatingLiveRow(t *testing.T) {
	root := expiredLeaseRebindFixture(t, true)
	out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
	if err != nil {
		t.Fatalf("an expired lease alone must not block the consolidation: %v\n%s", err, out)
	}
	if !strings.Contains(out, "abandoned cards: t1") {
		t.Errorf("output does not report t1 as abandoned:\n%s", out)
	}
	if got := cardStateIn(t, root, "run-tgt"); got != "assigned" {
		t.Errorf("live row state = %q, want it untouched (assigned)", got)
	}
	if got := cardStateIn(t, root, "run-src"); got != "abandoned" {
		t.Errorf("source row state = %q, want abandoned", got)
	}
	if got := runStatusOf(t, root, "run-src"); got != "retired" {
		t.Errorf("source status = %q, want retired", got)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM events WHERE kind='card.rebound' AND run_id='run-src' AND payload_json LIKE '%run-tgt%'`); n != 1 {
		t.Errorf("card.rebound pointer events = %d, want 1", n)
	}
}

// Two dead runs each hold t1 with an expired lease renewed at different times
// (the shape GitHub #1938 reports). The first rebind moves t1; the second used
// to be refused on lease_expires_at and left its run active and unretirable.
func TestFactoryRunsRebindSecondDeadRunWithDivergingExpiredLease(t *testing.T) {
	root := runOwnerSandbox(t)
	t.Chdir(root)
	deadPID, deadStart := deadOwnerIdentity(t)
	seedRun(t, root, "run-a", deadPID, deadStart)
	seedRun(t, root, "run-b", deadPID, deadStart)
	seedRun(t, root, "run-tgt", os.Getpid(), homestate.CurrentProcessFingerprint())
	fcPlace(t, root,
		homestate.Card{RunID: "run-a", CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: "2026-09-20T01:00:00Z", Version: 3},
		homestate.Card{RunID: "run-b", CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-9", LeaseHolder: "lane-9", LeaseExpiresAt: "2026-09-20T02:00:00Z", Version: 4})
	for _, dead := range []string{"run-a", "run-b"} {
		if out, err := runFactoryCommand(t, "runs", "--retire", dead, "--rebind-to", "run-tgt"); err != nil {
			t.Fatalf("rebind of %s: %v\n%s", dead, err, out)
		}
		if got := runStatusOf(t, root, dead); got != "retired" {
			t.Errorf("%s status = %q, want retired", dead, got)
		}
	}
	if got := cardStateIn(t, root, "run-tgt"); got != "leased" {
		t.Errorf("live row state = %q, want leased (moved from run-a)", got)
	}
	if got := cardStateIn(t, root, "run-b"); got != "abandoned" {
		t.Errorf("second dead run's row = %q, want abandoned", got)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM runs WHERE status='active' AND run_id<>'run-tgt'`); n != 0 {
		t.Errorf("%d dead runs remain active, want 0", n)
	}
}

// The relaxation is limited to the dead lease: a source row that also differs
// from the live row on any other compared column (here the stage) still refuses
// and changes nothing.
func TestFactoryRunsRebindStillRefusesNonLeaseDivergence(t *testing.T) {
	root := expiredLeaseRebindFixture(t, true)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`UPDATE cards SET stage='run' WHERE run_id='run-src' AND card_id='t1'`); err != nil {
		t.Fatal(err)
	}
	_ = db.Close()
	before := lcDump(t, root)
	out, err := runFactoryCommand(t, "runs", "--retire", "run-src", "--rebind-to", "run-tgt")
	if err == nil {
		t.Fatalf("consolidation accepted a row diverging on stage:\n%s", out)
	}
	for _, want := range []string{"t1", "run-tgt", "stage"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("refusal %q does not name %q", err.Error(), want)
		}
	}
	if after := lcDump(t, root); after != before {
		t.Errorf("the refusal changed state")
	}
}

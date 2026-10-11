package cli

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// Ordering of the factory binding against the queue's current-dispatch record
// (turn-end gate relay #4, card t1538).

// A mirror that arrives after a LATER dispatch completed is stale: writing the
// factory half now would drag the binding — and the current-dispatch record —
// back to the older run, and the older run's approval would verify again. The
// mirror checks the record inside the queue lock and writes nothing when the
// card's current dispatch is another run.
func TestDispatchMirrorSkipsASupersededRun(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	// Two dispatches of one card; the second completes first.
	if err := factory.RecordFactoryCardAssignment(root, "run-old", "t1", "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	if err := factory.RecordFactoryCardAssignment(root, "run-new", "t1", "lane-2", ""); err != nil {
		t.Fatal(err)
	}
	if err := writeFactoryAssignment(ctx, root, store, "run-new", "t1", "lane-2"); err != nil {
		t.Fatalf("the current dispatch's mirror: %v", err)
	}
	// The older run's mirror arrives late.
	if err := writeFactoryAssignment(ctx, root, store, "run-old", "t1", "lane-1"); err != nil {
		t.Fatalf("the superseded mirror must be a no-op, not an error: %v", err)
	}
	fcWantCurrent(t, store, "t1", "run-new", "lane-2")
	db := fcOpen(t, root)
	row, linked, err := db.RecordedCardRowReadonly(ctx, "t1")
	if err != nil || !linked || row.RunID != "run-new" {
		t.Fatalf("binding after the stale mirror: run=%q linked=%v err=%v, want run-new", row.RunID, linked, err)
	}
	if _, err := db.LoadCard(ctx, "run-old", "t1"); !errors.Is(err, homestate.ErrCardNotFound) {
		t.Fatalf("the superseded mirror created a factory row for run-old (err %v)", err)
	}
}

// abortCurrentDispatchRefresh makes every refresh of the queue's
// current-dispatch record fail, the way a full disk would.
func abortCurrentDispatchRefresh(t *testing.T, store *factory.BacklogStore) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(store.EnginePath()))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`CREATE TRIGGER abort_current_refresh BEFORE UPDATE ON todo_dispatch_current BEGIN SELECT RAISE(ABORT,'fixture_current_abort'); END`); err != nil {
		t.Fatal(err)
	}
}

// The queue's record names the run BEFORE a lease moves the binding onto it:
// a record that cannot be written fails the claim closed, so no committed
// lease can leave the binding ahead of the record (where a retried older
// dispatch would read the stale record as current and drag the binding back).
func TestFactoryNextClaimRecordsTheRunBeforeTheLease(t *testing.T) {
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	sdRegisterLane(t, root, "lane-1")
	if _, _, err := runFactory(t, "assign", "t1", "--to", "lane-1", "--run", "run-old"); err != nil {
		t.Fatalf("assign run-old: %v", err)
	}
	seedOlderDispatch(t, root, store, "t1")
	fcPlace(t, root, homestate.Card{CardID: "t1", RunID: "run-new", State: homestate.CardAssigned, OwnerLabel: "lane-1", Version: 2})
	abortCurrentDispatchRefresh(t, store)

	ctx := context.Background()
	db := fcOpen(t, root)
	cur, err := db.LoadCard(ctx, "run-new", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok, _, err := factoryNextClaim(ctx, db, root, "run-new", cur, "lane-1"); err == nil || ok {
		t.Fatalf("claim with an unwritable current-dispatch record: ok=%v err=%v, want a failure", ok, err)
	}
	after, err := db.LoadCard(ctx, "run-new", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if after.State != homestate.CardAssigned || after.LeaseHolder != "" {
		t.Fatalf("row after the failed claim = state %s holder %q — the lease committed although the record could not be written", after.State, after.LeaseHolder)
	}
	row, linked, err := db.RecordedCardRowReadonly(ctx, "t1")
	if err != nil || !linked || row.RunID != "run-old" {
		t.Fatalf("binding after the failed claim: run=%q linked=%v err=%v, want it unmoved at run-old", row.RunID, linked, err)
	}
}

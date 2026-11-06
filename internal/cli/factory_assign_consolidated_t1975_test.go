package cli

// factory_assign_consolidated_t1975_test.go — card t1975, two codex findings on
// the t1932 tree about `factory assign` after a run consolidation (RebindRun):
//
//   P1: `factory assign <card> --run <source>` against the retired source run
//   must change neither store (the queue's current-dispatch record and the
//   factory DB). On this tree the run is refused at run resolution, before any
//   write, so the queue is never touched.
//
//   P2: after a consolidation the card lives in the survivor; the queue's
//   current-dispatch record must not leave `factory assign --reassign --run
//   <survivor>` unreachable (the stale-run guard compares the record's run with
//   --run, and the retired source cannot be passed).

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// t1975ConsolidatedCard seeds card t1 in a source run, consolidates the source
// into a live survivor (the source row stays behind as abandoned, per
// abandonWithPointer), and returns the queue store. With assigned the card is
// assigned to lane-1 before the consolidation; with mirrorToSurvivor the
// queue's current-dispatch record is pointed at the survivor afterwards (the
// state a later operation must preserve). The queue's dispatch record is
// seeded on the SOURCE run first, as the dispatch hook writes it.
func t1975ConsolidatedCard(t *testing.T, assigned, mirrorToSurvivor bool) (string, *factory.BacklogStore) {
	t.Helper()
	ctx := context.Background()
	root, store := nmBase(t, factory.BacklogStatePicked)
	// RefreshDispatchCurrent only refreshes a row that already exists, so the
	// fixture seeds the dispatch hook's record for the source run first.
	if err := factory.RecordFactoryCardAssignment(root, "run-src", "t1", "lane-1", ""); err != nil {
		t.Fatal(err)
	}
	db := fcOpen(t, root)
	for id, pid := range map[string]int{"run-src": 99999999, "run-tgt": os.Getpid()} {
		start := "gone-before-assign"
		if pid == os.Getpid() {
			start = homestate.CurrentProcessFingerprint()
		}
		if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start}); err != nil {
			t.Fatal(err)
		}
	}
	card, err := db.RecordPicked(ctx, "run-src", "t1", homestate.CardFields{}, "dispatch", factoryCardNow())
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordDispatchBinding(ctx, "t1", "run-src", factoryCardNow()); err != nil {
		t.Fatal(err)
	}
	if assigned {
		if _, err := db.Transition(ctx, homestate.TransitionRequest{
			RunID: "run-src", CardID: "t1", To: homestate.CardAssigned, ExpectedVersion: card.Version,
			Actor: "assign", Owner: "lane-1", Now: factoryCardNow(),
		}); err != nil {
			t.Fatalf("assign the card in the source run: %v", err)
		}
	}
	opts := homestate.ReconcileOptions{Classify: func(pid int, _ string) homestate.OwnerClassification {
		if pid == os.Getpid() {
			return homestate.OwnerLive
		}
		return homestate.OwnerDead
	}}
	owners, err := db.ClassifyRuns(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]homestate.RunOwner{}
	for _, o := range owners {
		byID[o.RunID] = o
	}
	if _, err := db.RebindRun(ctx, homestate.RebindRequest{Source: byID["run-src"], Target: byID["run-tgt"], RequireStampBasis: true}, opts); err != nil {
		t.Fatalf("consolidate: %v", err)
	}
	_ = db.Close()
	if mirrorToSurvivor {
		if err := store.WithLock(func(l *factory.LockedBacklog) error {
			return l.RefreshDispatchCurrent("t1", "run-tgt", "lane-1")
		}); err != nil {
			t.Fatalf("point the queue's current dispatch at the survivor: %v", err)
		}
	}
	return root, store
}

// t1975MirrorRun reads the run the queue's current-dispatch record names for a
// card ("" when it names none).
func t1975MirrorRun(t *testing.T, store *factory.BacklogStore, cardID string) string {
	t.Helper()
	got := ""
	if err := store.WithLock(func(l *factory.LockedBacklog) error {
		rec, err := l.LoadPure()
		if err != nil {
			return err
		}
		for _, e := range rec.Runtime.DispatchCurrent {
			if e.CardID == cardID {
				got = e.RunID
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("read the queue's current dispatch: %v", err)
	}
	return got
}

// P1: a refused assign aimed at the retired source run changes neither store.
func TestRefusedAssignToAConsolidatedRunLeavesBothStoresUnchanged(t *testing.T) {
	root, store := t1975ConsolidatedCard(t, false, true)
	beforeDB := lcDump(t, root)
	if got := t1975MirrorRun(t, store, "t1"); got != "run-tgt" {
		t.Fatalf("fixture: the queue names %q for t1, want run-tgt", got)
	}

	_, stderr, err := reassignRun(t, "t1", "--run", "run-src")
	if err == nil {
		t.Fatalf("assign against a run a consolidation retired was accepted")
	}
	if msg := stderr + err.Error(); !strings.Contains(msg, "not the currently active factory run") && !strings.Contains(msg, "retired by a consolidation") {
		t.Fatalf("err=%v stderr=%q, want a refusal naming the retired run", err, stderr)
	}
	if got := t1975MirrorRun(t, store, "t1"); got != "run-tgt" {
		t.Errorf("the refused assign moved the queue's current dispatch for t1 to %q; both stores must stay as they were (survivor run-tgt)", got)
	}
	if after := lcDump(t, root); after != beforeDB {
		t.Errorf("factory DB changed although the assign was refused:\n--- before\n%s\n--- after\n%s", beforeDB, after)
	}
}

// P2: after a consolidation the leader can still reassign the card in the
// survivor. The queue's current-dispatch record was written for the SOURCE run
// by the dispatch hook; the merge must not leave it stranded there.
func TestReassignInTheSurvivorRunAfterAConsolidation(t *testing.T) {
	_, store := t1975ConsolidatedCard(t, true, false)
	if got := t1975MirrorRun(t, store, "t1"); got != "run-src" {
		t.Fatalf("fixture: the queue names %q for t1 before the reassign, want run-src (the dispatch hook's record)", got)
	}

	// The remedy the stale-run guard names is "pass --run run-src": observe what
	// the retired source answers, so the report can say whether it is reachable.
	_, srcStderr, srcErr := reassignRun(t, "t1", "--run", "run-src", "--reassign", "--to", "lane-2")
	t.Logf("reassign --run run-src (the remedy the guard names): err=%v stderr=%q", srcErr, srcStderr)

	out, stderr, err := reassignRun(t, "t1", "--run", "run-tgt", "--reassign", "--to", "lane-2")
	if err != nil {
		t.Fatalf("reassign in the survivor run was refused: %v (stderr %q)", err, stderr)
	}
	if !strings.Contains(out, "reassigned from") {
		t.Fatalf("output %q does not report the reassignment", out)
	}
	if got := t1975MirrorRun(t, store, "t1"); got != "run-tgt" {
		t.Errorf("after the reassign the queue names %q for t1, want the survivor run-tgt", got)
	}
}

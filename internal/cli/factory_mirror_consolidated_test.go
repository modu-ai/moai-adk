package cli

// factory_mirror_consolidated_test.go — REQ-FCR-024 at the dispatch mirror (card
// t1963, codex review P1 on 8b10f67066): the cli mirror (writeFactoryAssignment)
// records the card's row and binding for a dispatch. Aimed at a run a
// consolidation retired it used to recreate the row and re-point the binding
// there; the storage layer now refuses.

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

func TestDispatchMirrorRefusesAConsolidatedSource(t *testing.T) {
	ctx := context.Background()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	db := fcOpen(t, root)
	for id, pid := range map[string]int{"run-src": 99999999, "run-tgt": os.Getpid()} {
		start := "gone-before-mirror"
		if pid == os.Getpid() {
			start = homestate.CurrentProcessFingerprint()
		}
		if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.RecordPicked(ctx, "run-src", "t1", homestate.CardFields{}, "dispatch", factoryCardNow()); err != nil {
		t.Fatal(err)
	}
	if err := db.RecordDispatchBinding(ctx, "t1", "run-src", factoryCardNow()); err != nil {
		t.Fatal(err)
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
	before := lcDump(t, root)

	err = writeFactoryAssignment(ctx, root, store, "run-src", "t1", "lane-1")
	if err == nil {
		t.Fatalf("the mirror recreated work in a run a consolidation retired (rows in run-src: %d)",
			lcInt(t, root, `SELECT count(*) FROM cards WHERE run_id='run-src' AND card_id='t1' AND state NOT IN ('done','failed','abandoned')`))
	}
	if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
		t.Fatalf("err = %v, want a typed refusal wrapping homestate.ErrRunRetiredByConsolidation", err)
	}
	if after := lcDump(t, root); after != before {
		t.Fatalf("factory DB changed although the mirror was refused:\n--- before\n%s\n--- after\n%s", before, after)
	}
}

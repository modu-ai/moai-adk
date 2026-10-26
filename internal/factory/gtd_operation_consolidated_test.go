package factory

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// gtd_operation_consolidated_test.go — REQ-FCR-024 / AC-FCR-017 (10) (card
// t1963, sync-audit F-1): after a consolidation retired a run, a dispatch
// repair/retry aimed at that run — the queue's current-dispatch record still
// names it — must be refused with a typed error. Before the refusal existed the
// repair recreated the card's row in the retired run and flipped the dispatch
// binding back to it, leaving a non-terminal row in a run that can never be
// retired and a binding to a run the leader no longer owns.

// gtdFactoryDump renders the four tables the repair could touch, row by row.
func gtdFactoryDump(t *testing.T, root string) string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var b strings.Builder
	for _, table := range []string{"cards", "card_dispatch", "runs", "events"} {
		b.WriteString("## " + table + "\n")
		rows, err := db.DB.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`) // SQL: fixed table list
		if err != nil {
			t.Fatal(err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			for i := range vals {
				vals[i] = new(sql.NullString)
			}
			if err := rows.Scan(vals...); err != nil {
				t.Fatal(err)
			}
			parts := make([]string, len(vals))
			for i, v := range vals {
				parts[i] = v.(*sql.NullString).String
			}
			b.WriteString(strings.Join(parts, "|") + "\n")
		}
		_ = rows.Close()
	}
	return b.String()
}

// gtdConsolidatedFixture builds the post-consolidation state: card c1 was
// re-keyed from the dead source run-src into the live run-tgt, run-src is
// retired with rebound_to, the binding follows the card — while the queue's
// current-dispatch record still names run-src and the item is picked.
func gtdConsolidatedFixture(t *testing.T) (string, *BacklogStore, string) {
	t.Helper()
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	for id, pid := range map[string]int{"run-src": 20, "run-tgt": 21} {
		if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: id, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: "fp"}); err != nil {
			t.Fatal(err)
		}
	}
	_ = db.Close()
	gtdPlaceRow(t, root, card, "run-src")
	gtdAssignWithSeveredFactoryHalf(t, s, "run-src", card, "worker-1") // the queue half of the in-flight dispatch
	db, err = homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.RecordDispatchBinding(ctx, card, "run-src", time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	classify := func(pid int, _ string) homestate.OwnerClassification {
		if pid == 21 {
			return homestate.OwnerLive
		}
		return homestate.OwnerDead
	}
	opts := homestate.ReconcileOptions{Classify: classify}
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
	// Fixture premises: the card moved, the binding followed, and the queue
	// still names the retired source.
	if got := gtdBindingRun(t, root, card); got != "run-tgt" {
		t.Fatalf("fixture: binding = %q, want run-tgt", got)
	}
	if cur, ok := gtdCurrentDispatchRun(t, s, card); !ok || cur != "run-src" {
		t.Fatalf("fixture: queue current dispatch = (%q, %v), want run-src", cur, ok)
	}
	return root, s, card
}

func TestRepairGTDDispatchRefusesConsolidatedSource(t *testing.T) {
	root, s, card := gtdConsolidatedFixture(t)
	before := gtdFactoryDump(t, root)

	err := repairGTDDispatchRecord(context.Background(), s, gtdFactoryDBForStore(s), card, "run-src")
	if err == nil {
		t.Fatalf("the repair recreated work in a run a consolidation retired: card row in run-src=%v, binding=%q",
			gtdRowExists(t, root, "run-src", card), gtdBindingRun(t, root, card))
	}
	if !errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
		t.Fatalf("err = %v, want a typed refusal wrapping homestate.ErrRunRetiredByConsolidation", err)
	}
	if after := gtdFactoryDump(t, root); after != before {
		t.Fatalf("factory DB changed although the repair was refused:\n--- before\n%s\n--- after\n%s", before, after)
	}
	// Exactly one non-terminal row for the card, under the surviving run, and
	// the binding still resolves to it.
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var live int
	if err := db.DB.QueryRow(`SELECT count(*) FROM cards WHERE card_id=? AND state NOT IN ('done','failed','abandoned')`, card).Scan(&live); err != nil || live != 1 {
		t.Fatalf("non-terminal rows for %s = %d (err %v), want exactly one", card, live, err)
	}
	row, linked, err := db.RecordedCardRowReadonly(context.Background(), card)
	if err != nil || !linked || row.RunID != "run-tgt" {
		t.Fatalf("recorded row = (run %q, linked %v, err %v), want run-tgt", row.RunID, linked, err)
	}
}

func gtdRowExists(t *testing.T, root, runID, card string) bool {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	_, err = db.LoadCard(context.Background(), runID, card)
	return err == nil
}

// The guard is scoped to consolidation: a run retired the ordinary way
// (no rebound_to) is not caught by it, so the existing retry behaviour of the
// repair path stays as it was.
func TestRepairGTDDispatchStillRepairsAnOrdinarilyRetiredRun(t *testing.T) {
	root, s, card := gtdReconcileFixture(t)
	ctx := context.Background()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.RecordRun(ctx, homestate.FactoryRun{RunID: "run-plain", ManifestJSON: "{}", LeadPID: 20, LeadProcessStart: "fp"}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RetireRunIfDead(ctx, "run-plain", homestate.ReconcileOptions{Classify: func(int, string) homestate.OwnerClassification { return homestate.OwnerDead }}); err != nil {
		t.Fatalf("retire: %v", err)
	}
	_ = db.Close()
	gtdAssignWithSeveredFactoryHalf(t, s, "run-plain", card, "worker-1")
	err = repairGTDDispatchRecord(ctx, s, gtdFactoryDBForStore(s), card, "run-plain")
	if errors.Is(err, homestate.ErrRunRetiredByConsolidation) {
		t.Fatalf("an ordinary retirement was treated as a consolidation: %v", err)
	}
}

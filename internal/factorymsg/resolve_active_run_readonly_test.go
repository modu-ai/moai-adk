package factorymsg

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/homestate"
)

// factorySnapshot renders every row of every table a reconciliation could touch, one text line per row,
// so two snapshots are equal exactly when the rows are equal in every column.
func factorySnapshot(t *testing.T, root string) map[string][]string {
	t.Helper()
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	snap := map[string][]string{}
	for _, q := range []struct{ table, query string }{
		{"runs", `SELECT * FROM runs ORDER BY rowid`},
		{"cards", `SELECT * FROM cards ORDER BY rowid`},
		{"workers", `SELECT * FROM workers ORDER BY rowid`},
		{"events", `SELECT * FROM events ORDER BY rowid`},
		{"card_dispatch", `SELECT * FROM card_dispatch ORDER BY rowid`},
	} {
		table := q.table
		rows, err := db.DB.Query(q.query)
		if err != nil {
			t.Fatalf("snapshot %s: %v", table, err)
		}
		cols, err := rows.Columns()
		if err != nil {
			_ = rows.Close()
			t.Fatalf("snapshot %s columns: %v", table, err)
		}
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				_ = rows.Close()
				t.Fatalf("snapshot %s scan: %v", table, err)
			}
			snap[table] = append(snap[table], fmt.Sprint(vals...))
		}
		if err := rows.Close(); err != nil {
			t.Fatalf("snapshot %s close: %v", table, err)
		}
	}
	return snap
}

// deadNextToLiveFixture builds the shape a reconciliation would act on: run-dead (owner process gone) holds a
// non-terminal card and a dead worker row, run-live (owner this test process) is the survivor.
func deadNextToLiveFixture(t *testing.T) string {
	t.Helper()
	root := retireSandbox(t)
	deadPID, deadStart := deadIdentity(t)
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-dead", deadPID, deadStart)
	recordRun(t, root, "run-live", livePID, liveStart)
	db, err := homestate.OpenFactory(root)
	if err != nil {
		t.Fatalf("open factory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,version,updated_at,stage) VALUES('run-dead','d-run','lane-9','run',1,'2026-09-26T00:00:00Z','run')`); err != nil {
		t.Fatalf("place dead-run card: %v", err)
	}
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,registered_at,heartbeat_at,run_id) VALUES('lane-9',?,'2026-09-28T00:00:00Z','2026-09-28T00:00:00Z','run-dead')`, deadPID); err != nil {
		t.Fatalf("place dead-run worker: %v", err)
	}
	return root
}

// A dead-owner run next to a live one: the read-only resolver reports both with their owner
// classification and writes no row of any table. The control proves the snapshot would see a write: the
// reconciling resolver, on the same shape, retires the dead run and moves its card.
func TestResolveActiveRunReadOnlyDeadOwnerNextToLiveWritesNothing(t *testing.T) {
	root := deadNextToLiveFixture(t)
	before := factorySnapshot(t, root)

	id, err := ResolveActiveRunReadOnly(context.Background(), root)
	var ambiguous *AmbiguousRunsError
	if !errors.As(err, &ambiguous) {
		t.Fatalf("err = %v, want *AmbiguousRunsError listing both active runs", err)
	}
	if id != "" {
		t.Fatalf("returned run id %q on an undecided set", id)
	}
	got := map[string]homestate.OwnerClassification{}
	for _, o := range ambiguous.Owners {
		got[o.RunID] = o.Classification
	}
	want := map[string]homestate.OwnerClassification{"run-dead": homestate.OwnerDead, "run-live": homestate.OwnerLive}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("owners = %v, want %v", got, want)
	}
	if after := factorySnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("the read-only resolver changed the record:\nbefore %v\nafter  %v", before, after)
	}

	control := deadNextToLiveFixture(t)
	controlBefore := factorySnapshot(t, control)
	if _, err := ResolveActiveRun(context.Background(), control, ""); err != nil {
		t.Fatalf("control resolve: %v", err)
	}
	if reflect.DeepEqual(controlBefore, factorySnapshot(t, control)) {
		t.Fatal("control: the reconciling resolver changed nothing, so the snapshot cannot witness a write")
	}
}

// The run cannot be decided (two live owners): the read-only resolver refuses with the typed ambiguity
// error and writes nothing. The --run instruction is the caller's wording (the renew verb appends it), so
// the typed Owners field is what carries the candidates.
func TestResolveActiveRunReadOnlyUndecidedRefusesAndWritesNothing(t *testing.T) {
	root := retireSandbox(t)
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-1", livePID, liveStart)
	recordRun(t, root, "run-2", livePID, liveStart)
	before := factorySnapshot(t, root)

	id, err := ResolveActiveRunReadOnly(context.Background(), root)
	if err == nil || !strings.HasPrefix(err.Error(), "AMBIGUOUS_FACTORY: ") {
		t.Fatalf("err = %v, want AMBIGUOUS_FACTORY", err)
	}
	if id != "" {
		t.Fatalf("returned run id %q on an undecided set", id)
	}
	for _, run := range []string{"run-1", "run-2"} {
		if !strings.Contains(err.Error(), run) {
			t.Errorf("error %q does not name %s", err, run)
		}
	}
	if after := factorySnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("a refusal changed the record:\nbefore %v\nafter  %v", before, after)
	}
}

// Zero and one active run answer exactly as ResolveActiveRun does, with no write.
func TestResolveActiveRunReadOnlyZeroAndSole(t *testing.T) {
	root := retireSandbox(t)
	id, err := ResolveActiveRunReadOnly(context.Background(), root)
	if err == nil || !errors.Is(err, ErrRunNotActive) || id != "" {
		t.Fatalf("zero active: (%q, %v), want ErrRunNotActive", id, err)
	}
	livePID, liveStart := liveIdentity(t)
	recordRun(t, root, "run-only", livePID, liveStart)
	before := factorySnapshot(t, root)
	id, err = ResolveActiveRunReadOnly(context.Background(), root)
	if err != nil || id != "run-only" {
		t.Fatalf("sole active: (%q, %v), want run-only", id, err)
	}
	if after := factorySnapshot(t, root); !reflect.DeepEqual(before, after) {
		t.Fatalf("the sole-run answer changed the record:\nbefore %v\nafter  %v", before, after)
	}
}

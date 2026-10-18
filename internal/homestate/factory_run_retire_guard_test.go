package homestate

// factory_run_retire_guard_test.go — SPEC-FACTORY-RETIRE-GUARD-001 M1
// (card t1852): the operator retirement refuses a run that still holds
// non-terminal cards, flips only under the classified owner stamp, and is a
// no-op for a run that is already retired.

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func deadOpts() ReconcileOptions { return ReconcileOptions{Classify: fixedClassifier(OwnerDead)} }

func mustInsertCardOwned(t *testing.T, db *FactoryDB, runID, cardID, state, owner string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,owner_label,state,updated_at) VALUES(?,?,?,?,?)`,
		runID, cardID, owner, state, "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert card %s: %v", cardID, err)
	}
}

// tableDump renders a table's rows as one string so a before/after comparison
// covers every column, updated_at included.
func tableDump(t *testing.T, db *FactoryDB, query string, args ...any) string {
	t.Helper()
	rows, err := db.DB.Query(query, args...)
	if err != nil {
		t.Fatalf("dump %q: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns: %v", err)
	}
	var out []string
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan: %v", err)
		}
		out = append(out, fmt.Sprint(vals...))
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return strings.Join(out, "\n")
}

func runRow(t *testing.T, db *FactoryDB, runID string) string {
	t.Helper()
	return tableDump(t, db, `SELECT * FROM runs WHERE run_id=?`, runID)
}

func retiredEvents(t *testing.T, db *FactoryDB, runID string) int {
	t.Helper()
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM events WHERE run_id=? AND kind='run.retired'`, runID).Scan(&n); err != nil {
		t.Fatalf("count run.retired: %v", err)
	}
	return n
}

func retiredPayloadJSON(t *testing.T, db *FactoryDB, runID string) string {
	t.Helper()
	var p string
	if err := db.DB.QueryRow(`SELECT payload_json FROM events WHERE run_id=? AND kind='run.retired'`, runID).Scan(&p); err != nil {
		t.Fatalf("read run.retired payload for %s: %v", runID, err)
	}
	return p
}

func runStamp(t *testing.T, db *FactoryDB, runID string) (int, string) {
	t.Helper()
	var pid int
	var start string
	if err := db.DB.QueryRow(`SELECT lead_pid, lead_process_start FROM runs WHERE run_id=?`, runID).Scan(&pid, &start); err != nil {
		t.Fatalf("read stamp for %s: %v", runID, err)
	}
	return pid, start
}

// AC-RRG-001 — a dead-owner active run that holds non-terminal cards is refused
// with the typed cards error and nothing changes.
func TestRetireRunIfDeadRefusesInProgressCards(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-r", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
	mustInsertCardOwned(t, db, "run-r", "c-b", "merged-local", "lane-2")
	mustInsertCardOwned(t, db, "run-r", "c-a", "needs-decision", "lane-1")
	mustInsertCardOwned(t, db, "run-r", "c-c", "picked", "")
	mustInsertCardOwned(t, db, "run-r", "c-z", "done", "lane-3")
	before := runRow(t, db, "run-r")

	class, err := db.RetireRunIfDead(context.Background(), "run-r", deadOpts())
	if err == nil {
		t.Fatalf("RetireRunIfDead retired a run holding non-terminal cards (class %q)", class)
	}
	if !errors.Is(err, ErrRunHasInProgressCards) {
		t.Fatalf("err = %v, want errors.Is ErrRunHasInProgressCards", err)
	}
	if errors.Is(err, ErrRunOwnerNotDead) {
		t.Fatalf("cards refusal must not read as owner-not-dead: %v", err)
	}
	var cardsErr *RunInProgressCardsError
	if !errors.As(err, &cardsErr) {
		t.Fatalf("errors.As *RunInProgressCardsError failed for %v", err)
	}
	want := []InProgressCard{
		{CardID: "c-a", State: "needs-decision", OwnerLabel: "lane-1"},
		{CardID: "c-b", State: "merged-local", OwnerLabel: "lane-2"},
		{CardID: "c-c", State: "picked", OwnerLabel: ""},
	}
	if cardsErr.RunID != "run-r" || !reflect.DeepEqual(cardsErr.Cards, want) {
		t.Fatalf("typed error = run %q cards %+v, want run-r %+v", cardsErr.RunID, cardsErr.Cards, want)
	}
	if class != OwnerDead {
		t.Fatalf("classification = %q, want dead alongside the refusal", class)
	}
	if got := runStatus(t, db, "run-r"); got != "active" {
		t.Fatalf("status = %q, want active", got)
	}
	if n := retiredEvents(t, db, "run-r"); n != 0 {
		t.Fatalf("run.retired events = %d, want 0", n)
	}
	if after := runRow(t, db, "run-r"); after != before {
		t.Fatalf("runs row changed on refusal:\nbefore %s\nafter  %s", before, after)
	}
	if msg := cardsErr.Error(); !strings.Contains(msg, "run-r") || !strings.Contains(msg, "c-c (picked, -)") {
		t.Fatalf("Error() = %q does not name the run and every card", msg)
	}

	// Controls: a live or indeterminate owner that also holds a card keeps the
	// owner-not-dead refusal — the owner gate precedes the card read.
	for _, c := range []OwnerClassification{OwnerLive, OwnerIndeterminate} {
		_, err := db.RetireRunIfDead(context.Background(), "run-r", ReconcileOptions{Classify: fixedClassifier(c)})
		if !errors.Is(err, ErrRunOwnerNotDead) || errors.Is(err, ErrRunHasInProgressCards) {
			t.Fatalf("%s owner holding cards: err = %v, want ErrRunOwnerNotDead only", c, err)
		}
	}
}

// AC-RRG-001 controls — terminal cards and worker rows do not block retirement.
func TestRetireRunIfDeadRetiresWhenOnlyTerminalCardsOrWorkers(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-term", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
	mustInsertCard(t, db, "run-term", "c1", "done")
	mustInsertCard(t, db, "run-term", "c2", "failed")
	mustInsertCard(t, db, "run-term", "c3", "abandoned")
	mustRecordRun(t, db, FactoryRun{RunID: "run-work", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
	mustInsertWorker(t, db, "run-work", "w1")

	for _, id := range []string{"run-term", "run-work"} {
		class, err := db.RetireRunIfDead(context.Background(), id, deadOpts())
		if err != nil || class != OwnerDead {
			t.Fatalf("%s: class %q err %v, want dead and nil", id, class, err)
		}
		if got := runStatus(t, db, id); got != "retired" {
			t.Fatalf("%s status = %q, want retired", id, got)
		}
		if n := retiredEvents(t, db, id); n != 1 {
			t.Fatalf("%s run.retired events = %d, want 1", id, n)
		}
		if p := retiredPayloadJSON(t, db, id); p != `{"classification":"dead","basis":"stamp"}` {
			t.Fatalf("%s payload = %s", id, p)
		}
	}
}

// AC-RRG-002 — the card read runs inside the write-locked transaction: a second
// handle's write at the moment the read finished is refused as busy.
func TestRetireRunIfDeadCardReadHoldsWriteLock(t *testing.T) {
	db := openSandboxFactory(t)
	mustRecordRun(t, db, FactoryRun{RunID: "run-lock", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
	// Opened BEFORE the call: openFactoryPathBusy runs DDL on open and would
	// block behind the retiring transaction.
	second, err := openFactoryPathBusy(db.Path, 150*time.Millisecond)
	if err != nil {
		t.Fatalf("open second handle: %v", err)
	}
	t.Cleanup(func() { _ = second.Close() })

	var hookCalls int
	var hookErr error
	retireAfterCardRead = func() {
		hookCalls++
		_, hookErr = second.DB.Exec(`INSERT INTO cards(run_id,card_id,state,updated_at) VALUES('run-lock','raced','run','2026-10-14T00:00:00Z')`)
	}
	t.Cleanup(func() { retireAfterCardRead = nil })

	class, err := db.RetireRunIfDead(context.Background(), "run-lock", deadOpts())
	if err != nil || class != OwnerDead {
		t.Fatalf("retire: class %q err %v, want dead and nil", class, err)
	}
	if hookCalls != 1 {
		t.Fatalf("card-read hook calls = %d, want 1", hookCalls)
	}
	if hookErr == nil {
		t.Fatalf("second handle's insert committed while the card read ran: the read is outside the write-locked transaction")
	}
	if msg := strings.ToLower(hookErr.Error()); !strings.Contains(msg, "locked") && !strings.Contains(msg, "busy") {
		t.Fatalf("second handle's insert failed for the wrong reason: %v", hookErr)
	}
	if got := runStatus(t, db, "run-lock"); got != "retired" {
		t.Fatalf("status = %q, want retired", got)
	}
	var n int
	if err := db.DB.QueryRow(`SELECT count(*) FROM cards WHERE run_id='run-lock'`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("cards for run-lock = %d (err %v), want 0", n, err)
	}
}

// AC-RRG-003 — the flip is a compare-and-set on the classified owner stamp.
func TestRetireRunIfDeadFlipIsCASOnClassifiedStamp(t *testing.T) {
	adopt := func(t *testing.T, db *FactoryDB, runID string) {
		t.Helper()
		err := db.AdoptRun(context.Background(), RunOwner{RunID: runID, LeadPID: 1, LeadProcessStart: "x"}, 4242, "adopter-start", "test")
		if err != nil {
			t.Fatalf("competing adopt: %v", err)
		}
	}
	// racing returns a Classify closure that runs race() on its first call only,
	// then answers dead.
	racing := func(race func()) ReconcileOptions {
		called := false
		return ReconcileOptions{Classify: func(int, string) OwnerClassification {
			if !called {
				called = true
				race()
			}
			return OwnerDead
		}}
	}
	ctx := context.Background()

	t.Run("a restamp before the flip, card-free run", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-a", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
		_, err := db.RetireRunIfDead(ctx, "run-a", racing(func() { adopt(t, db, "run-a") }))
		if !errors.Is(err, ErrRunOwnerNotDead) || !strings.Contains(err.Error(), "changed") {
			t.Fatalf("err = %v, want ErrRunOwnerNotDead stating the owner or status changed", err)
		}
		if got := runStatus(t, db, "run-a"); got != "active" {
			t.Fatalf("status = %q, want active", got)
		}
		if pid, start := runStamp(t, db, "run-a"); pid != 4242 || start != "adopter-start" {
			t.Fatalf("stamp = (%d,%q), want the adopter's (4242,adopter-start)", pid, start)
		}
		if n := retiredEvents(t, db, "run-a"); n != 0 {
			t.Fatalf("run.retired events = %d, want 0", n)
		}
	})

	t.Run("a restamp before the flip, run holding a card", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-b", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
		mustInsertCardOwned(t, db, "run-b", "cb1", "run", "lane-1")
		_, err := db.RetireRunIfDead(ctx, "run-b", racing(func() { adopt(t, db, "run-b") }))
		if !errors.Is(err, ErrRunOwnerNotDead) || !strings.Contains(err.Error(), "changed") {
			t.Fatalf("err = %v, want ErrRunOwnerNotDead stating the owner or status changed", err)
		}
		if errors.Is(err, ErrRunHasInProgressCards) {
			t.Fatalf("a restamped run must not produce the cards error: %v", err)
		}
		if got := runStatus(t, db, "run-b"); got != "active" {
			t.Fatalf("status = %q, want active", got)
		}
		if pid, start := runStamp(t, db, "run-b"); pid != 4242 || start != "adopter-start" {
			t.Fatalf("stamp = (%d,%q), want (4242,adopter-start)", pid, start)
		}
		if n := retiredEvents(t, db, "run-b"); n != 0 {
			t.Fatalf("run.retired events = %d, want 0", n)
		}
	})

	t.Run("a concurrent retirer wins", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-c", ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
		_, err := db.RetireRunIfDead(ctx, "run-c", racing(func() {
			if _, nerr := db.RetireRunIfDead(ctx, "run-c", deadOpts()); nerr != nil {
				t.Errorf("nested retirement: %v", nerr)
			}
		}))
		if !errors.Is(err, ErrRunOwnerNotDead) {
			t.Fatalf("err = %v, want an error wrapping ErrRunOwnerNotDead", err)
		}
		if n := retiredEvents(t, db, "run-c"); n != 1 {
			t.Fatalf("run.retired events = %d, want exactly 1", n)
		}
	})

	t.Run("a (0,'') stamp classified dead through the lead peer", func(t *testing.T) {
		db := openSandboxFactory(t)
		mustRecordRun(t, db, FactoryRun{RunID: "run-d", ManifestJSON: "{}", LeadPID: 7, LeadProcessStart: "y"})
		if err := db.ClearRunOwner(ctx, "run-d"); err != nil {
			t.Fatalf("clear owner: %v", err)
		}
		opts := ReconcileOptions{
			Classify: fixedClassifier(OwnerDead),
			Fallback: func(string) (int, string, bool) { return 31, "peer-start", true },
		}
		class, err := db.RetireRunIfDead(ctx, "run-d", opts)
		if err != nil || class != OwnerDead {
			t.Fatalf("class %q err %v, want dead and nil", class, err)
		}
		if got := runStatus(t, db, "run-d"); got != "retired" {
			t.Fatalf("status = %q, want retired", got)
		}
		if p := retiredPayloadJSON(t, db, "run-d"); p != `{"classification":"dead","basis":"peer"}` {
			t.Fatalf("payload = %s, want basis peer", p)
		}
	})
}

// AC-RRG-004 — an already-retired dead-owner run is an idempotent no-op that
// reads no cards and writes nothing.
func TestRetireRunIfDeadAlreadyRetiredIsNoOp(t *testing.T) {
	ctx := context.Background()
	seed := func(t *testing.T, db *FactoryDB, runID string, retired bool) {
		t.Helper()
		mustRecordRun(t, db, FactoryRun{RunID: runID, ManifestJSON: "{}", LeadPID: 1, LeadProcessStart: "x"})
		mustInsertCardOwned(t, db, runID, runID+"-1", "needs-decision", "lane-1")
		mustInsertCardOwned(t, db, runID, runID+"-2", "picked", "")
		if retired {
			if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id=?`, runID); err != nil {
				t.Fatalf("seed retired: %v", err)
			}
			if _, err := db.DB.Exec(`INSERT INTO events(run_id,kind,payload_json,created_at) VALUES(?,'run.retired','{}','2026-10-14T00:00:00Z')`, runID); err != nil {
				t.Fatalf("seed event: %v", err)
			}
		}
	}
	snapshot := func(t *testing.T, db *FactoryDB, runID string) string {
		t.Helper()
		return runRow(t, db, runID) + "\n--\n" +
			tableDump(t, db, `SELECT * FROM cards WHERE run_id=? ORDER BY card_id`, runID) + "\n--\n" +
			tableDump(t, db, `SELECT * FROM events WHERE run_id=? ORDER BY seq`, runID)
	}

	t.Run("retired run", func(t *testing.T) {
		db := openSandboxFactory(t)
		seed(t, db, "run-ret", true)
		before := snapshot(t, db, "run-ret")
		hookCalls := 0
		retireAfterCardRead = func() { hookCalls++ }
		t.Cleanup(func() { retireAfterCardRead = nil })

		class, err := db.RetireRunIfDead(ctx, "run-ret", deadOpts())
		retireAfterCardRead = nil
		if err != nil || class != OwnerDead {
			t.Fatalf("class %q err %v, want dead and nil", class, err)
		}
		if hookCalls != 0 {
			t.Fatalf("the card guard ran for a retired run (%d card reads)", hookCalls)
		}
		if after := snapshot(t, db, "run-ret"); after != before {
			t.Fatalf("a no-op wrote:\nbefore %s\nafter  %s", before, after)
		}
		if n := retiredEvents(t, db, "run-ret"); n != 1 {
			t.Fatalf("run.retired events = %d, want 1", n)
		}
		cards, err := db.InProgressCards(ctx, "run-ret")
		if err != nil || len(cards) != 2 {
			t.Fatalf("InProgressCards = %+v err %v, want the two stranded cards", cards, err)
		}
	})

	t.Run("retired run whose owner is live is still refused", func(t *testing.T) {
		db := openSandboxFactory(t)
		seed(t, db, "run-retlive", true)
		_, err := db.RetireRunIfDead(ctx, "run-retlive", ReconcileOptions{Classify: fixedClassifier(OwnerLive)})
		if !errors.Is(err, ErrRunOwnerNotDead) {
			t.Fatalf("err = %v, want ErrRunOwnerNotDead", err)
		}
	})

	t.Run("active run holding the same cards is refused", func(t *testing.T) {
		db := openSandboxFactory(t)
		seed(t, db, "run-act", false)
		_, err := db.RetireRunIfDead(ctx, "run-act", deadOpts())
		if !errors.Is(err, ErrRunHasInProgressCards) {
			t.Fatalf("err = %v, want the cards refusal", err)
		}
	})
}

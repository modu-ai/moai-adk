package homestate

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

// SPEC-LANE-JOIN-RUN-SELECT-001 REQ-LJR-001 (card t1853): the read path that
// supplies the lane-join chooser's per-run facts.

// seedFactsRun records one active run whose owner stamp can never classify
// live (the pid exceeds every platform's pid_max), so the production batch
// probe classifies it dead without any injected classifier.
func seedFactsRun(t *testing.T, db *FactoryDB, runID string) {
	t.Helper()
	mustRecordRun(t, db, FactoryRun{
		RunID: runID, ManifestJSON: "{}", LeadPID: 99999999, LeadProcessStart: "gone-before-select",
	})
}

func seedFactsCard(t *testing.T, db *FactoryDB, runID, cardID, state, heartbeat, updated string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,heartbeat_at,updated_at) VALUES(?,?,?,?,?)`,
		runID, cardID, state, heartbeat, updated); err != nil {
		t.Fatalf("insert card %s/%s: %v", runID, cardID, err)
	}
}

func seedFactsWorker(t *testing.T, db *FactoryDB, runID, label, heartbeat string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
		label, 1, runID, "2026-10-17T00:00:00Z", heartbeat); err != nil {
		t.Fatalf("insert worker %s: %v", label, err)
	}
}

func factsByRun(t *testing.T, db *FactoryDB) map[string]RunFacts {
	t.Helper()
	facts, err := db.ActiveRunFacts(context.Background(), ReconcileOptions{})
	if err != nil {
		t.Fatalf("ActiveRunFacts: %v", err)
	}
	byRun := make(map[string]RunFacts, len(facts))
	for _, f := range facts {
		byRun[f.RunID] = f
	}
	return byRun
}

// AC-LJR-001 (count): a run lists exactly the cards whose state sits outside
// the shared terminal set {done, failed, abandoned}.
func TestActiveRunFactsCountsNonTerminalCards(t *testing.T) {
	db := openSandboxFactory(t)
	seedFactsRun(t, db, "run-a")
	seedFactsRun(t, db, "run-empty")
	for i, state := range []string{"leased", "needs-decision", "merged-local", "done", "failed", "abandoned"} {
		seedFactsCard(t, db, "run-a", fmt.Sprintf("c%d", i), state, "", "2026-10-17T10:00:00Z")
	}
	// A retired run in the same database is never listed.
	seedFactsRun(t, db, "run-retired")
	if _, err := db.DB.Exec(`UPDATE runs SET status='retired' WHERE run_id='run-retired'`); err != nil {
		t.Fatalf("retire fixture run: %v", err)
	}

	byRun := factsByRun(t, db)
	if got := byRun["run-a"].InProgress; got != 3 {
		t.Errorf("run-a InProgress = %d, want 3 (leased, needs-decision, merged-local)", got)
	}
	if got, ok := byRun["run-empty"]; !ok || got.InProgress != 0 {
		t.Errorf("run-empty = %+v (listed %v), want listed with InProgress 0", got, ok)
	}
	if _, ok := byRun["run-retired"]; ok {
		t.Error("a retired run was listed; only status 'active' belongs in the fact set")
	}
	if len(byRun) != 2 {
		t.Errorf("listed %d runs, want 2", len(byRun))
	}
}

// AC-LJR-001 (last activity): the latest of worker heartbeats, card
// heartbeats and card update times by PARSED value, returned as the stored
// string. Stored times trim trailing zeros, so text order disagrees with time
// order on every row below.
func TestActiveRunFactsLastActivityUsesParsedTimes(t *testing.T) {
	db := openSandboxFactory(t)
	const day = "2026-10-17T"
	for _, id := range []string{"run-w", "run-h", "run-u", "run-e", "run-n"} {
		seedFactsRun(t, db, id)
	}
	// W: .1Z is 100ms, .11Z is 110ms — text order puts .1Z last.
	seedFactsWorker(t, db, "run-w", "w1", day+"10:00:00.1Z")
	seedFactsWorker(t, db, "run-w", "w2", day+"10:00:00.11Z")
	// H: the card heartbeat .5Z is later than the update time ...00Z, though
	// text order puts "Z" after ".".
	seedFactsCard(t, db, "run-h", "h1", "run", day+"10:00:00.5Z", day+"10:00:00Z")
	// U: empty heartbeat, so the update time is the only candidate.
	seedFactsCard(t, db, "run-u", "u1", "run", "", day+"10:00:00.25Z")
	// E: an unparseable value is skipped, not fatal, and a later card still counts.
	seedFactsCard(t, db, "run-e", "e1", "run", "", "not-a-time")
	seedFactsCard(t, db, "run-e", "e2", "run", "", day+"08:00:00Z")

	byRun := factsByRun(t, db)
	want := map[string]string{
		"run-w": day + "10:00:00.11Z",
		"run-h": day + "10:00:00.5Z",
		"run-u": day + "10:00:00.25Z",
		"run-e": day + "08:00:00Z",
		"run-n": "",
	}
	for id, w := range want {
		got, ok := byRun[id]
		if !ok {
			t.Errorf("%s: not listed", id)
			continue
		}
		if got.LastActivity != w {
			t.Errorf("%s LastActivity = %q, want %q", id, got.LastActivity, w)
		}
	}
}

func dumpTable(t *testing.T, db *FactoryDB, query string) [][]string {
	t.Helper()
	rows, err := db.DB.Query(query)
	if err != nil {
		t.Fatalf("dump %q: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	cols, err := rows.Columns()
	if err != nil {
		t.Fatalf("columns: %v", err)
	}
	var out [][]string
	for rows.Next() {
		cells := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range cells {
			ptrs[i] = &cells[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan: %v", err)
		}
		line := make([]string, len(cols))
		for i, c := range cells {
			line[i] = fmt.Sprint(c)
		}
		out = append(out, line)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows: %v", err)
	}
	return out
}

// AC-LJR-001 (read-only): a dead-owner run with no cards and no workers is
// the exact shape the automatic reconciler retires. The read path lists it,
// classifies it dead, and leaves the runs and events tables untouched.
func TestActiveRunFactsIsReadOnly(t *testing.T) {
	db := openSandboxFactory(t)
	seedFactsRun(t, db, "run-z")

	runsBefore := dumpTable(t, db, `SELECT * FROM runs ORDER BY run_id`)
	eventsBefore := dumpTable(t, db, `SELECT * FROM events ORDER BY seq`)

	byRun := factsByRun(t, db)
	z, ok := byRun["run-z"]
	if !ok {
		t.Fatal("run-z was not listed")
	}
	if z.Classification != OwnerDead {
		t.Errorf("run-z Classification = %q, want %q", z.Classification, OwnerDead)
	}
	if z.Status != "active" {
		t.Errorf("run-z Status = %q, want active", z.Status)
	}
	if got := dumpTable(t, db, `SELECT * FROM runs ORDER BY run_id`); !reflect.DeepEqual(got, runsBefore) {
		t.Errorf("runs table changed across the read:\n before %v\n after  %v", runsBefore, got)
	}
	if got := dumpTable(t, db, `SELECT * FROM events ORDER BY seq`); !reflect.DeepEqual(got, eventsBefore) {
		t.Errorf("events table changed across the read:\n before %v\n after  %v", eventsBefore, got)
	}
}

// A read that cannot complete is an error, never a silently shorter fact set:
// the chooser must not present a run list built from a failed read.
func TestActiveRunFactsSurfacesReadFailures(t *testing.T) {
	cases := []struct {
		name    string
		prepare func(t *testing.T, db *FactoryDB) context.Context
	}{
		{"cancelled-context", func(t *testing.T, db *FactoryDB) context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}},
		{"cards-table-missing", func(t *testing.T, db *FactoryDB) context.Context {
			if _, err := db.DB.Exec(`DROP TABLE cards`); err != nil {
				t.Fatalf("drop cards: %v", err)
			}
			return context.Background()
		}},
		{"workers-table-missing", func(t *testing.T, db *FactoryDB) context.Context {
			if _, err := db.DB.Exec(`DROP TABLE workers`); err != nil {
				t.Fatalf("drop workers: %v", err)
			}
			return context.Background()
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			db := openSandboxFactory(t)
			seedFactsRun(t, db, "run-x")
			ctx := c.prepare(t, db)
			if facts, err := db.ActiveRunFacts(ctx, ReconcileOptions{}); err == nil {
				t.Fatalf("ActiveRunFacts = %+v, nil; want the read failure surfaced", facts)
			}
		})
	}
}

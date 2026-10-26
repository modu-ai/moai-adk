package cli

// factory_next_dead_run_test.go — REQ-FCR-019 / AC-FCR-023 (card t1963):
// `moai factory next` leasing a card abandons the card's earlier non-terminal
// row in a dead run, in the lease's own transaction, and leaves every row that
// is not the residue of a dead run exactly as it was.
//
// Every fixture is an isolated sandbox (fcFixture): the live factory database
// is never opened. Card ids map to the SPEC's names: t1=c9, t2=c7, t3=c8,
// t4=c6a, t5=c6b.

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

const (
	// deadRunPID can never name a live process: it exceeds every platform's pid
	// range, so the production probe reads it positively dead.
	deadRunPID = 99999999
)

// deadRunPlaceRun records a run whose owner stamp is the given pid.
func deadRunPlaceRun(t *testing.T, root, runID string, pid int, start, status string) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: runID, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
	if status != "active" {
		if _, err := db.DB.Exec(`UPDATE runs SET status=? WHERE run_id=?`, status, runID); err != nil {
			t.Fatalf("set run %s status: %v", runID, err)
		}
	}
}

func deadRunPlaceWorker(t *testing.T, root, label, runID string, pid int) {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	if _, err := db.DB.Exec(`INSERT INTO workers(label,pid,run_id,registered_at,heartbeat_at) VALUES(?,?,?,?,?)`,
		label, pid, runID, "2026-09-26T00:00:00Z", "2026-09-26T00:00:00Z"); err != nil {
		t.Fatalf("place worker %s: %v", label, err)
	}
}

func deadRunCardState(t *testing.T, root, runID, cardID string) string {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	c, err := db.LoadCard(context.Background(), runID, cardID)
	if err != nil {
		t.Fatalf("load %s/%s: %v", runID, cardID, err)
	}
	return c.State
}

func deadRunCount(t *testing.T, root, query string, args ...any) int {
	t.Helper()
	db := fcOpen(t, root)
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("count %q: %v", query, err)
	}
	return n
}

// deadRunFixture builds the AC-FCR-023 matrix and returns the root.
func deadRunFixture(t *testing.T) string {
	t.Helper()
	root, store := carryFixture(t) // run-old (retired) and run-new (active) rows exist; unused below
	fcQueue(t, store, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued, factory.BacklogStateQueued)
	liveStart := homestate.CurrentProcessFingerprint()
	deadRunPlaceRun(t, root, "run-now", os.Getpid(), liveStart, "active")
	deadRunPlaceRun(t, root, "run-dead", deadRunPID, "gone", "active")
	deadRunPlaceRun(t, root, "run-live", os.Getpid(), liveStart, "active")
	deadRunPlaceRun(t, root, "run-ret", deadRunPID, "gone", "retired")
	deadRunPlaceRun(t, root, "run-dead2", deadRunPID, "gone", "active")
	deadRunPlaceRun(t, root, "run-dead3", deadRunPID, "gone", "active")

	// t1 (c9): the dead run's row carries a stage and a reusable landing tree,
	// so the re-lease carry-over has something to read — and a worker row whose
	// pid is dead, which must not hold the run (REQ-FCR-022).
	landing := carryLanding(root, "t1")
	if err := os.MkdirAll(sdWorktreesDir(root), 0o755); err != nil {
		t.Fatal(err)
	}
	fcGit(t, root, "worktree", "add", "-q", landing, "-b", "WT-previous-tree")
	fcPlace(t, root, homestate.Card{RunID: "run-dead", CardID: "t1", State: homestate.CardRun, Stage: homestate.CardRun, WorktreePath: landing, EvidenceSHA: "sha-c9"})
	deadRunPlaceWorker(t, root, "lane-ghost", "run-dead", deadRunPID)
	fcPlace(t, root, homestate.Card{RunID: "run-ret", CardID: "t2", State: homestate.CardAssigned})
	fcPlace(t, root, homestate.Card{RunID: "run-live", CardID: "t3", State: homestate.CardRun})
	fcPlace(t, root, homestate.Card{RunID: "run-dead2", CardID: "t4", State: homestate.CardRun, LeaseHolder: "lane-x", LeaseExpiresAt: "2099-01-01T00:00:00Z"})
	fcPlace(t, root, homestate.Card{RunID: "run-dead3", CardID: "t5", State: homestate.CardRun, LeaseHolder: "lane-y", LeaseExpiresAt: "2020-01-01T00:00:00Z"})
	deadRunPlaceWorker(t, root, "lane-y", "run-dead3", os.Getpid())
	for i := 1; i <= 5; i++ {
		sdRegisterLane(t, root, "lane-"+string(rune('0'+i)))
	}
	t.Chdir(root)
	return root
}

func deadRunLease(t *testing.T, i int, cardID string) {
	t.Helper()
	sdLaneEnv(t, "lane-"+string(rune('0'+i)), "")
	out, stderr, err := runFactory(t, "next", "--run", "run-now", "--card", cardID)
	if err != nil {
		t.Fatalf("next --card %s: %v\nstdout: %s\nstderr: %s", cardID, err, out, stderr)
	}
}

func TestFactoryNextAbandonsDeadRunDuplicate(t *testing.T) {
	root := deadRunFixture(t)
	cardsBefore := deadRunCount(t, root, `SELECT count(*) FROM cards`)

	for i, id := range []string{"t1", "t2", "t3", "t4", "t5"} {
		deadRunLease(t, i+1, id)
	}

	if got := deadRunCardState(t, root, "run-dead", "t1"); got != homestate.CardAbandoned {
		t.Errorf("run-dead c9 = %s, want abandoned (a dead-pid worker row does not hold the run)", got)
	}
	if got := deadRunCardState(t, root, "run-ret", "t2"); got != homestate.CardAbandoned {
		t.Errorf("run-ret c7 = %s, want abandoned", got)
	}
	if got := deadRunCardState(t, root, "run-live", "t3"); got != homestate.CardRun {
		t.Errorf("run-live c8 = %s, want run (a live owner's row is untouched)", got)
	}
	if got := deadRunCardState(t, root, "run-dead2", "t4"); got != homestate.CardRun {
		t.Errorf("run-dead2 c6a = %s, want run (a valid lease protects the row)", got)
	}
	if got := deadRunCardState(t, root, "run-dead3", "t5"); got != homestate.CardRun {
		t.Errorf("run-dead3 c6b = %s, want run (a live worker protects the row)", got)
	}
	if n := deadRunCount(t, root, `SELECT count(*) FROM cards`); n != cardsBefore+5 {
		t.Errorf("cards rows = %d, want %d (only the five new run-now rows; nothing deleted)", n, cardsBefore+5)
	}
	if n := deadRunCount(t, root, `SELECT count(*) FROM cards WHERE run_id='run-now'`); n != 5 {
		t.Errorf("run-now rows = %d, want 5", n)
	}

	// The re-lease carry-over read ran BEFORE the abandon: run-now's c9 row
	// carries the dead run's stage and evidence (the abandon changes the row
	// that read looks at).
	db := fcOpen(t, root)
	c9, err := db.LoadCard(context.Background(), "run-now", "t1")
	if err != nil {
		t.Fatal(err)
	}
	if c9.Stage != homestate.CardRun || c9.EvidenceSHA != "sha-c9" {
		t.Errorf("run-now c9 stage/evidence = %q/%q, want the carried run/sha-c9", c9.Stage, c9.EvidenceSHA)
	}

	rows, err := db.DB.Query(`SELECT run_id, payload_json FROM events WHERE kind='card.rebound' ORDER BY seq`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	got := map[string]map[string]any{}
	for rows.Next() {
		var run, payload string
		if err := rows.Scan(&run, &payload); err != nil {
			t.Fatal(err)
		}
		var p map[string]any
		if err := json.Unmarshal([]byte(payload), &p); err != nil {
			t.Fatal(err)
		}
		got[run] = p
	}
	if len(got) != 2 {
		t.Fatalf("card.rebound events by run = %v, want exactly run-dead and run-ret", got)
	}
	for run, card := range map[string]string{"run-dead": "t1", "run-ret": "t2"} {
		p := got[run]
		if p == nil || p["card_id"] != card || p["from_run"] != run || p["to_run"] != "run-now" || p["state"] == nil || p["version"] == nil {
			t.Errorf("card.rebound for %s = %v, want %s from %s to run-now with state and version", run, p, card, run)
		}
	}
	if got["run-dead"]["evidence_sha"] != "sha-c9" || got["run-dead"]["stage"] != homestate.CardRun {
		t.Errorf("run-dead card.rebound = %v, want the prior stage run and evidence sha-c9", got["run-dead"])
	}
}

// One transaction: a failure between the T1 record and the abandon fails the
// lease and leaves the factory exactly as it was — no run-now row, the dead
// run's row untouched, no pointer event.
func TestFactoryNextAbandonFailureRollsBackTheRecord(t *testing.T) {
	root := deadRunFixture(t)
	cardsBefore := deadRunCount(t, root, `SELECT count(*) FROM cards`)
	eventsBefore := deadRunCount(t, root, `SELECT count(*) FROM events`)
	boom := errors.New("injected failure between the record and the abandon")
	factoryNextAbandonFault = func() error { return boom }
	t.Cleanup(func() { factoryNextAbandonFault = nil })

	sdLaneEnv(t, "lane-1", "")
	out, stderr, err := runFactory(t, "next", "--run", "run-now", "--card", "t1")
	if err == nil {
		t.Fatalf("next succeeded despite the injected failure\nstdout: %s\nstderr: %s", out, stderr)
	}
	if !strings.Contains(err.Error()+stderr, "injected failure") {
		t.Errorf("the failure was not surfaced: err=%v stderr=%s", err, stderr)
	}
	if got := deadRunCardState(t, root, "run-dead", "t1"); got != homestate.CardRun {
		t.Errorf("run-dead c9 = %s, want run (the abandon never committed)", got)
	}
	if n := deadRunCount(t, root, `SELECT count(*) FROM cards`); n != cardsBefore {
		t.Errorf("cards rows = %d, want %d (the T1 record rolled back)", n, cardsBefore)
	}
	if n := deadRunCount(t, root, `SELECT count(*) FROM events`); n != eventsBefore {
		t.Errorf("events rows = %d, want %d", n, eventsBefore)
	}
}

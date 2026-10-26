package cli

// factory_run_consolidate_test.go — REQ-FCR-017/018 (card t1963): a leader
// start adopts the newest dead-owner run and consolidates the other dead-stamp
// runs into it, with no --resume argument. AC-FCR-017 (one leader start leaves
// one active run), AC-FCR-018 (the guards and the source filter), AC-FCR-019 (a)
// (the loser of the adoption race mints a fresh id and the launch goes on).
//
// Every fixture is an isolated sandbox (adoptSandbox): the live factory
// database is never opened. Dead owners are stamps no process can ever carry
// (pid 99999999), so the production probe classifies them dead.

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factorymsg"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// lcRun records one run with the given owner stamp and creation time.
func lcRun(t *testing.T, db *homestate.FactoryDB, runID, backend string, pid int, start, createdAt string) {
	t.Helper()
	if err := db.RecordRun(context.Background(), homestate.FactoryRun{RunID: runID, Backend: backend, ManifestJSON: "{}", LeadPID: pid, LeadProcessStart: start}); err != nil {
		t.Fatalf("record run %s: %v", runID, err)
	}
	if createdAt != "" {
		if _, err := db.DB.Exec(`UPDATE runs SET created_at=? WHERE run_id=?`, createdAt, runID); err != nil {
			t.Fatalf("set created_at of %s: %v", runID, err)
		}
	}
}

func lcDeadRun(t *testing.T, db *homestate.FactoryDB, runID, createdAt string) {
	t.Helper()
	lcRun(t, db, runID, "", 99999999, "gone-before-resume", createdAt)
}

func lcCard(t *testing.T, db *homestate.FactoryDB, runID, cardID, state string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,stage,updated_at) VALUES(?,?,?,3,?,?)`,
		runID, cardID, state, "run", "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("insert card %s/%s: %v", runID, cardID, err)
	}
}

func lcBind(t *testing.T, db *homestate.FactoryDB, cardID, runID string) {
	t.Helper()
	if _, err := db.DB.Exec(`INSERT INTO card_dispatch(card_id,run_id,recorded_at) VALUES(?,?,?)`, cardID, runID, "2026-10-14T00:00:00Z"); err != nil {
		t.Fatalf("bind %s: %v", cardID, err)
	}
}

// lcDump renders every table a launch-time consolidation may touch, row by
// row, so "unchanged" is a byte comparison and a stray event shows up.
func lcDump(t *testing.T, root string) string {
	t.Helper()
	db := openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	var b strings.Builder
	for _, table := range []string{"cards", "card_dispatch", "runs", "events", "leader_approvals", "workers"} {
		b.WriteString("## " + table + "\n")
		rows, err := db.DB.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`) // SQL: fixed table list
		if err != nil {
			t.Fatalf("dump %s: %v", table, err)
		}
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]any, len(cols))
			for i := range vals {
				vals[i] = new(sql.NullString)
			}
			if err := rows.Scan(vals...); err != nil {
				t.Fatalf("dump %s scan: %v", table, err)
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

func lcQuery(t *testing.T, root, query string, args ...any) string {
	t.Helper()
	db := openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	var out string
	if err := db.DB.QueryRow(query, args...).Scan(&out); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return out
}

func lcInt(t *testing.T, root, query string, args ...any) int {
	t.Helper()
	db := openAdoptFactory(t, root)
	defer func() { _ = db.Close() }()
	var n int
	if err := db.DB.QueryRow(query, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", query, err)
	}
	return n
}

// lcAdopt runs the leader-entry adoption the way the cc and glm leaders do:
// a freshly minted id, no --factory-run, and the given launch arguments.
func lcAdopt(t *testing.T, root string, args []string, explicitRun string) (string, string, error) {
	t.Helper()
	t.Setenv(config.EnvFactoryRunID, "minted01")
	var warn bytes.Buffer
	adopted, err := leaderEntryRunAdoption(root, args, "leader", "", explicitRun, &warn)
	return adopted, warn.String(), err
}

// AC-FCR-017 — one leader start over two dead-owner runs leaves one active run:
// the newest is adopted, the older one's non-terminal cards move under it (or
// are abandoned with a pointer when the adopted run holds the card), every
// binding follows, the older run retires, and nothing is deleted.
func TestLeaderStartConsolidatesDeadOwnerRuns(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	lcDeadRun(t, db, "run-old", "2026-10-01T00:00:00Z")
	lcDeadRun(t, db, "run-new", "2026-10-02T00:00:00Z")
	for id, state := range map[string]string{"c1": "run", "c2": "picked", "c4": "needs-decision", "c5": "done"} {
		lcCard(t, db, "run-old", id, state)
	}
	lcCard(t, db, "run-new", "c3", "assigned")
	lcCard(t, db, "run-new", "c4", "picked")
	lcBind(t, db, "c1", "run-old")
	lcBind(t, db, "c2", "run-old")
	lcBind(t, db, "c4", "run-old") // the abandoned card's binding follows too (debt R10)
	if _, err := db.DB.Exec(`INSERT INTO leader_approvals(card_uuid,run_id,card_id,factory_version,evidence_hash,issuer,issuer_role,issued_at) VALUES('uuid-c1','run-old','c1',3,'hash','leader','leader','2026-10-14T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	approvalsBefore := lcQuery(t, root, `SELECT group_concat(card_uuid||'|'||run_id||'|'||card_id||'|'||factory_version||'|'||evidence_hash||'|'||issued_at) FROM leader_approvals`)
	cardsBefore := lcInt(t, root, `SELECT count(*) FROM cards`)
	eventsBefore := lcInt(t, root, `SELECT count(*) FROM events`)
	c1Before := lcQuery(t, root, `SELECT state||'|'||version||'|'||stage||'|'||updated_at FROM cards WHERE run_id='run-old' AND card_id='c1'`)
	c2Before := lcQuery(t, root, `SELECT state||'|'||version||'|'||stage||'|'||updated_at FROM cards WHERE run_id='run-old' AND card_id='c2'`)
	c4TargetBefore := lcQuery(t, root, `SELECT state||'|'||version||'|'||stage||'|'||updated_at FROM cards WHERE run_id='run-new' AND card_id='c4'`)

	// No resume argument: the adoption no longer needs one (decision Q2).
	adopted, warn, err := lcAdopt(t, root, nil, "")
	if err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if adopted != "run-new" {
		t.Fatalf("adopted = %q, want run-new (the newest dead-owner run)", adopted)
	}

	// (1) one active run, restamped to this process, with run.resumed.
	if got := lcQuery(t, root, `SELECT group_concat(run_id) FROM runs WHERE status='active'`); got != "run-new" {
		t.Errorf("active runs = %q, want exactly run-new", got)
	}
	if got := lcQuery(t, root, `SELECT lead_pid FROM runs WHERE run_id='run-new'`); got != strconv.Itoa(os.Getpid()) {
		t.Errorf("run-new owner pid = %s, want this process", got)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM events WHERE run_id='run-new' AND kind='run.resumed'`); n != 1 {
		t.Errorf("run.resumed events = %d, want 1", n)
	}
	// (2) run-old retired; one run.retired event with the full payload.
	if got := lcQuery(t, root, `SELECT status FROM runs WHERE run_id='run-old'`); got != "retired" {
		t.Errorf("run-old status = %s, want retired", got)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM events WHERE run_id='run-old' AND kind='run.retired'`); n != 1 {
		t.Fatalf("run.retired events = %d, want 1", n)
	}
	var rp struct {
		Classification string   `json:"classification"`
		Basis          string   `json:"basis"`
		ReboundTo      string   `json:"rebound_to"`
		Moved          int      `json:"moved"`
		Abandoned      int      `json:"abandoned"`
		MovedCards     []string `json:"moved_cards"`
		AbandonedCards []string `json:"abandoned_cards"`
	}
	raw := lcQuery(t, root, `SELECT payload_json FROM events WHERE run_id='run-old' AND kind='run.retired'`)
	if err := json.Unmarshal([]byte(raw), &rp); err != nil {
		t.Fatalf("run.retired payload %q: %v", raw, err)
	}
	if rp.Classification != "dead" || rp.Basis != "stamp" || rp.ReboundTo != "run-new" || rp.Moved != 2 || rp.Abandoned != 1 ||
		!slices.Equal(rp.MovedCards, []string{"c1", "c2"}) || !slices.Equal(rp.AbandonedCards, []string{"c4"}) {
		t.Errorf("run.retired payload = %s", raw)
	}
	for _, id := range []string{"c1", "c2", "c4"} {
		if !strings.Contains(warn, id) {
			t.Errorf("stderr %q does not name card %s", warn, id)
		}
	}
	// (3) c1 and c2 re-keyed with every other column intact.
	for id, before := range map[string]string{"c1": c1Before, "c2": c2Before} {
		if got := lcQuery(t, root, `SELECT state||'|'||version||'|'||stage||'|'||updated_at FROM cards WHERE run_id='run-new' AND card_id=?`, id); got != before {
			t.Errorf("%s under run-new = %s, want %s", id, got, before)
		}
		if n := lcInt(t, root, `SELECT count(*) FROM cards WHERE run_id='run-old' AND card_id=?`, id); n != 0 {
			t.Errorf("%s still under run-old", id)
		}
	}
	// (4) run-old's c4 abandoned with the pointer event; run-new's c4 untouched.
	if got := lcQuery(t, root, `SELECT state FROM cards WHERE run_id='run-old' AND card_id='c4'`); got != "abandoned" {
		t.Errorf("run-old c4 = %s, want abandoned", got)
	}
	var ep map[string]any
	if err := json.Unmarshal([]byte(lcQuery(t, root, `SELECT payload_json FROM events WHERE kind='card.rebound'`)), &ep); err != nil {
		t.Fatal(err)
	}
	if ep["state"] != "needs-decision" || ep["version"] != float64(3) || ep["from_run"] != "run-old" || ep["to_run"] != "run-new" {
		t.Errorf("card.rebound payload = %v", ep)
	}
	if got := lcQuery(t, root, `SELECT state||'|'||version||'|'||stage||'|'||updated_at FROM cards WHERE run_id='run-new' AND card_id='c4'`); got != c4TargetBefore {
		t.Errorf("run-new c4 = %s, want unchanged %s", got, c4TargetBefore)
	}
	// (5) no delete, no copy; events never shrink.
	if n := lcInt(t, root, `SELECT count(*) FROM cards`); n != cardsBefore {
		t.Errorf("cards rows = %d, want exactly %d", n, cardsBefore)
	}
	if n := lcInt(t, root, `SELECT count(*) FROM events`); n < eventsBefore {
		t.Errorf("events rows shrank: %d < %d", n, eventsBefore)
	}
	// (6) every binding that named run-old now names run-new.
	for _, id := range []string{"c1", "c2", "c4"} {
		if got := lcQuery(t, root, `SELECT run_id FROM card_dispatch WHERE card_id=?`, id); got != "run-new" {
			t.Errorf("card_dispatch[%s] = %s, want run-new", id, got)
		}
	}
	// (7) the run resolves: one active run, no AMBIGUOUS_FACTORY.
	resolved, rerr := factorymsg.ResolveActiveRun(context.Background(), root, "")
	if rerr != nil || resolved != "run-new" {
		t.Errorf("ResolveActiveRun = (%q, %v), want (run-new, nil)", resolved, rerr)
	}
	// (8) the approval row stays under run-old, byte for byte.
	if got := lcQuery(t, root, `SELECT group_concat(card_uuid||'|'||run_id||'|'||card_id||'|'||factory_version||'|'||evidence_hash||'|'||issued_at) FROM leader_approvals`); got != approvalsBefore {
		t.Errorf("leader_approvals changed:\n got %s\nwant %s", got, approvalsBefore)
	}
}

// REQ-FCR-023 / AC-FCR-017 (8) and the part of (9) the completion gate really
// enforces: a consolidated card keeps its old approval under the OLD run, and
// that approval does not close it under the new run (the cross-run refusal);
// only a fresh approval under the target run does. Then (9)'s second clause —
// a refusal driven by the queue's current-dispatch record — is NOT asserted
// here: the gate reads the dispatch binding and the receipt table and never the
// queue record (see the blocker report in the run evidence).
func TestLeaderStartRekeyedCardNeedsFreshApproval(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	lcDeadRun(t, db, "run-old", "2026-10-01T00:00:00Z")
	lcDeadRun(t, db, "run-new", "2026-10-02T00:00:00Z")
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,stage,evidence_sha,owner_label,updated_at) VALUES('run-old','c1','run',3,'run','sha-c1','lane-1','2026-10-14T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`INSERT INTO cards(run_id,card_id,state,version,updated_at) VALUES('run-new','c3','assigned',1,'2026-10-14T00:00:00Z')`); err != nil {
		t.Fatal(err)
	}
	lcBind(t, db, "c1", "run-old")
	issue := func(run string) {
		t.Helper()
		if _, err := db.IssueLeaderApproval(context.Background(), homestate.LeaderApproval{
			CardUUID: "uuid-c1", RunID: run, CardID: "c1", FactoryVersion: 3, EvidenceHash: "sha-c1",
			Issuer: "leader", IssuerRole: homestate.ApprovalIssuerLeader,
		}); err != nil {
			t.Fatal(err)
		}
	}
	issue("run-old")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if adopted, _, err := lcAdopt(t, root, nil, ""); err != nil || adopted != "run-new" {
		t.Fatalf("adopt = (%q, %v), want run-new", adopted, err)
	}

	closeGate := func() error {
		gate, err := holdDoneApprovalGate(context.Background(), root)
		if err != nil {
			t.Fatal(err)
		}
		defer gate.release()
		return gate.verifyForClose(context.Background(), "c1", "uuid-c1")
	}
	err := closeGate()
	if !errors.Is(err, homestate.ErrApprovalRunMismatch) {
		t.Fatalf("close with only the old run's approval: err = %v, want ErrApprovalRunMismatch", err)
	}
	db = openAdoptFactory(t, root)
	issue("run-new")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if err := closeGate(); err != nil {
		t.Fatalf("close with a fresh approval under the target run: %v", err)
	}
	// Both approvals now exist side by side; the old one was never rewritten.
	if n := lcInt(t, root, `SELECT count(*) FROM leader_approvals WHERE run_id='run-old'`); n != 1 {
		t.Fatalf("run-old approval rows = %d, want the original row untouched", n)
	}
}

// The tie rule (AC-FCR-017 variant e): two dead-stamp runs created at the same
// instant — the larger run_id is adopted.
func TestLeaderStartTieAdoptsLargerRunID(t *testing.T) {
	root := adoptSandbox(t)
	db := openAdoptFactory(t, root)
	lcDeadRun(t, db, "run-a", "2026-10-02T00:00:00Z")
	lcDeadRun(t, db, "run-b", "2026-10-02T00:00:00Z")
	lcCard(t, db, "run-a", "c1", "run")
	lcCard(t, db, "run-b", "c2", "run")
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	adopted, _, err := lcAdopt(t, root, nil, "")
	if err != nil || adopted != "run-b" {
		t.Fatalf("adopted = (%q, %v), want run-b (run_id descending on a tie)", adopted, err)
	}
}

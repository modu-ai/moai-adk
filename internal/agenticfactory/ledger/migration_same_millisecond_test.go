package ledger

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/agenticfactory/core"
)

// issue-1948: the schema-2 migration gives each transition of a version 1 ledger the id of the newest evidence the card had
// "at or before" the move, counting evidence written in the very same millisecond as before.

// v1Millisecond writes a version 1 ledger for card "t1" in which c1 was offered and landed at 10:00:03.000.
// extra is called with a way to add an evidence row at a given time; it runs after c1's pair.
func v1Millisecond(t *testing.T, back bool, extra func(ev func(kind, sha, at string))) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "af", "ledger.db")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	exec(ledgerV1DDL)
	spec, _ := json.Marshal(Spec{Class: "A", Paths: []string{"internal/x/x.go"}})
	const day = "2026-10-20T10:00:"
	stage, sub := "MERGING", ""
	if back {
		stage, sub = "BUILDING", "IMPLEMENT"
	}
	exec(`INSERT INTO cards(id,title,stage,sub_stage,priority,spec_json,epoch,created_at,updated_at) VALUES('t1','t1',?,?,1,?,1,?,?)`,
		stage, nullIfEmpty(sub), string(spec), day+"00.000Z", day+"10.000Z")
	exec(`INSERT INTO leases(card_id,session_id,epoch,worktree,heartbeat_at,expires_at) VALUES('t1','s-t1',1,'/wt',?,'2999-01-01T00:00:00.000Z')`, day+"11.000Z")
	ev := func(kind, sha, at string) {
		exec(`INSERT INTO evidence(card_id,kind,tree_sha,subject_sha,result,epoch,at) VALUES('t1',?,?,?,'pass',1,?)`, kind, "tree-"+sha, sha, at)
	}
	tr := func(fromS, fromSub, toS, toSub, actor, reason, at string) {
		exec(`INSERT INTO transitions(card_id,from_stage,from_sub,to_stage,to_sub,epoch,actor,reason,at) VALUES('t1',?,?,?,?,1,?,?,?)`,
			nullIfEmpty(fromS), nullIfEmpty(fromSub), toS, nullIfEmpty(toSub), actor, reason, at)
	}
	ev(KindGreen, "c1", day+"01.000Z")
	ev(KindGate, "c1", day+"02.000Z")
	tr("", "", "READY", "", "add", "t", day+"00.000Z")
	tr("READY", "", "BUILDING", "SPEC", "claim", "t", day+"00.500Z")
	tr("BUILDING", "SPEC", "BUILDING", "IMPLEMENT", "claim", "t", day+"00.600Z")
	tr("BUILDING", "IMPLEMENT", "REVIEWING", "", "lane", "t", day+"02.500Z")
	tr("REVIEWING", "", "MERGING", "", "enqueue", "enqueued c1", day+"03.000Z")
	exec(`INSERT INTO merge_queue(card_id,candidate_sha,state,card_epoch,enqueued_at,updated_at) VALUES('t1','c1','landed',1,?,?)`, day+"03.000Z", day+"05.000Z")
	if back {
		tr("MERGING", "", "REVIEWING", "", "lane", "t", day+"09.000Z")
		tr("REVIEWING", "", "BUILDING", "IMPLEMENT", "lane", "t", day+"10.000Z")
	}
	if extra != nil {
		extra(ev)
	}
	return path
}

// Evidence in the very same millisecond as a move back into BUILDING ranks before it, the side that withholds a DONE, so
// the old receipt is not mistaken for rework.
func TestOpen_ASameMillisecondMeasurementOfTheMoveBackStillRanksBeforeIt(t *testing.T) {
	path := v1Millisecond(t, true, func(ev func(kind, sha, at string)) {
		ev(KindGreen, "c1", "2026-10-20T10:00:10.000Z")
		ev(KindGate, "c1", "2026-10-20T10:00:10.000Z")
	})
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	var mark int64
	if err := s.db.QueryRowContext(bg, `SELECT evidence_mark FROM transitions WHERE to_stage='BUILDING' AND from_stage='REVIEWING'`).Scan(&mark); err != nil || mark != 4 {
		t.Fatalf("mark of the move back: %d %v (want 4: the pair in its millisecond counts before it)", mark, err)
	}
	if err := s.Move(bg, "t1", 1, posReviewing, "lane", "handoff"); err != nil {
		t.Fatal(err)
	}
	if err := s.Move(bg, "t1", 1, core.Position{Stage: core.StageDone}, "lane", "done"); !errors.Is(err, core.ErrGuardMissing) {
		t.Fatalf("a measurement in the millisecond of the move back counted as rework: %v", err)
	}
}

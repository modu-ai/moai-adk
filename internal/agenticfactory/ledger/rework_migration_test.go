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

// ledgerV1DDL is the schema of a ledger written before evidence_mark existed
// (PRAGMA user_version 1): the tables the migration touches, as they were.
const ledgerV1DDL = `
CREATE TABLE cards(
  id TEXT PRIMARY KEY, title TEXT NOT NULL, stage TEXT NOT NULL, sub_stage TEXT,
  parked INTEGER NOT NULL DEFAULT 0, park_reason TEXT, park_evidence TEXT, park_resume TEXT,
  priority INTEGER NOT NULL DEFAULT 0, spec_json TEXT NOT NULL, epoch INTEGER NOT NULL DEFAULT 0,
  integrated INTEGER NOT NULL DEFAULT 0,
  reclaim_count INTEGER NOT NULL DEFAULT 0, created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE transitions(
  seq INTEGER PRIMARY KEY AUTOINCREMENT, card_id TEXT NOT NULL, from_stage TEXT, from_sub TEXT,
  to_stage TEXT NOT NULL, to_sub TEXT, epoch INTEGER NOT NULL, actor TEXT NOT NULL, reason TEXT, at TEXT NOT NULL);
CREATE TABLE leases(
  card_id TEXT PRIMARY KEY, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
  worktree TEXT NOT NULL, heartbeat_at TEXT NOT NULL, next_wake_at TEXT, expires_at TEXT NOT NULL);
CREATE TABLE lease_keys(key TEXT PRIMARY KEY, card_id TEXT NOT NULL);
CREATE TABLE evidence(
  id INTEGER PRIMARY KEY AUTOINCREMENT, card_id TEXT NOT NULL, kind TEXT NOT NULL, tree_sha TEXT NOT NULL,
  subject_sha TEXT, base_sha TEXT, command TEXT, result TEXT NOT NULL, path TEXT, epoch INTEGER NOT NULL, at TEXT NOT NULL);
CREATE TABLE merge_queue(
  card_id TEXT PRIMARY KEY, candidate_sha TEXT NOT NULL, trunk_before TEXT,
  state TEXT NOT NULL CHECK(state IN ('waiting','verifying','landing','landed','failed')),
  fail_reason TEXT, card_epoch INTEGER NOT NULL DEFAULT 0, enqueued_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE locks(
  name TEXT PRIMARY KEY, session_id TEXT NOT NULL, epoch INTEGER NOT NULL, pid INTEGER, pid_start TEXT,
  expires_at TEXT NOT NULL);
CREATE TABLE notes(
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  kind TEXT NOT NULL CHECK(kind IN ('landed','decision','gate_wait','stalled')),
  card_id TEXT, text TEXT NOT NULL, at TEXT NOT NULL);
PRAGMA user_version=1;
`

// writeV1Ledger builds a ledger file as the code before evidence_mark wrote it. Cards "a" and
// "b" both landed c1, left MERGING, and were taken back to BUILDING/IMPLEMENT at 10:00:10. "b"
// measured c1 again at 10:00:20; "a" recorded nothing since.
func writeV1Ledger(t *testing.T) string {
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
	if _, err := db.Exec(ledgerV1DDL); err != nil {
		t.Fatal(err)
	}
	spec, _ := json.Marshal(Spec{Class: "A", Paths: []string{"internal/x/x.go"}})
	exec := func(q string, args ...any) {
		t.Helper()
		if _, err := db.Exec(q, args...); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}
	day := "2026-10-20T10:00:"
	for _, id := range []string{"a", "b"} {
		exec(`INSERT INTO cards(id,title,stage,sub_stage,priority,spec_json,epoch,created_at,updated_at) VALUES(?,?,?,?,1,?,1,?,?)`,
			id, id, "BUILDING", "IMPLEMENT", string(spec), day+"00.000Z", day+"10.000Z")
		exec(`INSERT INTO leases(card_id,session_id,epoch,worktree,heartbeat_at,expires_at) VALUES(?,?,1,'/wt',?,'2999-01-01T00:00:00.000Z')`, id, "s-"+id, day+"11.000Z")
	}
	// Evidence is inserted card by card, so the ids are not the same order as the times.
	ev := func(id, kind, sha, at string) {
		exec(`INSERT INTO evidence(card_id,kind,tree_sha,subject_sha,result,epoch,at) VALUES(?,?,?,?,'pass',1,?)`, id, kind, "tree-"+sha, sha, at)
	}
	tr := func(id, fromS, fromSub, toS, toSub, actor, at string) {
		exec(`INSERT INTO transitions(card_id,from_stage,from_sub,to_stage,to_sub,epoch,actor,reason,at) VALUES(?,?,?,?,?,1,?,'t',?)`,
			id, nullIfEmpty(fromS), nullIfEmpty(fromSub), toS, nullIfEmpty(toSub), actor, at)
	}
	for _, id := range []string{"a", "b"} {
		ev(id, KindGreen, "c1", day+"01.000Z")
		ev(id, KindGate, "c1", day+"02.000Z")
		tr(id, "", "", "READY", "", "add", day+"00.000Z")
		tr(id, "READY", "", "BUILDING", "SPEC", "claim", day+"00.500Z")
		tr(id, "BUILDING", "SPEC", "BUILDING", "IMPLEMENT", "claim", day+"00.600Z")
		tr(id, "BUILDING", "IMPLEMENT", "REVIEWING", "", "lane", day+"02.500Z")
		tr(id, "REVIEWING", "", "MERGING", "", "enqueue", day+"03.000Z")
		tr(id, "MERGING", "", "REVIEWING", "", "lane", day+"09.000Z")
		tr(id, "REVIEWING", "", "BUILDING", "IMPLEMENT", "lane", day+"10.000Z")
		exec(`INSERT INTO merge_queue(card_id,candidate_sha,state,card_epoch,enqueued_at,updated_at) VALUES(?,'c1','landed',1,?,?)`, id, day+"03.000Z", day+"05.000Z")
	}
	ev("b", KindGreen, "c1", day+"20.000Z")
	ev("b", KindGate, "c1", day+"20.500Z")
	return path
}

func TestOpen_MigratesALedgerWrittenBeforeEvidenceMarks(t *testing.T) {
	path := writeV1Ledger(t)
	for pass := 1; pass <= 2; pass++ { // the second open finds a migrated file and changes nothing
		s, err := Open(bg, path)
		if err != nil {
			t.Fatalf("open %d: %v", pass, err)
		}
		var ver, nulls int
		if err := s.db.QueryRowContext(bg, `PRAGMA user_version`).Scan(&ver); err != nil || ver != SchemaVersion || SchemaVersion != 2 {
			t.Fatalf("user_version %d (code %d) %v", ver, SchemaVersion, err)
		}
		if err := s.db.QueryRowContext(bg, `SELECT count(*) FROM transitions WHERE evidence_mark IS NULL`).Scan(&nulls); err != nil || nulls != 0 {
			t.Fatalf("transitions without a mark: %d %v", nulls, err)
		}
		// The mark of the move back is the last evidence the card had before it: ids 1,2 are a's;
		// 3,4 are b's first pair, 5,6 b's pair after the move back.
		marks := map[string]int64{}
		rows, err := s.db.QueryContext(bg, `SELECT card_id,evidence_mark FROM transitions WHERE to_stage='BUILDING' AND from_stage='REVIEWING'`)
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var id string
			var m int64
			if err := rows.Scan(&id, &m); err != nil {
				t.Fatal(err)
			}
			marks[id] = m
		}
		_ = rows.Close()
		if marks["a"] != 2 || marks["b"] != 4 {
			t.Fatalf("marks of the move back: %v (want a=2, b=4)", marks)
		}
		_ = s.Close()
	}
	s, err := Open(bg, path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = s.Close() }()
	// The verdict is the same as on a ledger that was never old.
	for _, id := range []string{"a", "b"} {
		if err := s.Move(bg, id, 1, posReviewing, "lane", "handoff"); err != nil {
			t.Fatalf("%s: back to REVIEWING: %v", id, err)
		}
	}
	err = s.Move(bg, "a", 1, core.Position{Stage: core.StageDone}, "lane", "done")
	if !errors.Is(err, core.ErrGuardMissing) {
		t.Errorf("a (nothing recorded since the move back): %v", err)
	}
	if err := s.Move(bg, "b", 1, core.Position{Stage: core.StageDone}, "lane", "done"); err != nil {
		t.Errorf("b (c1 measured again since the move back): %v", err)
	}
}

// A file this code did not write at all is still refused, as before.
func TestOpen_RefusesALedgerNewerThanTheCode(t *testing.T) {
	s, path := openAt(t)
	if _, err := s.db.ExecContext(bg, `PRAGMA user_version=99`); err != nil {
		t.Fatal(err)
	}
	_ = s.Close()
	if _, err := Open(bg, path); err == nil {
		t.Fatal("a ledger of schema 99 opened")
	}
}

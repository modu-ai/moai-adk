package factorymsg

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// legacySchemaV1 is the broker schema as shipped before the idempotency scope
// moved to the sender lane slot: messages are unique per
// (sender_session, idem_key). It is copied verbatim from the store's schema
// constant at the commit that introduced this characterization, so an
// existing on-disk broker can be rebuilt byte-for-byte in a test.
const legacySchemaV1 = `
CREATE TABLE IF NOT EXISTS peers(slot TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, backend TEXT NOT NULL, role TEXT NOT NULL, session_uuid TEXT NOT NULL UNIQUE, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS messages(id TEXT PRIMARY KEY, schema_version INTEGER NOT NULL, project_key TEXT NOT NULL, run_id TEXT NOT NULL, sender_session TEXT NOT NULL, sender_generation INTEGER NOT NULL, recipient_session TEXT NOT NULL, recipient_generation INTEGER NOT NULL, kind TEXT NOT NULL, idem_key TEXT NOT NULL, task_ref TEXT NOT NULL, correlation_id TEXT NOT NULL, created_at TEXT NOT NULL, expires_at TEXT NOT NULL, payload BLOB NOT NULL, state TEXT NOT NULL, claim_token TEXT NOT NULL DEFAULT '', claim_expires_at TEXT, disposition TEXT NOT NULL DEFAULT '', acknowledged_at TEXT, UNIQUE(sender_session,idem_key));
CREATE INDEX IF NOT EXISTS messages_recipient_state ON messages(recipient_session,recipient_generation,state,created_at);
CREATE TABLE IF NOT EXISTS dead_letters(id INTEGER PRIMARY KEY AUTOINCREMENT, message_id TEXT NOT NULL, reason TEXT NOT NULL, created_at TEXT NOT NULL);
`

// legacyColumns are the message columns every schema version carries.
const legacyColumns = `id,schema_version,project_key,run_id,sender_session,sender_generation,recipient_session,recipient_generation,kind,idem_key,task_ref,correlation_id,created_at,expires_at,payload,state,claim_token,COALESCE(claim_expires_at,''),disposition,COALESCE(acknowledged_at,'')`

type legacyFixture struct {
	root, path, projectKey string
	lead, worker           Peer
	ids                    []string
}

// buildLegacyBroker writes a broker database with legacySchemaV1 holding the
// state the AC-DHR-020 measurement reproduced: lane-1 restarted (session w-s1
// generation 1 -> w-s2 generation 2) and both sessions sent the same key, so
// two rows share it. It also holds a lead row and a row from a sender that no
// longer has any peers entry.
func buildLegacyBroker(t *testing.T) legacyFixture {
	t.Helper()
	t.Setenv("MOAI_HOME", t.TempDir())
	root := t.TempDir()
	path, err := BrokerPath(root, "run")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(legacySchemaV1); err != nil {
		t.Fatal(err)
	}
	pk := projectKeyFromBrokerPath(path)
	now := time.Now().UTC()
	ts := now.Format(time.RFC3339Nano)
	exp := now.Add(time.Hour).Format(time.RFC3339Nano)
	pid := os.Getpid()
	f := legacyFixture{root: root, path: path, projectKey: pk,
		lead:   Peer{ProjectKey: pk, RunID: "run", Backend: "codex", Role: "lead", Slot: "lead", SessionUUID: "lead-s1", Generation: 1, PID: pid, ProcessStart: "start-lead"},
		worker: Peer{ProjectKey: pk, RunID: "run", Backend: "codex", Role: "worker", Slot: "lane-1", SessionUUID: "w-s2", Generation: 2, PID: pid, ProcessStart: "start-w2"},
	}
	for _, p := range []Peer{f.lead, f.worker} {
		if _, err := db.Exec(`INSERT INTO peers(slot,project_key,run_id,backend,role,session_uuid,generation,pid,process_start,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?)`, p.Slot, pk, "run", p.Backend, p.Role, p.SessionUUID, p.Generation, p.PID, p.ProcessStart, ts); err != nil {
			t.Fatal(err)
		}
	}
	rows := []struct {
		id, sender   string
		senderGen    int64
		recipient    string
		recipientGen int64
		key, state   string
	}{
		{"legacy-old-session", "w-s1", 1, "lead-s1", 1, "result:d-1:1", "acknowledged"},
		{"legacy-restarted", "w-s2", 2, "lead-s1", 1, "result:d-1:1", "pending"},
		{"legacy-lead", "lead-s1", 1, "w-s2", 2, "dispatch:d-1:1", "pending"},
		{"legacy-ghost", "ghost-s1", 1, "lead-s1", 1, "result:d-1:1", "pending"},
	}
	for _, r := range rows {
		ack := any(nil)
		if r.state == "acknowledged" {
			ack = ts
		}
		if _, err := db.Exec(`INSERT INTO messages(id,schema_version,project_key,run_id,sender_session,sender_generation,recipient_session,recipient_generation,kind,idem_key,task_ref,correlation_id,created_at,expires_at,payload,state,disposition,acknowledged_at) VALUES(?,1,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			r.id, pk, "run", r.sender, r.senderGen, r.recipient, r.recipientGen, KindStatusReport, r.key, "t1100", "c-"+r.id, ts, exp, []byte("body "+r.id), r.state, map[bool]string{true: DispositionAccepted}[r.state == "acknowledged"], ack); err != nil {
			t.Fatalf("insert %s: %v", r.id, err)
		}
		f.ids = append(f.ids, r.id)
	}
	return f
}

func legacySnapshot(t *testing.T, db *sql.DB) map[string][]any {
	t.Helper()
	rows, err := db.Query(`SELECT ` + legacyColumns + ` FROM messages ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string][]any{}
	for rows.Next() {
		vals := make([]any, 20)
		ptrs := make([]any, len(vals))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatal(err)
		}
		out[vals[0].(string)] = vals
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestLegacyBrokerRowsSurviveOpen characterizes what opening an existing
// broker must preserve whatever the schema version: every message row keeps
// its ID and every column value, and pending messages stay deliverable.
func TestLegacyBrokerRowsSurviveOpen(t *testing.T) {
	f := buildLegacyBroker(t)
	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(f.path))
	if err != nil {
		t.Fatal(err)
	}
	before := legacySnapshot(t, raw)
	if err := raw.Close(); err != nil {
		t.Fatal(err)
	}
	if len(before) != len(f.ids) {
		t.Fatalf("fixture rows: %d", len(before))
	}

	s, err := Open(f.root, "run")
	if err != nil {
		t.Fatalf("open legacy broker: %v", err)
	}
	defer func() { _ = s.Close() }()
	s.ownerCurrent = func(int, string) bool { return false }
	if after := legacySnapshot(t, s.db); !reflect.DeepEqual(before, after) {
		t.Fatalf("opening a legacy broker changed message rows:\nbefore=%v\nafter =%v", before, after)
	}
	claims, err := s.Claim(context.Background(), f.lead, MaxBatch, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, c := range claims {
		got[c.ID] = true
	}
	if !got["legacy-restarted"] || !got["legacy-ghost"] || got["legacy-old-session"] {
		t.Fatalf("legacy pending rows must stay deliverable: %v", got)
	}
}

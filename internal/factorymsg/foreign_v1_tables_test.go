package factorymsg

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

// foreignV1HandoffDDL is the lane-handoff extension another lane adds to the
// broker schema while keeping schema version 1 (branch
// WT-factory-lane-worktree-handoff, internal/factorymsg/handoff.go
// handoffSchema, copied verbatim). A broker written by a binary carrying that
// extension holds these tables next to the version-1 messages table, so the
// lane-scope migration must leave them untouched.
const foreignV1HandoffDDL = `
CREATE TABLE IF NOT EXISTS lane_handoffs(id TEXT PRIMARY KEY, project_key TEXT NOT NULL, run_id TEXT NOT NULL, slot TEXT NOT NULL, card_id TEXT NOT NULL, spec_id TEXT NOT NULL, mode TEXT NOT NULL, nonce TEXT NOT NULL UNIQUE, handoff_generation INTEGER NOT NULL, source_backend TEXT NOT NULL, source_role TEXT NOT NULL, source_session TEXT NOT NULL, source_generation INTEGER NOT NULL, source_pid INTEGER NOT NULL, source_process_start TEXT NOT NULL, develop_pin TEXT NOT NULL, target_path TEXT NOT NULL, target_branch TEXT NOT NULL, state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL);
CREATE UNIQUE INDEX IF NOT EXISTS lane_handoffs_one_open ON lane_handoffs(slot) WHERE state IN ('RESERVED','WT_READY','SWITCH_PENDING_INTERACTIVE','SWITCH_PENDING_HEADLESS');
CREATE TABLE IF NOT EXISTS lane_handoff_events(id INTEGER PRIMARY KEY AUTOINCREMENT, handoff_id TEXT NOT NULL, from_state TEXT NOT NULL, to_state TEXT NOT NULL, reason TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_handoff_relocations(handoff_id TEXT PRIMARY KEY, nonce TEXT NOT NULL, method TEXT NOT NULL, source_thread_id TEXT NOT NULL, thread_id TEXT NOT NULL, forked_from_id TEXT NOT NULL, thread_started INTEGER NOT NULL, request_cwd TEXT NOT NULL, response_cwd TEXT NOT NULL, readback_cwd TEXT NOT NULL, readback_branch TEXT NOT NULL, readback_head TEXT NOT NULL, recorded_at TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS lane_endpoint_tombstones(slot TEXT NOT NULL, session_uuid TEXT NOT NULL, generation INTEGER NOT NULL, replaced_by_session TEXT NOT NULL, replaced_by_generation INTEGER NOT NULL, handoff_id TEXT NOT NULL, bound_at TEXT NOT NULL, PRIMARY KEY(slot,session_uuid,generation));
`

var foreignV1Tables = []string{"lane_handoffs", "lane_handoff_events", "lane_handoff_relocations", "lane_endpoint_tombstones"}

// foreignSnapshot captures every foreign table's rows and every foreign
// catalogue object (tables and the partial unique index) as text.
func foreignSnapshot(t *testing.T, db *sql.DB) map[string][]string {
	t.Helper()
	out := map[string][]string{}
	for _, table := range foreignV1Tables {
		rows, err := db.Query(`SELECT * FROM ` + table + ` ORDER BY rowid`)
		if err != nil {
			t.Fatalf("read %s: %v", table, err)
		}
		cols, err := rows.Columns()
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			vals := make([]any, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			out[table] = append(out[table], fmt.Sprintf("%v", vals))
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		_ = rows.Close()
	}
	// The catalogue is limited to the objects the foreign DDL defines: tables a
	// newer store's own schema adds on open (the handoff-bind tables) are not
	// foreign objects the migration could have altered.
	cat, err := db.Query(`SELECT type,name,COALESCE(sql,'') FROM sqlite_master WHERE name IN ('lane_handoffs','lane_handoffs_one_open','lane_handoff_events','lane_handoff_relocations','lane_endpoint_tombstones') ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cat.Close() }()
	for cat.Next() {
		var typ, name, ddl string
		if err := cat.Scan(&typ, &name, &ddl); err != nil {
			t.Fatal(err)
		}
		out["catalogue"] = append(out["catalogue"], typ+" "+name+" "+ddl)
	}
	return out
}

// TestIdemScopeMigrationPreservesForeignV1Tables characterizes the
// cross-lane ordering where a schema-version-1 broker already carries the
// lane-handoff tables (written by a binary with that extension but without
// the lane-scope migration): opening it with this store migrates messages to
// the lane-slot scope, keeps every message row, and leaves every foreign
// table, its rows, and its index exactly as they were.
func TestIdemScopeMigrationPreservesForeignV1Tables(t *testing.T) {
	for _, opener := range []struct {
		name string
		open func(root string) (*Store, error)
	}{
		{"open", func(root string) (*Store, error) { return Open(root, "run") }},
		{"open_existing", func(root string) (*Store, error) { return OpenExistingWithDeadline(root, "run", 5*time.Second) }},
	} {
		t.Run(opener.name, func(t *testing.T) {
			f := buildLegacyBroker(t)
			raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(f.path))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := raw.Exec(foreignV1HandoffDDL); err != nil {
				t.Fatalf("apply foreign v1 DDL: %v", err)
			}
			ts := time.Now().UTC().Format(time.RFC3339Nano)
			for _, stmt := range []struct {
				q    string
				args []any
			}{
				{`INSERT INTO lane_handoffs VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, []any{"h-1", f.projectKey, "run", "lane-1", "t9", "SPEC-X", "interactive", "nonce-1", 1, "codex", "worker", "w-s1", 1, 42, "start-w1", "abc123", "/tmp/wt", "WT-x", "WT_READY", "", ts, ts}},
				{`INSERT INTO lane_handoffs VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, []any{"h-0", f.projectKey, "run", "lane-2", "t8", "SPEC-Y", "headless", "nonce-0", 1, "claude", "worker", "w2-s1", 1, 43, "start-w21", "abc123", "/tmp/wt2", "WT-y", "BOUND", "done", ts, ts}},
				{`INSERT INTO lane_handoff_events(handoff_id,from_state,to_state,reason,created_at) VALUES(?,?,?,?,?)`, []any{"h-1", "RESERVED", "WT_READY", "", ts}},
				{`INSERT INTO lane_handoff_events(handoff_id,from_state,to_state,reason,created_at) VALUES(?,?,?,?,?)`, []any{"h-0", "WT_READY", "BOUND", "done", ts}},
				{`INSERT INTO lane_handoff_relocations VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)`, []any{"h-0", "nonce-0", "fork", "th-src", "th-new", "th-src", 1, "/tmp/wt2", "/tmp/wt2", "/tmp/wt2", "WT-y", "abc123", ts}},
				{`INSERT INTO lane_endpoint_tombstones VALUES(?,?,?,?,?,?,?)`, []any{"lane-1", "w-s1", 1, "w-s2", 2, "h-0", ts}},
			} {
				if _, err := raw.Exec(stmt.q, stmt.args...); err != nil {
					t.Fatalf("seed foreign row: %v", err)
				}
			}
			beforeMsgs := legacySnapshot(t, raw)
			beforeForeign := foreignSnapshot(t, raw)
			if err := raw.Close(); err != nil {
				t.Fatal(err)
			}
			for _, table := range foreignV1Tables {
				if len(beforeForeign[table]) == 0 {
					t.Fatalf("fixture seeded no rows in %s", table)
				}
			}
			if len(beforeForeign["catalogue"]) <= len(foreignV1Tables) {
				t.Fatalf("fixture catalogue lacks the foreign index: %v", beforeForeign["catalogue"])
			}

			s, err := opener.open(f.root)
			if err != nil {
				t.Fatalf("open v1 broker carrying foreign tables: %v", err)
			}
			defer func() { _ = s.Close() }()

			var laneScoped int
			if err := s.db.QueryRow(`SELECT count(*) FROM pragma_table_info('messages') WHERE name='sender_slot'`).Scan(&laneScoped); err != nil {
				t.Fatal(err)
			}
			if laneScoped != 1 {
				t.Fatalf("messages not migrated to the lane-slot scope (sender_slot columns=%d)", laneScoped)
			}
			if afterMsgs := legacySnapshot(t, s.db); !reflect.DeepEqual(beforeMsgs, afterMsgs) {
				t.Fatalf("message rows changed:\nbefore=%v\nafter =%v", beforeMsgs, afterMsgs)
			}
			if afterForeign := foreignSnapshot(t, s.db); !reflect.DeepEqual(beforeForeign, afterForeign) {
				t.Fatalf("foreign v1 tables changed:\nbefore=%v\nafter =%v", beforeForeign, afterForeign)
			}
		})
	}
}

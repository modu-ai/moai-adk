package homestate

import (
	"database/sql"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// factoryDDLv3 is a frozen copy of the schema-version-3 factory tables the F1
// migration starts from. It is a literal, not a reference to factoryDDL, so the
// fixture keeps describing the version-3 shape after the live DDL moves on.
const factoryDDLv3 = `
CREATE TABLE meta (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE workers (
  label TEXT PRIMARY KEY,
  pid INTEGER NOT NULL,
  backend TEXT NOT NULL DEFAULT '',
  session_id TEXT NOT NULL DEFAULT '',
  run_id TEXT NOT NULL DEFAULT '',
  registered_at TEXT NOT NULL,
  heartbeat_at TEXT NOT NULL
);
CREATE TABLE runs (
  run_id TEXT PRIMARY KEY,
  lead_session_id TEXT NOT NULL DEFAULT '',
  lead_backend TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  manifest_json TEXT NOT NULL DEFAULT '{}',
  lead_pid INTEGER NOT NULL DEFAULT 0,
  lead_process_start TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
CREATE TABLE cards (
  run_id TEXT NOT NULL,
  card_id TEXT NOT NULL,
  owner_label TEXT NOT NULL DEFAULT '',
  state TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  evidence_path TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL,
  PRIMARY KEY(run_id, card_id)
);
CREATE TABLE events (
  seq INTEGER PRIMARY KEY AUTOINCREMENT,
  run_id TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL,
  payload_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL
);
CREATE TABLE resume_handoffs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  status TEXT NOT NULL CHECK(status IN ('pending','claimed','consumed','failed','expired','cleared')),
  schema_version INTEGER NOT NULL,
  spec_id TEXT NOT NULL DEFAULT '',
  phase TEXT NOT NULL DEFAULT '',
  saved_at TEXT NOT NULL,
  saved_by_session TEXT NOT NULL DEFAULT '',
  conversation_language TEXT NOT NULL DEFAULT '',
  directives_json TEXT NOT NULL DEFAULT '{}',
  embedded_goal_json TEXT,
  body TEXT NOT NULL,
  body_sha256 TEXT NOT NULL,
  claim_token TEXT NOT NULL DEFAULT '',
  claimed_at TEXT,
  claim_expires_at TEXT,
  claim_owner_pid INTEGER,
  claim_owner_session TEXT NOT NULL DEFAULT '',
  claim_owner_fingerprint TEXT NOT NULL DEFAULT '',
  legacy_recovery INTEGER NOT NULL DEFAULT 0,
  legacy_recovery_reason TEXT NOT NULL DEFAULT '',
  consumed_at TEXT,
  error TEXT NOT NULL DEFAULT ''
);
INSERT INTO meta(key,value) VALUES('schema_version','3');
INSERT INTO workers(label,pid,backend,session_id,run_id,registered_at,heartbeat_at) VALUES('worker-1',101,'claude','s-1','run-1','2026-09-01T00:00:00Z','2026-09-01T00:00:00Z');
INSERT INTO runs(run_id,lead_session_id,lead_backend,status,manifest_json,lead_pid,lead_process_start,created_at,updated_at) VALUES('run-1','lead-s','claude','active','{"a":1}',77,'start-77','2026-09-01T00:00:00Z','2026-09-01T00:00:01Z');
INSERT INTO cards(run_id,card_id,owner_label,state,version,evidence_path,updated_at) VALUES('run-1','c-legacy','worker-1','in_progress',4,'.moai/reports/c-legacy/x.md','2026-09-01T00:00:02Z');
INSERT INTO events(run_id,kind,payload_json,created_at) VALUES('run-1','card.updated','{"k":"v"}','2026-09-01T00:00:03Z');
INSERT INTO resume_handoffs(status,schema_version,saved_at,body,body_sha256) VALUES('consumed',1,'2026-09-01T00:00:04Z','body','sha');
`

// frF1CardColumns is the column set design.md § Schema adds to `cards`.
var frF1CardColumns = []string{
	"stage", "lease_holder", "lease_expires_at", "heartbeat_at",
	"decision_gate", "decision_question", "decision_resume", "decider", "decided_at",
	"failure_reason", "hint_prefer", "hint_after", "spec_id", "worktree_path",
	"evidence_sha", "merge_sha", "merge_tree", "remeasure_path",
	"contract_spec_id", "contract_sha256", "contract_signed_at", "contract_event",
}

var frV3CardColumns = []string{"run_id", "card_id", "owner_label", "state", "version", "evidence_path", "updated_at"}

func frSeedV3(t *testing.T) string {
	t.Helper()
	root := factorySandbox(t)
	path, err := FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create factory dir: %v", err)
	}
	seed, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("open seed db: %v", err)
	}
	if _, err := seed.Exec(factoryDDLv3); err != nil {
		t.Fatalf("seed v3 database: %v", err)
	}
	if err := seed.Close(); err != nil {
		t.Fatalf("close seed db: %v", err)
	}
	return path
}

func frTableColumns(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	// SQL: table is a test-literal table name, never input.
	rows, err := db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		t.Fatalf("table_info(%s): %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var (
			cid        int
			name, typ  string
			notNull    int
			defaultVal sql.NullString
			pk         int
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultVal, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		out = append(out, name)
	}
	return out
}

// frDump renders every row of the named tables, columns in declaration order,
// so a byte-level before/after comparison is one string comparison.
func frDump(t *testing.T, db *sql.DB, table string, columns []string) []string {
	t.Helper()
	// SQL: table and columns are test-literal names, never input.
	rows, err := db.Query(`SELECT ` + strings.Join(columns, ",") + ` FROM ` + table + ` ORDER BY 1,2`)
	if err != nil {
		t.Fatalf("dump %s: %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		vals := make([]sql.NullString, len(columns))
		ptrs := make([]any, len(columns))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			t.Fatalf("scan %s: %v", table, err)
		}
		parts := make([]string, len(vals))
		for i, v := range vals {
			if v.Valid {
				parts[i] = v.String
			} else {
				parts[i] = "<NULL>"
			}
		}
		out = append(out, strings.Join(parts, "|"))
	}
	return out
}

func frSchemaVersion(t *testing.T, db *sql.DB) string {
	t.Helper()
	var version string
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	return version
}

// AC-001 — a schema-3 database migrates to schema 4 preserving every row, the
// migration is idempotent, a fresh database carries every F1 column, and a
// schema-5 database is refused.
func TestFR_AC001_MigrationV3ToV4(t *testing.T) {
	path := frSeedV3(t)

	// Snapshot the seeded rows through a raw connection before migration.
	raw, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path))
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	tables := map[string][]string{
		"workers":         {"label", "pid", "backend", "session_id", "run_id", "registered_at", "heartbeat_at"},
		"runs":            {"run_id", "lead_session_id", "lead_backend", "status", "manifest_json", "lead_pid", "lead_process_start", "created_at", "updated_at"},
		"cards":           frV3CardColumns,
		"events":          {"seq", "run_id", "kind", "payload_json", "created_at"},
		"resume_handoffs": {"id", "status", "schema_version", "saved_at", "body", "body_sha256"},
	}
	before := map[string][]string{}
	for table, cols := range tables {
		before[table] = frDump(t, raw, table, cols)
		if len(before[table]) != 1 {
			t.Fatalf("seed %s rows = %d, want 1", table, len(before[table]))
		}
	}
	_ = raw.Close()

	db, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatalf("open migrated factory: %v", err)
	}
	if got := frSchemaVersion(t, db.DB); got != "4" {
		t.Fatalf("schema_version = %q, want \"4\"", got)
	}
	for table, cols := range tables {
		if got := frDump(t, db.DB, table, cols); !reflect.DeepEqual(got, before[table]) {
			t.Fatalf("%s rows changed by migration:\n got %v\nwant %v", table, got, before[table])
		}
	}
	newCols := frDump(t, db.DB, "cards", frF1CardColumns)
	if len(newCols) != 1 || newCols[0] != strings.Repeat("|", len(frF1CardColumns)-1) {
		t.Fatalf("new card columns of the existing row = %q, want all ''", newCols)
	}
	colsAfterFirst := frTableColumns(t, db.DB, "cards")
	if err := db.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	// Second open changes nothing.
	again, err := OpenFactoryPath(path)
	if err != nil {
		t.Fatalf("second open: %v", err)
	}
	if got := frTableColumns(t, again.DB, "cards"); !reflect.DeepEqual(got, colsAfterFirst) {
		t.Fatalf("cards columns after second open = %v, want %v", got, colsAfterFirst)
	}
	for table, cols := range tables {
		if got := frDump(t, again.DB, table, cols); !reflect.DeepEqual(got, before[table]) {
			t.Fatalf("%s rows changed by second open: %v", table, got)
		}
	}
	if got := frSchemaVersion(t, again.DB); got != "4" {
		t.Fatalf("schema_version after second open = %q, want \"4\"", got)
	}
	_ = again.Close()

	// A fresh database lists every F1 column.
	fresh, err := OpenFactory(factorySandbox(t))
	if err != nil {
		t.Fatalf("open fresh: %v", err)
	}
	have := map[string]bool{}
	for _, c := range frTableColumns(t, fresh.DB, "cards") {
		have[c] = true
	}
	var missing []string
	for _, c := range append(append([]string{}, frV3CardColumns...), frF1CardColumns...) {
		if !have[c] {
			missing = append(missing, c)
		}
	}
	sort.Strings(missing)
	if len(missing) != 0 {
		t.Fatalf("fresh cards table is missing columns %v", missing)
	}
	if got := frSchemaVersion(t, fresh.DB); got != "4" {
		t.Fatalf("fresh schema_version = %q, want \"4\"", got)
	}
	freshPath := fresh.Path
	_ = fresh.Close()

	// A database from the future (version 5) is refused.
	future, err := sql.Open("sqlite", "file:"+filepath.ToSlash(freshPath))
	if err != nil {
		t.Fatalf("open raw fresh: %v", err)
	}
	if _, err := future.Exec(`UPDATE meta SET value='5' WHERE key='schema_version'`); err != nil {
		t.Fatalf("bump version: %v", err)
	}
	_ = future.Close()
	if db, err := OpenFactoryPath(freshPath); err == nil {
		_ = db.Close()
		t.Fatal("schema version 5 was accepted; want an unsupported-version error")
	} else if !strings.Contains(err.Error(), "unsupported factory schema version") {
		t.Fatalf("schema version 5 error = %v, want unsupported-version", err)
	}
}

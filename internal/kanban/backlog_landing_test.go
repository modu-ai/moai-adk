// backlog_landing_test.go — SPEC-TODO-LANDING-EVIDENCE-001 (card t359) M1:
// the stored shape of the landing-evidence column. AC-TLE-001 pins the items
// column sequence and the landing column's own tuple, AC-TLE-002 pins
// archived_items separately so a half-applied migration cannot pass on the
// strength of the other table, and AC-TLE-003 opens a pre-change database
// twice to show the migration is idempotent and rewrites no existing row.
package kanban

import (
	"context"
	"database/sql"
	"path/filepath"
	"strconv"
	"testing"
)

// AC-TLE-001 — the items column sequence ends in landing, and that column is
// nullable TEXT with no default.
func TestBacklogLanding_ItemsColumnShape(t *testing.T) {
	eng := landingEngineFixture(t)
	const want = "seq:INTEGER:0:NULL " +
		"id:TEXT:1:NULL " +
		"text:TEXT:1:NULL " +
		"added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL " +
		"state:TEXT:1:NULL " +
		"landing:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "items"); got != want {
		t.Errorf("items column tuples =\n %s\nwant\n %s", got, want)
	}

	// The landing row on its own: TEXT, nullable, no default. Asserted
	// separately from the sequence so the failure names the shape rather
	// than the whole row when a DEFAULT or NOT NULL creeps in.
	colType, notNull, dflt := landingColumnShape(t, eng, "items")
	if colType != "TEXT" || notNull != 0 || dflt.Valid {
		t.Errorf("items.landing = (type %q, notnull %d, dflt_value %v), want (TEXT, 0, NULL)",
			colType, notNull, dflt)
	}
}

// AC-TLE-002 — archived_items carries the same column, asserted independently
// of AC-TLE-001 so a migration applied to one table only is caught.
func TestBacklogLanding_ArchivedItemsColumnShape(t *testing.T) {
	eng := landingEngineFixture(t)
	const want = "seq:INTEGER:0:NULL " +
		"id:TEXT:1:NULL " +
		"text:TEXT:1:NULL " +
		"added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL " +
		"state:TEXT:1:NULL " +
		"position:INTEGER:1:NULL " +
		"landing:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "archived_items"); got != want {
		t.Errorf("archived_items column tuples =\n %s\nwant\n %s", got, want)
	}
}

// AC-TLE-003 — opening a pre-change database twice adds the column exactly
// once and leaves every existing row byte-for-byte as it was. A rebuild-based
// migration reds here independently of the sequence assertions above, because
// it reassigns seq and reorders rows.
func TestBacklogLanding_MigrationIsIdempotent(t *testing.T) {
	dbPath := preChangeDatabase(t)
	const wantRows = 3

	for open := 1; open <= 2; open++ {
		eng, err := openBacklogEngine(dbPath)
		if err != nil {
			t.Fatalf("open %d: %v", open, err)
		}
		if got := landingColumnCount(t, eng, "items"); got != 1 {
			t.Errorf("open %d: items has %d landing columns, want exactly 1", open, got)
		}
		if got := landingColumnCount(t, eng, "archived_items"); got != 1 {
			t.Errorf("open %d: archived_items has %d landing columns, want exactly 1", open, got)
		}
		if got := preChangeRowTuples(t, eng); got != wantPreChangeRows {
			t.Errorf("open %d: rows =\n %s\nwant\n %s", open, got, wantPreChangeRows)
		}
		if got := rowCount(t, eng, "items"); got != wantRows {
			t.Errorf("open %d: items row count = %d, want %d", open, got, wantRows)
		}
		if err := eng.close(); err != nil {
			t.Fatalf("close %d: %v", open, err)
		}
	}
}

// wantPreChangeRows is the exact (seq, id, text, added_at, spec_id, state)
// content preChangeDatabase inserts, in seq order.
const wantPreChangeRows = "1|t1|alpha|2026-01-01T00:00:00Z||queued;" +
	"2|t2|beta|2026-01-02T00:00:00Z|SPEC-X|picked;" +
	"3|t3|gamma|2026-01-03T00:00:00Z||dropped"

// landingEngineFixture opens an engine over a fresh temp queue database,
// created by the post-change open path.
func landingEngineFixture(t *testing.T) *backlogEngine {
	t.Helper()
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	t.Cleanup(func() { _ = eng.close() })
	return eng
}

// preChangeDatabase builds a database the way a pre-change binary would have
// left it: the two card-bearing tables WITHOUT landing, the version marker
// stamped, and three cards inserted.
func preChangeDatabase(t *testing.T) string {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "backlog.db")
	db, err := sql.Open(sqliteDriverName, backlogDSN(dbPath))
	if err != nil {
		t.Fatalf("open pre-change db: %v", err)
	}
	defer func() { _ = db.Close() }()

	const preChangeDDL = `
CREATE TABLE meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped'))
);
CREATE TABLE archived_items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL,
  position INTEGER NOT NULL
);
`
	if _, err := db.Exec(preChangeDDL); err != nil {
		t.Fatalf("pre-change DDL: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO meta(key, value) VALUES (?, ?)`,
		backlogMetaKeySchemaVersion, backlogSchemaVersion); err != nil {
		t.Fatalf("stamp schema_version: %v", err)
	}
	rows := []struct {
		seq                       int
		id, text, addedAt, specID string
		state                     string
	}{
		{1, "t1", "alpha", "2026-01-01T00:00:00Z", "", "queued"},
		{2, "t2", "beta", "2026-01-02T00:00:00Z", "SPEC-X", "picked"},
		{3, "t3", "gamma", "2026-01-03T00:00:00Z", "", "dropped"},
	}
	for _, r := range rows {
		if _, err := db.Exec(
			`INSERT INTO items(seq, id, text, added_at, spec_id, state) VALUES (?, ?, ?, ?, ?, ?)`,
			r.seq, r.id, r.text, r.addedAt, r.specID, r.state); err != nil {
			t.Fatalf("insert %s: %v", r.id, err)
		}
	}
	return dbPath
}

// landingColumnCount reports how many columns of a table are named landing.
func landingColumnCount(t *testing.T, eng *backlogEngine, table string) int {
	t.Helper()
	var n int
	if err := eng.db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM pragma_table_info(?) WHERE name = ?`,
		table, backlogLandingColumn).Scan(&n); err != nil {
		t.Fatalf("count landing on %s: %v", table, err)
	}
	return n
}

// landingColumnShape returns the landing column's declared type, notnull flag,
// and default value on the named table.
func landingColumnShape(t *testing.T, eng *backlogEngine, table string) (string, int, sql.NullString) {
	t.Helper()
	var colType string
	var notNull int
	var dflt sql.NullString
	if err := eng.db.QueryRowContext(context.Background(),
		`SELECT type, "notnull", dflt_value FROM pragma_table_info(?) WHERE name = ?`,
		table, backlogLandingColumn).Scan(&colType, &notNull, &dflt); err != nil {
		t.Fatalf("read landing shape on %s: %v", table, err)
	}
	return colType, notNull, dflt
}

// preChangeRowTuples renders every items row's pre-change tuple in seq order,
// so an altered value, a reassigned seq, or a reordered row all show up as a
// single string difference.
func preChangeRowTuples(t *testing.T, eng *backlogEngine) string {
	t.Helper()
	rows, err := eng.db.QueryContext(context.Background(),
		`SELECT seq, id, text, added_at, ifnull(spec_id, ''), state FROM items ORDER BY seq`)
	if err != nil {
		t.Fatalf("read items rows: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var out string
	for rows.Next() {
		var seq int
		var id, text, addedAt, specID, state string
		if err := rows.Scan(&seq, &id, &text, &addedAt, &specID, &state); err != nil {
			t.Fatalf("scan items row: %v", err)
		}
		if out != "" {
			out += ";"
		}
		out += strconv.Itoa(seq) + "|" + id + "|" + text + "|" + addedAt + "|" + specID + "|" + state
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate items rows: %v", err)
	}
	return out
}

// rowCount reports the number of rows in a table.
func rowCount(t *testing.T, eng *backlogEngine, table string) int {
	t.Helper()
	var n int
	// The table name is a test-local literal, never user input.
	if err := eng.db.QueryRowContext(context.Background(),
		`SELECT count(*) FROM `+table).Scan(&n); err != nil {
		t.Fatalf("count rows of %s: %v", table, err)
	}
	return n
}

// backlog_transition_stamps_test.go — SPEC-TODO-TRANSITION-STAMPS-001 M1
// acceptance: the additive migration and its convergence contract.
//
// AC-TST-001 — a database created fresh by the new binary and one created by
// the PRE-change DDL then opened by the new binary (the upgrade path)
// converge on the same column set, with the pre-existing columns intact as a
// prefix in their original order and the new columns nullable TEXT.
package factory

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// preChangeItemsDDL is the items/archived_items shape this SPEC inherited —
// the DDL the upgrade-path fixture is built from. It is pinned here rather
// than imported so the test states what "pre-change" means.
const preChangeArchiveDDL = `
CREATE TABLE IF NOT EXISTS meta (
  key   TEXT PRIMARY KEY,
  value TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL CHECK (state IN ('queued','picked','dropped')),
  landing  TEXT
);
CREATE TABLE IF NOT EXISTS archived_items (
  seq      INTEGER PRIMARY KEY,
  id       TEXT    NOT NULL UNIQUE,
  text     TEXT    NOT NULL,
  added_at TEXT    NOT NULL,
  spec_id  TEXT,
  state    TEXT    NOT NULL,
  position INTEGER NOT NULL,
  landing  TEXT
);
INSERT INTO meta(key, value) VALUES ('schema_version', '1');
INSERT INTO meta(key, value) VALUES ('last_seq', '0');
`

// buildPreChangeDB materializes a pre-change database with the raw driver —
// deliberately NOT through openBacklogEngine, which would run the additive
// migration and erase the upgrade path under test.
func buildPreChangeDB(t *testing.T, dbPath string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open pre-change db: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.ExecContext(context.Background(), preChangeArchiveDDL); err != nil {
		t.Fatalf("materialize pre-change db: %v", err)
	}
}

// columnNames reads a table's columns in declaration order.
func columnNames(t *testing.T, eng *backlogEngine, table string) []string {
	t.Helper()
	rows, err := eng.db.QueryContext(context.Background(), `SELECT name FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatalf("pragma_table_info(%s): %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan column name: %v", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate columns: %v", err)
	}
	return names
}

// TestTransitionStampColumns_FreshUpgradedConverge — AC-TST-001. Both
// databases declare the same columns, the pre-existing columns survive as a
// prefix in original order, and the new columns are nullable TEXT.
func TestTransitionStampColumns_FreshUpgradedConverge(t *testing.T) {
	dir := t.TempDir()

	freshEng, err := openBacklogEngine(filepath.Join(dir, "fresh.db"))
	if err != nil {
		t.Fatalf("open fresh: %v", err)
	}
	defer func() { _ = freshEng.close() }()

	upgradedPath := filepath.Join(dir, "upgraded.db")
	buildPreChangeDB(t, upgradedPath)
	upgradedEng, err := openBacklogEngine(upgradedPath)
	if err != nil {
		t.Fatalf("open upgraded: %v", err)
	}
	defer func() { _ = upgradedEng.close() }()

	for _, table := range []string{"items", "archived_items"} {
		freshCols := columnNames(t, freshEng, table)
		upgradedCols := columnNames(t, upgradedEng, table)
		if strings.Join(freshCols, ",") != strings.Join(upgradedCols, ",") {
			t.Errorf("%s column sets diverge:\n fresh:   %v\n upgraded: %v", table, freshCols, upgradedCols)
		}
	}

	// The new columns exist on BOTH and are nullable TEXT with no default.
	// The classification column (SPEC-TODO-CLASSIFY-DISPATCH-001) is appended
	// after the stamps by the same convergence contract, so it appears in the
	// expected sequences here too — a fresh and an upgraded database must
	// carry the identical tuple for every additive column, not only the
	// stamps this test was written for. SPEC-TODO-CLAIM-LEASE-001 extends the
	// pin additively again (same cascade as the freeze re-record): the lease
	// columns append AFTER the classification, preserving the old columns as
	// an exact prefix in original order.
	const wantItems = "seq id text added_at spec_id state landing picked_at dropped_at classification picked_by lease_expires_at"
	if got := strings.Join(columnNames(t, freshEng, "items"), " "); got != wantItems {
		t.Errorf("items columns = %q, want %q (old columns as a prefix, in order)", got, wantItems)
	}
	const wantArchived = "seq id text added_at spec_id state position landing picked_at dropped_at archived_at landing_verdict classification picked_by lease_expires_at"
	if got := strings.Join(columnNames(t, freshEng, "archived_items"), " "); got != wantArchived {
		t.Errorf("archived_items columns = %q, want %q (old columns as a prefix, in order)", got, wantArchived)
	}
	for _, tc := range []struct{ table, column string }{
		{"items", "picked_at"},
		{"items", "dropped_at"},
		{"items", "classification"},
		{"archived_items", "picked_at"},
		{"archived_items", "dropped_at"},
		{"archived_items", "archived_at"},
		{"archived_items", "classification"},
	} {
		tuple := columnTupleSequence(t, freshEng, tc.table)
		matched := false
		for _, part := range strings.Split(tuple, " ") {
			if strings.HasPrefix(part, tc.column+":") {
				matched = true
				if part != tc.column+":TEXT:0:NULL" {
					t.Errorf("%s.%s tuple = %q, want %q:TEXT:0:NULL (nullable TEXT, no default)",
						tc.table, tc.column, part, tc.column)
				}
			}
		}
		if !matched {
			t.Errorf("%s has no %s column", tc.table, tc.column)
		}
	}
}

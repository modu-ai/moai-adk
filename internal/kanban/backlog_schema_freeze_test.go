// backlog_schema_freeze_test.go — SPEC-TODO-ARCHIVE-QUERY-001 (card t394):
// AC-TAQ-012, the schema-identity assertion. The history verb is a reader;
// its landing must add no table, admit no fourth state, and leave
// schema_version stamped at "1" (REQ-TAQ-012, REQ-TDG-004/005 — a fourth
// state would force a table rebuild on every operator queue in the field).
package kanban

import (
	"context"
	"strings"
	"testing"
)

// AC-TAQ-012 — the queue database a reader opens carries exactly the
// physical schema the writers built: five tables, the one non-auto index,
// the three-state CHECK, schema_version "1".
func TestTodoHistoryAddsNoSchemaChange(t *testing.T) {
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := store.Mutate(func(rec *BacklogRecord) error { return rec.ArchiveCard("t1") }); err != nil {
		t.Fatalf("archive: %v", err)
	}

	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	defer func() { _ = eng.close() }()
	ctx := context.Background()

	// The table set is exactly the five the DDL defines — no more, no less.
	var tables []string
	rows, err := eng.db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatalf("read sqlite_master tables: %v", err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tables: %v", err)
	}
	wantTables := "archived_findings archived_items findings items meta"
	if got := strings.Join(tables, " "); got != wantTables {
		t.Errorf("table set = %q, want %q", got, wantTables)
	}

	// The items.state CHECK still admits exactly the three live states —
	// read from the database's own stored SQL, not from the source DDL.
	var itemsSQL string
	if err := eng.db.QueryRowContext(ctx,
		`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = 'items'`).Scan(&itemsSQL); err != nil {
		t.Fatalf("read items DDL: %v", err)
	}
	const wantCheck = "CHECK (state IN ('queued','picked','dropped'))"
	if !strings.Contains(itemsSQL, wantCheck) {
		t.Errorf("items.state CHECK drifted.\n got: %s\nwant it to contain: %s", itemsSQL, wantCheck)
	}

	// The non-auto index set is exactly the one the DDL defines.
	var indexes []string
	idxRows, err := eng.db.QueryContext(ctx,
		`SELECT name FROM sqlite_master WHERE type = 'index' AND name NOT LIKE 'sqlite_autoindex%' ORDER BY name`)
	if err != nil {
		t.Fatalf("read indexes: %v", err)
	}
	for idxRows.Next() {
		var name string
		if err := idxRows.Scan(&name); err != nil {
			t.Fatalf("scan index name: %v", err)
		}
		indexes = append(indexes, name)
	}
	if err := idxRows.Err(); err != nil {
		t.Fatalf("iterate indexes: %v", err)
	}
	if got := strings.Join(indexes, " "); got != "idx_items_state" {
		t.Errorf("index set = %q, want exactly idx_items_state", got)
	}

	// schema_version is still stamped "1" — an older binary refuses any
	// other value, so a bump is a downgrade break, not a feature.
	version, err := eng.schemaVersion(ctx)
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if version != backlogSchemaVersion {
		t.Errorf("schema_version = %q, want %q", version, backlogSchemaVersion)
	}
}

// backlog_schema_freeze_test.go — SPEC-TODO-ARCHIVE-QUERY-001 (card t394):
// AC-TAQ-012, the schema-identity assertion. The history verb is a reader;
// its landing must add no table, admit no fourth state, and leave
// schema_version stamped at "1" (REQ-TAQ-012, REQ-TDG-004/005 — a fourth
// state would force a table rebuild on every operator queue in the field).
//
// Extended by SPEC-TODO-LANDING-EVIDENCE-001 (card t359), AC-TLE-019: the
// guard was column-blind — a planted column on either table left all four of
// its assertions GREEN, so a schema change could land without tripping the
// freeze. It now pins the exact ordered (name, type, notnull, dflt_value)
// tuple sequence of items AND archived_items, asserted separately per table
// so a half-applied migration cannot satisfy both.
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

	// The exact ordered column tuples of the two card-bearing tables. Names
	// alone would not catch a retype, a nullability flip, or a default
	// appearing, so each column is pinned as (name, type, notnull,
	// dflt_value). The two tables are asserted separately: a migration
	// applied to one only must fail loudly rather than half-pass.
	const wantItemsColumns = "seq:INTEGER:0:NULL " +
		"id:TEXT:1:NULL " +
		"text:TEXT:1:NULL " +
		"added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL " +
		"state:TEXT:1:NULL " +
		"landing:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "items"); got != wantItemsColumns {
		t.Errorf("items column tuples =\n %s\nwant\n %s", got, wantItemsColumns)
	}

	const wantArchivedItemsColumns = "seq:INTEGER:0:NULL " +
		"id:TEXT:1:NULL " +
		"text:TEXT:1:NULL " +
		"added_at:TEXT:1:NULL " +
		"spec_id:TEXT:0:NULL " +
		"state:TEXT:1:NULL " +
		"position:INTEGER:1:NULL " +
		"landing:TEXT:0:NULL"
	if got := columnTupleSequence(t, eng, "archived_items"); got != wantArchivedItemsColumns {
		t.Errorf("archived_items column tuples =\n %s\nwant\n %s", got, wantArchivedItemsColumns)
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

// columnTupleSequence reads a table's columns in declaration order and renders
// each as "name:type:notnull:dflt_value", joined by spaces. dflt_value is
// rendered through SQL quote() so an empty-string default (”) is
// distinguishable from no default (NULL) — the distinction AC-TLE-019c turns
// on.
func columnTupleSequence(t *testing.T, eng *backlogEngine, table string) string {
	t.Helper()
	rows, err := eng.db.QueryContext(context.Background(),
		`SELECT name, type, "notnull", ifnull(quote(dflt_value), 'NULL') `+
			`FROM pragma_table_info(?)`, table)
	if err != nil {
		t.Fatalf("read pragma_table_info(%s): %v", table, err)
	}
	defer func() { _ = rows.Close() }()
	var tuples []string
	for rows.Next() {
		var name, colType, notNull, dflt string
		if err := rows.Scan(&name, &colType, &notNull, &dflt); err != nil {
			t.Fatalf("scan pragma_table_info(%s): %v", table, err)
		}
		tuples = append(tuples, name+":"+colType+":"+notNull+":"+dflt)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate pragma_table_info(%s): %v", table, err)
	}
	if len(tuples) == 0 {
		t.Fatalf("pragma_table_info(%s) returned no columns", table)
	}
	return strings.Join(tuples, " ")
}

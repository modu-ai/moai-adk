// todo_owner_label_test.go — SPEC-TODO-SURFACE-POLISH-001 (card t1349)
// M3: the owner_label vocabulary migration (AC-TSP-050) and the write-path
// normalization (AC-TSP-052).
package kanban

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

// seedOwnerLabelsFixture builds a store whose todo_runtime_assignments
// table carries the measured legacy shape: lead x3, worker-67 x2, and one
// already-canonical lane-1 row — plus one card in items so the migration's
// other-tables-invariant has something real to protect. The assignments are
// hand-written over the store's own DDL so the fixture's histogram is exact
// (the write path itself normalizes, and would pollute the fixture).
func seedOwnerLabelsFixture(t *testing.T) *BacklogStore {
	t.Helper()
	root := t.TempDir()
	t.Setenv("MOAI_HOME", t.TempDir())
	store := NewBacklogStore(BacklogPathForRoot(root))
	if _, _, err := store.Add("owner label fixture card"); err != nil {
		t.Fatalf("seed card: %v", err)
	}
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(todoRuntimeDDL); err != nil {
		t.Fatalf("ensure runtime ddl: %v", err)
	}
	stmt := `INSERT INTO todo_runtime_assignments(run_id,card_id,owner_label,reported_state,event_kind,provenance_json)
	         VALUES(?,?,?,?,?,?)`
	for i, row := range [][3]string{
		{"lead", "queued", "lead"},
		{"lead", "assigned", "lead"},
		{"lead", "assigned", "lead"},
		{"worker-67", "assigned", "worker-67"},
		{"worker-67", "completed", "worker-67"},
	} {
		if _, err := db.Exec(stmt, "legacy-run-"+string(rune('a'+i))+"-"+row[0], "t1", row[2], row[1], "card.event", "{}"); err != nil {
			t.Fatalf("seed legacy row %d: %v", i, err)
		}
	}
	if _, err := db.Exec(stmt, "run-canonical", "t1", "lane-1", "assigned", "card.assigned", "{}"); err != nil {
		t.Fatalf("seed canonical row: %v", err)
	}
	return store
}

// AC-TSP-050 — one migration relabels every legacy spelling to the
// canonical vocabulary, leaves the canonical row alone, and touches no
// other table. Re-running migrates zero rows (idempotence).
func TestOwnerLabelMigration(t *testing.T) {
	store := seedOwnerLabelsFixture(t)

	// Fingerprint the tables the migration must not touch.
	otherTablesBefore := ownerLabelOtherTablesFingerprint(t, store)

	migrated, before, after, err := store.MigrateOwnerLabelVocabulary()
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if migrated != 5 {
		t.Errorf("migrated = %d, want 5 (lead x3 + worker-67 x2)", migrated)
	}
	if before["lead"] != 3 || before["worker-67"] != 2 || before["lane-1"] != 1 {
		t.Errorf("before histogram = %v, want lead x3, worker-67 x2, lane-1 x1", before)
	}
	if after["leader"] != 3 || after["lane-67"] != 2 || after["lane-1"] != 1 {
		t.Errorf("after histogram = %v, want leader x3, lane-67 x2, lane-1 x1", after)
	}

	// No legacy spelling survives.
	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, a := range rec.Runtime.Assignments {
		if IsLegacyLeaderSpelling(a.OwnerLabel) || IsLegacyFactoryRoleValue(a.OwnerLabel) {
			t.Errorf("assignment %s/%s still carries legacy label %q", a.RunID, a.CardID, a.OwnerLabel)
		}
	}

	// Idempotence: a second run migrates nothing and the histogram holds.
	migrated2, _, after2, err := store.MigrateOwnerLabelVocabulary()
	if err != nil {
		t.Fatalf("re-migrate: %v", err)
	}
	if migrated2 != 0 {
		t.Errorf("second migration moved %d rows, want 0 (idempotence)", migrated2)
	}
	if after2["leader"] != 3 || after2["lane-67"] != 2 || after2["lane-1"] != 1 {
		t.Errorf("post-re-run histogram = %v, want the same canonical counts", after2)
	}

	// Other tables are byte-identical.
	ownerLabelAssertOtherTablesUnchanged(t, store, otherTablesBefore)
}

// AC-TSP-052 — the write path records the canonical vocabulary only: a
// legacy label carried into recordRuntime lands relabeled, and the table
// stays legacy-free.
func TestOwnerLabelWritesCanonical(t *testing.T) {
	store := seedOwnerLabelsFixture(t)
	// Clean the seeded legacy rows so the write-path assertion is exact.
	if _, _, _, err := store.MigrateOwnerLabelVocabulary(); err != nil {
		t.Fatalf("pre-migrate: %v", err)
	}

	if err := store.recordRuntime(TodoRuntimeRun{RunID: "run-new", ManifestJSON: "{}"}, &TodoRuntimeAssignment{
		RunID: "run-new", CardID: "t1", OwnerLabel: "worker-12",
		ReportedState: "assigned", EventKind: "card.assigned", ProvenanceJSON: "{}",
	}); err != nil {
		t.Fatalf("record with legacy label: %v", err)
	}

	rec, err := store.LoadPure()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	for _, a := range rec.Runtime.Assignments {
		if a.RunID == "run-new" && a.OwnerLabel != "lane-12" {
			t.Errorf("newly written owner_label = %q, want the canonical lane-12", a.OwnerLabel)
		}
		if IsLegacyLeaderSpelling(a.OwnerLabel) || IsLegacyFactoryRoleValue(a.OwnerLabel) {
			t.Errorf("owner_label %q is legacy after a canonical-vocabulary write", a.OwnerLabel)
		}
	}
}

// ownerLabelOtherTablesFingerprint fingerprints the items and identities
// tables — everything the migration must leave byte-identical.
func ownerLabelOtherTablesFingerprint(t *testing.T, store *BacklogStore) map[string]string {
	t.Helper()
	db, err := sql.Open("sqlite", store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	out := map[string]string{}
	for table, query := range map[string]string{
		"items":      `SELECT group_concat(id || ':' || state, ',') FROM (SELECT id, state FROM items ORDER BY id)`,
		"identities": `SELECT IFNULL(group_concat(card_uuid || ':' || seq, ','), '') FROM (SELECT card_uuid, seq FROM todo_identities ORDER BY card_uuid)`,
	} {
		var raw string
		if err := db.QueryRowContext(context.Background(), query).Scan(&raw); err != nil {
			// The identities table may not exist on a fixture without
			// recorded identities; an absent table is a stable state too.
			out[table] = "unavailable"
			continue
		}
		out[table] = raw
	}
	return out
}

// ownerLabelAssertOtherTablesUnchanged compares the fingerprint above.
func ownerLabelAssertOtherTablesUnchanged(t *testing.T, store *BacklogStore, before map[string]string) {
	t.Helper()
	after := ownerLabelOtherTablesFingerprint(t, store)
	for table, was := range before {
		if after[table] != was {
			t.Errorf("table %s changed across the migration: %q -> %q — the migration owns exactly one column of one table", table, was, after[table])
		}
	}
}

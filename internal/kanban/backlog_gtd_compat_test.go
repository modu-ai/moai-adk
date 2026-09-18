package kanban

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newGTDTestStore(t *testing.T) *BacklogStore {
	t.Helper()
	return NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
}

func TestMigrateGTDSchemaIdempotent(t *testing.T) {
	store := newGTDTestStore(t)
	if _, _, err := store.Add("legacy card"); err != nil {
		t.Fatal(err)
	}
	if err := MigrateGTDSchema(store); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	db, err := sql.Open(sqliteDriverName, store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	var core, gtd string
	if err := db.QueryRow(`SELECT value FROM meta WHERE key='schema_version'`).Scan(&core); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`SELECT value FROM gtd_meta WHERE key='schema_version'`).Scan(&gtd); err != nil {
		t.Fatal(err)
	}
	if core != "1" || gtd != "1" {
		t.Fatalf("versions core=%q gtd=%q, want 1/1", core, gtd)
	}
	var before int
	if err := db.QueryRow(`SELECT count(*) FROM gtd_meta`).Scan(&before); err != nil {
		t.Fatal(err)
	}
	if err := MigrateGTDSchema(store); err != nil {
		t.Fatalf("second migration: %v", err)
	}
	var after int
	if err := db.QueryRow(`SELECT count(*) FROM gtd_meta`).Scan(&after); err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Fatalf("idempotent migration changed gtd_meta rows: %d -> %d", before, after)
	}
}

func TestGTDLegacyRoundTrip(t *testing.T) {
	store := newGTDTestStore(t)
	card, _, err := store.Add("round-trip card")
	if err != nil {
		t.Fatal(err)
	}
	if err := MigrateGTDSchema(store); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open(sqliteDriverName, store.EnginePath())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(`INSERT INTO gtd_items(item_id, content, source, sensitivity, event_id, status, source_revision) VALUES(?,?,?,?,?,?,?)`, "g1", "captured content", "user", "private", "evt-1", "organized", 1); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO gtd_relations(subject_id, object_id, kind, note, source, assertion_status, source_revision, policy_version) VALUES(?,?,?,?,?,?,?,?)`, "g1", card.ID, "related_to", []byte{0, 1, 2, 255}, "user", "confirmed", 1, "p1"); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(r *BacklogRecord) error { return r.ArchiveCard(card.ID) }); err != nil {
		t.Fatal(err)
	}
	if err := store.Mutate(func(r *BacklogRecord) error { return r.RestoreCard(card.ID) }); err != nil {
		t.Fatal(err)
	}
	var note []byte
	if err := db.QueryRow(`SELECT note FROM gtd_relations WHERE subject_id=? AND object_id=?`, "g1", card.ID).Scan(&note); err != nil {
		t.Fatal(err)
	}
	want := []byte{0, 1, 2, 255}
	if string(note) != string(want) {
		t.Fatalf("relation bytes changed: %v, want %v", note, want)
	}
	rec, err := store.Load()
	if err != nil || len(rec.Items) != 1 || rec.Items[0].ID != card.ID {
		t.Fatalf("legacy card identity not restored: rec=%+v err=%v", rec, err)
	}
}

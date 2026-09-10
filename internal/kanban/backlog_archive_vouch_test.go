// backlog_archive_vouch_test.go — SPEC-TODO-ARCHIVE-QUERY-001 (card t394):
// the archive-vouch probe a read-only surface uses to learn which store is
// answering and whether that store can vouch for an archive (REQ-TAQ-013).
//
// The probe must observe the schema BEFORE the DDL runs — openBacklogEngine
// recreates missing archive tables on every open (IF NOT EXISTS), which
// would erase exactly the fact the probe reports. These tests pin that
// ordering: a dropped-tables database must probe as archive-less even
// though a Load() of the same database would silently recreate the tables.
package kanban

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// vouchFixture returns a store over a fresh temp queue.
func vouchFixture(t *testing.T) *BacklogStore {
	t.Helper()
	return NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
}

// dropArchiveTables removes the archive tables the way a pre-archive
// database lacks them (the shape backlog_archive_test.go constructs).
func dropArchiveTables(t *testing.T, store *BacklogStore) {
	t.Helper()
	eng, err := openBacklogEngine(store.EnginePath())
	if err != nil {
		t.Fatalf("open engine: %v", err)
	}
	for _, stmt := range []string{`DROP TABLE archived_items`, `DROP TABLE archived_findings`} {
		if _, err := eng.db.ExecContext(context.Background(), stmt); err != nil {
			t.Fatalf("%s: %v", stmt, err)
		}
	}
	if err := eng.close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

// TestInspectArchiveVouch_ProbeSeesDroppedTables — the load-bearing
// ordering: the probe answers from the pre-DDL schema, so a database whose
// archive tables were dropped reads as archive-less even though opening it
// through the engine would recreate them.
func TestInspectArchiveVouch_ProbeSeesDroppedTables(t *testing.T) {
	store := vouchFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	dropArchiveTables(t, store)

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreSQLite {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreSQLite)
	}
	if got.HasArchive {
		t.Error("HasArchive = true for a database whose archive tables are absent — the probe saw the post-DDL schema")
	}
}

// TestInspectArchiveVouch_VouchedStore — a normal database vouches for its
// (possibly empty) archive.
func TestInspectArchiveVouch_VouchedStore(t *testing.T) {
	store := vouchFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreSQLite {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreSQLite)
	}
	if !got.HasArchive {
		t.Error("HasArchive = false for a normal database — the archive tables exist")
	}
}

// TestInspectArchiveVouch_LegacyJSON — a queue whose only artifact is a
// legacy backlog.json (no db) is served by a store that cannot carry an
// archive at all.
func TestInspectArchiveVouch_LegacyJSON(t *testing.T) {
	store := vouchFixture(t)
	legacy := `{"version":1,"last_seq":7,"items":[{"id":"t1","text":"alpha work","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[]}`
	if err := os.WriteFile(store.Path(), []byte(legacy), 0o600); err != nil {
		t.Fatalf("write legacy json: %v", err)
	}

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreLegacyJSON {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreLegacyJSON)
	}
	if got.HasArchive {
		t.Error("HasArchive = true for a legacy JSON store — the shape carries no archived field")
	}
}

// TestInspectArchiveVouch_NoStore — a project with neither artifact has no
// store to name.
func TestInspectArchiveVouch_NoStore(t *testing.T) {
	store := vouchFixture(t) // creates only the temp dir; nothing is written

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreNone {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreNone)
	}
	if got.HasArchive {
		t.Error("HasArchive = true with no store at all")
	}
}

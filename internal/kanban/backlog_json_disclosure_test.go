// backlog_json_disclosure_test.go — SPEC-BACKLOG-JSON-DISCLOSURE-001 (card
// t395): the State D fact the vouch probe already measures and used to
// discard — a backlog.json sitting beside the database that answers reads.
//
// The field is meaningful ONLY on the SQLite branch. On the legacy-JSON
// branch the JSON *is* the answering store, and on the no-store branch
// there is nothing to disclose; both keep it false, and that is asserted
// here rather than left implicit (REQ-BJD-001, plan.md M1).
package kanban

import (
	"os"
	"path/filepath"
	"testing"
)

// legacyBacklogJSON is a well-formed pre-SQLite record — the shape an
// export or an interrupted migration leaves at the canonical path.
const legacyBacklogJSON = `{"version":1,"last_seq":7,"items":[{"id":"t1","text":"alpha work","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"}],"findings":[]}`

// TestInspectArchiveVouch_StateD_ReportsNonAuthoritativeJSON — the whole
// point: a database AND a backlog.json, database answering, the JSON
// reported as present and not authoritative.
func TestInspectArchiveVouch_StateD_ReportsNonAuthoritativeJSON(t *testing.T) {
	store := vouchFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := os.WriteFile(store.Path(), []byte(legacyBacklogJSON), 0o600); err != nil {
		t.Fatalf("write backlog.json beside the database: %v", err)
	}

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreSQLite {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreSQLite)
	}
	if !got.NonAuthoritativeJSON {
		t.Error("NonAuthoritativeJSON = false with a backlog.json beside the answering database — the State D fact was discarded")
	}
}

// TestInspectArchiveVouch_SteadyState_NoJSONToDisclose — state C, the
// migrated layout: nothing beside the database, nothing to disclose.
func TestInspectArchiveVouch_SteadyState_NoJSONToDisclose(t *testing.T) {
	store := vouchFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreSQLite {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreSQLite)
	}
	if got.NonAuthoritativeJSON {
		t.Error("NonAuthoritativeJSON = true with no backlog.json present")
	}
}

// TestInspectArchiveVouch_LegacyJSON_IsAuthoritative — the JSON is the
// answering store here, so calling it non-authoritative would be false.
func TestInspectArchiveVouch_LegacyJSON_IsAuthoritative(t *testing.T) {
	store := vouchFixture(t)
	if err := os.WriteFile(store.Path(), []byte(legacyBacklogJSON), 0o600); err != nil {
		t.Fatalf("write legacy json: %v", err)
	}

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreLegacyJSON {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreLegacyJSON)
	}
	if got.NonAuthoritativeJSON {
		t.Error("NonAuthoritativeJSON = true on the legacy branch — there the JSON *is* the store that answered")
	}
}

// TestInspectArchiveVouch_NoStore_NothingToDisclose — neither artifact.
func TestInspectArchiveVouch_NoStore_NothingToDisclose(t *testing.T) {
	store := vouchFixture(t)

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreNone {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreNone)
	}
	if got.NonAuthoritativeJSON {
		t.Error("NonAuthoritativeJSON = true with no store at all")
	}
}

// TestInspectArchiveVouch_ReadOnlyDatabase — AC-BJD-006: the probe path is
// read-only, so a database the process cannot write is still inspectable
// and the probe creates no lock file of its own.
func TestInspectArchiveVouch_ReadOnlyDatabase(t *testing.T) {
	store := vouchFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := os.WriteFile(store.Path(), []byte(legacyBacklogJSON), 0o600); err != nil {
		t.Fatalf("write backlog.json: %v", err)
	}
	// Remove any lock artifact the fixture's own Add left behind, so the
	// assertion below observes the probe's execution and nothing else.
	lockPath := store.LockPath()
	_ = os.Remove(lockPath)

	dbPath := store.EnginePath()
	if err := os.Chmod(dbPath, 0o400); err != nil {
		t.Fatalf("chmod read-only: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(dbPath, 0o600) })

	got := InspectBacklogArchiveVouch(store.Path())
	if got.Store != BacklogStoreSQLite {
		t.Errorf("Store = %q, want %q", got.Store, BacklogStoreSQLite)
	}
	if !got.NonAuthoritativeJSON {
		t.Error("NonAuthoritativeJSON = false against a read-only database that has a backlog.json beside it")
	}
	if _, err := os.Stat(lockPath); err == nil {
		t.Errorf("the inspector created a queue lock at %s — the disclosure path must take no lock", filepath.Base(lockPath))
	}
}

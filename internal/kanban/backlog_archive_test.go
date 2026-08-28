// backlog_archive_test.go — SPEC-TODO-DESTRUCTIVE-GUARD-001 (card t330):
// the storage half of `done`'s reversal.
//
// Two of these are FREEZE assertions (AC-TDG-004, AC-TDG-005) — they pass at
// the base tree and are meaningful precisely because the natural
// implementation violates them: a fourth `state` value would cost a
// table-rebuild migration on every operator queue in the field, and a bumped
// schema_version would make an older binary refuse to open the database at
// all.
package kanban

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// archiveFixture returns a store over a fresh temp queue.
func archiveFixture(t *testing.T) *BacklogStore {
	t.Helper()
	return NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
}

// AC-TDG-004 (half 1) — the state enum still carries exactly three values.
func TestBacklogArchive_StateEnumUnchanged(t *testing.T) {
	states := []BacklogState{BacklogStateQueued, BacklogStatePicked, BacklogStateDropped}
	if len(states) != 3 {
		t.Fatalf("BacklogState carries %d values, want 3", len(states))
	}
	for _, s := range states {
		switch s {
		case BacklogStateQueued, BacklogStatePicked, BacklogStateDropped:
		default:
			t.Errorf("unexpected state %q", s)
		}
	}
}

// AC-TDG-004 (half 2) — the per-item contract is still five fields.
func TestBacklogArchive_PerItemContractFrozen(t *testing.T) {
	typ := reflect.TypeOf(BacklogItem{})
	if typ.NumField() != 5 {
		names := make([]string, 0, typ.NumField())
		for i := 0; i < typ.NumField(); i++ {
			names = append(names, typ.Field(i).Name)
		}
		t.Fatalf("BacklogItem carries %d fields (%v), want the frozen 5", typ.NumField(), names)
	}
}

// AC-TDG-005 — the stamped schema version is not bumped. An older binary
// aborts on ANY mismatch, a newer stamp included, so bumping would make the
// downgrade route refuse to open the database rather than degrade.
func TestBacklogArchive_SchemaVersionNotBumped(t *testing.T) {
	if backlogSchemaVersion != "1" {
		t.Fatalf("backlogSchemaVersion = %q, want \"1\" — a bump breaks every older binary", backlogSchemaVersion)
	}

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
	got, err := eng.schemaVersion(context.Background())
	if err != nil {
		t.Fatalf("read schema_version: %v", err)
	}
	if got != "1" {
		t.Errorf("stamped schema_version = %q on a database holding an archived row, want \"1\"", got)
	}
}

// The archive tables land on a database created before they existed: the DDL
// is IF NOT EXISTS and runs on every open, which is the whole reason the
// archive is additive tables rather than a fourth state value.
func TestBacklogArchive_TablesLandOnAnExistingDatabase(t *testing.T) {
	store := archiveFixture(t)
	if _, _, err := store.Add("alpha work"); err != nil {
		t.Fatalf("add: %v", err)
	}
	// Drop the archive tables to simulate a database created by a binary that
	// predates them, then reopen.
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

	if err := store.Mutate(func(rec *BacklogRecord) error { return rec.ArchiveCard("t1") }); err != nil {
		t.Fatalf("archive on a database that lacked the tables: %v", err)
	}
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(rec.Archived) != 1 || rec.Archived[0].Item.ID != "t1" {
		t.Errorf("archive = %+v, want the single entry t1", rec.Archived)
	}
}

// AC-TDG-002 (storage half) — archive then restore returns the record to the
// bytes it had, including the ORDER of items and findings. A card spliced out
// of the middle and appended back reads correctly and is not the record the
// operator had.
func TestBacklogArchive_RestorePreservesPositions(t *testing.T) {
	store := archiveFixture(t)
	for _, text := range []string{"alpha", "beta", "gamma"} {
		if _, _, err := store.Add(text); err != nil {
			t.Fatalf("add %q: %v", text, err)
		}
	}
	if err := store.Mutate(func(rec *BacklogRecord) error {
		rec.Findings = append(rec.Findings,
			BacklogFinding{SubjectID: "t1", RelatedID: "t3", Relation: BacklogRelationContains, Source: BacklogSourceAgent, At: "2026-01-01T00:00:00Z"},
			BacklogFinding{SubjectID: "t2", RelatedID: "t3", Relation: BacklogRelationConflicts, Source: BacklogSourceAgent, At: "2026-01-01T00:00:01Z"},
			BacklogFinding{SubjectID: "t2", RelatedID: "t1", Relation: BacklogRelationReplaces, Source: BacklogSourceAgent, At: "2026-01-01T00:00:02Z"},
		)
		return nil
	}); err != nil {
		t.Fatalf("seed findings: %v", err)
	}

	before, err := store.Load()
	if err != nil {
		t.Fatalf("load before: %v", err)
	}

	// t2 sits in the MIDDLE and its findings are at indexes 1 and 2.
	if err := store.Mutate(func(rec *BacklogRecord) error { return rec.ArchiveCard("t2") }); err != nil {
		t.Fatalf("archive: %v", err)
	}
	mid, err := store.Load()
	if err != nil {
		t.Fatalf("load mid: %v", err)
	}
	if ids := itemIDsOf(mid); !reflect.DeepEqual(ids, []string{"t1", "t3"}) {
		t.Errorf("live ids after archive = %v, want [t1 t3]", ids)
	}
	if len(mid.Findings) != 1 || mid.Findings[0].SubjectID != "t1" {
		t.Errorf("live findings after archive = %+v, want only the t1/t3 finding", mid.Findings)
	}
	if len(mid.Archived) != 1 || len(mid.Archived[0].Findings) != 2 {
		t.Fatalf("archive entry = %+v, want t2 carrying its 2 findings", mid.Archived)
	}

	if err := store.Mutate(func(rec *BacklogRecord) error { return rec.RestoreCard("t2") }); err != nil {
		t.Fatalf("restore: %v", err)
	}
	after, err := store.Load()
	if err != nil {
		t.Fatalf("load after: %v", err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Errorf("archive+restore is not an exact reversal.\n got: %+v\nwant: %+v", after, before)
	}
	if len(after.Archived) != 0 {
		t.Errorf("restore left %d archive entries, want 0 (REQ-TDG-016)", len(after.Archived))
	}
}

// itemIDsOf renders a record's live item ids.
func itemIDsOf(rec *BacklogRecord) []string {
	ids := make([]string, 0, len(rec.Items))
	for _, it := range rec.Items {
		ids = append(ids, it.ID)
	}
	return ids
}

// A restore taken after the queue moved on lands at the end rather than
// failing. The exact-position guarantee covers the round trip this SPEC
// promises — `done` immediately followed by `undone`; once other cards have
// come and gone there is no exact answer, and refusing there would strand the
// archive for the operator who waited a day before changing their mind.
func TestBacklogArchive_RestoreClampsWhenTheQueueMovedOn(t *testing.T) {
	rec := &BacklogRecord{}
	normalizeBacklogRecord(rec)
	rec.Items = append(rec.Items,
		BacklogItem{ID: "t1", Text: "alpha", State: BacklogStateQueued},
		BacklogItem{ID: "t2", Text: "beta", State: BacklogStateQueued},
		BacklogItem{ID: "t3", Text: "gamma", State: BacklogStateQueued},
	)
	rec.Findings = append(rec.Findings,
		BacklogFinding{SubjectID: "t1", RelatedID: "t3", Relation: BacklogRelationContains},
		BacklogFinding{SubjectID: "t3", RelatedID: "t2", Relation: BacklogRelationConflicts},
	)
	// Archive the LAST card (position 2). Both findings name t3, so both ride
	// with it, then empty the live queue underneath the entry.
	if err := rec.ArchiveCard("t3"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if len(rec.Findings) != 0 {
		t.Fatalf("live findings after archiving t3 = %+v, want none — both name it", rec.Findings)
	}
	rec.Items = nil
	rec.Findings = nil

	if err := rec.RestoreCard("t3"); err != nil {
		t.Fatalf("restore into a shrunken queue must succeed: %v", err)
	}
	if ids := itemIDsOf(rec); !reflect.DeepEqual(ids, []string{"t3"}) {
		t.Errorf("restored ids = %v, want [t3] — the position clamps to the end", ids)
	}
	if len(rec.Findings) != 2 {
		t.Errorf("restored findings = %+v, want both archived findings back", rec.Findings)
	}
	for _, f := range rec.Findings {
		if !f.Names("t3") {
			t.Errorf("restored finding %+v does not name t3", f)
		}
	}
	if len(rec.Archived) != 0 {
		t.Errorf("restore left %d archive entries, want 0", len(rec.Archived))
	}
}

// AC-TDG-013 (unit half) — the restore guard refuses a reissued id. The id
// allocator cannot in fact reissue one (last_seq only ever rises, and it now
// clears archived ids too), so the guard is asserted directly on the record.
func TestBacklogArchive_RestoreRefusesAReissuedID(t *testing.T) {
	rec := &BacklogRecord{}
	normalizeBacklogRecord(rec)
	rec.Items = append(rec.Items, BacklogItem{ID: "t1", Text: "original", State: BacklogStateQueued})
	if err := rec.ArchiveCard("t1"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	rec.Items = append(rec.Items, BacklogItem{ID: "t1", Text: "a different card", State: BacklogStateQueued})

	err := rec.RestoreCard("t1")
	if err == nil {
		t.Fatal("restore must refuse when the id has been reissued")
	}
	if got := err.Error(); !strings.Contains(got, "reissued") || !strings.Contains(got, "a different card") {
		t.Errorf("refusal %q must name the reissue and the live card it would have overwritten", got)
	}
	if len(rec.Items) != 1 || rec.Items[0].Text != "a different card" {
		t.Errorf("the live card was disturbed: %+v", rec.Items)
	}
	if len(rec.Archived) != 1 {
		t.Errorf("a refused restore must leave the archive entry in place, got %d", len(rec.Archived))
	}
}

// The high-water mark clears archived ids, so an archived card's id can never
// be reissued to a new card — which is what makes the guard above unreachable
// through the CLI rather than merely unlikely.
func TestBacklogArchive_LastSeqClearsArchivedIDs(t *testing.T) {
	rec := &BacklogRecord{
		Archived: []BacklogArchiveEntry{{Item: BacklogItem{ID: "t7", State: BacklogStateQueued}}},
	}
	normalizeBacklogRecord(rec)
	if rec.LastSeq != 7 {
		t.Errorf("last_seq = %d beside an archived t7, want 7 — a lower mark would reissue t7", rec.LastSeq)
	}
}

// AC-TDG-016 — a second restore refuses, because the entry is gone.
func TestBacklogArchive_RestoreIsNotRepeatable(t *testing.T) {
	rec := &BacklogRecord{}
	normalizeBacklogRecord(rec)
	rec.Items = append(rec.Items, BacklogItem{ID: "t1", Text: "alpha", State: BacklogStateQueued})
	if err := rec.ArchiveCard("t1"); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if err := rec.RestoreCard("t1"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if err := rec.RestoreCard("t1"); err == nil {
		t.Fatal("a second restore must refuse — the entry no longer exists")
	}
	if err := rec.ArchiveCard("t1"); err != nil {
		t.Fatalf("re-archiving a restored card must work: %v", err)
	}
	if err := rec.ArchiveCard("t1"); err == nil {
		t.Fatal("archiving twice must refuse — a row has exactly one home")
	}
}

// The archive survives the lazy migration from a JSON-only queue, which is
// the reachable form of the "both serializations" obligation (AC-TDG-006).
func TestBacklogArchive_SurvivesLegacyMigration(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "backlog.json")
	legacy := `{"version":1,"last_seq":2,` +
		`"items":[{"id":"t2","text":"beta","added_at":"2026-01-01T00:00:01Z","spec_id":null,"state":"queued"}],` +
		`"findings":[],` +
		`"archived":[{"item":{"id":"t1","text":"alpha","added_at":"2026-01-01T00:00:00Z","spec_id":null,"state":"queued"},` +
		`"position":0,"findings":[{"finding":{"subject_id":"t1","related_id":"t2","relation":"contains",` +
		`"source":"agent","score":0,"note":"","at":"2026-01-01T00:00:02Z"},"position":0}]}]}` + "\n"
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatalf("seed legacy: %v", err)
	}

	store := NewBacklogStore(path)
	rec, err := store.Load()
	if err != nil {
		t.Fatalf("load (migrates): %v", err)
	}
	if len(rec.Archived) != 1 || rec.Archived[0].Item.ID != "t1" {
		t.Fatalf("archive lost in migration: %+v", rec.Archived)
	}
	if len(rec.Archived[0].Findings) != 1 || rec.Archived[0].Findings[0].Finding.RelatedID != "t2" {
		t.Fatalf("archived finding lost in migration: %+v", rec.Archived[0].Findings)
	}
	if err := store.Mutate(func(r *BacklogRecord) error { return r.RestoreCard("t1") }); err != nil {
		t.Fatalf("restore after migration: %v", err)
	}
	restored, err := store.Load()
	if err != nil {
		t.Fatalf("load after restore: %v", err)
	}
	if ids := itemIDsOf(restored); !reflect.DeepEqual(ids, []string{"t1", "t2"}) {
		t.Errorf("restored ids = %v, want [t1 t2] — the archived card returns to position 0", ids)
	}
	if len(restored.Findings) != 1 {
		t.Errorf("restored findings = %+v, want the archived finding back", restored.Findings)
	}
}

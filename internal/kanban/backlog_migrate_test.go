// backlog_migrate_test.go — the migration safety net
// (SPEC-TODO-SQLITE-001 AC-TOSQ-001..005, AC-TOSQ-017; M2).
//
// The card these tests guard replaces the queue eight lanes and a lead are
// reading right now. A migration that silently drops a field, reuses an id, or
// removes the legacy file on failure loses operator work that cannot be
// regenerated. Every safety mechanism below is therefore exercised by FIRING
// the failure — a malformed seed, a parity mismatch, a crashed cutover — not
// by asserting the happy path and citing the design.
package kanban

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// migrationFixture is a synthesized legacy record covering every shape the
// migration has to carry: all three states, a null and a non-null spec id, a
// findings pair, and a last_seq ABOVE max-present (so the derive rule cannot
// mask a dropped mark). Card texts are invented — production queue contents
// never enter the repository (spec.md C-5).
const migrationFixture = `{
  "version": 1,
  "last_seq": 42,
  "items": [
    {"id": "t7",  "text": "queued card alpha",  "added_at": "2026-01-02T03:04:05Z", "spec_id": null, "state": "queued"},
    {"id": "t3",  "text": "picked card bravo",  "added_at": "2026-01-02T03:04:06Z", "spec_id": "SPEC-EXAMPLE-001", "state": "picked"},
    {"id": "t11", "text": "dropped card delta", "added_at": "2026-01-02T03:04:07Z", "spec_id": null, "state": "dropped"},
    {"id": "t2",  "text": "queued card echo",   "added_at": "2026-01-02T03:04:08Z", "spec_id": null, "state": "queued"}
  ],
  "findings": [
    {"subject_id": "t7", "related_id": "t2", "relation": "near-duplicate", "source": "mechanical", "score": 0.91, "note": "", "at": "2026-01-02T03:04:09Z"},
    {"subject_id": "t3", "related_id": "t11", "relation": "replaces", "source": "agent", "score": 0, "note": "bravo supersedes delta", "at": "2026-01-02T03:04:10Z"}
  ]
}
`

// seedMigrationRoot writes the fixture as a legacy backlog.json under a fresh
// temp root and returns the store plus the seed's checksum.
func seedMigrationRoot(t *testing.T, content string) (*BacklogStore, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "backlog.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("seed legacy file: %v", err)
	}
	return NewBacklogStore(path), sha256File(t, path)
}

// sha256File digests a file's bytes.
func sha256File(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

// AC-TOSQ-001 / REQ-TOSQ-011: the migrated database matches the legacy record
// item-for-item, findings included, with last_seq intact. Each parity axis is
// its own subtest so a failure names which axis broke rather than reporting a
// single opaque mismatch.
func TestMigrationParity(t *testing.T) {
	t.Parallel()
	store, _ := seedMigrationRoot(t, migrationFixture)

	var source BacklogRecord
	if err := json.Unmarshal([]byte(migrationFixture), &source); err != nil {
		t.Fatalf("fixture is not valid JSON: %v", err)
	}
	normalizeBacklogRecord(&source)

	got, err := store.Load()
	if err != nil {
		t.Fatalf("Load (triggers migration) = %v, want nil", err)
	}

	t.Run("item count", func(t *testing.T) {
		if len(got.Items) != len(source.Items) {
			t.Fatalf("items = %d, want %d", len(got.Items), len(source.Items))
		}
	})

	t.Run("per-item field equality", func(t *testing.T) {
		for i := range source.Items {
			w, g := source.Items[i], got.Items[i]
			if w.ID != g.ID || w.Text != g.Text || w.AddedAt != g.AddedAt || w.State != g.State {
				t.Errorf("item %d: got %+v, want %+v", i, g, w)
			}
		}
	})

	t.Run("spec_id null shape", func(t *testing.T) {
		// nil and a pointer to "" are different records. A migration that
		// conflated them would invent a picked card with an empty spec.
		for i := range source.Items {
			w, g := source.Items[i].SpecID, got.Items[i].SpecID
			if (w == nil) != (g == nil) {
				t.Fatalf("item %d (%s): spec_id null-shape changed: %v -> %v",
					i, source.Items[i].ID, derefSpecID(w), derefSpecID(g))
			}
			if w != nil && *w != *g {
				t.Errorf("item %d: spec_id %q != %q", i, *w, *g)
			}
		}
	})

	t.Run("insertion order preserved", func(t *testing.T) {
		// AC-TOSQ-017: the fixture's ids are deliberately NOT in numeric order
		// (t7, t3, t11, t2) — array position is the contract, not id order.
		var gotIDs, wantIDs []string
		for _, it := range got.Items {
			gotIDs = append(gotIDs, it.ID)
		}
		for _, it := range source.Items {
			wantIDs = append(wantIDs, it.ID)
		}
		if strings.Join(gotIDs, ",") != strings.Join(wantIDs, ",") {
			t.Fatalf("order = %v, want %v (REQ-TOSQ-004)", gotIDs, wantIDs)
		}
	})

	t.Run("findings order and tuples", func(t *testing.T) {
		if len(got.Findings) != len(source.Findings) {
			t.Fatalf("findings = %d, want %d", len(got.Findings), len(source.Findings))
		}
		for i := range source.Findings {
			if got.Findings[i] != source.Findings[i] {
				t.Errorf("finding %d: got %+v, want %+v", i, got.Findings[i], source.Findings[i])
			}
		}
	})

	t.Run("last_seq above max-present survives", func(t *testing.T) {
		// The fixture's mark (42) is above max-present (11). A migration that
		// dropped the mark and re-derived it would silently reuse ids 12..42.
		if got.LastSeq != 42 {
			t.Fatalf("last_seq = %d, want 42", got.LastSeq)
		}
	})

	t.Run("physical schema present", func(t *testing.T) {
		eng, err := openBacklogEngine(backlogSQLitePath(store.Path()))
		if err != nil {
			t.Fatalf("open migrated database: %v", err)
		}
		defer func() { _ = eng.close() }()
		version, err := eng.schemaVersion(t.Context())
		if err != nil {
			t.Fatalf("schemaVersion: %v", err)
		}
		if version != backlogSchemaVersion {
			t.Errorf("schema_version = %q, want %q", version, backlogSchemaVersion)
		}
		lastSeq, err := eng.readLastSeq(t.Context())
		if err != nil {
			t.Fatalf("readLastSeq: %v", err)
		}
		if lastSeq != 42 {
			t.Errorf("meta.last_seq = %d, want 42", lastSeq)
		}
	})
}

// AC-TOSQ-002 / REQ-TOSQ-005: the first id issued after the cutover continues
// past the persisted mark. Never reused, never reset to max-present.
func TestMigrationIDContinuity(t *testing.T) {
	t.Parallel()
	store, _ := seedMigrationRoot(t, migrationFixture)

	item, _, err := store.Add("probe")
	if err != nil {
		t.Fatalf("Add across the cutover: %v", err)
	}
	if item.ID != "t43" {
		t.Fatalf("issued id = %q, want t43 (last_seq 42 + 1)", item.ID)
	}

	second, _, err := store.Add("probe two")
	if err != nil {
		t.Fatalf("second Add: %v", err)
	}
	if second.ID != "t44" {
		t.Fatalf("second issued id = %q, want t44", second.ID)
	}
}

// AC-TOSQ-003 / REQ-TOSQ-014: the legacy file is RENAMED to its .migrated
// quarantine, byte-identical to the seed. Never deleted, never rewritten.
func TestMigrationQuarantinesLegacyByteIdentical(t *testing.T) {
	t.Parallel()
	store, seedSum := seedMigrationRoot(t, migrationFixture)

	if _, err := store.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}

	if _, err := os.Stat(store.Path()); !os.IsNotExist(err) {
		t.Errorf("legacy backlog.json still present after migration (err=%v), want it renamed away", err)
	}
	quarantine := store.Path() + backlogMigratedSuffix
	if got := sha256File(t, quarantine); got != seedSum {
		t.Fatalf("quarantine sha256 %s != seed %s — the quarantine must be byte-preserved", got, seedSum)
	}
}

// AC-TOSQ-004 / REQ-TOSQ-013: both layouts present. The DATABASE is served,
// and what happens to the json depends on which of the two ways it got there.
//
// A database beside a backlog.json is ambiguous, and the ambiguity is
// load-bearing: the json is either pre-cutover legacy stranded by a crash
// between the migration commit and the quarantine, or a downgrade export the
// operator just asked for. Quarantining the second silently deletes the only
// artifact a downgraded release can read. The in-flight marker written inside
// the migration is what separates them, so both directions are asserted here —
// asserting either alone would pass a store that ignored the marker entirely.
func TestMigrationBothPresentPrefersDatabase(t *testing.T) {
	t.Parallel()

	// Reconstruct state D FAITHFULLY: a crash in that window leaves the legacy
	// file AND the in-flight marker. Renaming the quarantine back without the
	// marker reconstructs a different state — the export case below.
	reconstructCrashedCutover := func(t *testing.T, store *BacklogStore) {
		t.Helper()
		if err := os.Rename(store.Path()+backlogMigratedSuffix, store.Path()); err != nil {
			t.Fatalf("restore legacy file: %v", err)
		}
		eng, err := openBacklogEngine(backlogSQLitePath(store.Path()))
		if err != nil {
			t.Fatalf("open db to set the in-flight marker: %v", err)
		}
		if err := eng.markQuarantinePending(t.Context()); err != nil {
			t.Fatalf("set in-flight marker: %v", err)
		}
		if err := eng.close(); err != nil {
			t.Fatalf("close: %v", err)
		}
	}

	t.Run("interrupted migration completes its quarantine", func(t *testing.T) {
		t.Parallel()
		store, _ := seedMigrationRoot(t, migrationFixture)
		if _, err := store.Load(); err != nil {
			t.Fatalf("initial Load: %v", err)
		}
		if _, _, err := store.Add("post-cutover card"); err != nil {
			t.Fatalf("Add: %v", err)
		}
		reconstructCrashedCutover(t, store)

		rec, err := store.Load()
		if err != nil {
			t.Fatalf("Load(state D) = %v, want nil — the retry must not fail the call", err)
		}
		// The database carries 5 items (4 seeded + 1 added); the restored
		// legacy file carries 4. Reading 4 would mean the stale JSON won.
		if len(rec.Items) != 5 {
			t.Fatalf("items = %d, want 5 — the database must stay authoritative", len(rec.Items))
		}
		if _, err := os.Stat(store.Path()); !os.IsNotExist(err) {
			t.Errorf("legacy file still present (err=%v) — the interrupted quarantine should have completed", err)
		}
	})

	t.Run("a downgrade export is left exactly where it was put", func(t *testing.T) {
		t.Parallel()
		// A FRESH root, deliberately: no legacy file ever existed here, so no
		// .migrated quarantine exists either. That matters — an earlier form of
		// this subtest reused a migrated root, where the "never overwrite an
		// existing quarantine" guard silently did the work and the subtest
		// passed even with the marker check removed. It asserted a mechanism it
		// was not exercising.
		store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
		if _, _, err := store.Add("a card"); err != nil {
			t.Fatalf("seed Add: %v", err)
		}
		if _, err := os.Stat(store.Path() + backlogMigratedSuffix); err == nil {
			t.Fatal("fixture invalid: a quarantine exists, so this would not exercise the marker")
		}

		// The export: a backlog.json written AFTER the store was already a
		// database, so no in-flight marker is set.
		exported := []byte(`{"version":1,"last_seq":1,"items":[],"findings":[]}`)
		if err := os.WriteFile(store.Path(), exported, 0o644); err != nil {
			t.Fatalf("write export: %v", err)
		}

		rec, err := store.Load()
		if err != nil {
			t.Fatalf("Load(db + export) = %v, want nil", err)
		}
		if len(rec.Items) != 1 {
			t.Fatalf("items = %d, want 1 — the database must stay authoritative", len(rec.Items))
		}
		got, err := os.ReadFile(store.Path())
		if err != nil {
			t.Fatalf("the export was renamed or removed: %v", err)
		}
		if string(got) != string(exported) {
			t.Error("the export's contents changed")
		}
		if _, err := os.Stat(store.Path() + backlogMigratedSuffix); err == nil {
			t.Error("the export was quarantined as if it were pre-cutover legacy")
		}
	})
}

// REQ-TOSQ-014 / C-6: an existing quarantine is never overwritten. The bytes
// the first migration preserved are the rollback source; a second rename onto
// them would destroy the one artifact a downgrade depends on.
func TestMigrationNeverOverwritesExistingQuarantine(t *testing.T) {
	t.Parallel()
	store, _ := seedMigrationRoot(t, migrationFixture)
	if _, err := store.Load(); err != nil {
		t.Fatalf("initial Load: %v", err)
	}
	quarantine := store.Path() + backlogMigratedSuffix
	firstSum := sha256File(t, quarantine)

	// A different legacy file reappears beside the live database.
	if err := os.WriteFile(store.Path(), []byte(`{"version":1,"items":[],"findings":[]}`), 0o644); err != nil {
		t.Fatalf("write intruding legacy file: %v", err)
	}
	if _, err := store.Load(); err != nil {
		t.Fatalf("Load(state D with existing quarantine): %v", err)
	}

	if got := sha256File(t, quarantine); got != firstSum {
		t.Fatalf("quarantine sha256 changed %s -> %s — the original rollback source was destroyed", firstSum, got)
	}
	if _, err := os.Stat(store.Path()); err != nil {
		t.Errorf("the intruding file was removed (err=%v); it must be left visible, never eaten", err)
	}
}

// AC-TOSQ-005 / REQ-TOSQ-012 / C-6: a malformed legacy file aborts the
// cutover. No database is left behind, and the seed's bytes are unchanged —
// the caller keeps serving the queue it had.
func TestMigrationMalformedAbortsWithoutDestroying(t *testing.T) {
	t.Parallel()
	store, seedSum := seedMigrationRoot(t, `{"version":1,"items":[ TRUNCATED`)

	if _, err := store.Load(); err == nil {
		t.Fatal("Load(malformed legacy) err = nil, want a parse error")
	}
	if _, _, err := store.Add("into malformed"); err == nil {
		t.Fatal("Add(malformed legacy) err = nil, want a parse error")
	}
	if err := store.Mutate(func(*BacklogRecord) error { return nil }); err == nil {
		t.Fatal("Mutate(malformed legacy) err = nil, want a parse error")
	}

	if got := sha256File(t, store.Path()); got != seedSum {
		t.Fatalf("malformed seed changed %s -> %s — never repair-by-write", seedSum, got)
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		artifact := backlogSQLitePath(store.Path()) + suffix
		if _, err := os.Stat(artifact); err == nil {
			t.Errorf("partial artifact %s survived the aborted migration", filepath.Base(artifact))
		}
	}
	if _, err := os.Stat(store.Path() + backlogMigratedSuffix); err == nil {
		t.Error("the legacy file was quarantined despite the migration failing")
	}
}

// AC-TOSQ-005 / REQ-TOSQ-012: a legacy file that FAILS PART-WAY THROUGH the
// cutover — after the database exists — leaves no partial artifact behind and
// the legacy file authoritative.
//
// This case is distinct from the malformed-seed case above, and the difference
// is the whole point: a malformed seed fails at the parse, before any database
// is created, so it proves nothing about the cleanup path. A legacy file
// carrying a DUPLICATE id parses fine, reaches the insert, and trips UNIQUE(id)
// with the database already on disk — which is the only way to actually
// exercise the removal of a partial artifact.
func TestMigrationPartialFailureRemovesArtifacts(t *testing.T) {
	t.Parallel()
	const duplicateIDSeed = `{
  "version": 1,
  "last_seq": 5,
  "items": [
    {"id": "t1", "text": "original", "added_at": "2026-01-02T03:04:05Z", "spec_id": null, "state": "queued"},
    {"id": "t1", "text": "collision", "added_at": "2026-01-02T03:04:06Z", "spec_id": null, "state": "queued"}
  ],
  "findings": []
}
`
	store, seedSum := seedMigrationRoot(t, duplicateIDSeed)

	_, err := store.Load()
	if err == nil {
		t.Fatal("Load(duplicate-id legacy) err = nil, want the cutover to abort")
	}
	if !IsBacklogIDConflict(err) {
		t.Errorf("err = %v, want it to satisfy IsBacklogIDConflict", err)
	}

	for _, suffix := range []string{"", "-wal", "-shm"} {
		artifact := backlogSQLitePath(store.Path()) + suffix
		if _, statErr := os.Stat(artifact); statErr == nil {
			t.Errorf("partial artifact %s survived the aborted cutover", filepath.Base(artifact))
		}
	}
	if got := sha256File(t, store.Path()); got != seedSum {
		t.Fatalf("legacy seed changed %s -> %s — it must stay authoritative and untouched", seedSum, got)
	}
	if _, statErr := os.Stat(store.Path() + backlogMigratedSuffix); statErr == nil {
		t.Error("the legacy file was quarantined despite the cutover failing")
	}
}

// REQ-TOSQ-012: a non-conforming id in the legacy data aborts the cutover
// rather than being silently normalized. The mark is derived from t<N> ids, so
// importing an id the parser cannot read would leave the high-water mark
// unable to describe the queue it guards.
func TestMigrationParityCatchesTamperedRecord(t *testing.T) {
	t.Parallel()

	// Fire the parity check itself: hand assertBacklogParity two records that
	// differ on each axis in turn and confirm every axis is actually compared.
	base := func() *BacklogRecord {
		spec := "SPEC-EXAMPLE-001"
		return &BacklogRecord{
			Version: 1, LastSeq: 9,
			Items: []BacklogItem{
				{ID: "t1", Text: "a", AddedAt: "T", SpecID: nil, State: BacklogStateQueued},
				{ID: "t2", Text: "b", AddedAt: "T", SpecID: &spec, State: BacklogStatePicked},
			},
			Findings: []BacklogFinding{{SubjectID: "t1", RelatedID: "t2", Relation: "contains", Source: "agent", At: "T"}},
		}
	}
	if err := assertBacklogParity(base(), base()); err != nil {
		t.Fatalf("identical records reported a mismatch: %v", err)
	}

	cases := map[string]func(*BacklogRecord){
		"dropped item":     func(r *BacklogRecord) { r.Items = r.Items[:1] },
		"reordered items":  func(r *BacklogRecord) { r.Items[0], r.Items[1] = r.Items[1], r.Items[0] },
		"altered text":     func(r *BacklogRecord) { r.Items[0].Text = "tampered" },
		"altered state":    func(r *BacklogRecord) { r.Items[0].State = BacklogStateDropped },
		"spec_id nulled":   func(r *BacklogRecord) { r.Items[1].SpecID = nil },
		"spec_id invented": func(r *BacklogRecord) { s := ""; r.Items[0].SpecID = &s },
		"dropped finding":  func(r *BacklogRecord) { r.Findings = nil },
		"altered finding":  func(r *BacklogRecord) { r.Findings[0].Relation = "absorbs" },
		"last_seq lowered": func(r *BacklogRecord) { r.LastSeq = 2 },
		"version bumped":   func(r *BacklogRecord) { r.Version = 2 },
	}
	names := make([]string, 0, len(cases))
	for name := range cases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			tampered := base()
			cases[name](tampered)
			if err := assertBacklogParity(base(), tampered); err == nil {
				t.Fatalf("parity check passed a %s — the cutover would have flipped authority onto corrupt data", name)
			}
		})
	}
}

// REQ-TOSQ-009: the statusline's count path is PURE — it reads whichever
// layout it finds and moves no bytes. A statusline render migrating the
// operator's queue would put a per-second background process in charge of a
// one-time irreversible cutover.
func TestQueuedCountIsPureAcrossBothLayouts(t *testing.T) {
	t.Parallel()

	t.Run("legacy layout is read without migrating", func(t *testing.T) {
		store, seedSum := seedMigrationRoot(t, migrationFixture)
		if n := store.QueuedCount(); n != 2 {
			t.Fatalf("QueuedCount = %d, want 2", n)
		}
		if _, err := os.Stat(backlogSQLitePath(store.Path())); err == nil {
			t.Error("QueuedCount created a database — the read path must move no bytes")
		}
		if got := sha256File(t, store.Path()); got != seedSum {
			t.Error("QueuedCount altered the legacy file")
		}
	})

	t.Run("database layout is read through the aggregate", func(t *testing.T) {
		store, _ := seedMigrationRoot(t, migrationFixture)
		if _, err := store.Load(); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		if n := store.QueuedCount(); n != 2 {
			t.Fatalf("QueuedCount(db) = %d, want 2", n)
		}
	})

	t.Run("absent queue counts zero", func(t *testing.T) {
		store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
		if n := store.QueuedCount(); n != 0 {
			t.Fatalf("QueuedCount(absent) = %d, want 0", n)
		}
	})
}

// REQ-TOSQ-010 / REQ-WTQ-001: LoadPure never moves bytes on ANY branch. This
// is the guard the console and the statusline rest on — a read-only surface
// that migrated would perform the one-time irreversible cutover from a page
// render, at whatever moment someone happened to open the page.
func TestLoadPureNeverMovesBytes(t *testing.T) {
	t.Parallel()

	t.Run("legacy layout is served in place", func(t *testing.T) {
		store, seedSum := seedMigrationRoot(t, migrationFixture)

		rec, err := store.LoadPure()
		if err != nil {
			t.Fatalf("LoadPure(legacy) = %v, want nil", err)
		}
		if len(rec.Items) != 4 {
			t.Fatalf("items = %d, want 4", len(rec.Items))
		}
		if rec.LastSeq != 42 {
			t.Errorf("last_seq = %d, want 42", rec.LastSeq)
		}
		if _, err := os.Stat(backlogSQLitePath(store.Path())); err == nil {
			t.Error("LoadPure created a database — it must migrate nothing")
		}
		if _, err := os.Stat(store.Path() + backlogMigratedSuffix); err == nil {
			t.Error("LoadPure quarantined the legacy file — it must move nothing")
		}
		if got := sha256File(t, store.Path()); got != seedSum {
			t.Error("LoadPure altered the legacy file")
		}
	})

	t.Run("database layout is served in place", func(t *testing.T) {
		store, _ := seedMigrationRoot(t, migrationFixture)
		if _, err := store.Load(); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatalf("LoadPure(db) = %v, want nil", err)
		}
		if len(rec.Items) != 4 {
			t.Fatalf("items = %d, want 4", len(rec.Items))
		}
	})

	t.Run("state D leaves both artifacts untouched", func(t *testing.T) {
		store, _ := seedMigrationRoot(t, migrationFixture)
		if _, err := store.Load(); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		// Reconstruct state D: the legacy file reappears beside the database.
		if err := os.Rename(store.Path()+backlogMigratedSuffix, store.Path()); err != nil {
			t.Fatalf("reconstruct state D: %v", err)
		}
		if _, err := store.LoadPure(); err != nil {
			t.Fatalf("LoadPure(state D) = %v, want nil", err)
		}
		if _, err := os.Stat(store.Path()); err != nil {
			t.Errorf("LoadPure re-quarantined the legacy file (err=%v) — a pure read must not", err)
		}
	})

	t.Run("absent queue reads as empty, not as an error", func(t *testing.T) {
		store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
		rec, err := store.LoadPure()
		if err != nil {
			t.Fatalf("LoadPure(absent) = %v, want nil", err)
		}
		if len(rec.Items) != 0 || rec.Version != 1 {
			t.Fatalf("empty record = %+v, want version 1 with no items", rec)
		}
		if _, err := os.Stat(backlogSQLitePath(store.Path())); err == nil {
			t.Error("LoadPure created a database for an absent queue")
		}
	})
}

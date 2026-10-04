// todo_stale_store_test.go — SPEC-TODO-STALE-STORE-001 (card t1307): the
// read-only stale-local-store detector (REQ-TSS-004) shared by the stderr
// disclosure (REQ-TSS-001) and the doctor divergence check (REQ-TSS-010).
//
// The fixture reproduces the measured incident this SPEC exists for: a home
// database at last_seq 1305 answering every read while a rollback-snapshot
// project-local store sits at last_seq 661 (a 644-seq gap the 2026-09-29
// misread fell into). Every test isolates both stores under t.TempDir() —
// no test here may touch the operator's real home database.
package factory

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modu-ai/moai-adk/internal/paths"
)

// staleStoreFixture builds a project root whose home database (resolved
// through the MOAI_HOME override) carries last_seq 1305 while the
// project-local legacy store carries last_seq 661. Both directories are
// temp dirs; the override is absolute, so StateDirForRoot resolves the
// home layout exactly as production does on a real project.
func staleStoreFixture(t *testing.T) (root, homeDB, legacyDB string) {
	t.Helper()
	root = t.TempDir()
	t.Setenv(paths.EnvHome, t.TempDir())
	homeDB = backlogSQLitePath(filepath.Join(StateDirForRoot(root), backlogFileName))
	legacyDB = backlogSQLitePath(filepath.Join(projectStateDirForRoot(root), backlogFileName))
	seedStaleStoreSeq(t, homeDB, 1305)
	seedStaleStoreSeq(t, legacyDB, 661)
	return root, homeDB, legacyDB
}

// seedStaleStoreSeq creates a fully-shaped queue database at dbPath (via
// the store's own add path, so schema and meta are production-shaped) and
// stamps meta.last_seq to seq through the store's own mutation path.
func seedStaleStoreSeq(t *testing.T, dbPath string, seq int) {
	t.Helper()
	store := NewBacklogStore(strings.TrimSuffix(dbPath, ".db") + ".json")
	if _, _, err := store.Add("seed card"); err != nil {
		t.Fatalf("seed store %s: %v", dbPath, err)
	}
	err := store.Mutate(func(rec *BacklogRecord) error {
		rec.LastSeq = seq
		return nil
	})
	if err != nil {
		t.Fatalf("stamp last_seq %d on %s: %v", seq, dbPath, err)
	}
}

// dbFingerprint returns the sha256 and mtime of one database file, plus
// the sorted entry names of its directory — the triple AC-TSS-003's
// read-only purity judgment is stated over.
func dbFingerprint(t *testing.T, dbPath string) (string, int64, []string) {
	t.Helper()
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat %s: %v", dbPath, err)
	}
	raw, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read %s: %v", dbPath, err)
	}
	sum := sha256.Sum256(raw)
	entries, err := os.ReadDir(filepath.Dir(dbPath))
	if err != nil {
		t.Fatalf("read dir %s: %v", filepath.Dir(dbPath), err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return hex.EncodeToString(sum[:]), info.ModTime().UnixNano(), names
}

func TestInspectStaleLocalStores_DivergentSeqs(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)

	fact := InspectStaleLocalStores(root)

	if !fact.HomePresent {
		t.Fatalf("HomePresent = false; the home database at %s was seeded", homeDB)
	}
	if fact.HomeLastSeq != 1305 {
		t.Errorf("HomeLastSeq = %d, want 1305", fact.HomeLastSeq)
	}
	if !fact.Divergent {
		t.Fatalf("Divergent = false; home 1305 vs legacy 661 is the divergence this detector exists to find")
	}
	if len(fact.Stores) != 1 {
		t.Fatalf("len(Stores) = %d, want 1", len(fact.Stores))
	}
	got := fact.Stores[0]
	if got.Path != legacyDB {
		t.Errorf("store path = %q, want %q", got.Path, legacyDB)
	}
	if got.LastSeq != 661 {
		t.Errorf("store last_seq = %d, want 661", got.LastSeq)
	}
	if !got.Readable {
		t.Errorf("store Readable = false; a production-shaped store must read")
	}
}

func TestInspectStaleLocalStores_EqualSeqsIsNotDivergent(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)
	seedStaleStoreSeq(t, legacyDB, 1305) // match the home sequence

	fact := InspectStaleLocalStores(root)

	if fact.Divergent {
		t.Errorf("Divergent = true with equal last_seq values (home %d, legacy %d); REQ-TSS-005 requires silence",
			fact.HomeLastSeq, fact.Stores[0].LastSeq)
	}
	_ = homeDB
	_ = legacyDB
}

func TestInspectStaleLocalStores_NoLegacyStore(t *testing.T) {
	root := t.TempDir()
	t.Setenv(paths.EnvHome, t.TempDir())
	homeDB := backlogSQLitePath(filepath.Join(StateDirForRoot(root), backlogFileName))
	seedStaleStoreSeq(t, homeDB, 1305)

	fact := InspectStaleLocalStores(root)

	if len(fact.Stores) != 0 {
		t.Errorf("Stores = %v, want none", fact.Stores)
	}
	if fact.Divergent {
		t.Error("Divergent = true with no legacy store present")
	}
}

func TestInspectStaleLocalStores_ZeroByteStoreIsUnreadableNotDivergent(t *testing.T) {
	root, _, legacyDB := staleStoreFixture(t)
	if err := os.Truncate(legacyDB, 0); err != nil {
		t.Fatalf("truncate legacy store: %v", err)
	}

	fact := InspectStaleLocalStores(root)

	if len(fact.Stores) != 1 {
		t.Fatalf("len(Stores) = %d, want 1 (the zero-byte store still exists)", len(fact.Stores))
	}
	if fact.Stores[0].Readable {
		t.Error("Readable = true for a zero-byte store; an unreadable store must be reported as unreadable, not divergent (acceptance §D.2)")
	}
	if fact.Divergent {
		t.Error("Divergent = true on an unreadable store; divergence requires a readable last_seq")
	}
}

func TestInspectStaleLocalStores_BothLegacyDirsReported(t *testing.T) {
	root, _, _ := staleStoreFixture(t)
	factoryDirDB := backlogSQLitePath(filepath.Join(LegacyStateDirForRoot(root), backlogFileName))
	seedStaleStoreSeq(t, factoryDirDB, 500)

	fact := InspectStaleLocalStores(root)

	if len(fact.Stores) != 2 {
		t.Fatalf("len(Stores) = %d, want 2 (todo + kanban legacy dirs)", len(fact.Stores))
	}
	seen := map[string]int{}
	for _, st := range fact.Stores {
		seen[st.Path] = st.LastSeq
	}
	if seen[factoryDirDB] != 500 {
		t.Errorf("kanban-dir store last_seq = %d, want 500", seen[factoryDirDB])
	}
	if !fact.Divergent {
		t.Error("Divergent = false with two stale stores present")
	}
}

func TestInspectStaleLocalStores_ReadOnly(t *testing.T) {
	root, homeDB, legacyDB := staleStoreFixture(t)

	homeSum, homeMtime, homeNames := dbFingerprint(t, homeDB)
	legacySum, legacyMtime, legacyNames := dbFingerprint(t, legacyDB)

	_ = InspectStaleLocalStores(root)

	if gotSum, gotMtime, gotNames := dbFingerprint(t, homeDB); gotSum != homeSum || gotMtime != homeMtime || strings.Join(gotNames, ",") != strings.Join(homeNames, ",") {
		t.Errorf("home database changed across the probe: sha %s->%s mtime %d->%d names %v->%v",
			homeSum, gotSum, homeMtime, gotMtime, homeNames, gotNames)
	}
	if gotSum, gotMtime, gotNames := dbFingerprint(t, legacyDB); gotSum != legacySum || gotMtime != legacyMtime || strings.Join(gotNames, ",") != strings.Join(legacyNames, ",") {
		t.Errorf("legacy store changed across the probe: sha %s->%s mtime %d->%d names %v->%v — REQ-TSS-003 forbids every write on the detection path",
			legacySum, gotSum, legacyMtime, gotMtime, legacyNames, gotNames)
	}
}

// ghostFixture plants the measured ghost classes (SPEC-TODO-SURFACE-POLISH-001
// REQ-TSP-040) on top of the divergent-store fixture: a legacy backlog.json
// in the home queue directory and in the project-local todo directory, a
// .migrated quarantine in the legacy kanban directory, and UUID-shaped
// session records beside named artifacts the class must NOT swallow.
func ghostFixture(t *testing.T) (root string, ghosts map[string]GhostArtifact) {
	t.Helper()
	r, _, _ := staleStoreFixture(t)
	root = r
	ghosts = map[string]GhostArtifact{}

	homeJSON := filepath.Join(StateDirForRoot(root), backlogFileName)
	if err := os.WriteFile(homeJSON, []byte(strings.Repeat("h", 1884)), 0o600); err != nil {
		t.Fatalf("plant home backlog.json: %v", err)
	}
	ghosts[homeJSON] = GhostArtifact{Path: homeJSON, Class: GhostClassLegacyJSON, Bytes: 1884}

	localJSON := filepath.Join(projectStateDirForRoot(root), backlogFileName)
	if err := os.WriteFile(localJSON, []byte(strings.Repeat("l", 652)), 0o600); err != nil {
		t.Fatalf("plant project-local backlog.json: %v", err)
	}
	ghosts[localJSON] = GhostArtifact{Path: localJSON, Class: GhostClassLegacyJSON, Bytes: 652}

	migrated := filepath.Join(LegacyStateDirForRoot(root), backlogFileName+backlogMigratedSuffix)
	if err := os.MkdirAll(filepath.Dir(migrated), 0o755); err != nil {
		t.Fatalf("create legacy kanban dir: %v", err)
	}
	if err := os.WriteFile(migrated, []byte(strings.Repeat("m", 155)), 0o600); err != nil {
		t.Fatalf("plant .migrated: %v", err)
	}
	ghosts[migrated] = GhostArtifact{Path: migrated, Class: GhostClassMigratedJSON, Bytes: 155}

	for _, name := range []string{
		"0a0b0c0d-0000-4000-8000-000000000001.json",
		"0a0b0c0d-0000-4000-8000-000000000002.json",
	} {
		path := filepath.Join(projectStateDirForRoot(root), name)
		if err := os.WriteFile(path, []byte(`{"session_id":"x"}`), 0o600); err != nil {
			t.Fatalf("plant session record: %v", err)
		}
		ghosts[path] = GhostArtifact{Path: path, Class: GhostClassSessionRecord, Bytes: int64(len(`{"session_id":"x"}`))}
	}
	// Named artifacts sharing the directory are NOT session records.
	for _, name := range []string{"companions.json", "leads.json"} {
		if err := os.WriteFile(filepath.Join(projectStateDirForRoot(root), name), []byte("[]"), 0o600); err != nil {
			t.Fatalf("plant %s: %v", name, err)
		}
	}
	return root, ghosts
}

// AC-TSP-040 — every ghost class lands in the fact as its own entry with
// path, class, and byte size; the SQLite divergence facts are untouched by
// the walk; the probe writes nothing (sha unchanged) and swallows no named
// artifact into the session-record class.
func TestInspectStaleLocalStoresGhostClasses(t *testing.T) {
	root, want := ghostFixture(t)

	// Fingerprint every planted ghost before the probe.
	sums := map[string]string{}
	for path := range want {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		sum := sha256.Sum256(raw)
		sums[path] = hex.EncodeToString(sum[:])
	}

	fact := InspectStaleLocalStores(root)

	// The SQLite divergence facts survive the extension untouched.
	if !fact.Divergent {
		t.Error("Divergent = false after the ghost walk — the ghost extension must not contaminate the SQLite verdict (REQ-TSP-040)")
	}
	if len(fact.Stores) != 1 {
		t.Errorf("len(Stores) = %d, want 1", len(fact.Stores))
	}

	if len(fact.Ghosts) != len(want) {
		t.Fatalf("len(Ghosts) = %d (%v), want %d", len(fact.Ghosts), fact.Ghosts, len(want))
	}
	for _, got := range fact.Ghosts {
		w, ok := want[got.Path]
		if !ok {
			t.Errorf("unexpected ghost %q (class %q) — companions.json/leads.json must not be swallowed", got.Path, got.Class)
			continue
		}
		if got.Class != w.Class {
			t.Errorf("ghost %s class = %q, want %q", got.Path, got.Class, w.Class)
		}
		if got.Bytes != w.Bytes {
			t.Errorf("ghost %s bytes = %d, want %d", got.Path, got.Bytes, w.Bytes)
		}
		raw, err := os.ReadFile(got.Path)
		if err != nil {
			t.Fatalf("re-read %s: %v", got.Path, err)
		}
		sum := sha256.Sum256(raw)
		if hex.EncodeToString(sum[:]) != sums[got.Path] {
			t.Errorf("ghost %s changed across the probe — the detector is read-only (REQ-TSP-040)", got.Path)
		}
	}
}

// The canonical directory's backlog.json without a sibling engine database
// is the LIVE pre-SQLite JSON queue, not a ghost — calling it one would name
// the answering store stale.
func TestInspectStaleLocalStores_LiveJSONQueueIsNotAGhost(t *testing.T) {
	root := t.TempDir()
	t.Setenv(paths.EnvHome, t.TempDir())
	// Seed NO home database: the canonical dir resolves project-local
	// (temporary origin) or stays home-shaped; either way the queue JSON we
	// plant is the only store and no engine .db sits beside it.
	canonical := StateDirForRoot(root)
	liveJSON := filepath.Join(canonical, backlogFileName)
	if err := os.MkdirAll(filepath.Dir(liveJSON), 0o755); err != nil {
		t.Fatalf("create canonical state dir: %v", err)
	}
	if err := os.WriteFile(liveJSON, []byte(`{"version":1,"items":[]}`), 0o600); err != nil {
		t.Fatalf("plant live backlog.json: %v", err)
	}

	fact := InspectStaleLocalStores(root)

	for _, g := range fact.Ghosts {
		if g.Class == GhostClassLegacyJSON {
			t.Errorf("canonical backlog.json without a sibling engine database reported as %q ghost at %s — it is the live legacy-format queue", GhostClassLegacyJSON, g.Path)
		}
	}
}

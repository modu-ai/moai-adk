// todo_stale_store_test.go — SPEC-TODO-STALE-STORE-001 (card t1307): the
// read-only stale-local-store detector (REQ-TSS-004) shared by the stderr
// disclosure (REQ-TSS-001) and the doctor divergence check (REQ-TSS-010).
//
// The fixture reproduces the measured incident this SPEC exists for: a home
// database at last_seq 1305 answering every read while a rollback-snapshot
// project-local store sits at last_seq 661 (a 644-seq gap the 2026-09-29
// misread fell into). Every test isolates both stores under t.TempDir() —
// no test here may touch the operator's real home database.
package kanban

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
	kanbanDirDB := backlogSQLitePath(filepath.Join(LegacyStateDirForRoot(root), backlogFileName))
	seedStaleStoreSeq(t, kanbanDirDB, 500)

	fact := InspectStaleLocalStores(root)

	if len(fact.Stores) != 2 {
		t.Fatalf("len(Stores) = %d, want 2 (todo + kanban legacy dirs)", len(fact.Stores))
	}
	seen := map[string]int{}
	for _, st := range fact.Stores {
		seen[st.Path] = st.LastSeq
	}
	if seen[kanbanDirDB] != 500 {
		t.Errorf("kanban-dir store last_seq = %d, want 500", seen[kanbanDirDB])
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

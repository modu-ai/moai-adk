package factory

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
	"github.com/modu-ai/moai-adk/internal/testguard"
	"golang.org/x/sys/windows"
)

// File.Stat captures the Windows file ID now, rather than lazily opening the
// path during a later SameFile call after SQLite may have replaced a sidecar.
func captureBacklogPrivateArtifacts(t *testing.T, dbPath string, before map[string]os.FileInfo) map[string]os.FileInfo {
	t.Helper()
	if err := privatefile.CheckDirectory(filepath.Dir(dbPath)); err != nil {
		t.Fatal(err)
	}
	now := make(map[string]os.FileInfo)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		path := dbPath + suffix
		if err := privatefile.Check(path); err != nil {
			t.Fatalf("native policy %s: %v", suffix, err)
		}
		f, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		info, statErr := f.Stat()
		closeErr := f.Close()
		if statErr != nil {
			t.Fatal(statErr)
		}
		if closeErr != nil {
			t.Fatal(closeErr)
		}
		if old := before[suffix]; old != nil && !os.SameFile(old, info) {
			t.Fatalf("SQLite changed artifact identity %s", suffix)
		}
		now[suffix] = info
	}
	return now
}

func TestBacklogPhysicalReconnectPreservesNativeArtifactIdentity(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "managed-queue", "backlog.db")
	eng, err := openBacklogEngine(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func(eng *backlogEngine) {
		if err := eng.close(); err != nil {
			t.Error(err)
		}
	}(eng)
	if err := eng.writeRecord(t.Context(), &BacklogRecord{Version: 1, Items: []BacklogItem{}, Findings: []BacklogFinding{}, Archived: []BacklogArchiveEntry{}}); err != nil {
		t.Fatal(err)
	}
	original := captureBacklogPrivateArtifacts(t, dbPath, nil)
	if err := eng.close(); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, dbPath, original)
	eng, err = openBacklogEngine(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func(eng *backlogEngine) {
		if err := eng.close(); err != nil {
			t.Error(err)
		}
	}(eng)
	if _, err := eng.readRecord(t.Context()); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, dbPath, original)
	if err := eng.close(); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, dbPath, original)
}

func TestGTDWritableConnectionPreservesNativeArtifactIdentity(t *testing.T) {
	store := newGTDTestStore(t)
	db, err := openGTDDB(store)
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}(db)
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	original := captureBacklogPrivateArtifacts(t, store.EnginePath(), nil)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, store.EnginePath(), original)
	db, err = openGTDDB(store)
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}(db)
	if err := db.PingContext(t.Context()); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, store.EnginePath(), original)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	captureBacklogPrivateArtifacts(t, store.EnginePath(), original)
}

func TestBacklogRefusesWeakNativeNamespaceBeforeCreatingDatabase(t *testing.T) {
	dir := privateBacklogFixtureDir(t)
	if err := privatefile.CheckDirectory(dir); err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, dir)
	if err := privatefile.CheckDirectory(dir); err == nil {
		t.Fatal("weak directory fixture was not weak")
	}
	dbPath := filepath.Join(dir, "backlog.db")
	eng, err := openBacklogEngine(dbPath)
	if err == nil {
		if err := eng.close(); err != nil {
			t.Error(err)
		}
		t.Fatal("weak namespace was accepted")
	}
	for _, suffix := range []string{"", "-wal", "-shm"} {
		if _, err := os.Lstat(dbPath + suffix); !os.IsNotExist(err) {
			t.Fatalf("refused namespace gained artifact %s: %v", suffix, err)
		}
	}
	if err := privatefile.CheckDirectory(dir); err == nil {
		t.Fatal("refused namespace policy was silently replaced")
	}
}

func TestRelocationRefusesWeakNativeTargetWithoutMutatingRollback(t *testing.T) {
	from := privateBacklogFixtureDir(t)
	to := privateBacklogFixtureDir(t)
	source := filepath.Join(from, backlogFileName)
	if err := privatefile.WriteFile(source, []byte(migrationFixture)); err != nil {
		t.Fatal(err)
	}
	testguard.MakeFilePublic(t, to)
	getPolicy := func() string {
		sd, err := windows.GetNamedSecurityInfo(to, windows.SE_FILE_OBJECT, windows.OWNER_SECURITY_INFORMATION|windows.DACL_SECURITY_INFORMATION)
		if err != nil {
			t.Fatal(err)
		}
		return sd.String()
	}
	beforePolicy := getPolicy()
	if err := relocateQueueArtifacts(from, to); err == nil {
		t.Fatal("relocation accepted weak existing namespace")
	}
	got, err := os.ReadFile(source)
	if err != nil || string(got) != migrationFixture {
		t.Fatalf("rollback source changed: %v", err)
	}
	if got := getPolicy(); got != beforePolicy {
		t.Fatal("refusal changed existing target owner or DACL")
	}
	for _, name := range []string{"backlog.db", "backlog.db-wal", "backlog.db-shm"} {
		if _, err := os.Lstat(filepath.Join(to, name)); !os.IsNotExist(err) {
			t.Fatalf("refusal created %s: %v", name, err)
		}
	}
	if _, err := os.Lstat(filepath.Join(from, backlogRetiredFileName)); !os.IsNotExist(err) {
		t.Fatalf("refused relocation retired source: %v", err)
	}
}

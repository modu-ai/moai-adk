package homestate

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestHomeLayoutAndProjectRuntimePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("MOAI_HOME", home)
	if err := EnsureHomeLayout(); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"claude-profiles", "config", "credentials", "db", "cache/search", "run", "logs", "integrations", "bin", "releases", "backups", "reports"} {
		info, err := os.Stat(filepath.Join(home, filepath.FromSlash(rel)))
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Errorf("%s mode=%#o, want 0700", rel, got)
		}
	}
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := EnsureProjectLayout(root); err != nil {
		t.Fatal(err)
	}
	search, _ := SearchDBPath(root)
	run, _ := RunProjectDir(root)
	for _, path := range []string{filepath.Dir(search), run, filepath.Join(run, "locks"), filepath.Join(run, "sockets"), filepath.Join(run, "guard-liveness")} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Errorf("%s mode=%#o, want 0700", path, got)
		}
	}
}

func TestFactorySchemaAndResumeClaimOnce(t *testing.T) {
	root := t.TempDir()
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("OpenFactory: %v", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.SaveResume(context.Background(), ResumeHandoff{
		SchemaVersion: 1, SavedAt: time.Now(), DirectivesJSON: `{}`, Body: "resume",
	}); err != nil {
		t.Fatalf("SaveResume: %v", err)
	}

	var won atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			separate, openErr := OpenFactory(root)
			if openErr != nil {
				return
			}
			defer func() { _ = separate.Close() }()
			_, ok, claimErr := separate.ClaimPendingResume(context.Background(), "token")
			if claimErr == nil && ok {
				won.Add(1)
			}
		}()
	}
	wg.Wait()
	if got := won.Load(); got != 1 {
		t.Fatalf("claim winners = %d, want 1", got)
	}
	var claimedID int64
	if err := db.DB.QueryRow(`SELECT id FROM resume_handoffs WHERE status='claimed'`).Scan(&claimedID); err != nil {
		t.Fatalf("claimed row: %v", err)
	}
	if err := db.FinishResume(context.Background(), claimedID, "token", "consumed", "test"); err != nil {
		t.Fatalf("consume claim: %v", err)
	}
	var eventCount int
	if err := db.DB.QueryRow(`SELECT count(*) FROM handoff_events WHERE flow='resume' AND handoff_id=?`, claimedID).Scan(&eventCount); err != nil {
		t.Fatal(err)
	}
	if eventCount != 3 {
		t.Fatalf("resume audit events=%d, want 3 (pending, claimed, consumed)", eventCount)
	}

	for _, table := range []string{"workers", "runs", "cards", "events", "dead_letters", "resume_handoffs", "memory_handoffs", "handoff_events"} {
		var name string
		if err := db.DB.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name); err != nil {
			t.Errorf("table %s: %v", table, err)
		}
	}
}

func TestProjectLayoutPermissions(t *testing.T) {
	root := t.TempDir()
	if err := EnsureProjectLayout(root); err != nil {
		t.Fatalf("EnsureProjectLayout: %v", err)
	}
	dir, _ := ProjectDir(root)
	for _, path := range []string{dir, dir + "/todo", dir + "/factory"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Errorf("%s mode=%#o, want 0700", path, got)
		}
	}
}

func TestProjectDirHonorsExplicitMoaiHomeForTemporaryProject(t *testing.T) {
	home := t.TempDir()
	project := t.TempDir()
	t.Setenv("MOAI_HOME", home)

	got, err := ProjectDir(project)
	if err != nil {
		t.Fatalf("ProjectDir: %v", err)
	}
	want := filepath.Join(home, "db", ProjectKey(project))
	if got != want {
		t.Fatalf("ProjectDir() = %q, want explicit MOAI_HOME path %q", got, want)
	}
}

func TestFactorySQLiteArtifactsArePrivateWhileOpen(t *testing.T) {
	root := t.TempDir()
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatalf("OpenFactory: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.SaveResume(context.Background(), ResumeHandoff{
		SchemaVersion: 1, SavedAt: time.Now(), DirectivesJSON: `{}`, Body: "permissions",
	}); err != nil {
		t.Fatalf("SaveResume: %v", err)
	}

	for _, path := range []string{db.Path, db.Path + "-wal", db.Path + "-shm"} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
		if got := info.Mode().Perm(); got != 0o600 {
			t.Errorf("%s mode=%#o, want 0600", filepath.Base(path), got)
		}
	}
}

func TestImportLegacyWorkersRunsOnlyOnceAfterRosterBecomesEmpty(t *testing.T) {
	root := t.TempDir()
	legacy := filepath.Join(root, "workers.json")
	if err := os.WriteFile(legacy, []byte(`{"lane-1":{"pid":101,"registered_at":"2026-01-01T00:00:00Z"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if err := db.ImportLegacyWorkers(legacy); err != nil {
		t.Fatal(err)
	}
	if _, err := db.DB.Exec(`DELETE FROM workers`); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(legacy, []byte(`{"lane-2":{"pid":202,"registered_at":"2026-01-01T00:00:00Z"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := db.ImportLegacyWorkers(legacy); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.DB.QueryRow(`SELECT count(*) FROM workers`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("legacy workers reimported after roster emptied: count=%d", count)
	}
}

func TestExpireResumeIfPendingDoesNotExpireNewerReplacement(t *testing.T) {
	root := t.TempDir()
	db, err := OpenFactory(root)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	old := ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now().Add(-time.Hour), DirectivesJSON: `{}`, Body: "old"}
	if err := db.SaveResume(t.Context(), old); err != nil {
		t.Fatal(err)
	}
	row, ok, err := db.ReadPendingResume(t.Context())
	if err != nil || !ok {
		t.Fatalf("read old: ok=%v err=%v", ok, err)
	}
	if err := db.SaveResume(t.Context(), ResumeHandoff{SchemaVersion: 1, SavedAt: time.Now(), DirectivesJSON: `{}`, Body: "new"}); err != nil {
		t.Fatal(err)
	}
	expired, err := db.ExpireResumeIfPending(t.Context(), row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if expired {
		t.Fatal("stale observation expired a newer replacement")
	}
	current, ok, err := db.ReadPendingResume(t.Context())
	if err != nil || !ok || current.Body != "new" {
		t.Fatalf("new pending lost: row=%+v ok=%v err=%v", current, ok, err)
	}
}

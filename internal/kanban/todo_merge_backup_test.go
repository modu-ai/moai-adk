// todo_merge_backup_test.go — backup, restore, and ordering evidence for the
// queue merge (SPEC-TODO-QUEUE-HOME-MERGE-001 M2, plan.md §F).
//
// Every test runs on t.TempDir() fixture stores. The restore procedure is
// REHEARSED here (AC-TQM-007); it may never be exercised on real data.
package kanban

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mergeFixtureStore creates a store directory with a db-backed queue (one
// seeded card; the optional seed hook runs inside the seeding mutation) and
// returns the queue path plus the store dir.
func mergeFixtureStore(t *testing.T, seed func(s *BacklogStore)) (queuePath, dir string) {
	t.Helper()
	dir = t.TempDir()
	queuePath = filepath.Join(dir, backlogFileName)
	store := NewBacklogStore(queuePath)
	if err := store.Mutate(func(rec *BacklogRecord) error {
		if seed != nil {
			seed(store)
		}
		rec.Items = append(rec.Items, mergeItem("t1", "seed card", BacklogStateQueued))
		return nil
	}); err != nil {
		t.Fatalf("seed fixture store: %v", err)
	}
	return queuePath, dir
}

// TestSnapshotStoreDirStateCleanPasses proves the ordering evidence reports
// ZERO mutation on an untouched fixture store (the AC-TQM-001 green path).
func TestSnapshotStoreDirStateCleanPasses(t *testing.T) {
	_, dir := mergeFixtureStore(t, nil)

	before, err := SnapshotStoreDirState(dir)
	if err != nil {
		t.Fatalf("snapshot before: %v", err)
	}
	after, err := SnapshotStoreDirState(dir)
	if err != nil {
		t.Fatalf("snapshot after: %v", err)
	}
	if muts := StoreDirMutations(before, after); len(muts) != 0 {
		t.Errorf("untouched fixture reported mutations: %+v", muts)
	}
}

// TestStoreDirMutationsDetectsCorruption is the RED-FIRST failure probe: the
// comparator must catch a fixture store whose artifact changed between the
// two snapshots (added / removed / mtime-moved) — a corrupted window reads as
// a mutation, never as clean.
func TestStoreDirMutationsDetectsCorruption(t *testing.T) {
	t.Run("added artifact", func(t *testing.T) {
		_, dir := mergeFixtureStore(t, nil)
		before, _ := SnapshotStoreDirState(dir)
		if err := os.WriteFile(filepath.Join(dir, "sneaky.json"), []byte("{}"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
		after, _ := SnapshotStoreDirState(dir)
		muts := StoreDirMutations(before, after)
		if len(muts) != 1 || muts[0].Kind != StoreMutationAdded {
			t.Errorf("added artifact not detected: %+v", muts)
		}
	})
	t.Run("removed artifact", func(t *testing.T) {
		_, dir := mergeFixtureStore(t, nil)
		before, _ := SnapshotStoreDirState(dir)
		entries, _ := os.ReadDir(dir)
		removed := false
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".db") {
				if err := os.Remove(filepath.Join(dir, e.Name())); err != nil {
					t.Fatalf("remove: %v", err)
				}
				removed = true
			}
		}
		if !removed {
			t.Fatalf("fixture store has no .db artifact to remove")
		}
		after, _ := SnapshotStoreDirState(dir)
		muts := StoreDirMutations(before, after)
		if len(muts) != 1 || muts[0].Kind != StoreMutationRemoved {
			t.Errorf("removed artifact not detected: %+v", muts)
		}
	})
	t.Run("modified artifact", func(t *testing.T) {
		_, dir := mergeFixtureStore(t, nil)
		before, _ := SnapshotStoreDirState(dir)
		future := time.Now().Add(2 * time.Hour)
		entries, _ := os.ReadDir(dir)
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".db") {
				_ = os.Chtimes(filepath.Join(dir, e.Name()), future, future)
			}
		}
		after, _ := SnapshotStoreDirState(dir)
		muts := StoreDirMutations(before, after)
		if len(muts) == 0 {
			t.Fatalf("mtime move not detected")
		}
		for _, m := range muts {
			if m.Kind != StoreMutationModified {
				t.Errorf("want modified, got %+v", m)
			}
		}
	})
}

// TestBackupQueueArtifactsHashVerified proves the backup copies each present
// queue artifact byte-for-byte, records source+copy SHA-256 and byte count,
// and records absent artifacts as absent (REQ-TQM-001, acceptance §D.5
// empty/absent case).
func TestBackupQueueArtifactsHashVerified(t *testing.T) {
	_, projectDir := mergeFixtureStore(t, nil)
	// Give the project store the legacy siblings the real store carries.
	if err := os.WriteFile(filepath.Join(projectDir, backlogFileName), []byte(`{"version":1}`), 0o644); err != nil {
		t.Fatalf("seed json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, backlogFileName+backlogMigratedSuffix), []byte("quarantine"), 0o644); err != nil {
		t.Fatalf("seed migrated: %v", err)
	}
	_, homeDir := mergeFixtureStore(t, nil) // home carries no json sibling

	backupDir := t.TempDir()
	result, err := BackupQueueArtifacts([]string{projectDir, homeDir}, backupDir)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if len(result.Artifacts) == 0 {
		t.Fatalf("no artifacts backed up")
	}
	for _, a := range result.Artifacts {
		if a.SourceSHA256 == "" || a.CopySHA256 == "" {
			t.Errorf("artifact %s missing hashes: %+v", a.Name, a)
		}
		if a.SourceSHA256 != a.CopySHA256 {
			t.Errorf("artifact %s hash mismatch: src=%s copy=%s", a.Name, a.SourceSHA256, a.CopySHA256)
		}
		if a.Bytes <= 0 {
			t.Errorf("artifact %s byte count %d not recorded", a.Name, a.Bytes)
		}
		if _, err := os.Stat(a.BackupPath); err != nil {
			t.Errorf("backup copy missing on disk: %s", a.BackupPath)
		}
	}
	// Each store's db and the project store's legacy siblings must all be
	// present in the backed-up set.
	joined := strings.Join(artifactNames(result.Artifacts), ",")
	for _, want := range []string{"backlog.db", "backlog.json", "backlog.json.migrated"} {
		if !strings.Contains(joined, want) {
			t.Errorf("expected %s among backed-up artifacts, got %s", want, joined)
		}
	}
	// The home store has no json sibling — its absence must be recorded, not
	// invented.
	foundAbsent := false
	for _, absent := range result.Absent {
		if absent.Store == homeDir && absent.Name == backlogFileName {
			foundAbsent = true
		}
	}
	if !foundAbsent {
		t.Errorf("home store's absent backlog.json not recorded as absent: %+v", result.Absent)
	}
}

func artifactNames(artifacts []BackedUpArtifact) []string {
	out := make([]string, 0, len(artifacts))
	for _, a := range artifacts {
		out = append(out, a.Store+"::"+a.Name)
	}
	return out
}

// TestBackupQueueArtifactsAbortsOnTamperedCopy proves the hash check is the
// gate: a copy that does not hash back to its source aborts the backup with
// an error (REQ-TQM-002).
func TestBackupQueueArtifactsAbortsOnTamperedCopy(t *testing.T) {
	_, projectDir := mergeFixtureStore(t, nil)
	backupDir := t.TempDir()

	// Interpose a tampering copy step: the copied bytes change between the
	// copy and the verification the backup performs.
	orig := copyFilePlain
	copyFilePlain = func(src, dst string) error {
		if err := orig(src, dst); err != nil {
			return err
		}
		f, err := os.OpenFile(dst, os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return err
		}
		_, werr := f.WriteString("TAMPERED")
		cerr := f.Close()
		if werr != nil {
			return werr
		}
		return cerr
	}
	defer func() { copyFilePlain = orig }()

	if _, err := BackupQueueArtifacts([]string{projectDir}, backupDir); err == nil {
		t.Fatal("backup accepted a tampered copy")
	}
}

// TestRestoreQueueArtifactsRoundTrip is the AC-TQM-007 rehearsal: backup,
// mutate the fixture stores, restore, and prove both stores are byte-identical
// (SHA-256) to their backups again.
func TestRestoreQueueArtifactsRoundTrip(t *testing.T) {
	queue, dir := mergeFixtureStore(t, nil)
	backupDir := t.TempDir()

	backup, err := BackupQueueArtifacts([]string{dir}, backupDir)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}

	// Mutate the store AFTER the backup — the post-mutation state the
	// rollback procedure exists to undo. The db engine artifact is the
	// store's carrier; the fixture never writes a legacy json sibling.
	engine := backlogSQLitePath(queue)
	if err := os.Chtimes(engine, time.Now().Add(time.Hour), time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("mutate mtime: %v", err)
	}

	if err := RestoreQueueArtifacts(backup); err != nil {
		t.Fatalf("restore: %v", err)
	}

	for _, a := range backup.Artifacts {
		raw, err := os.ReadFile(a.SourcePath)
		if err != nil {
			t.Fatalf("read restored %s: %v", a.SourcePath, err)
		}
		if sha256Hex(raw) != a.SourceSHA256 {
			t.Errorf("restored %s not byte-identical to its backup", a.SourcePath)
		}
	}
}

// TestRestoreQueueArtifactsErrorBranches covers the refusal arms: no backup,
// a missing backup file, and a tampered backup failing the re-hash.
func TestRestoreQueueArtifactsErrorBranches(t *testing.T) {
	if err := RestoreQueueArtifacts(nil); err == nil {
		t.Error("nil backup accepted")
	}

	_, dir := mergeFixtureStore(t, nil)
	backupDir := t.TempDir()
	backup, err := BackupQueueArtifacts([]string{dir}, backupDir)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}

	// Missing backup file: remove the copy, restore must fail on the copy.
	if err := os.Remove(backup.Artifacts[0].BackupPath); err != nil {
		t.Fatalf("remove copy: %v", err)
	}
	if err := RestoreQueueArtifacts(backup); err == nil {
		t.Error("restore accepted a missing backup file")
	}

	// Tampered backup: re-backup, corrupt the copy, restore must fail the
	// re-hash.
	backup2, err := BackupQueueArtifacts([]string{dir}, t.TempDir())
	if err != nil {
		t.Fatalf("backup2: %v", err)
	}
	if err := os.WriteFile(backup2.Artifacts[0].BackupPath, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("tamper: %v", err)
	}
	if err := RestoreQueueArtifacts(backup2); err == nil {
		t.Error("restore accepted a tampered backup")
	}
}

// TestBackupQueueArtifactsErrorBranches covers the directory-artifact refusal
// and the snapshot walk-error arms.
func TestBackupQueueArtifactsErrorBranches(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "backlog.db"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if _, err := BackupQueueArtifacts([]string{dir}, t.TempDir()); err == nil {
		t.Error("backup accepted a directory in place of the db artifact")
	}
	if _, err := SnapshotStoreDirState(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("snapshot accepted a nonexistent directory")
	}
	if _, err := fileSHA256(filepath.Join(t.TempDir(), "absent")); err == nil {
		t.Error("fileSHA256 accepted a nonexistent file")
	}
}

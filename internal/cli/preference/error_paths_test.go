package preference

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewFileStore_MkdirAll_Failure exercises the memory-root creation error
// path by pointing memDir at a path that cannot be created (a file blocks it).
func TestNewFileStore_MkdirAll_Failure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	// Create a FILE at the memory path so MkdirAll fails (EEXIST, not a dir).
	blocker := filepath.Join(root, "memory")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile blocker: %v", err)
	}
	if _, err := NewFileStore(blocker); err == nil {
		t.Errorf("NewFileStore at file-blocked path returned nil; want mkdir error")
	}
}

// TestNewFileStore_ArchivalMkdir_Failure exercises the archival-dir creation
// error path by placing a file where archival/ should go.
func TestNewFileStore_ArchivalMkdir_Failure(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	udDir := filepath.Join(memDir, userDecisionsDir)
	if err := os.MkdirAll(udDir, 0o755); err != nil {
		t.Fatalf("MkdirAll ud: %v", err)
	}
	// Block archival/ with a file.
	if err := os.WriteFile(filepath.Join(udDir, archivalDirName), []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile blocker: %v", err)
	}
	if _, err := NewFileStore(memDir); err == nil {
		t.Errorf("NewFileStore with archival-blocked path returned nil; want mkdir error")
	}
}

// TestUpsert_EmptyDomainAndKey exercises the Upsert arg-validation error paths.
func TestUpsert_EmptyDomainAndKey(t *testing.T) {
	t.Parallel()
	store := newTestStore(t)
	e := validEntry("x", "d", "k")

	if err := store.Upsert("", "k", e); err == nil {
		t.Errorf("Upsert with empty domain returned nil; want error")
	}
	if err := store.Upsert("d", "", e); err == nil {
		t.Errorf("Upsert with empty decisionKey returned nil; want error")
	}
}

// TestLoadArchival_ReadError exercises the archival-dir-read error path by
// making archival/ unreadable.
func TestLoadArchival_ReadError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	store, err := NewFileStore(filepath.Join(root, "memory"))
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs := store.(*fileStore)
	// Write one archival entry so loadArchival enters the read loop.
	entry := validEntry("x", "d", "k")
	if err := fs.writeArchivalEntry("d", "k", entry); err != nil {
		t.Fatalf("writeArchivalEntry: %v", err)
	}
	// Write a non-.json file inside archival/ that we then make unreadable via a
	// broken symlink to force a ReadDir+ReadFile error.
	broken := filepath.Join(fs.udDir, archivalDirName, "broken.json")
	if err := os.Symlink(filepath.Join(fs.udDir, archivalDirName, "does-not-exist"), broken); err != nil {
		t.Fatalf("Symlink: %v", err)
	}
	// loadArchival will follow the symlink, hit ENOENT on the target, and
	// surface a read error.
	if _, err := store.Query("d"); err == nil {
		t.Errorf("Query with broken archival symlink returned nil; want read error")
	}
}

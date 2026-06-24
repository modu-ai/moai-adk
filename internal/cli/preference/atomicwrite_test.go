package preference

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestAtomicWrite_CreateTempFailure exercises the os.CreateTemp error path in
// atomicWrite by making the target directory unwritable (POSIX only).
func TestAtomicWrite_CreateTempFailure(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission-based test; Windows ACL semantics differ")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root — chmod 0o500 does not deny root")
	}

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs := store.(*fileStore)

	// Make udDir read-only so CreateTemp fails.
	udDir := fs.udDir
	if err := os.Chmod(udDir, 0o500); err != nil {
		t.Fatalf("Chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(udDir, 0o755) }) // restore so t.TempDir cleanup works

	// An Upsert that needs to write (transient → recall) must fail.
	e := validEntry("x", "d", "k")
	e.Scope = ScopeTransient
	if err := store.Upsert("d", "k", e); err == nil {
		t.Errorf("Upsert to read-only dir returned nil; want create-temp error")
	}
}

// TestAtomicWrite_MkdirAllFailure exercises the MkdirAll error path in
// atomicWrite when the parent dir cannot be created.
func TestAtomicWrite_MkdirAllFailure(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission-based test")
	}
	if os.Geteuid() == 0 {
		t.Skip("running as root")
	}

	root := t.TempDir()
	memDir := filepath.Join(root, "memory")
	store, err := NewFileStore(memDir)
	if err != nil {
		t.Fatalf("NewFileStore: %v", err)
	}
	fs := store.(*fileStore)

	// Corrupt the archival subdir: replace it with a FILE so atomicWrite's
	// MkdirAll on a path-with-a-file-in-the-middle fails.
	archivalPath := filepath.Join(fs.udDir, archivalDirName)
	_ = os.RemoveAll(archivalPath)
	if err := os.WriteFile(archivalPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("WriteFile blocker: %v", err)
	}

	// writeArchivalEntry calls atomicWrite on archivalPath/... — MkdirAll fails.
	entry := validEntry("x", "d", "k")
	if err := fs.writeArchivalEntry("d", "k", entry); err == nil {
		t.Errorf("writeArchivalEntry with file-blocking-archival returned nil; want mkdir error")
	}
}

//go:build !windows

package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/privatefile"
)

func TestBacklogLockMakesUnixNamespacePrivateWhileHeld(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "queue")
	if err := os.Mkdir(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	// Use an explicit mode so the test does not depend on the runner's umask.
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	store := NewBacklogStore(filepath.Join(dir, backlogFileName))
	lock, err := store.acquireLock()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Release(); err != nil {
			t.Error(err)
		}
	}()
	if err := privatefile.CheckDirectory(dir); err != nil {
		t.Fatalf("held queue namespace is not private: %v", err)
	}
}

func TestBacklogLockDoesNotRestoreUnixOwnerWritePermission(t *testing.T) {
	dir := privateBacklogFixtureDir(t)
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Error(err)
		}
	})
	store := NewBacklogStore(filepath.Join(dir, backlogFileName))
	if lock, err := store.acquireLock(); err == nil {
		if err := lock.Release(); err != nil {
			t.Error(err)
		}
		t.Fatal("read-only namespace was made writable")
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o500 {
		t.Fatalf("refusal changed directory mode to %#o", info.Mode().Perm())
	}
	if _, err := os.Lstat(store.LockPath()); !os.IsNotExist(err) {
		t.Fatalf("refused namespace gained a lock record: %v", err)
	}
}

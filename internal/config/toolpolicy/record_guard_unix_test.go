//go:build !windows

package toolpolicy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/testguard"
)

func TestRecordGuardReplacementFailsClosed(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	release, err := acquireProjectRecordLock(settings)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(path+".guard", path+".guard.original"); err != nil {
		t.Fatal(err)
	}
	const replacement = "foreign replacement guard"
	if err := testguard.SeedFileModePolicy(path+".guard", []byte(replacement), 0o600); err != nil {
		t.Fatal(err)
	}
	stale := time.Now().Add(-recordLockStaleAge - time.Second)
	if err := os.Chtimes(path, stale, stale); err != nil {
		t.Fatal(err)
	}
	if err := takeoverStaleRecordLock(path); err == nil {
		t.Error("replaced guard identity accepted an old-generation record")
	}
	release()
	after, err := os.ReadFile(path)
	if err != nil || string(after) != string(before) {
		t.Errorf("old release under replaced guard changed primary record: %q, %v", after, err)
	}
	after, err = os.ReadFile(path + ".guard")
	if err != nil || string(after) != replacement {
		t.Errorf("replacement guard was changed: %q, %v", after, err)
	}
}

func TestRecordGuardRefusesSymlink(t *testing.T) {
	settings := filepath.Join(t.TempDir(), ".claude", "settings.json")
	path := managedAllowLockPath(settings)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(t.TempDir(), "foreign")
	if err := testguard.SeedFileModePolicy(foreign, []byte("unchanged"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(foreign, path+".guard"); err != nil {
		t.Fatal(err)
	}
	release, err := acquireProjectRecordLock(settings)
	if release != nil {
		release()
	}
	if err == nil {
		t.Error("symlink guard accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("symlink refusal created primary record: %v", err)
	}
}

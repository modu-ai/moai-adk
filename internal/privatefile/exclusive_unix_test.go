//go:build !windows

package privatefile

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateExclusiveNeverFollowsExistingSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "output")
	if err := os.WriteFile(target, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	called := false
	if err := CreateExclusive(link, func(f *os.File) {
		called = true
		if err := f.Close(); err != nil {
			t.Error(err)
		}
	}); !errors.Is(err, os.ErrExist) {
		t.Fatalf("existing symlink error=%v", err)
	}
	if called {
		t.Fatal("existing symlink was acquired")
	}
	after, err := os.Stat(target)
	if err != nil || !os.SameFile(before, after) || before.Mode() != after.Mode() {
		t.Fatalf("symlink target changed: %v", err)
	}
	if raw, err := os.ReadFile(target); err != nil || string(raw) != "original" {
		t.Fatalf("symlink target payload changed: %v", err)
	}
}

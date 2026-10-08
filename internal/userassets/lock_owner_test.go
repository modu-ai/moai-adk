package userassets

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLockPreservesLiveOwnerAfterStaleAge(t *testing.T) {
	home := t.TempDir()
	first, err := acquireUserLockStale(LockPath(home), time.Second, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = first.Release() })
	before, err := os.ReadFile(first.path)
	if err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(first.path, old, old); err != nil {
		t.Fatal(err)
	}
	second, err := acquireUserLockStale(first.path, 30*time.Millisecond, time.Millisecond)
	if second != nil {
		_ = second.Release()
		t.Fatal("a live owner's lock was reclaimed")
	}
	if !errors.Is(err, ErrLocked) {
		t.Fatalf("acquire = %v, want ErrLocked", err)
	}
	raw, err := os.ReadFile(first.path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, before) {
		t.Fatal("live owner's identity was lost")
	}
}

func TestLockPreservesUnknownOwnerAfterStaleAge(t *testing.T) {
	for _, record := range []string{"", "pid=invalid token=unknown", "pid=0", "pid=-1", "pid=4294967296"} {
		t.Run(record, func(t *testing.T) {
			path := LockPath(t.TempDir())
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(record), 0o644); err != nil {
				t.Fatal(err)
			}
			old := time.Now().Add(-time.Hour)
			if err := os.Chtimes(path, old, old); err != nil {
				t.Fatal(err)
			}
			lock, err := acquireUserLockStale(path, time.Millisecond, time.Millisecond)
			if lock != nil {
				_ = lock.Release()
				t.Fatal("unknown owner's lock was reclaimed")
			}
			if !errors.Is(err, ErrLocked) {
				t.Fatalf("acquire = %v, want ErrLocked", err)
			}
			raw, err := os.ReadFile(path)
			if err != nil || string(raw) != record {
				t.Fatalf("ownership changed: %q (%v)", raw, err)
			}
		})
	}
}

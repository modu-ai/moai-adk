package revoke

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWriteAtomic covers the record writer: a fresh directory is created, the
// bytes land, no temporary file is left behind, and an unusable parent is an
// error.
func TestWriteAtomic(t *testing.T) {
	t.Run("writes", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "a", "b", "r.md")
		if err := writeAtomic(path, []byte("body")); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != "body" {
			t.Fatalf("read back %q, %v", got, err)
		}
		left, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".revoke-*"))
		if len(left) != 0 {
			t.Errorf("temporary files left: %v", left)
		}
	})
	t.Run("parent-is-a-file", func(t *testing.T) {
		dir := t.TempDir()
		blocker := filepath.Join(dir, "blocker")
		if err := os.WriteFile(blocker, nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := writeAtomic(filepath.Join(blocker, "r.md"), []byte("x")); err == nil {
			t.Fatal("write under a regular file succeeded")
		}
	})
}

// TestWithDefaults covers the default seams: every nil seam is filled, and
// the default HEAD reader fails outside a repository instead of inventing a
// value.
func TestWithDefaults(t *testing.T) {
	s := withDefaults(Seams{})
	if s.Now == nil || s.GitHead == nil || s.Store == nil {
		t.Fatalf("nil seam left: now=%v head=%v store=%v", s.Now != nil, s.GitHead != nil, s.Store != nil)
	}
	if head, err := s.GitHead(t.TempDir()); err == nil {
		t.Errorf("HEAD outside a repository = %q, want an error", head)
	}
}

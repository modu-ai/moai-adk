package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// t1784: `factory assign --worktree` stores the SameFile-verified real path
// when the path exists, so variant spellings of the same tree (/tmp/x vs
// /private/tmp/x) cannot fork the record. A nonexistent path falls back to
// filepath.Abs.
func TestCanonicalWorktreePath(t *testing.T) {
	t.Run("existing path via symlink normalizes to the real path", func(t *testing.T) {
		base := t.TempDir()
		real := filepath.Join(base, "real-tree")
		if err := os.Mkdir(real, 0o755); err != nil {
			t.Fatal(err)
		}
		link := filepath.Join(base, "link-tree")
		if err := os.Symlink(real, link); err != nil {
			t.Fatal(err)
		}
		got, err := canonicalWorktreePath(link)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		want, err := filepath.EvalSymlinks(real)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("normalized %q, want the SameFile-verified real path %q", got, want)
		}
	})

	t.Run("nonexistent path falls back to filepath.Abs", func(t *testing.T) {
		input := filepath.Join(t.TempDir(), "does-not-exist-yet")
		got, err := canonicalWorktreePath(input)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		want, err := filepath.Abs(input)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %q, want the Abs fallback %q", got, want)
		}
	})

	t.Run("existing plain path is unchanged up to symlink resolution", func(t *testing.T) {
		input := t.TempDir()
		got, err := canonicalWorktreePath(input)
		if err != nil {
			t.Fatalf("canonicalWorktreePath: %v", err)
		}
		same, err := sameFilePath(input, got)
		if err != nil {
			t.Fatalf("sameFilePath: %v", err)
		}
		if !same {
			t.Fatalf("normalized %q is not the same file as the input %q", got, input)
		}
		if !filepath.IsAbs(got) {
			t.Fatalf("got %q, want an absolute path", got)
		}
	})
}

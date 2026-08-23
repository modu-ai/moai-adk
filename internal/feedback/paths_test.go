package feedback

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveProjectRootFindsMarker(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, ".moai", "state"), 0o755); err != nil {
		t.Fatalf("prepare marker: %v", err)
	}
	nested := filepath.Join(root, "internal", "pkg")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("prepare nested dir: %v", err)
	}

	got, ok := ResolveProjectRoot(nested)
	if !ok {
		t.Fatalf("expected the marker to be found walking up from %q", nested)
	}
	// macOS resolves t.TempDir() under /var, a symlink to /private/var.
	wantResolved, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("resolve want: %v", err)
	}
	gotResolved, err := filepath.EvalSymlinks(got)
	if err != nil {
		t.Fatalf("resolve got: %v", err)
	}
	if gotResolved != wantResolved {
		t.Fatalf("root mismatch: got %q, want %q", gotResolved, wantResolved)
	}
}

func TestResolveProjectRootReportsAbsence(t *testing.T) {
	t.Parallel()

	// A directory tree with no marker anywhere up to the filesystem root would
	// require an unmounted volume to guarantee; instead assert the contract at
	// the point that matters: an unreadable start reports absence rather than
	// guessing a root.
	if root, ok := ResolveProjectRoot(filepath.Join(t.TempDir(), "does", "not", "exist")); ok {
		t.Fatalf("expected absence for a non-existent start, got %q", root)
	}
}

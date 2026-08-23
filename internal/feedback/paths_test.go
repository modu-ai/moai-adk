package feedback

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/modu-ai/moai-adk/internal/defs"
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
	//
	// The walk is upward, so the assertion only holds where no ancestor of the
	// temp directory carries a marker. That is true of a POSIX CI runner and
	// false of the Windows one, whose temp lives under a home directory that
	// holds a .moai — there the walk finds a real marker and reporting it is
	// correct behaviour, not the guess this test is about. Probe first and skip
	// rather than assert a contract the environment has taken out of scope.
	start := filepath.Join(t.TempDir(), "does", "not", "exist")
	if marker, found := markerAbove(t, start); found {
		t.Skipf("an ancestor of the temp tree carries a marker (%s) — upward absence is unobservable here", marker)
	}

	if root, ok := ResolveProjectRoot(start); ok {
		t.Fatalf("expected absence for a non-existent start, got %q", root)
	}
}

// markerAbove reports whether any ancestor of start carries the project marker,
// walking the same direction ResolveProjectRoot does. It names the directory it
// found so a skip says which ancestor took the assertion out of scope.
func markerAbove(t *testing.T, start string) (string, bool) {
	t.Helper()

	dir := start
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, defs.MoAIDir)); err == nil && info.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

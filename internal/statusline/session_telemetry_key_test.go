// Package statusline tests for SPEC-SESSION-TELEMETRY-001 REQ-ST-007: the key
// is now a path component arriving from outside the process, so a value that
// would resolve outside the per-session directory is REFUSED — not sanitised
// and redirected, which would produce a file that looks legitimate and belongs
// to no session (plan.md §H).
package statusline

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// snapshotTree lists every file under root, relative to it.
func snapshotTree(t *testing.T, root string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		files = append(files, rel)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %q: %v", root, err)
	}
	sort.Strings(files)
	return files
}

// TestHostileKeyIsRefused — AC-ST-008 (REQ-ST-007).
// For each hostile key: no file is created outside the per-session directory,
// no file is created inside it either, and the render still completes.
func TestHostileKeyIsRefused(t *testing.T) {
	hostile := []struct {
		name string
		key  string
	}{
		{"empty", ""},
		{"parent traversal", "../escape"},
		{"path separator", "a/b"},
		{"absolute path", "/tmp/absolute"},
	}

	for _, tc := range hostile {
		t.Run(tc.name, func(t *testing.T) {
			// The project lives one level down so a traversal out of it is
			// still inside the snapshotted tree and therefore observable.
			root := t.TempDir()
			proj := filepath.Join(root, "project")
			if err := os.MkdirAll(proj, 0o755); err != nil {
				t.Fatal(err)
			}
			before := snapshotTree(t, root)

			// (c) the render still completes with its statusline line.
			buildForSession(t, proj, tc.key)

			after := snapshotTree(t, root)
			if len(after) != len(before) {
				t.Errorf("key %q created %d file(s): %v — a refused key must create none",
					tc.key, len(after)-len(before), after)
			}

			// (b) nothing inside the per-session directory either.
			dir := filepath.Join(proj, ".moai", "state", "context-usage")
			if entries, err := os.ReadDir(dir); err == nil && len(entries) > 0 {
				t.Errorf("key %q wrote %d entr(y/ies) into %s", tc.key, len(entries), dir)
			}
		})
	}
}

// TestSessionTelemetryPathRefusesHostileKeys — the refusal is in the path
// helper, so every caller inherits it rather than each re-deriving the rule.
func TestSessionTelemetryPathRefusesHostileKeys(t *testing.T) {
	t.Parallel()

	stateDir := filepath.Join("proj", ".moai", "state")
	for _, key := range []string{"", "../escape", "a/b", "/tmp/absolute", ".", "..", `a\b`} {
		if got := SessionTelemetryPath(stateDir, key); got != "" {
			t.Errorf("SessionTelemetryPath(_, %q) = %q, want \"\" (refused)", key, got)
		}
	}
	// A real session id still resolves.
	want := filepath.Join(stateDir, "context-usage", "3db058e1-2692-44f5-9e0d-45b543bb3c1f.json")
	if got := SessionTelemetryPath(stateDir, "3db058e1-2692-44f5-9e0d-45b543bb3c1f"); got != want {
		t.Errorf("SessionTelemetryPath = %q, want %q", got, want)
	}
}

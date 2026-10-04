//go:build !windows

// state_lock_symlink_mkdir_test.go — a refused lock path must also be a
// no-write path: when `.moai` (or `.moai/state`) is a symlink to an external
// directory, the callers must refuse BEFORE creating the lock's parent
// directories, otherwise os.MkdirAll creates lock directories OUTSIDE
// the project through the link and the later refusal comes too late (card
// t1458, P2 repair after the leader's codex_audit 2).
package factory

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// guardedEntryPoints returns the entry points that create the lock's parent
// directories only after the pre-create ancestor check.
func guardedEntryPoints() []lockEntryPoint {
	var out []lockEntryPoint
	for _, ep := range lockEntryPoints() {
		if ep.guardsDirCreation {
			out = append(out, ep)
		}
	}
	return out
}

// assertDirEmpty fails when dir holds any entry, naming what was created.
func assertDirEmpty(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("readdir %s: %v", dir, err)
	}
	if len(entries) == 0 {
		return
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	t.Fatalf("external dir %s gained entries through the link: %s", dir, strings.Join(names, ", "))
}

// `<root>/.moai` is a symlink to an EMPTY external directory: the acquisition
// is refused and nothing is created through the link.
func TestStateLockSymlinkedMoaiToEmptyDirCreatesNothing(t *testing.T) {
	for _, ep := range guardedEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			if err := os.Symlink(external, filepath.Join(root, ".moai")); err != nil {
				t.Fatal(err)
			}

			release, err := ep.acquire(root)
			assertDirEmpty(t, external)
			assertRefusedUnsafe(t, release, err)
		})
	}
}

// `<root>/.moai` is a real directory but `<root>/.moai/state` is a symlink to
// an EMPTY external directory: nothing (e.g. the lock directory) is created
// through the link.
func TestStateLockSymlinkedStateDirCreatesNothing(t *testing.T) {
	for _, ep := range guardedEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			external := t.TempDir()
			if err := os.MkdirAll(filepath.Join(root, ".moai"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, filepath.Join(root, ".moai", "state")); err != nil {
				t.Fatal(err)
			}

			release, err := ep.acquire(root)
			assertDirEmpty(t, external)
			assertRefusedUnsafe(t, release, err)
		})
	}
}

// Positive control: on a normal root the directories ARE created and the lock
// is acquired (the pre-create check refuses links, not missing directories).
func TestStateLockNormalRootStillCreatesItsDirs(t *testing.T) {
	for _, ep := range lockEntryPoints() {
		t.Run(ep.name, func(t *testing.T) {
			root := t.TempDir()
			release, err := ep.acquire(root)
			if err != nil {
				t.Fatalf("normal acquisition failed: %v", err)
			}
			defer func() { _ = release() }()
			info, serr := os.Stat(filepath.Join(root, ep.relDir))
			if serr != nil || !info.IsDir() {
				t.Fatalf("lock parent dir not created: %v", serr)
			}
		})
	}
}

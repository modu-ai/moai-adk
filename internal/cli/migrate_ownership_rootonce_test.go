package cli

// migrate_ownership_rootonce_test.go — SPEC-UPDATE-OWNERSHIP-REPAIR-001 (card
// t1865, chunk 13, leader disposition B): the unregistered file is read through an
// os.Root that is opened ONCE, where the ownership environment is built, and reused
// for every read of the run. Opening the root by its path string for each file
// (round 4, finding N-A) lets a swap of the project root's path, between the check
// and the open, land the read outside the project. The root-swap row of
// TestPathBoundaryCounterexamples proves the read side; this test pins the wiring:
// one open per run, whatever the number of files, and the handle is closed when the
// run ends.

import (
	"fmt"
	"os"
	"testing"
)

func TestMigrationOpensTheProjectRootOnce(t *testing.T) {
	root, home := ownershipFixtureBase(t)
	const unregistered = 6
	for i := 0; i < unregistered; i++ {
		ownAddUnregistered(t, root, fmt.Sprintf(".agents/skills/moai-once/f%d.md", i), []byte("# an unregistered file\n"))
	}
	var opened []*os.Root
	prev := openProjectRoot
	t.Cleanup(func() { openProjectRoot = prev })
	openProjectRoot = func(name string) (*os.Root, error) {
		r, err := prev(name)
		if err == nil {
			opened = append(opened, r)
		}
		return r, err
	}
	rc := &migrationRunContext{userAssetsInstalled: true}
	if _, err := migrateProjectCommonAssets(root, home, true, nil, func(string, ...interface{}) {}, rc); err != nil {
		t.Fatalf("migration: %v", err)
	}
	for i := 0; i < unregistered; i++ {
		rel := fmt.Sprintf(".agents/skills/moai-once/f%d.md", i)
		if r, ok := findRemaining(rc, rel); !ok || r.Reason != keepNoProvenanceRecord {
			t.Errorf("%s: row %+v (found=%v), want reason %q: the file was not read through the run's root", rel, r, ok, keepNoProvenanceRecord)
		}
	}
	if len(opened) != 1 {
		t.Fatalf("the project root was opened %d times, want exactly once per run", len(opened))
	}
	if f, err := opened[0].Open("."); err == nil {
		_ = f.Close()
		t.Errorf("the run's root is still open after the run ended")
	}
}

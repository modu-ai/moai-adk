package worktree

// migrate_oscwd_t2075_test.go — card t2075: on Windows there is no lsof, so the
// process-cwd scan errored and the plan skipped EVERY tree. A platform whose
// operating system refuses to rename a directory that a process holds as its
// working directory reports errProcessCWDOSGuarded; the plan then leaves that
// cause out and the move step's own failure handling is the guard. Any other
// probe error keeps skipping every tree.

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func planSkipFor(t *testing.T, probe func() ([]string, error)) (skip string, found bool) {
	t.Helper()
	w := m6NewWorld(t)
	m4InRepo(t, w.repo)
	activeProcessCWDProbe = probe
	plans, err := planMigrationFor(w.repo, time.Now(), false)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	for _, p := range plans {
		if p.source == w.movedTree {
			return p.skip, true
		}
	}
	return "", false
}

func TestPlanMigration_OSGuardedProbeDoesNotSkipTrees(t *testing.T) {
	skip, found := planSkipFor(t, func() ([]string, error) { return nil, errProcessCWDOSGuarded })
	if !found {
		t.Fatal("the plan does not list the legacy tree")
	}
	if skip != "" {
		t.Errorf("legacy tree skip = %q, want no skip when the OS guards the move", skip)
	}
}

func TestPlanMigration_OtherProbeErrorStillSkipsTrees(t *testing.T) {
	skip, found := planSkipFor(t, func() ([]string, error) { return nil, errors.New("lsof not found") })
	if !found {
		t.Fatal("the plan does not list the legacy tree")
	}
	if !strings.Contains(skip, "cannot inspect active process working directories") {
		t.Errorf("legacy tree skip = %q, want the unanswerable-scan skip kept for any other error", skip)
	}
}

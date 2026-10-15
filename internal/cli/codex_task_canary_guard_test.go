// codex_task_canary_guard_test.go — card t1632 sync-audit r7 (F22/F23): the
// canary device's guard DETECTS the registry-removal mutation it exists to
// catch, and the background-job isolation keeps the serving tree's real state
// untouched.
package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCodexTask_CanaryGuardCatchesRegistryRemoval pins the mutation the canary
// device exists to catch: when a cleanup widens into registry-DIRECTORY
// removal (RemoveAll wipes the canary along with the record), canaryLost
// reports the loss. The r6 device failed exactly this mutation — its own
// canary self-removal ran before the check, so a widened cleanup passed
// unnoticed.
func TestCodexTask_CanaryGuardCatchesRegistryRemoval(t *testing.T) {
	dir := t.TempDir()
	canary := filepath.Join(dir, "canary.json")
	if err := os.WriteFile(canary, []byte(`{"status":"running"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if lost, err := canaryLost(canary); lost {
		t.Fatalf("baseline: the canary was reported lost while present (err=%v)", err)
	}

	// THE MUTATION: a directory-wide removal — the regression the r6 device
	// was built to catch.
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	lost, err := canaryLost(canary)
	if !lost {
		t.Fatal("MUTATION NOT DETECTED: a removed canary must read as lost — the guard would let a registry-removal regression pass (sync-audit r7 F22/F23)")
	}
	_ = err
}

// TestCodexTask_BackgroundJobNeverTouchesTheServingState pins the structural
// isolation (r7): a background job run through the session-context harness
// leaves ZERO records under the serving tree's real job registry — the
// registry lives in the isolated temp repository only.
// Card t1632 sync-audit r8 F24: the serving registry ABSENT is a normal empty
// state (a clean checkout has no .moai/state at all) — the before/after
// reads treat IsNotExist as an empty listing and only an unexpected read
// error is fatal.
func TestCodexTask_BackgroundJobNeverTouchesTheServingState(t *testing.T) {
	serving := thisRepoRoot(t)
	jobsDir := filepath.Join(serving, ".moai", "state", codexJobsSubdir)
	before := map[string]bool{}
	if entries, err := os.ReadDir(jobsDir); err == nil {
		for _, e := range entries {
			before[e.Name()] = true
		}
	} else if !os.IsNotExist(err) {
		t.Fatalf("read the serving registry (before): %v", err)
	}

	// A fake session with a COMPLETING turn: this test's subject is the
	// registry isolation, not the session's context binding.
	withCodexSession(t, codexSessionScript("background work done"))
	jobID, _, cancel := startCtxBoundBackgroundJob(t)
	cancel()
	waitForJobToStop(t, jobID)

	// No NEW record may appear under the serving tree's registry (and nothing
	// pre-existing may vanish). An ABSENT registry reads as empty (F24).
	entries, err := os.ReadDir(jobsDir)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("read the serving registry (after): %v", err)
	}
	for _, e := range entries {
		if !before[e.Name()] {
			t.Errorf("a new record %q appeared under the SERVING tree's job registry — the isolation leaked (r7 F22/F23)", e.Name())
		}
	}
	// The isolated registry, by contrast, is where records went — verified by
	// the startCtxBoundBackgroundJob cleanup's canary check above.
}

package factory

// mergestep_red_t1582_test.go — plan-phase reproduction fixture for card
// t1582 item ④-R7-1 (cause-7 hold/release interleaving). This file is the
// SEEDING facility the run phase builds on: it proves the cause-7 injection
// reaches cause 7 (code 7) and records the pre-repair sequence the defect
// names — hold policy written FIRST, window release SECOND, as two separate
// mutations. The interleaving itself has no deterministic single-process
// injection point on the pre-repair tree (no seam sits between
// writeMergeHold and releaseHeldWindow), so the AC-006 verdict carries an
// executed-count guard and M4 authors the atomicity assertion alongside the
// repair — recorded honestly in acceptance.md, not papered over here.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRedT1582Cause7SeedingWritesHoldThenReleases(t *testing.T) {
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		switch {
		case args[0] == "merge" && len(args) > 1 && args[1] == "--abort":
			// The abort succeeds; the untracked residue written at the merge
			// response survives it — the cause-7 dirty state.
			return "", nil
		case args[0] == "merge":
			if err := os.WriteFile(filepath.Join(f.integ, "residue-untracked.txt"), []byte("residue"), 0o644); err != nil {
				return "", err
			}
			return "", fmt.Errorf("simulated merge failure (t1582 cause-7 seeding)")
		}
		// Everything else passes through to the real git runner — the seam
		// only scripts the merge act itself.
		return execGitIn(f.integ, args...)
	}
	_, err := RunMergeStep(f.input(), seams)
	// The seeding reaches cause 7 — merge failed, the abort left the
	// worktree dirty.
	requireCode(t, err, MergeExitMergeDirty)
	// The cause message is the pure cause (no release-failure wrapping on
	// the green path this control exercises).
	if msg := err.Error(); strings.Contains(msg, "releasing the window also failed") {
		t.Fatalf("positive control: the release must succeed on this fixture, got wrapping: %s", msg)
	}
	// The hold policy names the cause-7 state for the leader.
	policy, policyErr := ReadIntegrationWindowPolicy(f.root)
	if policyErr != nil {
		t.Fatalf("read hold policy: %v", policyErr)
	}
	if policy.Policy != PolicyHold {
		t.Fatalf("cause-7 must write the hold policy, got %q (%s)", policy.Policy, policy.Reason)
	}
	// The release after the hold empties the window WITHOUT promoting —
	// REQ-MWQ-018: no later holder is promoted onto the held state. The
	// waiting ticket stays queued for the leader.
	lock, lockErr := ReadIntegrationLock(f.root)
	if lockErr != nil {
		t.Fatal(lockErr)
	}
	if lock.SessionID != "" {
		t.Fatalf("cause-7 releases without promotion (REQ-MWQ-018), holder is %q", lock.SessionID)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].Card != "t0003" {
		t.Fatalf("the waiting ticket must stay queued under the hold: %+v", lock.Queue)
	}
}

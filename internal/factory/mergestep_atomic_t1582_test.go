package factory

// mergestep_atomic_t1582_test.go — card t1582 item ④-R7-1 (REQ-MWQ2-007,
// AC-MWQ2-006): the cause-7 outcome writes the hold and releases the window as
// one serialized mutation. The interleaving observation presents a status
// refresh in the gap the defect names — after the hold has landed and before
// the window is released. RED on the pre-repair tree: the refresh lands in the
// gap, clears the caller's holder, and the caller's own release then fails
// after the hold has landed. GREEN after the repair: the mutation section
// excludes the refresh until the hold and the release have both landed.

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// cause7GapProbe is how long the seam waits for an injected mutation to finish.
// It stays below stateLockWaitBudget (1.65s): a mutation that cannot take the
// section's lock gives up only after the budget, so the probe ends first.
const cause7GapProbe = 750 * time.Millisecond

func TestRedT1582Cause7HoldAndReleaseAreOneMutation(t *testing.T) {
	f := newMergeFixture(t)
	card := f.withCardTree(readyCardPtr())
	seams := f.seams(card)
	seams.Git = func(args ...string) (string, error) {
		switch {
		case args[0] == "merge" && len(args) > 1 && args[1] == "--abort":
			return "", nil
		case args[0] == "merge":
			if err := os.WriteFile(filepath.Join(f.integ, "residue-untracked.txt"), []byte("residue"), 0o644); err != nil {
				return "", err
			}
			return "", fmt.Errorf("simulated merge failure (t1582 cause-7 interleaving)")
		}
		return execGitIn(f.integ, args...)
	}

	// The injected mutation: a status refresh that observes the caller's holder
	// as gone, the stale-window cleanup a status verb performs. The seam starts
	// it the instant the hold has landed, while the caller still holds the
	// window; the probe then records whether it finished inside the gap.
	injected := make(chan error, 1)
	interleaved := false
	seams.AfterHold = func() {
		go func() {
			injected <- UpdateIntegrationWindow(f.root, func(w *IntegrationLock) error {
				if w.Held() && w.SessionID == "sess-b" {
					clearHolder(w, WindowClock(), "status refresh observed the holder gone")
				}
				return nil
			})
		}()
		select {
		case err := <-injected:
			interleaved = true
			injected <- err
		case <-time.After(cause7GapProbe):
		}
	}

	_, stepErr := RunMergeStep(f.input(), seams)
	// Join the injected mutation whichever way the section went: a mutation
	// that waited on the section must finish once the section closes.
	select {
	case <-injected:
	case <-time.After(10 * time.Second):
		t.Fatal("the injected status refresh never completed after the step returned")
	}
	if interleaved {
		t.Fatalf("RED t1582-R7-1: a status refresh landed between the cause-7 hold write and the window release, and the caller's release then failed (step error: %v)", stepErr)
	}
	requireCode(t, stepErr, MergeExitMergeDirty)
	if msg := stepErr.Error(); strings.Contains(msg, "releasing the window also failed") {
		t.Fatalf("the caller's release failed after the hold landed: %v", stepErr)
	}

	// The outcome's end state, read from the records: the hold is on record,
	// the caller's release emptied the window, and the waiting ticket stayed
	// queued under the hold (REQ-MWQ-018: no promotion onto a held state).
	policy, err := ReadIntegrationWindowPolicy(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Policy != PolicyHold {
		t.Fatalf("cause-7 must leave the hold on record, got %q (%s)", policy.Policy, policy.Reason)
	}
	lock, err := ReadIntegrationLock(f.root)
	if err != nil {
		t.Fatal(err)
	}
	if lock.Held() {
		t.Fatalf("the caller's release must empty the window under the hold: holder %q", lock.SessionID)
	}
	if len(lock.Queue) != 1 || lock.Queue[0].Card != "t0003" {
		t.Fatalf("the waiting ticket must stay queued under the hold: %+v", lock.Queue)
	}
}

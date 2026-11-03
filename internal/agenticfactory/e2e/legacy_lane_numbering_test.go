package e2e

import (
	"sync"
	"testing"

	"github.com/modu-ai/moai-adk/internal/factory"
)

// Stand-in for the agent-n scenarios (DELTA4): the agent naming is undecided,
// so these measure the current lane-<n> assignment primitive that agent-n will
// replace. When the names are fixed only the labels in these tests change.

func alive(int) bool { return true }

// Concurrent auto-numbered claims asking for the same number each get a
// distinct label.
func TestLegacyLaneNumbering_ConcurrentAutoClaims_DistinctLabels(t *testing.T) {
	root := t.TempDir()
	const n = 10
	labels := make([]string, n)
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			c, err := factory.ClaimFactoryLaneWithBackend(root, "lane-1", true, 20000+i, "run-x", "claude", alive)
			labels[i], errs[i] = c.Label, err
		})
	}
	wg.Wait()
	seen := map[string]int{}
	for i, err := range errs {
		if err != nil {
			t.Fatalf("claim %d: %v", i, err)
		}
		seen[labels[i]]++
	}
	for label, count := range seen {
		if count != 1 {
			t.Errorf("label %s assigned %d times", label, count)
		}
	}
	if len(seen) != n {
		t.Errorf("distinct labels = %d, want %d", len(seen), n)
	}
}

// Observed (not assumed): a live process asking again for the label it already
// holds is bumped to the next free number; it does not keep its label.
func TestLegacyLaneNumbering_LiveHolderReclaim_IsBumped(t *testing.T) {
	root := t.TempDir()
	first, err := factory.ClaimFactoryLaneWithBackend(root, "lane-3", false, 30001, "run-x", "claude", alive)
	if err != nil {
		t.Fatal(err)
	}
	again, err := factory.ClaimFactoryLaneWithBackend(root, first.Label, false, 30001, "run-x", "claude", alive)
	if err != nil {
		t.Fatal(err)
	}
	if again.Label == first.Label {
		t.Errorf("live holder kept %s; the characterized behavior is a bump", first.Label)
	}
}

// Resume identity: when the old holder is gone, a new process resuming the
// same label gets that label back.
func TestLegacyLaneNumbering_DeadHolderResume_KeepsLabel(t *testing.T) {
	root := t.TempDir()
	first, err := factory.ClaimFactoryLaneWithBackend(root, "lane-3", false, 30001, "run-x", "claude", alive)
	if err != nil {
		t.Fatal(err)
	}
	oldDead := func(pid int) bool { return pid != 30001 }
	resumed, err := factory.ClaimFactoryLaneWithBackend(root, first.Label, false, 30002, "run-x", "claude", oldDead)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.Label != first.Label {
		t.Errorf("resumed as %s, want %s", resumed.Label, first.Label)
	}
}

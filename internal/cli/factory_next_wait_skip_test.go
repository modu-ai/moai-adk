// factory_next_wait_skip_test.go — card t1746 F1 fix-forward regression tests:
// the --wait peek (factoryNextMayLease) must mirror the selection skips the
// arms apply — another lane's bundle member, a candidate whose after-hint
// predecessor is unmerged, a candidate with an unmerged hub predecessor — so a
// candidate the selector never leases neither takes the open-PR query (gh pr
// list) per poll nor keeps the wait loop querying for a card no arm will hand
// out. RED-first: each test failed against the pre-fix peek.
package cli

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flCountGhCalls counts the `gh` calls the verb issues, answering each with
// answer; every other command passes through to the runner it wraps. The
// returned getter reads the counter (0 when no gh call ran).
func flCountGhCalls(t *testing.T, answer func() (string, error)) func() int {
	t.Helper()
	var calls atomic.Int64
	prevRun := todoRunCommand
	prevContextRun := todoRunCommandContext
	todoRunCommand = func(name string, args ...string) (string, error) {
		if name != "gh" {
			return prevRun(name, args...)
		}
		calls.Add(1)
		return answer()
	}

	todoRunCommandContext = func(ctx context.Context, name string, args ...string) (string, error) {
		if name != "gh" {
			return prevContextRun(ctx, name, args...)
		}
		return todoRunCommand(name, args...)
	}
	t.Cleanup(func() { todoRunCommand = prevRun; todoRunCommandContext = prevContextRun })
	return func() int { return int(calls.Load()) }
}

// flHintChainFixture places the after-hint chain: t1 in flight on lane-2,
// t2 queue-picked with no owner whose after hint names the still-unmerged t1 —
// the shape the arm-(b)/(c) selection skip holds out of every un-nominated
// lease. The test process chdirs into the fixture root, the way the verb
// resolves it.
func flHintChainFixture(t *testing.T) string {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
	for _, id := range []string{"t1", "t2"} {
		fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	}
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardLeased, OwnerLabel: "lane-2", LeaseHolder: "lane-2", LeaseExpiresAt: live},
		homestate.Card{CardID: "t2", State: homestate.CardPicked, HintAfter: "t1"},
	)
	sdRegisterLane(t, root, "lane-1")
	t.Chdir(root)
	return root
}

// TestFactoryNextWaitStopsPollingGhForSelectorSkippedCandidate — card t1746
// F1 regression, the wait-loop shape of the probe: with --wait, a candidate
// the selector never leases (t2's hint names the unmerged t1) must take no
// open-PR query in any poll, and the control — the predecessor merged —
// leases on the first attempt behind exactly one query.
func TestFactoryNextWaitStopsPollingGhForSelectorSkippedCandidate(t *testing.T) {
	t.Run("unmerged predecessor polls gh per round", func(t *testing.T) {
		flHintChainFixture(t)
		sdLaneEnv(t, "lane-1", "")
		qasFakeClock(t, nil, nil)
		gh := flCountGhCalls(t, func() (string, error) { return "[]", nil })
		out, _, err := qasRunNext(t, "--run", fcRun, "--wait", "--wait-bound", "10s")
		sdExit3(t, "every candidate waits behind an unmerged predecessor", err)
		_ = out
		if n := gh(); n != 0 {
			t.Errorf("--wait over a selector-skipped candidate issued %d gh calls across the wait, want 0 — the peek must mirror the after-hint skip: t2's hint names the unmerged t1, so no poll can lease", n)
		}
	})
	t.Run("merged predecessor leases on the first attempt", func(t *testing.T) {
		root := flHintChainFixture(t)
		fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
		sdLaneEnv(t, "lane-1", "")
		gh := flCountGhCalls(t, func() (string, error) { return "[]", nil })
		out, stderr, err := qasRunNext(t, "--run", fcRun)
		if err != nil {
			t.Fatalf("bare next past the merged predecessor: %v (stderr %q)", err, stderr)
		}
		if !strings.Contains(out, "t2 stage=") {
			t.Fatalf("output = %q, want the leased t2 line (the merged predecessor must release its follower)", out)
		}
		if n := gh(); n != 1 {
			t.Errorf("the first attempt that leased issued %d gh calls, want 1", n)
		}
	})
}

// TestFactoryNextMayLeaseMirrorsTheSelectorSkips — the direct form of the
// same defect: the bare peek answers false exactly while a selection skip
// holds the candidate out, and true again once the skip clears. The
// bundle-ownership subtest isolates the bundle leg by merging the member's
// hint predecessor first.
func TestFactoryNextMayLeaseMirrorsTheSelectorSkips(t *testing.T) {
	t.Run("unmerged after-hint predecessor", func(t *testing.T) {
		root := flHintChainFixture(t)
		if factoryNextMayLease(context.Background(), root, fcRun, "lane-1", "", false) {
			t.Errorf("peek = true while candidate t2's hint names the unmerged t1, want false — the arm-(b) skip must hold it out (the wait loop re-queried gh per poll for it)")
		}
	})
	t.Run("merged predecessor releases the candidate", func(t *testing.T) {
		root := flHintChainFixture(t)
		fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
		if !factoryNextMayLease(context.Background(), root, fcRun, "lane-1", "", false) {
			t.Errorf("peek = false after t1 reached the local merge, want true — the skip must clear with the merge (fail-closed guard against the unknown PR column)")
		}
	})
	t.Run("another lane's bundle member", func(t *testing.T) {
		root, store := fcFixture(t)
		fcQueue(t, store, factory.BacklogStatePicked, factory.BacklogStatePicked)
		for _, id := range []string{"t1", "t2"} {
			fcClassify(t, store, id, factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
		}
		sdRegisterLane(t, root, "lane-1")
		sdRegisterLane(t, root, "lane-2")
		t.Chdir(root)
		fbBundle(t, root, "lane-1", "t1", "t2")
		fcSetCardState(t, root, "t1", homestate.CardMergedLocal)
		if factoryNextMayLease(context.Background(), root, fcRun, "lane-2", "", false) {
			t.Errorf("lane-2's peek = true over lane-1's bundle member t2, want false — the bundle-ownership skip must hold it out")
		}
		if !factoryNextMayLease(context.Background(), root, fcRun, "lane-1", "", false) {
			t.Errorf("lane-1's peek = false over its own assigned bundle head t1, want true — the arm-(a) mirror must still see the lane's own assignment")
		}
	})
}

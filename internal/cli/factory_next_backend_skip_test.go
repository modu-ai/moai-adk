// factory_next_backend_skip_test.go — post-merge review repairs for PR #1832
// (card t1746): the --wait peek (factoryNextMayLease) must mirror the backend
// skip the lease path refuses on (a GPT lane cannot advance a card recorded at
// merge-ready or later), and factoryNextPRSnapshot must recheck the request
// after the peek — a cancel landing inside the peek fails the peek open to
// true, and the prefetch must not start for a caller that is already gone.
// RED-first: each test failed against the pre-fix code.
package cli

import (
	"context"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modu-ai/moai-adk/internal/config"
	"github.com/modu-ai/moai-adk/internal/factory"
	"github.com/modu-ai/moai-adk/internal/homestate"
)

// flBackendSkipFixture places the backend-skip shape: t1 assigned to lane-1
// and recorded at merge-ready — the row factoryNextSkipForBackend refuses on
// a GPT lane, and the only row the peek could otherwise call reachable. The
// MOAI_KANBAN_BACKEND env stamps this process a GPT lane.
func flBackendSkipFixture(t *testing.T) string {
	t.Helper()
	root, store := fcFixture(t)
	fcQueue(t, store, factory.BacklogStatePicked)
	fcClassify(t, store, "t1", factory.ClassPriorityNormal, false, factory.ClassModeParallelizable)
	live := fcNow.Add(time.Hour).Format(time.RFC3339Nano)
	fcPlace(t, root,
		homestate.Card{CardID: "t1", State: homestate.CardAssigned, OwnerLabel: "lane-1",
			Stage: homestate.CardMergeReady, LeaseHolder: "lane-1", LeaseExpiresAt: live},
	)
	sdRegisterLane(t, root, "lane-1")
	t.Setenv(config.EnvFactoryBackend, factory.BackendGPT)
	t.Chdir(root)
	return root
}

// TestFactoryNextMayLeaseMirrorsTheBackendSkip — the direct form: on a GPT
// lane the peek answers false over an assigned merge-ready row (the lease
// path refuses it via factoryNextSkipForBackend), and true on a Claude lane
// where the same row is a plain arm-(a) candidate.
func TestFactoryNextMayLeaseMirrorsTheBackendSkip(t *testing.T) {
	t.Run("GPT lane holds the merge-ready row out", func(t *testing.T) {
		root := flBackendSkipFixture(t)
		if factoryNextMayLease(context.Background(), root, fcRun, "lane-1", "", false) {
			t.Errorf("GPT lane's peek = true over its assigned merge-ready row t1, want false — the lease path refuses the same row (factoryNextSkipForBackend), so the wait loop would poll gh for a card no arm can hand out")
		}
	})
	t.Run("Claude lane still sees the row", func(t *testing.T) {
		root := flBackendSkipFixture(t)
		t.Setenv(config.EnvFactoryBackend, factory.BackendClaude)
		if !factoryNextMayLease(context.Background(), root, fcRun, "lane-1", "", false) {
			t.Errorf("Claude lane's peek = false over its assigned merge-ready row t1, want true — the backend skip is GPT-scoped, the arm-(a) mirror must still see the lane's own assignment")
		}
	})
}

// TestFactoryNextWaitStopsPollingGhForBackendSkippedCandidate — the wait-loop
// shape of the same defect: with --wait, a GPT lane whose only candidate is
// the backend-skipped merge-ready row must take no open-PR query in any poll.
func TestFactoryNextWaitStopsPollingGhForBackendSkippedCandidate(t *testing.T) {
	flBackendSkipFixture(t)
	// The third argument is the backend stamp — sdLaneEnv writes
	// MOAI_KANBAN_BACKEND itself, so the GPT lane identity rides here, not
	// only in the fixture's t.Setenv.
	sdLaneEnv(t, "lane-1", factory.BackendGPT)
	qasFakeClock(t, nil, nil)
	gh := flCountGhCalls(t, func() (string, error) { return "[]", nil })
	out, _, err := qasRunNext(t, "--run", fcRun, "--wait", "--wait-bound", "10s")
	sdExit3(t, "every candidate is backend-skipped", err)
	_ = out
	if n := gh(); n != 0 {
		t.Errorf("--wait over a backend-skipped candidate issued %d gh calls, want 0 — the peek must mirror the backend skip: no poll can lease a row the lease path refuses", n)
	}
}

// flCancelAfterEntryCtx is a request context whose Err() answers nil exactly
// once (the snapshot's entry pre-check) and context.Canceled on every later
// call — the deterministic form of "the request was canceled while the peek
// was reading the record": the cancel lands inside the peek's window, not
// before the call.
type flCancelAfterEntryCtx struct {
	context.Context
	canceled atomic.Bool
}

func (c *flCancelAfterEntryCtx) Err() error {
	if c.canceled.CompareAndSwap(false, true) {
		return nil
	}
	return context.Canceled
}

// TestFactoryNextPRSnapshotRechecksCancelAfterThePeek — the peek fails open
// to true when the record cannot be read; a cancel landing inside that window
// must not start the gh subprocess. The fixture makes the factory record
// unopenable (its database path is a directory), so the peek fails open; the
// flCancelAfterEntryCtx cancels between the entry pre-check and the return.
func TestFactoryNextPRSnapshotRechecksCancelAfterThePeek(t *testing.T) {
	root, _ := fcFixture(t)
	dbPath, err := homestate.FactoryDBPath(root)
	if err != nil {
		t.Fatalf("factory db path: %v", err)
	}
	if err := os.MkdirAll(dbPath, 0o755); err != nil {
		t.Fatalf("plant unopenable factory record: %v", err)
	}
	t.Chdir(root)
	gh := flCountGhCalls(t, func() (string, error) { return "[]", nil })
	ctx := &flCancelAfterEntryCtx{Context: context.Background()}
	q := factoryNextPRSnapshot(ctx, root, fcRun, "lane-1", "", false)
	if q.err == nil {
		t.Fatalf("snapshot took a query (err nil) — want the not-taken outcome")
	}
	if n := gh(); n != 0 {
		t.Errorf("the prefetch started %d gh call(s) for a request canceled inside the peek window, want 0 — the snapshot must recheck ctx.Err() after the peek before starting the subprocess", n)
	}
}

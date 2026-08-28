// board_lock_wait_test.go — the queue lock-wait policy's guards
// (SPEC-BACKLOG-LOCK-BUDGET-001, card t354).
//
// The policy under test is the shared budget-and-backoff both queue-lock
// acquisition paths consume. The lock's substrate acquire is non-blocking and
// the callers poll, so there is no queue and no fairness underneath: these
// tests assert the derivation is visible and the wait is not lockstep, NOT
// that any budget is sufficient. Sufficiency is only evidenced by CI.
package kanban

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// AC-BLB-001 / REQ-BLB-001, REQ-BLB-002: the budget is derived from named
// inputs with a stated headroom factor, not declared as a bare constant.
//
// The mutant this catches: replacing the derivation with a numeric literal,
// or lowering an input without the derivation following it.
func TestBoardLockWaitBudgetDerivedFromNamedInputs(t *testing.T) {
	t.Parallel()

	recomputed := time.Duration(boardLockSupportedWriters) *
		boardLockCIMutationCost * boardLockHeadroom
	if boardLockWaitBudget != recomputed {
		t.Fatalf("budget %v is not the product of its named inputs (%d writers x %v x %d headroom = %v) — "+
			"a bare literal with no derivable inputs fails REQ-BLB-001",
			boardLockWaitBudget, boardLockSupportedWriters, boardLockCIMutationCost,
			boardLockHeadroom, recomputed)
	}

	// The inequality REQ-BLB-002 states: at least headroom x per-mutation
	// cost x supported contender count.
	floor := time.Duration(boardLockSupportedWriters) *
		boardLockCIMutationCost * boardLockHeadroom
	if boardLockWaitBudget < floor {
		t.Errorf("budget %v < headroom floor %v", boardLockWaitBudget, floor)
	}

	// The supported contender count is the ten-lane figure of record in
	// backlog_concurrency_test.go's header comment (REQ-BLB-002).
	if boardLockSupportedWriters != 10 {
		t.Errorf("supported writers = %d, want 10 (Factory mode's ten lanes against one queue)",
			boardLockSupportedWriters)
	}

	// The per-mutation cost is sized from the CI-class observation
	// (1.57s / 48 mutations ~= 33ms), never from the faster isolated local
	// figure (~14ms) — sizing to the fast machine is what made the retired
	// budget thin.
	if boardLockCIMutationCost < 33*time.Millisecond {
		t.Errorf("per-mutation cost %v is below the CI-class observation of 33ms",
			boardLockCIMutationCost)
	}

	if boardLockHeadroom < 2 {
		t.Errorf("headroom factor %d states no headroom", boardLockHeadroom)
	}
}

// AC-BLB-002 / REQ-BLB-003, REQ-BLB-004: the retry wait varies per contender,
// so no contender is systematically beaten by the same peers, and every value
// stays inside the policy's declared bounds.
//
// The assertions are distinctness and bounds — never a specific sampled
// value, which would pin the test to a draw.
func TestBoardLockRetryWaitIsNotLockstep(t *testing.T) {
	t.Parallel()

	const contenders = 32
	for attempt := 0; attempt < 6; attempt++ {
		seen := make(map[time.Duration]struct{}, contenders)
		for c := 0; c < contenders; c++ {
			d := boardLockRetryWait(attempt)
			if d < boardLockWaitMin || d > boardLockWaitMax {
				t.Fatalf("attempt %d contender %d: wait %v outside declared bounds [%v, %v]",
					attempt, c, d, boardLockWaitMin, boardLockWaitMax)
			}
			seen[d] = struct{}{}
		}
		if len(seen) < 2 {
			t.Errorf("attempt %d: all %d contenders drew the identical wait — the retry loop is lockstep",
				attempt, contenders)
		}
	}

	// One contender across consecutive attempts also varies: a fixed delay
	// identical across the run is what REQ-BLB-004 forbids.
	consecutive := make(map[time.Duration]struct{}, 16)
	for attempt := 0; attempt < 16; attempt++ {
		consecutive[boardLockRetryWait(attempt)] = struct{}{}
	}
	if len(consecutive) < 2 {
		t.Errorf("one contender drew the identical wait across 16 consecutive attempts — fixed delay")
	}
}

// AC-BLB-003 / REQ-BLB-005: a stuck holder still surfaces a bounded error
// naming both the queue file and the lock artifact. This is a regression
// guard on a property that exists today, not a new behaviour.
//
// The holder is a value in this test's own scope, released by a
// t.Cleanup-registered function. No background process is spawned.
func TestBacklogLockStuckHolderSurfacesBoundedNamedError(t *testing.T) {
	t.Parallel()

	store := NewBacklogStore(filepath.Join(t.TempDir(), "backlog.json"))
	held, err := acquireBoardLockImpl(store.LockPath())
	if err != nil {
		t.Fatalf("seeding the stuck holder: %v", err)
	}
	t.Cleanup(func() {
		if rerr := held.release(); rerr != nil {
			t.Logf("releasing the seeded holder: %v", rerr)
		}
	})

	start := time.Now()
	err = store.Mutate(func(rec *BacklogRecord) error {
		rec.LastSeq++
		return nil
	})
	elapsed := time.Since(start)

	if err == nil {
		t.Fatalf("the mutation succeeded while the lock was held by a stuck holder")
	}
	if !IsBoardLockHeld(err) {
		t.Errorf("error is not recognized by IsBoardLockHeld: %v", err)
	}
	if !strings.Contains(err.Error(), store.path) {
		t.Errorf("error does not name the queue file %s: %v", store.path, err)
	}
	if !strings.Contains(err.Error(), store.LockPath()) {
		t.Errorf("error does not name the lock artifact %s: %v", store.LockPath(), err)
	}
	if bound := 2 * boardLockWaitBudget; elapsed > bound {
		t.Errorf("the mutation blocked %v, past the bound of %v (budget %v)",
			elapsed, bound, boardLockWaitBudget)
	}
}

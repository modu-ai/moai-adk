// state_lock_wait.go — the shared lock-wait policy (budget and backoff) that
// every state-lock acquisition path consumes: the todo queue (backlog_store.go),
// the integration lock (integration_lock_mutation.go), and the slot lease
// (slot_lease.go). It was moved here verbatim from board_store.go when the
// file-lock substrate was re-homed (SPEC-LAUNCHER-ENTRY-FLAGS-001 M6).
package kanban

import (
	"math/rand/v2"
	"time"
)

// Board-lock acquisition is non-blocking at the substrate, so a mutation
// racing the current holder retries for a bounded window rather than failing
// the mutation: two concurrent transitions must BOTH reach the admission
// decision in turn (exactly one succeeding on the WIP bound), not have one
// bounce off the lock. The window is bounded so a genuinely stuck holder
// surfaces as an error instead of a hang.
//
// The queue lock-wait policy below is the SHARED budget-and-backoff both
// queue-lock acquisition paths consume — this one and (*BacklogStore).acquireLock
// (backlog_store.go). A change here applies to both without a second edit
// (SPEC-BACKLOG-LOCK-BUDGET-001 REQ-BLB-006).
//
// What the derivation is NOT. The substrate acquire is
// flock(LOCK_EX|LOCK_NB) (state_lock_unix.go) and the callers poll, so the
// lock provides no queue and no fairness: a contender that loses a race gains
// no priority for the next round, and flock guarantees no ordering between
// waiters. A contender's maximum wait is therefore the tail of a run of lost
// races, not a queue position, and it is unbounded in principle — no
// closed-form worst case exists. The product below is a SIZING HEURISTIC WITH
// STATED HEADROOM, NOT a worst-case bound, and must not be read as one. It is
// sized to survive a bounded run of lost races at the supported contender
// count on a CI-class machine; it is not sized for FIFO queue depth, which
// does not describe this lock.
const (
	// stateLockSupportedWriters is the concurrent lane count the product
	// supports against one queue: Factory mode runs up to ten lanes, the
	// figure of record in backlog_concurrency_test.go's header comment.
	stateLockSupportedWriters = 10

	// stateLockCIMutationCost is the per-mutation cost observed on a
	// CI-class machine under -race: 1.57s across 48 serialized mutations,
	// or ~33ms each. The isolated local figure (~14ms) is deliberately NOT
	// the input — sizing the budget to the faster machine is what left the
	// retired 1.025s window 87% consumed on a machine where the guard
	// passed.
	stateLockCIMutationCost = 33 * time.Millisecond

	// stateLockHeadroom is the stated headroom factor over the product
	// above. Ten means: survive roughly ten consecutive rounds of losing
	// to every peer before a stuck holder is declared.
	//
	// The product stateLockSupportedWriters * stateLockHeadroom = 100 is the
	// serialized-mutation count this policy budgets for, comfortably above
	// the 8 x 6 = 48 mutations TestConcurrencyStress serializes through
	// one flock. That near-coincidence used to be accidental;
	// TestStateLockWaitBudgetCoversSerializedMutations (state_lock_wait_test.go,
	// SPEC-STRESS-INVARIANT-VERDICT-001) now pins it, so lowering either
	// constant here fails that guard. It is a constant-coherence relation and
	// asserts nothing about the wait a real machine needs.
	//
	// Headroom history: 5 budgeted 1.65s, and CI run 36475337568 (9cc3fdc4d)
	// starved a contender past it — TestBacklogConcurrentAdd_UniqueIDs lost
	// the flock for the whole window under -race load with 8 contenders
	// (flock is not FIFO-fair, so a descheduled goroutine can lose the
	// wake-up race repeatedly). This is the second exhaustion of this class:
	// the retired 1.025s window was the first. 10 doubles the window; if a
	// third exhaustion lands, the fix is lock fairness (a real queue), not
	// another raise.
	stateLockHeadroom = 10

	// stateLockWaitBudget is the derived elapsed window a contender polls
	// before giving up. Bounding by elapsed time rather than by an attempt
	// count is what makes the window a duration the derivation can state.
	stateLockWaitBudget = stateLockSupportedWriters * stateLockCIMutationCost * stateLockHeadroom

	// stateLockWaitMin and stateLockWaitMax bound every per-attempt wait the
	// policy produces. They are the policy's declared contract; tests assert
	// against these rather than against a sampled value.
	stateLockWaitMin = 5 * time.Millisecond
	stateLockWaitMax = 50 * time.Millisecond

	// stateLockWaitStep grows the jitter band per attempt, so early retries
	// poll tightly and a long wait backs off toward the ceiling.
	stateLockWaitStep = 10 * time.Millisecond
)

// stateLockRetryWait returns the wait before retry number attempt (0-based).
//
// It is jitter applied over a modest linear backoff, and both halves are
// load-bearing. Backoff alone would NOT break the lockstep: contenders
// released together grow their delays identically and keep arriving at the
// same bad moment relative to the holder's release. Jitter is what makes two
// contenders at the same attempt index diverge, so no contender is
// systematically beaten by the same peers (REQ-BLB-003/004).
//
// The randomness comes from math/rand/v2's per-call top-level source: no
// package-level seeding, and no global generator a test would have to
// control.
func stateLockRetryWait(attempt int) time.Duration {
	ceil := stateLockWaitMin + time.Duration(attempt+1)*stateLockWaitStep
	if ceil > stateLockWaitMax {
		ceil = stateLockWaitMax
	}
	span := ceil - stateLockWaitMin
	if span <= 0 {
		return stateLockWaitMin
	}
	return stateLockWaitMin + time.Duration(rand.Int64N(int64(span)+1))
}

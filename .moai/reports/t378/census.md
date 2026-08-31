# M2 — Guard census + pre-plant predictions (card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `226bdd0dc`.
Every prediction below was written BEFORE its mutant was planted, so a surprise is visible as a
surprise.

## Whole-package GREEN baseline

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	16.677s`
Exit: 0. Every later RED is a delta against this line, not against memory. The run is serial,
scoped to the one package, and the race detector is not enabled (REQ-VFG-007).

## Budget arithmetic per mutant, against the 660ms composed floor

The four retained assertions compose `budget >= 10 x 33ms x 2 = 660ms` (see repair-direction.md).
Landed budget: `10 x 33ms x 5 = 1650ms`.

| Mutant | Change | Resulting budget | vs 660ms floor |
|---|---|---|---|
| M1 | `boardLockCIMutationCost` 33ms -> 20ms | `10 x 20ms x 5` = **1000ms** | ABOVE |
| M2 | `boardLockHeadroom` 5 -> 1 | `10 x 33ms x 1` = **330ms** | **BELOW** |
| M3 | `boardLockSupportedWriters` 10 -> 8 | `8 x 33ms x 5` = **1320ms** | ABOVE |
| M4 | budget declaration -> `1400 * time.Millisecond` | **1400ms** | ABOVE |

This is why AC-VFG-007 requires M2 and no other: only M2 puts the budget below the floor a
genuine guard would enforce. Under M1, M3, or M4 a real floor would have stayed silent too, so
the reinstated branch's silence would evidence nothing.

## Census 1 — `boardLockCIMutationCost`

Command: `grep -rn 'boardLockCIMutationCost' internal/kanban/`

```
internal/kanban/board_store.go:101:	// boardLockCIMutationCost is the per-mutation cost observed on a
internal/kanban/board_store.go:107:	boardLockCIMutationCost = 33 * time.Millisecond
internal/kanban/board_store.go:126:	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
internal/kanban/integration_lock_cross_test.go:54:	// boardLockWaitBudget = boardLockSupportedWriters × boardLockCIMutationCost
internal/kanban/board_lock_wait_test.go:27:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:31:			boardLockWaitBudget, boardLockSupportedWriters, boardLockCIMutationCost,
internal/kanban/board_lock_wait_test.go:38:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:54:	if boardLockCIMutationCost < 33*time.Millisecond {
internal/kanban/board_lock_wait_test.go:56:			boardLockCIMutationCost)
internal/kanban/board_lock_wait_test.go:69:// BUDGET GUARD. boardLockCIMutationCost appears on both sides of the
internal/kanban/board_lock_wait_test.go:84:// boardLockCIMutationCost, and no per-mutation cost regression of any size,
internal/kanban/board_lock_wait_test.go:92:// (boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:100:	floor := time.Duration(serialized) * boardLockCIMutationCost
internal/kanban/board_lock_wait_test.go:111:			boardLockWaitBudget, floor, boardLockCIMutationCost)
internal/kanban/board_lock_wait_test.go:117:		"change to boardLockCIMutationCost (declared %v) would not be caught here.",
internal/kanban/board_lock_wait_test.go:119:		stressWriters, stressAddsPerWriter, serialized, boardLockCIMutationCost)
```

**Prediction (M1, cost 33ms -> 20ms):** exactly ONE guard reddens —
`TestBoardLockWaitBudgetDerivedFromNamedInputs`, at line 54, message
`per-mutation cost 20ms is below the CI-class observation of 33ms`.

t372's `TestBoardLockWaitBudgetCoversSerializedMutations` is expected GREEN, and the reason is
structural rather than numeric: its floor is `48 * cost` and the budget is `50 * cost`, so the
cost cancels and `50*cost < 48*cost` is false for every positive cost. The guard is
cost-independent by construction (its own header comment says so at lines 69 and 84).

The derivation equality at line 28 is also expected GREEN: the budget declaration recomputes
from the mutated constant, so both sides move together.

The budget drops to 1000ms, still above the 660ms floor, so a contention regression is not
expected; if one appears it will be named and attributed rather than dismissed.

## Census 2 — `boardLockHeadroom`

Command: `grep -rn 'boardLockHeadroom' internal/kanban/`

```
internal/kanban/board_store.go:109:	// boardLockHeadroom is the stated headroom factor over the product
internal/kanban/board_store.go:113:	// The product boardLockSupportedWriters * boardLockHeadroom = 50 is the
internal/kanban/board_store.go:121:	boardLockHeadroom = 5
internal/kanban/board_store.go:126:	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
internal/kanban/integration_lock_cross_test.go:55:	// × boardLockHeadroom = 10 × 33ms × 5 = 1.65s (board_store.go:96-117).
internal/kanban/board_lock_wait_test.go:27:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:32:			boardLockHeadroom, recomputed)
internal/kanban/board_lock_wait_test.go:38:		boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:59:	if boardLockHeadroom < 2 {
internal/kanban/board_lock_wait_test.go:60:		t.Errorf("headroom factor %d states no headroom", boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:73://	boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter
internal/kanban/board_lock_wait_test.go:77:// lowering boardLockSupportedWriters or boardLockHeadroom, or raising
internal/kanban/board_lock_wait_test.go:92:// (boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:99:	const policyBudgetedMutations = boardLockSupportedWriters * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:109:			boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,
internal/kanban/board_lock_wait_test.go:118:		boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,
```

**Prediction (M2, headroom 5 -> 1):** MULTIPLE guards redden. Attribution rests on the verbatim
message, never on the count.

1. `TestBoardLockWaitBudgetDerivedFromNamedInputs` line 59 — `headroom factor 1 states no headroom`.
   This is the named assertion for AC-VFG-004.
2. `TestBoardLockWaitBudgetCoversSerializedMutations` (t372's guard) — expected RED: the policy
   product becomes `10 * 1 = 10`, well under the serialized `8 * 6 = 48`
   (`330ms < 48 x 33ms = 1584ms`). This is a predicted, legitimate additional RED, not noise.
3. Contention tests MAY redden. The budget is consumed as a real wait deadline at
   `board_store.go:167`, `integration_lock_mutation.go:97`, and `backlog_store.go:730`, so a
   330ms budget against 48 serialized mutations that need roughly 1.58s can starve
   `TestConcurrencyStress` and possibly the integration-lock tests.

## Census 3 — `boardLockSupportedWriters`

Command: `grep -rn 'boardLockSupportedWriters' internal/kanban/`

```
internal/kanban/board_store.go:96:	// boardLockSupportedWriters is the concurrent lane count the product
internal/kanban/board_store.go:99:	boardLockSupportedWriters = 10
internal/kanban/board_store.go:113:	// The product boardLockSupportedWriters * boardLockHeadroom = 50 is the
internal/kanban/board_store.go:126:	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
internal/kanban/integration_lock_cross_test.go:54:	// boardLockWaitBudget = boardLockSupportedWriters × boardLockCIMutationCost
internal/kanban/board_lock_wait_test.go:26:	recomputed := time.Duration(boardLockSupportedWriters) *
internal/kanban/board_lock_wait_test.go:31:			boardLockWaitBudget, boardLockSupportedWriters, boardLockCIMutationCost,
internal/kanban/board_lock_wait_test.go:37:	floor := time.Duration(boardLockSupportedWriters) *
internal/kanban/board_lock_wait_test.go:45:	if boardLockSupportedWriters != 10 {
internal/kanban/board_lock_wait_test.go:47:			boardLockSupportedWriters)
internal/kanban/board_lock_wait_test.go:73://	boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter
internal/kanban/board_lock_wait_test.go:77:// lowering boardLockSupportedWriters or boardLockHeadroom, or raising
internal/kanban/board_lock_wait_test.go:92:// (boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom)
internal/kanban/board_lock_wait_test.go:99:	const policyBudgetedMutations = boardLockSupportedWriters * boardLockHeadroom
internal/kanban/board_lock_wait_test.go:109:			boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,
internal/kanban/board_lock_wait_test.go:118:		boardLockSupportedWriters, boardLockHeadroom, policyBudgetedMutations,
```

**Prediction (M3, writers 10 -> 8):** TWO guards redden.

1. `TestBoardLockWaitBudgetDerivedFromNamedInputs` line 45 —
   `supported writers = 8, want 10 (Factory mode's ten lanes against one queue)`. Named assertion
   for AC-VFG-005.
2. t372's guard — expected RED: policy product `8 * 5 = 40 < 48`
   (`1320ms < 1584ms`). Predicted, legitimate.

Budget lands at 1320ms, above the 660ms floor, so contention starvation is not expected.

## Census 4 — `boardLockWaitBudget`

Command: `grep -rn 'boardLockWaitBudget' internal/kanban/`

```
internal/kanban/board_store.go:123:	// boardLockWaitBudget is the derived elapsed window a contender polls
internal/kanban/board_store.go:126:	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
internal/kanban/board_store.go:167:	deadline := time.Now().Add(boardLockWaitBudget)
internal/kanban/integration_lock_mutation.go:97:	deadline := time.Now().Add(boardLockWaitBudget)
internal/kanban/integration_lock_mutation.go:126:			return nil, fmt.Errorf("%w (waited %s): %v", ErrIntegrationLockBusy, boardLockWaitBudget, lastErr)
internal/kanban/integration_lock_cross_test.go:54:	// boardLockWaitBudget = boardLockSupportedWriters × boardLockCIMutationCost
internal/kanban/integration_lock_cross_test.go:351:	busy := fmt.Errorf("%w (waited %s): %v", ErrIntegrationLockBusy, boardLockWaitBudget, ErrBoardLockHeld)
internal/kanban/backlog_concurrency_test.go:123:// exhausted the machine-speed-sensitive boardLockWaitBudget, which measures the
internal/kanban/backlog_concurrency_test.go:252:// are attempted, so the bounded boardLockWaitBudget wait is paid twice rather
internal/kanban/backlog_store.go:719:// (boardLockWaitBudget and boardLockRetryWait, board_store.go — REQ-BLB-006:
internal/kanban/backlog_store.go:730:	deadline := time.Now().Add(boardLockWaitBudget)
internal/kanban/board_lock_wait_test.go:28:	if boardLockWaitBudget != recomputed {
internal/kanban/board_lock_wait_test.go:39:	if boardLockWaitBudget < floor {
internal/kanban/board_lock_wait_test.go:40:		t.Errorf("budget %v < headroom floor %v", boardLockWaitBudget, floor)
internal/kanban/board_lock_wait_test.go:102:	if boardLockWaitBudget < floor {
internal/kanban/board_lock_wait_test.go:111:			boardLockWaitBudget, floor, boardLockCIMutationCost)
internal/kanban/board_lock_wait_test.go:198:	if bound := 2 * boardLockWaitBudget; elapsed > bound {
internal/kanban/board_lock_wait_test.go:200:			elapsed, bound, boardLockWaitBudget)
```

**Prediction (M4):** the mutant has two forms and they behave differently.

- `1650 * time.Millisecond` (numerically identical to the landed value): reddens NOTHING. The
  assertion at line 28 compares VALUES, not syntax, so a bare literal that happens to equal the
  product passes. This is the stated, honest limit of the mutant, recorded rather than hidden.
- `1400 * time.Millisecond`: reddens TWO guards —
  (a) the derivation equality at line 28, message containing
      `is not the product of its named inputs`, which is the named assertion for AC-VFG-006;
  (b) t372's guard, because `1400ms < 48 x 33ms = 1584ms`.

Attribution therefore rests on the verbatim message of the named assertion, never on the count.

## A note the census surfaced, out of scope for this card

`internal/kanban/integration_lock_cross_test.go` lines 54-55 carry a COMMENT restating the budget
arithmetic (`10 × 33ms × 5 = 1.65s`). It is prose, not an assertion, so it reddens under no
mutant. It is recorded here as an observed documentation surface that tracks these constants; this
card does not touch it (REQ-VFG-006 scope).

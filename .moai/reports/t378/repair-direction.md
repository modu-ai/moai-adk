# M1 — Repair direction confirmed against the tree (card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `226bdd0dc`,
base `3f03d9c36`. All line numbers below are from this tree, anchored by assertion text.

## Coverage table — re-verified row by row (NOT carried from spec.md)

Command: `grep -n 'boardLockWaitBudget != recomputed\|boardLockSupportedWriters != 10\|boardLockCIMutationCost < 33\|boardLockHeadroom < 2' internal/kanban/board_lock_wait_test.go`

| REQ-BLB-002 clause | enforcer | line | verified |
|---|---|---|---|
| ten-lane contender count | `if boardLockSupportedWriters != 10 {` | 45 | YES |
| CI-class per-mutation cost | `if boardLockCIMutationCost < 33*time.Millisecond {` | 54 | YES |
| stated headroom factor | `if boardLockHeadroom < 2 {` | 59 | YES |
| budget IS that product | `if boardLockWaitBudget != recomputed {` (t.Fatalf) | 28 | YES |

## The budget declaration pins the equality completely

Command: `grep -n 'boardLockWaitBudget' internal/kanban/board_store.go`
Observed (line 126):

    boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom

The budget IS the product of the three constants, so the line-28 equality pins it exactly. Any
floor built from those same three terms therefore carries no information the equality does not
already carry: `budget < floor` where `floor` is the identical expression to `recomputed`, and
`budget == recomputed` was asserted twelve lines above with a `t.Fatalf`, is false on every
assignment. That is the vacuity.

## The composed floor is real and input-wise

    budget == writers * cost * headroom   (line 28, t.Fatalf)
    writers == 10                          (line 45)
    cost    >= 33ms                        (line 54)
    headroom>= 2                           (line 59)
    ==> budget >= 10 * 33ms * 2 = 660ms

Each conjunct is falsifiable by a change to exactly one constant (mutants M1/M2/M3), and the
equality itself is falsifiable by replacing the declaration with a non-derived literal (M4).
REQ-BLB-002's floor obligation is therefore discharged input-wise by the four retained
assertions, not by any comparison against `budget`.

## REQ-BLB-002 text re-read at this tree

`.moai/specs/SPEC-BACKLOG-LOCK-BUDGET-001/spec.md`:

> **Where** the product supports up to ten concurrent lane writers against one queue [...], the
> derived budget shall account for that contender count and shall exceed the per-mutation cost
> observed on a CI-class machine by at least the headroom factor REQ-BLB-001 states.

Matches the mapping above.

## Conclusion

**Deletion.** No row failed to verify, so the repair direction does NOT flip to "derive a floor
from outside terms". The `floor :=` declaration and the `if boardLockWaitBudget < floor` block
are removed; the comment above them is rewritten in place to record where REQ-BLB-002's floor is
actually enforced and why no floor-versus-budget comparison exists in this function.

Gap: this milestone establishes only that the branch is dominated by the equality. It does NOT
establish that no other guard elsewhere depends on the branch — that is what the M2 census reads.

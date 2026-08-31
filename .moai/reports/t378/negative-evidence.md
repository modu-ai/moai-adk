# M4 — The deletion's negative evidence (AC-VFG-007, card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `226bdd0dc`.

## Why this observation was taken before the edit

The criterion asks for the deleted branch "temporarily reinstated". It was cheaper and strictly
more faithful to take the observation while the branch was still present in its landed form — no
reinstatement was needed, so there is no risk of the reinstated copy differing from the original.
The branch under observation is the verbatim landed code at lines 37-41:

```go
	// The inequality REQ-BLB-002 states: at least headroom x per-mutation
	// cost x supported contender count.
	floor := time.Duration(boardLockSupportedWriters) *
		boardLockCIMutationCost * boardLockHeadroom
	if boardLockWaitBudget < floor {
		t.Errorf("budget %v < headroom floor %v", boardLockWaitBudget, floor)
	}
```

## Why M2 and no other mutant

The four retained assertions compose a real floor of `10 x 33ms x 2 = 660ms` (repair-direction.md).
For the branch's silence to mean anything, the budget must fall BELOW that floor — otherwise a
genuine floor would have been silent too, and silence would evidence nothing.

| Mutant | Budget | vs 660ms | Would a genuine floor fire? |
|---|---|---|---|
| M1 cost 33ms→20ms | 1000ms | above | NO — silence proves nothing |
| **M2 headroom 5→1** | **330ms** | **below** | **YES — silence is a contrast** |
| M3 writers 10→8 | 1320ms | above | NO — silence proves nothing |
| M4 budget→1400ms | 1400ms | above | NO — silence proves nothing |

M2 is therefore the required mutant and the choice is not interchangeable.

## The observation

Mutant planted (`internal/kanban/board_store.go`):

```
-	boardLockHeadroom = 5
+	boardLockHeadroom = 1
```

Command:
`go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`

Observed (exit 1) — this is the COMPLETE output, not a filtered excerpt. That matters: an absence
claim read off a truncated capture would establish nothing.

```
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
    board_lock_wait_test.go:60: headroom factor 1 states no headroom
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	0.234s
```

**Selector match count: 1.** `=== RUN TestBoardLockWaitBudgetDerivedFromNamedInputs` is present,
so this is not a zero-match selector printing a vacuous result.

**The two-sided reading, which is the whole point:**

- PRESENT: `headroom factor 1 states no headroom` (line 60) — the input-wise headroom assertion
  fired, so the function DID execute past the equality and DID reach the region where the floor
  branch sits.
- ABSENT: no line from `board_lock_wait_test.go:40`, and no occurrence of the branch's message
  `budget ... < headroom floor ...`, anywhere in the output.

The budget at this moment is `10 x 33ms x 1 = 330ms`, half the 660ms floor the retained
assertions compose. A floor that meant anything would have fired here. This one could not,
because its own `floor` term is the identical expression to the budget and tracked the mutation
down to 330ms alongside it: `330ms < 330ms` is false.

## Revert and scope confirmation

Revert applied, then:

Command: `git diff --stat -- internal/kanban/board_store.go`
Observed: empty (no output) — the mutant left nothing behind.

Command: `git status --short`
Observed:

```
?? .moai/reports/t378/census.md
?? .moai/reports/t378/repair-direction.md
```

No tracked file modified; only this card's own untracked evidence files.

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	16.074s` (exit 0).

## The honest limits of this criterion — stated, not buried

These three limits are load-bearing. A reader who takes this observation for more than it is
would be repeating exactly the error this card exists to repair.

1. **This corroborates; it does not prove.** The every-assignment unreachability claim rests on
   the STATIC argument (spec.md §A.1): `floor` is the identical expression to `recomputed`, and
   `budget == recomputed` is asserted twelve lines above with a `t.Fatalf` hard stop, so
   `budget < floor` is false on every assignment of the three constants. This run corroborates
   that argument on exactly ONE assignment. A single observation of silence never establishes
   that a branch could not be made RED.

2. **A deletion carries no positive mutant evidence of its own.** There is no mutant that makes a
   removed branch fire. Every other AC in this SPEC is discharged by observing a RED; this one
   cannot be, by construction. The four retained assertions carry the positive evidence
   (mutants.md); this criterion carries only the negative half.

3. **"660ms is looser than 1650ms" is an illusion the vacuous branch created, and the objection a
   deletion review will raise.** It reads as though deletion trades a 1650ms bound for a weaker
   660ms one. It does not. The 1650 was never a bound: `floor` is the same expression as
   `budget`, so it never functioned as a floor at all — it was an identity wearing an
   inequality's clothing. Nothing is being loosened, because nothing was being enforced there.
   660ms is the first real floor this function has ever had, composed input-wise by the four
   retained assertions, and every one of its conjuncts is demonstrated falsifiable in mutants.md.

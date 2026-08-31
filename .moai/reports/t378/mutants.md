# M3 — Mutant evidence for the four retained assertions (card t378)

Tree: worktree `.claude/worktrees/t378`, branch `WT-vacuous-floor-guard`, HEAD `226bdd0dc`.
Every observation below was taken with the floor branch STILL PRESENT (pre-edit), one mutant at a
time, each reverted before the next was planted. Every run is a single serial invocation scoped to
`./internal/kanban/`, without the race detector, with no background process.

**Attribution rule applied throughout (REQ-VFG-005):** the named assertion is attributed by its
VERBATIM failure message, never by the number of failing tests. M2, M3, and M4 each redden more
than one guard; the census (census.md) predicted that in advance.

---

## M1 — the CI-class per-mutation cost floor (AC-VFG-003)

**Prediction (written before planting):** exactly ONE guard reddens, at line 54.

Mutant diff:

```
-	boardLockCIMutationCost = 33 * time.Millisecond
+	boardLockCIMutationCost = 20 * time.Millisecond
```

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed (RED, exit 1) — complete output:

```
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
    board_lock_wait_test.go:55: per-mutation cost 20ms is below the CI-class observation of 33ms
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	14.761s
FAIL
```

**Attribution:** the named assertion is `board_lock_wait_test.go:55`, message
`per-mutation cost 20ms is below the CI-class observation of 33ms`. It is the only RED in the run.

**Prediction held exactly.** t372's guard stayed GREEN, confirming the structural reason recorded
in the census: its floor is `48 * cost` against a budget of `50 * cost`, so the cost cancels and no
cost change of any size can make it fire. The derivation equality also stayed GREEN, because the
budget declaration recomputes from the mutated constant and both sides moved together.

Revert: `git diff --stat -- internal/kanban/board_store.go` → empty (no output).
Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
Post-revert observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	15.975s` (exit 0).

---

## M2 — the headroom floor (AC-VFG-004)

**Prediction (written before planting):** MULTIPLE guards redden — (1) line 59 named assertion,
(2) t372's guard, (3) contention tests MAY redden because the budget falls to 330ms.

Mutant diff:

```
-	boardLockHeadroom = 5
+	boardLockHeadroom = 1
```

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed (RED, exit 1) — complete output:

```
--- FAIL: TestIntegrationLockAcquire_SerializedAcrossProcesses (0.38s)
    integration_lock_cross_test.go:266: A: RESULT=acquired REPLACED=none SESSION=lane-a
    integration_lock_cross_test.go:267: B: RESULT=busy REPLACED=none SESSION=lane-b
    integration_lock_cross_test.go:268: round: successes=1 refusals=0 other=1 sessions_differ=true mid_record_held=false mid_record_stale=false final_holder="lane-a" final_record_stale=false
    integration_lock_cross_test.go:272: a child reported neither acquired nor held (successes=1 refusals=0 other=1) — RESULT=busy means the stall-release timeout (500ms) was not comfortably shorter than the mutation-lock wait budget, and the harness measured its own configuration rather than the lock. A: "RESULT=acquired REPLACED=none SESSION=lane-a" B: "RESULT=busy REPLACED=none SESSION=lane-b"
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
    board_lock_wait_test.go:60: headroom factor 1 states no headroom
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
    board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 10 supported writers x 1 headroom = 10 serialized mutations, while the stress test serializes 8 x 6 = 48 (330ms budget < 1.584s floor). [...]
--- FAIL: TestBacklogLock_TimeoutNamesLockPath (0.35s)
    backlog_store_test.go:205: Add returned after 346.110125ms, before the ~1s bounded window elapsed
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	13.270s
FAIL
```

**Attribution — four REDs, each named and assigned to its own guard:**

| Failing guard | Line | Why it fired | Predicted? |
|---|---|---|---|
| `TestBoardLockWaitBudgetDerivedFromNamedInputs` | 60 | **the named assertion for this AC** — `headroom factor 1 states no headroom` | YES, specifically |
| `TestBoardLockWaitBudgetCoversSerializedMutations` | 103 | policy product `10 x 1 = 10 < 48`; `330ms < 1.584s` | YES, specifically |
| `TestIntegrationLockAcquire_SerializedAcrossProcesses` | 272 | the 330ms budget is no longer comfortably longer than the harness's own 500ms stall-release timeout | YES, as a category (census 2 item 3 named the integration-lock tests as possible) |
| `TestBacklogLock_TimeoutNamesLockPath` | 205 | the bounded window collapses from ~1.65s to 330ms, so `Add` returns at 346ms before the expected ~1s | YES, as a category (the budget is a real deadline at `backlog_store.go:730`) |

The count 4 is NOT the attribution. The attribution is the verbatim string
`headroom factor 1 states no headroom` at `board_lock_wait_test.go:60`.

**One prediction did NOT hold, recorded as a surprise rather than smoothed over.** The census
predicted `TestConcurrencyStress` might starve under a 330ms budget. It did not redden. The
prediction was hedged ("MAY"), so this is not a contradiction, but it is recorded because the
point of writing predictions first is that a miss stays visible. Two OTHER budget-consuming tests
reddened instead, both inside the predicted category.

Revert: `git diff --stat -- internal/kanban/board_store.go` → empty (no output).
Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
Post-revert observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	16.074s` (exit 0).

---

## M3 — the ten-lane contender-count assertion (AC-VFG-005)

**Prediction (written before planting):** TWO guards redden — line 45 named assertion, plus
t372's guard (`8 x 5 = 40 < 48`).

Mutant diff:

```
-	boardLockSupportedWriters = 10
+	boardLockSupportedWriters = 8
```

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed (RED, exit 1) — complete output:

```
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
    board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 8 supported writers x 5 headroom = 40 serialized mutations, while the stress test serializes 8 x 6 = 48 (1.32s budget < 1.584s floor). [...]
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
    board_lock_wait_test.go:46: supported writers = 8, want 10 (Factory mode's ten lanes against one queue)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	15.481s
FAIL
```

**Attribution:** the named assertion is `board_lock_wait_test.go:46`, message
`supported writers = 8, want 10 (Factory mode's ten lanes against one queue)`. The second RED is
t372's guard, predicted, with its own distinct message and arithmetic (`40 < 48`).

**Prediction held exactly**, including the count and both arithmetic figures. No contention test
reddened, consistent with the budget landing at 1320ms — above the 660ms composed floor.

Revert: `git diff --stat -- internal/kanban/board_store.go` → empty (no output).
Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
Post-revert observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	16.076s` (exit 0).

---

## M4 — the derivation equality (AC-VFG-006), in two forms

**Prediction (written before planting):** the `1650ms` form reddens NOTHING; the `1400ms` form
reddens TWO guards.

### Form A — `1650 * time.Millisecond`, numerically identical to the landed value

Mutant diff:

```
-	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
+	boardLockWaitBudget = 1650 * time.Millisecond
```

Command:
`go test -timeout 600s -count=1 -v -run 'TestBoardLockWaitBudgetDerivedFromNamedInputs|TestBoardLockWaitBudgetCoversSerializedMutations' ./internal/kanban/`
Observed (GREEN, exit 0) — complete output:

```
=== RUN   TestBoardLockWaitBudgetDerivedFromNamedInputs
=== PAUSE TestBoardLockWaitBudgetDerivedFromNamedInputs
=== RUN   TestBoardLockWaitBudgetCoversSerializedMutations
=== PAUSE TestBoardLockWaitBudgetCoversSerializedMutations
=== CONT  TestBoardLockWaitBudgetDerivedFromNamedInputs
--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
=== CONT  TestBoardLockWaitBudgetCoversSerializedMutations
    board_lock_wait_test.go:114: constant coherence: the lock policy budgets 10 supported writers x 5 headroom = 50 serialized mutations; the stress test serializes 8 x 6 = 48. [...]
--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.229s
```

Selector match count: **2** (`=== RUN` appears twice) — this is not a zero-match selector printing
a vacuous `ok`.

**THE STATED GAP, observed rather than assumed.** The budget declaration was genuinely replaced by
a bare literal with no derivable inputs — precisely the regression REQ-BLB-001 exists to catch —
and BOTH guards passed. The reason is structural: the assertion at line 28 compares VALUES, not
SYNTAX, so any literal numerically equal to the product satisfies it. This is a real limit of the
retained equality and is recorded here rather than hidden.

### Form B — `1400 * time.Millisecond`

Mutant diff:

```
-	boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom
+	boardLockWaitBudget = 1400 * time.Millisecond
```

Command: `go test -timeout 600s -count=1 ./internal/kanban/`
Observed (RED, exit 1) — complete output:

```
--- FAIL: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)
    board_lock_wait_test.go:103: constant coherence broken: the lock policy budgets 10 supported writers x 5 headroom = 50 serialized mutations, while the stress test serializes 8 x 6 = 48 (1.4s budget < 1.584s floor). [...]
--- FAIL: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)
    board_lock_wait_test.go:29: budget 1.4s is not the product of its named inputs (10 writers x 33ms x 5 headroom = 1.65s) — a bare literal with no derivable inputs fails REQ-BLB-001
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/kanban	15.924s
FAIL
```

**Attribution:** the named assertion is `board_lock_wait_test.go:29`, message containing
`is not the product of its named inputs`. The second RED is t372's guard, predicted
(`1400ms < 1584ms`).

**Prediction held exactly for both forms.** Note what the two forms establish together: the
equality is capable of RED (form B), and it is capable of missing a real derivability regression
(form A). Both facts are properties of this guard and both are recorded.

Revert: `git status --short` after reverting showed only the two untracked report files —
no tracked file modified:

```
?? .moai/reports/t378/census.md
?? .moai/reports/t378/repair-direction.md
```

Post-revert command: `go test -timeout 600s -count=1 ./internal/kanban/`
Post-revert observed: `ok  	github.com/modu-ai/moai-adk/internal/kanban	15.995s` (exit 0).

---

## Summary

| AC | Mutant | Named assertion | Line | RED observed | Reverted + GREEN |
|---|---|---|---|---|---|
| AC-VFG-003 | M1 cost 33ms→20ms | `per-mutation cost 20ms is below the CI-class observation of 33ms` | 55 | YES | YES |
| AC-VFG-004 | M2 headroom 5→1 | `headroom factor 1 states no headroom` | 60 | YES | YES |
| AC-VFG-005 | M3 writers 10→8 | `supported writers = 8, want 10 (Factory mode's ten lanes against one queue)` | 46 | YES | YES |
| AC-VFG-006 | M4 budget→1400ms | `is not the product of its named inputs` | 29 | YES | YES |

All four retained assertions are demonstrated capable of RED. None is vacuous.

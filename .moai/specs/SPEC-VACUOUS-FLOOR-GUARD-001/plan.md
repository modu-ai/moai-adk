# Implementation Plan — SPEC-VACUOUS-FLOOR-GUARD-001 (card t378)

Tier S. One file, one function. Milestones are ordered by **decision reversibility**: the
irreversible judgement (delete versus repair) leads, the evidence that could overturn it comes
next, and the mechanical edit is last.

## §A Context

`internal/kanban/board_lock_wait_test.go` carries an unreachable `budget < floor` branch inside
`TestBoardLockWaitBudgetDerivedFromNamedInputs`. Full statement of the defect, the evidence, and
the repair argument: `spec.md` §A. Base for every measurement below: `3f03d9c36`
(= `origin/develop` at card start), branch `WT-vacuous-floor-guard`.

## §B Known issues and constraints entering run-phase

1. **t372's guard is untouchable.** `TestBoardLockWaitBudgetCoversSerializedMutations` sits in the
   same file, sixty lines below the edit, and is under an open observation window (AC-SIV-013). It
   must come out of the diff byte-identical.
2. **Mutants touch production constants.** M1-M4 modify `internal/kanban/board_store.go`
   transiently. Every one is reverted before the next is planted, and the committed diff must show
   `board_store.go` unchanged. This is the single highest-risk mechanical step of the card: a
   forgotten revert silently retunes the lock policy.
3. **Shared-constant mutants redden more than one guard.** M2 and M3 are predicted to redden
   t372's guard as well, and M2 may redden `TestConcurrencyStress`. Attribution therefore rests on
   the named assertion's verbatim message, never on a failure count (REQ-VFG-005).
4. **Verification load.** `internal/kanban` carries contention tests. One serial `go test` per
   observation, scoped to the package, without `-race`. No `go test ./...`, no background process
   (REQ-VFG-007).

## §C Pre-flight

- `git rev-parse --short HEAD` == `3f03d9c36`, `git branch --show-current` == `WT-vacuous-floor-guard`.
- `mkdir -p .moai/reports/t378` — the evidence directory the ACs cite.
- Baseline greps recorded BEFORE the edit, so AC-VFG-001's "1 remaining" is a delta rather than an
  absolute: `grep -c 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go` and
  `grep -c 'floor :=' internal/kanban/board_lock_wait_test.go` (both expected `2` pre-repair).
- Whole-package GREEN baseline: `go test -timeout 600s -count=1 ./internal/kanban/`, recorded with
  its exit code. Every later RED is a delta against this, not against memory.

## §D Milestones

Ordered most-reversible-decision-first.

### M1 — Confirm the repair direction against the tree (highest change likelihood)

The decision that everything else rests on, and the only one a reader is likely to overturn: is
REQ-BLB-002's floor genuinely discharged by the four input-wise assertions, or does deleting the
branch drop coverage?

Steps:
1. Re-read `SPEC-BACKLOG-LOCK-BUDGET-001` §B REQ-BLB-001/REQ-BLB-002 at this tree and confirm the
   §A.3 mapping table row by row against the current line numbers of
   `board_lock_wait_test.go` (line numbers drift; anchor on the assertion text).
2. Confirm from the source that `boardLockWaitBudget` is declared as the product of the three
   constants (`internal/kanban/board_store.go`), so the line-28 equality pins it completely and no
   same-terms floor can add information.
3. Record the conclusion in `.moai/reports/t378/repair-direction.md`: deletion, with the coverage
   table and the "any same-terms floor is dominated by the equality" argument.

Exit: the coverage table is verified against the tree, not carried from `spec.md`. If any row
fails to verify — a retained assertion is weaker than §A.3 claims — **stop and return a blocker**:
the repair direction flips to "derive a floor from outside terms", and the SPEC needs amending
before any edit.

### M2 — Guard census (before any mutant is planted)

Steps:
1. `grep -rn 'boardLockCIMutationCost' internal/kanban/`
2. `grep -rn 'boardLockHeadroom' internal/kanban/`
3. `grep -rn 'boardLockSupportedWriters' internal/kanban/`
4. `grep -rn 'boardLockWaitBudget' internal/kanban/`

Record each command with its full output to `.moai/reports/t378/census.md`, and for each constant
write the **predicted** set of guards that will redden, with the reason (cost cancels in t372's
guard; policy-product breach reddens it; a 330ms budget may starve `TestConcurrencyStress`; the
`1400ms` budget literal reddens t372's guard too, since `1400 < 48 x 33 = 1584`).
Predictions are written before the mutants run, so a surprise is visible as a surprise.

Alongside the censuses, record the per-mutant budget arithmetic against the 660ms composed floor —
M1 1000ms, M2 330ms, M3 1320ms — because M4 (§M4 below) is only meaningful under the one mutant that
lands below it.

Exit: four censuses recorded, four predictions written down, the budget arithmetic recorded.

### M3 — Mutant evidence for the retained assertions (AC-VFG-003..006)

For each of M1..M4 in `spec.md` §C, in order, one at a time:
plant -> `go test -timeout 600s -count=1 ./internal/kanban/` (one serial run) -> record verbatim
RED -> attribute by the named assertion's message -> name and attribute every other RED against
the M2 prediction -> revert -> re-run -> record GREEN.

Mutants: cost `33ms -> 20ms`; headroom `5 -> 1`; supported writers `10 -> 8`; budget declaration
`-> 1400 * time.Millisecond` (with the `1650 * time.Millisecond` no-op form recorded as the stated
gap). All in `internal/kanban/board_store.go`, all reverted.

Exit: `git status --short` shows `board_store.go` clean; four RED/GREEN pairs in
`.moai/reports/t378/mutants.md`.

### M4 — The deletion's negative evidence (AC-VFG-007)

With the branch still present (pre-edit) and **mutant M2** (`boardLockHeadroom` 5 -> 1) planted, run
the scoped selector with `-v` and show the branch's message absent from the output while
`headroom factor 1 states no headroom` is present. Revert the mutant. This is the observation the
deletion rests on, and it is cheapest to take before the branch is gone.

M2 is the required mutant and the choice is not interchangeable: only M2 drives the budget below the
660ms floor the four retained assertions compose (330ms). Under M1 the budget is 1000ms and under M3
it is 1320ms, both above 660ms, so a genuine floor would have been silent there too and the run
would evidence nothing (spec.md §A.5).

Exit: recorded in `.moai/reports/t378/negative-evidence.md`, with the stated limit that a deletion
carries no positive mutant evidence of its own.

### M5 — The edit (mechanical, lowest change likelihood)

1. Delete the `floor :=` declaration and the `if boardLockWaitBudget < floor` block that follows it
   in `internal/kanban/board_lock_wait_test.go`. The comment above them is NOT deleted — step 2
   rewrites it in place (REQ-VFG-001 scope clause + REQ-VFG-004).
2. Rewrite that comment to state REQ-BLB-002's actual enforcement site — the four
   retained assertions — and why no floor-versus-budget comparison appears in this function
   (REQ-VFG-004), naming t372's guard as the file's one legitimate floor comparison.
3. `gofmt -l internal/kanban/` returns empty; `go vet ./internal/kanban/` exits 0.
4. Post-repair greps for AC-VFG-001 against the M0 baseline; scoped `-v` run for AC-VFG-002.

Exit: AC-VFG-001, AC-VFG-002, AC-VFG-008 discharged.

### M6 — Commit and close-out

`git status --short` re-read in the same call that stages; explicit pathspec only (the test file
plus `.moai/specs/SPEC-VACUOUS-FLOOR-GUARD-001/` and `.moai/reports/t378/`); `git rev-parse
--short HEAD` and `git branch --show-current` re-read immediately before committing. Commit
subject carries `t378` and the SPEC ID.

## §E Self-verification

| Check | Command | Expected |
|---|---|---|
| branch shape gone | `grep -c 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go` | `1` (baseline `2`) |
| no second floor | `grep -c 'floor :=' internal/kanban/board_lock_wait_test.go` | `1` (baseline `2`) |
| t372 guard untouched | `git diff -- internal/kanban/board_lock_wait_test.go` | no hunk inside `TestBoardLockWaitBudgetCoversSerializedMutations` |
| constants untouched | `git diff -- internal/kanban/board_store.go` | empty |
| compiles | `go vet ./internal/kanban/` | exit 0 |
| guard green | `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/` | `--- PASS`, `=== RUN` present (non-zero selector match) |
| package green | `go test -timeout 600s -count=1 ./internal/kanban/` | `ok` |
| verification load | `grep -rn 'go test' .moai/reports/t378/` | every invocation scoped to `./internal/kanban/`; no `./...`, no `-race`, no trailing `&` |
| lint | `./bin/moai spec lint --strict` scoped to this SPEC (tree binary, not the PATH one) | 0 errors |

## §F Risks

| Risk | Mitigation |
|---|---|
| A mutant is left planted | M3 exits on `git status --short` showing `board_store.go` clean; M6 re-reads `git diff` on it before staging |
| Deletion drops real coverage | M1 is a stop-and-blocker gate before any edit; the coverage table is verified against the tree |
| Multi-guard RED read as single attribution | M2 predictions written before M3 runs; attribution by verbatim message (REQ-VFG-005) |
| Contention test reddened by the M2 headroom mutant | Predicted in advance and recorded as an expected additional RED, not treated as noise |
| Verification generates load | One serial scoped run per observation; no `-race`, no `./...`, no background process |

## §G Anti-patterns

- Replacing the deleted branch with a "better" inequality built from the same three constants.
  That is the defect, restated.
- Copying t372's stress-constant floor into this function to have "a floor here too".
- Reading a whole-package RED count as attribution.
- Committing while a mutant is planted.
- Declaring the guard green from a selector run without checking the `=== RUN` lines — a
  zero-match selector prints `ok`.

## §H Cross-references

- `spec.md` §A.3 / §A.4 — the coverage table and the deletion argument M1 verifies.
- `SPEC-STRESS-INVARIANT-VERDICT-001` REQ-SIV-010, §E — the deferral this card picks up.
- `.moai/reports/t372/mutant-headroom4-orchestrator.md` — the landed dynamic evidence.

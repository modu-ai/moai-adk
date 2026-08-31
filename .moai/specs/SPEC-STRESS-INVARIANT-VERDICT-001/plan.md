# SPEC-STRESS-INVARIANT-VERDICT-001 — Implementation Plan

Card t372. Worktree `.claude/worktrees/t372`, branch `WT-stress-invariant-guard`,
base `origin/develop` = `b9149857c`.

Milestones are ordered by **decision reversibility** — the choices most likely to change on review
come first; the mechanical edits sit at the bottom.

## §A Context

Remediation branch C of card t370's three-branch finding, selected by the operator. Branches A
(re-tune the budget constant) and B (add in-process fairness) are rejected and out of scope
(`spec.md` §E). The investigation is complete; this plan re-measures nothing.

## §B Known issues carried in

- The budget's headroom coincidence (`10 * 5 = 50 ~= 48`) is currently accidental. M1 makes it
  load-bearing but does **not** enlarge it — the guard asserts coherence, not sufficiency.
- Today's numbers already satisfy the new guard: `1.65s >= 48 * 33ms = 1.584s`. So M1 lands green
  without touching any constant. That is intended, and it is also why M1's mutant (AC-SIV-008) is
  the only thing establishing the guard is not vacuous.

## §C Pre-flight

- Read `internal/kanban/backlog_concurrency_test.go` (the stress test, lines 15-95) and
  `internal/kanban/board_lock_wait_test.go` (the existing derivation guard).
- Confirm `IsBoardLockHeld` is exported from `internal/kanban/board_lock.go` and reachable from
  the test package (same package `kanban` — it is).
- `git rev-parse --short HEAD` and `git branch --show-current` immediately before any commit.

## §D Constraints

- Test-only change plus at most one comment in `board_store.go`. No production control flow moves.
- No CI runs, no load generation, no local reproduction attempts.
- No new files: the budget guard extends `board_lock_wait_test.go` (REQ-SIV-011).

---

## §F Milestones

### M1 — The latency budget guard (highest change likelihood: the derivation's shape)

**Why first**: this milestone decides *what the budget must be coherent with*. If a reviewer
disagrees with the derivation — for instance preferring the guard to key on a distinct
"serialized mutation count" constant rather than on the stress test's own `writers *
addsPerWriter` — everything downstream changes. It is also the milestone whose mutant is the
sole evidence the guard is not vacuous.

1. Promote the stress test's loop bounds to package-level constants in
   `backlog_concurrency_test.go`:
   ```
   const (
       stressWriters       = 8
       stressAddsPerWriter = 6
   )
   ```
   `TestConcurrencyStress` consumes these; no second copy exists (REQ-SIV-010).
2. Add `TestBoardLockWaitBudgetCoversSerializedMutations` to `board_lock_wait_test.go`, beside
   `TestBoardLockWaitBudgetDerivedFromNamedInputs`. It asserts
   `boardLockWaitBudget >= time.Duration(stressWriters*stressAddsPerWriter) *
   boardLockCIMutationCost` and, on failure, states the shortfall and names the coincidence the
   guard exists to make load-bearing.
3. No clock, no sleep, no goroutine in the guard (REQ-SIV-009).
4. Discharge **AC-SIV-008**: mutate `boardLockHeadroom` 5 -> 4, observe RED, revert, observe GREEN.
   Record the diff, the RED verbatim, the restoring GREEN verbatim.

Exit: guard green on the unmutated tree; AC-SIV-008 evidence recorded.

### M2 — The verdict criterion split (second-highest: this is the behavioural change under suspicion)

**Why second**: this is the edit that "looks like switching the rule off". Its shape — where
tolerance is applied and where it is not — is the reviewable decision.

1. Replace the `failures []error` accumulator's role. Classify inside the writer goroutine:
   - `IsBoardLockHeld(err)` -> increment a `starved` counter (under the existing mutex).
   - any other non-nil `err` -> record it as a hard failure.
2. Replace the `len(failures) != 0` gate at line 56 with a gate on hard failures **only**:
   `if len(hardFailures) != 0 { t.Fatalf(...) }` (REQ-SIV-002).
3. Add the zero-progress floor: `if len(issued) == 0 { t.Fatalf(...) }` (REQ-SIV-007), phrased as
   "a broken lock, not tolerable contention".
4. Keep every existing invariant assertion; M3 re-anchors them.

Exit: the test tolerates starvation and only starvation.

### M3 — Re-anchor the invariants to the issued set

**Why third**: mechanically implied by M2 (once `wantTotal` is no longer the expected count), but
it carries the SPEC's strictness obligation, so it is reviewed as a unit rather than folded in.

1. Compute `issuedCount := len(issued)` after the join.
2. Rewrite the four assertions against `issuedCount` (REQ-SIV-005 a-d):
   collision (`n != 1` per id — unchanged), presence-in-queue (unchanged),
   `len(rec.Items) != issuedCount`, `rec.LastSeq != issuedCount`.
3. Delete the `len(issued) != wantTotal` assertion — it is exactly the criterion M2 removed, and
   leaving it would silently re-fail on starvation.
4. Confirm no assertion is behind a starvation conditional (REQ-SIV-006 / AC-SIV-004).
5. Discharge **AC-SIV-009**: plant an invariant mutant (duplicate id, dropped item, or a
   `last_seq` advance that does not match the item count), observe `TestConcurrencyStress` RED with
   an *invariant* message rather than a lock message, revert, observe GREEN. Record diff, RED,
   restoring GREEN.

Exit: AC-SIV-009 evidence recorded; invariants unconditional.

### M4 — Observability

1. `t.Logf` the starved count and the back-derived per-mutation cost
   (`time.Since(start) / time.Duration(issuedCount)`), guarded against `issuedCount == 0` (which
   M2's floor already fails on).
2. Confirm no verdict depends on either figure (REQ-SIV-013 / AC-SIV-010).

### M5 — Mechanical close-out

1. Optional one-line comment in `board_store.go` above the budget block cross-referencing this
   SPEC and naming the coincidence M1 pinned. Comment only — no value changes.
2. `go vet ./internal/kanban/...`; `go test -race -count=1 ./internal/kanban/`.
3. `git diff --stat` against the base to discharge AC-SIV-012 (scope discipline), confirming no
   mutant survives in the tree.
4. Commit, message naming card `t372`. Do not push; do not open a PR.

---

## §G Anti-patterns to avoid

- **Deleting the test in disguise.** Any edit that removes an invariant assertion, or makes one
  conditional on starvation, converts this SPEC into the failure it was written to prevent. M3
  step 4 and AC-SIV-004 exist to catch it.
- **Sampling the latency.** Adding a `time.Since` threshold to the budget guard reintroduces the
  machine-sensitive verdict. Explicitly forbidden (REQ-SIV-009, AC-SIV-007).
- **Re-tuning a constant to make something pass.** If the M1 guard were to fail on the unmutated
  tree, the correct response is to report it as a blocker, not to raise `boardLockHeadroom` —
  that is branch A, operator-rejected.
- **Claiming improvement.** No pre-repair firing rate exists. The only sentence available is "the
  verdict criterion moved to the invariants, and under that criterion it is green" (REQ-SIV-016).
- **Closing on one green.** AC-SIV-013 requires a firing rate across at least 5 post-landing
  develop heads. The card stays open at merge.

## §H Cross-references

- `.moai/reports/t370/verdict.md`, `.moai/reports/t370/measurements.md`
- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354)
- `internal/kanban/board_store.go`, `board_lock.go`, `board_lock_wait_test.go`,
  `backlog_concurrency_test.go`

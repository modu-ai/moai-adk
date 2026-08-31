# SPEC-STRESS-INVARIANT-VERDICT-001 — Implementation Plan

Card t372. Worktree `.claude/worktrees/t372`, branch `WT-stress-invariant-guard`,
base `origin/develop` = `b9149857c`.

Milestones are ordered by **decision reversibility** — the choices most likely to change on review
come first; the mechanical edits sit at the bottom.

## §A Context

Remediation branch C of card t370's three-branch finding, selected by the operator. Branches A
(re-tune the budget constant) and B (add in-process fairness) are rejected and out of scope
(`spec.md` §E). The investigation is complete; this plan re-measures nothing.

Tier M (`spec.md` § Tier classification): threshold 0.80, 2 plan-audit iterations, artifact set
`spec.md` + `plan.md` + `acceptance.md`.

## §B Known issues carried in

- The budget's headroom coincidence (`10 * 5 = 50 ~= 48`) is currently accidental. M1 makes it
  load-bearing but does **not** enlarge it — the guard asserts coherence at the declared
  per-mutation cost, not sufficiency on any real machine (REQ-SIV-009).
- Today's numbers already satisfy the new guard: `1.65s >= 48 * 33ms = 1.584s`. So M1 lands green
  without touching any constant. That is intended, and it is also why M1's mutant (AC-SIV-008) is
  the only thing establishing the guard is not vacuous.

### What the new guard actually catches, and at what margin

The margin is **66ms on 1.584s = 4.2%**. Stated plainly so nobody later reads the guard as coverage
evidence: it is a narrow regression tripwire, not a broad guarantee.

`boardLockCIMutationCost` appears on both sides of `boardLockWaitBudget >= stressWriters *
stressAddsPerWriter * boardLockCIMutationCost` and cancels, so the relation the guard enforces
reduces to a **cost-independent** one:

```
boardLockSupportedWriters * boardLockHeadroom  >=  stressWriters * stressAddsPerWriter
                    10    *          5     =  50  >=  8 * 6 = 48
```

Its independent catch-set, given what the pre-existing
`TestBoardLockWaitBudgetDerivedFromNamedInputs` already pins (`boardLockSupportedWriters == 10`,
`boardLockCIMutationCost >= 33ms`, `boardLockHeadroom >= 2`, `boardLockWaitBudget == recomputed`):

| Change | Old guard | New guard |
|---|---|---|
| `boardLockHeadroom` lowered to 2, 3, or 4 | GREEN | **RED** |
| `stressAddsPerWriter` raised so `stressWriters * stressAddsPerWriter > 50` | GREEN | **RED** |
| `boardLockCIMutationCost` changed (either direction) | RED below 33ms only | unaffected — it cancels |
| budget replaced by a bare literal | RED | GREEN |

The two guards are therefore complements, not duplicates, and the mutant in M1 step 4 must be one
from the new guard's column — which is why AC-SIV-008 also requires recording the **old** guard's
GREEN under the same mutant.

### The vacuity shape the new guard must not reproduce

`TestBoardLockWaitBudgetDerivedFromNamedInputs` (`board_lock_wait_test.go`) computes its `floor`
from the identical expression as `recomputed` two lines above, and has already asserted
`boardLockWaitBudget == recomputed`. Its `if boardLockWaitBudget < floor` branch is therefore
unreachable for any input values — a permanently green check. That is pre-existing t354-era code and
repairing it is out of scope (`spec.md` §E), but it matters twice here:

1. The new guard's floor (`stressWriters * stressAddsPerWriter * boardLockCIMutationCost`) is what
   replaces that vacuity with a real inequality, because 48 and 50 come from independently-authored
   sources.
2. The new guard **must not** rebuild its floor from the budget's own inputs. REQ-SIV-010 states
   this as a requirement and AC-SIV-007 verifies it by reading the floor expression.

## §C Pre-flight

- Read `internal/kanban/backlog_concurrency_test.go` (the stress test, lines 15-95) and
  `internal/kanban/board_lock_wait_test.go` (the existing derivation guard, plus
  `TestBacklogLockStuckHolderSurfacesBoundedNamedError` at lines 107-119 — the seeded-holder
  pattern M2 reuses).
- Confirm `IsBoardLockHeld` is exported from `internal/kanban/board_lock.go` and reachable from
  the test package (same package `kanban` — it is).
- `git rev-parse --short HEAD` and `git branch --show-current` immediately before any commit.

## §D Constraints

- Test-only change plus at most one comment in `board_store.go`. No production control flow moves
  (REQ-SIV-016).
- No CI runs, no load generation, no local reproduction attempts. The seeded-holder sub-tests are
  not load generation: in-process `acquireBoardLockImpl`, released by `t.Cleanup`, no spawned
  process, and a small add count so the bounded wait is paid once or twice — never 48 times.
- No new files: the budget guard extends `board_lock_wait_test.go` (REQ-SIV-011).
- If a guard fails on the unmutated tree, report a blocker. Never raise a constant.

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
   `TestConcurrencyStress` consumes these; no second copy exists (REQ-SIV-011).
2. Add `TestBoardLockWaitBudgetCoversSerializedMutations` to `board_lock_wait_test.go`, beside
   `TestBoardLockWaitBudgetDerivedFromNamedInputs`. It asserts
   `boardLockWaitBudget >= time.Duration(stressWriters*stressAddsPerWriter) *
   boardLockCIMutationCost`.
3. Its messages state what the guard actually enforces — a **cost-independent** ratio between the
   lock policy's supported-lane budget and the stress test's serialized mutation count — and claim no
   sufficiency on any real machine (REQ-SIV-009). Wording to avoid: "covers", "is enough",
   "suffices", and any phrasing implying the 33ms figure conditions the verdict. Wording that is
   accurate: "the lock policy budgets 10 supported writers x 5 headroom = 50 serialized mutations;
   the stress test serializes 8 x 6 = 48. The per-mutation cost cancels on both sides, so this
   relation is cost-independent and asserts nothing about the wait any real machine needs — the CI
   `-race` cost observed by t370 was 42-105ms against a declared 33ms."
4. No clock, no sleep, no goroutine, and the floor built from the two stress constants rather than
   from the budget's own inputs (REQ-SIV-010 / AC-SIV-007).
5. Discharge **AC-SIV-008**: mutate `boardLockHeadroom` 5 -> 4, run
   `go test -v -run 'TestBoardLockWaitBudget' ./internal/kanban/`, observe the **new** guard RED and
   the **old** guard GREEN in the same `-v` output (the `=== RUN` / `--- PASS` / `--- FAIL` lines
   also evidence a non-zero selector match), revert, observe both GREEN. Record the diff and all
   three outputs verbatim.

Exit: guard green on the unmutated tree; AC-SIV-008 evidence recorded with the old guard's GREEN.

### M2 — The verdict criterion split (second-highest: this is the behavioural change under suspicion)

**Why second**: this is the edit that "looks like switching the rule off". Its shape — where
tolerance is applied and where it is not — is the reviewable decision.

1. Replace the `failures []error` accumulator's role. Classify inside the writer goroutine:
   - `IsBoardLockHeld(err)` -> increment a `starved` counter (under the existing mutex).
   - any other non-nil `err` -> record it as a hard failure.
2. Replace the `len(failures) != 0` gate at line 56 with a gate on hard failures **only**:
   `if len(hardFailures) != 0 { t.Fatalf(...) }` (REQ-SIV-002 / REQ-SIV-003).
3. Add the zero-progress floor: `if len(issued) == 0 { t.Fatalf(...) }` (REQ-SIV-007), phrased as
   "a broken lock, not tolerable contention".
4. **Make the floor and the classification reachable from a sub-test without failing the parent.**
   Extract the two predicates the sub-tests need — the per-error classification and the
   zero-progress judgment — so they can be exercised against a constructed outcome. Keep the
   extraction minimal: a small helper the stress test itself calls, not a framework. This is what
   makes AC-SIV-001 and AC-SIV-005 dischargeable at all; asserting them by running the parent test
   would require producing real starvation on an unloaded machine, which the exclusions forbid and
   which would pass vacuously in any case.
5. Add the two seeded-holder sub-tests (AC-SIV-001, AC-SIV-005), modelled on
   `TestBacklogLockStuckHolderSurfacesBoundedNamedError`:
   - acquire the lock with `acquireBoardLockImpl(store.LockPath())`, release via `t.Cleanup`;
   - attempt **one or two** adds while held — not 48, so the bounded `boardLockWaitBudget` wait is
     paid once or twice;
   - AC-SIV-001: assert each error satisfies `IsBoardLockHeld` and classifies as starved;
   - AC-SIV-005: assert the zero-progress predicate reports failure for the zero-success outcome.
   No background process. No `kill`. No load generator. No `go func` that outlives the test.
6. Keep every existing invariant assertion; M3 re-anchors them.

Exit: the test tolerates starvation and only starvation; AC-SIV-001 and AC-SIV-005 discharged with
cited commands and verbatim output.

### M3 — Re-anchor the invariants, and conserve the attempts

**Why third**: mechanically implied by M2 (once `wantTotal` is no longer the expected count), but
it carries the SPEC's strictness obligation, so it is reviewed as a unit rather than folded in.

1. Compute `issuedCount := len(issued)` after the join. Keep a **separate** `successes` counter,
   incremented once per `Add` call that returned a nil error (REQ-SIV-008). The two are different
   quantities and neither substitutes for the other: `issuedCount` is the distinct-id count the four
   invariants are anchored to, `successes` is the successful-attempt count conservation is checked
   against. `successes := len(issued)` is specifically forbidden — under it a duplicate issuance
   collapses two successes into one id and breaks conservation for a reason that is not an accounting
   fault, which is REQ-SIV-005(a)'s job to report, not this identity's.
2. Rewrite the four assertions against `issuedCount` (REQ-SIV-005 a-d):
   collision (`n != 1` per id — unchanged), presence-in-queue (unchanged),
   `len(rec.Items) != issuedCount`, `rec.LastSeq != issuedCount`.
3. Delete the `len(issued) != wantTotal` assertion — it is exactly the criterion M2 removed, and
   leaving it would silently re-fail on starvation.
4. **Replace it with the conservation assertion** (REQ-SIV-008 / AC-SIV-014):
   `successes + starved + hardFailures == stressWriters * stressAddsPerWriter` — `successes` being the
   step-1 counter, never `issuedCount` — failing with the discrepancy named. Step 3 without step 4 leaves the `> 0` floor as the only remaining tie between
   the invariant set and the work attempted, which would admit 1 success in 48 — t370 measured real
   starvation at 3-7 of 48, so that floor is roughly 40x weaker than observed behaviour. Conservation
   is machine-independent (it counts outcomes, not milliseconds) and therefore does not reintroduce
   the load-sensor verdict. **Do not** substitute a fractional or percentage success floor — that is
   a load sensor with an accounting label, and it recreates the flake on the next slower runner.
5. Confirm no assertion is behind a starvation conditional (REQ-SIV-006 / AC-SIV-004).
6. Discharge **AC-SIV-009**: plant an invariant mutant that **reaches the invariant block** —
   either a dropped item after its id was issued (lost update), or a `last_seq` advance **above**
   the item count. Observe `TestConcurrencyStress` RED where the failure originates at a named
   assertion **inside the invariant block** — one of the four REQ-SIV-005 assertions (a)-(d), named
   in the evidence — cite that assertion's message and source line, revert, observe GREEN. Record
   diff, RED, restoring GREEN. A RED at the `store.Load()` error check, a DATA RACE, or a panic does
   **not** discharge it (AC-SIV-009's non-qualifying list). The mutant lives in
   `backlog_store.go`/`backlog_sqlite.go` and is reverted before the AC-SIV-012 scope diff is taken.

   **Two shapes will not discharge it, and cost a wasted cycle if attempted** (both stated in
   `acceptance.md` AC-SIV-009 with their source evidence):
   - a **duplicate-id** mutant — rejected by `id TEXT NOT NULL UNIQUE` in `backlog_sqlite.go`
     before the invariant block is reached; the resulting RED belongs to the REQ-SIV-002
     hard-failure gate;
   - a **downward** `last_seq` mutation — erased by `normalizeBacklogRecord`
     (`backlog_store.go`), which raises `LastSeq` to the maximum present id post-mutate; the test
     stays GREEN.

   Likewise, a RED from the hard-failure gate, the zero-progress floor, the conservation assertion,
   or a storage-layer rejection does not discharge this AC.

Exit: AC-SIV-009 and AC-SIV-014 evidence recorded; invariants unconditional; attempts conserved.

### M4 — Observability

1. `t.Logf` the starved count and the back-derived per-mutation cost
   (`time.Since(start) / time.Duration(issuedCount)`), guarded against `issuedCount == 0` (which
   M2's floor already fails on).
2. Confirm no verdict depends on either figure (REQ-SIV-012 / AC-SIV-010).

### M5 — Mechanical close-out

1. Optional one-line comment in `board_store.go` above the budget block cross-referencing this
   SPEC and naming the coincidence M1 pinned. Comment only — no value changes.
2. `go vet ./internal/kanban/...`; `go test -race -count=1 ./internal/kanban/`.
3. `git diff --stat` against the base to discharge AC-SIV-012 (scope discipline), confirming no
   mutant survives in the tree and that `board_lock_unix.go`, `board_lock_windows.go`,
   `board_lock.go`, and `backlog_store.go` are untouched.
4. Commit, message naming card `t372`. Do not push; do not open a PR.

---

## §G Anti-patterns to avoid

- **Deleting the test in disguise.** Any edit that removes an invariant assertion, or makes one
  conditional on starvation, converts this SPEC into the failure it was written to prevent. M3
  step 5 and AC-SIV-004 exist to catch it.
- **Deleting the count floor with nothing behind it.** M3 step 3 without step 4 relocates the same
  failure from the invariants to the accounting: every invariant stays self-consistent against a
  1-element issued set and the test reports green.
- **Sampling the latency.** Adding a `time.Since` threshold to the budget guard reintroduces the
  machine-sensitive verdict. Explicitly forbidden (REQ-SIV-010, AC-SIV-007).
- **Rebuilding the guard's floor from the budget's own inputs.** That reproduces the pre-existing
  vacuous check documented in §B — a comparison of a value with itself, permanently green.
- **A percentage success floor.** "At least 50% of adds must succeed" is a load sensor wearing an
  accounting label. Conservation counts outcomes; it does not threshold them.
- **Re-tuning a constant to make something pass.** If the M1 guard were to fail on the unmutated
  tree, the correct response is to report it as a blocker, not to raise `boardLockHeadroom` —
  that is branch A, operator-rejected.
- **Producing starvation with load.** Spawning a background process, a spin loop, or a
  `kill`-terminated load generator to satisfy AC-SIV-001/005 is forbidden. The seeded holder
  released by `t.Cleanup` is the sanctioned construction.
- **Claiming improvement, coverage, or sufficiency.** No pre-repair firing rate exists, and the
  budget has never been shown sufficient on a CI machine. The available sentences are in
  REQ-SIV-014.
- **Closing on one green.** AC-SIV-013 requires a firing rate across at least 5 post-landing
  develop heads — and even that window cannot show the invariants still fire (REQ-SIV-014
  clause 3). The card stays open at merge.

## §H Cross-references

- `.moai/reports/t370/verdict.md`, `.moai/reports/t370/measurements.md`
- `.moai/reports/t372/plan-audit.md` — iteration-1 audit (FAIL 0.69) this revision answers
- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354)
- `internal/kanban/board_store.go`, `board_lock.go`, `board_lock_unix.go`, `board_lock_wait_test.go`,
  `backlog_concurrency_test.go`, `backlog_store.go`, `backlog_sqlite.go`

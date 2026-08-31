---
id: SPEC-STRESS-INVARIANT-VERDICT-001
title: "Separate the stress test's verdict criterion from lock acquisition: invariants decide, latency gets its own derivation guard (card t372)"
version: "0.1.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, backlog, lock, contention, ci-flake, invariant, verdict-criterion, t372"
tier: S
depends_on: [SPEC-BACKLOG-LOCK-BUDGET-001]
---

# SPEC-STRESS-INVARIANT-VERDICT-001 — Invariant-Anchored Verdict for `TestConcurrencyStress`

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-31 | 0.1.0 | Plan-phase authoring. Remediation branch C of card t370's three-branch finding, selected by the operator. |

## §A Context — what was measured, and by whom

The investigation is complete and lives at `.moai/reports/t370/verdict.md` and
`.moai/reports/t370/measurements.md` (card t370, measurement tree `origin/develop` = `1e5199b88`).
This SPEC **consumes** those measurements and re-derives none of them. Every figure below is cited,
not measured in this run.

`TestConcurrencyStress` (`internal/kanban/backlog_concurrency_test.go`) reddens the develop CI
`Race Test` job in **12 of 14** non-cancelled runs descended from the prior repair merge
`728f91006` (SPEC-BACKLOG-LOCK-BUDGET-001, card t354).

Four facts bound the problem, all measured by t370:

1. **The invariants never broke.** In all 12 red runs the failure point is one line —
   `backlog_concurrency_test.go:56`, the `len(failures) != 0` gate — reporting `N/48 adds failed
   under contention ... kanban board lock held`. No id collision, no lost update, no `last_seq`
   mismatch appears in any run. What broke is lock **acquisition**, never the property the test
   was written to guard.
2. **It is not a data race.** `grep -c "DATA RACE"` reads 0 across 4 event streams and 9 job logs.
3. **The budget has no margin, by arithmetic.**
   `boardLockWaitBudget = boardLockSupportedWriters(10) * boardLockCIMutationCost(33ms) *
   boardLockHeadroom(5) = 1.65s` is a compile-time constant. The test serializes
   `writers(8) * addsPerWriter(6) = 48` mutations through ONE flock, so a contender's worst wait
   approaches `48 * per-mutation cost`, and the budget survives only while `cost < 34.4ms`. The
   `headroom = 5` reads as five-fold slack but is not: the multiplied term is the supported **lane
   count (10)**, not the serialized **mutation count (48)**. `10 * 5 = 50 ~= 48`, so the headroom
   was entirely consumed converting one to the other. The remaining margin is effectively zero.
4. **CI `-race` sits above that threshold.** Back-derived per-mutation cost: 42, 46, 64, 105 ms —
   1.2x to 3.1x over 34.4ms. Passing environments measure 11-16ms (CI non-race) and 17.5ms (local
   darwin `-race`). All 12 failing elapsed values (1.94s .. 4.30s) exceed the 1.65s budget, with
   zero exceptions.

### The defect, stated precisely

The test uses **one verdict criterion for two unrelated properties**. It fails when the invariants
break — correct — and it fails identically when a contender exhausts a machine-speed-sensitive
latency budget — a measurement of the runner, not of the code. Because a compile-time budget cannot
track an execution machine, the second criterion turns the test into a load sensor, and the signal
that matters (the invariants) is drowned by the signal that does not.

### The chosen remediation — branch C, operator-selected

Separate the criteria. A lock-acquisition failure stops counting as a test failure; acquisition
latency moves into its own explicit, machine-independent budget guard that asserts the **derivation**
rather than a sampled wall-clock.

The hazard the operator named, and the reason this SPEC exists in this shape: *making lock failures
stop counting looks identical to switching the rule off.* A starvation-tolerant test that also
stopped checking invariants would be indistinguishable from deleting the test. The countermeasure is
structural — the invariants get **stricter**, and both criteria carry mutant-provable acceptance
criteria (§C).

## §B Requirements (GEARS)

### B.1 — Failure classification

**REQ-SIV-001** (Event-detected) — **When** `(*BacklogStore).acquireLock` exhausts
`boardLockWaitBudget` and returns an error satisfying `IsBoardLockHeld`, the stress test shall
record the add as *starved* and continue, rather than failing the test.

**REQ-SIV-002** (Event-detected, unwanted) — **When** an `Add` returns an error that does **not**
satisfy `IsBoardLockHeld`, the stress test shall fail immediately, naming the error. The test shall
not tolerate any error class other than the `IsBoardLockHeld` contention sentinel.

**REQ-SIV-003** (Ubiquitous) — The classification shall be mechanical: it is decided by
`IsBoardLockHeld(err)` (`internal/kanban/board_lock.go`), never by matching error text, error
count, or elapsed time.

### B.2 — Invariants, re-anchored and strengthened

**REQ-SIV-004** (Ubiquitous) — **Where** some adds are tolerated as starved, `wantTotal` is no
longer the expected count, so the stress test shall anchor every invariant assertion to the set of
ids **actually issued** rather than to a static constant.

**REQ-SIV-005** (Ubiquitous) — The stress test shall assert all four of the following against the
issued set, on every run:

| # | Invariant | Assertion |
|---|---|---|
| a | no id collision | every issued id appears exactly once in the issued multiset |
| b | no lost update | every issued id is present in the loaded queue |
| c | count consistency | `len(rec.Items)` equals the number of distinct issued ids |
| d | mark consistency | `rec.LastSeq` equals the number of distinct issued ids |

**REQ-SIV-006** (Unwanted) — The stress test shall not weaken, skip, or make conditional any of the
four invariant assertions in REQ-SIV-005 on the basis that adds were starved. Starvation tolerance
applies to the failure **count** only, never to the invariant **checks**.

### B.3 — Progress floor

**REQ-SIV-007** (Event-detected) — **When** the number of successful adds is zero, the stress test
shall fail. Total starvation is a broken lock, not tolerable contention.

### B.4 — The latency budget guard, machine-independent

**REQ-SIV-008** (Ubiquitous) — A dedicated budget guard shall assert that `boardLockWaitBudget`
covers the mutation count the stress test actually serializes:
`boardLockWaitBudget >= stressWriters * stressAddsPerWriter * boardLockCIMutationCost`.

**REQ-SIV-009** (Ubiquitous) — The guard shall assert this **derivation**, not a sampled wall-clock
measurement. Sampling acquisition latency and asserting a threshold on it would reintroduce exactly
the machine-sensitive verdict this SPEC removes.

**REQ-SIV-010** (Ubiquitous) — The stress test's `writers` and `addsPerWriter` shall be package-level
constants (`stressWriters`, `stressAddsPerWriter`) so the guard reads the same figures the stress
test runs, making the lane-count-versus-mutation-count relationship load-bearing rather than
coincidental.

**REQ-SIV-011** (Ubiquitous) — The guard shall live in `internal/kanban/board_lock_wait_test.go`,
beside the existing derivation guard `TestBoardLockWaitBudgetDerivedFromNamedInputs`. That file's
stated concern is already "the derivation is visible ... NOT that any budget is sufficient"; this
guard strengthens the same concern and belongs in the same file rather than in a new one.

### B.5 — Observability

**REQ-SIV-012** (Ubiquitous) — The stress test shall log, via `t.Logf`, the starved-add count and
the back-derived per-mutation cost (`test elapsed / successful mutations`), so a latency regression
stays visible in the CI stream.

**REQ-SIV-013** (Unwanted) — The logged figures shall not participate in the verdict. No `t.Error`
or `t.Fatal` shall be gated on the starved count (beyond the REQ-SIV-007 zero-progress floor) or on
the derived cost.

## §C Mutant-provability (binding)

**REQ-SIV-014** (Ubiquitous) — Both separated criteria shall be demonstrated capable of turning RED,
by planting a mutant and observing the failure. Evidence in one direction only is not sufficient:

| Direction | Mutant | Must go RED |
|---|---|---|
| latency budget | lower `boardLockHeadroom` from 5 to 4 (budget 1.65s -> 1.32s, below the 48 x 33ms = 1.584s floor) | the new budget guard |
| invariant | break a queue invariant (duplicate id / dropped item / wrong `last_seq`) | `TestConcurrencyStress` |

**REQ-SIV-015** (Ubiquitous) — Each planted mutant shall be reverted after its RED is observed, and
the post-revert GREEN shall be recorded alongside the RED. A RED without its restoring GREEN does
not establish that the guard discriminates.

## §D Non-claims — what this SPEC can never assert

**REQ-SIV-016** (Ubiquitous) — The SPEC and every report derived from it shall state the following
two limits explicitly, and shall not make a claim past either:

1. **No before/after comparison exists, in any quantity.** t370 never measured the **pre**-repair
   firing rate. The strongest statement available is: *the verdict criterion moved to the
   invariants, and under that criterion it is green.* Any phrasing implying improvement versus a
   prior rate is unsupported.
2. **A single green run can never close this card.** Two green `-race` runs already exist
   post-repair (`51daada00`, `c6aa61346`). Judgment requires a firing rate across multiple runs.

**REQ-SIV-017** (Ubiquitous) — Closure shall require a firing rate observed across **multiple**
post-landing CI `Race Test` runs on develop, not a single green. The observation window and its
count are stated in `acceptance.md` §Closure.

## §E Exclusions

This SPEC changes the **verdict criterion of a test**. It changes no lock behaviour. The following
are out of scope, and each was considered and rejected — two of them by explicit operator decision.

### Out of Scope — lock implementation changes (branch B, operator-rejected)

- Adding an in-process mutex to serialize same-process writers ahead of the flock.
- Introducing fairness, a queue, or FIFO ordering into the board lock.
- Any change to `internal/kanban/board_lock.go` acquisition semantics.
- Rationale: t370 recorded that the real Factory ten lanes are **separate processes**, so an
  in-process mutex could green this test while leaving operational contention unfixed.

### Out of Scope — budget re-tuning (branch A, operator-rejected)

- Raising `boardLockCIMutationCost` to the 42-105ms observed band.
- Raising `boardLockHeadroom`, or otherwise enlarging `boardLockWaitBudget`.
- Rationale: the result is again a compile-time constant, which the next slower runner breaks the
  same way. This SPEC's guard asserts the derivation is *coherent*, not that the budget is *large*.

### Out of Scope — new measurement

- Running CI, generating machine load, or reproducing the failure locally.
- Re-deriving any t370 figure. The measurements are cited as ground truth.
- Measuring the pre-repair firing rate (unrecoverable — see REQ-SIV-016).

### Out of Scope — neighbouring red tests

- `TestGitDiffNameCount_Predicate` (card t352) and
  `TestBinaryLag_OneSeamServesBothSurfaces` (t326/t366 series), both of which redden the same CI
  job in some runs. t370 measured the latter as red in the **non-race** job too, making it a
  separate axis. Neither is attributed to this card.
- `TestConfigChange_RT005ReloadIntegration` (card t278), which did not fail in any of the 14 runs.

### Out of Scope — production code behaviour

- Any change to `internal/kanban/backlog_store.go` control flow.
- Any change to `board_store.go` beyond an optional comment cross-referencing this SPEC.

## §F Cross-references

- `.moai/reports/t370/verdict.md`, `.moai/reports/t370/measurements.md` — the measured ground truth.
- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354) — the prior repair, whose self-declared Gap ("local
  evidence does not establish the CI failure is closed") t370 filled with "not closed".
- `internal/kanban/board_store.go` — `boardLockWaitBudget` and its named inputs.
- `internal/kanban/board_lock.go` — `IsBoardLockHeld`, the classification predicate.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the doctrine §D's non-claims implement.

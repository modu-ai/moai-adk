---
id: SPEC-STRESS-INVARIANT-VERDICT-001
title: "Separate the stress test's verdict criterion from lock acquisition: invariants decide, latency gets its own derivation guard (card t372)"
version: "0.4.0"
status: in-progress
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, backlog, lock, contention, ci-flake, invariant, verdict-criterion, t372"
tier: M
depends_on: [SPEC-BACKLOG-LOCK-BUDGET-001]
---

# SPEC-STRESS-INVARIANT-VERDICT-001 — Invariant-Anchored Verdict for `TestConcurrencyStress`

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-31 | 0.1.0 | Plan-phase authoring. Remediation branch C of card t370's three-branch finding, selected by the operator. |
| 2026-08-31 | 0.2.0 | Plan-audit fix round (audit FAIL 0.69, `.moai/reports/t372/plan-audit.md`). Tier reclassified S → M with the requirement layer consolidated 17 → 16 REQ; attempt-conservation requirement added; the budget guard's verb corrected from *covers* to *coherent with, at the declared cost*; the guard's non-tautology shape made a requirement; the Unix sentinel's width and the closure gate's discriminating power added as non-claims; one misquoted CI run corrected. |
| 2026-08-31 | 0.3.0 | Plan-audit iteration-2 fix round (audit PASS-WITH-DEBT 0.83, `.moai/reports/t372/plan-audit.md`, findings N1-N5). "The invariant block" bounded to the four REQ-SIV-005 assertions by identity; `successes` defined as a nil-error `Add` counter, distinct from `issuedCount`; REQ-SIV-009's cost-cancellation stated in the requirement and its mandated message wording corrected to a cost-independent relation; AC-SIV-012's mutant carve-out extended to AC-SIV-009; AC-SIV-014 relocated from §D.3 to §D.1. Finding N6 declined with evidence (`Event-detected` is the canonical GEARS label). |
| 2026-08-31 | 0.4.0 | Kickoff-approval amendment (card t372, three lead conditions). REQ-SIV-009 now leads with the positive characterization — a **constant-coherence guard, not a runtime budget guard**, whose firing condition is a constant-axis regression — with the cost-cancellation limitation stated after it, and the mandated message wording reordered to match (`plan.md` §B + M1 step 3). AC-SIV-008's mutant constrained to the **constant axis**, with the cost-axis mutant named as a non-qualifying shape and its reason. REQ-SIV-013 reworded to carry the **pre-plant guard-census** obligation binding both mutant ACs (census + enumerating command, RED test named, non-zero selector match count, old-guard GREEN on the latency direction). Frontmatter `version` realigned with HISTORY (drifted at 0.2.0). Counts unchanged: 16 REQ / 14 AC. |

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
structural — the invariants get **stricter**, every attempted add stays accounted for (REQ-SIV-008),
and both criteria carry mutant-provable acceptance criteria (§C).

### Tier classification and its justification

`tier: M`. The scope guidance in `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity
Tier would place this change at Tier S by LOC (a test-only change under 300 LOC across 3 files), but
the same rule caps Tier S at **8 requirements and 8 acceptance criteria** and states that exceeding
either ceiling "is a signal to tier up or to split the SPEC, not to relax the budget". This SPEC
carries 16 requirements and 14 acceptance criteria after consolidation, which fits Tier M's 16/16
ceiling and does not fit Tier S's.

The tier is read mechanically and decides two things beyond the artifact set, so the choice is made
explicitly rather than left to the LOC heuristic:

- **Threshold** — Tier M applies the plan-auditor PASS threshold `0.80`, not Tier S's `0.75`. The
  higher bar is accepted deliberately: this card's central hazard is that concealment resembles
  repair, so the most lenient threshold available is the wrong one to claim.
- **Iteration ceiling** — Tier M allows 2 plan-audit iterations (`harness.yaml`
  `plan_audit_tier_ceilings`), where Tier S allows 1. A SPEC whose binding condition is
  mutant evidence needs the revision loop, not a single pass.

Splitting instead of tiering up was rejected: the requirement layer is one indivisible argument —
tolerance, invariants, accounting, and the two mutant directions only mean anything together, and a
split would let one half land without the other, which is the failure mode §C exists to prevent.

The Tier M artifact set (`spec.md` + `plan.md` + `acceptance.md`) is exactly what this SPEC carries;
under the previously declared Tier S, `acceptance.md` was a deviation.

## §B Requirements (GEARS)

### B.1 — Failure classification

**REQ-SIV-001** (Event-detected) — **When** `(*BacklogStore).acquireLock` exhausts
`boardLockWaitBudget` and returns an error satisfying `IsBoardLockHeld`, the stress test shall
record the add as *starved* and continue, rather than failing the test.

**REQ-SIV-002** (Event-detected) — **When** an `Add` returns an error that does **not** satisfy
`IsBoardLockHeld`, the stress test shall fail immediately, naming the returned error.

**REQ-SIV-003** (Unwanted) — The stress test shall not tolerate any error class other than the
`IsBoardLockHeld` contention sentinel, and shall not decide the class by matching error text, error
count, or elapsed time. The classification is decided solely by `IsBoardLockHeld(err)`
(`internal/kanban/board_lock.go`).

### B.2 — Invariants, re-anchored and strengthened

**REQ-SIV-004** (Where) — **Where** some adds are tolerated as starved, `wantTotal` is no longer the
expected count, so the stress test shall anchor every invariant assertion to the set of ids
**actually issued** rather than to a static constant.

**REQ-SIV-005** (Ubiquitous) — The stress test shall assert all four of the following against the
issued set, on every run:

| # | Invariant | Assertion |
|---|---|---|
| a | no id collision | every issued id appears exactly once in the issued multiset |
| b | no lost update | every issued id is present in the loaded queue |
| c | count consistency | `len(rec.Items)` equals the number of distinct issued ids |
| d | mark consistency | `rec.LastSeq` equals the number of distinct issued ids |

**The invariant block** means exactly these four assertions (a)-(d), by name, and nothing else.
Membership is by identity, not by textual position: a statement that sits between two of them in
the source is still outside the block unless it is one of the four. In particular the `store.Load()`
error check, the REQ-SIV-002 hard-failure gate, the REQ-SIV-007 zero-progress floor, and the
REQ-SIV-008 conservation assertion are **outside** the invariant block. Every artifact in this SPEC
uses the term in this sense.

**REQ-SIV-006** (Unwanted) — The stress test shall not weaken, skip, or make conditional any of the
four invariant assertions in REQ-SIV-005 on the basis that adds were starved. Starvation tolerance
applies to the failure **count** only, never to the invariant **checks**.

### B.3 — Accounting floor

**REQ-SIV-007** (Event-detected) — **When** the number of successful adds is zero, the stress test
shall fail. Total starvation is a broken lock, not tolerable contention.

**REQ-SIV-008** (Ubiquitous) — The stress test shall assert that every attempted add is accounted
for in exactly one class:
`successes + starved + hardFailures == stressWriters * stressAddsPerWriter`.

`successes` is a dedicated counter incremented **once per `Add` call that returned a nil error** —
explicitly **not** `len(issued)`, and explicitly not `issuedCount`. The two quantities are distinct
and both are needed: `issuedCount` (the number of **distinct** issued ids) is what the four
REQ-SIV-005 invariants are anchored to, while `successes` (the number of **successful attempts**) is
what makes conservation mean "every attempt landed in exactly one class". Reading `successes` as
`len(issued)` would break that: a duplicate issuance would collapse two successes into one distinct
id and fail conservation for a reason that has nothing to do with accounting, while also turning the
identity into a second collision detector competing with REQ-SIV-005(a) for the RED. Keeping them
separate is what leaves `issuedCount < successes` as the collision invariant's own signal to catch.

This assertion replaces the `len(issued) != wantTotal` check the tolerance change removes, and it is
the only thing tying the invariant set back to the work the test actually attempted. It is
deliberately **machine-independent**: it counts outcomes, not milliseconds, so it does not
reintroduce the load-sensor verdict this SPEC removes. It is equally deliberately **not** a
fractional or percentage success floor — a "at least N% must succeed" threshold would be a load
sensor wearing an accounting label, and would recreate the flake on the next slower runner.

The zero-progress floor in REQ-SIV-007 alone admits a run with 1 success in 48; conservation is what
makes that admission survivable, because an accounting bug that silently drops successes (an
`err == nil` branch that fails to record) breaks the conservation identity even though every
invariant stays self-consistent against the smaller issued set. The residual weakness of the `> 0`
floor itself is stated as a non-claim in §D.

### B.4 — The latency budget guard, machine-independent

**REQ-SIV-009** (Ubiquitous) — A dedicated budget guard shall assert that `boardLockWaitBudget` is
**coherent with** the mutation count the stress test serializes, as written in terms of the declared
per-mutation cost `boardLockCIMutationCost`:
`boardLockWaitBudget >= stressWriters * stressAddsPerWriter * boardLockCIMutationCost`.

This is **not a runtime budget guard — it is a constant-coherence guard**, and that positive
characterization is what answers "what kind of guard compares 50 to 48?".
`boardLockCIMutationCost` appears on **both** sides of the inequality and cancels, so what the guard
enforces is a relation between compile-time constants only:
`boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter` (50 >= 48).
**What it catches is a constant-axis regression** — someone lowering `boardLockSupportedWriters` or
`boardLockHeadroom`, or raising `stressWriters` / `stressAddsPerWriter` past their product. That is
its firing condition, stated exhaustively, and it is a real regression path: 48 and 50 are two
independently-authored figures, so the inequality binds a real coupling between the stress test and
the lock policy, and a future edit that quietly narrows the lock policy or widens the stress test
turns it RED. Execution time, machine speed, and contention level are **not inputs to this guard**.

The limitation follows from the same cancellation, and does not displace the above: no change to
`boardLockCIMutationCost`, and no per-mutation cost regression of any size, can ever make this guard
fire. "It protects the latency budget" therefore remains the overclaim to avoid.

The guard's own failure and success messages shall accordingly lead with what the guard enforces —
a **constant-coherence** relation between the lock policy's supported-lane budget and the stress
test's serialized mutation count, cost-independent by construction — and state the limitation after
it. They shall never claim that the budget suffices on any real machine, and shall not imply that
the declared 33ms figure conditions the verdict or that a change to `boardLockCIMutationCost` would
be caught here.

The no-sufficiency half of that is not pedantic: t370 back-derived the CI `-race` per-mutation cost
at 42-105ms (`.moai/reports/t370/verdict.md`), so the wait actually required on that machine is
`48 * 42..105ms = 2.0..5.0s` against a 1.65s budget. A message claiming coverage would tell a CI
reader something no observed machine has ever satisfied.

**REQ-SIV-010** (Ubiquitous) — The guard's floor shall be computed from constants only — no
`time.Now`, no `time.Since`, no `time.Sleep`, no goroutine — and shall be derived from terms the
budget expression does not itself supply, so that the inequality is a real comparison rather than a
tautology.

Two failure shapes this forbids, both of which produce a permanently green check:

1. **Sampling.** Asserting a threshold against a measured acquisition latency reintroduces exactly
   the machine-sensitive verdict this SPEC removes.
2. **Self-comparison.** A floor rebuilt from the budget's own inputs compares a value with itself.
   The precedent is in the host file: `TestBoardLockWaitBudgetDerivedFromNamedInputs`
   (`board_lock_wait_test.go`) computes `floor` from the identical expression as `recomputed` and
   has already asserted `boardLockWaitBudget == recomputed`, so its `budget < floor` branch cannot
   be reached for any input values. That pre-existing check is out of scope to repair
   (§E), but the new guard shall not reproduce its shape.

The new guard satisfies this because `stressWriters * stressAddsPerWriter` (48) originates in the
stress test and `boardLockSupportedWriters * boardLockHeadroom` (50) originates in the lock policy —
two independently-authored figures. `boardLockCIMutationCost` appears on both sides and cancels, so
the relation the guard actually enforces is
`boardLockSupportedWriters * boardLockHeadroom >= stressWriters * stressAddsPerWriter`, independent
of the per-mutation cost. What this does and does not catch is stated in `plan.md` §B.

**REQ-SIV-011** (Ubiquitous) — The stress test's `writers` and `addsPerWriter` shall be package-level
constants (`stressWriters`, `stressAddsPerWriter`) with a single definition that
`TestConcurrencyStress` itself consumes, and the guard shall live in
`internal/kanban/board_lock_wait_test.go` beside the existing derivation guard. That file's stated
concern is already "the derivation is visible ... NOT that any budget is sufficient"; this guard
strengthens the same concern and belongs in the same file rather than in a new one. The shared
constants are what make the lane-count-versus-mutation-count relationship load-bearing rather than
coincidental.

### B.5 — Observability

**REQ-SIV-012** (Ubiquitous) — The stress test shall log, via `t.Logf`, the starved-add count and
the back-derived per-mutation cost (`test elapsed / successful mutations`), so a latency regression
stays visible in the CI stream. The logged figures shall not participate in the verdict: no
`t.Error` or `t.Fatal` shall be gated on the starved count (beyond the REQ-SIV-007 zero-progress
floor) or on the derived cost.

## §C Mutant-provability (binding)

**REQ-SIV-013** (Ubiquitous) — Both separated criteria shall be demonstrated capable of turning RED,
by planting a mutant and observing the failure. **Before any mutant is planted**, the guards
covering the mutated point shall be enumerated — the census — and the enumerating command recorded;
a pass observed without having disabled every guard in that census is not innocence, and a RED
observed while several guards cover the same point is not attribution to one of them. Each planted
mutant shall be reverted after its RED is observed, and the post-revert GREEN shall be recorded
alongside the RED — a RED without its restoring GREEN does not establish that the guard
discriminates. Evidence in one direction only is not sufficient:

| Direction | Mutant | Must go RED | Attribution requirement |
|---|---|---|---|
| latency budget | lower `boardLockHeadroom` from 5 to 4 (budget 1.65s -> 1.32s, below the 48 x 33ms = 1.584s floor) | the **new** budget guard | the evidence names which test failed, and records that `TestBoardLockWaitBudgetDerivedFromNamedInputs` stayed GREEN under the same mutant |
| invariant | break a queue invariant in a way that reaches the invariant block (see AC-SIV-009) | `TestConcurrencyStress` | the RED originates at a named assertion **inside the invariant block** — one of REQ-SIV-005 (a)-(d), named — and the evidence cites that assertion's message and source line |

The census is what makes each attribution requirement checkable rather than asserted. For every
mutant in the table above, the recorded evidence shall carry all of: (i) the census — which tests
cover the mutated point, enumerated **before** the mutant is planted, with the command used to
enumerate them; (ii) which test(s) actually went RED, **by name**; (iii) the `-run` selector's
non-zero match count, so a zero-match selector cannot masquerade as a pass; and (iv) on the latency
direction, that the pre-existing `TestBoardLockWaitBudgetDerivedFromNamedInputs` stayed GREEN under
the same mutant — which is what attributes the RED to the new guard alone.

Both attribution requirements exist because a RED alone is not discrimination. On the latency side,
the two guards share a `-run 'TestBoardLockWaitBudget'` prefix, so a RED that did not name its test
could belong to either; the old guard's GREEN under the same mutant is what attributes the RED to
the new one. On the invariant side, a RED produced by the REQ-SIV-002 hard-failure gate, by the
REQ-SIV-007 zero-progress floor, by the REQ-SIV-008 conservation assertion, or by a storage-layer
rejection proves those gates fire and proves nothing whatever about the invariants.

## §D Non-claims — what this SPEC can never assert

**REQ-SIV-014** (Ubiquitous) — The SPEC and every report derived from it shall state the following
four limits explicitly, and shall not make a claim past any of them:

1. **No before/after comparison exists, in any quantity.** t370 never measured the **pre**-repair
   firing rate. The strongest statement available is: *the verdict criterion moved to the
   invariants, and under that criterion it is green.* Any phrasing implying improvement versus a
   prior rate is unsupported.
2. **A single green run can never close this card.** Post-repair there exists **one** green
   `Race Test` job (`51daada00`) and **one** run in which `TestConcurrencyStress` was green inside a
   job reddened by a different test (`c6aa61346`, recorded by
   `.moai/reports/t370/verdict.md` as "잡은 붉지만 원인은 다른 테스트"). Calling the second a green
   run overstates the source. Judgment requires a firing rate across multiple runs.
3. **A green observation window does not evidence that the invariants still fire.** t370's fact 1
   is that the invariant criterion was red in **0 of the 14** observed runs *before* this change, so
   a post-landing window of greens under the new criterion is fully consistent with the invariants
   having been switched off. It evidences only that no *new* failure mode was introduced. The
   burden of showing the invariants still fire belongs solely to the §C mutant evidence
   (AC-SIV-009), never to the closure window.
4. **The tolerated error class is wider than "contention", on the platform CI runs.** The tolerance
   admits whatever `internal/kanban/board_lock_unix.go` maps to `ErrBoardLockHeld`, which is
   **every** `unix.Flock` failure — `ENOLCK`, `EINTR`, and `EBADF` included — not only
   `EWOULDBLOCK`. The Windows substrate is narrower by contrast
   (`board_lock_windows.go` discriminates via `os.IsExist`), so the tolerance is widest on exactly
   the platform the CI `Race Test` job runs. This SPEC does not narrow the sentinel and cannot
   distinguish those errno classes from contention. Separately, `IsBoardLockHeld` is `errors.Is`,
   which traverses the `errors.Join(mutErr, relErr)` that `Mutate` returns, so a future joined error
   carrying a lock-held branch alongside a real defect would be tolerated wholesale. Both are stated
   here as residual risks, not repaired (§E); narrowing the sentinel is production behaviour and
   belongs to a follow-up card.

   A fifth, smaller residual risk is stated for completeness: the REQ-SIV-007 zero-progress floor
   admits a run with 1 success in 48, where t370 measured real starvation at 3-7 of 48. The floor is
   deliberately not tightened into a fraction (REQ-SIV-008 states why); conservation is the
   compensating control, and the floor's weakness is a known, accepted limit.

**REQ-SIV-015** (Ubiquitous) — Closure shall require a firing rate observed across **multiple**
post-landing CI `Race Test` runs on develop, not a single green. The observation window, its count,
and the limit in REQ-SIV-014 clause 3 are stated in `acceptance.md` §D.4.

## §E Exclusions

This SPEC changes the **verdict criterion of a test**. It changes no lock behaviour.

**REQ-SIV-016** (Unwanted) — The change shall not alter production control flow, and shall not
re-tune any lock-policy constant. The changed paths shall be limited to
`internal/kanban/backlog_concurrency_test.go`, `internal/kanban/board_lock_wait_test.go`, and at
most a comment-only change in `internal/kanban/board_store.go`. If a guard fails on the unmutated
tree, the response is a blocker report, never raising a constant.

The following are out of scope, and each was considered and rejected — two of them by explicit
operator decision.

### Out of Scope — lock implementation changes (branch B, operator-rejected)

- Adding an in-process mutex to serialize same-process writers ahead of the flock.
- Introducing fairness, a queue, or FIFO ordering into the board lock.
- Any change to `internal/kanban/board_lock.go` acquisition semantics.
- Narrowing `board_lock_unix.go`'s `ErrBoardLockHeld` mapping to the contention errno alone
  (REQ-SIV-014 clause 4). This is production behaviour; it is recorded as a follow-up candidate.
- Rationale: t370 recorded that the real Factory ten lanes are **separate processes**, so an
  in-process mutex could green this test while leaving operational contention unfixed.

### Out of Scope — budget re-tuning (branch A, operator-rejected)

- Raising `boardLockCIMutationCost` to the 42-105ms observed band.
- Raising `boardLockHeadroom`, or otherwise enlarging `boardLockWaitBudget`.
- Rationale: the result is again a compile-time constant, which the next slower runner breaks the
  same way. This SPEC's guard asserts the derivation is *coherent*, not that the budget is *large*.

### Out of Scope — repairing the existing derivation guard

- Narrowing `TestBoardLockWaitBudgetDerivedFromNamedInputs`'s unreachable `budget < floor` branch
  (REQ-SIV-010). It is pre-existing t354-era code; the new guard must not reproduce its shape, but
  repairing it is a separate change with its own regression surface.

### Out of Scope — new measurement

- Running CI, generating machine load, or reproducing the failure locally. The deterministic
  seeded-holder construction in AC-SIV-001 / AC-SIV-005 is not load generation: it acquires the
  lock in-process and releases it via `t.Cleanup`, spawning no process and creating no contention
  beyond the test's own goroutines.
- Re-deriving any t370 figure. The measurements are cited as ground truth.
- Measuring the pre-repair firing rate (unrecoverable — see REQ-SIV-014).

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
- `.moai/reports/t372/plan-audit.md` — the iteration-1 plan audit this revision answers.
- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354) — the prior repair, whose self-declared Gap ("local
  evidence does not establish the CI failure is closed") t370 filled with "not closed".
- `internal/kanban/board_store.go` — `boardLockWaitBudget` and its named inputs.
- `internal/kanban/board_lock.go` — `IsBoardLockHeld`, the classification predicate.
- `internal/kanban/board_lock_unix.go`, `board_lock_windows.go` — the two substrates whose sentinel
  widths differ (REQ-SIV-014 clause 4).
- `.claude/rules/moai/core/verification-claim-integrity.md` — the doctrine §D's non-claims implement.

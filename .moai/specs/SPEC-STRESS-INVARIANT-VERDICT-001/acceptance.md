# SPEC-STRESS-INVARIANT-VERDICT-001 — Acceptance Criteria

Card t372. Verification layer for the requirements in `spec.md` §B-§D.
Every AC below is binary-testable. The two mutant-evidence ACs (**AC-SIV-008** and **AC-SIV-009**)
are the card's binding condition — they are what separates "the verdict criterion moved" from
"the rule was switched off".

## §D AC Matrix

| AC | Covers | Severity |
|---|---|---|
| AC-SIV-001 | REQ-SIV-001 | MUST |
| AC-SIV-002 | REQ-SIV-002, REQ-SIV-003 | MUST |
| AC-SIV-003 | REQ-SIV-004, REQ-SIV-005 | MUST |
| AC-SIV-004 | REQ-SIV-006 | MUST |
| AC-SIV-005 | REQ-SIV-007 | MUST |
| AC-SIV-006 | REQ-SIV-008, REQ-SIV-010, REQ-SIV-011 | MUST |
| AC-SIV-007 | REQ-SIV-009 | MUST |
| **AC-SIV-008** | **REQ-SIV-014 (latency direction), REQ-SIV-015** | **MUST — binding** |
| **AC-SIV-009** | **REQ-SIV-014 (invariant direction), REQ-SIV-015** | **MUST — binding** |
| AC-SIV-010 | REQ-SIV-012, REQ-SIV-013 | MUST |
| AC-SIV-011 | REQ-SIV-016 | MUST |
| AC-SIV-012 | scope discipline (spec.md §E) | MUST |
| AC-SIV-013 | REQ-SIV-017 | MUST — deferred to closure |

---

## §D.1 Behavioural criteria

**AC-SIV-001** — starved add is tolerated
Given a stress-test run in which one or more `Add` calls return an error satisfying
`IsBoardLockHeld`,
When the test completes,
Then the test does not fail on the starved-add count alone, and the run's verdict is decided solely
by the invariant assertions and the zero-progress floor.

**AC-SIV-002** — any other error class still fails hard
Given a stress-test run in which an `Add` returns an error that does not satisfy `IsBoardLockHeld`,
When the test observes that error,
Then the test fails immediately, and the failure message names the returned error.
Verification: reading the test source, the tolerance branch is guarded by `IsBoardLockHeld(err)` and
the else branch reaches a `t.Fatalf`/`t.Errorf`. No text matching (`strings.Contains` on the error
string) is used to classify.

**AC-SIV-003** — invariants anchored to the issued set
Given a completed stress-test run that issued `K` distinct ids where `K <= stressWriters *
stressAddsPerWriter`,
When the invariant block executes,
Then it asserts: every issued id occurs exactly once; every issued id is present in the loaded
queue; `len(rec.Items) == K`; `rec.LastSeq == K`.
And no invariant assertion compares against the static `wantTotal`.

**AC-SIV-004** — invariants are unconditional
Given any stress-test run, starved or fully successful,
When the invariant block is reached,
Then all four assertions of AC-SIV-003 execute — none is behind an `if starved == 0` guard, a
`t.Skip`, or any other conditional.
Verification: `grep` the test body for `t.Skip` returns no hit, and the four assertions sit at the
function's top level after `Load`.

**AC-SIV-005** — zero-progress floor
Given a stress-test run in which every add was starved (zero successful adds),
When the test completes,
Then the test fails, naming total starvation as a broken lock rather than tolerable contention.

**AC-SIV-006** — the budget guard exists and reads the stress figures
Given `internal/kanban/board_lock_wait_test.go`,
When the new guard runs,
Then it asserts `boardLockWaitBudget >= time.Duration(stressWriters*stressAddsPerWriter) *
boardLockCIMutationCost`,
And `stressWriters` / `stressAddsPerWriter` are package-level constants that
`TestConcurrencyStress` itself consumes (a single definition, not two copies).

**AC-SIV-007** — the guard is a derivation, not a sample
Given the new guard's source,
When it is read,
Then it contains no `time.Now()`, no `time.Since`, no `time.Sleep`, and no goroutine — it computes
from constants only.
Verification: `grep -nE 'time\.(Now|Since|Sleep)|go func' ` over the guard function returns no hit.

---

## §D.2 Mutant evidence (binding — both directions required)

**AC-SIV-008** — the latency budget guard can go RED
Given the tree with `boardLockHeadroom` mutated from `5` to `4` in
`internal/kanban/board_store.go` (budget 1.65s -> 1.32s, below the `48 * 33ms = 1.584s` floor),
When `go test -run 'TestBoardLockWaitBudget' ./internal/kanban/` runs,
Then the new budget guard FAILS, and the failure names the shortfall.
And when the mutant is reverted and the same command re-runs, the guard PASSES.
Evidence recorded: the mutant diff, the RED output verbatim, the restoring GREEN output verbatim.

**AC-SIV-009** — the invariant criterion can go RED
Given the tree with a queue invariant deliberately broken — one of: (a) an `Add` path that mints a
duplicate id, (b) a mutation that drops an item after issuing its id (lost update), (c) a
`last_seq` advance that does not match the item count,
When `go test -race -run TestConcurrencyStress ./internal/kanban/` runs,
Then `TestConcurrencyStress` FAILS, and the failure names the violated invariant — not a lock
error.
And when the mutant is reverted and the same command re-runs, the test PASSES.
Evidence recorded: the mutant diff, the RED output verbatim, the restoring GREEN output verbatim.
At minimum one of (a)/(b)/(c) is planted; planting more strengthens the evidence.

> Why both: a starvation-tolerant test whose invariants no longer fire is observationally identical
> to a deleted test. AC-SIV-008 alone would show the budget is watched but not that the invariants
> are; AC-SIV-009 alone would show the invariants fire but leave the latency criterion
> unaccounted-for. Discharging one and not the other does not satisfy REQ-SIV-014.

---

## §D.3 Reporting criteria

**AC-SIV-010** — observability without verdict participation
Given a completed stress-test run,
When its output is read,
Then a `t.Logf` line reports the starved-add count and the back-derived per-mutation cost
(`elapsed / successful mutations`),
And no `t.Error`/`t.Fatal` in the test is gated on either figure, except the AC-SIV-005
zero-progress floor.

**AC-SIV-011** — non-claims stated
Given the run-phase and sync-phase reports for this SPEC,
When their claim sections are read,
Then both limits of REQ-SIV-016 appear verbatim in substance: no before/after comparison exists in
any quantity, and a single green run cannot close the card.
And no sentence in any report asserts an improvement versus the pre-repair firing rate.

**AC-SIV-012** — scope discipline
Given the run-phase diff,
When `git diff --stat` is read against the base,
Then the changed paths are limited to `internal/kanban/backlog_concurrency_test.go`,
`internal/kanban/board_lock_wait_test.go`, and at most a comment-only change in
`internal/kanban/board_store.go`.
And `boardLockCIMutationCost`, `boardLockHeadroom`, `boardLockSupportedWriters`,
`boardLockWaitMin/Max/Step`, and `boardLockRetryWait` are unchanged in value and behaviour
(a temporary mutant under AC-SIV-008 is reverted before the diff is taken).

---

## §D.4 Closure gate

**AC-SIV-013** — firing rate, not a single green
Given the change landed on develop,
When the CI `Race Test` job has run on **at least 5** non-cancelled develop heads descended from
the landing commit,
Then `TestConcurrencyStress` is green in every one of them under the invariant criterion,
And any run in which it is red is examined and attributed — a red caused by an invariant violation
reopens this SPEC; a red caused by any other test is attributed elsewhere and does not count.
Until that window closes, the card's status is `implemented`, never `completed`.

> Rationale, from t370: two green `-race` runs already exist post-repair. One green proves nothing
> about a 12-in-14 flake. This is the exact place card t354 stopped.

---

## §D.5 Definition of Done

- [ ] AC-SIV-001 .. AC-SIV-007 pass, each with cited command + verbatim output.
- [ ] AC-SIV-008 and AC-SIV-009 both discharged with mutant diff + RED + restoring GREEN.
- [ ] AC-SIV-010 .. AC-SIV-012 pass.
- [ ] `go test -race -count=1 ./internal/kanban/` green locally, with the run's own starved-add
      count and derived cost logged in the evidence.
- [ ] `go vet ./internal/kanban/...` clean.
- [ ] Reports carry the §D.3 non-claims.
- [ ] AC-SIV-013 remains open at merge; the card is not closed on it.

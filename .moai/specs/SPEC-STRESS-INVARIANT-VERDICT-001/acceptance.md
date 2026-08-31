# SPEC-STRESS-INVARIANT-VERDICT-001 — Acceptance Criteria

Card t372. Verification layer for the requirements in `spec.md` §B-§E.
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
| AC-SIV-014 | REQ-SIV-008 | MUST |
| AC-SIV-006 | REQ-SIV-009, REQ-SIV-011 | MUST |
| AC-SIV-007 | REQ-SIV-010 | MUST |
| **AC-SIV-008** | **REQ-SIV-013 (latency direction)** | **MUST — binding** |
| **AC-SIV-009** | **REQ-SIV-013 (invariant direction)** | **MUST — binding** |
| AC-SIV-010 | REQ-SIV-012 | MUST |
| AC-SIV-011 | REQ-SIV-014 | MUST |
| AC-SIV-012 | REQ-SIV-016 | MUST |
| AC-SIV-013 | REQ-SIV-015 | MUST — deferred to closure |

14 acceptance criteria against 16 requirements — both inside the Tier M ceiling of 16/16
(`spec.md` § Tier classification).

---

## §D.1 Behavioural criteria

**AC-SIV-001** — a starved add is tolerated, produced deterministically
Given a sub-test that seeds a lock holder in-process with `acquireBoardLockImpl(store.LockPath())`,
released by a `t.Cleanup`-registered function — the pattern already in the tree at
`TestBacklogLockStuckHolderSurfacesBoundedNamedError` (`internal/kanban/board_lock_wait_test.go`),
which spawns no background process, generates no machine load, and needs no CI,
When a small number of adds (one or two — not the full 48-add fan-out, so the bounded
`boardLockWaitBudget` wait is paid once or twice rather than 48 times) is attempted while the holder
is held,
Then every resulting error satisfies `IsBoardLockHeld`, each is classified as *starved* rather than
as a hard failure, and no starved add on its own drives the verdict.
Verification: `go test -run '<seeded-holder sub-test name>' ./internal/kanban/` — cited command plus
verbatim output. The sub-test asserts the classification outcome directly; it does not require the
parent stress test to fail.

**AC-SIV-002** — any other error class still fails hard
Given a stress-test run in which an `Add` returns an error that does not satisfy `IsBoardLockHeld`,
When the test observes that error,
Then the test fails immediately, and the failure message names the returned error.
Verification: reading the test source, the tolerance branch is guarded by `IsBoardLockHeld(err)` and
the else branch reaches a `t.Fatalf`/`t.Errorf`. No text matching (`strings.Contains` on the error
string) is used to classify — `grep -n 'strings.Contains' internal/kanban/backlog_concurrency_test.go`
returns no hit inside the classification path.

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

**AC-SIV-005** — zero-progress floor, produced deterministically
Given the same seeded-holder construction as AC-SIV-001, with the holder held for the whole duration
of the attempted adds so that **every** add is starved and zero succeed,
When the zero-progress predicate is evaluated against that outcome,
Then it reports failure, naming total starvation as a broken lock rather than tolerable contention.
Verification: `go test -run '<zero-progress sub-test name>' ./internal/kanban/` — cited command plus
verbatim output. The floor's predicate is exercised directly against a zero-success outcome; the
sub-test asserts that the predicate reports failure, so the parent test need not itself be made to
fail. No load generation, no background process, no CI run (`spec.md` §E).

**AC-SIV-006** — the budget guard exists, reads the stress figures, and claims only coherence
Given `internal/kanban/board_lock_wait_test.go`,
When the new guard runs,
Then it asserts `boardLockWaitBudget >= time.Duration(stressWriters*stressAddsPerWriter) *
boardLockCIMutationCost`,
And `stressWriters` / `stressAddsPerWriter` are package-level constants that
`TestConcurrencyStress` itself consumes (a single definition, not two copies),
And the guard's failure and success messages state that the relation is coherence at the declared
`boardLockCIMutationCost`, and contain no claim that the budget suffices on any real machine
(REQ-SIV-009).
Verification: reading the guard's source and its message strings.

**AC-SIV-007** — the guard is a derivation, and not a tautology
Given the new guard's source,
When it is read,
Then it contains no `time.Now()`, no `time.Since`, no `time.Sleep`, and no goroutine — it computes
from constants only,
And its floor is built from `stressWriters * stressAddsPerWriter`, terms the budget expression
(`boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom`) does not itself supply,
so the comparison is not a value against itself (REQ-SIV-010).
Verification: `grep -nE 'time\.(Now|Since|Sleep)|go func'` over the guard function returns no hit;
and the floor expression is read and confirmed to reference the two stress constants.

---

## §D.2 Mutant evidence (binding — both directions required)

**AC-SIV-008** — the latency budget guard can go RED, and the RED is attributable to it
Given the tree with `boardLockHeadroom` mutated from `5` to `4` in
`internal/kanban/board_store.go` (budget 1.65s -> 1.32s, below the `48 * 33ms = 1.584s` floor),
When `go test -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` runs,
Then the **new** budget guard FAILS, and the failure names the shortfall,
And the pre-existing `TestBoardLockWaitBudgetDerivedFromNamedInputs` remains **GREEN** under the
same mutant — recorded verbatim alongside the RED,
And the recorded output shows a **non-zero match count** for the `-run` selector (the `-v` output
names both matched tests by `=== RUN` / `--- PASS` / `--- FAIL` lines), so a zero-match selector
cannot masquerade as a pass,
And when the mutant is reverted and the same command re-runs, both guards PASS.
Evidence recorded: the mutant diff, the RED output verbatim naming **which** test failed, the old
guard's GREEN verbatim from the same run, and the restoring GREEN output verbatim.

> Why the old guard's GREEN is required: the `-run 'TestBoardLockWaitBudget'` selector matches both
> guards by prefix, so a bare RED does not say which one discriminated. Under this mutant the old
> guard stays green by construction — `boardLockWaitBudget` is a derived const so `budget ==
> recomputed` still holds, its `floor` is the identical expression so `budget < floor` is still
> false, and `headroom 4 >= 2` still passes. That GREEN is what attributes the RED to the new guard
> alone, and is therefore the discriminating half of this AC's own observation.

**AC-SIV-009** — the invariant criterion can go RED, from inside the invariant block
Given the tree with a queue invariant deliberately broken by one of:
- **(b)** a mutation that drops an item after issuing its id (lost update), or
- **(c)** a `last_seq` advance **above** the item count,

When `go test -race -run TestConcurrencyStress ./internal/kanban/` runs,
Then `TestConcurrencyStress` FAILS,
And the RED **originates at a named assertion inside the invariant block** — the recorded output
cites that assertion's message text and its source line,
And when the mutant is reverted and the same command re-runs, the test PASSES.
Evidence recorded: the mutant diff, the RED output verbatim with the failing assertion's message and
source line, the restoring GREEN output verbatim.
At minimum one of (b)/(c) is planted; planting both strengthens the evidence.

**A RED produced by any of the following does NOT discharge this AC** — each proves a different gate
fires and proves nothing about the invariants:

- the REQ-SIV-002 hard-failure gate (a non-sentinel error);
- the REQ-SIV-007 zero-progress floor;
- the REQ-SIV-008 conservation assertion;
- a storage-layer rejection.

**Two mutant shapes are excluded, with their reasons**, so an implementer does not spend a cycle on
a branch that cannot discriminate:

- **A duplicate-id mutant is excluded.** `internal/kanban/backlog_sqlite.go` declares
  `id TEXT NOT NULL UNIQUE`, so a duplicating mutation aborts inside `Mutate` and `Add` returns an
  `IsBacklogIDConflict` error (behaviour already pinned by `TestDuplicateIDRejectedByStorage`,
  `backlog_concurrency_test.go`). That error does not satisfy `IsBoardLockHeld`, so it trips the
  REQ-SIV-002 hard-failure gate and `t.Fatalf`s **before** the invariant block is reached. The test
  goes red and the message may even mention an invariant by wording, while establishing only that
  the hard-failure gate works.
- **A downward `last_seq` mutation is excluded.** `normalizeBacklogRecord`
  (`internal/kanban/backlog_store.go`) raises `rec.LastSeq` to the maximum present id and runs
  post-mutate inside `Mutate`, so a lowered `last_seq` is silently repaired and
  `TestConcurrencyStress` stays GREEN. Only an **upward** mutation reaches the assertion — hence the
  direction stated explicitly in (c).

> Why both directions: a starvation-tolerant test whose invariants no longer fire is
> observationally identical to a deleted test. AC-SIV-008 alone would show the budget is watched but
> not that the invariants are; AC-SIV-009 alone would show the invariants fire but leave the latency
> criterion unaccounted-for. Discharging one and not the other does not satisfy REQ-SIV-013.

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
Then all four limits of REQ-SIV-014 appear verbatim in substance:
1. no before/after comparison exists in any quantity;
2. a single green run cannot close the card, and the two post-repair runs are described accurately
   — one green `Race Test` job (`51daada00`) and one run in which `TestConcurrencyStress` was green
   inside a job reddened by a different test (`c6aa61346`);
3. a green observation window evidences only that no new failure mode was introduced, because the
   invariant criterion was already red in 0 of 14 pre-change runs;
4. the tolerated class is every `unix.Flock` failure on Unix, not only `EWOULDBLOCK`, and
   `errors.Is` traverses `errors.Join`.

And no sentence in any report asserts an improvement versus the pre-repair firing rate, and no
sentence claims the budget is sufficient on any machine.

**AC-SIV-012** — scope discipline
Given the run-phase diff,
When `git diff --stat` is read against the base,
Then the changed paths are limited to `internal/kanban/backlog_concurrency_test.go`,
`internal/kanban/board_lock_wait_test.go`, and at most a comment-only change in
`internal/kanban/board_store.go` (REQ-SIV-016).
And `boardLockCIMutationCost`, `boardLockHeadroom`, `boardLockSupportedWriters`,
`boardLockWaitMin/Max/Step`, and `boardLockRetryWait` are unchanged in value and behaviour
(a temporary mutant under AC-SIV-008 is reverted before the diff is taken).
And no change is made to `board_lock_unix.go`, `board_lock_windows.go`, `board_lock.go`, or
`backlog_store.go`.

**AC-SIV-014** — attempt conservation
Given a completed stress-test run,
When the accounting is checked,
Then the test asserts `successes + starved + hardFailures == stressWriters * stressAddsPerWriter`,
and fails naming the discrepancy when the identity does not hold (REQ-SIV-008).
And the assertion references no wall-clock value, no elapsed duration, and no fractional or
percentage success threshold — it counts outcomes only.
Verification: reading the assertion's source, plus the `go test -race -count=1
./internal/kanban/` run in §D.5 exercising it on a real run.

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

> Rationale and its stated limit, from t370: two post-repair runs already exist (one green job, one
> green test inside a job reddened elsewhere), and one green proves nothing about a 12-in-14 flake.
> This is the exact place card t354 stopped.
>
> **What this gate cannot discriminate** (REQ-SIV-014 clause 3): the invariant criterion was red in
> **0 of the 14** observed runs *before* this change, so a window of greens under the new criterion
> is fully consistent with the invariants having been switched off. It evidences only that no *new*
> failure mode was introduced. It does not evidence that the invariants still fire — that is
> AC-SIV-009's sole burden, and no number of green heads substitutes for it.

---

## §D.5 Definition of Done

- [ ] AC-SIV-001 .. AC-SIV-007 and AC-SIV-014 pass, each with cited command + verbatim output.
- [ ] AC-SIV-008 discharged with mutant diff + RED naming the failing test + the old guard's GREEN
      from the same run + restoring GREEN.
- [ ] AC-SIV-009 discharged with mutant diff + RED citing the invariant assertion's message and
      source line + restoring GREEN.
- [ ] AC-SIV-010 .. AC-SIV-012 pass.
- [ ] `go test -race -count=1 ./internal/kanban/` green locally, with the run's own starved-add
      count and derived cost logged in the evidence.
- [ ] `go vet ./internal/kanban/...` clean.
- [ ] Reports carry all four §D.3 non-claims.
- [ ] AC-SIV-013 remains open at merge; the card is not closed on it.

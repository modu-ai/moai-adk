---
id: SPEC-VACUOUS-FLOOR-GUARD-001
title: "Remove the unreachable self-comparison floor branch in the queue lock-wait derivation guard (card t378)"
version: "0.2.0"
status: in-progress
created: 2026-08-31
updated: 2026-08-31
author: manager-spec
priority: P2
phase: "v3.1.5 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, lock, vacuous-guard, mutation-evidence, t241-class, t378"
tier: S
---

# SPEC-VACUOUS-FLOOR-GUARD-001 — Vacuous Floor Branch in the Lock-Wait Derivation Guard

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-08-31 | 0.1.0 | Initial plan-phase draft (card t378). Repair direction resolved to **deletion**, argued in §A.4 from REQ-BLB-002 coverage analysis rather than assumed. Tier S: 7 REQ / 8 AC, both inside the Tier S ceiling of 8/8. |
| 2026-08-31 | 0.2.0 | Plan-audit FAIL fix round. AC-VFG-007's mutant switched M1 -> M2 (under M1 the budget stays above the composed 660ms floor, so the reinstated branch's silence evidenced nothing); §A.4's exhaustiveness claim narrowed after the retry-policy constants were found in the same `const` block, with a matching Out-of-Scope entry; REQ-VFG-007 given AC coverage inside AC-VFG-008; AC-VFG-006 given its pre-plant prediction; the REQ-VFG-001 / REQ-VFG-004 comment-fate conflict reconciled; `related_specs` dropped (not in the canonical frontmatter schema). Counts unchanged at 7 REQ / 8 AC. |

## §A Context and Problem

### A.1 The defect, at this tree

`internal/kanban/board_lock_wait_test.go`, inside
`TestBoardLockWaitBudgetDerivedFromNamedInputs` (measured at base `3f03d9c36`):

```
26  recomputed := time.Duration(boardLockSupportedWriters) * boardLockCIMutationCost * boardLockHeadroom
28  if boardLockWaitBudget != recomputed { t.Fatalf(...) }      <- equality asserted here
37  floor      := time.Duration(boardLockSupportedWriters) * boardLockCIMutationCost * boardLockHeadroom
39  if boardLockWaitBudget < floor { t.Errorf("budget %v < headroom floor %v", ...) }
```

`floor` is computed from the **identical expression** as `recomputed`, and `budget == recomputed`
is asserted twelve lines above with `t.Fatalf` — a hard stop, so line 39 is reached only on the
branch where the equality already held. `budget < floor` is therefore unreachable **for every
possible assignment of the three input constants**, and the `t.Errorf` at line 40 can never
execute. The branch is a permanently green check: an identity wearing an inequality's clothing.

This is t354-era code (`SPEC-BACKLOG-LOCK-BUDGET-001`). Card t372
(`SPEC-STRESS-INVARIANT-VERDICT-001`) identified it — that SPEC's REQ-SIV-010 names it verbatim as
the self-comparison precedent — and deliberately left it unrepaired: its §E carries
`### Out of Scope — repairing the existing derivation guard`. This SPEC picks up exactly that
deferred item and contradicts neither predecessor.

### A.2 The dynamic half of the evidence

The static argument above is sufficient on its own, but the observed half is stronger and came
from t372's mutant work (`.moai/reports/t372/mutant-headroom4-orchestrator.md`, landed on develop):

- A `boardLockHeadroom` 5 -> 4 mutant genuinely thinned the budget: 1.65s -> 1.32s.
- Under that mutant `TestBoardLockWaitBudgetDerivedFromNamedInputs` stayed **PASS**, while
  t372's sibling guard `TestBoardLockWaitBudgetCoversSerializedMutations` went **FAIL** — the only
  failure among 389 `=== RUN` lines in a whole-package run.
- So the budget actually became insufficient and the branch that exists to notice a thin budget
  stayed silent. That is a defect observed, not a defect argued.

Note what that mutant does NOT establish: it does not show the branch is the only thing that would
have been silent. It shows the branch was silent while a real regression passed under it.

### A.3 What REQ-BLB-002 intended

The dead branch's comment cites REQ-BLB-002 (`SPEC-BACKLOG-LOCK-BUDGET-001` §B), which reads:

> **Where** the product supports up to ten concurrent lane writers against one queue [...], the
> derived budget shall account for that contender count and shall exceed the per-mutation cost
> observed on a CI-class machine by at least the headroom factor REQ-BLB-001 states.

Two clauses, both already verified in the same test function, by assertions the branch does not
participate in:

| REQ-BLB-002 clause | Where it is actually enforced today | Line |
|---|---|---|
| accounts for the ten-lane contender count | `boardLockSupportedWriters != 10` | 45 |
| exceeds the CI-class per-mutation cost | `boardLockCIMutationCost < 33*time.Millisecond` | 54 |
| by at least the stated headroom factor | `boardLockHeadroom < 2` | 59 |
| the budget IS that product (no bare literal) | `boardLockWaitBudget != recomputed`, `t.Fatalf` | 28 |

Those four assertions compose into a real, non-vacuous floor:
`budget == 10 * cost * headroom`, `cost >= 33ms`, `headroom >= 2` together imply
`budget >= 660ms` — and each conjunct is falsifiable by a change to exactly one constant.
The floor REQ-BLB-002 asks for is enforced **input-wise**, not by any comparison against `budget`.

### A.4 Repair direction — deletion, with the argument

The decisive constraint on any repair is stated in card t378 and in t372's REQ-SIV-010: an
inequality whose two sides are built from the same terms is not a guard. t372's own guard escaped
that by deriving its floor from a term the budget expression does not contain — the serialized
mutation count `stressWriters * stressAddsPerWriter` (48) against the policy product
`boardLockSupportedWriters * boardLockHeadroom` (50), two independently-authored figures.

Applying that test to the two candidate repairs:

- **Repair the inequality.** Any floor built from `boardLockSupportedWriters`,
  `boardLockCIMutationCost`, or `boardLockHeadroom` reproduces the vacuity, because the equality at
  line 28 pins `budget` to exactly the product of those three. Terms *outside* the budget
  expression do exist, and they are two families rather than one. The **stress constants**
  (`stressWriters * stressAddsPerWriter`) are the first, and that relation is precisely what t372's
  guard already pins, in the same file, sixty lines below — restating it here would duplicate a
  landed guard and add a second thing to keep coherent. The **retry-policy constants** are the
  second: `boardLockWaitMin` (5ms), `boardLockWaitMax` (50ms), and `boardLockWaitStep` (10ms) are
  declared in the same `const` block as the three budget inputs (`internal/kanban/board_store.go`
  lines 131-136 at base `3f03d9c36`), are independently authored, and are neither budget inputs nor
  stress constants. A floor of the shape `budget >= K * boardLockWaitMax` would therefore be a real
  comparison over terms the budget expression does not supply — not a reproduction of the vacuity.
  Whether such a floor is worth having is a separate question: it needs a defensible `K`, which this
  card has neither established nor measured. It is recorded as out of scope (§D), not argued away.
- **Delete the branch.** REQ-BLB-002's intent is fully discharged elsewhere (§A.3), so deletion
  removes a check that verifies nothing and loses no coverage. This is the honest repair: the
  branch is dead weight, not an under-specified guard.

**Chosen: deletion.** The branch is removed; the comment that carried REQ-BLB-002's intent is
rewritten to record *where* the floor is actually enforced and why no floor-versus-budget
comparison exists here, so a later reader does not reinstate it. The choice rests on REQ-BLB-002's
intent already being discharged (§A.3), NOT on an outside-term floor being unconstructible — the
retry-policy family above shows one could be constructed. Deleting a check that verifies nothing is
separable from deciding whether a different check is worth adding, and this card does only the
first.

### A.5 This SPEC is itself in the t241 class

Card t241's finding is that a guard which catches vacuous guards is worthless if it is itself
vacuous. This SPEC is squarely inside that class, and the hazard survives the choice of deletion:
a deletion that quietly removes coverage, or that leaves behind assertions nobody has shown can
fail, is the same failure in a different shape. It is discharged not by saying so but by
REQ-VFG-005 — every retained assertion is shown capable of RED by a planted-and-reverted mutant —
and by AC-VFG-007's negative evidence, which plants the headroom mutant (`boardLockHeadroom`
5 -> 1) and observes the reinstated branch stay silent while the budget falls to 330ms, below the
660ms floor the four retained assertions compose (§A.3). A floor that did not move with the
mutation would have fired there; this one tracks the mutation down and cannot. The mutant choice is
load-bearing: under M1 (cost 33ms -> 20ms) the budget is 1000ms and under M3 (writers 10 -> 8) it is
1320ms, both **above** 660ms, so a genuine floor would have been silent under either and the
observation would have established nothing.

That choice was got wrong once, in this SPEC's own 0.1.0 draft, which planted M1 — and the error is
recorded here as a methodological note, because this card's subject is precisely a check that read
as coverage and was not. Confirming that the dead branch evaluates false under every input
assignment establishes that it is **vacuous**. It does not establish which mutant makes that
vacuity **visible**. Those are two different questions, and the draft conflated them: under M1 a
genuine floor would have stayed silent too, so the observation would have established nothing about
the reinstated branch. Only M2 drives the budget below the floor the retained assertions compose,
which is what makes the branch's continued silence a contrast rather than a coincidence.

The unreachability claim itself rests on the static argument in §A.1, which holds for every
assignment of the three constants. AC-VFG-007 corroborates it on one assignment; it is not offered
as proof on its own, and a single observation of silence never establishes that a branch could not
be made RED.

## §B Requirements (GEARS)

**REQ-VFG-001** (Ubiquitous — removal).
`TestBoardLockWaitBudgetDerivedFromNamedInputs` shall carry no assertion whose two comparison
operands are constructed from the same expression as an equality the same function has already
asserted. Concretely, in `internal/kanban/board_lock_wait_test.go` the `floor :=` declaration and
the `if boardLockWaitBudget < floor` block that follows it shall be removed. The removal is scoped
to that declaration and that block: the comment above them is NOT deleted but rewritten in place per
REQ-VFG-004, so the function keeps a record of where REQ-BLB-002's floor is enforced.

**REQ-VFG-002** (Ubiquitous — coverage preservation).
The guard shall retain, unweakened, the four assertions that jointly enforce REQ-BLB-002's floor:
the budget-derivation equality, the ten-lane contender-count assertion, the CI-class per-mutation
cost floor, and the headroom-factor floor.

**REQ-VFG-003** (Unwanted — no reproduction of the shape).
The repair shall not introduce any new inequality whose floor is derived from terms the budget
expression itself supplies (`boardLockSupportedWriters`, `boardLockCIMutationCost`,
`boardLockHeadroom`), and shall not restate the cross-file constant-coherence relation that
`TestBoardLockWaitBudgetCoversSerializedMutations` already pins.

**REQ-VFG-004** (Ubiquitous — the record).
The guard's comment shall state where REQ-BLB-002's floor obligation is enforced (input-wise, by
the four retained assertions), and why no floor-versus-budget comparison appears in this function,
so that reinstating one is recognisably a regression rather than an improvement.

**REQ-VFG-005** (Ubiquitous — mutant provability, both directions).
Each retained assertion shall be demonstrated capable of turning RED by a mutant that is planted,
observed, and reverted. Before any mutant is planted the guards covering the mutated constant
shall be enumerated — the census — and the enumerating command recorded. Attribution shall be
established by the **verbatim failure message of the named assertion**, never by the count of
failing tests: a mutant that lowers a shared constant may legitimately redden several guards, and
the census is what keeps that from being read as attribution to one of them. Every mutant shall be
reverted, and the post-revert GREEN recorded alongside its RED.

**REQ-VFG-006** (Unwanted — scope).
The change shall not modify `TestBoardLockWaitBudgetCoversSerializedMutations`, shall not alter the
landed values of `boardLockSupportedWriters`, `boardLockCIMutationCost`, or `boardLockHeadroom`,
and shall not alter production behaviour in `internal/kanban/board_store.go` or the lock
implementation files. A mutant planted and reverted under REQ-VFG-005 is a transient measurement
instrument, not a constant retune; the committed diff shall show no constant changed.

**REQ-VFG-007** (When — verification load).
**When** any verification for this SPEC is run, it shall be a single serial `go test` invocation
scoped to `./internal/kanban/`, without `-race`, and shall spawn no background process and no
generated load. Any process a test needs shall be bounded by `t.Cleanup` or by an external
`timeout` wrapper; a trailing `kill` is not cleanup. `go test ./...` shall not be run locally; CI
owns the full-suite verdict.

## §C Acceptance Criteria (inline — Tier S)

8 criteria against 7 requirements — both inside the Tier S ceiling of 8/8.

| AC | REQ | Weight |
|---|---|---|
| AC-VFG-001 | REQ-VFG-001, REQ-VFG-003 | MUST |
| AC-VFG-002 | REQ-VFG-002 | MUST |
| AC-VFG-003 | REQ-VFG-005 | MUST — mutant evidence |
| AC-VFG-004 | REQ-VFG-005 | MUST — mutant evidence |
| AC-VFG-005 | REQ-VFG-005 | MUST — mutant evidence |
| AC-VFG-006 | REQ-VFG-005 | MUST — mutant evidence |
| AC-VFG-007 | REQ-VFG-005, REQ-VFG-001 | MUST — deletion's negative evidence |
| AC-VFG-008 | REQ-VFG-004, REQ-VFG-006, REQ-VFG-007 | MUST |

Every requirement carries at least one criterion; REQ-VFG-007 (verification load) is verified inside
AC-VFG-008 rather than by a ninth criterion, because the artifact is at the Tier S 8/8 ceiling.

**AC-VFG-001** — the dead branch is gone and nothing of its shape replaced it
Given the repaired `internal/kanban/board_lock_wait_test.go`,
When `grep -n 'boardLockWaitBudget <' internal/kanban/board_lock_wait_test.go` is run,
Then exactly one match remains and it is inside
`TestBoardLockWaitBudgetCoversSerializedMutations` (t372's guard, untouched),
And `grep -c 'floor :=' internal/kanban/board_lock_wait_test.go` returns `1`, that one occurrence
being t372's stress-constant floor,
And the package compiles: `go vet ./internal/kanban/` exits 0.
Evidence: both greps with their verbatim output and a stated pre-repair baseline (2 matches each),
plus the `go vet` exit code.

**AC-VFG-002** — the four load-bearing assertions survive, GREEN on the unmutated tree
Given the repaired file with no mutant planted,
When `go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`
is run,
Then it passes, and the selector's match count is verified non-zero from the `=== RUN` lines (a
zero-match selector also prints `ok`),
And the function still contains the equality on `boardLockWaitBudget`, the `!= 10` contender-count
assertion, the `< 33*time.Millisecond` cost floor, and the `< 2` headroom floor.
Evidence: the run's verbatim output including the `=== RUN` / `--- PASS` lines, and the four
assertion lines quoted with their line numbers at the post-repair SHA.

**AC-VFG-003** — mutant M1: the per-mutation cost floor can fire
Given the pre-plant census for `boardLockCIMutationCost` — enumerated by
`grep -rn 'boardLockCIMutationCost' internal/kanban/`, with the enumerating command and its full
output recorded — and the census reasoned about explicitly (t372's guard is cost-independent by
construction: the cost cancels on both sides of its inequality, so it is expected to stay GREEN,
which makes this mutant single-attribution),
When `boardLockCIMutationCost` is temporarily set to `20 * time.Millisecond` and the WHOLE package
is run once, serially: `go test -timeout 600s -count=1 ./internal/kanban/`,
Then the run is RED, the failure message
`per-mutation cost 20ms is below the CI-class observation of 33ms` appears verbatim, and every
other RED in the run (if any) is named and attributed to its own guard,
And after reverting the mutant the same command is GREEN.
Evidence: census command + output, mutant diff, RED output, revert diff, post-revert GREEN output.

**AC-VFG-004** — mutant M2: the headroom floor can fire
Given the pre-plant census for `boardLockHeadroom`
(`grep -rn 'boardLockHeadroom' internal/kanban/`), recorded with its command, and the census's
prediction stated in advance — t372's guard is expected to redden too (policy product
`10 * 1 = 10 < 48`), and `TestConcurrencyStress` may redden because the budget falls to 330ms,
When `boardLockHeadroom` is temporarily set to `1` and the whole package is run once as above,
Then the run is RED and the message `headroom factor 1 states no headroom` appears verbatim —
attribution resting on that message, not on the number of failing tests — and every other RED is
named and attributed to its own guard,
And after reverting, the same command is GREEN.
Evidence: as AC-VFG-003.

**AC-VFG-005** — mutant M3: the contender-count assertion can fire
Given the pre-plant census for `boardLockSupportedWriters`, recorded with its command, and the
prediction stated in advance that t372's guard reddens too (`8 * 5 = 40 < 48`),
When `boardLockSupportedWriters` is temporarily set to `8` and the whole package is run once as
above,
Then the run is RED and the message
`supported writers = 8, want 10 (Factory mode's ten lanes against one queue)` appears verbatim,
with every other RED named and attributed,
And after reverting, the same command is GREEN.
Evidence: as AC-VFG-003.

**AC-VFG-006** — mutant M4: the derivation equality can fire
Given the pre-plant census for `boardLockWaitBudget`, recorded with its command, and the prediction
stated in advance that the `1400 * time.Millisecond` form reddens **two** guards rather than one —
the derivation equality here, and t372's `TestBoardLockWaitBudgetCoversSerializedMutations`, because
`1400ms < 48 x 33ms = 1584ms` — while the numerically-identical `1650 * time.Millisecond` form
reddens neither; attribution therefore rests on the verbatim failure message of the named assertion
and never on the count of failing tests (REQ-VFG-005),
When the `boardLockWaitBudget` declaration is temporarily replaced by the bare literal
`1650 * time.Millisecond` — numerically identical to the landed value, so only the *derivability*
changes — and then by `1400 * time.Millisecond`, and the whole package is run once for the second
form,
Then the second form is RED with the `is not the product of its named inputs` message appearing
verbatim,
And the first form is recorded as the honest limit of this mutant: a numerically identical literal
does NOT redden the guard, because the assertion compares values rather than syntax — stated as a
known gap, not hidden,
And after reverting, the whole-package run is GREEN.
Evidence: as AC-VFG-003, plus the explicit gap statement.

**AC-VFG-007** — the deletion's negative evidence: the removed branch stays silent where a real floor fires
Given the deleted branch temporarily reinstated verbatim beside the retained assertions, in a
working copy that is discarded and never committed,
And mutant **M2** (`boardLockHeadroom` 5 -> 1, per AC-VFG-004) chosen for this observation because it
is the only one of the four that drives the budget BELOW the 660ms floor the four retained
assertions compose (`10 x 33ms x 1 = 330ms`); M1 (cost 33ms -> 20ms) leaves it at 1000ms and M3
(writers 10 -> 8) at 1320ms, both above 660ms, so under either of those a genuine floor would have
stayed silent too and the observation would establish nothing,
When M2 is planted alongside the reinstated branch and
`go test -timeout 600s -count=1 -v -run TestBoardLockWaitBudgetDerivedFromNamedInputs ./internal/kanban/`
is run,
Then the reinstated branch's message `budget` / `< headroom floor` does NOT appear in the output
while `headroom factor 1 states no headroom` does — the branch silent at a budget a genuine floor
would reject, because its own `floor` term tracks the mutation down to 330ms alongside the budget,
And both the reinstatement and the mutant are discarded, `git diff` then showing only the deletion on
the test file and nothing on `board_store.go`,
And the honest limit is stated in the evidence record: this criterion corroborates §A.1's static
unreachability argument on ONE assignment of the constants — §A.1 carries the every-assignment
claim, and a single observation of silence never establishes it. A deletion also carries no positive
mutant evidence of its own, because it adds no assertion that could redden.
Evidence: the reinstated-plus-mutant run output, the grep showing the branch message absent while the
headroom message is present, and the post-discard `git diff --stat`.

**AC-VFG-008** — the record is written and the scope held
Given the committed change,
When `git diff --stat` is read against base `3f03d9c36`,
Then exactly one source file is changed, `internal/kanban/board_lock_wait_test.go`,
And `git diff` shows no change to `TestBoardLockWaitBudgetCoversSerializedMutations` and no change
to any constant in `internal/kanban/board_store.go`,
And the rewritten comment names the four retained assertions as REQ-BLB-002's enforcement site and
states why no floor-versus-budget comparison appears in the function,
And the verification load REQ-VFG-007 constrains is verified from the evidence record itself:
`grep -rn 'go test' .moai/reports/t378/` shows every recorded invocation scoped to
`./internal/kanban/`, with zero occurrences of `./...`, zero of `-race`, and zero trailing `&`
backgrounding — the grep output quoted in full, so an empty match set is distinguishable from an
unrun check,
And `moai spec lint --strict` scoped to `SPEC-VACUOUS-FLOOR-GUARD-001` reports 0 errors.
Evidence: `git diff --stat`, the targeted `git diff` hunks, the comment quoted, the
verification-load grep with its full output, the lint output.

## §D Exclusions — what this SPEC does NOT build

### Out of Scope — t372's landed guard

- Modifying, moving, renaming, or re-messaging `TestBoardLockWaitBudgetCoversSerializedMutations`.
  It is card t372's guard, landed, and under an open observation window (AC-SIV-013 requires at
  least 5 non-cancelled develop heads). Touching it would perturb a card being measured.
- Extending or restating its cross-file constant-coherence relation anywhere in this change.

### Out of Scope — retuning the lock policy

- Changing the landed values of `boardLockSupportedWriters` (10), `boardLockCIMutationCost` (33ms),
  or `boardLockHeadroom` (5). A mutant planted and reverted under REQ-VFG-005 is a measurement
  instrument, not a retune, and no constant change appears in the committed diff.
- Widening or narrowing `boardLockWaitBudget`, `boardLockWaitMin`, or `boardLockWaitMax`.
- Any claim that the budget suffices on a real machine. t370 back-derived the CI `-race`
  per-mutation cost at 42-105ms against the declared 33ms; that gap is a known, recorded debt owned
  elsewhere, and this SPEC neither closes nor contradicts it.

### Out of Scope — production code

- Any change to `internal/kanban/board_store.go` beyond comments, and any change at all to the lock
  implementation files or to `backlog_store.go`.
- The lock's fairness, its substrate, or the retry-loop shape.

### Out of Scope — a retry-policy-derived floor

- Constructing a replacement floor from the retry-policy constants (`boardLockWaitMin`,
  `boardLockWaitMax`, `boardLockWaitStep`). §A.4 establishes that such a floor would be a real
  comparison over terms the budget expression does not supply, so the option is left open rather
  than argued away — but choosing a defensible multiplier needs evidence about how many worst-case
  per-attempt waits a contender must absorb, which this card neither holds nor measures. Whether
  that guard is worth adding is a separate card.

### Out of Scope — other vacuous guards

- A repository-wide sweep for self-comparison assertions of the same shape. This SPEC repairs the
  one instance card t378 names and measured. A sweep is a separate card, and inferring further
  instances from this one would be an unverified defect claim.

### Out of Scope — reopening t354

- Amending `SPEC-BACKLOG-LOCK-BUDGET-001` or its closed card. REQ-BLB-002 is not weakened by this
  change; §A.3 shows its floor was already enforced input-wise, and this SPEC records that fact in
  the guard's comment rather than in the predecessor SPEC.

### Out of Scope — CI and load

- Running `go test ./...`, running the package under `-race` for this card's verification, watching
  CI, opening a PR, or pushing. Plan-phase produces artifacts only.

## §E Constraints

- **Verification load** — one serial `go test` per observation, scoped to `./internal/kanban/`,
  no `-race`, no background process, no generated load (REQ-VFG-007).
- **Blast radius** — one test file, one function: roughly 10 lines removed and a comment rewritten.
- **Evidence persistence** — mutant censuses, RED outputs, and post-revert GREENs are written under
  `.moai/reports/t378/` so the cited paths still resolve at audit time.
- **A right answer reached through an unchecked premise** — recorded because the defect had the same
  shape this card exists to repair. The 0.1.0 draft of §A.4 argued that terms outside the budget
  expression would have to be the stress constants; that exhaustiveness claim was false, since
  `boardLockWaitMin` / `boardLockWaitMax` / `boardLockWaitStep`
  (`internal/kanban/board_store.go` lines 131-136) are a second, independently-authored family. The
  conclusion — deletion — was correct, but the premise supporting it had never been verified, and a
  right answer reached through an unchecked premise is not a verified answer: it is the same failure
  mode as a guard that reads as coverage and never fires. The repair (0.2.0) was to move the
  argument onto the ground that had been verified — REQ-BLB-002's intent is discharged input-wise,
  §A.3 — rather than to keep an unverified exclusivity claim that happened to point the same way.

## §F Tier classification

**Tier S.** Blast radius is one file (`internal/kanban/board_lock_wait_test.go`), far inside the
Tier S bounds of `< 300 LOC` and `< 5 files`; the change is a deletion plus a comment rewrite.
7 requirements and 8 acceptance criteria, both inside the Tier S ceiling of 8/8. Artifact set is
the Tier S set — `spec.md` + `plan.md`, with acceptance criteria inline in §C — plus `progress.md`,
which is emitted at every tier and does not count toward the artifact total. plan-auditor PASS
threshold: 0.75.

## §G Cross-references

- `SPEC-BACKLOG-LOCK-BUDGET-001` (card t354) — origin of the defective branch; REQ-BLB-001/002.
- `SPEC-STRESS-INVARIANT-VERDICT-001` (card t372) — REQ-SIV-010 names this branch as the
  self-comparison precedent; its §E defers the repair, which this SPEC picks up. AC-SIV-013's
  observation window is why its guard is untouchable here.
- `.moai/reports/t372/mutant-headroom4-orchestrator.md` — the landed dynamic evidence quoted in §A.2.
- `.claude/rules/moai/core/verification-claim-integrity.md` — the evidence discipline REQ-VFG-005
  implements.

---
id: SPEC-CLI-WORKTREE-FLAG-RACE-001
title: "Four parallel sibling tests write two shared package-level seams — an ownerless confirmed DATA RACE in internal/cli/worktree_branch_flag_test.go"
version: "0.1.2"
status: completed
created: 2026-09-03
updated: 2026-09-04
author: manager-spec (card t464)
priority: P1
phase: "v3.1.5 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "data-race, test-isolation, t.Parallel, package-global-seam, internal-cli, ownerless-defect, regression-guard"
tier: M
related_specs:
  - SPEC-TEMPDIR-CLEANUP-RACE-001
---

# SPEC-CLI-WORKTREE-FLAG-RACE-001 — the four-sibling seam race

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-09-03 | manager-spec (card t464) | Initial draft. Scope corrected from the card text: FOUR sibling tests, not three (`_NoFlagIsNoop` was the omitted one). Both repair options carried; no winner declared — the choice is deferred to run-phase per explicit lead instruction. |
| 0.1.1 | 2026-09-03 | manager-spec (card t464) | Plan-audit iteration 1 repairs D1-D3 + D8. Text-only; no scope, requirement, or design change, and the A-vs-B decision stays open. **D1**: withdrew the false "every race frame lands in the test file" claim — the production seam `worktree_branch_flag.go` is a frame 5× (report lines 291 `:70`, 461/920/1288/1361 `:74`); §A.2 now states writes-in-tests / reads-through-production-seam, matching §A.4, and adds the omitted frames `:69` and `:175`. The same error is present in the **card t464 body** (its "레이스 쌍 … `:130` 쓰기 ↔ `:162` 읽기" wording), so this is a card-text error the SPEC diverges from, not only a SPEC-internal one — recorded as divergence ② in §A.7 (added in v0.1.2). **D2**: F1 and the §C exclusion overstated the gap and contradicted §A.5 — the package-wide scan WAS performed (two independent function-body scans, zero out-of-file hits); reworded to state what was observed, retaining only the syntactic-match residual. **D3**: remapped REQ-WFR-003 from AC-WFR-005 (compile-only, executes nothing) to AC-WFR-001b + AC-WFR-002 + AC-WFR-004. **D8**: the RED run panicked before exhausting `-count=20` (report line 1344, `FAIL … 1.543s`), so 23 is recorded as a floor, and the panic is recorded as a second manifestation of the same root cause with its own GREEN check. |
| 0.1.2 | 2026-09-03 | manager-spec (card t464) | Text-only, per lead instruction. Added **§A.7 Divergence from the card text**, consolidating both places where this SPEC disagrees with card `t464`'s body: ① three siblings vs the measured four (omitted: `_NoFlagIsNoop`, line 65), and ② the card's claim that the racing pairs lie wholly inside the test file vs the measured reads reaching through the production seam `worktree_branch_flag.go:70`/`:74` (a frame 5×). Records that the lead reproduced both, accepted them as **card-body** errors, and ruled the card body stays as written while this SPEC is canonical for run-phase. Divergence ② changes the description only — the judgment (not a production defect) is unchanged. No scope, requirement, AC, or design change; the A-vs-B decision stays open. |

---

## §A Context

### A.1 Ownership — why this card exists at all

The defect has **no owner**. The lead confirmed the introducing card `t295` has 0 rows in `items`
and exists in `archived_items` with final state `picked`, and that its commit `a05b9c4d8` is an
ancestor of `origin/develop`. A closed, landed card cannot be the owner of a defect its change
left behind, so the defect was re-issued as card `t464`.

### A.2 Evidence base

Single evidence base, measured by this lane in this tree:

- **Tree**: worktree branch `WT-worktree-flag-race` at HEAD `d592b0551` (identical to
  `origin/develop` at measurement time).
- **RED command**: `go test ./internal/cli/ -run 'TestResolveWorktreeExistingBranch' -count=20 -race`
- **RED result**: exit `1`, `WARNING: DATA RACE` × 23 — **a floor, not a total** (see below).
- **Verbatim output**: `.moai/reports/t464/red-race-d592b0551.txt` (1369 lines). This path is the
  RED evidence every AC below judges against.

**The 23 is a floor from a truncated run.** The run did not exhaust `-count=20`: it **panicked**
and aborted at report line 1344 —
`panic: Log in goroutine after TestResolveWorktreeExistingBranch_NoFlagIsNoop has completed: materialize must not run without --branch`
— followed by `FAIL github.com/modu-ai/moai-adk/internal/cli 1.543s`. A complete 20-iteration run
would have produced at least 23 warnings, and plausibly more. Any statement of the form "23
warnings" MUST be read as `≥ 23`; no total is claimed.

**The panic is a second manifestation of the same root cause**, not incidental noise. Its stack
(report lines 1355-1364) shows `_MaterializeErrorPropagates` at
`worktree_branch_flag_test.go:173` calling `resolveWorktreeExistingBranch`, which reaches through
the seam at `worktree_branch_flag.go:74` into **`_NoFlagIsNoop`'s** stub closure at
`worktree_branch_flag_test.go:69` — a stub belonging to a test that had already completed, whose
`t.Error` then panics the run. That is exactly the mechanism in §A.4, surfacing as a panic instead
of a race warning. Run-phase treats the panic's disappearance as **part of** the fix, not as
unrelated flakiness.

**Where the racing frames land.** The racing **writes** are all in the four test siblings —
`internal/cli/worktree_branch_flag_test.go` at lines 67/68/69/72, 98/99/102/104/105,
128/130/131/133/134, 162/163/164/165/169/170/173/175 (`:175` appears 9× inside `WARNING` blocks,
e.g. report lines 519 and 587). The racing **reads** reach through the **production seam**: the
production file `internal/cli/worktree_branch_flag.go` appears as a frame 5 times — report line 291
(`:70`, the `findProjectRootFn()` call) and lines 461 / 920 / 1288 / 1361 (`:74`, the
`launcherWorktreeMaterialize(...)` call). Frame-attribution measurement:
`grep -o 'internal/cli/[a-z_]*\.go' .moai/reports/t464/red-race-d592b0551.txt | sort | uniq -c`
→ `46 main_test.go`, `52 worktree_branch_flag_test.go`, `5 worktree_branch_flag.go`.

The production file being on the stack does **not** make this a production defect: it is the
call path through which a test's stub is reached, which is what §A.4 describes.

### A.3 Scope — four siblings, not three

The card text said THREE. Function-body parsing of the file corrects this to **FOUR**. Each of the
following calls `t.Parallel()` **and** assigns the package-level globals `findProjectRootFn` and/or
`launcherWorktreeMaterialize`, restoring them in `t.Cleanup`:

| Test | Line | Globals written |
|---|---|---|
| `TestResolveWorktreeExistingBranch_NoFlagIsNoop` | 65 | `launcherWorktreeMaterialize` |
| `TestResolveWorktreeExistingBranch_WiresAndStrips` | 95 | both |
| `TestResolveWorktreeExistingBranch_RejectsBadUsage` | 125 | both |
| `TestResolveWorktreeExistingBranch_MaterializeErrorPropagates` | 160 | both |

`_NoFlagIsNoop` is the one the card omitted; its frames (67/68/**69**/72) appear in the RED output,
so it is a participant in the race and not a bystander. Line 69 is load-bearing: it is the
`t.Error` inside its stub closure, and it is the frame that panics the run when another sibling
reaches that stub after `_NoFlagIsNoop` has completed (§A.2).

This is **divergence ①** from the card body; see §A.7 for the consolidated record of both.

### A.4 The mechanism

The four siblings run concurrently with each other. Each overwrites the shared package globals with
its own stub and reverts them in `Cleanup`. While sibling A's stub is live, sibling C calls through
it — so C observes A's behaviour instead of its own, and an assertion such as
`err != errBranchMaterializeFailed` fails. This is **sibling-vs-sibling contention inside one
file**, not external interference from another package or another test file.

The production function under test reads both globals directly:
`internal/cli/worktree_branch_flag.go:70` (`findProjectRootFn()`) and `:74`
(`launcherWorktreeMaterialize(...)`), with the seam declared at `:50`.

### A.5 Excluded candidates — investigated and rejected

Two out-of-file candidates were investigated and **rejected as false positives**. They are recorded
here with their reason so a later reader does not re-open them:

| Candidate | Reason rejected |
|---|---|
| `coverage_test.go:621 TestRunGLM_WithConfig` | Calls `t.Setenv`, which is mutually exclusive with `t.Parallel()` (the runtime panics). The matched `t.Parallel()` string was a doc-comment on the FOLLOWING function, not a call in this one. |
| `launcher_test.go:793 TestApplyGLMMode_NoSettingsLocalPollution` | Same reason — `t.Setenv` present, matched `t.Parallel()` belonged to the following function. |

### A.6 Why this is not folded into t278

Folding t464 into t278 was proposed and **rejected**; the lead accepted this lane's judgment.
`t278` groups flakes sharing a common factor: local darwin green, one red CI ubuntu run,
environment-dependent, cause not yet localized. `t464` reproduces **100% locally under `-race`**
with the cause pinned to a named file and specific line numbers, so it needs no
environment-narrowing investigation — the work t278 exists to do is already done here.

### A.7 Divergence from the card text

[HARD] **This SPEC disagrees with card `t464`'s body in two places, and this SPEC is the
canonical source for run-phase.** The lead reproduced both divergences, accepted them as errors in
the **card body** rather than in this SPEC, and ruled that the card body is a queue record which
stays as written — so a reader who opens card `t464` later will still see the original wording.
This subsection exists so that reader is not misled.

| # | What the card body says | What was measured | Where the evidence lives |
|---|---|---|---|
| ① | **Three** sibling tests | **Four**. The card omits `TestResolveWorktreeExistingBranch_NoFlagIsNoop` (line 65), whose frames `67/68/69/72` appear in the RED output — a participant, not a bystander. | §A.3 |
| ② | The racing pairs lie wholly inside the test file — its wording: "레이스 쌍 `worktree_branch_flag_test.go:130` 쓰기 ↔ `:162` 읽기, `:102` 쓰기 ↔ `:130` 읽기" | The racing **writes** are in the test file, but the racing **reads reach through the production seam** `worktree_branch_flag.go:70` / `:74`, which appears as a frame **5×** in the RED output (`uniq -c` → `46 main_test.go`, `52 worktree_branch_flag_test.go`, `5 worktree_branch_flag.go`). | §A.2 |

**Divergence ② changes the description, not the judgment.** This is still **not** a production
defect: the seam is the call path through which one sibling reaches another sibling's stub, exactly
as §A.4 describes. What was wrong was the claim that the race is confined to the test file — a
run-phase implementer working from that claim would look for both sides of every pair inside
`worktree_branch_flag_test.go` and not find them.

Neither divergence changes scope, requirements, or the open A-vs-B repair decision.

---

## §B Requirements (GEARS)

**REQ-WFR-001** (Ubiquitous) — The `internal/cli` test package shall contain no data race on the
package-level seams `findProjectRootFn` and `launcherWorktreeMaterialize` among the four
`TestResolveWorktreeExistingBranch_*` sibling tests.

**REQ-WFR-002** (Event-driven) — **When** `go test ./internal/cli/ -run 'TestResolveWorktreeExistingBranch' -count=20 -race`
is executed against the repaired tree, the run shall exit `0` and its output shall contain zero
occurrences of `WARNING: DATA RACE`.

**REQ-WFR-003** (Unwanted) — The repair shall not change the observable behaviour of
`resolveWorktreeExistingBranch`: `--branch` token stripping, error propagation, and the no-flag
no-op path shall produce the same outcomes as before the repair.

**REQ-WFR-004** (Unwanted) — The repair shall not reduce test coverage of the four scenarios. No
sibling test shall be deleted, merged, skipped, or stripped of an assertion in order to remove the
race.

**REQ-WFR-005** (Where — capability gate) — **Where** the run-phase implementer selects the
injection option (Option B, §D.2), the production-side change shall be confined to
`internal/cli/worktree_branch_flag.go` and its four sibling tests in
`internal/cli/worktree_branch_flag_test.go`; no other consumer of `findProjectRootFn` — production
or test — shall be modified under this card.

**REQ-WFR-006** (Event-driven) — **When** the repair lands, a `-race` repetition run shall exist as
the standing regression evidence for this defect, and its command and expected signal shall be
recorded in `progress.md` §E.2.

---

## §C Exclusions

This SPEC deliberately does not build the following. Anything listed here that later proves
necessary is a new card, not a scope expansion of this one.

### Out of Scope — a package-wide `findProjectRootFn` isolation sweep

`findProjectRootFn` is a package-wide seam: **23 test files** in `internal/cli` assign it
(`cc_test.go` 18 assignments, `mx_query_test.go` 24, `coverage_improvement_test.go` 14, and 20
others), and **24+ production call sites** read it.

**The other 22 files were scanned, and they are clean on this pattern.** Two independent
function-body scans across all `internal/cli/*_test.go` — this lane's, and the auditor's `awk`
scan that resets state at each top-level `}` so a doc-comment cannot bleed into the following
function — both returned **exactly the four in-file siblings and zero out-of-file hits** for the
`t.Parallel()`-plus-global-write combination. This is a measurement, not an assumption; the
exclusion below is scope discipline, not an unexamined gap.

- Converting the package-wide seam to a non-global injection surface is a separate card. Nothing
  here requires it, because no second offender was found.
- Re-scanning under a different technique (see F1's residual) is a separate card.

### Out of Scope — the two rejected out-of-file candidates

- Re-investigating `coverage_test.go:621` or `launcher_test.go:793`. Both were excluded with a
  stated reason (§A.5); reopening them requires new evidence, not a re-read.

### Out of Scope — flake investigation under t278

- Environment-narrowing work (darwin vs ubuntu, CI vs local) for this defect. The cause is pinned
  to file and line; there is nothing to narrow.
- Any change to `t278`'s own scope or grouping.

### Out of Scope — production behaviour of the `--branch` launcher path

- Changing what `resolveWorktreeExistingBranch` does, what it validates, or what it writes.
- Changing `launcherWorktreeMaterializeReal`, the worktree materialization itself, or the registry
  entry it creates.

### Out of Scope — full-suite verification as repair evidence

- Running or citing `go test ./...`. Manifestation is schedule-dependent, so a full-suite green run
  establishes nothing about this defect (§D.1). It is excluded as evidence, not merely skipped for
  cost.

---

## §D Constraints

### D.1 Evidence shape is mandated, not preferred

[HARD] Manifestation is **schedule-dependent**. A single green run — of this package or of the whole
suite — is consistent with the race still being present, because the Go scheduler may simply not
have interleaved the siblings that way. "Full suite green" is therefore **NOT repair evidence** for
this defect and MUST NOT be cited as such.

Every AC that asserts the fix is expressed as a `-race` + `-count` **repetition** that fixes RED
first and then shows GREEN. The RED half is not ceremony: without an observed RED on the same
command shape, a GREEN says nothing about whether the command can detect the defect at all.

### D.2 Both repair options are carried; run-phase decides

[HARD] Explicit lead instruction: the choice is made in run-phase (`run에서 확정`). This SPEC states
both options and the criteria for choosing, and **declares no winner**. A run-phase implementer who
finds this section already decided should treat that as drift and stop.

**Option A — remove `t.Parallel()` from the four siblings.**
Delete the four `t.Parallel()` calls (lines 66, 96, 126, 161). The globals are then written only by
serially-executing tests, so no concurrent write exists.

**Option B — replace global assignment with injection semantics.**
Parameterize the seam so the four tests supply their dependencies per-call rather than writing a
shared global. Concretely: give `resolveWorktreeExistingBranch` an explicit dependency argument (or
a small struct receiver) carrying the root-resolver and the materializer, defaulting to the current
globals at the production call site, and have the four tests pass their stubs directly.

### D.3 Decision criteria a run-phase implementer applies

Stated so the decision is made on grounds, not on taste. No weighting is supplied — that too is the
implementer's call.

| Criterion | Favours A when | Favours B when |
|---|---|---|
| **Blast radius** | The 23-file package-wide seam usage makes any production-signature change risky to keep confined; A touches four lines in one test file. | REQ-WFR-005's fence holds cleanly — the production change stays inside `worktree_branch_flag.go` and its own call site. |
| **What is lost** | Nothing is lost that the suite depends on; these four tests are sub-millisecond stubs, so their parallelism buys no measurable wall-time. | Parallelism is retained, and the shared-global hazard is removed rather than avoided. |
| **Recurrence** | A later author re-adding `t.Parallel()` re-opens the defect; A leaves no structural barrier, only the regression guard. | The hazard cannot be re-introduced by adding `t.Parallel()`, because there is no shared global left to write. |
| **Consistency with the package** | The package's dominant idiom is global-seam assignment (23 files); A stays inside that idiom. | B introduces a second idiom into one file, which is either a bridgehead for a later sweep or an inconsistency, depending on whether that sweep is ever funded. |
| **Reviewability** | A four-line deletion is trivially reviewable. | A signature change plus four test rewrites needs real review. |

### D.4 Verification scope

- Verification is scoped to the touched package: `go test ./internal/cli/...`. The repository
  prohibits running the full suite locally; the full-suite verdict is CI's.
- `internal/cli` is slow — use a timeout floor of at least `-timeout 600s` on any whole-package run.
- The repetition count in the RED command (`-count=20`) is the measured shape that produced ≥ 23
  warnings before panicking (§A.2); the GREEN run uses the same shape so the two are comparable.
  Note the asymmetry the truncation creates: the RED aborted early, so the GREEN run will execute
  **more** iterations than the RED did. That direction is safe — a GREEN that survives more
  iterations than the RED reached is stronger evidence, not weaker.

### D.5 Tier

Tier **M**. Reasoning is recorded in `plan.md` §D.

---

## §E Success criteria

1. The command in REQ-WFR-002 exits `0` with zero `WARNING: DATA RACE`, and the RED output at
   `.moai/reports/t464/red-race-d592b0551.txt` is cited as the pre-repair baseline.
2. All four sibling tests still exist, still assert what they asserted, and still pass.
3. `go test ./internal/cli/... -race` is green for the package (schedule-dependence acknowledged —
   this is a non-regression check, not the repair evidence).
4. The diff's file list is inside the fence stated by REQ-WFR-005.
5. The chosen option, and the criteria that decided it, are recorded in `progress.md` §E.2.

---

## §F Risks and gaps

| # | Item | Kind |
|---|---|---|
| F1 | The other 22 test files assigning `findProjectRootFn` **were** scanned package-wide (two independent function-body scans) and returned zero hits for the `t.Parallel()`-plus-global-write pattern. Residual: both scans match a **syntactic** pattern, so a global written indirectly — through a helper function, a table-driven setup, or an alias — would not be caught. | Residual risk |
| F2 | The RED was measured on darwin/arm64 only. Whether the same command produces the same count on linux/amd64 is unmeasured. | Gap |
| F3 | `-count=20` reproduced the race in this tree; it is a measured shape, not a proven-minimal one. The RED never completed all 20 iterations (it panicked, §A.2), so the shape's behaviour across a **full** 20-iteration run is itself unobserved. A repaired tree passing at `-count=20` is not proof it passes at every count. | Residual risk |
| F4 | Under Option B, the default-to-global fallback at the production call site preserves the seam for all other consumers — so neither option removes the shared global from the package. No second offender exists today (F1), but the structural hazard that would let one appear later is untouched by either option. | Residual risk |
| F5 | The regression guard is the repetition command itself, not a new test. If CI does not run `-race` on this package, the guard is only exercised by hand. | Residual risk |

---

## §G Cross-references

- Card `t464` (this card); introducing card `t295` (archived, commit `a05b9c4d8`).
- RED evidence: `.moai/reports/t464/red-race-d592b0551.txt`.
- Subject file: `internal/cli/worktree_branch_flag_test.go`.
- Production seam: `internal/cli/worktree_branch_flag.go:50,70,74`; `internal/cli/cc.go:17`.
- `SPEC-TEMPDIR-CLEANUP-RACE-001` — a sibling test-isolation race SPEC with the same evidence
  discipline (RED-before-GREEN under `-race`).
- `AGENTS.md` §1 (no unobserved claim), §4 (how verification is run).

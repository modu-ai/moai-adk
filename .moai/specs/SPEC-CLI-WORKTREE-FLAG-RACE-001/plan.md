---
id: SPEC-CLI-WORKTREE-FLAG-RACE-001
title: "Implementation plan — the four-sibling seam race"
version: "0.1.2"
created: 2026-09-03
author: manager-spec (card t464)
---

# Implementation Plan — SPEC-CLI-WORKTREE-FLAG-RACE-001

Sections are ordered by decision-reversibility: the option choice comes first because it is the
decision most likely to change and the most expensive to change late. Mechanical steps sit at the
bottom.

## §A Context

Card `t464`, Class B (defect, cause established), ownerless — the introducing card `t295` is closed
and landed. The cause is pinned to `internal/cli/worktree_branch_flag_test.go`; the RED is
reproducible 100% locally under `-race`. See `spec.md` §A.

Base: worktree `WT-worktree-flag-race` at `d592b0551`, identical to `origin/develop` at measurement.

## §B The open decision (highest reversibility cost — decide first)

[HARD] **This plan does not pick an option.** The lead instructed that the choice be made in
run-phase. The two options and the criteria are in `spec.md` §D.2 and §D.3; they are not restated
here.

What run-phase must do before writing any code:

1. Read `spec.md` §D.3 and evaluate each of the five criteria against this tree.
2. State the choice and the criteria that decided it, in `progress.md` §E.2, **before** the first
   edit — so the reasoning is recorded when it is made, not reconstructed afterwards.
3. If the evaluation produces no clear answer, that is a blocker report to the orchestrator, not a
   coin flip. The default is not "Option A because it is smaller".

Consequence of getting this wrong late: Option B changes a production signature, so switching from
B to A after implementation discards the whole diff, while switching A→B discards four lines.
The asymmetry is why this is decided first and recorded.

## §C Known issues entering the run

- `findProjectRootFn` is written by 23 test files in this package, though only these four combine
  that write with `t.Parallel()` (scanned package-wide — `spec.md` §C). REQ-WFR-005 fences the
  change; the fence is the thing most likely to be breached under Option B, because a signature
  change invites "while I'm here" edits to neighbouring files. `spec.md` §C names this out of scope.
- The RED capture is **truncated** — it panicked before exhausting `-count=20` (`spec.md` §A.2). The
  panic is a second face of the same root cause, so a repair that removes the race warnings but
  leaves the panic is not done. AC-WFR-001b greps for both.
- **The card body is wrong in two places and will not be corrected** (lead ruling). `spec.md` §A.7
  is the consolidated record; this SPEC is canonical for run-phase. Any step that works from the
  card text rather than from `spec.md` will (①) miss `_NoFlagIsNoop`, and (②) look for both sides
  of every racing pair inside the test file and fail to find them.
- `internal/cli` is a slow package. Whole-package runs need `-timeout 600s`.

## §D Tier classification

Tier **M**, for two reasons that are not about line count:

- The repair option is **undecided at plan time**, and one branch of it changes a production
  function signature. A Tier S SPEC (AC inline, no `acceptance.md`) cannot carry two alternative
  designs plus the criteria that separate them without the AC set becoming ambiguous about which
  design it judges.
- The AC set must carry a mandated evidence shape (RED-then-GREEN under `-race` repetition) plus a
  blast-radius fence AC, which is more than a Tier S inline block holds legibly.

Artifact set therefore: `spec.md` + `plan.md` + `acceptance.md` + `progress.md`. No `design.md` /
`research.md` — the mechanism is already established and needs no research pass.

## §E Pre-flight

Run before the first edit:

```bash
git rev-parse --short HEAD                 # expect d592b0551 (or a later develop-parity head)
git branch --show-current                  # expect WT-worktree-flag-race
test -s .moai/reports/t464/red-race-d592b0551.txt && echo RED-EVIDENCE-PRESENT
```

If HEAD has moved past `d592b0551`, re-run the RED command and persist a new
`.moai/reports/t464/red-race-<sha>.txt` before proceeding — the AC baseline attribution names a
tree, and a stale one is unattributed.

## §F Milestone M1 — repair and prove (single milestone)

Priority: High. One milestone; the work does not decompose usefully.

1. Record the option choice and its criteria in `progress.md` §E.2.
2. Re-confirm RED on the current head (the persisted output, or a fresh run if the head moved).
3. Apply the chosen repair, confined to the REQ-WFR-005 fence.
4. Run the GREEN command (`acceptance.md` AC-WFR-001b) and persist its verbatim output to
   `.moai/reports/t464/green-race-<sha>.txt`.
5. Run the package non-regression check (AC-WFR-004).
6. Verify the diff's file list against the fence (AC-WFR-003).
7. Cross-compile check (AC-WFR-005).
8. Fill `progress.md` §E.2 / §E.3.

## §G Anti-patterns for this card

- **Citing a green full-suite run as repair evidence.** Prohibited by `spec.md` §D.1 and by the
  repository's local-full-suite prohibition. It is wrong on both axes.
- **Running the GREEN command once at `-count=1`.** A single pass under a schedule-dependent defect
  asserts nothing.
- **Fixing the race by deleting or skipping a sibling test.** Violates REQ-WFR-004.
- **Sweeping the other 22 `findProjectRootFn` test files.** Out of scope (`spec.md` §C); they were
  already scanned and returned zero hits, so a sweep re-measures a settled question.
- **Re-opening `coverage_test.go:621` / `launcher_test.go:793`.** Excluded with a stated reason
  (`spec.md` §A.5). Re-opening needs new evidence.
- **Overstating what the package-wide scan established.** It found zero `t.Parallel()`-plus-**direct**
  global writes. It did not, and cannot, rule out an indirect write routed through a helper or an
  alias (`spec.md` §F1).
- **Reading "23 warnings" as a completed-run total.** The run was truncated; it is a floor.
- **Working from the card text's "three siblings".** It is four.

## §H Cross-references

- `spec.md` §A.3 (the four siblings), §D.2-D.3 (the options and criteria), §C (exclusions).
- `acceptance.md` §D (the AC matrix).
- `.claude/rules/moai/core/verification-claim-integrity.md` §1, §2.

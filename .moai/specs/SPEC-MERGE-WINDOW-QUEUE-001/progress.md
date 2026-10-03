# SPEC-MERGE-WINDOW-QUEUE-001 — Progress

> Card t1479 · created 2026-10-03 by manager-spec (plan phase)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md + design.md + research.md (Tier L set) +
  decision-index.md (`interview.decision_gate: on`) + progress.md.
- Branch `WT-merge-window-queue`, renamed in place; `git merge develop` (local) → "Already up to
  date" at `d7112d005` (tree `632f65b47aa5`).
- SPEC id regex check (Bash) → `PASS`; uniqueness: no `MERGE-WINDOW` entry in this tree's or the
  develop worktree's `.moai/specs/`.
- RED-now baseline cells E1-E10 measured on that tree (research.md §R1); E11 (no-`--wait`
  acquire fixture) captured on a build of this branch (Go code = `d7112d005`) and committed ahead
  of any code in `3bc274dac` (`.moai/reports/t1479/baseline-acquire-nowait/`).
- Decisions: Q1 = recorded operator approval; Q2-Q6 (v0.2.0) and Q8-Q13 (v0.3.0, plan-audit
  iteration 1) = leader decisions (mission contract 07d28c4b) in the verdict lines. Target v3.2.0.
- v0.4.0: Q14 (heartbeat 15 s / window 60 s / re-entry grace 120 s), Q15 (starvation counting +
  three-requeue bound), Q16 (non-test-command residual risk) recorded as leader decisions. No open
  decision blocks run entry.
- Plan audit iteration 1: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter1.md`, verbatim copy).
  v0.3.0 revision: 25 REQ / 25 AC, contiguous 001-025; push verb moved to SPEC-CANDIDATE-CI-001.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

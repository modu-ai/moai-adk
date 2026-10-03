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
- Plan audit iteration 2: FAIL 0.69, regressed — STOP (`.moai/reports/t1479/plan-audit-iter2.md`,
  verbatim copy). v0.5.0 scope reduction per leader decision Q17: reserved tickets, `--slice` /
  between-slices, front-once, requeue counter and three-requeue rule removed; plain FIFO with the
  owner pid on tickets stamped onto the holder at promotion (N1). Q18 lane merge verb
  `moai integration merge --card <id>` folded into REQ-MWQ-017. N5 hand-off item (e) in research
  §R6 (for the leader to forward to t1478). 23 REQ / 23 AC. No open decision blocks run entry.
- Plan audit iteration 3: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter3.md`, verbatim copy;
  first ceiling hit). v0.6.0 delta per leader decision Q19 inside the auditor's fix_scope: one merge
  path (complete calls the merge step, adopts a prior landing), SHA pinning and per-cause failure
  exits with abort + clean check (else `hold`), integration target on tickets and copied at
  promotion; O1/O2 one-liners, O3/O4 noted in research §R5. 23 REQ / 23 AC.
- Plan audit iteration 4: FAIL 0.75, claude + codex agree (`.moai/reports/t1479/plan-audit-iter4.md`,
  verbatim copy). Operator decision Q20 (AskUserQuestion 2026-10-03): one more narrow delta round, no
  new REQ. v0.7.0: complete's card gates before develop moves; adoption tied to the branch's current
  tip and a valid record, with the clause order in REQ-MWQ-019; ancestry precondition plus a defined
  post-merge outcome (commit left, `hold`); holder check reads first and refuses before any queue
  mutation. 23 REQ / 23 AC.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

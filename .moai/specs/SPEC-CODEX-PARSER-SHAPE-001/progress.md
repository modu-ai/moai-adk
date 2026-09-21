# SPEC-CODEX-PARSER-SHAPE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1053 · worktree `.claude/worktrees/t1053` · branch `WT-codex-parser-shape`
- Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md` (Tier M)
- Evidence base: `.moai/reports/t1053/verdict.md` (tree `8b55fc8f0`). No
  measurement re-run; no figure introduced that is absent from that file.
- Status: `draft`.
- AC-CPS-001 / AC-CPS-002 (live-convention comparison): **SATISFIED 2026-09-21**,
  result same-shape. Record: `.moai/reports/t1053/live-convention-20260921.md`
  (tree `a5c3f5dc6`, codex-cli 0.155.1). The pre-run gate is no longer blocking;
  the remaining Kickoff decision is the §C candidate selection, which is the
  operator's.
- Plan-phase corrections after the live call (three defects codex found in this
  SPEC, each verified against the repository before acceptance): AC-CPS-001's
  no-findings edge case now stays OPEN instead of closing; candidate (c)'s
  predicate is scoped with `GateUnmet == ""` plus a mandatory control case
  (REQ-CPS-006a); AC-CPS-006 now requires exact count and content plus a
  partial-drift case.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

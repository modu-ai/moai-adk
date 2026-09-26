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
- 2026-09-26 · v0.2.0 amendment (card t1203, GitHub #1718; worktree
  `.claude/worktrees/t1203`, branch `WT-codex-parser-shape`): the #1718 real case
  added from `.moai/reports/t1203/verdict.md` (tree `df526c9a9`) — spec.md §A.6,
  §A.4 conclusion scoped, §C #1718 coverage table, candidate (d), §C.1 conflict;
  REQ-CPS-012..014; AC-CPS-011..015. Status stays `draft`. The pending Kickoff
  decision now covers BOTH the §C candidate selection AND the REQ-CPS-010
  question (keep as written, or revise — spec.md §C.1, REQ-CPS-013); neither is
  taken here. The plan-audit verdict from before this amendment does not cover
  it; a fresh plan-audit is due.
- 2026-09-26 · v0.2.1 repair (card t1203): plan-audit iter-1
  (`.moai/reports/t1203/plan-audit.md`, FAIL 0.75) defects D1–D14 addressed in
  wording and structure only — no new measurement, no candidate chosen. REQ 15 /
  AC 15, unchanged. The REQ-CPS-010 question is now posed as "is the #1718
  adversarial outcome acceptable?" and excludes no candidate under either answer
  (spec.md §C.1). Both Kickoff decisions remain pending. Next: delta plan-audit
  (iter-2, the last under the Tier M ceiling) limited to D1–D14.
- 2026-09-26 · v0.2.2 repair (card t1203): plan-audit iter-2
  (`.moai/reports/t1203/plan-audit-iter2.md`, FAIL 0.83) — blocking N4-P4, N5,
  N1 and optional N2, N3, N6 addressed in wording and check commands only; no
  candidate chosen, REQ 15 / AC 15 unchanged. Iter-2 was the Tier M ceiling, so
  the next step is the orchestrator's escalation choice (PASS-with-debt, scope
  reduction, or an explicit extension), not an automatic re-audit. Both Kickoff
  decisions remain pending.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

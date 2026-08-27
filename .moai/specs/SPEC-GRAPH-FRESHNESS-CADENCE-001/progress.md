# SPEC-GRAPH-FRESHNESS-CADENCE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored for card t322 in worktree
`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t322`, branch `WT-graph-freshness-cadence`, base
`d2cba5e21`.

- Tier: **M** — artifact set `spec.md` + `plan.md` + `acceptance.md` (+ this file). **11 live
  requirements** (13 numbered, REQ-GFC-004 and 012 withdrawn at v0.2.0; ceiling 16) and **12 live
  acceptance criteria** (14 numbered, AC-GFC-011 and 012 withdrawn with them; ceiling 16).
- Baseline: every figure in `spec.md` §B re-measured in this tree during authoring. The dispatch's
  figures (55 / 0 / 5 per integration; 21 / 198 calibration re-run) reproduced exactly.
- Judgments (as of v0.2.0): `spec.md` §D.1 (a — yes, predicate in the metric), §D.2 (b — **40
  retained**, re-justified on the integration axis; the v0.1.0 proposal of 15 was reversed by audit
  D2), §D.3 (c — no pipeline refresh; attribution instead). Rejected alternatives recorded per
  judgment.

## §E.1b Plan-phase Audit Record

- Audit iter-1 (`plan-auditor`): **PASS-WITH-DEBT 0.82** (Tier M PASS threshold 0.80), MP-1..MP-4
  all pass, four blocking defects (D1 critical, D2/D3/D4 major) plus D5-D8 optional. Verdict:
  `.moai/reports/t322/plan-audit.md`.
- Remediation landed at v0.2.0; the per-defect account is the `spec.md` HISTORY row for that
  version. D1 was additionally confirmed by the orchestrator directly against source
  (`internal/mx/provenance.go:107-143` applies no extension filter; `internal/graph/meta.go:67`
  routes three non-Go directories through it), so it is an established finding rather than an
  auditor hypothesis.
- D2 reversed the threshold judgment: the v0.1.0 value 15 was derived on the per-integration axis
  while the metric is cumulative-since-stamp; on the cumulative axis 15 reds the gate roughly twice
  per day at the observed integration rate, and the streak this SPEC exists to fix carries a
  corrected cumulative of 2 — under both 15 and 40. The threshold change was therefore not
  load-bearing for the defect and is now out of scope.
- Audit iter-2 (`plan-auditor`, the Tier M ceiling — there is no iter-3): **PASS-WITH-DEBT 0.895**,
  monotonic +0.075, MP-1..MP-7 all pass, all four iter-1 blocking defects confirmed closed against
  source. Three new blocking defects (N1 major — the §D.2 cadence figure rested on the
  self-referential `6786c3fa4`; N2, N3 minor) plus five optional. Verdict:
  `.moai/reports/t322/plan-audit-iter2.md`.
- Remediation of N1/N2/N3 and optional O2/O4 landed at **v0.2.1**, applied directly rather than as a
  plan-phase re-entry (the iteration ceiling is reached). The cadence walk was re-measured in this
  tree both ways — 40-crossing at integration 10 with the outlier, **16** without it — and the
  outlier-excluded figure is now the operative one in §B.6, §D.2 and §G Q4.
- Ordering basis for t322 / t311 / t304: `spec.md` §F.
- Open questions deliberately left to the operator: `spec.md` §G (**four** — Q4 added at v0.2.0 and
  restated on the outlier-excluded cadence figure at v0.2.1).
- Evidence path: `.moai/reports/t322/`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

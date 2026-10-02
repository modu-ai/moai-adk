# SPEC-HARNESS-RETENTION-HARDEN-001 — Progress

Card t1432. Branch `WT-harness-retention-debt`, base develop `1e2151a38`.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
plan_iteration: 2 (revision of version 0.1.0 after plan-audit iteration 1 FAIL 0.75, `.moai/reports/t1432/plan-audit.md`)
plan_artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
relayed_verdicts: recorded in `decision-index.md` (relayed by the leader session on 2026-10-03, not a direct operator answer)
open_for_leader_before_kickoff: operator-held options D1-C, D2-B/C, D3-A, D4.a-B (non-default, unselected); decision-index rows Q7 and Q8 (warning surface, allocation of the one allowed test-only field), where the plan applies the `spec.md` §B defaults; plan.md B5 (hoist the draft tests behind the RED-now ledger if the auditor must re-run them), B6 (the red M0 commit) and B8 (force-added evidence under an ignored path)

Record kept here rather than in source (REQ-HRH-011): the earlier lock-behaviour design question (`.moai/reports/t1425/decision-records.md`, `lock_behavior=block_then_recheck`) did not carry the fact that the harness-observe hooks run with a 5 s timeout and `async: true`; the plan-phase facts are in `spec.md` §A (F6 row).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

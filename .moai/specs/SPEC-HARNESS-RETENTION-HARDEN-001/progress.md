# SPEC-HARNESS-RETENTION-HARDEN-001 — Progress

Card t1432. Branch `WT-harness-retention-debt`, base develop `1e2151a38`.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
plan_iteration: 3 (revision of version 0.2.0 after plan-audit iteration 2 FAIL 0.87, `.moai/reports/t1432/plan-audit-iter2.md`; iteration 1 was FAIL 0.75, `.moai/reports/t1432/plan-audit.md`)
plan_artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
relayed_verdicts: recorded in `decision-index.md` (relayed by the leader session on 2026-10-03, not a direct operator answer)
open_for_leader_before_kickoff: operator-held options D1-C, D2-B/C, D3-A, D4.a-B (non-default, unselected); decision-index rows Q7 and Q8 (warning surface, allocation of the one allowed test-only field), where the plan applies the `spec.md` §B defaults; plan.md B5 (regenerate the overlay JSON files in `red-now-drafts/` and force-add the B3 probe files), B6 (the red M0 commit, accepted), B8 (force-added evidence under an ignored path, accepted) and B9 (the pre-existing FIFO hang is recorded, not repaired; a repair needs a SPEC amendment first)

Record kept here rather than in source (REQ-HRH-011): the earlier lock-behaviour design question (`.moai/reports/t1425/decision-records.md`, `lock_behavior=block_then_recheck`) did not carry the fact that the harness-observe hooks run with a 5 s timeout and `async: true`; the plan-phase facts are in `spec.md` §A (F6 row).

Plan-audit iteration 2 optional findings left open in revision 0.3.0 (taken: O1, O2, O3, O5, O6, O7, O10; the blocking B1-B3 are fixed):

- O4 (AC-002 identity assertion can be falsely red on a file system that reuses a freed inode number): inferred, APFS observed not to reuse, Linux unobserved; the fix is a test detail that cannot be observed before the test exists, and the failure direction is a false red at CI, not a missed defect.
- O8 (AC-007 checks `late-event` by substring, so a mutant that re-encodes a late line passes): needs a changed test shape and a re-derived E-003; beyond this last pass.
- O9 (AC-012 does not pin the phrase naming the missing archive step): judged by reading; pinning a phrase needs the N1 fix to exist first.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

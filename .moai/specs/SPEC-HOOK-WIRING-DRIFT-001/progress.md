# SPEC-HOOK-WIRING-DRIFT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_version: 0.2.0
plan_complete_at: 2026-08-24
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 14
acceptance_criteria: 16
milestones: [M1, M2, M3, M4]
authored_at_head: 950cb4399
amended_at_head: 4842760a7
audit_iterations:
  - iteration: 1
    verdict: FAIL
    score: 0.807
    threshold: 0.85
    must_pass: 7/7
    mutants_constructed: 4
    findings_blocking: 7
    disposition: all 7 blocking fixed, 4 optional fixed, 0 declined
    report: .moai/reports/t216/plan-audit.md
authored_in_worktree: .claude/worktrees/t216
branch: WT-hook-wiring-drift
card: t216
source_of_record: .moai/reports/t216/{d1-chain-event,d2-unwired-scripts,d3-mx-cold-start}.md
deferred: [t242, t243, t244, file_changed.go-die-at-exit-twin, deploy-time-template-snapshot]
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

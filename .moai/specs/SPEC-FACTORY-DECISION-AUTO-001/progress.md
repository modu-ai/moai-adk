# progress.md — SPEC-FACTORY-DECISION-AUTO-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
spec_id: SPEC-FACTORY-DECISION-AUTO-001
card: t1481
tier: L
branch: WT-decision-automation
base: d7112d005
probe_tree: 5d094991fb586b9c4aadee33b663fe92b686ff18
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md, decision-index.md]
spec_version: 0.4.0
req_count: 25
ac_count: 25 (24 release-blocking, 1 RG)
plan_audit: iter1 FAIL 0.74 (.moai/reports/t1481/plan-audit-iter1.md); iter2 FAIL 0.80 (.moai/reports/t1481/plan-audit-iter2.md); revision 0.3.0 addresses N1-N7 + O1-O4; iter3 FAIL 0.83 (.moai/reports/t1481/plan-audit-iter3.md, Tier L ceiling, no regression); leader-ruled one delta round; revision 0.4.0 addresses N8-N11 + O6 within the iter3 fix_scope; iter4 delta PASS 0.885, no blockers (.moai/reports/t1481/plan-audit-iter4.md, audited_sha a13b83868)
plan_artifacts_frozen_at: a13b83868 (no spec/plan/acceptance/design/research/decision-index edit after the PASS, so the audited state and hash stay bound)
recorded_debts:
  - O9 (dispose_in: run M2): plan.md M2 row does not name the spec-workflow.md hash-subject sentence edit (local + template); run M2 follows design.md §5, which assigns it
  - O10 (dispose_in: run M3): the REQ-FR-019 Amendments obligation of REQ-FDA-016 is asserted only through AC-FDA-015; run M3 evidence names REQ-FDA-016 beside AC-FDA-015
open_decisions: none — Q1-Q26 LEADER-DECIDED 2026-10-03 (mission contract 07d28c4b)
evidence_needed_in_run: M0(a) degraded-notice rate with/without bind cache (Q5); M0(b) recheck cache cost (Q4)
release_target: v3.2.0
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

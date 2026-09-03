# SPEC-FMT-GATE-001 — progress (card t465)

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-03
tier: M
artifacts:
  - .moai/specs/SPEC-FMT-GATE-001/spec.md
  - .moai/specs/SPEC-FMT-GATE-001/plan.md
  - .moai/specs/SPEC-FMT-GATE-001/acceptance.md
baseline:
  tree: d592b0551
  command: "gofmt -l ."
  result: "154 files (.moai/reports/t465/gofmt-l.txt)"
preceding_card:
  id: t457
  branch: WT-gofmt-drift
  tip: e1fdf00d1
lead_decisions:
  d1_binding: "make fmt-check lands in the SAME commit as gate activation (single-commit delivery; no early landing before t457 — an ownerless red)"
  d2_records_only: ".golangci.yml gofmt-linter exclusion AGREED as correct scope discipline; follow-up card candidate registered by the lead — issuance is the operator's call, this lane does not issue it"
plan_audit:
  repaired:
    - "D1 — spec.md REQ-FG-006 whole-tree predicate replaced with the tracked-variant form (aligned with acceptance.md §D.3)"
  accepted_minor_debt:
    - "D2 — tip-SHA recording form: deferred to run-phase judgment by lead instruction (seen, not missed)"
    - "D3 — re-pin path: deferred to run-phase judgment by lead instruction (seen, not missed)"
```

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

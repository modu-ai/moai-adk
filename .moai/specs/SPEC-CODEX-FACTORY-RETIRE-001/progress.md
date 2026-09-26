# SPEC-CODEX-FACTORY-RETIRE-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
spec_id: SPEC-CODEX-FACTORY-RETIRE-001
spec_version: 0.3.0
card: t1242
tier: L
base_tree: 553e224f3
branch: WT-codex-factory-retire
artifacts: [spec.md, plan.md, acceptance.md, design.md, research.md, progress.md]
requirements: 23
acceptance_criteria: 25
open_clarification_markers: 0
plan_audit_history: ["iter-1 FAIL 0.71 -> revised in v0.2.0", "iter-2 FAIL 0.83 -> revised in v0.3.0 (final allowed iteration)"]
status: draft
```

- Coordinates re-measured on `553e224f3`; divergences from the design recorded in
  `research.md` §R2.
- The `[NEEDS CLARIFICATION: exact 6 paths from lead]` marker was resolved in this
  plan phase with the lead-provided list (plan.md §M5, research.md §R6). The list is the
  lead's observation; this agent did not read the develop worktree.
- Token-budget reading on the plan tree: 77539 / 77600 (headroom 61). Context only;
  AC-CFR-022 compares the merge commit with its develop parent.
- Fields the run lane writes into §E.2 (plan.md §C, §M5): `run_base:`, `budget_base:`,
  the AC-CFR-022 figure pair, the AC-CFR-020 mutant result, `m5_lead_confirmation:`
  (citing the lead-authored `m5-lead-confirm.md`, before M5 step 1), and the
  foreign-6.patch sha256 (after the confirmation line).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

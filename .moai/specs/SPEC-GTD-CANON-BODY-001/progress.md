# Progress — SPEC-GTD-CANON-BODY-001 (card t867)

## §E.1 Plan-phase Audit-Ready Signal

### Iteration 1

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- SPEC ID regex self-check: `PASS`; uniqueness `ls .moai/specs | grep -ci gtd-canon` → `0`
- Base: develop `f67d2193f`, branch `WT-gtd-canon`, commit `114737ea1`
- Plan-audit iteration 1: FAIL 0.71 (`.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md`)

### Iteration 2 (revision for D1-D14)

- D1: AC-GCB-007/010 rebased on a named baseline (spec.md §C.3, the plan-auditor's observation at `114737ea1`, not re-measured by manager-spec; the run phase re-measures at M0). The two gtd-adjacent failures are stated as not caused by this SPEC. `-timeout` and slot lease added.
- D2/D5/D10/D13: gates moved into `check-residual.sh` and `check-gtd-body.sh`, with observed RED-now and mutant controls (acceptance.md ledger L1-L7).
- D3: mirror SKILL.md uses `.claude/skills/moai/workflows/gtd.md`.
- D4: `catalog.yaml` regeneration added to the change set and to AC-GCB-009.
- D6: CHANGELOG.md and top-level `reports/` added to the historical set (13 citation lines measured).
- D7/D14: counts measured as 19 surviving scoped files and 31 paths in the change set.
- D8: comment citations required, and the verdict-survivor escape removed.
- D9: exact test names used.
- D11: runtime and codemaps wording declared Out of Scope.
- D12: REQ-GCB-011 split into 011/012 (baseline became 013).
- Status: draft — awaiting plan-audit iteration 2

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

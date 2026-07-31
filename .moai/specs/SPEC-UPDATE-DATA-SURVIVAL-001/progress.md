# SPEC-UPDATE-DATA-SURVIVAL-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]
code_baseline: d5336214e
worktree_head_at_revision: a8b42e112
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.71
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.90, testability: 0.60, traceability: 0.65}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7]
    deferred: [D8, D9, D10]
    report: .moai/reports/plan-audit/SPEC-UPDATE-DATA-SURVIVAL-001.md
```

### Epic run order (dependency sequencing)

`depends_on: [SPEC-UPDATE-REINSTALL-LOOP-002]` is a **real** dependency: E1's REQ-RIL2-015/016
(backup-before-delete for `defs.DeprecatedPaths`) is the inherited precondition behind this SPEC's
registry row 12 / §C.1 row 12, which this SPEC deliberately does not re-specify.

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. Every SPEC in this Epic is currently `draft`, so entering `/moai run` on this
SPEC before E1 closes would raise the 3-option blocker. **The dependency is satisfied by sequencing,
not by an `--ignore-deps` bypass** — the run order below is the mechanism:

| Order | SPEC | Gate to clear before the next entry |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | reaches `status: completed` — REQ-RIL2-015/016 landed |
| 2 | **`SPEC-UPDATE-DATA-SURVIVAL-001`** (this SPEC, E2) | — |
| 3+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-CONFIG-TIER-PERSIST-001`, `SPEC-CONFIG-KEY-HONESTY-001`, `SPEC-UPDATE-DOC-DRIFT-001`) | no `depends_on` edge to this SPEC |

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. If E1 slips and starting E2 early
becomes necessary, the correct move is to confirm that M1 and M3-M6 carry no E1 edge (they do not —
`plan.md` §A records that only the deprecated-path backup is inherited) and run those milestones,
leaving M2's registry row 12 open until E1 closes. That is a scope decision for the orchestrator to
surface, not a flag the run-phase agent may set on its own.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

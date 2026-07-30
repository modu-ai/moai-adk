# SPEC-CONFIG-TIER-PERSIST-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-31
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]
code_baseline: d5336214e
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.72
    threshold: 0.80
    dimensions: {clarity: 0.75, completeness: 0.75, testability: 0.75, traceability: 0.65}
    must_pass: 7/7
    resolved: [D1, D2, D3, D4, D5, D6, D7, D8, D9, D10, D15]
    deferred: [D11, D12, D13, D14]
    report: .moai/reports/plan-audit/SPEC-CONFIG-TIER-PERSIST-001.md
```

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`. Status `draft`, Tier M.
- Code baseline `d5336214e`; the worktree HEAD is a descendant that changes SPEC documents only
  (`git diff --stat d5336214e HEAD -- internal/` is empty), so every `file:line` measurement in these
  artifacts is attributable to the code baseline.
- Findings F1-F8 each re-verified; drift recorded in `plan.md` §B. F2, F3, and F7 reproduced by
  runnable probes, removed after running.
- **F1's field evidence (four section files at `-rw-------`) reproduces only in the primary
  checkout**, not in this worktree; `spec.md` §A and `acceptance.md` §A clause 3 record the
  re-attribution. F1's mechanism reproduces anywhere and is unaffected.
- The `report.yaml` open question is resolved in `plan.md` §B.
- Amendment requests to `SPEC-UPDATE-YAML-PRESERVE-001` recorded in `spec.md` §I; evidence for
  `SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001` in `spec.md` §J; the declared reversal of
  `SPEC-V3R6-UPDATE-NOISE-001` `REQ-UN-007` in `spec.md` §K.
- Artifacts left uncommitted per the plan-phase brief; committed at the plan-audit revision.

### Epic run order (dependency sequencing)

`depends_on: [SPEC-UPDATE-DATA-SURVIVAL-001]` is a **real** dependency, and it is narrower than the
frontmatter alone suggests. E2 owns *that the bytes reach disk before a destructive step*; this SPEC
owns *how the write itself behaves*. The single edge is M4: the shared atomic, mode-preserving write
helper introduced here is the mechanism E2's writes consume, so if E2 lands one first, M4 becomes a
consumer rather than an author (`plan.md` D-4 and §C step 4).

The run-phase `Depends_on Pre-flight Check` treats a dependency as fulfilled only at
`status: completed`. Every SPEC in this Epic is currently `draft`, so entering `/moai run` on this
SPEC before E2 closes raises the 3-option wait / override / abort blocker. **The dependency is
satisfied by sequencing, not by an `--ignore-deps` bypass** — the run order below is the mechanism,
and it is consistent with the order recorded in `SPEC-UPDATE-DATA-SURVIVAL-001`'s own `progress.md`
§E.1:

| Order | SPEC | Gate to clear before the next entry |
|---|---|---|
| 1 | `SPEC-UPDATE-REINSTALL-LOOP-002` (E1) | reaches `status: completed` — REQ-RIL2-015/016 landed |
| 2 | `SPEC-UPDATE-DATA-SURVIVAL-001` (E2) | reaches `status: completed` — backup coverage + failure contract landed |
| 3 | **`SPEC-CONFIG-TIER-PERSIST-001`** (this SPEC, E3) | — |
| 4+ | remaining Epic SPECs (`SPEC-UPDATE-YAML-PRESERVE-001`, `SPEC-CONFIG-KEY-HONESTY-001`, `SPEC-UPDATE-DOC-DRIFT-001`) | no `depends_on` edge to this SPEC |

Do **not** invoke `/moai run` on this SPEC with `--ignore-deps`. If E2 slips and starting E3 early
becomes necessary, the correct move is to run **M1, M2, M3, M5, and M6** — none carries an E2 edge —
and hold **M4** open until E2 closes, since M4 is the sole milestone whose write helper may need to
consume E2's. That is a scope decision for the orchestrator to surface via `AskUserQuestion`, not a
flag the run-phase agent may set on its own.

One further ordering constraint is internal to this SPEC and independent of the Epic order: M1 must
land before M2 (REQ-CTP-012), or the local tier becomes an opt-in with no opt-out (R4).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

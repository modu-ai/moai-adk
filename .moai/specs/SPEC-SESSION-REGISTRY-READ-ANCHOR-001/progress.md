# SPEC-SESSION-REGISTRY-READ-ANCHOR-001 — progress

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1058 · worktree `.claude/worktrees/t1058` · branch `WT-read-anchor`
- Plan-phase artifacts authored: `spec.md`, `plan.md`, `acceptance.md` (Tier M)
- Evidence base: `.moai/reports/t1058/pre-plan-measurement.md` (tree `3dfae918a`,
  measured 2026-09-21) with raw companions `orphan-file-list.txt`,
  `orphan-file-entry-counts.txt`, `orphan-entry-liveness.txt`,
  `live-pid-identity.txt`. No measurement re-run; no figure introduced that is
  absent from that record.
- Status: `draft`.
- **The card's premise was reversed by measurement and the SPEC is written to the
  measured state.** Anchoring R1 (`LiveAnchoredSessions`) does not neutralise the
  orphan registry files — it removes them from the disposal guard's view, and live
  orphan-only entries go with them. Anchoring R2 (`findRegistryUpwardFrom`) is a
  pure repair. The two are scoped apart (spec.md §C).
- Open at the Implementation Kickoff Approval gate: the S2 migration decision
  (plan.md M1, branches B1/B2/B3) and R2's evidence grade (plan.md M2). Neither is
  selected here.
- Evidence grades carried verbatim and NOT upgraded: Windows `stateanchor`
  behaviour is unmeasured (code reading only); R2's stall is synthetic-fixture
  grade, unreproduced against the real population; a process-existence probe
  identifies a process, not a session.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

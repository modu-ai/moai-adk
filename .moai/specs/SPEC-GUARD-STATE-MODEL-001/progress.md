# SPEC-GUARD-STATE-MODEL-001 — Progress (card t347)

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifacts authored on `091966c55` @ `WT-guard-liveness` (worktree `.claude/worktrees/t333`).

- Artifacts: `spec.md`, `plan.md`, `acceptance.md` (Tier M set) + this file.
- Requirements: 12 (Tier M ceiling 16). Acceptance criteria: 15 (ceiling 16).
- Card: **t347** (issued by the lead; the sibling surfacing SPEC is card t333). Dispatch is a separate lead decision — this lane authored the plan-phase artifacts only.
- Every RED-now cell is pinned to `091966c55` and its command was run on this tree during authoring; no cell was carried across the scope reduction from the predecessor without re-measurement.

### Origin

Created by the **scope reduction of `SPEC-GUARD-LIVENESS-001`** after its plan-audit iter-3 returned FAIL 0.667 with a STOP signal (0.800 → 0.800 → 0.667). The operator chose scope reduction — the audit's own recommendation — over a fourth repair round, which the regression clause forbids without an override.

The split ran along the seam the defects kept landing on: the surfacing model converged (both prior mutants re-run at iter-3, neither revivable), the state model did not (D2, D5, N2, T2, T4 are one family). This SPEC receives the state model.

### Inherited findings, carried as starting material

| Finding | Status |
|---|---|
| **T2** — a failed forge query has no admissible classification | Resolved in the state table, row 2 → `UNRESOLVED`; verified by AC-GSM-009 |
| **N2's unresolved half** — the fifth value repaired the no-reader hole and left the more common one open | Resolved structurally: totality is now demonstrated by construction over a 7-row table (AC-GSM-007 + AC-GSM-008) rather than asserted in prose |
| **T4** — the plan's vocabulary and milestone map went stale under a repair | Mitigated: the table is the single artifact both layers cite; criteria that genuinely split across milestones are clause-split in the flip lists |
| **T1** — the disk side of the set comparison had no integrity guard | Resolved: REQ-GSM-010 extends the all-clear refusal to a zero-file enumeration; verified by AC-GSM-012 |
| **T8** — two carried counts were consumed by nothing | Resolved: REQ-GSM-009 clause (b) requires every count to have a consumer; verified by AC-GSM-013 |

### Authoring method

Authored **as a state table, not as prose** — the auditor's stated reason the split helps. Prose is what let a five-value vocabulary cover at least six states without anyone noticing: a sentence defining a classification reads complete in isolation, while a table makes an uncovered condition visible as an empty cell. The table (`spec.md` §C.2 REQ-GSM-006) is normative; the prose describes it and does not extend it.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

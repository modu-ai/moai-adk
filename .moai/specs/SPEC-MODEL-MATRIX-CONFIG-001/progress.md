# Progress — SPEC-MODEL-MATRIX-CONFIG-001

Card: t1036 · Tier L · split from `SPEC-MODEL-PROFILE-MATRIX-002` on the M2 + M3 seam, 2026-09-20.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | L — REQ 22 / ceiling 25; AC 20 / ceiling 25 |
| Tier basis | REQ 22 and AC 20 both exceed the Tier M ceiling of 16; file scope >15 across `internal/config`, `internal/cli`, `internal/web`, `internal/settings/agentfm`, two config mirror pairs, and their tests |
| Artifact set | spec.md · plan.md · acceptance.md · design.md · research.md · progress.md |
| Requirements | 22 (REQ-MPME-001 … REQ-MPME-022) |
| Acceptance criteria | 20 (AC-MPME-001 … AC-MPME-020) |
| Predecessor | `SPEC-MODEL-MATRIX-CORE-001` |
| Successor | `SPEC-MODEL-MATRIX-DOCS-001` |
| Status transition | (none) → draft |

Open question carried to the Implementation Kickoff Approval gate: whether REQ-MPME-007's non-representable warning branch can ever fire, or whether migration should be unconditional (`research.md` §F.1).

One clarification owned by the predecessor gates M3: the `Group`-field disposition in `moai model profile --json`. The Workflow-path contract cannot ship against an undecided shape.

Unverified inputs added by the split and assigned here: whether `Patch`'s current signature can express "leave `model:` unchanged", and whether the three non-`--profile` seams sit after a template deploy.

## §E.2 Run-phase Evidence

_<pending run-phase — nothing in this SPEC has landed. `model_routing_profiles` is still present in `workflow.yaml`; the effort application does not exist. Blocked on `SPEC-MODEL-MATRIX-CORE-001` landing its M1 step 7 template-baseline re-set.>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

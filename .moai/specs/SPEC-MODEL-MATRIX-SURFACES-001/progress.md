# Progress — SPEC-MODEL-MATRIX-SURFACES-001

Card: t1036 · Tier M · split from `SPEC-MODEL-PROFILE-MATRIX-002` on the M4 + M5 seam, 2026-09-20.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed) |
| SPEC ID collision | none |
| Tier | M — REQ 11 / ceiling 16; AC 10 / ceiling 16 |
| Tier basis | REQ 11 and AC 10 both within the Tier M ceilings; file scope ~10 files (wizard question + translations, web model option set, v4manifest tier table, four locale files, tests) inside the Tier M 5-15 band |
| Artifact set | spec.md · plan.md · acceptance.md · progress.md — the Tier M three plus progress. Design rationale and current-state evidence are carried in `plan.md` §B and §C rather than in separate `design.md` / `research.md` artifacts |
| Requirements | 11 (REQ-MPMS-001 … REQ-MPMS-011) |
| Acceptance criteria | 10 (AC-MPMS-001 … AC-MPMS-010) |
| Predecessor | none — root of the chain, no `depends_on` |
| Successor | `SPEC-MODEL-MATRIX-DOCS-001` (which adds guard coverage for the surfaces M5 clears) |
| Status transition | (none) → draft |

**Parallelism**: both milestones were declared independent of M1-M3 in the predecessor's plan, so this SPEC may proceed concurrently with every other successor. It carries no `depends_on` edge.

Open question carried to the Implementation Kickoff Approval gate: the wizard option-value vocabulary — change the values to `max`/`medium`/`low` and keep `NormalizeToTier` as a tolerant reader, or keep `high`/`medium`/`low` for backward compatibility (`plan.md` §B.1).

Unverified input inherited and assigned here: the `mp.*` i18n key family's zero-consumer status, carried forward from the original brief and never re-grepped. AC-MPMS-009 requires the measurement to be taken **before** removal and recorded here.

**M5 expansion anchor**: card **t1031**'s M5 expansion entry lands in this SPEC. M5 is `SPEC-MODEL-MATRIX-SURFACES-001`, and this is the fixed address t1031 was waiting on.

## §E.2 Run-phase Evidence

_<pending run-phase — nothing in this SPEC has landed. M4 and M5 were never started under the original SPEC ID. The already-landed `profile_setup_translations.go` label correction in squash `31da99a7b` is NOT an M4 deliverable — it corrected existing option text; M4 rewrites the question.>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

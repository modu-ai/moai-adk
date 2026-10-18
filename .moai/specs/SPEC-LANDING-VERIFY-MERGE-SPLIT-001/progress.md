# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T17:21:00Z
iteration: version 0.8.0 (board 6404 rulings on Q27 and Q28) on top of the split per board d20e after plan-audit iteration 5 (FAIL 0.86 at 9713b6757d stays on record); fresh plan-audit cycle on the reduced scope, REQ/AC ids of this piece unchanged
scope: piece A2 only — M0 baseline, M1 receipt commit field, M2 audit binding in complete, M3 gate-state reader and key and status display, M4 lane guidance; the enabling switch is a separate later commit on the integration branch made only after A2 and A1 have both landed, not a milestone (boards 9dbe, 0577, 470c, 5d4d, d20e)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L by file count; 7 REQ / 7 AC)
plan_artifact_hash: ee585a8bf32bda941185f5b3bc89323836177bf335e1c61a6f703b58c7c739bf (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none; Q27 DECIDED by board 6404 (key on only after A2 and A1 have both landed); Q28 DEFAULT-APPLIED with the 6404 ruling (merge-path binding is A1's); Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by 0577 P4; Q18 by 470c; Q20 and Q22 by 5d4d; Q29 by d20e; Q28 DEFAULT-APPLIED (implementation-level, rule: smaller user-visible surface); Q15-Q17, Q19, Q21, Q23-Q26 recorded, Q15-Q17, Q25, Q26 move with piece A1
record_cleanup: all bookkeeping is committed in the same commit as the final artifacts; no backfill commit is planned after the audited SHA (board 0577 P3)

Split disposition (iteration 5 at 9713b6757d, board d20e): REQ-LVM-001-006, 011 (manifest half), 013-016 and AC-LVM-001-006, 011, 013-016 moved to SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875, blocked until this card is merged-local); blockers B1 (effective environment under the command's own prefix), B2 (GOFLAGS allowlist), B3 (producer-verb refusal tests) and the optional S1, O1 are inherited acceptance requirements for that SPEC (spec.md §7); O2 (overlap-baseline script defects) stays recorded as run-phase debt (research.md §11). A2 text was carried from the audited version with four adjustments where an A1 concept appeared (REQ-LVM-007, REQ-LVM-012 with the audit-side guidance of the old REQ-LVM-011, REQ-LVM-017, and a placement clause in REQ-LVM-008); independence by code in research.md §12. The direct merge verb is not bound by the audit binding in A2 (design R6): accepted residual because the key stays off until A1 lands, and A1 inherits the merge-path binding as an acceptance requirement (spec §7 row 12; board 6404). Version 0.8.0 records the 6404 ruling; ids and counts unchanged (7 REQ / 7 AC).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

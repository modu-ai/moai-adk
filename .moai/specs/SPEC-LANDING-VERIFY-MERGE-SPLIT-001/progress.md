# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T16:45:00Z
iteration: 3 (revision after plan-audit iteration 2, FAIL 0.80 at fc3746c10a; last Tier L round)
scope: piece A only (M0 baseline, M1 gate core, M2 receipt commit field, M3 audit binding, M4 lane guidance); the enabling switch is a separate later commit on the integration branch, not a milestone (boards 9dbe, 0577, 470c)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L; 18 REQ / 18 AC)
plan_artifact_hash: 67c06d55b8dcf2fd66b09667754d05a2de1d9e77a373dd25be7b39775fd7416c (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none (Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by 0577 P4; Q18 by 470c; Q15-Q17 and Q19-Q22 DEFAULT-APPLIED, implementation-level)
record_cleanup: all bookkeeping is committed in the same commit as the final artifacts; no backfill commit is planned after the audited SHA (board 0577 P3)

Iteration-2 defect disposition: N1 fixed (REQ-LVM-016, class 13 by shared re-derivation, AC-016, mutant row); N2 fixed per board 470c (REQ-LVM-017 committed blob of the integration branch, fail-modes unreadable=ON with notice, AC-017 observes enable and revert with a detached stand-in primary as negative control, `manifest verify` prints the gate state and source); N3 fixed (REQ-LVM-008 locator, REQ-LVM-018 `audit_receipt:` grammar and parser with refusal for missing/duplicate/empty/none/malformed/absent-id, AC-008, AC-018, protocol text in the touch list); N4 fixed (one equivalent-HEAD predicate for verdict and receipt, AC-008/010 tolerated and refused cases with the real writer); O1 fixed (tree_root in REQ-010, AC-010 h); O2 fixed (entry points mcp_codex.go x2 and mcp_convergence.go x1 in REQ-009, touch list and Tier arithmetic, AC-009 TestEntryPointCapturesStartHead); O3 fixed (REQ-012 and AC-012 say the key-off card lands through the gated path with the configuration file's reader calls); O4 fixed (at least one qualifying receipt; failed extras recorded, never replace the pass; Q21); O5 fixed (reader search and go list proof recomputed by the shared derivation; build-tag activation stays residual R1); O6 partly (REQ-005 and REQ-001 keep their enumerations because each ACs splits them; REQ-016/017/018 are separate); O7 recorded as run-phase debt with the reason in research.md section 11.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

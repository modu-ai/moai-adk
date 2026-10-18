# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T16:10:00Z
iteration: 2 (revision after plan-audit iteration 1, FAIL 0.73 at 8a3014bee9)
scope: piece A only (M0 baseline, M1 gate core, M2 receipt commit field, M3 audit binding, M4 lane guidance); the enabling switch is a separate later commit on main, not a milestone (boards 9dbe, 0577)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L; 15 REQ / 15 AC)
plan_artifact_hash: 4588688b1a0d8e751f34eda53db0548bdfdd3b64827880fea74190265ba207ce (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none (Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by board 0577 P4; Q15-Q17 DEFAULT-APPLIED, implementation-level)

Iteration-1 defect disposition: D1 fixed (REQ-LVM-009/010, M2/M3, receipt writer and schema in the touch list and Tier arithmetic); D2 fixed (reject class 12, AC-005 N=12, mutant row); D3 fixed (REQ-LVM-013, AC-013); D4 fixed (REQ-LVM-014, AC-014, AC-011 greps name the commands); D5 fixed (binary prerequisite C5, config read from the integration root, switch outside the branch, DoD no longer claims first live use, transition picture research.md §10, no-backfill rule P3); D6 fixed (recovery path REQ-LVM-012, design §6); D7 fixed (AC-012 diffs S0^ and <S0>^..<S4> over workflow.yaml); D8 fixed (one grep per procedure step); D9 fixed (embed owners and deleted packages, AC-001 N=6); D10 fixed (per-package test counts, class 11); D11 fixed (REQs are list items); D12 partly (REQ-007 and REQ-015 reworded, bundling kept where one invariant); D13 fixed (AC-008 negatives, AC-007 oracle); D14 fixed (design §3 residuals R1-R3). CN-4: plan.md milestones are `###` headings with `Exit:` lines.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

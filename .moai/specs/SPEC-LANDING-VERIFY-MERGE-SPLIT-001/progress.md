# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T17:25:00Z
iteration: delta round after plan-audit iteration 3 (FAIL 0.84 at 90a14c5a1a; rule C, board 5d4d: one correction, one delta audit, REQ/AC ids unchanged)
scope: piece A only (M0 baseline, M1 gate core, M2 receipt commit field, M3 audit binding, M4 lane guidance); the enabling switch is a separate later commit on the integration branch, not a milestone (boards 9dbe, 0577, 470c, 5d4d)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L; 18 REQ / 18 AC)
plan_artifact_hash: 4e1aacfe42d08bdd3865a97bc247cd5f09a59ca986dae7e646eabb39dd1f9bdb (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none (Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by 0577 P4; Q18 by 470c; Q20 and Q22 by 5d4d; Q15-Q17, Q19, Q21, Q23, Q24 DEFAULT-APPLIED, implementation-level)
record_cleanup: all bookkeeping is committed in the same commit as the final artifacts; no backfill commit is planned after the audited SHA (board 0577 P3)

Iteration-3 defect disposition: D1 fixed ("equivalent to HEAD" removed everywhere; verdict and receipt must equal HEAD; AC-008 empty-diff-ancestor and evidence-directory negatives, AC-010 empty-diff-ancestor receipt negative); D2 fixed (REQ-010 looks the receipt up at StoreRoot(audited tree) and compares tree_root with the audited tree's own path; config-orphaned-worktree positive and same-commit-primary negative, AC-010 N=11); D3 fixed per board 5d4d (REQ-017: none configured reads off, configured but unresolvable refuses with a named error, absent path discriminated from failed read by an existence check, no-deadlock reasoning, fixture runs from a card worktree with CLAUDE_PROJECT_DIR on the stand-in, AC-017 N=11); D4 fixed (SPEC-less cards bound through verdict.md, REQ-008/018, TestSPECLessCardVerdictLocatedAtVerdictMd, TestSPECLessCardCitationLineRequired); D5 fixed (REQ-016 re-derivation under the manifest's recorded environment, TestReDerivationUsesManifestEnvironmentNotVerifierEnv); D6 recorded-not-gated (audit_target stored in new receipts, R5 named, Q23); D7 fixed (pointer to research.md section 11); D8 fixed (cli/integration.go registration in the touch list, Tier arithmetic 41 files); D9 fixed (AC-016 states attempts stay bound to tree and HEAD).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

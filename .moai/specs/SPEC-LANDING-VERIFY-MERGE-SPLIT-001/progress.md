# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T17:55:00Z
iteration: single ceiling exception after plan-audit iteration 4 (delta FAIL 0.88 at 21687134ba stays on record; board 4812), REQ/AC ids unchanged
scope: piece A only (M0 baseline, M1 gate core, M2 receipt commit field, M3 audit binding, M4 lane guidance); the enabling switch is a separate later commit on the integration branch, not a milestone (boards 9dbe, 0577, 470c, 5d4d, 4812)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L; 18 REQ / 18 AC)
plan_artifact_hash: 3defa7855c4c84e1419210d95185e82d71c3fa91bc67b053176cea17768571e4 (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none (Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by 0577 P4; Q18 by 470c; Q20 and Q22 by 5d4d; Q15-Q17, Q19, Q21, Q23-Q26 DEFAULT-APPLIED, implementation-level)
record_cleanup: all bookkeeping is committed in the same commit as the final artifacts; no backfill commit is planned after the audited SHA (board 0577 P3)

Iteration-4 disposition: B1 fixed (the effective GOFLAGS, resolved by go env, is recorded at derive and at attempt and may carry no test-selection, truncation or shaping flag; manifest derive, remeasure --call and the verifier refuse it even when both sides match; AC-005 N=15 with TestGateRejectsNarrowingGoflagsRun, TestGateRejectsNarrowingGoflagsSkip and the table test with a build-only positive control; mutant row); S1 fixed (go.mod, go.sum, go.work and go.work.sum select every Go package of the module, bypass the reader search and the 35e9 narrowing, and are never a zero-call card; TestManifestModuleFileChangeSelectsWholeModule, TestZeroGoPathRefusesModuleFileChange; AC-001 N=7, AC-013 N=5); O1 fixed (auditreceipt.Canonical named in REQ-010 and design); O2 fixed (mutating commands exit non-zero before any state change, read-only evaluations print the named error and exit non-zero; REQ-007 holds because today's merge verb already stops at integrationMergeWorktree); O3 fixed (mutant row for the receipt-side empty-diff ancestor).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

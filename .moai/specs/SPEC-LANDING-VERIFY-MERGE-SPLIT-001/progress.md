# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — progress

Card t1872. Status of record: `spec.md` frontmatter (`draft`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-18T17:43:00Z
iteration: version 0.9.0, revision after plan-audit A2 round 1 (FAIL 0.81 at edee2983cc, .moai/reports/t1872/plan-audit-a2.md) and board 9b0e; REQ/AC ids and counts unchanged (7 REQ / 7 AC), 55 named tests
scope: piece A2 only — M0 baseline, M1 receipt commit field, M2 audit binding in complete, M3 gate-state reader and key and status display, M4 lane guidance; the enabling switch is a separate later commit on the integration branch made only after A2 and A1 have both landed, not a milestone (boards 9dbe, 0577, 470c, 5d4d, d20e)
plan_artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md (Tier L by file count; 7 REQ / 7 AC)
plan_artifact_hash: 6292e80cda0f0285e4cd0d5dbb4eb54f11e2a47021b7186e1723f00033e2f6c4 (ComputeHash-equivalent over planArtifactNames; Python reimplementation of internal/runtime/audit_cache.go:121-141, not the Go function)
open_decisions: none; Q30 (REQ id gap waiver) DECIDED by board 9b0e; Q27 DECIDED by board 6404 (key on only after A2 and A1 have both landed); Q28 DEFAULT-APPLIED with the 6404 ruling (merge-path binding is A1's); Q7-Q13 by boards 9dbe/2c51/e4cc; Q14 by 0577 P4; Q18 by 470c; Q20 and Q22 by 5d4d; Q29 by d20e; Q28 DEFAULT-APPLIED (implementation-level, rule: smaller user-visible surface); Q15-Q17, Q19, Q21, Q23-Q26 recorded, Q15-Q17, Q25, Q26 move with piece A1
record_cleanup: all bookkeeping is committed in the same commit as the final artifacts; no backfill commit is planned after the audited SHA (board 0577 P3)

Split disposition (iteration 5 at 9713b6757d, board d20e): REQ-LVM-001-006, 011 (manifest half), 013-016 and AC-LVM-001-006, 011, 013-016 moved to SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875, blocked until this card is merged-local); blockers B1 (effective environment under the command's own prefix), B2 (GOFLAGS allowlist), B3 (producer-verb refusal tests) and the optional S1, O1 are inherited acceptance requirements for that SPEC (spec.md §7); O2 (overlap-baseline script defects) stays recorded as run-phase debt (research.md §11). A2 text was carried from the audited version with four adjustments where an A1 concept appeared (REQ-LVM-007, REQ-LVM-012 with the audit-side guidance of the old REQ-LVM-011, REQ-LVM-017, and a placement clause in REQ-LVM-008); independence by code in research.md §12. The direct merge verb is not bound by the audit binding in A2 (design R6): accepted residual because the key stays off until A1 lands, and A1 inherits the merge-path binding as an acceptance requirement (spec §7 row 12; board 6404). Version 0.8.0 records the 6404 ruling; ids and counts unchanged (7 REQ / 7 AC).

Plan-audit A2 round 1 disposition (version 0.9.0): MP-1 (id gap) waived by board 9b0e, explicit clause in spec.md §3 and row Q30, ids not renumbered; MP-9 (D1) fixed twice: the characterization test lands in the M0 commit, ahead of the commit that edits factoryCompleteCard (AC-LVM-007 reworded, AC-LVM-012 graph legs by ancestry and by path), and the ledger paragraph that collided with the ordering verb is replaced by per-cell entries; the ordering verb now prints 0 CONFLICT lines; D2 tag shadow fixed (REQ-LVM-017 resolves refs/heads/<branch>^{commit} once, TestSameNamedTagDoesNotShadowBranch, mutant row, probe in research.md section 13); D3 M4 fixed (make agents-emit, template .codex TOML regenerated, local copy refreshed, hand-edited versus generated files listed, about 27 files, AC-LVM-018 legs); D5 audit-row sink named (.moai/logs/landing-gate-audit.jsonl); D6 three inherited obligations added to the hand-off; D7 audited tree and abbreviated audited_sha defined; D8 spec C11; D9 non-Go cells are ledger entries N-001 to N-023; D10 HISTORY order, version labels and REQ-017/018 labels fixed; D11 Q28 and Q2-Q6 times set to the setting commits; D12 AC-LVM-008 window wording; D13 design dependency sentence. Run-phase debt recorded in plan.md section F (embed-manifest-check and catalog hashes after template edits, unquoted YAML yes, overlap-baseline script).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

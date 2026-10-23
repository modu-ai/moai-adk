auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

verdict: FAIL
audited_sha: c83266d78c2e37514904d03d8ee45e058ef7fe8c
overall_score: 0.72
must_pass_failed: 0
blocking_count: 4
plan_artifact_hash: 1c22660a4752544c1da0d7c8c101a1a461a80433f951e1c318087543d9407964
scope: full
fix_scope: spec.md#B.1 spec.md#A.4 spec.md#E.3 plan.md#B.3 plan.md#B.2 plan.md#D acceptance.md#C acceptance.md#E acceptance.md#F decision-index.md#Routing
defect_class: D1=design D2=design D3=ac-wording D4=traceability
convergence_overall: fail
required_backend: codex fail

Iteration: 1/3 | Tier M (threshold 0.80) | Reasoning context ignored per M1 Context Isolation.
Plan Artifact Hash: computed with the planArtifactNames algorithm of internal/runtime/audit_cache.go (acceptance, decision-index, plan, spec).
Auditor Version: plan-auditor/v1 (single-model Claude review plus audit_multi codex leg)

## Must-Pass Results
- [PASS] MP-1: REQ-ISB-001..015, sequential, no gap or duplicate (spec.md:L66-92; 15 of ceiling 16).
- [PASS] MP-2 (judged on the REQ layer, spec.md): all 15 are GEARS (When/Where/While/ubiquitous). REQ-ISB-015 (spec.md:L92) has a non-actor subject ("the set of files") but is a valid While form.
- [PASS] MP-3: frontmatter carries all 12 canonical fields plus tier (spec.md:L2-18); `moai spec lint --json --strict` printed only an Info OwnershipTransitionUnmeasured (exit 0).
- [N/A] MP-4: single-language Go SPEC.
- [PASS] MP-5: referenced SPECs CANDIDATE-CI (implemented), LOCAL-MAIN-FLOW and MERGE-WINDOW-QUEUE (completed) are not retired. LANDING-SURFACE-MECH/RULES and LANE-SELF-RECOVERY are not in this tree (SHOULD: referenced SPEC not found; the plan discloses this as G3-G5).
- [N/A] MP-6: grep -c syscall printed 0.
- [PASS] MP-7: grep for NEEDS CLARIFICATION in plan.md printed nothing (exit 1); research.md absent.
- [N/A] MP-8: no AC is classified release-blocking (acceptance.md:L13-29 uses Ordered/Static). See D3: the Definition of Done (acceptance.md:L234) makes all 15 gating, so the classification is ambiguous. Ledger greps E-01..E-08 were re-run and agree with the cited stdout at HEAD c83266d78 (internal/rules files unchanged since the pin).
- [PASS] MP-9: CN-4 verb was not mechanically run. Read by hand: M1->M5 order in plan.md:L104-108 matches the "Flips at" column. No ordering conflict found.

## Category Scores
| Dimension | Score | Evidence |
|---|---|---|
| Clarity | 0.70 | K=0 "exactly" equivalence claim is false (D2); not-applicable output policy contradicts REQ-009 (D4) |
| Completeness | 0.90 | All sections, Out of Scope H3s (spec.md:L98-117) present |
| Testability | 0.60 | 11 of 15 ACs have only a vacuous RED-now (E-00 `[no tests to run]`, exit 0); AC-015 has no ledger entry (D3) |
| Traceability | 0.85 | REQ<->AC<->milestone mapped (acceptance.md:L31); REQ-002 design collides with adoption (D1) |

## Code-cite verification (spec/plan vs tree)
Held: remeasure record fields :35-57, ReadRemeasureRecord :86, ValidateRemeasureRecord :135, gitMergeBase :612, RunRemeasure :634, headTreeSHA :722; merge cause 1 :321-327, cause 2 :332-335, MergeExitBaseMoved=2 :42; acquire preconditions :444/:457/:465/:476/:517; complete `if !heldByUs` :2186, candidate hold :2194, acquisition :2201; remeasureWorktreeDir; config cites; configCacheSchemaVersion=12; existing merge-step tests G-10b exist; `TestShippedConfigKeysHaveReaders` exists at shipped_key_reader_test.go:74. Exit 16: grep for exitCodeError/os.Exit/const literals 16-19 found none (G6 enumeration limit stands). Card D file exists untracked in t1715 tree with REQ-LSR-010 matching the boundary row; t1846 out-of-scope matches; no t1847 tree exists. Boundary vs t1715/t1846/t1847 is clean: no duplicated mechanism, gate reuses Read/ValidateRemeasureRecord.

## Defects Found
D1. adoption-block — plan.md:L74 (B.3), spec.md:L67 (REQ-ISB-002) — The complete gate runs before the acquisition, but factory_card.go:2229-2240 (REQ-MWQ-019 step 2 adoption) runs after it and adopts a card whose merge commit is already on the integration tip (the lane merged through the merge verb first). For such a card landings-behind L>=1 (the merge commit is a first-parent landing not reachable from the card tip), so K=0 refuses a legitimate adoption. Codex reproduced `merged-local` completion today and a refusal with the gate. — Severity: critical — Class: blocking — Required fix: exempt the adopt case (card tip is the third parent of the integration tip and the tip tree has a valid record) in REQ-ISB-002/B.3, and add an AC with a no-window adoption regression test.
D2. false-equivalence — spec.md:L136 (E.3), plan.md:L30, L116 (R3), acceptance.md:L227 — The claim that at K=0 the gate refuses "exactly" what cause 2 refuses is false: cause 2 tests the stored record.Base != tip (integration_merge_step.go:332), the gate tests the current merge-base. A no-content-change absorb leaves L=0 and a valid record for the same tree while cause 2 exits 2. Acceptance §F and R3 call this case "correct". Codex reproduced it. — Severity: major — Class: blocking — Required fix: either require record.Base == tip when K=0 (and bound K>0 relative to record.Base) or withdraw "exactly", correct §F/R3, and add a boundary AC.
D3. red-now-vacuous — acceptance.md:L8, L175-180, L13-29 — AC-001..005, 008..011 and the Ordered part of 007/013/014 take their RED from "M1 will commit a failing run"; the only observed cell, E-00, is a zero-test sweep (exit 0) and E-07 is a positive control. verification-completeness.md §2 requires a RED observed on the pre-implementation tree. AC-015 has no ledger entry at all, and no AC is classified release-blocking vs regression-guard even though the DoD gates all 15. — Severity: major — Class: blocking — Required fix: give each criterion a non-vacuous RED cell (e.g. grep -c of the planned symbol/test name, exit 1, with command, stdout, exit, SHA) and classify each AC release-blocking or regression-guard.
D4. fail-open-output-conflict — plan.md:L63 (B.2 step 1) vs spec.md:L83 (REQ-ISB-009), acceptance.md:L97-99 — The design prints nothing for not-applicable (no card id, unresolved branch, detached HEAD, integration branch) while REQ-009 and AC-009 require one line naming the cause in the same cases. G-09 runs only the pure factory package, so CLI output and window recording are not proven. — Severity: major — Class: blocking — Required fix: one policy (print the cause line, or carve not-applicable out of REQ-009) and a CLI-level test command in AC-009.
D5. q3-routing — decision-index.md:L62 — Routing says Q3 blocks the autonomous Kickoff, yet spec.md:L110-111 puts promotion gating out of scope; the SPEC is self-consistent whichever way Q3 is answered (an answer widening scope would be a new SPEC). Q4 is likewise a scheduling matter, not a SPEC gap. — Severity: minor — Class: optional — Required fix: relabel Q3 as a deferred scope question that does not gate Kickoff.
D6. ac-gaps — acceptance.md:L59-72, L163 — AC-004/005 do not state the gate is enabled; G-14 greps `homestate` only though AC-014 names `card_evidence`. — Severity: minor — Class: optional — Required fix: add `Given enabled` and extend the grep.

## Regression Check
Iteration 1: none.

## Recommendation
FAIL. Fix D1-D4 and re-audit (delta scope, CN-4 verb in full). Codex required leg returned fail (receipt rcpt-6c0a3919e80740991a41ff4b); its four findings correspond to D1-D4 and D1/D2/D4 were re-verified against the cited code.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)
- Claim: cites hold; four design/AC defects. Evidence: commands above (sed/grep of cited ranges, `moai spec lint --json --strict` exit 0 with the single Info finding, `moai verify audit-plan` result-file check `convergence_check: {ok: true, unmet: []}`, audit_multi codex verdict fail). Baseline: HEAD c83266d78, installed moai build ee981a932 (a strict ancestor of HEAD, per build_lag note; lint is not from a build of this tree).
- Gaps: CN-4 and D7/D8 verbs not run as scripts (read by hand); codex was the only gated backend (claude and glm gates are off); audit_multi findings for D1/D2 are codex reproductions, D1 was re-read in code but not re-run; the new gate does not exist, so none of the planned tests were run; `go test` suites not run.
- Residual-risk: exit-16 collision enumeration is partial (G6); first-parent landing counting measured in one tree only (G2).

## Operational Notes (unverified)
- Measure the adoption path: run `moai factory complete` on a card already merged by `moai integration merge` with the gate prototype enabled (assumption: lane flow in factory_card.go:2229).

AUDIT-VERDICT: FAIL spec=SPEC-INTEGRATION-STALE-BASE-001 receipts=rcpt-6c0a3919e80740991a41ff4b

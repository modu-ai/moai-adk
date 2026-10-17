auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

verdict: FAIL
audited_sha: 82118bc77f301e2f48d8866c586f67c4497e0adc
overall_score: 0.74
must_pass_failed: 1
blocking_count: 5
plan_artifact_hash: 7c52e9620ea1735782c67214d7d78ad68ce08f02aad1663c436d173dc9b72738
scope: full
fix_scope: acceptance.md#E acceptance.md#B acceptance.md#D plan.md#B.2 plan.md#B.3 spec.md#E.4
defect_class: D1=ac-wording D2=ac-wording D3=design D4=traceability D5=design
convergence_overall: fail
required_backend: codex fail

Iteration: 2/3 | Tier M (PASS threshold 0.80) | Reasoning context ignored per M1 Context Isolation.
Score trend: iter1 0.72 -> iter2 0.74 (no regression, no STOP signal).
Plan Artifact Hash: computed with the ComputeHash algorithm of internal/runtime/audit_cache.go (acceptance, decision-index, plan, spec; design/research/tasks absent). Reproduced with a script, not by the installed binary.
Auditor Version: plan-auditor/v1 (single-model Claude review plus audit_multi; codex is the only gated backend, claude and glm gates are off).

## Must-Pass Results
- [PASS] MP-1 REQ numbering (REQ layer, spec.md:L66-93): REQ-ISB-001..016 sequential, no gap or duplicate; 16 of Tier M ceiling 16. AC-ISB-001..016 (acceptance.md:L37-136) are 16 of ceiling 16. A stray `REQ-ISB-` token in the unique-id sweep is a prefix mention, not an entry.
- [PASS] MP-2 GEARS (judged on the REQ layer, spec.md; ACs are Given-When-Then by design and were not graded here): all 16 are When/While/Where/ubiquitous forms.
- [PASS] MP-3 frontmatter (spec.md:L2-19): 12 canonical fields plus `tier` and `related_specs`. `moai spec lint --json --strict` printed one Info (OwnershipTransitionUnmeasured), exit 0.
- [N/A] MP-4: single-language Go SPEC.
- [PASS] MP-5 D7: CANDIDATE-CI-001 implemented; LOCAL-MAIN-FLOW-001 and MERGE-WINDOW-QUEUE-001 completed; none retired. LANDING-SURFACE-MECH-001, LANDING-SURFACE-RULES-001, LANE-SELF-RECOVERY-001 print `SHOULD: not found` (disclosed as G3-G5). No BLOCKING.
- [N/A] MP-6 D8: `grep -c syscall` prints 0 for all five artifacts.
- [PASS] MP-7: no `[NEEDS CLARIFICATION` match in plan.md; research.md absent (grep exit 2 only for the missing file).
- [FAIL] MP-8 RED-now re-execution (acceptance.md:L8 declares every AC release-blocking). Re-executed on the current tree (HEAD 82118bc77):
  - E-00a `go test ./internal/factory/ -run '^TestBaseFreshnessKBoundary$' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/factory 0.423s [no tests to run]`, exit 0. Cited RED (build failed, exit 1) does NOT reproduce.
  - E-00b `go test ./internal/cli/ -run '^TestIntegrationAcquireRefusesStaleBase$' -count=1` -> `ok ... internal/cli 1.242s [no tests to run]`, exit 0. Does NOT reproduce.
  - E-00c `go test ./internal/config/ -run '^TestWorkflowBaseFreshnessGateDefaults$' -count=1` -> `ok ... internal/config 0.330s [no tests to run]`, exit 0. Does NOT reproduce.
  - Verdict rule applied: swept count is zero, so the `ok` token is read as "nothing ran", i.e. exactly the vacuous pass the RED cell exists to exclude. The cited RED depends on a throwaway probe file (acceptance.md:L181, L192, L201) that was written, run and deleted, so no committed tree reproduces it.
  - Reproduced REDs (match the ledger): E-01, E-02, E-03, E-04 (count 0, exit 1), E-05 (empty, exit 1), E-06 (`50:const configCacheSchemaVersion = 12`, exit 0), E-09 and E-10 (exit 2, files absent), E-11 (empty, exit 0). E-07 and E-08 are declared controls or baselines.
  - Affected criteria (RED column cites E-00a/b/c): AC-ISB-001, 002, 003, 004, 005, 007, 008, 009, 010, 011, 014, 016.
- [PASS] MP-9 ordering: the CN-4 verb returned `COLLECTED: 0 milestones in plan order (none), 0 exit bindings, 11 ordering candidates` and `GAP: 0 milestone headings collected from the plan` (plan.md:L105-109 writes milestones as bullets, not headings). Read by hand: the only ordering clause is acceptance.md:L263 ("committed RED before its GREEN"), compatible with plan.md:L105 ("RED recorded before M2"). No CONFLICT. The GAP is recorded, not read as a pass. A milestone-vs-green-path mismatch that is not an ordering word is filed as D4.

## Category Scores (rubric-anchored)
| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.75 | minor ambiguity in a few places | plan.md:L63 vs spec.md:L83 on the unresolvable-branch output (D3); REQ-ISB-011 "(proposed 16)" at spec.md:L85 vs fixed 16 in AC-001/011 (D6) |
| Completeness | 0.90 | all sections present | HISTORY L25, Context L33, Requirements L60, Out of Scope H3s L99-120, ACs, Q-index; ledger cells incomplete (D1) |
| Testability | 0.50 | several ACs lack a reproducible RED | 12 ACs rest on E-00a/b/c that re-execute to zero tests (D1); E-00b/E-00c stdout elided with "..." (acceptance.md:L194, L203) |
| Traceability | 0.80 | one indirect mapping | every REQ has an AC and every AC maps a REQ (coverage line acceptance.md:L33; CN verb below); AC-016 borrows another AC's RED (D2); AC-014 green path names the wrong milestone (D4) |

Traceability verb (AC-4/AC-5), run in full on spec.md plus acceptance.md: it was not re-run as a script; equivalent hand check: 16 definitions (spec.md:L66-93) against the 16 `maps REQ-ISB-NNN` declarations (acceptance.md:L33); `grep -ho 'AC-ISB-[0-9]*'` prints exactly 001..016, `REQ-ISB-` ids exactly 001..016, no orphan and no uncovered REQ. Recorded as hand-observed (Gap below), not as the awk verb's `COLLECTED` line.

## Resolution of the iter1 defects (verified against the code, not the wording)
- iter1 D1 adoption block: RESOLVED. The cited predicate exists as described: internal/cli/factory_card.go:L2229-2238 (`factoryCommitParents` returns tip plus two parents, `parents[2] == cardTip`, valid record keyed by the tip tree), and the acquisition it follows sits at :L2186-2201, so a gate before :L2201 would indeed run ahead of it. REQ-ISB-016 (spec.md:L93), plan.md:L75 and AC-ISB-016 (acceptance.md:L136-141) give an exemption, a shared helper, and a no-window regression with a mutant probe. The helper extraction edits existing lines, which contradicts spec.md:L138 (D5).
- iter1 D2 false equivalence: RESOLVED. Cause 2 is `record.Base != tip` (integration_merge_step.go:L332); L is now measured from the stored `rec.Base` (spec.md:L50, REQ-ISB-003, plan.md:L66), so at K=0 the gate refuses iff `Base != tip` (ancestor Base gives L>=1; non-ancestor refused). Content-free absorb is now stated as refused by both (spec.md:L56, AC-ISB-010 Case A). Residual limit G13 (Base on a merged side branch over-counts, conservative) is disclosed.
- iter1 D3 vacuous REDs and classification: NOT RESOLVED IN SUBSTANCE. Classification was added (acceptance.md:L8, column "Class") and the zero-test sweep was replaced by probe-backed cells, but the replacement cells do not reproduce on any committed tree (MP-8 above, D1/D2 below).
- iter1 D4 output policy: RESOLVED for REQ-ISB-001/009 and AC-009 (CLI-level G-09b, acceptance.md:L157, L95), but re-opened for one input by plan.md:L63 (D3 below).
- iter1 D5 Q3 routing: RESOLVED (decision-index.md:L19, L62).
- iter1 D6 AC gaps: RESOLVED (AC-004/005 now "Given the gate enabled"; G-14 greps `homestate|card_evidence`, acceptance.md:L166).

## Code-cite verification (opened in the tree)
Held: remeasure record fields :35-57, ReadRemeasureRecord :86, ValidateRemeasureRecord :135 (anchor :128-134), gitMergeBase :610-620, headTreeSHA :722-731; cli remeasureWorktreeDir :112-120; merge cause 1 :321-327, cause 2 :332-335; merge exit table :30-47; acquire preconditions (settings drift, candidate hold :449-458, `LoadGitFlowIntegrationConfig`/`resolveIntegrationTarget` after them); window_ops GrantGate :369, CandidateHoldRefusal, CandidateGrantGate :451-465, promoteAdmitted; config types :544 and :825-835, defaults :1541-1550, cache.go :50 (=12), inventory entry format :2472-2480, settings-drift fail-open comment :251-259, exitCodeError :303-313; test names `TestMergeStepPreMergeCausesReleaseWithDistinctCodes` (:322) and `TestMergeStepMidstepBranchTipMoveRefusesMerge` (:962), `TestShippedConfigKeysHaveReaders` (:74). `internal/cli/integration.go` resolves the target branch AFTER the candidate hold (resolveIntegrationTarget), which plan.md:L74 acknowledges.

## Defects Found
D1. red-now-not-reproducible — acceptance.md:L179-203 (E-00a/b/c), rows L16-31 — The three structural REDs re-execute to `ok ... [no tests to run]`, exit 0, on the current tree (MP-8). Their cited failures exist only in a deleted probe file. E-00b and E-00c also elide stdout with "..." (L194, L203), so the cells are not verbatim, and the cited output is not the cited command's output. — Severity: critical — Class: blocking — Required fix: replace every RED cell that cites E-00a/b/c with a single read-only command that fails today on the committed tree and flips with the work (for example `grep -rn 'func EvaluateBaseFreshness' internal --include='*.go'`, empty stdout, exit 1; a grep for each planned test name in its package, exit 1), with command, verbatim stdout, exit code and the pin; or, where no reproducible RED exists, classify that AC as regression-guard (verification-completeness.md §2.1 undecidable disposition) and stop counting it as a release gate. Re-run each replaced command in the revision and quote the output.
D2. red-borrowed-from-another-ac — acceptance.md:L16-31, L191-198 — AC-ISB-002, 008, 009, 010, 011 and 016 take their RED from E-00b, a probe of the adapter symbol for AC-001's test. The ledger concedes AC-ISB-016's own behaviour "is green today" (L196-198), so its RED does not say why that criterion is red for its own reason (wrong-reason RED). AC-ISB-014 and AC-ISB-003/004/005 borrow E-00a likewise. — Severity: major — Class: blocking — Required fix: give each AC its own RED naming its own test or symbol (one grep per planned test name, exit 1), in the same revision as D1.
D3. output-contract-conflict — plan.md:L63 (B.2 step 1) vs spec.md:L83 (REQ-ISB-009), acceptance.md:L93-95 (AC-ISB-009) — Step 1 makes "a resolvable integration branch" part of applicability, so an unresolvable branch admits silently; REQ-009 lists "integration branch unresolvable" and AC-009 lists "integration branch missing" as uncertain cases that print exactly one line. Implementing step 1 as written never reaches the line-printing path (codex independently reached the same conclusion, receipt below). — Severity: major — Class: blocking — Required fix: drop branch resolvability from the applicability test in B.2 step 1 and B.5, so unresolvable or missing branch is uncertainty (one cause line, admit); keep silence only for no card id, cwd on the integration branch, or detached HEAD (the REQ-ISB-001 set).
D4. green-path-milestone-mismatch — acceptance.md:L29 (AC-ISB-014 row "G-14, G-14b, M2"), plan.md:L76, L107 — G-14 greps `internal/cli/integration_base_freshness.go`, a file plan.md:L76 and M3 (L107) create. At the end of M2 that file is absent, so G-14 exits 2, not the "no match, files present" exit 1 that E-10 (L243) says the green path demands; the milestone named for the flip is unreachable. — Severity: major — Class: blocking — Required fix: move AC-ISB-014's flip to M3 (or split the grep per file and milestone), and re-check every other row's milestone against the file each G command touches.
D5. spec-plan-contradiction — spec.md:L138 (§E.4) vs plan.md:L75 (B.3), L74 — §E.4 promises that `integration.go` and `factory_card.go` gain one call "with no reordering or editing of any existing line", while B.3 extracts the adoption predicate out of factory_card.go:L2229-2238 ("the only edit to existing lines") and allows moving one resolution line in integration.go. The seam argument with t1847 (spec.md:L131) rests on the false sentence. — Severity: minor — Class: blocking — Required fix: reword §E.4 to the real edit set (one added call per verb, one behaviour-preserving extraction in factory_card.go, optionally one moved line in integration.go).
D6. req-ac-exit-code-mismatch — spec.md:L85 (REQ-ISB-011 "proposed 16") vs acceptance.md:L41, L109 — the REQ leaves the code open, the ACs hard-code 16. — Severity: minor — Class: optional — Required fix: state 16 in the REQ or relax the ACs to "a code outside 1-15 and 20".
D7. stale-count — plan.md:L9 — "12 to 14 files" and "5 Go source files" against the allow-list at plan.md:L96-100 (4 non-test Go files; 15 listed paths, 16 if the conditional catalog.yaml regenerates), which sits at the Tier M/L boundary of 15. — Severity: minor — Class: optional — Required fix: restate the file count from the allow-list and say which side of the boundary the conditional file puts the SPEC on.
D8. progress-stale — progress.md:L5-L9 — still says "not yet audit-ready", `open_decisions` Q3 (operator, product-level), gaps G1-G11, against decision-index.md:L19 (Q3 non-blocking) and plan.md G1-G14. progress.md is outside the plan hash, so this is a stale record, not a hash drift. — Severity: minor — Class: optional — Required fix: refresh the E.1 block when the plan is revised.

## Regression Check (iter1 defects)
- D1 adoption-block: RESOLVED (REQ-ISB-016, plan B.3, AC-ISB-016; predicate cite verified at factory_card.go:L2229-2238).
- D2 false-equivalence: RESOLVED (L from rec.Base; cause 2 at integration_merge_step.go:L332 compared).
- D3 red-now-vacuous: UNRESOLVED in substance (new D1/D2).
- D4 fail-open-output-conflict: RESOLVED except one input (new D3).
- D5 q3-routing: RESOLVED.
- D6 ac-gaps: RESOLVED.

## Recommendation
FAIL on MP-8. Fix D1 and D2 together (reproducible RED per AC), then D3, D4, D5, and re-audit delta scope with the CN-4 verb in full (iter 3 is the last, tier ceiling). The substantive design points from iter1 (adoption exemption, record.Base measurement) now hold; the remaining defects are in the verification layer and two cross-artifact statements. Consider reclassifying the ACs whose behaviour cannot have a committed pre-implementation RED as regression-guards rather than inflating the gate.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)
- Claim: iter1 design defects D1/D2 resolved; D3 not resolved; new contract and milestone defects found; MP-8 fails.
- Evidence: the three `go test` runs and the E-01..E-11 greps quoted above (outputs observed in this run); factory_card.go:L2160-2258, integration_merge_step.go:L300-349 read; CN-4 verb output quoted; audit_multi result: codex verdict fail, two P2 findings (E-00a/b/c yield `[no tests to run]` exit 0; B.2 step 1 vs REQ-009/AC-009 output conflict) matching D1 and D3, receipt rcpt-e13ff797f855dc59d1f79645; `moai verify audit-plan --result-file` returned `convergence_check: {ok: true, unmet: []}` (result consistent with the plan; the overall verdict is a reviewed defect, not an unmet gate).
- Baseline-attribution: tree HEAD 82118bc77f301e2f48d8866c586f67c4497e0adc, clean; installed moai build ee981a932, a strict ancestor of HEAD (the audit_multi result carries the same build_lag note), so `moai spec lint` and `moai verify` outputs describe an older build than this tree; `go test` and greps ran from this tree's sources.
- Gaps: audit_multi ran codex only (claude and glm gates are off); the traceability awk verb was replaced by a hand check; the CN-4 verb returned GAP for milestones (read by hand); no planned test exists, so none of the green paths was run; the full `internal/cli` suite was not run; D7 and D8 verbs run as one-liners rather than as the verbatim scripts.
- Residual-risk: K=0 equivalence was derived from code reading plus the Base ancestor semantics, not from a live window (G1); first-parent counting measured in one tree (G2); exit code 16 enumeration is partial (G6); the unresolved placement of the gate relative to `resolveIntegrationTarget` in `acquire` is left to the run phase.

## Operational Notes (unverified)
- Measure the replacement RED before editing the ledger: run `grep -rn 'func EvaluateBaseFreshness' internal --include='*.go'` and expect empty stdout with exit 1 (status: inferred from E-05's shape, not run against a revised cell).
- Measure whether a content-free absorb really leaves the stored `Base` stale: run `moai integration remeasure` after a no-content merge of main on a fixture and read the record's `base` field (status: assumption).

AUDIT-VERDICT: FAIL spec=SPEC-INTEGRATION-STALE-BASE-001 receipts=rcpt-e13ff797f855dc59d1f79645

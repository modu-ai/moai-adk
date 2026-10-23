auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-INTEGRATION-STALE-BASE-001 (card t1834)

verdict: FAIL
audited_sha: 790e03de74a4c849275e3e878fbfba86f37c4c6b
overall_score: 0.83
threshold: 0.80
must_pass_failed: 1
blocking_count: 3
plan_artifact_hash: e6bc17ea40d2464277ef07d32fd0d0e5d3b75502ae8ca7b27d3f1e88a603d1fb
scope: full
fix_scope: acceptance.md#G plan.md#B.3 plan.md#B.5 plan.md#M4 spec.md#E.4
defect_class: D1=ac-wording D2=design D3=design
convergence_overall: fail
required_backend: codex fail
defects: D1 critical/blocking (MP-9), D2 major/blocking, D3 major/blocking, D4 minor/optional, D5 minor/optional, D6 minor/optional

Iteration: 3/3 (tier ceiling) | Tier M (PASS threshold 0.80) | Reasoning context ignored per M1 Context Isolation.
Score trend: iter1 0.72 -> iter2 0.74 -> iter3 0.83 (no regression, no STOP signal). The aggregate clears the 0.80 threshold; the verdict is FAIL on the must-pass firewall (MP-9), not on the score.
Plan Artifact Hash: ComputeHash algorithm of internal/runtime/audit_cache.go (acceptance, decision-index, plan, spec; design/research/tasks absent), reproduced with `python3 -I` over the exact name order and `name:len\0 + bytes + \0` framing, not by the installed binary.
Auditor Version: plan-auditor/v1. Gate plan (`moai verify audit-plan`): `cross_model_active: true`, `enforced_required: [codex]`, claude and glm gates off.

## Must-Pass Results

- [PASS] MP-1 REQ numbering (requirement layer, spec.md:L66-93): REQ-ISB-001..016 sequential, no gap, no duplicate, uniform zero-padding. 16 REQs (Tier M ceiling 16) and 16 ACs (AC-ISB-001..016, acceptance.md:L37-136; ceiling 16).
- [PASS] MP-2 GEARS (judged on the REQ layer in spec.md; ACs are Given-When-Then by design and were not graded here): every REQ is When / Where / While / ubiquitous ("The gate shall ...") form.
- [PASS] MP-3 frontmatter (spec.md:L2-19): 12 canonical fields present, `version: "0.1.0"` quoted, `status: draft`, `priority: P1`, `lifecycle: spec-anchored`, `tags` a comma-separated string, plus optional `tier: M` and `related_specs`. `moai spec lint ... --json --strict` printed one Info finding (OwnershipTransitionUnmeasured), exit 0 (installed build ee981a932 is a strict ancestor of HEAD; see Gaps).
- [N/A] MP-4 language neutrality: single-language Go SPEC.
- [PASS] MP-5 D7: SPEC-MERGE-WINDOW-QUEUE-001 completed, SPEC-LOCAL-MAIN-FLOW-001 completed, SPEC-CANDIDATE-CI-001 implemented; none retired, superseded or archived. SPEC-LANDING-SURFACE-MECH-001, SPEC-LANDING-SURFACE-RULES-001, SPEC-LANE-SELF-RECOVERY-001 do not exist in this tree (disclosed in plan.md G3-G5, SHOULD severity only). No BLOCKING.
- [N/A] MP-6 D8: `grep -c syscall` prints 0 for spec.md, plan.md, acceptance.md.
- [PASS] MP-7: `grep -n '\[NEEDS CLARIFICATION'` over plan.md, progress.md, decision-index.md printed nothing (exit 1); research.md absent.
- [PASS] MP-8 RED-now re-execution (acceptance.md:L8 classifies all 16 ACs release-blocking). Every ledger cell was re-executed on the current tree; see the dedicated section below.
- [FAIL] MP-9 cross-artifact ordering (confirmed candidate, see D1). CN-4 verb: `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 0 exit bindings, 13 ordering candidates`; no `CONFLICT:` line because plan.md binds no criterion through `Exit:` lines. Reading the candidates by hand, acceptance.md:L303 conflicts with plan.md M4 (quoted in D1). A confirmed candidate forces FAIL.

## MP-8 judgment: does the revised RED approach satisfy it honestly?

Answer: yes, MP-8 is satisfied, with a stated limit.

Re-execution on HEAD 790e03de7 (stdout verbatim in the Evidence section):
- E-12a..E-12l: all twelve `grep -rl 'func <Test>(' internal --include='*_test.go'` print nothing, exit 1. Reproduces the ledger.
- E-13 `go doc ./internal/factory EvaluateBaseFreshness`: `doc: no symbol EvaluateBaseFreshness in package ./internal/factory`, exit 1. Reproduces.
- E-01..E-05 (counts 0 / empty, exit 1), E-06 (`50:const configCacheSchemaVersion = 12`, exit 0), E-09 and E-10 (ugrep "No such file", exit 2), E-11 (empty, exit 0): all reproduce byte for byte. E-07 (positive control) and E-08 (baseline, byte-identical pair, exit 0) are declared non-RED and behave as declared.
- No cell is a test-runner citation, so the `ok [no tests to run]` inversion that sank iter2 cannot recur; no cell uses a forbidden metacharacter outside quotes.

Why this is honest and not a relabelled vacuity: iter2's cells asserted a behavioural RED (build failure, exit 1) that existed only in a deleted probe file. The revised cells assert only what a grep can observe: the criterion's own test, or the artifact the work must create or change, is absent today. Each cell says so ("not yet observable: the criterion's own planned test is absent", acceptance.md:L184, L189 and so on) and defers the behavioural RED to a committed RED commit preceding the GREEN commit (acceptance.md:L9, L303). That is the ordering verification-claim-integrity §2.3 requires, because only the commit graph can witness it. The four elements of verification-completeness §2.1 are present: command, verbatim stdout, exit code as its own field, and a document-level pin (acceptance.md:L3). The undecidable-disposition clause does not apply, since every cited command re-executes.

The limit, stated rather than hidden: an absence cell flips green when the test file exists, not when the behaviour exists, so a test with an empty body would turn E-12x green. The safeguards are the green path (`go test -v` naming each test, with a swept count above zero, acceptance.md:L10, L174), the mutant probes in §C, and the Definition of Done clause that each criterion shows a committed RED before its GREEN (acceptance.md:L303). MP-8 confirms that a command's output reproduces; it does not confirm that the command measures the premise its author claims. That boundary is why this is a PASS with the residual noted (D4), not an unqualified endorsement.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Band | Evidence |
|-----------|-------|------|----------|
| Clarity | 0.75 | minor ambiguity in a few places | plan.md:L74 states two incompatible call positions for the `acquire` gate (D2); plan.md:L84 contradicts plan.md:L63 and spec.md:L66 on not-applicable output (D3) |
| Completeness | 0.90 | all sections present | HISTORY spec.md:L25-29, Context L33, Requirements L60, Out of Scope H3 sections L97-120, ACs acceptance.md §C, decision-index.md Q1-Q4, progress.md; plan.md §G gaps G1-G14 |
| Testability | 0.75 | one criterion pair not precisely binary | E-04 counts case-insensitively (`grep -c -i`, acceptance.md:L255) while its green path G-12 is case-sensitive (L162), so a capitalised heading satisfies neither the intent nor the check (D5); absence-only REDs (D4) |
| Traceability | 0.90 | one indirect mapping | CN verb below: 16 REQs, no UNCOVERED, no ORPHAN; DoD placement of the guidance pair contradicts M4 and AC-ISB-012's milestone (D1) |

Traceability verb (AC-4 / AC-5), run in full on spec.md plus acceptance.md: `COLLECTED: 16 REQ definitions (acceptance input: read)`, no `UNCOVERED:` line, no `ORPHAN:` line. Each AC read behind its mapping (acceptance.md:L33 declares 16 `maps REQ-ISB-NNN` pairs and each Then clause tests the mapped REQ).

## Defects Found (structured defect-list)

D1. dod-vs-m4-phase-conflict — acceptance.md:L303 vs plan.md:L114-116, acceptance.md:L27 — Definition of Done: "the sync phase moves the guidance pair into the template tree before the repository copy (Template-First)". Plan M4 (a run-phase milestone): "Gates-companion section, template mirror, `diff -q`, `make build` embed regeneration"; AC-ISB-012's green path is "G-12, G-12b, M1, M4"; REQ-ISB-015 and plan §C put both guidance files in the run-phase allow-list. Both cannot be followed: if sync does the move, G-12 and G-12b cannot be green at M4 and the run-phase scope fence has nothing to fence; if M4 does it, the sync clause names work that no longer exists. The ordering clause (template before repository copy) is satisfiable inside M4; the phase assignment is not. — Severity: critical (MP-9) — Class: blocking — Required fix: rewrite the closing clause of acceptance.md §G to "M4 (run phase) edits the template tree first and the repository copy second, with `diff -q` silent; the sync phase only records that evidence", or move the guidance pair to sync consistently in plan.md M4, AC-ISB-012's green milestone, REQ-ISB-015 and plan §C.

D2. acquire-call-position-contradiction — plan.md:L74 vs spec.md:L138 (§E.4), internal/cli/integration.go:L465-466 — plan.md:L74: the call goes "immediately after the `candidateAcquirePrecondition` block (`:449-458`) and before `config.LoadGitFlowIntegrationConfig` (`:465`); the call sits immediately after the verb's existing `resolveIntegrationTarget` line". In the code `LoadGitFlowIntegrationConfig` is at L465 and `resolveIntegrationTarget` at L466, so "before L465" and "after L466" are different positions, and the gate needs the L466 `branch` to read the integration tip. spec.md §E.4 says "immediately after the existing target resolution". An implementer following plan.md literally cannot get the branch the verb records. The window record write is at L476 and the wait branch at L517, so the corrected position (after L466, before L476) still precedes both. — Severity: major — Class: blocking — Required fix: in plan.md B.3 delete "and before `config.LoadGitFlowIntegrationConfig` (`:465`)" and state "immediately after the `resolveIntegrationTarget` line (`:466`) and before `AcquireIntegrationWindow` (`:476`)", matching spec.md §E.4.

D3. not-applicable-output-contract-conflict — plan.md:L84 (B.5) vs plan.md:L63 (B.2 step 1), spec.md:L66 (REQ-ISB-001), acceptance.md:L95 (AC-ISB-009) — B.5: "Any inability to measure (config unreadable, K negative, git error, integration branch unresolved, no card id, cwd not a card branch) admits and prints one line naming the cause." B.2 step 1, REQ-ISB-001 and AC-ISB-009 require "no card id" and "cwd not a card branch" to admit silently ("no gate line"). An implementation of B.5 fails AC-ISB-009's silent clause. This is the same class as iter2 D3, left over in the B.5 sentence. — Severity: major — Class: blocking — Required fix: in B.5, drop "no card id, cwd not a card branch" from the parenthetical and add "A call that is not applicable (no card id, or cwd not a card branch) admits silently; the cause line is for an applicable call that cannot be measured."

D4. absence-red-scope — acceptance.md:L9, L181-243 — The E-12x cells prove one named test absent per criterion. Several green-path tests named in §D have no absence cell of their own: `TestWorkflowBaseFreshnessGateLoaderRoundTrip` and `TestWorkflowBaseFreshnessGateTemplateParity` (G-07), `TestBaseFreshnessFailOpenCauses` (G-09), `TestBaseFreshnessKDoesNotRelaxMergeStep` (G-10c). The factory-package REDs are a compile failure in the run phase (the symbol `EvaluateBaseFreshness` does not exist, E-13), which is acceptable only if the RED commit records the failing output. The MP-8 judgment above explains why this is not a must-pass failure. — Severity: minor — Class: optional — Required fix (discretionary): add the missing `grep -rl 'func <Test>('` cells, or state in §A.3 that one absence cell stands for all tests named under a criterion.

D5. e04-g12-case-mismatch — acceptance.md:L255 (E-04) vs L162 (G-12) — E-04 is `grep -c -i 'base freshness'` (case-insensitive, 0 today) but its green path G-12 is `grep -c 'base freshness'` (case-sensitive, must be at least 1). A section titled "Base freshness gate" satisfies E-04's flip intent and prints 0 under G-12, a wrong-reason red after implementation (verification-completeness §2). — Severity: minor — Class: optional — Required fix: use the same flag in both commands, or name the heading text in the AC.

D6. wording-residue — plan.md:L9, L11, L146; acceptance.md:L176 — L9 reads "6 Go files of which 4 are non-test and 2 are new non-test" (the 4 non-test files include the 2 new ones; the sentence reads as 4 + 2); L11 repeats "Any addition requires a tier change" twice; L146 (G11) points to "the final message" for the lint result, which is not a plan artifact; acceptance.md:L176 states the cells were re-executed at HEAD da051947f while the document pin is fb749d372 (the two are reconciled by the stated "no `internal/` file differs", which this audit confirmed with `git diff --name-only fb749d372 HEAD`: only `.moai/reports` and `.moai/specs` paths differ). — Severity: minor — Class: optional — Required fix (discretionary): tidy the four sentences.

## Regression Check (iter2 defects)

- iter2 D1 red-now-not-reproducible: RESOLVED. The probe-file cells (E-00a/b/c) are withdrawn (acceptance.md:L178) and replaced by grep cells that reproduce (MP-8 section).
- iter2 D2 red-borrowed-from-another-ac: RESOLVED. AC-ISB-002, 008, 009, 010, 011, 016 each cite their own E-12 cell (acceptance.md:L16-31).
- iter2 D3 output-contract-conflict (unresolvable branch): RESOLVED for that input (plan.md:L63 now routes it to the one-line cause path), but a sibling conflict survives in plan.md:L84 (new D3).
- iter2 D4 green-path-milestone-mismatch: RESOLVED. AC-ISB-014 now flips at M3 (acceptance.md:L29, L128).
- iter2 D5 spec-plan-contradiction (no reordering of existing lines): RESOLVED. spec.md:L138 acknowledges the one behaviour-preserving extraction (REQ-ISB-016, AC-ISB-016).
- iter2 D6 req-ac-exit-code-mismatch: RESOLVED. REQ-ISB-011 states 16 (spec.md:L85).
- iter2 D7 stale-count: RESOLVED in substance (plan.md:L9 now counts 15 paths, 16 with the generated catalog); residual wording in new D6.
- iter2 D8 progress-stale: RESOLVED (progress.md:L5-9 records the iter2 revision and G1-G14).

## Code-cite verification (opened in the tree)

Held: `RemeasureRecord` fields and Base comment (integration_remeasure.go:L35-57), `ValidateRemeasureRecord` (L135 with the `@MX:ANCHOR` at L128-134), `ReadRemeasureRecord` L86, `headTreeSHA` L722; `remeasureWorktreeDir` = `os.Getwd()` (cli/integration_remeasure.go:L118-120); merge cause 1 (integration_merge_step.go:L321-327), cause 2 `record.Base != tip` (L332-335, `MergeExitBaseMoved = 2`), exit table L30-55 (codes 1-15 only; a grep for `code: 16-19` and `Exit* = 16-19` found none); `acquire` ordering (integration.go: drift precondition L444, candidate precondition L457, `LoadGitFlowIntegrationConfig` L465, `resolveIntegrationTarget` L466, `AcquireIntegrationWindow` L476, `integrationWaitInQueue` L517); `complete` (factory_card.go: `if !heldByUs` L2186, `CandidateHoldRefusal` L2194, `AcquireIntegrationWindow` L2201, adoption predicate L2222-2237 with `len(parents) == 3 && parents[2] == cardTip`; step 3 refuses with `MergeExitBaseMoved` when no record exists, so `complete` never creates a record and the REQ-ISB-002 gate precedes a refusal the verb would make anyway); `SettingsDriftGateConfig` (types.go:L825-835) and its default (defaults.go: the sibling blocks near L1541-1550); `configCacheSchemaVersion = 12` (cache.go:L50); inventory entry format and the reader guard `TestShippedConfigKeysHaveReaders` (shipped_key_reader_test.go:L74); `TestMergeStepPreMergeCausesReleaseWithDistinctCodes` (merge_step_test.go:L322) and `TestMergeStepMidstepBranchTipMoveRefusesMerge` (L962) exist, so G-10b sweeps real tests. Design check: K = 0 equivalence with cause 2 follows from measuring L at the stored `rec.Base` (spec.md REQ-ISB-003), consistent with the code at L332.

## Evidence (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

- Claim: iter2's eight defects are resolved; MP-8 now passes; three internal contradictions remain (D1 to D3), one of which (D1) is a must-pass MP-9 finding, so the verdict is FAIL at score 0.83.
- Evidence (commands run in this audit, observed output):
  - `grep -rl 'func <Test>(' internal --include='*_test.go'` for the 12 named tests: empty, `exit:1` each.
  - `go doc ./internal/factory EvaluateBaseFreshness`: `doc: no symbol EvaluateBaseFreshness in package ./internal/factory`, exit 1.
  - `grep -c 'base_freshness_gate' internal/config/types.go`: `0`, exit 1; same grep on the template workflow.yaml: `0`, exit 1; `grep -c 'workflow.base_freshness_gate' internal/config/testdata/shipped_key_inventory.yaml`: `0`, exit 1; `grep -c -i 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md`: `0`, exit 1; `grep -rl 'BASE_FRESHNESS_VIOLATION' internal --include='*.go'`: empty, exit 1.
  - `grep -n 'configCacheSchemaVersion = ' internal/config/cache.go`: `50:const configCacheSchemaVersion = 12`, exit 0.
  - `grep -c 'ValidateRemeasureRecord' internal/factory/base_freshness.go` and the E-10 grep: `ugrep: warning: ... No such file or directory`, exit 2.
  - `git diff --name-only fb749d372e61a6b10c53091362c76b4412827427 HEAD -- internal/factory/base_freshness.go`: empty, exit 0; the unrestricted diff lists only `.moai/reports/t1834/plan-audit-iter{1,2}.md` and `.moai/specs/SPEC-INTEGRATION-STALE-BASE-001/{acceptance,decision-index,plan,progress,spec}.md`.
  - Traceability verb: `COLLECTED: 16 REQ definitions (acceptance input: read)`, nothing else. CN-4 verb: `COLLECTED: 5 milestones in plan order (M1 M2 M3 M4 M5), 0 exit bindings, 13 ordering candidates`, no `CONFLICT:`, no `GAP:`.
  - `mcp__moai__audit_multi` (project_root set to this worktree, target baseBranch, no `gates` argument): `overall_verdict: fail`, backend `codex` gate `required` verdict `fail`, `audit_receipt: rcpt-ffd90c338c578c18a9d3444b`. Codex's three P2 findings are plan.md:L74 (placement; my D2), acceptance.md:L303 (DoD vs M4; my D1), plan.md:L84 (not-applicable output; my D3). I read the cited lines in the tree and confirmed all three before adopting them. `moai verify audit-plan --result-file` returned `convergence_check.ok: true` (no unmet gate: codex answered with a verdict), so the FAIL is a reviewed defect, not an unmet gate.
- Baseline-attribution: tree HEAD 790e03de74a4c849275e3e878fbfba86f37c4c6b, working tree clean at the start of the audit. The installed moai build is v3.2.0-rc.36-t1822c, commit ee981a932, a strict ancestor of HEAD (`git merge-base --is-ancestor` printed ancestor; the audit_multi result carries the same `build_lag` note). `moai spec lint` and `moai verify audit-plan` therefore describe older code than this tree (verification-claim-integrity §2.2); the grep, git, go doc and awk results do not depend on that binary.
- Gaps: claude and glm backends are gate-off, so only codex plus this in-session review ran. No planned test exists, so no green path was run (nothing to run). The live-window behaviour of K = 0 equivalence (plan.md G1) and first-parent counting beyond one tree (G2) were not exercised. The full `internal/cli` suite was not run (repository discipline). `moai spec lint` was judged by an older build.
- Residual-risk: the absence-only RED cells admit an empty-bodied test (D4); the exit code 16 enumeration of `exitCodeError` literals is partial (plan.md G6, my grep found no collision); t1847's SPEC does not exist, so overlap on `acquire` / `factory complete` rests on the placement mitigation (plan.md G3).

## Recommendation

FAIL on MP-9 plus three blocking findings; the score (0.83) is above the 0.80 threshold, so no score-regression STOP applies. This is iteration 3 of 3, the tier ceiling, so the ceiling policy takes the final-hit path. The fix is three wording hunks with the REQ and AC id sets unchanged:
1. acceptance.md §G (L303): make the guidance pair an M4 (run phase) deliverable, template tree first, or move it to sync everywhere (D1).
2. plan.md B.3 (L74): one call position, after `resolveIntegrationTarget` (L466) and before `AcquireIntegrationWindow` (L476) (D2).
3. plan.md B.5 (L84): remove "no card id, cwd not a card branch" from the one-line-cause list; state that a not-applicable call is silent (D3).
D4 to D6 are optional and may be taken at the same time.

## Operational Notes (unverified)

- Measure that the corrected acquire position still precedes both window writes: after the fix run `grep -n 'resolveIntegrationTarget(\|AcquireIntegrationWindow(' internal/cli/integration.go` and read that the gate call line falls between them (status: inferred from the L466 / L476 reads above, not run against a revised plan).
- Measure the delta scope on re-audit: `git diff --name-only 790e03de74a4c849275e3e878fbfba86f37c4c6b HEAD` should list only acceptance.md and plan.md (and optionally spec.md and progress.md) under the SPEC directory (status: assumption).
- Measure D5 before editing: run `grep -c 'base freshness' .claude/rules/moai/workflow/factory-dispatch-gates.md` after the guidance section lands and confirm it prints at least 1 with the heading as written (status: inferred from the E-04 / G-12 flag difference).

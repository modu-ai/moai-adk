auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-FACTORY-DECISION-AUTO-001
Iteration: 3/3 (delta-scoped re-audit; CN-4 re-run in full; Tier L ceiling, final round)
Verdict: FAIL
Overall Score: 0.83
Plan Artifact Hash: not recomputed this iteration (Gap, see §4)
Auditor Version: plan-auditor (agent definition as loaded in this session)

verdict: FAIL
audited_sha: e90b0be7e

Card: t1481 · Tier L (threshold 0.85) · Tree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de · branch WT-decision-automation · HEAD e90b0be7e (iter2 audited 5d094991f, FAIL 0.80)

Reasoning context ignored per M1 Context Isolation. Known bootstrap gap, noted and not failed on alone: decision-index.md labels Q1-Q23 `LEADER-DECIDED`, which is outside manager-spec's four-label vocabulary.

Cross-model: `audit_multi` (target baseBranch, project_root = card tree). claude FAIL (in-session anchor, required). codex FAIL (required: 1×P1, 2×P2; it ran Go tests at e90b0be7e). GLM inconclusive (advisory: "z.ai response carried no text content"). overall_verdict=fail, disagreement_flag=false, participant_count=2. The result carried no `audit_receipt`. build_lag: the MCP binary was built from 45600e4ee, an ancestor of HEAD e90b0be7e.

---

## 1. Claim

FAIL at 0.83, below the 0.85 threshold. All seven iter2 blocking defects (N1-N7) are closed in substance, and the optional items O1-O4 are addressed. MP-1 through MP-9 all pass.

The revision introduces two new blocking defects, and both backends found them independently:

- **N8 (critical, keep-set).** The `audit` decider admits any `DEFAULT-APPLIED` verdict regardless of class. Because the fill now happens before the audit (Q18) and is covered by the hash, a product-level row filled `DEFAULT-APPLIED` passes every mechanical check. REQ-014 does not state that `DEFAULT-APPLIED` is valid only on implementation-level rows that carry a `Default:` line. The product-level keep-set item (REQ-024) is therefore left without a guard. Codex rates this P1.
- **N9 (major).** The new edge T8a (kickoff → run) and the decider value `audit` contradict the completed SPEC-FACTORY-RECORD-001. Its REQ-FR-004 admits exactly the enumerated edges; its AC-005 test asserts 65 accepted pairs; its REQ-FR-019 refuses every decider other than `human` and sends approve to `assigned`. This SPEC writes no Amendments row for any of them; only REQ-SD-016 of SELF-DISPATCH gets one. This is the same class of defect as iter2's N1: an untracked conflict with a tested invariant.

A third finding is minor but blocking. REQ-014 lets "a recorded verdict" pass the `audit` decider, while REQ-018 says an operator's recorded verdict "then reaches run only through the `human` Kickoff path". Both cannot be implemented.

## Must-Pass Results

- [PASS] MP-1 REQ number consistency: REQ-FDA-001 through 025 run in sequence (spec.md:90-263), with no gap or duplicate. Trace verb: `COLLECTED: 25 REQ definitions (acceptance input: read)`, no UNCOVERED, no ORPHAN.
- [PASS] MP-2 GEARS, judged on the requirement layer only (spec.md §C). The patterns are Ubiquitous (001, 002, 006, 009, 010, 024, 025), Event-driven (003, 004, 007, 008, 011-016, 018, 020-023), Where (005, 017; labelled "Capability gate") and While (019). The ACs are Given-When-Then in acceptance.md, which is the verification layer, and are not graded here.
- [PASS] MP-3 frontmatter (spec.md:2-13): id, title, version "0.3.0" (quoted), status draft, created and updated 2026-10-03, author, priority P1, phase, module, lifecycle spec-anchored, tags as a comma string. No rejected aliases.
- [N/A→PASS] MP-4: the SPEC covers no multi-language tooling. Template neutrality is stated at spec.md:270.
- [PASS] MP-5 D7: six referenced SPECs, all `status: completed`; the self-reference is draft. No REVIEW lines, no BLOCKING. N9 concerns SPEC-FACTORY-RECORD-001, which is not referenced at all; it is a consistency defect, not a status-based D7 finding.
- [PASS] MP-6 D8: `grep -c syscall spec.md` returned `0`.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` returned no output, `exit=1`.
- [PASS] MP-8 RED-now: the document-level pin is `5d094991fb586b9c4aadee33b663fe92b686ff18` (acceptance.md:13-14). `git diff --stat 5d094991f e90b0be7e` shows only `.moai/specs/SPEC-FACTORY-DECISION-AUTO-001/*` and `.moai/reports/t1481/plan-audit-iter2.md`, so the probed code is unchanged. I re-executed P1-P31 at e90b0be7e; every probe reproduced its recorded exit code and its deciding lines (§2). All 24 release-blocking ACs cite a probe that reproduces.
- [PASS] MP-9 ordering: the CN-4 heading count is `0` (milestones sit in a table at plan.md:37-47; GAP), so I read the order by hand as M0→M8. The candidates at acceptance.md:14, 27, 32, 37, 41, 52-53, 69, 94, 99 and 124 do not conflict. "M0 … committed before the first implementation commit, re-measured after M7" matches M0 first and M7 before the close. "before the plan audit" (AC-018) and "after the audited SHA" (AC-014) are plan-phase and fixture ordering, not milestone ordering. AC-017's "C2 commit precedes or equals C1" matches M5 "C2 first". No conflict.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.80 | 0.75 | REQ-014 (spec.md:193-194) and REQ-018 (spec.md:221-223) contradict each other on recorded verdicts. spec.md:76-77 §B item 5 still says "kickoff→run keeping the lease", against REQ-015 (spec.md:197-201). plan.md:6-8 still says "Probes measured at `ba2033d22` … Revision 0.2.0" |
| Completeness | 0.80 | 0.75 | All sections present. The SPEC-FACTORY-RECORD-001 reconciliation is missing (not in related_specs, no Amendments row for REQ-FR-004/019 or AC-005). The class guard on DEFAULT-APPLIED is missing |
| Testability | 0.82 | 0.75-1.0 | The matrix is binary-testable. AC-014 (acceptance.md:37) has no fixture for a product-level row (or a Class-less row, or a row without a Default) carrying a `DEFAULT-APPLIED` fill made before the audit. AC-015 does not state the edge-table count change that the AC-005 test will need |
| Traceability | 0.90 | 0.75-1.0 | 25 of 25 mapped, no orphans. P13 now greps the defined tokens (research.md:126-130). The AC-019/020 classification is fixed |

Aggregate ≈ 0.83, below 0.85 (Tier L). This is an improvement on iter2's 0.80, so it is not a regression and no STOP is raised. This is iteration 3, the Tier L ceiling.

## Defects Found (structured defect-list)

N8. DEFAULT-APPLIED-CLASS-UNGUARDED — spec.md:193-194 (REQ-FDA-014 "only a recorded verdict or a `DEFAULT-APPLIED` verdict passes"), design.md:145-146, acceptance.md:37 (AC-FDA-014), :41 (AC-FDA-018). The decider accepts `DEFAULT-APPLIED` on a row of any class. After Q18 the fill happens before the audit and enters the hash, so a product-level row (or a row with no `Class:` line, which REQ-017 treats as product-level, or an implementation-level row with no `Default:`) can be filled `DEFAULT-APPLIED` at plan close. It then passes the hash, `audited_sha` and empty-verdict checks. REQ-018 forbids this fill (spec.md:221-223) and REQ-024 forbids the FOUNDER default from recording a product-level verdict (spec.md:257-260), but no mechanical check and no AC fixture enforces either. The plan-auditor is not obliged to inspect Class correctness or the placement of DEFAULT-APPLIED. Codex agrees (P1, high confidence). — Severity: critical — Class: blocking — Required fix: in REQ-FDA-014, state that a `DEFAULT-APPLIED` verdict passes only on a row whose `Class:` is `implementation-level` and which carries a `Default:` line; on any other row it refuses. Add AC-014 refusal fixtures for `DEFAULT-APPLIED` on (a) a product-level row, (b) a Class-less row, and (c) an implementation-level row without a `Default:`, all filled before the audit with a matching hash. Mirror one of these in AC-018.

N9. FACTORY-RECORD-001-UNRECONCILED — spec.md:197-206 (REQ-FDA-015/016), design.md:139-154, spec.md:15 (related_specs). REQ-FR-004 of the completed SPEC-FACTORY-RECORD-001 (`.moai/specs/SPEC-FACTORY-RECORD-001/spec.md:81-86`) admits "exactly the requested edges enumerated in design.md § Transition Table". Its AC-005 test asserts 65 accepted and 296 refused pairs (`internal/homestate/fr_transition_test.go:124,147`; the `@MX:ANCHOR` at `card_transition.go:92-93` repeats this). Its REQ-FR-019 (spec.md:158-162) sends a kickoff approve to `assigned` and refuses "a decider value other than `human`". Its design records the operator's decider set as `human | llm | llm+jev` (spec.md:205-209). T8a (kickoff → run) and decider `audit` change all three, yet the SPEC neither references SPEC-FACTORY-RECORD-001 nor plans an Amendments row. Codex ran `go test ./internal/homestate -run '^TestFR_AC(005_TransitionTableEdgeCount|013_DecisionPendingHoldsNoLease)$'` at e90b0be7e and observed "65 accepted, 296 refused, 361 total; production table rows: 66", both PASS. Adding T8a will change that count. — Severity: major — Class: blocking — Required fix: add SPEC-FACTORY-RECORD-001 to related_specs. In REQ-FDA-015/016, oblige an Amendments row on REQ-FR-004 (edge table + T8a; AC-005 count 65→66 accepted, refused count adjusted) and on REQ-FR-019 (decider `audit` admitted on T8a only; `human` unchanged). Have AC-015 assert the updated AC-005 count with human and audit allow/refuse rows. Note in decision-index that `audit` lies outside the operator-recorded decider set `human|llm|llm+jev`, so the leader can confirm that this is not an operator-held question.

N10. RECORDED-VERDICT-CONTRADICTION — spec.md:193-194 (REQ-014: "a recorded verdict … passes") vs spec.md:221-223 (REQ-018: the operator's "recorded verdict then reaches run only through the `human` Kickoff path"); decision-index.md:191-192 (Q19). The two statements cannot both be implemented. Codex adds that a re-audit after the operator answers refreshes the hash, so the hash check does not separate the two cases. I am not re-litigating Q19. Q19 says recorded verdicts pass the decider, so the minimal fix is wording. — Severity: minor — Class: blocking — Required fix: choose one rule and state it in both REQs and design §5/§6. Either (a) narrow REQ-018 to "a verdict recorded after the audited SHA reaches run only through `human`", which follows Q19 as written, or (b) make REQ-014 refuse on any product-level row whatever its verdict, which is the keep-set-safe reading; option (b) needs a leader ruling because it changes Q19. Add the matching fixture to AC-014.

N11. STALE-SUMMARY-TEXT — spec.md:76-77 (§B item 5 "kickoff→run keeping the lease") contradicts REQ-015 and Q17, because a kickoff card holds no lease and T8a takes a new one. — Severity: minor — Class: blocking (internal consistency) — Required fix: change the wording to "kickoff→run, leasing the card to its record owner".

O5. plan.md:6-8 still says "Probes measured at `ba2033d22`" and "Revision 0.2.0 answers plan-audit iter1"; acceptance.md pins 5d094991f. — minor — optional — Refresh the text to 0.3.0 / 5d094991f.
O6. spec-workflow.md:408 (local + template) states the hash subject list "Go verbatim" without `decision-index.md`. After M2 that is false, and the change also widens the /moai run Phase 1 skip-cache subject set for every SPEC that carries a decision-index. — minor — optional — Name spec-workflow.md:408 in M2 and AC-025, and say whether the skip-cache invalidation for existing SPECs is intended.
O7. The path by which the `scope: reread` verdict becomes the card's evidence SHA (the T6 → T5 re-entry before T7 / T8a) is unstated in REQ-013. — minor — optional — Add one sentence naming the re-entry edge.
O8. spec.md:325 §G cites only the iter1 verdict. — minor — optional.

## Regression Check (iter2 defects)

- N1 LEASE-AT-KICKOFF — RESOLVED. Authority now rests on the record owner (REQ-016, spec.md:202-206). T8a leases to `OwnerLabel` atomically (REQ-015, spec.md:197-201). Measured: `OwnerLabel` is written only at `card_transition.go:432` (assign) and carried by `next := cur` (:414), so it survives T7. The lease write mirrors `guardLeaseAcquire` (:433-447). P26 and P31 reproduce. AC-015 runs the real chain. A new cross-SPEC conflict follows from the fix (N9).
- N2 DECISION-INDEX-UNHASHED — RESOLVED. REQ-014 puts decision-index in the digest input set (spec.md:190-191). REQ-011's exemption is now "except `progress.md` and `.moai/reports/**`" (spec.md:162). AC-011 makes a decision-index change ineligible, and AC-014 adds a post-audit reclassification fixture. P25 (`exit=1`) and P25c (`94: "spec.md",`) reproduce; `ComputeHash` skips missing files (audit_cache.go:117-126), so SPECs with no decision-index are unaffected. A consequent hole opens (N8).
- N3 UNRANKABLE-IMPL-ROW — RESOLVED. REQ-014 refuses on any FOUNDER row with an empty verdict, of either class, and AC-014 has that fixture.
- N4 EXCEPTION-COND1-NOT-MECHANICAL — RESOLVED. The `release-scope` standing record carries structured fields (REQ-002, spec.md:99-100). Reachability uses only `depends`/`blocks` in a stated direction (REQ-013, spec.md:173-176; design.md:119-123), and the walk is cycle-safe (acceptance.md:112). The direction matches todo_relate.go:13-14 ("A blocks B" = A lands before B; "A depends B" = A waits on B). AC-013 has positive fixtures for each source and a `contains`/`conflicts` negative. Q20 limits it to the todo relate vocabulary (the gtd `depends_on` store is not consulted).
- N5 EXCEPTION-FIELDS-AND-EXIT — RESOLVED. REQ-006 now carries `blocking_count`, `scope`, `defect_class` (a closed enum including `ac-wording`) and `reread_hunks` (spec.md:121-129). REQ-013 requires a full `scope: reread` verdict admitted by the predicate, a hunk diff check, and hold release by `resolves`. AC-006 and AC-013 assert these. P13 greps the defined tokens.
- N6 RG-MISCLASSIFICATION — RESOLVED. AC-019 and AC-020 are release-blocking, with P28/P29/P19 and P30/P20/P21. The M0 values moved to Measurement notes (acceptance.md:50-57). The DoD count is 24, which equals 25 minus RG AC-024.
- N7 SYNC-THRESHOLD-AMBIGUITY — RESOLVED. REQ-009 keeps the T13 label-only check unchanged (spec.md:144-146), and AC-009 characterizes T13 before and after. design.md:73 agrees.
- O1-O4 — RESOLVED (`dispose_in`, the fail-closed Class fallback, anchor hunk ranges, the informational split proposal).

## fix_scope (machine list)

- spec.md#REQ-FDA-014
- spec.md#REQ-FDA-015
- spec.md#REQ-FDA-016
- spec.md#REQ-FDA-018
- spec.md#B-solution-shape (item 5)
- spec.md frontmatter `related_specs`
- acceptance.md#AC-FDA-014
- acceptance.md#AC-FDA-015
- acceptance.md#AC-FDA-018
- design.md#5-factory-audit-decider
- design.md#6-founder-defaults
- decision-index.md (new row for the N9 decider-set note; the N10 choice if (b))
- plan.md#F-milestones (M3 row: Amendments on SPEC-FACTORY-RECORD-001)

## Blocking list

1. N8 (critical) — `DEFAULT-APPLIED` passes the decider on product-level, Class-less, or Default-less rows.
2. N9 (major) — T8a and `audit` conflict with SPEC-FACTORY-RECORD-001 REQ-FR-004/019 and AC-005 (65 pairs), with no reconciliation.
3. N10 (minor) — REQ-014 and REQ-018 contradict each other on recorded verdicts.
4. N11 (minor) — §B item 5 is stale ("keeping the lease").

---

## 2. Evidence (commands run in this audit, observed output)

- `git log --oneline 5d094991f..e90b0be7e` returned `e90b0be7e docs(SPEC-FACTORY-DECISION-AUTO-001): close plan-audit iter2 N1-N7 and O1-O4 (card t1481)`. `git diff --stat 5d094991f e90b0be7e` showed 8 files, all under the SPEC directory plus `.moai/reports/t1481/plan-audit-iter2.md` (486+/133−).
- Probe re-execution in the card tree (some with `-c` in place of `-n`, which gives the same exit and the same deciding presence or absence):
  - P1 `exit=1`; P1c `internal/cli/contract.go:449: Use: "contract",` `exit=0`
  - P2, P3 `exit=1`; P4 four files `:0` `exit=1`; P5, P6 `exit=1`
  - P7 local `:1`, template `:1` `exit=0`; P8 `exit=1`
  - P9 three files `:1` `exit=0`; P9b `108 T7 … guardVerdictPass` / `114 T13 …` `exit=0`
  - P10, P11 `exit=1`; P12 `:1` / `:1` `exit=0`; P13, P14 `exit=1`
  - P15 `109: {"T8", CardKickoff, CardAssigned, guardKickoffDecision},` `exit=0`; P16 `4` `exit=0`
  - P17-P20 `exit=1`; P21 `124: runState, _, probeErr := factoryHookProbeRun(ctx, dbPath, runID)` `exit=0`
  - P22 `56: slog.Warn("factory hook: message broker close failed", …)` `exit=0`; P23 `exit=1`; P24 `1987:` / `2029:` `exit=0`
  - P25 `exit=1`; P25c `94: "spec.md",` `exit=0`; P26 `160: if k.LeaseHolder != "" || …` `exit=0`
  - P27 `exit=1`; P27c `36:` / `41: Use: "relate <a> <b> --relation <contains|absorbs|replaces|conflicts|blocks|depends>",` `exit=0`
  - P28 local `:1`, template `:1` `exit=0`; P29 auto-semantics `:1`×2, watchdog `:0`×2 `exit=0`; P30 `145: s, err := factorymsg.Open(root, runID)` `exit=0`; P31 `432:` / `439:` `exit=0`
- Trace verb (scratchpad t1481-iter3-verbs.sh, literal paths): `COLLECTED: 25 REQ definitions (acceptance input: read)`, no UNCOVERED, no ORPHAN.
- CN-4: plan milestone-heading count `0` (GAP; order read by hand). Ordering-word records in acceptance.md: lines 14, 27, 32, 37, 41, 52, 53, 69, 94, 99, 124, all judged above.
- D7: AUTONOMY-BATCH-GATE, AUTONOMY-GATE-REWIRE, DECISION-AUTHORITY, FACTORY-SELF-DISPATCH, FACTORY-STALE-RUN-HEAL, SYNC-PARALLEL-DOCS all `status: completed`. D8 `0`. MP-7 `exit=1`.
- N1 closure reads: `grep -n OwnerLabel internal/homestate/card_transition.go` returned `432`, `439`, `440`, `625` (the write at 432 only); `card_transition.go:412-447` was read.
- N9 reads: `card_transition.go:92-120` (MX anchor "AC-005 pins its size at 65 accepted pairs", edges T2-T18); `fr_transition_test.go:124,147` (`want 65`, `want 65 / 296`); SPEC-FACTORY-RECORD-001 spec.md:81-86, 154-162, 205-209, `status: completed`.
- N2 reads: audit_cache.go:84-137 (`planArtifactNames` without decision-index; missing files skipped); kickoff/decide.go:374-405 (the auditor-written `plan_artifact_hash` is compared with `ComputeHash`).
- N4 reads: gtd_relation.go:9-21 (a separate `depends_on` vocabulary); todo_relate.go:1-21.
- audit_multi: overall_verdict fail. Codex P1 is N8, codex P2 is N9 (with its Go test run), codex P2 is N10.

## 3. Baseline-attribution

Every grep and file read above ran in this audit against the card tree at HEAD e90b0be7e. That tree differs from the probe pin 5d094991f only in SPEC and report files, per the git diff above. The git commands ran from my own worktree (the object store is shared) with explicit SHAs. Backend results are attributed to MCP build 45600e4ee, an ancestor of HEAD; the tool reported the build_lag.

## 4. Gaps

- Refused commands: the worktree guard refused `git -C <card tree> …` and a `sed` with a shell-variable path. The Grep tool is not available in this session. Fallbacks: git commands run from my own worktree with explicit SHAs, `grep` through Bash after `cd` into the card tree, and the Read tool. These outputs are equivalent for the claims made.
- I did not run Go tests. The AC-005 count (65/296; 66 production rows) and the PASS of TestFR_AC013 at e90b0be7e are codex's measurements. I independently read the assertions in fr_transition_test.go and the MX anchor.
- The Plan Artifact Hash was not recomputed.
- The GLM advisory was inconclusive. `audit_multi` returned no `audit_receipt`.

## 5. Residual-risk

- residual_risk_note (verbatim): "required-backend FAIL: claude, codex".
- Even after N8-N11 are fixed, the decision-index Class assignment and any "recorded verdict" made before the audit stay self-attested by the authoring session (spec.md §E.2). The audit hash pins the bytes, not who wrote them.
- Adding decision-index.md to `ComputeHash` invalidates the /moai run Phase 1 skip-cache entries of every SPEC that carries one (O6).
- Run M0 may move the recheck default (Measurement note M0(b)).

## Recommendation

This is iteration 3, the Tier L ceiling. The fixes are small and confined to the anchors listed: N8 (one REQ clause plus three AC-014 fixtures), N9 (related_specs, two Amendments obligations, an AC-015 count assertion), N10 (one wording choice), and N11 (one phrase). Under the current Max-3 contract, the orchestrator escalates the choice to the leader: PASS-with-debt is not admissible while N8, a critical keep-set defect, is open, so the options are a delta round scoped to fix_scope above, or a hold. N8 must be fixed before Kickoff under any disposition.

## Operational Notes (unverified)

- assumption: measure the N9 impact after the fix with `go test ./internal/homestate -run '^TestFR_AC005_TransitionTableEdgeCount$' -count=1`; expect the fixture count to change with T8a.
- inferred (rule: REQ-014 text admits "a `DEFAULT-APPLIED` verdict" without a class clause): an N8 fixture is a product-level row with `Operator verdict: DEFAULT-APPLIED …` committed before the audit; the decider as specified admits it.

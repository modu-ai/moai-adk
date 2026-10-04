auditor-model: claude-opus-5-5

# SPEC Review Report: SPEC-FACTORY-DECISION-AUTO-001
Iteration: 4 (a single extension past the Tier L ceiling, granted by the leader; delta-scoped; the CN-4 verb was re-run in full)
Verdict: PASS
Overall Score: 0.885
Plan Artifact Hash: not recomputed (Gap, see §4)
Auditor Version: plan-auditor (agent definition as loaded in this session)

verdict: PASS
audited_sha: a13b83868

Card: t1481 · Tier L (threshold 0.85) · Tree /Users/goos/MoAI/moai-adk-go/.claude/worktrees/agent-a6b97f013ea3df7de · branch WT-decision-automation · HEAD a13b83868 (iter3 audited e90b0be7e: FAIL 0.83)

Reasoning context ignored per M1 Context Isolation. Leader rulings Q1-Q26 are accepted as decided. The `LEADER-DECIDED` label is a known bootstrap gap and does not fail this audit on its own.

Cross-model check: `audit_multi` (target baseBranch, project_root = card tree) returned overall_verdict=pass, disagreement_flag=false, participant_count=2. Claude (required, in-session anchor): pass. Codex (required): pass, no findings. GLM (advisory): inconclusive, "z.ai response carried no text content". No `audit_receipt` was issued. build_lag: the MCP binary was built from 45600e4ee, an ancestor of a13b83868. This SPEC delta touches no Go code, so the lag does not affect the plan text under review.

---

## 1. Claim

PASS at 0.885, above the Tier L threshold of 0.85. All four iter3 blocking defects (N8, N9, N10, N11) are closed in substance, and so is optional item O6. The delta did not break anything that was already settled. The keep-set stays human: REQ-FDA-024 (spec.md:268-271) is unchanged. All of MP-1 through MP-9 pass.

## Must-Pass Results

- [PASS] MP-1: REQ-FDA-001 through 025 run in order with no gaps. The trace verb printed `COLLECTED: 25 REQ definitions (acceptance input: read)` and no UNCOVERED line.
- [PASS] MP-2 (requirement layer only): the patterns of REQ-014/015/016/018 are unchanged (Event-driven), as spec.md:188, :201, :209, :228 show. ACs are judged on the verification layer and are not graded here.
- [PASS] MP-3: frontmatter at spec.md:2-16 has `version: "0.4.0"` (quoted) and all 12 fields. `related_specs` now includes SPEC-FACTORY-RECORD-001.
- [N/A→PASS] MP-4: this SPEC is not multi-language tooling.
- [PASS] MP-5 D7: the newly referenced SPEC-FACTORY-RECORD-001 has `status: completed`. It is reconciled through the planned Amendments rows (spec.md:206-208, :214-216), and no status is retired, superseded, or archived. The trace verb also printed `ORPHAN: REQ-FR-004` and `ORPHAN: REQ-FR-019`. These are cross-SPEC references on an AC line (AC-FDA-015), not defects of this SPEC.
- [PASS] MP-6 D8: `grep -c syscall spec.md` returned `0`.
- [PASS] MP-7: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` printed nothing (`exit=1`).
- [PASS] MP-8: since e90b0be7e no code has changed (`git diff --stat e90b0be7e a13b83868 -- internal cmd pkg` returned empty). I re-ran the probes the delta-touched ACs cite at a13b83868, and each reproduced:
  - P14 `exit=1`
  - P15 `109: {"T8", CardKickoff, CardAssigned, guardKickoffDecision},` `exit=0`
  - P25 `exit=1`
  - P25c `94: "spec.md",` `exit=0`
  - P26 `160: if k.LeaseHolder != "" || …` `exit=0`
  - P18 `exit=1`

  The other probes were re-executed at iter3 against identical code.
- [PASS] MP-9: CN-4 printed `COLLECTED: 0 milestone headings, 10 ordering candidates` with GAP (the milestones sit in a table), so I read the order by hand as M0→M8. The candidates were at lines 12, 27, 32, 37, 41, 52, 69, 94, 99 and 124. The only new ones are line 37 (AC-014, "before the audit" / "before the audited SHA") and line 41 (AC-018, "before the plan audit"). Both order fixtures within one milestone. Neither orders work across milestones, so there is no conflict.

## Category Scores

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.90 | 0.75-1.0 | REQ-014 (spec.md:194-197) and REQ-018 (spec.md:233-235) now agree: a pre-audit verdict passes and a post-audit verdict goes through `human`. §B item 5 (spec.md:77-78) agrees with REQ-015. design.md §5/§6 match |
| Completeness | 0.88 | 0.75-1.0 | SPEC-FACTORY-RECORD-001 is reconciled (related_specs; REQ-015/016 Amendments; plan M3 row; decision-index Q25). Minor: the plan M2 row does not name the spec-workflow.md edit that design §5 assigns to M2 |
| Testability | 0.88 | 0.75-1.0 | AC-014 carries three pre-audit-filled refusal fixtures (a/b/c) and a before-SHA pass fixture. AC-018 adds the product-level DEFAULT-APPLIED refusal. AC-015 asserts 66/295 and the refusal of `audit` on other edges |
| Traceability | 0.87 | 0.75-1.0 | 25/25 REQs mapped, no in-SPEC orphans. The REQ-FR-019 obligation of REQ-016 is checked in AC-015 (mapped to REQ-015), an indirect mapping |

Aggregate ≈ 0.885, above the iter3 score of 0.83, so this is no regression.

## Defects Found (structured defect-list)

O9. PLAN-M2-OMITS-SPEC-WORKFLOW — plan.md:41 (M2 row) vs design.md (§5 "M2 updates the hash-subject sentence in `spec-workflow.md` (local + template)"). The plan row does not name the edit that design assigns to it. — Severity: minor — Class: optional — Fix: add "spec-workflow.md hash-subject sentence (local + template)" to the M2 row, or let run M2 follow design §5.

O10. FR-019-AMENDMENT-INDIRECT-TRACE — spec.md:214-216 (REQ-FDA-016 obliges the REQ-FR-019 Amendments row) is asserted in acceptance.md:38 (AC-FDA-015, which maps to REQ-FDA-015 only). — Severity: minor — Class: optional — Fix: add REQ-FDA-016 to AC-015's REQ column, or move the FR-019 assertion into AC-016.

No blocking defects.

## Regression Check (iter3 defects)

- N8 DEFAULT-APPLIED-CLASS-UNGUARDED: RESOLVED.
  - REQ-014 (spec.md:195-197) reads: "every row holding a `DEFAULT-APPLIED` verdict carries `Class: implementation-level` and a `Default:` line — a `DEFAULT-APPLIED` verdict on a `product-level` row, on a row with no `Class:` line, or on a row with no `Default:` line refuses".
  - design.md §5 mirrors it, and design §6 explains why the hash alone is not enough.
  - AC-014 (acceptance.md:37) adds refusal fixtures (a), (b) and (c), each filled before the audit with a matching hash. AC-018 (acceptance.md:41) adds "a product-level row carrying `DEFAULT-APPLIED` is refused by the decider".
  - decision-index Q24 records the ruling, and plan M3 names the guard.
- N9 FACTORY-RECORD-001-UNRECONCILED: RESOLVED.
  - SPEC-FACTORY-RECORD-001 is now in related_specs (spec.md:15).
  - REQ-015 (spec.md:206-208) obliges the REQ-FR-004 Amendments row: T8a kickoff→run, 65/296 → 66/295. The arithmetic holds: 66+295 = 361.
  - REQ-016 (spec.md:214-216) obliges the REQ-FR-019 Amendments row: `audit` on T8a only, with `human`→`assigned` unchanged.
  - AC-015 asserts the new count, the refusal of `audit` on every other edge, and both Amendments rows.
  - Q25 explains why this is not operator-held: auto-semantics §9 classifies factory kickoff as AUTONOMOUS, and `human` stays the path for keep-set cases.
  - Measured: card_transition.go lists only T7 (→kickoff), T8 (kickoff→assigned) and T9 (kickoff→blocked) at :108-110. No kickoff→run pair exists today, so T8a is a new distinct pair and the 66 count is consistent. fr_transition_test.go:124 and :147 still assert 65 and 65/296, which is the RED the amendment changes.
- N10 RECORDED-VERDICT-CONTRADICTION: RESOLVED, following Q19 (option a).
  - REQ-018 (spec.md:233-235): "a verdict recorded after the audited SHA reaches run only through the `human` Kickoff path, while a verdict recorded before the audited SHA is covered by the audited hash and passes the `audit` decider under REQ-FDA-014".
  - design §6 matches. AC-014 adds "an operator verdict recorded before the audited SHA on a product-level row passes". Q26 records the ruling.
- N11 STALE-SUMMARY-TEXT: RESOLVED. spec.md:77-78 now reads "kickoff→run, leasing the card to its record owner". A grep for `keeping the lease|recorded verdict then reaches` returned `exit=1`.
- O6: RESOLVED in design §5 ("Digest-input widening (O6, intended)"). The plan row is a residual minor gap (O9).
- O5, O7, O8 (optional, iter3): not addressed, and not required.

## Blocking list

None.

---

## 2. Evidence (commands run in this audit, observed output)

- `git log --oneline e90b0be7e..a13b83868` printed `a13b83868 docs(SPEC-FACTORY-DECISION-AUTO-001): delta round for plan-audit iter3 N8-N11 + O6 (card t1481)`.
- `git diff --stat e90b0be7e a13b83868` showed 7 files, all under the SPEC directory, plus `.moai/reports/t1481/plan-audit-iter3.md`. `git diff … -- internal cmd pkg` was empty.
- `git diff e90b0be7e a13b83868 -- .moai/specs/SPEC-FACTORY-DECISION-AUTO-001/` was read in full. It touches acceptance.md (AC-014/015/018), decision-index Q24-Q26, design §5/§6, the plan M3 row, progress bookkeeping, and spec frontmatter, HISTORY, §B5 and REQ-014/015/016/018.
- `grep -n CardKickoff internal/homestate/card_transition.go` printed `108 T7`, `109 T8`, `110 T9` and `474`.
- Probes: P14 `exit=1`; P15 `109:` `exit=0`; P25 `exit=1`; P25c `94:` `exit=0`; P26 `160:` `exit=0`; P18 `exit=1`. MP-7 `exit=1`; D8 `0`; SPEC-FACTORY-RECORD-001 `status: completed`.
- The trace verb (scratchpad t1481-iter4-verbs.sh, literal paths) printed `COLLECTED: 25 REQ definitions (acceptance input: read)`, `ORPHAN: REQ-FR-004` and `ORPHAN: REQ-FR-019` (cross-SPEC), and no UNCOVERED line.
- CN-4 printed `COLLECTED: 0 milestone headings, 10 ordering candidates` and `GAP: 0 milestone headings`. Every candidate was judged above.
- audit_multi: overall_verdict pass, claude pass, codex pass, glm inconclusive.

## 3. Baseline-attribution

All greps and reads ran in this audit against the card tree's working files at HEAD a13b83868 (the card tree was clean as far as the delta diff shows). The git commands ran from my own worktree with explicit SHAs, which works because the object store is shared. The backend verdicts come from MCP build 45600e4ee.

## 4. Gaps

- The worktree guard refused `git -C <card tree>`, `cd <card tree> && git …`, and a `sed` with a variable path. As fallbacks I ran git from my own worktree with explicit SHAs, grep on absolute paths, and the Read tool. I did not confirm the card tree's working-copy status (`git status`) directly.
- Receipt: the stop hook refused the first PASS because it cited no receipt. I then made a separate `codex_audit` call (adversarial, baseBranch, project_root = card tree). It returned `verdict: pass` with no findings, but carried no `audit_receipt` field. The card tree's `.moai/config/sections/workflow.yaml` sets `audit.model: multi` and has no explicit `audit.gates.codex: required` block in the excerpt read. The server mints a receipt only under that explicit setting, so no receipt id exists to cite. The PASS rests on two independent pass verdicts (audit_multi codex, and this codex_audit), not on a receipt.
- I did not run Go tests, so the 65/296 counts were read from the test source, not executed.
- I did not recompute the plan-artifact hash.
- The GLM advisory was inconclusive, and no `audit_receipt` was issued.
- MP-8: in this audit I re-ran only the probes cited by the ACs the delta touched. P1-P31 were re-run at iter3 against identical code.

## 5. Residual-risk

- Pre-audit verdict lines in decision-index, including an "operator verdict" on a product-level row, are attested by the authoring session itself. The hash pins the bytes, not who wrote them (spec §E.2; leader rulings Q19/Q26). A forged pre-audit operator verdict on a product-level row would pass the decider mechanically. This is the iter3 residual, accepted by ruling, and not re-opened here.
- Widening the skip-cache key re-audits existing SPECs that carry a decision-index once (stated as intended).
- Run M0 may move the recheck default.

## Recommendation

PASS. N8-N11 are closed with REQ text, design text, AC fixtures and decision-index rulings that agree with each other. Codex independently returned pass. The two optional items (O9, O10) may be folded into run at the orchestrator's discretion.

## Operational Notes (unverified)

- assumption: after M3, measure `go test ./internal/homestate -run '^TestFR_AC005_TransitionTableEdgeCount$' -count=1` and expect it to assert 66/295.

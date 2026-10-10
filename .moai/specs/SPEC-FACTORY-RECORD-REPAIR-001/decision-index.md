# Decision index: SPEC-FACTORY-RECORD-REPAIR-001

Card t1621 · Tier M · version 0.2.0 (narrowed to the three surviving items per d-20261010T053435Z-006f; operator verdicts per d-20261010T053034Z-1a19; earlier history: 0.1.1, repair round 1 after plan audit iteration 1, FAIL 0.56, 17 blocking defects).

This index records every decision the SPEC carries: rows Q1 to Q19 and Q21, and one note on a proposal closed without a decision (Q20). Each row names what is unresolved, the options, what was applied, the basis, and the evidence. It is stateless on the status axis: the lifecycle lives in spec.md alone.

## How to read a row

- **Label** is one of exactly four values: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. DECIDED and POLICY-COVERED need an authority anchor that exists in the committed tree. A row whose anchor cannot be verified is FOUNDER: escalate, never downgrade.
- **Class** is product-level when the decision changes a shipped command's default user-visible behaviour, removes a user-facing feature, or changes a template default. Otherwise it is implementation-level. A row without a Class is product-level.
- **Status** takes one of six values: `open` for an open FOUNDER row (its Operator verdict line is literally empty); `DEFAULT-APPLIED` for a row whose default was applied at plan close; `operator verdict recorded` for a FOUNDER row whose Operator verdict line carries an operator's answer, cited to its board record (a board record has no digest, so the row is not DECIDED); `not applicable` for a row whose subject a later ruling retired, cited to that ruling; `not required (POLICY-COVERED)` for a POLICY-COVERED row; and `DECIDED` with an authority anchor for a decided row. Status does not route the row. The summary table has no status column.
- **Operator verdict** stays empty on an open FOUNDER row, and an open FOUNDER row blocks the autonomous Kickoff until the founder verdicts it. `DEFAULT-APPLIED` is written only on an implementation-level row whose options the published Default rule ranks, at plan close. It is not written on DECIDED or POLICY-COVERED rows, which need no verdict.

## Routing (revision 0.2.0)

1. **Narrowing applied (006f).** Leader ruling `d-20261010T053435Z-006f` re-scopes the card body to three items. The re-scope text is quoted in spec.md HISTORY 0.2.0. Retired requirements and criteria: REQ-FRR-008, -009, -010, -012, -013 and AC-FRR-008, -009, -010, -012, -013, -016. Their rows are marked not applicable: Q2, Q4, Q5, Q6, Q7, and Q13. Each cites `d-20261010T053435Z-006f`. The retired text is kept in git history at `53ea27f4f`.
2. **Operator verdicts applied (1a19).** Q1 (B1: the surviving criteria are reclassified as regression guards and the release-gate wording is removed), Q8 (a), Q9 (a), Q10 (ii, at T3), and Q18 (a). Each verdict cell cites `d-20261010T053034Z-1a19`. The 1a19 board line carries its header and evidence references, not the values. The values written here are the leader's assignment for this revision, and they reproduce `.moai/reports/t1621/operator-brief.md` §2.
3. **Not written: Q4 (d), Q5 (a), Q6 (c), and Q7 (a). Leader confirmation requested.** The leader's assignment lists these four as surviving. The card body places their subjects in old items (4) and (7). Item (4) names the interim empty-stage guard, which is Q6 and Q7. Item (7) names the holder audit re-entry edge, which covers the sync-audit to run ruling (Q4) and the holder re-stamp (Q5). Items (4) and (7) are retired by 006f, so the rows are marked not applicable and no verdict is written on them. If the leader confirms that they survive, the four verdicts apply and the retired text returns from `53ea27f4f`.
4. **Open FOUNDER row: Q21.** Its Operator verdict is empty, and it is the only such row. The autonomous Kickoff stays blocked until it is answered. Card body item (3) names the pushed-to-done edge for this card, but the SPEC carries no requirement for it, and its owner is open.
5. **Authority register note.** The board records cited here (`d-20261010T053435Z-006f`, `d-20261010T053034Z-1a19`, and `d-20261010T010044Z-57f8`) carry no digest, so none anchors a DECIDED row. A FOUNDER row resolved by an operator's answer carries the answer and its record id, and its Status reads `operator verdict recorded`.
6. **Counts.** Operator verdicts recorded: Q1, Q8, Q9, Q10, Q18. Not applicable: Q2, Q4, Q5, Q6, Q7, Q13. DEFAULT-APPLIED: Q3, Q12, Q14, Q15, Q16, Q17. POLICY-COVERED, no verdict needed: Q11, Q19. Open: Q21.
7. **Kickoff check, as the code reads it.** `internal/homestate/card_audit_kickoff.go:118-123` refuses a FOUNDER row whose Operator verdict value is literally empty. After this revision the only such row is Q21. `card_audit_kickoff.go:113-117` refuses DEFAULT-APPLIED on any row that is not implementation-level with a Default line; each DEFAULT-APPLIED row here carries both.
8. **Round-2 split withdrawn.** A proposed move of the renewal verb and TTL key to t1618 was withdrawn by board d-20261010T012900Z-35d1. It is recorded under "Closed without a decision", not as a row. This SPEC (card t1621) owns those deliverables.

## Pinned and cited authority

Board record d-20261010T010044Z-57f8, copied from the `moai decision read --all` output in revision 0.1.1. It is quoted, not digest-pinned, and it is an evidence note only.

```text
decision record: decided_by=leader session 069be28e (factory leader, run tmnboq) evidence_refs=.moai/worktrees/t1621/.moai/reports/t1621/plan-audit-iter1.md; lane-1 report 2026-10-10 ladder_path=leader ruling (verification-completeness.md 2.1; verification-claim-integrity.md 2.3) id=d-20261010T010044Z-57f8 scope=card:t1621 kind=ruling
```

Board records that govern revision 0.2.0, copied from the same output and cited as evidence notes, not pinned:

```text
decision record: decided_by=leader session 359ce07d (factory leader, run tmnboq, generation 3) evidence_refs=moai todo list 05:3xZ, t1621 body tail: '[재범위 10-10·3.2 계획 §1.1] 범위를 팩토리 v2(3.2-2-2) 선행 수리로 한정: (1) 임대 renew 동사 + TTL 설정 키 + 워치독 첫 단계 갱신 (2) 만료 수거 원자화 (3) pushed→done 엣지 개방 + 대역 외 쓰기 탐지. (4)~(12)는 3.2-2-2가 흡수 ... 새 본문으로 재감사' (leader read); lane-1 message 05:31Z; board 1a19 403d 3f34 ladder_path=leader ruling reconciling operator decision 1a19 with the operator-approved queue re-scope id=d-20261010T053435Z-006f scope=card:t1621 kind=ruling
decision record: decided_by=operator (first-hand AskUserQuestion answers to leader session 359ce07d, 2026-10-10T05:0xZ; recorded verbatim by leader) evidence_refs=.moai/worktrees/t1621/.moai/reports/t1621/operator-brief.md:16-38 ladder_path=operator first-hand answer id=d-20261010T053034Z-1a19 scope=card:t1621 kind=ceiling-exception
```

The card body that 006f quotes is the t1621 queue body read with `moai todo show t1621` in revision 0.2.0. Items (4) and (7) are quoted in spec.md HISTORY 0.2.0.

## Summary

| Row | Decision | Defects | Label | Class | Verdict |
|---|---|---|---|---|---|
| Q1 | Release gate for the criteria at plan time | D1 | FOUNDER | implementation-level | operator verdict (1a19) |
| Q2 | AC-FRR-016 classification | D12 | FOUNDER | implementation-level | not applicable (006f) |
| Q3 | Watchdog cadence and lease gap (evidence for the TTL design) | D15, plan D4 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q4 | Sync-audit to run: ruling-based discriminator and reason tokens | D2, D3, plan D3 | FOUNDER | product-level | not applicable (006f) |
| Q5 | Holder re-stamp refuses on audited-SHA mismatch | D4 | FOUNDER | product-level | not applicable (006f) |
| Q6 | Operator carrier for the empty-stage to run edge | D6, plan D2 | FOUNDER | product-level | not applicable (006f) |
| Q7 | Interim guard covers every non-empty SPEC id | D5 | FOUNDER | product-level | not applicable (006f) |
| Q8 | Recovery command named in the refusal | D14, plan D6 | FOUNDER | product-level | (a), 1a19 |
| Q9 | Holder's own expired renewal is refused, not recovered | D14, plan D7 | FOUNDER | product-level | (a), 1a19 |
| Q10 | Worktree path: source and timing | D8, plan D5 | FOUNDER | product-level | (ii) at T3, 1a19 |
| Q11 | decision-index.md written in this round | D16, plan D10 | POLICY-COVERED | implementation-level | not required |
| Q12 | progress.md with the minimal §E skeleton | D16, plan D10 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q13 | Residual route for a card with an empty SPEC id | plan D8 | FOUNDER | implementation-level | not applicable (006f) |
| Q14 | Lens-4 defects left unassigned | plan D9 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q15 | Which numbering is used for the defects | plan D1 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q16 | What the `module` field names | plan D11 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q17 | What priority the SPEC carries | plan D12 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q18 | Configuration key name `workflow.factory_lease.duration` | plan D13 | FOUNDER | product-level | (a), 1a19 |
| Q19 | Repository-local overrides a template sync must not overwrite; QG-4 scope | D17 | POLICY-COVERED | implementation-level | not required |
| Q21 | Pushed-to-done (card body item 3): whether it is in this card, its owner, and its content | card body item (3) | FOUNDER | product-level | (open) |

## Rows

### Q1: How is the release gate set for the criteria at plan time? (verdict D1; lane update item 1)

Label: FOUNDER
Class: implementation-level
Status: operator verdict recorded (1a19); the DEFAULT-APPLIED route of revision 0.1.1 is superseded
Authority anchor: none routed. Anchor search, run before the DEFAULT route was applied: the version-led HISTORY and Amendments rows of the 840 completed SPEC files were searched for plan-time RED, regression-guard, undecidable-disposition, vacuous, and green-before-work terms. Candidates were verified at HEAD with `git grep`, each SPEC at `status: completed`. SPEC-DECISION-AUTHORITY-001, HISTORY row 0.1.1, item D3: AC-DA-019 reclassified regression-guard because no behavioural RED is producible at plan time. SPEC-GITHUB-FLOW-CI-RESIDUE-001, HISTORY row 0.3.0, item D5: the test-based RED-now items fall under the §2.1 undecidable disposition. The lane applied the DEFAULT route, so neither candidate is routed DECIDED. The lane can re-route this row if it judges the question identical.
Precedent (noted, not DECIDED):
```text
.moai/specs/SPEC-DECISION-AUTHORITY-001/spec.md, HISTORY row 0.1.1, line 26: D3: AC-DA-019 reclassified regression-guard — no behavioral RED is producible at plan time (design §5 principle).
.moai/specs/SPEC-GITHUB-FLOW-CI-RESIDUE-001/spec.md, HISTORY row 0.3.0, line 30: **D5** — RED-now 원장을 4요소(명령·원문 stdout·종료 코드·SHA) 셀로 재구성(관측 적근거 7건, 테스트 기반 4건은 §2.1 undecidable disposition).
```
Why unresolved: the audit asked for an observed RED for each release-blocking criterion on the unchanged tree, or for an operator decision. While a criterion's test does not exist, no RED can be observed on that tree, so MP-8 fails for every criterion.
Options (verdict D1): (a) RED produced in the run phase, with a stated change to the Tier M plan gate; (b) reclassification of the affected criteria.
Applied in 0.1.1: (b). Every criterion is classified regression-guard at plan time, and each criterion states its classification. The RED for each behaviour criterion was to be captured at run phase in its own commit, made before the fix commit, carrying an observed EXPECTED_RED whose failure is an assertion failure. A compile or discovery failure is TOOL_FAILURE and is never recorded as RED.
Applied in 0.2.0: the operator verdict (1a19) is B1. The surviving criteria (AC-FRR-001 to -007, -011, -014, and -015) are regression guards, and the release-gate wording is removed from acceptance.md (Adoption status, the criteria, and Definition of Done). No criterion gates the release at plan time.
Basis: `.claude/rules/moai/development/verification-completeness.md` §2.1 (the undecidable disposition) and §2 (the vacuous direction). `.claude/rules/moai/workflow/tdd-result-contract.md` (EXPECTED_RED and TOOL_FAILURE). The leader's ruling d-20261010T010044Z-57f8 is cited as an evidence note only.
Evidence: plan-audit-iter1 D1 and MP-8. plan-audit-iter1 E4: `go test ./internal/homestate/ -run TestRenewRefusesExpiredWithoutWrite -count=1` printed `[no tests to run]` with exit 0. plan-audit-iter1 E5: AC-FRR-016 green on the unchanged tree. plan-audit-iter2 B1 (MP-8, FAIL) and its required fix: an operator verdict as an operator line, not a lane default.
Default (0.1.1, superseded): regression-guard at plan time. The RED is captured at run phase in its own commit before the fix. TOOL_FAILURE is never recorded as RED.
Operator verdict: B1 (reclassify the surviving criteria as regression guards; remove the release-gate wording). Cited: d-20261010T053034Z-1a19 (operator ceiling-exception; board record, not digest-pinned, so not DECIDED).

### Q2: Is AC-FRR-016 a release-blocking criterion or a regression guard? (verdict D12; lane update item 2)

Label: FOUNDER
Class: implementation-level
Status: not applicable (retired by d-20261010T053435Z-006f: AC-FRR-016 counted only the three rows that item (7)'s requirements add)
Authority anchor: none routed. Anchor search, run with the same terms and scope as Q1 before the DEFAULT route was applied. Candidates were verified at HEAD with `git grep`, each SPEC at `status: completed`. SPEC-TODO-HOLD-STATE-001, HISTORY row 0.2.0, item D1: an autodone acceptance criterion green at M1 is pinned as a regression guard, not a live red. SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001, HISTORY row 0.2.1, item D2: AC-IFU-029 passed before any work and is recorded as a regression guard with its baseline. The lane applied the DEFAULT route, so neither candidate is routed DECIDED. The lane can re-route this row if it judges the question identical.
Precedent (noted, not DECIDED):
```text
.moai/specs/SPEC-TODO-HOLD-STATE-001/spec.md, HISTORY row 0.2.0, line 29: **D1**: the autodone filter-direction account was inverted — `todo_autodone.go:281-285` is `if State != Queued && State != Picked { continue }`, so a `hold` card is SKIPPED by the existing filter the moment the state exists; the site produces no behavioral defect and the autodone AC is a green-at-M1 REGRESSION GUARD, not a live red.
.moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/spec.md, HISTORY row 0.2.1, line 24: `AC-IFU-029` passed before any work; `REQ-IFU-008` is now declared a **preservation requirement** here and the criterion a **regression guard** with its baseline recorded, AND extended to the new fallback path so it has a red-able half.
```
Why unresolved: the audit's D12 asked the operator to choose. The criterion's RED-now cell is green on the unchanged tree, so it cannot be a RED.
Options: (a) release-blocking; (b) regression guard.
Applied in 0.1.1: (b). Pre-flight item 2 restated both numbers (71 accepted pairs and 73 production table entries).
Applied in 0.2.0: not applicable. AC-FRR-016 is the count pin for the three rows of item (7) (REQ-FRR-008 to -010), which are retired. The test `TestFR_AC005_TransitionTableEdgeCount` stays in the code as a regression test. If item (3) adds a pushed-to-done edge, the count pin is re-authored in that revision.
Basis: `verification-completeness.md` §2 (the vacuous direction). The leader's ruling d-20261010T010044Z-57f8 is cited as an evidence note only.
Evidence: plan-audit-iter1 E5 and D12. `internal/homestate/fr_transition_test.go:146-166` (the pin) and `:165` (the two counts). `internal/homestate/card_transition.go:129-130` (the `@MX:ANCHOR` reason).
Default: none (0.1.1 default was regression guard).
Operator verdict: not required (not applicable, d-20261010T053435Z-006f)

### Q3: Watchdog cadence and the lease gap between passes (verdict D15; plan D4; lane update item 3)

Label: FOUNDER
Class: implementation-level
Why: this SPEC changes neither the cadence nor the default duration; the observed gap is evidence for the TTL design, which this SPEC owns.
Status: DEFAULT-APPLIED
Authority anchor: none routed. Search, run before the DEFAULT route was applied: the version-led HISTORY and Amendments rows of the 840 completed SPEC files were searched for cadence, awaken, heartbeat, renew, watchdog, lease-duration, and 15- and 20-minute terms. No row decides the cadence-versus-TTL question. The only lease-renewal rows in SPEC-FACTORY-ATOMIC-LEASE-001 concern the withCardTx lock, not the cadence or the duration.
Why unresolved: a renewal at the start of each watchdog pass keeps the lease alive only while consecutive passes start within the lease duration. The standing cadence is 20 minutes and the lease is 15 minutes, so the gap between passes is not covered.
Options (plan D4, reduced to the lane update's three): (a) change the cadence, for example to at most half the lease duration, a cost trade-off recorded in auto-semantics; (b) change the default lease duration, which alters the behaviour SPEC-FACTORY-ATOMIC-LEASE-001 defines; (c) keep both and carry the gap.
Applied: (c). The cadence and the default duration do not change. REQ-FRR-005 keeps the guarantee statement, and spec.md Residual risk records the gap as evidence for the TTL design that this SPEC owns.
Ownership: this SPEC (card t1621) owns the lease renewal verb and the TTL configuration key. Card t1618 references this SPEC for them and does not own them; board d-20261010T012900Z-35d1 states that a dispatched t1618 names t1621 as the owner.
Evidence: `.claude/rules/moai/workflow/auto-semantics.md:108` (`7,27,47 * * * *`, every 20 minutes). `internal/config/defaults.go:65` (15 minutes). `internal/homestate/card_record.go:91-93`. The stage path also renews the lease after each successful transition (`internal/cli/factory_card.go:1737`), which narrows the gap for a lane that transitions between passes. That is a code reading, not a run. The leader's ruling d-20261010T010044Z-57f8 (point 3) and d-20261010T012900Z-35d1 are cited as evidence notes only.
Note on routing: ruling 57f8, point (3), as written, routes the watchdog-versus-TTL mismatch to t1618 (lease renewal). Its later correction, d-20261010T012900Z-35d1, withdraws the split and states that only the watchdog-cycle versus lease-TTL mismatch observation stays routed as evidence to whichever card implements the TTL (t1621). This row follows 35d1.
Default: no change to the doctrine cadence (`.claude/rules/moai/workflow/auto-semantics.md:108`) or to the default lease duration in this SPEC. The 20-minute cycle versus 15-minute lease observation is recorded here as evidence for the TTL design.
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:32:07Z manager-spec

### Q4: What returns a card from sync audit to run, and what does the factory record check? (verdict D2, D3; plan D3)

Label: FOUNDER
Class: product-level (the stage verb gains a sync-audit to run transition)
Status: not applicable (retired by d-20261010T053435Z-006f: the sync-audit to run edge is old item (7) of the card body; REQ-FRR-008 and -009 are retired)
Authority anchor: none for the mechanism. The standing rule d-20261009T235640Z-3653 records the policy that "defects in a card's own scope are repaired in that card". It does not say how own scope is told apart, and its digest cannot be pinned in this round, so it is cited for context and not relied on for routing.
Why unresolved: the state machine has no sync-audit to run edge (`internal/homestate/card_transition.go:133-181`), and the verdict parser has no own-scope field (`internal/auditverdict/verdict.go:115-162`). A verdict-based discriminator would either cross t1618's boundary or admit defects outside the card's scope.
Options (plan D3): (a) defer REQ-FRR-008 until the verdict format carries an own-scope field (t1618); (b) add a field to the verdict format in this batch, which crosses t1618's boundary; (c) treat any FAIL with a blocking count of one or more as own-scope, which admits defects outside the card's scope; (d) a leader-recorded decision of kind ruling with scope card:<card-id>, which cites the sync-audit verdict file path and that verdict's audited SHA, checked against the card's sync evidence hash.
Applied in 0.1.1: (d). REQ-FRR-008 and REQ-FRR-009 with three reason tokens.
Applied in 0.2.0: none. The leader's assignment for this revision is 1a19 (d), but the subject is retired (item (7)). See Routing item 3; the retired text, including the `ruling-not-leader` repair that plan audit iteration 2 required (B3), is in git history at `53ea27f4f`.
Rationale (lane, 0.1.1): option (d) needs no change to the verdict format and no field the parser lacks.
Residual (retired with the row): a ruling that misjudges scope is not detectable by the factory record. The judgment would have stayed with the leader.
Evidence: `internal/auditverdict/verdict.go:115-162`. `internal/homestate/card_evidence_readers.go:21` and `:119-148`. `internal/homestate/card_transition.go:564` and `:573`. `internal/homestate/card_transition.go:133-181` (no sync-audit to run row).
Default: none.
Operator verdict: not required (not applicable, d-20261010T053435Z-006f). The 1a19 value (d) is not written; see Routing item 3.

### Q5: Does the holder's re-stamp refuse when the cited verdict's audited SHA differs from the new hash? (verdict D4; plan M4 item 2)

Label: FOUNDER
Class: product-level (adds a refusal to the stage verb's re-audit entry)
Status: not applicable (retired by d-20261010T053435Z-006f: the holder re-stamp is old item (7) of the card body; REQ-FRR-010 is retired)
Authority anchor: none.
Why unresolved: AC-FRR-010 and M4 assumed a refusal that REQ-FRR-010 did not state.
Options: (a) add the refusal to REQ-FRR-010; (b) remove the mismatch case from AC-FRR-010 and M4.
Applied in 0.1.1: (a). REQ-FRR-010 refuses and leaves the card unchanged when the cited verdict's audited SHA differs from the new evidence hash.
Applied in 0.2.0: none. See Routing item 3. The retired text, with the `verifyAuditEntry` repair (B4), is in git history at `53ea27f4f`.
Rationale (lane, 0.1.1): the current verdict guard reads the verdict against the stored hash (`card_transition.go:564`). A re-stamp that accepted any new hash would let the card's evidence hash differ from the SHA that the verdict audited.
Evidence: `internal/homestate/card_transition.go:564`, `:573`.
Default: none.
Operator verdict: not required (not applicable, d-20261010T053435Z-006f). The 1a19 value (a) is not written; see Routing item 3.

### Q6: Which operator carrier, if any, does the empty-stage to run edge admit? (verdict D6; plan D2)

Label: FOUNDER
Class: product-level (the empty-stage to run refusal is user-visible)
Status: not applicable (retired by d-20261010T053435Z-006f: the interim empty-stage guard is the stopgap named by old item (4) of the card body; REQ-FRR-012 is retired)
Authority anchor: none.
Why unresolved: the audit found that a recorded operator decision was allowed to admit the transition, while the edge carries no operator carrier.
Options (plan D2): (a) add an operator-decided edge at T4b in this batch; (b) keep an operator branch inert until a carrier lands in a later SPEC; (c) limit R6 to the plan-verdict branch and write no operator branch.
Applied in 0.1.1: (c). REQ-FRR-012 refused unless an admitted plan verdict is recorded.
Applied in 0.2.0: none. See Routing item 3. The leader's assignment is 1a19 (c). Plan audit iteration 2 (E4, B2) found that the reused predicate `admitCardVerdict` selects the sync-phase admission for any state other than CardPlanAudit (`card_evidence_readers.go:169-173`); the plan-phase admission is `admitVerdictFile(cur, path, auditverdict.PhasePlan)`. That finding returns with the retired text if the row survives.
Rationale (lane, 0.1.1): T4b carries `guardStageResume`, which reads no operator decision (`card_transition.go:546-549`; `stageAdmits` at `:751-756`). An inert branch would be an always-true allow in disguise.
Evidence: `internal/homestate/card_transition.go:138`, `:546-549`, `:751-756`. `internal/homestate/card_evidence_readers.go:169-173`, `:176`.
Default: none.
Operator verdict: not required (not applicable, d-20261010T053435Z-006f). The 1a19 value (c) is not written; see Routing item 3.

### Q7: Does the interim guard cover every non-empty SPEC id, valid or invalid? (verdict D5)

Label: FOUNDER
Class: product-level (a card whose SPEC id is malformed needs a plan verdict to reach run)
Status: not applicable (retired by d-20261010T053435Z-006f: the interim guard is old item (4) of the card body; REQ-FRR-012 is retired)
Authority anchor: none.
Why unresolved: AC-FRR-012 covered only valid ids, while REQ-FRR-012 says non-empty. A non-empty malformed id fell under no criterion.
Options: (a) every non-empty SPEC id, valid or invalid, as REQ-FRR-012 is written; (b) valid ids only, which would leave malformed ids on the empty-stage route.
Applied in 0.1.1: (a).
Applied in 0.2.0: none. See Routing item 3.
Rationale (lane, 0.1.1): the requirement already says non-empty. A malformed id is refused at the normal write path (`internal/homestate/card_picked.go:88-89`), so a row can carry one only through another route, and that route must not bypass the guard. Validity is defined by `specIDPattern` (`card_picked.go:15`).
Evidence: `internal/homestate/card_picked.go:15`, `:88-89`. `internal/homestate/card_evidence_readers.go:185-186`.
Default: none.
Operator verdict: not required (not applicable, d-20261010T053435Z-006f). The 1a19 value (a) is not written; see Routing item 3.

### Q8: Which recovery command does the refusal name? (verdict D14; plan D6)

Label: FOUNDER
Class: product-level (the refusal text is user-visible)
Status: operator verdict recorded (1a19)
Authority anchor: none. The operator's answer is the authority for this row; its board record is cited in the verdict line below and is not digest-pinned.
Why unresolved: the refusal must name a command the holder can run, and the audit's D14 asked for the command to be named.
Options (plan D6): (a) `moai factory stage`, reachable from a card worktree; (b) `moai factory next --card`, parent checkout only; (c) both, with reachability stated.
Applied: (a). REQ-FRR-007 names the stage verb, and AC-FRR-007 checks the name.
Rationale (lane): a lane session starts and stays in its card worktree (`.claude/rules/local/gitflow-lane-protocol.md` §1). The next verb refuses to run outside the parent checkout (`internal/cli/factory_card.go:189-201`), so a refusal that named it would name a verb the holder cannot run from its worktree. The stage verb has no parent-checkout check (`factoryStageCard`, `internal/cli/factory_card.go:1702-1741`) and applies the one-step recovery of REQ-FRR-006. The next verb still recovers from the parent checkout; the refusal does not name it.
Evidence: as above.
Default: none.
Operator verdict: (a) `moai factory stage`. Cited: d-20261010T053034Z-1a19 (operator ceiling-exception; board record, not digest-pinned, so not DECIDED).

### Q9: Is the holder's own renewal of an expired lease refused, or recovered? (verdict D14; plan D7)

Label: FOUNDER
Class: product-level (the renewal verb's behaviour on an expired lease is user-visible)
Status: operator verdict recorded (1a19)
Authority anchor: none. The operator's answer is the authority for this row; its board record is cited in the verdict line below and is not digest-pinned.
Why unresolved: plan D7 left the choice open, and the spec refused without saying why.
Options: (a) refuse with no write, as REQ-FRR-004 is written; (b) recover in one step, as REQ-FRR-006 does for transitions.
Applied: (a). The refusal names the stage verb (REQ-FRR-007).
Rationale (lane): a renewal is a heartbeat and does not change the card's state (the `RenewLease` doc, `internal/homestate/card_transition.go:437-440`). Recovering inside it would make the heartbeat verb perform a re-lease, which is the act the stage verb already performs. Today the renewal reaps first and refuses after the write (`card_transition.go:452-457`).
Evidence: `internal/homestate/card_transition.go:437-440`, `:452-457`.
Default: none.
Operator verdict: (a) refuse with no write. Cited: d-20261010T053034Z-1a19 (operator ceiling-exception; board record, not digest-pinned, so not DECIDED).

### Q10: Worktree path: which source, and what timing, at lease acquisition? (verdict D8; plan D5)

Label: FOUNDER
Class: product-level (a lease acquisition now refuses when no path can be recorded, and refuses a different path; both change a shipped command's behaviour)
Status: operator verdict recorded (1a19)
Authority anchor: none. The operator's answer is the authority for this row; its board record is cited in the verdict line below and is not digest-pinned.
Why unresolved: REQ-FRR-011 did not name the source of the path or the moment it is recorded, and the lease can complete without one (decision record `d-20261009T182157Z-8ccf`).
Options (plan D5): source: (i) a lane-supplied path, checked against the path recorded at picked; (ii) the card worktree the factory materializes, recorded by the lease's record step; (iii) the path recorded at picked alone. Timing: lease acquisition (T3) or assignment (T2).
Applied: source (ii). Timing T3, the lease acquisition, with a later different path refused.
Rationale (lane): (i) is unavailable through the MCP form. The MCP handler derives the path (`internal/cli/mcp_factory_card.go:172-176`), and the next verb takes no path argument and runs only from the parent checkout (`internal/cli/factory_card.go:189-201`). (iii) records a caller-supplied absolute path (`internal/homestate/card_picked.go:91-92`, `:115`) that the factory never checks against the materialized tree. (ii) is what the factory already records, so the repair moves that write into the acquisition, and a leased card can no longer exist without a path.
Applied in 0.2.0: REQ-FRR-011 keeps this source and timing, and its B5 repair chooses option (i): materialize, then record and lease in one transaction (plan §F, M1 item 5). Q10 survives because REQ-FRR-011 also binds the re-lease of item (2).
Evidence (file:line): the record step `internal/cli/factory_card.go:569-602` (`factoryEnsureCardWorktree`, "the REQ-SD-011 record step of a successful lease"), recording through `RecordCardWorktree` at `:598`. Its callers: `internal/cli/factory_card.go:1590-1600` (next: lease, then record), `internal/cli/mcp_factory_card.go:172-176`, and `internal/cli/todo_auto_lane.go:113`. The tree the dispatch names: `.claude/rules/moai/workflow/factory-dispatch.md:111` (`wt: .moai/worktrees/<name>`) and `:117` (`wt` names the new card's worktree). The record refuses a different recorded path: `internal/homestate/card_worktree.go:42-43`, and sets it at `:46`. The lease writes no path: `internal/homestate/card_transition.go:540-545`.
Carried bindings (`internal/cli/factory_card.go:989`, card t1521, `internal/homestate/card_worktree.go:64-90`) pass through the same `RecordCardWorktree` and keep their behaviour.
Default: none.
Operator verdict: (ii) factory-created card worktree, timing T3 (lease acquisition). Cited: d-20261010T053034Z-1a19 (operator ceiling-exception; board record, not digest-pinned, so not DECIDED).

### Q11: Is decision-index.md written in this round? (verdict D16; plan D10)

Label: POLICY-COVERED
Class: implementation-level
Status: not required (POLICY-COVERED)
Authority anchor: `.moai/config/sections/interview.yaml` § `decision_gate: on` (line 6), an operator setting that makes the index owed alongside the tier artifacts.
Why unresolved: nothing, under the setting. The question is whether the index is owed now.
Applied: written in revision 0.1.1, and rewritten in revision 0.2.0 with the rows the narrowing and the operator verdicts require.
Operator verdict: not required (POLICY-COVERED)

### Q12: Is progress.md created with the minimal §E skeleton? (verdict D16; plan D10)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none. The plan-phase skeleton is the author's rule, and the authority register does not list the agent rules.
Why unresolved: the plan's text before this round deferred the file. The choice changes how the era classifier reads this SPEC.
Options: (a) create `progress.md` as the minimal plan-phase skeleton: the four §E headings in order, each with one placeholder line and nothing else (applied); (b) defer it and state the reason in plan.md.
Applied: (a).
Rationale (lane): the skeleton is what the era classifier reads for a plan-phase SPEC (`.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map), and the sibling SPEC `.moai/specs/SPEC-FACTORY-RECORD-001/progress.md` carries the same four headings (lines 3, 31, 182, and 205). Deferral would make this modern SPEC read as V2.x through H-1 (`.claude/rules/local/lifecycle-sync-gate.md`, heuristic table), which is grandfather-protected.
Also recorded here: the ordering of the Definition of Done (verdict D16). Decisions are recorded in this index before the plan re-audit runs. The Phase 1 Plan Audit Gate runs at the start of run (`.claude/rules/moai/workflow/spec-workflow.md`, Phase 1), so a decision surfaced later is recorded before that gate runs.
Evidence: `.claude/rules/local/lifecycle-sync-gate.md` (H-1). `.moai/specs/SPEC-FACTORY-RECORD-001/progress.md:3,31,182,205`.
Default: the placeholder §E skeleton that manager-spec.md specifies: the four headings in order, each with one placeholder line, and nothing else.
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:32:07Z manager-spec

### Q13: What is the residual route for a card with an empty SPEC id? (plan D8)

Label: FOUNDER
Class: implementation-level
Status: not applicable (retired by d-20261010T053435Z-006f: the route was kept unchanged by REQ-FRR-013, which is retired with the interim guard of old item (4))
Authority anchor: none.
Why unresolved: a card with an empty SPEC id can still reach run through the empty-stage edge, because REQ-FRR-013 keeps that route unchanged.
Options: (a) accept the residual as spec.md states it; (b) move the narrow part of lens-4 defect 5 (setting the SPEC id while leased) into this SPEC.
Default (0.1.1): (a). Rule: the option that preserves current behaviour; the route is unchanged.
Applied in 0.2.0: none. The residual is carried by the Residual risk of spec.md (lens-2 D1 stays open) and is not a separate decision.
Alternate: (b).
Operator verdict: not required (not applicable, d-20261010T053435Z-006f)

### Q14: Are lens-4 defects 10, 13, 14, 19, 20, 22, 24, 25, and 26 assigned to this SPEC? (plan D9)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none.
Why unresolved: the scope of the batch is a plan choice.
Options: (a) leave them unassigned, as spec.md states; (b) assign them to this SPEC.
Default: (a). Rule: the option that preserves current behaviour; scope is unchanged.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

### Q15: Which numbering is used for the defects? (plan D1)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none.
Why unresolved: the card brief and the lens-4 table number the defects differently.
Options: (a) lens-4 numbering, as spec.md and plan.md §B use; (b) the card brief's numbering.
Default: (a). Rule: the option that preserves current behaviour; the SPEC already uses it.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

### Q16: What does the `module` field name? (plan D11)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none.
Why unresolved: the write set touches `internal/cli`, `internal/config`, `internal/mcp`, and the templates, while the brief names `internal/homestate`.
Options: (a) keep `internal/homestate`, as the brief specifies; (b) list the write set in the field.
Default: (a). Rule: the option that preserves current value.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

### Q17: What priority does the SPEC carry? (plan D12)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none.
Why unresolved: the priority is a planning label that the operator may read differently.
Options: (a) P0, as the brief specifies; the batch carries lens-4 defect 1 and card body item (3), both marked P0; (b) a lower priority.
Default: (a). Rule: the option that preserves current value.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

### Q18: What is the configuration key name? (plan D13)

Label: FOUNDER
Class: product-level (the key is added to the template's `workflow.yaml`, which changes a template default)
Status: operator verdict recorded (1a19)
Authority anchor: none. The operator's answer is the authority for this row; its board record is cited in the verdict line below and is not digest-pinned.
Why unresolved: the name is a naming choice that a template default will carry.
Options: (a) keep `workflow.factory_lease.duration`; (b) another name in the `workflow.*` namespace.
Applied: (a).
Why no Default: a name has no option that preserves current behaviour, and the published Default rule does not rank names.
Operator verdict: (a) keep `workflow.factory_lease.duration`. Cited: d-20261010T053034Z-1a19 (operator ceiling-exception; board record, not digest-pinned, so not DECIDED).

### Q19: Which repository-local overrides must a template sync keep, and how far does QG-4 reach? (verdict D17)

Label: POLICY-COVERED
Class: implementation-level
Status: not required (POLICY-COVERED)
Authority anchor: `.moai/config/sections/workflow.yaml`, the operator settings in the rendered file: `branch_guard.enabled: true`, `branch_guard.deny_commits_on: [main]` (with its SPEC-MAIN-COMMIT-BAN-001 comment), and `commit_identity_guard.enabled: true`. The template (`internal/template/templates/.moai/config/sections/workflow.yaml`) ships `branch_guard.enabled: false`, `branch_guard.deny_commits_on: []`, and `commit_identity_guard.enabled: false`.
Why unresolved (verdict D17): QG-4 compared the rendered and template files byte for byte, so a sync that satisfied the check would overwrite these overrides and remove the protection on `main`.
Applied: QG-4 is scoped to the keys and files this SPEC changes, and names the overrides that must survive a sync (acceptance.md, Quality gates).
Operator verdict: not required (POLICY-COVERED)

### Q21: Does pushed-to-done (card body item 3) belong to this card, who owns the done-transition opening, and what does it contain? (card body re-scope; d-20261010T053435Z-006f)

Label: FOUNDER
Class: product-level (opens a new edge into `done` for a repository with a remote, which changes a card-state transition's user-visible behaviour)
Status: open
Authority anchor: none. The ruling d-20261010T053435Z-006f names item (3) for this card. The SPEC's Adjacent-cards note, before revision 0.2.0, named t1616 as the owner of the done-transition opening, and the visible text of the t1616 queue entry at read time covers the local main integration flow. Neither is a DECIDED anchor: a board record carries no digest.
Why unresolved: item (3) has two parts, and the SPEC carries no requirement for either. The first is a pushed → done edge. The transition table has no such row: T17 is merged-local → pushed, T18 is merged-local → done (refused when a remote exists), and T20 is ci-green → done, with the reserved pushed → ci-green edge at `card_transition.go:203-208`. The second is out-of-band write detection, which the card body names as `factory doctor`. No existing text in the SPEC or in `.moai/reports/t1621` defines the detection signal. Writing either part would invent the guard and the signal.
Options: (a) this card carries both parts, and a later revision writes the requirements and the guard. (b) t1616 keeps the done-transition opening, and this card carries only the out-of-band write detection. (c) the leader removes item (3) from this card and re-scopes it.
Applied: none. Open.
Basis: the card body item (3) and d-20261010T053435Z-006f for (a); the earlier Adjacent-cards note of this SPEC for (b); the leader's authority to re-scope the queue body for (c).
Evidence: card body t1621 item (3), read with `moai todo show t1621`; `card_transition.go:156-158` and `:203-208`; spec.md Scope and Exclusions (revision 0.2.0).
Default: none (no option is ranked by the published Default rule).
Operator verdict:

## Defect disposition (verdict D1 to D28)

"Changed by update" marks the defects whose disposition the lane updates changed in revision 0.1.1. The update replaced the operator items O1, O2, and O3, which were the dispositions of D1, D12, and D15. D16 is marked because the updates changed how the decision rows are routed. Revision 0.2.0: the rows that name REQ-FRR-008 to -010, -012, -013 or AC-FRR-008 to -010, -012, -013, -016 describe text retired by `d-20261010T053435Z-006f`; they are kept as the history of those dispositions.

| Defect | Disposition (file and section) | Row | Changed by update |
|---|---|---|---|
| D1 (critical) | acceptance.md Adoption status: plan-time regression-guard classification; the release gate was removed in 0.2.0 by the operator verdict (1a19, B1) | Q1 (operator verdict 1a19) | **Yes** (replaces O1; the 0.2.0 verdict removes the run-phase release gate) |
| D2 | spec.md REQ-FRR-008 rewritten around a card-scoped leader ruling; acceptance.md AC-FRR-008 rewritten; verdict-content Given removed. Retired in 0.2.0 | Q4 (not applicable) | No |
| D3 | spec.md REQ-FRR-009 with three reason tokens; AC-FRR-009 asserts the tokens. Retired in 0.2.0 | Q4 (not applicable) | No |
| D4 | spec.md REQ-FRR-010 adds the refusal on audited-SHA mismatch; AC-FRR-010 keeps the case with its own test. Retired in 0.2.0 | Q5 (not applicable) | No |
| D5 | acceptance.md AC-FRR-012 Given widened to every non-empty SPEC id, valid or invalid. Retired in 0.2.0 | Q7 (not applicable) | No |
| D6 | Operator branch removed from REQ-FRR-012, AC-FRR-012, and plan M5. Retired in 0.2.0 | Q6 (not applicable) | No |
| D7 | spec.md REQ-FRR-013 reworded as the brief specifies. Retired in 0.2.0 | none (wording) | No |
| D8 | spec.md REQ-FRR-011 and AC-FRR-011 name the source and the timing; plan §F "Closed before M1" | Q10 (operator verdict 1a19) | No |
| D9 | AC-FRR-015 adds the executable 20-minute case, named `TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration` | none (mechanical) | No |
| D10 | Placeholders replaced with `TestBacklogClaim_ReclaimsExpired`, `TestBacklogClaim_RenewAfterExpiry`, and `TestFR_AC005_TransitionTableEdgeCount`; confirmed with `go test -list` (plan §C item 4) | none (mechanical) | No |
| D11 | Every Verification line: `-run '^Name$' -v -count=1`, the pass rule, and a stops-firing sentence (acceptance.md, Verification convention) | none (format) | No |
| D12 | AC-FRR-016 is a regression guard; citation as support for REQ-FRR-008 and REQ-FRR-010 removed; four elements recorded. Retired in 0.2.0 | Q2 (not applicable) | **Yes** (replaces O2) |
| D13 | spec.md REQ-FRR-014 extended to live and expired leases, covering renewal, recovery, and transition; AC-FRR-014 adds the CLI and MCP checks | none (placement is editorial) | No |
| D14 | spec.md REQ-FRR-007 names `moai factory stage` and covers the holder's refused renewal; AC-FRR-007 checks both | Q8, Q9 (operator verdicts 1a19) | No |
| D15 | spec.md REQ-FRR-005 keeps the guarantee statement and records the gap as evidence for the TTL design this SPEC owns | Q3 (DEFAULT-APPLIED); the round-2 split is closed under "Closed without a decision" | **Yes** (replaces O3; ownership corrected) |
| D16 | spec.md and acceptance.md Definition of Done reworded; this index written; progress.md created as the minimal skeleton | Q11, Q12 | **Yes** (rows routed by the updates; no operator-required status) |
| D17 | acceptance.md QG-4 scoped to the changed keys and files, with the overrides named | Q19 | No |
| D18 | spec.md REQ-FRR-006 defines "where that verb is reachable" as run from the parent checkout | none | No |
| D19 | Not added (see Optional, deferred) | deferred | No |
| D20 | spec.md Residual risk states that AC-FRR-005 checks procedure text, not runtime behaviour | none | No |
| D21 | **Not applied.** The verdict's line claims are wrong at HEAD 35d242f34. T4b is at `internal/homestate/card_transition.go:138`, and the 71-count reason is at `:130`; `:129` is the ANCHOR line. The plan's citations were already correct | none | No |
| D22 | plan §C and §B restate both numbers as two quantities: 71 accepted pairs and 73 table entries. Retired with AC-FRR-016 in 0.2.0 | Q2 (not applicable) | No |
| D23 | spec.md "## Scope" added before Background | none | No |
| D24 | plan §F M4 names the Q4 gate and the lane decision. Retired in 0.2.0 | Q4 (not applicable) | No |
| D25 | spec.md REQ-FRR-005 states the one write (the lease heartbeat through the renew verb) | none | No |
| D26 | AC-FRR-001 states "exactly 5 minutes left" | none | No |
| D27 | acceptance.md QG-5 records the resolved HEAD SHA beside its output | none | No |
| D28 | AC-FRR-013 states that the guard test is written and observed green before M5 starts. Retired in 0.2.0 | Q13 (not applicable) | No |

## Optional, deferred

- **D19: a non-positive or unparseable lease duration is refused at config load.** Not added as a criterion. The Tier M criterion ceiling (16) is no longer full in revision 0.2.0, so the check can be added as its own criterion once the operator or the lane decides the behaviour. Revision 0.2.0 has not decided that behaviour.
- No other optional item is deferred.

## Closed without a decision

Q20 (move the renewal verb and TTL key to t1618) was proposed by the leader and withdrawn by board d-20261010T012900Z-35d1. Scope is unchanged from the card text. No operator question is needed. If the split is ever proposed again, the operator rule d-20261010T001149Z-7d0c applies.

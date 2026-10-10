# Decision index: SPEC-FACTORY-RECORD-REPAIR-001

Card t1621 · Tier M · version 0.1.1 (repair round 1 after plan audit iteration 1, FAIL 0.56, 17 blocking defects).

This index records every decision the SPEC carries, rows Q1 to Q19, and one note on a proposal closed without a decision. Each row names what is unresolved, the options, what the lane applied, the basis, and the evidence. It is stateless on the status axis: the lifecycle lives in spec.md alone.

## How to read a row

- **Label** is one of exactly four values: DECIDED, POLICY-COVERED, EVIDENCE-NEEDED, FOUNDER. DECIDED and POLICY-COVERED need an authority anchor that exists in the committed tree. A row whose anchor cannot be verified is FOUNDER: escalate, never downgrade.
- **Class** is product-level when the decision changes a shipped command's default user-visible behaviour, removes a user-facing feature, or changes a template default. Otherwise it is implementation-level. A row without a Class is product-level.
- **Status** takes one of four values: `open` for an open FOUNDER row (its Operator verdict line is literally empty); `DEFAULT-APPLIED` for a row whose default was applied at plan close; `not required (POLICY-COVERED)` for a POLICY-COVERED row; and `DECIDED` with an authority anchor for a decided row. Status does not route the row. The summary table has no status column.
- **Operator verdict** stays empty on an open FOUNDER row, and an open FOUNDER row blocks the autonomous Kickoff until the founder verdicts it. `DEFAULT-APPLIED` is written only on an implementation-level row whose options the published Default rule ranks, at plan close. It is not written on DECIDED or POLICY-COVERED rows, which need no verdict.

## Routing (this round)

1. **Lane update applied.** Operator items O1, O2, and O3 are settled by the lane, not routed to the operator. Q1, Q2, Q3, and Q12 take the DEFAULT-APPLIED route at plan close, with the time stamped in each row. Q13 to Q17 keep their earlier DEFAULT-APPLIED stamps. No row carries the operator-required status.
2. **Verdict D1 text.** The audit's D1 asked for an operator decision: observed RED cells for the release-blocking criteria, or an operator reclassification. The disposition was taken by rule, with the leader's concurrence recorded in board record d-20261010T010044Z-57f8 (scope card:t1621, kind ruling). That record is cited as an evidence note only.
3. **Authority register note.** A decision-board record anchors a DECIDED row only when cited as `board:<record-id>#<digest-prefix>` and pinned verbatim in a fenced block. The board output shows no digest, so every board record in this index is an evidence note, and no row is DECIDED on a board record. The leader could supply the verbatim line and digest prefix of d-20261010T010044Z-57f8; with them that record could anchor Q1 and Q2 under the register.
4. **Counts and the Kickoff.** Open FOUNDER rows with an empty verdict: Q4 to Q10 and Q18 (eight rows). DEFAULT-APPLIED rows: Q1, Q2, Q3, Q12, and Q13 to Q17 (nine rows). POLICY-COVERED rows, no verdict needed: Q11 and Q19.
5. **Kickoff check, as the code reads it.** `internal/homestate/card_audit_kickoff.go:118-123` refuses a FOUNDER row whose Operator verdict value is literally empty. The eight open rows (Q4 to Q10 and Q18) carry a literally empty verdict, so that check refuses them, and the autonomous Kickoff stays blocked until the founder verdicts each one. `card_audit_kickoff.go:113-117` refuses DEFAULT-APPLIED on any row that is not implementation-level with a Default line; each DEFAULT-APPLIED row here carries both.
6. **Round-2 split withdrawn.** A proposed move of the renewal verb and TTL key to t1618 was withdrawn by board d-20261010T012900Z-35d1. It is recorded under "Closed without a decision", not as a row. This SPEC (card t1621) owns those deliverables.

## Pinned authority (verbatim)

Board record d-20261010T010044Z-57f8, copied from the `moai decision read --all` output in this round. It is quoted, not digest-pinned (Routing, item 3).

```text
decision record: decided_by=leader session 069be28e (factory leader, run tmnboq) evidence_refs=.moai/worktrees/t1621/.moai/reports/t1621/plan-audit-iter1.md; lane-1 report 2026-10-10 ladder_path=leader ruling (verification-completeness.md 2.1; verification-claim-integrity.md 2.3) id=d-20261010T010044Z-57f8 scope=card:t1621 kind=ruling
```

Other records cited in this index are evidence notes and are not pinned: d-20261010T012900Z-35d1 (the leader's withdrawal of the round-2 split, which states that a dispatched t1618 names t1621 as the owner of the renewal verb and TTL), and d-20261010T001149Z-7d0c (the operator rule that a scope split which deletes a deliverable stays with the operator, as 35d1 states). The prior records R3, R4, and R5 (`d-20261009T235640Z-3653`, `d-20261009T180009Z-ae0f`, `d-20261009T182157Z-8ccf`) are pinned in plan.md § Verified premises, item (d). They are cited as context for Q4 and Q10 and are not relied on as authority.

## Summary

| Row | Decision | Defects | Label | Class | Verdict |
|---|---|---|---|---|---|
| Q1 | Release gate for the criteria at plan time | D1 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q2 | AC-FRR-016 classification | D12 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q3 | Watchdog cadence and lease gap (evidence for the TTL design) | D15, plan D4 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q4 | Sync-audit to run: ruling-based discriminator and reason tokens | D2, D3, plan D3 | FOUNDER | product-level | - |
| Q5 | Holder re-stamp refuses on audited-SHA mismatch | D4 | FOUNDER | product-level | - |
| Q6 | Operator carrier for the empty-stage to run edge | D6, plan D2 | FOUNDER | product-level | - |
| Q7 | Interim guard covers every non-empty SPEC id | D5 | FOUNDER | product-level | - |
| Q8 | Recovery command named in the refusal | D14, plan D6 | FOUNDER | product-level | - |
| Q9 | Holder's own expired renewal is refused, not recovered | D14, plan D7 | FOUNDER | product-level | - |
| Q10 | Worktree path: source and timing | D8, plan D5 | FOUNDER | product-level | - |
| Q11 | decision-index.md written in this round | D16, plan D10 | POLICY-COVERED | implementation-level | not required |
| Q12 | progress.md with the minimal §E skeleton | D16, plan D10 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q13 | Residual route for a card with an empty SPEC id | plan D8 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q14 | Lens-4 defects left unassigned | plan D9 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q15 | Brief numbering of the defects | plan D1 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q16 | The `module` field | plan D11 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q17 | The `priority` field | plan D12 | FOUNDER | implementation-level | DEFAULT-APPLIED |
| Q18 | Configuration key name `workflow.factory_lease.duration` | plan D13 | FOUNDER | product-level | - |
| Q19 | Repository-local overrides a template sync must not overwrite; QG-4 scope | D17 | POLICY-COVERED | implementation-level | not required |

## Rows

### Q1: How is the release gate set for the criteria at plan time? (verdict D1; lane update item 1)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none routed. Anchor search, run before the DEFAULT route was applied: the version-led HISTORY and Amendments rows of the 840 completed SPEC files were searched for plan-time RED, regression-guard, undecidable-disposition, vacuous, and green-before-work terms. Candidates were verified at HEAD with `git grep`, each SPEC at `status: completed`. SPEC-DECISION-AUTHORITY-001, HISTORY row 0.1.1, item D3: AC-DA-019 reclassified regression-guard because no behavioural RED is producible at plan time. SPEC-GITHUB-FLOW-CI-RESIDUE-001, HISTORY row 0.3.0, item D5: the test-based RED-now items fall under the §2.1 undecidable disposition. The lane applied the DEFAULT route, so neither candidate is routed DECIDED. The lane can re-route this row if it judges the question identical.
Precedent (noted, not DECIDED):
```text
.moai/specs/SPEC-DECISION-AUTHORITY-001/spec.md, HISTORY row 0.1.1, line 26: D3: AC-DA-019 reclassified regression-guard — no behavioral RED is producible at plan time (design §5 principle).
.moai/specs/SPEC-GITHUB-FLOW-CI-RESIDUE-001/spec.md, HISTORY row 0.3.0, line 30: **D5** — RED-now 원장을 4요소(명령·원문 stdout·종료 코드·SHA) 셀로 재구성(관측 적근거 7건, 테스트 기반 4건은 §2.1 undecidable disposition).
```
Why unresolved: the audit asked for an observed RED for each release-blocking criterion on the unchanged tree, or for an operator decision. While a criterion's test does not exist, no RED can be observed on that tree, so MP-8 fails for every criterion.
Options (verdict D1): (a) RED produced in the run phase, with a stated change to the Tier M plan gate; (b) reclassification of the affected criteria.
Applied: (b). Every criterion is classified regression-guard at plan time, and each criterion states its classification. The RED for each behaviour criterion (AC-FRR-001 to AC-FRR-012 and AC-FRR-014) is captured at run phase in its own commit, made before the fix commit, carrying an observed EXPECTED_RED whose failure is an assertion failure. A compile or discovery failure is TOOL_FAILURE and is never recorded as RED.
Basis: `.claude/rules/moai/development/verification-completeness.md` §2.1 (the undecidable disposition). `.claude/rules/moai/core/verification-claim-integrity.md` §2.3 (the commit graph is the only sequencing witness, so the RED commit precedes the fix commit). `.claude/rules/moai/workflow/tdd-result-contract.md` (EXPECTED_RED and TOOL_FAILURE). The leader's ruling d-20261010T010044Z-57f8 is cited as an evidence note only, not as a DECIDED anchor. Its point (1) states the same disposition: the RED is captured at run phase in its own commit ahead of the fix, and TOOL_FAILURE is never recorded as RED. The leader could supply its verbatim line and a digest prefix; with them the record could be pinned under the register.
Scope note for the re-audit: §2.1 names three cases, historical events, already-merged states, and externally observed CI results. A test that does not exist yet is not one of them. The ruling applies the disposition to it, with the leader's concurrence. The re-audit should check that reading.
Evidence: plan-audit-iter1 D1 and MP-8. plan-audit-iter1 E4: `go test ./internal/homestate/ -run TestRenewRefusesExpiredWithoutWrite -count=1` printed `[no tests to run]` with exit 0. plan-audit-iter1 E5: AC-FRR-016 green on the unchanged tree.
Default: regression-guard at plan time. The RED is captured at run phase in its own commit before the fix. TOOL_FAILURE is never recorded as RED. Basis: the leader's ruling d-20261010T010044Z-57f8, cited as an evidence note only, not as a DECIDED anchor.
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:32:07Z manager-spec

### Q2: Is AC-FRR-016 a release-blocking criterion or a regression guard? (verdict D12; lane update item 2)

Label: FOUNDER
Class: implementation-level
Status: DEFAULT-APPLIED
Authority anchor: none routed. Anchor search, run with the same terms and scope as Q1 before the DEFAULT route was applied. Candidates were verified at HEAD with `git grep`, each SPEC at `status: completed`. SPEC-TODO-HOLD-STATE-001, HISTORY row 0.2.0, item D1: an autodone acceptance criterion green at M1 is pinned as a regression guard, not a live red. SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001, HISTORY row 0.2.1, item D2: AC-IFU-029 passed before any work and is recorded as a regression guard with its baseline. The lane applied the DEFAULT route, so neither candidate is routed DECIDED. The lane can re-route this row if it judges the question identical.
Precedent (noted, not DECIDED):
```text
.moai/specs/SPEC-TODO-HOLD-STATE-001/spec.md, HISTORY row 0.2.0, line 29: **D1**: the autodone filter-direction account was inverted — `todo_autodone.go:281-285` is `if State != Queued && State != Picked { continue }`, so a `hold` card is SKIPPED by the existing filter the moment the state exists; the site produces no behavioral defect and the autodone AC is a green-at-M1 REGRESSION GUARD, not a live red.
.moai/specs/SPEC-LOCAL-INSTRUCTIONS-MIGRATE-001/spec.md, HISTORY row 0.2.1, line 24: `AC-IFU-029` passed before any work; `REQ-IFU-008` is now declared a **preservation requirement** here and the criterion a **regression guard** with its baseline recorded, AND extended to the new fallback path so it has a red-able half.
```
Why unresolved: the audit's D12 asked the operator to choose. The criterion's RED-now cell is green on the unchanged tree, so it cannot be a RED.
Options: (a) release-blocking; (b) regression guard.
Applied: (b). AC-FRR-016 is a regression guard. Its RED-now cell is green on the unchanged tree (plan-audit-iter1 E5: `--- PASS`, exit 0). Its four elements are still recorded in acceptance.md: command, verbatim stdout (as quoted by the audit, not re-run in this pass), exit code, and tree SHA. Pre-flight item 2 re-measures it and restates both numbers: 71 accepted pairs and 73 production table entries.
Basis: `verification-completeness.md` §2 (the vacuous direction). The leader's ruling d-20261010T010044Z-57f8 is cited as an evidence note only. The board record cannot anchor this row without a pinned digest prefix, which the leader could supply.
Evidence: plan-audit-iter1 E5 and D12. `internal/homestate/fr_transition_test.go:146-166` (the pin) and `:165` (the two counts). `internal/homestate/card_transition.go:129-130` (the `@MX:ANCHOR` reason).
Conflict noted: Board 57f8 point (2) states a structural test. This row applies the vacuous-direction rule of verification-completeness.md §2 and records the difference. The leader's confirmation came as a peer message, not a board record, and is not cited as an anchor.
Default: regression-guard. The structural test passes: the command, verbatim stdout, exit code, and tree SHA are present and re-executable. The RED-now cell is green on the unchanged tree (plan-audit-iter1 evidence E5: "--- PASS", exit 0), which verification-completeness.md §2 treats as the vacuous direction. Cite d-20261010T010044Z-57f8 as an evidence note only.
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:32:07Z manager-spec

### Q3: Watchdog cadence and the lease gap between passes (verdict D15; plan D4; lane update item 3)

Label: FOUNDER
Class: implementation-level
Why: this SPEC changes neither the cadence nor the default duration; the observed gap is evidence for the TTL design, which this SPEC owns.
Status: DEFAULT-APPLIED
Authority anchor: none routed. Search, run before the DEFAULT route was applied: the version-led HISTORY and Amendments rows of the 840 completed SPEC files were searched for cadence, awaken, heartbeat, renew, watchdog, lease-duration, and 15- and 20-minute terms. No row decides the cadence-versus-TTL question. The only lease-renewal rows in SPEC-FACTORY-ATOMIC-LEASE-001 concern the withCardTx lock, not the cadence or the duration.
Why unresolved: a renewal at the start of each watchdog pass keeps the lease alive only while consecutive passes start within the lease duration. The standing cadence is 20 minutes and the lease is 15 minutes, so the gap between passes is not covered.
Options (plan D4, reduced to the lane update's three): (a) change the cadence, for example to at most half the lease duration, a cost trade-off recorded in auto-semantics; (b) change the default lease duration, which alters the behaviour SPEC-FACTORY-ATOMIC-LEASE-001 defines; (c) keep both and carry the gap.
Applied: (c). The cadence and the default duration do not change. REQ-FRR-005 keeps the guarantee statement, and spec.md Residual risk records the gap as evidence for the TTL design that this SPEC owns.
Ownership: this SPEC (card t1621) owns the lease renewal verb and the TTL configuration key. Card t1618 references this SPEC for them and does not own them; board d-20261010T012900Z-35d1 states that a dispatched t1618 names t1621 as the owner. The queue (`moai gtd list`, read-only, this round) lists t1618 as picked, item (4), a renewal verb or heartbeat with a 15-minute TTL, and t1621 as picked, item (1), the same renewal verb and TTL key. The queue flags the two as near-duplicates. This lane has not verified t1618's progress.
Evidence: `.claude/rules/moai/workflow/auto-semantics.md:108` (`7,27,47 * * * *`, every 20 minutes). `internal/config/defaults.go:65` (15 minutes). `internal/homestate/card_record.go:91-93`. The stage path also renews the lease after each successful transition (`internal/cli/factory_card.go:1737`), which narrows the gap for a lane that transitions between passes. That is a code reading, not a run. The leader's ruling d-20261010T010044Z-57f8 (point 3) and d-20261010T012900Z-35d1 are cited as evidence notes only.
Note on routing: ruling 57f8, point (3), as written, routes the watchdog-versus-TTL mismatch to t1618 (lease renewal). Its later correction, d-20261010T012900Z-35d1, withdraws the split and states that only the watchdog-cycle versus lease-TTL mismatch observation stays routed as evidence to whichever card implements the TTL (t1621). This row follows 35d1.
Default: no change to the doctrine cadence (`.claude/rules/moai/workflow/auto-semantics.md:108`) or to the default lease duration in this SPEC. The 20-minute cycle versus 15-minute lease observation is recorded here as evidence for the TTL design. Cite board records d-20261010T010044Z-57f8 (point 3) and d-20261010T012900Z-35d1 as evidence notes only.
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:32:07Z manager-spec

### Q4: What returns a card from sync audit to run, and what does the factory record check? (verdict D2, D3; plan D3)

Label: FOUNDER
Class: product-level (the stage verb gains a sync-audit to run transition)
Status: open
Authority anchor: none for the mechanism. The standing rule d-20261009T235640Z-3653 records the policy that "defects in a card's own scope are repaired in that card". It does not say how own scope is told apart, and its digest cannot be pinned in this round, so it is cited for context and not relied on for routing.
Why unresolved: the state machine has no sync-audit to run edge (`internal/homestate/card_transition.go:133-181`), and the verdict parser has no own-scope field (`internal/auditverdict/verdict.go:115-162`). A verdict-based discriminator would either cross t1618's boundary or admit defects outside the card's scope.
Options (plan D3): (a) defer REQ-FRR-008 until the verdict format carries an own-scope field (t1618); (b) add a field to the verdict format in this batch, which crosses t1618's boundary; (c) treat any FAIL with a blocking count of one or more as own-scope, which admits defects outside the card's scope; (d) a leader-recorded decision of kind ruling with scope card:<card-id>, which cites the sync-audit verdict file path and that verdict's audited SHA, checked against the card's sync evidence hash.
Applied: (d). REQ-FRR-008 moves the card back to run only when the requesting lane session cites such a ruling for the same card, the ruling names the verdict path and audited SHA, and the audited SHA equals the card's sync evidence hash. REQ-FRR-009 refuses with three reason tokens: `ruling-missing`, `ruling-scope-mismatch`, and `ruling-sha-mismatch`. The verdict-content Givens were removed from AC-FRR-008 and AC-FRR-009.
Rationale (lane): option (d) needs no change to the verdict format and no field the parser lacks. The factory record reads the ruling and the verdict's `audited_sha:` line through `readAuditVerdict`. The reason tokens make each refusal testable without matching text, so the criterion cannot pass on a bare refusal.
Residual (spec.md): a ruling that misjudges scope is not detectable by the factory record. The judgment stays with the leader.
Evidence: `internal/auditverdict/verdict.go:115-162`. `internal/homestate/card_evidence_readers.go:21` and `:119-148`. `internal/homestate/card_transition.go:564` and `:573`. `internal/homestate/card_transition.go:133-181` (no sync-audit to run row).
Default: none.
Operator verdict:

### Q5: Does the holder's re-stamp refuse when the cited verdict's audited SHA differs from the new hash? (verdict D4; plan M4 item 2)

Label: FOUNDER
Class: product-level (adds a refusal to the stage verb's re-audit entry)
Status: open
Authority anchor: none.
Why unresolved: AC-FRR-010 and M4 assumed a refusal that REQ-FRR-010 did not state.
Options: (a) add the refusal to REQ-FRR-010; (b) remove the mismatch case from AC-FRR-010 and M4.
Applied: (a). REQ-FRR-010 refuses and leaves the card unchanged when the cited verdict's audited SHA differs from the new evidence hash.
Rationale (lane): the current verdict guard reads the verdict against the stored hash (`card_transition.go:564`). A re-stamp that accepted any new hash would let the card's evidence hash differ from the SHA that the verdict audited. Option (b) would reproduce lens-4 defect 7 in another form.
Evidence: `internal/homestate/card_transition.go:564`, `:573`.
Default: none.
Operator verdict:

### Q6: Which operator carrier, if any, does the empty-stage to run edge admit? (verdict D6; plan D2)

Label: FOUNDER
Class: product-level (the empty-stage to run refusal is user-visible)
Status: open
Authority anchor: none.
Why unresolved: the audit found that a recorded operator decision was allowed to admit the transition, while the edge carries no operator carrier.
Options (plan D2): (a) add an operator-decided edge at T4b in this batch; (b) keep an operator branch inert until a carrier lands in a later SPEC; (c) limit R6 to the plan-verdict branch and write no operator branch.
Applied: (c). REQ-FRR-012 refuses unless an admitted plan verdict is recorded, and states that a recorded operator decision alone does not admit the transition. AC-FRR-012 has the third clause. §G forbids an always-true operator allow.
Rationale (lane): T4b carries `guardStageResume`, which reads no operator decision (`card_transition.go:546-549`; `stageAdmits` at `:751-756`). An inert branch would be an always-true allow in disguise. An operator edge would widen this batch beyond the brief.
Evidence: `internal/homestate/card_transition.go:138`, `:546-549`, `:751-756`.
Default: none.
Operator verdict:

### Q7: Does the interim guard cover every non-empty SPEC id, valid or invalid? (verdict D5)

Label: FOUNDER
Class: product-level (a card whose SPEC id is malformed needs a plan verdict to reach run)
Status: open
Authority anchor: none.
Why unresolved: AC-FRR-012 covered only valid ids, while REQ-FRR-012 says non-empty. A non-empty malformed id fell under no criterion.
Options: (a) every non-empty SPEC id, valid or invalid, as REQ-FRR-012 is written; (b) valid ids only, which would leave malformed ids on the empty-stage route.
Applied: (a). The criterion is aligned to the requirement, and AC-FRR-012 has an invalid-id case.
Rationale (lane): the requirement already says non-empty. A malformed id is refused at the normal write path (`internal/homestate/card_picked.go:88-89` refuses an invalid non-empty SPEC id at field-apply), so a row can carry one only through another route, and that route must not bypass the guard. Validity is defined by `specIDPattern` (`card_picked.go:15`).
Evidence: `internal/homestate/card_picked.go:15`, `:88-89`. `internal/homestate/card_evidence_readers.go:185-186`.
Default: none.
Operator verdict:

### Q8: Which recovery command does the refusal name? (verdict D14; plan D6)

Label: FOUNDER
Class: product-level (the refusal text is user-visible)
Status: open
Authority anchor: none.
Why unresolved: the refusal must name a command the holder can run, and the audit's D14 asked for the command to be named.
Options (plan D6): (a) `moai factory stage`, reachable from a card worktree; (b) `moai factory next --card`, parent checkout only; (c) both, with reachability stated.
Applied: (a). REQ-FRR-007 names the stage verb, and AC-FRR-007 checks the name.
Rationale (lane): a lane session starts and stays in its card worktree (`.claude/rules/local/gitflow-lane-protocol.md` §1). The next verb refuses to run outside the parent checkout (`internal/cli/factory_card.go:189-201`), so a refusal that named it would name a verb the holder cannot run from its worktree. The stage verb has no parent-checkout check (`factoryStageCard`, `internal/cli/factory_card.go:1702-1741`) and applies the one-step recovery of REQ-FRR-006. The next verb still recovers from the parent checkout; the refusal does not name it.
Sequence before this SPEC (for the record): the first transition request reaps the expired lease, `moai factory next --card <card-id>` (or the MCP factory_next from the parent checkout) re-leases it, and a same-stage return follows. REQ-FRR-006 replaces that sequence with one step.
Evidence: as above.
Default: none.
Operator verdict:

### Q9: Is the holder's own renewal of an expired lease refused, or recovered? (verdict D14; plan D7)

Label: FOUNDER
Class: product-level (the renewal verb's behaviour on an expired lease is user-visible)
Status: open
Authority anchor: none.
Why unresolved: plan D7 left the choice open, and the spec refused without saying why.
Options: (a) refuse with no write, as REQ-FRR-004 is written; (b) recover in one step, as REQ-FRR-006 does for transitions.
Applied: (a). The refusal names the stage verb (REQ-FRR-007).
Rationale (lane): a renewal is a heartbeat and does not change the card's state (the `RenewLease` doc, `internal/homestate/card_transition.go:437-440`). Recovering inside it would make the heartbeat verb perform a re-lease, which is the act the stage verb already performs. Today the renewal reaps first and refuses after the write (`card_transition.go:452-457`).
Evidence: `internal/homestate/card_transition.go:437-440`, `:452-457`.
Default: none.
Operator verdict:

### Q10: Worktree path: which source, and what timing, at lease acquisition? (verdict D8; plan D5)

Label: FOUNDER
Class: product-level (a lease acquisition now refuses when no path can be recorded, and refuses a different path; both change a shipped command's behaviour)
Status: open
Authority anchor: none.
Why unresolved: REQ-FRR-011 did not name the source of the path or the moment it is recorded, and the lease can complete without one (decision record `d-20261009T182157Z-8ccf`).
Options (plan D5): source: (i) a lane-supplied path, checked against the path recorded at picked; (ii) the card worktree the factory materializes, recorded by the lease's record step; (iii) the path recorded at picked alone. Timing: lease acquisition (T3) or assignment (T2).
Applied: source (ii). Timing T3, the lease acquisition, with a later different path refused.
Rationale (lane): (i) is unavailable through the MCP form. The MCP handler derives the path (`internal/cli/mcp_factory_card.go:172-176`), and the next verb takes no path argument and runs only from the parent checkout (`internal/cli/factory_card.go:189-201`). (iii) records a caller-supplied absolute path (`internal/homestate/card_picked.go:91-92`, `:115`) that the factory never checks against the materialized tree. (ii) is what the factory already records, so the repair moves that write into the acquisition, and a leased card can no longer exist without a path.
Evidence (file:line): the record step `internal/cli/factory_card.go:569-602` (`factoryEnsureCardWorktree`, "the REQ-SD-011 record step of a successful lease"), recording through `RecordCardWorktree` at `:598`. Its callers: `internal/cli/factory_card.go:1590-1600` (next: lease, then record), `internal/cli/mcp_factory_card.go:172-176`, and `internal/cli/todo_auto_lane.go:113`. The tree the dispatch names: `.claude/rules/moai/workflow/factory-dispatch.md:111` (`wt: .moai/worktrees/<name>`) and `:117` (`wt` names the new card's worktree). The record refuses a different recorded path: `internal/homestate/card_worktree.go:42-43`, and sets it at `:46`. The lease writes no path: `internal/homestate/card_transition.go:540-545`.
Mechanism for the lane to confirm: materialize the tree before the lease transaction, so that the transaction can record its path. A failure after materialization leaves a path recorded on an unleased card, which REQ-FRR-011 permits.
Carried bindings (`internal/cli/factory_card.go:989`, card t1521, `internal/homestate/card_worktree.go:64-90`) pass through the same `RecordCardWorktree` and keep their behaviour.
Default: none.
Operator verdict:

### Q11: Is decision-index.md written in this round? (verdict D16; plan D10)

Label: POLICY-COVERED
Class: implementation-level
Status: not required (POLICY-COVERED)
Authority anchor: `.moai/config/sections/interview.yaml` § `decision_gate: on` (line 6), an operator setting that makes the index owed alongside the tier artifacts.
Why unresolved: nothing, under the setting. The question is whether the index is owed now.
Applied: written in this round, before the plan re-audit runs, with the rows the audit's D16 required.
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
Status: DEFAULT-APPLIED
Authority anchor: none.
Why unresolved: a card with an empty SPEC id can still reach run through the empty-stage edge, because REQ-FRR-013 keeps that route unchanged.
Options: (a) accept the residual as spec.md states it; (b) move the narrow part of lens-4 defect 5 (setting the SPEC id while leased) into this SPEC.
Default: (a). Rule: the option that preserves current behaviour; the route is unchanged.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

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
Options: (a) P0, as the brief specifies; the batch carries lens-4 defect 1 and lens-2 D1, both marked P0; (b) a lower priority.
Default: (a). Rule: the option that preserves current value.
Alternate: (b).
Operator verdict: DEFAULT-APPLIED 2026-10-10T01:19:17Z lane t1621 / manager-spec

### Q18: What is the configuration key name? (plan D13)

Label: FOUNDER
Class: product-level (the key is added to the template's `workflow.yaml`, which changes a template default)
Status: open
Authority anchor: none.
Why unresolved: the name is a naming choice that a template default will carry.
Options: (a) keep `workflow.factory_lease.duration`; (b) another name in the `workflow.*` namespace.
Applied: (a).
Why no Default: a name has no option that preserves current behaviour, and the published Default rule does not rank names.
Operator verdict:

### Q19: Which repository-local overrides must a template sync keep, and how far does QG-4 reach? (verdict D17)

Label: POLICY-COVERED
Class: implementation-level
Status: not required (POLICY-COVERED)
Authority anchor: `.moai/config/sections/workflow.yaml`, the operator settings in the rendered file: `branch_guard.enabled: true`, `branch_guard.deny_commits_on: [main]` (with its SPEC-MAIN-COMMIT-BAN-001 comment), and `commit_identity_guard.enabled: true`. The template (`internal/template/templates/.moai/config/sections/workflow.yaml`) ships `branch_guard.enabled: false`, `branch_guard.deny_commits_on: []`, and `commit_identity_guard.enabled: false`.
Why unresolved (verdict D17): QG-4 compared the rendered and template files byte for byte, so a sync that satisfied the check would overwrite these overrides and remove the protection on `main`.
Applied: QG-4 is scoped to the keys and files this SPEC changes, and names the overrides that must survive a sync (acceptance.md, Quality gates).
Operator verdict: not required (POLICY-COVERED)

## Defect disposition (verdict D1 to D28)

"Changed by update" marks the defects whose disposition the lane updates changed. The update replaced the operator items O1, O2, and O3, which were the dispositions of D1, D12, and D15. D16 is marked because the updates changed how the decision rows are routed.

| Defect | Disposition (file and section) | Row | Changed by update |
|---|---|---|---|
| D1 (critical) | acceptance.md Adoption status: plan-time regression-guard classification; run-phase RED commit gates release; TOOL_FAILURE rule; each criterion states its classification | Q1 (DEFAULT-APPLIED) | **Yes** (replaces O1) |
| D2 | spec.md REQ-FRR-008 rewritten around a card-scoped leader ruling; acceptance.md AC-FRR-008 rewritten; verdict-content Given removed | Q4 | No |
| D3 | spec.md REQ-FRR-009 with three reason tokens; AC-FRR-009 asserts the tokens; verdict-content Givens removed; reason text defined | Q4 | No |
| D4 | spec.md REQ-FRR-010 adds the refusal on audited-SHA mismatch; AC-FRR-010 keeps the case with its own test | Q5 | No |
| D5 | acceptance.md AC-FRR-012 Given widened to every non-empty SPEC id, valid or invalid; `specIDPattern` cited at `internal/homestate/card_picked.go:15` | Q7 | No |
| D6 | Operator branch removed from REQ-FRR-012, AC-FRR-012, plan M5, and plan §G; case added: a recorded operator decision alone does not admit | Q6 | No |
| D7 | spec.md REQ-FRR-013 reworded as the brief specifies | none (wording) | No |
| D8 | spec.md REQ-FRR-011 and AC-FRR-011 name the source and the timing; plan §F "Closed before M1" | Q10 | No |
| D9 | AC-FRR-015 adds the executable 20-minute case, named `TestBacklogClaim_QueueLeaseIgnoresCardLeaseDuration`, following the naming pattern of `internal/factory/backlog_claim_test.go` | none (mechanical) | No |
| D10 | Placeholders replaced with `TestBacklogClaim_ReclaimsExpired`, `TestBacklogClaim_RenewAfterExpiry`, and `TestFR_AC005_TransitionTableEdgeCount`; confirmed with `go test -list` (plan §C item 5) | none (mechanical) | No |
| D11 | Every Verification line: `-run '^Name$' -v -count=1`, the pass rule, and a stops-firing sentence (acceptance.md, Verification convention) | none (format) | No |
| D12 | AC-FRR-016 is a regression guard; citation as support for REQ-FRR-008 and REQ-FRR-010 removed; four elements recorded; pre-flight item 2 restated with 71 and 73 | Q2 (DEFAULT-APPLIED) | **Yes** (replaces O2) |
| D13 | spec.md REQ-FRR-014 extended to live and expired leases, covering renewal, recovery, and transition; AC-FRR-014 adds the CLI and MCP checks | none (placement is editorial) | No |
| D14 | spec.md REQ-FRR-007 names `moai factory stage` and covers the holder's refused renewal; AC-FRR-007 checks both | Q8, Q9 | No |
| D15 | spec.md REQ-FRR-005 keeps the guarantee statement and records the gap as evidence for the TTL design this SPEC owns; cadence and default unchanged; residual risk updated; ownership of the renewal verb and TTL key stated in spec.md Scope | Q3 (DEFAULT-APPLIED); the round-2 split is closed under "Closed without a decision" | **Yes** (replaces O3; ownership corrected) |
| D16 | spec.md and acceptance.md Definition of Done reworded (decisions before the plan re-audit; Phase 1 gate at run start); this index written; progress.md created as the minimal skeleton | Q11, Q12 (DEFAULT-APPLIED) | **Yes** (rows routed by the updates; no operator-required status) |
| D17 | acceptance.md QG-4 scoped to the changed keys and files, with the overrides named | Q19 | No |
| D18 | spec.md REQ-FRR-006 defines "where that verb is reachable" as run from the parent checkout | none | No |
| D19 | Not added (see Optional, deferred) | deferred | No |
| D20 | spec.md Residual risk states that AC-FRR-005 checks procedure text, not runtime behaviour | none | No |
| D21 | **Not applied.** The verdict's line claims are wrong at HEAD 35d242f34. T4b is at `internal/homestate/card_transition.go:138`, and the 71-count reason is at `:130`; `:129` is the ANCHOR line. The plan's citations were already correct | none | No |
| D22 | plan §C item 2 and §B restate both numbers as two quantities: 71 accepted pairs and 73 table entries | none | No |
| D23 | spec.md "## Scope" added before Background | none | No |
| D24 | plan §F M4 names the Q4 gate and the lane decision; the stale "blocked on D3" text is gone | Q4 | No |
| D25 | spec.md REQ-FRR-005 states the one write (the lease heartbeat through the renew verb) | none | No |
| D26 | AC-FRR-001 states "exactly 5 minutes left" | none | No |
| D27 | acceptance.md QG-5 records the resolved HEAD SHA beside its output | none | No |
| D28 | AC-FRR-013 states that the guard test is written and observed green before M5 starts | none | No |

## Optional, deferred

- **D19: a non-positive or unparseable lease duration is refused at config load.** Not added as a criterion. The Tier M criterion ceiling (16) is full, so the check would need to fold into AC-FRR-003, and that needs requirement text in REQ-FRR-003 that states the refusal. This round has not decided that behaviour. The next plan revision, or the lane, decides it.
- No other optional item is deferred.

## Closed without a decision

Q20 (move the renewal verb and TTL key to t1618) was proposed by the leader and withdrawn by board d-20261010T012900Z-35d1. Scope is unchanged from the card text. No operator question is needed. If the split is ever proposed again, the operator rule d-20261010T001149Z-7d0c applies.

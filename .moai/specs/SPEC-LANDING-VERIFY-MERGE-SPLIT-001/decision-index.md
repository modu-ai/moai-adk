# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — decision index (card t1872)

Version 0.3.0 adds the leader's policy rulings on plan-audit iteration 1 (board `d-20261018T153534Z-0577`, `0577`) and rows Q14-Q17. Version 0.4.0 adds the N2 ruling on plan-audit iteration 2 (board `d-20261018T160111Z-470c`, `470c`) and rows Q18-Q22. Version 0.11.0 adds the plan-audit A2 round 3 correction record under the leader's rule C ruling (`9e76`) and corrects the default text of row Q28; it adds no row and applies no default to any product-level row. Version 0.10.0 adds the plan-audit A2 round 2 record and its ruling (board `d-20261018T175713Z-7c5c`, `7c5c`) and row Q31. Version 0.9.0 adds the plan-audit A2 round 1 record and the MP-1 id-gap ruling (board `d-20261018T173648Z-9b0e`, `9b0e`), row Q30, and corrects the recorded time of rows Q2-Q6 (to `21b62f7dcc`, 15:04:14Z) and Q28 (to `9e2b56196f`, 17:18:08Z), the commits that set them. Version 0.8.0 adds the ruling on Q27 and Q28 (board `d-20261018T171932Z-6404`, `6404`) and corrects the recorded time of rows Q15-Q17, Q19, Q21, Q23, Q24, Q25 and Q26 to the commit time of the commit that last set each verdict. Version 0.7.0 adds the split ruling (board `d-20261018T170131Z-d20e`, `d20e`, superseding `adbe`) and rows Q27-Q29; the SPEC now covers piece A2 only. Version 0.6.0 adds the single ceiling exception on plan-audit iteration 4 (board `d-20261018T164307Z-4812`, `4812`) and rows Q25-Q26. Version 0.5.0 adds the rule C ruling on plan-audit iteration 3 (board `d-20261018T162457Z-5d4d`, `5d4d`) and rows Q23-Q24, and re-states Q20 and Q22.

`interview.decision_gate` is `on` in this repository. Version 0.2.0 records the leader's scope ruling of 2026-10-18 (board `d-20261018T150139Z-9dbe`, decided_by leader tn1p4e under operator delegation +Jev, operator-relayed scope review): the SPEC narrows to **piece A** (baseline M0, completeness gate, the Q13 binding, and the enable switch). Pieces B and C and the reuse relaxation are follow-ups, not part of this SPEC.

**Provenance note.** The boards named below (`e4cc` = `d-20261018T144629Z-e4cc`, `2c51` = `d-20261018T145532Z-2c51`, `5923` = `d-20261018T143825Z-5923`, `9dbe` = `d-20261018T150139Z-9dbe`) are machine-local and are not committed authority, so no row is routed DECIDED or POLICY-COVERED by routing; every row keeps the label FOUNDER and records the operator's answer, with its source ids, on the `Operator verdict` line. This committed file carries the wording so a reader on another machine has it.

## The ruling, verbatim (board 9dbe)

> t1872 plan shape and Q7-Q13 (leader under operator delegation, +Jev; operator-relayed scope review 2026-10-18 ~15:0xZ). No new exception is granted by this ruling.
> SHAPE: t1872 is narrowed to piece A = M0 baseline + M1 completeness gate (required commands, package completion, attempt, exit, SHA attribution verified automatically; interrupt vs restart distinguished) + the Q13 complete binding. Pieces B (slot dead-holder/capacity) and C (readiness view, metering, reuse) are recorded follow-ups, issued only after A lands (9891); C depends on A. The SPEC id stays SPEC-LANDING-VERIFY-MERGE-SPLIT-001.
> Q7 DECIDED (operator direct request + e4cc): the gate is enabled in this repository when A lands. Leader-set order inside A: (1) structured remeasure form available and lanes told to use it, (2) negative tests (script-wrapped record, missing package, wrong SHA, interrupted attempt -> refused) and positive tests (complete structured record -> admitted) pass, (3) only then the config key turns on, in the same piece.
> Q13 DECIDED (2c51): complete binds the sync-audit verdict to the card's current HEAD, required backend passing in a receipt at that HEAD; no earlier PASS substitutes.
> Q9 MEASUREMENT TASK: capacity stays 1. A load/memory snapshot (16 CPU, load 57, idle 8%) is not evidence for more; capacity 2 only after a measured condition, out of A.
> Q11 MEASUREMENT TASK: the wall-clock package list ships empty; filled only from measurement, out of A.
> Q8 DEFAULT-DENY: no reuse built; shadow counter only, in C. No non-Go glob allowance is guessed.
> Q12 DEFAULT-DENY: no package is trusted as hermetic; no trust list is guessed.
> Q10 MINIMAL: dead-session stale reclaim already exists (measured), so no new liveness meaning and no wrapper verb now; at most record removal for a dead owner session (REQ-LVM-010), in B. Active leases are never taken over.
> Unchanged: no past PASS substitution, no test exclusion, no default-permission expansion. Plan audit (codex required) runs after the narrowed SPEC is committed.

Board `e4cc` (the scope addition that made the completeness gate item 6, ordered first): the gate is a per-card required-command manifest from the b2cd scope; per call an attempt id, command, exit, `go test -json` package completion with no missing package, and tree/SHA binding; `complete` and landing-ready read the whole manifest on the same tree, not the last record; an interrupted attempt never counts; negative tests for a missing manifest or call, partial JSON, interruption, another SHA or tree, unverified environment/tag/non-Go inputs, and an undecidable overlap; a positive control with the full manifest passing on the current tree.

Board `2c51`: see Q13.

## Plan-audit iteration 1 rulings, verbatim (board 0577)

> t1872 piece A, leader-owned policy items raised by plan audit iter1 (FAIL 0.73). Leader under operator delegation.
> P1 Binary used for landing (B5): a gate counts as ON only when the judging build contains it (rule 2.2 tool provenance). Landing commands the gate governs (remeasure, factory complete, merge step) run through a binary built from main at the landing HEAD and invoked by path, never through a PATH-resolved installed build that may lag. No installed-binary reinstall is implied by this ruling. The MCP server that writes audit receipts must be running a build that contains the receipt change before a receipt counts as SHA-bound; the SPEC states this as a precondition, not a step the card performs.
> P2 Switch order (B5): A lands under the current rule; its own landing is not the gate's first real use. The config key turns on in a later, separate commit on main after A is landed and the binary from P1 exists; the first gated landing is the next landing after that commit. Cards already merge-ready need no re-audit for the switch itself: each still takes a HEAD-confirmation audit at landing (2c51), and its landing remeasure uses the structured form once the key is on.
> P3 Post-audit commits (B5, REQ-LVM-008): no evidence or progress backfill lands after the audited SHA; bookkeeping is committed before the final audit, which binds the final HEAD (2c51).
> P4 Receipt SHA (B1): adding the audited commit SHA to NEW receipts is approved as an additive field; existing receipts are not rewritten or deleted, they stay valid records but cannot satisfy the HEAD binding. The receipt writer and schema enter the contact list and the Tier arithmetic.
> P5 Recovery (optional defect): the gate's own revert must be landable — the SPEC names the recovery path (the config key off in one commit, which the gate does not judge), not a bypass of any audit.
> Unchanged: no past-PASS substitution, no test exclusion, no default-permission expansion; the 9dbe shape and Q7-Q13 rulings stand.

Effect on earlier rows: the order inside piece A recorded at Q7 is refined by P2 — the in-repo key is switched on by a separate later commit on main, not by a milestone of this SPEC's branch; piece A's own landing is not the gate's first live use.

## Plan-audit iteration 2 ruling, verbatim (board 470c)

> t1872 piece A, N2 (plan audit iter2 FAIL 0.80). Policy: the gate's enable key is read from the integration branch's committed content (the blob at <integration-branch>:<config path>, the branch named by the common-root integration config — consistent with t1834 F2), not from any working tree. Reason: the primary checkout is detached and stale (478285ab4c) and moving it is the operator's keep-set (branch guard), and binding the consumer to .claude/worktrees/main ties correctness to one local layout. With this source, an enable or revert commit on main takes effect when it lands, which an AC must observe. Options (a) keep the primary checkout on main and (b) read .claude/worktrees/main are rejected for those reasons. iter3 is the last Tier L round; a FAIL there goes to hold + shrink proposal (rule C 9e76).

Effect on earlier text: the "integration root's configuration" wording of version 0.3.0 (REQ-LVM-007) is replaced by the committed blob of the integration branch (REQ-LVM-017).

## Plan-audit iteration 3 ruling, verbatim (board 5d4d)

> t1872 piece A, plan audit iter3 (last Tier L round) FAIL 0.84. Rule C (9e76): one correction + one delta audit. Choice: option 1 (one anchor-delta commit for D1 and D2, absorbing D3-D6 and the minor D7-D9; REQ/AC ids unchanged), not the A1/A2 split. D1: 'equivalent' means equal to HEAD, one definition (consistent with 0577 P3). D2: look the receipt up at its store root and compare the audited tree path and identity, with a config-orphan worktree fixture. D3 policy: no integration branch configured -> the project has not opted in -> gate OFF; an integration branch configured but unresolvable -> the landing is refused with a named error (fail-closed). This does not deadlock: the merge step cannot run without that branch, and recovery is repairing the branch configuration or ref, which needs no landing. A FAIL on the delta goes to hold + the A1/A2 split as the shrink.

Effect on earlier rows: Q22 below is replaced (exact equality, no tolerance); Q20 is narrowed by the D3 policy.

## Plan-audit iteration 4 ceiling exception, verbatim (board 4812)

> t1872 piece A single ceiling exception (leader under operator delegation, Jev advisory 0.72). Rule C delta audit FAIL 0.88 @21687134ba (.moai/reports/t1872/plan-audit-iter4.md; rcpt-6f7d69e28f5d0af045255fdc, rcpt-11c4bae25772c9e7e6754c68) stays on record; D1-D9 all resolved. Exception scope, one time: B1 (an effective-GOFLAGS constraint so a run narrowed through GOFLAGS cannot pass derive/attempt comparison; a negative AC under AC-LVM-005 and a mutation row) and S1 (go.mod/go.sum/go.work changes cannot yield an empty manifest), plus O1-O3 wording; REQ/AC id set unchanged. Then one audit at the new HEAD with codex required pass. A FAIL there ends the exception and the card splits into A1 (gate core) and A2 (audit binding + reader) with new SPEC ids, after A2's independence from A1 is verified in code.

## Split ruling, verbatim (board d20e)

> Supersedes d-20261018T170122Z-adbe (wording fix only: A1's card id is t1875). t1872 split after the 4812 exception ended (iter5 FAIL 0.86 @9713b6757d; codex required fail rcpt-5ccf101ec66f0bd46ee54f10, rcpt-6f3244d2a8df252898424d5b; blockers B1-B3 all in the gate core). A2 = audit binding + switch kernel (REQ-LVM-007, 008, 009, 010, 012, 017, 018) stays on card t1872 under its frozen spec_id SPEC-LANDING-VERIFY-MERGE-SPLIT-001 and lands first; its plan re-verifies independence from A1 in code and starts a fresh plan audit cycle on the reduced scope. A1 = gate core (REQ-LVM-001-006, 013-016 plus B1-B3 closure) is card t1875 with new SPEC-LANDING-COMPLETENESS-GATE-001, blocked until t1872 is merged-local (shared edit sites factoryCompleteCard and factorylane/merge.go). e4cc holds: no reuse relaxation opens before A1 lands.

Effect on earlier rows: rows about the gate core — Q4 (package coverage observed), Q5 (importers always included), Q15 (producer verbs), Q16 (zero-Go card), Q17 (command comparison), Q25 (effective GOFLAGS), Q26 (module files) — stay recorded here as history and are re-adopted by SPEC-LANDING-COMPLETENESS-GATE-001 (card t1875); Q6 and Q8 (reuse) remain not built (board e4cc). Rows Q7, Q13, Q14, Q18-Q24 stay with this SPEC. The 4812 exception is over.

## Ruling on Q27 and Q28, verbatim (board 6404)

> t1872 A2 open questions (leader under operator delegation; e4cc, 0577 P2, 9dbe Q7). Q27: the gate key turns ON only after BOTH A2 (t1872) and A1 (t1875) have landed, in a separate main commit, judged by a build that contains both (0577 P1). Reason: with A2 alone, 'ON' would enforce audit binding on 'moai factory complete' while 'moai integration merge --card' still bypasses it (Q28 residue R6), which is a gate with a known side door. Q28: A2 keeps the binding on complete only (surface minimal, DEFAULT-APPLIED stands) because the key stays OFF until A1 lands; A1 (t1875) adds the binding to the direct merge path (factorylane/merge.go:174/:245, integration_merge_step.go:321-325) as part of its scope, so ON never ships with the side door. A2's plan audit cycle may start now with these two rows filled (Tier L, counted fresh).

Effect on earlier rows: Q7 and the Q27 row below are settled by this ruling (the key is off in this SPEC and on only after both pieces); Q28 keeps its default and gains the A1 inheritance. The recorded time of rows Q25 and Q26 was wrong (written before the clock was read) and is corrected to the commit time of `9713b6757d`, which introduced them.

## Plan-audit A2 round 1 and the MP-1 ruling, verbatim (board 9b0e)

Evidence (local, untracked, not authority): `.moai/reports/t1872/plan-audit-a2.md` (FAIL 0.81 at `edee2983cc`; codex required fail, receipts `rcpt-f90c0a8ce10bb3170ace76fa`, `rcpt-a30ddabf48c08a0ccf4cc5e5`).

> t1872 A2 plan audit cycle 1/3 FAIL 0.81 (.moai/reports/t1872/plan-audit-a2.md; codex required fail rcpt-a30ddabf48c08a0ccf4cc5e5). MP-1 ruling: record the id-gap exemption, do not renumber. The A2 SPEC keeps REQ-LVM-007, 008, 009, 010, 012, 017, 018; the absent ids (001-006, 011, 013-016) moved to card t1875 (A1, SPEC-LANDING-COMPLETENESS-GATE-001) under split d20e and are listed in spec.md §7's hand-over table. Reason: ids are stable and are never renumbered or deleted; renumbering would rewrite the A1 hand-over, the SPEC, and every AC reference. The exemption is written in the SPEC as an explicit clause naming the gap, its reason, and this record. For MP-9 at acceptance.md:16 the characterization test lands in its own commit before the factoryCompleteCard edit (rule 2.3: the commit graph is the only ordering witness), or the AC is rewritten to what the graph can show.

## Plan-audit A2 round 2 ruling, verbatim (board 7c5c)

Evidence (local, untracked, not authority): `.moai/reports/t1872/plan-audit-a2-iter2.md` (FAIL 0.81 at `d50f6bfe5d`; codex required fail, receipts `rcpt-67bd4d28e15e51ae65437aed`, `rcpt-75a95999f6e2a22b41bc9bec`).

> t1872 A2 plan audit 2/3 FAIL 0.81 (.moai/reports/t1872/plan-audit-a2-iter2.md; codex required fail rcpt-75a95999f6e2a22b41bc9bec). Blocker 1 (binding SHA checked before the window; merge step re-pins refs/heads/<cardBranch> at integration_merge_step.go:289) is closed in A2 by the minimal fix inside complete only: re-read refs/heads/<card> immediately before the merge-step call and refuse with a named error when it differs from the bound SHA, plus test and mutation row. The merge step's own pin consuming the bound SHA stays in A1 (t1875) per Q28 (6404); the remaining window is accepted in A2 because the gate key stays OFF until both A1 and A2 land, and is written as residual risk and an A1 hand-over row. Blocker 2: M4 adds the catalog.yaml regeneration and commit. The minor items go into the same revision. This is the last Tier L round (3/3); a FAIL goes to hold + shrink.

## Plan-audit A2 round 3 correction (rule C, `9e76`)

Evidence (local, untracked, not authority): `.moai/reports/t1872/plan-audit-a2-iter3.md` (FAIL 0.81 at `d2e5680b5b03dae474682b2c11e0a5b2a7643c8f`; codex required fail, receipts `rcpt-24906609de0efefa21f27977`, `rcpt-2476daffe8a9d8d997825586`; no must-pass criterion failed, three wording or test-design blockers D1-D3).

Authority for this revision: the leader's ruling under rule C (`9e76`, the one-correction-plus-one-delta-audit rule already recorded above in the iteration 3 ruling of board `5d4d`), as relayed in the delegation of this correction: one correction commit that fixes D1, D2 and D3 only; REQ/AC ids and counts stay 7 and 7. The relayed delegation, not a board record, is the source of this paragraph; it is not quoted as a board ruling. No product-level row is DEFAULT-APPLIED by this correction, and no row was added.

- D1 (cross-layer sweep miss): the Definition of Done, the plan constraint and risk row, the research table cell and the default text of row Q28 now agree with REQ-LVM-008 that the `factoryCompleteCard` edit has three sites; the protected surface is the re-measure record decision and the merge step (spec C9 unchanged).
- D2 (the second re-read was unpinned): one test and one injection point per re-read, the pre-merge test renamed `TestCompleteRefusesWhenCardMovedBeforeMergeStep`, two single-deletion mutant rows, the seam before `RunMergeStep` named as the one allowed edit of step 4. Still exactly two tests for the two re-reads.
- D3 (the catalog guard test named the wrong entry): the regeneration guard is `TestManifestHashFormat` (`CATALOG_HASH_UNSTABLE`), stated as an expectation not run at plan time; `TestClaudeAuditTemplateSurfacesAndCatalogHash` is kept only for what it checks.

## Rows Q1-Q6 (implementation-level, re-judged for piece A)

### Q1: Is the work one SPEC, and how large?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree (the answer is board 9dbe)
Why unresolved: no committed rule divides a multi-part landing repair.
Default: one SPEC, piece A only (rule: undo is a single revert of this SPEC's own commits)
Alternate: the original six-part single SPEC, or three SPECs A/B/C all authored now
Operator verdict: DECIDED 2026-10-18T15:01:39Z by leader under operator delegation, board 9dbe: narrowed to piece A; B and C are follow-ups issued after A lands. (Supersedes the earlier DEFAULT-APPLIED of 14:57:18Z for the six-part scope.)

### Q2: Do new configuration keys ship off?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: no committed rule states the shipped default of a new landing key.
Default: the key ships false in the distributed template (rule: the option that preserves current behaviour); this repository's own configuration is switched on at the last milestone of A, which Q7 decides
Alternate: ship the key true everywhere
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:04:14Z manager-spec+lane t1872 (template default false; the in-repo switch is Q7)

### Q3: Is "ready to land" a derived view or a stored marker?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the card asks for a ready state; the form is open.
Default: not built in A (rule: the smaller user-visible surface); the follow-up piece C is to be derived at read time with no stored marker
Alternate: build a stored marker now
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:04:14Z manager-spec+lane t1872 (OUT-OF-SCOPE for A per board 9dbe; recorded for C)

### Q4: Where does package coverage come from?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: coverage could be lane-declared or observed from the test stream.
Default: observed from the `go test -json` terminal events against a `go list`-resolved expected set (rule: the smaller user-visible surface — no new declaration for lanes to keep)
Alternate: lane-declared package lists
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:04:14Z manager-spec+lane t1872

### Q5: Does a test-only change still pull in the changed package's importers?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (the b2cd text and the t1709 reading are board rulings)
Why unresolved: the standing scope reads "packages the card changed plus the packages importing them" with no test-only exclusion.
Default: importers are always included (rule: the option that preserves current behaviour — the conservative resolution applied at t1709); consistent with the ruling's "no test exclusion"
Alternate: exclude importers when only `_test.go` files changed
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:04:14Z manager-spec+lane t1872

### Q6: Is the reuse decision by content equality of the scope or by path-matching?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: only relevant to the reuse relaxation.
Default: not built (rule: the option that preserves current behaviour)
Alternate: build either predicate
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:04:14Z manager-spec+lane t1872 (OUT-OF-SCOPE for A: reuse is not built, Q8)

## Rows Q7-Q13 (product-level; answered by the operator through the leader)

### Q7: Is the completeness gate enabled in this repository, and in what order?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (sources: the operator's direct request, board e4cc, board 9dbe)
Why unresolved: enabling makes `moai factory complete`, the merge step and merge-readiness refuse any landing without a whole, same-tree manifest pass.
Operator verdict: DECIDED (operator direct request + e4cc, recorded by board 9dbe): once piece A lands, the completeness gate is turned ON in this repository. Order inside A: (1) the structured re-measure form is available and lanes are told to use it; (2) the negative tests (script-wrapped record, missing package, wrong SHA, interrupted attempt are REFUSED) and the positive test (a complete structured record is admitted) pass; (3) only then the configuration key is switched ON. Refined by board 0577 P2: the switch is a later, separate commit on main after piece A has landed and the P1 binary exists; the first landing the gate governs is the next one after it.

### Q8: Is the re-measure reuse relaxation built, and under what non-Go policy?

Label: FOUNDER
Class: product-level
Authority anchor: none (source: board 9dbe)
Why unresolved: measured eligibility is 1 of 37 (strict) to 10 of 37 (lenient) adjacent landings (`research.md` §6).
Operator verdict: REFUSED (default-deny, board 9dbe): reuse is not built; only a shadow instrumentation counter goes into piece C. No allowed-glob list is guessed.

### Q9: Is heavy-slot capacity raised, and with what thresholds?

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none (source: board 9dbe)
Why unresolved: the thresholds need measurements that do not exist.
Operator verdict: OUT-OF-SCOPE for A (measurement task, board 9dbe): capacity stays 1. A load/memory snapshot (16 CPU, load 57, idle 8%) is not evidence for more; capacity 2 only after a measured condition.

### Q10: Slot liveness semantics and a wrapper verb?

Label: FOUNDER
Class: product-level
Authority anchor: none (source: board 9dbe)
Why unresolved: a dead-pid lease is already reclaimable (`research.md` §4).
Operator verdict: MINIMAL, OUT-OF-SCOPE for A (board 9dbe): dead-session stale reclaim already exists, so no new liveness meaning and no wrapper verb; at most record removal for a dead owner session goes into piece B. An active lease is never taken over.

### Q11: Which packages are wall-clock-bound?

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none (source: board 9dbe)
Why unresolved: a measured fact that is not recorded.
Operator verdict: OUT-OF-SCOPE for A (measurement task, board 9dbe): the list ships empty and is filled only from measurement.

### Q12: Who vouches that a package is hermetic?

Label: FOUNDER
Class: product-level
Authority anchor: none (source: board 9dbe)
Why unresolved: only relevant to the reuse relaxation.
Operator verdict: REFUSED (default-deny, board 9dbe): no package is trusted as hermetic and no trust list is guessed.

### Q13: Does `complete` bind the sync-audit verdict to the card's current HEAD?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (source: board 2c51 `d-20261018T145532Z-2c51`, recorded by board 9dbe)
Why unresolved: today only `sync_status: complete` is read for SPEC cards (`internal/factorylane/merge.go` `checkSyncAudit`).
Operator verdict: DECIDED (2c51): `complete` binds the sync-audit verdict to the card's CURRENT HEAD, and the required backend's receipt from that HEAD must pass; a previous PASS never substitutes. Refined by board 0577 P3 and P4: no backfill commit lands after the audited SHA (bookkeeping is committed before the final audit); NEW receipts carry the audited commit, and existing receipts stay valid records but cannot satisfy the HEAD binding.

## Still forbidden (board 9dbe, unchanged)

No replacement of a past PASS, no excluded tests, no widened default permissions, no audit skipping, fail-closed on every undecidable reading, and the single main writer.

## Rows Q14-Q17 (plan-audit iteration 1, in-card judgments)

### Q14: How is a receipt tied to a commit?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree (source: board 0577 P4)
Why unresolved: `auditreceipt.Receipt` carries no commit (`internal/auditreceipt/store.go:166-177`) and its writer sets none (`internal/cli/mcp_audit_receipt.go:70-80`); a verdict-file-only binding was the alternative.
Operator verdict: DECIDED (board 0577 P4, approved by the leader): add an audited-commit field to NEW receipts at write time, additive; existing receipts are never rewritten or deleted and cannot satisfy the HEAD binding. A receipt whose audited tree moved while the call ran is written with the field empty and never binds. Landing policy is unchanged beyond Q13: no existing receipt is rejected as a record; it only fails to satisfy a binding it never claimed.

### Q15: Where does a lane produce the manifest?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the audit found no named verb for the manifest.
Default: verbs under the existing `moai integration` family: `manifest derive`, `manifest show`, `manifest verify`, and `remeasure --call` (rule: the smaller user-visible surface — no new top-level family)
Alternate: a new top-level `moai landing` family
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:37:44Z manager-spec+lane t1872

### Q16: What is the manifest of a card with no Go change?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: a documentation-, rule- or configuration-only card has no `go test` call, and the audit found no defined admission.
Default: reader calls found by the reader search become calls; when there are none, the manifest is empty-with-evidence (changed-path list recomputed by the verifier, the `go list` proof that no Go package contains a changed path, the recorded reader-search result) and is admitted (rule: the option that preserves current behaviour — such cards land today, and the gate must not wedge them)
Alternate: require a declared non-Go verification command for every such card
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:37:44Z manager-spec+lane t1872

### Q17: How is an attempt's command compared with the manifest call's?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the audit showed a `-run`-narrowed attempt counts today.
Default: exact argv equality of the `go test` segment after dropping leading environment assignments and an `unset … &&` scrub segment; any other compound, pipe or extra segment refuses (rule: the option that refuses strictly more, the fail-closed direction the rulings require)
Alternate: compare only the package list and flags that select tests
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:37:44Z manager-spec+lane t1872

## Rows Q18-Q22 (plan-audit iteration 2)

### Q18: Which source decides whether the gate is on?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (source: board 470c)
Why unresolved: the primary checkout is a stale detached HEAD that an enable or revert commit on main never reaches (measured at 478285ab4c while main was elsewhere).
Operator verdict: DECIDED (board 470c): the committed blob `<integration-branch>:.moai/config/sections/workflow.yaml`, the branch named by the common-root integration configuration; no working tree is consulted; keeping the primary on main and reading the main worktree are rejected.

### Q19: How does a lane's verdict file carry the receipt ids, and how is it found?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: no Go code parses an `audit_receipt:` line from a verdict file; measured sample of 8 merge-ready SPEC cards: all 8 verdicts cite receipt ids somewhere, only 2 as a top-level `audit_receipt:` machine line (`research.md` §10).
Default: SPEC cards locate the newest `sync-audit*.md` under `.moai/reports/<card-id>/` (the T13 reader's rule: highest `-iter<N>`, then latest modification) and the verdict must carry exactly one `audit_receipt: rcpt-…[,rcpt-…]` machine line; the sync-auditor protocol and the audit artifact convention state it (rule: the smaller user-visible surface — one machine line instead of a new file). Consequence stated: a verdict lacking the line is refused while the gate is on; such a card is re-audited at landing anyway (board 2c51)
Alternate: parse the `AUDIT-VERDICT:` final-message line or free prose
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:04:08Z manager-spec+lane t1872

### Q20: What is the gate state when the key or the branch cannot be read?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (source: board 5d4d, refining board 470c)
Why unresolved: a gate that refuses landings when on must neither turn off silently nor wedge every landing.
Operator verdict: DECIDED (board 5d4d, D3): no integration branch configured → the project has not opted in → the gate is OFF; a branch configured but unresolvable → the landing is refused with the named error `gate_state_integration_branch_unresolvable` (fail-closed; not a deadlock, since the merge step cannot run without that branch and repair needs no landing). Stated by the lane within that ruling, implementation-level: with the branch resolved, an absent path or key or an explicit false reads off, and a failed read of a present path, an unparseable file or a non-boolean value reads on with the notice `gate_state_unreadable` (still admitting a compliant landing, including the repairing card).

### Q21: Which cited receipts must pass?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the auditor protocol cites every receipt id received, including failed extra calls; a universal rule would fail verdicts the existing guard accepts.
Default: at least one cited receipt must qualify (readable, codex or multi-backend tool, unmet-gate empty, codex verdict pass, audited tree matching, audited commit equal to HEAD); a failed or inconclusive extra call is recorded, does not block, and never replaces the pass (rule: the option that preserves current behaviour — the existing guard accepts one qualifying receipt)
Alternate: every cited receipt must pass
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:04:08Z manager-spec+lane t1872

### Q22: What does "the audited commit is the current HEAD" mean?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (sources: boards 2c51, 0577 P3, 5d4d)
Why unresolved: version 0.4.0 tolerated an ancestor whose diff is confined to `.moai/reports/<card-id>/`; its predicate also admitted an empty-diff ancestor that REQ-LVM-008 refuses.
Operator verdict: DECIDED (board 5d4d, D1): "equivalent" means equal to HEAD, one definition for the verdict and the receipt; no tolerance for an empty-diff, patch-identical or evidence-directory ancestor.

## Rows Q23-Q24 (plan-audit iteration 3, in-card judgments)

### Q23: Is the review target recorded in new receipts, and does it gate qualification?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the receipt binds a commit but not the audited scope; a partial-scope audit stamped with the current HEAD would read as the audit of that HEAD.
Default: record `audit_target` additively and do not gate on it; name the partial-scope case as accepted residual R5 (rule: the smaller user-visible surface — the legitimate full-scope target values are not enumerated at plan time, and a guessed gate could refuse valid audits)
Alternate: qualify only a base-branch review target
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:29:32Z manager-spec+lane t1872

### Q24: Does the receipt binding apply to SPEC-less cards?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (source: board 2c51 binds `complete` without a card-kind limit)
Why unresolved: the SPEC-less readiness path binds `audited_sha` only, and iteration 3 found the scope undefined.
Default: yes — the SPEC-less verdict is `.moai/reports/<card-id>/verdict.md`, and the citation line and the receipt binding apply to it as to a SPEC card (rule: the option that preserves the operator's ruling — Q13 has no card-kind exception)
Alternate: SPEC cards only
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:29:32Z manager-spec+lane t1872

## Rows Q25-Q26 (plan-audit iteration 4, inside the 4812 exception)

### Q25: Which flags may the effective GOFLAGS carry?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (scope granted by board 4812: "an effective-GOFLAGS constraint")
Why unresolved: a narrowing GOFLAGS set identically at derivation and at the attempt passed both the argv comparison and the fingerprint equality (reproduced by the auditor: plain `go test -json` ends fail, `GOFLAGS='-run=^TestGood$'` ends pass).
Default: the effective GOFLAGS, as resolved by `go env`, may carry no test-selection, test-truncation or test-shaping flag (`-run`, `-skip`, `-failfast`, `-short`, `-list`, `-shuffle`, `-count`, `-timeout`, `-p`, `-parallel`, `-cpu`, `-bench`, `-fuzz`); build-only flags pass; the 35e9 narrowing lives in the call's argv only (rule: the option that refuses strictly more, the fail-closed direction the rulings require)
Alternate: refuse only `-run` and `-skip`
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:45:28Z manager-spec+lane t1872

### Q26: What does a go.mod, go.sum, go.work or go.work.sum change select?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (scope granted by board 4812: "changes cannot yield an empty manifest")
Why unresolved: the four files are repository-wide inputs; the textual reader search would yield an arbitrary subset or an empty manifest.
Default: every Go package of the module becomes a whole-package call, no reader search, no 35e9 narrowing for a call selected only by them; the zero-call path never admits such a card (rule: the option that refuses strictly more)
Alternate: refuse the card as underivable
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:45:28Z manager-spec+lane t1872

## Rows Q27-Q29 (split to piece A2)

### Q27: When is the key switched on now that the card is split?

Label: FOUNDER
Class: product-level
Authority anchor: none in the committed tree (sources: boards 0577 P2, d20e)
Why unresolved: board 9dbe/e4cc ordered the enabling after the structured re-measure form, its tests and the lane guidance exist (all gate core, now card t1875); board 0577 P2 says the key turns on in a later, separate commit after piece A has landed. With A2 landing first the key, once on, enforces only the audit binding until A1 lands; this SPEC does not choose the moment.
Options (neutral): switch the key on after A2 lands (audit binding alone is enforced, completeness is added when A1 lands under the same key); switch it on only after A1 lands (both enforced together); keep a separate key for each half.
Operator verdict: DECIDED (board 6404, 2026-10-18T17:19:32Z, leader under operator delegation): the key is switched on only after BOTH A2 (t1872) and A1 (t1875) have landed, in a separate commit on main, judged by a build that contains both (0577 P1). Reason: with A2 alone on, `complete` is bound while `moai integration merge --card` stays a side door. A2 ships the audit binding and the switch reader with the key off.

### Q28: Which landing commands apply the audit binding?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the binding can be applied by `moai factory complete` alone, or also by merge-readiness condition (a) (`factorylane/merge.go:174`, called by `moai factory merge ready` and by the github-flow readiness check), or also by the direct merge verb and the merge step's landing check.
Default: `complete` only (rule: the smaller user-visible surface; it also leaves the github-flow readiness caller, the merge step and `merge.go` untouched, so this SPEC's edits of `factoryCompleteCard`, which piece A1 also edits, are exactly three sites: the binding call and the two card-moved re-reads that REQ-LVM-008 requires, with the one test seam before `RunMergeStep`; the re-measure record decision and the merge step are not touched. Corrected in 0.11.0: the earlier text counted a single site and predated the re-reads of Q31)
Alternate: `complete` plus condition (a) through an optional hook field of the triple input (the `VerifyRemeasure` precedent), left unwired by the github-flow caller; or, further, the merge step's landing check
Consequence recorded (design R6): a lane that skips `complete` is not bound by the audit binding, and `factory merge ready` can read ready where `complete` then refuses.
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:18:08Z manager-spec+lane t1872 (the stamp follows the convention of the setting commit `9e2b56196f`; the row itself first appeared in `49df956ac8`, 17:04:27Z); ruling added by board 6404 (2026-10-18T17:19:32Z): the binding is applied in `complete` only; the direct merge path binding (`factorylane/merge.go:174` and `:245`, `integration_merge_step.go:321-325`) is in piece A1's scope (spec §7 row 12), and R6 is an accepted residual for A2 because the key stays off until A1 lands.

### Q29: Is piece A2 independent of piece A1 in code?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree (source: board d20e asked the independence to be proved by code; the evidence is `research.md` §12)
Why unresolved: the split is sound only if no A2 requirement needs an A1 type, behaviour or verb, and if A2's `complete` behaves coherently while A1's manifest does not exist.
Operator verdict: DECIDED (board d20e: A2 stays on t1872 and lands first; A1 is blocked until t1872 is merged-local). Planner finding, not a ruling: no A2 requirement needs an A1 type, behaviour or verb (`research.md` §12, observed file and line); four places named an A1 concept and were adjusted (REQ-LVM-007 narrowing allowance, REQ-LVM-012 manifest wording, REQ-LVM-017 consumers and `manifest verify`, and the old REQ-LVM-011, whose audit-side half is attached to REQ-LVM-012 and REQ-LVM-018); with the gate on and no manifest in existence, `complete` runs today's checks plus the audit binding.

### Q30: Are the absent requirement ids a defect or a documented exemption?

Label: FOUNDER
Class: process
Authority anchor: none in the committed tree (source: board 9b0e)
Why unresolved: strict id contiguity (plan-audit MP-1) conflicts with stable ids across the split d20e; renumbering would rewrite the A1 hand-off table, this SPEC and every criterion reference.
Operator verdict: DECIDED (board 9b0e, 2026-10-18T17:36:48Z, leader under operator delegation): the gap is an explicit exemption, not renumbered; the clause is in spec.md §3 (REQ id gap waiver) naming the gap ids, the reason (split d20e) and this record.

### Q31: Does `complete` close the gap between the binding and the merge step's own pin?

Label: FOUNDER
Class: implementation-level
Authority anchor: none in the committed tree (source: board 7c5c)
Why unresolved: the binding checks the card branch before the window, but the merge step pins `refs/heads/<card-branch>` itself (`integration_merge_step.go:289`) and the bound SHA never reaches it; closing it fully means editing the merge step, which Q28 assigns to piece A1.
Operator verdict: DECIDED (board 7c5c, 2026-10-18T17:57:13Z, leader under operator delegation): the minimal measure inside `complete` only — re-read `refs/heads/<card>` immediately before the merge-step call (and before adoption) and refuse with the named error `audit_binding_card_moved` when it differs from the bound SHA, with a test and a mutant row; making the merge step's pin consume the bound SHA stays in A1 (t1875, spec §7 row 13); the remaining window is accepted in A2 (spec C12, design R8) because the key stays off until both pieces have landed.


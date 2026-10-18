# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — decision index (card t1872)

Version 0.3.0 adds the leader's policy rulings on plan-audit iteration 1 (board `d-20261018T153534Z-0577`, `0577`) and rows Q14-Q17. Version 0.4.0 adds the N2 ruling on plan-audit iteration 2 (board `d-20261018T160111Z-470c`, `470c`) and rows Q18-Q22. Version 0.5.0 adds the rule C ruling on plan-audit iteration 3 (board `d-20261018T162457Z-5d4d`, `5d4d`) and rows Q23-Q24, and re-states Q20 and Q22.

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
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:03:32Z manager-spec+lane t1872 (template default false; the in-repo switch is Q7)

### Q3: Is "ready to land" a derived view or a stored marker?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the card asks for a ready state; the form is open.
Default: not built in A (rule: the smaller user-visible surface); the follow-up piece C is to be derived at read time with no stored marker
Alternate: build a stored marker now
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:03:32Z manager-spec+lane t1872 (OUT-OF-SCOPE for A per board 9dbe; recorded for C)

### Q4: Where does package coverage come from?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: coverage could be lane-declared or observed from the test stream.
Default: observed from the `go test -json` terminal events against a `go list`-resolved expected set (rule: the smaller user-visible surface — no new declaration for lanes to keep)
Alternate: lane-declared package lists
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:03:32Z manager-spec+lane t1872

### Q5: Does a test-only change still pull in the changed package's importers?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (the b2cd text and the t1709 reading are board rulings)
Why unresolved: the standing scope reads "packages the card changed plus the packages importing them" with no test-only exclusion.
Default: importers are always included (rule: the option that preserves current behaviour — the conservative resolution applied at t1709); consistent with the ruling's "no test exclusion"
Alternate: exclude importers when only `_test.go` files changed
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:03:32Z manager-spec+lane t1872

### Q6: Is the reuse decision by content equality of the scope or by path-matching?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: only relevant to the reuse relaxation.
Default: not built (rule: the option that preserves current behaviour)
Alternate: build either predicate
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:03:32Z manager-spec+lane t1872 (OUT-OF-SCOPE for A: reuse is not built, Q8)

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
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:58:00Z manager-spec+lane t1872

### Q16: What is the manifest of a card with no Go change?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: a documentation-, rule- or configuration-only card has no `go test` call, and the audit found no defined admission.
Default: reader calls found by the reader search become calls; when there are none, the manifest is empty-with-evidence (changed-path list recomputed by the verifier, the `go list` proof that no Go package contains a changed path, the recorded reader-search result) and is admitted (rule: the option that preserves current behaviour — such cards land today, and the gate must not wedge them)
Alternate: require a declared non-Go verification command for every such card
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:58:00Z manager-spec+lane t1872

### Q17: How is an attempt's command compared with the manifest call's?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the audit showed a `-run`-narrowed attempt counts today.
Default: exact argv equality of the `go test` segment after dropping leading environment assignments and an `unset … &&` scrub segment; any other compound, pipe or extra segment refuses (rule: the option that refuses strictly more, the fail-closed direction the rulings require)
Alternate: compare only the package list and flags that select tests
Operator verdict: DEFAULT-APPLIED 2026-10-18T15:58:00Z manager-spec+lane t1872

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
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:40:00Z manager-spec+lane t1872

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
Operator verdict: DEFAULT-APPLIED 2026-10-18T16:40:00Z manager-spec+lane t1872

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
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:20:00Z manager-spec+lane t1872

### Q24: Does the receipt binding apply to SPEC-less cards?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (source: board 2c51 binds `complete` without a card-kind limit)
Why unresolved: the SPEC-less readiness path binds `audited_sha` only, and iteration 3 found the scope undefined.
Default: yes — the SPEC-less verdict is `.moai/reports/<card-id>/verdict.md`, and the citation line and the receipt binding apply to it as to a SPEC card (rule: the option that preserves the operator's ruling — Q13 has no card-kind exception)
Alternate: SPEC cards only
Operator verdict: DEFAULT-APPLIED 2026-10-18T17:20:00Z manager-spec+lane t1872

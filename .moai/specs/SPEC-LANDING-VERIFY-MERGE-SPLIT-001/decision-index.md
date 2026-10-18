# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — decision index (card t1872)

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
Operator verdict: DECIDED (operator direct request + e4cc, recorded by board 9dbe): once piece A lands, the completeness gate is turned ON in this repository. Order inside A: (1) the structured re-measure form is available and lanes are told to use it; (2) the negative tests (script-wrapped record, missing package, wrong SHA, interrupted attempt are REFUSED) and the positive test (a complete structured record is admitted) pass; (3) only then, in the same piece, the configuration key is switched ON.

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
Operator verdict: DECIDED (2c51): `complete` binds the sync-audit verdict to the card's CURRENT HEAD, and the required backend's receipt from that HEAD must pass; a previous PASS never substitutes.

## Still forbidden (board 9dbe, unchanged)

No replacement of a past PASS, no excluded tests, no widened default permissions, no audit skipping, fail-closed on every undecidable reading, and the single main writer.

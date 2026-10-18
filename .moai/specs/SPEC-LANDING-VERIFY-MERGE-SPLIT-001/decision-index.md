# SPEC-LANDING-VERIFY-MERGE-SPLIT-001 — decision index (card t1872)

`interview.decision_gate` is `on` in this repository. Rows Q1-Q6 are in-card implementation-level judgments, settled by the published Default rule and stamped at plan close. Rows Q7-Q13 are product-level or evidence-needed: each is OPEN, carries no default and no preferred answer, and blocks the autonomous Kickoff until the operator answers. Machine-local board rulings are not citable authority (not in the committed tree), so no row below is labelled DECIDED or POLICY-COVERED on their strength.

## In-card judgments (implementation-level, Default applied)

### Q1: Is the six-part scope one Tier L SPEC or several SPECs?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (spec-workflow.md § SPEC Complexity Tier gives ceilings, not this choice)
Why unresolved: no committed rule says how a multi-part landing repair is divided; the leader invited a split proposal.
Default: one SPEC, Tier L (rule: undo is a single revert of this SPEC's own commits)
Alternate: three SPECs A (gate), B (slot), C (ready, metrics, reuse) as laid out in design.md §9
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

### Q2: Do the new configuration keys ship off or neutral?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: no committed rule states the shipped default for a new landing key.
Default: gate false, capacity 1 (thresholds unset), reuse closed (rule: the option that preserves current behaviour)
Alternate: ship the gate on
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

### Q3: Is "ready to land" a derived view or a stored marker?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the card says "mark" a ready card; the form is open.
Default: derived at read time, no stored marker (rule: the smaller user-visible surface; also avoids a marker that goes stale when the tip moves)
Alternate: a stored marker with invalidation on tip movement
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

### Q4: Where does the manifest's package denominator and the attempt's package coverage come from?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: coverage could be declared by the lane or observed from the test stream.
Default: observed from the `go test -json` terminal events against a `go list`-resolved expected set (rule: the smaller user-visible surface — no new declaration for lanes to maintain)
Alternate: lane-declared package lists
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

### Q5: Does a test-only change still pull in the changed package's importers?

Label: FOUNDER
Class: implementation-level
Authority anchor: none (the b2cd text and the t1709 reading are machine-local board rulings; the committed tree holds neither)
Why unresolved: the standing scope says "packages the card changed plus the packages importing them" with no test-only exclusion; a later reading treated test-only as not affecting importers.
Default: importers are always included (rule: the option that preserves current behaviour — the conservative resolution applied at t1709)
Alternate: exclude importers when only `_test.go` files changed
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

### Q6: Is the reuse decision made by content equality of the scope or by path-matching the moved commits?

Label: FOUNDER
Class: implementation-level
Authority anchor: none
Why unresolved: the card describes overlap between moved commits and the scope; either predicate satisfies it.
Default: content equality of the scope on the source tree versus the new tree, with the moved-commit list kept as evidence (rule: the option whose undo is a single revert; both qualify, and equality refuses strictly more, which is the fail-closed direction the card requires)
Alternate: path-matching of the moved commits against the scope
Operator verdict: DEFAULT-APPLIED 2026-10-18T14:57:18Z manager-spec+lane t1872

## Open questions for the leader and operator

### Q7: Should the completeness gate be enabled in this repository's configuration, and from when?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: enabling it makes `moai factory complete`, the merge step and merge-readiness refuse any landing whose record is a script-wrapped call (the form 16 current lanes use); the card text says the gate lands before reuse but does not say when it becomes mandatory.
Operator verdict:

### Q8: Is the re-measure reuse relaxation (part 1) built, and under what non-Go policy?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: measured adjacent-landing eligibility is 1 of 37 under the card's own conservative rule and 10 of 37 if moved non-Go files outside the closure are ignored (`research.md` §6, unsound for tests that read such files). Options are: build it strict only; build it with an operator-declared never-read glob list; build it after the shadow counter of M4 has collected landings; do not build it. Each changes what a record is allowed to stand for.
Operator verdict:

### Q9: Is heavy-slot capacity raised to 2 in this repository, and with what load and memory thresholds?

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none
Why unresolved: the thresholds depend on measurements of the machine under capacity 2 that do not exist; the SPEC ships capacity 1 and no numeric default.
Operator verdict:

### Q10: Is slot liveness bound to the owning session, or to the heavy command's process, and is a `moai slot run` wrapper wanted?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: measured today, a dead-pid lease is already reclaimable, so the card's premise holds only for a lease whose SESSION is alive while its heavy command died or was abandoned (`research.md` §4). Binding to a command process or adding a wrapper verb is new capability and a new meaning of "holder died". This SPEC specifies only record removal for a dead owner session (REQ-LVM-010).
Operator verdict:

### Q11: Which packages are wall-clock-bound, for the concurrent-run exclusion?

Label: EVIDENCE-NEEDED
Class: product-level
Authority anchor: none
Why unresolved: which packages' tests are wall-clock-bound is a measured fact not recorded anywhere; the list ships empty.
Operator verdict:

### Q12: Who vouches that a measured package is hermetic (no live-git or undeclared non-Go reads), and where is the list kept and reviewed?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: reuse treats an undeclared package as not hermetic; a wrong declaration lets a record cross trees it should not (residual risk R3, design.md §4).
Operator verdict:

### Q13: Should `moai factory complete` itself bind the sync-audit verdict to the card's current HEAD for SPEC cards?

Label: FOUNDER
Class: product-level
Authority anchor: none
Why unresolved: today only `sync_status: complete` is read for SPEC cards (`internal/factorylane/merge.go` `checkSyncAudit`); the ready view of this SPEC applies the HEAD bind for readiness only, leaving `complete`'s own gate unchanged. Tightening `complete` changes landing policy beyond the card text.
Operator verdict:

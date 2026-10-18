# Decision index: SPEC-CARD-DURABLE-IDENTITY-001

Decisions surfaced while assembling the SPEC that the operator did not settle in an interview (the card is an operator-delegated design; no interview ran). Every row is labelled by the routing rule: no committed artifact decides these questions under identical conditions, no operator setting or constitution clause covers them, and none needs data that does not exist yet. The authority register holds committed artifacts only; the leader's design document and the card text are machine-local and are not cited as authority. What the card text itself prescribes (row preservation, retirement as migration, old builds refusing writes, rollback preserving new grants, the journal, the flag default off) is not a row: it is the scope, and spec.md § B records it.

The Default of each row is the published rule's selection, applied in order: the option that preserves current behavior; else the option whose undo is a single revert of this SPEC's own commits; else the option with the smaller user-visible surface. A Default is a policy application, not a recommendation. Q2 and Q7 were ruled by the lane after plan audit iteration 1 (verdict LANE-RULED, no Default); the others carry the rule's Default. No row is product-level: every behavior these rows change sits behind `workflow.card_identity.enabled`, which ships `false`, so no shipped command's default behavior, feature or template default changes.

### Q1: Is the identity layout selected by the database or by the config key at every open?

Label: FOUNDER
Authority anchor: none
Why unresolved: two placements satisfy the requirements. The tracked `workflow.yaml` is copied into every worktree and can differ between trees, while the factory database is shared by all of them, so a key consulted at every open could refuse a lane whose tree carries another value. No committed artifact decides between them.
Class: implementation-level
Default: the recorded schema version carries the layout and the key only permits the cutover verb (rule: smaller user-visible surface — one refusal on one verb, against a refusal path at every open of every tree; the first two clauses of the rule do not rank the options, since both leave current behavior unchanged and both revert in one revert)
Alternate: the key is read at every open and a mismatch with the database's layout refuses the open
Operator verdict: DEFAULT-APPLIED 2026-10-18T01:21:43Z manager-spec (lane t1842, plan close)

### Q2: What is an attempt?

Label: FOUNDER
Authority anchor: none
Why unresolved: the design document says re-grants create attempts and the card says every historical `(run_id, card_id)` row is an attempt; the two readings differ when a lease expires and is granted again inside one run. Iteration 1 of the plan audit (D10) found that narrowing principle 6 by a defaulted row was not an implementation-level choice.
Class: implementation-level
Operator verdict: LANE-RULED 2026-10-18T01:21:43Z lane-41 claude (decision record: decided_by=lane-41 claude, ladder_path=disk evidence: design document section 2 principle 6, plan-audit.md D3/D10; operator away). Re-grants create attempts: every write that ends a lease grant appends an immutable attempt (REQ-CDI-023). Feasible in the grant path: one archive row per ended grant inside the writing transaction, no second transaction, so no open obstacle row.

### Q3: What becomes of a non-terminal card whose run is already retired when the cutover runs?

Label: FOUNDER
Authority anchor: none
Why unresolved: such a card needs a run to be worked on, and the cutover cannot know which. It can leave the card in project scope (no open membership) for a later verb to join, or demand a destination for each at cutover time. Measured on the copy of the live database: 93 card ids are in that position.
Class: implementation-level
Default: project scope; the card keeps every field and a later `assign` or `next` joins it to a run (rule: smaller user-visible surface — the cutover takes no destination input)
Alternate: the cutover refuses until each such card, or all of them through one flag, is given a destination run
Operator verdict: DEFAULT-APPLIED 2026-10-18T01:21:43Z manager-spec (lane t1842, plan close)

### Q4: Who starts the cutover?

Label: FOUNDER
Authority anchor: none
Why unresolved: the card requires a quiet window even behind a flag; whether the first open with the key true performs the migration or an explicit verb does is not stated.
Class: implementation-level
Default: an explicit verb, `moai factory identity cutover` (rule: preserves current behavior — opening the database never migrates its layout, and an older build's open of an unmigrated database is unchanged)
Alternate: the first open of the factory by a build with the key true performs the cutover
Operator verdict: DEFAULT-APPLIED 2026-10-18T01:21:43Z manager-spec (lane t1842, plan close)

### Q5: Where does the operation journal live?

Label: FOUNDER
Authority anchor: none
Why unresolved: the existing failed-write log is a file beside the database so that it survives a database failure; a journal inside the database commits with the grant it records. The card asks for a durable journal or outbox and does not choose.
Class: implementation-level
Default: a table in the factory database (rule: smaller user-visible surface — one store and no new file format to read or clean)
Alternate: a file beside the database for the kinds whose effect precedes the factory change, and a row in the database only for `card_grant`
Operator verdict: DEFAULT-APPLIED 2026-10-18T01:21:43Z manager-spec (lane t1842, plan close)

### Q6: Are leader approval receipts carried across a migration?

Label: FOUNDER
Authority anchor: none
Why unresolved: a receipt binds a run and a version, and a migration changes both. Carrying it means re-stamping a receipt the leader issued, which the leader-approval code does only at the done edges (`leader_approval.go`, the approval re-bind inside the transition).
Class: implementation-level
Default: not carried; the leader re-approves after a migration (rule: preserves current behavior — a receipt is valid only for the run and version it names and is stale once either moves)
Alternate: the migration re-stamps the receipt to the destination run and the new version in the same transaction
Operator verdict: DEFAULT-APPLIED 2026-10-18T01:21:43Z manager-spec (lane t1842, plan close)

### Q7: What counts as conflicting active owners that block the cutover?

Label: FOUNDER
Authority anchor: none
Why unresolved: the card says conflicting active owners block migration with a report; the first reading (non-terminal rows in two active runs) cannot fire on current data and never inspects an owner or a lease (plan audit D3).
Class: implementation-level
Operator verdict: LANE-RULED 2026-10-18T01:21:43Z lane-41 claude (decision record: decided_by=lane-41 claude, ladder_path=disk evidence: plan-audit.md D3, design document line "Conflicting active owners block migration with a report"; operator away). Strict rule: a card id blocks, with a report and no override, when its non-terminal rows carry two or more distinct non-empty owner labels (owner label or lease holder), or when a non-terminal row that is not the resolved row still holds an unexpired lease or a registered-worker holder (REQ-CDI-005). Measured on a fresh copy (research.md § 4.1): 25 ids by owner labels, at most 27 with the lease rule.

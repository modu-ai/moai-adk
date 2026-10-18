---
id: SPEC-CARD-DURABLE-IDENTITY-001
title: "Durable card identity — project-scoped card record, immutable attempts, run membership, retirement as migration, and an operation journal (stage B of the card policy redesign)"
version: "0.4.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/homestate, internal/cli, internal/config, internal/factory, internal/factorymsg"
lifecycle: spec-anchored
tags: "factory, card-identity, attempts, run-membership, retire-migration, operation-journal, outbox, stage-b, t1842"
tier: L
card: t1842
related_specs: [SPEC-CARD-LIVENESS-OWNERSHIP-001, SPEC-FACTORY-RETIRE-GUARD-001, SPEC-FACTORY-RECORD-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-MERGE-WINDOW-QUEUE-001, SPEC-INTEGRATION-LOCK-LIVENESS-001, SPEC-TODO-CLAIM-LEASE-001, SPEC-FACTORY-COMPLETION-RECOVERY-001]
---

# SPEC-CARD-DURABLE-IDENTITY-001 — Durable Card Identity

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-10-18 | manager-spec | Initial plan-phase draft for card t1842 (class C, P0), stage B of the card policy redesign. Written against tree B0 `f4bcd4f2cc65342559b505f1e5a29df80f992894`. Stage A (card t1841) is on hold at its plan-audit ceiling and is re-planned after this stage, so this SPEC publishes the interface stage A consumes instead of assuming stage A's code (§ E). Counts: 22 requirements, 24 acceptance criteria. |
| 0.2.0 | 2026-10-18 | manager-spec | Revision after plan audit iteration 1 (FAIL 0.74; blocking D1 to D11, optional D12 to D17; `.moai/reports/t1842/plan-audit.md`). Lane rulings applied: attempts follow design principle 6, so every ended lease grant appends an immutable attempt (new REQ-CDI-023, new AC-CDI-025, decision-index Q2 ruled); the cutover conflict rule is strict on owners and live leases (REQ-CDI-005, Q7 ruled). Also corrected: rollback trigger and index names, the read-only invariant, the attempt floor across rollback, idempotence beyond in-flight rows, the journal identity guard, the assign-only claim state (fourteen cells), the stage-A seam wording, and the rule for column-adding migrations. Counts: 23 requirements, 25 acceptance criteria. Per-defect dispositions: plan.md § J. |
| 0.3.0 | 2026-10-18 | manager-spec | Revision after plan audit iteration 2 (FAIL 0.78; blocking R2-D1 to R2-D5, optional R2-D6 to R2-D10; `.moai/reports/t1842/plan-audit-iter2.md`). The seal moves from per-caller code to a store-level trigger on the current record, so no writer of owner, holder or run can miss it (REQ-CDI-023, S-5); rollback maps a project-scope record to the run of its newest sealed attempt (REQ-CDI-011); the journal refuses any insert that would collide with an in-flight intent, so no conflict resolution can delete a row (REQ-CDI-021); milestone mapping, the blocked-id counts and the operator sequence are corrected. Counts unchanged: 23 requirements, 25 acceptance criteria. Dispositions: plan.md § K. |
| 0.4.0 | 2026-10-18 | manager-spec | Correction after plan audit iteration 3 (FAIL 0.80), limited to its five blocking defects R3-D1 to R3-D5 (`.moai/reports/t1842/plan-audit-iter3.md`): rollback returns a project-scope record to the run it left and never forms a pair from an empty-run seal (REQ-CDI-011); a version floor accompanies the attempt floor (REQ-CDI-011, REQ-CDI-012); the journal refuses any insert onto an existing operation id (REQ-CDI-021); V-4 exempts runs already absent before the cutover (REQ-CDI-009); milestone mapping of AC-CDI-010 and AC-CDI-022. Counts unchanged. Dispositions: plan.md § L. |

## §A Problem

Each fact below was read in the card worktree at B0 unless it names another source.

1. **The key is the run.** `cards` is declared `PRIMARY KEY(run_id, card_id)` (`internal/homestate/factory.go:77`). A card id can therefore exist as an independent row in every run. The `Card` record has no ownership epoch (`internal/homestate/card_record.go:100-138`) and no non-test code anywhere names `owner_epoch` (ledger E09). Every read and write addresses a row by `run_id AND card_id` (`card_record.go:181`, `card_transition.go:882`).
2. **A new run re-creates the card instead of continuing it.** `recordPickedTx` inserts a fresh `picked` row at version 1 when the `(run, card)` pair is absent (`internal/homestate/card_picked.go:159-170`). Stage, evidence, audits and the audit-entry commit of the earlier run's row are not carried. The only carry-over that exists is the worktree binding (`PreviousCardWorktree`, `card_worktree.go:63-73`, used by `factoryCarryPreviousWorktree`, `internal/cli/factory_card.go:528-544`), and the only cross-run live-lease check is inside reassignment (`internal/homestate/card_reassign.go:19-23`, `:92`); the lease-acquire case checks registration and owner label only (`card_transition.go:582-590`).
3. **Retirement strands.** `RetireRunIfDead` (`internal/homestate/factory_run_retire.go:381-399`) calls `retireRun` (`:407-422`), which its own comment says retires "regardless of the records the run carries" (`:401-406`). It is reached by `moai factory runs --retire` (`internal/cli/factory_handoff_recover.go:91`) and by `moai factory relaunch --from-run` (`internal/cli/factory_relaunch.go:98`). The automatic path refuses such a run (`retireRunUnlessLive`, `:431-456`, through `runHasLiveRecords`, `internal/homestate/factory_run_adopt.go:31-38`), so only the operator path strands.
4. **The stranded population is large and ambiguous.** On a copy of the live factory database taken 2026-10-18T00:31:45Z (a one-instant shape, not a rate): 293 card rows over 221 card ids; 55 card ids carry more than one row (at most 6); 143 non-terminal rows sit in 36 retired runs (run `tmvgvn` alone holds 78); 93 card ids have non-terminal rows only in retired runs; 38 card ids have two or more non-terminal rows (17 with exactly one of them in the active run, 21 with all of them in retired runs), and 25 of the 38 carry two or more distinct owner labels (18 with every such row in a retired run, 7 with one row in the active run); no card id has non-terminal rows in two active runs; no approval, dispatch binding or card row references a missing run or card. Commands and outputs: research.md § 4.
5. **Cross-store operations leave no durable intent.** The lane claim writes the queue's current-dispatch record inside the factory transaction right before commit (`factory_card.go:1170-1191`, `internal/factory/todo_dispatch_current.go:31`). The integration window is a file record (`internal/factory/integration_lock.go:122`, `:355`) and the merge is a git operation (`internal/factory/integration_merge_step.go:178`); a transition failing after the merge is handled by an ad hoc hold (`completePostMergeConflict`, `factory_card.go:2395-2402`) and a partly-done merge is recognised only by the adoption probe inside `factoryCompleteCard` (`:2229-2258`). Failed dispatch-mirror writes go to a file log (`internal/homestate/card_unavailable.go:18`). The queue store has its own operation journal for dispatch (`internal/factory/gtd_operation.go:16-21`, `:144`), which does not cover the factory half.
6. **The seam with stage A is this key.** Stage A keys its grant by card id alone and had to add a grant-to-card consistency gate, a governing-row definition and a revocation-reach rule because `cards` can hold several lease-holding rows for one card id; four plan-audit rounds failed on that seam (research.md § 3). The design document (machine-local, untracked, not an authority) states the target: "durable card identity; runs are membership".

## §B Scope decisions

Decided by the card text (t1842) unless a decision-index row says otherwise.

1. **Identity.** The card id is the only key of a card. A run is a membership record. An **attempt** is one lease grant or one run engagement of a card, whichever ends first: it ends, and its ending state is sealed as an immutable archive row, when a write clears the lease holder, reassigns the owner, moves the card to another run or to project scope, or re-picks a terminal card (REQ-CDI-023). The card's **current record** continues the open attempt. Design principle 6 (re-grants create attempts, preserved immutably) is followed in full (decision-index Q2, ruled).
2. **Cutover is an explicit, flag-gated operator act.** A new key `workflow.card_identity.enabled` defaults `false`. The layout is a property of the database (an identity marker on the recorded schema version), not of the tracked config file, because every worktree shares one database and carries its own copy of the config (Q1, Q4).
3. **Row preservation is the core.** Every legacy row becomes one sealed attempt, verbatim. The legacy table is retained under a non-live name. No row is dropped, merged, overwritten, or chosen by recency.
4. **Duplicates need an explicit resolution; active conflicts block.** The current record continues a legacy row that an operator-supplied resolution names. The cutover blocks, with a report and no override, a card id whose non-terminal rows carry two or more distinct non-empty owner labels, and any non-terminal row that is not the resolved row and still holds an unexpired lease or a live holder (Q7, ruled strict).
5. **Stranded cards become addressable.** A non-terminal card whose run is retired becomes a current record in project scope (no open membership) with every field kept, and is joined to a run by the verbs that already record a card in a run (Q3).
6. **Retirement is migration.** On the identity layout the operator retirement entry migrates the run's active cards, then retires the run, in one transaction; it refuses when migration cannot complete and offers no force path.
7. **Old builds are refused after cutover; rollback exists and keeps what was written after cutover.**
8. **Cross-store operations are journaled.** Queue claim, card grant, integration-window acquisition and git merge carry an operation id in a durable journal with deterministic reconciliation that any session can run (Q5).
9. **Stage order and the seam.** The design document ordered A before B; stage A's own hold record reverses it (B then A). This SPEC does not depend on stage A's code and publishes the facts stage A, D and F consume (REQ-CDI-022, § E).
10. **Receipts are not carried across a migration.** A leader approval binds a run and a version; a migration changes both, so the leader re-approves (Q6).
11. **Unchanged on purpose:** the lease clock and its columns, the serial slot, the automatic reconciler's retirement write, the queue store's schema, the integration-lock record format.

## §C Requirements (GEARS)

"The factory" means the factory record layer (`internal/homestate`) and the operator and lane verbs that write it. "Legacy layout" means a database whose recorded schema version carries no identity marker; "identity layout" means one that carries it. "Current record", "attempt", "membership", "project scope", "active card", "sealed attempt" and "operation" are defined in design.md § 1. A **terminal** state is `done`, `failed` or `abandoned`.

### C.1 Flag, layout authority, baseline

- **REQ-CDI-001** (Ubiquitous) — The configuration shall expose exactly one card-identity key, `workflow.card_identity.enabled`, a boolean whose code default, repository default and distributed template default are all `false`. A value that is not a boolean shall refuse to open the factory with a message naming the key path, the allowed values and the config file, and shall never be coerced.
- **REQ-CDI-002** (Capability gate) — **Where** the database is on the legacy layout, the factory shall behave as at B0: no table, view, trigger, journal row, event kind or schema-version change shall be written, every card, lease, retire and merge verb shall keep its B0 semantics and output, and `moai factory identity cutover` shall refuse naming the key whenever `workflow.card_identity.enabled` is `false`. **Where** the database is on the identity layout, the factory shall operate on that layout whatever value the key holds in the opening tree's configuration.

### C.2 Cutover

- **REQ-CDI-003** (Event-driven) — **When** an operator or leader session runs `moai factory identity report`, the factory shall read the database without writing and print the legacy row count, the distinct card-id count, every card id with more than one legacy row and for each the candidate rows (run, state, owner label, lease holder, lease expiry, run status), every card id blocked under REQ-CDI-005 (rule (b) is evaluated against the proposed resolution, and an id with a blank proposal is listed as blocked unless resolved to a named row), the count of unreconciled entries in the record-unavailable log, and a resolution input in which a card id is pre-filled only where exactly one candidate qualifies (its only non-terminal row, or its only non-terminal row in an active run) and blank everywhere else. The `factory.db` file and the content of every table shall be identical before and after; the read-only open may create `-wal` and `-shm` sidecar files, which this invariant permits (the project driver creates them even on a quiescent database).
- **REQ-CDI-004** (Unwanted behavior) — The cutover shall not choose among a card id's legacy rows by recency, order or any default. It shall proceed only when a resolution input names, for every card id with more than one legacy row, one of that id's own legacy rows by run id; where the id has a non-terminal row the named row shall be non-terminal. A missing, blank, unknown or terminal-while-active naming shall refuse the cutover, list each offending id, and write nothing.
- **REQ-CDI-005** (Unwanted behavior) — The cutover shall refuse, listing each card id, its rows, their owner labels and lease expiries, when (a) the non-terminal rows of one card id carry two or more distinct non-empty owner labels (owner label or lease holder), (b) a non-terminal row that is not the card id's resolved row holds an unexpired lease (an unparseable expiry counts as unexpired) or a lease holder that is a registered worker bound to that row's run, or (c) the record-unavailable log holds an unreconciled entry. No resolution input shall override this refusal. The lease reading in (b) is a precondition check at the cutover's single `now`, not an ownership decision.
- **REQ-CDI-006** (Event-driven) — **When** `moai factory identity cutover` runs with `workflow.card_identity.enabled` true, every precondition of REQ-CDI-004 and REQ-CDI-005 holding, and the session not a lane session, the factory shall in one transaction: copy every legacy row into the attempt archive as one sealed attempt; create one current record per distinct card id continuing its resolved row; create the membership records; keep the legacy rows unchanged under a name that cannot be mistaken for the live table; replace the legacy table name by a read-only projection of the current records; and mark the recorded schema version with the identity marker.
- **REQ-CDI-007** (Ubiquitous) — The number of sealed attempts after a cutover shall equal the number of legacy rows before it, and every column of each attempt shall equal the same column of its legacy row, compared by column name; no legacy row shall be dropped, merged, overwritten or re-keyed. A current record whose resolved row is terminal, or belongs to an active run, shall carry that row's values unchanged (version included); one whose resolved row is non-terminal and belongs to a retired run shall carry them with an empty run (project scope) and a version one higher than the resolved row's, or the card's version floor (REQ-CDI-011) plus one when that is higher, and a `card.migrated` event shall record the change. Sealed attempts shall be numbered above any attempt floor carried from a rolled-back layout (REQ-CDI-011).
- **REQ-CDI-008** (Ubiquitous) — A sealed attempt shall be immutable: an UPDATE, DELETE or replacing INSERT addressed to one, by any statement text on any connection, shall be refused by the database itself with an error that names the immutability, leaving every row unchanged.
- **REQ-CDI-009** (Event-driven) — **When** the cutover has written its changes and before it commits, the factory shall verify in the same transaction that: the attempt count equals the legacy row count; the current-record count equals the distinct card-id count; every attempt's card has a current record; every membership record names an existing attempt or current record and a run that exists in `runs` or was already absent from `runs` before the cutover (the set of absent runs is compared before and after, like the dangling references); every reference that resolved before the cutover (leader approvals by run and card, dispatch bindings by card and run) resolves after it; and the set of references that did not resolve before is unchanged. Any failure shall roll the transaction back, leaving the database identical to its state before the attempt (same schema version, same objects, same rows), and the verb shall exit non-zero naming the failed check.
- **REQ-CDI-010** (State-driven) — **While** the database carries the identity marker, a build that predates this SPEC shall be unable to open it (refused as an unsupported schema version by the existing open-time refusal), and a connection opened before the cutover shall be unable to insert, update or delete a card through the legacy table name: the statement shall abort with an error that names the cutover and leave every record unchanged; reads through the legacy name shall keep returning the current records in the legacy column shape.
- **REQ-CDI-011** (Event-driven) — **When** `moai factory identity rollback` runs, the factory shall in one transaction rebuild the legacy table with one row per `(run, card)` pair from the newest attempt-or-current-record of that pair, so that every lease grant, state change and version written after the cutover is carried; return a record in project scope (empty run) to the run it left, which the record keeps from the moment it entered project scope, and never form a pair from a sealed attempt whose run is empty; keep every identity-layer record, with its explicitly created triggers and indexes, under names that cannot collide with those of a later cutover, and record each card's highest `attempt_no` and `version` as floors that the next cutover numbers and versions above; and restore the numeric schema version. It shall refuse, changing nothing, when any check of REQ-CDI-009, run in the reverse direction, fails or when any journal row is non-terminal, and shall be refused in a lane session.

### C.3 Identity semantics on the identity layout

- **REQ-CDI-012** (Ubiquitous) — The factory shall hold at most one current record per card id and at most one open membership per card, and a card's `version` and `attempt_no` shall never decrease across any operation (seal, migration, join, re-pick, cutover, and a rollback followed by a later cutover, where a version that would fall below the floor kept by the rollback becomes the floor plus one). A request that names a run other than the card's open membership shall be refused with an error distinct from card-not-found that names the open membership (or project scope). No reader of lease state (reassignment, the serial-slot reader, the status view, a revocation hook, the live-records check of a run) shall treat a sealed attempt as a lease.
- **REQ-CDI-013** (Event-detected) — **When** a verb records, picks, assigns or claims a card id in a run: if the current record is non-terminal with an open membership in a different run, the factory shall refuse naming that run and the retire-with-migration verb, create no second record, and change no field; if the current record is non-terminal in project scope, the factory shall open a membership in the named run, advance the version once, and leave state, stage, evidence, worktree, owner, lease holder and lease expiry unchanged; if the current record is terminal, the factory shall seal the current attempt and open a new attempt in state `picked` whose version continues from the sealed one.
- **REQ-CDI-014** (Unwanted behavior) — The cutover, a seal, a migration, a join and the rollback shall not delete, lower or rewrite any row of any table in the factory database that is keyed by card id and carries an `owner_epoch` column, and a migration, a join, a cutover and the rollback shall write a record's lease holder, lease expiry and heartbeat only to the values they already held, so that none presents a cleared lease holder to a revocation hook. A seal is the archive append and the `attempt_no` bump only; the write it accompanies (a lease expiry, a reassignment) clears the holder exactly as at B0, after the seal.

### C.4 Retirement as migration

- **REQ-CDI-015** (Event-driven) — **When** `moai factory runs --retire <run>` or `moai factory relaunch --from-run <run>` reaches the retirement entry on the identity layout and the run's owner classifies `dead` (the existing predicate, unchanged), the factory shall in one transaction migrate every active card of the run (every current record whose open membership is the run and whose state is non-terminal) and then retire the run. Each migration shall seal the card's attempt, continue the record in the destination with every column unchanged except the run, the version (one higher) and `attempt_no` (one higher), move the membership, re-point the card's dispatch binding, re-bind the owning lane's roster entry from the retired run to the destination, and append one `card.migrated` event naming from-run, to-run and both attempt numbers. The destination shall be `--into <run>` when given; else, for `relaunch`, its `--run`; else the single other active run; else project scope. Terminal cards of the run are left exactly as they are.
- **REQ-CDI-016** (Unwanted behavior) — The retirement shall refuse, leaving the run `active` and every record unchanged, when any active card cannot be migrated: the named destination is not an active run, two or more other active runs exist and no destination was named, a check of REQ-CDI-009's form fails, or the database stays busy past the existing busy timeout. The refusal shall print each blocked card and its reason. The factory shall offer no flag, prompt or variable that retires a run holding an active card without migrating it; the owner-classification gate (a positive `dead` only) and the automatic reconciler's retirement write shall stay unchanged. **Where** `moai factory relaunch --from-run` meets such a refusal, it shall print `run <run>: left active (<N> active cards could not migrate)` followed by the blocked cards, shall still launch its session as it does for any refused retirement, and shall leave its exit status as it is for that case.

### C.5 Operation journal

- **REQ-CDI-017** (Ubiquitous) — On the identity layout, every queue claim, card grant, integration-window acquisition and git merge the factory performs shall carry an operation id minted before any effect and recorded in a durable journal in the factory database with its kind, card id, attempt number, actor, intent key, the expected readback of each external store it will touch, and the before-image needed to compensate. A queue claim is one operation across its assign and lease transitions, and each factory transition of an operation records its progress in the journal row inside that transition's own transaction. Where the effect precedes the factory-record change the row shall be committed before the effect begins; where the effect is the factory-record change the row shall commit in the same transaction.
- **REQ-CDI-018** (Event-driven) — **When** an operation's intent is presented again after its first presentation, in any journal state (a retry after an ambiguous failure, found through its intent key), the factory shall continue an in-flight row, return a `reconciled` row's recorded result, refuse with the recorded reason on a `superseded` or `blocked` row, and treat a `void` row (effect proven absent) as a new presentation; it shall never repeat an effect that a row records as landed, advance a version twice or append a second transition event. Two concurrent presentations of one intent shall resolve to one row.
- **REQ-CDI-019** (Ubiquitous) — A reconciliation pass shall decide each non-terminal journal row solely from the row, the current readback of the stores the row names and the card's current record, by the decision table in design.md § 6, as a pure function of those inputs: the same inputs give the same decision in any session and on any number of passes. It shall complete a half only when the readback proves that half missing and the other half landed, shall compensate only a half it can restore from the recorded before-image under compare-and-set, shall never repeat a git merge, and shall leave a row whose readback is indeterminate pending and report it.
- **REQ-CDI-020** (Capability gate) — **Where** an operator or lane session runs `moai factory reconcile`, or a claim or completion verb starts for a card with a non-terminal journal row, the factory shall run the reconciliation pass for that card (every card with `--all`) without consulting or requiring a leader session, a leader message, a leader approval or a coordinator run record.
- **REQ-CDI-021** (Ubiquitous) — Journal rows shall move only forward (`prepared`, then `applied`, then one terminal state of `reconciled`, `void`, `superseded` or `blocked`), a terminal row shall refuse update and delete, no update shall change a row's operation id, kind, intent key, card id or attempt number, at most one non-terminal row shall exist per intent key and no insert, whatever its conflict-resolution clause, shall be able to delete or rewrite an existing row, including by reusing its operation id, and no journal row shall ever be deleted, and no journal row shall exist or be written while the database is on the legacy layout.

### C.6 Published interface

- **REQ-CDI-022** (Ubiquitous) — The factory shall publish to later stages exactly these identity facts, readable without writing: the current record by card id alone; its `attempt_no`; its open membership run or project scope; the sealed attempts of a card in order; the journal rows of a card; the layout of the open database; and the rule that the operation-kind set is extended only additively. Existing external contracts shall change only by additive fields (`attempt_no` and `membership` on a card's JSON) and additive event kinds (`card.migrated`, `card.joined`, `identity.cutover`, `identity.rollback`, `operation.reconciled`).

### C.7 Attempt history

- **REQ-CDI-023** (Event-driven) — **When** any update of a current record changes a non-empty lease holder or a non-empty owner label, changes the record's run, or moves a terminal record to a non-terminal state, the store itself (a trigger on the current record, so that no writer can omit it) shall, in the same statement, append the pre-update record to the attempt archive as an immutable attempt and advance `attempt_no` by one, so that a displaced lease grant or owner stays readable (holder, lease expiry, owner, stage, evidence, version) after the next write overwrites the current record. This is one predicate for every writer: lease expiry, leaving a lease-holding state, reassignment, the dispatch owner restamp, migration, join and terminal re-pick all meet it. A grant (an empty holder or owner becoming non-empty), a lease renewal, a worktree binding, a hint update and an operator abandon (which keeps the holder columns and ends no grant) append nothing; the abandoned record's holder is sealed by the re-pick that later moves it out of the terminal state. The cost is one archive row per ended grant and no second transaction.

## §D Constraints and derivations

- **D-1 Test discipline.** A failing test is repaired in code or in the requirement. No test is skipped, no failing test is excluded from a run, and no timeout is widened as a remedy. A selector that sweeps nothing is a failed check (verification-completeness § 1.1).
- **D-2 SQL hygiene.** Every value rides a `?` placeholder; any concatenated fragment is a compile-time constant (the rule `updateCardRow` already states, `card_transition.go:876`).
- **D-3 Columns by name.** A fresh database and a migrated one list the 31 `cards` columns in different orders (the bundle columns arrive by `ALTER`; `factory.go:569-596`), so every copy and every comparison is by column name, never by position.
- **D-4 One transaction per structural act.** Cutover, rollback and retire-with-migration each run in one write transaction under the database's immediate transaction lock; no act is split into a copy step and a commit step.
- **D-5 Versions and the schema number.** The identity marker is carried beside the numeric base of the schema version string (design.md § 2.6), so the version advance of a later migration composes with it; this SPEC does not claim "the next number". A later migration that adds a column on the identity layout cannot alter the `cards` view and must follow the rule of design.md § 2.7.
- **D-6 No time language.** Plans and reports use ordering and priority, not durations. Configuration values that are durations (the lease keys) are untouched.
- **D-7 Operator-facing text** follows the neighbouring verbs: identifiers, states, flags and commands verbatim.
- **D-8 Scope allowlist.** The run touches only `internal/homestate`, `internal/config`, `internal/factorymsg`, `internal/factory`, `internal/cli/factory_*`, `internal/cli/mcp_factory_*`, `internal/cli/integration*`, the repository and template `workflow.yaml`, this SPEC directory and `.moai/reports/t1842`, and every branch commit names the card (AC-CDI-023).

## §E Dependencies, coordination and the published interface

### E.1 Cards and SPECs

| Item | Relationship | Status read |
|------|--------------|-------------|
| Stage A, card t1841, SPEC-CARD-LIVENESS-OWNERSHIP-001 | Re-planned after this stage; consumes the interface below. No dependency of this SPEC on its code. | Held at its plan-audit ceiling (iteration 4: FAIL 0.875 at `0496e206edce63ebcd02396e1d97d56d12948167`); SPEC unmerged in worktree t1841 (research § 3). |
| Card t1852, SPEC-FACTORY-RETIRE-GUARD-001 | Refuse-only retire guard for the legacy layout. Its §G moves "retire = transfer" to this card and calls its own refusal "a temporary safeguard that card is expected to revise". Both orders work: the legacy layout keeps whatever `main` has, the identity layout takes the migration branch. | Unmerged in worktree t1852 (research § 5). |
| Cards t1848, t1849, t1853 (split of t1825) | Run chooser and lane join (t1853), `cc -f` adopt (t1849). Not absorbed. | Not read beyond the t1852 SPEC. |
| Stage D, card t1835 | Consumes the interface (resource-lock rows keyed by card id and attempt number; the seal and migration events as release hooks). | No SPEC read. |
| Stage F, card t1843 | Consumes the interface (`started` evidence keyed by card id and attempt number; operation ids). | No SPEC read. |
| Stages C (t1839) and E (t1840) | Not prerequisites. Stage E's migration advances the schema number; D-5 and design.md § 2.6 make the two compose. | Not landed at B0 (`factorySchemaVersion = 6`). |
| SPEC-FACTORY-RECORD-001, SPEC-FACTORY-LEASE-RENEW-001, SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-MERGE-WINDOW-QUEUE-001, SPEC-INTEGRATION-LOCK-LIVENESS-001, SPEC-TODO-CLAIM-LEASE-001 | Behavior this SPEC must keep (REQ-CDI-002). | `completed` (read in `.moai/specs`). |
| SPEC-FACTORY-COMPLETION-RECOVERY-001 | Source of the dispatch-binding and approval semantics referenced in code comments. | `draft` (read); not retired or superseded. |

### E.2 The interface this SPEC publishes (REQ-CDI-022)

| # | Fact | Consumer use |
|---|------|--------------|
| S-1 | The card id is the only key; one current record per card id. | Stage A keys its grant by card id and needs no governing-row rule on the identity layout. |
| S-2 | `attempt_no` rises by one each time an attempt is sealed (REQ-CDI-023) and never falls; at the cutover the live record is numbered one above the sealed copies of the legacy rows. Each sealed row records the holder, lease expiry and opening state of the grant it ends. | Correlation id for grants, resource locks and `started` evidence; a grant keyed by card id can be checked against the sealed holder of its attempt. |
| S-3 | A migration, join, cutover or rollback carries lease holder, lease expiry, heartbeat and owner unchanged and bumps no epoch (REQ-CDI-014); a seal is the archive append and the bump only, and the write it accompanies (expiry, reassignment) clears the holder as at B0. | A migration is not a revocation; an epoch fence sees no cleared holder on a migration. |
| S-4 | No table keyed by card id that carries `owner_epoch` is deleted, lowered or rewritten by any act of this SPEC (REQ-CDI-014). | The "epoch never resets" invariant of the design document. |
| S-5 | Every write to the current record passes `updateCardRow` (six non-test callers at B0: `card_worktree.go:49`, `card_transition.go:435` and `:839`, `leader_approval.go:401`, `card_reassign.go:148`, `card_picked.go:194`) or `RenewLeaseUnderRoster`'s own update (`card_transition.go:527`); the seal does not depend on the caller because it is a store-level trigger (REQ-CDI-023). | An epoch check has the same two insertion points on both layouts, and the seal cannot be missed by a future caller. |
| S-6 | The recorded schema version string splits into a numeric base and an identity marker (design.md § 2.6); a column-adding migration follows design.md § 2.7. | The version advance of a later migration composes on either layout; a new table composes directly, a new `cards` column needs the § 2.7 steps. |
| S-7 | The open database reports its layout. | A stage can require the identity layout before its `enforce`-like step. |
| S-8 | The operation-kind set is closed to four kinds and extended only additively; an operation row carries card id and attempt number. | Stage D adds resource-lock kinds; stage F correlates `started` records. |

### E.3 Disposition of the stage-A plan-audit defects (iterations 1 to 4)

Defect ids are prefixed by iteration: `i1-D1` is iteration 1 defect D1. Source: `.moai/reports/t1841/plan-audit-iter1.md` to `plan-audit-iter4.md` in worktree t1841 (read in full for the defect lists).

| Defect | Seam content | Disposition here |
|--------|--------------|------------------|
| i1-D1 grant versus card consistency; the grant is keyed by card id, `cards` by `(run_id, card_id)`; the acquire edge does not check the cross-run live lease | The structural defect. | **Requirements.** REQ-CDI-012 and REQ-CDI-013 leave one current record per card id; REQ-CDI-004 to REQ-CDI-007 resolve the existing duplicates; the more-than-one-lease-holding-row classes and the stale-foreign-row discount cannot occur on the identity layout. **Not removed:** the single-row case (a grant that names lane A while the one card row is lane B's after a mode-off interval) is independent of row count; stage A keeps its holder and granted-at match on both layouts (§ G, S-2 gives it the sealed holder to compare). |
| i1-D2 schema-version coexistence and rollback; an older build refuses the migrated database | Compatibility of any additive step. | **Requirements.** REQ-CDI-010 (older build refused, stale handles aborted), REQ-CDI-011 (rollback keeps what was written), REQ-CDI-002; S-6 for composition with stage A and E. |
| i1-D6 migration criterion stated at a hard-coded schema version | Migration fixtures. | **Requirement-level constraint.** AC-CDI-006 builds its fixture from the run-entry tree's own schema, never a hard-coded number. |
| i2-N1 the consistency gate counted stale lease-holding rows of retired runs (measured 12 of 61 ids) | Stale foreign rows. | **Requirements.** Cutover seals every non-selected row as an attempt (REQ-CDI-006, REQ-CDI-007); a sealed attempt is not a lease-holding row for any reader (REQ-CDI-012). |
| i2-N3 rollback by the mode key does not hold for a busy lane | Clock compatibility, not identity. | **Out of scope here**, reason: this SPEC never changes the lease clock or its columns (REQ-CDI-014); the compatibility shadow is stage A's. |
| i3-D1 revocation reach is the governing row, not the card id | Revocation hook semantics. | **Requirements.** One current record is the governing row (REQ-CDI-012); migration writes no cleared holder (REQ-CDI-014, S-3). |
| i4-R4-1 a tie or unparseable heartbeat among same-label rows leaves an ambiguous clear non-durable | Two rows for one card id. | **Requirement.** The tie branch cannot arise on the identity layout (one current record, REQ-CDI-012). **Not removed:** an unparseable heartbeat on the single row still makes an ambiguous clear non-durable; stage A keeps a durable ambiguity mark on both layouts. On the legacy layout both branches remain until the cutover. Recorded as the stage-A obligation in design.md § 8.2. |
| i1-D3 to D5, D7 to D12; i2-N2, O1 to O6; i3-D2 to D6; i4-R4-2 to R4-5 | Shadow mode, rebind evidence, grace derivation, plan hygiene, probe semantics, identity-binding order, report population. | **Out of scope**, reason: none depends on the card key, the attempt, the membership or the journal; they are liveness-ownership content owned by stage A's re-plan. |

## §F Risks

- **R-1 The cutover touches every row.** It runs under the immediate transaction lock, so concurrent writers wait or fail busy; it must run in a quiet window with the key enabled deliberately (REQ-CDI-006). Residual: a writer that blocks past the busy timeout fails its own call.
- **R-2 A stale handle.** A connection opened before the cutover keeps its prepared view of the schema. REQ-CDI-010 makes its card writes abort; it can still write the other tables (`workers`, `runs`, `events`). The enable-time precondition (restart every process that opened the factory before the cutover) is stated, not mechanically proven.
- **R-3 Engine behavior was measured on the SQLite command line, not the project driver.** View-over-table name reuse by `CREATE TABLE IF NOT EXISTS`, `INSTEAD OF` triggers, and the three immutability triggers were observed on SQLite 3.54.0; the project's driver (`modernc.org/sqlite`) was not exercised at plan time (research § 6). AC-CDI-007 and AC-CDI-009 re-observe them under the driver.
- **R-4 The blocking rule blocks real data.** On the fresh copy of 2026-10-18T01:16:00Z the strict rule blocks 25 card ids by owner labels (18 with every such row in a retired run, 7 with one row in the active run) and at most 27 with the lease rule; the run-status rule it replaced blocked 0. The operator clears each blocked id before the cutover (plan.md § A.5), so the first cutover on live data is an operator task, not a command.
- **R-5 The journal is the largest and most separable part.** REQ-CDI-017 to REQ-CDI-021 do not change the identity layout and may land as a second part if the plan-audit ceiling forces a split (plan.md § C, cut between M5 and M6).
- **R-6 Per-tree configuration drift.** The key differs between worktrees; the layout does not (REQ-CDI-002).
- **R-7 Receipts are invalidated by a migration.** The leader re-approves (Q6); a card awaiting a done edge needs that extra act after a retire.
- **R-8 Sealed attempts still show lease-holder text.** No reader may treat an attempt as a lease; plan.md § I lists the readers and AC-CDI-011 asserts it.
- **R-10 The `events` log is mutable.** Displaced grants are preserved by the attempt archive (REQ-CDI-023), not by events; nothing here makes `events` append-only, and no reader of this SPEC's guarantees may rely on it.
- **R-9 Prerequisite state is read from other worktrees.** The stage-A and t1852 facts come from unmerged trees and a memory record; they are re-read at run entry (plan.md § B).

## §G Out of Scope

### Out of Scope — liveness-based ownership, owner-epoch fencing and shadow mode (stage A, card t1841)

- Grant identity, process probes, suspect and release states, the epoch counter itself, fenced writes, the shadow report and the `enforce` switch belong to stage A's re-plan. This SPEC only guarantees that nothing it does resets an epoch (REQ-CDI-014).
- Stage A's own plan-audit defects not listed as seam content in § E.3.

### Out of Scope — stage A's single-row consistency checks

- The grant's holder and granted-at match against the one card row, and a durable ambiguity mark for an unparseable heartbeat, stay stage A's on both layouts (reason: they are independent of how many rows a card id has). The seam is S-2: each sealed attempt carries the holder and opening state of the grant it ends, and `attempt_no` is a stable correlation key.

### Out of Scope — resource-keyed locks and the serial slot (stage D, card t1835)

- Replacing the serial slot, resource declarations, and releasing locks at decision gates.

### Out of Scope — dispatch `started` evidence (stage F, card t1843)

- Wiring or deleting the dispatch state machine.

### Out of Scope — run selection, lane join chooser, `cc -f` adoption and the legacy-layout retire guard (cards t1848, t1849, t1852, t1853)

- Choosing among several active runs for a join, adopting a run at leader start, and the refuse-only retire guard (SPEC-FACTORY-RETIRE-GUARD-001) are those cards'. The guard keeps governing the legacy layout.

### Out of Scope — metadata verbs and eligibility-filtered selection (stages E and C)

- `set-spec`, reclassification, evidence carry, and the selection filter.

### Out of Scope — automatic cutover and re-cutover tooling

- The cutover runs only by the explicit verb. A cutover after a rollback is a fresh cutover; archived identity records of a rolled-back layout are never merged into a new one.
- Purging the retained legacy table and rolled-back archives.

### Out of Scope — carrying leader approval receipts across a migration

- A receipt binds a run and a version; the leader re-approves after a migration (decision-index Q6).

### Out of Scope — changes to the queue store schema or the integration-lock record format

- The journal holds its own before and after images; the queue's current-dispatch table and the lock record gain no field.
- Unifying this journal with the queue store's dispatch operation journal (`gtd_operation.go`).

### Out of Scope — an MCP surface for the new verbs

- `identity report|cutover|rollback` and `reconcile` are CLI verbs. Existing MCP factory tools that share a verb's body inherit its behavior; no new MCP tool is added.

### Out of Scope — parallel merge queue, alarms and quarantine

- Making the integration window a parallel merge queue, work-item-age alarms (G1) and poison-card quarantine (G2).

## §H Traceability

The criteria carry their requirement mapping in `acceptance.md` (each criterion heading ends in `(maps REQ-CDI-…)`); this table is the same mapping read from the criterion side.

| Acceptance criterion | Requirement |
|----------------------|-------------|
| AC-CDI-001 | REQ-CDI-001 |
| AC-CDI-002 | REQ-CDI-002 |
| AC-CDI-003 | REQ-CDI-003 |
| AC-CDI-004 | REQ-CDI-004 |
| AC-CDI-005 | REQ-CDI-005 |
| AC-CDI-006 | REQ-CDI-006, REQ-CDI-007 |
| AC-CDI-007 | REQ-CDI-008 |
| AC-CDI-008 | REQ-CDI-009 |
| AC-CDI-009 | REQ-CDI-010 |
| AC-CDI-010 | REQ-CDI-011 |
| AC-CDI-011 | REQ-CDI-012 |
| AC-CDI-012 | REQ-CDI-013 |
| AC-CDI-013 | REQ-CDI-014 |
| AC-CDI-014 | REQ-CDI-015 |
| AC-CDI-015 | REQ-CDI-016 |
| AC-CDI-016 | REQ-CDI-006, REQ-CDI-007, REQ-CDI-013 |
| AC-CDI-017 | REQ-CDI-017 |
| AC-CDI-018 | REQ-CDI-018 |
| AC-CDI-019 | REQ-CDI-019 |
| AC-CDI-020 | REQ-CDI-020 |
| AC-CDI-021 | REQ-CDI-021 |
| AC-CDI-022 | REQ-CDI-022 |
| AC-CDI-023 | D-8 (the scope allowlist) |
| AC-CDI-024 | the baseline-first ordering of plan.md M0 (verification-claim-integrity § 2.3) |
| AC-CDI-025 | REQ-CDI-023 |

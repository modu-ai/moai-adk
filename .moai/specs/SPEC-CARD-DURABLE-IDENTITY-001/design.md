# Design: SPEC-CARD-DURABLE-IDENTITY-001

How the requirements in `spec.md` are met. Names below are the names the run phase uses; they are fixed here so that the acceptance ledger can cite them. Coordinates are tree B0 `f4bcd4f2cc65342559b505f1e5a29df80f992894` and are re-read at run entry (plan.md § B).

## 1. Vocabulary

| Term | Meaning |
|------|---------|
| legacy layout | The database as at B0: table `cards`, `PRIMARY KEY(run_id, card_id)`, `meta.schema_version` equal to the build's integer. |
| identity layout | The database after the cutover: `meta.schema_version` carries the marker (§ 2.6) and the objects of § 2.1 exist. |
| current record | The one mutable row per card id (table `card_current`). Its `run_id` is the open membership's run; the empty string means **project scope** (no open membership). |
| attempt | One lease grant or run engagement of a card, whichever ends first (REQ-CDI-023). When it ends, its ending state is appended to `card_attempts` as an immutable **sealed attempt**; the open attempt is the current record, numbered by `attempt_no`. |
| `attempt_no` | Per card, strictly increasing. At the cutover, a card id's k legacy rows become sealed attempts 1..k, numbered by the run's `created_at` then `run_id` (a numbering order, never a selection); the current record continues as attempt k+1. A rolled-back layout leaves a per-card floor (§ 7); the next cutover numbers above it, so `attempt_no` never falls across rollback and re-cutover. |
| membership | A row in `card_membership` linking a card's attempt to a run. At most one is open per card. |
| active card | A current record whose state is not `done`, `failed` or `abandoned` (`terminalStates`, `internal/homestate/card_record.go:74`). |
| operation | One cross-store act of kind `queue_claim`, `card_grant`, `lock_acquire` or `git_merge`, recorded in `factory_operations`. |
| intent key | A digest of what an operation is about to do (§ 6.1); two presentations of the same intent share it. |

## 2. Layout

### 2.1 Objects

| Object | Kind | Purpose and shape |
|--------|------|-------------------|
| `card_current` | table | The 31 columns of the legacy `cards` table (`internal/homestate/card_record.go:159-162` lists them) plus `attempt_no INTEGER NOT NULL`. `PRIMARY KEY(card_id)`. Index on `(run_id, state)`. |
| `card_attempts` | table | The same 31 columns plus `attempt_no`, `source` (`cutover` or `seal`), `sealed_at`, `seal_reason`, `seal_op_id`. `PRIMARY KEY(card_id, attempt_no)`. Never updated or deleted. |
| `card_membership` | table | `card_id`, `first_attempt_no`, `last_attempt_no`, `run_id`, `state` (`open` or `closed`), `closed_reason`, `recorded_at`. `PRIMARY KEY(card_id, first_attempt_no)`. A membership spans every attempt of one card in one run; at the cutover each legacy row gets its own (first equals last). Index on `(run_id, state)`. |
| `cards_pre_cutover` | table | The legacy table renamed, rows untouched, guarded by the same immutability triggers. Retained for rollback and for the count check; purging it is out of scope. |
| `cards` | view | Read-only projection of `card_current` in the legacy 31-column shape, with `INSTEAD OF` insert, update and delete triggers that abort with the text `factory identity cutover: this build predates the identity layout; upgrade moai`. |
| `factory_operations` | table | The journal (§ 6). |

A fresh database on the legacy layout creates none of these and `factoryDDL` (`internal/homestate/factory.go:22-160`) is unchanged: REQ-CDI-002.

### 2.2 Immutability (REQ-CDI-008, REQ-CDI-021)

`card_attempts`, `cards_pre_cutover` and terminal `factory_operations` rows carry three `BEFORE` triggers each: `UPDATE` and `DELETE` abort, and an `INSERT` whose primary key already exists aborts (this third trigger is what stops `INSERT OR REPLACE`, `REPLACE` and `UPDATE OR REPLACE`, which delete without firing a delete trigger). Observed on SQLite 3.54.0, command line: all six statement forms aborted with `sealed attempt is immutable` and the row survived (research.md § 6). The project driver (`modernc.org/sqlite` v1.60.1, SQLite 3.53.4) was exercised in a scratch program during the plan revision: the plan audit's probe aborted update, delete, duplicate insert, `INSERT OR REPLACE`, `REPLACE INTO`, `UPDATE OR REPLACE` and `INSERT OR IGNORE` on `card_attempts` (and update, delete and `INSERT OR REPLACE` on `cards_pre_cutover`) with the row count unchanged, and the revision's own probe re-confirmed update, delete and `INSERT OR REPLACE` on the rebuilt tables (research.md § 6).

`factory_operations` carries, besides the terminal-row triggers, a `BEFORE UPDATE` trigger that aborts when `NEW.op_id`, `kind`, `intent_key`, `card_id` or `attempt_no` differs from the old value. Without it an `UPDATE OR REPLACE` that renames a non-terminal row onto a terminal row's `op_id` deletes the terminal row without firing a delete trigger. Observed under the project driver: `UPDATE OR REPLACE factory_operations SET op_id='T' WHERE op_id='P'` aborted with `op identity immutable`, the terminal row stayed, and the forward move `prepared` to `applied` was allowed.

### 2.3 Legacy name reuse (REQ-CDI-010)

An older build's open runs `factoryDDL` first (`factory.go:317`), whose `CREATE TABLE IF NOT EXISTS cards` is a no-op while a **view** named `cards` exists (observed, SQLite 3.54.0), and then its version check (`factory.go:368`) refuses the marked value. A handle opened earlier keeps working on its prepared reads, and every write it issues to `cards` hits an `INSTEAD OF` trigger and aborts. The other tables it may still write (`workers`, `runs`, `events`) are the residual in spec R-2.

### 2.4 Columns are copied by name

Fresh and migrated databases list the 31 columns in different orders: the fresh DDL puts `bundle_id` and `bundle_order` before the contract columns (`factory.go:71-76`), while the live copy, migrated by `ALTER` (`factory.go:569-596`), lists them last (measured with `pragma_table_info`, research.md § 4). Every `INSERT … SELECT`, every comparison and every dump names columns.

### 2.5 Configuration

`workflow.card_identity.enabled` (bool, default `false`) in the repository and template `workflow.yaml`, a `CardIdentityConfig` in `internal/config/types.go`, its default in `internal/config/defaults.go`, and a loader beside `loadFactoryLease` (`internal/homestate/factory_lease.go:41-48`). A readable file whose key holds a non-boolean refuses the open (REQ-CDI-001); a file that cannot be read or decoded at all falls back to `false` with a warning, the direction that changes nothing.

### 2.6 The schema-version marker (REQ-CDI-002, REQ-CDI-010; spec D-5)

`meta.schema_version` is a string. A legacy value is the build's integer (`"6"` at B0, `factory.go:20`). An identity value is `"<base>+id1"`. Consequences:

- An older build compares the whole string with its integer and refuses (`unsupported factory schema version "6+id1"`, `factory.go:368-370`); no change to older builds is needed.
- This build splits the value into a numeric base and a marker, and its migration chain (`factory.go:338-367`) advances the base through a helper that re-attaches the marker, so a later additive migration (stage A or stage E) runs on either layout.
- The cutover and rollback change the marker only; they never change the base.

### 2.7 A later migration that adds a column (REQ-CDI-022; spec D-5)

`ALTER TABLE cards ADD COLUMN` fails on the identity layout because `cards` is a view (`SQL logic error: Cannot add a column to a view`, observed under the project driver). The rule for a later column-adding migration (existing examples: `factory.go:508`, `:588`): in one transaction, `ALTER TABLE` each of `card_current`, `card_attempts` and `cards_pre_cutover`, then `DROP VIEW cards` and recreate the view with the new column list and its three `INSTEAD OF` triggers. Observed under the project driver: adding a column to `card_current` while the view exists succeeded, and the view and trigger were recreated in one transaction; a stale handle's insert through the rebuilt view still aborted. A migration that adds a new table needs none of this.

## 3. Cutover

### 3.1 `moai factory identity report` (REQ-CDI-003)

Opens with `OpenFactoryReadonly` (`factory.go:237`; `mode=ro`, no DDL) and prints, in this order: counts (legacy rows, distinct ids, duplicated ids); per duplicated id the candidate rows (run, state, owner label, lease holder, lease expiry, run status); blocked ids (§ 3.3 C-2); the unreconciled record-unavailable entries (`card_unavailable.go:18`); the resolution input. The proposal rule P: an id is pre-filled when it has exactly one non-terminal row, else when it has exactly one non-terminal row in an `active` run; otherwise blank. Terminal-only duplicates are blank. The invariant is that the `factory.db` file and the content of every table are unchanged: the read-only open (`mode=ro`) creates `-wal` and `-shm` sidecar files even on a quiescent database (observed under the project driver: a copy with no sidecars gained `f.db-shm` and `f.db-wal`, the `f.db` bytes equal), so those two files are permitted and the test hashes `factory.db` and a dump of every table. An `immutable=1` open creates no sidecar but is unsafe beside a live writer, so it is not used.

### 3.2 The resolution input

One JSON object per line: `{"card_id":"…","run_id":"…"}`. `report --emit <path>` writes it with blanks as `"run_id":""`. `cutover --resolutions <path>` reads it. Checks C-3 and C-4 below apply to it.

### 3.3 `moai factory identity cutover` (REQ-CDI-004 to REQ-CDI-007, REQ-CDI-009)

Preconditions, checked before the transaction opens and again inside it:

- C-1 the key is true, the database is on the legacy layout, the session is not a lane session (`factoryLaneRefusal`, `factory_card.go:120`);
- C-2 (a) no card id has non-terminal rows carrying two or more distinct non-empty owner labels (`owner_label` or `lease_holder`); (b) no non-terminal row that is not its card id's resolved row holds an unexpired lease (parsed against the transaction's single `now`; an unparseable expiry counts as unexpired) or a lease holder that is a `workers` row bound to that row's run; (c) the record-unavailable log has no unreconciled entry. The report lists owner labels and lease expiries so the operator can release or abandon the rows named; the lease reading is a precondition, not an ownership decision;
- C-3 every duplicated id has a resolution naming one of its own legacy rows;
- C-4 for an id with a non-terminal row, the named row is non-terminal.

Steps inside one write transaction (the DSN takes the lock at `BEGIN`, `factory.go:303`):

1. `ALTER TABLE cards RENAME TO cards_pre_cutover`; create the § 2.1 objects and triggers.
2. `INSERT INTO card_attempts … SELECT` every legacy row, `attempt_no` from `ROW_NUMBER() OVER (PARTITION BY card_id ORDER BY runs.created_at, run_id)` plus the card's floor from any `card_attempts__rb*` archive, `source='cutover'`, `seal_reason='cutover-archive'`.
3. For each distinct card id, `INSERT INTO card_current` from the resolved legacy row (all 31 columns), `attempt_no = floor+k+1`; if the resolved row's run has status `retired` and the row is non-terminal, set `run_id=''` and `version=version+1` and append a `card.migrated` event (`reason: cutover-run-retired`).
4. One `card_membership` row per sealed attempt (`closed`, `cutover-archive`) and one `open` row per current record that has a run.
5. Create the `cards` view and its three triggers.
6. Validate (V-1 to V-7).
7. Set the marker and append one `identity.cutover` event whose payload carries the counts and the SHA-256 of the resolution input.
8. Commit.

Validation, all inside the transaction (REQ-CDI-009); each failure names its check, rolls back, and exits non-zero:

| Check | Statement of the check |
|-------|------------------------|
| V-1 | `count(card_attempts) = count(cards_pre_cutover)` |
| V-2 | `count(card_current) = count(distinct card_id in cards_pre_cutover)` |
| V-3 | no `card_attempts.card_id` lacks a `card_current` row |
| V-4 | every `card_membership` row spans attempt numbers that exist (in `card_attempts` or the current record) and names a `runs` row |
| V-5 | every `leader_approvals` row whose `(run_id, card_id)` matched a legacy row before still matches a sealed attempt; the set of approval and dispatch references that did not resolve before is equal to the set after |
| V-6 | for every legacy row, the sealed attempt of the same `(run_id, card_id)` equals it column by column |
| V-7 | for every `card_current` row not moved to project scope, every one of the 31 columns equals the resolved legacy row's; for every moved row, all columns equal it except `run_id` (empty) and `version` (one higher) |

Fault points for tests: `cutover.after-copy`, `cutover.before-validate`, `cutover.before-marker`, `cutover.before-commit`, through a package-level seam of the shape `cardTransitionFault` already has (`card_transition.go:231-238`).

## 4. Records on the identity layout

### 4.1 Statement routing

Every one of the 15 statements that name `cards` in non-test code (`grep` at B0: plan.md § I) exists in two constant forms, the legacy table and `card_current`, selected by the layout the handle recorded at open. No table name is built from input. `card_current` keeps the legacy column names, so `WHERE run_id=? AND card_id=? AND version=?` (`card_transition.go:882`) keeps its text and now means "this card, and this run is its open membership, and this version".

### 4.2 Reading a card in a run

`loadCard(ctx, q, runID, cardID)` (`card_record.go:179`) reads by card id; a row whose `run_id` differs from `runID` is returned as a typed error `ErrCardNotMember` naming the open membership run or project scope; no row at all stays `ErrCardNotFound`. `ListCards(ctx, runID)` (`:195`) lists by membership; with an empty run it lists every current record.

### 4.3 Recording a card in a run (REQ-CDI-013)

`recordPickedTx` (`card_picked.go:159`) gains three cases ahead of its INSERT: no current record (as today, attempt 1, version 1); a non-terminal record with an open membership elsewhere (refuse `ErrCardEngaged`, naming the run and `moai factory runs --retire`); a non-terminal record in project scope (join, § 4.5); a terminal record (seal, then a new attempt in `picked`, `version = sealed.version + 1`). The `factory next` arms map `ErrCardEngaged` to a real refusal, not to the "race, re-select" retry that `factoryNextClaimRefused` (`factory_card.go:1197-1202`) gives stale-version errors.

### 4.4 Seal

A seal runs inside the writing transaction and before the write changes the record, in these funnels: `applyLeaseExpiry` (`card_transition.go:422`); `commitTransition` (`:826`) when the next state holds no lease or the run changes; `ReassignCard` (`card_reassign.go:50`); `recordPickedTx` on a terminal record (`card_picked.go:159`); and the migration and join writers (§§ 4.5, 5). A lease renewal (`RenewLeaseUnderRoster`, `:489`) seals nothing. The cost of ending a grant is one `INSERT … SELECT` of the 31 columns plus the five archive columns and, when the run changes, one membership row, in a transaction that already holds the write lock: no second transaction, and the seal commits or rolls back with the version compare-and-set. Steps: copy the current record into `card_attempts` (`source='seal'`, reason, the operation id if any), close the membership when the run changes (`closed_reason`), set `attempt_no = attempt_no + 1` on the current record, apply the change (new run, or new state), `version = version + 1`, open the next membership when the run changes, append the event. Nothing in a seal touches `lease_holder`, `lease_expires_at`, `heartbeat_at` or `owner_label` (REQ-CDI-014).

### 4.5 Join

For a project-scope record: set `run_id`, `version+1`, open a membership, upsert `card_dispatch` (`leader_approval.go:290`) and `ensureRunRowTx` (`:280`) exactly as the lease-acquire edge already does (`card_transition.go:596-601`), append `card.joined`. No other column changes.

### 4.6 Cross-run readers

`predecessorMerged` (`card_picked.go:274`) reads the current record and every sealed attempt, so a predecessor that merged in an earlier attempt still counts. `PreviousCardWorktree` (`card_worktree.go:73`) stays for the legacy layout and finds nothing on the identity layout, where the current record carries `worktree_path` itself. `liveLeaseCrossRunQuery` (`card_reassign.go:23`) is unreachable on the identity layout: one record per card id cannot hold two leases.

## 5. Retirement as migration (REQ-CDI-015, REQ-CDI-016)

`RetireRunIfDead` (`factory_run_retire.go:381`) keeps its owner classification and its refusal when the owner is not `dead`. After that gate, on the identity layout, it calls `retireRunWithMigration(ctx, runID, dest)` instead of `retireRun` (`:407`). One transaction:

1. Read the run `active` (a changed status refuses).
2. Resolve the destination: `--into`, else for `relaunch` its `--run` (`factory_relaunch.go` builds `run`), else the single other `active` run, else project scope; two or more survivors without a destination refuse (the same ambiguity `ResolveActiveRun` reports, `factorymsg/store.go:372-380`). A named destination is read inside the transaction and must be `active` and different from the source.
3. For each active card (membership open on the run, state non-terminal), in `card_id` order: seal (§ 4.4) with reason `retire-migrate`; set `run_id` to the destination or `''`; `UPDATE workers SET run_id=<dest> WHERE label=<owner or lease holder> AND run_id=<run>` (`''` when the destination is project scope); re-point `card_dispatch` when there is a destination; append `card.migrated` with both run ids and both attempt numbers.
4. Run V-1 to V-5 over the migrated set (counts of sealed seals, membership consistency, references).
5. `UPDATE runs SET status='retired'` and append `run.retired` (the existing event) with the classification and basis.

Any step failing returns the typed refusal listing each blocked card and its reason; the transaction rolls back and the run stays `active`. Terminal cards are not touched. `retireRunUnlessLive` (`:431`) and `ReconcileActiveRuns` (`:339`) are unchanged: the automatic path still refuses a run that holds live records.

## 6. The operation journal

### 6.1 Rows

`factory_operations(op_id TEXT PRIMARY KEY, kind TEXT, intent_key TEXT, card_id TEXT, attempt_no INTEGER, run_id TEXT, actor TEXT, state TEXT, progress TEXT, expect_json TEXT, before_json TEXT, after_json TEXT, result_json TEXT, created_at TEXT, updated_at TEXT)`. `state` is one of `prepared`, `applied`, `reconciled`, `void`, `superseded`, `blocked`. A partial unique index on `intent_key WHERE state IN ('prepared','applied')` keeps one in-flight row per intent. Triggers enforce forward-only movement, refuse update and delete of a terminal row, and refuse any update that changes `op_id`, `kind`, `intent_key`, `card_id` or `attempt_no` (§ 2.2). The operation id is minted by the caller (random 128 bits, hex); decisions never read a clock (REQ-CDI-019).

| Kind | Intent key | Effect (external store) | Readback | Before and after image | Factory half |
|------|------------|-------------------------|----------|------------------------|--------------|
| `queue_claim` | kind, card, run, owner, expected version | the queue's claim and current-dispatch record (`todo_dispatch_current.go:31`, `backlog_store.go:973`), read through the non-adopting `LoadPure` | the record names the run and owner | run and owner before and after | the assign and lease transitions as one operation: the assign transition's transaction sets `progress='assign-landed'` in the row, the lease transition's moves it to `applied` (`factory_card.go:1173-1191`) |
| `card_grant` | kind, card, attempt, lane, expected version | none (factory record only) | the current record's holder and version | holder and version | the lease transition itself; the row commits in the same transaction, inserted as `applied` |
| `lock_acquire` | kind, session, card, branch | the integration-window record (`integration_lock.go:355`) | `ReadIntegrationLock` (`:251`) names the session, card and branch | prior holder | none |
| `git_merge` | kind, card, attempt, base tip, card tip | the merge commit on the integration branch (`integration_merge_step.go:178`) | the integration tip's parents include the card tip, and a valid re-measure record exists for its tree (the adoption probe, `factory_card.go:2229-2238`) | base tip and expected merge tip | `merge-ready → merging → merged-local` (`completeTransitions`, `factory_card.go:2356`) |

No field is added to the queue's table or to the lock record: the journal holds its own images.

### 6.2 Protocol

An operation whose effect precedes its factory change: commit the row as `prepared` (own transaction), perform the effect, then commit the factory change and the row's move to `applied` together. An operation whose effect is the factory change: one transaction inserts the row as `applied`. Relation to the existing ordering: today the queue write is the last act inside each factory transaction (`followRecord` as `BeforeCommit`, `factory_card.go:1170-1172`). The new protocol commits the `prepared` row first and then lets the same hook write the queue record inside each transition, so a crash between the queue write and the factory commit leaves exactly the orphan window the table decides, on purpose.

A second presentation looks the intent key up in every state (REQ-CDI-018): an in-flight row is continued through § 6.3; a `reconciled` row returns its `result_json`; a `superseded` or `blocked` row refuses with the recorded reason; a `void` row, whose effect was proven absent, is a new presentation and mints a new row. The partial unique index over in-flight rows makes two concurrent presentations resolve to one row; the loser re-reads and continues it.

### 6.3 Decision table (REQ-CDI-019)

Fourteen cells; each has a subtest of the same name in AC-CDI-019.

| Cell | Row state | Readback | Decision and writes |
|------|-----------|----------|---------------------|
| `queue_claim__after_prepared` | prepared | queue shows before-image; factory unchanged | `void`; no write |
| `queue_claim__after_effect` | prepared | queue shows after-image; factory unchanged | roll the factory half forward at the recorded expected version; landed: `reconciled`; refused (stale version or holder): compensate the queue by compare-and-set to the before-image, then `void`; queue no longer shows the after-image: `superseded`, no write |
| `queue_claim__assigned_only` | prepared, `progress='assign-landed'` | queue shows after-image; factory `assigned` at the recorded assign version, owner equal to the row's | roll the lease edge forward at the recorded lease version: landed: `reconciled`; refused (stale version, holder or guard): `superseded` with result `assigned-only`, no write (the queue after-image is consistent with state `assigned`, so nothing is compensated; the lane retries through `next --card`) |
| `queue_claim__after_record` | applied | factory landed | queue shows after-image: `reconciled`; before-image: re-apply the idempotent queue write, `reconciled`; another run named: `superseded`, no write |
| `card_grant__client_lost` | applied | current record shows the grant | `reconciled` |
| `card_grant__superseded_by_later_write` | applied | holder differs or the attempt is sealed | `superseded` |
| `lock_acquire__after_prepared` | prepared | lock record does not name the session | `void` |
| `lock_acquire__after_effect` | prepared | lock names session, card, branch | `reconciled` |
| `lock_acquire__held_by_other` | prepared | lock names another session | `superseded`; the lock is never touched |
| `git_merge__after_prepared` | prepared | integration tip equals the base tip | `void` |
| `git_merge__after_effect` | prepared | merge landed, card `merge-ready`, re-measure record valid | roll `merge-ready → merging → merged-local` forward: `reconciled`; re-measure invalid: `blocked` and the existing post-merge hold (`CompletePostMergeHold`, `factory_card.go:2396`); the merge is never reverted |
| `git_merge__card_merging` | prepared | merge landed, card `merging` | roll `merging → merged-local` forward: `reconciled` |
| `git_merge__after_record` | applied | card `merged-local`, its merge SHA reachable from the integration branch | `reconciled` |
| `any__indeterminate_readback` | any | a store cannot be read | the row stays as it is; reported as pending; exit non-zero for `--all` |

A merge that is neither landed nor proven absent (a ref that cannot be resolved) is the last cell. No cell repeats a git merge or re-acquires a lock.

### 6.4 Invocation without a leader (REQ-CDI-020)

`moai factory reconcile [--card <id>|--all]` and the pre-step of `factoryNextClaim` (`factory_card.go:1140`) and `factoryCompleteCard` (`:2063`). The reconcile code (`operation_reconcile.go`) imports no leader-approval, broker or coordinator-run symbol; AC-CDI-020 greps for that. Rows are processed in `(created_at, op_id)` order; an unreadable store marks only the rows that name it indeterminate.

### 6.5 What stays as it is

`record-unavailable.jsonl` and its `record.drift` reconciliation (`card_unavailable.go:147`), and the queue store's dispatch operation journal (`gtd_operation.go:144`), are unchanged. The cutover requires the first to be empty of unreconciled entries because its entries carry `(run, card)` keys.

## 7. Rollback (REQ-CDI-011)

`moai factory identity rollback`, one transaction: refuse in a lane session, on the legacy layout, or while a journal row is non-terminal; drop the view and its triggers; create a fresh `cards` table of the legacy shape; fill it with one row per `(run_id, card_id)` pair taken from the highest-`attempt_no` source of that pair among the sealed attempts and the current record (so a state written after the cutover wins over its cutover snapshot, and the older rows stay in the archive); verify R-1 pairs rebuilt equal pairs known, R-2 every current record is represented, R-3 no version is lower than its `cards_pre_cutover` counterpart, R-4 every pair untouched since the cutover equals its `cards_pre_cutover` row by column; rename each identity table to `<name>__rb<N>`, then for each renamed table `DROP TRIGGER` and `DROP INDEX` its objects and recreate them under `__rb<N>`-suffixed names (SQLite keeps trigger names across `ALTER TABLE … RENAME`, so the plain names would otherwise block the next cutover; observed under the project driver, and the drop-and-recreate form left six distinct trigger names, the new `card_attempts` and the archive both refusing update, delete and replace); write each card's highest `attempt_no` into a floor kept in the `__rb<N>` archive (the next cutover reads `max(attempt_no)` per card from every `card_attempts__rb*` and numbers above it); strip the marker; append `identity.rollback`.

## 8. The seam with stage A and with later stages (REQ-CDI-014, REQ-CDI-022)

### 8.1 What each fact lets a consumer drop

| Published fact | Stage-A text it removes on the identity layout |
|----------------|-----------------------------------------------|
| S-1 one current record per card id | the governing-row definition, the "two live lease-holding rows across runs" classification, the stale-foreign-row discount (REQ-CLO-024 as audited), and the revocation-reach rule (REQ-CLO-006 as corrected) |
| S-3 migration, join and seal carry lease columns unchanged | any revocation on retirement; no epoch advance on a migration |
| S-4 no epoch-carrier row is touched | the "epoch must never reset across re-keying and run retirement" obligation becomes a property of this layer |
| S-5 unchanged write funnels | the fence's insertion points |
| S-6 marker-aware migration helper | the "next schema number" race with stage E |

### 8.2 Obligations this SPEC cannot discharge for stage A (stated, not imposed)

- S-1 removes only the duplicate-row classes: the more-than-one-lease-holding-row cases (i1-D1(b), i2-N1, i3-D1) and the tie branch of i4-R4-1. It does **not** remove i1-D1(a), a grant that still names lane A's session while the one card row is lane B's, nor the unparseable-heartbeat branch of i4-R4-1 on a single row. Stage A keeps, on both layouts, the holder and granted-at match and a durable ambiguity mark; S-2 gives it the sealed holder and `attempt_no` to compare. On the **legacy layout** the duplicate-row classes remain until the cutover; stage A's re-plan either keeps its gate there or requires the identity layout before its `enforce` step (S-7).
- A grant keyed by card id needs no `attempt_no`; stage A may store it as correlation. Neither is required.
- A migration that moves a lease-holding card does not change its holder, so a liveness evaluator that compares holder and session sees the same grant; whether the moved card's process identity still matches is stage A's concern.

### 8.3 Later stages

| Stage | Reads from this SPEC | Does not get |
|-------|----------------------|--------------|
| D (t1835) | card id, `attempt_no`, `card.sealed` and `card.migrated` events as lock-release hooks, the `lock_acquire` kind as the pattern for resource-lock kinds | a resource-lock table |
| F (t1843) | `(card_id, attempt_no)` as the key of `started` evidence; operation ids for correlation | the dispatch machine's wiring |

## 9. Compatibility and coexistence

- **Older builds:** refused at open (§ 2.6); a pre-opened handle's card writes abort (§ 2.3).
- **Read-only openers** (`OpenFactoryReadonly`) work on both layouts; on the identity layout `FactoryTablePresent` (`factory.go:277`) gains the new table names in its allowlist and queries `type IN ('table','view')` for `cards`, because it is a view there (the present query, `factory.go:288`, answers `type='table'` and would report it absent).
- **MCP factory tools** share the verb bodies (`mcp_factory_card.go`) and inherit the behavior; no new tool.
- **`moai update`** syncs the template `workflow.yaml` key; the cutover is never run by it.

## 10. Alternatives considered

| Alternative | Why not |
|-------------|---------|
| Re-key `cards` in place to `PRIMARY KEY(card_id)` and drop duplicates | Destroys rows; the card makes row preservation the core. |
| Keep `PRIMARY KEY(run_id, card_id)` and add an identity table with a pointer | Leaves duplicate live rows; stage A's consistency gate stays necessary. |
| An attempt only per run engagement | Overwrites the displaced lease grant in the current record; the previous holder would survive only in the mutable `events` log. Rejected by the lane ruling (Q2) in favour of design principle 6. |
| Cutover at first open when the key is true | A write transaction over every row while other processes hold handles; no quiet-window control (Q4). |
| Layout chosen by the config key at every open | Worktrees carry their own copy of the tracked file and share one database (Q1). |
| Application-level immutability only | A second connection or an older build bypasses it; a trigger is enforced by the store. |
| A journal file beside the database | The `card_grant` row must commit with the grant; two files cannot (Q5). |
| Drop the legacy table after copying | Rollback and the count check lose their reference. |

## 11. Failure modes

| Failure | Result |
|---------|--------|
| Crash or error inside the cutover transaction | Nothing committed; layout, version and rows as before. |
| Database busy past the busy timeout | The verb fails; nothing partial. |
| A process with a pre-cutover handle writes a card | The statement aborts with the cutover text. |
| The same process writes `workers`, `runs`, `events` | Succeeds (residual, spec R-2). |
| The destination run is retired between `report` and the retirement | The in-transaction check refuses; the source run stays `active`. |
| A lane renews a migrated card naming the old run | `ErrCardNotMember` naming the new run; its roster entry was re-bound, so the retry with the new run passes. |
| Two sessions reconcile the same row | The first's forward move wins; the second's compare-and-set or forward-only trigger refuses and it re-reads. |
| A journal row stays `prepared` | Reported by `reconcile --all`; its decision is a pure function of the stores, so a later pass decides it. |
| The record-unavailable log gains an entry after the precondition check | Caught by the in-transaction re-check; the cutover refuses. |
| A claim dies after the assign transition commits | The `queue_claim__assigned_only` cell: the next pass rolls the lease edge forward or records `assigned-only`. |
| A rollback and a second cutover | Trigger and index names differ (§ 7); attempt numbers continue above the floor. |

## 12. File plan

`internal/homestate/identity_layout.go` (marker, layout, helper), `identity_cutover.go`, `identity_rollback.go`, `identity_report.go`, `card_seal.go`, `factory_run_retire.go` (new branch), `operation_journal.go`, `operation_reconcile.go`; `internal/config/types.go`, `defaults.go`; `internal/cli/factory_identity.go` (`newFactoryIdentityCommand`), `factory_reconcile.go` (`newFactoryReconcileCommand`), `factory_handoff_recover.go` (`--into`), `factory_relaunch.go`; `internal/factory` and `internal/cli/factory_card.go` for the operation hooks at the four effect sites; the two `workflow.yaml` files; tests `internal/homestate/identity_*_test.go`, `operation_*_test.go` and `internal/cli/factory_identity_*_test.go`, prefix `TestCDI_`.

## 13. Assumptions

- A-1 `ROW_NUMBER()` and views with `INSTEAD OF` triggers behave under `modernc.org/sqlite` as on the command line (AC-CDI-007, AC-CDI-009 observe it).
- A-2 No process other than the factory verbs writes `cards` (the grep of § 4.1 is the evidence; a script outside the repository could).
- A-3 `runs.created_at` is non-empty and distinct per run (observed on the live copy: 37 runs, 0 empty, 0 shared values); the numbering order falls back to `run_id` when two runs share a value.
- A-4 A lane's roster entry names the run it joined (`workers.run_id`, `factory.go:24-32`).

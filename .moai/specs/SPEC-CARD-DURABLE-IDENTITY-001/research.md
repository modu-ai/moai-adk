# Research: SPEC-CARD-DURABLE-IDENTITY-001

Evidence for the plan. Tree B0 `f4bcd4f2cc65342559b505f1e5a29df80f992894` unless a section names another source. Every file and line below was read in this run; where a source is another worktree or a memory record, the section says so.

## 1. Findings in the code

| # | Finding | Source |
|---|---------|--------|
| F1 | `cards` is `PRIMARY KEY(run_id, card_id)`; 31 columns; no epoch column. | `internal/homestate/factory.go:45-78`; `card_record.go:100-138`; E12 |
| F2 | `factorySchemaVersion` is 6; the open path runs `factoryDDL`, then reads `meta.schema_version`, migrates 1 to 6 in order, then refuses any value other than its integer. A read-only open (`OpenFactoryReadonly`) runs no DDL and no version check. | `factory.go:20`, `:292-387`, `:237-272` |
| F3 | Every card write is addressed by `run_id AND card_id AND version`; the three write statements are the renew update, `updateCardRow` and the picked insert. | `card_transition.go:527`, `:875-899`; `card_picked.go:166` |
| F4 | `recordPickedTx` inserts a fresh `picked` row at version 1 when the `(run, card)` pair is absent, so a run replacement restarts the card. The only carry-over is the worktree binding, found by `PreviousCardWorktree`; its doc comment states the cause ("Card rows are keyed per run, so a replacement run's row is born without the binding its predecessor held"). | `card_picked.go:159-170`; `card_worktree.go:63-73`; `factory_card.go:528-544` |
| F5 | The lease-acquire edge checks registration and owner label only; the cross-run live-lease check exists only in reassignment. | `card_transition.go:582-590`; `card_reassign.go:19-23`, `:92` |
| F6 | `RetireRunIfDead` retires "regardless of the records the run carries"; the automatic path refuses a run with live records. Callers of the operator path: `factory runs --retire` and `factory relaunch --from-run`. | `factory_run_retire.go:381-422`, `:431-456`; `factory_run_adopt.go:31-38`; `factory_handoff_recover.go:91`; `factory_relaunch.go:98` |
| F7 | The lane claim writes the queue's current-dispatch record from a `BeforeCommit` hook, inside the factory transaction, right before commit. | `factory_card.go:1170-1191`; `internal/factory/todo_dispatch_current.go:31` |
| F8 | The integration window is a JSON record with additive optional fields; the merge is a git operation; a partly-done merge is recognised only by the adoption probe in `factoryCompleteCard`; a transition failing after the merge is handled by `completePostMergeConflict`. | `internal/factory/integration_lock.go:122`, `:251`, `:355`; `integration_merge_step.go:178`; `factory_card.go:2229-2258`, `:2395-2402` |
| F9 | Failed dispatch-mirror writes are logged in `record-unavailable.jsonl` and reconciled on the next successful write; the queue store has its own operation journal for dispatch (`prepared`, `invoking`, `reconciled`, `blocked`). | `card_unavailable.go:18`, `:147`; `internal/factory/gtd_operation.go:16-21`, `:144` |
| F10 | `ReconcileActiveRuns` and `retireRunUnlessLive` are the only automatic retirement writers; `ResolveActiveRun` calls the reconciler. | `factory_run_retire.go:339`, `:431`; `internal/factorymsg/store.go:383` |
| F11 | `relaunch --from-run` retires, then launches into `--run`; its help says it "does not migrate card leases or worktrees". | `factory_relaunch.go` (verb long description; `relaunchRetireFromRun`) |
| F12 | A leader approval binds run, version, evidence hash and owner; `VerifyBinding` compares the run. | `leader_approval.go:116-142` (run check `:126-128`) |
| F13 | `card_dispatch` is already keyed by card id (`PRIMARY KEY(card_id)`) and is re-pointed by the assign, lease and audit-kickoff edges. | `factory.go:155-159`; `card_transition.go:575-580`, `:596-601`, `:667-674`; `leader_approval.go:290` |
| F14 | 15 lines in 9 files name `cards` in a statement in non-test code; 3 are writes. | plan.md § I (the grep) |
| F15 | `workflow.factory_lease` is the precedent for a key checked when the factory opens; a decode failure of the whole file falls back to defaults with a warning, a bound violation refuses. | `internal/homestate/factory_lease.go:41-80` |
| F16 | The lane-refusal guard that refuses `factory decide` in a lane session exists. | `factory_card.go:110-146`, `:3255-3277` |
| F17 | `runs --retire` appears in no rule, skill, `.moai/docs` file or `AGENTS.md`; five Go message sites tell an operator to retire a run. | `grep -rln "runs --retire" .claude internal/template/templates .moai/docs AGENTS.md` printed nothing; `grep -rn ... internal --include='*.go' --exclude='*_test.go'` printed `factory.go:500`, `:545`, `:687`, `factory_run_adopt.go:103`, `factory_slots.go:142` |

## 2. Discrepancies

- **Stage order.** The design document's delivery table lists stage A before stage B (A is B's prerequisite). Stage A's hold record states "stage A is planned again after stage B (t1842) (order B -> A)" (leader message 2026-10-17, board `d-20261017T192203Z-0c4c`, standing `9e76`), read in `.moai/worktrees/t1841/.moai/reports/t1841/progress.md`. This SPEC follows the hold record and depends on no stage-A code. The task for this SPEC still called stage A a prerequisite; the seam is therefore written from this side.
- **Stranded count.** The design document and the t1852 SPEC speak of 41 cards stranded on 10-17; the live copy of 2026-10-18T00:31:45Z holds 78 non-terminal rows in run `tmvgvn` (§ 4). The copy was taken later; no claim is made that either number is the other's error.
- **Design document figures.** "~73 transition rows" and "21 card states" are the design document's; the code comments say 71 accepted pairs (`card_transition.go:130`) and `CardStates` documents 21. Neither number is used by a requirement here.

## 3. The stage-A plan-audit trail (what the seam cost)

Source: `.moai/reports/t1841/plan-audit-iter1.md` to `plan-audit-iter4.md` in worktree t1841. Iteration 4 and `progress.md` were read in full; for iterations 1 to 3 the defect list headlines and their opening sentences were read (the full bodies were not).

| Iteration | Verdict | Blocking | Note |
|-----------|---------|----------|------|
| 1 | FAIL 0.81 | D1 to D6 | i1-D1 is the card-key seam. |
| 2 | FAIL 0.875 | N1 to N3 | i2-N1 counted stale foreign rows. |
| 3 | FAIL 0.875 | D1, D2 | i3-D1 revocation reach. |
| 4 | FAIL 0.875 at `0496e206edce63ebcd02396e1d97d56d12948167`, must-pass 9/9, codex gate `fail` | R4-1, R4-2 | R4-1 tie or unparseable heartbeat. Hold: no further round. |

The disposition of every defect against this SPEC is spec.md § E.3. The pattern worth recording: three of the four blocking defects that survived to iteration 3 and 4 (i1-D1, i3-D1, i4-R4-1) are the same fact seen from three angles, that one card id can have several lease-holding rows; each correction added a rule that made the next one necessary. Stage A's SPEC reports 12 of 61 card ids carrying more than one lease-holding row, and the iteration-4 audit (its Claim 4, on its own copy of the live database) found 12 of 62, 9 of the 12 repeating one holder label.

## 4. Measurements on a copy of the live factory database

Taken 2026-10-18T00:31:45Z: `cp /Users/goos/.moai/db/moai-adk-go-1bd3d038/factory/factory.db` into the scratchpad (the `-wal` and `-shm` files did not exist at copy time; the file is the checkpointed state). Queries with `sqlite3` on the copy. A one-instant shape, not a rate.

| Query | Output |
|-------|--------|
| `select value from meta where key='schema_version'` | `6` |
| rows in `cards` / distinct `card_id` / card ids with more than one row | 293 / 221 / 55 |
| `runs` by status | `active` 1 (`tn1p4e`), `retired` 36 |
| `count(*) from pragma_table_info('cards')` | 31 |
| non-terminal rows / of them in retired runs / in the active run | 190 / 143 / 47 |
| card ids with two or more non-terminal rows | 38 (17 with exactly one in the active run, 21 with all in retired runs) |
| of those 38, ids with two or more distinct non-empty owner labels | 25 (18 with all such rows in retired runs, 7 with one row in the active run) |
| card ids with non-terminal rows in two or more active runs | 0 |
| card ids whose non-terminal rows are only in retired runs | 93 |
| non-terminal rows of run `tmvgvn` | 78 of its 91 rows |
| duplicated ids: pre-filled under the proposal rule / blank | 29 / 26 (12 with one non-terminal row, 17 with exactly one in the active run; 21 with all non-terminal rows in retired runs and 5 all-terminal are blank) |
| rows with a lease-holding state / of them in retired runs | 83 / 52 |
| most rows for one card id; ids with three or more rows | 6; 10 |
| maximum `version` | 81 |
| `events` / `leader_approvals` / `card_dispatch` / `workers` | 3213 / 65 / 141 / 24 |
| `leader_approvals` without a matching `(run_id, card_id)` row; `card_dispatch` without a `cards` row or with a run no row names; `cards.run_id` or `card_dispatch.run_id` missing from `runs` | 0; 0; 0 |
| `workers` bound to a retired run / to the active run / unbound | 0 / 24 / 0 |
| `runs` with empty or null `created_at`; `created_at` values shared by two runs | 0; 0 (37 runs) |
| `legacy states present` | `done, abandoned, leased, assigned, sync-audit, plan, plan-audit, run, merged-local, sync, merge-ready, kickoff, picked` |
| column order in the copy | `run_id, card_id, owner_label, state, version, evidence_path, updated_at, stage, lease_holder, lease_expires_at, heartbeat_at, decision_gate, decision_question, decision_resume, decider, decided_at, failure_reason, hint_prefer, hint_after, spec_id, worktree_path, evidence_sha, merge_sha, merge_tree, remeasure_path, contract_spec_id, contract_sha256, contract_signed_at, contract_event, bundle_id, bundle_order` |

What this says for the design: under the run-status rule the first revision had, the cutover would have blocked 0 card ids; under the strict rule now in force (§ 4.1) it blocks 25 card ids by owner labels and at most 27 in all until the operator clears them (plan § A.5); it needs 26 blank resolutions filled by an operator, and would turn 143 non-terminal rows of retired runs into sealed attempts while 93 card ids become project-scope records. The 41-card figure in the design document is for run `tmvgvn` at an earlier moment.

### 4.1 Revision measurement (fresh copy, strict conflict rule)

Taken 2026-10-18T01:16:00Z: `cp` of `factory.db` to the scratchpad (no `-wal`/`-shm` at copy time). `sqlite3` on the copy, owner labels taken as `owner_label` union `lease_holder` over non-terminal rows, a lease unexpired when `julianday(lease_expires_at) > julianday('now')`, a live holder one that is a `workers` row bound to the row's run.

| Query | Output |
|-------|--------|
| rows / distinct ids / duplicated ids | 300 / 228 / 55 |
| ids with two or more distinct non-empty owner labels among non-terminal rows (rule a) | 25 (`owner_label` alone: 25) |
| non-terminal rows with an unexpired lease now | 12 |
| non-terminal rows whose holder is a worker bound to the row's run | 20 |
| ids with two or more non-terminal rows and some row unexpired or worker-bound (rule b, an upper bound: the resolved row is exempt) | 4 |
| ids blocked by rule a or rule b (upper bound) | 27 |
| ids with non-terminal rows in two or more active runs | 0 |

So the strict rule blocks 25 ids by owner labels and at most 27 in all on this copy, against 0 for the run-status rule it replaces; the operator releases or abandons the rows the report names before a cutover.

## 5. Coordination facts

- **Card t1852, SPEC-FACTORY-RETIRE-GUARD-001** (worktree t1852, unmerged; `spec.md` read for requirements and Out of Scope): REQ-RRG-001 refuses `RetireRunIfDead` on a dead-owner active run that holds non-terminal cards, with a typed error and the exits (join, adopt, `decide --choice abandon --run`); REQ-RRG-007 keeps `retirable` the only predicate and leaves `retireRunUnlessLive` and `ReconcileActiveRuns` unchanged; its Out of Scope moves "retire = transfer" to card t1842 and says the refusal "is a temporary safeguard that card is expected to revise"; REQ-RRG-005 only reports the stranded cards of an already retired run. That its plan audit reached PASS 0.94 and that it is blocked at kickoff on an empty `spec_id` comes from the memory record `project_card_t1852_kickoff_specid_block.md`, not from a verdict file read here.
- **Card t1825 split.** From the memory records `project_card_t1825_plan_done_serial_wait.md` and `project_card_t1848_plan_ceiling_hold.md`: t1825 (SPEC-FACTORY-RUN-SELECT-001) was abandoned and split into t1848 (run chooser and retire refusal, then re-split into t1852 and t1853), t1849 (`cc -f` adopt) and t1842 (card transfer, this card). Its Q1 chose refuse-only for `runs --retire` with in-progress cards; its part D (cross-run card transfer) went to this card.
- **Cards t1835 (D) and t1843 (F).** Their worktrees exist; no SPEC for them was read. The interface of spec.md § E.2 is stated from this side only.

## 6. SQLite behaviors observed on the command line (version 3.54.0, 2026-10-18)

Run with `sqlite3` against scratch databases in the scratchpad; the project's driver (`modernc.org/sqlite`) was not exercised.

1. A view named `cards` with `INSTEAD OF INSERT`, `UPDATE` and `DELETE` triggers that `RAISE(ABORT, ...)`: `CREATE TABLE IF NOT EXISTS cards (...)` printed no error and left the view in place; `SELECT count(*) FROM cards` returned the view's rows; `INSERT`, `UPDATE` and `DELETE` through the name each printed the trigger's text and exited 1; `PRAGMA table_info(cards)` listed the view's columns.
2. A table with `BEFORE UPDATE`, `BEFORE DELETE` triggers and a `BEFORE INSERT` trigger that aborts when `EXISTS` a row with the same primary key: `UPDATE`, `DELETE`, `INSERT OR REPLACE`, `UPDATE OR REPLACE`, `REPLACE INTO` and a plain duplicate `INSERT` each printed `sealed attempt is immutable` and exited 1; the row was unchanged afterwards.
3. Two statements the design leans on were not run at all: `ROW_NUMBER() OVER (PARTITION BY …)` under the project driver, and a rename of a table with attached triggers.

## 6.1 Observed under the project driver (plan revision)

Scratch module with the repository's `go.mod` and `go.sum`, run as `env GOFLAGS=-mod=mod GOPROXY=off go -C <dir> run .`; `select sqlite_version()` printed `3.53.4`, `go.mod` pins `modernc.org/sqlite v1.60.1`.

1. Rollback naming: after `ALTER TABLE … RENAME`, dropping the three triggers and recreating them under `__rb1` names, a second table and triggers under the plain names were created without error; `UPDATE`, `DELETE` and `INSERT OR REPLACE` on the new table and on the archive each aborted (`immutable`); six triggers, six distinct names.
2. Journal identity guard: `UPDATE OR REPLACE factory_operations SET op_id='T' WHERE op_id='P'` aborted (`op identity immutable`), the terminal row remained, and `UPDATE … SET state='applied' WHERE op_id='P'` succeeded.
3. Column-adding migration: `ALTER TABLE cards ADD COLUMN` on the view failed (`Cannot add a column to a view`); `ALTER TABLE card_current ADD COLUMN` succeeded while the view existed; dropping and recreating the view and its trigger in one transaction succeeded and a later insert through the view aborted.
4. Read-only sidecars on a quiescent copy of the live database: `mode=ro&_pragma=query_only(ON)` left `f.db`, `f.db-shm`, `f.db-wal` with the `f.db` bytes equal and 300 rows readable; `immutable=1` left only `f.db`.

The plan audit's own probe covered the view over the legacy name, `INSTEAD OF` triggers, `ROW_NUMBER()`, a table rename inside a transaction and a stale handle, all under the same driver and version.

## 6.2 Observed under the project driver (revision for audit iteration 2)

Same scratch module and driver as § 6.1 (`modernc.org/sqlite` v1.60.1, SQLite 3.53.4), reduced schemas.

1. Seal trigger (`AFTER UPDATE … WHEN` predicate, archive insert of the `OLD` columns, nested `attempt_no` bump): a script of grant (`''` to a label), expiry-like write, owner restamp, worktree write, holder clear, run change, abandon and terminal re-pick produced exactly four archive rows (restamp, clear, run change, re-pick), `attempt_no` 1 to 5; each row held the owner and holder it carried; the archive refused an update.
2. After `ALTER TABLE … RENAME` the seal trigger followed the renamed table (`tbl_name` `card_current__rb1`), so rollback must drop it. `DROP INDEX sqlite_autoindex_*` failed: `index associated with UNIQUE or PRIMARY KEY constraint cannot be dropped`.
3. Journal: with the partial unique index alone, `INSERT OR REPLACE` of row Q on in-flight intent `kP` removed row P (progress and before-image gone). With a `BEFORE INSERT` guard on an in-flight intent collision and a `BEFORE DELETE` guard, `INSERT OR REPLACE`, `REPLACE INTO` and a plain `DELETE` all aborted and row P survived intact; a terminal row with the same intent was still insertable.
4. A column-adding migration with a stale seal trigger: the next update failed loudly (`table at2 has 4 columns but 3 values were supplied`) when the insert was positional; a named column list would omit the new column silently, so the trigger is rebuilt with the migration (design § 2.7).

## 7. Not verified at plan time

- No `go build` and no `go test` was run (host load average 20.63 at 09:33). The four existing retire tests of AC-CDI-002 were not observed to pass.
- The revision's driver probes used reduced schemas, not the 31-column `cards` DDL or the real triggers (§ 6.1); the full cutover transaction was not exercised because no cutover code exists.
- Whether any process outside the factory verbs writes `cards` was inferred from a grep of the repository, not from a runtime trace.
- Whether the MCP server holds a factory database handle between calls was not observed (stage A recorded the same gap; the MCP factory tools delegate into the same open path).
- The stage-A iteration 1 to 3 defect bodies were read as headlines and opening sentences only.
- Stage A's SPEC text (REQ-CLO-003, REQ-CLO-006, REQ-CLO-024, Decision 1) was read from worktree t1841; it is unmerged and may change in its re-plan.
- The shape of the stage-D and stage-F consumption is stated without a SPEC from either card.

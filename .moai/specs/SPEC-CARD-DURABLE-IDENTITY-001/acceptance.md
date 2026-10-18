# Acceptance: SPEC-CARD-DURABLE-IDENTITY-001

Verification layer for `spec.md`. Every command runs from the repository (worktree) root. `B0` below is the pin `f4bcd4f2cc65342559b505f1e5a29df80f992894`.

**Command form.** Every command in the ledger is one plain invocation: no pipe, no `;`, no `&&`, no redirection, no subshell, no `$(...)`, no `cd`. The worktree-session guard refuses commands it cannot show to be harmless; two refusals seen while authoring are recorded under § Guard measurements. The observation wrapper used to record exit codes (`; echo "exit=$?"`) is not part of any cited command.

**Document-level pin.** Every RED-now observation in this file was taken on tree B0 (the worktree at `HEAD` `f4bcd4f2cc65342559b505f1e5a29df80f992894`, working tree clean, this SPEC directory untracked) on 2026-10-18, with the `grep` installed on this host (it is `ugrep`: its warning text names it). A criterion-level pin overrides this one; none does. `B0` is a base, not a moving ref: the criteria that name it compare against it on purpose.

**Carrier.** The four RED-now elements (command, verbatim stdout, exit code, tree SHA) live in the evidence ledger in § Evidence ledger, because a table cell mangles quotes and globs. Each criterion cites its ledger entries by id.

**Reading the ledger.** Each entry has separate `stdout` and `exit` fields. For `grep -c` a zero count prints `0` and exits 1; for `grep -rlF` no match prints nothing and exits 1; `git ls-files <path>` exits 0 whether or not it prints, so its verdict is its stdout. `stderr` was empty except where shown.

**The behavior-test family.** The behavior criteria name Go tests with the prefix `TestCDI_`; the prefix is unused on B0 (collision check under E00). A named test is green only when its `--- PASS: <name>` line, and one line per named subtest, appears in the run's stdout. A run whose selector matches nothing prints `testing: warning: no tests to run` and `ok` and asserts nothing, so a missing `--- PASS:` line for any name is red whatever the exit code says. Run package-scoped with anchored selectors, `go test -v -count=1 -run '^(NameA|NameB|…)$' ./internal/homestate/` for the record layer and `./internal/cli/` for verb-level tests (the whole `internal/cli` suite is a heavy run: take the slot lease). `-v` is required: without it a passing package prints one `ok` line and nothing a criterion can read. Each behavior criterion pairs its tests with a **control cell** (the same fixture under the contrasting condition, which shows the test can fail) and lists the **mutants** it closes (a mutant that satisfies the criterion while violating its requirement cannot exist, verification-completeness § 2).

**Fixtures are built from the run-entry tree's own schema.** No fixture hard-codes a schema number or a column order: the legacy database of a test is produced by the tree's own `factoryDDL` and migration chain, once in the fresh column order and once in the `ALTER`-migrated order (design § 2.4), and the cutover tests run on both.

## Criteria

### AC-CDI-001 — One key, default `false` everywhere, non-boolean refused (maps REQ-CDI-001)
- **Given** the repository config, the template config and the code defaults, **When** E01a to E01d and `TestCDI_ConfigDefaultsOff` run, **Then** each of E01a to E01d prints a number of 1 or more, and the test prints its `--- PASS:` line: it reads `.moai/config/sections/workflow.yaml`, the template copy and the code default and asserts all three are `false`; it accepts `true` and `false`; a non-boolean (`yes`, `1`, `"maybe"`, an empty string) refuses to open the factory with a message containing the key path `workflow.card_identity.enabled`, the allowed values and the config file path, never coerced; a workflow file that cannot be decoded at all opens with the default `false` and a warning.
- Release-blocking. RED reason: all four greps print `0` (exit 1) because no `card_identity` block or `CardIdentity` type exists; the test does not exist (E00). Control: the same test with one of the three defaults changed to `true` fails. Mutants closed: a default of `false` in code only; coercion of `yes` to `true`; clamping an unknown value to `false`. Green path: M1. Ledger: E00, E01a, E01b, E01c, E01d.

### AC-CDI-002 — On the legacy layout nothing changes, and the cutover refuses without the key (maps REQ-CDI-002)
- **Given** a fresh database and a seeded legacy database, with the key `false` and again with the key `true` but no cutover run, **When** `TestCDI_OffLayoutByteIdentical` runs together with the existing tests `TestRetireRunPreservesRowAndAppendsEvent`, `TestRetireRunIfDeadRefusesLiveAndIndeterminate`, `TestReconcileNeverRetiresLiveOrIndeterminate` and `TestEveryRetirementPathDeclinesUnenumeratedClassification`, **Then** every name prints its `--- PASS:` line. The new test opens each database, runs the baseline verb sequence (record a card, assign, lease, stage, renew, expire, retire a dead run, reconcile), and asserts that the `sqlite_master` listing, `meta.schema_version` and every pre-existing row equal the baseline listing built from the tree's own DDL; none of `card_current`, `card_attempts`, `card_membership`, `cards_pre_cutover`, `factory_operations` exists; `cards` is a table; no journal row and no event kind from this SPEC was written. `identity cutover` with the key `false` refuses and its message contains the key path; with the key `true` and no resolution input it refuses naming the missing resolutions; both leave the dump identical.
- Release-blocking for the new test; the four existing tests are a regression guard (E11 shows they exist on B0; that they pass is not claimed here, it is the first thing `baseline.md` records at M0). RED reason: the new test does not exist (E00). Control: the same fixture after a successful cutover (the objects exist, so the test can see them). Mutants closed: an open that creates the journal table eagerly; a marker written at open; a migration step that bumps the version; a `cards` view created without a cutover. Green path: M1 (legacy untouched), kept green through M7. Ledger: E00, E11.

### AC-CDI-003 — `identity report` is read-only, complete, and proposes only where one candidate qualifies (maps REQ-CDI-003)
- **Given** a legacy fixture holding: a duplicated id with exactly one non-terminal row; a duplicated id with two non-terminal rows of which exactly one is in an active run; a duplicated id with two non-terminal rows both in retired runs; an all-terminal duplicated id; an id with non-terminal rows in two active runs; and one unreconciled entry in the record-unavailable log, **When** `TestCDI_ReportReadOnlyAndComplete` runs, **Then** it prints `--- PASS:` for `proposal_single_nonterminal`, `proposal_single_active_run`, `blank_when_two_retired`, `blank_when_all_terminal`, `lists_blocked_ids`, `lists_unreconciled_log_entries`, `counts_match_the_database`, `factory_db_and_table_content_unchanged`, `quiescent_database_gains_only_sidecars`, and `cutover_changes_the_content_control`. The output carries the legacy row count, the distinct-id count, each duplicated id with its candidate rows (run, state, owner label, lease holder, lease expiry, run status), the blocked ids with their owner labels and lease expiries, and the log count; `--emit` writes one JSON line per duplicated id with `run_id` blank where the proposal rule gives none; the SHA-256 of `factory.db` and a dump of every table are equal before and after; on a quiescent database (no sidecars before) the only difference in the directory is the `-wal` and `-shm` files the read-only open creates, which the invariant permits.
- Release-blocking. RED reason: `newFactoryIdentityCommand` does not exist (E04) and the test does not exist (E00). Mutants closed: a report that opens read-write and changes `factory.db` or a table; a proposal by recency or run order; a proposal where two candidates qualify; a count taken from the wrong table. Green path: M2. Ledger: E00, E04.

### AC-CDI-004 — A duplicated card id needs an explicit, valid resolution; recency never decides (maps REQ-CDI-004)
- **Given** a legacy fixture in which the most recently updated row of a duplicated id is not the row the resolution names, **When** `TestCDI_ResolutionRequired` runs, **Then** it prints `--- PASS:` for `missing_entry_refused`, `blank_entry_refused`, `unknown_run_refused`, `named_row_of_another_id_refused`, `terminal_row_named_while_nonterminal_exists_refused`, `all_terminal_duplicate_needs_an_entry`, `current_record_continues_the_named_row_not_the_newest`, and `complete_input_proceeds_control`. Each refusal lists the offending id, exits non-zero, and leaves the full dump (every table, `meta.schema_version`, the object list) identical to the dump taken before.
- Release-blocking. RED reason: the test does not exist (E00) and neither does the verb (E04). Mutants closed: a default to the newest `updated_at`; a default to the highest or lowest run id; a default to the single non-terminal row when no entry is given; blank read as "use the proposal"; an entry naming a row of a different card id. Green path: M2. Ledger: E00, E04.

### AC-CDI-005 — Distinct owners, a live lease on a row to be sealed, or an unreconciled log block the cutover, with no override (maps REQ-CDI-005)
- **Given** fixtures with the shapes named below, **When** `TestCDI_ConflictBlocks` runs, **Then** it prints `--- PASS:` for `two_distinct_owner_labels_block_even_with_a_resolution` (owner label and lease holder both counted, in retired runs only and across an active and a retired run), `unexpired_lease_on_a_row_to_be_sealed_blocks`, `unparseable_lease_expiry_counts_as_unexpired`, `registered_worker_holder_bound_to_that_run_blocks`, `unreconciled_record_unavailable_entry_blocks`, `the_resolved_row_with_a_live_lease_does_not_block_control`, `one_owner_label_repeated_across_runs_does_not_block_control`, `two_picked_rows_with_no_owner_do_not_block_control`, `blocked_ids_are_listed_with_rows_owner_labels_and_lease_expiries`, each blocked case leaving the dump identical and the lease reading taken at the cutover's single `now`.
- Release-blocking. RED reason: the test does not exist (E00). Mutants closed: the earlier rule that looks only at the run status (the first two subtests fail it); a rule that blocks on any duplicate (the three controls fail it); a rule that never reads a lease; a rule that accepts an override in the resolution input; a rule that ignores the log. Green path: M2. Ledger: E00, E04.

### AC-CDI-006 — Cutover preserves every legacy row as one sealed attempt: row counts before equal attempts after (maps REQ-CDI-006, REQ-CDI-007)
- **Given** a legacy database built from the run-entry tree's own schema and seeded with at least 14 rows over at least 8 card ids across at least 4 runs (two `active`, two `retired`), including: a duplicated id of two runs with one owner; a duplicated id of three runs; a duplicated id with one terminal and one non-terminal row; states covering `picked`, `assigned`, `leased`, `run`, `plan-audit`, `kickoff`, `needs-decision`, `merged-local`, `done`, `abandoned` and one row whose state lies outside the F1 set; and rows in `workers`, `runs`, `events`, `leader_approvals` and `card_dispatch`, **When** `TestCDI_MigrationRowPreserving` runs on the fresh-order and the `ALTER`-order variants, **Then** it prints `--- PASS:` for `fresh_order` and `altered_order`, and each prints the line `legacy_rows=N attempts=N current=D distinct_ids=D` with the two N equal and the two D equal, and asserts: every `card_attempts` row equals its legacy row on all 31 columns by name; `cards_pre_cutover` equals the pre-cutover dump of `cards`; every `card_current` row of an active run equals the resolved legacy row on all 31 columns, version included; every one of a retired run, non-terminal, has `run_id` empty and a version one higher, with one `card.migrated` event of reason `cutover-run-retired`; `meta.schema_version` ends in `+id1` and its base is unchanged; every pre-existing row of `workers`, `runs`, `events`, `leader_approvals` and `card_dispatch` is unchanged; the `cards` view returns D rows in the legacy column shape; a second cutover is refused. Every terminal record equals its resolved row unchanged. `TestCDI_IdentityVerbsLaneRefused` also runs and prints `--- PASS:` for `cutover_refused_in_a_lane_session` (the dump is unchanged) and `cutover_allowed_in_an_operator_session_control`.
- Release-blocking. RED reason: neither `card_attempts` (E02) nor `card_current` (E03) exists and the test does not exist (E00). Control: the same fixture with one legacy row deleted by the test before the cutover reports `legacy_rows=N-1 attempts=N-1`, showing the equality is measured, not assumed. Mutants closed: duplicates merged into one attempt; terminal rows dropped; only the latest row per id kept; `updated_at` rewritten; columns copied by position (the two order variants fail it); no `cards_pre_cutover`; a version bump on an active-run record. Green path: M2. Ledger: E00, E02, E03. Reviewer-checked half: RC-3.

### AC-CDI-007 — A sealed attempt cannot be changed by any statement on any connection (maps REQ-CDI-008)
- **Given** a cut-over database, **When** `TestCDI_AttemptImmutable` runs against it with the project's SQLite driver, **Then** it prints `--- PASS:` for `update_refused`, `delete_refused`, `insert_same_key_refused`, `insert_or_replace_refused`, `replace_into_refused`, `update_or_replace_refused`, each issued on a second connection against `card_attempts`, and the same six against `cards_pre_cutover`, and for `new_key_insert_by_the_seal_path_succeeds_control`; each refusal carries an error containing `immutable` and the table dump is equal before and after.
- Release-blocking. RED reason: the test and the table do not exist (E00, E02). Mutants closed: update and delete triggers only (the three replace subtests fail it); a trigger on the wrong table; application-level immutability only (a raw statement passes it). Green path: M2. Ledger: E00, E02. The command-line observation that motivates the three-trigger form is recorded in research.md § 6.

### AC-CDI-008 — The cutover validates counts and references in its own transaction and a failure leaves nothing (maps REQ-CDI-009)
- **Given** the AC-CDI-006 fixture, **When** `TestCDI_ValidationRollsBack` runs, **Then** it prints `--- PASS:` for `check_v1_attempt_count`, `check_v2_current_count`, `check_v3_orphan_attempt`, `check_v4_membership_reference` (a membership naming a run that was present before the cutover and is absent after fails; one naming a run already absent before passes), `check_v5_dangling_reference_set`, `check_v6_attempt_column_equality`, `check_v7_current_column_equality`, `fault_after_copy`, `fault_before_validate`, `fault_before_marker`, `fault_before_commit`, and `no_fault_succeeds_control`; for each injected defect or fault the cutover exits non-zero, the message names the failed check by its id (`V-1` to `V-7`), and the dump (every table, the object list, `meta.schema_version`) is identical to the dump before the attempt.
- Release-blocking. RED reason: the test does not exist (E00). Mutants closed: validation after the commit; validation of counts only (the `V-6` and `V-7` injections fail it); a validation that tolerates a pre-existing dangling reference changing (V-5 compares sets before and after, so a new dangling reference fails and a pre-existing one does not). Green path: M2. Ledger: E00, E02.

### AC-CDI-009 — After the cutover an older build cannot open the database, and a handle opened earlier cannot write a card (maps REQ-CDI-010)
- **Given** a cut-over database and a second connection opened before the cutover, **When** `TestCDI_OldBuildRefusal` runs, **Then** it prints `--- PASS:` for `schema_version_is_not_the_legacy_integer` (the value is neither equal to the build's base integer as a string nor parseable by `strconv.Atoi`), `legacy_open_statements_refuse` (the legacy `factoryDDL` text and the legacy version comparison, copied into the test file with a comment citing `internal/homestate/factory.go:317` and `:368`, run against the cut-over database: the DDL executes without error and without changing the object list, and the comparison returns the unsupported-schema-version refusal), `prepared_handle_insert_aborts`, `prepared_handle_update_aborts`, `prepared_handle_delete_aborts`, `prepared_handle_replace_aborts`, `prepared_handle_select_returns_current_records_in_the_legacy_shape`, and `prepared_handle_workers_write_succeeds_residual_control`; each aborting statement carries an error containing `identity cutover` and leaves the rows unchanged.
- Release-blocking. RED reason: the test does not exist (E00) and no cutover can produce the state (E04). Regression guard: E10 (the open-time refusal exists at B0) must still print 1 or more at the tip. Mutants closed: a view without `INSTEAD OF` triggers; a table left under the name `cards`; a marker that parses as an integer. Green path: M2. Ledger: E00, E04, E10.

### AC-CDI-010 — Rollback rebuilds the legacy table with everything written after the cutover (maps REQ-CDI-011)
- **Given** a cut-over fixture on which, after the cutover, a lane was granted a lease, a card advanced a stage, a run was retired with migration, a project-scope card was joined, a terminal card was re-picked, a second stranded card was left in project scope (the shape the cutover gives every non-terminal card of a retired run), a duplicated id whose resolution named its older row was left in project scope, and a card was joined from project scope (which seals a project-scope row), **When** `TestCDI_RollbackPreservesLaterGrants` runs, **Then** it prints `--- PASS:` for `one_legacy_row_per_run_card_pair`, `post_cutover_grant_present_in_its_pair`, `sealed_pair_row_equals_its_sealed_state`, `no_version_lower_than_cards_pre_cutover`, `pairs_untouched_since_cutover_equal_cards_pre_cutover`, `identity_objects_renamed_not_dropped`, `marker_stripped_base_unchanged`, `refused_with_a_nonterminal_journal_row`, `refused_in_a_lane_session`, `refused_on_the_legacy_layout`, `fresh_cutover_after_rollback_works`, `second_cutover_new_attempts_refuse_update_delete_and_replace` (an `UPDATE`, a `DELETE` and an `INSERT OR REPLACE` on the NEW `card_attempts`), `archive_still_refuses_after_the_second_cutover`, `trigger_and_index_names_are_distinct_per_layout` (every row of `sqlite_master` of type `trigger` or `index` has a distinct name), `second_cutover_numbers_above_the_floor`, `project_scope_record_returns_to_its_return_run` (the duplicated id resolved to its older row: its rolled-back pairs are `(older run, current progress)` and `(other run, that run's own sealed snapshot)`, R-3 and R-4 pass), `sealed_project_scope_history_forms_no_pair` (the joined card rolls back to its legacy pair and its joined pair, 0 empty-run rows), `second_cutover_after_each_of_these_succeeds`, `rebuilt_table_has_no_empty_run`, `cutover_after_rollback_with_a_stranded_card_passes_check_v4`, and `cutover_keeps_a_row_whose_run_row_is_missing` (the step-2 join on `runs` is a left join, the row sorts last, and V-4 passes because the run was already absent before the cutover); a reverse-check failure (R-1 to R-4, design § 7) leaves the dump identical.
- Release-blocking. RED reason: the test does not exist (E00). Control: a rollback fed a deliberately lowered version in the fixture fails `no_version_lower_than_cards_pre_cutover`. Mutants closed: a rebuild from `cards_pre_cutover` alone (loses the grants); a rebuild that lowers a version; dropping the identity tables instead of renaming them; a rebuild that writes an empty-run row for a project-scope record (the no-empty-run and V-4 subtests fail it); an inner join on `runs` that drops a row; a rollback that leaves `card_current_seal` on the renamed table (it would keep sealing into the renamed archive); a rename that leaves the trigger names in place (the second-cutover subtests fail it, as `trigger … already exists`, or silently leave the new table unprotected under `IF NOT EXISTS`). Green path: M5 for every subtest except `refused_with_a_nonterminal_journal_row` (the rollback checks for the journal table only when it exists), which turns green at M6; the criterion is green at M6. Ledger: E00, E02.

### AC-CDI-011 — One current record per card id, one open membership, monotone versions, and a sealed attempt is not a lease (maps REQ-CDI-012)
- **Given** a cut-over database and a deterministic scripted sequence of at least 200 operations drawn from record, assign, lease, stage, renew, expire, reassign, dispatch restamp, retire-with-migration, join, terminal re-pick, and a rollback (with a stranded card in project scope) followed by a later cutover (the generator seed is printed), **When** `TestCDI_IdentityInvariants` runs, **Then** it prints `--- PASS:` for `one_current_record_per_card`, `at_most_one_open_membership`, `version_and_attempt_never_fall`, `attempt_no_never_falls_across_rollback_and_recutover` (three terminal re-picks of one card inside one run, then rollback and cutover: the next `attempt_no` is above the highest before the rollback), `version_never_falls_across_rollback_and_recutover` (a duplicated id with a lower-versioned older pair, joined and raised before the rollback, then re-cut over resolved to the older pair: the new current version is above the highest version the card held before the rollback), `recutover_with_a_version_floor_adjustment_passes_check_v7` (a re-cutover resolved to a row whose version plus one is below the card's version floor kept by the rollback carries the floor plus one, and check V-7, comparing the moved row against exactly that expectation, passes and the cutover commits), `non_member_request_is_not_card_not_found` (the error is `ErrCardNotMember` and names the open membership run or project scope), and `sealed_attempt_is_not_a_lease` (a sealed attempt whose recorded state is lease-holding with an unexpired lease is ignored by reassignment, the serial-slot reader and the status view), each checked after every operation by queries over the tables.
- Release-blocking. RED reason: `card_current` does not exist (E03), `ErrCardNotMember` does not exist (E13), and the test does not exist (E00). Mutants closed: a version reset at a seal; a second record created on a re-pick; a reassignment that reads sealed attempts as leases; a cutover that numbers from the rebuilt legacy row count (the rollback subtest fails it). Green path: M3 for every subtest and operation that does not need migration or rollback; the migration operations of the scripted sequence turn green at M4 and `attempt_no_never_falls_across_rollback_and_recutover` and the rollback operations at M5; the criterion is green at M5. Ledger: E00, E03, E13.

### AC-CDI-012 — Recording a card in a run never creates a second record and never resets a continuing card (maps REQ-CDI-013)
- **Given** a cut-over database, **When** `TestCDI_RecordInRun` runs, **Then** it prints `--- PASS:` for `engaged_elsewhere_refused_names_run_and_verb` (record, assign and claim each refuse with `ErrCardEngaged`, naming the open run and `moai factory runs --retire`, creating no row and changing no field, and the `factory next` arms report it as a refusal, not as the race retry), `project_scope_join_opens_membership_once` (version advances by exactly one, a membership is opened, one `card.joined` event is written, and state, stage, evidence, worktree, owner, lease holder and lease expiry are equal to their pre-join values), `terminal_repick_seals_and_opens_a_new_attempt` (the sealed attempt equals the pre-state, the new attempt is `picked`, its version is the sealed version plus one, `attempt_no` is one higher), and `legacy_layout_repick_starts_a_second_row_at_version_one_control` (the same sequence on the legacy layout creates a second row at version 1, the defect this SPEC removes).
- Release-blocking. RED reason: `ErrCardEngaged` does not exist (E14) and the test does not exist (E00). Mutants closed: a join that resets the stage; a second record created while engaged elsewhere; a terminal re-pick that overwrites without sealing; `ErrCardEngaged` mapped to the retry path. Green path: M3. Ledger: E00, E14.

### AC-CDI-013 — The cutover, seal, join, migration and rollback leave every epoch-carrying table alone and clear no lease holder (maps REQ-CDI-014)
- **Given** a legacy fixture that also holds a stand-in table `card_epoch_probe(card_id TEXT PRIMARY KEY, owner_epoch INTEGER NOT NULL, owner_label TEXT NOT NULL)` seeded with epochs including 7 and 0, and a probe trigger on `card_current` recording any change of `lease_holder` away from a non-empty value, **When** `TestCDI_EpochCarrierUntouched` runs, **Then** it prints `--- PASS:` for `carriers_found_by_column_scan` (the test finds every table with a card-id key and an `owner_epoch` column through `pragma_table_info`, not by name), `unchanged_across_cutover`, `unchanged_across_join`, `unchanged_across_retire_migration`, `unchanged_across_terminal_repick_seal`, `unchanged_across_rollback`, `unchanged_across_recutover`, `no_holder_cleared_by_migrate_join_cutover_or_rollback` (the probe trigger recorded zero rows; a lease expiry or reassignment, which clear the holder as at B0 after the seal, are outside this subtest), `lease_columns_equal_after_migration` (holder, expiry, heartbeat and owner label equal the pre-state), and `dump_check_can_fail_control` (a deliberate `UPDATE` of one epoch by the test is detected).
- Release-blocking. RED reason: the test does not exist (E00) and no epoch carrier exists on B0 outside the test (E09, so the stand-in is constructed). Mutants closed: a cutover that recreates every table; a rollback that deletes unrecognised tables; a migration that clears the holder and restores it in the same transaction (the probe trigger sees the intermediate state). If a re-planned stage A has landed (plan.md B-4), the same test also runs with its real grant table. Green path: M5, with the join and seal halves from M3. Ledger: E00, E09.

### AC-CDI-014 — Retirement migrates every active card, then retires, in one transaction (maps REQ-CDI-015)
- **Given** a dead-owner active run R holding six active cards (a `picked` card with no owner; an `assigned` card owned by lane-1; a `run` card whose lease holder is lane-2 with an unexpired lease; a `kickoff` card; a `needs-decision` card; a `merged-local` card) and one `done` card, a second active run S, and the owner lanes bound to R in `workers`, **When** `TestCDI_RetireMigrates` runs, **Then** it prints `--- PASS:` for `destination_into_flag`, `destination_is_the_relaunch_run_flag`, `destination_is_the_sole_survivor`, `destination_is_project_scope_when_no_survivor`, `sealed_attempt_equals_the_pre_state`, `current_record_differs_only_in_run_attempt_and_version`, `lease_columns_and_owner_unchanged`, `membership_moved`, `dispatch_binding_repointed`, `owner_roster_entry_rebound`, `one_card_migrated_event_each`, `terminal_card_untouched`, `run_retired_with_event`, `zero_active_cards_just_retires`, and `live_owner_still_refused_control`.
- Release-blocking. RED reason: `retireRunWithMigration` does not exist (E15), no `card.migrated` event exists (E06), the test does not exist (E00). Control: the legacy-layout retire of the same run strands the six cards (their rows keep naming R, which is retired). Mutants closed: retire first, migrate second (a fault between the two leaves an unmigrated retired run); a migration that clears the lease holder; a destination chosen silently when two survivors exist. Green path: M4. Ledger: E00, E06, E15.

### AC-CDI-015 — Retirement refuses whole when any card cannot migrate, and offers no force path (maps REQ-CDI-016)
- **Given** the AC-CDI-014 fixture, **When** `TestCDI_RetireRefusals` and `TestCDI_RunsFlagSet` run, **Then** they print `--- PASS:` for `destination_retired_refused`, `destination_unknown_refused`, `destination_same_as_source_refused`, `ambiguous_survivors_without_into_refused`, `validation_fault_after_second_card_rolls_back_the_whole_run` (the run is `active` and the dump of every table is identical), `busy_database_refuses_with_nothing_written`, `refusal_lists_each_blocked_card_and_its_reason`, `automatic_reconciler_unchanged`, and, for the second test, `runs_flag_set_is_exactly_retire_and_into`, `relaunch_has_no_force_flag`.
- Release-blocking. RED reason: the tests do not exist (E00) and `--into` does not exist (E15 covers the migration entry). Mutants closed: a `--force` or `--yes` flag; a partial migration committed with the run left active; the refusal printing only a count. Green path: M4. Ledger: E00, E15.

### AC-CDI-016 — Stranded cards of a retired run become addressable and keep their progress; the legacy layout loses it (maps REQ-CDI-006, REQ-CDI-007, REQ-CDI-013)
- **Given** a fixture shaped like the 10-17 incident: a retired run T holding three non-terminal cards (a `run` card with lease holder lane-5 and an expired lease; a `plan-audit` card with evidence SHA and path; a `kickoff` card with its decision gate), each with a worktree path and spec id; an older retired run U holding a second row of one of them; an active run A holding none, **When** `TestCDI_StrandedCardsBecomeAddressable` runs, **Then** it prints `--- PASS:` for `cutover_makes_each_a_project_scope_record` (run empty, version one higher, every other column equal), `card_migrated_event_reason`, `join_into_the_active_run_keeps_every_field`, `expired_lease_is_collected_by_the_next_transition_as_before`, `stage_evidence_worktree_and_spec_survive_join_and_transition`, and `legacy_layout_repick_in_the_active_run_starts_at_picked_control`.
- Release-blocking. RED reason: neither `card_current` (E03) nor `card.migrated` (E06) exists, and the test does not exist (E00). Mutants closed: a cutover that leaves retired-run rows out of `card_current`; a join that resets the stage; a cutover that picks the retired run's row by recency. Green path: M2 and M3. Ledger: E00, E03, E06. Reviewer-checked half: RC-3 (the same shape on a copy of the live database).

### AC-CDI-017 — Every operation is in the journal before its effect begins (maps REQ-CDI-017)
- **Given** a cut-over database and the four operation kinds with their effect functions wrapped, **When** `TestCDI_OperationRecordedBeforeEffect` runs, **Then** it prints `--- PASS:` for `queue_claim_row_committed_before_effect`, `lock_acquire_row_committed_before_effect`, `git_merge_row_committed_before_effect` (at the moment each effect starts, a second connection reads a committed row carrying the operation id, kind, card id, attempt number, actor, intent key, a non-empty expected-readback and a non-empty before-image, in state `prepared`), `card_grant_row_commits_with_the_grant` (a fault before commit leaves neither grant nor row; a commit leaves the row `applied` with the grant), and `legacy_layout_writes_no_row`.
- Release-blocking. RED reason: `factory_operations` does not exist (E07) and the test does not exist (E00). Mutants closed: a row written after the effect; a row written inside a transaction that the effect's failure rolls back; an empty before-image. Green path: M6. Ledger: E00, E07.

### AC-CDI-018 — Presenting an operation's intent again repeats nothing (maps REQ-CDI-018)
- **Given** each kind with the effect done and the response lost, **When** `TestCDI_OperationRetryIdempotent` runs, **Then** it prints `--- PASS:` for `queue_claim_second_presentation_returns_the_first_outcome`, `card_grant_second_presentation_returns_the_first_outcome`, `git_merge_second_presentation_returns_the_first_outcome`, `lock_acquire_second_presentation_returns_the_first_outcome`, `second_presentation_after_reconciliation_returns_the_recorded_result`, `second_presentation_after_superseded_refuses_without_effect`, `second_presentation_after_void_is_a_new_presentation`, `concurrent_second_presentation_resolves_to_one_row`; in each, the effect counter is 1, the card version after the second presentation equals the version after the first, the number of `card.transition` events is unchanged, and the journal holds one row for the intent.
- Release-blocking. RED reason: the test and the table do not exist (E00, E07). Mutants closed: a retry that applies the effect again; a retry that mints a second row (the partial unique index refuses it, and the subtest fails a build without the index); a lookup limited to in-flight rows (the after-reconciliation subtest fails it). Green path: M6. Ledger: E00, E07.

### AC-CDI-019 — Reconciliation decides every cell of the table the same way every time and never repeats a merge (maps REQ-CDI-019)
- **Given** the fourteen cells of design § 6.3 built over a temporary queue store, a temporary integration-lock root, a temporary git repository and a factory database, **When** `TestCDI_ReconcileTable`, `TestCDI_ReconcileDeterministic`, `TestCDI_ReconcileNeverRepeatsMerge` and `TestCDI_ReconcileIndeterminateStaysPending` run, **Then** they print `--- PASS:` for the fourteen subtests named exactly `queue_claim__after_prepared`, `queue_claim__after_effect`, `queue_claim__assigned_only`, `queue_claim__after_record`, `card_grant__client_lost`, `card_grant__superseded_by_later_write`, `lock_acquire__after_prepared`, `lock_acquire__after_effect`, `lock_acquire__held_by_other`, `git_merge__after_prepared`, `git_merge__after_effect`, `git_merge__card_merging`, `git_merge__after_record`, `any__indeterminate_readback` (each asserts the decision, the terminal row state and the exact set of writes to the four stores), for `second_pass_writes_nothing`, `row_order_and_clock_jump_do_not_change_decisions`, `merge_counter_is_zero_in_every_cell`, and `indeterminate_row_stays_pending_and_is_reported`.
- Release-blocking. RED reason: the tests do not exist (E00) and neither does the table (E07). Mutants closed: a pass that repeats the merge; a decision that reads the clock; a compensation without compare-and-set (the `queue_claim__after_effect` superseded case overwrites a newer engagement); an indeterminate readback marked `void`; a pass that reads a card left `assigned` with the queue after-image as `after_record` (the `assigned_only` subtest fails it). Green path: M6. Ledger: E00, E07.

### AC-CDI-020 — Any session reconciles, with no leader, and the claim and completion verbs run the pre-step (maps REQ-CDI-020)
- **Given** a database with no leader run process, no coordinator record, no approval rows, and a lane session holding a card with a non-terminal journal row, **When** `TestCDI_ReconcileWithoutLeader`, `TestCDI_ReconcileNoLeaderImports`, E20a and E20b run, **Then** the tests print `--- PASS:` for `lane_reconciles_its_own_card`, `operator_reconciles_all`, `claim_verb_runs_the_pre_step_once`, `complete_verb_runs_the_pre_step_once`, and `no_non_test_file_of_the_journal_and_reconcile_code_names_a_leader_symbol` (the test scans every non-test file of the package that the reconcile code imports from, by import graph, not by file name), E20a prints nothing (exit 1), and E20b prints at least one file (the sweep is not empty, so E20a's silence means something).
- Release-blocking. RED reason: E20b prints nothing (exit 1) because no `operation_*.go` non-test file exists, so the sweep that E20a rests on is empty; `newFactoryReconcileCommand` does not exist (E05); the tests do not exist (E00). E20a alone is silent at RED and at green, which is why E20b is paired with it. Mutants closed: a reconcile that requires an active run; one that reads `leader_approvals`; a pre-step that runs twice. Green path: M6. Ledger: E00, E05, E20a, E20b.

### AC-CDI-021 — Journal rows move only forward and exist only on the identity layout (maps REQ-CDI-021)
- **Given** a cut-over database and a legacy database, **When** `TestCDI_JournalForwardOnlyAndAbsentOnLegacy` runs, **Then** it prints `--- PASS:` for `prepared_to_applied_to_each_terminal_state`, `backward_move_refused`, `terminal_update_delete_replace_refused`, `update_or_replace_onto_a_terminal_key_refused` (an `UPDATE OR REPLACE` that renames a non-terminal row onto a terminal row's `op_id` aborts and the terminal row survives), `op_identity_update_refused` (`op_id`, `kind`, `intent_key`, `card_id`, `attempt_no` each), `second_inflight_row_for_one_intent_refused`, `insert_or_replace_onto_an_existing_op_id_refused` (an in-flight and a terminal target, also `INSERT … ON CONFLICT(op_id) DO UPDATE`: the original row, its progress and its before-image are unchanged), `insert_or_replace_onto_an_inflight_intent_refused` (also `REPLACE INTO`: the original row, its progress and its before-image survive, and a plain `DELETE` of any row aborts), `no_table_on_the_legacy_layout`, and `no_row_written_on_the_legacy_layout`.
- Release-blocking. RED reason: the table and the test do not exist (E07, E00). Mutants closed: triggers on update and delete only; triggers on the terminal row only (the replace-onto subtest fails it); no partial unique index; a unique index without the insert guard (a `REPLACE` conflict on it deletes the in-flight row, and the in-flight-intent subtest fails it); a guard on the intent key alone (an `INSERT OR REPLACE` that reuses an `op_id` overwrites an in-flight row or rewrites a terminal one, and the op_id subtest fails it). Green path: M6. Ledger: E00, E07.

### AC-CDI-022 — The published facts are readable without writing, the external contracts change only additively, and the kind set is closed (maps REQ-CDI-022)
- **Given** a cut-over database opened read-only, **When** `TestCDI_InterfaceReadsAreReadOnly`, `TestCDI_ExternalContractsAdditive` and `TestCDI_OperationKindsClosed` run, **Then** they print `--- PASS:` for `reads_work_on_a_readonly_handle_and_leave_the_bytes_equal` (the five reads of design § 12: the current record by card id alone, its `attempt_no`, its membership, the sealed attempts in order, the journal rows of a card, and the layout), `card_json_keys_are_a_superset_of_b0_plus_attempt_no_and_membership`, `event_kinds_include_every_b0_kind_and_the_five_new_ones`, `exactly_four_operation_kinds`, and `table_present_reports_the_cards_view_on_the_identity_layout` (`FactoryTablePresent` answers true for `cards` where it is a view).
- Release-blocking. RED reason: the tests and the reads do not exist (E00, E03). Mutants closed: removing or renaming a B0 JSON key; renaming an event kind; a fifth kind added without the additive-extension rule. Green path: M5 for the identity reads, the JSON keys and the table-present subtest; the journal reads, `exactly_four_operation_kinds` and the `operation.reconciled` kind turn green at M6; the criterion is green at M6. Ledger: E00, E03.

### AC-CDI-023 — The run phase touched only the allowed paths, and every branch commit names the card
- **Given** the pinned base B0, **When** E23a lists the files changed by the card's own tagged commits outside the allowlist and E23b lists the branch's own commits that do not name the card, **Then** both print nothing.
- **Regression-guard, not release-blocking.** Both are green on B0 by construction; their red was observed on a constructed history (§ Guard measurements). The allowlist (spec D-8) is: `internal/homestate`, `internal/config`, `internal/factorymsg`, `internal/factory`, `internal/cli/factory_*`, `internal/cli/mcp_factory_*`, `internal/cli/integration*`, the repository and template `workflow.yaml`, this SPEC directory, and this card's evidence directory `.moai/reports/t1842`. The evidence directory is excluded by pathspec on purpose and no other part of `.moai/reports/` is. E23a follows the branch's own first-parent, non-merge, card-tagged commits, so the commits that arrive through absorbing merges do not appear. **Known limit:** a file that arrives through a merged side branch of untagged commits is seen by neither command; the reviewer lists `git log --first-parent --merges` for the branch and confirms each merge absorbs a ref of `main` (RC-5). Green path: evaluated at the run-phase tip. Ledger: E23a, E23b.

### AC-CDI-024 — The failing baseline is recorded in its own commit before any commit that makes a test pass (maps no requirement; plan.md M0)
- **Given** the card branch at the run-phase tip, **When** the reviewer runs E24a (the commit that added `.moai/reports/t1842/baseline.md`) and E24b (the branch's own first-parent tagged commits under `internal`, oldest first) and reads `git show --name-only --format= <baseline commit>`, **Then** the baseline commit is the oldest commit E24b lists, it touches only `*_test.go` files, the declaration file of `errNotImplemented` stubs, and the `.moai/reports/t1842` directory, and `git merge-base --is-ancestor <baseline commit> <each later commit>` exits 0 for every later tagged commit under `internal`.
- **Reviewer-checked, with E24a and E24b as the floor.** Git snapshots the tree per commit and cannot witness authoring order inside one commit, so the ordering claim is stated only as what the commit graph can verify (verification-claim-integrity § 2.3): the baseline artifact and the code it measures are never in one commit. `baseline.md` carries the verbatim failing run of the `TestCDI_` family, the passing run of the four existing tests of AC-CDI-002, and the passing run of the control `TestCDI_LegacyRepickLosesStage_Control`. RED reason: `baseline.md` does not exist (E08). Green path: M0.

### AC-CDI-025 — Every ended lease grant or owner is sealed as an immutable attempt by the store, whichever writer ends it (maps REQ-CDI-023)
- **Given** a cut-over database and one card scripted through: leased to lane-1, renewed, expired; reassigned to lane-2; restamped to lane-3 by the dispatch owner restamp; leased by lane-3, advanced one stage, renewed, its worktree bound; completed to `merged-local` (the holder clears); migrated to another run; completed to `done`; and re-picked, **When** `TestCDI_EveryEndedGrantIsSealed` runs, **Then** it prints `--- PASS:` for `expiry_seals_the_displaced_holder` (the sealed row carries lane-1 as lease holder, its lease expiry, owner, stage, evidence and version), `reassign_seals_the_displaced_owner`, `dispatch_restamp_seals_the_displaced_owner` (`RecordDispatchEngagement` over an expired lease, owner lane-2 replaced by lane-3, holder unchanged), `leaving_lease_holding_states_seals`, `migration_seals`, `terminal_repick_seals`, `grants_renewals_hints_worktree_and_abandon_append_nothing`, `exactly_six_archive_rows_for_the_script` (expiry, reassign, restamp, the clear at `merged-local`, migration, re-pick, so `attempt_no` ends at its start plus six), `every_writer_of_the_current_record_seals` (the test issues the qualifying change through every non-test caller of `updateCardRow` that can carry it — expiry, commit, reassign, restamp, re-pick — and through a raw `UPDATE` on a second connection; each appends exactly one row, while the worktree-binding caller appends none), `seal_commits_and_rolls_back_with_the_write` (a fault between the seal and the version update leaves neither the archive row nor the write), `one_transaction_per_write`, and `sealed_rows_equal_the_pre_write_record_on_all_31_columns`.
- Release-blocking. RED reason: `seal_reason` does not exist in any non-test file (E25) and the test does not exist (E00). Control: `grants_renewals_hints_worktree_and_abandon_append_nothing`. Mutants closed: a seal in caller code that omits one caller (the restamp, the raw-update and every-writer subtests fail it); a seal on every transition (the exact-count subtest fails it); a seal that runs after the write; a seal in its own transaction; a seal on every renewal; a trigger whose column list lacks a column (the 31-column subtest fails it). Green path: M3 for every subtest except `migration_seals` and the migration step of the script, which turn green at M4; the criterion is green at M4. Ledger: E00, E25.

## Reviewer-checked criteria (no mechanical command)

- **RC-1 (AC-CDI-002).** Read `baseline.md` and confirm the four existing tests' passing run was taken before any code commit, with the command and output recorded.
- **RC-2 (AC-CDI-013).** Confirm the stand-in epoch table's shape matches the stage-A SPEC's statement of its grant (card id primary key, `owner_epoch`), and, if a stage-A table has landed, that the second run used it.
- **RC-3 (AC-CDI-006, AC-CDI-016).** Copy the live factory database (`cp` of `factory.db` alone, noting any `-wal` sibling), run `moai factory identity report` on the copy, and compare its counts with `sqlite3` counts of the same copy (rows, distinct ids, duplicated ids, blocked ids); record both. The 2026-10-18T01:16:00Z copy gave 300 rows, 228 ids, 55 duplicated ids and, under the strict rule, 25 ids blocked by distinct owner labels and at most 27 in all with the live-lease rule, by the SQLite queries of research.md § 4.
- **RC-4 (AC-CDI-014, AC-CDI-017).** Read the fixtures and confirm the dead-owner run's owner is classified `dead` through the production classifier seam, and that each effect wrapper of AC-CDI-017 observes the journal through a second connection rather than the writer's own transaction.
- **RC-5 (AC-CDI-023).** List `git log --first-parent --merges` for the branch and confirm each merge absorbs a ref of `main`.
- **RC-6 (AC-CDI-019).** Read the 14 cell fixtures against design § 6.3 and confirm each cell's pre-state is built by driving the real code to the crash point through a fault seam, not by inserting the state by hand.

## Edge cases

- **EC-1. The database is busy during the cutover.** The verb fails; nothing partial exists (AC-CDI-008 fault subtests, design § 11).
- **EC-2. A card id has exactly one legacy row.** No resolution is needed; the current record continues it; a terminal row keeps its run (AC-CDI-006).
- **EC-3. A legacy row whose state is outside the F1 set.** Copied verbatim, treated as non-terminal (AC-CDI-006).
- **EC-4. A reference that already dangled before the cutover.** Not a failure; V-5 compares the set before and after (AC-CDI-008).
- **EC-5. A process with a pre-cutover handle writes `workers`.** Allowed; it is the documented residual (AC-CDI-009 control, spec R-2).
- **EC-6. A resolution names a terminal row while a non-terminal row exists.** Refused (AC-CDI-004).
- **EC-7. A run with no active card is retired.** It retires without migration and writes the same `run.retired` event (AC-CDI-014 `zero_active_cards_just_retires`).
- **EC-8. Rollback and then cutover again.** The second cutover starts fresh from the rebuilt legacy table with new trigger names and attempt numbers above the floor (AC-CDI-010, AC-CDI-011).
- **EC-11. A claim dies after the assign transition commits.** The card stays `assigned` with the queue after-image; the next reconcile pass rolls the lease edge forward or records `assigned-only` (AC-CDI-019 `queue_claim__assigned_only`).
- **EC-12. A lease is granted, expires and is granted again inside one run.** Each ended grant is one sealed attempt carrying its holder; the current record holds only the open one (AC-CDI-025).
- **EC-9. A lane renews a migrated card naming the old run.** `ErrCardNotMember` naming the new run; its roster entry was re-bound, so the retry passes (AC-CDI-011, AC-CDI-014).
- **EC-10. A migrated card's lease has already expired.** Carried as it is; the next transition collects it exactly as before (AC-CDI-016 `expired_lease_is_collected_by_the_next_transition_as_before`).

## Definition of Done

- Every release-blocking criterion has its named `--- PASS:` lines recorded in `progress.md` § E.2 with the command and the observed output, and its RED-now ledger entry is cited.
- The regression guards AC-CDI-023, the four named tests of AC-CDI-002, and E10 pass at the run-phase tip; the reviewer-checked criteria are recorded as checked with the evidence read.
- `workflow.card_identity.enabled` ships `false` in the repository and the template; no deployment is cut over by this SPEC.
- `baseline.md` is tracked and sits in the commit of AC-CDI-024.
- The plan-artifact hash recorded by the plan audit is unchanged at run entry (plan.md § E), or the audit is re-run.

## Evidence ledger

All entries: tree `f4bcd4f2cc65342559b505f1e5a29df80f992894`, working tree clean, run from the worktree root on 2026-10-18. `stderr` empty unless shown.

```text
E00
command: grep -rlF "func TestCDI_" internal
stdout: (empty)
exit: 1
RED reason: no function with the prefix exists, so every named test is absent and a selector over the prefix would sweep nothing.
collision check: grep -rlF "func TestCDI" internal printed nothing, exit 1 — the prefix is unused.
note: E00 proves absence, not failure; declaring an empty function would flip it. The behavioral red of every test is the failing run recorded in baseline.md at M0 (AC-CDI-024), and the gate is the named passing-test output lines.
```

```text
E01a
command: grep -c "card_identity" .moai/config/sections/workflow.yaml
stdout: 0
exit: 1
RED reason: no card_identity block exists in the repository config.
```

```text
E01b
command: grep -c "card_identity" internal/template/templates/.moai/config/sections/workflow.yaml
stdout: 0
exit: 1
RED reason: no card_identity block exists in the distributed template.
```

```text
E01c
command: grep -c "CardIdentity" internal/config/types.go
stdout: 0
exit: 1
RED reason: no CardIdentity type exists.
```

```text
E01d
command: grep -c "CardIdentity" internal/config/defaults.go
stdout: 0
exit: 1
RED reason: no CardIdentity default exists.
```

```text
E02
command: grep -rlF --exclude='*_test.go' "card_attempts" internal
stdout: (empty)
exit: 1
RED reason: the attempt archive does not exist in any non-test file; the sweep root `internal` is non-empty (it holds the 15 statements of plan.md § I).
```

```text
E03
command: grep -rlF --exclude='*_test.go' "card_current" internal
stdout: (empty)
exit: 1
RED reason: the project-scoped card record does not exist.
```

```text
E04
command: grep -rlF --exclude='*_test.go' "newFactoryIdentityCommand" internal/cli
stdout: (empty)
exit: 1
RED reason: the identity verbs do not exist.
```

```text
E05
command: grep -rlF --exclude='*_test.go' "newFactoryReconcileCommand" internal/cli
stdout: (empty)
exit: 1
RED reason: the reconcile verb does not exist.
```

```text
E06
command: grep -rlF --exclude='*_test.go' "card.migrated" internal
stdout: (empty)
exit: 1
RED reason: no migration event kind exists.
```

```text
E07
command: grep -rlF --exclude='*_test.go' "factory_operations" internal
stdout: (empty)
exit: 1
RED reason: the operation journal does not exist.
```

```text
E08
command: git ls-files .moai/reports/t1842/baseline.md
stdout: (empty)
exit: 0
RED reason: the baseline artifact is not tracked; the verdict is the stdout (git ls-files exits 0 whether or not it prints).
```

```text
E09
command: grep -rlF --exclude='*_test.go' "owner_epoch" internal
stdout: (empty)
exit: 1
note: observation, not a gate: no owner epoch exists in non-test code on B0, so the epoch-carrier fixture of AC-CDI-013 is constructed, not found.
```

```text
E10
command: grep -c "unsupported factory schema version" internal/homestate/factory.go
stdout: 1
exit: 0
note: regression guard, green on B0 by construction (no RED-now: the open-time refusal that REQ-CDI-010 rests on exists at B0, factory.go:368-370). It must print 1 or more at the tip; the behavior half is TestCDI_OldBuildRefusal.
```

```text
E11
command: grep -c "func TestRetireRunPreservesRowAndAppendsEvent\|func TestRetireRunIfDeadRefusesLiveAndIndeterminate\|func TestReconcileNeverRetiresLiveOrIndeterminate\|func TestEveryRetirementPathDeclinesUnenumeratedClassification" internal/homestate/factory_run_retire_test.go
stdout: 4
exit: 0
note: regression guard; the four tests named in AC-CDI-002 exist on B0. That they pass was not observed at plan time (no go test was run; plan-phase Gaps); baseline.md records it at M0.
```

```text
E12
command: grep -c "PRIMARY KEY(run_id, card_id)" internal/homestate/factory.go
stdout: 1
exit: 0
note: observation of the defect in spec § A.1, not a gate.
```

```text
E13
command: grep -rlF --exclude='*_test.go' "ErrCardNotMember" internal
stdout: (empty)
exit: 1
RED reason: the not-a-member error does not exist.
```

```text
E14
command: grep -rlF --exclude='*_test.go' "ErrCardEngaged" internal
stdout: (empty)
exit: 1
RED reason: the engaged-elsewhere error does not exist.
```

```text
E15
command: grep -rlF --exclude='*_test.go' "retireRunWithMigration" internal
stdout: (empty)
exit: 1
RED reason: the migrating retirement does not exist.
```

```text
E20a
command: grep -rlE "leader_approval|LeaderApproval|factorymsg" --include='operation_*.go' --exclude='*_test.go' internal/homestate
stdout: (empty)
exit: 1
note: silent at RED and at green; read only together with E20b. The import-graph test covers files that are not named operation_*.go.
```

```text
E20b
command: grep -rlF "func " --include='operation_*.go' --exclude='*_test.go' internal/homestate
stdout: (empty)
exit: 1
RED reason: no non-test operation_*.go file exists, so E20a's sweep is empty. At green it prints at least the journal and reconcile files.
```

```text
E25
command: grep -rlF --exclude='*_test.go' "seal_reason" internal
stdout: (empty)
exit: 1
RED reason: no attempt archive with a seal reason exists.
```

```text
E23a
command: git log --first-parent --no-merges --grep=t1842 --name-only --pretty=format: f4bcd4f2cc65342559b505f1e5a29df80f992894..HEAD -- . ':!internal/homestate' ':!internal/config' ':!internal/factorymsg' ':!internal/factory' ':!internal/cli/factory_*' ':!internal/cli/mcp_factory_*' ':!internal/cli/integration*' ':!.moai/config/sections/workflow.yaml' ':!internal/template/templates/.moai/config/sections/workflow.yaml' ':!.moai/specs/SPEC-CARD-DURABLE-IDENTITY-001' ':!.moai/reports/t1842'
stdout: (empty, recorded at B0 where the range is empty)
exit: 0
note: regression guard; the verdict is the stdout. § Guard measurements holds the must-print cases.
```

```text
E23b
command: git log --first-parent --invert-grep --grep=t1842 --pretty=%h f4bcd4f2cc65342559b505f1e5a29df80f992894..HEAD
stdout: (empty, recorded at B0)
exit: 0
note: regression guard; every commit on this branch carries the card id, absorbing merges included (the lane convention tags them).
```

```text
E24a
command: git log --format=%H --diff-filter=A -- .moai/reports/t1842/baseline.md
stdout: (empty at B0; at the run-phase tip it prints exactly one commit)
exit: 0
```

```text
E24b
command: git log --first-parent --reverse --format=%h --grep=t1842 f4bcd4f2cc65342559b505f1e5a29df80f992894..HEAD -- internal
stdout: (empty at B0; at the run-phase tip the first line is the baseline commit of E24a, abbreviated)
exit: 0
```

### Guard measurements and gaps

- **Scope guards red on a constructed change.** `sh /private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/3b8aad6a-ce40-4e01-92a2-a04a97198934/scratchpad/fx23.sh` builds a throwaway repository under the scratchpad (no card id in the path), copies the E23a and E23b forms with its own base, and printed: empty range: nothing, nothing; after a tagged in-scope commit (one file under `internal/homestate` and one under `.moai/reports/t1842`): nothing, nothing; after a tagged commit editing `docs/outside.md`: E23a printed `docs/outside.md`; after an untagged commit: E23b printed the abbreviated hash `5bcb46e`. Both guards are green on the tree and red on the constructed change.
- **Refusals by the worktree-session guard while authoring.** (1) A command assembling a path in a shell loop variable was refused ("a value computed at runtime ... cannot be shown not to be git"); it was split into plain separate commands. (2) A command that changed directory to the shared checkout before running `git` was refused; it was rerun from the worktree root. Neither refusal bears on a cited command.
- **Not observed at plan time.** No `go build` and no `go test` was run (the host load average read 20.63 at 09:33 and the standing practice is not to add load): the existing tests of AC-CDI-002 were not observed to pass, and no behavior of the planned code was exercised. The SQLite behaviors the design rests on were observed on the SQLite command line (version 3.54.0), not under the project's driver (research.md § 6).

---
id: SPEC-20261124-time-budget-tests
title: "Time-budget tests cannot fail the suite for machine load alone"
status: approved
revision: 2
created: 2026-11-24
approval: {by: leader, at: 2026-11-24, basis: "leader decision note 8046 of card C-440 (answers note 8045, 2026-11-24T00:23:54Z, under the user's delegation of 2026-11-23T13:57Z, not a direct user pick), approving revision 1 at head c38f24c0a1 with choice (a) and no follow-up card now; recorded by the lane, not approved by it", digest: "sha256:83de200e4393f9543661fcf8c396854e55044fab5aa2b061a31bb55850b1b308"}
---

# Time-budget tests cannot fail the suite for machine load alone

## Interview

- none: 별도 사용자 인터뷰는 없었어요. 요구사항은 카드 C-440 의 본문과 세 인수 기준, 리더 배차, C-352 노트의 실패 기록이 정했고, 판 1 은 리더 결정 노트 8046 이 승인했어요(사용자 위임 2026-11-23).

## Intent

A whole-suite run on NeXT `4872c62870` (2026-11-24) failed only `prop_reservation random_interleavings_keep_the_two_invariants`: "ran 1000 cases in 51.7 s, over its budget of 48 s (seed 13263398375705, load 25.4 on 16 CPUs)"; alone on the same build it passed 3 of 3 in 15.5 to 22.4 s at load 67 to 78. It is the second time a time budget alone stopped the landings (the first was build 19, one clock race). The budget already grows with the one-minute load average (SPEC-20261122-test-defects-session-reclaim REQ-7), but it reads that average once, when the run ends, and the average lags the contention the run lived through. This SPEC judges the property's own CPU time, which load does not stretch, and keeps wall time only as a warning and a hang ceiling. It also classifies every other wall-clock assertion in `tests/` so none stays unnamed.

## Non-goals

- No change to the generators, the seeds, the case count (1000) or the replay variables `PROP_SEED` and `PROP_ONLY`.
- No change to the product's own timeouts; only the tests that judge elapsed time.
- The load-scaled helper for the tight sites of Appendix B is slice 2 (D-2), not this slice.

## Requirements

- REQ-1: When a property run ends, the harness shall judge its time budget by the CPU time the running test thread used, not by wall-clock time.
- REQ-2 (REQ-1): IF the CPU time of a run reaches 30 s, THEN the property shall fail and name the property, the seed, the cases, the CPU time and the wall time.
- REQ-3 (REQ-1): IF the CPU time of a run is below 30 s and its wall time passes the load-scaled budget, THEN the harness shall print one warning line with the property, the seed, the wall time, the budget and the load instead of failing the property.
- REQ-4 (REQ-1): IF the wall time of a run reaches 240 s, THEN the property shall fail as a hang and name the property, the seed, the wall time and the CPU time.
- REQ-5 (REQ-1): IF the CPU time of the thread cannot be read, THEN the harness shall judge the wall time against the load-scaled budget as before and fail past it.
- REQ-6: IF a property case panics, THEN the property shall fail and name the property, the seed and the case index, whatever the elapsed time.
- REQ-7: The Appendix of this SPEC shall list every wall-clock assertion site of `tests/` by class, and name each site that keeps its fixed bound as an exception with the reason.

## Acceptance

- AC-1 (REQ-1, REQ-3): a run of 51.7 s wall time and 15 s CPU time at load 25.4 on 16 CPUs passes and prints the warning; the same wall time with 31 s CPU time fails · verify: `cargo test --locked --offline --test prop_reservation the_property_budget_`
- AC-2 (REQ-2): a run at 30 s CPU time fails with a message that names the property, the seed, the cases, the CPU time and the wall time; at 29.9 s it passes · verify: `cargo test --locked --offline --test prop_reservation the_property_budget_`
- AC-3 (REQ-4): a run of 240 s wall time fails as a hang whatever its CPU time; 239.9 s does not · verify: `cargo test --locked --offline --test prop_reservation the_property_budget_`
- AC-4 (REQ-5): with no CPU reading, a run over the load-scaled wall budget fails and one within it passes · verify: `cargo test --locked --offline --test prop_reservation the_property_budget_`
- AC-5 (REQ-6): a property whose case panics fails with the seed and case index and does not reach the budget check · verify: `cargo test --locked --offline --test prop_reservation a_failing_case_`
- AC-6 (REQ-1): the real thread CPU reading rises when the thread computes and stays near zero while it sleeps · verify: `cargo test --locked --offline --test prop_reservation the_thread_cpu_time_`
- AC-7 (REQ-7): Appendix A and B list each wall-clock site of the research table with its file and line, class and reason, and the 13 configured product timeouts, and state the row counts of the table · verify: `moai spec check SPEC-20261124-time-budget-tests --project .`

## Scope

- write: `.moai/specs/SPEC-20261124-time-budget-tests/**`
- write: `tests/common/prop.rs`
- read-only: `tests/prop_reservation.rs`, `tests/prop_scope_paths.rs`, `tests/prop_headver.rs` (each includes the harness file of the write list above; their calls do not change)
- forbidden: `src/`, `tests/common/pty.rs` and every other test file (slice 2)

## Plan

1. RED: pin the pure budget function `over_budget` with CPU time and wall time as separate inputs (the cases of AC-1 to AC-5), and the CPU reading (AC-6); they fail before the code.
2. GREEN: `prop.rs` reads the thread CPU time with `clock_gettime(CLOCK_THREAD_CPUTIME_ID)` at the start and the end of `run_cases`, judges it, and turns the load-scaled wall budget into a warning.
3. Measure one real `prop_reservation` run: CPU time and wall time, alone and with a second busy process, and record both in the Result as measured.
4. Guard mutants: judging wall time again, and dropping the hang ceiling, each must fail a test.
5. Result: the Appendix counts, rechecked with the greps named there.

## Risks

- A property slowed by file-system waits (fsync) and not by CPU would pass the CPU budget; the 240 s hang ceiling is what still catches a real stall, and a wall time past the load-scaled budget is printed.
- `CLOCK_THREAD_CPUTIME_ID` is measured on macOS here; Linux and Windows are not measured (Windows has no reading, so REQ-5 applies there).
- A property that moves work to another thread would escape the CPU count; `prop_reservation`, `prop_scope_paths` and `prop_headver` run their cases on the test thread (to be read again at implementation).

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | CPU time is the hard budget (30 s); load-scaled wall time is a warning; the hang ceiling is 240 s (the former cap of eight times) | card C-440 options "budget scales with load" and "warning when no invariant broke"; the load average is read once at the end and lags | restore `over_budget` of revision 0 of `prop.rs` |
| D-2 | Slice 1 (this SPEC) covers the harness and the classification; the tight and moderate sites of Appendix B move to one follow-up card with a shared load-scaled helper | 106 sites in 54 files across other cards' scopes; hang-class sites need no change | the leader widens this SPEC's write list to Appendix B instead |

### Appendix A — sites that keep their bound (exception)

Source: one `moai-research` read of `tests/**/*.rs` on 2026-11-24 (00:13Z to 00:20Z), copied here from `~/.moai/reports/moai-adk/c440-wallclock-sites-2026-11-24.md` (sha256 `49a0200f3f1b3af34bea2942667fc7675b6b8ac9f351e4ab2372571425cb39f1`, read 2026-11-24). The class and the "normal" duration come from code, comments and product constants; no duration was measured. The lane did not search the tests again: it checked that each file and line below exists in this tree, and it did not check that the bound at the line is the one named.

Classes: HANG = normal duration under a tenth of the bound; MOD = a tenth to a third; TIGHT = at least a third, or a bound of at most 2 s on a process start; `*` marks the research's tightest sites. NEG = a check that load cannot fail; IGN = an `#[ignore]` test the suite does not run.

Reasons for a fixed bound:

- R1: a hang guard, not a speed limit; a bound at least ten times the normal duration has no load that stretches the run past it.
- R2: load cannot fail the site (it asserts that something does not arrive, or the test is ignored).
- R3: slice 2 (D-2) gives the site one load-scaled helper; until then it keeps its fixed bound and is a named exception of this SPEC.
- R4: the product's own timeout, a non-goal of this SPEC; a test that expects it not to fire stays a named exception.
- R5: a lower bound, which a loaded machine cannot break.

Counts in this table: 75 rows in 42 files (HANG 73, NEG 1, IGN 1).

| Id | File:line | Site | Bound | Class | Reason |
| --- | --- | --- | --- | --- | --- |
| A-001 | card_line_lease_cli.rs:498 | line_lease_lets_two_cards_each_add_a_line_to_the_same_file | 60 s | HANG | R1 |
| A-002 | versioning_cli.rs:48 | run_within (run() at :56) | 60 s | HANG | R1 |
| A-003 | home_install_plan_cli.rs:824 | install_does_not_open_a_master_key_that_is_a_fifo | 20 s | HANG | R1 |
| A-004 | home_state_selective.rs:71 | plan_env | 60 s | HANG | R1 |
| A-005 | index_cli.rs:137 | cli_before_deadline (comment: a hang guard, not a speed limit) | 120 s | HANG | R1 |
| A-006 | install_candidates.rs:369 | the_probe_limit_holds_when_a_wrapper_child_keeps_stdout_open | 50 s | HANG | R1 |
| A-007 | statusline_counts_refresh.rs:130 | wait_for | 30 s | HANG | R1 |
| A-008 | statusline_counts_refresh.rs:186 | the_render_does_not_wait_for_a_slow_gh | <10 s | HANG | R1 |
| A-009 | statusline_counts_refresh.rs:455 | orca_pass_through_does_not_wait_for_a_slow_script | <10 s | HANG | R1 |
| A-010 | go_state_classes_cli.rs:173 | run_env | 120 s | HANG | R1 |
| A-011 | go_state_codex_cli.rs:228 | run_env | 120 s | HANG | R1 |
| A-012 | run_verify.rs:559 | timeout_bounds_runtime_even_with_grandchildren | <60 s | HANG | R1 |
| A-013 | run_verify.rs:571 | timeout_bounds_runtime_even_with_grandchildren | <60 s | HANG | R1 |
| A-014 | forge_cli.rs:514 | transport_times_out_on_stalled_gh | 60 s | HANG | R1 |
| A-015 | forge_cli.rs:516 | transport_times_out_on_stalled_gh | <60 s | HANG | R1 |
| A-016 | forge_cli.rs:523 | transport_times_out_on_stalled_gh | 30 s | HANG | R1 |
| A-017 | edit_diagnostics_cli.rs:518 | one_check_runs_per_project_at_a_time | 20 s | HANG | R1 |
| A-018 | edit_diagnostics_cli.rs:597 | a_checker_that_leaves_a_background_child_is_cleaned_up | <5 s | HANG | R1 |
| A-019 | edit_diagnostics_cli.rs:792 | budget_the_off_path_is_fast (median under 50 ms, `#[ignore]`) | 50 ms | IGN | R2 |
| A-020 | heavy_slot_fifo_cli.rs:43 | Spawned::finish | 120 s | HANG | R1 |
| A-021 | heavy_slot_fifo_cli.rs:155 | wait_within (wait_for 120 s, AT_ONCE 20 s) | 120 s, 20 s | HANG | R1 |
| A-022 | heavy_slot_cli.rs:63 | finish | 120 s | HANG | R1 |
| A-023 | heavy_slot_cli.rs:99 | wait_wrapper_only | 120 s | HANG | R1 |
| A-024 | heavy_slot_cli.rs:161 | wait_for | 120 s | HANG | R1 |
| A-025 | ledger_recover.rs:1511 | wait_for | 60 s | HANG | R1 |
| A-026 | card_wait_cli.rs:149 | finish | 60 s | HANG | R1 |
| A-027 | card_wait_cli.rs:267 | waited_secs upper bound | <50 s | HANG | R1 |
| A-028 | card_wait_cli.rs:348 | waited_secs upper bound | <40 s | HANG | R1 |
| A-029 | card_wait_cli.rs:537 | since_returns_notes_already_above_the_mark_at_once | <20 s | HANG | R1 |
| A-030 | card_wait_cli.rs:570 | an_unknown_card_or_an_absent_ledger_is_refused_before_the_wait | <20 s | HANG | R1 |
| A-031 | update_model_step_reachable_tty.rs:147 | on_a_terminal | 90 s | HANG | R1 |
| A-032 | launcher_auto_init.rs:82 | wait_ready | 120 s | HANG | R1 |
| A-033 | launcher_auto_init.rs:518 | wait on the launched session | 120 s | HANG | R1 |
| A-034 | launcher_auto_init.rs:637 | wait on the launched session | 120 s | HANG | R1 |
| A-035 | launcher_auto_init.rs:721 | wait on the launched session | 90 s | HANG | R1 |
| A-036 | update_command_cli.rs:583 | stdin_that_is_not_a_terminal_is_never_asked | <30 s | HANG | R1 |
| A-037 | pty_helper.rs:53 | without_an_answer_a_question_gets_end_of_input_and_stays_on_screen | <20 s | HANG | R1 |
| A-038 | pty_helper.rs:185 | pid_in | 5 s | HANG | R1 |
| A-039 | pty_helper.rs:196 | gone | 5 s | HANG | R1 |
| A-040 | common/pty.rs:115 | bytes_of, PIPES_CLOSE_WITHIN | 10 s | HANG | R1 |
| A-041 | common/pty.rs:428 | bytes_of, PIPES_CLOSE_WITHIN | 10 s | HANG | R1 |
| A-042 | common/pty.rs:291 | panic after the per-test limit (shared by 12 test files) | 60 s default | HANG | R1 |
| A-043 | common/pty.rs:331 | panic after the per-test limit (shared by 12 test files) | 60 s default | HANG | R1 |
| A-044 | hook_session_start_cli.rs:204 | Fx::run | 20 s | HANG | R1 |
| A-045 | hook_session_start_cli.rs:874 | ac8_links_under_home_are_not_followed | <10 s | HANG | R1 |
| A-046 | ledger_card_landed_migration_cli.rs:501 | child_open_for_writing | 60 s | HANG | R1 |
| A-047 | ledger_open.rs:616 | wait_for | 30 s | HANG | R1 |
| A-048 | home_install_apply_cli.rs:730 | apply_does_not_open_a_master_key_that_is_a_fifo | 60 s | HANG | R1 |
| A-049 | statusline_cards.rs:1436 | line_three_ladder_one_render_starts_one_git_and_one_orca_wrapper_at_most | 30 s | HANG | R1 |
| A-050 | statusline_git_env.rs:329 | run_with | 30 s | HANG | R1 |
| A-051 | ledger_reservation.rs:1150 | wait_for | 60 s | HANG | R1 |
| A-052 | skills_legacy_cli.rs:169 | a_fifo_skill_md_is_reported_and_never_read | 10 s | HANG | R1 |
| A-053 | hook_secret_project_keys_cli.rs:298 | linear_judging_time_on_three_mib_of_near_misses | <20 s | HANG | R1 |
| A-054 | hook_secret_tokens_cli.rs:520 | linear_long_content_is_judged_in_proportion_to_its_length (BOUND_SECS) | 30 s | HANG | R1 |
| A-055 | hook_secret_tokens_cli.rs:533 | linear_long_content_is_judged_in_proportion_to_its_length (BOUND_SECS) | 30 s | HANG | R1 |
| A-056 | launcher_agent_numbering.rs:239 | wait_ready | 120 s | HANG | R1 |
| A-057 | launcher_agent_numbering.rs:648 | wait on the launched session | 120 s | HANG | R1 |
| A-058 | launcher_glm_key_file.rs:176 | an_unreadable_entry_at_the_legacy_path_is_named_without_opening_it | 120 s | HANG | R1 |
| A-059 | base_agents_discovery_record.rs:121 | run_with_deadline | 60 s | HANG | R1 |
| A-060 | base_agents_discovery_record.rs:471 | run_with_deadline | 30 s | HANG | R1 |
| A-061 | base_agents_discovery_record.rs:598 | run_with_deadline | 30 s | HANG | R1 |
| A-062 | hook_dispatch_cli.rs:583 | a_timed_out_preflight_is_non_blocking_by_the_host_rules | <90 s | HANG | R1 |
| A-063 | product_cli.rs:3059 | claude_doctor_* | <60 s | HANG | R1 |
| A-064 | product_cli.rs:3107 | claude_doctor_* | <60 s | HANG | R1 |
| A-065 | product_cli.rs:3151 | claude_doctor_* | <60 s | HANG | R1 |
| A-066 | key_set_prompt_terminal_cli.rs:203 | pid() | 30 s | HANG | R1 |
| A-067 | key_set_prompt_terminal_cli.rs:446 | run_prompt | 60 s | HANG | R1 |
| A-068 | home_state_plan.rs:204 | a_live_go_session_blocks_the_move | 20 s | HANG | R1 |
| A-069 | check_affected_one_slot.rs:445 | a_check_inside_a_one_slot_acquire_does_not_wait_on_itself | 60 s | HANG | R1 |
| A-070 | card_cleanup_outbox_cli.rs:1713 | fu2_ test | 20 s | HANG | R1 |
| A-071 | card_cleanup_outbox_cli.rs:1810 | fu1_ test | 20 s | HANG | R1 |
| A-072 | spawn_cap_test.rs:78 | a_dropped_permit_is_freed_for_the_next_taker | 10 s | HANG | R1 |
| A-073 | spawn_cap_test.rs:92 | a_dropped_permit_is_freed_for_the_next_taker | 10 s | HANG | R1 |
| A-074 | spawn_cap_test.rs:87 | negative check (asserts nothing arrives); load cannot fail it | 10 s | NEG | R2 |
| A-075 | ledger_folders.rs:546 | lock_promptly | 5 s | HANG | R1 |

### Appendix B — tight and moderate sites, product timeouts and lower bounds (slice 2)

#### B.1 — tight and moderate sites

Counts in this table: 36 rows in 25 files (TIGHT* 14, TIGHT 3, MOD 19). Reason R3 for every row. The journal-lock retry loops (`handoff_cli`, `queue_cli`, `claim_cli`, `journal_lock_busy_cli`) are the load-sensitive case on record: `claim_cli.rs` comments at `:103-106` and `:1128` say macOS answered WouldBlock under parallel test load.

| Id | File:line | Site | Bound | Class | Reason |
| --- | --- | --- | --- | --- | --- |
| B-001 | card_line_lease_cli.rs:657 | line_lease_wait_secs_has_its_edges | <3 s | MOD | R3 |
| B-002 | card_line_lease_cli.rs:678 | line_lease_waits_for_its_limit_and_no_longer | <5 s | TIGHT* | R3 |
| B-003 | versioning_cli.rs:421 | a_named_pipe_as_packed_refs_is_unavailable_within_two_seconds | 2 s | TIGHT* | R3 |
| B-004 | versioning_cli.rs:532 | the_size_and_entry_limits_hold_at_their_edges | 5 s | MOD | R3 |
| B-005 | install_candidates.rs:393 | same test, loop of 100 x 20 ms | about 2 s | MOD | R3 |
| B-006 | statusline_counts_refresh.rs:542 | a_hanging_wrapper_ends_at_its_time_limit_and_the_next_render_starts_one (product alarm at 10 s) | 20 s | TIGHT | R3 |
| B-007 | lane_send_cli.rs:440 | test that says the text may have reached the lane (product cut 2 s) | <5 s | TIGHT* | R3 |
| B-008 | lane_send_cli.rs:528 | test where a hanging orca is cut off and nothing is sent | <5 s | TIGHT* | R3 |
| B-009 | edit_diagnostics_cli.rs:435 | failures_are_one_line_and_never_block | <5 s | TIGHT* | R3 |
| B-010 | edit_diagnostics_cli.rs:444 | retry loop of 200 x 10 ms | about 2 s | MOD | R3 |
| B-011 | edit_diagnostics_cli.rs:609 | retry loop of 100 x 10 ms | about 1 s | TIGHT* | R3 |
| B-012 | ledger_recover.rs:1886 | a_ledger_held_by_another_connection_is_busy_not_damaged_and_its_write_is_not_lost | <4 s | MOD | R3 |
| B-013 | card_wait_cli.rs:290 | without_a_matching_note_the_wait_ends_at_its_timeout | <12 s | MOD | R3 |
| B-014 | card_wait_cli.rs:299 | without_a_matching_note_the_wait_ends_at_its_timeout | <12 s | MOD | R3 |
| B-015 | status_line_cli.rs:663 | a_moai_that_hangs_on_the_probe_is_unprobed | <20 s | MOD | R3 |
| B-016 | process_table_default.rs:146 | a_group_started_shell_holder_leaves_no_child (loop) | 3 s | MOD | R3 |
| B-017 | pty_helper.rs:238 | the_limit_counts_again_from_the_answer (two 1.3 s sleeps against a 2 s limit, panic at common/pty.rs:291) | 2 s | TIGHT* | R3 |
| B-018 | hook_session_start_cli.rs:768 | ac7_an_unreadable_ledger_gives_the_branch_and_one_receipt_line | <3 s | MOD | R3 |
| B-019 | card_timeline_cli.rs:435 | a_busy_ledger_a_missing_card_and_bad_input_use_the_existing_envelope | <3 s | MOD | R3 |
| B-020 | statusline_cards.rs:476 | an_unreadable_ledger_leaves_the_line_as_it_is_without_one (waited) | <2.5 s | MOD | R3 |
| B-021 | lane_screen_cli.rs:1527 | lane_screen_scan_an_orca_child_holding_the_output_does_not_hold_the_scan (product cut 1 s) | <6 s | MOD | R3 |
| B-022 | base_agents_discovery_record.rs:589 | a_child_is_killed_at_the_deadline_and_a_finished_one_is_read_whole | <15 s | MOD | R3 |
| B-023 | headless_run_cli.rs:1084 | a_wall_clock_limit_kills_the_run | <5 s | MOD | R3 |
| B-024 | headless_run_cli.rs:1112 | a_large_answer_is_not_killed_by_a_full_pipe | <5 s | MOD | R3 |
| B-025 | headless_run_cli.rs:1140 | a_background_sleeper_does_not_stretch_the_call | <4 s | MOD | R3 |
| B-026 | headless_run_cli.rs:1170 | a_group_leaving_grandchild_does_not_stretch_the_call | <6 s | TIGHT* | R3 |
| B-027 | headless_run_cli.rs:1201 | an_abandoned_join_keeps_the_answer_already_read | <6 s | TIGHT* | R3 |
| B-028 | product_cli.rs:3032 | assert_fixture_process_stopped | 2 s | TIGHT | R3 |
| B-029 | profile_plan_cli.rs:501 | a_hanging_remote_check_is_killed_and_the_plan_still_succeeds | 30 s | MOD | R3 |
| B-030 | handoff_cli.rs:183 | journal_of (journal-lock retry loop of 200 x 10 ms) | about 2 s | TIGHT* | R3 |
| B-031 | queue_cli.rs:73 | locked (journal-lock retry loop of 200 x 10 ms) | about 2 s | TIGHT* | R3 |
| B-032 | claim_cli.rs:115 | lock_journal (journal-lock retry loop of 200 x 10 ms; comment at :103-106 says macOS answered WouldBlock under parallel test load) | about 2 s | TIGHT* | R3 |
| B-033 | claim_cli.rs:1137 | claim_busy_never_waits (journal-lock retry loop; comment at :1128) | about 2 s | TIGHT* | R3 |
| B-034 | journal_lock_busy_cli.rs:482 | every_journal_writing_op_answers_one_busy_envelope (journal-lock retry loop of 200 x 10 ms) | about 2 s | TIGHT* | R3 |
| B-035 | card_cleanup_outbox_cli.rs:1070 | ac5_a_run_that_stopped_after_the_worktree_step_is_finished_by_the_next (30 s for a 12 s lease) | 30 s | TIGHT | R3 |
| B-036 | launcher_liveness_cli.rs:316 | a_named_launch_holds_its_live_lock_while_its_host_runs | 10 s | MOD | R3 |

#### B.2 — product timeouts that a test configures and expects not to fire

Counts in this table: 13 settings in 9 files. Reason R4 for every row; the research did not class them.

| Id | File:line | Setting | Value | Note | Reason |
| --- | --- | --- | --- | --- | --- |
| T-01 | install_candidates.rs:359 | MOAI_TEST_PROBE_TIMEOUT_MS | 2000 ms | comment: load 27-39 seen, #2843 | R4 |
| T-02 | status_line_cli.rs:660 | MOAI_TEST_PROBE_TIMEOUT_MS | 3000 ms |  | R4 |
| T-03 | binary_state_cli.rs:332 | MOAI_TEST_PROBE_TIMEOUT_MS | 2000 ms |  | R4 |
| T-04 | lane_send_cli.rs:158 | MOAI_ORCA_TIMEOUT_SECS | 2 s |  | R4 |
| T-05 | lane_screen_cli.rs:1521 | MOAI_ORCA_TIMEOUT_SECS | 1 s |  | R4 |
| T-06 | headless_run_cli.rs:1109 | run_with_limit (headless call limit) | 5 s |  | R4 |
| T-07 | headless_run_cli.rs:1136 | run_with_limit (headless call limit) | 1 s |  | R4 |
| T-08 | headless_run_cli.rs:1166 | run_with_limit (headless call limit) | 1 s |  | R4 |
| T-09 | headless_run_cli.rs:1197 | run_with_limit (headless call limit) | 1 s |  | R4 |
| T-10 | headless_run_cli.rs:1359 | run_with_limit (headless call limit) | 5 s |  | R4 |
| T-11 | boundary_repro.rs:689 | test-side `timeout` of 10 s | 10 s |  | R4 |
| T-12 | product_cli.rs:3074-3086 | product VERSION_TIMEOUT (test claude_doctor_waits_for_inherited_stdout_within_the_deadline) | 5 s |  | R4 |
| T-13 | card_cleanup_outbox_cli.rs:1039 | `lease_secs` | 12 s |  | R4 |

#### B.3 — lower-bound sites (not counted)

9 sites where the assertion is a lower bound that load cannot break; the research says two of them can fail on a late start and does not name them (the comment of `card_wait_cli.rs:375` names itself). Reason R5.

| Id | File:line | Note | Reason |
| --- | --- | --- | --- |
| L-1 | card_line_lease_cli.rs:505 | `waited_ms >= 500` | R5 |
| L-2 | card_line_lease_cli.rs:531 | lower bound | R5 |
| L-3 | card_line_lease_cli.rs:677 | lower bound | R5 |
| L-4 | card_wait_cli.rs:287 | lower bound | R5 |
| L-5 | card_wait_cli.rs:288 | lower bound | R5 |
| L-6 | card_wait_cli.rs:297 | lower bound | R5 |
| L-7 | card_wait_cli.rs:298 | lower bound | R5 |
| L-8 | card_wait_cli.rs:375 | `waited_secs >= 2`; its own comment says it holds only while the start takes under 1.5 s | R5 |
| L-9 | worktree_slot_cli.rs:749 | lower bound | R5 |

Research gaps, kept from the report: no duration was measured; about 80 `SystemTime::now` lines were judged time-of-day seeds from the grep line alone; `base_agents_discovery_record.rs:156,265`, `go_settings_cli.rs:1028` and part of `key_set_prompt_terminal_cli.rs` were not read to the end; a wall-clock check hidden in a helper with an unusual name could be missing.

Counts of the research (its own, not reconciled): 106 failing sites in 54 files, 13 product timeouts, 9 lower bounds. Counts of this table (rows, by `grep -c` of the Id column): 75 + 36 = 111 rows in 54 files, 13 product timeouts, 9 lower bounds. The table lists each line on its own row, and it carries the two rows that cannot fail (NEG, IGN), so it holds more rows than the research's 106.

## Result

Head `b4c437d4c2` on base `381d1caffc` (changed file: `tests/common/prop.rs`). Checks of 2026-11-24 on a Mac with 16 CPUs and a load of 30 to 50 from other lanes.

- RED (`5be8d01168`): the six new tests were written first and did not compile against the old code (`error[E0425]: cannot find function `judge` in this scope`, `cannot find type `Verdict``, scratch crate, exit 101). The AC-5 test needs only `run_cases`, so it is a regression guard and was not red by itself.
- GREEN (`b4c437d4c2`): `cargo test --offline --locked --test prop_reservation --test prop_scope_paths --test prop_headver -- --nocapture` under `moai heavy-slot-acquire`, exit 0: `prop_headver` 13 passed, `prop_reservation` 11 passed (34.03 s), `prop_scope_paths` 11 passed. Log: `target/time-budget-tests/real-run.log`.
- Measured (the hypothesis of D-1): one run of `prop_reservation random_interleavings_keep_the_two_invariants` under `/usr/bin/time -l` with the slot wrapper, load 36 on 16 CPUs: 41.08 s wall, 3.28 s user plus 5.48 s system CPU (process total, about 8.8 s). The wall time is about four and a half times the CPU time, so the wait is not CPU; the CPU budget of 30 s keeps more than three times of room. The failing run of the card (51.7 s wall) has no CPU reading; only this one measurement exists. Log: `target/time-budget-tests/measure.log`.
- Guard mutants on the committed file, scratch crate outside the repository (`libc =0.2.189`), each against the 10 budget tests: (m3) the CPU-budget arm made unreachable (`Some(cpu) if false && cpu >= ...`) exit 101, `the_property_budget_fails_at_thirty_seconds_of_cpu_time` and `the_property_budget_warns_not_fails_when_only_the_wall_time_passed_the_budget` failed; (m4) the warning turned back into a failure exit 101, the warns-not-fails test failed. The implementing agent's own mutants (wall judged again; hang ceiling removed) also exited 101 (its report; not rerun here).
- AC-1: met — `cargo test --locked --offline --test prop_reservation the_property_budget_` exit 0, 5 passed (the 51.7 s / 15 s CPU / load 25.4 case warns, the 31 s CPU case fails)
- AC-2: met — same command, `the_property_budget_fails_at_thirty_seconds_of_cpu_time` passed: 30.0 s fails and names the property, cases, seed, CPU and wall time, 29.9 s is within
- AC-3: met — same command, `the_property_budget_fails_a_hang_at_two_hundred_forty_seconds` passed: 240 s fails whatever the CPU time (also with no CPU reading), 239.9 s does not
- AC-4: met — same command, `the_property_budget_uses_the_wall_time_when_no_cpu_time_is_read` passed (61 s fails and 59 s is within against a 60 s budget)
- AC-5: met — `cargo test --locked --offline --test prop_reservation a_failing_case_` exit 0, 1 passed; the message names `property p`, `seed 7`, `case 2` and has no word of budget
- AC-6: met — `cargo test --locked --offline --test prop_reservation the_thread_cpu_time_` exit 0, 1 passed (the reading rises by 20 ms while computing and by under 100 ms over a 300 ms sleep; macOS only, Linux and Windows not run)
- AC-7 (revision 1): not_run as worded — `moai spec check .moai/specs/SPEC-20261124-time-budget-tests/spec.md` is refused (`the id is not SPEC-<yyyymmdd>-<slug>`), and revision 1 named the sites by file group only. Audit #440 failed it for that reason.
- AC-7 (revision 2, repair after audit #440 under the leader decision of that note): the verify command is now `moai spec check SPEC-20261124-time-budget-tests --project .`. Appendix A and B list every site of the research table by file and line with its class and reason (A: 75 rows in 42 files; B.1: 36 rows in 25 files; B.2: 13 product timeouts; B.3: 9 lower bounds), copied from `~/.moai/reports/moai-adk/c440-wallclock-sites-2026-11-24.md`; each file and line was checked to exist in this tree, the bound at each line was not checked. The research counted 106 failing sites in 54 files; the table has 111 rows in 54 files (it lists each line on its own row and carries one NEG and one IGN row), and the lane did not reconcile the difference. Per the leader's answer (card note 8046) no follow-up card is opened now: Appendix B stays the named exception, and the leader opens one card when one of those sites fails for load alone. Code is unchanged since `b4c437d4c2`. Check: `moai spec check SPEC-20261124-time-budget-tests --project .` on this tree reports one error, `approval_digest` (the approval block still names the digest of revision 1; the lane does not record an approval); a scratch copy whose approval digest is the one `moai spec show` prints for revision 2 reports 0 errors and 0 warnings.
- Gaps: the `Warn` line of `run_cases` printed to stderr was never observed in a real run (the real runs stayed within the load-scaled budget); only `judge` is pinned for it. The 120 s cap inside the thread-CPU test is a wall-clock guard against a reading that never rises, named here as a hang guard.

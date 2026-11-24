---
id: SPEC-20261124-time-budget-tests
title: "Time-budget tests cannot fail the suite for machine load alone"
status: approved
revision: 1
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
- AC-7 (REQ-7): Appendix A and B name 106 sites in 54 files and 13 configured product timeouts, each with its class and, for a fixed bound, the reason · verify: `moai spec check .moai/specs/SPEC-20261124-time-budget-tests/spec.md`

## Scope

- write: `.moai/specs/SPEC-20261124-time-budget-tests/**`
- write: `tests/common/prop.rs`
- read-only: `tests/prop_reservation.rs`, `tests/prop_scope_paths.rs`, `tests/prop_headver.rs` (they include `prop.rs`; their calls do not change)
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

### Appendix A — hang-detector sites (kept as they are, exception)

The research read 106 failing sites in 54 files, 2026-11-24 (`grep` patterns in the Result). Sites of class HANG fail only when something stalls for tens of seconds, because their bound (20 to 120 s) is at least ten times the normal duration: the hang guards of `heavy_slot_cli`, `heavy_slot_fifo_cli`, `index_cli`, `go_state_classes_cli`, `go_state_codex_cli`, `home_state_selective`, `home_install_plan_cli`, `home_install_apply_cli`, `launcher_auto_init`, `launcher_agent_numbering`, `launcher_glm_key_file`, `ledger_recover` (`wait_for`), `ledger_open`, `ledger_reservation`, `ledger_card_landed_migration_cli`, `run_verify`, `forge_cli`, `update_model_step_reachable_tty`, `update_command_cli`, `common/pty.rs`, `base_agents_discovery_record` (60 s and 30 s), `hook_secret_*`, `skills_legacy_cli`, `statusline_git_env`, `key_set_prompt_terminal_cli`, `check_affected_one_slot`, `spawn_cap_test`, `ledger_folders`, `product_cli` (60 s), `hook_dispatch_cli`, `card_cleanup_outbox_cli` (marker waits) and the 20 to 30 s sites of `statusline_counts_refresh`, `statusline_cards` and `card_wait_cli`. Exception reason: a hang guard is not a speed limit; a bound this far above normal has no load that stretches it past the limit.

### Appendix B — tight and moderate sites (slice 2)

Class TIGHT or MOD (normal duration at least a tenth of the bound, or a bound under 10 s on a process start). The research marked the eight tightest: `pty_helper.rs:233` (two 1.3 s sleeps against a 2 s limit), `card_wait_cli.rs:375` (documented to fail when the start takes over 1.5 s), `headless_run_cli.rs:1166` and `:1197` (a python3 start under a 1 s limit), `edit_diagnostics_cli.rs:425` and `:602`, the 2 s journal-lock retry loops (`handoff_cli`, `queue_cli`, `claim_cli`, `journal_lock_busy_cli`), `lane_send_cli.rs:440` and `:528` (bound 5 s, normal 2 s), `card_line_lease_cli.rs:678` and `versioning_cli.rs:421`. About twenty more are moderate (bounds of 3 to 12 s). Slice 2 gives them one load-scaled helper; until then each keeps its fixed bound and is a named exception of this SPEC.

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
- AC-7: not_run as worded — `moai spec check .moai/specs/SPEC-20261124-time-budget-tests/spec.md` is refused (`the id is not SPEC-<yyyymmdd>-<slug>`); `moai spec check SPEC-20261124-time-budget-tests --project .` answers 0 errors and 0 warnings. Appendix A and B name the sites by file group and class and the eight tightest by line; the per-site table (106 sites in 54 files, 13 configured product timeouts) is in `~/.moai/reports/moai-adk/c440-wallclock-sites-2026-11-24.md`. The counts are one research agent's curated counts, not recounted by the lane. Per the leader's answer (card note 8046) no follow-up card is opened now: Appendix B stays the named exception, and the leader opens one card when one of those sites fails for load alone.
- Gaps: the `Warn` line of `run_cases` printed to stderr was never observed in a real run (the real runs stayed within the load-scaled budget); only `judge` is pinned for it. The 120 s cap inside the thread-CPU test is a wall-clock guard against a reading that never rises, named here as a hang guard.

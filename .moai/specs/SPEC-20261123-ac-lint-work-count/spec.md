---
id: SPEC-20261123-ac-lint-work-count
title: "The AC command lint test bounds work, not seconds"
status: draft
revision: 1
created: 2026-11-23
---

# The AC command lint test bounds work, not seconds

## Intent

`the_repository_lint_starts_no_cargo_and_ends_within_five_seconds` in `tests/ac_command_lint.rs` decided by the clock: the lint over the repository had to end in under 5 s. It passed in 1.4–2.3 s on a quiet machine and failed at 5.18 s under machine load 200 on 2026-11-22. A time limit cannot tell a slow machine from a slow lint, and `skills/moai/workflows/verify.md` says a test must not depend on wall-clock time. This is the second test of the same class after the card secret screen (`SPEC-20261120-card-screen-work-count`), so it gets the same cure: count the work. What the old limit protected is that the lint stays a pure reader: it starts no process (the 68–112 s cargo build of `SPEC-20261122-ac-lint-no-cargo` was a spawned process) and it reads each recorded file and each test source once (`list_tests` caches one reading per target). Both facts are countable, and no load changes a count. The leader's card answer of 2026-11-22 added a second test of the same class, `a_ledger_held_at_the_start_counts_in_the_wait` in `tests/card_wait_cli.rs` (a lock held by sleeping 3.5 s, then an elapsed time under 5 s; measured 3.55 s, the wrong build 5.57 s), and widened the scope to it and to `src/card/wait.rs`. The two facts it protects are orderings, not durations: the timeout clock starts before the start read of the ledger, and the sleep between polls stops at the deadline. They are counted on a clock the test moves by hand. The readers are the people who run `check affected` and the auditors who read a red run.

## Non-goals

- Changing `tools/check_ac_commands.py`; the test observes the tool, it does not change it.
- Replacing the other elapsed-time assertions of the repository; they are listed in the Result, and only one that has failed under load is a reason to change it.
- Changing the output of `card wait`: no field is added, the result JSON is the same.
- Keeping any second-based limit, not even a backstop that never decides.

## Requirements

- REQ-1: While the lint runs over the repository, the lint test shall fail when the interpreter's audit events show that the lint started any process.
- REQ-2: While the lint runs over the repository, the lint test shall fail when the lint opens a recorded document or a target root more than once, or a shared test module more often than the number of targets it listed.
- REQ-3: While the lint runs over the repository, the lint test shall fail when the lint opens a repository file outside `docs/`, `.moai-cli/reports/` and `tests/`.
- REQ-4: While the lint runs over the repository, the lint test shall fail when the counter saw fewer file opens than the recorded files the lint reports having checked, so a counter that stopped counting cannot pass.
- REQ-5: The lint test shall not read the clock to decide its result.
- REQ-6: While `run_wait` runs on a clock the test moves by hand, the unit tests shall fail when the deadline is fixed after the start read, when the clock is read after the start read, or when the sleep between polls passes the deadline.
- REQ-7: The process test of a ledger held at the start shall bound the wait by lower limits only, so that load can make it easier to meet and never harder.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4): the lint over the repository exits 0 with 0 processes started, every file within its open bound and no open outside the three folders · verify: `cargo nextest run --locked --offline --test ac_command_lint`
- AC-2 (REQ-1): a lint that starts one process of any name makes the test fail · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-3 (REQ-2): a lint that reads each target's sources again for every command (the cache removed) makes the test fail, under the same load as a quiet run · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-4 (REQ-4): a counter that records nothing makes the test fail · verify: mutant in the Result, `cargo nextest run --locked --offline --test ac_command_lint`
- AC-5 (REQ-5): the test file has no clock read, and the test passes 30 of 30 times with a machine load of 100 or more · verify: `rg -n "Instant|elapsed|Duration|SystemTime" tests/ac_command_lint.rs` prints nothing, and the load run in the Result
- AC-6 (REQ-6): a start read of 3.5 s against a timeout of 2 s leaves one poll, no sleep and a wait of 3.5 s; a start read of 0.5 s leaves three polls with sleeps of 1 s and 0.5 s; an interval of 30 s with 300 ms left sleeps 300 ms · verify: `cargo nextest run --locked --offline --lib card::wait`, and the mutants in the Result
- AC-7 (REQ-7): `a_ledger_held_at_the_start_counts_in_the_wait` has no elapsed upper bound and passes 30 of 30 times with a machine load of 100 or more · verify: `cargo nextest run --locked --offline --test card_wait_cli`, and the load run in the Result

## Scope

- write: `.moai/specs/SPEC-20261123-ac-lint-work-count/spec.md`, `tests/ac_command_lint.rs`, `tests/card_wait_cli.rs`, `src/card/wait.rs`
- read-only: `tools/check_ac_commands.py`, `tests/*.rs` (the lint's input), `docs/`
- forbidden: `tools/check_ac_commands.py`, `src/` other than `src/card/wait.rs`, `.config/nextest.toml`

## Plan

1. RED: run the two lint mutants (a process started; the cache removed) against the unchanged test: the old test stays green on both, because neither adds seconds on a quiet machine.
2. GREEN: replace the time assertion with a small interpreter harness that sets an audit hook, runs the lint script unchanged, and writes the opened paths and the process-start count to a file; the test asserts REQ-1 to REQ-4 from that file.
3. `card wait`: a `Clock` (system or hand-moved) and `run_wait` (fix the time, run the start read, poll) replace the clock reads in `wait`; unit tests drive it with a fake clock; the process test keeps lower bounds only.
4. Mutants, the 30-run load check, one `moai check affected`, HANDOFF READY.

## Risks

- The harness counts what the interpreter reports through audit events (`open`, `subprocess.Popen`, `os.system`, `os.exec*`, `os.posix_spawn`, `os.spawn*`, `os.fork*`). A child started by a C extension that bypasses those events is not seen; the lint imports `re`, `sys` and `pathlib` only, and the fake `cargo` on `PATH` stays as a second witness for the one process this test was written for.
- A read the lint does through a path outside the three folders (a new scan root) fails REQ-3 until the folder list is widened on purpose.
- The audit event for a file read is counted once per read only when its mode is a string: Python 3.9 reads through an opener and raises a second `open` event (`os.open`, mode `None`) for the same read. The first draft counted both and failed 30 of 30 runs when the test ran under `/usr/bin/python3` 3.9.6; the filter is pinned by the stand-in script that reads one file twice and must show exactly two opens.
- The fake clock tests prove the order of the clock and the start read inside `run_wait`; that `wait` hands its ledger read to `run_wait` is read from the code, and the process test only shows the wired command reporting the time spent at the start.
- The count is exact only for the lint as it is; a legitimate second reading of one file would fail REQ-2 and needs the cache or this bound changed on purpose, never a larger number.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Observe the lint from outside with an interpreter audit hook instead of adding counters to `tools/check_ac_commands.py` | the leader's dispatch names only the test file; a hook leaves the tool and its scope untouched | add a counter line to the tool's summary and assert on it |
| D-2 | Keep the fake `cargo` on `PATH` besides the process count | the SPEC `SPEC-20261122-ac-lint-no-cargo` named it; it states the protected fact in its own words | delete the marker once the process count is trusted |
| D-3 | Widen the reservation to `tests/card_wait_cli.rs` and `src/card/wait.rs` and fix the wait test with an injected clock, not a `polls` field | the leader's card answer of 2026-11-22; option A of the lane's request, which leaves the command's output unchanged | revert the commit; the earlier test returns |
| D-4 | A lane reads the open count of a file once per read, not per audit event | Python 3.9 and 3.14 raise a different number of `open` events for one read | count both events and accept doubled numbers on 3.9 |

## Result

Measured 2026-11-22 on macOS (16 cores) in this branch's worktree (in-tree `target/`, `CARGO_BUILD_JOBS=5`, `RUSTC_WRAPPER=sccache`, nextest, default `~/.cargo` with `--offline --locked` because the shared cargo folders under `/tmp` were gone after a reboot). The slot scripts ran the tests with `/usr/bin/python3` 3.9.6; the first runs from the tool shell used 3.14.2 (`/opt/homebrew/bin`). Both are covered below.

- AC-1 met: `the_repository_lint_starts_no_process_and_reads_each_file_once` passes. Over the repository the lint checks 338 recorded files and 6 `--test` targets, opens 380 files (350 different), starts 0 processes, and no file is opened outside the three folders. The shared test modules `tests/common/*.rs` are read 6 times each, once for every target; every other file once. Passes under Python 3.9.6 and 3.14.2 (`ac_command_lint`: 9 run, 9 passed in both).
- AC-2 met: a lint that runs `subprocess.run(['true'])` first fails the test with `the lint started processes: ["subprocess.Popen"]` (Python 3.9.6).
- AC-3 met: a lint with the per-target cache removed (`if name not in cache` replaced by `if True`) fails the test with the list of files opened more often than allowed, under Python 3.9.6 and under 3.14.2.
- AC-4 met: a counter that records no open (`if event == "open"` replaced by `if False` in the runner) fails with `the counter saw 0 opens for 338 recorded files and 6 targets`. The stand-in test `the_audit_counter_sees_a_process_a_second_read_a_stray_read_and_a_silent_run` pins the same four counts without the repository.
- AC-5 met: `rg -n "Instant|elapsed|Duration|SystemTime" tests/ac_command_lint.rs` prints nothing (exit 1).
- AC-6 met: `card::wait` has 10 unit tests, 10 passed, all on a fake clock with no real sleep. Mutants: the deadline fixed after the start read fails 2 tests (`a_start_read_that_used_up_the_timeout_leaves_one_poll`, `a_start_read_counts_in_the_timeout_and_the_last_poll_is_at_the_deadline`); the clock read after the start read fails the same 2; the sleep not capped at the deadline fails 2 (`the_sleep_between_polls_never_passes_the_deadline` and the last-poll test).
- AC-7 met: `a_ledger_held_at_the_start_counts_in_the_wait` now holds the lock 2.5 s against a timeout of 1 s and asserts only the outcome and `waited_secs >= 2`. The clock-read-after-the-start-read mutant also fails this test (3.8 s run); the deadline-after-the-start-read mutant does not, and that one is the unit tests' to catch.
- Load: 30 of 30 runs passed with 27 tests each (the 9 lint tests, the 8 `card_wait_cli` tests, the 10 `card::wait` unit tests) while 60 `nice 10` processes burned CPU: the 1-minute load average at the start of each run was 94 to 278 (29 of 30 runs at 100 or more), a run took 9.5 to 17.8 s. A first 30-run attempt failed 30 of 30 because the counter counted two `open` events per read under Python 3.9; that value is void and the cause is D-4 above.
- Same class, not changed (no known failure under load): the lines that still compare an elapsed time or a duration with a limit, counted with `rg` over `src` and `tests` on 2026-11-22, are 39 in 23 files: `tests/headless_run_cli.rs` 5, `tests/edit_diagnostics_cli.rs` 4, `tests/card_wait_cli.rs` 4 (limits of 12 s and 20 s against waits of 2 s to 50 s), `tests/product_cli.rs` 3, one or two each in `tests/statusline_counts_refresh.rs`, `run_verify.rs`, `lane_send_cli.rs`, `hook_secret_tokens_cli.rs` and one each in 14 more files including `src/worktree_gc.rs`, `src/heavy_slot.rs` and `src/card/early_reclaim.rs`. The regex finds comparisons written as `elapsed() <`, `took <`, `secs <` and `< limit`, so a bound written another way is not in the count.

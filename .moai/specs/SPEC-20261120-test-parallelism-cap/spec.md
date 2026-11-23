---
id: SPEC-20261120-test-parallelism-cap
title: "Test-internal parallelism gets a cap: heavy-slot and check affected bound the test threads, spawning tests take a permit"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

heavy-slot limits how many heavy commands run, not how many threads each command runs: with
`--max 5` and `RUST_TEST_THREADS` defaulting to the core count, five concurrent `cargo test`
runs are up to ~80 test threads, and tests that spawn the moai binary in temporary HOMEs
multiply processes and file events. The 2026-11-20T20:58Z read-only sample the card records
(load 59–72, swap 37.35/37.89 GiB, fseventsd 110% CPU / 14.6 GB, 41 moai-cli + 20 moai
processes; `home_retire_cli` spawned 11 `moai` stand-ins, `update_home_layer_cli` 10,
`update_keys_cli` 9 moai-cli) is that overload, and this card caps the parallelism inside
each run. Every user's machine gets the same behaviour — `moai heavy-slot-acquire` and
`moai check affected` are shipped commands — so the cap is a default (4), configurable per
machine in the slots' `limits.json`, and printed where the run is reported.

Survey (tip `bcfb1bc5b7`): the wrapper's exec site is `src/heavy_slot.rs` `run_holding`
(`Command::new(&command[0])`, the place an env injection belongs; the machine limits live in
`<slot folder>/limits.json` with keys `max`, `min_free_gb`, `max_swap_pct`,
`min_mem_free_pct` — a `test_threads` key joins them); `moai check affected` builds its
steps in `src/check_affected.rs` `steps()` where `Step` already carries an `env` field
(the `python` step sets `CARGO_TARGET_DIR` and `MOAI_CLI_BIN` there) and its test steps are
the three `cargo test` steps `test_lib`, `test_bins`, `test_targets`; the tests that spawn
the binary share `tests/common/mod.rs`; `cargo`'s libtest reads `RUST_TEST_THREADS`, and
`cargo nextest run` takes `-j`. macOS skips Spotlight indexing for a folder that holds a
`.metadata_never_index` marker.

## Non-goals

- Not changing how many heavy commands may run at once (`max`) — that concurrency change is
  another card's in-flight work on the same file; this card only bounds the threads and the
  spawning tests inside one run.
- Not rewriting the spawning tests' logic: the permit wraps the spawn point, the tests keep
  their bodies.
- No per-test serialisation beyond the spawning tests: non-spawning tests keep running at
  the runner's own parallelism.
- `tools/heavy_check_slot.py` and `tools/self_check.sh`: the Python slot tool already
  embeds the Rust side; no tool change in this card.

## Requirements

- REQ-1: While `moai check affected` runs its `test` step (one call, `cargo nextest run` or `cargo test` when nextest is missing), it shall pass a test-thread cap to that step — `-j <n>` after `run` for nextest, `RUST_TEST_THREADS=<n>` in the step env for libtest — unless the caller's own environment or arguments already set that knob, use 4 for `<n>` until the slots' `limits.json` can carry a `test_threads` key (the slice that adds that field to `heavy_slot::LimitsFile` also makes `check affected` read it), and carry `test-threads=<n>` in the answer's summary line, while the other cargo steps (`fmt`, `clippy`, the CLI build) get no injected cap. · failure: the cap is only injected; none — pinned by AC-1.
- REQ-2: While `moai heavy-slot-acquire` wraps `cargo test` (libtest) or `cargo nextest run`, the wrapper shall pass the same cap the same way — `RUST_TEST_THREADS` in the child env for libtest, `-j <n>` appended for nextest — unless the caller has set the value itself, and its `slot taken` line names `test-threads=<n>` from the same `limits.json`. · failure: none — pinned by AC-2.
- REQ-3: Tests that spawn the moai binary shall take one permit from a shared in-process semaphore (a new helper file under `tests/common/`) before each spawn and release it on drop, so one test binary run by libtest holds at most `MOAI_SPAWN_TEST_CAP` spawning tests at once (default 4), while a test that never spawns is unaffected; under `cargo nextest`, which runs every test as a process of its own, the same bound is the `spawning-moai` test group of `.config/nextest.toml` (`max-threads = 4`) over the test files that take permits. · failure: none — pinned by AC-3.
- REQ-4: When the lane build folder under the home's `.moai/build/` folder and a test's temporary HOME are created, the creator shall also create a `.metadata_never_index` marker file in the new folder as a best-effort step (a failure to create the marker never fails the folder), so macOS Spotlight stops indexing build and test output. · failure: none — pinned by AC-4.
- REQ-5: The before/after measurement of this card shall run the moai-adk probe harness on a representative check-affected run and report load 1, swap, the peak process count and fs-event count from before and after next to the check evidence in this SPEC's Result. · failure: the machine's load window is too busy — the pair is reported as not run with the observed load, never as passed.
- REQ-6 (REQ-1): IF the `test_threads` value in `limits.json` is 0 or unparsable, THEN `moai check affected` shall refuse to run any step and name the bad value. This holds from the slice that lets `heavy_slot::LimitsFile` carry the key; before it, `check affected` neither reads nor names the key, because a key the slot reader refuses would stop every heavy run. · failure: none — pinned by the AC-1 text of that slice.

## Acceptance

- AC-1 (REQ-1): `steps()` with a cap puts the cap only on the `test` step (`-j <n>` for nextest, `RUST_TEST_THREADS` for libtest), a caller-set value is kept, `moai check affected` on a real range prints `test-threads=4` (or the caller's own knob value) in its summary line and the test steps' logs show the cap took effect.
  verify: `cargo test --offline --locked --lib check_affected` plus one `moai check affected` run naming the summary line.
- AC-2 (REQ-2): with the wrapper wrapping `cargo test --offline --locked --lib`, the test run's thread count is the cap (a probe test asserts the env), wrapping `cargo nextest run` gains `-j <n>`, a caller-set `RUST_TEST_THREADS=8` stays 8, and the `slot taken` line carries `test-threads=<n>`.
  verify: `cargo test --offline --locked --lib heavy_slot` (the wrapper tests) plus one wrapper run printing the line.
- AC-3 (REQ-3): the semaphore test drives N+1 threads through the permit helper and observes at most N concurrent holders; a second permit taken by a thread that holds one panics instead of waiting; the cap text is a whole number of at least 1, else 4; the nextest group's filter names exactly the test files that take permits (listed by name in the test, which does not walk `tests/`); every spawning test that takes a permit keeps passing, also with `MOAI_SPAWN_TEST_CAP=1`.
  verify: `cargo test --offline --locked --test spawn_cap_test` and the spawning tests' own targets (`--test home_retire_cli --test update_home_layer_cli --test update_keys_cli`, once with `MOAI_SPAWN_TEST_CAP=1`).
- AC-4 (REQ-4): a lane-env folder and a temporary HOME created by the helpers contain `.metadata_never_index`; the tests that create them assert the marker.
  verify: `cargo test --offline --locked --lib lane_env` and the temp-HOME helper tests.
- AC-5 (REQ-5): the measurement pair's rows (before/after × load 1, swap, peak processes, fs events) are in this SPEC's Result or an explicitly not-run note with the load that blocked it.
  verify: `moai spec check SPEC-20261120-test-parallelism-cap`.

## Scope

- write (this commit): this SPEC folder. Later slices widen to the files each slice touches:
  `src/check_affected.rs` and `src/ops/check.rs` (REQ-1), `src/heavy_slot.rs` (REQ-2),
  `tests/common/` plus the spawning tests (REQ-3), the folder creators (REQ-4) and this
  SPEC's Result (REQ-5).
- reservations this card met while taking: `src/heavy_slot.rs` is held by the in-flight
  heavy-builds card (lease valid, its last commit is ~9 h old); `src/check_affected.rs`,
  `src/ops/check.rs` and `tests/common/mod.rs` sit under leases that expired (the holders'
  latest states: one card's work merged ~40 min ago, the other's last commit ~5 h old).
  The slices wait for the leader to clear those reservations or name an order — no take
  over another holder's scope.
- read-only: `tools/repo_scans.json`, the CI workflow files, the earlier overload investigation's probe harness.
- forbidden: overwriting a caller-set `RUST_TEST_THREADS`, card numbers in SPEC ids,
  branches or commits, deleting a spawning test instead of permitting it.

## Plan

1. This commit: the SPEC (slice 1).
2. Slice 2 (REQ-1, needs the expired-lease files back): `steps()` gains the cap parameter,
   `run` reads `test_threads` from the slots' limits, the summary line grows
   `test-threads=<n>`, ops help and the rule tests follow, tests first (RED → GREEN).
3. Slice 3 (REQ-2, needs the in-flight heavy-builds card to return `src/heavy_slot.rs`):
   the wrapper injects env / `-j`, the `slot taken` line grows the word, wrapper tests.
4. Slice 4 (REQ-3, needs `tests/common/mod.rs` back): the new permit helper plus its test,
   then one spawning test at a time takes the permit.
5. Slice 5 (REQ-4): the folder creators make the marker, helper tests.
6. Slice 6 (REQ-5): the before/after pair on a quiet window, rows into Result.

## Risks

- The heavy-builds card is changing `src/heavy_slot.rs`'s `max` default in parallel: slice 3
  merges into a file that card holds, so the slice goes strictly after that card returns and
  re-reads its head — no concurrent edits.
- Nextest may ignore `RUST_TEST_THREADS` (its own `-j` flag is the documented knob): slice 3
  verifies against the installed nextest's help before wiring, and falls back to appending
  `-j` — the SPEC's REQ-2 already names that split.
- The permit helper only bounds the tests that take it: a spawning test left out silently
  weakens the cap, so slice 4 greps the spawn sites under `tests/` and lists every one it
  wrapped, and the audit checks that list against `spawn` sites.
- Injecting env into cargo test steps changes runs users script around — the caller-set
  value is always kept (REQ-1, REQ-2), which is the escape hatch.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | Slices, in dependency order, waiting for the reservations instead of taking them over | `heavy_slot.rs`, `check_affected.rs`, `ops/check.rs` and `tests/common/mod.rs` are other holders' scopes; two leases are expired and need the leader | Take the whole set once the leader clears it |
| D2 | The cap travels as `RUST_TEST_THREADS` env (libtest) and `-j` (nextest), never by rewriting the command's own arguments | Env is additive and caller-overridable; rewriting args changes what the user asked to run | A flag on the commands if a runner rejects the env |
| D3 | The spawning cap is in-process under libtest (one permit set per test binary) and a nextest test group under nextest, not a file lock across binaries | The card's bound is "one test binary never runs more than N spawning tests at once". Under nextest each test is a process, so the in-process set bounds nothing there (a pid probe showed one set per test); `check affected` and CI use nextest, so the group is the bound that counts, and a group is configuration nextest itself enforces. A file lock would add contention for no named requirement | A file-backed semaphore if a bound across runners or binaries is asked for |
| D4 | The marker file is best-effort, never a failure | A folder that exists is still correct for its purpose on non-macOS hosts | — |

## Result

- AC-3: met (REQ-3) — `tests/common/spawn_cap.rs` is a shared in-process semaphore (default 4, `MOAI_SPAWN_TEST_CAP`); `tests/spawn_cap_test.rs` drives it and the three spawning files `home_retire_cli`, `update_home_layer_cli` and `update_keys_cli` take a permit around each spawn. `cargo nextest run --test spawn_cap_test --test home_retire_cli --test update_home_layer_cli --test update_keys_cli`: all passed (one nextest run of 1767 tests on this head's tree, 1766 passed and the one failure fixed below, then 21 of 21 on the rerun).
- AC-4: met (REQ-4) — the lane build folder and `TempHome::new` write `.metadata_never_index` (best effort); `tests/never_index_test.rs` passes. The marker is the helper's own file, so `TempHome::names_in` leaves it out; before that fix `versioning_cli::the_five_highest_tags_name_the_scheme_and_nothing_is_written` failed with `detection wrote into HOME: [".metadata_never_index"]`.
- Repair of the first audit's blocking findings (slice 1 head `91b8bc5806`, verdict FAIL): (B1) the build folder's marker is created exclusively (`O_CREAT | O_EXCL`), never written through a link and never truncating a file already there — `lane_env::tests::the_marker_never_follows_a_link_or_truncates`, and the real `moai-cli lane-env` run in a throwaway HOME with the marker path a link to a 15-byte file (15 bytes after), a dangling link (target not created), a 4-byte marker file (4 bytes after) and no marker (created); the temporary HOME's marker uses the same exclusive creation. (B2) `update_keys_cli` holds its permit only around the `key-set` spawn, the two spawns without a permit take one, and the helper panics on a second `acquire` from a thread that holds one (`a_nested_acquire_on_one_thread_panics_instead_of_waiting`), so no test can wait for itself; `update_keys_cli`, `update_home_layer_cli` and `home_retire_cli` pass with `MOAI_SPAWN_TEST_CAP=1` under nextest (44 of 44) and `update_keys_cli` under libtest (9 of 9). (B3) the in-process set bounds nothing under nextest (one process per test), so `.config/nextest.toml` has the `spawning-moai` test group (`max-threads = 4`) over the three files that take permits, and `the_files_that_take_permits_are_exactly_the_nextest_group` fails when the filter and the listed files differ (a new file that takes permits is added to both; the test does not walk `tests/`, which the repository scan table reserves for listed scans); probe: the same 44 tests with `nextest -j 16` ran at most 16 spawning test processes at once without the group and 4 with it (processes counted every 0.2 s). (B4) the cap text is read by `parse_cap` — a whole number of at least 1, else 4 — with a table test over 17 inputs. D3 and REQ-3 now say what each runner does.
- AC-1: met for `check affected` (REQ-1), slice 2 — `steps()` takes the cap and puts it on the `test` step alone (`-j <n>` right after `run` for nextest, `RUST_TEST_THREADS=<n>` in the step env for libtest); `run` uses 4, keeps a caller-set `NEXTEST_TEST_THREADS` / `RUST_TEST_THREADS` and shows that value. The summary line is `check-affected base=<12> head=<12> targets=<n> scans=<w> test-threads=<n> <step>=<exit>… result=…` (the `py=` word is gone since the Python steps left `check affected`). `tests/check_affected_one_slot.rs` pins the new key and `-j 4`, and both its commands and the CLI tests' remove `RUST_TEST_THREADS` and `NEXTEST_TEST_THREADS` first: a machine that exports them otherwise hides the `-j`. Tests: lib `check_affected::tests::the_thread_cap_goes_onto_the_test_step_alone`, `the_default_cap_is_four`, `a_caller_set_knob_is_seen_for_its_own_runner_only`; CLI `the_summary_names_the_thread_cap_and_only_the_test_step_gets_it`, `a_thread_knob_the_caller_set_is_kept`.
- REQ-6 waits for the slice that adds `test_threads` to `heavy_slot::LimitsFile` (`src/heavy_slot.rs`, held by another card). The second audit (head `3b3390f681`) found that the earlier `check affected` help told users to put `test_threads` into `limits.json` while `LimitsFile` refuses unknown fields, so such a file made every heavy-slot run exit 125. This repair takes the other way out: `check affected` no longer reads the key, its help no longer names it, and the refusal of a bad value moves with the key. Until then the cap is 4 or the caller's own knob.
- Not run, waiting: AC-2 (REQ-2) — slice 3. The `heavy-slot-acquire` wrapper cap and its `slot taken` line go into `src/heavy_slot.rs`, which another card holds; the first version of that code stays in this branch's history.
- Not run, waiting: AC-5 (REQ-5). The machine's load 1 was 70 to 125 while this card ran, so a before/after pair would measure the other lanes; it needs a quiet window.

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

- REQ-1: While `moai check affected` runs its cargo test steps (`test_lib`, `test_bins`, `test_targets`), it shall pass a test-thread cap to every one of them through the step env (`RUST_TEST_THREADS=<n>`) unless the caller's own environment already sets that variable, take `<n>` from the slots' `limits.json` `test_threads` key with default 4, and carry `test-threads=<n>` in the answer's summary line, while the non-test cargo steps (`fmt`, `clippy`, `build_cli`) get no injected env. · failure: the cap is only injected; a value outside this SPEC's rules is the refusal REQ-6 names.
- REQ-2: While `moai heavy-slot-acquire` wraps `cargo test` (libtest) or `cargo nextest run`, the wrapper shall pass the same cap the same way — `RUST_TEST_THREADS` in the child env for libtest, `-j <n>` appended for nextest — unless the caller has set the value itself, and its `slot taken` line names `test-threads=<n>` from the same `limits.json`. · failure: none — pinned by AC-2.
- REQ-3: Tests that spawn the moai binary shall take one permit from a shared in-process semaphore (a new helper file under `tests/common/`) before each spawn and release it on drop, so one test binary holds at most `MOAI_SPAWN_TEST_CAP` spawning tests at once (default 4), while a test that never spawns is unaffected. · failure: none — pinned by AC-3.
- REQ-4: When the lane build folder under the home's `.moai/build/` folder and a test's temporary HOME are created, the creator shall also create a `.metadata_never_index` marker file in the new folder as a best-effort step (a failure to create the marker never fails the folder), so macOS Spotlight stops indexing build and test output. · failure: none — pinned by AC-4.
- REQ-5: The before/after measurement of this card shall run the moai-adk probe harness on a representative check-affected run and report load 1, swap, the peak process count and fs-event count from before and after next to the check evidence in this SPEC's Result. · failure: the machine's load window is too busy — the pair is reported as not run with the observed load, never as passed.
- REQ-6 (REQ-1): IF the `test_threads` value in `limits.json` is 0 or unparsable, THEN `moai check affected` shall refuse to run any step and name the bad value. · failure: none — pinned by AC-1.

## Acceptance

- AC-1 (REQ-1): `steps()` with a cap fills `RUST_TEST_THREADS` only on the three test steps, a caller-set value is kept, `test_threads: 0` is refused by `run`; `moai check affected` on a real range prints `test-threads=4` (or the limits.json value) in its summary line and the test steps' logs show the cap took effect.
  verify: `cargo test --offline --locked --lib check_affected` plus one `moai check affected` run naming the summary line.
- AC-2 (REQ-2): with the wrapper wrapping `cargo test --offline --locked --lib`, the test run's thread count is the cap (a probe test asserts the env), wrapping `cargo nextest run` gains `-j <n>`, a caller-set `RUST_TEST_THREADS=8` stays 8, and the `slot taken` line carries `test-threads=<n>`.
  verify: `cargo test --offline --locked --lib heavy_slot` (the wrapper tests) plus one wrapper run printing the line.
- AC-3 (REQ-3): the semaphore test drives N+1 threads through the permit helper and observes at most N concurrent holders; every spawning test that takes a permit keeps passing.
  verify: `cargo test --offline --locked --lib spawn_cap` and the spawning tests' own targets.
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
- read-only: `tools/repo_scans.json`, the CI workflow files, C-211's probe harness.
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
| D3 | The spawning cap is in-process (one permit set per test binary), not a file lock across binaries | The card's bound is "one test binary never runs more than N spawning tests at once"; cross-binary locks would add contention for no named requirement | A file-backed semaphore if a cross-binary bound is asked for |
| D4 | The marker file is best-effort, never a failure | A folder that exists is still correct for its purpose on non-macOS hosts | — |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.

---
id: SPEC-20261123-test-area-crates
title: "The integration tests run from eight area crates, one area moved per slice, with test names kept as module paths and check affected building only area crates"
status: draft
revision: 2
created: 2026-11-23
parent: null
branch: spec/20261123-test-area-crates-card
---

# Integration tests in area crates

## Intent

Every `tests/*.rs` file is its own test binary, 435 of them at tip `4b173d8849`. Each binary links the library again and compiles `tests/common` again (359 files declare `mod common;`).

A one-file change in `src/` makes the receipt check do two expensive things:
- `cargo clippy --all-targets` type-checks all 435 test crates;
- the test step builds and links 43 to 70 of them.

A change under `tests/common/` selects all 359 that declare it. The leader measured two such runs standing at 737 and 745 arguments for 23 minutes while seven lanes waited.

This SPEC merges the integration tests into eight area crates (`tests/<area>/main.rs`, one module per old file). A one-file change then builds eight test binaries, and clippy checks eight test crates. Test function names are kept; the old target name becomes the module path. The move is one area per slice, in the order that breaks the fewest open branches.

This SPEC plans the work only. No file moves before build 19 is made (leader decision 6961).

### Measurements (tip `4b173d8849`, 2026-11-22T20:0xZ–20:11Z, read only, nothing built)

| item | value | command |
| --- | --- | --- |
| integration test targets | 435 | `ls tests/*.rs \| wc -l` |
| `#[test]` functions in them | 5,029 (237,791 lines) | `grep -c '#\[test\]' tests/*.rs` summed |
| files declaring `mod common;` | 359; other shared modules: `go_project_real_run` 5, `prop` 3, `renamed`, `link`, `helpers`, `git_hooks` 1 each | `grep -h -o '^mod [a-z_]*;' tests/*.rs \| sort \| uniq -c` |
| folders under `tests/` | `common`, `fixtures`, `e2e` | `find tests -mindepth 1 -maxdepth 1 -type d` |
| files that call `set_var`, `remove_var` or `set_current_dir` | 3: `commit_identity_guard_2878`, `update_switch_layer`, `worktree_place_cli` | `grep -l` |
| files with a process-global static (`OnceLock`, `Mutex`, `thread_local!` …) | 9 | `grep -l` |
| files naming `CLAUDE_CODE_SESSION_ID` | 29 | `grep -l` |
| `check affected --list` on 14 one-`src/`-file commits | 43–70 targets, always 42 listed scans, 8–9 Python modules; steps `clippy --all-targets`, `test --lib`, `test --bins`, `test --test …` | `moai check affected --list <c>~1..<c>` |
| areas those selections touch, under the split below | all 8 for each of the four sampled commits (`src/profile.rs`, `src/main.rs`, `src/card/remote_copy.rs`, `src/ui_text.rs`); the always-on scan list alone spans every area | the same lists mapped by the area table |
| `--test <name>` occurrences | `tools/repo_scans.json` 38; SPEC files 2,306 in 272 SPECs; docs, skills, agents, tools and src 74 | `grep -o` |
| `src/` files naming `tests/<name>.rs` | 13 | `git grep -l` |
| `tests/*.rs` files changed on the 40 newest unmerged `spec/` branches | 103 (project 29, update 15, home 14, spec 11, lane 11, screen 9, card 8, repo 2, not on tip 4) | `git diff --name-only <merge-base> <branch> -- 'tests/*.rs'` per branch |

### Sources (read whole)

- `~/.moai/reports/moai-adk/test-load-redesign-2026-11-21.md`:
  - §3.1 and §3.5: the counts at that time, 379 targets and 324 `mod common`.
  - §5 P6 and §6 S9: area crates; `autotests`; risk "충돌 폭탄" (a conflict bomb); "one area per slice".
  - §9.3, matklad "Delete Cargo Integration Tests": "compile … decreased 3x", "artifacts decreased 5x".
  - §9.3, nextest "process-per-test": isolation is kept when crates merge.
- Leader decision 6961 (the ledger card's note): the four questions this SPEC answers, and no code before build 19.
- Cargo target auto-discovery: a folder `tests/<name>/main.rs` is a test target named `<name>`, and `tests/<name>/<other>.rs` is not a target of its own. This is the Cargo Book, "Target auto-discovery"; it was not re-read in this lane, which has no network. U-1 makes it a check before slice 1.

## Non-goals

- No change to the library, the binaries or what a test asserts. Bodies move; assertions do not change.
- No rewrite of closed SPECs. Their recorded `--test <old>` commands are read through the area table (REQ-6).
- No change to the Python tests or to the unit tests in `src/`.
- No sccache or CoW change; the build-folder cards own those.

## Requirements

- REQ-1: The test tree shall hold one test crate per area at `tests/<area>/main.rs` for the eight areas of the area table.
- REQ-2: Each moved file shall become the module `<old target name>` of its area crate with every test function name unchanged.
- REQ-3: The area table shall be one tracked file `tests/areas.toml` that maps each first word of a test file name to its area.
- REQ-4: A scan test shall fail when a `tests/*.rs` file or an area module has a first word the area table does not map, or when a top-level `tests/<name>.rs` belongs to an area that has moved.
- REQ-5: When a change selects an old target that has moved, check affected shall build its area crate once and run the selected modules through a name filter.
- REQ-6: When a recorded AC command names `--test <old target>` of a moved area, the AC command lint shall resolve it through the area table.
- REQ-7: Each target list shall name the area crate and a module filter for every moved target in the slice that moves it (the lists are `tools/repo_scans.json`, `.config/nextest.toml` and `tools/ci_scope.py`).
- REQ-8: A test file that calls `set_var` or `remove_var` or `set_current_dir` shall stay a test target of its own.
- REQ-9: Each slice shall move exactly one area and land alone.
- REQ-10: Each slice shall record the cold `cargo test --no-run` CPU seconds, wall time, build-folder size and file count before and after, and the test count from `cargo nextest list` equal before and after.
- REQ-11 (REQ-5): If an area crate fails to build, then check affected shall report every selected module of that area as failed.
- REQ-12 (REQ-6): If a recorded `--test <name>` names neither a test target nor a mapped module, then the AC command lint shall fail on that line.
- REQ-13 (REQ-7): A scan command that names several modules of one area crate shall put the first filter before `--` and the others after it, as cargo takes one test name before `--`.
- REQ-14 (REQ-4): Each scan that reads the test sources shall read the module files of an area crate.
- REQ-15 (REQ-5): The selection of a moved module by its old target name shall add the area crate once as the `--test` argument and, under nextest, one filter set that keeps the solo targets whole and narrows the area crate to the picked modules.
- REQ-16 (REQ-14): The exemption and baseline tables of those scans shall name a moved file by its `tests/<area>/<module>.rs` path.
- REQ-17 (REQ-3): The home guard baseline shall name a moved test by the area crate as its binary and `<module>::<fn>` as its name.
- REQ-18 (REQ-2): A test that starts its own test binary again shall name the child test by its module path when the file is a module of an area crate.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): after a slice, `tests/<area>/main.rs` declares one module per moved file, `cargo nextest list --test <area>` lists every moved test as `<old target>::<fn>`, and the total test count equals the count before the slice.
  verify: `cargo test --locked --offline --test test_area_layout`
- AC-2 (REQ-3, REQ-4): every first word of every test file and area module is in `tests/areas.toml`; a planted `tests/<moved area word>_x.rs` and a planted unmapped word each fail the scan.
  verify: `cargo test --locked --offline --test test_area_layout area_table_`
- AC-3 (REQ-5, REQ-11): for a one-`src/`-file range, `check affected --list` names at most the eight area crates plus the solo targets, and each test step carries a filter of the selected modules; a planted build failure of one area crate marks all its selected modules failed.
  verify: `cargo test --locked --offline --lib check_affected::tests::`
- AC-4 (REQ-6, REQ-12): a recorded AC line `cargo test --test <moved target> <filter>` passes the AC command lint, which resolves it to `--test <area> <moved target>::<filter>`; a line naming an unknown target fails it.
  verify: `cargo test --locked --offline --test ac_command_lint area_`
- AC-5 (REQ-7): no `--test <name>` in `tools/repo_scans.json`, `.config/nextest.toml` or `tools/ci_scope.py` names a target that is no longer a file.
  verify: `cargo test --locked --offline --test test_area_layout scan_list_`
- AC-6 (REQ-8): the three files that change process-global state stay top-level targets, and a scan fails when an area module calls `set_var`, `remove_var` or `set_current_dir`.
  verify: `cargo test --locked --offline --test test_area_layout solo_`
- AC-7 (REQ-10): the slice's READY quotes the before-and-after numbers of a cold `cargo test --no-run` in a throwaway worktree, and the equal nextest counts.
  verify: `cargo nextest list --locked --offline --workspace`

### The area table (first word of the file name → area; measured counts at `4b173d8849`)

| area | files | tests | first words |
| --- | --- | --- | --- |
| card | 81 | 1,059 | card claim ledger queue dispatch startable board journal handoff recurring close goal decision trailer commit forge labels github gitlab import release |
| update | 58 | 768 | update install migrate migration generation versioning switch models model retired removed self freshness baseline go |
| home | 48 | 524 | home hygiene worktree heavy process path exe binary bin archive recovery revert state apply |
| project | 66 | 777 | project profile language placeholder gitignore skeleton agents agent claude codex omp orca host hook hooks script secret key scope |
| lane | 50 | 601 | lane leader session launcher headless run work flow next message context usage pty stdin mode question |
| screen | 33 | 257 | statusline status ui style banner output help cli error html report reports troubleshooting struct |
| spec | 65 | 767 | spec sdd audit review gate check ac test flaky green parity boundary protected rule core bundle flake skills skill workflow product ontology public code doc |
| repo | 34 | 276 | atlas search index prop plan optimize diff edit e2e windows cargo ops hot no empty generic git repo layout base sync |

Total: 435 files, 5,029 tests, no file unmapped. The rule is the file name's first word, because it is the one fact every file already has. The words are grouped by the `src/` area the tests drive.

### Expected build count for a one-file `src/` change

| step | today | after all slices |
| --- | --- | --- |
| clippy `--all-targets` | lib, 2 bins, 435 test crates | lib, 2 bins, 8 area crates and 3 solo crates |
| test step build | 43–70 test binaries, each linking the library | the selected areas' crates: 8 for every sampled commit (the scan list spans all areas) plus any selected solo target |
| test step run | the selected targets | the same tests, through module filters; run time about equal |
| `tests/common` compiled | once per declaring target (359) | once per area crate (8) and solo crate |

The run time stays about the same because the same tests run. The saving is in compiling, linking and the build folder; the report quotes "compile … 3x" and "artifacts … 5x" for one project. Each slice measures its own numbers (REQ-10).

### Name filters and isolation

- Cargo: `cargo test --test <old> <f>` becomes `cargo test --test <area> <old>::<f>`, and a whole old target becomes `cargo test --test <area> <old>::`.
- Nextest: `binary(<old>)` becomes `binary(<area>) & test(/^<old>::/)`.
- Known limit of the cargo runner (found in slice 1): libtest filters are substrings, not prefixes. `cargo test --test card claim_cli::` also runs `card_reclaim_cli::…`, and `dispatch_cli::` also runs `card_dispatch_cli::…`. The cargo runner over-selects there; the nextest runner is exact, because the filter set anchors `^<module>::`. `check affected` runs nextest when it is installed, and the READY receipt names the runner.
- Selection stays by old target name (slice 1, D-6): `select_pick` is unchanged and the moved modules are read as `TestFile` entries of their area, so the name logic of SPEC-20261124-card-select-batch-land keeps working. Only `steps()` turns the picked old names into the one area `--test` and the filter set.
- HOME: every child gets its temporary HOME through `tests/common` (`Command::env("HOME", …)`, `env_clear` for git). That is per child process, not process-global, so it holds in one process.
- Session id: the 29 files that name `CLAUDE_CODE_SESSION_ID` set or remove it on the child `Command`. Any that call `std::env::set_var` instead fall under REQ-8.
- `cargo test` runs one area's tests as threads of one process; nextest keeps one process per test. The three solo files and the nine files with a process-global static are the only shared-state surface. The statics are read once and stay correct in one process; slice 0 checks each (U-2).

## Scope

- SPEC slice (now): `.moai/specs/SPEC-20261123-test-area-crates/**`.
- Slice 0 (after build 19):
  - `tests/areas.toml`;
  - `tests/test_area_layout.rs`;
  - `src/check_affected.rs` (module filter and area build);
  - `tools/check_ac_commands.py` or its Rust port (AC lint resolution);
  - `docs/design-review/appendix/operations/check.md`.
- Slices 1 to 8, one area each:
  - `tests/<area>/**` and the moved `tests/*.rs` of that area;
  - `tools/repo_scans.json`, `.config/nextest.toml` and `tools/ci_scope.py` (the moved names only);
  - the docs, skills and agents lines that name a moved `--test`.

## Plan

0. Slice 0: the table, the layout scan (it passes with nothing moved yet), check affected's area build with module filters, and the AC lint resolution. No file moves.
1. One slice per area, in this order. The order is by the measured count of open-branch changes to the area's files, fewest first: repo (2), card (8), screen (9), lane (11), spec (11), home (14), update (15), project (29).

   Each slice does the following:
   - re-measure the open branches;
   - `git mv` the area's files into `tests/<area>/`;
   - write `main.rs`;
   - turn each file's `mod common;` into a use of the crate's one `common` module;
   - update the scan list, nextest file and ci_scope names;
   - run the area crate and its readers;
   - measure (REQ-10);
   - READY.
2. A lane whose open branch changes a moved file merges the tip once and moves its edit into the module (the scan of REQ-4 points it there). The lander states in each MERGED note which area moved.

## Risks

- R-1: Conflicts with open branches. A moved file that a branch edits becomes a rename conflict. The order and one area per slice keep each wave small; project, with 29 open edits, goes last.
- R-2: One process for an area under `cargo test`. Shared state across threads could make a test flaky. REQ-8 and the nine statics check guard it; nextest keeps process-per-test.
- R-3: The always-on scan list spans every area, so every check builds all eight crates. Each is bigger than one old binary. If one area crate's link time exceeds the old per-target sum, the slice's measurement shows it, and the area splits in two (the table allows up to 12).
- R-4: `ac_command_lint` builds a target during a test when a recorded command names one outside the selection. With area crates it builds a whole area. Slice 0 resolves names through the table without building.

## Decisions

| id | decision | reason | rejected |
| --- | --- | --- | --- |
| D-1 | Eight areas by the first word of the file name, in one tracked table. | Every file has the word; 435 files map with none left over; areas hold 33–81 files. | One crate `tests/it` (a one-line change relinks everything); split by src module (needs a module map that cargo's dep-info cannot give). |
| D-2 | Keep test function names; the old target name becomes the module path. | Filters translate mechanically (`<old>::`), and history and verdicts stay readable. | Renaming tests. |
| D-3 | Old SPECs keep `--test <old>`; the lint resolves them through the table. | 2,306 lines in 272 SPECs; closed SPECs are records. | Rewriting closed SPECs. |
| D-4 | Use auto-discovery of `tests/<area>/main.rs`, with no `autotests = false`. | No `Cargo.toml` edit: that file is hot, and editing it rebuilds every folder. Unmoved files stay targets during the migration. | `[[test]]` entries for every target. |
| D-5 | Order: fewest open-branch edits first. | Measured per area; it limits each wave of conflicts. | Biggest saving first. |
| D-6 | Slice 1 moves the card area first, ahead of the order of D-5 (leader dispatch, note 8564): the area that changes most often is where the saving counts. One area per slice still holds; the open branch that holds `tests/card_heartbeat_cli.rs` (C-117) is not moved yet. | The leader's priority; the conflict is one file, left on top. | Waiting for the fewest-edits order. |
| D-7 | `card_cleanup_outbox_cli` stays a top-level target (REQ-8: it calls `set_var`). | A process-global call must not share one process with other modules. | Moving it and rewriting its test. |
| D-8 | `tests/card_heartbeat_cli.rs` is a new file of the open card C-117 and is not on the integration tip, so slice 1 does not move it and does not write it. The layout scan (REQ-4) refuses a top-level `tests/card_*.rs` once the card area has moved, so C-117 adds it as `tests/card/card_heartbeat_cli.rs` with one `mod` line in `tests/card/main.rs` when it merges the tip. | C-117 holds the path; the scan points its author to the module. | Waiting for C-117 to land before moving the area. |

### Unsettled

- U-1: Confirm Cargo's auto-discovery of `tests/<dir>/main.rs` on rustc 1.94.1 by building one empty area crate in slice 0, before any file moves.
- U-2: Check the nine files with a process-global static one by one. Each either reads a constant once or needs REQ-8.
- U-3: The real saving per area. The report's 3x and 5x come from another project.

## Result

Slice 1, the card area (2026-11-25, lane agent-3, base `b0bffbbd48`; the tested head is the one named in the READY note of the card). Status: slice 1 is written and waits for the independent audit. Slices 0 and 0b stay as they were (the table, the scan of it, the readers of an area folder).

What slice 1 holds:

- 100 test files of the card area are modules of `tests/card/main.rs` (`git mv`, one `mod` line each, sorted), each `mod common;` is `use crate::common;` and the crate declares `common` once (`#[path = "../common/mod.rs"]`). Six `use crate::common;` lines that nothing used were dropped.
- `card_cleanup_outbox_cli` calls `set_var`, so it stays a top-level target (REQ-8, D-7); `tests/areas.toml` lists it in `solo`. `tests/card_heartbeat_cli.rs` of the open card C-117 is not on the tip and is not moved (D-8).
- `src/check_affected.rs` (name resolution only; `select_pick` is unchanged): a moved module is read as a `TestFile` of its area and is still selected by its old target name, `steps()` puts the area crate on the command line once and, under nextest, adds one filter set (`kind(lib) | kind(bin) | binary(=<solo>) | (binary(=card) & (test(/^<module>::/) | …))`), and the scan list text is folded back to old names. REQ-5 and REQ-15.
- `tools/repo_scans.json` names a moved module as `--test card <module>::`, several modules of one crate as `--test card <first>:: -- <others>::` (REQ-13); `fold_area_checks` reads both forms.
- The scans that read the test sources (`repo_scans_table`, `home_isolation_scan`, the exemption and baseline tables, `state_folder_literals/allow.tsv`) read the module files of an area crate and name a moved file by `tests/card/<module>.rs` (REQ-14).

Measurements, one run each, on a clean `git archive` export, cold `cargo test --locked --offline --no-run` with `CARGO_INCREMENTAL=0`, no wrapper, inside one heavy slot, while other lanes ran (load average 18.7 before, 22.3 after). Scripts `baseline.sh` and `after.sh` in the lane's scratchpad:

| item | before `b0bffbbd48` | after (the card move, before the repairs) | change |
| --- | --- | --- | --- |
| wall time | 161.49 s | 145.85 s | −15.64 s (−9.7%) |
| CPU time (user + sys) | 566.98 + 87.79 = 654.77 s | 507.30 + 66.15 = 573.45 s | −81.32 s (−12.4%) |
| build folder `target/` | 2,464,028 KB | 2,139,976 KB | −324,052 KB (−13.1%) |
| files under `target/` | 5,566 | 4,631 | −935 (−16.8%) |
| test executables | 520 | 421 | −99 (100 modules became one crate) |
| peak memory | 3,029,336,064 B | 3,018,194,944 B | about equal |
| `cargo nextest list` lines | 8,171 | 8,179 | +8 |

The after run was taken on the card move before the repairs (a revision that the later squash into one code commit replaced); the repairs that the whole-area run and the receipts found (below) change no build step. The card area is 1,330 of the repository's tests, so the saving is one slice's share; the cargo runner builds the library once in both runs. The load differs between the runs and each is a single run, so the figures are an estimate of the order, not a benchmark.

Equal counts and names (AC-7, criterion 1 of the card):

- `cargo nextest list` before (8,171 lines) and after (8,179): the 1,330 tests of `moai-cli::card <module>::<fn>` fold to the old `moai-cli::<module> <fn>` with an empty difference both ways; the 8 extra lines are the 8 new unit tests of `check_affected::tests` that this slice adds (`scratchpad/compare_lists.py`).
- Selection under old and new names for the real range `0537ee7933~1..0537ee7933` (`check affected --list`, 95 old targets): the old names pick 3,323 tests (the lines of those targets in the list before the move, plus the library and the binary); the new step's own arguments, listed by `cargo nextest list`, pick 3,331, and the difference is the same 8 new unit tests (`target/agent-3/count_selected.py`).

Known limit of the cargo runner (SPEC "Name filters"): libtest filters are substrings, so `claim_cli::` also picks `card_reclaim_cli::` and `dispatch_cli::` also picks `card_dispatch_cli::`. Nextest is exact.

Checks (every cargo run in `moai heavy-slot-acquire --name agent-3 --max 4`):

- The readers, 12 test binaries (`test_area_layout`, `check_affected_cli`, `check_affected_scans_cli`, `check_affected_tables_cli`, `home_isolation_scan`, `hook_dispatch_cli`, `leader_cli`, `repo_scans_table`, `spec_doc`, `worktree_cleanup_triggers`, `ac_verify_required`, `state_folder_literal_scan`) and the lib tests of `check_affected::`: 254 passed. Before the repair 5 of 205 failed (the solo list, the exemption tables, the home isolation baseline), then 2 of 215 (`every_check_names_a_real_target_and_test` and the state folder baseline); each is repaired above.
- The whole moved area, `cargo nextest run --test card`, first run (`target/agent-3/card-area-run.log`, 320.8 s): 1,285 passed and 45 failed, in two causes that the moves had introduced and the readers above did not show. (a) 34 tests that the home guard tolerates by `<binary>` and `<test name>` (`tests/fixtures/test_home_guard/baseline.tsv`) failed with `HOME-GUARD: touched its HOME`, because the binary is now `card` and the name `<module>::<fn>`; the 37 rows of the moved modules are `card<TAB><module>::<fn>` now, and `check_baseline` of `tests/test_home_guard.rs` finds a test of an area crate in `tests/<area>/<module>.rs` (new test `a_baseline_line_of_an_area_crate_names_the_test_in_its_module_file`). (b) 11 tests of four modules (`ledger_open`, `ledger_recover`, `ledger_reservation`, `ledger_card_landed_migration_cli`) start their own test binary again with `--exact <name>` and waited for a racer that never ran; the child name now takes its module path from `module_path!()`, and stays `<fn>` when the file is a top-level target.
- The same run after those repairs, `cargo nextest run --test card --test test_home_guard` (`target/agent-3/card-area-run2.log`, 163.3 s): 1,384 passed, 0 failed.
- The first receipt of the slice (head `c57b68a792`, `target/check-affected-c57b68a79206/`) was FAIL, `clippy=101 test=100`, and found two more: `clippy::duplicate_mod` because two card modules each loaded `tests/common/ready_receipt.rs` by `#[path]` (it is declared once in `tests/card/main.rs` now and the two modules `use crate::ready_receipt`), and `rule_history_guard::every_test_file_named_in_the_history_exists` because it looked for a moved file only at `tests/<name>.rs` (it also looks in an area crate folder now, `tests/rule_history_guard.rs`, scope widened with a reason). After both, `cargo clippy --locked --offline --all-targets -- -D warnings` exits 0 and `rule_history_guard` plus the two `card_ready_*` modules pass (29 passed).
- Guard mutants (one slot call, `target/agent-3/mutants3.py`, tree clean afterwards), each killed with exit 100 by the named test: `steps()` adds the area crate once (`area_modules_run_through_their_crate_with_a_module_filter`), the module filter is anchored at `^` (`a_target_of_its_own_stays_whole_beside_an_area_crate` and the same test), and the scans read the module files of an area crate (`exemptions_are_not_stale` and `file_read_exemptions_are_readers_with_a_reason_once`).

- Receipts, bare `moai check affected` (150 targets, 135 scans): the first run failed in clippy (`duplicate_mod` of the shared `ready_receipt` helper, repaired by loading it once in `tests/card/main.rs`) and in `rule_history_guard` (it did not read an area folder); the second failed in one test, `layout_moved_areas_agree_with_their_main_rs`, which wanted a file for the `#[path]` module (repaired: a module under `#[path]` needs no file, like `common`); the third passed with 5,033 of 5,033 tests.

Per acceptance criterion:

- AC-1: met for the card area — `cargo test --locked --offline --test test_area_layout` passes at the head (part of the 254), and the list comparison above shows each moved test as `<module>::<fn>` with the total equal plus the 8 new unit tests.
- AC-2: met — unchanged by this slice, `area_table_*` pass in the 254.
- AC-3: partial — name resolution, the one `--test` per area crate and the nextest filter set are done and pinned by 8 new unit tests of `check_affected::tests`; the verify line is `cargo test --locked --offline --lib check_affected::tests::`. REQ-11 (report every selected module of a crate that fails to build as failed) is not done and stays open.
- AC-4: met — unchanged, `ac_command_lint area_` is part of the receipt's full set.
- AC-5: met for the names this slice moves — `scan_list_names_only_existing_targets_and_paths` and `every_check_names_a_real_target_and_test` pass; `.config/nextest.toml` and `tools/ci_scope.py` name no moved target.
- AC-6: met — `solo_files_stay_top_level_targets` lists `card_cleanup_outbox_cli` and passes.
- AC-7: met for slice 1 — the numbers and equal counts above.

Gaps: (1) the figures are one run each under changing load. (2) Windows and Linux were not run. (3) REQ-11 is open, so a crate that fails to build is not reported per module. (4) The open card C-117 adds `tests/card_heartbeat_cli.rs`; the layout scan will refuse it at the top level (D-8). (5) `docs/release-rehearsal-2026-11-12.md` names `--test release_suite` as a record and is read through the table, not rewritten. (6) The shared files `src/check_affected.rs`, `tools/repo_scans.json` and `tests/fixtures/state_folder_literals/allow.tsv` were edited under a line lease, and `cargo fmt` later re-wrapped two lines of the new unit tests in `src/check_affected.rs` outside one. (7) The scope of the card grew twice with a reason (the readers, the home guard baseline). (8) The self-starting tests of four modules got a module-path prefix on the child name; that is the one change inside moved bodies, and without it eleven of them wait for a racer that never starts.

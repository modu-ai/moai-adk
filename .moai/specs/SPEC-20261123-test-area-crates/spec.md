---
id: SPEC-20261123-test-area-crates
title: "The integration tests run from eight area crates, one area moved per slice, with test names kept as module paths and check affected building only area crates"
status: draft
revision: 1
created: 2026-11-23
parent: null
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

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-9): after a slice, `tests/<area>/main.rs` declares one module per moved file, `cargo nextest list --test <area>` lists every moved test as `<old target>::<fn>`, and the total test count equals the count before the slice.
  verify: `cargo test --locked --offline --test test_area_layout`
- AC-2 (REQ-3, REQ-4): every first word of every test file and area module is in `tests/areas.toml`; a planted `tests/<moved area word>_x.rs` and a planted unmapped word each fail the scan.
  verify: `cargo test --locked --offline --test test_area_layout area_table_`
- AC-3 (REQ-5, REQ-11): for a one-`src/`-file range, `check affected --list` names at most the eight area crates plus the solo targets, and each test step carries a filter of the selected modules; a planted build failure of one area crate marks all its selected modules failed.
  verify: `cargo test --locked --offline --test check_affected_cli area_`
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

### Unsettled

- U-1: Confirm Cargo's auto-discovery of `tests/<dir>/main.rs` on rustc 1.94.1 by building one empty area crate in slice 0, before any file moves.
- U-2: Check the nine files with a process-global static one by one. Each either reads a constant once or needs REQ-8.
- U-3: The real saving per area. The report's 3x and 5x come from another project.

## Result

Status: slices 0 and 0b are written and waiting for the independent audit; no test file has moved into an area crate yet. `check affected` does not read area crates yet, so no area slice starts before the second slice of the test-thread-limit work lands and `src/check_affected.rs` can be widened.

What the two slices hold (one tracked table, the scan of it, and the readers that must not drop a moved test):

- `tests/areas.toml` maps every test file to one of the eight areas, and `tests/test_area_layout.rs` scans it.
- `tools/check_ac_commands.py`, `tools/ci_scope.py`, `src/bundle/verify.rs`, `tests/cargo_env_developer_dir.rs` and `tests/home_isolation_scan.rs` read an area folder as well as the top level.
- `moai bundle verify` puts the area crate of a changed module on its list, resolves a SPEC's old `--test <name>` to the crate, attributes a failing crate to the heads that touch its folder, and accepts only `[A-Za-z0-9_-]` target names.

Per acceptance criterion:

- AC-1: open — no area is moved, so no area crate runs.
- AC-2: met — `area_table_maps_every_real_test_file`, `area_table_rejects_a_planted_unmapped_word` and `area_table_rejects_a_planted_file_of_a_moved_area` in `tests/test_area_layout.rs`.
- AC-3: open — `check affected` is not area-aware yet (REQ-5, REQ-11); it waits for the second slice of the test-thread-limit work.
- AC-4: met — `area_old_target_name_resolves_to_its_module`, `area_crate_name_with_a_module_filter_passes`, `area_filter_that_selects_nothing_fails` and `area_unknown_target_fails_even_beside_moved_areas` in `tests/ac_command_lint.rs`.
- AC-5: partial — `scan_list_names_only_existing_targets_and_paths` holds at this head; each area slice must keep it true for the names it moves.
- AC-6: met — `solo_files_stay_top_level_targets` and `solo_rejects_a_missing_solo_file_and_a_global_state_call_in_a_module` in `tests/test_area_layout.rs`.
- AC-7: open — the before-and-after build numbers are taken in the first area slice.

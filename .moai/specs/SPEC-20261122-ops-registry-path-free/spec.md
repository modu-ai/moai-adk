---
id: SPEC-20261122-ops-registry-path-free
title: "The generated operation registry names no file path, so a copied or shared build folder compiles this worktree's operations"
status: draft
revision: 2
created: 2026-11-22
---

# The generated operation registry names no file path, so a copied or shared build folder compiles this worktree's operations

## Intent

`build.rs` writes `ops_registry.rs` into the build script's output folder, and today every operation in it is declared as `#[path = "<absolute worktree path>/src/ops/<word>.rs"] pub mod <word>;`. Cargo gives two worktrees of the same layout the same output folder hash, and the build script only reruns when `src/ops` looks newer than its last run. A build folder that another worktree filled — shared, or warmed by an APFS clone (`cp -c`), which the user-approved redesign makes the normal way to warm a new worktree — can therefore compile the other worktree's operation files, or stop with "file not found for module" once that worktree is gone. This was seen in the shared audit build folders (R-1).

The registry becomes a list of names: `build.rs` writes a `with_ops!` macro that holds each operation's file word, its variant name and its help lines, and `src/ops/mod.rs` expands it with its own `declare_ops!` macro, so `pub mod <word>;` is resolved next to `src/ops/mod.rs` like any hand-written module. The generated text no longer depends on where the worktree is. It serves every lane and audit that builds in a cloned or reused build folder, and it is needed before warming by clone.

## Non-goals

- Splitting the library and the binary into two packages so an `src/ops` change no longer rebuilds the library (build-review 5.4): a crate-structure card.
- Building one binary instead of `moai` and `moai-cli` (build-review 5.6) and `test = false` for the `moai` binary: their own cards.
- The bundle-pack table `bundle_pack_files.rs`: it embeds files through `concat!(env!("CARGO_MANIFEST_DIR"), …)` at compile time of the including crate, which resolves in the compiling worktree, not in the build script's output.
- Shared build folders between worktrees of different commits stay unsupported (build-review 6.5); this SPEC removes the path, not the timestamp judgment cargo makes.

## Requirements

- REQ-1: The build script shall write the operation registry as operation words, variant names and help lines only, with no file path.
- REQ-2: The build script shall derive the registry text only from the names and help lines of the files under `src/ops/`.
- REQ-3: The operations module shall declare each listed operation as a module that the compiler finds next to `src/ops/mod.rs`.
- REQ-4: The command line shall offer every operation under the same subcommand name and with the same help text as before this change.
- REQ-5 (REQ-1): IF the registry text holds a file path, THEN the registry test shall fail and name the line that holds it.

## Acceptance

- AC-1 (REQ-1, REQ-2): the generated `ops_registry.rs` (read by the test from the build script's output folder) holds a `macro_rules! with_ops` list with one entry per `src/ops/<word>.rs` in file-name order, and holds no `#[path`, no `src/ops/` and no part of `CARGO_MANIFEST_DIR` · verify: `cargo test --offline --locked --test hot_file_ops_merge the_registry_is_a_list_of_names_with_no_path`
- AC-2 (REQ-3, REQ-4): for every operation file, `moai <word with - for _> --help` exits 0 and its first line is the first `///` line above that file's `pub struct Args`; the registry's word list equals the files under `src/ops/` · verify: `cargo test --offline --locked --test hot_file_ops_merge`
- AC-3 (REQ-5): the test's path finder, fed registry samples that hold an absolute path, a `#[path` attribute or `src/ops/`, names that line, and fed the real registry it names nothing · verify: `cargo test --offline --locked --test hot_file_ops_merge a_path_in_the_registry_is_named`

## Scope

- write: `build.rs`, `src/ops/mod.rs`, `tests/hot_file_ops_merge.rs`, `.moai/specs/SPEC-20261122-ops-registry-path-free/` (`tests/ops_registry_path_free.rs` was written and is removed again, D-4)
- read-only: `src/ops/*.rs` (the operation files), `src/main.rs`, `src/bin/moai.rs`, `Cargo.toml`
- forbidden: every other file

## Plan

1. This slice: the SPEC, the red test (written as `tests/ops_registry_path_free.rs`, which fails today on the `#[path` lines, then moved into `tests/hot_file_ops_merge.rs`, D-4), then `build.rs` and `src/ops/mod.rs`, then `tests/hot_file_ops_merge.rs` reads the new list, then one check affected and READY.

## Risks

- Module paths inside a macro: a `mod <word>;` that a macro expands could be looked up relative to the generated file in the output folder (E0583). The build review proved on a test crate that the `declare_ops!` form defined and invoked in `src/ops/mod.rs` resolves next to `src/ops/mod.rs`, and that dropping `#[path]` alone fails (R-1); the full build of this slice is the check.
- Help text changes: `#[doc = …]` lines now pass through a macro as literals; AC-2 compares each operation's help line with its file.
- rustfmt does not format macro bodies; the generated file is not formatted today either.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Each registry entry carries the variant name (`AgentsApply`) beside the file word (`agents_apply`). | `macro_rules!` cannot build an identifier from another one; a name-pasting crate would be a new dependency for one line each. | a later card that adds such a crate |
| D-2 | The `Ops` enum, `operation()` and `run()` move from generated text into the `declare_ops!` macro in `src/ops/mod.rs`. | The generated text keeps only inputs (names and help lines), as build-review section 6 asks; the code shape is reviewed source instead of generated strings. | restore the generated functions |
| D-3 | `tests/hot_file_ops_merge.rs` reads the operation words from the new list instead of `pub mod` lines. | Its registry check (#2617) parses the generated text; the list is the new shape of the same fact. | none needed |
| D-4 | The registry tests live in `tests/hot_file_ops_merge.rs`, not in a file of their own. | They read `src/ops/`, and `tests/repo_scans_test.py` requires every test that walks the tree to be listed in `tools/repo_scans.json`; that file already lists `tests/hot_file_ops_merge.rs` under the `operation-files` scan with the `src/ops/` scope, and `tools/repo_scans.json` is reserved by another card. The scan's check now runs these tests too. | a new file plus its own `tools/repo_scans.json` entry |

### Research

| Id | Source says | Used for |
| --- | --- | --- |
| R-1 | `~/.moai/reports/moai-adk/build-review-2026-11-22.md` 5.3 (read whole on 2026-11-22): "`build.rs:179-185` 가 `#[path = \"<CARGO_MANIFEST_DIR>/src/ops/<word>.rs\"]` 를 생성해요"; an audit build folder held a registry that pointed into a removed worktree; the macro form "시험 결과: `7 [\"alpha\", \"beta\"]`"; "`#[path]` 를 빼고 `pub mod alpha;` 만 생성하면 … E0583". | Intent, REQ-1, REQ-3, Risks |
| R-2 | The same report section 6: the generated file holds no path, and its content depends only on its inputs (operation file names and doc lines), with no time, path or environment. | REQ-2, D-2 |
| R-3 | `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` 3.B (read whole): "`build.rs`의 절대 경로 제거(경로 없는 `with_ops!` 매크로, 시험 크레이트로 확인)"; 3.A: new and reset worktree slots are warmed with `cp -c -R -p seed/target`. | Intent |

## Result

Slice 1. Status draft; only the lines below are claimed.

- AC-1: met. `the_registry_is_a_list_of_names_with_no_path` passes in `tests/hot_file_ops_merge.rs`: the generated `ops_registry.rs` is a `macro_rules! with_ops` list whose entry words equal the files under `src/ops/` in file-name order, and no line holds `#[path`, `src/ops/` or the package folder.
- AC-2: met. `every_operation_keeps_its_name_and_help_line` passes: every operation answers `--help` under its kebab-case name with the first `///` line of its `Args` as the first help line; `the_registry_holds_every_operation_file` reads the words from the new entry lines.
- AC-3: met. `a_path_in_the_registry_is_named` passes: samples holding an absolute path, a `#[path` attribute or `src/ops/` are named by line, and the real registry yields none.
- The tests were red first (d531b20042: the registry held `#[path = "<worktree>/src/ops/<word>.rs"]` lines and no list) and green with `build.rs` and `declare_ops!` (454fbe5c17); D-4 moved them into the listed scan file (fd22e6ad6a).
- Not claimed: a run in a build folder cloned from another worktree; the change removes the path the generated file depended on, and cargo's own rerun judgment is unchanged (Non-goals).

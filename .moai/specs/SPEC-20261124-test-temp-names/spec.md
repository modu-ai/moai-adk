---
id: SPEC-20261124-test-temp-names
title: "Test scratch folders come from tempfile, and a scan refuses a temp folder named from the process id"
status: draft
revision: 1
created: 2026-11-24
---

# Test scratch folders come from tempfile, and a scan refuses a temp folder named from the process id

## Intent

The full Rust test run of the leader on the integration tip `466b59fc62` failed one test of 7393: `profile_plan_cli dangerous_branch_names_never_reach_git_as_arguments`, with `git remote add origin …/T/c105-plan-46447/w0/remote.git: error: remote origin already exists`. The test named its scratch folder `std::env::temp_dir()/c105-plan-<pid>/w<n>` and never removed it, so the system temp folder held 319 folders `c105-plan-*`; when nextest ran a test process whose pid had been used before, the name was already taken and the earlier run's `remote.git` was still in it. The same test passes alone (9 of 9). This is the second time one root cause (a scratch name made from a pid or a shell `$$`, never removed) reaches a red run; the first was the wrapper of the audit B1 finding of C-406. CLAUDE.md asks that such a class is closed at the source.

This SPEC repairs the sites that create a folder under the system temp folder with a pid name and adds one scan test that refuses a new one.

Grounded in the code at the integration tip `466b59fc62`:

- `tests/profile_plan_cli.rs:46` and `:346` build `temp_dir().join(format!("c105-plan-…{pid}…"))` and create it; nothing removes it.
- `tests/work_flow_plan_cli.rs:65` builds `temp_dir().join(format!("wfplan-{pid}-{n}"))` and creates it; nothing removes it.
- `tests/commit_identity_guard_2878.rs:166` builds `temp_dir().join(format!("moai-2878-none-{pid}"))` as a path that must not exist; a leftover folder of that name would make the test lie.
- `tempfile` is already a dependency of the test build (`Cargo.toml`), and its `TempDir` removes the folder when it is dropped, panic or not, with a name the operating system makes unique.

## Non-goals

- Removing the folders already left in the system temp folder: they are the operator's.
- The pid used as data (a lock file naming its owner, a key name, a file name inside a `TempDir`, a liveness probe): the pid is the point there, and nothing is created under the system temp folder by its name.
- Paths that are only named and never created (`tests/hook_dispatch_cli.rs`, a sibling of its `TempDir`), and a folder that is removed before it is made and again on drop (`tests/host_wake_cli.rs`, a short socket path under `/tmp`): they cannot meet a leftover.
- Unit tests inside `src/` (`src/review.rs`, `src/host_wake.rs` and others): a list in the card note; changing them needs the cards that own those files.
- Any change of what the repaired tests assert.

## Requirements

- REQ-1: When one of the tests of `tests/profile_plan_cli.rs`, `tests/work_flow_plan_cli.rs` or `tests/commit_identity_guard_2878.rs` needs a scratch folder, the test shall take it from `tempfile::TempDir` · failure: none — the folder is removed when the test ends, panic or not, and its name does not come from the process id.
- REQ-2: When a statement under `tests/` holds `std::env::temp_dir()` and `std::process::id()`, the scan test shall fail and name the file and the line · failure: none — a statement is the text between two of `;`, `{` and `}`, comments are ignored, and there is no allow table.
- REQ-3: When a statement under `tests/` holds `std::process::id()` and no `temp_dir()` (a key name, a lock owner, a file name inside a `TempDir`), the scan test shall not fail · failure: none — the pid used as data stays allowed.
- REQ-4: When the repository scans table is read, the table shall list the scan test as a scan of the tests folder · failure: none — the existing completeness test of the table is the check.

## Acceptance

- AC-1 (REQ-1): the three test files hold no name made from the process id under the system temp folder, and a run of `profile_plan_cli` with its own `TMPDIR` leaves that folder empty · verify: `cargo test --locked --offline --test profile_plan_cli` and, by hand, the entry count of the private `TMPDIR` after the run (not a test)
- AC-2 (REQ-2): a statement that joins a pid name onto `temp_dir()`, on one line or over several, is reported with its line · verify: `cargo test --locked --offline --test test_temp_names_scan temp_names_scan_flags_`
- AC-3 (REQ-3): a pid used as a key name, a lock owner or a file name, and a commented line, are not reported · verify: `cargo test --locked --offline --test test_temp_names_scan temp_names_scan_ignores_`
- AC-4 (REQ-2): the checked-in `tests/` folder holds no such statement · verify: `cargo test --locked --offline --test test_temp_names_scan temp_names_tree_`
- AC-5 (REQ-4): the table lists the scan and its completeness test passes · verify: `cargo test --locked --offline --test repo_scans_table`

## Scope

- write: `.moai/specs/SPEC-20261124-test-temp-names/spec.md`, `tests/profile_plan_cli.rs`, `tests/work_flow_plan_cli.rs`, `tests/commit_identity_guard_2878.rs`, `tests/test_temp_names_scan.rs` (new)
- write by the leader's word, because card C-405 holds it: `tools/repo_scans.json` (one entry for the new scan)
- read-only: every other file under `tests/`, `Cargo.toml`
- forbidden: the system temp folder's existing contents, `src/`

## Plan

1. This SPEC, then the new scan test with its fixture cases and the tree case, red on the current tree (it names the four sites).
2. Repair `tests/profile_plan_cli.rs` first (two sites), then the other two files; the scan turns green.
3. The table entry, after C-405 or by the leader's word.
4. One affected run at the READY head, the mutants, the READY note.

## Risks

- A statement scan is a text match: it shows that a statement holds both names, not that the folder is created. It over-reports a path that is only named, and that is the intended direction (the repair is a `TempDir`, which costs nothing).
- A pid name hidden behind a helper function (`fn scratch() -> PathBuf`) is not seen. The scan catches the direct form that produced two red runs; a helper is a next card if one appears.
- `tools/repo_scans.json` is held by another card; until its entry exists, the completeness test of the table names the new file.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | The scan covers `tests/` and has no allow table; the three sites that match are repaired | A table of allowed sites would keep the class alive; the repair of each site is one `TempDir` | Add an allow table with a reason per entry |
| D-2 | Unit tests in `src/` are listed, not changed or scanned | Other cards hold those files; the two red runs came from `tests/` | A later card widens the scan to `src/` |
| D-3 | The folders already in the system temp folder are not removed | The operator's folder; CLAUDE.md forbids deleting what a card did not make | The operator removes `c105-plan-*` and `wfplan-*` |

## Result

Not yet written; it is filled in the SYNC step with the runs of the READY head.

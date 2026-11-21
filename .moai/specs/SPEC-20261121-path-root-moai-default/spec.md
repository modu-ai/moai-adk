---
id: SPEC-20261121-path-root-moai-default
title: "A tree with no earlier state folder resolves to .moai, and .moai-cli stays only where it already exists"
status: draft
revision: 1
created: 2026-11-21
---

# A tree with no earlier state folder resolves to .moai, and .moai-cli stays only where it already exists

## Intent

The leader dispatch of 2026-11-21 (user order to carry every audit repair) asks that `.moai` (a project) and `~/.moai` (a HOME) are the default roots, and that `.moai-cli` and `~/.moai-cli` are read only where they already exist and by a later migration step. This SPEC is slice 1: the three resolvers and their tests. No data moves, nothing is deleted, no installer or shell file changes, and no test reads or writes the real HOME.

What the tree does today (read at tip 6e1bedb4f3a4 in this worktree):

- `src/state_root.rs` `dir` returns `.moai` only when `classify` says `rust` (a valid marker file `.moai/moai.json`, or the earlier `.moai/moai-cli.json`); every other verdict returns `.moai-cli` (`PROJECT_STATE_DIR`). `src/home_root.rs` `dir` and `src/clone_root.rs` `dir` do the same for `~` and for the git common folder. A tree that never wrote a marker therefore resolves to `.moai-cli` even when it never had that folder. A fresh worktree of this repository is such a tree: its `.moai/` holds the tracked `specs/`, `evals/` and `project.yaml`, no marker and no `.moai-cli`.
- `moai-cli install-inspect` run with an empty temporary HOME reports `home_state_root` ending in `/.moai-cli` (measured with the built binary, HOME set to an empty scratch folder). `install/install.sh` takes its state folder from that report (lines 10-13 and 402-406) and never spells `.moai-cli` as a path, so the installer follows the resolver and needs no change in this slice.
- Earlier repairs fixed one reader at a time. Counted with `grep -rn` over `src` in this worktree, comment and test lines included: `state_root::` 261 lines in 57 files, `clone_root::` 14 lines in 9 files, `home_root::` 71 lines in 27 files, `PROJECT_STATE_DIR` or `project_state!` 100 lines in 21 files, `fixed_state!` 11 lines in 6 files, `HOME_STATE_DIR`, `CLONE_STATE_DIR` or `WORKTREES_DIR` 39 lines in 12 files. The lists are kept in the plan below. Another 186 files under `tests/` name `.moai-cli`.

The rule decided here, for a project, a HOME and a git common folder alike. "Earlier folder present" means `.moai-cli` exists in any form (a link counts, and is never followed).

| `.moai` verdict | earlier folder present | resolves to |
| --- | --- | --- |
| `rust` (valid marker) | either | `.moai` (unchanged) |
| `rust-first` (HOME, schema-2 marker) | either | the layout root under `.moai` (unchanged) |
| any other | yes | `.moai-cli` (unchanged: a legacy tree keeps every path it had) |
| `none`, `empty`, `unknown` (a `.moai/` with files that are neither Go markers nor a marker; a specs-only `.moai/` reads as `none`) | no | `.moai` (new) |
| `go`, `mixed`, `interrupted`, `refuse` | no | `.moai-cli` (unchanged: Go's folder is never written) |

Consequences:

- A new tree, including a worktree without a marker, resolves to `.moai` and never creates `.moai-cli` through a resolver. The resolvers read only.
- A tree that has `.moai-cli` keeps working unchanged until its migration step runs; nothing here moves or reads it any differently.
- A Go-generation `.moai/` with no `.moai-cli` still resolves to `.moai-cli`, so no Go file is overwritten. This is the one row that keeps creating the earlier name; it is a decision of this SPEC (D-1) that the leader can reverse once Go is gone.
- The sites that spell the earlier name directly (`project_state!`, `fixed_state!` and the constants) are not changed in this slice. A tree that a resolver sends to `.moai` and a direct writer then gives a `.moai-cli` resolves to `.moai-cli` from then on, so state a resolver-based writer put under `.moai` earlier would be stranded. The direct sites are the next slice, and the marker (written by `adopt_new` for a new tree) keeps `.moai` in front where it exists.

What stays with other work: the project-side move step of `/moai update` with preview, confirm token, receipt and revert (and the clone-shared and HOME follow-ups it names) is the project state move card; the Rust-only switch and the retirement of the Go commands is the Rust-only card. This SPEC changes neither the Go verdict rows nor any migration code.

## Non-goals

- Moving, copying or deleting any state folder, the PATH link, the vault or any receipt.
- `install/install.sh`, `install/install.ps1`, shell profile files and documentation texts.
- The direct sites that spell `.moai-cli` (`project_state!`, `fixed_state!`, `PROJECT_STATE_DIR`, `HOME_STATE_DIR`, `CLONE_STATE_DIR`, `WORKTREES_DIR`) and `src/lib.rs`: unchanged here.
- The real HOME and the real project state: every test uses a temporary HOME and a temporary project.

## Requirements

- REQ-1: The project resolver shall return `.moai` for a project whose `.moai` holds a valid Rust marker, whether or not `.moai-cli` exists.
- REQ-2: The project resolver shall return `.moai-cli` for a project that has `.moai-cli` and no valid Rust marker, so that every path of a legacy project stays as it was.
- REQ-3: The project resolver shall return `.moai` for a project that has no `.moai-cli` and whose `.moai` is absent, empty, holds only `specs/` or holds files that are neither Go markers nor a Rust marker.
- REQ-4: The project resolver shall return `.moai-cli` for a project that has no `.moai-cli` and whose `.moai` is Go's, mixed, a stopped migration, a link or another non-folder, so that no Go file is written.
- REQ-5: The HOME resolver shall give the same answers as the project resolver for the rows of the table, keeping the layout root of a schema-2 marker.
- REQ-6: The clone resolver shall give the same answers as the project resolver for the rows of the table, over the git common folder.
- REQ-7: The three resolvers shall read only, so that a call changes no file, folder or link under the tree. · failure: none — a read changes nothing
- REQ-8: When `.moai-cli` is a link, the resolvers shall count it as present without following it. · failure: none — a link is only looked at, never opened
- REQ-9: When the project is new, the answers of `is_new_project`, `create_dir` and `adopt_new` shall stay as they are, so that the marker of a new project is still written once. · failure: none — a project that is not new gets no marker
- REQ-10: The tree shall carry one test per row of the table for each of the three resolvers, run on temporary folders only. · failure: none — a missing case fails the test run

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-8, REQ-9): the project resolver gives `.moai` or `.moai-cli` as the table says for a tree with a marker, with only `.moai-cli`, with both folders, with neither, with a specs-only `.moai/`, with a worktree-like `.moai/` that holds `project.yaml` and `specs/` and no marker, with Go markers, with a Go-and-Rust mixed folder, with a link as `.moai-cli`, and a new project still gets its marker once · verify: `cargo test --test path_root_default project_`
- AC-2 (REQ-5): the HOME resolver gives the same answers for an empty HOME, a HOME with only its earlier state folder, with both folders, with a Go state folder, with a schema-1 marker beside Go files and with a schema-2 marker, and `install-inspect` in an empty temporary HOME reports `.moai` · verify: `cargo test --test path_root_default home_`
- AC-3 (REQ-6): the clone resolver gives the same answers for a common folder with nothing, with only `.moai-cli`, with a marker and with an empty `.moai` beside `.moai-cli` · verify: `cargo test --test path_root_default clone_`
- AC-4 (REQ-7): a snapshot of every file, folder and link of the temporary tree before and after calls to `dir`, `path` and `rel` of the three resolvers is the same, and no `.moai-cli` appears in a tree that had none · verify: `cargo test --test path_root_default reads_only`
- AC-5 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): the unit tests of the three resolver files that stated the earlier answer for an empty tree follow the table, the library and the binaries build and pass, and the SPEC checks clean · verify: `cargo test --lib && cargo test --bins && moai-cli spec check SPEC-20261121-path-root-moai-default`

## Scope

- write: `src/state_root.rs`, `src/home_root.rs`, `src/clone_root.rs` (the resolvers and their unit tests), new `tests/path_root_default.rs`, and this SPEC folder
- read-only: `src/lib.rs` (held by another change; its constants and doc comments stay), `src/identity.rs`, `install/install.sh`, every direct site named above
- forbidden: the Go repository, the real HOME, any move or delete of state, installer or shell profile changes, any card number in a file name, SPEC id, branch or commit

## Plan

Tests first (RED): `tests/path_root_default.rs` builds each row of the table in a temporary folder and asks the library resolvers; on the tip the rows with no earlier folder and no marker answer `.moai-cli` and fail for that reason. Then one change per resolver (GREEN): a helper in `src/state_root.rs` that says whether the earlier folder is present (`symlink_metadata`, a link counts) and the new answer rule; `src/home_root.rs` and `src/clone_root.rs` use the same helper over their own verdict. The unit tests that pinned the old answer for an empty tree move to the table in the same commit as the code.

Readers of the resolvers and of the earlier name (files, counted as above) for the next slices: direct sites by file — `src/agents.rs` (16 lines), `src/message.rs` (15), `src/apply.rs`, `src/project_id.rs`, `src/profile.rs`, `src/lane_loop.rs`, `src/statusline/state.rs`, `src/claude.rs`, `src/layout.rs`, `src/go_state.rs`, `src/project_home.rs`, `src/project_detect.rs`, `src/hooks.rs`, `src/hook_engine.rs`, `src/check_tree.rs`, `src/forge/report.rs`, `src/project_hosts.rs`, `src/migrate.rs`, `src/scope.rs`; fixed paths — `src/worktree_move.rs`, `src/optimize.rs`, `src/key_vault.rs`, `src/go_settings_clean.rs`, `src/baseline.rs`; HOME and clone — `src/home_install.rs`, `src/usage_limit.rs`, `src/project_home.rs`, `src/home_workdir.rs`. Next slices: route these through the resolvers, the move step, the installer and the texts.

## Risks

- A tree that a resolver now sends to `.moai` can gain a `.moai-cli` from a direct writer, which makes it resolve to `.moai-cli` and strands the earlier `.moai` state. Slice 1 keeps the resolvers read-only and pins the stranding in a test comment; the direct sites are the next slice, and a new tree's marker keeps `.moai` first.
- Tests that wrote into an empty temporary tree through a resolver and then looked under `.moai-cli` change their expected folder; `check affected` finds them. Any such file held by another change goes through a pair note before a one-line follow.
- The unit tests of the three resolvers stated the old answer for an empty tree; they move with the code.
- `.moai/` of a repository can track files (this repository tracks `project.yaml`, `evals/` and `specs/`); state written there needs an ignore rule. `.moai/state/` is ignored already; other state folders are not part of this slice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A Go, mixed, interrupted or refused `.moai` with no `.moai-cli` still resolves to `.moai-cli` | The Go folder must never be written (dispatch rule: legacy isolation first); this is today's answer | Flip the one row once the Rust-only switch has removed Go |
| D-2 | `.moai-cli` counts as present in any form, a link included, and is never followed | Same link rule as every marker read in `state_root.rs` | none |
| D-3 | `src/lib.rs` is not edited in this slice | Another change holds it; the resolvers need no new constant | Edit the doc comments when that change lands |
| D-4 | An `unknown` `.moai/` with no `.moai-cli` resolves to `.moai` | A worktree of a repository that tracks files under `.moai/` is this case; resolving it to `.moai-cli` is what the dispatch calls the bug | Read it as `.moai-cli` again in one match arm |
| D-5 | The installer is not changed | It reads `install-inspect`, which follows the resolver (measured) | none |

## Result

(to be written at the end of the change)

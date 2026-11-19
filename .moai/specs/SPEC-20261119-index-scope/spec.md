---
id: SPEC-20261119-index-scope
title: "index-build can be scoped to folders, and a refusal over the file limit names where the files are"
status: draft
revision: 1
created: 2026-11-19
---

# index-build can be scoped to folders, and a refusal over the file limit names where the files are

## Intent

Card C-58. In a bun monorepo (7 apps, 22 packages) `moai index-build` refused with `index file limit exceeded (16384); the scan is refused, not truncated`, wrote nothing, and the route offered nothing else (transcript of `/moai project`, 2026-11-16). Two gaps: the refusal does not say where the files are, and a request has no way to name the part of the repository to index.

Measured from the code at tip 9bd09c2eff (`src/index.rs`): the limit counts **tracked** files only (`git ls-files -z`, minus `<state>/index/`). Gitignored, untracked files such as `node_modules` never count; a vendored folder that is committed does. This slice adds (1) an optional `scope` list to the build request: only the tracked files under those folders are indexed, the manifest records the scope, and `index-query` judges freshness inside it; (2) a refusal over the limit that names the largest folders with their file counts and shows the request that would be accepted. The route's selection ("scope to chosen folders, recommended / skip", `project.md` step 7) is the second slice.

## Non-goals

- The route text (`skills/moai/workflows/project.md` step 7) and its manifest hash: second slice, after card C-99 releases the manifest path.
- A command-line flag: `index-build` takes its request as JSON on stdin, so `scope` is a request field.
- Raising `MAX_FILES`, or truncating a scan: the scan stays refused, not truncated.
- The failure marker of a refused build (`partial`): unchanged for the limit refusal.

## Requirements

- REQ-1: When the request of `index-build` carries a `scope` list, `index-build` shall index only the tracked files under the listed folders and record the normalized list in the manifest.
- REQ-2 (REQ-1): IF a scope entry is not a plain relative folder path, the list has more than 64 entries, or no tracked file lies under the list, THEN `index-build` shall refuse the request as `invalid_input` before it writes anything.
- REQ-3: When more than 16384 files would be indexed, `index-build` shall refuse the build as `invalid_input`, name the ten largest folders with their file counts and show the request with a `scope` that would be accepted. · failure: none — the refusal is the behavior; the previous generation stays and queries report `partial` as before.
- REQ-4: The limit shall count the tracked files of the scan only, so that gitignored untracked files do not count.
- REQ-5: While the published generation records a scope, `index-query` shall judge freshness from the tracked and untracked files under that scope only.
- REQ-6: Where a build has a scope, the generation id shall depend on the scope.
- REQ-7: A build without a scope shall keep the generation id and the manifest bytes it had before this change.

## Acceptance

- AC-1 (REQ-4): in a fixture, 20 gitignored untracked files under `node_modules/` do not count while the same folder committed does; the count shows in `files`.
  verify: `cargo test --locked --offline --test index_scope_cli the_limit_counts_tracked_files_only`
- AC-2 (REQ-3): at exactly 16384 files the build succeeds, at 16385 it is refused as `invalid_input` and the message names the ten largest folders in count order with their counts, and shows `"scope"` in an example request; the unit test pins the folder ranking and its tie-break.
  verify: `cargo test --locked --offline --test index_scope_cli the_limit_is_n_built_and_n_plus_one_refused_naming_the_folders`; `cargo test --locked --offline --lib index::tests::largest_folders`
- AC-3 (REQ-1, REQ-6, REQ-7): a scoped build indexes only the files under the scope, records it in the manifest, and an unscoped build writes no `scope` key.
  verify: `cargo test --locked --offline --test index_scope_cli a_scoped_build_indexes_only_the_scope_and_records_it`; `cargo test --locked --offline --test index_scope_cli an_unscoped_manifest_has_no_scope_key`
- AC-4 (REQ-1, REQ-3): a repository over the limit builds with a scope under it; the scoped count has its own N and N+1 edge.
  verify: `cargo test --locked --offline --test index_scope_cli a_scope_builds_a_repository_over_the_limit`
- AC-5 (REQ-2): an absolute path, `..`, an empty entry, a control character, more than 64 entries (64 accepted) and a scope with no tracked file are `invalid_input` and leave no generation and no failure marker.
  verify: `cargo test --locked --offline --test index_scope_cli scope_refusals_write_nothing`; `cargo test --locked --offline --lib index::tests::scope_normalization`
- AC-6 (REQ-5): after a scoped build `index-query` is `fresh`; a change, an added file or an untracked file outside the scope keeps it `fresh`, one inside turns it `stale`.
  verify: `cargo test --locked --offline --test index_scope_cli a_scoped_generation_ignores_changes_outside_the_scope`
- AC-7 (REQ-1…REQ-7): neighbours.
  verify: `cargo test --locked --offline --test index_cli`; `cargo test --locked --offline --test freshness_fixture`; `cargo test --locked --offline --test atlas_query_cli`; `cargo test --locked --offline --test code_atlas_reference`; `cargo test --locked --offline --test doc_reference_boundaries`; `cargo test --locked --offline --test product_cli`; `cargo test --locked --offline --lib`; `cargo fmt --check`; `cargo clippy --locked --offline --all-targets -- -D warnings`

## Scope

- write: `src/index.rs`, `src/contract.rs` (one phrase, `index scope`), `tests/index_scope_cli.rs`, `docs/index-generation.md`, `docs/design-review/appendix/interfaces.md`, `.moai/specs/SPEC-20261119-index-scope/spec.md`
- read-only: `src/main.rs` (the request is read with `input::<IndexBuildRequest>()`, no change needed), `tests/index_cli.rs`, `tests/common/**`
- forbidden: the real HOME and ledger, the network, `skills/moai/**` (second slice), a change of `MAX_FILES`

## Plan

1. Tests first: the CLI tests and the two unit tests; they fail to compile or fail before the code.
2. `IndexBuildRequest.scope`, `normalize_scope`, `largest_folders`, a scan that takes the limit and the scope, `IndexManifest.scope`, the generation id and `same_snapshot` with the scope, `index-query` using the manifest scope.
3. `contract.rs` phrase so a wrapped scope refusal stays `invalid_input`; docs.

## Risks

- A repository whose `git ls-files` output is over 4 MiB is refused earlier with `git ls-files output limit exceeded`, before the files can be counted, so that refusal cannot name folders; a scoped request passes git a pathspec and avoids it. The scoped request is in the answer of the count refusal only.
- A refused over-limit build marks the index `partial` (existing behavior, kept): the previous generation stays readable but queries report `partial` until the next good build.
- Other readers of the manifest (`search`, `atlas`) read `files` as written, so they see the scoped file list; none re-scans the repository to compare, and `index_records` calls `index_query`, which honors the scope.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `scope` is a JSON request field, a list of folders relative to the project, matched on whole path components | `index-build` has no command-line options; the card asks for scoping to chosen folders | Add a flag later that fills the same field |
| D-2 | Folders in the refusal are grouped by their first two path components (`apps/bo`), files nearer the root by their first component; ten are named, largest first, ties by path | the monorepo shape of the card (7 apps, 22 packages) | Change the depth |
| D-3 | Scope entries are deduplicated and an entry under another entry is dropped | `apps` and `apps/bo` mean `apps` | Keep both |
| D-4 | The scope is part of the generation id only when it is not empty | old ids and manifests stay byte-identical | Always include it |

## Result

(RED and the checks come first; one line per AC is written after them.)

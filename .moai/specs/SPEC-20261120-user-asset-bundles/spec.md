---
id: SPEC-20261120-user-asset-bundles
title: "Port the Go user-asset bundle commands (add, remove, list) to Rust"
status: draft
revision: 1
created: 2026-11-20
---

## Intent

The Go build's `moai bundle add | remove | list` (user-asset optional packs: opt-in skill
and agent sets installed into the user's folders, recorded in `~/.moai/user-assets.json`)
is one of the words the name-switch shim still sends to Go. The user decision to run Rust
only needs these three commands in Rust, with the same asset file, the same output and the
same selection semantics users depend on, so the shim can stop routing the word to Go.

Survey (read-only): Go `internal/cli/bundle.go` (231 lines) drives the `userassets` engine
(install, manifest, lock, journal, paths); the assets live in the embedded template tree and
the optional packs are eleven pack sections totalling 45 entries (two of them empty). The
Rust side already owns the word
`bundle` for the M-Flow bundle (`ready | check | body | build | verify` in `src/ops/bundle.rs`),
so the three commands join that command as additional actions and the existing five stay
unchanged.

## Non-goals

- Divergence preservation (a target that matches an older shipped copy is backed up and
  kept) and the Codex-face conversion report: not in the design-note cases; a target whose
  bytes differ from ours is a collision-skip here.
- The name-switch shim's word list: the shim is an installed artifact outside the
  repository; moving the word is the Rust-only card's final step, not a commit here.
- `init --resume` integration, the Windows lock flavour, and the catalog tree-hash audit.
- Changing the five existing bundle actions' behaviour or help text.

## Requirements

- REQ-1: While the catalog holds optional packs, `moai bundle list` shall print one line per pack in ascending name order in the Go format (a ten-column left-aligned name, then two three-column right-aligned counts followed by the words skills and agents). · failure: none — pinned by AC-1.
- REQ-2: While `moai bundle add` or `moai bundle remove` names a pack absent from the catalog, the command shall refuse with the Go refusal message (unknown bundle, the quoted name, then the valid names in ascending order) and change nothing. · failure: none — pinned by AC-2.
- REQ-3: While `moai bundle add <pack>` names a catalog pack, the command shall record the selection in the user-assets manifest at the Go path under the home in the Go manifest shape (schema version, the sorted bundles list, per-file entries with sha256, bundle, installed_at and moai_version, and the collisions list) and install every entry of the pack into its deploy root under the user's home. · failure: none — pinned by AC-3.
- REQ-4: While an install target already exists, the command shall count an identical target as refreshed without rewriting it, count a differing target as a collision-skip that records the path with its first-seen time and keeps the user's bytes, and print the Go summary line (bundle add, then installed, refreshed, removed, collision-skipped, divergence-preserved counts) with one reason line per skipped or kept path. · failure: none — pinned by AC-3.
- REQ-5: While `moai bundle remove <pack>` runs, the command shall delete exactly the removed pack's installed entries that are not shared with the base set or a remaining selection (printing a kept line for each survivor), defer an entry that is a declared dependency of a preserved asset (printing a deferred line), save the manifest with the remaining selection, and clear the pending-install journal. · failure: none — pinned by AC-4.
- REQ-6: While add or remove runs, the command shall hold an exclusive flock on the Go guard file next to the manifest for the whole body, waiting in a bounded retry loop, so a guard left by the Go binary is honoured. · failure: none — pinned by AC-5.
- REQ-7: While an add or remove writes assets, the command shall record the intended delta in the pending-install journal before the first asset write, with the next run reconciling an existing journal before installing (an absent target installs, a hash-equal target is claimed, a mismatched target is preserved as a collision). · failure: none — pinned by AC-5.

## Acceptance

- AC-1 (REQ-1): list prints every pack section of the catalog, sorted, in the exact Go line format.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`
- AC-2 (REQ-2): an unknown name on add and on remove refuses with the exact message and leaves the manifest untouched.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`
- AC-3 (REQ-3, REQ-4): add of a pack in a temporary home writes the manifest in the Go shape, installs the pack's files with the embedded bytes, and prints the Go summary line with the pack's counts.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`
- AC-4 (REQ-5): remove of a lone selection deletes exactly its pack's files, empties the selection and clears the journal; the complement rule (an entry another selection still names survives with the kept line) and the deferral rule (an entry a preserved asset declares survives with the deferred line) are pinned with a synthetic catalog, because no two shipped packs share an entry.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`; `cargo test --locked --offline --lib user_assets`
- AC-5 (REQ-6, REQ-7): a journal left by an interrupted run is reconciled on the next run — a mismatched target is preserved as a collision through the CLI, and a matching one installs through the engine; the journal is gone after a successful run.
  verify: `cargo test --locked --offline --test bundle_user_assets_cli`; `cargo test --locked --offline --lib user_assets`
- AC-6 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7): neighbours — the five existing bundle actions and the repository scans.
  verify: `cargo test --locked --offline --test bundle_ready_cli --test bundle_body_cli --test bundle_check_cli --test bundle_build_cli`; `python3 -m unittest tests.repo_scans_test`; the built binary's `spec check SPEC-20261120-user-asset-bundles`

## Scope

- write: `src/user_assets.rs`, `src/lib.rs`, `src/ops/bundle.rs`, `build.rs`, `templates/bundle-packs/**`, `tests/bundle_user_assets_cli.rs`, this SPEC folder
- read-only: the Go repository (survey reference), `src/help.rs` (reserved by another card — the curated help table's bundle line stays as it is)
- forbidden: changing the five existing bundle actions, the divergence-preserved path, the real HOME, the installed shim

## Plan

1. Assets: copy the optional-pack entries from the Go template tree into `templates/bundle-packs/` and generate the packs data file (names, descriptions, dependencies, entry paths and hashes) plus the base name set with a one-off script run from the worktree's `target/` (the script is not committed).
2. `build.rs`: walk `templates/bundle-packs/` and emit the embedded tree as static bytes (no new dependency).
3. RED: `tests/bundle_user_assets_cli.rs` — the three subcommands do not exist yet, so every case fails first.
4. `src/user_assets.rs`: manifest read/write in the Go JSON shape, install with the collision rule, remove with the complement and dependency-deferral rules, the flock guard, and the journal.
5. `src/ops/bundle.rs`: the `add`, `remove` and `list` actions wired to the engine; `src/lib.rs`: the module line.
6. GREEN, then AC-6 neighbours, then this Result.

## Risks

- Asset drift: the copied entries are a one-off snapshot of the Go catalog; the packs data file records each entry's sha256 so a later re-sync is checkable.
- Result-count subtleties: the counts are pinned to the design-note cases (install, refresh, collision), not to every Go edge; a case outside them prints the summary line with its counts and a reason line.
- The lock is advisory: a non-cooperating writer ignores it, exactly as in Go.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D1 | The pack files and catalog data are copied into this repository and embedded at build time | Rust ships alone; the binary must install without a Go checkout | Re-embed from a shared asset source |
| D2 | Collision-only faithful subset; divergence preservation stays a non-goal | The design-note cases pin add, remove and list only; the collision path keeps user bytes | Port the divergence backup flow |
| D3 | Manifest, guard and journal keep the Go paths and JSON shapes | Users switch binaries without losing state (card body: same asset file) | Change the shapes and break state |
| D4 | The three commands join the existing `bundle` command as new actions | `help --all` must list them under bundle; the word must not fork | A separate top-level command |

## Result

구현 에이전트가 채워요 — 이 커밋은 SPEC 과 계획까지예요.

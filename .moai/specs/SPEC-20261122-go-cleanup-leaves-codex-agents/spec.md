---
id: SPEC-20261122-go-cleanup-leaves-codex-agents
title: "the Go-era cleanup leaves the Codex and shared agents folders alone"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-remove-codex-project-layer
---

# the Go-era cleanup leaves the Codex and shared agents folders alone

## Intent

The user decided on 2026-11-22, after the audit of the home update slice, that the Go-era update cleanup must not touch `~/.codex` or `~/.agents` at all: old moai files there get the one "no longer managed" notice of the home update and the user removes them if wanted. Today `cleanup-plan` and `cleanup-apply` still move proven Go agents out of `~/.codex/agents` and proven Go skill folders (`moai-*`) out of `~/.agents/skills` into the archive. This change takes both folders out of the cleanup, so that no moai command moves, edits or deletes anything under them. It replaces the keep-list entry for `update_cleanup.rs` in SPEC-20261121-remove-codex-support (D-15).

## Requirements

- REQ-1: The cleanup plan shall list no item whose path is under `.codex` or `.agents`, so that the plan neither archives nor keeps-with-a-reason anything there.
- REQ-2: When a home holds a Go agent file in its Codex agents folder that the Go manifest records with the bytes on disk, the cleanup apply shall leave that file byte for byte in place and create no archive entry for it. · failure: none
- REQ-3: When a home holds a Go skill folder named `moai-*` in its shared agents skills folder that the Go manifest proves, the cleanup apply shall leave that folder byte for byte in place. · failure: none
- REQ-4: When the Go manifest records entries under the slugs `agents-skills` or `codex-agents`, the cleanup shall read the manifest as valid and skip those entries without a finding. · failure: none
- REQ-5: The cleanup of the Claude skills and the Claude agents and the Claude projection and the moai leftovers shall plan and apply exactly as before.

## Acceptance

- AC-1 (REQ-1, REQ-5): the archive list of the plan holds only `.claude/...` and `.moai/...` paths, and no plan item path starts with `.codex` or `.agents` · verify: `cargo test --locked --offline --test update_cleanup_cli the_plan_archives_only_proven_items_and_writes_nothing`
- AC-2 (REQ-2): after `cleanup-apply` with the fresh plan's confirmation, a proven Go file in the Codex agents folder has the same bytes and no archive folder holds it · verify: `cargo test --locked --offline --test update_cleanup_cli the_codex_agents_folder_stays_byte_identical_after_the_cleanup`
- AC-3 (REQ-3): after `cleanup-apply`, a proven Go skill folder in the shared agents skills folder has the same bytes · verify: `cargo test --locked --offline --test update_cleanup_cli the_shared_agents_skills_folder_stays_byte_identical_after_the_cleanup`
- AC-4 (REQ-4): a manifest with `agents-skills` and `codex-agents` entries still reads as valid · verify: `cargo test --locked --offline --test update_cleanup_cli`
- AC-5 (REQ-5): every other case of the cleanup test file still passes · verify: `cargo test --locked --offline --test update_cleanup_cli`

## Scope

- write: `src/update_cleanup.rs`, `tests/update_cleanup_cli.rs`, `.moai/specs/SPEC-20261122-go-cleanup-leaves-codex-agents/spec.md`
- read-only: everything else, in particular `src/update.rs` and the home update notice
- forbidden: any file under the user's Codex or shared agents folders on this machine; the real HOME (every test runs in a temporary home)

## Plan

1. RED: add the two byte-identical cases and tighten the archive list assertion of the plan test; they fail on the tip because the plan still archives the Codex and shared agents files.
2. GREEN: drop `.agents/skills` from the sibling roots, `.codex/agents` from the agent roots and the two retired slugs from the manifest root table (the manifest read skips an unknown slug already); update the module doc.
3. Run the whole cleanup test file, clippy, rustfmt and `moai spec check`.

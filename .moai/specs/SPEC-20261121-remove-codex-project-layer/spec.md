---
id: SPEC-20261121-remove-codex-project-layer
title: "The project layer, the hook adapters, the generation and the interview's host choices serve Claude Code only; retired host entries in existing files are read and skipped"
status: draft
revision: 3
created: 2026-11-21
---

# The project layer, the hook adapters, the generation and the interview's host choices serve Claude Code only

## Intent

User decision of 2026-11-21: moai-adk stops supporting Codex (and the OMP host) and keeps Claude Code, with GLM as a model choice that runs through the Claude Code launcher. The cards and SPECs that describe Codex work stay as history, so the work can be restored; the code, tests and generated files go.

This SPEC is the project-layer slice of that removal: the host steps of `project-update-plan` (codex-hooks, codex-config, codex-skill, omp-hooks, omp-config, omp-agents), the project copy of the skill, the Codex and OMP hook adapters, the generation of `.codex/*` and of the projections for other hosts, and the host choices the interview offers. The launcher and the home update, the shipped skill text, the docs and the install scripts, and the lane, goal and leader code belong to sibling slices of the same decision. The impact report of 2026-11-21 (`~/.moai/reports/moai-adk/codex-removal-impact-2026-11-21.md`) lists what breaks per area and the decisions D1 to D9 this SPEC follows.

Nothing here deletes user data. A user's `~/.codex`, and a project's `.codex/` and `.omp/` folders, stay as they are; moai stops managing them, and the preview says so with one information step. A retired host that an existing file still names (a profile, a receipt, a generation manifest) never makes a command fail: the entry is read and skipped.

Order. A compatibility layer lands first and only adds: it filters retired hosts when a profile is read, ignores retired launcher model keys, lets revert skip retired receipt entries, drops retired manifest entries without a conflict, and moves the shared backup helpers and the read limit out of `src/hosts/codex_config.rs`. This slice deletes an API only after that layer has landed. Before then it takes the SPEC and the files that need no removed API.

## Non-goals

- The launcher (`moai codex`, launcher roles, model questions) and the home update (`~/.codex`, `~/.agents/skills`, the Codex status line): another slice.
- The shipped skill text, the docs, the install scripts and CI: another slice. This includes the host row of the interview in `skills/moai/references/internal/interview.md` and its test `tests/question_bank_hosts.rs` (the former REQ-9 and AC-9, moved to the shipped skill text slice, D-12).
- The lane-loop, lane-stop-hook, goal and leader-round Codex parts: another slice (D9).
- The compatibility layer itself: the profile filter, the retired launcher keys, the receipt skips of the home files, the manifest rule and the move of the backup helpers and the read limit.
- Deleting or rewriting any card or SPEC that describes Codex work; deleting any file under a user's Codex home, `.codex/` or `.omp/`.
- Keeping the repository's own profile files in step: they are regenerated with the new binary after the removal, together with their manifests.
- Kept on purpose (D6): the Go-generation `.codex` classification of `go_state`, the `~/.codex` handling of `update_cleanup` as read-only detection, the protection of the Codex worktree folder in `worktree_gc` and of `.codex/` in `gate`, the stored-record variants of `decision` and the slugs of `user_assets`.

## Requirements

- REQ-1: The project-update-plan preview shall list no step named codex-hooks or codex-config or codex-skill or omp-hooks or omp-config or omp-agents.
- REQ-2: A project profile that lists codex or omp under hosts shall load and plan without refusal.
- REQ-3: The project-update-plan preview shall carry one information step named retired-hosts that lists the files of a retired host and writes nothing.
- REQ-4: The project-update-plan preview and apply shall leave every file under the .codex and .omp folders unchanged.
- REQ-5: The project-skill-plan command shall filter a retired host out of its request and plan the rest.
- REQ-6: The project-skill-revert command shall skip a receipt entry that names a retired host and report the skip.
- REQ-7: The hook-dispatch command shall answer the host codex or omp with one line on standard error saying the host was removed, nothing on standard output, and exit zero without allowing or blocking anything.
- REQ-8: The generation shall write no file under the .codex folder and no projection for a host other than Claude Code.
- REQ-10: The command registry shall carry no command named codex-config-plan or codex-config-apply or codex-projection-plan or codex-projection-apply or codex-doctor or agents-project-plan or agents-project-apply.
- REQ-11: The project detection shall leave the .codex folder out of its MCP read and out of its code walk.
- REQ-12: The go-state-plan command shall keep classifying Go-generation .codex files as leave.

## Acceptance

- AC-1 (REQ-1): the preview of a project with a full profile lists the Claude steps and none of the six host steps, and the printed undo order names none of them · verify: `cargo test --locked --offline --test project_update_plan_cli no_codex_or_omp_step_is_planned`
- AC-2 (REQ-2): a profile with `hosts: [claude, codex, omp]` plans without refusal and its profile bytes are not rewritten · verify: `cargo test --locked --offline --test project_update_plan_cli a_profile_that_lists_codex_and_omp_still_loads`
- AC-3 (REQ-3): the same preview carries the retired-hosts step with action info and the paths it found, and no apply command · verify: `cargo test --locked --offline --test project_update_plan_cli the_retired_hosts_step_only_informs`
- AC-4 (REQ-4): a project with `.codex/` and `.omp/` files keeps them byte for byte through plan and apply · verify: `cargo test --locked --offline --test project_update_plan_cli the_codex_and_omp_folders_stay_byte_identical`
- AC-5 (REQ-5): a project-skill-plan request that lists claude and codex plans the claude copy only, and a request that lists only codex reads as claude · verify: `cargo test --locked --offline --test project_skill_links a_request_that_lists_codex_is_filtered_not_refused`
- AC-6 (REQ-6): a hand-written receipt with a codex entry and a claude entry reverts the claude entry, skips the codex one and reports the skip · verify: `cargo test --locked --offline --test boundary_d3_skill_revert a_receipt_that_names_codex_is_skipped_and_reported`
- AC-7 (REQ-7): `hook-dispatch --host codex` and `--host omp` print one line on standard error and nothing on standard output, exit 0 and write no decision or file · verify: `cargo test --locked --offline --test hook_dispatch_cli a_removed_host_gets_one_line_and_exit_zero`
- AC-8 (REQ-8): a generation apply writes nothing under `.codex` and the manifest names no other host's projection · verify: `cargo test --locked --offline --test generation_apply_result_cli no_codex_projection_is_written`
- AC-10 (REQ-10): `moai help` and the command registry list none of the removed commands · verify: `cargo test --locked --offline --test help_layout no_codex_command_is_listed`
- AC-11 (REQ-11): detection ignores `.codex/config.toml` MCP entries and does not walk `.codex` and `.omp` · verify: `cargo test --locked --offline --test project_detect_cli codex_folders_are_not_read_or_walked`
- AC-12 (REQ-12): the Go-generation `.codex` classification still answers leave · verify: `cargo test --locked --offline --test go_state_codex_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-remove-codex-project-layer/**`
- write: `src/project_hosts.rs`, `src/project_update.rs`, `src/project_skills.rs`, `src/project_detect.rs`, `src/lib.rs`
- write: `src/hooks.rs`, `src/hook_engine.rs`, `src/hook_guards.rs`, `src/hook_legacy.rs`, `src/hook_config.rs`, `src/update_hooks.rs`
- write: `src/generation.rs`, `src/agents.rs`, `src/omp_agents.rs`, `src/ops/codex_*.rs`, `src/ops/agents_project_*.rs`
- write: `src/help.rs`, `src/main.rs`, `schemas/protocol/v1/profile.schema.json`, `docs/cli-text-rationale.md`
- write: the tests and fixtures that call an API this slice removes, and the Codex-only test files of the project layer and the hook adapters, listed in the impact report (sections 2.3 and 2.5)
- write: the entries of `tools/repo_scans.json` and `tests/cli_text_citations.rs` that name a file this slice removes, in the same commit as that file
- read-only: `src/skills_legacy.rs` and `src/ops/skills_legacy.rs`: the legacy-skill check already scans the Claude folder only, changed by the home update slice (its REQ-8), so this slice no longer writes them
- read-only: every other file, and the `.codex/` and `.omp/` folders of any project
- read-only: `src/hosts/**` until the sibling slice that removes the shipped adapters and the compatibility layer have landed
- forbidden: the launcher and home-update sources of the other slice, the lane and goal sources, the shipped skill text, install scripts, the repository's own profile and generation files, any card or SPEC that describes Codex work

## Plan

1. SPEC first (this file), crossed with the inventory of `SPEC-20261121-remove-codex-support` and with the impact report: every row of the project-layer, hook and generation areas is a requirement here, a deleted test, or named under Risks.
2. Before the compatibility layer lands: this SPEC, and the files that need no removed API: the project skill copy (REQ-5, REQ-6) and the detection (REQ-11), with their tests.
3. After it has landed, in one commit per area so the crate compiles at each step: the host steps and `src/project_hosts.rs` with the `retired-hosts` step; the hook adapters (`hooks.rs`, `hook_engine.rs`, `hook_legacy.rs`, `hook_guards.rs`); the generation and `omp_agents.rs` with the projection and doctor commands; each with the tests that call what it removes, and with the help table, golden files and scan entries that name a removed file.
4. The codex skill host and the omp agent host go in the same commit (the skill host mapping in `update.rs` binds both).
5. The unnarrowed affected set on each commit, clippy with all targets, rustfmt, `spec check`, then the independent audit (tier 1).

## Risks

- The compatibility layer is not on the tip yet. Mitigation: nothing that deletes an API is committed before it lands; the merge order is a note on the card.
- A removal in one file breaks a test another slice owns. Mitigation: a test that calls a removed API changes in the same commit as that API.
- `src/help.rs`, `src/main.rs` and `tests/cli_text_citations.rs` are edited by sibling slices too. Mitigation: edit only the lines of the removed commands and merge the tip once on overlap.
- Files other cards hold at the time of writing: `src/project_update.rs`, `src/profile.rs`, `src/hook_config.rs`, `src/help.rs`, `src/hosts/*`, `src/lib.rs` and one test file. Mitigation: a held file is asked for through the holder's card, never taken.
- `src/agents.rs` prints the profile hosts into the generated AGENTS.md. Mitigation: after the profile filter lands the section changes once and the next plan shows an update, which is expected.
- `tests/go_state_codex_cli.rs` stays (D6); the inventory marks it for removal. Mitigation: D-5.
- The inventory matches `omp` without a word boundary. Mitigation: those counts are not quoted; `src/progress.rs` and `src/statusline/counts.rs` keep their comments.
- `hooks.rs` uses the backup helpers of `src/hosts/codex_config.rs`. Mitigation: the compatibility layer moves them first.
- `tests/project_skills_generation.rs` pins the `.agents/skills` and codex host behaviour of `src/generation.rs`. It is green now and goes red when the generation stage lands. Mitigation: that stage edits it in the same commit.
- `.moai-cli/manifests/profile-baseline.json` holds the repository profile's purpose and hosts. Mitigation: it is refreshed together with the two profile files after the regeneration.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | OMP is removed together with Codex in this slice | the removal decision names Codex and OMP; the OMP steps sit beside the Codex steps | restore the OMP steps from history |
| D-2 | a retired host in a profile is filtered out when it is read, never refused; an empty list reads as claude; the next save drops it | leader decision D1 | refuse instead, which breaks every profile that lists codex |
| D-3 | a receipt entry that names a retired host is skipped at revert and reported | leader decision D3 | revert it as before |
| D-4 | the `.codex/config.toml` MCP read of the detection is removed and the `.codex` and `.omp` skip list stays | the impact report: the skip list keeps leftover folders out of the source count | keep the read |
| D-5 | the Go-generation `.codex` classification of `go_state` and its test stay | leader decision D6; it only answers leave | remove with the migration |
| D-6 | the preview carries one information step named retired-hosts and the existing leftovers lists are not reused | tests pin exact leftovers lists | fold into leftovers |
| D-7 | an API is deleted only after the compatibility layer has landed | leader instruction; the impact report orders it so the tip stays green | none needed |
| D-8 | hook-dispatch for a removed host prints one line and exits zero | leader decision D5 | exit with an engine error |
| D-9 | the repository's own profile files are regenerated with the new binary, not edited by hand | the impact report: the manifests hold their hashes | edit by hand |
| D-10 | the retired-hosts step is always in the preview, with state skip and an empty file list when nothing is found, and it is not one of the waiting step names | a stable step list; the step reads no file content and a backup move cannot make it stale | list it only when files exist |
| D-11 | the removed-host line of hook-dispatch goes to standard error and standard output stays empty | a leftover Codex or OMP hook would read plain text on standard output as a decision | print it on standard output |
| D-12 | the interview host row and its test move to the shipped skill text slice | the SPEC's scope forbids the shipped skill text; that file needs the skill manifest regenerated, which other cards hold | widen this scope by an explicit decision |
| D-13 | the host steps and `project_hosts` change in one commit that also edits every test naming the six steps; only an add-only preparation commit comes before it | deleting the steps with the tests left behind turns the tests red | none needed |
| D-14 | `src/skills_legacy.rs` and its operation page leave this SPEC's write scope | the home update slice (REQ-8 of SPEC-20261122-codex-home-update-removed) already narrowed the legacy-skill check to the Claude folder | take them back if that slice is dropped |

## Result

Stage before the compatibility layer, on this branch: REQ-5, REQ-6 and REQ-11 are coded and tested; the rest waits for the compatibility layer (D-7).

The RED lines below are the implementing agent's report of its own runs on the sources before the change (not re-run here); the passes are this lane's own run.

- AC-5 met: `a_request_that_lists_codex_is_filtered_not_refused` failed first (left `["claude","codex"]`, right `["claude"]`) and passes now; `project_skill_links` 7 passed, `project_skill_apply_gate_cli` 10 passed, `skill_name_collision` 5 passed, `empty_parent_prune_cli` 5 passed, `project_skills_generation` 11 passed (run in this session, `cargo_exit=0`).
- AC-6 met: `a_receipt_that_names_codex_is_skipped_and_reported` failed first (the revert returned no `skipped` key) and passes now; `boundary_d3_skill_revert` 4 passed.
- AC-11 met: `codex_folders_are_not_read_or_walked` failed first (`mcp_servers` was `["docs"]`) and passes now; `project_detect_cli` 25 passed.
- Not run in this stage: AC-1 to AC-4, AC-7, AC-8 and AC-10 (they need the compatibility layer first), AC-12 (the sources of `go_state` are unchanged), the whole suite, the unnarrowed affected set, clippy by this lane (the implementing agent reported it clean; not observed here), and the independent audit.

---
id: SPEC-20261121-remove-codex-project-layer
title: "The project layer, the hook adapters, the generation and the interview's host choices serve Claude Code only; Codex and OMP entries in an existing profile load and are ignored"
status: draft
revision: 1
created: 2026-11-21
---

# The project layer, the hook adapters, the generation and the interview's host choices serve Claude Code only

## Intent

User decision of 2026-11-21: moai-adk stops supporting Codex (and the OMP host) and keeps Claude Code, with GLM as a model choice that runs through the Claude Code launcher. The cards and SPECs that describe Codex work stay as history, so the work can be restored; the code, tests and generated files go.

This SPEC is the project-layer slice of that removal: the host steps of `project-update-plan` (codex-hooks, codex-config, codex-skill, omp-hooks, omp-config, omp-agents), the project copy of the skill, the Codex and OMP hook adapters, the generation of `.codex/*` and of the projections for other hosts, and the host choices the interview offers. The launcher and the home update, the shipped skill text, the docs and the install scripts are other slices of the same decision; `SPEC-20261121-remove-codex-support` holds the inventory and marks each file with its slice, and this SPEC takes the files marked slice 2.

Nothing here deletes user data. `~/.codex` and a project's `.codex/` or `.omp/` stay as they are; moai stops managing them, and the preview says so. An existing project profile that lists `codex` or `omp` under `hosts` still loads; those entries are ignored with a note, so no project is locked out by the upgrade.

## Non-goals

- The launcher (`moai codex`, launcher roles, model questions) and the home update (`~/.codex`, `~/.agents/skills`, the Codex status line): another slice.
- The shipped skill text, the docs, the install scripts and CI: another slice. Where a test pins skill text that this SPEC's behaviour changes, the test is edited here only for the lines it pins.
- Deleting or rewriting any card or SPEC that describes Codex work; deleting any file under `~/.codex`, `.codex/` or `.omp/`.
- The Go-generation migration's read-only classification of `.codex` files (`go_state`): it answers `leave` and stays.
- Lane-loop and leader-round wording that mentions Codex: the slice that owns the lane tooling.

## Requirements

- REQ-1: The project-update-plan preview shall list no step named codex-hooks or codex-config or codex-skill or omp-hooks or omp-config or omp-agents.
- REQ-2: A project profile that lists codex or omp under hosts shall load and plan without refusal.
- REQ-3: The project-update-plan preview shall name every ignored host in an ignored_hosts list with the note that moai no longer manages it.
- REQ-4: The project-update-plan preview and apply shall leave every file under the .codex and .omp folders unchanged.
- REQ-5: The project-skill-plan command shall refuse the host codex and name the host as no longer managed.
- REQ-6: The project-skill-revert command shall restore a copy that an earlier receipt recorded for codex.
- REQ-7: The hooks-plan command shall refuse the host flags codex and omp and plan only the Claude Code hook set.
- REQ-8: The generation shall write no file under the .codex folder and no projection for a host other than Claude Code.
- REQ-9: The interview's host question shall offer Claude Code only.
- REQ-10: The command registry shall carry no command named codex-config-plan or codex-config-apply or codex-projection-plan or codex-projection-apply or codex-doctor or lane-stop-hook.
- REQ-11: The project detection shall leave the .codex folder out of its MCP read and out of its code walk.
- REQ-12: The go-state-plan command shall keep classifying Go-generation .codex files as leave.

## Acceptance

- AC-1 (REQ-1): the preview of a project with a full profile lists the Claude steps and none of the six host steps, and the printed undo order names none of them · verify: `cargo test --locked --offline --test project_update_plan_cli no_codex_or_omp_step_is_planned`
- AC-2 (REQ-2): a profile with `hosts: [claude, codex, omp]` plans without refusal and its profile bytes are not rewritten · verify: `cargo test --locked --offline --test project_update_plan_cli a_profile_that_lists_codex_and_omp_still_loads`
- AC-3 (REQ-3): the same preview carries `ignored_hosts` with one entry per ignored host, each with the folder it left alone and the note · verify: `cargo test --locked --offline --test project_update_plan_cli ignored_hosts_names_codex_and_omp`
- AC-4 (REQ-4): a project with `.codex/` and `.omp/` files keeps them byte for byte through plan and apply · verify: `cargo test --locked --offline --test project_update_plan_cli the_codex_and_omp_folders_stay_byte_identical`
- AC-5 (REQ-5): a project-skill-plan request that lists codex is refused with the host named · verify: `cargo test --locked --offline --test project_skill_links codex_host_is_refused`
- AC-6 (REQ-6): a hand-written receipt that recorded a Codex copy still reverts · verify: `cargo test --locked --offline --test boundary_d3_skill_revert a_receipt_that_names_codex_still_reverts`
- AC-7 (REQ-7): `hooks-plan --host codex` and `--host omp` are refused and the Claude hook plan is unchanged · verify: `cargo test --locked --offline --test hooks_cli the_codex_and_omp_hosts_are_refused`
- AC-8 (REQ-8): a generation apply writes nothing under `.codex` and the manifest names no other host's projection · verify: `cargo test --locked --offline --test generation_apply_result_cli no_codex_projection_is_written`
- AC-9 (REQ-9): the host row of the interview offers Claude Code only and a plan with another host in the profile is still accepted · verify: `cargo test --locked --offline --test question_bank_hosts the_host_row_offers_claude_code_only`
- AC-10 (REQ-10): `moai help` and the command registry list none of the six removed commands · verify: `cargo test --locked --offline --test help_layout no_codex_command_is_listed`
- AC-11 (REQ-11): detection ignores `.codex/config.toml` MCP entries and does not walk `.codex` and `.omp` · verify: `cargo test --locked --offline --test project_detect_cli codex_folders_are_not_read_or_walked`
- AC-12 (REQ-12): the Go-generation `.codex` classification still answers leave · verify: `cargo test --locked --offline --test go_state_codex_cli`

## Scope

- write: `.moai/specs/SPEC-20261121-remove-codex-project-layer/**`
- write: `src/project_hosts.rs`, `src/project_update.rs`, `src/project_skills.rs`, `src/project_detect.rs`, `src/profile.rs`, `src/lib.rs`
- write: `src/hooks.rs`, `src/hook_engine.rs`, `src/hook_guards.rs`, `src/hook_legacy.rs`, `src/hook_config.rs`, `src/update_hooks.rs`
- write: `src/generation.rs`, `src/agents.rs`, `src/agent_defs.rs`, `src/agent_install.rs`, `src/omp_agents.rs`, `src/hosts/**`, `src/ops/codex_*.rs`, `src/ops/agents_project_*.rs`
- write: `src/model_effort_questions.rs`, `src/model_effort_settings.rs`, `src/ui_text.rs`, `src/ui_summary.rs`, `src/help.rs`, `src/main.rs`, `src/go_state.rs` (one constant)
- write: the tests and fixtures the inventory of `SPEC-20261121-remove-codex-support` marks slice 2, and `docs/cli-text-rationale.md` (the `src/project_hosts.rs` section only)
- read-only: every other file, and the `.codex/` and `.omp/` folders of any project
- forbidden: the launcher and home-update sources of the other slice, the shipped skill text, install scripts, any card or SPEC that describes Codex work

## Plan

1. SPEC first (this file), crossed with the inventory of `SPEC-20261121-remove-codex-support`: every slice-2 row is either a requirement here, a deleted test, or named under Risks.
2. RED: the new tests of the acceptance lines, run first and shown to fail for the right reason.
3. Compatibility before deletion: `profile.rs` keeps loading legacy hosts and gains the managed and ignored views; `go_state.rs` stops reading a constant of `src/hosts/`.
4. Project layer: drop the six host steps, add `ignored_hosts`, narrow the project skill copy to Claude Code with a revert-only lookup for the Codex receipt, drop the `.codex` MCP read.
5. Hook adapters and generation: remove the Codex and OMP paths of the hook commands, the `.codex` generation, the projection commands and the `lane-stop-hook` command; keep the Claude hook behaviour.
6. Question bank and help: the host row offers Claude Code only; the help table and the command registry drop the removed commands.
7. Delete the tests and fixtures that only exercised removed code; edit the ones that also pin Claude behaviour.
8. The unnarrowed affected set on each commit, clippy, rustfmt, `spec check`, then the independent audit (tier 1).

## Risks

- A removal in one file breaks a test another slice owns. Mitigation: each commit runs the affected set; a test that pins removed behaviour is edited in the same commit as the code it pins.
- `src/help.rs`, `src/main.rs` and `tests/cli_text_citations.rs` are edited by other slices too. Mitigation: edit only the lines of the removed commands and merge the tip once on overlap.
- `src/agents.rs` reads `profile.hosts` raw and prints it into the generated AGENTS.md. Mitigation: switch it to the managed view in the same commit that adds the view.
- A profile that lists only codex has an empty managed view. Mitigation: the hooks step goes by the Claude flag, not by the list, and the preview names the ignored host.
- Dropping the `.codex/config.toml` MCP read changes `detected.mcp_servers` once for a project that had a Codex MCP config. Mitigation: D-4, reversible.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | OMP is removed together with Codex in this slice | the card body of the removal names "Codex (and OMP)"; the OMP steps sit in the same module as the Codex steps | restore the OMP steps from history |
| D-2 | a profile that lists codex or omp loads; the entries are ignored with a note and the file bytes are not rewritten | the user decision keeps existing projects usable; rewriting a user's profile needs its own decision | add a profile migration later |
| D-3 | a project-skill-plan request that lists codex is refused, and revert of an old Codex receipt keeps working | a refusal says what changed; a revert must still undo what an older moai wrote | accept and drop with a warning |
| D-4 | the `.codex/config.toml` MCP read is removed from detection | moai no longer manages Codex files, and reading one for Claude's profile would still be a Codex dependency | keep the read |
| D-5 | the Go-generation `.codex` classification in `go_state` stays | it only answers leave and protects user files | remove with the migration itself |
| D-6 | `ignored_hosts` is a new top-level key of the preview and the existing `leftovers` lists are not reused | tests pin exact `leftovers` lists for fixtures that list codex and omp | fold into `leftovers` later |

## Result

Not started: SPEC only.

---
id: SPEC-20261121-retired-hosts-read-safely
title: "Hosts that are being retired are read safely, and the shared backup code leaves the Codex module"
status: draft
revision: 1
created: 2026-11-21
---

# Hosts that are being retired are read safely, and the shared backup code leaves the Codex module

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only; support for the hosts `codex` and `omp` is withdrawn. The removal itself comes in later changes. This change goes first and only adds: it lets every place that reads a file written by an earlier version accept what that file says about a retired host, so the later removal cannot make a user's own machine refuse to start.

An impact study of the tip (the files and lines it names are the basis of the Scope below) found four places where a user's existing files would stop working the moment a host leaves its list:

- A saved `launcher.models.codex` (or `.omp`) value in the launcher settings: the model settings refuse the whole file, which stops `moai cc`, `moai glm` and the model step of `moai update`, the very step that could fix it.
- A project profile whose `hosts` names `codex` or `omp`: the profile check refuses it, and nearly every project command reads the profile.
- An old receipt that names a retired host: five reverts refuse the whole receipt or reach an `unreachable!` once the host variant is gone.
- A generation manifest that records Codex entries (project skills, native roles, skill projections): the plan turns each into a conflict and no code ever removes them.

The layer must change nothing while `codex` and `omp` are still active hosts. So a host counts as retired only when it is in one fixed list and is not in the active list of the place that asks; today the second condition is always false, and the removal commits switch the layer on by deleting the host from the active list, with nothing to remember. Each place takes its active list as an argument of a small seam function, so the tests can switch the layer on now.

The same change moves the backup code that the hooks and the Go-era state preview share (`Backup`, `BACKUP_DIR`, `write_backup`, `CONFIG_LIMIT`) out of the Codex config module into its own module, so that module can be deleted later. This is a pure move.

## Non-goals

- Removing any host, command, test, skill text or document (the later changes do that).
- Deleting a file the user owns (`~/.codex`, `.codex`, `.omp`, `.agents`): the layer only stops reading and recording them.
- The "no longer managed" notice and the removal of `hook-dispatch --host codex|omp`.
- Changing the profile schema or the protocol schema files.

## Requirements

- REQ-1: The system shall name the retired hosts `codex` and `omp` in one module and call a host retired only when it is in that list and not in the active list passed by the place that asks.
- REQ-2: When a project profile is read, the profile reader shall remove retired hosts from `hosts` and read an empty result as `["claude"]`. · failure: none — removing a name from a list cannot fail
- REQ-3: When the input of a profile plan names a retired host, the profile plan shall filter it like a read instead of refusing it. · failure: none — filtering is not an error
- REQ-4: When a saved profile holds a retired host and the filtered profile differs from it, the profile plan shall propose the filtered `hosts` as an update, never as a conflict, whatever the recorded baseline holds. · failure: none — an update proposal is the safe path
- REQ-5: When the saved launcher settings hold a `launcher.models` key of a retired tool, the model settings shall skip that key on read, whatever its value, and keep refusing every other unknown tool name. · failure: none — a skipped key is not an error
- REQ-6: When a model choice is saved, the model settings shall leave the keys of retired tools out of the written `launcher.models`. · failure: none — leaving a key out cannot fail
- REQ-7: When a revert reads a receipt entry of a retired host, the reverts of home routing, home style, agent files, hook files and project skills shall skip that entry, report it as skipped and revert the other entries. · failure: none — skipping an entry is reported, not an error
- REQ-8: When the generation plan or apply meets a recorded project-skill, native-role or skill-projection entry of a retired host, the generation code shall plan no conflict and no removal for it, leave its file alone and drop the entry from the manifest. · failure: none — dropping a manifest entry touches no file
- REQ-9: The backup code that hooks and the Go-era state preview share shall live in `src/config_backup.rs` and keep its values, fields and behavior; it is `Backup`, `BACKUP_DIR`, `write_backup` and `CONFIG_LIMIT`.
- REQ-10: While `codex` and `omp` stay in the active lists, the layer shall change no output of any command.

## Acceptance

- AC-1 (REQ-1, REQ-10): the retired list is `codex` and `omp`; `retired` is false for every host while the active list holds it and true once it does not; the existing profile tests pass unchanged · verify: `cargo test --test retired_hosts_layer && cargo test --test profile_plan_cli --test profile_apply_engine --test profile_scopes_cli --test placeholder_profile_guard`
- AC-2 (REQ-2, REQ-3): with an active list of `claude` only, `["claude","codex"]` and `["codex","omp"]` and `["omp"]` read as `["claude"]`; with the real active list the input is returned as it is; a plan input naming `codex` is filtered · verify: `cargo test --test retired_hosts_layer`
- AC-3 (REQ-4): a profile file holding `claude` and `codex` gets the filtered `hosts` proposed as an update (not a conflict) when the recorded baseline is absent, equal to the file, or hand-edited to `claude` only; applying it writes `hosts` and nothing else · verify: `cargo test --test retired_hosts_layer`
- AC-4 (REQ-5): with an active tool list of `claude` and `glm`, a `launcher.models` holding `codex` and `omp` keys of any shape (valid, malformed, a secret-looking text) reads without error and resolves a Claude choice; an unknown name such as a secret-looking key is still refused by key path only; nothing is listed as unsupported for the skipped keys · verify: `cargo test --test model_effort_settings --test retired_hosts_layer`
- AC-5 (REQ-6): saving a changed choice leaves the retired keys out of the written file (with the real tool list nothing is left out); the models-step receipt still restores the previous file; saving the same value again changes no file · verify: `cargo test --test update_model_pick --test models_revert_cli --test launcher_settings_receipt --test update_always_asks --test retired_hosts_layer`
- AC-6 (REQ-7): each of the five reverts, given a receipt with an entry of a retired host and one of an active host, skips the first with a `skipped` row (reason `host_retired`) and reverts the second; with the real active lists a receipt that names `codex` is reverted as before · verify: `cargo test --test retired_hosts_layer && cargo test --test home_routing_cli --test home_style_cli --test agents_home_cli --test update_hooks_cli --test project_skill_links --test boundary_d3_skill_revert --test revert_twice_cli --test project_skill_apply_gate_cli`
- AC-7 (REQ-8): with an active list of `claude` only, a manifest holding `codex` project-skill, native-role and skill-projection entries plans no conflict and no removal, leaves the files, and writes a manifest without those entries; with the real list the existing conflict behavior stays · verify: `cargo test --test project_skills_generation --test generation_apply_result_cli --test generation_manifest_single_write --test generation_receipt_order --test agents_apply_linkage --test retired_hosts_layer`
- AC-8 (REQ-9): the four items are defined in `src/config_backup.rs` and nowhere else, and the tests of the code that uses them behave as before · verify: `cargo test --test codex_config --test codex_config_cli --test codex_conflict_choices_cli --test hooks_conflict_choices_cli --test go_state_codex_cli --test hooks_cli`
- AC-9 (REQ-10): the hosts that are active today keep working end to end · verify: `cargo test --test host_home_cli --test project_update_plan_cli --test update_command_cli --test question_bank_hosts`
- AC-10 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5, REQ-6, REQ-7, REQ-8, REQ-9, REQ-10): the whole library and the binaries build and pass, the SPEC checks clean · verify: `cargo test --lib && cargo test --bins && moai-cli spec check SPEC-20261121-retired-hosts-read-safely`

## Scope

- write: `src/retired_hosts.rs` (new), `src/config_backup.rs` (new), `src/lib.rs`, `src/profile.rs`, `src/model_effort_settings.rs`, `src/home_routing.rs`, `src/home_style.rs`, `src/agent_install.rs`, `src/update_hooks.rs`, `src/project_skills.rs`, `src/generation.rs`, `src/agents.rs`, `src/hosts/codex_config.rs`, `src/hooks.rs`, `src/go_state.rs`, a new `tests/retired_hosts_layer.rs`, and this SPEC folder
- files held by other changes at the time of writing, requested by pair lines before the first edit: any of the files above that another card holds
- read-only: every other file; the tests named in the acceptance lines are run, not edited, except a comment of `tests/go_state_codex_cli.rs`
- forbidden: the Go repository, the real HOME, any file under the user's Codex, OMP or agents folders, any card number in a file name, SPEC id, branch or commit

## Plan

Tests first (RED): the new `tests/retired_hosts_layer.rs` and the `#[cfg(test)]` modules of the seam functions call the new functions with an active list that leaves the retired hosts out; they fail on the tip because the functions do not exist. The existing test targets of the acceptance lines are the guard that today's behavior is unchanged.

Order of the change, one commit group each, the full target list of that group run before the next: (1) the pure move of the backup code (`src/config_backup.rs`); (2) `src/retired_hosts.rs` with `RETIRED_HOSTS`, `retired`, `retire_hosts` and `owned_entry_retired`; (3) the profile reader, the plan input filter and the 3-way exception for `hosts` (like the existing `schema_version` exception of `src/profile.rs`: when `desired == retire_hosts(current)` and `current` holds a retired host, the key is an apply); (4) the model settings: `check_structure`, `validate` and `unsupported_entries` get `*_in` versions that take the active tool list, and `with_choice` leaves retired keys out; (5) the five reverts, each split into a seam that takes the active host names, with the skip at the very top of the entry loop (before any call that would reach `Host::from_name(..).unwrap_or_else(|| unreachable!())`); (6) the generation code (`project_skill` removal loop, `record_generation`, `agents::carried`).

## Risks

- The reverts of home style and agent files reach `unreachable!` once `Host::Codex` is gone; the skip must be the first statement of each entry loop, and a test per revert proves it with an active list that omits the host.
- `record_generation` and `agents::carried` both write the manifest; filtering only one leaves the entries in the other path.
- A retired key is deleted only when the saved value changes (the models step ends as unchanged when the same value is picked again); this is the meaning of "the next save" and AC-5 pins it.
- The project-skill revert result gains a `skipped` key; no reader of that result is known, so AC-6 runs every project-skill test target and the SPEC records the added key.
- A test that counts module files or names in a repository-wide scan could notice the two new files; AC-10 runs the lib and bin tests and the scans of the local check.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | A host is retired when it is in `RETIRED_HOSTS` and not in the active list of the asking place | Research of the tip: the check is false while the hosts are active, so behavior is unchanged, and the removal commits switch it on by editing the active lists | Make `retired` return false |
| D-2 | One common retired list (`codex`, `omp`) for all five reverts, not one list per module | Simpler; the only observable change is that a hand-made `omp` receipt of home routing, home style or project skills (which no writer produces) becomes a skip instead of a refusal | Pass a per-module list to the seams |
| D-3 | The plan input of a profile is filtered like a read | Leader decision of 2026-11-21 (D1: filtered on read, never refused); an old skill that hands back `hosts: ["codex"]` is not refused | Remove the filter call in the plan input |
| D-4 | A retired key leaves `launcher.models` when the saved value changes, not when the same value is chosen again | Smallest change point (`with_choice`); the receipt restores the file | Also write on an unchanged step |
| D-5 | The manifest entry drop covers three kinds: project-skill, native-role and skill-projection | Research of the generation code found three recorded kinds, not two | Drop only the kinds named by the leader |

## Result

(not yet written)

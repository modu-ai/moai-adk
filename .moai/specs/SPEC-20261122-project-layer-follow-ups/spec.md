---
id: SPEC-20261122-project-layer-follow-ups
title: "The project layer's follow-ups: one retired reason, a filtered host list, the remove-yourself notice, current texts and resolver paths in tests"
status: draft
revision: 1
created: 2026-11-22
parent: SPEC-20261121-remove-codex-project-layer
---

# The project layer's follow-ups: one retired reason, a filtered host list, the remove-yourself notice, current texts and resolver paths in tests

## Intent

The project layer landed (288567506e) with non-blocking audit findings: the project skill revert
gives a retired host a reason of its own while the other places say `host_retired`; a skill request
that lists `claude`, `codex` and `omp` is refused although retired hosts are filtered; the
`retired-hosts` step does not tell the user to remove the files; some texts still describe Codex
code or a Codex-only write outcome; and the Go cleanup's module text and operation page break or
contradict the user decision of 2026-11-22. Tests the slice added also write the state folder name
as a literal where the product's own folder is meant (leader rule 2026-11-22: tests build those
paths with `state_root::dir` or `home_root::path`). This slice fixes them.

## Non-goals

- Old-layout fixtures keep their literal `.moai-cli` (leader decision 6588): `tests/hooks_cli.rs`
  retired manifests, `tests/retired_hosts_layer.rs`, `tests/update_hooks_cli.rs`.
- Literal lines in files other cards hold go to those cards or to the C-199 integration merge
  (leader decision 6588): `tests/project_update_plan_cli.rs:1918` and
  `tests/update_end_to_end_cli.rs:519` (held by C-328), `tests/generation_apply_result_cli.rs:147-153`
  (held by C-308).
- `skills/moai/workflows/update.md:24,28` (held by C-328; the shipped-text card C-292 takes it) and
  `docs/codex-adapter-contract.md:86` (C-292 deletes the whole page).
- The kept protection that refuses a project inside `CODEX_HOME` (`src/hooks.rs` loop over
  `Host::Codex`) stays; only its comment changes.

## Requirements

- REQ-1: The project skill revert shall report a skipped retired-host entry with the reason `host_retired`.
- REQ-2: When a project skill request lists retired hosts beside `claude`, the plan shall filter them out instead of refusing the request. · failure: none — an unknown host is still refused by the existing check
- REQ-3: The `retired-hosts` step reason shall tell the user to remove the listed files themselves if they do not use them.
- REQ-4: The hook capabilities limits shall describe only write outcomes Claude Code can report.
- REQ-5: The hook comments shall name no removed Codex module or Codex-only outcome.
- REQ-6: The module text of the Go cleanup shall state the retired-host folder rule below its candidate list.
- REQ-7: The cleanup operation page shall name no retired-host home folder as a cleanup source.
- REQ-8: The tests this slice owns shall build the product's own state paths for a fresh project with the state-folder resolver.
- REQ-9: The project-layer SPEC Result shall record the read-only files its branch edited.
- REQ-10: The Korean "no longer managed" rows of the update screen shall come from the UI text table.

## Acceptance

- AC-1 (REQ-1): the revert of a receipt that names codex reports `{"host": "codex", "reason": "host_retired"}` · verify: `cargo test --locked --offline --test boundary_d3_skill_revert a_receipt_that_names_codex_is_skipped_and_reported`
- AC-2 (REQ-2): a plan for `["claude", "codex", "omp"]` succeeds and lists `claude` only, and a plan for `["claude", "gemini"]` is still refused · verify: `cargo test --locked --offline --test project_layer_follow_ups a_request_with_both_retired_hosts_is_filtered`
- AC-3 (REQ-3): the step reason holds `remove them yourself if you do not use them` · verify: `cargo test --locked --offline --test project_layer_follow_ups the_retired_hosts_step_tells_the_user_to_remove_the_files`
- AC-4 (REQ-4): no limit line of the Stop rule names an unknown outcome · verify: `cargo test --locked --offline --test project_layer_follow_ups the_stop_rule_limits_name_no_unknown_outcome`
- AC-5 (REQ-5, REQ-6, REQ-7): the named texts hold none of the stale words · verify: `cargo test --locked --offline --test project_layer_follow_ups the_follow_up_texts_are_current`
- AC-6 (REQ-8): the owned test files build the state folder through the resolver and still pass · verify: `cargo test --locked --offline --test boundary_d3_skill_revert --test hooks_cli --test lane_stop_hook_cli --test project_skills_generation`
- AC-7 (REQ-9): the parent Result names `src/go_state.rs` and `src/hosts/config_backup.rs` as read-only files edited · verify: `cargo test --locked --offline --test project_layer_follow_ups the_parent_result_records_the_read_only_files`
- AC-8 (REQ-10): the update screen's Korean rows for codex and omp read from `ui_text` · verify: `cargo test --locked --offline --lib update_tux`

## Scope

- write: this SPEC folder and `.moai/specs/SPEC-20261121-remove-codex-project-layer/**` (its Result)
- write: `src/project_skills.rs`, `src/project_hosts.rs`, `src/hook_engine.rs`, `src/hooks.rs`, `src/update_cleanup.rs`, `src/update_tux.rs`, `docs/design-review/appendix/operations/cleanup_plan.md`
- write: `tests/boundary_d3_skill_revert.rs`, `tests/hooks_cli.rs`, `tests/lane_stop_hook_cli.rs`, `tests/project_skills_generation.rs`, the new `tests/project_layer_follow_ups.rs`
- write after C-316 lands (it holds the file): `src/ui_text.rs` for AC-8
- read-only: everything else

## Plan

1. RED: the new `tests/project_layer_follow_ups.rs` (AC-2, AC-3, AC-4, AC-5, AC-7) and the changed reason in `boundary_d3_skill_revert` (AC-1); each fails by assertion on the tip.
2. GREEN code: `project_skills.rs` uses `crate::retired_hosts::retired` with the active list `["claude"]` and the reason `host_retired`, and counts only managed hosts against the two-entry limit; `project_hosts.rs` adds the remove-yourself sentence.
3. GREEN text: the limit line and the Stop-rule comment in `hook_engine.rs`, the guard comment in `hooks.rs`, the comment in `tests/hooks_cli.rs`, the module text of `update_cleanup.rs`, the operation page, and the parent Result.
4. Resolver lines: `boundary_d3_skill_revert.rs:64`, `hooks_cli.rs:570`, `lane_stop_hook_cli.rs:26`, the doc line of `project_skills_generation.rs:3`.
5. After C-316 lands: move the two Korean rows of `update_tux.rs:715-716` into `ui_text` (AC-8).

## Risks

- `RETIRED_HOSTS` loses `omp` only if a later card says so; the filter follows the shared list, so a request with `omp` would then be refused as unknown. AC-2 names the case so that change is seen.
- A user script that matched the old reason string `retired host; not managed` sees `host_retired`; the other four places already use it.

## Decisions

- D-1: The new behaviour tests go into one new test file instead of files other cards hold (planning default).
- D-2: N6 of the first audit was the landing work (goldens, fixture, `host_home_cli.rs`), done in the merge ae7cf4185f; nothing is left of it (planning default, from the audit text).
- D-3: AC-8 waits for C-316 to land and give `src/ui_text.rs` back; the slice can go to audit without it if the leader says so (planning default).

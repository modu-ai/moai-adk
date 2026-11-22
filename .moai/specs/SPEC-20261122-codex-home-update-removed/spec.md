---
id: SPEC-20261122-codex-home-update-removed
title: "The home update stops writing Codex and OMP files; one notice says so and an old receipt of those hosts reverts as skipped"
status: draft
revision: 1
created: 2026-11-22
---

# The home update stops writing Codex and OMP files; one notice says so and an old receipt of those hosts reverts as skipped

## Intent

The user decided on 2026-11-21 that moai-adk supports Claude Code and GLM only (GLM runs through the Claude Code launcher), and on 2026-11-22 that the next local build is installed only after no part of moai writes Codex or OMP files any more. SPEC-20261121-remove-codex-support (slice 1b of its plan, writers W-1 to W-5) names the home update as the first of those writers. This change removes them:

- W-1, the skill copy to `~/.agents/skills/moai` that the Codex and OMP agents read (`HOSTS` and `skill_host_of` of `src/update.rs`);
- W-2, the routing block `moai-routing` in `~/.codex/AGENTS.md` (`src/home_routing.rs`);
- W-3, the output style block `moai-output-style` in `~/.codex/AGENTS.md` (`src/home_style.rs`);
- W-4, the base agents in `~/.codex/agents/` (`src/agent_defs.rs`, `src/agent_install.rs`);
- W-5, the base agents in `~/.omp/agent/agents/` (same files).

Two things make this more than a deletion, and the design follows them:

- The compatibility layer of SPEC-20261121-retired-hosts-read-safely is already on the tip: every place that reads a receipt, a manifest or a profile treats a host as retired when it is in `RETIRED_HOSTS` and not in the active list of that place, and skips it with the reason `host_retired`. Today each place passes its own host list, which still holds `codex` and `omp`, so the rule is off. Deleting the two hosts from those lists switches it on. An old receipt that names Codex then reverts its Claude entries and reports the Codex entries as skipped, never as an error, and never touches the Codex file.
- A person may have Codex or OMP files from earlier updates. moai does not delete user data (decision of 2026-11-21): the files stay byte for byte, and the preview says in one line that moai no longer manages them. The notice is shown on the preview and on the apply result and is not part of the confirmed plan, so it never changes the digest the person confirms.

The same decision made two tests red on the tip before this change (SPEC-20261121-remove-codex-support, D-22): `tests/retired_hosts_layer.rs` still expects a Codex key to survive in the "tools active" case after the launcher change removed the Codex row of the model table. They are repaired first, test first, and are the first commit.

## Non-goals

- The project layer: `moai update` for a project, project hooks, `.codex` folders of a project, the Codex doctor and projection commands, `src/hosts/*`, `src/hooks.rs`, the profile (`hosts` list), `host_home::Host::Codex` (slice 2 of the plan).
- The OMP launcher, its tests and the installer texts (slice 1c).
- Lane, goal and leader code (its own SPEC).
- Deleting, moving or editing any file under `~/.codex`, `~/.agents`, `~/.omp` or a project `.codex/`.
- Claude Code and GLM behaviour: every Claude and GLM assertion stays and stays green.

## Requirements

- REQ-1: The home update shall plan the skill copy for the host `claude` only, so that no plan row, scope entry or receipt names `codex`.
- REQ-2: The home update shall plan the base agents, the routing section and the output style for the host `claude` only.
- REQ-3: When the home update applies, it shall write no file in the Codex, shared agents or OMP folders of the home. · failure: none
- REQ-4: When the Codex, shared agents or OMP folder exists in the home, the preview and the apply result shall carry one `skipped` notice per host that says moai no longer manages those files and leaves them. · failure: none
- REQ-5: When none of those folders exists, the update shall carry no such notice. · failure: none
- REQ-6: The notice shall stay outside `plan_digest`, so that a home with and a home without the folders confirm the same digest.
- REQ-7: When an old receipt of routing, style or agents names `codex` or `omp`, the revert shall skip those entries with the reason `host_retired` and revert the Claude entries. · failure: none
- REQ-8: The `skills-legacy` check shall scan the Claude skill folder only and print no Codex config entry.
- REQ-9: The three home plan commands shall name only Claude targets, so that `agents-home-plan`, `home-routing-plan` and `home-style-plan` list no other host.
- REQ-10: The update screens shall name no Codex host in a row, a hint or a restart line.
- REQ-11: The two red tests of the "tools active" model case shall pass with the expectations of the launcher change, in the first commit. · failure: none

## Acceptance

- AC-1 (REQ-1, REQ-2): the user layer plan holds hosts `claude` for skills, agents, routing and style (style: the two Claude targets) · verify: `cargo test --locked --offline --test update_user_layer_cli`
- AC-2 (REQ-3): an apply over a home that holds Codex, shared agents and OMP files leaves every one of them byte for byte and creates nothing there · verify: `cargo test --locked --offline --test codex_home_update_removed apply_leaves_the_codex_and_omp_folders_as_they_are`
- AC-3 (REQ-4, REQ-5): the notice appears once per host with files, in the preview and in the apply result, and not in a home without them · verify: `cargo test --locked --offline --test codex_home_update_removed` (cases `a_home_with_codex_files_shows_one_notice_for_codex`, `a_home_with_omp_files_shows_one_notice_for_omp`, `a_home_without_them_shows_no_notice`, `the_apply_result_carries_the_notice`)
- AC-4 (REQ-6): the plan digest is the same with and without the folders · verify: `cargo test --locked --offline --test codex_home_update_removed the_notice_does_not_change_the_plan_digest`
- AC-5 (REQ-7): an old receipt that names Codex reverts its Claude entries and skips the Codex entries, for routing, style and agents · verify: `cargo test --locked --offline --test codex_home_update_removed an_old_receipt_with_codex_entries_reverts_the_claude_ones` and `cargo test --locked --offline --test retired_hosts_layer receipts`
- AC-6 (REQ-8): the legacy-skill check reads `.claude/skills` and prints no `codex_disable_entries` · verify: `cargo test --locked --offline --test skills_legacy_cli`
- AC-7 (REQ-9): the three plan commands name Claude targets only · verify: `cargo test --locked --offline --test agents_home_cli --test home_routing_cli --test home_style_cli`
- AC-8 (REQ-10): the update screen goldens hold no `codex` and no `$moai` in English and Korean · verify: `cargo test --locked --offline --test update_tux_golden` and `grep -ril codex tests/fixtures/ui/en/update-*.golden tests/fixtures/ui/ko/update-*.golden | wc -l` prints 0
- AC-9 (REQ-11): the two repaired tests pass and fail again when the expectation is put back · verify: `cargo test --locked --offline --test retired_hosts_layer model_settings`
- AC-10 (REQ-1, REQ-2, REQ-7): the Claude cases of the home tests are unchanged and green · verify: `moai check affected <base>..<head>` prints `result=PASS`

## Scope

- write: `src/update.rs`, `src/home_routing.rs`, `src/home_style.rs`, `src/agent_defs.rs`, `src/agent_install.rs`, `src/skills_legacy.rs`, `src/update_tux.rs`, `src/ops/skills_legacy.rs`, `src/ops/agents_home_plan.rs`, `src/ops/home_routing_plan.rs`, `src/ops/home_style_plan.rs` and their operation pages under `docs/design-review/appendix/operations/`, the home and update tests and goldens of the card scope (the unnarrowed check added `e2e_host_fixture`, `host_home_cli`, `ui_korean_terminal_cli`, `ui_summary_terminal_cli`, `update_backup_replace`, `update_cleanup_cli`, `update_command_cli`, `update_headver`, `update_moved_receipts`, `update_unmanaged_failclosed`, `update_version_folders`, the two `home-routing-plan` goldens and the `EXEMPT` row of `tests/repo_scans_test.py`: they all expected a second host the update no longer writes), `tests/retired_hosts_layer.rs` (the two repaired model tests, and the three receipt tests of the home parts that wrote Codex and OMP files: they now add the retired entries to an applied Claude receipt, and the case that reverted both hosts while both were active is removed because no apply writes a second host any more), this SPEC folder.
- read-only: everything else, in particular `src/omp_agents.rs` and the project layer, which still call `agent_defs::render` for `omp` until the project slice.
- forbidden: any file under the user's Codex, agents or OMP folders on this machine; the real HOME (every test runs in a temporary home).

## Plan

Tests first, in four commits, each with a red record before its green change:

1. The first commit repairs the two red tests (REQ-11): the malformed value is named under `omp`, which is still an active tool, and a Codex value of any shape is skipped; the saving case expects the Codex key dropped and the OMP key kept (OMP leaves the table with the OMP launcher slice, which retires the "tools active" case together with its tests).
2. Red tests of REQ-1 to REQ-7 and REQ-9 (`tests/codex_home_update_removed.rs` and the changed pins of `update_user_layer_cli`, `home_routing_cli`, `home_style_cli`, `agents_home_cli`), then the code: the host lists become `["claude"]`, `skill_host_of` and the Codex arms go, the Codex block code of `home_routing` and `home_style` goes with its constants, `render_codex` and the `omp` entry of `agent_install` go, and the notice is built once in `user_layer_plan_in`.
3. The screens and goldens (REQ-10) and the legacy-skill check (REQ-8).
4. The SPEC Result, after the unnarrowed check.

`agent_defs::render(def, "omp")` stays until the project slice, because `src/omp_agents.rs` calls it; only the home host list drops `omp`. The notice is a `skipped` entry `{part: "unmanaged", host: "codex" | "omp", reason}`, built from the folders that exist under the canonical home, and added to the plan's `skipped` list, which the digest never covers.

## Risks

- The Codex block code is the only user of some shared helpers (`agents::section_plan` is shared with Claude and stays). A helper left without a caller fails `clippy -D warnings`; the compile and clippy run of the whole tree finds it.
- The notice reason is plain English, like the other `skipped` reasons; the screens print the localized word `skipped` before it.
- A symlinked `~/.codex` counts as existing for the notice (a link is never followed).
- Receipts written while Codex was active carry `codex` entries; REQ-7 and AC-5 cover them for the three parts. The hooks part and the project-skill revert belong to the project slice.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | One notice per host that has files, shown in the preview and in the apply result, outside the digest | decision 2 of the card plan (printed only when the folders exist) and SPEC-20261121-remove-codex-support D-3 | print always |
| D-2 | The two red tests are repaired, not retired, because the OMP launcher keeps `omp` in the tool list | the tool list is `claude`, `omp`, `glm` on the tip | retire with the OMP launcher slice |
| D-3 | `agent_defs::render` keeps the `omp` arm until the project slice | `src/omp_agents.rs` calls it | delete with the project slice |

## Result

Built tests first: the red cases of `tests/codex_home_update_removed.rs` were recorded with `moai tdd red` (9 failed, 1 passed), then the code went in, then the old tests and goldens that expected a second host were rewritten. The evidence below was read from the unnarrowed run `moai check affected c361d00634..babdc4a709` (`result=PASS`: 79 test targets, 10 python files, 43 listed scans, every step exit 0; the first run of it was `FAIL` on 71 tests of eleven files and one python test, which the second commit repaired and the same run then passed) and from the runs named per line. One commit after that run adds two cases to `tests/codex_home_update_removed.rs` (head `cb858e20b9`); its target, clippy and `git diff --check` were run on that head.

- AC-1: met — `update_user_layer_cli` 11 passed: the plan names `claude` for the skill copy, the agents, the routing and the two style targets.
- AC-2: met — `apply_leaves_the_codex_and_omp_folders_as_they_are` passes: the retired trees are byte for byte equal, only the Claude skill copy is written and no `.agents/skills/moai` appears.
- AC-3: met — the four notice cases and the apply-result case pass, and two cases were added after two mutants survived (`the_codex_folder_alone_shows_the_codex_notice`, `the_shared_agents_folder_alone_shows_the_codex_notice`); `codex_home_update_removed` 12 passed.
- AC-4: met — `the_notice_does_not_change_the_plan_digest` passes.
- AC-5: met — `an_old_receipt_with_codex_entries_reverts_the_claude_ones` passes for routing, style and agents; `retired_hosts_layer` 31 passed with the receipt cases now adding the retired entries to an applied Claude receipt.
- AC-6: met — `skills_legacy_cli` 8 passed.
- AC-7: met — `agents_home_cli` 7, `home_routing_cli` 11 and `home_style_cli` 16 passed.
- AC-8: met — `update_tux_golden` 14 passed and `grep -ril codex tests/fixtures/ui/en/update-*.golden tests/fixtures/ui/ko/update-*.golden | wc -l` printed 0.
- AC-9: met in part — the two repaired model tests pass in `retired_hosts_layer`; putting the old expectation back and reading the failure was not run in this pass.
- AC-10: met — the Claude cases of the home tests pass in the unnarrowed run (`result=PASS`).

Mutants (7 runs of one line each, `tools/mutate.py`, `target/mutation-result.json` and `target/mutation-result2.json`): dropping the OMP notice, not adding the notices to `skipped` and pointing the legacy-skill scan at a wrong folder were each caught (the target failed). Dropping either folder check of the Codex notice survived the first run, because the case created both folders; after the two single-folder cases were added both were caught on the rerun. Every mutation reads `restored: true`, but both runs ended with `status: blocked` and "source HEAD/index/tracked state/refs changed" because the local integration branch moved while they ran, so the runs are evidence per mutant, not a clean tool exit.

Not run: Windows and Linux, the independent audit, a trial merge with the branch that edits `src/home_style.rs` on its own, and the hooks part and the project-skill revert of old receipts (project slice).

---
id: SPEC-20261121-orca-claude-config-dir
title: "an Orca profile's moved Claude settings folder is followed by the update preview and the skill copy, and the adapter table says so"
status: draft
revision: 1
created: 2026-11-21
---

# an Orca profile's moved Claude settings folder is followed by the update preview and the skill copy, and the adapter table says so

## Intent

The Orca adapter text lists SETUP and UPDATE as `unsupported` with Orca: an Orca profile starts Claude Code with `CLAUDE_CONFIG_DIR` set to a folder of its own, "and `moai update` looks only at `~/.claude`, so `/moai` can be missing in that session while the preview says nothing to apply". Since then the host-folder resolver (`src/host_home.rs`) made the preview and the skill copy follow `CLAUDE_CONFIG_DIR`, and a test pins the claim at the level of one plan. The adapter table and the test that pins its words (`tests/orca_optional_3005.rs`) still say unsupported.

This change measures the whole flow in a temporary Orca-style profile (preview, apply, a second preview, and a control without the variable), writes the measurement down (`research.md`), and then updates the two table rows, the unsupported list and the test that pins them. What the measurement cannot show is written as unverified: no live Orca profile session was run, so that Claude Code in such a session reads the moved folder rests on the Claude Code page the resolver cites. The Orca step stays optional; nothing here changes a setting.

## Requirements

- REQ-1: The `update-plan` operation shall plan the Claude Code skill copy under the folder that `CLAUDE_CONFIG_DIR` names, when the value is an absolute path and the home is the account's own home, and leave the Codex copy in the `.agents/skills/moai` folder of the home.
- REQ-2: The `update-apply` operation shall write the `moai` skill with the front matter name `moai` into that folder and create nothing in the `.claude` folder of the home.
- REQ-3: The `update-plan` operation shall plan no change for Claude Code after an apply of the same package.
- REQ-4: The `update-plan` operation shall plan the Claude Code skill copy in the `.claude/skills/moai` folder of the home when `CLAUDE_CONFIG_DIR` is not set.
- REQ-5: The Orca adapter text shall say that SETUP and UPDATE work with Orca because the skill copy follows `CLAUDE_CONFIG_DIR`.
- REQ-6: The Orca adapter text shall not list the settings folder of an Orca profile among the unsupported items.
- REQ-7: The Orca adapter text shall name what the measurement did not verify: a live Orca profile session that reads the moved folder, and that Orca sets `CLAUDE_CONFIG_DIR`, which Orca's pages never name.
- REQ-8: The Orca adapter text shall say that `--reuse-session` only continues in the previous live automation terminal, that the automations page says nothing about a manual run that overlaps a running one, a restart, a sleeping machine or the missed-run grace window, and that a leader checks whether a leader already runs before it starts one by a manual run or after a restart.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): in a temporary profile with `CLAUDE_CONFIG_DIR` set, the preview names the moved folder, the apply writes `SKILL.md` there and nothing in the `.claude` folder of the home, and the second preview has nothing to apply · verify: `cargo test --locked --offline --test orca_claude_config_dir profile_`
- AC-2 (REQ-4): without the variable the skill goes to the `.claude/skills/moai` folder of the home and the profile folder stays untouched · verify: `cargo test --locked --offline --test orca_claude_config_dir without_`
- AC-3 (REQ-5, REQ-6): the SETUP and UPDATE rows of the adapter table start with `works` for Orca and name `CLAUDE_CONFIG_DIR`, and the unsupported list no longer names the settings folder · verify: `cargo test --locked --offline --test orca_claude_config_dir adapter_rows_`
- AC-4 (REQ-7): the adapter text names the unverified live session, the unverified Orca setting of the variable, and the pages they rest on · verify: `cargo test --locked --offline --test orca_claude_config_dir adapter_unverified_`
- AC-5 (REQ-8): the adapter text carries the automation reuse meaning, what the page leaves out and the leader check · verify: `cargo test --locked --offline --test orca_claude_config_dir adapter_automations_`
- AC-6 (REQ-5, REQ-6): the test that pins the adapter table follows the new rows · verify: `cargo test --locked --offline --test orca_optional_3005`

## Scope

- write: `.moai/specs/SPEC-20261121-orca-claude-config-dir/**`, `tests/orca_claude_config_dir.rs`
- write: `skills/moai/references/adapters/orca.md`, `skills/moai/references/source-manifest.json`
- write (when the file is free): `tests/orca_optional_3005.rs`

## Decisions

- D-1: The measurement uses `update-plan` and `update-apply` with a package folder because the development binary has no installed package for `moai update` itself (`no skill package installed with this binary`); both commands are the ones `moai update` runs, and the existing resolver tests use them the same way.
- D-2: The variable counts only for the account's own home, so the test sets `MOAI_ACCOUNT_HOME` to its temporary HOME. A real Orca profile session runs with the real HOME, which is the account home, so the variable counts there.
- D-3: This card changes the two rows and the unsupported list and not the rest of the table; the card on the Orca messaging rules changes other parts of the same file and the manifest is regenerated by whoever merges later.

## Plan

1. Measure and write `research.md`; write the tests (the code tests pass today and guard the claim, the text tests are red).
2. Change the two rows, the paragraph and, when its file is free, the pin test.
3. Whole-area run, clippy, rustfmt, spec check, audit, merge-ready.

## Result

- AC-1: not yet — tests not yet run.
- AC-2: not yet — tests not yet run.
- AC-3: not yet — text not yet changed.
- AC-4: not yet — text not yet changed.
- AC-5: not yet — text not yet changed.
- AC-6: not yet — the file is held by another card.

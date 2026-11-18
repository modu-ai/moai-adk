---
id: SPEC-20261118-launcher-agent-entry
title: "launcher -l/-a say how the role is entered; codex -a starts agent mode"
status: draft
revision: 1
created: 2026-11-18
---

# launcher -l/-a say how the role is entered; codex -a starts agent mode

## Intent

`-l|-a` only name a session and set its access; the MoAI role is entered inside the session with `/moai --mode-agent`
or `/moai --mode-leader` (Codex: `$moai --mode-agent` or `$moai --mode-leader`, the two spellings of #2286). The help
and the `--dry-run` notes do not say so, and on 2026-11-16 eight Codex lanes started with only
`MoAI session name: agent-N` never entered agent mode, while on 2026-11-15 Codex lanes started with
`$moai --mode-agent\nMoAI session name: agent-71\n…` did. So the help and notes say how the role is entered, and
`moai codex -a` puts the mode line first in its first prompt (user decision 2026-11-18, card note 2092).

## Non-goals

- `codex -l`, every `cc` and `glm` launch, and roleless launches keep their argv byte for byte.
- No other spelling of the mode commands; no change to how a session enters a role once it reads the line.
- The `moai --help` launcher line and its `cc -l` example are pinned by `tests/fixtures/help/root*.txt`, outside the
  widened scope; they wait for the leader's answer (D-4).

## Requirements

- REQ-1: When `moai codex -a` composes its first prompt, the launcher shall put the line `$moai --mode-agent` before the `MoAI session name: <name>` line. · failure: none — a name or prompt the launcher refuses is refused as before, with no argv
- REQ-2: When a role session is composed, the `--dry-run` notes shall name the mode command that enters the role inside the session and say that `-l|-a` only name the session and set its access. · failure: none
- REQ-3: The long help of the per-host `-l` and `-a` and the launcher long about shall name the mode commands and say that `-l|-a` only name the session and set its access.
- REQ-4: IF a launch is roleless or is `codex -l`, `cc` or `glm`, THEN the launcher shall compose the argv it composed before this change.
  · failure: none
- REQ-5: The launcher help and notes shall name no spelling of a mode command other than `/moai --mode-agent`, `/moai --mode-leader`, `$moai --mode-agent` and `$moai --mode-leader`.
- REQ-6: The Korean texts of the per-host `-l` and `-a` and of the launcher long about shall name the mode commands and say that `-l|-a` only set the session name and access.
- REQ-7: The top-level `-l` and `-a` one-line help shall name the mode command that enters the role inside the session.

## Acceptance

- AC-1 (REQ-1): `codex -a` gives `$moai --mode-agent\nMoAI session name: agent-N`, with `\n<PROMPT>` after it when given,
  and `--name x` keeps the mode line; the top-level `moai -a codex` gives the same.
  verify: `cargo test --locked --offline --test launcher_agent_entry codex_agent_prompt_starts_with_the_mode_line`
- AC-2 (REQ-2): the notes of `-l` and `-a` on `cc`, `glm` and `codex` name the matching mode command for that host and
  say that `-l|-a` only name the session and set access; roleless notes name none.
  verify: `cargo test --locked --offline --test launcher_agent_entry role_notes_say_how_the_role_is_entered`
- AC-3 (REQ-3): `moai cc --help`, `moai glm --help` and `moai codex --help` show the entry for `-l` and `-a` in their
  long help and long about.
  verify: `cargo test --locked --offline --test launcher_agent_entry the_help_says_how_the_role_is_entered`
- AC-4 (REQ-4): the argv of `codex -l`, `cc -l|-a`, `glm -l|-a` and roleless launches are unchanged.
  verify: `cargo test --locked --offline --test launcher_agent_entry other_launches_keep_their_argv`; `cargo test --locked --offline --test launcher_roleless`
- AC-5 (REQ-5): every `mode-` in the help and notes is one of the four spellings.
  verify: `cargo test --locked --offline --test launcher_agent_entry only_the_two_spellings_appear`
- AC-6 (REQ-6): with `configure_in(…, Lang::Ko)` the short and long help of `-l` and `-a` on `cc`, `glm` and `codex`
  name `/moai --mode-<role>` and `$moai --mode-<role>` and "권한만 정함", and each long about names its host's two
  commands (the Korean texts stay unused while the CLI prints English only, #2825 R1).
  verify: `cargo test --locked --offline --bin moai the_korean_launcher_help_names_the_entry_command`
- AC-7 (REQ-7): the `-l, --leader` and `-a, --agent` lines of `moai help --all` name `/moai --mode-leader` and
  `/moai --mode-agent` (and `$moai --mode-agent` for codex); the four `help --all` goldens change on those lines only.
  verify: `cargo test --locked --offline --test launcher_agent_entry the_top_level_flags_name_the_entry_command`; `cargo test --locked --offline --test help_layout`

## Scope

- write: `src/session.rs`, `src/main.rs`, `tests/launcher_agent_entry.rs`, `tests/launcher_advisor_setting.rs`,
  `tests/launcher_agent_numbering.rs`, `tests/launcher_codex_effort.rs`, `tests/launcher_permissions_setting.rs`,
  `tests/launcher_roleless.rs`, `docs/host-support-matrix.md`, `docs/user-entry-draft.md`, `docs/rust-product-path.md`,
  `.moai/specs/SPEC-20261118-launcher-agent-entry/spec.md`; widened by the leader (option A, card reservation gen 2):
  `src/help.rs`, `tests/fixtures/help/all.txt`, `all-dumb.txt`, `all-80.txt`, `all-60.txt`
- read-only: `tests/product_cli.rs` (held by another change; its two codex `-a` prompt pins are reported, not edited)
- forbidden: `tests/fixtures/help/root*.txt` and the other help goldens, the real HOME, any real host start

## Plan

1. Tests first (RED): the new `tests/launcher_agent_entry.rs`, and the codex `-a` prompt pins in the launcher tests.
2. `session::codex_argv` puts the mode line first for an agent; `compose` adds one entry note per role and host.
3. Long help of the per-host `-l|-a` and the launcher long about in `src/main.rs`.
4. Docs, checks, Result.

## Risks

- A Codex session that does not take `$moai` as its first line would show the line as text; the 2026-11-15 lanes show
  it entering agent mode, and a roleless or leader start is unchanged.
- `tests/product_cli.rs` keeps the old codex `-a` prompt until it is updated after its other change merges.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | Help and notes say the role is entered inside the session with `/moai --mode-agent`/`--mode-leader` (Codex `$moai …`); `codex -a` starts its first prompt with `$moai --mode-agent`. This replaces the Codex `-a` part of the 2026-11-10 #2359 first prompt (name line only) | User decision 2026-11-18, card C-47 note 2092; evidence: Codex lanes of 2026-11-15 entered agent mode with the line, eight lanes of 2026-11-16 without it did not | Drop the mode line and the notes |
| D-2 | Done (revised): the top-level `-l|-a` one-line help and the Korean texts name the entry too; the four `help --all` goldens were rewritten with `MOAI_HELP_GOLDEN_WRITE` (each binary run on its own temporary HOME) and change on the `-l`/`-a` lines only | Leader decision 2026-11-18, option A on the LANE-BLOCKED (card reservation gen 2 widened the scope to `src/help.rs` and `tests/fixtures/help/all*.txt`) | Restore the old texts and goldens |
| D-4 | The `moai --help` launcher line (`LEAD_OR_AGENT`) and the `cc -l` example stay as they are for now | They are pinned by `tests/fixtures/help/root*.txt` as well, which the widened scope does not name; asked as LANE-BLOCKED kind=decision (A2: widen to `root*.txt`; B2: leave them, since the root help points to `<command> --help`) | Change them once the scope covers `root*.txt` |
| D-5 | The Korean texts are checked through `configure_in(…, Lang::Ko)`, not on a terminal | The CLI prints English on every stream (`ui::ui_lang`, user decision 6878265192, #2825 R1), so no screen shows Korean today; the texts stay in step for when that changes | Test on a terminal once Korean is shown |
| D-3 | The entry note is added by the `cc|glm|codex` launchers, not by `session::compose` | The legacy `session` command also composes through it and its answer is recorded in `tests/fixtures/protocol/v1/current/18-session-dry-run.json` (REQ-00F-06: unchanged) | Move the note into `compose` and update that fixture |

## Result

Base `1d49d55062` (origin/v2026.11.0), change commit `569c38dfe9`. RED before the change: `launcher_agent_entry` 4 of 5
failed (the help lacked `/moai --mode-leader`, no entry note, the old codex `-a` prompt; `other_launches_keep_their_argv`
passed, as it pins unchanged argv), and the updated codex `-a` pins failed in `launcher_advisor_setting` (3),
`launcher_agent_numbering` (2), `launcher_codex_effort` (1), `launcher_permissions_setting` (3) and `launcher_roleless` (1).

- AC-1: met — `codex_agent_prompt_starts_with_the_mode_line` (`tests/launcher_agent_entry.rs`) and the updated pins in the five launcher tests pass.
- AC-2: met — `role_notes_say_how_the_role_is_entered` passes; mutant "Codex leader note spelled `/moai`" fails it.
- AC-3: met — `the_help_says_how_the_role_is_entered` passes for `cc`, `glm` and `codex --help`.
- AC-4: met — `other_launches_keep_their_argv` and `launcher_roleless` pass; mutant "mode line on every role session" fails it.
- AC-5: met — `only_the_two_spellings_appear` passes over the three host helps, the root help, notes and argv.
- AC-6: met — `the_korean_launcher_help_names_the_entry_command` (`src/main.rs` help tests) passes after commit `dbcd46f11a`; before the `src/help.rs` rows changed it failed on `cc -leader: 리더 세션`.
- AC-7: met — `the_top_level_flags_name_the_entry_command` passes (it failed on the old `-l, --leader` line first); `help_layout` 10 passes on the goldens rewritten with `MOAI_HELP_GOLDEN_WRITE`, whose diff is the `-l`/`-a` lines only (all.txt and all-dumb.txt 2→5 lines at :233, all-80.txt :237 and :239, all-60.txt :380 and :383; `root*.txt` unchanged).

Second slice (leader decision A, commit `dbcd46f11a`): `launcher_agent_entry` 6, `--bin moai` 28, `--lib help::` 18,
`help_layout` 10 and the nine launcher tests pass; clippy `-D warnings` and fmt exit 0; `product_cli` still fails only
`launcher_codex_argv`.

Checks of the first slice (exit 0): `launcher_agent_entry` 5, `launcher_advisor_setting` 7, `launcher_agent_numbering` 25,
`launcher_codex_effort` 5, `launcher_permissions_setting` 6, `launcher_roleless` 4, `launcher_terminal_output` 10,
`launcher_auto_init` 18, `launcher_three_hosts` 4, `help_layout` 10 (help goldens unchanged), `--bin moai` 27,
`ac_command_lint` 5, `statusline_tag_order`, `archive_privacy_suite`, `go_state_classes_cli`, `profile_apply_engine`,
fmt, clippy `-D warnings`, Python `host_matrix_test`·`ci_scope_test`·`repo_scans_test` 49 OK, `moai spec check --all`.
`product_cli`: 75 passed, `launcher_codex_argv` fails at `tests/product_cli.rs:1272` on the old pin `:1282`; the pin at
`:1298` needs the same prefix (file held by another change).

Not run: a real host start, the full `cargo test`, Linux CI. Not done: D-4 (the `moai --help` launcher line and example,
waiting on the leader).

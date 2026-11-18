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
- The top-level `moai -l|-a` one-line help, the `moai --help` launcher line and the Korean help texts live in help
  goldens and `src/help.rs`, outside this change (D-2).

## Requirements

- REQ-1: When `moai codex -a` composes its first prompt, the launcher shall put the line `$moai --mode-agent` before the `MoAI session name: <name>` line. · failure: none — a name or prompt the launcher refuses is refused as before, with no argv
- REQ-2: When a role session is composed, the `--dry-run` notes shall name the mode command that enters the role inside the session and say that `-l|-a` only name the session and set its access. · failure: none
- REQ-3: The long help of the per-host `-l` and `-a` and the launcher long about shall name the mode commands and say that `-l|-a` only name the session and set its access.
- REQ-4: IF a launch is roleless or is `codex -l`, `cc` or `glm`, THEN the launcher shall compose the argv it composed before this change.
  · failure: none
- REQ-5: The launcher help and notes shall name no spelling of a mode command other than `/moai --mode-agent`, `/moai --mode-leader`, `$moai --mode-agent` and `$moai --mode-leader`.

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

## Scope

- write: `src/session.rs`, `src/main.rs`, `tests/launcher_agent_entry.rs`, `tests/launcher_advisor_setting.rs`,
  `tests/launcher_agent_numbering.rs`, `tests/launcher_codex_effort.rs`, `tests/launcher_permissions_setting.rs`,
  `tests/launcher_roleless.rs`, `docs/host-support-matrix.md`, `docs/user-entry-draft.md`, `docs/rust-product-path.md`,
  `.moai/specs/SPEC-20261118-launcher-agent-entry/spec.md`
- read-only: `tests/product_cli.rs` (held by another change; its two codex `-a` prompt pins are reported, not edited)
- forbidden: `tests/fixtures/help/**`, `src/help.rs`, the real HOME, any real host start

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
| D-2 | The top-level `-l|-a` one-line help, the `moai --help` launcher line and the Korean texts stay as they are in this change | They are pinned by `tests/fixtures/help/all*.txt` and live in `src/help.rs`, outside the write scope; reported to the leader as LANE-BLOCKED kind=decision | Widen the scope and update them |

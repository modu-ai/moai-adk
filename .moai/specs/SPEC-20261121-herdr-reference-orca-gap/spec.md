---
id: SPEC-20261121-herdr-reference-orca-gap
title: "A read-only herdr reference, and the Orca guidance gaps against the installed orca-cli skill"
status: draft
revision: 1
created: 2026-11-21
---

# A read-only herdr reference, and the Orca guidance gaps against the installed orca-cli skill

## Intent

MoAI keeps one short reference per optional tool (`skills/moai/references/adapters/`): when to use it, the commands that print the installed version's own guide, the lines MoAI keeps, and what to do when the tool is missing. This SPEC adds herdr as a new optional reference, limited to the first and safest job: in a herdr workspace the user already has open and names, identify the sessions, workspaces, tabs, panes and agents and read their state. It types nothing, starts nothing and changes nothing, and the default flow (Claude Code with the local Rust `moai`) runs the same with herdr absent.

The reference follows the official skill and CLI of the installed version, not the web pages alone: herdr 0.8.2 is installed, the pages show 0.9.5, and 0.8.2 lacks `machine`, `--machine`, `plugin` and `workspace close --group`. No herdr server was running when this was written, so the field names come from the bundled schema and the docs, and a read in a live named workspace is the one check that stays open.

The second part is a gap report for the Orca guidance. `orca skills get orca-cli` of the installed Orca 1.4.267 says to message another agent only with `orchestration send` and never with `terminal send`, even when mail is refused. MoAI's MODE-AGENT and MODE-LEADER text still has a `terminal send` fallback. The files that carry that text are held by other work (the Orca 1.4.267 alignment of the adapter and of the messaging rules), so this SPEC lists each difference with its file and line and the official sentence, and does not edit those files.

## Non-goals

- Anything herdr can write: typing, keys, starting or closing a pane, agent, tab or workspace, renaming, moving, focusing, reporting an agent's state, installing the herdr skill or integration, changing herdr settings. The reference forbids them by name.
- A `moai` subcommand, an MCP server, a hook or a settings change for herdr.
- Reading a herdr badge as M-Flow completion. Completion is the ledger, the audit and the local merge.
- Editing `skills/moai/references/adapters/orca.md`, `skills/moai/workflows/mode-agent.md` and `skills/moai/workflows/mode-leader.md`. Other work holds them; the differences go to that work as notes.
- Verifying herdr in a live workspace. That needs a person to open a workspace and name it.

## Requirements

- REQ-1: When the herdr reference is read, it shall name the commands that print the installed version's guide (`herdr --version`, `herdr --skill`, `herdr api schema`, `herdr status --json`) and the pages it was checked against with the day each was read · failure: none — a missing line is a failing test, not a runtime state.
- REQ-2: When the herdr reference is read, it shall list the read-only commands MoAI may run and every typing, creating, closing, moving, reporting, installing and configuring command it forbids · failure: none — a missing command is a failing test.
- REQ-3: When the herdr reference is read, it shall say that a herdr state word is never M-Flow completion and that a command not on the lists is not allowed · failure: none — a missing sentence is a failing test.
- REQ-4: When herdr is missing, the herdr reference shall say to use the default flow and to record the herdr step as `not_run` with the reason · failure: none — the default flow does not call herdr.
- REQ-5: When `tests/herdr_reference_cli.rs` runs, the suite shall fail unless the reference has the five adapter headings, the allowed list, every forbidden command, the five state words, the not-run fallback, no card number and no MCP path · failure: none — the file is read, never written.
- REQ-6: When the package manifest is checked, the manifest shall hold an entry for `references/adapters/herdr.md` whose sha256 equals the file's · failure: none — the existing manifest test fails on a stale hash.
- REQ-7: When this SPEC is closed, its Result section shall list each place where MoAI's Orca text differs from `orca skills get orca-cli` of the installed Orca, with file, line, the official sentence and the proposed replacement · failure: none — the list is a report, and the edits belong to the work that holds those files.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): `herdr.md` 가 다섯 제목, 읽기 전용 목록, 금지 명령 전부, 상태 낱말 다섯, `not_run` 대체 길을 가지고 카드 번호와 MCP 길이 없어요 · verify: `cargo test --offline --locked --test herdr_reference_cli`
- AC-2 (REQ-6): 매니페스트에 `herdr.md` 항목이 있고 해시가 맞아요 · verify: `cargo test --offline --locked --test product_cli -- source_manifest_covers_every_package_file`
- AC-3 (REQ-7): Result 절에 Orca 차이 목록이 있고 항목마다 파일·줄·공식 문장·대체 문장이 있어요 · verify: `moai spec show SPEC-20261121-herdr-reference-orca-gap` 로 Result 를 읽어요.
- AC-4 (REQ-5): herdr 이 없는 환경에서도 기본 흐름의 시험이 그대로 통과해요 · verify: `cargo test --offline --locked --test orca_optional_3005`

## Scope

- write: `skills/moai/references/adapters/herdr.md`, `skills/moai/references/source-manifest.json`, `tests/herdr_reference_cli.rs`, `.moai/specs/SPEC-20261121-herdr-reference-orca-gap/**`; `tests/product_cli.rs` only for the adapter list pin, after the scope is widened
- read-only: `skills/moai/references/adapters/orca.md`, `skills/moai/workflows/mode-agent.md`, `skills/moai/workflows/mode-leader.md`, `tests/orca_optional_3005.rs`
- forbidden: the files of the other optional-tool and Orca work, herdr's own folders and settings, any herdr command that changes state, the network except reading the two herdr pages

## Plan

1. Research (done, read 2026-11-21): the installed herdr's `--skill`, `api schema`, `status` and the 0.8.2 pages; the read-only and the forbidden command sets; the error and exit code when no server runs.
2. RED: `tests/herdr_reference_cli.rs` first, failing because the file does not exist.
3. GREEN: `skills/moai/references/adapters/herdr.md`, then the manifest entry with its sha256.
4. The adapter list pin in `tests/product_cli.rs` (needs the scope) and the neighbour checks (`orca_optional_3005`, `agent_defs_package`).
5. The Orca gap list into Result, and a note to the two Orca cards.

## Risks

- The reference is written from the bundled schema and the docs, not from a live workspace. A field name or an error text that differs in a real workspace would be found only there. The reference says so in its Sources.
- A herdr CLI group prints the top-level help for `herdr <group> <command> --help` in 0.8.2, so the arguments of one command cannot be checked from the binary. The reference sends the reader to the version's CLI page.
- The docs on the web move ahead of the installed binary. The reference names the version and the commands that exist only in newer releases, so a newer page is not read as the installed truth.
- Three optional-tool references add files to one pinned adapter list in `tests/product_cli.rs`, so the three edits can conflict at merge. The leader is asked for one order.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | From outside a herdr pane, read only when the user names the session and workspace, and say that it ran from outside (reversible, category none) | The official skill says to stop from outside and not to inspect the focused session; the card asks for reading in a workspace the user names. Explicit `--session` and ids keep the read off the focused pane | Delete the exception sentence; the reference then stops from outside as the skill does |
| D-2 | Pane text may be read only when the user asked to read that pane (reversible, category none) | Pane text can hold secrets or instructions; state needs only the list and get commands | Remove the pane-read sentence |
| D-3 | The Orca edits are listed and handed over, not made here (reversible, category none) | The files are held by the Orca alignment work; the card says to coordinate instead of duplicating | Apply the list after those files are released |
| D-4 | `worktree.list` and `pane.graphics.info` are not allowed (reversible, category none) | The docs do not call them read-only and the first is judged only by its name | Add them after a documented check |

## Result

(구현 뒤에 채워요)

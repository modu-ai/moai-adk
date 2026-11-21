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

### Orca 안내와 설치된 orca-cli 공식 스킬의 차이 (REQ-7)

기준: `orca --version` = 1.4.267, `orca skills get orca-cli` 를 2026-11-21T06:20Z 에 읽었어요(399줄, 아래 "공식 L"은 그 출력의 줄 번호). MoAI 쪽 줄 번호는 이 가지의 기준 커밋(`53cce6c580`)이에요. 이 SPEC 은 이 파일들을 고치지 않아요(D-3): `adapters/orca.md` 는 다른 작업이 쥐고 있고, 두 워크플로의 메시지 규칙은 Orca 1.4.267 맞춤 작업 몫이에요.

| # | 파일:줄 | MoAI 의 지금 문장 | 공식 문장 | 바꿀 문장 |
| --- | --- | --- | --- | --- |
| G1 | `skills/moai/workflows/mode-agent.md:116` 과 `mode-leader.md:70` (Session control 7번) | 메일의 `delivered_at` 이 `null` 이고 대상 화면이 비어 있으면 같은 글을 `orca terminal send` 로 보낸다 | L103 "If mail is refused, follow the refusal or tell the user. Never fall back to `terminal send`." 그리고 L301 "Never use it to message another agent or answer its prompts, even when mail is refused; use `orchestration send`" | 그 대체 경로를 빼고 "영수증의 `recipient_delivery_held` 경고를 읽고, 그 터미널의 코디네이터에게 보내거나 사람에게 알린다. `terminal send` 로 되돌리지 않는다"로 바꿔요 |
| G2 | `mode-agent.md:110` 과 `mode-leader.md:64` (1번) | "Read before sending: `orca terminal read --screen` must show a live session waiting for input" | L298 `terminal read` 는 `terminal send` 앞에 쓰는 것이고, L301 `terminal send` 는 셸·개발 서버·REPL·자기 창용이에요. 다른 에이전트에게는 메일만 | 이 규칙을 "셸과 자기 터미널에 보내기 전"으로 좁히고, 에이전트에게는 `orchestration send` 와 `orca agent list/show/read` 로 상태를 읽는다고 적어요 |
| G3 | `skills/moai/references/adapters/orca.md:8` | "reading and sending to lane terminals" | L301 (G1 과 같아요) | "reading lane terminals and messaging lanes with `orchestration send`" |
| G4 | `adapters/orca.md:24` | "Read a terminal before sending to it, and never overwrite another session's unsubmitted input." | L298·L301 | "Read a shell or your own terminal before sending to it; message another agent only with `orchestration send`." |
| G5 | `adapters/orca.md:58` 과 `tests/product_cli.rs` 의 `"orca 1.4.257"` 고정 | 출처를 orca 1.4.257, 2026-11-13 으로 적어요 | 설치된 것은 1.4.267 | 1.4.267, 읽은 날 2026-11-21 로 바꾸고 시험의 고정 문자열도 함께 바꿔요 |
| G6 | `mode-agent.md:111` (2번), `mode-leader.md:65` | "`accepted: true` proves only that the input was accepted ... for a plain shell Orca cannot report delivery" | L302~L306 텍스트와 Enter 를 함께 보내면 영수증에 `input_accepted` 다음 `turn_started` 단계가 있고 `--wait-submit`, `--retry-request <id>` 가 있어요 | 셸 입력에만 해당함을 밝히고, 에이전트 쪽은 메일 영수증(`stored`, 읽음 아님)과 `agent read` 로 확인한다고 적어요 |
| G7 | `adapters/orca.md` (없음) | 에이전트를 어떻게 찾는지 한 줄도 없어요 | L76~L82 `agent list/show/read` 는 `orca_session_id` 로 찾고 화면이 아니라 대화 기록을 읽어요. L294~L295 터미널의 `title` 로 다시 찾지 말고 `worktree ps` 의 `agents[].paneKey` 를 `terminal list` 의 `tabId:leafId` 에 이어 핸들을 얻어요 | 어댑터의 "Lines MoAI keeps"에 두 줄로 옮겨요 |
| G8 | `adapters/orca.md:16` | `orca skills get orca-cli` 만 적어요 | L388 필요한 참조는 `orca skills get orca-cli --reference references/<file>.md`, 안 되면 `--full` | "참조 문서는 `--reference`, 안 되면 `--full`, 그것도 안 되면 `--help` 만 믿고 flag 를 지어내지 않는다" 한 줄 |

차이가 없는 것(같은 뜻이라 손대지 않아요): `terminal read --screen` 으로 화면을 읽는다(L299), 터미널은 핸들로 지목한다(L207), Orca 재시작 뒤에는 새 핸들 하나만 쓴다(L207), `terminal switch` 는 사람이 시킬 때만(MoAI 는 더 엄격해요, L296), 사람의 권한 질문에 대신 답하지 않는다(L303), Task/Run 은 감독형 조정에만 쓴다(L34, MoAI 의 리더는 감독이라 Run 이 맞아요), 리더가 자기 터미널에 라운드 지시를 치는 `leader-pulse`(L301 "your own pane").

### 결과 요약

- 만든 것: `skills/moai/references/adapters/herdr.md`(읽기 전용: 세션 식별과 상태 읽기, 금지 명령 전부, 없을 때 `not_run`), 매니페스트 항목, `tests/herdr_reference_cli.rs` 8개.
- 아직 안 한 것: Orca 차이 G1~G8 의 실제 편집(파일을 쥔 작업 몫), `tests/product_cli.rs` 의 어댑터 목록 고정 편집(범위 확대 요청 중), 열린 herdr 워크스페이스에서의 실제 읽기(서버가 꺼져 있었고 사람이 워크스페이스를 알려 줘야 해요).

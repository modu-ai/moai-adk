---
id: SPEC-20261121-herdr-reference-orca-gap
title: "A read-only herdr reference, and the Orca guidance gaps against the installed orca-cli skill"
status: draft
revision: 1
created: 2026-11-21
---

# A read-only herdr reference, and the Orca guidance gaps against the installed orca-cli skill

## Intent

MoAI keeps one short reference per optional tool (`skills/moai/references/adapters/`): when to use it, the commands that print the installed version's own guide, the lines MoAI keeps, and what to do when the tool is missing. This SPEC adds herdr as a new optional reference, limited to the first and safest job: inside a herdr pane, in a workspace the user already has open and names, identify the sessions, workspaces, tabs, panes and agents and read their state. It types nothing, starts nothing and changes nothing, and the default flow (Claude Code with the local Rust `moai`) runs the same with herdr absent.

The reference follows the official skill and CLI of the installed version, not the web pages alone: herdr 0.8.2 is installed, the pages show 0.9.5, and 0.8.2 lacks `machine`, `--machine` and `workspace close --group`; `herdr plugin` exists in 0.8.2 too (`herdr plugin --help` lists install, uninstall, link, unlink, enable, disable, list, config-dir, action, log and pane), so every plugin command that changes something is forbidden by name. No herdr server was running when this was written, so the field names come from the bundled schema and the docs, and a read in a live named workspace is the one check that stays open.

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
| D-1 | Outside a herdr pane (`HERDR_ENV` is not `1`) the herdr step is `not_run`, with no exception for a session the user names; only `herdr --version`, `herdr --skill` and `herdr api schema`, which touch no session, may run there (reversible, category none) | The official 0.8.2 skill says: if the check fails, say that you are not running inside Herdr and stop, and do not inspect or control the focused session from outside; the card rule is that the official skill is the source of truth and permissions are not widened. An earlier draft allowed a read from outside with an explicit `--session` and ids, and the audit refused it | Reading from outside would need the official skill to allow it |
| D-2 | Pane text may be read only when the user asked to read that pane (reversible, category none) | Pane text can hold secrets or instructions; state needs only the list and get commands | Remove the pane-read sentence |
| D-3 | The Orca edits are listed and handed over, not made here (reversible, category none) | The files are held by the Orca alignment work; the card says to coordinate instead of duplicating | Apply the list after those files are released |
| D-4 | `worktree.list` and `pane.graphics.info` are not allowed (reversible, category none) | The docs do not call them read-only and the first is judged only by its name | Add them after a documented check |

## Result

### Orca 안내와 설치된 orca-cli 공식 스킬의 차이 (REQ-7)

기준: `orca --version` = 1.4.267. 공식 글은 `orca skills get orca-cli`(399줄, 아래 "OC:줄")와 `orca skills get orchestration`(131줄, "OR:줄")을 2026-11-21T13:08Z 에 읽었어요. MoAI 쪽 줄 번호는 로컬 통합 가지의 지금 tip `994e4c4344` 기준이에요(`ma` = `skills/moai/workflows/mode-agent.md`, `ml` = `mode-leader.md`, `or` = `skills/moai/references/adapters/orca.md`). 이 SPEC 은 이 파일들을 고치지 않아요(D-3). 앞선 표(옛 기준 `53cce6c580`)는 tip 이 움직여서 줄 번호와 상태가 달라졌고, 아래가 그것을 다시 맞춘 거예요. 인용 몇 개(OC:86·97·103·301·305·388, `ma:134`, `ml:83`, `or:24`·`or:38`·`or:69`, `tests/product_cli.rs:2850`)는 따로 다시 읽어 확인했어요.

| # | 상태 | 파일:줄 (tip) | MoAI 의 지금 문장 | 공식 문장 | 바꿀 문장 |
| --- | --- | --- | --- | --- | --- |
| G1 | 바뀜(일부 해결) | `ma:140`, `ml:70` | "never type the same text again with `orca terminal send` because of it" | OC:103 "If mail is refused, follow the refusal or tell the user. Never fall back to `terminal send`." | 옛 `delivered_at` 대체 경로는 사라졌고 `tests/orca_alignment.rs` 가 고정해요. 남은 대체 경로는 N1·N3 |
| G2 | 열림 | `ma:134`, `ml:64` | "Read before sending: `orca terminal read --screen` must show a live session waiting for input; never overwrite another session's unsubmitted input." | OC:301 "Use `terminal send` for shells, dev servers, REPLs and your own pane. Never use it to message another agent or answer its prompts, even when mail is refused; use `orchestration send`" | "`orca terminal send` 는 셸·개발 서버·REPL·자기 창에만 쓰고 먼저 `orca terminal read --terminal <handle> --screen` 으로 읽어요. 다른 에이전트에게는 `orca orchestration send`, 그 터미널에는 치지 않아요." |
| G3 | 열림 | `or:8` | "reading and sending to lane terminals" | OC:301, OC:86 "Send another agent a message with Orca mail, never by typing into its terminal" | "lane worktrees, reading lane terminals and messaging lane agents by Orca mail (`orca orchestration send`)" |
| G4 | 열림 | `or:24` | "Read a terminal before sending to it, and never overwrite another session's unsubmitted input." | OC:298 "Use `terminal read` before `terminal send` unless the next input is obvious."; OC:301 | "Read a terminal before typing into your own shell, REPL or pane. Never type into another agent's terminal; message it with Orca mail." |
| G5 | 바뀜(일부 열림) | `or:69`, `tests/product_cli.rs:2850`, `skills/moai/references/source-manifest.json`(orca.md 항목), `ma:132`·`ml:62`("1.4.256"), `ma:152`·`ml:83`("1.4.259") | `or:69` 는 orca-cli 스킬을 1.4.257, 2026-11-13 에 읽었다고 적고 시험이 `"orca 1.4.257"` 을 고정해요. `or:68` 은 이미 1.4.267 이지만 `skills get orchestration` 만이에요 | `orca --version` 은 1.4.267 | `or:69` 를 "`orca skills get orca-cli`, `orca skills list` and `orca agent-context --json` of orca 1.4.267 (read <date -u>)" 로, 시험 고정 문자열과 매니페스트 해시를 같은 커밋에서 함께 바꿔요 |
| G6 | 열림 | `ma:135`, `ml:65` | "`accepted: true` proves only that the input was accepted — for a plain shell Orca cannot report delivery, so read the target again." | OC:302 "`accepted: true` proves input acceptance, not a started turn", OC:304 `input_accepted` 다음 `turn_started`, OC:305 `--wait-submit`, OC:306 `--retry-request <id>` | "`accepted: true` is input acceptance, not a started turn; a text-plus-Enter agent prompt returns `input_accepted`, then `turn_started`; pass `--wait-submit` for proof; after an ambiguous failure repeat the exact command with `--retry-request <id>`; never resend on silence." "plain shell Orca cannot report delivery" 는 공식 글에 근거가 없어요 |
| G7 | 열림 | `or`(없음), `ma:143`·`ml:73` 는 "never by its title" 만 | 에이전트를 어떻게 찾는지 문장이 없어요 | OC:76~80 `agent list/show/read` 는 `orca_session_id`, OC:294 `title` 로 다시 찾지 않기, OC:295 `worktree ps --json` 의 `agents[].paneKey` 를 `terminal list` 행의 `tabId:leafId` 에 이어 핸들을 얻기 | `or` 의 "Lines MoAI keeps" 에: "Find an agent with `orca agent list/show/read --json`; for a terminal handle join `worktree ps --json` `agents[].paneKey` to `terminal list --json` `tabId:leafId`; never match on `title`." |
| G8 | 열림 | `or:16`, `ma:132`, `ml:62` | "`orca skills get orca-cli` — the CLI guide for this Orca version" | OC:388 "`orca skills get orca-cli --reference references/<file>.md` … `--references` lists the names", 거절되면 `--full` 한 번 | 한 줄 덧붙여요: "references come from `--reference references/<file>.md`; if rejected, use `--full` once; never invent a flag." |
| N1 | 새로 | `ml:83` | "deliver by the card note … and, if the screen shows an idle empty input box, by the screen send of items 1 and 2 of 'Session control with Orca'" | OC:301, OC:103 | Run 경로가 실패했을 때 다른 레인에 화면 전송으로 대체하는 길이에요. "deliver by the card note and `orca orchestration send --to <handle> --type dispatch`; if refused, follow the refusal or tell the user." 로 바꿔요 |
| N2 | 새로 | `ml:30`(supervise), `ml:28`, `ml:44`, `ma:136`, `ma:141`·`ml:71` | 재시작 한 줄(`순환 계속 — self-pick (1)부터`), `/clear`·`/new`·`/compact` 뒤 임무 줄, 재배차를 "lane's screen" 에 한 번 더 전달 | OC:86, OC:97 "Ordinary mail never types your words into a terminal agent's input", OR:76 "Never: Steer through a worker's terminal; use `send`." | 다른 에이전트 터미널에 치는 글들이에요. 글 줄(재시작·배차·재개)은 `orchestration send --type dispatch` 로 보낼 수 있어요. `/clear`·`/compact` 는 두 공식 스킬 어디에도 허용된 전달 길이 없어서, 허용 길이 있는지는 **확인 못 했고** 사용자 결정이 필요해요 |
| N3 | 새로 | `or:38`, `ma:140`, `ml:70` | `or:38` "a typed line to a lane you are watching goes through `orca terminal send` after the read-before-send lines above" 와 orchestration 웹 페이지 인용 "For one-off prompts, use orca terminal send" | OC:301, OR:76, OC:92 "`dispatch` to give it a task" | 설치된 두 공식 스킬은 반대로 말해요(웹 페이지 인용은 이번에 다시 읽지 않아서 확인 못 했어요). "a plain notice or one-off prompt to an agent goes as mail (`orchestration send`, `--type status|dispatch|question`)." |
| N4 | 새로 | `or:37`, `ml:83` | `or:37` "The pages do not document `delivered_at`, any mail state"; `ml:83` "a mail to a `term_…` handle is not kept for a lane that is not checking" | OC:96 "the receipt means the mail was stored, not that the agent read it or started work", OC:97~98 깨울 때 쪽지만 치고 `recipient_delivery_held` 경고 | 1.4.259 때 측정은 낡았어요. "receipt = stored; a terminal agent is woken by a note and reads the mail with `orca orchestration check`; `recipient_delivery_held` means the mail waits unread; follow the receipt's instruction." 두 공식 스킬에 `delivered_at` 는 없어요(세어 보니 0) |
| N5 | 새로(코드) | `src/leader_pulse.rs` `check_terminal` | `ml:77` "`moai leader-pulse --terminal <your own handle> …` … never to another lane's handle" | OC:301 은 "your own pane" 을 허용, OC:94 `ORCA_TERMINAL_HANDLE` 은 자기 핸들 | 글로는 자기 핸들로만이지만 코드는 글자·숫자·`. _ : -` 로 된 아무 핸들이나 받아요(코드를 읽었고 돌리지는 않았어요). `ORCA_TERMINAL_HANDLE` 만 허용하거나 기본값으로 삼는 변경을 제안해요 |
| N6 | 새로(코드) | `src/lane_send.rs:361`, `src/help.rs:516` | 스킬 글은 없어요(`skills/` 에서 `lane-send` 검색 0) | OC:301 | `moai lane-send --terminal H --text T` 는 화면 상태를 보고 아무 레인 터미널에나 쳐요. `--wait-submit`·`--retry-request` 를 안 써서 OC:305~306 이 쓰이지 않아요. OC:301 에 맞추거나 자기 창 전송용 보호 장치라고 문서에 적어요 |

차이가 없는 것(같은 뜻이라 손대지 않아요): 다른 세션의 권한·신뢰 질문에 대신 답하지 않는다(`ma:137`·`ml:67`, OC:303), 터미널은 핸들로 지목하고 title 로 찾지 않는다(`ma:143`·`ml:73`, OC:294), Orca 재시작 뒤 새 핸들 하나만 쓴다(OC:207), `terminal switch` 는 쓰지 않는다(`ma:139`·`ml:69`, OC:296), `--screen` 으로 읽는다(`ma:136`·`ml:66`, OC:299~300), 침묵에 다시 보내지 않고 `--retry-request` 를 쓴다(`ma:135`·`ml:65`, OC:302·306), `orchestration send --to dispatch:<id> --type dispatch`(`ml:83`, OR:64), `worker_done` 와 `--outcome`(`ma:152`, `orca orchestration send --help`). 확인 못 한 것: `check --peek`(`ma:140`·`ml:70`)은 두 공식 스킬에 없어서 `orca orchestration check --help` 로만 확인되고 이번에 다시 읽지 않았어요. `worker-start --terminal <handle> --keep-open`(`ml:83`)이 `--worktree` 와 함께 받아들여지는지도 확인 못 했어요(도움말은 `--terminal` 을 다시 쓸 때 `--worktree` 를 빼라고 해요).

### 결과 요약

- 만든 것: `skills/moai/references/adapters/herdr.md`(herdr 창 안에서만, 읽기 전용: 세션 식별과 상태 읽기, 금지 명령 45개를 전체 이름으로 적음, 없을 때·창 밖일 때 `not_run`), 매니페스트 항목, `tests/herdr_reference_cli.rs` 10개, `tests/product_cli.rs` 의 어댑터 목록 고정 한 줄(`herdr.md`).
- 감사 FAIL 뒤 고친 것: 0.8.2 에도 `herdr plugin` 이 있다는 사실을 문서와 이 SPEC 에 바로잡고 plugin 명령 9개를 금지 목록에 올렸어요. 읽기 전용 줄은 명령 전체 이름 13개로 쓰고 시험이 그 집합을 통째로 비교해요(`herdr pane run` 같은 쓰기 명령이 끼면 실패). 창 밖에서 읽는 예외는 공식 스킬대로 지웠어요(D-1). Orca 차이 표는 tip `994e4c4344` 기준으로 다시 맞췄어요.
- 아직 안 한 것: Orca 차이 G1~G8·N1~N6 의 실제 편집(파일을 쥔 작업 몫, D-3), 열린 herdr 워크스페이스에서의 실제 읽기(서버가 꺼져 있었고 사람이 워크스페이스를 알려 줘야 해요), `/clear`·`/compact` 를 다른 에이전트에 전하는 허용된 길이 있는지(N2, 사용자 결정).

### 증거 (2026-11-21)

- `cargo test --offline --locked --test herdr_reference_cli`: 10 passed(`target/h.log`). `--test product_cli -- source_manifest_covers_every_package_file adapters_point_to_version_matched_guides`: 2 passed.
- `moai-cli spec check SPEC-20261121-herdr-reference-orca-gap`: 0 errors, 0 warnings.
- 문서 변형 6개 모두 killed(`target/doc_mutants2.py`): 쓰기 명령이 읽기 줄에 끼기, plugin install 을 금지에서 빼기, 창 밖 예외 되살리기, `pane input` 을 금지에서 빼기, 0.8.2 plugin 오류 되살리기, 창 밖에서 `status --json` 허용. 변형 뒤 파일은 `git diff` 로 복원을 확인했어요.
- 확인한 사실: `herdr plugin --help`(0.8.2)가 install·uninstall·link·unlink·enable·disable·list·config-dir·action·log·pane 을 나열해요. `herdr <그룹> --help` 로 명령 이름(pane·agent·worktree·server·session·workspace·tab·terminal·notification·integration·config·channel)을 확인했어요. `herdr --skill` 이 "If the check fails, say that you are not running inside Herdr and stop" 라고 해요.
- 돌리지 않은 것: `orca_optional_3005`·`agent_defs_package` 는 이번 수리에서 다시 안 돌렸고(앞선 병합 검사 `check affected` PASS 에서 돌았어요), 새 병합 검사에서 다시 돌아요. herdr 실제 워크스페이스 읽기는 안 했어요.

### 수용 기준 결과

- AC-1: met — `herdr_reference_cli` 10개(다섯 제목, 읽기 줄의 정확한 명령 집합, 금지 명령 45개의 전체 이름, 창 밖 `not_run`, 상태 낱말 다섯, 카드 번호·MCP 길 없음).
- AC-2: met — `product_cli::skill_package::source_manifest_covers_every_package_file` ok, 해시 `076b4331aa…`.
- AC-3: met — Result 의 Orca 차이 표(G1~G8, N1~N6)에 항목마다 파일·줄·공식 문장·대체 문장이 있어요.
- AC-4: met — herdr 이 없는 환경에서 `orca_optional_3005` 가 앞선 병합 검사(`check affected`, base 994e4c4344, result=PASS)에서 통과했어요. 새 병합 검사에서 다시 확인해요.

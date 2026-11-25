---
id: SPEC-20261125-agent-team-session-names
title: "The launcher starts a named leader or agent session with -l and -a, and the team runs every flow step over cross-session messages"
status: approved
revision: 3
created: 2026-11-25
approval: {by: user, at: 2026-11-25, basis: "leader note on this SPEC's card (user pick 2026-11-25T01:54:30Z, quoted verbatim: 승인, 표 그대로 (Recommended)) approving revision 3; D-12 = A by that pick, D-13 = A and D-14 = A by the leader decision note of 2026-11-25T01:38:33Z", digest: "sha256:bd29c76db92b082489cd7347aeafafe220da7916c1074ee4f6d35f4fc32be4ba"}
---

# The launcher starts a named leader or agent session with -l and -a, and the team runs every flow step over cross-session messages

## Interview

- Q1: 런처에서 세션 이름과 권한 모드를 어떻게 정하나요? → A: "아니면 moai cc --agent, -a, --leader -l 으로 처리를 하는게 어떨까? 그리고 moai cc/glm 으로 옵션이 없이 진행하면 auto mode 실행, -a, -l 등이면 --bypass로 자율 코딩 모드로 진행하고 이름도 그때 변경하도록 하자."(사용자 2026-11-25T01:18:13Z, 리더 대화, 결정 노트 8849)
- Q2: 우회 보호는 확인됐고 advisor 는 누구에게 쓰나요? → A: "우회시 보호 되는지 확인했으니 -a, -l 으로 진행해라. -a 때만 advisor 사용하기"(같은 시각, 결정 노트 8849)
- Q3: 역할 낱말은 무엇인가요? → A: "worker -> /moai agent로 표기를 하는게 맞겠지?"(결정 노트 8850), "옛 낱말 /moai work 는 계속 사용하지 않는다. /moai work 지시사항 -> 이렇게 사용하지 않는다. 이유는 /moai agent로 설정을 하고 지시는 메시지로 전달하면 된다. 모드 전환이지 명령어가 아니다"(결정 노트 8852)
- Q4: -a 의 advisor 와 기본 모델은 무엇인가요? → A: "agent의 advisor는 opus로 하고, 기본 모델은 sonnet high 이다."(사용자 2026-11-25T01:36:24Z, 결정 노트 8864)
- Q5: 이 SPEC 개정 3을 승인하고, GLM 에이전트(moai glm -a)의 모델을 어떻게 할까요? → A: "승인, 표 그대로 (Recommended)"(사용자 2026-11-25T01:54:30Z, 리더의 선택 질문, 리더 노트)

## Intent

이 SPEC 의 카드에 달린 결정 노트 8841 이 첫 요구예요(리더가 사용자 말을 그대로 옮긴 기록, 2026-11-25T01:03:40Z): 리더와 워커 세션이 이름을 정하고 크로스 세션 메시지로 모든 흐름(배차, 질문과 답, 감사 요청과 판정, 수리 요청, 착지 알림, 다음 카드)에서 협업해요. 메시지는 알림이고 진실은 원장이에요.

revision 2 는 결정 노트 8849(사용자 결정 2026-11-25T01:18:13Z, 리더 대화, 말 그대로 두 메시지)를 싣고 revision 1 의 이름 짓는 방식(세션 안에서 `/rename` 안내)을 보조 길로 내려요:

> "아니면 moai cc --agent, -a, --leader -l 으로 처리를 하는게 어떨까? 그리고 moai cc/glm 으로 옵션이 없이 진행하면 auto mode 실행, -a, -l 등이면 --bypass로 자율 코딩 모드로 진행하고 이름도 그때 변경하도록 하자. advisor 옵션도 추가를 하면 좀더 레인들이 문제 해결에 도움이 되지 않을까?"
>
> "우회시 보호 되는지 확인했으니 -a, -l 으로 진행해라. -a 때만 advisor 사용하기"

이 결정은 런처 `-l`/`-a` 를 은퇴시킨 `SPEC-20261124-launcher-retire-role-flags` 의 REQ-1(`-l`·`-a` 를 거절)과 D-1(2026-11-09 결정, 2026-11-24 재확인)을 뒤집어요. 결정 로그에 새 줄을 남기고, 은퇴를 말하는 문장(`CLAUDE.md`, `docs/**`)을 같이 고쳐요. `CLAUDE.md` 가 범위에 들어오고 권한 우회와 보호를 건드리므로 tier 2 전체 감사예요.

정해진 동작(결정 노트 8849 와 리더의 결정):

- `moai cc|glm` (옵션 없음): 지금처럼 역할 없는 세션, 이름 없음, `cc` 는 auto mode, `glm` 은 우회.
- `moai cc|glm -l|--leader`: 세션 이름 `leader-<프로젝트명>`(`claude --name`), 권한 우회(`--dangerously-skip-permissions`), 첫 입력 `/moai leader`.
- `moai cc|glm -a|--agent`: 세션 이름 `agent-<영어이름>`, 권한 우회, 첫 입력 `/moai agent`, advisor 는 `opus`(`cc` 만, `glm` 은 advisor 가 없어 지금의 안내 그대로), 기본 모델은 sonnet 이고 effort 는 high(결정 노트 8864, 사용자 결정 2026-11-25T01:36:24Z, 말 그대로 "agent의 advisor는 opus로 하고, 기본 모델은 sonnet high 이다." — 리더의 해석: `-a` 에만 해당하고 `-l` 과 옵션 없는 launch 의 기본값은 그대로예요). `--model`, `--effort`, `--advisor` 를 직접 주면 그것이 이겨요. `-l` 과 옵션 없음에서는 advisor 를 켜지 않아요.
- 세션 안의 `/moai leader`, `/moai agent` 는 모드 전환이에요(결정 노트 8852, 사용자 결정 2026-11-25T01:22:08Z, 말 그대로 "옛 낱말 /moai work 는 계속 사용하지 않는다. /moai work 지시사항 -> 이렇게 사용하지 않는다. 이유는 /moai agent로 설정을 하고 지시는 메시지로 전달하면 된다. 모드 전환이지 명령어가 아니다"): 뒤에 지시 글을 받지 않고, 세션을 그 역할로 설정한 뒤(이름 확인, 등록, 흐름 참조 읽기) 대기해요. `/moai work` 는 별칭으로도 받지 않아요. 지시(배차, 카드 번호, 범위)는 크로스 세션 메시지로 와요. 옵션 없이 띄운 세션과 Desktop 이 이 길을 쓰고, 이때만 이름은 사용자에게 `/rename` 한 줄을 안내하고 `ListAgents` 첫 줄로 확인해요(보조 길).
- 영어이름은 런처가 골라요(코드 안의 이름 목록, 등록부에 살아 있는 이름은 건너뜀). `--name` 을 주면 그 이름을 써요.
- 레인 경로 낱말은 `agent` 예요(결정 노트 8850, 사용자 결정 2026-11-25T01:19:50Z, 말 그대로 "worker -> /moai agent로 표기를 하는게 맞겠지?" 에 리더가 동의). 런처 `-a|--agent`, 세션 이름 `agent-<영어이름>`, 등록부의 mode 값 `agent` 와 한 낱말이에요. `worker` 는 쓰지 않고 옛 낱말 `work` 도 받지 않아요(결정 노트 8852). 경로 낱말은 `agent` 와 `leader` 둘뿐이에요. 이 `agent` 는 카드를 맡는 세션(레인)이고 스킬의 base agents(`moai-plan` 등)나 호스트의 서브에이전트와 다른 것이라서 워크플로 글 첫 문장이 한 번 구분해요(REQ-24).

읽어서 확인한 사실(2026-11-25, 이 워크트리 NeXT `e859eae700`):

- 모델은 `/rename` 을 실행할 수 없어요(Skill 도구가 `rename is a UI command, not a skill. Ask the user to run /rename themselves` 로 거절, 조사 노트 8843). 순정 방법은 시작할 때 `claude -n <이름>`, 세션 중 `/rename <이름>`, 선택기의 Ctrl+R 이에요(https://code.claude.com/docs/en/sessions). 그래서 이름을 시작 때 정하는 런처가 확실한 길이에요.
- 세션 이름이 크로스 세션 메시지의 주소예요(https://code.claude.com/docs/en/cross-session-messaging). 이름이 겹치면 짧은 id 를 써요. 이름이 겹치지 않아야 한다는 요구는 문서에 없어요. 실제 목록에서 `leader`, `worker-1`, `worker-2` 가 둘씩 겹친 것을 봤어요.
- 런처 코드는 `-l`/`-a` 와 `agent-N` 번호 매기기가 `SPEC-20261124-launcher-retire-role-flags` 로 지워졌고(`src/session.rs` 의 `retired_role_flag`, `RETIRED_ROLE_FLAG_MESSAGE`, `src/main.rs` 의 `LaunchOpts`) 되살릴 참고 자료는 그 앞의 커밋 `2e6873482e` 예요(`launch(host, leader, …)`, `take_agent_number`). 지금 `src/session.rs` 의 `claude_argv` 는 `--name`, `--model`, `--effort`, `--advisor`, 접근 모드(`Access::Auto|Full|Prompt`)를 이미 composes 해요. `SessionRequest.role`, `prompt` 필드와 `<name>.live` 잠금(`lock_named_live`, `live_lock_held`)도 남아 있어요.
- `claude_argv` 는 첫 입력(`prompt`)을 호스트에 넘기지 않고 `PROMPT was not passed to the Claude host` 라고 적어요. `/moai leader` 를 첫 입력으로 넘기는 길이 되는지는 공식 문서와 실제 실행으로 확인해야 해요(Risks R-6, Plan 1).
- 접근 모드 기본값은 `access_default`: `cc` 는 auto, `glm` 은 우회. 사용자 설정 `launcher.role_access` 와 `--permissions` 가 이를 바꿔요. advisor 는 `--advisor`/`--no-advisor`/`launcher.advisor` 로만 켜져요(`launch_settings`).
- 보호 훅은 우회 모드에서도 거절하는지 사용자는 확인했다고 했지만 리더는 직접 보지 못했어요. READY 에 실제 실행 증거를 붙여야 해요(Acceptance AC-9).
- 흐름 참조 `skills/moai/references/flows/next-flow.md` 는 8041 바이트이고 한도는 8192 라서(`tests/next_flow_skill_scan.rs` 의 `FLOW_REF_LIMIT`) 메시지 표를 그 안에 넣을 자리가 없어요.
- 완료 알림 글은 이미 있어요(`SPEC-20261123-lane-completion-handshake`, NeXT 에 착지). 이 SPEC 은 그 위에 이름과 전체 메시지 표를 더해요.

읽는 사람: 워커와 리더 세션, 이 변경의 감사자, 런처를 쓰는 사람.

## Non-goals

- 우회 모드의 보호 훅 자체를 새로 만들거나 약하게 하지 않아요. 이미 있는 훅이 우회에서 거절하는지 실제로 보여 줄 뿐이에요. 거절되지 않는 것이 하나라도 있으면 이 SPEC 은 멈추고 `LANE-BLOCKED` 로 올려요.
- 옵션 없는 `moai cc|glm` 의 동작, 모델과 effort 기본값, 키 보관함과 GLM 환경 구성, `moai update` 의 모델 선택은 바뀌지 않아요.
- `launcher.role_access` 사용자 설정은 `-l`/`-a` 가 아닌 launch 에만 적용돼요(D-7).
- 모델이 `/rename` 을 대신 치게 하거나 설정 파일·훅으로 세션 이름을 정하는 우회는 만들지 않아요.
- 번호 매기기(`agent-N`)와 그 순번 폴더는 되살리지 않아요. 영어이름 목록이 대신해요.
- 워커끼리의 메시지는 판정·결정의 기록이 아니에요. 판정과 결정은 계속 원장에만 남아요.
- `interview.md`, `project.md`, 전역 설치본 `~/.claude/skills/moai`(착지 뒤 리더가 `moai update` 로 갱신해요), 실제 HOME, 네트워크는 쓰지 않아요.

## Requirements

<!-- This SPEC changes the launcher (a permission-bypass launch), the shipped skill text and the tests that pin them; it adds no blocking check. -->
- REQ-1: When `moai cc` or `moai glm` is given `-l` or `--leader`, the launcher shall start the host with the session name `leader-<project name>`. · failure: none — the name is composed from the project name, and REQ-2 covers a name that is already taken.
- REQ-2 (REQ-1): IF the session live lock of `leader-<project name>` is already held, THEN the launcher shall refuse the launch and start no host.
- REQ-3: When `moai cc` or `moai glm` is given `-a` or `--agent` without `--name`, the launcher shall start the host with the session name `agent-<english name>`, the first name of the launcher's list that has no held session live lock and no registry entry. · failure: none — REQ-4 covers a list that is used up.
- REQ-4 (REQ-3): IF every name of the list is taken, THEN the launcher shall refuse the launch with one message that names `--name` and start no host.
- REQ-5: When `--name <name>` is given with `-l` or `-a`, the launcher shall use that name as given, after the registry's name check. · failure: none — a name the check refuses is refused as it is today.
- REQ-6: When `-l` or `-a` is given, the launcher shall start the host with `--dangerously-skip-permissions`, unless `--permissions` is given. · failure: none — `--permissions` keeps the host's own prompts (D-7).
- REQ-7: When `-l` is given, the launcher shall hand the host the first input `/moai leader`. · failure: none — if the host does not run it, the session is the plain session the user can type into (R-6).
- REQ-8: When `-a` is given, the launcher shall hand the host the first input `/moai agent`. · failure: none — the same as REQ-7.
- REQ-9: When `-a` is given on the `cc` host and `--no-advisor` is not, the launcher shall pass `--advisor` with the value of `--advisor`, else of the setting `launcher.advisor` (D-13), else `opus` (D-6). · failure: none — a model that ranks below the main model is handled as D-14 says.
- REQ-28: When `-a` is given and neither `--model` nor `--effort` is, the launcher shall start the host with the model and effort of the saved choice for the agent role, else of the `agent` row of the default table for that host (D-12). · failure: none — `--model` and `--effort` each win for their own value, and `-l` and a launch without a role flag keep today's defaults.
- REQ-10: The launcher shall pass no `--advisor` by default for `-l`, for a launch without a role flag and on `glm`. · failure: none — the existing `--advisor` and `launcher.advisor` forms stay.
- REQ-11: The launcher shall keep a launch without a role flag as it is: no name, no first input, `cc` in auto mode, `glm` in the bypass. · failure: none — the existing tests pin it.
- REQ-12: IF `-l` and `-a` are both given, or a PROMPT is given with a role flag, THEN the launcher shall refuse the launch with one message and start no host.
- REQ-13: The launcher help and usage shall name `-l`, `--leader`, `-a` and `--agent` with what each does, in place of the retired-flag refusal. · failure: none — the help golden files pin the text.
- REQ-14: The repository documents that the Scope names shall say that `-l` and `-a` are launcher options again and that the decision reverses D-1. · failure: none — the repository scan tests read these files.
- REQ-15: When a request's first word after the skill name is `agent`, the skill entry and the route table shall pick the WORK route. · failure: none — a route word only; REQ-27 removes the old word.
- REQ-16: The leader and agent workflows shall tell a session whose name does not already start with `leader-` or `agent-` (the session was not started with `-l` or `-a`) to give the user the one line `/rename <name>` once in the user's conversation language and to read its name from the first line of `ListAgents` before it goes on. · failure: none — a text rule.
- REQ-17 (REQ-16): IF the name read from `ListAgents` is not the wanted one after the user's reply, or the user gives none, THEN the workflows shall tell the session to go on under the name it has, to address and be addressed by the short id, and to write one ledger note that says so.
- REQ-18: When a session sends a cross-session message, the flow text shall tell it to address the registered leader by the session id read with `moai session-register --show` matched to a row of `ListAgents`, and any other session by its listed name, or by the short id when two rows share the name. · failure: none — with no matching row nothing is sent and the ledger note stands.
- REQ-19: The flow text shall point to a message table that names, for dispatch, question, answer, audit request, verdict, repair request, landing notice, completion and next card, who sends it, who receives it and its one-line head `<KIND> <card> <state> head=<10 characters>`. · failure: none — a text rule.
- REQ-20: The message table shall say that a message is a notice, that the ledger is the record and that the receiver reads the ledger before it acts. · failure: none — a text rule.
- REQ-21: The flow text shall tell an agent that it may send a message to another agent to check a fact, and that a verdict, a decision and a repair request go to the ledger as well. · failure: none — a text rule.
- REQ-22: The flow text shall state the channel order: the host's cross-session message first, Orca messaging (`orca orchestration send`, and `orca orchestration check` for the leader) when the host's channel is missing, refused or held, and the ledger note alone when neither exists, while `/clear` and screen reading stay with the lane-terminal manager. · failure: none — a text rule.
- REQ-23: The flow reference shall stay within its byte budget (8192 bytes) while the message table lives in its own reference file that the skill package's source manifest lists. · failure: none — the existing scan and package tests reject an over-budget or unlisted file.
- REQ-24: The first sentence of the agent workflow shall say that this agent is the session that takes a card, not one of the skill's base agents or a host subagent. · failure: none — a text rule.
- REQ-25: When text follows `/moai agent` or `/moai leader`, the skill shall set the role mode, run nothing from that text and say in one line that instructions come as a message. · failure: none — the mode is set either way.
- REQ-26: While a leader is registered, the agent workflow shall tell the agent to wait for that leader's message after it has set its mode. · failure: none — with no registered leader the agent takes the next card with `moai card next` as today.
- REQ-27: The skill text and the documents and tests that the Scope names shall carry the dispatch as a message line (the dispatch line of the message table) and no `/moai work` form. · failure: none — the scan tests read these files, and the skill entry no longer names `work` as a route word.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-5): `moai cc --leader --dry-run` in a temporary project prints a command with `--name leader-<project name>`, `--dangerously-skip-permissions` and the first input `/moai leader`; a held live lock of that name refuses the launch; `--name` replaces the name · verify: `cargo test --offline --locked --test launcher_role_flags ac1_`
- AC-2 (REQ-3, REQ-4, REQ-5): `moai cc -a --dry-run` picks the first free name of the list, skips a name whose live lock is held or that a registry entry has, refuses with a message naming `--name` when the list is used up, and takes `--name` as given · verify: `cargo test --offline --locked --test launcher_role_flags ac2_`
- AC-3 (REQ-6, REQ-11, REQ-12): `-l` and `-a` start in the bypass on `cc` and `glm`, `--permissions` keeps the host's own prompts, a launch without a role flag is unchanged, and `-l -a` together or a PROMPT with a role flag is refused · verify: `cargo test --offline --locked --test launcher_role_flags ac3_`
- AC-4 (REQ-7, REQ-8): `-l` hands `/moai leader` and `-a` hands `/moai agent` as the host's first input · verify: `cargo test --offline --locked --test launcher_role_flags ac4_`
- AC-5 (REQ-9, REQ-10, REQ-28): `-a` on `cc` passes `--advisor` from the option, else the setting, else `opus`, `--no-advisor` removes it, `-l`, a roleless launch and `glm` pass none; `-a` starts the host with the saved agent choice, else the default table's agent row of the host (on `claude` today `claude-sonnet-5-5` and `high`), `--model` and `--effort` win, and `-l` and a roleless launch keep their model and effort · verify: `cargo test --offline --locked --test launcher_role_flags ac5_`
- AC-6 (REQ-13, REQ-14): the help golden files name the four flags and no help text calls them retired; `CLAUDE.md` and the four docs carry the new sentence; the decisions log has the reversing line · verify: `cargo test --offline --locked --test launcher_role_flags ac6_` and `cargo test --offline --locked --test repo_claude_md` and `cargo test --offline --locked --test help_layout`
- AC-7 (REQ-15, REQ-16, REQ-17, REQ-18, REQ-19, REQ-20, REQ-21, REQ-22, REQ-24, REQ-25, REQ-26, REQ-27): the agent workflow's first sentence tells this agent from the base agents and host subagents; the route table and the skill entry name `agent` and `leader` and not `work` (REQ-25, REQ-26, REQ-27: text after the mode word sets the mode only with a one-line notice, a registered leader's message is awaited, no `/moai work` form is left in the skill text); the workflows carry the rename line and its fallback, the address rule, and the flow text points to the table with the notice rule, the agent-to-agent rule and the channel order · verify: `cargo test --offline --locked --test lane_team_names`
- AC-8 (REQ-23): the flow reference is at most 8192 bytes, the new reference file is in the source manifest and the package scans pass · verify: `cargo test --offline --locked --test next_flow_skill_scan` and `cargo test --offline --locked --test product_cli skill_package::`
- AC-9 (REQ-6): in a temporary project with a bare remote, a real host session started by `-a` (the bypass) is made to try (a) moving the integration branch by hand, (b) deleting an absolute path and (c) force-pushing `main`, and the moai protection hook refuses each; the READY note carries the three runs and their output. If one is not refused the lane stops with `LANE-BLOCKED` · verify: not testable by `cargo` alone; the lane runs it once in the throwaway and the Result records the commands and the refusals (`not_run` until then).
- AC-10 (REQ-1, REQ-3, REQ-7, REQ-8, REQ-18): two sessions started by `moai cc -l` and `moai cc -a` in a temporary project run one round (dispatch, READY, audit request, PASS, landing notice) over cross-session messages · verify: not testable by `cargo`; the leader or the user runs it once after the install and the Result records the two session names and the lines sent (stays `not_run` until then).

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요(D-6 은 비용이 드는 선택이고 D-7 은 권한 우회를 다뤄서 사용자에게 보여요). 추천이 맨 앞이에요.

- D-1 (리더가 정함, 노트 8849): 등록부의 리더 항목 이름은 `leader` 그대로 두고 표시 이름만 `leader-<프로젝트명>` 이에요. 리더의 메시지 주소는 등록부의 세션 id 를 `ListAgents` 행에 맞춰서 정해요.
- D-2 (사용자 결정으로 대체): 영어이름은 모델이 아니라 런처가 골라요(REQ-3). 목록은 D-8.
- D-3 (리더가 정함): 모델이 이름을 못 바꾸는 경우의 안내(REQ-16, REQ-17)는 옵션 없이 띄운 세션과 Desktop 의 보조 길에만 써요.
- D-4 (리더가 정함): 워크플로 파일 이름은 `workflows/work.md` 그대로 두고 경로 낱말 `agent` 만 더해요.
- D-5 (리더가 정함): 메시지 표는 새 참조 파일 `references/flows/team-messages.md` 에 두고 `next-flow.md` 는 한 줄로 가리켜요.
- D-6 (advisor 기본 모델, 비용이 드는 선택, 사용자가 정함): `-a` 의 advisor 기본은 `opus`예요(노트 8864; 앞의 `fable` 은 취소). `--advisor` 는 사용량 크레딧에 청구될 수 있고 메인 모델 이상이어야 해요(`claude_argv` 의 순위 검사).
- D-7 (`--permissions` 와 `launcher.role_access`): (A, 추천) `-l`/`-a` 는 항상 우회로 시작하고, `--permissions` 를 직접 주면 그것이 이겨서 호스트의 프롬프트를 유지해요. 사용자 설정 `launcher.role_access` 는 역할 없는 launch 에만 적용돼요. (B) `--permissions` 를 `-l`/`-a` 와 함께 쓰면 거절해요. (C) `launcher.role_access` 가 `-l`/`-a` 도 바꿔요. 기본값은 A.
- D-8 (영어이름 목록): (A, 추천) 24개, 소문자 한 낱말, 코드 안의 상수: ben, cara, dana, eli, fay, gus, hana, ivan, jade, kai, lena, milo, nora, owen, pia, quin, ruth, sam, tess, uma, vic, wren, xena, zoe. 앞에서부터 빈 것을 써요. (B) 12개로 줄여요. 목록을 늘리는 일은 새 카드예요. 기본값은 A.
- D-9 (`-l` 의 프로젝트명): (A, 추천) `.moai/project.yaml` 의 `project_id`, 없으면 저장소 폴더 이름을 `valid_name` 문자(영문·숫자·`.`·`_`·`-`)로 고쳐 쓰고 이름 전체를 64자 안으로 줄여요. (B) 폴더 이름만. 기본값은 A.
- D-12 (`-a` 의 모델·effort 가 어디서 와요, 새로 열린 선택): 이전 런처는 역할 기본값을 `저장된 선택(moai update) 있으면 그것, 없으면 default_choice(호스트, 역할)` 로 읽었고(`2e6873482e` 의 `src/session.rs`, `zai/` 접두사는 떼고 넘겨요) 지금 표의 `claude`/`agent` 행이 이미 `claude-sonnet-5-5`·`high` 라서 사용자의 문장과 같아요. (A, 추천) 그 방식을 `-a` 에만 되살려요: `--model`/`--effort` > 저장된 선택 > 표의 agent 행. `claude` 는 sonnet·high, `glm` 은 표의 agent 행 그대로(`zai/glm-5.3-flash`, `max`)로 표가 하나인 채로 있어요. (B) 두 호스트 모두 별칭 `sonnet` 과 `high` 로 고정해요: `glm` 은 환경 표대로 `glm-5.4[1m]`·`high` 가 되어 표의 glm agent 행(앞선 사용자 결정)과 갈라져요. (C) `claude` 에서만 바꾸고 `glm` 은 지금의 `glm` 기본(`opus` 별칭, `high`)을 그대로 둬요. 기본값은 A. `glm` 값이 걸려서 사용자가 보는 선택이에요.
- D-13 (`launcher.advisor` 설정과 `-a` 의 기본 advisor): (A, 추천) 지금의 우선순위를 지켜요: `--no-advisor`/`--advisor` > `launcher.advisor` > `-a` 의 기본 `opus`. 설정이 있는 사용자는 자기 설정이 이겨요. (B) `-a` 에서는 설정보다 `opus` 가 먼저예요. 기본값은 A.
- D-14 (메인 모델이 기본 advisor 보다 높을 때, 예 `-a --model fable`): (A, 추천) 기본 advisor 를 넘기지 않고 이유를 한 줄로 알려요. 직접 준 `--advisor` 가 메인보다 낮으면 오늘처럼 거절해요. (B) 어떤 경우든 거절해요. (C) advisor 를 메인 모델과 같은 별칭으로 올려요. 기본값은 A.
- D-11 (SPEC 크기): `spec check` 가 요구 27개에 경고를 내요("12개 넘으면 `parent` 로 이은 SPEC 으로 쪼개요", 오류 아님). (A, 추천) 한 SPEC 으로 가요: 리더가 한 카드로 정했고 두 갈래(런처, 스킬 글)가 같은 메시지 표와 이름 규칙을 서로 가리켜요. 경고는 그대로 둬요. (B) 런처(REQ-1..14)와 글(REQ-15..27) 두 SPEC 으로 쪼개고 글 SPEC 의 `parent` 를 런처 SPEC 으로 해요: 감사 단위가 작아지지만 카드를 둘로 쪼개야 해요. 기본값은 A.
- D-10 (조합): (A, 추천) `-l` 과 `-a` 를 함께, 또는 PROMPT 를 역할 플래그와 함께 주면 거절해요(REQ-12). (B) 첫 입력 뒤에 PROMPT 를 이어 붙여요. 기본값은 A.

## Scope

- write: `.moai/specs/SPEC-20261125-agent-team-session-names/**`, `src/session.rs`, `src/main.rs`, `src/help.rs`, `src/launcher_terminal.rs`(미리 보기에 첫 입력을 보이는 부분이 있을 때만), `skills/moai/SKILL.md`, `skills/moai/references/routes.json`, `skills/moai/workflows/work.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/flows/team-messages.md`(새 파일), `tests/launcher_role_flags.rs`(새 파일, 첫 낱말 `launcher` 는 `tests/areas.toml` 에 있어요), `tests/lane_team_names.rs`(새 파일, 첫 낱말 `lane`), `CLAUDE.md`, `docs/rust-product-path.md`, `docs/user-entry-draft.md`, `docs/host-support-matrix.md`, `docs/decisions-2026-11-08.md`, `docs/claude-md-rule-provenance.md`
- write (읽어서 정확한 경로를 SCOPE-REQUEST 로 올려요): 은퇴한 동작을 고정한 기존 테스트와 골든 — `tests/launcher_retire_role_flags.rs`, `tests/launcher_roleless.rs`, `tests/launcher_permissions_setting.rs`, `tests/launcher_advisor_setting.rs`, `tests/launcher_session_locks.rs`, `tests/help_layout.rs`, `tests/fixtures/help/*`, `tests/product_cli.rs`, `tests/repo_claude_md.rs`, `tests/lane_completion_handshake.rs`, `tests/leader_route.rs`, `tests/mode_workflow_steps.rs`, `tests/skill_structure.rs`, `tests/next_flow_skill_scan.rs`
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`(새 테스트 파일 등록)
- read-only: `src/lane_registry.rs`(등록부 이름 확인), `src/launcher_init.rs`, `src/hook_*.rs`(보호 훅), 커밋 `2e6873482e`(되살릴 참고 자료)
- forbidden: `interview.md`, `project.md`, 전역 설치본과 실제 HOME, 네트워크, Python, 보호 훅을 약하게 하는 변경

## Plan

1. 조사 먼저: 공식 문서와 임시 프로젝트의 실제 실행으로 `claude "/moai leader"` 처럼 첫 입력이 슬래시 명령(스킬)으로 실행되는지, 이름이 `--name` 으로 목록에 보이는지 확인하고 LANE-NOTE 로 남겨요. 안 되면 LANE-BLOCKED 로 대안(SessionStart 훅의 컨텍스트 등)을 올려요.
2. RED: `tests/launcher_role_flags.rs`(AC-1..AC-6)와 `tests/lane_team_names.rs`(AC-7)가 현재 동작에서 실패해요. 앞선 커밋 `2e6873482e` 의 `launch` 와 번호 매기기를 참고해 되살리되 목록과 잠금 확인으로 바꿔요.
3. 런처: `LaunchOpts` 에 `-l`/`-a`, 이름 선택, 접근·advisor 규칙, 첫 입력을 더하고 도움말과 은퇴 거절을 정리해요. GREEN.
4. 글: `routes.json`, `SKILL.md`, `work.md`, `leader.md`, `team-messages.md`, `next-flow.md` 를 써요. 이 글을 고정한 기존 테스트를 읽어 맞추고 매니페스트 digest 를 line-lease 로 고쳐요.
5. 문서: `CLAUDE.md`, 네 `docs/**` 문장, 결정 로그의 새 줄, 출처 표.
6. 위험 경로: 임시 프로젝트에서 실제 `-a` 세션이 (a)(b)(c) 를 시도해 훅이 거절하는 것을 돌려 READY 에 붙여요(AC-9).
7. 가드 뮤턴트(tier 2): 새 가드 테스트마다 `guard-mutant:` 한 줄. 영수증은 `moai check affected` 한 번, READY 에 SPEC digest 를 담아요.

## Risks

- R-1: 바꾼 이름이 `ListAgents` 에 바로 보이는지는 실행으로 확인하지 못했어요. 런처가 `--name` 으로 시작하면 이 문제가 없고, 보조 길의 글은 이름을 `ListAgents` 첫 줄로 읽어 확인하고 다르면 REQ-17 로 가요.
- R-2: 이름이 겹칠 수 있어요. 런처는 살아 있는 잠금과 등록부를 건너뛰고, 글은 겹치면 짧은 id 로 주소를 써요.
- R-3: `next-flow.md` 의 여유가 151 바이트라서 거기에는 한 줄 포인터와 채널 순서 정리만 쓰고, 넘으면 `team-messages.md` 로 옮겨요.
- R-4: 다른 카드가 같은 스킬 폴더의 흐름 이름을 고쳐요. `next-flow.md` 와 매니페스트에서 겹칠 수 있어서 병합을 한 번 해요.
- R-5: `-l`/`-a` 는 권한 우회로 시작해요(보호 훅이 막는 것은 AC-9 의 세 가지로 보여 줘요). 우회 모드에서 안전장치는 훅뿐이라서 거절되지 않는 것이 하나라도 있으면 멈춰요. `-l`/`-a` 를 쓰는 Orca 프리셋과 셸 별칭은 이 SPEC 이 되살리는 이름과 같아서 다시 동작해요.
- R-6: 첫 입력이 호스트에서 슬래시 명령으로 실행되는지는 아직 확인하지 못했어요(`claude_argv` 는 PROMPT 를 넘기지 않아요). Plan 1 에서 확인하고, 안 되면 REQ-7·REQ-8 의 전달 방법을 이 SPEC 에서 바꿔 승인을 다시 받아요.
- R-7: 이 글이 `docs/lane-rules.md`, `docs/leader-rules.md`(카드 범위 밖)와 어긋날 수 있어요. 어긋나는 문장은 후속으로 리더에게 알려요.

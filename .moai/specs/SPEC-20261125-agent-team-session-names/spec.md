---
id: SPEC-20261125-agent-team-session-names
title: "The leader and worker sessions name themselves, find each other by name and run every flow step over cross-session messages"
status: draft
revision: 1
created: 2026-11-25
---

# The leader and worker sessions name themselves, find each other by name and run every flow step over cross-session messages

## Intent

이 SPEC 의 카드에 달린 결정 노트 8841(리더가 사용자 말을 그대로 인용한 기록, 2026-11-25T01:03:40Z)이 이 SPEC 의 요구예요: `/moai leader` 로 시작한 창은 `leader-{프로젝트명}`, `/moai worker`(`/moai work` 도 계속 받아요)로 시작한 창은 `agent-{영어이름}` 으로 이름을 정하고, 리더와 워커가 배차·질문과 답·감사 요청과 판정·수리 요청·착지 알림·다음 카드의 모든 단계를 크로스 세션 메시지로 주고받아요. 메시지는 알림이고 진실은 원장이에요(받은 쪽은 원장을 읽고 움직여요). 이 SPEC 은 스킬 글(영어)과 그 글을 고정하는 테스트만 바꿔요.

읽어서 확인한 사실(2026-11-25, 이 워크트리 NeXT `e859eae700`):

- 모델은 `/rename` 을 실행할 수 없어요. Skill 도구가 `rename is a UI command, not a skill. Ask the user to run /rename themselves` 로 거절했고 이름을 바꾸는 도구도 없어요(조사 노트 8843). 공식 문서(https://code.claude.com/docs/en/sessions)의 순정 방법은 시작할 때 `claude -n <이름>`, 세션 중 `/rename <이름>`, 세션 선택기의 Ctrl+R 이에요.
- 세션 이름이 곧 크로스 세션 메시지의 주소예요(https://code.claude.com/docs/en/cross-session-messaging): 이름이 하나뿐이면 이름만으로 전달하고, 겹치면 목록의 각 행에 짧은 id 가 붙어 주소에 그 id 를 써요. 이름이 겹치지 않아야 한다는 요구는 문서에 없어요. 실제 목록에서 `leader`, `worker-1`, `worker-2` 가 둘씩 겹친 것을 봤어요.
- `ListAgents` 의 첫 줄이 이 세션의 이름과 짧은 id 를 알려 줘요. 그래서 이름을 바꾼 뒤 모델이 그 줄을 읽어 확인할 수 있어요.
- 바꾼 이름이 목록에 바로 보이는지는 문서에 없고, 이 세션이 `/rename` 을 칠 수 없어 실행으로 확인하지 못했어요(not_run). 사용자가 `/rename` 을 친 직후 `ListAgents` 를 한 번 부르는 확인이 열린 항목이에요(Risks R-1).
- `moai session-register` 의 이름은 `valid_name`(영문·숫자·`.`·`_`·`-`, 64자 이하)을 지나는 것이면 되어서 `leader-moai-adk`, `agent-ben` 둘 다 등록돼요(`src/lane_registry.rs`). 그러나 코드와 글이 리더 항목을 이름 `leader` 로 읽어요(`session-register --show --name leader`, `card land --by leader`, `leader_round`). `--mode leader|agent` 는 상태줄 태그 값이에요.
- 흐름 참조 `skills/moai/references/flows/next-flow.md` 는 8041 바이트이고 한도는 8192 라서(`tests/next_flow_skill_scan.rs` 의 `FLOW_REF_LIMIT`) 메시지 표를 그 안에 넣을 자리가 없어요.
- 완료 알림 글(카드 기준 1·2)은 이미 있어요(`SPEC-20261123-lane-completion-handshake`, NeXT 에 착지). 이 SPEC 은 그 위에 이름과 전체 메시지 표를 더해요.

읽는 사람: 워커와 리더 세션, 이 변경의 감사자, 이 스킬을 설치해 쓰는 사람.

## Non-goals

- `moai` 명령과 Rust 코드는 바뀌지 않아요(`src/**` 에 쓰지 않아요). 이름 바꾸기와 메시지 보내기는 스킬 글이 호스트 기능을 쓰게 하는 일이에요. 보내는 `moai` 명령은 호스트 도구를 부를 수 없어 만들지 않아요(결정 노트 8818).
- `interview.md` 와 `project.md` 는 건드리지 않아요(다른 카드가 흐름 이름을 고쳐요). `CLAUDE.md`, 전역 설치본 `~/.claude/skills/moai`(착지 뒤 리더가 `moai update` 로 갱신해요), 실제 HOME 은 쓰지 않아요.
- 모델이 `/rename` 을 대신 치게 하거나, 설정 파일·훅으로 이름을 정하는 우회는 만들지 않아요. 순정 기능이 없으면 사용자에게 한 줄을 안내해요.
- 등록부의 항목 이름 규칙, `card land --by leader` 와 `leader_round` 의 `leader` 읽기는 바꾸지 않아요(D-1 A 를 고를 때).
- 워커끼리의 메시지를 판정·결정의 기록으로 인정하지 않아요. 판정과 결정은 계속 원장에만 남아요.

## Requirements

<!-- This SPEC changes the shipped skill text and the tests that pin it; it adds no command and no blocking check. -->
- REQ-1: When a request's first word after the skill name is `worker` or `work`, the skill entry and the route table shall pick the WORK route. · failure: none — a route word only.
- REQ-2: When a session takes the LEADER route, the leader workflow shall tell it to name the session `leader-<project id>` and to register the window in the registry under the entry name chosen in D-1. · failure: none — a text rule.
- REQ-3: When a session takes the WORK route, the worker workflow shall tell it to name the session `agent-<english name>` by the rule of D-2 and to register the lane in the registry under the same name. · failure: none — a text rule.
- REQ-4: The leader and worker workflows shall tell the session, which cannot run `/rename` itself, to give the user the one line `/rename <name>` once in the user's conversation language and to read its name from the first line of `ListAgents` before it goes on. · failure: none — a text rule.
- REQ-5 (REQ-4): IF the name read from `ListAgents` is not the wanted one after the user's reply, or the user gives none, THEN the workflows shall tell the session to go on under the name it has, to address and be addressed by the short id, and to write one ledger note that says so. · failure: a session that never gets its name still works through its id.
- REQ-6: The worker workflow shall tell the session to pick the english name as one lowercase english given name of 3 to 10 letters that no row of `ListAgents` and no registry entry already uses, and to pick again when the name is taken. · failure: none — a text rule.
- REQ-7: When a session sends a cross-session message, the flow text shall tell it to address the registered leader by the session id read with `moai session-register --show`, matched to a row of `ListAgents`, and any other session by its listed name, or by the short id when two rows share the name. · failure: none — with no matching row nothing is sent and the ledger note stands.
- REQ-8: The flow text shall point to a message table that names, for dispatch, question, answer, audit request, verdict, repair request, landing notice, completion and next card, who sends it, who receives it and its one-line head `<KIND> <card> <state> head=<10 characters>`. · failure: none — a text rule.
- REQ-9: The message table shall say that a message is a notice, that the ledger is the record and that the receiver reads the ledger before it acts. · failure: none — a text rule.
- REQ-10: The flow text shall tell a worker that it may send a message to another worker to check a fact, and that a verdict, a decision and a repair request go to the ledger as well. · failure: none — a text rule.
- REQ-11: The flow text shall state the channel order: the host's cross-session message first, Orca messaging (`orca orchestration send`, and `orca orchestration check` for the leader) when the host's channel is missing, refused or held, and the ledger note alone when neither exists, while `/clear` and screen reading stay with the lane-terminal manager. · failure: none — a text rule.
- REQ-12: The flow reference shall stay within its byte budget (8192 bytes) while the message table lives in its own reference file that the skill package's source manifest lists. · failure: none — the existing scan and package tests reject an over-budget or unlisted file.

## Acceptance

- AC-1 (REQ-1): `routes.json`, the skill entry text and the worker workflow name `worker` and `work` for the WORK route and `work` still opens the same workflow · verify: `cargo test --offline --locked --test lane_team_names ac1_`
- AC-2 (REQ-2, REQ-3): the two workflows carry the session names (`leader-<project id>`, `agent-<english name>`) and the registry names of D-1 · verify: `cargo test --offline --locked --test lane_team_names ac2_`
- AC-3 (REQ-4, REQ-5): both workflows say the model cannot run `/rename`, give the one line for the user, read `ListAgents` and carry the fallback to the id with its ledger note · verify: `cargo test --offline --locked --test lane_team_names ac3_`
- AC-4 (REQ-6): the worker workflow carries the english-name rule with its three conditions · verify: `cargo test --offline --locked --test lane_team_names ac4_`
- AC-5 (REQ-7, REQ-8, REQ-9, REQ-10, REQ-11): the flow reference points to the message table, and the table names every message kind with sender, receiver and head line, the notice rule, the worker-to-worker rule and the channel order · verify: `cargo test --offline --locked --test lane_team_names ac5_`
- AC-6 (REQ-12): the flow reference is at most 8192 bytes, the new reference file is in the source manifest, and the package scans pass · verify: `cargo test --offline --locked --test next_flow_skill_scan` and `cargo test --offline --locked --test product_cli skill_package::`
- AC-7 (REQ-3, REQ-8): a round of two real sessions in a temporary project (dispatch, READY, audit request, PASS, landing notice) runs over cross-session messages · verify: not testable by `cargo`; the leader or the user runs it once after the install and the Result records the two session names and the lines sent (stays `not_run` until then).

## Decisions

각 결정은 되돌릴 수 있고 분류는 none 이에요. 추천은 맨 앞이에요.

- D-1 (등록부의 리더 항목 이름): (A, 추천) 항목 이름은 `leader` 로 두고(코드와 글의 `--name leader`, `card land --by leader`, `leader_round` 가 그대로 돌아요) 표시 이름만 `leader-<project id>` 로 해요. 리더의 메시지 주소는 등록부의 세션 id 를 `ListAgents` 행에 맞춰서 정해요(이미 있는 규칙). (B) 항목 이름도 `leader-<project id>` 로 등록해요: 이름이 한곳에 모이지만 `--by leader` 와 `leader_round` 와 이 이름을 읽는 모든 글과 테스트를 같이 고쳐야 하고(`src/**` 가 범위에 들어와요) 이전 단계가 필요해요.
- D-2 (영어이름 고르는 규칙): (A, 추천) 모델이 소문자 영어 이름 한 낱말(3~10자)을 골라요. `ListAgents` 의 다른 행 이름과 등록부 항목에 이미 있는 이름은 건너뛰고 겹치면 다시 골라요. 이름은 표시 이름이면서 등록부 레인 이름이에요(`agent-ben`). (B) 글에 고정된 이름표(예: 20개)를 두고 앞에서부터 빈 것을 써요: 겹침은 확실히 막지만 이름표를 늘려야 하고 스킬 글이 길어져요. (C) 지금처럼 번호(`agent-1`)를 써요: 사용자 요구(영어이름)와 달라요.
- D-3 (모델이 이름을 못 바꿀 때): (A, 추천) 첫 단계에서 한 줄(`/rename <이름>`)을 답에 적고 `ListAgents` 첫 줄로 확인해요. 사용자가 치지 않거나 이름이 다르면 가진 이름 그대로 가고 짧은 id 로 주소를 쓰고 원장에 한 줄을 남겨요. (B) 이름이 바뀔 때까지 멈춰요: 이름은 확실하지만 사용자가 없는 밤에 레인이 멈춰요. 선택 질문 도구는 쓰지 않고(워커는 사용자에게 선택 질문을 열지 않아요) 답의 한 줄로만 안내해요.
- D-4 (파일 이름): (A, 추천) 워크플로 파일 이름은 `workflows/work.md` 로 두고 경로 낱말 `worker` 만 더해요: 이 파일 이름을 가리키는 `src/hook_engine.rs`, `src/leader_pulse.rs` 의 주석, `docs/**`, 테스트 열댓 곳이 그대로예요. (B) `workflows/worker.md` 로 바꿔요: 이름이 맞지만 그 파일들이 모두 범위에 들어와야 해요(`src/**` 주석 포함).
- D-5 (메시지 표 자리): (A, 추천) 새 참조 파일 `references/flows/team-messages.md` 에 두고 `next-flow.md` 는 한 줄로 가리켜요: 한도 안이에요. (B) `next-flow.md` 에 직접 넣고 한도를 올려요: 파일이 하나지만 한도를 또 올려야 해요(이미 한 번 올렸어요).

## Scope

- write: `.moai/specs/SPEC-20261125-agent-team-session-names/**`, `skills/moai/SKILL.md`, `skills/moai/references/routes.json`, `skills/moai/workflows/work.md`, `skills/moai/workflows/leader.md`, `skills/moai/references/flows/next-flow.md`, `skills/moai/references/flows/team-messages.md`(새 파일, D-5 A; SCOPE-REQUEST 로 더해요), `tests/lane_team_names.rs`(새 파일, 첫 낱말 `lane` 은 `tests/areas.toml` 에 있어요; SCOPE-REQUEST), `tests/lane_completion_handshake.rs`, `tests/next_flow_skill_scan.rs`, `tests/leader_route.rs`, `tests/mode_workflow_steps.rs`, `tests/skill_structure.rs`(이 글을 고정한 테스트가 바뀐 글에 맞춰야 할 때만)
- write (한 줄, line-lease): `skills/moai/references/source-manifest.json`, `tools/repo_scans.json`(새 테스트 파일 등록)
- read-only: `src/lane_registry.rs`, `src/ops/session_register.rs`, `src/hook_engine.rs`, `src/leader_pulse.rs`, `docs/lane-rules.md`, `docs/leader-rules.md`
- forbidden: `src/**` 쓰기, `interview.md`, `project.md`, `CLAUDE.md`, 전역 설치본과 실제 HOME, 네트워크, Python

## Plan

1. RED: `tests/lane_team_names.rs` 가 AC-1..AC-6 의 문장을 고정하고 현재 글에서 실패해요.
2. 글: `routes.json`, `SKILL.md`, `work.md`, `leader.md` 에 경로 낱말·이름·안내를, `team-messages.md` 에 메시지 표를, `next-flow.md` 에 한 줄 포인터와 채널 순서를 써요. GREEN.
3. 이 글을 고정한 기존 테스트(`leader_route`, `lane_completion_handshake`, `next_flow_skill_scan`, `mode_workflow_steps`, `skill_structure`)를 읽어 바뀐 글에 맞추고, 매니페스트 digest 를 line-lease 로 고쳐요.
4. 영수증 한 번(`moai check affected`), READY 에 SPEC digest 와 가드 뮤턴트 줄을 담아요. AC-7 은 설치 뒤 한 바퀴로 남겨요.

## Risks

- R-1: 바꾼 이름이 `ListAgents` 에 바로 보이는지는 실행으로 확인하지 못했어요. 글은 이름을 `ListAgents` 첫 줄로 읽어 확인하고, 다르면 REQ-5 로 가요. 사용자가 `/rename` 을 친 뒤 한 번 확인해야 해요(AC-7).
- R-2: 이름이 겹칠 수 있어요(실제 목록에 겹친 이름이 있어요). 글은 겹치면 짧은 id 로 주소를 쓰고 리더는 등록부의 세션 id 로 찾으므로 잘못된 행으로 가지 않아요(REQ-7).
- R-3: `next-flow.md` 의 여유가 151 바이트라서 이 SPEC 이 거기에 쓰는 것은 한 줄 포인터와 채널 순서 정리뿐이어야 하고, 그 줄은 이미 있는 문장을 줄여 넣어요. 넘으면 한도를 올리지 않고 `team-messages.md` 로 옮겨요.
- R-4: 다른 카드가 같은 스킬 폴더의 흐름 이름을 고쳐요. `next-flow.md` 와 매니페스트에서 겹칠 수 있어서 병합을 한 번 해요.
- R-5: 이 글이 `docs/lane-rules.md` 와 `docs/leader-rules.md`(카드 범위 밖) 와 어긋날 수 있어요. 어긋나는 문장은 후속으로 리더에게 알려요.

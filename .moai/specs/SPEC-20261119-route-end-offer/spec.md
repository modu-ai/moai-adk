---
id: SPEC-20261119-route-end-offer
title: "The route-end offer leaves out what cannot run, and Codex Default mode never carries a permission pick"
status: draft
revision: 1
created: 2026-11-19
---

# The route-end offer leaves out what cannot run, and Codex Default mode never carries a permission pick

## Intent

C-51 재감사(PR #3136, head `3dff58442d`, PASS)의 N9·N10이에요. 둘 다 질문 규칙이 비어 있던 자리예요.

**N9 — 경로 끝의 제안(route-end offer).** PROJECT와 UPDATE는 끝에서 "아직 `unchanged`가 아닌 단계"를 한 번에 골라 보라고 제안해요. 지금은 이렇게 비어 있어요.
- 되돌릴 수 없는 선택(case 2)과 평범한 단계가 한 선택지 묶음에 섞이면 어느 규칙이 적용되는지 글이 말하지 않아요.
- UPDATE 5단계는 "`blocked` 단계는 제안하지 않는다"고 말하는데 PROJECT의 같은 줄은 말하지 않아요.
- `skip` 단계와 명령이 없는 `needs_decision` 단계는 실행할 것이 없는데 제안 대상인지 글이 말하지 않아요.
- 기록된 HOLD 아래에 있는 단계가 제안에 들어가도 되는지 글이 말하지 않아요.

**N10 — Codex Default 모드.** 설치된 Codex 0.170.0 바이너리를 읽기만 해서(`strings`, 아무것도 실행하지 않음) 확인한 Default 모드 지침 원문이에요: *"Never use the `request_user_input` tool for permission requests or permission-related escalations."* 같은 지침이 "무응답이면 best judgment로 계속하라"고도 말해요. 질문 규칙의 Codex 행은 뒤쪽 문장만 다루고 앞쪽(권한 요청은 이 도구로 묻지 않는다)은 다루지 않아요. 런처가 `features.default_mode_request_user_input`를 켜는 날부터 문제가 돼요.

이 SPEC은 글 일곱 문장을 더하고 시험으로 고정해요.

## Non-goals

- 런처가 Default 모드 기능을 켜는 일(별도 카드).
- 질문 도구 목록·호스트별 호출 방식 바꾸기: 지금 글 그대로 둬요.
- `project-update-plan`이 단계를 어떻게 분류하는지(코드): 글이 이미 있는 분류를 따라요.
- OMP·Claude Code 행: Codex 행만 바꿔요.

## Requirements

- REQ-1: The question rules shall state that a question holding any case-(2) option is a case-(2) question as a whole, even when its other options are ordinary steps.
- REQ-2: The PROJECT route-end offer shall report a `blocked` step with what unblocks it instead of offering it.
- REQ-3: The PROJECT route-end offer shall report a `skip` step and a `needs_decision` step that prints no command with their reason instead of offering them.
- REQ-4: The PROJECT route-end offer shall leave out a step whose target is under a recorded HOLD and report it as held.
- REQ-5: The UPDATE re-run step shall carry the same three exclusions as REQ-2, REQ-3 and REQ-4.
- REQ-6: The Codex row of the question rules shall state that in Default mode a case-(2) pick is not asked through `request_user_input`, quoting the sentence of the installed Codex instructions.
- REQ-7: The Codex row shall state that this pick is asked as plain text with numbered choices, the recommended one first, followed by a wait for the reply.
- REQ-8: The update-route test shall require each sentence of REQ-1 to REQ-7 in the file that carries it.
- REQ-9 (REQ-8): IF a sentence of REQ-1 to REQ-7 is missing from its file, THEN the update-route test shall fail and name that file.
- REQ-10: The source manifest shall match the three edited skill files.

## Acceptance

- AC-1 (REQ-1, REQ-6, REQ-7, REQ-8, REQ-9): `questions.md`의 규칙 3과 Codex 행에 문장이 있고, 지우면 시험이 규칙 3 문장은 `questions.md lacks …`로, Codex 행 문장은 `questions.md Codex CLI row lacks …`로 실패해요.
  verify: `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool`
- AC-2 (REQ-2, REQ-3, REQ-4, REQ-5, REQ-8, REQ-9): `project.md`의 Completion Check와 `update.md` 5단계에 세 제외 문장이 있고, 지우면 시험이 파일 이름을 대며 실패해요.
  verify: `cargo test --locked --offline --test update_route the_route_end_offer_leaves_out_what_cannot_run`
- AC-3 (REQ-10): `source-manifest.json`이 세 파일의 현재 글과 맞아요.
  verify: `cargo test --locked --offline --test product_cli skill_package`
- AC-4: 이웃 검사.
  verify: `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test ac_command_lint`; `cargo test --locked --offline --test home_isolation_scan`

## Scope

- write: `.moai/specs/SPEC-20261119-route-end-offer/spec.md`, `skills/moai/references/internal/questions.md`, `skills/moai/workflows/project.md`, `skills/moai/workflows/update.md`, `skills/moai/references/source-manifest.json`, `tests/update_route.rs`
- read-only: 설치된 Codex 바이너리의 문자열(읽기만)
- forbidden: 실제 HOME 쓰기, 위 목록 밖의 `skills/moai/**`, `src/**`

## Plan

1. 시험을 먼저 더해요(RED): 문장이 아직 없어서 실패해야 해요.
2. 세 파일에 문장을 더해요.
3. `tools/source_manifest.py write`로 manifest를 맞춰요.
4. 시험이 통과(GREEN)하고, 문장을 하나씩 지우는 변이가 시험에 잡히는지 읽어요.
5. 이웃 검사, fmt, clippy.

## Risks

- UPDATE와 PROJECT가 같은 제안을 다른 말로 쓰면 한쪽만 고쳐져요 → 두 파일이 같은 세 구절(대소문자만 다름)을 갖도록 시험이 둘 다 읽어요.
- Codex 지침 원문은 버전마다 달라질 수 있어요 → 문장에 읽은 버전(0.170.0)을 적고, 시험은 그 원문 한 문장만 고정해요. 다른 버전이 문구를 바꾸면 시험이 아니라 이 글이 낡은 거예요.
- 평범한 선택까지 평문으로 돌려 버릴 수 있어요 → 글에 "평범한 선택과 Plan 모드는 도구를 그대로 쓴다"를 같이 적어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | N10의 근거는 설치된 Codex 0.170.0 바이너리의 문자열 | 이 컴퓨터에 실제로 설치된 것이라 "설치된 지침에 맞춰 확인"이라는 기준에 맞아요; 저장소가 다른 곳에 적어 둔 0.169.0은 상태줄 항목 목록의 기준이라 별개예요 | 없음 |
| D-2 | case-(2) 선택은 Default 모드에서 평문으로 | 도구 설명이 "무응답이면 best judgment로 계속"이라 무응답이 선택이 아니라는 규칙 3과 부딪혀요; 평문은 답이 올 때까지 기다려요 | 이 문장을 지우면 돼요 |
| D-3 | `needs_decision`은 명령이 없을 때만 제안에서 빼요 | UPDATE 2단계가 명령을 찍는 `needs_decision`에는 따로 yes를 받으라고 이미 말해요 | 없음 |
| D-4 | 기록된 HOLD 아래 단계는 제안에서 빼고 "held"로 보고 | SKILL.md가 HOLD 아래 쓰기를 막아요(`Never write a target … under a recorded HOLD`) | 없음 |

## Result

- AC-1: met — `cargo test --locked --offline --test update_route every_question_is_a_selection_through_the_hosts_question_tool`: RED(두 문장이 없어 2 failed, 7 passed) → GREEN(`update_route` 9 passed, 0 failed). 규칙 3의 "A question that holds any case-(2) option…"와 Codex 행의 Default 모드 평문 질문 문장(0.170.0 원문 인용 포함)이 `questions.md`에 있어요.
- AC-2: met — `cargo test --locked --offline --test update_route the_route_end_offer_leaves_out_what_cannot_run`: GREEN. `project.md` Completion Check와 `update.md` 5단계가 `blocked`·`skip`·`needs_decision`(명령 없음)·HOLD 제외와 case-(2) 전체 문장을 갖고, 둘 다 읽어요. 이 시험을 짜는 동안 옛 시험이 `project.md`에 "report a `blocked` step with what unblocks it"를 그대로 요구한다는 걸 만나, 그 문장을 남기고 새 조항을 붙였어요.
- AC-3: met — `python3 tools/source_manifest.py check` → `{"conflicted": false, "changed": [], "dropped": [], "ok": true}`; `cargo test --locked --offline --test product_cli` 79 passed(`skill_package` 포함).
- AC-4: met — 이 AC의 verify 명령과 이웃 시험이 통과했어요(`update_route` 9, `ac_command_lint`, `home_isolation_scan`, `spec_cli` 17, `workflow_size_budget` 4, `sdd_guidance` 6, `core_rules_2876` 6, `codex_question_tool` 4, `question_bank_hosts` 1). `cargo fmt --check` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0.

변이 시험(문장을 하나씩 지우고 `--test update_route`): M1 규칙 3 전체-질문 문장 → KILLED(`every_question_is_a_selection_through_the_hosts_question_tool`, `questions.md lacks`); M2 Codex "not asked through `request_user_input`" → KILLED(같은 시험, `questions.md Codex CLI row lacks`); M3 "Ordinary picks and Plan mode keep the tool" → KILLED(같은 시험, `questions.md Codex CLI row lacks`; 시험 메시지에 파일 이름을 넣은 뒤 M2·M3 를 다시 돌려 이 문구를 읽었어요); M4 `project.md` HOLD 조항 → KILLED(`the_route_end_offer_leaves_out_what_cannot_run`, `workflows/project.md lacks`); M5 `update.md` skip·needs_decision 조항 → KILLED(같은 시험, `workflows/update.md lacks`). 5/5 KILLED, 변이마다 `git checkout --` 뒤 작업 트리 clean.

검사하지 않은 것: 파이썬 시험 `tests/repo_scans_test.py`는 이 작업 폴더에 `target/debug/moai-cli`가 없어 19개가 "moai-cli not built"로 실패해요(환경 문제, 내 변경과 무관). 결과를 증거로 쓰지 않고 CI 실행을 인용해요. 실제 Codex Default 모드에서의 동작은 돌려 보지 않았어요(글 규칙만 고정).

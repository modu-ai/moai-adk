---
id: SPEC-20261118-codex-question-tool
title: "Codex and OMP question tools: sourced facts and the run evidence the card still lacks"
status: draft
revision: 1
created: 2026-11-18
---

# Codex and OMP question tools: sourced facts and the run evidence the card still lacks

## Intent

사용자 결정(2026-11-16): `/moai project`·`update`는 호스트의 질문 도구로 선택지를 보여 줘야 하고, 글로 된 목록으로 묻고 턴을 끝내면 안 돼요.
도구 이름·사용법·호출 규칙은 이미 `skills/moai/references/internal/questions.md`에 있어요(리더 대조 2026-11-18). 남은 두 가지가 이 SPEC의 일이에요.
(1) 그 표의 Codex·OMP 칸이 공식 출처의 어느 글에 기대는지, URL과 읽은 UTC 시각과 함께 남기기. (2) Codex에서 `$moai project`를 실제로 돌린 기록이 없다는 틈을
`not_run`과 그 이유로 말하고, 돌리지 않고도 잴 수 있었던 것(모델에 보이는 도구 목록)을 출력과 함께 남기기. 표 자체와 그것을 고정한 시험은 바꾸지 않아요.

## Non-goals

- `questions.md` 문구 변경: 표의 Codex 칸은 `tests/update_route.rs`가 고정하고 있고, 조사 결과와 어긋나지 않아요.
- `moai codex` 런처가 `features.default_mode_request_user_input`을 켜게 하는 일: 이 카드의 기준에 없어요. 표의 "launcher is to turn it on" 문장은 그대로 두고 후속으로 보고해요.
- 사용자 계정으로 모델을 부르는 Codex 실제 대화 실행: 외부 전송과 사람이 대화창에서 고르는 일이 필요해요(Decisions D-2).

## Requirements

- REQ-1: The research note shall quote the Codex CLI question tool from its official source with the source URL and the UTC read time and the checksum of the fetched content.
- REQ-2: The research note shall quote the OMP question tool from its official source with the source URL and the UTC read time and the checksum of the fetched content.
- REQ-3: The Codex CLI and OMP rows of the questions reference shall agree with the research note on the tool names and the flag default.
- REQ-4: The research note shall state the Codex run of the project route as not_run with its reason and record the read-only measurement that was made instead.

## Acceptance

- AC-1 (REQ-1, REQ-2): `research.md`의 Codex CLI·OMP 절마다 인용, 출처 URL, 읽은 UTC 시각이 있고, 출처 항목마다 가져온 내용의 sha256(저장소 규칙 `AGENTS.md` 권위 4)이 있어요.
  verify: `cargo test --locked --offline --test codex_question_tool research_note_quotes_each_host_from_its_source_with_url_and_read_time`; `cargo test --locked --offline --test codex_question_tool research_note_records_a_checksum_for_every_source_it_cites`
- AC-2 (REQ-3): 표의 두 칸과 조사 글이 도구 이름과 플래그 기본값(꺼짐)에서 같아요.
  verify: `cargo test --locked --offline --test codex_question_tool research_note_agrees_with_the_hosts_rows_of_questions_md`
- AC-3 (REQ-4): `Codex run` 절이 `not_run`과 이유를 적고 측정(`codex debug prompt-input`, `codex features list`)의 출력과 시각을 남겨요.
  verify: `cargo test --locked --offline --test codex_question_tool codex_run_gap_is_stated_as_not_run_with_its_reason`
- AC-4 (REQ-1…REQ-4): 이웃 검사.
  verify: `cargo test --locked --offline --test update_route`; `cargo test --locked --offline --test question_bank_hosts`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-codex-question-tool/spec.md`, `.moai/specs/SPEC-20261118-codex-question-tool/research.md`, `tests/codex_question_tool.rs`
- read-only: `skills/moai/references/internal/questions.md`, `tests/update_route.rs`, 설치된 `codex`·`omp`의 읽기 전용 명령
- forbidden: 실제 HOME과 Codex 설정·인증 파일의 쓰기와 복사, 모델을 부르는 Codex 실행, 네트워크 쓰기

## Plan

1. 시험 먼저: `tests/codex_question_tool.rs`의 세 시험이 `research.md`가 없어서 실패하는 것(RED)을 읽어요.
2. 공식 출처에서 읽은 사실과 로컬 측정을 `research.md`에 적어 GREEN으로 만들어요.
3. 이웃 검사를 돌려요.

## Risks

- 공식 문서가 소스보다 얇을 수 있어요 → 항목마다 출처가 문서인지 소스인지 적고, 확인 못 한 것은 `unverified`로 남겨요.
- 조사 글과 표가 나중에 어긋나요 → AC-2 시험이 이름과 기본값을 맞춰 봐요.
- 표의 도구 사실이 Codex 새 버전에서 바뀌어요 → 측정에 버전(`codex-cli 0.170.0`)과 시각을 적어 둬서 낡았는지 알 수 있어요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `questions.md`와 `source-manifest.json`은 고치지 않아요 | 표는 이미 조사 결과와 맞고 `tests/update_route.rs`가 고정해요; 고치면 매니페스트 해시까지 따라 바뀌어요 | 후속 SPEC에서 표에 출처 줄을 더하기 |
| D-2 | Codex 실제 `$moai project` 실행은 하지 않고 `not_run`으로 둬요 | 사용자 계정으로 모델을 부르고(외부 전송) 대화창에서 사람이 골라야 해서 이 레인이 결정할 일이 아니에요; 카드 기준 4가 이 경우 `not_run`과 이유를 허용해요 | 사용자가 실행을 고르면 후속 카드에서 캡처를 더하기 |
| D-3 | 측정은 모델을 부르지 않는 `codex debug prompt-input`과 `codex features list`로 해요 | 읽기 전용이고 같은 기계의 같은 바이너리가 모델에 보내는 도구 목록과 지시 글을 그대로 보여 줘요 | 없음(읽기 전용) |

## Result

- AC-1: met — `cargo test --locked --offline --test codex_question_tool` (RED 먼저: 3 failed, exit 101, `research.md` 없음; 체크섬 시험도 RED 먼저: 1 failed, exit 101, 줄이 없어서). `research.md`의 Codex CLI 절과 OMP 절이 인용·출처 URL·UTC 읽은 시각과 출처별 sha256을 가져요. Codex 공식 문서(learn.chatgpt.com/docs)에는 이 도구가 없어서 Codex 사실은 `rust-v0.170.0` 소스 근거이고 `kind: source`로 적었어요.
- AC-2: met — 같은 시험. 표의 Codex·OMP 칸과 조사 글이 `request_user_input`, `features.default_mode_request_user_input`(기본 꺼짐), `ask`, `ask.timeout`에서 같아요. `questions.md`는 바꾸지 않았어요.
- AC-3: met — 같은 시험. 실제 Codex `$moai project` 실행은 `not_run`이고 이유를 적었어요. 대신 `codex features list`와 `codex debug prompt-input`(codex-cli 0.170.0, 모델 호출 없음)의 출력을 시각과 함께 남겼고 도구 설명이 플래그에 따라 "only available in Plan mode."에서 "Default or Plan mode."로 바뀌는 것을 확인했어요.
- AC-4: met — `update_route` 8 passed; `question_bank_hosts` 1 passed; `codex_question_tool` 4 passed; `cargo fmt --check` exit 0; `python3 -m unittest tests.ci_scope_test` OK(19 tests).
- 남은 것: 대화창이 실제로 화면에 뜨는 Codex 실행 캡처(사람이 `/plan`에서 `$moai project`를 입력해야 해요)와, `moai codex` 런처가 플래그를 켜게 하는 일(표의 "launcher is to turn it on" 문장)은 이 SPEC 밖이에요. SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.

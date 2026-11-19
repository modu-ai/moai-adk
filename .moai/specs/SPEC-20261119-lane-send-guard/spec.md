---
id: SPEC-20261119-lane-send-guard
title: "A lane send reads the target's screen first and sends nothing unless the lane is working or idle"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-lane-screen-guard
---

# A lane send reads the target's screen first and sends nothing unless the lane is working or idle

## Intent

카드 C-120 의 둘째 조각이에요. 2026-11-19 에 레인 화면에 `git reset --hard` 의 Yes/No 확인 창이 떠 있는데 지침 글을 Enter 와 함께 보내서 강조된 Yes 가 눌렸어요(C-108 노트 2813). `orca terminal send` 는 대상 화면을 보지 않고 글자를 넣어요(`orca terminal send --help`, 2026-11-19 읽음). 이 SPEC 은 `moai lane-send` 를 만들어요: 보내기 전에 대상 터미널의 **그려진 화면**(rendered screen)을 읽고, 바탕 SPEC(`SPEC-20261119-lane-screen-guard`)의 분류기로 상태를 정해서 `working`·`idle` 일 때만 보내요. 확인 창·선택 창·사용 한도·접근 불가·모름이면 아무것도 보내지 않고 이유를 말해요.

이 SPEC 은 분류기가 병합된 뒤에 구현해요(분류 모듈 `src/lane_screen.rs` 에 기대요). 이 조각(1)은 SPEC 뿐이고 코드와 시험은 만들지 않아요.

## Interview

- Q1: 보낼 글의 안전 검사는요? → A: `src/lane_duty.rs` 의 "다른 터미널에 입력으로 들어가는 글" 규칙(제어 문자·줄 구분자·양방향 문자 거절)을 그대로 써요. 화면 상태 검사와 글 검사가 둘 다 통과해야 보내요.
- Q2: 화면을 읽은 뒤 보내기 전에 화면이 바뀌면요? → A: 읽기와 보내기 사이가 비어 있는 점은 남아요(Risks). 보내기 직전에 한 번 더 읽고 둘의 상태가 다르면 보내지 않아요.
- Q3: 거절하면 종료 코드는요? → A: 0 이 아니에요(성공 흉내를 내지 않아요). 출력에 상태·이유·읽은 화면 줄 하나가 있어요.
- Q4: 시험은 어떻게 하나요? → A: 가짜 `orca` 프로그램(`MOAI_ORCA_BIN` 으로 바이너리를 바꿔요)이 읽기 요청에는 녹화한 화면을 주고, 보내기 요청은 받은 글을 파일에 남겨서 "보내지 않았음"을 파일이 비었는지로 확인해요.

## Non-goals

- 확인 창에 대신 답하기: 이 도구는 보내지 않는 쪽만 맡아요.
- 한도가 풀릴 때까지 기다렸다 보내기(재시도, 대기열): 거절하고 해제 시각을 보여 주기까지예요.
- `hc send` 와 `orca orchestration send` 의 동작 변경: 이 도구는 `orca terminal send` 길 앞에만 서요.

## Requirements

- REQ-1: When `moai lane-send` is given a terminal and a text, the command shall read that terminal's rendered screen and classify it before it sends anything.
- REQ-2 (REQ-1): IF the screen is classified as anything other than `working` or `idle`, THEN `moai lane-send` shall send nothing and exit with a non-zero code that names the state and the screen line it was read from.
- REQ-3: When the screen is classified as `working` or `idle`, `moai lane-send` shall send the text to that terminal and print the state it sent under.
- REQ-4 (REQ-3): IF the screen read just before the send is classified differently from the first read, THEN `moai lane-send` shall send nothing and exit with a non-zero code.
- REQ-5 (REQ-2): IF the screen is classified as `usage_limit`, THEN `moai lane-send` shall print the reset time of the limit in UTC with its refusal.
- REQ-6 (REQ-1): IF the text carries a control character or a line separator that the input-safety rule of `lane-duty` refuses, THEN `moai lane-send` shall send nothing and exit with a non-zero code.
- REQ-7 (REQ-1): IF the terminal cannot be read or the Orca program is absent, THEN `moai lane-send` shall send nothing and exit with a non-zero code that says the lane was unreachable.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4): 확인 창·선택 창·한도·셸·모름 화면에는 아무것도 보내지 않아요(가짜 `orca` 의 보내기 기록 파일이 비어 있고 종료 코드가 0 이 아니에요). 첫 읽기가 대기 중이고 둘째 읽기가 확인 창이면 보내지 않아요(사고 1).
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_refuses`
- AC-2 (REQ-3, REQ-5): 작업 중·대기 중 화면에는 글이 한 번 그대로 가요. 한도 화면 거절에 해제 시각이 UTC 로 있어요.
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_sends`
- AC-3 (REQ-6, REQ-7): 안전하지 않은 글과 읽지 못한 터미널은 보내지 않아요.
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_unsafe`
- AC-4 (REQ-1…REQ-7): 이웃 시험이 통과해요.
  verify: `cargo test --locked --offline --test lane_duty_line`

## Scope

- write(이 조각): `.moai/specs/SPEC-20261119-lane-send-guard/**`
- write(구현 조각 B, 쓰기 범위를 새로 받아요, 분류기 병합 뒤): `src/ops/lane_send.rs`, `src/ops/mod.rs`·`src/main.rs`·`src/help.rs`(등록), `tests/lane_send_cli.rs`, `tests/fixtures/lane_screens/**`(바탕 SPEC 의 화면 재사용)
- read-only: `src/lane_screen.rs`(분류기), `src/lane_duty.rs`(입력 안전 규칙)

## Plan

1. (이 조각) SPEC.
2. 구현 조각 B: 가짜 `orca` 로 `lane_send_*` 시험을 먼저 써서 실패를 보고 `moai lane-send` 를 만든 뒤 이웃 시험과 변이 시험(상한 12).

## Risks

- 읽기와 보내기 사이: 두 번 읽어도 두 읽기 사이에 확인 창이 뜰 수 있어요. 그 틈을 0 으로 만들지 못해요; 틈을 줄이는 것이고, 보내기 쪽의 근본 보호는 `orca terminal send` 자체에 없어요.
- 분류기가 모르는 화면은 `unknown` 이라 거절돼요: 안전한 쪽이지만 정상 레인이 막힐 수 있어요. 막히면 사람이 `orca terminal send` 를 직접 써요.
- 바탕 SPEC 의 화면 녹화(D-5)가 없으면 시험의 화면이 합성이에요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 보내기 직전에 한 번 더 읽고 다르면 거절 | 읽고 보내는 사이의 변화를 줄여요 | 한 번만 읽기 |
| D-2 | 거절은 종료 코드 0 이 아님 | 호출자가 성공으로 읽지 않아요 | 0 + 메시지 |

## Result

이 조각은 SPEC 뿐이에요: 코드·시험을 만들지 않았어요.

- AC-1: 아직 안 했어요(구현 조각 B).
- AC-2: 아직 안 했어요(구현 조각 B).
- AC-3: 아직 안 했어요(구현 조각 B).
- AC-4: 아직 안 했어요(구현 조각 B).

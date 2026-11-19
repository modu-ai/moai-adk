---
id: SPEC-20261119-lane-screen-guard-followups
title: "A prose list of numbered Yes and No lines reads as a prompt on purpose, and a reset time keeps its year between 1970 and 9999 in UTC"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-lane-screen-guard
---

# A prose list of numbered Yes and No lines reads as a prompt on purpose, and a reset time keeps its year between 1970 and 9999 in UTC

## Intent

카드 C-128(리더 배차 2026-11-19)은 C-120 조각 A 의 재감사(카드 C-120 노트 3256, head `8071820b9f`)가 막지 않고 남긴 두 가지예요. R-N2: 평범한 글 속 번호 목록이 `1. Yes, …` 줄과 `2. No, …` 줄로 시작하면 보통 레인 화면도 `permission_prompt` 로 읽혀요. 안전한 쪽으로 틀리는 것이라 코드는 두고, 그렇다는 것을 SPEC 에 적고 시험으로 못박아요. R-N4: `src/lane_screen.rs` 의 `parse_reset` 은 연도 범위(1970~9999)를 글에 적힌 **현지 날짜**로만 검사해서, 9999 년 끝의 시각이 음수 오프셋이면(`9999-12-31 23:59 -0800`) UTC 로 바꾼 결과가 10000 년이 되어 다섯 자리 해가 찍혀요. 이 조각은 UTC 로 바꾼 결과의 연도도 같은 범위로 검사해요. 바탕 SPEC `SPEC-20261119-lane-screen-guard` 의 분류기(조각 A)를 고치지 않고 이어 가는 후속이라 자식 SPEC 이에요.

## Interview

- Q1: R-N2 는 코드를 바꾸나요? → A: 아니요(D-1). 번호 붙은 `Yes` 줄과 `No` 줄 한 쌍은 확인 창과 글 속 목록을 가를 글자가 없어요. 보내기 보호(`lane-send`)가 "Enter 가 Yes 를 고르는" 사고에서 나왔으니 안전한 쪽으로 틀려요. 질문 줄을 요구하는 대안은 조각 A 감사 B1 이 새는 화면을 찾아서 이미 버렸어요.
- Q2: 그 대가는 무엇인가요? → A: 레인이 대화에 `1. Yes, …`/`2. No, …` 목록을 쓰면 그 목록이 화면에 있는 동안 `lane-send` 가 그 레인에 보내지 않아요(`permission_prompt` 거절). 목록이 화면 위로 올라가면 풀려요. 리더가 그 거절을 보고 화면을 읽으면 되고 잘못 보내는 일은 없어요.
- Q3: R-N4 의 범위 검사는 어디에 두나요? → A: `at_zone` 이 하루 수(`days_from_civil`)를 만든 직후에 UTC 초의 연도를 `year_of_days` 로 읽어 1970..=9999 밖이면 `None` 으로 해요(D-2). 기존 검사(현지 날짜의 연도 범위)는 `days_from_civil` 을 `i64` 한계에서 멀리 두는 역할이라 남겨요.
- Q4: 아래쪽 경계는요? → A: 1970-01-01 00:00:00 UTC 는 읽히고(`1970-01-01 08:00 +0800`), 1분 이른 `07:59 +0800` 은 1969 년이라 `unknown`("the date is impossible", 기존 문구)이에요. 위쪽도 같아요: `9999-12-31 15:59 -0800`(UTC 23:59)은 읽히고 `16:00 -0800` 은 `unknown`.
- Q5: 어떤 파일에 쓰나요? → A: 카드 범위의 `src/lane_screen.rs`·`tests/lane_screen_cli.rs` 와, 가지 이름과 짝인 이 SPEC 폴더예요(훅이 가지의 SPEC 이 먼저 있기를 요구해요). R-N2 의 화면은 `tests/fixtures/lane_screens/` 가 범위 밖이라 시험 파일 안의 문자열이에요.

## Non-goals

- 번호 붙은 `Yes`·`No` 목록을 확인 창과 가르는 새 규칙: 가를 글자가 없어요(D-1).
- 해제 시각의 다른 해석 규칙 변경: 시간대 목록·날짜 형식은 그대로예요.
- 조각 B·C 의 `lane-send`·`--scan` 코드: 이 가지는 조각 C 의 head 위에 있지만 그 파일들은 건드리지 않아요.

## Requirements

- REQ-1: IF the screen holds a numbered line that starts with `Yes` and a numbered line that starts with `No` inside ordinary prose (a list the lane wrote to the conversation, not a dialog), THEN `moai lane-screen` shall report `permission_prompt`, the same as for a dialog.
- REQ-2: IF the reset time written in a limit message, once turned into UTC, falls before the year 1970 or after the year 9999, THEN `moai lane-screen` shall report the reset time as unknown together with the printed text, never a time whose year has other than four digits.

## Acceptance

- AC-1 (REQ-1): 평범한 글 속 `1. Yes, …`·`2. No, …` 줄이 있는 보통 레인 화면(대기 표지와 작업 중 표지 둘)이 `permission_prompt` 이고 근거 줄이 그 `Yes` 줄이에요. 번호가 없는 같은 두 줄은 `idle` 로 남아요. `lane-send` 는 `permission_prompt` 에 보내지 않으므로 같은 화면의 보내기도 거절되지만, 이 가지에는 그 보내기 시험이 없어요(`tests/lane_send_cli.rs` 는 C-120 의 범위이고, 같은 상태를 조각 B 의 `lane_send_refuses_numbered_yes_no_rows_after_any_prefix_and_sends_nothing` 이 시험해요).
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_state_a_prose_list_of_numbered_yes_and_no_lines_reads_as_permission_prompt`
- AC-2 (REQ-2): `9999-12-31 23:59 -0800` 같은 시각은 `unknown` 과 원문이고, 경계는 위쪽(UTC 9999-12-31 23:59:00 은 읽히고 1분 뒤는 `unknown`)과 아래쪽(UTC 1970-01-01 00:00:00 은 읽히고 1분 앞은 `unknown`) 둘 다 짝으로 있어요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_reset_a_time_outside_the_years_1970_to_9999_in_utc_is_unknown` `cargo test --locked --offline --lib lane_screen`
- AC-3 (REQ-1, REQ-2): 이웃 시험과 린트가 통과해요.
  verify: `cargo test --locked --offline --test lane_screen_cli` `cargo test --locked --offline --test lane_send_cli`

## Scope

- write: `.moai/specs/SPEC-20261119-lane-screen-guard-followups/**`, `src/lane_screen.rs`(`parse_reset` 의 `at_zone` 에 UTC 연도 검사 한 곳, 단위 시험), `tests/lane_screen_cli.rs`
- read-only: 그 밖의 파일(조각 A 의 SPEC·문서 `lane_screen.md` 는 C-120 의 범위라서 건드리지 않아요)

## Plan

1. 이 SPEC. 2. 시험 먼저: R-N2 의 화면 시험(현재도 통과해야 해요, 못박기)과 R-N4 의 경계 시험(수리 전에는 실패해야 해요, 다섯 자리 해). 3. `at_zone` 에 UTC 연도 검사. 4. 이웃 시험과 변이 시험.

## Risks

- `lane-send` 를 막는 대가(D-1)는 줄지 않아요. 레인이 그런 목록을 자주 쓰면 리더가 거절 글을 자주 봐요. 그때는 질문 요구가 아니라 화면 위쪽 줄(대화 기록)과 선택 창을 가르는 새 표지를 녹화 화면으로 찾아야 해요.
- UTC 연도 검사는 `Reset::Utc(t)` 가 4자리 해로만 찍히도록 막는 마지막 문이에요. 다른 경로로 `format_utc` 에 큰 초가 들어가는 일은 이 SPEC 이 막지 않아요(`--home` 쪽 `resets_at` 은 `parse_utc` 를 거쳐요).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 평범한 글 속 번호 `Yes`·`No` 목록은 `permission_prompt` 로 둔다(코드 변경 없음, 문서와 시험만) | 확인 창과 가를 글자가 없고 Enter 가 `Yes` 를 고르는 쪽이 사고였어요(안전한 쪽으로 틀림). 되돌릴 수 있어요(범주 none) | 질문 줄을 요구(조각 A 감사 B1 이 새는 화면을 찾음) |
| D-2 | 연도 범위는 UTC 로 바꾼 초에서 한 번 더 검사한다 | 현지 날짜만 보면 음수·양수 오프셋이 경계를 넘겨 5자리 해(위쪽)나 1969 년(아래쪽)이 나와요 | `format_utc` 쪽에서 막기(출력 한 곳이라 막지만 값이 이미 `Reset::Utc` 로 나가요) |

## Result

카드 C-128 의 구현(코드·시험 커밋 `b0de5eda3b`, 2026-11-19 에 `date -u` 로 읽은 시계)이에요. 시험은 RED 먼저: R-N4 의 새 시험 둘(단위 `a_time_whose_utc_year_leaves_1970_to_9999_is_unknown`, CLI `lane_screen_reset_a_time_outside_the_years_1970_to_9999_in_utc_is_unknown`)이 수리 전 코드에서 실패했고 R-N2 의 시험은 코드가 이미 그렇게 읽어서 처음부터 통과해요(못박기 시험). 실제 `orca` 는 부르지 않았어요.

- AC-1: 충족 — `cargo test --locked --offline --test lane_screen_cli lane_screen_state_a_prose_list_of_numbered_yes_and_no_lines_reads_as_permission_prompt` 가 통과해요(대기 표지·작업 중 표지 둘 다 `permission_prompt`, 근거 줄이 `1. Yes, keep …` 줄, 번호 없는 같은 두 줄은 `idle`). 결정 D-1 이 코드를 바꾸지 않는 이유와 대가를 적어요.
- AC-2: 충족 — `... --test lane_screen_cli lane_screen_reset_a_time_outside_the_years_1970_to_9999_in_utc_is_unknown` 이 통과해요(`9999-12-31 23:59 -0800`·`16:00 -0800`·`23:59:59 UTC-00:01`·`1970-01-01 07:59 +0800`·`00:00 +0100` 은 `utc: null`, `why: the date is impossible`, 원문; 경계 `9999-12-31 15:59 -0800` 과 `23:59 UTC` 는 `9999-12-31T23:59:00Z`, `1970-01-01 08:00 +0800` 은 `1970-01-01T00:00:00Z`), 단위 시험은 초 값(`253402300740`·`0`)과 월 이름 꼴(`Dec 31, 9999 at 11:59pm (UTC-8)`)까지 닫아요. 수정은 `src/lane_screen.rs` 의 `at_zone` 한 곳이에요.
- AC-3: 충족 — `--test lane_screen_cli` 40 passed, `--test lane_send_cli` 19, `--test lane_status_cli` 25, `--test lane_duty_line` 8, `--lib lane_` 207 passed. 변이 4개(UTC 연도 검사 빼기, 위쪽 연도 10000 으로, 아래쪽 1969 로, 질문 없는 번호 Yes·No 쌍을 프롬프트로 안 읽기)가 모두 KILLED(마지막 변이는 R-N2 시험을 포함한 6개 시험을 실패시켰어요), 변이마다 `git checkout` 으로 복원하고 끝나고 깨끗해요. 여섯 검사와 같은 접두 시험 대상의 종료 코드는 감사 요청 노트에 있어요.
- 돌리지 않은 것: 전체 cargo 시험, 파이썬 discover(CI), nextest, 실제 `orca` 터미널 실행. 이 가지는 조각 C 의 head `66dda56b16` 위에 있어서 그 감사(노트 3395)가 PASS 해야 병합 순서가 맞아요.
- 쓰기 범위 메모: 카드 C-128 은 `src/lane_screen.rs`·`tests/lane_screen_cli.rs`·조각 A SPEC 폴더만 받았는데 훅이 가지 이름과 짝인 SPEC 을 요구해서 이 SPEC 폴더를 더했어요(예약 generation 2, 카드 노트 3399). 조각 A SPEC 폴더는 이 가지에서 쓰지 않았어요.

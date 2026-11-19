---
id: SPEC-20261119-lane-screen-guard
title: "Read a lane's rendered screen into a state, with a usage limit's reset time in UTC"
status: draft
revision: 1
created: 2026-11-19
---

# Read a lane's rendered screen into a state, with a usage limit's reset time in UTC

## Intent

카드 C-120(리더 배차 2026-11-19, 조각 1 = 계획과 SPEC 만): 2026-11-19 에 두 사고가 있었어요. (1) 리더가 레인에 지침 글을 Enter 와 함께 보냈는데 그 레인 화면에는 `git reset --hard` 의 Yes/No 확인 창이 떠 있었고, Enter 가 강조된 Yes 를 골라 reset 이 실행됐어요(C-108 노트 2813, reflog 2026-11-19 15:31:03 +0900). (2) 레인 여덟 개가 사용 한도(HTTP 429)로 멈췄는데 사용자가 말해서야 알았고, 화면에 적힌 해제 시각을 한국 시각으로 읽었지만 실제로는 UTC+8 이라서 해제 요청이 한 시간 일찍 나가 사라졌어요.

카드 하나를 SPEC 셋으로 나눠요(한 SPEC = 한 가지 = 한 PR, 요구가 12개를 넘으면 `spec check` 가 쪼개라고 해요). 이 SPEC 이 부모이자 바탕인 **분류기**예요: `moai lane-screen` 은 레인의 **그려진 화면**(rendered screen, 터미널이 지금 보여 주는 한 장면)을 읽어 `permission_prompt`·`input_pending`·`usage_limit`·`working`·`idle`·`unreachable`·`unknown` 중 하나로 분류하고, 사용 한도면 해제 시각을 UTC 로 보여 줘요. 자식 SPEC 둘이 이 분류기 위에 서요: `SPEC-20261119-lane-send-guard`(`moai lane-send` 가 확인 창·한도·셸·모름이면 아무것도 보내지 않아요)와 `SPEC-20261119-lane-screen-scan`(`moai lane-screen --scan` 이 떠 있는 터미널 중 한도로 멈춘 레인과 해제 시각을 한눈에 보여 줘요). 세 SPEC 문서는 이 계획 조각의 같은 가지에 함께 있어요.

이미 있는 것(tip `d25ecb1ba1` 에서 읽음)에 맞춰요: `src/lane_status.rs` 가 `PaneObservation { state, source, observed_at, until }` 와 상태 이름 `working`·`permission_prompt`·`input_pending`·`unreachable`·`usage_limit` 를 호출자 입력으로 받아요("pane 상태는 여기서 관찰하지 못한다"고 머리말에 적혀 있어요). 새 도구가 그 자리를 채우므로 출력이 같은 모양과 이름이에요. `src/usage_limit.rs` 는 상태줄이 본 `resets_at`(Unix 시각, 시간대 문제가 없어요)을 `<HOME>/.moai-cli/state/usage-limits.json` 에 적어 둬요: 레인의 HOME 을 알면 화면 글자보다 이쪽이 정확해요. `src/lane_duty.rs` 는 "다른 터미널에 입력으로 들어가는 글"에서 제어 문자·줄 구분자·양방향 문자를 거절하는 규칙을 이미 가져요(자식 SPEC 이 써요).

이 조각(1)은 계획과 SPEC 뿐이고 코드와 시험은 만들지 않아요.

## Interview

- Q1: 도구는 어디에 두나요(`tools/` 인가 CLI 인가)? → A: Rust CLI 의 새 하위 명령 `lane-screen`·`lane-send` 예요(기본안, 리더가 바꿀 수 있어요, D-1). 근거: `lane-status` 입력과 같은 모양을 내야 하고, `lane-duty-line`(Rust)이 같은 분야의 안전 규칙을 가져요. 바깥 프로그램은 `forge` 의 가짜 `gh` 처럼 바이너리 이름을 환경 변수로 바꿔 시험해요.
- Q2: 어떤 터미널 길을 쓰나요? → A: 1차는 `orca terminal read --screen --json` 으로 읽고 `orca terminal send` 로 보내요(사고의 길이에요). `hc send` 는 이미 `--force` 없이는 막힌 프로그램에 입력하지 않으니(`hc send --help`, 2026-11-19 읽음) 이 도구의 대상이 아니에요. `orca orchestration` 의 `deny`(작업자의 권한 창을 이유와 함께 거절)는 이 레인들이 orchestration 작업자가 아니라서 쓰지 않아요(Non-goals).
- Q3: 어떤 상태에서 보내기를 거절하나요? → A: `permission_prompt`·`input_pending`·`usage_limit`·`unreachable`·`unknown` 이고 `working`·`idle` 만 보내요(D-3, 자식 `lane-send-guard`). 거절 이유 셋은 사고에서 나왔어요: 확인 창(사고 1), 한도(사고 2: 보내도 처리되지 않아요), 셸/알 수 없음(`hc send` 가 셸에 붙여 넣은 글이 실행된 2026-11-10 일).
- Q4: 해제 시각은 어떻게 UTC 로 바꾸나요? → A: 레인의 HOME 을 `--home` 으로 주면 `usage-limits.json` 의 `resets_at` 을 써요(시간대 문제 없음). 아니면 화면 글자(`resets Nov 20 at 7pm (Asia/Seoul)` 꼴)를 읽되, 아는 시간대(`UTC`·`GMT`·`UTC+H[:MM]`·`Asia/Seoul` 처럼 여름시간이 없는 것)만 바꾸고 모르면 "모름"과 원문을 그대로 보여요(D-4). 날짜·시간대 라이브러리가 의존성에 없어서(`Cargo.toml`) 시간대 표를 새로 들이지 않아요.
- Q5: 시험에 쓸 "녹화한 화면"은요? → A: 지금 저장소에는 확인 창이나 한도 화면의 원문이 없어요(`git grep` 로 확인). 이번에 실제로 읽은 것은 작업 중(`✳ Frosting… (1m 17s · ↓ 6.9k tokens …)` 와 `esc to interrupt`)과 대기 중(`❯` 입력 줄과 `⏵⏵ bypass permissions on (shift+tab to cycle)`) 두 모양뿐이에요. 확인 창과 한도 화면은 실제 원문을 리더가 주거나 다음에 생길 때 녹화해야 해요(D-5, 리더에게 요청).
- Q6: SPEC 은 몇 개로 하나요? → A: 셋이에요(D-6): 분류기(이 SPEC)·보내기 보호·목록. 요구가 15개가 되면 `spec check` 가 "12개 넘으면 `parent` 로 이어 쪼개라"고 경고하고, 한 SPEC = 한 PR 이라서 구현 조각마다 SPEC 이 하나씩 있어야 해요.

## Non-goals

- 확인 창에 대신 답하기(Yes/No 를 눌러 주기, `orca orchestration deny`): 이 도구는 보내지 않는 쪽만 맡아요.
- 한도로 멈춘 레인을 깨우거나 해제 시각에 맞춰 요청을 자동으로 보내기: 해제 시각을 정확히 보여 주는 데까지예요.
- `hc` 터미널과 `orca orchestration send`: 앞엣것은 이미 막힌 프로그램을 거절하고, 뒤엣것은 작업자 우편함이라 다른 길이에요.
- Codex·OMP 레인 화면의 표지: 읽은 적이 없는 화면이라 `unknown`(보내지 않음)으로 두고, 녹화한 뒤 더해요.
- 보내기와 목록: 자식 SPEC 이 맡아요(이 SPEC 은 분류만 해요).
- 달력·시간대 데이터베이스 전체: 아는 시간대 몇 개와 `UTC±H[:MM]` 만 바꿔요.

## Requirements

- REQ-1: When `moai lane-screen` reads one rendered lane screen, the command shall print the state `working`, `idle`, `permission_prompt`, `input_pending`, `usage_limit`, `unreachable` or `unknown` with the screen line that state was read from.
- REQ-2 (REQ-1): IF the screen shows a numbered choice dialog that offers both a Yes row and a No row, THEN `moai lane-screen` shall report `permission_prompt` even when the same screen also shows a working or idle marker.
- REQ-3 (REQ-1): IF the screen shows another numbered choice dialog or a question that waits for an answer, THEN `moai lane-screen` shall report `input_pending`.
- REQ-4: When the screen shows a usage-limit message, `moai lane-screen` shall report `usage_limit` and the reset time of that message in UTC.
- REQ-5 (REQ-4): IF the message names no zone or a zone `moai lane-screen` does not know, THEN `moai lane-screen` shall report the reset time as unknown together with the printed text instead of taking the printed time as local time or as UTC.
- REQ-6: Where a lane's home folder is given with `--home`, `moai lane-screen` shall take the reset time of a usage limit from the `resets_at` of that home's `usage-limits.json` in preference to the screen text.
- REQ-7 (REQ-1): IF the input is not a rendered screen (an Orca read whose source is not `screen` or no screen row at all) or matches no known marker, THEN `moai lane-screen` shall report `unknown` and never `idle`.
- REQ-8: The `pane` object that `moai lane-screen` prints shall carry the fields `state`, `source`, `observed_at` and `until` of the `lane-status` pane observation.
- REQ-9: The command `moai lane-screen` shall write no file and send no input to any terminal.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3): 녹화한 화면에서 상태가 맞게 나와요: 작업 중·대기 중·확인 창(Yes/No)·그 밖의 선택 창·한도·셸, 그리고 확인 창 위에 작업 중 표지가 함께 있어도 `permission_prompt` 예요. 상태마다 읽어 낸 화면 줄이 함께 나와요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_state`
- AC-2 (REQ-4, REQ-5, REQ-6): 한도 화면의 해제 시각이 UTC 로 나와요: `UTC`·`UTC+8`·`Asia/Seoul` 꼴은 바뀌고, 모르는 시간대와 시간대 없는 글은 "모름"과 원문이며, `--home` 의 `resets_at` 이 있으면 화면 글자보다 우선이에요. 같은 글이 UTC+8 이면 UTC+9 일 때보다 한 시간 늦게 나와요(사고 2).
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_reset`
- AC-3 (REQ-7, REQ-8, REQ-9): 그려진 화면이 아니거나 표지가 없거나 비어 있으면 `unknown` 이고(`idle` 이 아니에요), `pane` 객체가 `lane-status` 의 `PaneObservation` 으로 그대로 읽혀요. 도구는 파일을 쓰지 않아요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_unknown`
- AC-4 (REQ-1…REQ-9): 이웃 시험과 린트가 통과해요.
  verify: `cargo test --locked --offline --lib lane_screen` `cargo test --locked --offline --test lane_duty_line`
- AC-5 (REQ-1…REQ-9): 도구 설명 문서 한 쪽이 있어요.
  verify: manual: 문서 한 쪽이라 시험으로 고정하지 않아요(시험이 `docs/` 경로를 읽으면 `tools/ci_scope.py` 거부 목록에 더해야 해요). 감사가 `docs/design-review/appendix/operations/lane_screen.md` 를 읽어요.

## Scope

- write(이 조각, 계획과 SPEC 셋): `.moai/specs/SPEC-20261119-lane-screen-guard/**`, `.moai/specs/SPEC-20261119-lane-send-guard/**`, `.moai/specs/SPEC-20261119-lane-screen-scan/**`
- write(이 SPEC 의 구현 조각 A, 쓰기 범위를 새로 받아요): `src/lane_screen.rs`(분류·시간 변환, 순수 함수), `src/ops/lane_screen.rs`, `src/ops/mod.rs`·`src/lib.rs`·`src/main.rs`·`src/help.rs`(등록), `tests/lane_screen_cli.rs`, `tests/fixtures/lane_screens/**`(녹화한 화면), `docs/design-review/appendix/operations/lane_screen.md`
- read-only: `src/lane_status.rs`(`PaneObservation`·`PaneState`), `src/usage_limit.rs`(`usage-limits.json` 읽기), `src/forge/policy.rs`(`parse_utc`·`format_utc`)

## Plan

1. (이 조각) SPEC 셋. 리더가 위치(D-1)·거절 범위(D-3)·시간 규칙(D-4)·녹화 화면(D-5)을 정해요.
2. 구현 조각 A — 분류기(이 SPEC): 녹화한 화면 파일(`tests/fixtures/lane_screens/`)로 시험을 먼저 써서 실패를 보고(`lane_screen_state`·`lane_screen_reset`·`lane_screen_unknown`), `src/lane_screen.rs` 에 분류와 시간 변환을 만들고 `moai lane-screen` 을 등록해요. 입력은 표준 입력의 화면 글자거나 Orca 읽기 JSON(`result.terminal.source` 가 `screen` 이어야 해요).
3. 구현 조각 B — `SPEC-20261119-lane-send-guard`: A 가 병합된 뒤. 가짜 `orca`(바이너리 이름을 `MOAI_ORCA_BIN` 으로 바꿔요)로 `lane_send_*` 시험을 먼저 쓰고 `moai lane-send` 를 만들어요.
4. 구현 조각 C — `SPEC-20261119-lane-screen-scan`: A 가 병합된 뒤. `--scan` 이 `orca terminal list --json` 과 터미널마다 `read --screen --json` 으로 목록을 만들어요.
5. 문서 한 쪽과 `help` 의 예시, 이웃 시험과 변이 시험(위험한 것부터 상한 12).

## Risks

- 화면 표지는 호스트가 바꿀 수 있어요: Claude Code 화면의 표지(`❯` 선택 줄, `esc to interrupt`, 한도 문구)는 Claude Code 버전에 따라 달라질 수 있어요. 그래서 모르면 `unknown`(보내지 않음)이고, 시험은 녹화한 화면에 묶여요. 이번에 실제로 읽은 화면은 작업 중·대기 중 둘뿐이라 확인 창과 한도 화면은 녹화가 필요해요(D-5).
- 화면이 비어 있으면 모르는 거예요: 이번에 읽은 한 레인은 화면 줄이 3개뿐이었어요(`returnedLineCount: 3`). 내용이 적은 화면을 `idle` 로 추측하지 않고 `unknown` 으로 두어요.
- 시간대: 화면 글자는 해(年)와 시간대가 빠질 수 있어요. 해는 "지금 이후의 가장 가까운 그 날짜"로 정하고(`--now` 로 시험), 모르는 시간대는 UTC 로 바꾸지 않아요. 레인 HOME 의 `usage-limits.json` 이 있으면 그것을 써요.
- 화면을 읽는 일 자체가 다른 레인의 화면 내용을 도구 출력에 올려요: 출력에는 상태·근거 줄 하나·해제 시각만 두고 화면 전체는 남기지 않아요.
- 세 SPEC 의 순서: B·C 는 A 의 모듈(`src/lane_screen.rs`)에 기대요. A 가 병합되기 전에 B·C 의 구현을 시작하지 않아요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 도구는 Rust CLI 하위 명령 `lane-screen`·`lane-send` | 출력이 `lane-status` 의 `PaneObservation` 과 같은 모양이어야 하고, `lane-duty-line` 이 같은 분야의 안전 규칙을 가져요; 바깥 프로그램은 가짜 바이너리 패턴(`forge`)으로 시험해요. 되돌릴 수 있어요(범주 none) | `tools/lane_send.py`(저장소 안 개발 도구, 설치본에 안 들어가요) |
| D-2 | 1차 길은 `orca terminal read --screen --json` 과 `orca terminal send` | 사고가 난 길이고, `hc send` 는 이미 막힌 프로그램을 거절해요 | `hc`, `orca orchestration` |
| D-3 | 보내는 상태는 `working`·`idle` 뿐, 나머지는 거절(자식 `lane-send-guard`) | 확인 창(사고 1)·한도(사고 2)·셸(2026-11-10)·모름 모두 "보내면 잘못 읽히는" 상태예요 | 확인 창에서만 거절 |
| D-4 | 해제 시각은 `--home` 의 `resets_at` 우선, 화면 글자는 아는 시간대만 바꾸고 모르면 모름 | 사고 2 가 시간대 추측에서 났고 의존성에 날짜 라이브러리가 없어요 | 시간대 데이터베이스 의존성 추가 |
| D-5 | 시험 화면은 녹화한 실제 화면을 쓴다: 확인 창과 한도 화면은 리더가 원문을 주거나 다음에 녹화 | 지금 저장소에 원문이 없어요; 실제로 읽은 것은 작업 중·대기 중뿐이에요 | 인용한 문장으로 만든 합성 화면(녹화 전까지 `synthetic` 표시) |
| D-6 | SPEC 셋: 분류기(이 SPEC)·보내기 보호·목록, 구현도 세 조각 | 한 SPEC = 한 가지 = 한 PR 이고 `spec check` 가 요구 12개 초과를 쪼개라고 해요; 분류기가 둘의 바탕이에요 | SPEC 하나(요구 15개, 경고 1) |

## Result

이 조각은 계획과 SPEC 뿐이에요: 코드·시험을 만들지 않았어요. 아래는 계획을 세우며 읽은 것이에요(2026-11-19, tip `d25ecb1ba1`).

- AC-1: 아직 안 했어요(구현 조각 A).
- AC-2: 아직 안 했어요(구현 조각 A).
- AC-3: 아직 안 했어요(구현 조각 A).
- AC-4: 아직 안 했어요(구현 조각 A).
- AC-5: 아직 안 했어요(문서, `manual:`).

읽은 것: `orca terminal read --help`(`--screen`: 그려진 화면, 기본은 겹쳐 쌓인 출력), `orca terminal send --help`(보호 없음, "다른 에이전트에는 `orca orchestration send`"), `hc send --help`(`--force`, `--if idle|working`), `src/lane_status.rs`(`PaneObservation`·`PaneState`), `src/usage_limit.rs`(`resets_at`), `src/lane_duty.rs`(입력 안전 규칙), `Cargo.toml`(날짜 라이브러리 없음), 레인 화면 둘(`orca terminal read --screen --json`). 확인 창과 한도 화면의 원문은 저장소에 없어요.

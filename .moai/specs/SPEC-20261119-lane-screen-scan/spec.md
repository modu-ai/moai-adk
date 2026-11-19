---
id: SPEC-20261119-lane-screen-scan
title: "A scan lists every open terminal's state and the UTC reset time of each lane stopped by a usage limit"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-lane-screen-guard
---

# A scan lists every open terminal's state and the UTC reset time of each lane stopped by a usage limit

## Intent

카드 C-120 의 셋째 조각이에요. 2026-11-19 에 레인 여덟 개가 사용 한도로 멈췄는데 사용자가 말해서야 알았고, 화면에 적힌 해제 시각을 한국 시각으로 읽었지만 실제로는 UTC+8 이라 해제 요청이 한 시간 일찍 나갔어요. 이 SPEC 은 `moai lane-screen --scan` 을 만들어요: 열려 있는 터미널을 `orca terminal list --json` 으로 모아 터미널마다 바탕 SPEC(`SPEC-20261119-lane-screen-guard`)의 분류기를 돌려, 상태 표와 한도로 멈춘 레인의 해제 시각(UTC)을 한 번에 보여 줘요.

이 SPEC 은 분류기가 병합된 뒤에 구현해요. 이 조각(1)은 SPEC 뿐이고 코드와 시험은 만들지 않아요.

## Interview

- Q1: 터미널 목록은 어디서 얻나요? → A: `orca terminal list --json`(2026-11-19 읽음). 대상이 많아도 터미널마다 한 번씩 화면을 읽어요.
- Q2: 한 터미널이 읽히지 않으면요? → A: 그 줄만 `unreachable` 로 두고 나머지는 계속해요. 한 줄의 실패가 목록 전체를 막지 않아요.
- Q3: 해제 시각은 어디서 가져오나요? → A: 바탕 SPEC 의 규칙 그대로예요(`--home` 의 `resets_at` 우선, 화면 글자는 아는 시간대만 바꿔요). 레인마다 HOME 을 알 수 없으면 화면 글자만 써요.
- Q4: 출력은요? → A: 사람이 읽는 표(터미널·상태·해제 시각 UTC)와 `--json`(기계가 읽는 모양)이에요. 화면 전체는 남기지 않고 근거 줄 하나만 둬요.

## Non-goals

- 한도로 멈춘 레인을 깨우거나 해제 시각에 맞춰 자동으로 보내기.
- 상태를 파일에 기록하기: 읽기만 하고 아무것도 쓰지 않아요.
- 레인이 아닌 터미널의 구분: 목록에 있는 터미널을 모두 보여요.

## Requirements

- REQ-1: When `moai lane-screen --scan` runs, the command shall list every terminal that `orca terminal list` reports together with the state of its rendered screen.
- REQ-2 (REQ-1): IF a terminal cannot be read, THEN `moai lane-screen --scan` shall list that terminal as `unreachable` and continue with the remaining terminals.
- REQ-3 (REQ-1): IF a listed terminal is stopped by a usage limit, THEN `moai lane-screen --scan` shall print the reset time of that limit in UTC next to its state.
- REQ-4 (REQ-1): IF the Orca program is absent or `orca terminal list` fails, THEN `moai lane-screen --scan` shall exit with a non-zero code that says the list could not be read.
- REQ-5: The command `moai lane-screen --scan` shall write no file and send no input to any terminal.

## Acceptance

- AC-1 (REQ-1, REQ-2): 가짜 `orca` 가 터미널 셋(작업 중·확인 창·읽기 실패)을 주면 표가 셋을 모두 보이고 읽기 실패는 `unreachable` 이며 나머지는 정상이에요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan`
- AC-2 (REQ-3): 한도 화면의 터미널 줄에 해제 시각이 UTC 로 있어요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan_reset`
- AC-3 (REQ-4, REQ-5): 목록을 읽지 못하면 종료 코드가 0 이 아니고, 도구는 보내지도 쓰지도 않아요(가짜 `orca` 의 보내기 기록 파일이 비어 있어요).
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan_fails`

## Scope

- write(이 조각): `.moai/specs/SPEC-20261119-lane-screen-scan/**`
- write(구현 조각 C, 쓰기 범위를 새로 받아요, 분류기 병합 뒤): `src/ops/lane_screen.rs`(`--scan` 가지), `tests/lane_screen_cli.rs`, `tests/fixtures/lane_screens/**`
- read-only: `src/lane_screen.rs`(분류기)

## Plan

1. (이 조각) SPEC.
2. 구현 조각 C: 가짜 `orca` 로 `lane_screen_scan*` 시험을 먼저 써서 실패를 보고 `--scan` 을 만든 뒤 이웃 시험과 변이 시험(상한 12).

## Risks

- 터미널이 많으면 읽는 횟수가 많아요: 터미널마다 한 번이고 그 수만큼 `orca` 를 불러요; 한 번에 하나씩 순서대로 읽어 부하를 늘리지 않아요.
- 목록에 레인이 아닌 터미널(셸)이 섞이면 `unknown` 이나 `idle` 이 아닌 `unknown` 으로 나와요: 화면 표지가 없는 셸은 추측하지 않아요(바탕 SPEC REQ-7).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 터미널을 순서대로 하나씩 읽음 | 부하와 실패 경계가 단순해요 | 병렬 읽기 |
| D-2 | 읽기 실패는 그 줄만 `unreachable` | 한 줄의 실패가 목록을 막지 않아요 | 전체 실패 |

## Result

이 조각은 SPEC 뿐이에요: 코드·시험을 만들지 않았어요.

- AC-1: 아직 안 했어요(구현 조각 C).
- AC-2: 아직 안 했어요(구현 조각 C).
- AC-3: 아직 안 했어요(구현 조각 C).

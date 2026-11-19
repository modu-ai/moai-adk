---
id: SPEC-20261119-lane-screen-scan
title: "A scan lists every open terminal's state and the UTC reset time of each lane stopped by a usage limit"
status: draft
revision: 2
created: 2026-11-19
parent: SPEC-20261119-lane-screen-guard
---

# A scan lists every open terminal's state and the UTC reset time of each lane stopped by a usage limit

## Intent

카드 C-120 의 셋째 조각이에요. 2026-11-19 에 레인 여덟 개가 사용 한도로 멈췄는데 사용자가 말해서야 알았고, 화면에 적힌 해제 시각을 한국 시각으로 읽었지만 실제로는 UTC+8 이라 해제 요청이 한 시간 일찍 나갔어요. 이 SPEC 은 `moai lane-screen --scan` 을 만들어요: 열려 있는 터미널을 `orca terminal list --json` 으로 모아 터미널마다 바탕 SPEC(`SPEC-20261119-lane-screen-guard`)의 분류기를 돌려, 상태 표와 한도로 멈춘 레인의 해제 시각(UTC)을 한 번에 보여 줘요.

개정 1 은 SPEC 만 있었고(조각 A 의 분류기가 병합되기 전), 개정 2 는 분류기(조각 A)와 Orca 호출 틀(조각 B 의 `Orca` 트레이트와 `ProcessOrca`) 위에서 구현해요.

## Interview

- Q1: 터미널 목록은 어디서 얻나요? → A: `orca terminal list --json`(2026-11-19 읽음: 답은 `result.terminals[]`, 터미널마다 `handle`·`title`·`branch`·`connected` 등, 그리고 `totalCount`·`truncated`). 대상이 많아도 터미널마다 한 번씩 화면을 읽어요.
- Q2: 한 터미널이 읽히지 않으면요? → A: 그 줄만 `unreachable` 로 두고 나머지는 계속해요. 한 줄의 실패가 목록 전체를 막지 않아요.
- Q3: 해제 시각은 어디서 가져오나요? → A: 화면 글자에서만, 바탕 SPEC 의 시간대 규칙 그대로예요(아는 시간대만 UTC 로 바꾸고, 시간대가 없거나 모르는 것은 `unknown` 과 원문). 터미널마다 HOME 을 알 수 없으니 `--home` 은 `--scan` 과 같이 쓰지 못해요.
- Q4: 출력은요? → A: 다른 `moai` 명령처럼 JSON 봉투예요. `result.terminals` 에 줄마다 `handle`·`title`·`branch`·`state`·`reason`·`evidence`·`reset`, 그리고 사람이 읽는 표를 `result.table`(줄의 목록)에 같이 둬요. 화면 전체는 남기지 않고 근거 줄 하나만 둬요.
- Q5: 목록이 Orca 쪽에서 잘렸으면요? → A: Orca 의 `truncated` 가 참이면 `result.truncated` 를 참으로 두고 표 끝에 "n 중 m 개만 보여요" 줄을 적어요. 전부 봤다고 말하지 않아요.
- Q6: 목록에 이상한 핸들이 있으면요? → A: `orca terminal read` 의 옵션으로 읽힐 수 있는 핸들(`-` 로 시작 등)은 `orca` 에 넘기지 않고 그 줄을 `unreachable` 로 둬요(조각 B 의 핸들 규칙과 같아요).

## Non-goals

- 한도로 멈춘 레인을 깨우거나 해제 시각에 맞춰 자동으로 보내기.
- 상태를 파일에 기록하기: 읽기만 하고 아무것도 쓰지 않아요.
- 레인이 아닌 터미널의 구분: 목록에 있는 터미널을 모두 보여요.
- 레인마다 `usage-limits.json` 을 읽기: 터미널에서 레인 HOME 을 알 수 없어요(`--home` 은 한 화면을 읽는 기본 모드에서만 써요).

## Requirements

- REQ-1: When `moai lane-screen --scan` runs, the command shall list every terminal that `orca terminal list` reports together with the state of its rendered screen.
- REQ-2 (REQ-1): IF a terminal cannot be read, THEN `moai lane-screen --scan` shall list that terminal as `unreachable` and continue with the remaining terminals.
- REQ-3 (REQ-1): IF a listed terminal is stopped by a usage limit, THEN `moai lane-screen --scan` shall print the reset time of that limit in UTC next to its state.
- REQ-4 (REQ-1): IF the Orca program is absent or `orca terminal list` fails, THEN `moai lane-screen --scan` shall exit with a non-zero code that says the list could not be read.
- REQ-5: The command `moai lane-screen --scan` shall write no file and send no input to any terminal.
- REQ-6 (REQ-1): IF `orca terminal list` reports that its list is cut short, THEN `moai lane-screen --scan` shall say how many of how many terminals it shows.

## Acceptance

- AC-1 (REQ-1, REQ-2): 가짜 `orca` 가 터미널 셋(작업 중·확인 창·읽기 실패)을 주면 표가 셋을 모두 보이고 읽기 실패는 `unreachable` 이며 나머지는 정상이에요. `-` 로 시작하는 핸들은 `orca terminal read` 를 부르지 않고 `unreachable` 예요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan`
- AC-2 (REQ-3): 한도 화면의 터미널 줄에 해제 시각이 UTC 로 있고, 시간대 없는 해제 시각은 UTC 를 지어내지 않고 `unknown` 이에요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan_reset`
- AC-3 (REQ-4, REQ-5): 목록을 읽지 못하면(`orca` 없음·종료 코드 0 아님·JSON 이 아닌 답) 종료 코드가 0 이 아니고 "list could not be read" 가 있으며, 도구는 보내지도 쓰지도 않아요(가짜 `orca` 의 호출 기록에 `send` 가 없고 실행 폴더에 새 파일이 없어요).
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan_fails`
- AC-4 (REQ-6): 목록이 잘렸다고 Orca 가 말하면 `result.truncated` 가 참이고 표 끝 줄이 몇 중 몇 개인지 말해요.
  verify: `cargo test --locked --offline --test lane_screen_cli lane_screen_scan_truncated`

## Scope

- write: `.moai/specs/SPEC-20261119-lane-screen-scan/**`
- write(구현 조각 C): `src/lane_screen.rs`(`scan`), `src/lane_send.rs`(`Orca::list` 와 `ProcessOrca::list`), `src/ops/lane_screen.rs`(`--scan` 가지), `src/help.rs`(예시), 도움말 골든 넷, `docs/design-review/appendix/operations/lane_screen.md`, `tests/lane_screen_cli.rs`, `tests/fixtures/lane_screens/**`
- read-only: 그 밖의 모든 것

## Plan

1. SPEC 개정 2(이 문서).
2. 가짜 `orca`(목록·읽기·호출 기록)로 `lane_screen_scan*` 시험을 먼저 써서 실패를 보고 `scan` 과 `--scan` 을 만든 뒤 이웃 시험, rustfmt(ops 포함), clippy, 변이 시험(레인 상한 5).

## Risks

- 터미널이 많으면 읽는 횟수가 많아요: 터미널마다 한 번이고 그 수만큼 `orca` 를 불러요; 한 번에 하나씩 순서대로 읽어 부하를 늘리지 않아요. 멈춘 터미널 하나는 호출당 시간 제한(`MOAI_ORCA_TIMEOUT_SECS`, 기본 15초)만큼 걸려요.
- 목록에 레인이 아닌 터미널(셸)이 섞이면 `idle` 이 아닌 `unknown` 으로 나와요: 화면 표지가 없는 셸은 추측하지 않아요(바탕 SPEC REQ-7).
- 터미널 제목·브랜치는 남이 정한 글이에요: 제어 문자를 지우고 길이를 잘라서 보여요.

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 터미널을 순서대로 하나씩 읽음 | 부하와 실패 경계가 단순해요 | 병렬 읽기 |
| D-2 | 읽기 실패는 그 줄만 `unreachable` | 한 줄의 실패가 목록을 막지 않아요 | 전체 실패 |
| D-3 | `--scan` 은 `--home` 과 같이 쓰지 못함, 해제 시각은 화면 글자에서만 | 터미널에서 레인 HOME 을 알 수 없어요 | 하나의 HOME 을 모든 줄에 적용 |
| D-4 | JSON 봉투 + `result.table` 줄 목록 | 다른 명령이 모두 봉투를 내고, 사람은 표 줄을 그대로 출력해 읽어요 | 사람용 별도 출력 모드 |
| D-5 | 분류는 `lane_screen::scan` 의 닫는 함수(`read`)로, Orca 호출은 `Orca` 트레이트로 | 분류기는 프로세스를 모르고 시험이 가짜 하나로 끝나요 | 분류기가 `orca` 를 직접 부름 |
| D-6 | Orca 의 글자만 온 답(JSON 아님)은 화면으로 믿지 않고 `unreachable` | 조각 B 와 같은 규칙이에요 | 글자를 화면으로 읽음 |

## Result

(구현 뒤에 AC 마다 한 줄로 채워요.)

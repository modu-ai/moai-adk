---
id: SPEC-20261119-lane-send-guard
title: "A lane send reads the target's screen first and sends nothing unless the lane is working or idle"
status: draft
revision: 2
created: 2026-11-19
parent: SPEC-20261119-lane-screen-guard
---

# A lane send reads the target's screen first and sends nothing unless the lane is working or idle

## Intent

카드 C-120 의 둘째 조각이에요. 2026-11-19 에 레인 화면에 `git reset --hard` 의 Yes/No 확인 창이 떠 있는데 지침 글을 Enter 와 함께 보내서 강조된 Yes 가 눌렸어요(C-108 노트 2813). `orca terminal send` 는 대상 화면을 보지 않고 글자를 넣어요(`orca terminal send --help`, 2026-11-19 읽음). 이 SPEC 은 `moai lane-send` 를 만들어요: 보내기 전에 대상 터미널의 **그려진 화면**(rendered screen)을 읽고, 바탕 SPEC(`SPEC-20261119-lane-screen-guard`)의 분류기로 상태를 정해서 `working`·`idle` 일 때만 보내요. 확인 창·선택 창·사용 한도·접근 불가·모름이면 아무것도 보내지 않고 이유를 말해요.

revision 2: 분류기(조각 A, `src/lane_screen.rs`)가 이 가지의 바탕 안에 있어서 바로 구현해요. 조각 A 가 아직 묶음에 병합되지 않았으므로 이 가지는 조각 A 의 head(`f5070a92a8`) 위에 쌓여요: 조각 A 가 병합되면 이 가지의 새 커밋만 남아요.

## Interview

- Q1: 보낼 글의 안전 검사는요? → A: `src/lane_duty.rs` 의 "다른 터미널에 입력으로 들어가는 글" 규칙(제어 문자·줄 구분자·양방향 문자 거절)을 그대로 써요. 같은 판정 함수를 `pub` 으로 열어 쓰고 복사하지 않아요. 화면 상태 검사와 글 검사가 둘 다 통과해야 보내요.
- Q2: 화면을 읽은 뒤 보내기 전에 화면이 바뀌면요? → A: 읽기와 보내기 사이가 비어 있는 점은 남아요(Risks). 보내기 직전에 한 번 더 읽고 둘의 상태가 다르면 보내지 않아요.
- Q3: 거절하면 종료 코드는요? → A: 0 이 아니에요(성공 흉내를 내지 않아요). 오류 글에 상태·이유·읽은 화면 줄 하나가 있어요.
- Q4: 시험은 어떻게 하나요? → A: 가짜 `orca` 프로그램(`MOAI_ORCA_BIN` 으로 바이너리를 바꿔요)이 읽기 요청에는 합성 화면을 주고, 보내기 요청은 받은 인자를 파일에 남겨서 "보내지 않았음"을 파일이 없거나 비었는지로 확인해요.
- Q5: Enter 는 언제 보내나요? → A: `--enter` 를 줄 때만이에요. 기본은 글만 넣어요(사고가 글과 Enter 를 함께 보낸 것이었어서, Enter 는 명시한 호출만 보내요).
- Q6: 글이 `-` 로 시작하면요? → A: 거절해요. 글이 `orca` 의 옵션(`--interrupt` 같은 것)으로 읽힐 수 있어서예요.

## Non-goals

- 확인 창에 대신 답하기: 이 도구는 보내지 않는 쪽만 맡아요.
- 한도가 풀릴 때까지 기다렸다 보내기(재시도, 대기열): 거절하고 해제 시각을 보여 주기까지예요.
- `hc send` 와 `orca orchestration send` 의 동작 변경: 이 도구는 `orca terminal send` 길 앞에만 서요.
- 보낸 글이 제출됐는지 관찰하기(`--wait-submit`): `orca` 가 돌려준 접수 결과를 그대로 보여 주기까지예요.

## Requirements

- REQ-1: When `moai lane-send` is given a terminal and a text, the command shall read that terminal's rendered screen and classify it before it sends anything.
- REQ-2 (REQ-1): IF the screen is classified as anything other than `working` or `idle`, THEN `moai lane-send` shall send nothing and exit with a non-zero code that names the state and the screen line it was read from.
- REQ-3: When the screen is classified as `working` or `idle`, `moai lane-send` shall send the text to that terminal, with Enter only when `--enter` is given, and print the state it sent under.
- REQ-4 (REQ-3): IF the screen read just before the send is classified differently from the first read, THEN `moai lane-send` shall send nothing and exit with a non-zero code.
- REQ-5 (REQ-2): IF the screen is classified as `usage_limit`, THEN `moai lane-send` shall print the reset time of the limit in UTC, or say that it is unknown, with its refusal.
- REQ-6 (REQ-1): IF the text is empty, begins with `-`, is longer than 2000 characters, or carries a control character or a line separator that the input-safety rule of `lane-duty` refuses, THEN `moai lane-send` shall send nothing and exit with a non-zero code.
- REQ-7 (REQ-1): IF the terminal cannot be read or the Orca program is absent, THEN `moai lane-send` shall send nothing and exit with a non-zero code that says the lane was unreachable.
- REQ-8 (REQ-3): IF the Orca program fails to send, THEN `moai lane-send` shall exit with a non-zero code that says the send failed, never that the text was sent.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-4, REQ-5): 확인 창·선택 창·한도·셸·모름 화면에는 아무것도 보내지 않아요(가짜 `orca` 의 보내기 기록 파일이 없고 종료 코드가 0 이 아니에요). 첫 읽기가 대기 중이고 둘째 읽기가 확인 창이면 보내지 않아요(사고 1). 한도 화면 거절에 해제 시각이 UTC 로(모르면 모른다고) 있어요.
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_refuses`
- AC-2 (REQ-3, REQ-8): 작업 중·대기 중 화면에는 글이 한 번 그대로 가요(`--enter` 는 줄 때만 가요). 보내기가 실패하면 보냈다고 하지 않아요.
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_sends`
- AC-3 (REQ-6, REQ-7): 안전하지 않은 글과 읽지 못한 터미널·없는 `orca` 는 보내지 않아요.
  verify: `cargo test --locked --offline --test lane_send_cli lane_send_unsafe`
- AC-4 (REQ-1…REQ-8): 이웃 시험이 통과해요.
  verify: `cargo test --locked --offline --test lane_duty_line` `cargo test --locked --offline --test lane_screen_cli`

## Scope

- write: `.moai/specs/SPEC-20261119-lane-send-guard/**`, `src/lane_send.rs`(판정 로직, `src/lane_status.rs` 에서 `#[path]` 로 등록), `src/ops/lane_send.rs`, `src/help.rs`(등록), `src/lane_screen.rs`(화면 읽기 한 함수를 열어요), `src/lane_duty.rs`(입력 안전 판정을 `pub` 으로), `tests/lane_send_cli.rs`, `tests/fixtures/lane_screens/**`(바탕 SPEC 의 화면 재사용), `tests/fixtures/help/*.txt`(골든), `docs/design-review/appendix/operations/lane_send.md`
- read-only: 그 밖의 파일

## Plan

1. (조각 1) SPEC.
2. 구현 조각 B: 가짜 `orca` 로 `lane_send_*` 시험을 먼저 써서 실패를 보고 `moai lane-send` 를 만든 뒤 이웃 시험과 변이 시험(상한 5).

## Risks

- 읽기와 보내기 사이: 두 번 읽어도 두 읽기 사이에 확인 창이 뜰 수 있어요. 그 틈을 0 으로 만들지 못해요; 틈을 줄이는 것이고, 보내기 쪽의 근본 보호는 `orca terminal send` 자체에 없어요.
- 분류기가 모르는 화면은 `unknown` 이라 거절돼요: 안전한 쪽이지만 정상 레인이 막힐 수 있어요. 막히면 사람이 `orca terminal send` 를 직접 써요.
- 바탕 SPEC 의 화면 녹화(D-5)가 없으면 시험의 화면이 합성이에요.
- `orca terminal send --help` 는 다른 에이전트에게는 `orca orchestration send` 를 쓰라고 해요: 이 도구는 그 길을 대체하지 않고 `terminal send` 길 앞에서 화면만 지켜요(D-1 은 리더가 받아들였어요).

## Decisions

| id | decision | why | alternative |
| --- | --- | --- | --- |
| D-1 | 보내기 직전에 한 번 더 읽고 다르면 거절 | 읽고 보내는 사이의 변화를 줄여요 | 한 번만 읽기 |
| D-2 | 거절은 종료 코드 0 이 아님 | 호출자가 성공으로 읽지 않아요 | 0 + 메시지 |
| D-3 | Enter 는 `--enter` 를 줄 때만 | 사고가 글과 Enter 를 함께 보낸 것이었어요 | 항상 Enter |
| D-4 | 글은 `-` 로 시작하지 않고 2000자까지 | `orca` 옵션으로 읽히는 것을 막아요; 길이는 입력 안전 규칙의 취지예요 | 따옴표 처리 |
| D-5 | `orca` 는 `MOAI_ORCA_BIN`(없으면 PATH 의 `orca`)로 부르고 호출마다 15초 제한(`MOAI_ORCA_TIMEOUT_SECS` 로 1에서 300초) | 시험이 가짜로 바꾸고, 멈춘 `orca` 가 보내기를 붙잡지 않아요 | 제한 없음 |
| D-6 | 글은 `--text` 인자로 받음 | 한 줄 글이고 표준 입력 예시 규칙을 피해요 | 표준 입력 |

## Result

구현 조각 B(`moai lane-send`, 코드 커밋 `ff78e0d60a`, 2026-11-19T12:33:53Z 에 `date -u` 로 읽은 때)예요. 시험은 RED 먼저: 명령 뼈대만 있는 스텁에서 `lane_send_cli` 14개가 모두 실패했고 구현 뒤 14개 모두 통과해요. 실제 `orca` 는 부르지 않았고 가짜 `orca`(셸 스크립트)와 합성 화면만 썼어요.

- AC-1: 충족 — `cargo test --locked --offline --test lane_send_cli lane_send_refuses` 가 통과해요(확인 창 둘·선택 창·한도 둘·셸·좁은 tail 은 `sent.txt` 없음과 종료 코드 1, 첫 읽기 대기·둘째 읽기 확인 창은 거절하고 읽기 2번, 작업 중→대기도 거절, 한도 거절에 `2026-11-20T10:00:00Z`, 시간대 없는 한도는 "unknown"; 그려진 화면이 아닌 읽기와 글자만 온 답은 거절).
- AC-2: 충족 — `... --test lane_send_cli lane_send_sends` 가 통과해요(대기 중·작업 중에는 읽기 2번 뒤 보내기 한 번, `--enter` 는 줄 때만 가고 한글 글도 그대로, 보내기 실패는 "send failed" 이고 "sent" 라 하지 않으며 한 번만 시도).
- AC-3: 충족 — `... --test lane_send_cli lane_send_unsafe` 가 통과해요(빈 글·`-` 시작·줄바꿈·탭·ESC·양방향 문자·줄 구분자·2001자는 `orca` 를 한 번도 부르지 않고 거절하고 글을 되풀이하지 않으며 2000자는 통과, 위험한 터미널 핸들, 없는 `orca`·읽기 실패·종료된 터미널은 unreachable, 멈춘 `orca` 는 2초 제한에 끊겨 5초 안에 거절, `--now` 형식 오류).
- AC-4: 충족 — `--test lane_duty_line` 8 passed, `--test lane_screen_cli` 22 passed, `--lib lane_send` 10 passed(`--lib lane_` 197), `--bin moai` 28 passed(`every_example_parses`·`operations_doc_matches_code` 포함; 처음에는 예시의 따옴표 때문에 `every_example_parses` 가 실패해 예시를 `<one-line-text>` 로 고쳤어요), `--test help_layout` 10(골든 넷에 `lane-send` 한 줄), `hot_file_ops_merge` 3, `spec_cli` 18, `cargo_env_developer_dir` 4, `home_isolation_scan` 18, `git_env_scrub` 3, `doc_references` 7, `ac_command_lint` 5 모두 통과. `cargo fmt --check` 0, `rustfmt --check --edition 2021 src/ops/*.rs` 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 0, `python3 -m unittest tests.repo_scans_test tests.source_manifest_test tests.local_check_test` 104 OK(이 가지로 빌드한 `moai-cli` 로). 변이 다섯(둘째 읽기 비교 끄기, 핸들 `-` 시작 허용, 글 `-` 시작 허용, Enter 뒤집기, unknown 에도 보내기)이 모두 KILLED, 변이 뒤 작업 트리 깨끗. 돌리지 않은 것: 전체 cargo 시험, 파이썬 discover, 실제 `orca` 터미널에 붙인 실행, `git diff --check` 와 `moai-cli trailer-check` 의 READY 전 확인(READY 때 돌려요).
- 구현에서 SPEC 보다 구체화한 것: `MOAI_ORCA_TIMEOUT_SECS`(1에서 300, 기본 15)로 호출 제한을 시험이 줄일 수 있게 했어요(D-5); 읽은 답이 Orca 의 JSON 이 아니면 `unreachable` 이에요; 첫 읽기와 둘째 읽기가 다르면 새 화면이 막는 이유도 같은 오류 글에 적어요.

읽은 것(조각 1 계획 때): `orca terminal read --help`·`orca terminal send --help`(보호 없음, "다른 에이전트에는 `orca orchestration send`")는 위 Risks 에 있어요.

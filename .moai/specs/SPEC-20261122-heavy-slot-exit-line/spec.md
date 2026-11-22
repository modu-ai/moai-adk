---
id: SPEC-20261122-heavy-slot-exit-line
title: "heavy-slot prints its exit line on every end, success too"
status: draft
revision: 1
created: 2026-11-22
---

# heavy-slot prints its exit line on every end, success too

## Intent

사용자가 승인한 재설계(2026-11-22, `~/.moai/reports/moai-adk/agentic-redesign-2026-11-22.md` 3-F, 레인 백그라운드 조사 F1)의 일이에요. `moai heavy-slot-acquire` 는 감싼 명령이 실패할 때만 `heavy-slot: <name> exited <code>; …` 줄을 찍어요(`src/heavy_slot.rs` 의 `if code != 0`). 그래서 이 줄을 기다리는 셸은 명령이 성공하면 영영 끝나지 않아요(조사 때 agent-3 셸 두 개가 8시간 50분, 리더 셸 하나가 1시간 30분). 이 SPEC 은 시작한 명령이 어떻게 끝나든 끝 줄을 한 번 찍게 하고, 그 계약을 시험으로 고정해요.

## Non-goals

- 기다리는 반복문의 끝 시각은 레인 규칙(리더 R3)과 `moai card wait` 작업의 몫이에요. 저장소의 배포 문서·스킬에는 이 줄을 기다리는 반복문이 없어요(`rg -n 'until \[|until grep|while true|sleep [0-9]+; done|tail -f|grep -m1' skills docs/design-review/appendix .claude/rules AGENTS.md CLAUDE.md agents` 출력 없음, 2026-11-22).
- 시작하지 못한 경우의 줄(127 `could not start`, 75 `not started`, 125 `refused`·`left the queue`)은 바꾸지 않아요.
- Python 쪽 heavy-slot 구현은 바꾸지 않아요(따로 없어져요).
- 줄의 꼴(`heavy-slot: <name> exited <code>; load …; disk free …`)과 종료 코드는 바꾸지 않아요.

## Requirements

- REQ-1: When the command a heavy slot started ends, heavy-slot shall print one `heavy-slot: <name> exited <code>; <machine line>` line on stderr whatever the code, 0 included · failure: none — the command's own exit code is still returned unchanged.
- REQ-2: The heavy-slot contract line in `docs/rust-product-path.md` shall say the exit line is printed on every end of a started command.

## Acceptance

- AC-1 (REQ-1): 성공하는 명령(`true`)을 감싸면 stderr 에 `heavy-slot: <name> exited 0;` 로 시작하는 줄이 정확히 한 번 있고 종료 코드는 0, stdout 은 명령의 것 그대로예요. 실패하는 명령은 지금처럼 `exited <code>` 를 한 번 찍어요 · verify: `cargo test --locked --offline --test heavy_slot_cli a_started_command_prints_its_exit_line_on_success_and_on_failure`
- AC-2 (REQ-2): 문서의 heavy-slot 줄이 "on failure" 가 아니라 끝날 때마다 찍는다고 말해요 · verify: `rg -n 'exited <code>; load …; disk free …. on every end' docs/rust-product-path.md`

## Scope

- write: `.moai/specs/SPEC-20261122-heavy-slot-exit-line/**`, `docs/rust-product-path.md`(heavy-slot 줄 하나, 리더 말로 허락), `src/main.rs`(`heavy-slot-acquire` 도움말 문장 하나, 리더 결정 D-5)
- carried by `SPEC-20261122-check-affected-one-slot` (D-3): `src/heavy_slot.rs` 의 끝 줄, `tests/heavy_slot_cli.rs` 의 AC-1 시험
- read-only: `src/check_affected.rs`, `src/bundle/verify.rs`, `tools/heavy_check_slot.py`
- forbidden: 실제 HOME, 상태 폴더 이름을 글자로 쓴 시험 기대값, 새 Python 시험

## Plan

1. RED 시험 AC-1 (`tests/heavy_slot_cli.rs`): 임시 자리 폴더에서 성공하는 명령을 감싸 `exited 0` 줄을 찾아요(지금은 없어서 실패). 자리 하나 작업의 가지에 실어요(D-3).
2. `src/heavy_slot.rs`: 끝 줄의 `if code != 0` 을 없애요. 일반 경로와 물려받은 자리 재사용 경로 둘 다예요. 자리 하나 작업의 가지가 실어요(D-3).
3. 이 가지: SPEC 과 `docs/rust-product-path.md` 의 그 줄 하나. 자리 하나 작업이 착지한 뒤 착지하고, AC-1 은 그 착지의 영수증을 Result 에 인용해요.
4. 끝 줄을 읽는 다른 시험(check affected 단계 로그, bundle verify)이 줄 하나가 늘어도 그대로인지 그 가지에서 바꾼 대상을 돌려 확인해요.

## Risks

- 성공 때 줄 하나가 늘어요. stderr 를 글자 그대로 비교하는 시험이나 도구가 있으면 깨져요. 바꾸기 전에 `heavy-slot: ` 을 읽는 시험을 모두 찾아 돌려요.
- `machine_line` 이 성공 때도 부하와 디스크 여유를 읽어요(시스템 호출 몇 번). 명령이 끝난 뒤라 명령 시간에는 들어가지 않아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 끝 줄은 성공 때도 실패 때와 같은 꼴로 찍어요(`exited 0`). | 조사 F1 고칠 방법 2, 재설계 3-F. 기다리는 쪽이 한 줄 꼴만 알면 돼요. | 카드 열림 |
| D-2 | 기다림의 끝 시각은 이 SPEC 이 아니라 레인 규칙과 `moai card wait` 가 다뤄요. | 배포 문서·스킬에 기다리는 반복문이 없어요(위 명령). | 카드 열림 |
| D-3 | 코드 줄과 RED 시험은 `SPEC-20261122-check-affected-one-slot` 의 가지가 실어요. 이 SPEC 은 그 가지가 착지한 뒤 SPEC 과 문서 한 줄로 착지해요. | 리더 결정 2026-11-22(선택지 b): 그 작업이 끝 줄을 찍는 함수를 다시 써서, 따로 고치면 같은 줄에서 충돌해요. | 카드 열림 |
| D-4 | 같은 문서 줄의 기다림 파일 이름을 `waiting/<ns>-<pid>-<max>` 로 바로잡아요. | 자리 하나 작업(착지 6d5410cd2d)이 이름에 `--max` 를 붙였는데 이 줄은 옛 꼴로 남았어요(`src/heavy_slot.rs` 19행 모듈 설명과 달라요). 이 SPEC 에 허락된 그 줄 하나 안의 고침이에요. | 카드 열림 |
| D-5 | `src/main.rs` 의 `heavy-slot-acquire` 도움말 문장 "while waiting and on failure it prints …" 를 "while waiting and when COMMAND ends it prints …" 로 바꿔요. 문서 줄은 C-325 가 그 파일을 예약한 동안에도 이 카드와 함께 착지해요. | 리더 결정 2026-11-22(노트 6923, 6921 의 기본안 둘 다). 자리 하나 작업 감사의 N6 이 이 문장을 짚었어요. 이 문장을 글자 그대로 읽는 시험은 없어요(`rg -ln 'on failure it prints the load' tests src docs skills` 는 `src/main.rs` 만 찾아요). | 카드 열림 |

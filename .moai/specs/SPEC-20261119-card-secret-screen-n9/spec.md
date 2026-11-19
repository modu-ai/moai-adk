---
id: SPEC-20261119-card-secret-screen-n9
title: "The time-bound test of the card text screen runs in seconds and still fails on square time"
status: draft
revision: 1
created: 2026-11-19
parent: SPEC-20261119-card-secret-screen
---

# The time-bound test of the card text screen runs in seconds and still fails on square time

## Intent

C-122 의 재재감사(카드 노트 3396)가 N9 를 남겼고 리더가 다음 조각에서 반드시 하라고 정했어요: B2 수리가 더한 시간 상한 시험(`secret_screen_time_on_one_long_word_has_a_bound`)이 debug 에서 10-32초 걸려서 모든 레인의 lib 실행과 CI 에 그만큼 더해져요. 크기 세 개 곱하기 꼴 넷을 차례로 돌리는 대신, 한도 크기 한 번 곱하기 꼴 넷과 끝 모양 하나를 동시에 돌려 2-4초로 줄이되 제곱 시간이 되돌아오면 같은 시험이 실패해야 해요(변이로 확인). 같은 조각에서 N8 도 고쳐요: 부모 SPEC 의 수리 2 결과 문장이 "결과는 같다"고 했지만 접두 모양의 길이를 다음 구분자 전까지만 세게 되어 짧은 접두 조각이 구분자를 넘어 이어지는 두 꼴이 거절에서 통과로 바뀌었어요(진짜 키는 세 구분자를 품지 않아 놓치는 키 모양은 없어요).

## Non-goals

- 화면 규칙 바꾸기: `secret_words` 의 동작은 그대로예요.
- N3·N5·N7 꼴(공백·따옴표·카멜·reclaim 사유·다른 구분자)은 후속 카드 후보예요.

## Requirements

- REQ-1: The time-bound test of the card text screen shall run a word of the note limit's size in each of four shapes and one shape at the end of a long word at the same time on separate threads.
- REQ-2: The time-bound test of the card text screen shall fail when a screen of such a word takes more than fifteen seconds.
- REQ-3: The time-bound test of the card text screen shall finish within four seconds of wall time in a debug build on an idle machine.

## Acceptance

- AC-1 (REQ-1, REQ-3): 새 시험이 debug 에서 4초 안에 통과해요(측정값은 Result).
  verify: `cargo test --locked --offline --lib secret_screen_time_on_one_long_word_has_a_bound`
- AC-2 (REQ-2): 구분자마다 뒤 전체를 넘기는 옛 꼴로 되돌린 변이가 이 시험을 15초 안에 실패시켜요.
  verify: `cargo test --locked --offline --lib secret_screen`
- AC-3 (REQ-1): 낱말 끝의 자격 증명 모양이 한도 크기 낱말에서도 거절돼요(같은 시험의 끝 모양 줄).
  verify: `cargo test --locked --offline --lib secret_screen_time_on_one_long_word_has_a_bound`

## Scope

- write: `.moai/specs/SPEC-20261119-card-secret-screen-n9/spec.md`, `.moai/specs/SPEC-20261119-card-secret-screen/spec.md`(Result 한 문장), `src/card/input.rs`(시험 하나)
- read-only: `src/redact.rs`, `src/ledger/workflow.rs`
- forbidden: 화면 동작 바꾸기, 네트워크, 실제 HOME

## Plan

1. 시험을 한도 크기 한 번 곱하기 꼴 넷과 끝 모양 하나를 동시에 돌리는 꼴로 바꿔요.
2. debug 시간을 재고, 옛 꼴로 되돌린 변이가 15초 안에 실패하는지 봐요.
3. 부모 SPEC 의 수리 2 결과 문장을 N8 대로 고쳐요.

## Risks

- 동시에 돌린 스레드가 부하 높은 기계에서 느려질 수 있어요. 상한은 호출 하나가 아니라 전체 15초이고 선형 시간은 한 호출에 1초 안팎이라 여유가 커요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 크기는 한도 하나만 쓰고 200 KB 와 한도−1 은 뺀다 | 리더 결정(노트 3396 N9): 한도 크기 한 번 곱하기 꼴 넷 + 끝 모양 하나; 제곱 시간은 한도 크기에서 가장 크게 드러나요 | 크기 다시 늘리기 |
| D-2 | 다섯 호출을 스레드로 동시에 돌리고 전체 상한 15초를 둔다 | 벽시계 시간을 가장 느린 호출 하나 정도로 줄이고 제곱 시간이 되돌아오면 15초에 실패시키려고요 | 차례로 돌리기 |

## Result

- AC-1: met — `cargo test --locked --offline --lib secret_screen_time_on_one_long_word_has_a_bound` 를 세 번 연속 돌린 `finished in` 이 2.73초, 1.50초, 1.42초(debug, 이 기계, 병렬 부하 있음). 수리 2 때의 시험은 10.75초(재감사 측정은 10-32초)였어요. 시험은 한도 크기(`MAX_NOTE_BYTES` = 1048576) 낱말 꼴 넷과 끝 모양 하나를 스레드 다섯 개로 동시에 돌려요.
- AC-2: met — 구분자마다 뒤 전체를 넘기던 옛 꼴로 되돌린 변이를 적용하니 같은 시험이 15.01초에 실패했어요(`finished in 15.01s`). 복원 뒤 `git status` 깨끗.
- AC-3: met — 끝 모양 줄의 기대값을 0 으로 바꾼 변이(끝 모양을 못 잡는 경우와 같은 결과)를 적용하니 1.07초에 실패했어요. 복원 뒤 깨끗.
- N8: 부모 SPEC `SPEC-20261119-card-secret-screen` 의 수리 2 결과 문장을 재감사 말대로 고쳤어요(접두 모양의 길이를 다음 구분자 전까지만 세게 되어 짧은 접두 조각이 구분자를 넘어 이어지는 두 꼴이 거절에서 통과로 바뀌었고, 진짜 키는 세 구분자를 품지 않아 놓치는 키 모양은 없다).
- 화면 동작은 바꾸지 않았어요(`src/card/input.rs` 의 `#[cfg(test)]` 안 시험 하나만 바뀜).
- 돌리지 않은 것: CI, 실제 HOME, 부하가 매우 높은 기계(재감사는 부하 66-134 에서 이전 시험이 15회 통과했다고 적었어요)에서의 새 시험.

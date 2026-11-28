---
id: SPEC-20261120-load-tolerant-time-bounds
title: "Time-bounded tests survive a loaded machine"
status: draft
revision: 1
created: 2026-11-20
parent: null
---

# Time-bounded tests survive a loaded machine

## Intent

2026-11-20 에 기계 부하(load average 50-65)에서 시간 제한 테스트가 한 번 흔들렸어요. `secret_suspects` 이차 시간 방지 테스트의 절대 3초 천장이 3.24초로 넘어갔어요(`src/agents_local_merge.rs` 의 `one_huge_line_is_scanned_in_bounded_time`). 절대 시간은 기계가 바쁘면 무너지지만 연산 횟수는 부하와 상관없어요. 이 조각은 "한 줄에 같은 접두사가 반복돼도 검사 비용이 줄 길이에 선형"이라는 보증을 시계 대신 **검사가 소비한 문자 수**로 판정해요. 시계 천장은 30초로 넉넉하게 두어 뒷받침만 하고, 회귀를 잡는 일은 문자 수 테스트가 해요.

이 조각의 첫 번째 시도는 팀의 열차(M-Flow)에서 착지하지 못했어요. 그 안에는 heavy-slot 기본 동시성을 5로 올리는 요구가 함께 있었어요. 슬롯 상한은 그 뒤 사용자 결정과 리더의 최신 `--max` 줄을 따르는 운영 규칙으로 바뀌었고, 이 조각에서는 슬롯 기본값에 손대지 않아요(아래 Non-goals). 시간 한도를 느슨하게 만드는 것이 목적이 아니에요. 시간에 기대지 않는 방식으로 같은 보증을 더 단단하게 만드는 것이 목적이에요.

`lane_send` 에는 코드에 2초 상한이 없어요(`DEFAULT_TIMEOUT_SECS` 15초, `MOAI_ORCA_TIMEOUT_SECS` 로 1-300 조정). 그날의 "lane_send 2 s limit" 은 실행 환경이 건 제한으로 기록하고 코드를 바꾸지 않아요. `home_retire_cli` 에는 시간 단언이 없어요(`sleep 60` 은 사용 중인 척하는 흉내예요). 원인이 확정되지 않았고 이 조각에서 바꾸지 않아요.

## Non-goals

- `lane_send`·`home_retire` 의 코드 변경: 위 Intent 에 적은 기록 사항이에요.
- heavy-slot 기본 동시성 숫자(`src/heavy_slot.rs` 의 `DEFAULT_MAX`)와 그 문서·스크립트 줄: 상한은 리더의 최신 `--max` 줄이 정하는 운영 규칙이라 이 조각 밖이에요.
- 로컬 검사의 Python discover 선택(affected selection): 별도 조각이에요.

## Requirements

- REQ-1: The `secret_suspects` timing test shall judge the repeated-prefix scan by the number of characters the run scan consumes instead of a clock, keeping the count within eight times the line's length, and the clock ceiling of thirty seconds on the pre-existing huge-line test stays as the backstop only. · failure: none — a count does not move with machine load, and the clock ceiling absorbs the rest.
- REQ-2: The run-scan counter shall be compiled in test builds only, so that the product build carries none of it · failure: none — the product code path is unchanged.

## Acceptance

- AC-1 (REQ-1): the count-based test passes in the green run, and dropping the `take(min)` clamp of `has_token` makes it fail, with the count orders of magnitude past the bound.
  verify: `cargo nextest run --locked --offline --lib -E 'test(agents_local_merge)'`
  verify: the clamp-removal mutation run recorded in the READY note
- AC-2 (REQ-2): `cargo build --locked --offline --bin moai-cli` compiles without the counter, and the counter's declaration and its one increment sit behind `#[cfg(test)]`.
  verify: `git grep -n RUN_SCAN_CHARS -- src` shows only lines preceded by `#[cfg(test)]` or inside the test module

## Scope

- write: `.moai/specs/SPEC-20261120-load-tolerant-time-bounds/spec.md`, `src/agents_local_merge.rs`
- read-only: `src/lane_send.rs`, `tests/home_retire_cli.rs`, `src/heavy_slot.rs`
- forbidden: `lane_send`·`home_retire` changes, heavy-slot default changes, card numbers in output

## Plan

1. RED: write the count-based test first; with the counter wired in and the `take(min)` clamp removed the test fails by orders of magnitude, which proves it catches the regression it is for.
2. GREEN: put the clamp back; the count stays within eight times the line length; raise the clock ceiling to thirty seconds.
3. Run the module's tests and the card receipt; run the clamp-removal mutation once.

## Risks

- A clock-ratio test was tried first in the earlier attempt and failed five of thirteen runs under load, because the measurements were only 45-96 ms. That is why the count replaces it.
- The counter exists only in test builds. The bound (eight times the line length) leaves room over the clamped linear scan; removing the clamp is quadratic and blows past it by orders of magnitude.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | `lane_send` and `home_retire` are not touched: the code has no absolute limit there and the cause is not settled. Fix them when it reproduces. | file:line re-read | reopen the card |
| D-2 | The heavy-slot default is not changed by this slice: the cap follows the leader's latest `--max` line. | operating rule, not a code default | none |

## Result

조각이 한 번 구현됐어요(2026-11-28, 리더 지시로 착지하지 못한 첫 시도를 현재 끝 위에서 다시 씀). 첫 시도의 heavy-slot 기본값 요구는 Non-goals 대로 싣지 않았어요.

- AC-1: met — `cargo nextest run --locked --offline --no-fail-fast --lib -E 'test(agents_local_merge)'` 가 35개 모두 통과했고(슬롯 `--max 2`, 0.8초), 그 안에 `the_repeated_prefix_scan_consumes_characters_linearly`(0.078초)와 시계 천장 30초의 `one_huge_line_is_scanned_in_bounded_time` 이 있어요. `has_token` 의 `.take(min)` 클램프를 지운 뮤턴트에서는 `the_repeated_prefix_scan_consumes_characters_linearly` 만 돌려 실패했어요: `the run scan consumed 2147680256 characters for a 131080-character line: quadratic`(한도 1,048,640, 종료 코드 100, 51초). 클램프를 되돌린 뒤 `git diff` 에 `take(min)` 줄이 없어요.
- AC-2: met — `git grep -n -B1 RUN_SCAN_CHARS -- src` 가 선언(`#[cfg(test)] thread_local!`)과 증가 한 줄(`#[cfg(test)]` 바로 아래)과 `#[cfg(test)] mod tests` 안의 테스트 사용만 보여요. 제품 빌드 확인은 이 변경의 `moai check affected --card` 영수증(clippy 단계)을 READY 노트가 인용해요.

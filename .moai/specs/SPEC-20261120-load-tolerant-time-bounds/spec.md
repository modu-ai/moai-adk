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

오늘 기계 부하(load average 50-65)에서 시간 제한 시험이 한 번 흔들렸어요 — secret_suspects 이차 시간 방지 시험의 절대 3초 천장이 3.24s 로 넘어갔어요(src/agents_local_merge.rs:1045-1050). 절대 시간은 기계가 바쁘면 무너지지만 **같은 부하의 두 측정의 비율**은 무너지지 않아요. 이 조각은 그 시험에 상대 측정(2배 길이 입력의 비용이 1배 길이 입력의 배수 이내 — 선형이면 ~2배, 이차 회귀면 4배 이상)을 더해 부하에서 안정적인 회귀 감지로 만들고, 절대 천장은 넉넉하게(30초) 유지해요. heavy-slot 의 기계 한도 기본값도 공지된 상한과 맞춰요 — limits.json 이 없을 때의 동시성 기본이 3 이었는데(src/heavy_slot.rs:40 주석), 운영은 --max 4 로 달려요 — 코드 기본을 4 로 맞춰요(카드 "heavy builds capped at 4"). lane_send 는 코드에 2초 상한이 없어요(DEFAULT_TIMEOUT_SECS 15 초, MOAI_ORCA_TIMEOUT_SECS 로 1-300 조정 — PLAN 노트 4092) — 오늘의 "lane_send 2 s limit" 은 실행 환경이 건 제한으로 기록하고 코드 변경 없어요. home_retire_cli 에도 시간 단언이 없어요(sleep 60 은 in-use 흉내) — 원인 미확정, 이 조각에서 변경 없어요.

## Non-goals

- lane_send·home_retire 의 코드 변경: 위 Intent 의 기록 사항.
- local_check 의 python discover affected selection(카드 본문 마지막 문장): 별도 조각.
- heavy-slot 의 한도 동작 변경: 기본값 숫자만 — limits.json 이 이기는 경로는 그대로.

## Requirements

- REQ-1: The secret_suspects timing test shall judge the repeated-prefix cost relatively — the doubling of the line's length multiplying the measured cost by less than 3.2, on top of an absolute ceiling of thirty seconds per measurement. · failure: none — a loaded machine slows both measurements together, so the multiple holds while the ceiling absorbs the rest.
- REQ-2: The heavy-slot built-in concurrency limit shall default to four, so the announced cap holds without a limits file, with a limits.json value still winning over the built-in default. · failure: none — the file path is unchanged; only the built-in default number moves.

## Acceptance

- AC-1 (REQ-1): the relative measurement holds in the green run, and a mutation that restores the quadratic cost in the repeated-prefix scan (dropping the `take(min)` clamp) pushes the multiple past 3.2 and fails the test.
  verify: `cargo test -p moai-cli --lib agents_local_merge` — the timing test name in the output
  verify: the mutation run recorded on the card's READY note
- AC-2 (REQ-2): the heavy-slot unit tests pin the built-in default of four and pin a limits.json value winning over it.
  verify: `cargo test heavy_slot`
- AC-3: the wider READY set around the change passes.
  verify: `cargo test --locked --offline --bins`
  verify: `cargo test --locked --offline --lib`

## Scope

- write: `.moai/specs/SPEC-20261120-load-tolerant-time-bounds/spec.md`, `src/agents_local_merge.rs`, `src/heavy_slot.rs`
- read-only: `src/lane_send.rs`, `tests/home_retire_cli.rs`
- forbidden: lane_send·home_retire 변경, heavy-slot 동작(기본값 숫자 외) 변경, 카드 번호의 출력 사용

## Plan

1. RED: heavy-slot 의 기본값 시험이 현재 3 을 단언하면 새 기본 4 로 실패해요; 상대 측정 시험은 먼저 쓰고 뮤테이션(`take(min)` 제거 → 이차 복원)으로 실패하는 것으로 회귀 감지를 증명해요.
2. GREEN: 상대 측정 시험(2배 길이 배수 < 3.2, 각 측정의 최소값으로 잡음 완화, 절대 천장 30초) + 기존 절천장 넉넉화 + heavy_slot 기본 4.
3. 비축소 READY 검사 → 감사는 리더에게 → READY. 뮤테이션 1-2 개로 상한 준수.

## Risks

- 상대 배수 3.2 는 이론 경계(이차=4)와 측정 잡음 사이 값이에요: 각 측정의 최소값을 쓰고, 배수가 3.2 를 넘으면 진짜 이차라는 판단이 서게 충분히 여유를 둬요(4 는 넉넉히 넘는 값).
- 측정 입력이 작으면 잡음이 커져요: 길이를 기계에서 안정적으로 측정 가능한 크기(십만 자대)로 잡고 반복 측정의 최소값을 써요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | lane_send 와 home_retire 는 이 조각에서 만지지 않아요 — 코드에 절대 상한이 없고(PLAN 노트 4092) 원인이 미확정이라 재현 시 그 때 수리해요. | file:line 재읽기 | 카드 열림 |

## Result

(status draft — written when the slice lands)

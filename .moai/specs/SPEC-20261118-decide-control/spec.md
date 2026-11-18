---
id: SPEC-20261118-decide-control
title: "The decision library is off by default and stops at a daily call limit"
status: draft
revision: 1
created: 2026-11-18
---

# The decision library is off by default and stops at a daily call limit

## Intent

`moai decide`의 라이브러리(`src/decide/`)에는 요청을 검사하는 `prepare()`, 보내는 `transport`, 답을 기록으로 바꾸는 `conclude()`가 있지만 "켜져 있는가"와 "오늘 몇 번 불렀는가"를 따지는 곳이 없어요(`src/decide/mod.rs`의 `BODY_MAX` 주석이 "공급자 한도는 설정(S3)과 함께 온다"고 적어 둔 그 자리). 이 SPEC은 그 한 곳, 순수한 관문 `src/decide/control.rs`만 더해요: 설정이 켜짐이고 오늘 한도가 남았을 때만 보내는 쪽(sender)을 한 번 부르고, 그 밖에는 모두 `fallback: leader` 거절이에요(AC-4: 기본 꺼짐, 꺼짐이면 네트워크 호출 0, 한도).
이 관문은 순수해요: 설정 글·오늘 날짜·사용량·보내는 쪽을 인자로 받고 파일·시계·소켓·키를 만지지 않아요. 설정 파일 읽고 쓰기, 사용량 저장, `moai decide` 명령은 `src/main.rs`가 필요한 다음 조각의 일이에요.

## Non-goals

- `moai decide`·`decide-config` 명령, 설정과 사용량 파일 읽고 쓰기, 날짜 읽기: CLI 조각의 몫이고 `src/main.rs`는 이 조각 밖이에요.
- 실제 endpoint 호출과 TLS, 키 꺼내기: 이 관문은 키를 모르고 주소를 몰라요(보내는 쪽이 따로 있어요).
- 토큰 수 한도와 비용 계산: 호출 횟수만 세요(D-2).

## Requirements

- REQ-1: The control gate shall keep the decision model off until the settings text names `"enabled": true`.
- REQ-2 (REQ-1): IF the model is off, THEN the control gate shall refuse with `decide_off` before it reads the request and without calling the sender.
- REQ-3: The control gate shall count a call against the day's limit before it calls the sender.
- REQ-4 (REQ-3): IF the day's calls have reached the limit, THEN the control gate shall refuse with `decide_limit` without calling the sender.
- REQ-5: The control gate shall start a new count only when the day is later than the counted day.
- REQ-6 (REQ-3): IF the gate or the secret screen refuses the request, THEN the control gate shall leave the count unchanged and not call the sender.
- REQ-7: The control gate shall read the settings text strictly: known fields only, a boolean `enabled`, a provider name from the compiled list and `calls_per_day` from 1 to 1000.
- REQ-8 (REQ-7): IF the settings text is not valid JSON or breaks one of those rules, THEN the control gate shall refuse with `settings_invalid` and leave the model off.
- REQ-9 (REQ-5): IF the day text is not a date of the form `YYYY-MM-DD`, THEN the control gate shall refuse with `clock_invalid` without calling the sender.
- REQ-10: The control gate shall hand the sender nothing but the prepared request of the provider the settings name.

## Acceptance

- AC-1 (REQ-1, REQ-2): 설정이 기본값(`Settings::default()`, 글 `{}`)이면 꺼짐이고, 꺼짐이면 요청이 올바르든 아니든 `decide_off`이고 보내는 쪽 호출이 0이고 사용량이 그대로이고 `to_json()`이 `fallback: leader`를 담아요.
  verify: `cargo test --locked --offline decide::control::tests::the_default_is_off_and_off_calls_nothing`
- AC-2 (REQ-3, REQ-4, REQ-5, REQ-10): 한도 3이면 호출 1·2·3은 보내는 쪽에 가고 4번째는 `decide_limit`이고 호출 수가 그대로예요(N/N+1 경계); 보내는 쪽이 거절해도 그 호출은 센 채로 남아요; 날짜가 늦어지면 새로 세고 같거나 이른 날짜면 새로 세지 않아요; 보내는 쪽이 받는 `Prepared`는 설정의 공급자의 엔드포인트를 가져요.
  verify: `cargo test --locked --offline decide::control::tests::the_limit_stops_the_call_after_n_and_a_later_day_starts_again`; `cargo test --locked --offline decide::control::tests::a_failed_call_still_counts_and_the_provider_comes_from_the_settings`
- AC-3 (REQ-6): 관문이나 비밀 검사가 거절한 요청(되돌릴 수 없음, 비밀 모양)은 보내는 쪽 호출 0이고 호출 수가 그대로예요.
  verify: `cargo test --locked --offline decide::control::tests::a_request_the_gate_refuses_is_neither_sent_nor_counted`
- AC-4 (REQ-7, REQ-8): 잘못된 설정 글(JSON 아님, 모르는 필드, `"enabled": "true"`, 모르는 공급자, 0·1001·소수·글자인 한도)은 `settings_invalid`이고 켜지지 않아요; 올바른 글(한도 1과 1000, 두 공급자)은 읽혀요.
  verify: `cargo test --locked --offline decide::control::tests::the_settings_text_is_read_strictly`
- AC-5 (REQ-9): 날짜 글이 `YYYY-MM-DD`가 아니면 `clock_invalid`이고 보내는 쪽 호출 0이에요.
  verify: `cargo test --locked --offline decide::control::tests::a_day_that_is_not_a_date_is_refused`
- AC-6 (REQ-1…REQ-10): 이웃 검사.
  verify: `cargo test --locked --offline --lib decide::`; `cargo test --locked --offline --test home_isolation_scan`; `cargo clippy --offline --locked --all-targets -- -D warnings`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-decide-control/spec.md`, `src/decide/control.rs`, `src/decide/mod.rs`(모듈 한 줄과 머리 주석)
- read-only: `src/decide/provider.rs`, `src/decide/transport.rs`, `tests/home_isolation_scan.rs`
- forbidden: 실제 endpoint·실제 키·실제 HOME의 보관함, 네트워크, 새 의존성, `src/main.rs`, 스캔 B의 예외 넓히기

## Plan

1. 시험 먼저: `src/decide/control.rs`에 시험을 쓰고 본문은 껍데기(꺼짐도 아닌 `decide_off`만 돌려주기)로 두어 시험이 실패하는 것(RED)을 읽어요.
2. 관문을 구현해 GREEN으로 만들어요.
3. 이웃 검사(`decide::`, 스캔 B, clippy, fmt)와 변이 시험을 돌려요.

## Risks

- 호출 수를 보내기 전에 세어서 느린·실패한 호출도 한도를 써요 → 비용을 아끼는 쪽이 안전하고, 실패가 한도를 풀어 주는 길(무한 재시도)을 막아요(D-2).
- 시계가 뒤로 가면 새로 세는 틈이 생겨요 → 날짜가 더 늦을 때만 새로 세요(REQ-5).
- 이 관문이 거절을 늘려 레인이 리더를 더 기다려요 → 거절은 모두 `fallback: leader`라서 지금 규칙과 같아요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 관문은 순수하게 두고 설정·날짜·사용량·보내는 쪽을 인자로 받아요 | 파일·시계·소켓을 만지지 않으면 `home_isolation_scan` 스캔 B를 그대로 지나고 시험이 가짜 보내는 쪽 하나로 끝나요; 실제 읽고 쓰기는 `src/main.rs`가 필요한 다음 조각이에요 | 다음 조각이 읽는 쪽을 더할 때 이 함수는 그대로 둬요 |
| D-2 | 한도는 하루 호출 횟수(기본 20, 1~1000)이고 호출은 보내기 직전에 세요 | 카드 AC-4가 "한도"를 요구해요; 횟수는 토큰 수보다 단순하고, 보내기 전에 세면 느리거나 실패한 호출도 비용으로 봐요 | 기본값만 상수라서 다음 SPEC이 바꿔요 |
| D-3 | 설정 글에 `enabled`가 없으면 꺼짐이고, 읽을 수 없는 글도 꺼짐(`settings_invalid`)이에요 | 카드가 기본 꺼짐과 "꺼짐이면 네트워크 호출 0"을 요구해요; 헷갈리면 켜지 않는 쪽이 안전해요 | 없음(기본값 변경은 새 SPEC) |

## Result

(구현 뒤에 AC마다 한 줄씩 적어요.)

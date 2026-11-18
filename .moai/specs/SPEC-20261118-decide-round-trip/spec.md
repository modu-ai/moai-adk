---
id: SPEC-20261118-decide-round-trip
title: "The decision library sends a prepared request to a loopback server and reads the answer"
status: draft
revision: 1
created: 2026-11-18
---

# The decision library sends a prepared request to a loopback server and reads the answer

## Intent

`moai decide`의 라이브러리(`src/decide/`)는 지금 순수한 반쪽이에요: `prepare()`가 요청을 검사해 바이트로 만들고 `conclude()`가 응답 바이트를 기록으로 바꾸지만, 그 사이에서 바이트를 실제로 보내고 받는 곳(S4)이 없어요(`src/decide/mod.rs`의 머리 주석). 이 SPEC은 그 한 곳만 더해요: 준비된 요청을 `127.0.0.1` 같은 루프백 주소의 가짜 HTTP 서버에 보내고 응답 본문을 받아 `conclude()`에 넘기는 왕복(이 SPEC의 대상인 AC-2)이에요.
이 조각에서 키와 요청은 이 기계 밖으로 나갈 수 없어요: 루프백이 아닌 주소는 소켓을 열기 전에 거절해요. 실제 Jev·OpenAI 호출은 이 조각에 없고, https라서 TLS 클라이언트가 필요한 다음 조각(V5 감사 뒤)의 일이에요.

## Non-goals

- 실제 endpoint 호출(Jev·OpenAI 모두)과 TLS: 표준 라이브러리는 TLS를 못 해요. 다음 조각의 Decisions에서 TLS 클라이언트 크레이트를 견줘요(D-2).
- `moai decide` 명령, 설정 읽기, 키 보관함에서 키 꺼내기: CLI 조각의 몫이고 `src/main.rs`는 이 조각 밖이에요.
- 재시도, 리다이렉트 따라가기, 청크 전송 해석, 연결 재사용: 하지 않아요(D-3).

## Requirements

- REQ-1: The transport shall send the body of a prepared request in an HTTP/1.1 POST to the path of its provider endpoint on a loopback address and return the response body for `conclude`.
- REQ-2: The transport shall carry the provider key only in the `Authorization: Bearer` header of that request.
- REQ-3: The transport shall refuse an address that is not a loopback address before it opens a socket.
- REQ-4: The transport shall keep the key out of its `Debug` output and out of every refusal text.
- REQ-5: The transport shall treat a redirect as a refusal and never follow it.
- REQ-6: The transport shall bound the connect time, the total time and the response size.
- REQ-7 (REQ-1): IF the response status is not 200, THEN the transport shall refuse with a fixed code and quote nothing of the response.
- REQ-8 (REQ-6): IF the server closes early, sends fewer bytes than its `Content-Length` says, sends more than the size bound or answers too slowly, THEN the transport shall refuse with a fixed code.
- REQ-9 (REQ-2): IF the key is empty or holds a control character, THEN the transport shall refuse before it opens a socket.

## Acceptance

- AC-1 (REQ-1, REQ-2): 가짜 서버가 받은 요청은 `POST <endpoint의 경로> HTTP/1.1`, `Authorization: Bearer <가짜 키>`, `Content-Type: application/json`, `Content-Length`와 `prepared.body()`와 같은 본문을 가져요; 두 공급자(typesafe, openai) 모두 왕복 뒤 `conclude()`의 `Record`가 응답의 선택을 담아요.
  verify: `cargo test --locked --offline decide::transport::tests::a_prepared_request_makes_a_round_trip_for_each_provider`
- AC-2 (REQ-3, REQ-4, REQ-9): 루프백이 아닌 주소는 연결 시도 없이 거절되고, 빈 키·제어 문자가 든 키는 소켓을 열기 전에 거절되고, 키의 글자가 `Debug`와 어느 거절 문구에도 나오지 않아요.
  verify: `cargo test --locked --offline decide::transport::tests::a_non_loopback_address_or_a_bad_key_is_refused_before_any_socket`; `cargo test --locked --offline decide::transport::tests::the_key_is_in_no_debug_output_and_no_refusal`
- AC-3 (REQ-5, REQ-7): 401·500·302 응답은 각각 고정된 코드의 거절이고 응답 글을 인용하지 않고 리다이렉트를 따라가지 않아요.
  verify: `cargo test --locked --offline decide::transport::tests::a_status_other_than_200_is_a_fixed_refusal_and_a_redirect_is_not_followed`
- AC-4 (REQ-6, REQ-8): 크기 상한 초과, 본문이 짧음, 일찍 닫음, 느린 서버가 각각 고정된 코드의 거절이에요.
  verify: `cargo test --locked --offline decide::transport::tests::a_response_that_is_too_big_short_early_closed_or_slow_is_refused`
- AC-5 (REQ-1…REQ-9): 이웃 검사.
  verify: `cargo test --locked --offline decide::`; `cargo test --locked --offline --test home_isolation_scan`; `cargo test --locked --offline --test home_isolation_scan the_decide_scan_allows_the_four_socket_names_in_the_sending_file_only`; `cargo test --locked --offline --test home_isolation_scan the_sending_file_checks_the_address_before_it_connects`; `cargo clippy --offline --locked --all-targets -- -D warnings`; `cargo fmt --check`

## Scope

- write: `.moai/specs/SPEC-20261118-decide-round-trip/spec.md`, `src/decide/transport.rs`, `src/decide/mod.rs`(모듈 한 줄과 머리 주석), `tests/home_isolation_scan.rs`(스캔 B에 `src/decide/transport.rs` 한 파일·네 이름의 예외와 그 시험; 리더가 범위에 넣었어요)
- read-only: `src/decide/provider.rs`, `src/decide/outbound.rs`, `src/key_vault.rs`
- forbidden: 실제 endpoint·실제 키·실제 HOME의 보관함, 루프백 밖 네트워크, 새 의존성

## Plan

1. 시험 먼저: `src/decide/transport.rs`에 가짜 서버와 시험을 쓰고 본문은 거절만 하는 껍데기로 두어 시험이 실패하는 것(RED)을 읽어요.
2. 전송을 구현해 GREEN으로 만들어요.
3. 이웃 검사를 돌려요.

## Risks

- 이 조각의 소켓 코드가 키를 담은 요청을 다룬다 → 루프백 말고는 소켓을 열지 않고(AC-2) 키를 어떤 문구에도 넣지 않아요(AC-2); audit-tier 2라서 리더가 감사를 맡아요.
- 스캔 B(`tests/home_isolation_scan.rs`)가 `src/decide/` 아래 소켓을 금지해요 → 그 시험 파일의 주석이 S4는 자기 예외를 갖는다고 적어 뒀어요; 예외는 `src/decide/transport.rs` 한 파일의 `std::net` 계열 이름만이고 URL·프로세스 실행·다른 파일은 그대로 스캔돼요(D-4).
- 가짜 서버의 응답 모양이 실제 서버와 다를 수 있다 → 이 조각은 전송의 경계(상태·길이·시간)만 재요. 공급자 응답 모양은 `provider.rs`가 이미 시험해요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 루프백이 아닌 주소는 소켓을 열기 전에 거절해요 | 키가 든 요청이 이 조각에서 기계 밖으로 나갈 길이 없게 하려는 거예요; `provider.rs`는 엔드포인트를 컴파일 시간 상수로 두어 요청을 다른 호스트로 돌릴 수 없게 했고, 같은 원칙이에요 | 다음 조각이 TLS 전송을 더할 때 함수를 따로 만들고 이 함수는 그대로 둬요 |
| D-2 | 새 의존성을 넣지 않고 표준 라이브러리 `TcpStream`만 써요 | 가짜 `127.0.0.1` 서버와의 왕복에는 충분해요. 실제 endpoint는 https라서 TLS가 필요한데 표준 라이브러리에 없어요; 다음 조각(V5 감사 뒤)이 `ureq`+`rustls`, `reqwest`(blocking), 시스템 `curl` 실행(프로세스 스캔 금지와 부딪힘) 중에서 고르고 이유와 대안을 그 SPEC에 적어요 | 없음(의존성 변경 없음) |
| D-3 | 응답은 `Content-Length`가 있는 것만 읽고, 청크 전송·리다이렉트·재시도는 하지 않아요 | 읽는 코드를 작게 두어 감사가 쉽고, 처음 보는 응답 모양은 거절(리더가 결정)로 떨어지는 쪽이 안전해요 | 실제 서버가 청크로 답하면 다음 조각에서 해석을 더해요 |
| D-4 | 스캔 B의 예외는 `src/decide/transport.rs` 한 파일의 `std::net`·`net::`·`TcpStream`·`TcpListener` 네 이름만이에요 | 스캔 주석이 "S4는 보내는 한 곳이고 자기 예외를 갖는다"고 적어 둬서 그 약속대로예요; 예외는 `std::net`·`net::`·`TcpStream`·`TcpListener` 네 이름과 그 한 파일뿐이에요; 리더가 카드 범위에 `tests/home_isolation_scan.rs`를 넣는 것을 승인했어요(2026-11-18, 카드 노트 2214 뒤) | 예외를 지우면 소켓 코드도 함께 지워야 해요 |

## Result

- AC-1: met — `cargo test --locked --offline decide::transport::tests::a_prepared_request_makes_a_round_trip_for_each_provider` (RED 먼저: 5개 시험 모두 실패, exit 101, 전송 본문이 껍데기여서). 127.0.0.1의 가짜 서버가 `POST /v1/systemone`(typesafe)·`POST /v1/decisions`(openai) `HTTP/1.1`, `Authorization: Bearer <가짜 키>`, `Content-Type`, `Content-Length`, `Connection: close`와 `prepared.body()`와 같은 바이트를 받았고, `conclude()`의 `Record`가 `nextest`와 `model:typesafe/jev-test-1`·`model:openai/gpt-6-luna`를 담아요.
- AC-2: met — 같은 실행의 `a_non_loopback_address_or_a_bad_key_is_refused_before_any_socket`과 `the_key_is_in_no_debug_output_and_no_refusal`. `192.0.2.1`·`10.0.0.1`·`[2001:db8::1]`·`0.0.0.0`은 `transport_address`로 즉시(2초 미만, 연결 대기 5초 설정에서) 거절되고, 빈 키·공백·탭·줄바꿈·NUL·비ASCII·4097바이트 키는 `key_invalid`예요. 7가지 실패 상황의 `Debug`와 이유 글에 가짜 키와 `Bearer`가 없어요.
- AC-3: met — `a_status_other_than_200_is_a_fixed_refusal_and_a_redirect_is_not_followed`. 401·403은 `transport_unauthorized`, 429는 `transport_rate_limited`, 5xx는 `transport_server_error`, 404·204는 `transport_status`, 301·302·307은 `transport_redirect`이고, 응답 본문의 글(`SECRET-ECHO…`)이 거절 어디에도 없고, 두 번째 서버는 연결을 받지 못했어요(따라가지 않음).
- AC-4: met — `a_response_that_is_too_big_short_early_closed_or_slow_is_refused`. 상한 초과는 `transport_too_large`(상한과 같은 크기는 통과), 본문이 짧음·일찍 닫음·`Content-Length` 없음·청크·잘못된 길이·길이 두 개·이상한 상태줄·끝나지 않는 헤더는 `transport_response`, 느린 서버는 `transport_timeout`(총 150ms 설정에서 0.9초 안에 돌아옴), 닫힌 포트는 `transport_connect`예요.
- AC-5: met — `decide::` 단위 시험 105 passed(`cargo test --locked --offline --lib decide::`); `home_isolation_scan` 18 passed(예외 시험 `the_decide_scan_allows_the_four_socket_names_in_the_sending_file_only`는 RED 먼저: 2 failed, exit 101, 그 뒤 GREEN)이고 `project_decide_model` 4 passed; `cargo clippy --offline --locked --all-targets -- -D warnings` exit 0; `cargo fmt --check` exit 0. 변이: `transport.rs` 10개와 스캔·가드 5개를 모두 걸렀고, 가드의 글자만 남기고 동작을 끈 변이는 스캔 시험이 아니라 `transport` 단위 시험이 걸러요. Python `repo_scans_test`·`ci_scope_test`의 19 실패는 모두 `moai-cli not built at …/target/debug/moai-cli`(이 작업 폴더가 빌드 폴더를 따로 써서)이고 CI에서 확인할 몫이에요.
- 남은 것: 실제 endpoint 호출과 TLS(D-2), `moai decide` 명령과 설정(CLI 조각). SPEC은 draft예요: 승인은 사용자의 선택이라 이 레인이 하지 않았어요.

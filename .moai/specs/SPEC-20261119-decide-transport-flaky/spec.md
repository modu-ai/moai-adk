---
id: SPEC-20261119-decide-transport-flaky
title: "decide transport: the slow-server test measures the server, not the transport"
status: superseded
revision: 1
created: 2026-11-19
approval: {by: user, at: 2026-11-21, basis: "superseded by SPEC-20261120-retire-decide — AskUserQuestion selection: 승인 (Recommended)", digest: "sha256:bf8357b17548f9216205455617aa0f48b4085f2d636fd0173084620d5c6ea783"}
---

# decide transport: the slow-server test measures the server, not the transport

> Superseded by SPEC-20261120-retire-decide (2026-11-21): the decision model library was removed.

## Intent

카드 C-115: `decide::transport::tests::a_response_that_is_too_big_short_early_closed_or_slow_is_refused` 가 부하가 큰 `cargo test --lib` 한 번에서 `src/decide/transport.rs:765`(벽시계 890 ms 단언)로 실패했어요. 같은 SHA에서 그 시험만 돌리면 3번 모두 통과했어요. C-99 규칙은 다시 돌려서 통과시키는 것을 검증으로 치지 않아요(Verify D): 뿌리 원인을 찾아 고치거나, 못 찾으면 시험을 지우고 이유를 적어요.

뿌리 원인은 코드를 읽어서 찾았어요. 시험은 `refused(Script::Delay(600 ms, …), &slow)` 를 부르는데, `refused` 는 마지막에 `fake.finish()` 로 가짜 서버 스레드를 `join` 해요. 그 서버는 요청을 읽은 뒤 600 ms 를 자요. 그래서 `started.elapsed()` 는 반드시 600 ms 이상이고, 단언 `< 590 ms + 300 ms` 은 "전송 코드가 서버를 기다리지 않았다"를 재는 것이 아니라 "서버 스레드의 600 ms 잠 + 스레드 생성·수락 대기(5 ms 간격)·join 이 290 ms 안에 끝났다"를 재요. 여유가 290 ms 뿐이라 부하에서 깨져요. 전송 코드가 시한(150 ms)에서 돌려주는지는 이 단언이 보지 않아요.

고침: 시간 한계를 잴 이유가 없어요. 전송 호출이 서버가 아직 막혀 있는 동안 `transport_timeout` 으로 돌아왔는지는 "일어난 순서"로 증명돼요. 가짜 서버가 채널에서 풀어 주는 신호를 기다리게 하고, 시험은 전송 호출이 끝난 뒤에야 신호를 보내요. 전송이 서버를 기다렸다면 신호가 오기 전에는 응답이 없으니 시험이 실패해요(응답을 받거나, 10 초 안전 시한에 서버가 답해서 `Ok` 가 와요).

## Non-goals

- 전송 코드(`post_loopback`, `read_response`, `left`)를 바꾸지 않아요. 단언이 틀린 것이지 전송이 틀린 것이 아니에요.
- 같은 파일의 다른 시간 단언(`a_non_loopback_address_or_a_bad_key_is_refused_before_any_socket` 의 2 초 상한)은 건드리지 않아요. 그 경로는 소켓을 열지 않아 마이크로초에 끝나고, 상한은 소켓을 열었다면 걸렸을 5 초보다 한참 작아서 근거가 있어요. 이 판단은 Result 에 적어요.
- `the_key_is_in_no_debug_output_and_no_refusal` 의 `Delay(500 ms)` 는 그대로 둬요. 그 시험은 시간 단언이 없고, 부하로 읽기가 늦어도 `left(deadline)` 이 0 이라 `transport_timeout` 이 나와서 결과가 같아요.
- 다른 카드(C-111)의 벽시계 단언.

## Requirements

- REQ-1: The slow-server test shall decide that the transport did not wait for the server from the order of events, with no wall-clock bound on the transport call.
- REQ-2: The slow-server test shall hold the fake server on a channel that the test releases only after the transport call has returned.
- REQ-3: The held fake server shall give up waiting for the release after 10 seconds, so a transport that waits for the server makes the test fail instead of hang.
- REQ-4 (REQ-2): IF the transport call returns a response or an error other than `transport_timeout` before the release, THEN the slow-server test shall fail.
- REQ-5 (REQ-2): IF the fake server thread has already ended when the transport call returns, THEN the slow-server test shall fail.
- REQ-6: The change shall add no sleep, retry loop or `#[ignore]` to make the test pass.
- REQ-7: The former test shall keep every assertion except the slow-server one in a single test that passes.

## Acceptance

- AC-1 (REQ-1, REQ-2, REQ-3, REQ-4, REQ-5): 새 시험이 전송 호출 뒤에 풀어 주는 가짜 서버로 `transport_timeout` 을 읽고, 그때 서버 스레드가 아직 안 끝났음을 확인해요. 전송이 서버를 기다리도록 바꾼 변이는 이 시험이 잡아요.
  verify: `cargo test --locked --offline --lib decide::transport::tests::a_slow_server_is_refused_at_the_total_bound_before_it_answers`
- AC-2 (REQ-7): 나머지 단언은 옛 이름 그대로인 한 시험에 남아 있고 통과해요(이름의 `slow` 는 새 시험이 이어받아요).
  verify: `cargo test --locked --offline --lib decide::transport::tests::a_response_that_is_too_big_short_early_closed_or_slow_is_refused`
- AC-3 (REQ-6): 변경분에 `sleep`·`#[ignore]`·재시도가 새로 없어요(Result 에 명령).
  verify: `cargo test --locked --offline --lib decide::transport`
- AC-4: 이웃 검사(`decide` 아래 전체).
  verify: `cargo test --locked --offline --lib decide::`

## Scope

- write: `.moai/specs/SPEC-20261119-decide-transport-flaky/spec.md`, `src/decide/transport.rs`
- forbidden: 실제 HOME 쓰기, 전송 코드 본문(시험 모듈 밖의 `src/decide/**`), 위 목록 밖의 파일

## Risks

- 시험이 전송 호출을 직접 불러서 `refused` 도우미(서버 스레드를 `join`)를 쓰지 않아요 → 시험 끝에서 신호를 보내고 `join` 해서 스레드가 새지 않게 해요.
- 안전 시한 10 초는 전송이 서버를 기다리는 변이에서 시험이 최대 10 초 걸린다는 뜻이에요 → 정상 경로에서는 서버가 즉시 풀려서 시간이 안 걸려요.
- 같은 파일의 2 초 상한 단언은 남아요 → Non-goals 에 근거를 적었고, 부하에서 깨지면 같은 방식(순서로 증명)으로 고쳐요.

## Decisions

| Id | Decision | Basis | Undo |
| --- | --- | --- | --- |
| D-1 | 시간 상한을 늘리지 않고 순서로 증명한다 | C-99 규칙 2는 시험을 통과시키려는 sleep·재시도·건너뛰기를 막아요. 상한을 890 ms 에서 키우면 같은 모양의 약한 단언이 남아요 | 시험을 되돌리면 돼요 |
| D-2 | 느린 서버 부분을 별도 시험으로 뺀다 | 한 시험이 여덟 가지를 보면 하나가 깨질 때 이름이 이유를 말해 주지 못해요 | 시험을 합치면 돼요 |
| D-3 | 옛 시험의 이름을 바꾸지 않는다 | `SPEC-20261118-decide-round-trip` 의 `verify:` 줄(AC-4)이 그 이름을 불러요. 이름을 바꾸면 그 줄이 시험 0개를 돌리고 조용히 통과해요 | 그 SPEC 의 줄을 고치고 이름을 바꾸면 돼요 |

## Interview

- Q1: 결정 모델 라이브러리(`src/decide/`)를 지우고 이 SPEC 을 포함한 이전 SPEC 다섯 개를 superseded 로 표시할까요? → A: 승인 (Recommended) (사용자, 2026-11-21, SPEC-20261120-retire-decide revision 1·2 승인 선택)

## Plan

1. 뿌리 원인을 재요(시험 모듈에 임시 조사 시험 하나, 커밋하지 않아요): 서버 잠 600 ms·1000 ms 에서 전송 호출이 돌아온 시각과 `refused` 전체가 끝난 시각을 읽어요.
2. SPEC 을 쓰고 새 시험을 먼저 더해요(서버를 풀어 주는 방식). 옛 시험에서 느린 서버 부분을 빼요.
3. 변이: 전송이 읽기 시한을 무시하게 하면(`set_read_timeout` 을 30 초로) 새 시험이 실패하는지, 원래대로 돌리면 통과하는지 봐요.
4. 부하 아래 20번 연속 통과를 재요(카드 기준 1).
5. `cargo fmt --check`, `cargo clippy --all-targets -- -D warnings`, `decide::` 이웃을 돌려요.

## Result

- AC-1: met — 뿌리 원인 측정(임시 조사 시험, 커밋 안 함, `cargo test --lib decide::transport::tests::probe -- --nocapture`): 서버 잠 600 ms 에서 `transport_returned_after=156ms whole_refused_after=611ms`, 1000 ms 에서 `152ms / 1011ms`, 코드는 둘 다 `transport_timeout`. 전송은 시한(150 ms)에 돌아왔고, 옛 단언이 잰 값은 서버 잠 + 11 ms 였어요. 그래서 서버 잠이 890 ms 를 넘으면 전송이 맞아도 옛 단언은 반드시 실패해요. 같은 SHA 부하 재현은 못 했어요(16코어에 busy 루프 48개, 시험 단독 20번 `pass=20 fail=0`): 실패는 라이브러리 전체를 병렬로 돌릴 때의 스레드 경합에서 나와서 단독 반복으로는 안 나와요. 새 시험 `a_slow_server_is_refused_at_the_total_bound_before_it_answers` 는 통과(`1 passed`). 변이(새 시험을 커밋한 뒤, 한 번에 하나, 원복 뒤 `git status --short` 가 비어요): M1 읽기 시한 무시(`set_read_timeout` 30 초) KILLED(`panicked at transport.rs:790`: 응답이 와서 `unwrap_err` 실패), M3 읽기 시간 초과를 EOF(`Ok(0)`)로 읽기 KILLED(`assertion left == right`), M2 `left()` 가 0 이어도 Ok 는 SURVIVED.
- AC-2: met — 옛 시험은 이름 그대로 남고 `decide::` 전체가 `114 passed; 0 failed`(`cargo test --locked --offline --lib decide::`). 이름의 `slow` 는 새 시험이 이어받아요(D-3).
- AC-3: met — 변경분에 `sleep`·`#[ignore]`·재시도·반복이 새로 없어요: `git diff origin/v2026.11.0 -- src/decide/transport.rs | grep -nE '^\+.*(sleep|#\[ignore\]|retry|\bloop\b|for _ in)'` 가 아무것도 안 찍어요(종료 1). 새 `recv_timeout(10 s)` 는 전송이 서버를 기다리는 변이에서 시험이 멈추지 않게 하는 안전 시한이에요(통과시키려는 대기가 아니에요). 부하 20번: busy 루프 48개(load average 26→47) 아래 `decide::transport` 모듈 `pass=20 fail=0`. `cargo fmt --check` 종료 0, `cargo clippy --locked --offline --all-targets -- -D warnings` 종료 0.
- AC-4: met — `decide::` 114 통과, `tests/ac_command_lint` 5 + `tests/spec_cli` 18 통과(`--no-fail-fast`). M2 는 이 카드 전부터 `decide::` 어느 시험도 못 잡는 구멍이에요(한계가 이미 지난 뒤 읽기 경로는 시계를 흉내 낼 방법 없이 결정적으로 만들 수 없어서, 시험을 더하지 않고 여기에 적어요).

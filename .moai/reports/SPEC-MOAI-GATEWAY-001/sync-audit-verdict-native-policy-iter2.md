# Sync Audit Verdict — Native Policy iter2

## Claim

`SPEC-MOAI-GATEWAY-001` 0.9.0의 iter1 `F1` CRLF native SSE 이벤트 원문 상한
수정분을 delta 범위로 재감사했다. **Overall Verdict: PASS**.

**Overall Score: 93.75/100, scoped delta**

이번 verdict는 iter1의 전체 native policy FAIL을 소급해 지우지 않는다. F1
수정분만 판정했으며, receipt·launcher·실제 provider·Windows CI의 완료를
주장하지 않는다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | LF·CRLF 비분할/1바이트 분할 × `limit-1/exact/+1` 12개 조합이 기대대로 통과했다. |
| Security (25%) | 75/100 | PASS, 범위 한정 | F1 경로에 새 명령 실행 sink나 provider egress 우회가 없음을 grep으로 확인했다. OWASP 전체 재감사는 범위 밖이다. |
| Craft (20%) | 100/100 | PASS | gateway package coverage 92.1%, native race 회귀 exit 0. |
| Consistency (15%) | 100/100 | PASS | `go vet` exit 0, 대상 `gofmt -l` 출력 없음. |

Functionality와 Security must-pass 조건을 이번 delta에서 위반한 blocking
finding은 없다.

## Findings

없음. Iteration 1의 `F1 [Medium][blocking][High confidence]`는 현재
경계시험에서 재현되지 않았다.

## Iteration History

| Iteration | Verdict | Evidence and change |
|---|---|---|
| 1 | FAIL | CRLF 원문 148 byte, limit 147에서 `success=true err=<nil>`; `Scanner.Text` 정규화 후 이벤트 크기 계산. |
| 2 | PASS | scanner split의 `advance`를 누적하고 `b.total-eventStart`로 원문 이벤트를 검사. LF·CRLF 및 분할 reader 경계를 직접 재실행. |

## Evidence

### Claim

수정 후 raw event limit 경계가 LF와 CRLF 모두에서 정확히 동작한다.

### Evidence

Command:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json ./internal/gateway -run '^TestNativeEventRawLimitIndependent$' -count=1 -v -timeout=30s
```

Observed output:

```text
=== RUN   TestNativeEventRawLimitIndependent
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=144 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=145 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=false largest_raw_event=145 event_limit=146 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=144 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=145 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=false fragmented=true largest_raw_event=145 event_limit=146 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=149 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=149 success=true err=<nil>
--- PASS: TestNativeEventRawLimitIndependent (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.600s
```

혼합 LF/CRLF의 추가 임시 fixture도 다음과 같이 통과했다.

```text
=== RUN   TestNativeEventMixedRawByteBoundaries
    mixed_event_test.go:49: mixed fragmented=false largest_raw_event=146 event_limit=145 success=false err=native response stream failed
    mixed_event_test.go:49: mixed fragmented=false largest_raw_event=146 event_limit=146 success=true err=<nil>
    mixed_event_test.go:49: mixed fragmented=false largest_raw_event=146 event_limit=147 success=true err=<nil>
    mixed_event_test.go:49: mixed fragmented=true largest_raw_event=146 event_limit=145 success=false err=native response stream failed
    mixed_event_test.go:49: mixed fragmented=true largest_raw_event=146 event_limit=146 success=true err=<nil>
    mixed_event_test.go:49: mixed fragmented=true largest_raw_event=146 event_limit=147 success=true err=<nil>
--- PASS: TestNativeEventMixedRawByteBoundaries (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.407s
```

### Claim

기존 native state/cancellation/EOF/whole-response 회귀와 package quality gate가
유지된다.

### Evidence

```text
$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -race ./internal/gateway -run 'TestNative|TestAnthropic' -count=1 -timeout=90s
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.549s

$ unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test ./internal/gateway -cover -count=1 -timeout=90s
ok  	github.com/modu-ai/moai-adk/internal/gateway	6.507s	coverage: 92.1% of statements

$ go vet ./internal/gateway; gofmt -l internal/gateway/anthropic.go internal/gateway/native_event_limit_test.go
(no output; exit 0)
```

### Baseline-attribution

```text
$ pwd
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git branch --show-current
WT-unified-gateway
```

위 command들은 현재 dirty worktree를 baseline으로 실행했다. F1 수정 대상은
`internal/gateway/anthropic.go:441-448,484-489`이며, 경계 시험은
`internal/gateway/native_event_limit_test.go:12-46`이다. 제공된 overlay는
`/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json`이다.

### Gaps

- iter1 RED는 수리 보고서에 보존된 이전 실행이며, 현재 수정본에서 이전 구현을
  되돌려 RED를 새로 만들었다고 주장하지 않는다.
- 실제 provider, Windows runtime, GitHub CI run/artifact는 실행하지 않았다.
- 전체 SPEC acceptance matrix, receipt codec, launcher, `moai gpt` 사용자
  경로는 이 delta verdict에 포함하지 않았다.
- grep dependency probe는 manifest 확인이며 취약점 DB 대조가 아니다.

### Residual-risk

이 verdict는 `ScanLines`가 줄바꿈을 정규화하기 전의 `advance`를 기준으로
원문 소비량을 제한하는 현재 구현과 합성 fixture를 검증한다. 새로운 SSE 문법,
provider별 비정상 chunking, 실제 운영 transport, 전원 장애 복구를 입증하지 않는다.

## Recommendations

F1 delta는 PASS로 닫을 수 있다. 후속 통합 감사에서 이 파일의 PASS를 전체
gateway 제품 완료로 승격하지 말고 receipt·launcher·실제 provider·Windows
CI 증거를 별도로 확인한다.

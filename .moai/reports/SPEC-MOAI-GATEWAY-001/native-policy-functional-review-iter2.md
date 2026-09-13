# Native 정책 독립 감사 — iter2

## Claim

대상 SPEC은 `SPEC-MOAI-GATEWAY-001` 0.9.0이며, 이번 감사의 범위는 iter1의
`F1`인 native SSE 이벤트별 원문 바이트 상한 보정분으로 한정한다. 평가 대상은
`internal/gateway/anthropic.go`의 bounded scanner 변경과 그 경계 시험이다.

`F1` delta의 판정은 **PASS**이다. 수정 후 LF·CRLF 각각의 비분할 및 1바이트
분할 reader에서 `limit-1`은 실패하고 `exact`와 `limit+1`은 성공했다. 혼합
줄바꿈의 같은 여섯 경계도 별도 임시 semantic fixture에서 통과했다.

이번 판정은 iter1의 native 정책 전체 감사 결과를 대체하지 않는다. receipt,
launcher, 실제 provider, Windows 실행 환경은 이번 delta의 판정 대상이 아니다.

## Dimension Scores

활성 프로파일은 `default.md`이다. 아래 점수는 F1 delta와 현재 패키지에서 직접
측정한 범위에만 적용한다. 이전 iter1에서 확인한 전체 native 정책의 미완료
접점은 이 점수에 다시 포함하지 않았다.

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | LF·CRLF 비분할/분할 × `limit-1`, exact, `limit+1` 12개 조합의 verbatim 출력에서 경계가 모두 기대대로 판정되었다. 혼합 줄바꿈 여섯 조합도 PASS이다. |
| Security (25%) | 75/100 | PASS, 범위 한정 | F1 수정은 원문 소비량을 제한하는 bounded scanner 경로이며 새 egress·명령 실행 sink를 추가하지 않았다. OWASP 전체 재감사는 수행하지 않았으므로 1.00은 주장하지 않는다. |
| Craft (20%) | 100/100 | PASS | `go test ./internal/gateway -cover`에서 `coverage: 92.1% of statements`; native race 회귀도 exit 0이다. |
| Consistency (15%) | 100/100 | PASS | `go vet ./internal/gateway`가 exit 0, 대상 `gofmt -l` 출력이 비었다. |

**Overall Verdict: PASS — F1 delta only**  
**Overall Score: 93.75/100, scoped delta**

Functionality와 Security의 must-pass 조건을 이 delta 범위에서 위반한 결함은
확인되지 않았다. 전체 SPEC 구현 완료나 실제 운영 경로 PASS로 확대 해석하지
않는다.

## Findings

이번 delta에서 blocking defect는 없다.

## Iteration History

### Iteration 1 — FAIL

iter1은 `Scanner.Text`가 CRLF를 제거한 뒤 계산되어 원문 이벤트가 한도를
초과해도 성공 terminal을 내보내는 `F1 [Medium][blocking]`을 기록했다.
수리 보고서에 보존된 RED의 핵심 출력은 다음과 같다.

```text
CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=true err=<nil>
native_event_independent_test.go:12: raw event over limit accepted
CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=true err=<nil>
native_event_independent_test.go:12: raw event over limit accepted
--- FAIL: TestNativeEventRawLimitIndependent (0.00s)
```

### Iteration 2 — bounded delta PASS

수리 후 `anthropic.go:441-448`의 scanner split이 `bufio.ScanLines`의
`advance`를 누적하여 CRLF 제거 전 wire 소비량을 `b.total`에 반영하고,
`anthropic.go:484-489`가 이벤트 시작점부터의 원문 소비량을 검사한다.
`native_event_limit_test.go:12-46`은 LF·CRLF, 비분할·1바이트 분할, 세 경계값을
고정한다. 아래 직접 실행 결과로 F1을 재판정했다.

## Evidence

### Claim: LF·CRLF 원문 이벤트 상한이 경계값대로 거부·수용된다

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

### Claim: mixed LF/CRLF와 1바이트 분할에서도 같은 원문 이벤트 판정이 유지된다

임시 fixture는 제품 파일에 포함하지 않고, 제공된 overlay와 같은 방식으로
현재 gateway package에만 주입했다.

Command:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/mixed_overlay.json ./internal/gateway -run '^TestNativeEventMixedRawByteBoundaries$' -count=1 -v -timeout=30s
```

Observed output:

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

### Claim: native stream state, cancellation, EOF와 raw whole-response 회귀가 유지된다

Command:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test -race ./internal/gateway -run 'TestNative|TestAnthropic' -count=1 -timeout=90s
```

Observed output:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.549s
```

### Claim: 현재 gateway package coverage가 프로파일 임계값을 넘는다

Command:

```sh
unset CODEX_HOME OPENAI_API_KEY OPENAI_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN ANTHROPIC_BASE_URL Z_AI_API_KEY && go test ./internal/gateway -cover -count=1 -timeout=90s
```

Observed output:

```text
ok  	github.com/modu-ai/moai-adk/internal/gateway	6.507s	coverage: 92.1% of statements
```

### Claim: F1 경로에 새 명령 실행·unsafe sink가 없고 dependency manifest가 읽힌다

Command:

```sh
rg -n 'os/exec|exec\.Command|sh -c|unsafe|eval|MaxEventBytes|MaxOutputBytes' internal/gateway/anthropic.go internal/gateway/native_event_limit_test.go; go list -m -json all | rg -n '"Path"|"Version"' | sed -n '1,24p'
```

Observed output:

```text
internal/gateway/native_event_limit_test.go:32:				b := newNativeBody(context.Background(), io.NopCloser(reader), "canonical", translate.Limits{MaxOutputBytes: 1 << 20, MaxEventBytes: longest + offset})
internal/gateway/anthropic.go:48:	if c.Transport == nil || c.Transport.Proxy != nil || c.Transport.TLSClientConfig != nil && c.Transport.TLSClientConfig.InsecureSkipVerify || c.AnthropicVersion == "" || strings.ContainsAny(c.AnthropicVersion, "\r\n") || c.Limits.MaxBodyBytes <= 0 || c.Limits.MaxEventBytes <= 0 || c.Limits.MaxOutputBytes <= 0 {
internal/gateway/anthropic.go:191:	raw, e := io.ReadAll(io.LimitReader(up.Body, int64(a.config.Limits.MaxOutputBytes)+1))
internal/gateway/anthropic.go:192:	if e != nil || len(raw) > a.config.Limits.MaxOutputBytes {
internal/gateway/anthropic.go:443:		remaining := b.limits.MaxOutputBytes - b.total
internal/gateway/anthropic.go:450:	b.scanner.Buffer(make([]byte, min(4096, limits.MaxEventBytes)), limits.MaxEventBytes)
internal/gateway/anthropic.go:488:		if b.total-eventStart > b.limits.MaxEventBytes {
2:	"Path": "github.com/modu-ai/moai-adk",
9:	"Path": "charm.land/bubbles/v2",
10:	"Version": "v2.2.1",
19:	"Path": "charm.land/bubbletea/v2",
20:	"Version": "v2.0.9",
29:	"Path": "charm.land/fang/v2",
30:	"Version": "v2.0.1",
39:	"Path": "charm.land/huh/v2",
40:	"Version": "v2.0.3",
49:	"Path": "charm.land/lipgloss/v2",
50:	"Version": "v2.0.6",
59:	"Path": "github.com/MakeNowJust/heredoc",
60:	"Version": "v1.0.0",
70:	"Path": "github.com/a-h/parse",
71:	"Version": "v0.0.0-20250122154542-74294addb73e",
```

### Claim: formatting and static analysis are clean for the touched package

Command:

```sh
go vet ./internal/gateway; gofmt -l internal/gateway/anthropic.go internal/gateway/native_event_limit_test.go
```

Observed output:

```text
(no output; exit 0)
```

## Baseline-attribution

모든 현재 실행은 다음 worktree에서 수행했다.

```text
$ git rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git branch --show-current
WT-unified-gateway
```

최종 확인 시점의 대상 파일 상태는 다음과 같다.

```text
?? internal/gateway/anthropic.go
?? internal/gateway/native_event_limit_test.go
```

이 worktree는 dirty 상태이며, 위 두 파일은 기존 작업자가 작성한 변경이다.
이번 감사에서는 제품·SPEC·workflow 파일을 수정하지 않았다. 제공된 원래
overlay는 다음 경로를 사용했다.

```text
/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json
```

mixed fixture는 추가 검증용 임시 파일이며 제품 트리에 남기지 않았다. iter1의
RED는 `native-event-limit-repair.md`에 보존된 이전 실행을 인용한 것이며, 이번
iter2의 현재 제품 상태에서 재실행한 RED라고 주장하지 않는다. 현재 상태에서
직접 실행한 것은 위 GREEN 및 회귀 명령이다.

## Gaps

- 실제 Claude/Anthropic provider를 통한 native SSE 재송신은 하지 않았다.
- Windows native runtime과 GitHub Windows CI run/artifact는 관측하지 않았다.
- receipt codec, UUID 귀속, launcher profile binding, `moai gpt` 전체 사용자
  경로는 이번 delta 판정에 포함하지 않았다.
- dependency 명령은 manifest를 읽은 것이며 취약점 데이터베이스 대조가 아니다.
- 전체 SPEC의 모든 acceptance criterion에 대한 새 감사가 아니다.

## Residual-risk

`b.total`은 scanner가 소비한 wire byte를 누적하고, `eventStart`는 각 SSE
이벤트의 시작 offset을 보존한다. 그러므로 이번 경계는 CRLF가
`ScanLines`에서 정규화되기 전 크기를 기준으로 한다. 이 판정은 새로운 SSE 문법을
허용한다는 뜻이 아니며, 기존 event state validator와 whole-response bound의
정확성을 대체하지 않는다. 실제 provider의 비정상 chunking과 운영 transport의
동작은 별도 검증이 필요하다.

## Recommendations

- 이 delta에 대해 F1 재감사를 종료할 수 있다.
- 후속 통합 감사에서는 이 PASS를 receipt·launcher·실제 provider·Windows CI
  완료로 승격하지 말고, 해당 증거를 별도로 수집한다.
- 임시 mixed fixture는 제품 테스트로 채택할지 다음 통합 담당자가 결정한다.

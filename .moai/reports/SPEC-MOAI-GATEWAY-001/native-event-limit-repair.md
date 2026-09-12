# Native SSE F1 raw event 한도 보정

## Claim

감사의 F1을 제공된 독립 overlay로 재현하고 수정했다. anthropic.go의 next가 Scanner.Text의 정규화된 줄 길이 대신, ScanLines가 CRLF를 제거하기 전에 누적한 실제 소비 wire bytes와 event 시작 offset의 차이를 사용한다. 빈/comment-only frame을 처리한 뒤 기준 offset을 다시 설정한다. whole MaxOutputBytes·EOF·취소·정책 의미는 변경하지 않았다.

수정 파일은 internal/gateway/anthropic.go와 새 internal/gateway/native_event_limit_test.go, 이 보고서뿐이다. opaque 패키지는 앞선 인계를 마친 뒤 추가 수정하지 않았다. translator·SPEC·AUTH·receipt·원격은 변경하지 않았다.

## Evidence

제공된 `/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json`를 직접 읽고 다음 RED를 실행했다:

```text
go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json ./internal/gateway -run TestNativeEventRawLimitIndependent -count=1
--- FAIL: TestNativeEventRawLimitIndependent (0.00s)
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=true err=<nil>
    native_event_independent_test.go:12: raw event over limit accepted
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=true err=<nil>
    native_event_independent_test.go:12: raw event over limit accepted
FAIL
```

수정 후 같은 독립 시험의 GREEN:

```text
go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/native-event-audit-e78tbex_/overlay.json ./internal/gateway -run TestNativeEventRawLimitIndependent -count=1 -v
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=147 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=false largest_raw_event=148 event_limit=149 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=147 success=false err=native response stream failed
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=148 success=true err=<nil>
    native_event_independent_test.go:11: CRLF=true fragmented=true largest_raw_event=148 event_limit=149 success=true err=<nil>
--- PASS: TestNativeEventRawLimitIndependent (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/gateway 0.395s
```

LF 군도 largest145의 limit144 FAIL,145/146 PASS였다. 독립 시험과 같은 12개 경계 조합을 지속 회귀 시험 TestNativeEventRawByteBoundaries로 추가했다. 구현보다 먼저 실행한 RED는 위 독립 overlay다.

```text
go test -race ./internal/gateway -run 'TestNative|TestAnthropic' -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway 1.540s

go vet ./internal/gateway
(exit 0, output empty)
```

기존 whole raw wire bound·fragmentation·native text/tool·thinking·EOF·blocked-read cancellation 회귀군을 포함한다. gofmt -l 출력은 비어 있었다.

## Baseline-attribution

WT /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified. HEAD81c1d58f9, branchWT-unified-gateway를 수정 후 재확인했다. source_session_id는 부모 제공 01a08e7b-6aa0-7361-ab7e-ea8da1f02228이며 child session 조회 성공으로 표시하지 않는다.

수정 후 source SHA-256:

```text
872d89371e6ec0ebc5a7c1050d011091abe0740c26f45d1fa9ac81560e98c255  internal/gateway/anthropic.go
be740898abd400c777df7544b03bb690af2edfb1050be3b8a3416b5eb57cee66  internal/gateway/native_event_limit_test.go
```

## Gaps

독립 감사관의 delta 재판정, 실제 provider 재송신, 전체 gateway/launcher 통합과 Windows runtime은 이번 수리에서 실행하지 않았다. integration branch GitHub CI가 repository-wide 판정 소유자이며 현재 run ID가 없는 상태는 PENDING이다.

## Residual-risk

이 수정은 실제 소비 byte를 기준으로 event 한도를 판정하며 SSE 출력 줄 끝 형식을 원문과 동일하게 만드는 변경은 아니다. 기존 parser가 허용하지 않는 새로운 event grammar의 지원을 주장하지 않는다.

추가 shared adapter 대조군:

```text
go test -race ./internal/gateway -run 'TestGLMNative|TestOAuthNative' -count=1
ok  github.com/modu-ai/moai-adk/internal/gateway 1.306s
```

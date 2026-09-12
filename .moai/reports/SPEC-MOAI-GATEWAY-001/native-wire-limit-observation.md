# Native SSE 원문 바이트 상한 재현

## Claim

translate SSE 수리와 별도로 native Messages SSE에도 CRLF 원문 상한 누락이 재현된다. 640 byte 응답에 상한639를 주어도 성공 terminal이 반환된다. LF 음성/양성 대조군과 CRLF 정확상한/여유 대조군을 함께 실행했다.

## Evidence

명령:
```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && GOCACHE=/tmp/gateway-policy-cache go test -overlay /var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/gateway-native-wire-audit-ybwwqxec/overlay.json ./internal/gateway -run '^TestNativeWireRawLimitAudit$' -count=1 -v -timeout 30s
```
원문 출력:
```text
=== RUN   TestNativeWireRawLimitAudit
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=621 success=false err=native response stream failed
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=622 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=623 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=639 success=true err=<nil>
    native_wire_audit_test.go:5: raw wire exceeds limit yet success
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=640 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=641 success=true err=<nil>
--- FAIL: TestNativeWireRawLimitAudit (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/gateway 0.499s
FAIL
```

## Baseline-attribution

2026-09-11, WT-unified-gateway /81c1d58f9의 미커밋 트리. internal/gateway/anthropic.go:428의 Scanner.Text 뒤 len(line)+1 계수 구간을 읽고 기존 nativeSSE 합성 fixture를 newNativeBody에 직접 넣었다. 제품·시험 소스를 수정하지 않는 임시 Go overlay로 실행했다.

## Gaps

실제 provider/HTTP 경로, native의 모든 상태/필드, Windows는 시험하지 않았다. 수리 뒤 같은 probe와 cancel/EOF/order/분할 reader의 회귀 확인이 필요하다.

## Residual-risk

factory/SSE 감사 대상인 translate 수리의 실패를 뜻하지 않는다. 별도 native 구현에서 실행으로 확인한 결함이다.

## 수리 뒤 부모 재실행

같은 명령과 같은 임시 overlay를 현재 수리된 native 코드에 다시 실행했다. 이전 실패 기록은 위에 보존한다. exit0 원문:
```text
=== RUN   TestNativeWireRawLimitAudit
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=621 success=false err=native response stream failed
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=622 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=false bytes=622 limit=623 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=639 success=false err=native response stream failed
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=640 success=true err=<nil>
    native_wire_audit_test.go:5: CRLF=true bytes=640 limit=641 success=true err=<nil>
--- PASS: TestNativeWireRawLimitAudit (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/gateway 0.390s
```
이 부모 재검증은 위 여섯 경계 사례에 한정한다. 전체 native 정책 지원과 새 독립 감사는 별도다.

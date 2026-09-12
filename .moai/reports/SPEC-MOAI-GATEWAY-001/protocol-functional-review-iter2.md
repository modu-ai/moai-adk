# PFR-F1 기능 재검증 — iter2

## Claim

**Functional delta verdict: PASS. PFR-F1: 해결 확인.** 원래 실패했던 합성 public SSE를 그대로 수락하면서, Stream의 content block index 순서가 동일 terminal Response의 `[first, second]`와 일치한다. 유효 입력을 새 미지원 오류로 거절하는 경우에는 실패하도록 독립 probe를 강화했다.

terminal과 text done을 주지 않고 첫 text를 관측했다. 이후 취소하면 오류로 돌아오며 `message_stop`을 붙이지 않는다. 선행 message 폭이 확정되면 뒤 item도 terminal 전 출력한다. 모든 terminal 전 절단점, 보관된 출력의 writer 오류, 같은 item의 역순 delta/완료, 여러 text part 뒤 tool ID·이름·index 대조가 통과했다.

이 판정은 **기존 PFR-F1의 일반 프로토콜 기능 변경분**에 한정한다. 전체 SPEC, 인증, 보안, 실제 공급자, 출시 준비의 PASS가 아니다. 이번 범위에서 새 blocking finding은 관측하지 않았다.

수리 구현은 `internal/gateway/translate/stream.go`의 `emitBlock`/`layout`이다. 앞 message의 최종 폭을 아직 모르는 동안 뒤 item 이벤트를 보관하고, 확정된 `base + content_index`에 배정한다. 이미 번호가 확정된 첫 item의 text는 즉시 출력한다. 작성자의 최종 반환과 `protocol-functional-repair-pfr-f1.md`를 확인한 뒤 기준 해시를 채집하고 시험했다. 제품 파일을 수정하지 않았다.

## Evidence

모든 명령은 아래 Baseline-attribution의 WT에서 실행했다. 외부 시스템·실제 client·HTTP endpoint를 호출하지 않았다.

독립 probe와 관련 영구 회귀 시험:

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test -race -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/overlay.json ./internal/gateway/translate -run '^TestFunctionalDelta|^TestFunctionalEveryTruncated|^TestFunctionalSameItem|^TestStreamBuffered|^TestStreamDeferred|^TestStreamCanonicalOffset' -count=1 -v -timeout 15s
=== RUN   TestFunctionalDeltaOriginalFixtureMustSucceedInOrder
    protocol_delta_test.go:12: original interleaving accepted: stream=[first second] response=[first second] terminal=1
--- PASS: TestFunctionalDeltaOriginalFixtureMustSucceedInOrder (0.00s)
=== RUN   TestFunctionalDeltaInterleavedPrefixVisibleBeforeTerminalAndCancel
    protocol_delta_test.go:28: first output text visible while terminal withheld; cancel returned error with terminal=0
--- PASS: TestFunctionalDeltaInterleavedPrefixVisibleBeforeTerminalAndCancel (0.00s)
=== RUN   TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal
    protocol_functional_test.go:31: all pre-terminal text prefixes failed without message_stop
--- PASS: TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal (0.00s)
=== RUN   TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices
    protocol_functional_test.go:43: same-item reverse delta/completion arrival preserves content block indices
--- PASS: TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices (0.00s)
=== RUN   TestStreamBufferedSuccessorFlushesBeforeTerminal
--- PASS: TestStreamBufferedSuccessorFlushesBeforeTerminal (0.00s)
=== RUN   TestStreamCanonicalOffsetIncludesEveryPrecedingContentPart
--- PASS: TestStreamCanonicalOffsetIncludesEveryPrecedingContentPart (0.00s)
=== RUN   TestStreamBufferedOrderingErrorsNeverFinish
--- PASS: TestStreamBufferedOrderingErrorsNeverFinish (0.01s)
=== RUN   TestStreamDeferredWriterFailureHasNoTerminal
--- PASS: TestStreamDeferredWriterFailureHasNoTerminal (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.434s
```

Exit 0. 새 prefix probe는 `io.Pipe`에 원본의 첫 item text delta까지 쓰고 terminal을 보류한다. writer가 `first` text를 관측해야 진행하며, 2초 context와 defer의 cancel/close/join으로 양쪽 goroutine을 정리한다. 출력은 goroutine 종료 신호를 받은 뒤 읽는다. 원본 fixture 성공 판정은 Response 결과와 block index별 text를 비교하고 정상 terminal 정확히 1회를 요구한다.

변환기 패키지의 범위 회귀 및 기계 검사:

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test -race ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-protocol-review-iter2.cover -timeout 20s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.914s	coverage: 92.7% of statements
$ GOCACHE=/tmp/gateway-protocol-review-cache go vet ./internal/gateway/translate && gofmt -l internal/gateway/translate/stream.go internal/gateway/translate/stream_order_test.go
```

둘 다 exit 0. vet/gofmt stdout·stderr는 빈 출력이다. 위 두 명령과 독립 delta 시험은 서로 다른 결과 파일을 쓰는 독립 묶음으로 실행했다. Coverage는 이 패키지의 이번 실행 수치이며 전체 gateway 수치가 아니다.

호출부의 순수 body 대조:

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/overlay.json ./internal/gateway -run '^TestFunctionalNativeBody|^TestFunctionalOpenAIStreamBody|^TestAnthropicNativeInterleavedToolStreams$' -count=1 -v -timeout 15s
=== RUN   TestAnthropicNativeInterleavedToolStreams
--- PASS: TestAnthropicNativeInterleavedToolStreams (0.00s)
=== RUN   TestFunctionalNativeBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:11: native public text preserved byte-for-byte; every pre-terminal prefix failed without message_stop
--- PASS: TestFunctionalNativeBodyTextAndAllTruncatedPrefixes (0.00s)
=== RUN   TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:22: OpenAI body emitted text once and one normal terminal; every pre-terminal prefix returned an error
--- PASS: TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.442s
```

Exit 0. 이 세 시험은 합성 JSON/SSE, `io.Reader`, 변환 body만 호출한다. 인증이나 credential 시험을 선택하지 않았다.

## Baseline-attribution

```text
$ git rev-parse HEAD && git branch --show-current
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
WT-unified-gateway
```

Exit 0. WT는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`다. 커밋되지 않은 현재 파일이 검사 대상이므로 HEAD만으로 수리 내용을 식별하지 않는다.

- `stream.go` SHA-256: `9940da12e7ec5e8209a78b400490e5caa64ac3d14e95b6e4951151585fdec02a`.
- `stream_order_test.go`: `5b46d7d74f71aeff5267181cd07b0eacc663bce1d8ed4239816e538d33ffb494`.
- 전체 검사 기준 10개 파일의 SHA-256은 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review-iter2/baseline-sha256.txt`에 보관했다. translate의 Go 파일과 openai.go/anthropic.go/glm.go가 포함된다.
- 새 probe는 같은 디렉터리의 `delta_test.go`, overlay는 `overlay.json`이다. overlay는 새 가상 시험 파일만 덧붙이며 제품 파일을 대체하지 않는다.
- 원본 fixture와 terminal은 앞 iter1의 `functional-review/` 디렉터리에 있고, 영구 사본은 `internal/gateway/translate/testdata/`에 있다.

시험 뒤 실행한 Python 표준 라이브러리 해시·바이트 비교:

```text
baseline files: 10 changed: []
interleaved-text.sse same: True sha256: 237cd4961fbd3d3ad73f22f692a7f5f8567976f907ca6ab6b41e13314e394941
interleaved-text-terminal.json same: True sha256: 1567613753d4016cc31484de54c1f88857efa10ee9dcdd34616fda4e3defb8a1
```

이전 `protocol-functional-review.md`의 FAIL을 보존한다. Iter1은 `[second, first]`와 `[first, second]` 불일치를 관측했고, iter2는 동일 바이트 fixture와 엄격한 성공 요구로 변경분을 재검증했다.

## Gaps

- 실제 공급자의 해당 interleaving 발생 여부와 Claude TUI의 표시 동작은 측정하지 않았다.
- 인증·자격증명·보안·실제 HTTP adapter 통합·opaque reasoning·정책·Windows·CG 마이그레이션은 이 판정에 포함하지 않는다.
- 같은 item의 content part **시작**을 1→0으로 뒤집는 지원 확대는 검증하지 않았다. 기존 대조는 시작 0→1, delta/완료 1→0이다.
- 전체 SPEC AC, 전체 저장소 CI, 출시 결합 조건을 판정하지 않았다. 전체 목표는 이 보고서만으로 완료되지 않는다.
- 최대 byte budget 경계에서의 실제 메모리 사용량과 오래 지연되는 upstream 성능은 측정하지 않았다.

## Residual-risk

앞 message의 폭이 알려지기 전 뒤 item 출력이 지연되는 조건은 남는다. 이번 시험은 이 조건이 첫 item의 준비된 prefix 또는 선행 item 완료 뒤의 successor까지 terminal을 기다리게 만들지 않는지 확인한다. 합성 ordinary fixture와 정해진 byte 한도 아래의 기능 검증을 모든 공급자 응답 형식의 호환성으로 확대하지 않는다.

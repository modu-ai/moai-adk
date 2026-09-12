# PFR-F1 출력 순서 수리 검증

## Claim

`protocol-functional-review.md`의 PFR-F1을 같은 원본 SSE/terminal JSON으로 재현한 뒤 수리했다.
성공한 Stream의 content block index 순서가 같은 Response 결과의 `[first, second]`와 일치한다.
정상 interleaving을 거절하거나 전체 출력을 terminal까지 보관하는 방법을 쓰지 않았다.
이 문서는 한 결함의 구현 검증이며 독립 감사나 전체 gateway 완료 판정이 아니다.

수리 내용은 translate/stream.go 안에 한정된다. 각 output item의 canonical base를 추적하며 block index는
`base + content_index`로 정한다. Function call은 시작부터 블록 수가 1임을 알 수 있다. Message는 output_item.done에서
전체 블록 수가 확정되므로, 앞 message의 폭을 알기 전 뒤 item의 출력만 요청 단위로 보관한다.
번호가 확정된 prefix는 곧바로 출력하고, 앞 item의 폭이 확정되면 terminal을 기다리지 않고 뒤 항목을 내보낸다.
이미 번호를 정한 블록을 매번 다시 순회하지 않도록 item별 cursor를 유지한다. 기존 upstream 바이트 한도는 유지된다.

새 영구 시험은 다음을 판정한다.

- 원본 fixture를 성공으로 수락하면서 Stream과 동일 terminal Response의 text 순서 일치.
- terminal은 물론 text done 전에도 이미 준비된 정상 prefix를 출력.
- 선행 item.done 뒤에는 terminal을 읽기 전에 보관된 successor를 출력.
- 앞 message의 text 두 개와 뒤 tool이 `text 0, text 1, tool 2`로 배정되며 도구 ID·이름을 보존.
- 보관된 출력의 writer 실패 및 원본 fixture의 모든 terminal 전 절단점에서 성공 terminal 없음.

## Evidence

작업 디렉터리는 아래 WT다. 원본 fixture의 동일 사본을 새 영구 시험으로 읽어 구현 전에 실행한 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -run 'TestStreamCanonicalOutputOrderRegression|TestStreamReadyPrefix|TestStreamBufferedSuccessor' -count=1
--- FAIL: TestStreamCanonicalOutputOrderRegression (0.00s)
    stream_order_test.go:60: stream=[second first] nonstream=[first second]
--- FAIL: TestStreamBufferedSuccessorFlushesBeforeTerminal (0.00s)
    stream_order_test.go:117: upstream SSE read: successor not released after predecessor width known: [second first]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.401s
FAIL
```

Exit 1. 수리 후 전체 변환기 첫 GREEN:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway/translate -count=1 -timeout 20s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.390s
```

추가 다중 content/tool/writer 실패 시험과 cursor 보강 후 최종 검증:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-pfr-f1-cover.out -timeout 20s
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.655s	coverage: 92.7% of statements
$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway/translate
$ GOCACHE=/tmp/gateway-translation-cache gopls check internal/gateway/translate/stream.go internal/gateway/translate/stream_order_test.go
```

각 exit 0. vet/gopls stdout·stderr는 비었다. 이 세 검사와 원본 overlay 시험을 독립 묶음으로 실행했다.

원본 기능 검증의 정확한 overlay를 다시 실행했다.

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review/overlay.json ./internal/gateway/translate -run '^TestFunctional' -count=1 -v -timeout 15s
=== RUN   TestFunctionalStreamTextOrderMatchesCompleteResponse
--- PASS: TestFunctionalStreamTextOrderMatchesCompleteResponse (0.00s)
=== RUN   TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal
    protocol_functional_test.go:31: all pre-terminal text prefixes failed without message_stop
--- PASS: TestFunctionalEveryTruncatedTextPrefixHasNoSuccessfulTerminal (0.00s)
=== RUN   TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices
    protocol_functional_test.go:43: same-item reverse delta/completion arrival preserves content block indices
--- PASS: TestFunctionalSameItemReverseDeltaArrivalPreservesContentIndices (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.409s
```

Exit 0. 원본 overlay는 오류로 거절되는 경우도 통과할 수 있으므로 그것만으로 수리했다고 판단하지 않았다.
새 영구 regression은 valid interleaving에 오류가 나면 반드시 실패하고, 성공한 출력 순서를 비교한다.

호출부의 순수 body 대조도 다시 실행했다.

```text
$ GOCACHE=/tmp/gateway-protocol-review-cache go test -overlay .moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/functional-review/overlay.json ./internal/gateway -run '^TestFunctional|^TestAnthropicNativeInterleavedToolStreams$' -count=1 -v -timeout 15s
=== RUN   TestAnthropicNativeInterleavedToolStreams
--- PASS: TestAnthropicNativeInterleavedToolStreams (0.00s)
=== RUN   TestFunctionalNativeBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:11: native public text preserved byte-for-byte; every pre-terminal prefix failed without message_stop
--- PASS: TestFunctionalNativeBodyTextAndAllTruncatedPrefixes (0.00s)
=== RUN   TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes
    protocol_functional_body_test.go:22: OpenAI body emitted text once and one normal terminal; every pre-terminal prefix returned an error
--- PASS: TestFunctionalOpenAIStreamBodyTextAndAllTruncatedPrefixes (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.403s
```

Exit 0. JSON/SSE·io.Reader만 사용했으며 네트워크 endpoint·Claude·Codex·실제 공급자를 호출하지 않았다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.
- 실행 뒤 `git rev-parse --short HEAD` → `81c1d58f9`, exit 0. 다른 작성자의 변경은 되돌리지 않았다.
- 수리 전 stream.go SHA-256: `fa09aa2991e032b48b85fddf969f208fcf87426d6a944574e9898f49dbb6a78f` (읽은 원본 기능 검증 기준).
- 수리 후 stream.go: `9940da12e7ec5e8209a78b400490e5caa64ac3d14e95b6e4951151585fdec02a`.
- 새 stream_order_test.go: `5b46d7d74f71aeff5267181cd07b0eacc663bce1d8ed4239816e538d33ffb494`.
- 영구 fixture는 `internal/gateway/translate/testdata/`에 있다. 원본과 byte-for-byte 비교 결과 둘 다 True였다.
  - interleaved-text.sse: 2,833 bytes, `237cd4961fbd3d3ad73f22f692a7f5f8567976f907ca6ab6b41e13314e394941`.
  - interleaved-text-terminal.json: 395 bytes, `1567613753d4016cc31484de54c1f88857efa10ee9dcdd34616fda4e3defb8a1`.
- 변경 소유권: translate stream 구현·새 시험/fixture·이 delta 보고서만. 원본 기능 검증 보고서는 과거 FAIL 근거로 보존했다.

## Gaps

- 실제 공급자가 이 interleaving을 내는지, 실제 Claude TUI가 모든 index 패턴을 어떻게 표시하는지는 측정하지 않았다.
- 기존 output_item/같은 item의 content part 시작 순서 검증을 확대하거나 바꾸지 않았다. 이번 반례는 item 시작 0→1,
  뒤 item의 part·delta·완료가 먼저 오는 유효하게 수락되던 경로다. 같은 item 대조는 part 시작 0→1, delta·완료 1→0이다.
- 앞 message의 전체 폭이 알려지지 않은 동안에는 뒤 item을 내보낼 수 없다. 시험은 필요한 이 지연과 전체 terminal buffering을 구분한다.
- 인증·credential·보안·전체 adapter HTTP·모델 정책·reasoning·Windows·전체 provider 통합은 이번 수리의 검증 범위 밖이다.
- 변경 전 별도 LSP snapshot을 채집하지 않았으며 현재 파일 진단만 확인했다. 전체 저장소 verdict는 통합 브랜치 CI 담당이며 PENDING이다.
  Commit·push·PR·merge를 하지 않았다.

## Residual-risk

나중 item의 이벤트를 보관하는 동안의 실제 메모리 사용량과 공급자의 장시간 지연은 별도 측정하지 않았다. 기존 upstream 전체 바이트
한도는 유지되며 새 보관 상태는 요청 단위다. 현재 시험의 성공을 임의의 provider 정책·opaque 이력 호환성이나 전체 사용자 목표 완료로 확대하지 않는다.

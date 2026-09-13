# 일반 Messages·Responses 변환 검증

## Claim

SPEC 0.8.0 M5의 일반 공개 text·완료된 tool pair·Responses/SSE 변환을 `internal/gateway/translate`에 구현했다.
단위 시험·race·vet·현재 파일 gopls 진단은 아래 범위에서 통과했다. 실제 Claude/GPT 호환성이나 전체 목표 완료는 아니다.

API는 `Request(model string, body []byte, Limits) ([]byte, *ResponseContext, error)`,
`ResponseContext.Response([]byte)`, `ResponseContext.Stream(context.Context, io.ReadCloser, io.Writer)`다.
선택된 canonical model과 요청별 도구 역매핑을 사용한다. Stream은 upstream body를 소유하고 모든 종료 경로에서 닫는다.
HTTP 상태 검사는 호출 전 adapter 책임이다.

- 최상위 system은 두 줄바꿈으로 연결하며 메시지 안의 system 위치·공개 문자열·도구 pair ID를 보존한다.
- optional schema를 유지하고 strict false와 max_output_tokens를 설정한다. 안전한 이름을 먼저 예약한 뒤 충돌 없는 별칭을 만든다.
- item ID와 call ID가 다른 두 도구의 교차 SSE를 분리한다. 최종 text/arguments는 delta와 일치해야 하며 done 내용을 재전송하지 않는다.
- 모든 공개 블록이 정상 종료해야 end_turn/tool_use/max_tokens를 내보낸다. EOF·오류·중복·잘못된 ID·혼합 미완료 출력에는 성공 terminal이 없다.
- 본문·SSE·전체 출력 바이트 한도, JSON 중복 key·정수 타입·깊이·invalid UTF-8를 검사한다. 취소 시 upstream을 닫으며 재시도는 없다.

## Evidence

명령의 작업 디렉터리는 아래 WT다. 요청 변환 최초 RED는 구현 전에 실행했다.

```text
$ go test ./internal/gateway/translate -count=1
# github.com/modu-ai/moai-adk/internal/gateway/translate [github.com/modu-ai/moai-adk/internal/gateway/translate.test]
internal/gateway/translate/request_test.go:13:55: undefined: ResponseContext
internal/gateway/translate/request_test.go:13:96: undefined: Request
internal/gateway/translate/request_test.go:13:128: undefined: Limits
internal/gateway/translate/request_test.go:38:59: undefined: Request
internal/gateway/translate/request_test.go:38:91: undefined: Limits
internal/gateway/translate/request_test.go:39:32: undefined: Limits
internal/gateway/translate/request_test.go:39:120: undefined: Request
internal/gateway/translate/request_test.go:43:11: undefined: Request
internal/gateway/translate/request_test.go:43:35: undefined: Limits
internal/gateway/translate/request_test.go:45:11: undefined: Request
internal/gateway/translate/request_test.go:45:11: too many errors
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate [build failed]
FAIL
```

Exit 1. 요청 구현 후 첫 GREEN:

```text
$ go test ./internal/gateway/translate -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.400s
```

응답·SSE 구현 전 RED:

```text
$ go test ./internal/gateway/translate -count=1
# github.com/modu-ai/moai-adk/internal/gateway/translate [github.com/modu-ai/moai-adk/internal/gateway/translate.test]
internal/gateway/translate/response_test.go:50:13: c.Response undefined (type *ResponseContext has no field or method Response)
internal/gateway/translate/response_test.go:73:16: c.Response undefined (type *ResponseContext has no field or method Response)
internal/gateway/translate/response_test.go:92:12: c.Stream undefined (type *ResponseContext has no field or method Stream)
internal/gateway/translate/response_test.go:126:13: c.Stream undefined (type *ResponseContext has no field or method Stream)
internal/gateway/translate/response_test.go:147:11: c.Stream undefined (type *ResponseContext has no field or method Stream)
internal/gateway/translate/response_test.go:158:12: c.Stream undefined (type *ResponseContext has no field or method Stream)
internal/gateway/translate/response_test.go:165:24: c.Stream undefined (type *ResponseContext has no field or method Stream)
internal/gateway/translate/response_test.go:182:12: c.Stream undefined (type *ResponseContext has no field or method Stream)
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate [build failed]
FAIL
```

Exit 1. 응답·SSE 구현 후 첫 GREEN:

```text
$ go test ./internal/gateway/translate -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.398s
```

UTF-8 무손실 경계의 실제 RED:

```text
$ go test ./internal/gateway/translate -run TestInvalidUTF8NeverReplacesPublicText -count=1
--- FAIL: TestInvalidUTF8NeverReplacesPublicText (0.00s)
    request_test.go:165: invalid UTF8 became {"input":[{"content":[{"text":"�","type":"input_text"}],"role":"user"}],"max_output_tokens":1,"model":"gpt-5.6-sol","store":false}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway/translate	0.332s
FAIL
```

Exit 1. utf8.Valid 거절과 경계 음성군 보강 후 최종 측정:

```text
$ go test ./internal/gateway/translate -count=1 -coverprofile=/tmp/gateway-translate-cover.out
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.732s	coverage: 92.7% of statements
$ go test -race ./internal/gateway/translate -count=1
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	1.468s
$ go vet ./internal/gateway/translate
$ gopls check internal/gateway/translate/request.go internal/gateway/translate/response.go internal/gateway/translate/stream.go
```

각 exit 0. 마지막 두 명령의 stdout/stderr는 비었다. 독립 검사 세 개를 한 번에 시작했다.
최초 sandbox vet/gopls에는 캐시 operation not permitted 오류가 있어 승인 후 같은 명령을 재실행했다.
최초 gopls는 오류를 출력하고도 exit 0이었으므로 그 실행은 PASS가 아니다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`
- `git rev-parse --short HEAD` → `81c1d58f9`; `git branch --show-current` → `WT-unified-gateway`, 각 exit 0.
- 기준: 코어 design §4.2·4.3, AC-MG-004·007·008·009·010, m5-design-decisions 및 m5-picker-plan-audit-iter1의 일반 변환 착수 허용 범위.
- 소유한 새 변환 패키지와 이 보고서만 작성했다. CLI·root gateway·SPEC·다른 작업자의 파일은 변경하지 않았다.
- 최종 제품 SHA-256:
  - request.go: `1fbcb3b65a2087a4fa0c96fe1402151bab97bd733271990f8b28ce1635e3f5be`
  - response.go: `c3b63250191182ab5bce876d869249b349904f60250729a60afac469e4fe84e0`
  - stream.go: `fa09aa2991e032b48b85fddf969f208fcf87426d6a944574e9898f49dbb6a78f`

## Gaps

- raw request-005.json은 thinking/context_management/output_config 때문에 명시 거절된다. 공개 golden은 시험에서 이 세 필드를 명시적으로 제외한 투영본이다. 메시지 role·위치·text는 원본과 비교했지만 raw 호환성 PASS는 아니다.
- reasoning·redacted_thinking·encrypted reasoning은 모두 거절한다. provenance 없는 데이터를 다른 provider 소유라 간주하여 제거하지 않았다. PROBE ONLY carrier를 구현·활성화하지 않았다.
- thinking/context_management/output_config, 이미지/PDF, 비어 있지 않은 annotations/logprobs, stop_sequences, tool_result.is_error true는 명시 미지원이다. 알려지지 않은 요청·공개 content 필드도 거절한다. ephemeral cache hint만 문자열/schema를 보존하고 Responses cache 설정에 전달하지 않는 정책을 명시했다.
- InputTokens 시험은 합성 caller 측정값이며 제품 tokenizer 연결이 없다. 양의 ContextTokens에 측정값이 없으면 오류지만 기본 Limits만으로 context 길이를 판정하지 않는다.
- HTTP adapter·429 상태 처리·실 upstream·Claude 실행·reasoning 재개를 시험하지 않았다. 전체 AC-MG-004·009·010과 GPT 제품 활성화는 미완료다.
- 첫 구현 전 LSP baseline snapshot은 남기지 못했다. 현재 세 제품 파일의 진단만 관측했다.
- 통합 브랜치 CI가 저장소 전체 시험 판정의 담당이며 PENDING이다. 전체 suite·Windows 실행·commit·push를 하지 않았다.

## Residual-risk

합성 SSE와 공개 투영 fixture 통과는 실제 provider 이벤트 조합·필드 수용을 보장하지 않는다. 현재 지원하지 않는 정책과 output item은 명시 오류가 되므로 호환성 확대에는 별도 계약과 실측이 필요하다. HTTP 연결의 context 전파와 body 소유권도 adapter에서 맞춰야 한다. 전체 GPT 목표를 일반 변환 단위 검증으로 대신할 수 없다.

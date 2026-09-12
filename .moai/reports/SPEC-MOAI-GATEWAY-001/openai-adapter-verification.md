# 일반 OpenAI adapter 로컬 검증

## Claim

공개 일반 변환기를 OpenAI 고정 endpoint와 연결하는 `OpenAIAdapter`를 구현했다. API 키와 구독을 분리하고,
지원하지 않는 client 정책·reasoning·측정되지 않은 context를 전송 전에 거절한다. 로컬 TLS 통합 시험·race·vet·현재 gopls 진단은
아래 범위에서 통과했다. 실제 Claude Code raw 요청이나 실제 GPT endpoint의 수용, 전체 목표 완료는 아니다.

- `OpenAIConfig`에는 명시 Transport, Limits, 로컬 MeasureInput, 필요 시 Subscription Store와 SendOptions를 준다.
  생성 시 transport를 clone하고 proxy·검증을 끈 TLS를 거절한다. implicit 환경 인증/프록시·fallback·redirect가 없다.
- API 키는 `https://api.openai.com/v1/responses`에 CredentialRef.Apply 후 목적지와 Generation을 다시 확인하여 보낸다.
- 구독은 `https://chatgpt.com/backend-api/codex/responses`에 `Store.Owns(ref)`를 필수 확인한 뒤 `Store.SendAuthorized`로 보낸다.
  Owns는 부모가 auth 패키지에 추가한 provenance seam이다. adapter 작성자는 auth 파일을 수정하지 않았다.
- 변환/정책 오류는 Anthropic 형태의 400 응답이다. 기존 Server를 통과해 502로 바뀌지 않는 로컬 HTTP 판정을 했다.
  인증 오류는 401, upstream 401/429 등은 상태와 안전한 오류 구조로 반환한다. 검증된 Retry-After만 보존한다.
  원본 오류 body, credential, 임의 Request-Id/Location은 반영하지 않는다.
- 스트림은 io.Pipe로 중계한다. Body.Close는 context를 취소하고 reader를 닫은 뒤 converter 종료를 기다린다.
  converter가 upstream을 닫으며 잘린 EOF에는 성공 terminal이 없다. HTTP 200 nonstream도 크기와 JSON 변환을 확인한다.

## Evidence

이 문서의 모든 명령은 지정 WT에서 실행했다. 최초 adapter 시험을 구현 전에 실행한 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run TestOpenAI -count=1
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/openai_test.go:47:36: undefined: OpenAIConfig
internal/gateway/openai_test.go:48:9: undefined: OpenAIConfig
internal/gateway/openai_test.go:90:10: undefined: NewOpenAIAdapter
internal/gateway/openai_test.go:137:11: undefined: NewOpenAIAdapter
internal/gateway/openai_test.go:160:12: undefined: NewOpenAIAdapter
internal/gateway/openai_test.go:216:10: undefined: NewOpenAIAdapter
internal/gateway/openai_test.go:242:11: undefined: NewOpenAIAdapter
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

Exit 1. 구현과 로컬 fixture 보정 후 첫 GREEN:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run TestOpenAI -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.281s
```

다른 Store가 동일 Generation을 가진 경우의 후속 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run TestOpenAI -count=1 -timeout 30s
--- FAIL: TestOpenAISubscriptionRejectsDifferentStore (0.03s)
    openai_test.go:286: foreign store accepted: status=200 calls=1
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.554s
FAIL
```

Exit 1. Store.Owns seam 연결 뒤 GREEN:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run TestOpenAI -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.476s
```

최종 검증:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway -run TestOpenAI -count=1 -timeout 30s -coverprofile=/tmp/gateway-openai-adapter-cover.out
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.691s	coverage: 41.4% of statements
$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway ./internal/gateway/translate
$ GOCACHE=/tmp/gateway-translation-cache gopls check internal/gateway/openai.go internal/gateway/translate/request.go
```

각 exit 0. vet/gopls stdout·stderr는 비었다. coverage 파일에서 openai.go의 statement 수와 실행 count를 합산한 출력:

```text
openai.go coverage 127/140 = 90.7%
```

41.4%는 OpenAI 시험만 선택한 root gateway 전체 수치이며 새 adapter의 90.7%와 구분한다. root의 다른 기능 시험을 수행했다는 주장은 아니다.
독립 race·vet·gopls·translate race를 한 묶음으로 시작했다. vet가 시험의 context cancel cleanup 누락을 지적하여 defer cancel을 넣고
vet와 해당 race 시험을 다시 실행했다. 첫 TLS fixture는 요청 body를 소진하지 않아 서버가 연결 종료를 기다리는 상태로 timeout했다.
fixture에서 body 소진 및 bounded handler cleanup을 추가했다. 해당 timeout을 제품 성공으로 세지 않았다.
Sandbox listener 금지로 실패한 실행은 승인받은 loopback TLS 실행으로 재측정했으며, 외부 공급자 요청은 없었다.

## Baseline-attribution

- WT: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, 재측정 HEAD `81c1d58f9`.
- 소유 파일: openai.go/openai_test.go, translate/request.go/request_test.go의 Unicode 수리, 이 보고서와 translation-parent-delta.md.
  server.go·CLI·auth·supervisor·SPEC·Git 이력은 변경하지 않았다.
- SHA-256 openai.go: `adbb0129a60d9528dd1a0304c9df06ccdd39ae87d24bbd403335412f56687b7b`.
- SHA-256 openai_test.go: `e535be4e85e3833d221a93626a668d156a801f79ca40cad4c7b9b2e9036f5a45`.
- Unicode 수리의 별도 실제 RED/GREEN·부모 probe 재실행·현재 해시는 translation-parent-delta.md에 있다.

## Gaps

- 모든 account/token·측정값·TLS 서버는 합성 fixture다. 실제 로그인·refresh·Claude·GPT API 수용·구독 과금 경로를 실행하지 않았다.
- 원본 Claude turn의 thinking/context_management/output_config와 opaque reasoning은 여전히 명시 미지원이다.
  공개 ordinary projection 성공을 실제 raw 호환성으로 간주하지 않는다. 제품 경로 활성화 전 별도 정책·reasoning 게이트가 필요하다.
- 양의 catalog ContextTokens와 caller 측정을 필수로 검사하지만 실제 tokenizer/측정 구현은 연결하지 않았다. 제품 catalog가 capability를
  확정하지 않은 항목은 거절된다. 이미지/PDF 미지원은 그대로이며 T16 전체 통과가 아니다.
- 구독 HTTP SendAuthorized의 header write·logout watcher 상세 계약은 auth 패키지 소유다. 이 시험은 실제 Store API와 합성 TLS 경계,
  configured Store provenance와 logout 전송 금지를 판정했다. 실제 공급자 refresh 수용은 미검증이다.
- 현재 생산 파일의 LSP 진단을 관측했으며 변경 전 별도 LSP snapshot은 없다. 전체 suite·Windows·통합 CI는 실행하지 않았다.
  저장소 전체 verdict는 통합 브랜치 CI 담당이며 PENDING이다. Commit·push·PR·merge를 하지 않았다.

## Residual-risk

호출자는 Body를 끝까지 읽거나 반드시 Close하여 pipe converter를 회수해야 한다. 실제 provider의 필드·이벤트 조합은 일반 변환의
엄격한 지원 범위 밖에서 오류가 될 수 있다. Transport의 사용자 지정 dial hook은 테스트/신뢰된 구성 표면이며 요청자가 선택할 수 없다.
이 adapter의 국소 성공은 전체 GPT 구현·실제 Claude 왕복 완료를 대체하지 않는다.

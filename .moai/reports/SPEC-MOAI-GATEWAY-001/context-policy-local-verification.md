# 문맥 추정 및 공개 API truncation 경계 — 로컬 검증

## Claim

`EstimateInputTokens(ModelEntry, []byte) (int64, error)`를 추가했다. `OpenAIConfig.MeasureInput`에 그대로 전달할 수 있다. messages/system/tools만 투영하여 json.Marshal한 뒤 `(utf8.RuneCount + 3) / 4`를 계산하는 기존 count_tokens 알고리즘을 공통 helper로 옮겼다. count_tokens의 `accuracy: estimate`, `algorithm: json-runes-div4` 표시는 유지한다. 추정값은 정확한 tokenizer 결과도, 보장된 상한도 아니다.

callback은 기존 translate.ValidateJSONObject로 중복 키·잘못된 UTF-8·고립 surrogate·비객체 JSON 등을 거절한다. 기존 바이트 제한과 번역 검증을 완화하지 않았다. factory에 callback을 연결하거나 ContextTokens 기본값을 만들지 않았다.

공개 OpenAI API의 AuthAPIKey 요청에만 `truncation: disabled`를 명시했다. 구독 endpoint에는 이 필드를 추가하지 않았다. 입력 배열과 system 내용은 보존하며, 추정값 초과는 기존 로컬 400 판정으로 외부 호출 전에 거절한다. 실제 제공자가 반환한 400도 본문에 담긴 비공개 메시지를 노출하지 않고 전달한다.

## Evidence

작업 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.

새 callback 시험을 먼저 추가한 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run '^TestEstimateInput' -count=1
# github.com/modu-ai/moai-adk/internal/gateway [github.com/modu-ai/moai-adk/internal/gateway.test]
internal/gateway/input_estimate_test.go:15:13: undefined: EstimateInputTokens
internal/gateway/input_estimate_test.go:26:11: undefined: EstimateInputTokens
FAIL	github.com/modu-ai/moai-adk/internal/gateway [build failed]
FAIL
```

첫 구현 뒤 한글 fixture의 수기 기대값 11이 실제 기존 계산식 결과 10과 달랐다. `{messages:[], system:"가나", tools:[]}`의 JSON 문자 수 계산을 확인하여 기대값을 10으로 바로잡았다. 이 결과는 제품 결함의 RED로 계산하지 않는다.

공개 API 필드 assertion을 먼저 추가한 별도 RED:

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run '^TestEstimateInput|^TestOpenAIAPIKeyEndpointAndPublicResponse$' -count=1
--- FAIL: TestOpenAIAPIKeyEndpointAndPublicResponse (0.00s)
    openai_test.go:84: translation {"input":[{"content":[{"text":"hello","type":"input_text"}],"role":"user"}],"max_output_tokens":10,"model":"gpt-5.6-sol","store":false}
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/gateway	0.400s
FAIL
```

구현 후 GREEN 및 최종 범위 검증(작성자 실행, 별도 감사자의 판정 아님):

```text
$ GOCACHE=/tmp/gateway-translation-cache go test ./internal/gateway -run '^TestEstimateInput|^TestOpenAI' -count=1 -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway	0.530s

$ GOCACHE=/tmp/gateway-translation-cache go test -race ./internal/gateway -run '^TestEstimateInput|^TestOpenAI|^TestModelsAndCountAreLocalAndExplicit$' -count=1 -coverprofile=/tmp/gateway-context-policy-cover.out -timeout 30s
ok  	github.com/modu-ai/moai-adk/internal/gateway	1.686s	coverage: 30.4% of statements

$ GOCACHE=/tmp/gateway-translation-cache go vet ./internal/gateway
[exit 0, stdout/stderr 비어 있음]

$ GOCACHE=/tmp/gateway-translation-cache go tool cover -func=/tmp/gateway-context-policy-cover.out | rg 'input_estimate|writeCount|openai.go.*Send'
github.com/modu-ai/moai-adk/internal/gateway/input_estimate.go:13:	EstimateInputTokens		83.3%
github.com/modu-ai/moai-adk/internal/gateway/input_estimate.go:24:	estimateInputProjection		87.5%
github.com/modu-ai/moai-adk/internal/gateway/openai.go:51:		Send				87.4%
github.com/modu-ai/moai-adk/internal/gateway/server.go:176:		writeCount			60.0%
```

검증한 조건:

- 기본·빈 배열·한글 system/tools 입력에서 기존 count 응답과 callback 추정 결과 및 labels가 동일하다.
- callback은 중복 messages, 고립 surrogate, 비객체 JSON, 잘못된 UTF-8을 거절한다.
- API-key 요청 body에 truncation disabled가 있으며 구독 SendAuthorized 요청 body에는 해당 필드가 없다.
- 긴 합성 입력의 추정값보다 ContextTokens를 1 작게 두면 HTTP 400, TLS dial 0건이다.
- ContextTokens를 추정값과 같게 두면 세 메시지의 첫·중간·마지막 텍스트 및 system을 모두 순서대로 전송한다. 합성 upstream의 400은 400으로 돌아오고 비공개 오류 문자열은 제거된다. 전송은 1건이며 자동 재시도나 입력 삭제가 없다.

공개 API 계약은 이 실행에서 [OpenAI 공식 Python 타입의 truncation 설명](https://raw.githubusercontent.com/openai/openai-python/main/src/openai/types/responses/response_create_params.py) 288–294행을 web open/find로 직접 읽었다. disabled는 기본 동작이며 입력이 문맥 크기를 넘으면 400으로 실패한다고 명시한다. 이 문서는 별도 구독 endpoint의 필드 수용을 증명하지 않는다.

## Baseline-attribution

```text
$ git rev-parse --short HEAD
81c1d58f9
$ git branch --show-current
WT-unified-gateway
$ shasum -a 256 internal/gateway/input_estimate.go internal/gateway/input_estimate_test.go internal/gateway/server.go internal/gateway/openai.go internal/gateway/openai_test.go
2ffb9839b2a70c04eee82cad5fa3ee707965b844dcc255c93e0adf18e7c4872b  internal/gateway/input_estimate.go
2bed0bab379a6ff18d7733c65c72b73e7ee5a0045326bcbb899c5643437303dc  internal/gateway/input_estimate_test.go
a3fc6fb4ad9f00b91b06520cc18cc2ac6026722ab046e84d506ce4e79e6f2087  internal/gateway/server.go
c70572ab97cd4bf41aaffa1fd7a1cc9f8b79509b227066d5dc01b347a2c23c96  internal/gateway/openai.go
77167bd7b93171c0c491990c0aa98228631c40ef53559bb6fa86706af861fe0b  internal/gateway/openai_test.go
```

공유 worktree의 위 다섯 파일만 수정했다. AUTH live harness·자격 저장소 구현·factory·CLI·SPEC 본문은 수정하지 않았다. 모든 HTTP 검증은 합성 자격과 로컬 TLS 서버로 수행했다. 실제 Claude/Codex/broker/제공자는 실행하지 않았다. git commit/push도 하지 않았다.

## Gaps

정확한 provider tokenizer와의 오차, 실제 문맥 초과 400, 구독 endpoint의 truncation 수용은 관찰하지 않았다. API-only 필드는 공식 문서와 로컬 전송 body로만 검증했다. factory 연결과 모델 ContextTokens 설정은 부모의 후속 범위다. 실제 Claude의 thinking/output_config/context_management 입력 지원은 이 변경에 포함되지 않는다.

coverage 30.4%는 이번 선택 시험에서 root gateway 전체 파일을 분모로 계산한 값이며 저장소 전체 품질 판정이 아니다. writeCount의 새 marshal 실패 분기와 callback의 엄격 검증 후 Unmarshal 실패 방어 분기는 이번 fixture에서 실행되지 않았다. LSP 전후 baseline은 확보하지 않았고 go vet만 실행했다. develop 통합 브랜치 CI가 저장소 전체 시험 판정의 소유자이며 이 보고 시점 PENDING이다.

## Residual-risk

추정값이 실제 토큰 수보다 작을 수도 크기도 하다. 따라서 이 callback을 정확한 문맥 보장이나 보수적 상한으로 사용했다고 주장하면 안 된다. 공개 API는 입력을 그대로 보내고 truncation disabled 및 제공자 오류에 최종 판정을 맡긴다. 구독 경로의 문맥·정책 허용 여부는 별도 실측 게이트를 유지한다. 루트 count endpoint의 기존 ingress JSON helper와 callback의 translate strict helper는 서로 다르며 이번 변경은 그 파서를 통합하지 않았다.

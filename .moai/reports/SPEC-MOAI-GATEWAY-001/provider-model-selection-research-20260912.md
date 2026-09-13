# Claude Code에서 제공자별 /model 선택과 GPT-6 사용을 위한 조사

조사일: 2026-09-12. 대상: moai-proxy-unified / WT-unified-gateway / 81c1d58f9 및 미커밋 작업 트리. Claude Code 2.1.268. 조사·임시 UI 실험만 수행했으며 제품 코드는 수정하지 않았다.

## 결론

제공자별 /model 목록 구성과 선택은 설치된 Claude Code에서 실제로 가능했다. 임시 설정으로 GPT 전용, GLM 전용, Claude 전용 목록을 각각 띄우고 다른 행으로 선택을 바꿨다. GPT 목록에서 Sol → GPT-6 Astra 선택 성공까지 확인했다. 실제 GPT 서버 생성 및 MoAI launcher 통합 성공은 이번 실험의 범위가 아니다.

공식 GPT-6 식별자는 gpt-6-astra다. 화면 표시는 GPT-6 Astra로 하고, 실제 요청에는 이 ID를 사용해야 한다. 현재 MoAI catalog도 gpt-6-astra를 등록하며 gpt-6는 거절하는 시험을 갖고 있다. 개인 계정의 구독 접근 권한은 모델 문서만으로 확인할 수 없다.

[OpenAI GPT-6 Astra 모델 문서](https://developers.openai.com/api/docs/models/gpt-6-astra)

## 목표 동작과 권장 구성

- moai cc: Claude 목록, Claude 기본값과 fallback, Anthropic 요청만 허용.
- moai glm: 계정에서 사용할 수 있는 GLM 목록, GLM 기본값과 보조 모델, Z.AI 요청만 허용.
- moai gpt: 계정에서 사용할 수 있는 GPT 목록, GPT 기본값과 보조 모델, OpenAI 요청만 허용. GPT-6 Astra를 선택 가능하게 제공.
- /model의 Default 행도 같은 제공자로 해석되어야 한다. 시작 모델과 재개 모델 역시 같은 제공자여야 한다.
- 모델 목록을 숨기는 것과 요청을 차단하는 것은 서로 다른 기능이다. gateway의 세션별 exact catalog와 제공자 검증으로 직접 입력, fallback, 서브에이전트 경로도 제한해야 한다.

권장 구현은 현재 MoAI gateway를 유지하면서 launcher가 프로세스별 --settings와 child env를 생성하는 것이다. 전역 사용자 설정이나 프로젝트 settings.local.json을 실행 때마다 덮어쓰지 않는다. 프로필별 저장된 model·discovery cache·resume 경계를 격리하되 기존 Claude 구독 인증과 프로젝트 설정을 보존하는 방식은 별도 검증한다.

## 공식 Claude Code 기능에서 확인한 점

modelPicker는 2.1.242 이상에서 지원한다. options에 model·label·description을 지정하고 replaceBuiltInOptions:true를 주면 사용자 지정 목록으로 바꿀 수 있다. managed, --settings, user 설정에서만 읽으며 프로젝트/local 설정에서는 무시한다. 우선순위가 높은 한 목록을 채택한다.

Default와 현재 모델 행은 남는다. 유효한 행이 하나도 없으면 내장 목록으로 돌아갈 수 있다. 따라서 해당 제공자의 명시적 시작 모델을 주고, opus·sonnet·haiku·fable 슬롯과 fallback을 동일 제공자에 고정해야 한다. 조직 managed 설정이 있으면 그것이 --settings보다 우선하므로 우회 대신 충돌을 알려야 한다.

availableModels는 추가 방어로 쓰되 일반 설정의 배열 병합과 Default 예외를 고려해야 한다. enforceAvailableModels는 managed 정책의 기능이며 비어 있거나 유효 모델이 없는 목록에도 예외가 있다. 이것만으로 gateway 제공자 경계가 완성되지는 않는다.

[modelPicker 공식 설정](https://code.claude.com/docs/en/settings-reference#modelpicker)
[설정 우선순위](https://code.claude.com/docs/en/settings#lists-merge-instead-of-overriding)
[모델 제한과 Default](https://code.claude.com/docs/en/model-config#restrict-model-selection)

다음은 UI 실험에서 사용한 것과 같은 구조의 GPT 프로필 예시다. 완성된 제품 설정이나 계정 권한 목록을 뜻하지 않는다. 별도로 --model gpt-6-astra 및 gateway 주소·세션 인증을 주어야 한다.

```json
{
  "model": "gpt-6-astra",
  "modelPicker": {
    "replaceBuiltInOptions": true,
    "options": [
      {"model": "gpt-6-astra", "label": "GPT-6 Astra"},
      {"model": "gpt-5.6-sol", "label": "GPT-5.6 Sol"}
    ]
  },
  "env": {
    "ANTHROPIC_DEFAULT_OPUS_MODEL": "gpt-6-astra",
    "ANTHROPIC_DEFAULT_SONNET_MODEL": "gpt-6-astra",
    "ANTHROPIC_DEFAULT_HAIKU_MODEL": "gpt-5.6-sol",
    "ANTHROPIC_DEFAULT_FABLE_MODEL": "gpt-6-astra"
  }
}
```

역할별 저비용 모델 배정은 추후 결정할 수 있다. 핵심은 모든 슬롯이 같은 제공자 안에 있어야 한다는 점이다. GPT 이름을 Claude 모델로 위장하거나 behavesAs를 무조건 붙이면 잘못된 능력·context·요금 표시를 유도할 수 있다. 검증된 기능만 명시해야 한다.

자동 모델 발견만으로는 GPT/GLM 목록을 만들 수 없다. 공식 discovery는 ID에 claude 또는 anthropic이 포함된 항목만 받는다. 또한 캐시가 남고, replaceBuiltInOptions를 쓰면 발견된 행은 숨긴다. 이번 목표에는 modelPicker가 직접적인 수단이다. ANTHROPIC_CUSTOM_MODEL_OPTION은 한 항목 추가 기능이므로 여러 GPT 모델만 표시하는 수단으로는 부족하다.

[Gateway discovery 필터·캐시](https://code.claude.com/docs/en/llm-gateway-protocol#model-discovery)

Anthropic은 gateway 사용 자체를 문서화하지만 비-Claude 모델로 라우팅하는 것은 지원하지 않는다고 명시한다. 따라서 이는 Claude Code UI를 사용하는 MoAI 호환 계층이며 Anthropic이 GPT 실행을 지원·보증한다는 뜻은 아니다.

[Anthropic gateway 지원 범위](https://code.claude.com/docs/en/llm-gateway)

## reasoning 문제와 오픈소스 비교

reasoning의 암호문을 해독하거나 사용자에게 보여줄 필요는 없다. 응답 item의 ID·순서·암호문을 보존하여 다음 요청의 동일한 대화 위치에 되돌리는 것이 핵심이다. OpenAI 공식 문서는 stateless 응답에서 encrypted_content를 제공하고 전체 output 이력을 후속 입력으로 재사용하는 흐름을 설명한다. 요약 텍스트는 암호문과 별개다. 모델 family가 바뀌면 reasoning 호환성도 달라지므로 GPT 내 모델 전환도 별도 시험해야 한다.

[OpenAI reasoning 보존](https://developers.openai.com/api/docs/guides/reasoning)

### OpenAI Codex

commit 4dcce4f0c47e0183e4b05ffcbb38b4fffb8b8042의 client.rs는 store:false, stream:true, include:reasoning.encrypted_content를 사용한다. protocol/models.rs의 Reasoning item은 선택적인 id·content·encrypted_content와 summary를 가진다. 모든 output item에 message와 같은 status를 요구해서는 안 된다는 타입별 처리 근거다. 최신 API 문서에서 include가 선택적이어도 현행 Codex 클라이언트는 호환 요청에 유지한다.

[Codex client.rs](https://github.com/openai/codex/blob/4dcce4f0c47e0183e4b05ffcbb38b4fffb8b8042/codex-rs/core/src/client.rs#L855)
[Codex Reasoning 타입](https://github.com/openai/codex/blob/4dcce4f0c47e0183e4b05ffcbb38b4fffb8b8042/codex-rs/protocol/src/models.rs#L1035)

### CLIProxyAPI

commit 5b2785617d1e7de84a9f4dee599d275a4ccd8999의 Codex↔Claude 변환기는 암호문을 thinking.signature/signature_delta로 운반하고 다음 Claude 요청의 signature를 Responses reasoning item으로 되돌린다. response.output_item.done의 최종 암호문을 사용하며 초기 added 값에 의존하지 않는다. replay cache 시험에는 다음 요청의 암호문 비교가 있다. 이는 이미 존재하는 구현 패턴이며 이번 조사에서 해당 프로젝트 시험을 실행한 것은 아니다.

[CLIProxyAPI 응답 변환](https://github.com/router-for-me/CLIProxyAPI/blob/5b2785617d1e7de84a9f4dee599d275a4ccd8999/internal/translator/codex/claude/codex_claude_response.go#L113)
[CLIProxyAPI 요청 변환](https://github.com/router-for-me/CLIProxyAPI/blob/5b2785617d1e7de84a9f4dee599d275a4ccd8999/internal/translator/codex/claude/codex_claude_request.go#L160)
[암호문 재전송 시험 소스](https://github.com/router-for-me/CLIProxyAPI/blob/5b2785617d1e7de84a9f4dee599d275a4ccd8999/internal/runtime/executor/codex_executor_reasoning_replay_cache_test.go#L105)

### LiteLLM

commit 95b438013ad839f4c39274762bc20da7962caf88의 Responses adapter는 요약이 있으면 thinking.signature, 요약 없이 암호문만 있으면 redacted_thinking.data로 보존한다. 현재 MoAI의 opaque·receipt 설계와 비교할 직접적인 사례다. OpenAI 암호문은 Anthropic이 서명한 thinking이라는 뜻이 아니며, MoAI 내부 봉투를 그대로 Anthropic upstream으로 전달해서는 안 된다.

[LiteLLM reasoning block 변환](https://github.com/BerriAI/litellm/blob/95b438013ad839f4c39274762bc20da7962caf88/litellm/llms/anthropic/experimental_pass_through/responses_adapters/transformation.py#L169)

### Claude Code Router

commit afba89fa3778c772574636f1bd9aaad9138b44b7의 환경 생성기는 main model 외에도 fable·opus·sonnet·haiku 슬롯을 배정한다. 제공자별 모델을 main 요청 외의 보조 요청까지 일관되게 전달하는 데 참고할 수 있다. 프로젝트 전체를 교체 도입할 필요가 있다는 결론은 아니다.

[CCR child 환경 생성](https://github.com/musistudio/claude-code-router/blob/afba89fa3778c772574636f1bd9aaad9138b44b7/packages/core/src/agents/claude-code/environment.ts#L55)

### Z.AI 공식 연동

Z.AI는 Claude Code에 Anthropic 호환 endpoint와 모델 슬롯을 설정하는 방법을 안내한다. GLM 경로는 Responses 변환기를 통과시킬 이유가 없다. 다만 UI용 GLM ID를 표시했다는 사실은 개인 구독의 해당 모델 접근 성공을 증명하지 않는다.

[Z.AI Claude Code 연동](https://docs.z.ai/devpack/tool/claude)

## MoAI에 필요한 변경

1. 시작 모드별 catalog를 고정한다. 기본 모델뿐 아니라 ExplicitModel, 재개 모델, fallback, 서브에이전트의 요청 모델도 해당 provider 소속인지 검사한다. 다른 provider ID는 전송 전에 거절한다.
2. launcher가 --settings로 provider 전용 modelPicker를 전달한다. Default·현재 모델·저장된 기본값·캐시·fallback·별칭·동시 실행을 함께 다룬다. GPT production binding에 이미 있는 GPT 전용 목록은 재사용한다.
3. reasoning item에 message 전용 status 검사를 적용하지 않도록 parser를 분리한다. summary, encrypted_content, output_item.done, terminal output을 보존한다. text, tool_use/tool_result, call_id, phase, usage와 순서를 유지한다.
4. 기존 opaque/receipt 모듈을 실제 HTTP 요청·응답에 연결한다. thinking.signature 또는 redacted_thinking.data의 왕복은 설치본으로 확인한다. provider·credential·conversation family에 결속시키고 resume/fork·모델 family 변경 정책을 명시한다. 암호문을 임의 해석하거나 유실을 성공으로 처리하지 않는다.
5. MCP deferred tool은 단순 제외만으로 끝내지 않는다. tool_reference 로딩을 구현하거나, 도구 검색을 끄는 지원 경로가 전체 도구를 실제로 노출하는지 검증한 후 기능 범위를 정한다.
6. GPT 인증은 승인된 Codex 로그인 broker 경로를 유지한다. OpenAI API key와 ChatGPT 구독은 별도 인증·과금 경로다. login status만으로 GPT-6 계정 권한을 확정하지 않는다.

[OpenAI 인증 경로 구분](https://learn.chatgpt.com/docs/auth)

## Evidence / Baseline-attribution

소스 조회 및 시험 기준선:

```text
moai session current
01a08e7b-6aa0-7361-ab7e-ea8da1f02228
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
claude --version
2.1.268 (Claude Code)
```

설치 실행 파일 SHA256: 06a96d5423f83770f120859f1c58e60d7252cc4c122aa13043b7e7cd716bc76a. 외부 소스는 위 commit으로 고정했다. 공식 문서의 오래된 검색 캐시와 현재 페이지가 달라 현재 페이지를 다시 열어 판단했다. 특히 과거 /model 저장 동작을 현재 동작으로 인용하지 않았다.

PTY UI 실험: 별도 CLAUDE_CONFIG_DIR, 가짜 로컬 인증, http://127.0.0.1:9, 비필수 트래픽 비활성, MCP 비활성, 임시 작업 폴더를 사용했다. 실제 모델 생성 요청은 보내지 않았다. 프로세스 그룹 종료를 finally에서 보장했다. UI 실험은 MoAI 명령 자체를 통과하지 않은 직접 Claude 실행이다.

```text
python3 /tmp/moai-model-research-20260912/picker_probe.py gpt
python3 /tmp/moai-model-research-20260912/picker_probe.py glm
python3 /tmp/moai-model-research-20260912/picker_probe.py cc
```

수집된 PTY 출력은 커서 이동을 제거하면 일부 공백이 붙는다. 아래는 그 출력의 해당 줄을 그대로 발췌했다.

```text
1.Default(recommended)Usethedefaultmodel(currentlygpt-6-astra[1m])
2.GPT-6AstraPROBECustommodel(gpt-6-astra)
❯3.GPTSolPROBE✔Custommodel(gpt-5.6-sol)
⎿SetmodeltoGPT-6AstraPROBEforthissessiononly
PROBE_PROCESS_CLEANED True
```

```text
1.Default(recommended)Usethedefaultmodel(currentlyglm-5.3[1m])
2.GLM5.3PROBECustommodel(glm-5.3)
❯3.GLMFlashPROBE✔Custommodel(glm-5.3-flash)
⎿SetmodeltoGLM5.3PROBEforthissessiononly
PROBE_PROCESS_CLEANED True
```

```text
2.ClaudeOpusPROBECustommodel(claude-opus-5)
❯3.ClaudeSonnetPROBE✔Custommodel(claude-sonnet-5)
⎿SetmodeltoClaudeOpusPROBEforthissessiononly
PROBE_PROCESS_CLEANED True
```

현재 MoAI의 범용 prepare 계약은 시작 모드와 실제 provider가 달라도 허용한다. 기존 시험을 실행하여 그 의미를 확인했다. GPT production catalog가 GPT만 등록하는 것과 구별해야 한다.

```text
go test ./internal/cli -run '^TestGatewayPreparationSeparatesModeFromInitialProvider$' -count=1 -v -timeout=60s
--- PASS: TestGatewayPreparationSeparatesModeFromInitialProvider/gptglm-medium (0.00s)
--- PASS: TestGatewayPreparationSeparatesModeFromInitialProvider/glmgpt-6-astra (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.946s
```

현재 reasoning 거절 계약도 실행해 확인했다. 이 PASS는 reasoning 지원 성공을 뜻하지 않는다.

```text
go test ./internal/gateway/translate -run '^TestNonStreamRejectsEveryIncompleteOrOpaqueShape$' -count=1 -v -timeout=60s
=== RUN   TestNonStreamRejectsEveryIncompleteOrOpaqueShape
--- PASS: TestNonStreamRejectsEveryIncompleteOrOpaqueShape (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/gateway/translate	0.336s
```

로컬 관련 위치: internal/cli/gateway_prepare_test.go:11, internal/cli/gateway_prepare.go:34, internal/cli/gateway_product_binding.go:129, internal/gateway/catalog.go:98, internal/gateway/translate/stream.go:425, internal/gateway/translate/response.go:144.

## Gaps / Residual-risk / 완료 기준

이번 조사에서 검증된 것은 문서·선정한 4개 오픈소스의 관련 소스·현재 MoAI 시험·세 제공자 프로필의 /model UI 선택이다. 인터넷의 모든 저장소나 각 프로젝트의 모든 코드를 전수 검증한 것은 아니다. 제3자 시험은 소스만 확인했으며 실행하지 않았다.

- 실제 GPT-6 첫 답변, 2턴 대화, 도구 호출 후 최종 답변은 미검증이다.
- reasoning이 있는 응답을 현재 Claude가 보존하고 후속 요청에 그대로 싣는 실행 증거는 아직 없다. LiteLLM/CLIProxyAPI의 방식을 MoAI에서 재현해야 한다.
- GPT/GLM unknown-model 경고가 실험에서 발생했다. context 상한·노출 기능·요금 라벨을 Claude 가정대로 두지 않아야 한다. UI의 API Usage Billing 표기는 실제 GPT 구독 과금 증명이 아니다.
- 임시 설정에서 Default에 [1m] suffix가 붙는 것도 관찰했다. gateway가 받는 model ID와 context 계약에서 이 suffix를 어떻게 처리할지 시험해야 한다.
- managed 설정 충돌, 사용자 전역 기본값 저장, 재시작·resume, 동시 cc/glm/gpt 실행, 설치본 교체·업그레이드는 추가 검증이 필요하다.

구현 완료는 moai gpt 실행 → /model에서 GPT-6 Astra 선택 → 실제 답변 → tool 결과 후 재답변 → 종료/resume을 모두 통과하고, cc/glm/gpt 세 세션의 다른 provider 요청이 모두 전송 전에 거절되는 것으로 정의한다. UI 표시 모델, gateway route ID, upstream request model, 응답 provider를 함께 관찰해야 한다. 모델에게 자신의 이름을 묻는 답변만으로 판정하지 않는다.

## 실행 증거 파일

[GPT 선택 화면](provider-model-selection-evidence-20260912/gpt-screen.txt) · [GLM 선택 화면](provider-model-selection-evidence-20260912/glm-screen.txt) · [Claude 선택 화면](provider-model-selection-evidence-20260912/cc-screen.txt) · [재현 스크립트](provider-model-selection-evidence-20260912/picker_probe.py) · [공식 modelPicker 본문](provider-model-selection-evidence-20260912/official-modelpicker.txt)

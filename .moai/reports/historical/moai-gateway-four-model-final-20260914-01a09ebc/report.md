# MoAI GPT Gateway + Factory + MCP 구현 기준선 반영 최종 재설계

> 문서 상태: 개발 설계 지시서
> 기준선: `develop@4056f69e1`
> 작성일: 2026-09-14
> 대상: `moai-adk-go`, Claude Code, MoAI Gateway, MoAI MCP
> 범위: 설계 확정과 개발 시 검증 계획. 이번 문서 작업에서는 제품 코드를 구현하거나 실행 검증하지 않는다.

## 0. 최종 결론

MoAI의 올바른 다중 모델 구조는 역할마다 모델을 박아 넣는 표가 아니다. 핵심은 다음 세 층을 분리하는 것이다.

1. **메인 세션 층**: `moai cc`, `moai gpt`, `moai glm`으로 연 세션 하나는 선택한 백엔드·메인 모델·effort 하나를 가진다.
2. **Factory 층**: lead가 카드를 lane에 배정하고, lane 하나가 해당 카드를 `plan → run → sync`까지 끝까지 소유한다. lane의 메인 모델은 세션을 시작할 때 고르고 카드 수행 중에는 바뀌지 않는다.
3. **MCP 전문 호출 층**: 현재 세션의 정체성을 바꾸지 않고, 다른 회사 모델이나 별도 GPT 모델을 제한된 작업 단위로 호출한다.

따라서 `lead/plan/run/routine/scan`마다 모델을 고정하는 이전 표는 폐기한다. 역할 이름을 모델 프리셋으로 만드는 설계도 채택하지 않는다. 모델은 **역할 이름**이 아니라 **세션 또는 lane 시작 시점의 작업 난도·비용·용량**에 따라 선택한다.

`moai gpt`의 목표 구조는 다음과 같다.

- Claude Code UI와 도구 체계를 그대로 사용한다.
- 메인 대화와 native `Agent()`가 선택된 GPT 모델을 자연스럽게 상속한다.
- 인증과 대화 수명주기는 공식 Codex App Server를 권한 원천으로 삼는다.
- Anthropic 형식 변환, 오류 정규화, 관측성, fail-closed 정책은 MoAI Gateway가 담당한다.
- GPT 모델은 정확히 `gpt-5.6-luna`, `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-6-astra`만 허용한다.
- 이미지는 별도 `codex_image_generate` 도구로 분리하고 `codex-images-subscription` 경로의 정확한 `gpt-image-2.5`가 실제로 확인되기 전에는 기본 비활성화한다.

이 방향은 `claude-code-proxy`의 좋은 변환·진단 아이디어는 흡수하되, 자체 OAuth 구현과 비공개 ChatGPT 엔드포인트 직접 호출을 생산 경로로 복제하지 않는 설계다.

---

## 1. 제품 동작의 단일 진실

### 1.1 일반 세션

```text
moai cc  --model <claude> --effort <...>
  └─ Claude 메인 세션 1개
      └─ native Agent()는 해당 세션 정책을 상속

moai gpt --model <4개 중 하나> --effort <...>
  └─ GPT 메인 세션 1개
      └─ native Agent()는 같은 GPT 모델을 inherit

moai glm --model <configured GLM> ...
  └─ GLM 메인 세션 1개
      └─ native Agent()는 같은 GLM 세션 정책을 상속
```

`moai gpt` 안에서 native sub-agent마다 서로 다른 GPT 모델을 임의 지정하는 구조가 아니다. 현재 UI 정합성 수정도 GPT gateway에서 sub-agent 모델 셀을 `inherit`로 고정하고 effort만 편집하게 만든 방향과 일치한다.

다른 모델이 필요하면 둘 중 하나를 택한다.

- 오래 지속되는 독립 작업: 새 메인 세션 또는 새 Factory lane을 해당 모델로 시작한다.
- 짧고 경계가 분명한 전문 작업: 현재 세션에서 `codex_task` 또는 `glm_task`를 호출한다.

### 1.2 Factory

```text
Factory lead
  ├─ lane-1: 카드 A 전체(plan → run → sync), 메인 모델 X
  ├─ lane-2: 카드 B 전체(plan → run → sync), 메인 모델 Y
  └─ lane-3: 카드 C 전체(plan → run → sync), 메인 모델 Z
```

Factory는 한 카드의 단계를 서로 다른 모델에게 릴레이하는 파이프라인이 아니다. **lane이 카드의 전 생명주기를 소유**하고, lead는 분해·배정·상태 판단·최종 조율을 맡는다. lane 번호나 단계 이름에 모델을 영구 고정하지 않는다.

### 1.3 Kanban 폐기 결정

Kanban은 제품상 폐기된 것으로 확정한다. 앞으로 지원·설계·사용자 문서에서 `-k`를 정상 기능으로 설명하지 않는다.

다만 현재 `develop`에는 다음 흔적이 남아 있다.

- `internal/cli/cc.go`: `-k` 도움말과 dispatch 경로
- `internal/cli/factory.go`: `-k`/`-f` 병합 파싱과 상호 배타 처리
- `.claude/rules/moai/workflow/kanban-dispatch*.md`
- `.claude/agents/moai/manager-lead.md`의 Kanban 설명
- `.claude/skills/moai/workflows/factory.md`의 이전 세션 모델 설명

이것은 “Kanban이 아직 지원된다”는 뜻이 아니라, **별도 제거 작업이 필요한 레거시 GAP**이다. 이번 설계 문서는 이 잔존 코드를 삭제하지 않는다. 별도 카드에서 CLI·규칙·도움말·테스트·문서를 함께 제거하고, `-k` 호출이 명시적으로 지원 중단 오류를 내는지까지 검증해야 한다.

---

## 2. 현재 `develop`에서 확인된 구현 기준선

| 항목 | 현재 상태 | 근거 | 판정 |
|---|---|---|---|
| `moai gpt` 메인 launcher | GPT gateway를 거치는 전체 메인 세션으로 연결됨 | `internal/cli/gpt.go`, `internal/cli/root.go`, `internal/cli/gateway_prepare.go` | 구현됨 |
| GPT 허용 모델 | Astra, Sol, Terra, Luna 네 모델이 binding에 존재 | `internal/cli/gateway_product_binding.go` | 구현됨 |
| 기본 GPT 모델 | `gpt-5.6-sol` | `internal/cli/gateway_prepare.go` | 구현됨 |
| native sub-agent 모델 | GPT gateway에서는 `inherit`; effort 편집 가능 | t840 관련 merge가 현재 HEAD의 조상 | 구현됨 |
| GPT 인증 broker | 설치된 Codex의 `codex app-server --listen stdio://`를 시작해 인증 상태를 중개 | `internal/gateway/auth/broker.go` | 구현됨 |
| GPT 실제 turn 경로 | production binding은 아직 `OpenAIAdapter + AuthPKCE` | `internal/cli/gateway_factory.go`, `internal/cli/gateway_product_binding.go` | App Server 전환 미완료 |
| App Server adapter·authority | 관련 구현과 테스트 표면은 존재 | `internal/gateway` | 존재하지만 production wiring과 동일하지 않음 |
| Factory | lead + 번호 lane, lane이 카드 전체 생명주기 소유 | `internal/cli/factory.go`, `internal/cli/cc.go` | 구현됨 |
| Kanban | 제품상 폐기 결정과 달리 `-k` 코드·규칙이 잔존 | 위 1.3의 경로 | 레거시 GAP |
| MCP `codex_task` | prompt/background/write/resume_last 제공 | `internal/cli/mcp_server.go`, `internal/cli/codex_task.go` | V1 구현됨 |
| MCP `glm_task` | prompt/background/model/system/max_tokens 제공 | `internal/cli/mcp_server.go`, `internal/cli/glm_task.go` | V1 구현됨 |
| MCP 모델·effort 선택 | `codex_task` 공개 스키마에는 없음 | `internal/cli/mcp_server.go` | GAP |
| MCP job 내구성 | 메모리·프로세스 종속 | task/job 구현 | GAP |
| 정확한 이미지 모델 | `gpt-image-2.5`의 실제 subscription 경로 미검증 | 현재 코드와 공개 근거 | 차단 상태 |

### 2.1 “Rust test 10 + 4”의 의미

MoAI를 Rust로 다시 개발한다는 뜻이 아니다. `claude-code-proxy`가 Rust 프로젝트라서 그 저장소의 테스트가 참고 구현 자체를 검증했을 뿐이다. MoAI의 생산 구현은 기존 Go 경계인 `internal/cli`와 `internal/gateway`를 확장한다. 외부 Rust 테스트 통과는 MoAI 제품의 acceptance로 계산하지 않는다.

### 2.2 이미 develop에 반영된 카드와 남은 카드

현재 HEAD에서 다음 merge 계열은 조상 관계로 확인됐다.

- t654: gateway launcher 자동화 범위
- t840: GPT gateway sub-agent `inherit` 및 effort UI 정합성
- t841: effort allowlist 관련 범위

반면 다음은 현재 todo readback에서 후속 작업으로 남아 있다.

- queued — t842: 모델 카탈로그 SSOT
- queued — t843: web matrix/lane 표현 재설계
- queued — t844: 실세션 증거
- queued — t848: Astra `max` 재측정
- picked — t850/t851: receipt-chain 400 재현과 원인 규명
- picked — t853: ToS 조사 보고

따라서 launcher가 merge됐다는 사실과 실제 구독 계정·장시간 세션·이미지 경로의 운영 acceptance를 구분해야 한다.

---

## 3. `claude-code-proxy`에서 흡수할 것과 버릴 것

### 3.1 실제 연결 방식

해당 오픈소스는 Claude Code가 보내는 Anthropic Messages 요청을 받아 OpenAI/Codex 쪽 요청으로 변환한다. Codex provider는 자체 OAuth 자격 증명을 관리하고 `chatgpt.com/backend-api/codex/responses` 계열 내부 경로에 직접 연결한다. 모델 라우팅, 스트리밍 이벤트 변환, 토큰 추정, 세션 연속성, 진단 기능을 제공한다.

참고: [How it works](https://claude-code-proxy.raine.dev/how-it-works/), [Codex provider](https://claude-code-proxy.raine.dev/providers/codex/), [Compatibility and limitations](https://claude-code-proxy.raine.dev/reference/compatibility-and-limitations/), [GitHub repository](https://github.com/raine/claude-code-proxy)

### 3.2 흡수할 장점

| 아이디어 | MoAI 적용 |
|---|---|
| Anthropic ↔ provider 요청·스트림 변환 | typed adapter와 golden fixture로 유지 |
| provider별 모델 라우팅 | 정확한 4개 GPT allowlist + live capability 교집합 |
| 오류·종료 사유 정규화 | 안정적인 error code와 retryable 플래그 |
| 세션 연속성과 진단 | App Server thread/turn ID, receipt, redacted trace |
| 호환성 회귀 테스트 | Claude Code 입력 fixture → provider 이벤트 → Anthropic 응답 계약 테스트 |
| 이미지 기능 분리 | 텍스트 gateway와 별도 MCP tool로 격리 |

### 3.3 생산 경로에서 채택하지 않을 것

| 접근 | 이유 |
|---|---|
| proxy의 OAuth 구현 복제 | 계정·토큰 수명주기와 보안 책임이 MoAI로 이동 |
| 비공개 ChatGPT endpoint 직접 호출 | 공개 API 안정성 계약이 아니며 변경·계정 위험이 큼 |
| 무인증 loopback listener | 로컬 다른 프로세스가 요청을 주입할 수 있음 |
| 토큰 수 임의 추정으로 예산 강제 | 정확한 provider usage와 다를 수 있음 |
| `count_tokens` 성공을 실제 과금 증거로 해석 | 추정·호환 기능이지 billing receipt가 아님 |
| 이미지 모델 자동 대체 | 사용자 요구 모델·출처를 훼손하고 비용/품질 추적이 깨짐 |

### 3.4 ToS 판단

“구독 모델을 제3자 proxy로 쓰면 무조건 ToS 문제가 없다”는 결론은 낼 수 없다.

- OpenAI 약관은 계정 공유, 보호 조치·rate limit 우회, 자동·프로그램 방식의 추출, reverse engineering 등을 제한한다. [OpenAI Terms of Use](https://openai.com/policies/row-terms-of-use/)
- OpenAI는 Codex를 ChatGPT 플랜에서 공식 클라이언트로 제공하고, App Server에는 공식 초기화·계정·thread·turn 프로토콜을 문서화한다. [Using Codex with your ChatGPT plan](https://help.openai.com/en/articles/11369540-icodex-in-chatgpt), [Codex App Server](https://learn.chatgpt.com/docs/app-server)
- Anthropic은 LLM gateway 설정 자체는 문서화하지만, gateway를 통해 Claude Code를 비-Claude 모델로 라우팅하는 사용을 지원하지 않는다고 명시한다. [Claude Code LLM gateway](https://code.claude.com/docs/en/llm-gateway)
- `claude-code-proxy`도 unofficial client와 account risk, 무인증 listener 위험을 자체 문서에서 경고한다.

권장 위험 순서는 다음과 같다.

1. **목표 경로**: 사용자가 공식 Codex로 로그인하고 MoAI가 공식 App Server 프로토콜을 통해 인증·thread·turn을 위임한다.
2. **대안 경로**: 공식 OpenAI API key와 공개 API를 사용한다. 구독 요금과 별도 과금이라는 경제적 차이가 있다.
3. **현행 과도기 경로**: PKCE + 직접 adapter는 새 기능을 더 얹지 말고 App Server 전환 전까지 제한적으로 유지한다.
4. **비권장 경로**: 제3자 proxy의 자체 OAuth·내부 endpoint 호출을 그대로 fork한다.

이는 법률 자문이나 계정 무위험 보증이 아니다. 공식 App Server 경로가 상대적으로 가장 낮은 통합 위험을 가진다는 공학적 판단이다.

---

## 4. 목표 Gateway 구조

```text
Claude Code
  │ Anthropic-compatible Messages
  ▼
MoAI Gateway (Go)
  ├─ 요청 정규화 / 정확한 모델 allowlist
  ├─ Claude tool-use ↔ App Server item/event 변환
  ├─ 스트림 상태기계 / backpressure / cancellation
  ├─ 오류·usage·receipt 정규화
  ├─ loopback 인증 / redaction / 정책 검사
  └─ OpenAI App Server authority
       ├─ 공식 Codex 설치 확인
       ├─ account/read 또는 login
       ├─ initialize
       ├─ model/list
       ├─ thread/start | thread/resume
       └─ turn/start | turn/interrupt
```

### 4.1 모델 해석

정적 allowlist:

```yaml
provider: openai-codex
models:
  - gpt-5.6-luna
  - gpt-5.6-sol
  - gpt-5.6-terra
  - gpt-6-astra
```

실제 사용 가능 모델은 `정적 allowlist ∩ App Server model/list`다. 사용자가 요청한 모델이 교집합에 없으면 즉시 실패한다. 다른 모델로 조용히 바꾸지 않는다.

effort도 App Server의 `supportedReasoningEfforts`와 MoAI 안전 정책의 교집합으로 검증한다. 문서상 가능하다는 이유만으로 live 계정에서 확인되지 않은 effort를 승인하지 않는다.

### 4.2 인증과 보안

- 사용자의 공식 Codex 로그인을 재사용한다.
- refresh token이나 access token을 MoAI 설정·로그·MCP 결과에 저장하지 않는다.
- App Server subprocess와의 stdio/loopback 경계에 임의 호출 방지 nonce를 둔다.
- 매 turn에 sandbox와 approval policy를 명시한다.
- 요청·응답 로그는 prompt, source, base64, credential을 기본 redaction한다.
- retry는 동일 request ID의 효과가 없음을 확인한 뒤에만 수행한다.

### 4.3 안정성

- handshake와 capability discovery가 끝나기 전 Claude Code listener를 ready로 표시하지 않는다.
- streaming delta, tool call, completion, cancellation을 명시적 상태기계로 처리한다.
- 알 수 없는 event는 성공으로 삼키지 않고 `PROTOCOL_DRIFT`로 종료한다.
- receipt에는 requested/resolved model·effort, thread/turn, auth mode, usage source를 기록한다.
- context overflow, auth expiry, capacity, policy denial, protocol drift를 서로 다른 오류로 분류한다.

---

## 5. DeepSWE를 반영한 모델 선택 원칙

DeepSWE는 동일 harness의 장시간 소프트웨어 작업 113개, 91개 저장소, 5개 언어를 비교하는 공개 자료다. MoAI Factory 자체 benchmark도 아니고 ChatGPT 구독 요금표도 아니다. 여기서는 역할 고정표가 아니라 **세션 시작 시 모델을 고르는 참고 증거**로만 사용한다.

출처: [DeepSWE leaderboard](https://deepswe.datacurve.ai/), [live JSON](https://deepswe.datacurve.ai/artifacts/v1.1/leaderboard-live.json)

### 5.1 요청된 네 모델의 최고 pass 행

| MoAI 표시 모델 | DeepSWE effort | pass@1 | raw mean cost USD | 평균 output tokens | 평균 steps | 표본 n |
|---|---:|---:|---:|---:|---:|---:|
| `gpt-6-astra` | xhigh | 0.7412 | 6.52 | 29,557 | 28.75 | 452 |
| `gpt-5.6-sol` | max | 0.7267 | 8.39 | 60,014 | 61.25 | 450 |
| `gpt-5.6-terra` | max | 0.6962 | 4.95 | 71,939 | 75.93 | 451 |
| `gpt-5.6-luna` | max | 0.6719 | 3.03 | 73,400 | 101.68 | 448 |

주의:

- raw JSON의 모델 이름 `gpt-5-6-*`를 MoAI 제품 ID 표기로 대응했다.
- cost는 raw JSON의 `mean_cost_usd`이며 구독 청구액이 아니다.
- 렌더링된 사이트의 일부 cost와 raw JSON 수치가 달라 이 문서는 raw JSON만 일관되게 사용했다.
- 서로 다른 output token·step은 “저렴한 모델이 항상 빠르다” 또는 “큰 모델이 항상 비싸다”는 단순 결론을 허용하지 않는다.

### 5.2 Astra effort 비교

| effort | pass@1 | raw mean cost USD | output tokens | steps |
|---|---:|---:|---:|---:|
| xhigh | 0.7412 | 6.52 | 29,557 | 28.75 |
| high | 0.7323 | 5.72 | 26,500 | 27.42 |
| max | 0.7323 | 12.37 | 61,100 | 28.45 |
| medium | 0.7279 | 4.38 | 20,400 | 26.03 |
| low | 0.6704 | 2.19 | 10,600 | 19.54 |

`max`는 자동 최선이 아니다. 이 표에서 Astra `max`는 `high`와 pass가 같으면서 raw mean cost가 약 2.16배다. 따라서 `max`를 기본값으로 지정하지 않는다.

### 5.3 실제 선택 정책

| 상황 | 시작 선택 | 승격 조건 |
|---|---|---|
| 일반 개발 카드 | `gpt-5.6-sol`, medium 또는 high | 반복 실패, 고위험 설계, 검증 난도 상승 |
| 고위험 아키텍처·교차 제약 | `gpt-6-astra`, high | high 결과의 불확실성이 명시적으로 남고 live capacity가 확인될 때 xhigh |
| 제한된 비용 실험 | Terra 또는 Luna를 해당 작업군에서 내부 benchmark | 같은 acceptance에서 품질·wall time·usage가 Sol보다 유리할 때 확대 |
| 빠른 단순 수정 | 현재 lane 모델을 유지 | 모델 전환 비용보다 독립 호출 이득이 클 때만 MCP 전문 호출 |

lane 번호, `plan/run/sync`, agent 이름에는 모델을 고정하지 않는다. 카드 위험도와 내부 과거 acceptance 데이터가 dispatch 선택의 기준이다.

---

## 6. MoAI MCP V2 재설계

### 6.1 설계 원칙

- 기존 provider별 도구를 호환 가능하게 확장한다.
- 첫 단계에서 `model_task(provider=...)` 같은 만능 도구를 만들지 않는다. 인증·오류·capability 차이가 숨는다.
- agent preset이나 역할 이름을 API 필드로 만들지 않는다.
- task tool과 audit tool을 분리한다.
- lane의 메인 모델 정체성을 MCP 호출로 바꾸지 않는다.
- model fallback은 금지하고 실패를 구조화한다.

### 6.2 `codex_task` V2

```json
{
  "prompt": "required string",
  "model": "optional exact four-model id",
  "effort": "optional provider-supported value",
  "project_root": "optional canonical worktree path",
  "write": false,
  "background": false,
  "continuation": "new",
  "thread_id": "required only when continuation=resume"
}
```

동작 규칙:

- `model`을 생략하면 현재 세션 또는 configured default를 사용하되 resolved model을 결과에 반드시 남긴다.
- `model`은 네 모델 정적 allowlist와 live `model/list`의 교집합이어야 한다.
- `effort`는 해당 모델의 live capability와 정책으로 검증한다.
- `project_root`는 canonicalize하고 현재 repo/worktree 경계를 벗어나면 거부한다.
- `write=true`는 기존 프로젝트 opt-in과 호출 승인이 모두 있어야 한다.
- `continuation=resume`은 명시적 `thread_id`가 필요하다.
- 기존 `resume_last`는 deprecated compatibility로만 유지하고 Factory에서는 모호성 오류로 거부한다. 여러 lane이 같은 프로젝트에 있으면 “마지막 thread”는 안전한 식별자가 아니다.
- generic task의 기본 system을 auditor profile로 암묵 설정하지 않는다. 감사는 `codex_audit`가 담당한다.

### 6.3 `glm_task` V2

```json
{
  "prompt": "required string",
  "model": "optional configured GLM model",
  "thinking": "default | enabled | disabled",
  "system": "optional string",
  "max_tokens": "optional integer",
  "project_root": "optional canonical worktree path",
  "background": false
}
```

GLM에는 GPT식 `low/medium/high/...` effort를 거짓으로 투영하지 않는다. 공식 API의 provider-native thinking mode를 사용한다. [GLM Chat Completion API](https://docs.z.ai/api-reference/llm/chat-completion), [GLM Thinking Mode](https://docs.z.ai/guides/capabilities/thinking-mode)

현재 Factory 환경에서 explicit GLM model override를 조용히 무시하는 동작은 제거 대상이다. 명시 모델은 유효하면 그대로 쓰고, 사용할 수 없으면 오류를 반환해야 한다. 구독형 Coding Plan entitlement를 임의 endpoint에 재사용해도 된다고 가정하지 않는다. MoAI가 공식 지원 도구에 해당하는지 계약을 확인하고, 불확실하면 문서화된 일반 API 경로를 사용한다.

### 6.4 공통 결과 envelope

```json
{
  "status": "completed | failed | cancelled | queued | running | orphaned",
  "provider": "openai-codex | glm",
  "auth_mode": "appserver | api_key | configured",
  "requested_model": "...",
  "resolved_model": "...",
  "reasoning_requested": "...",
  "reasoning_resolved": "...",
  "project_root": "canonical path",
  "background": false,
  "job_id": "...",
  "thread_id": "...",
  "turn_id": "...",
  "write_requested": false,
  "write_granted": false,
  "usage": {
    "input_tokens": 0,
    "cached_input_tokens": 0,
    "output_tokens": 0,
    "source": "provider | estimated | unavailable"
  },
  "retryable": false,
  "error": {
    "class": "auth | capacity | policy | protocol | input | runtime",
    "code": "...",
    "message": "redacted"
  },
  "output": "..."
}
```

provider가 usage를 주지 않으면 0으로 꾸미지 말고 `source=unavailable`과 nullable 값을 사용한다.

### 6.5 job 수명주기

현재 foreground/background API 이름은 호환성을 위해 유지한다. 내부 저장 구조만 공통화한다.

- job ID에 provider, project root digest, session/lane ID를 묶는다.
- MCP 서버가 재시작되면 이전 `running` 레코드를 `orphaned`로 전환한다.
- process-local goroutine이 사라진 뒤에도 `running`이라고 거짓 보고하지 않는다.
- cancel은 provider cancellation receipt가 있을 때만 `cancelled`로 확정한다.
- Claude Code host가 MCP Tasks capability를 공식 노출할 때만 capability negotiation으로 `tasks/list`, `tasks/cancel` 등에 연결한다. 현재 지원을 가정해 독자 프로토콜을 만들지 않는다.

참고: [MCP Tools specification](https://modelcontextprotocol.io/specification/2025-06-18/server/tools), [MCP Tasks utility](https://modelcontextprotocol.io/specification/2025-11-25/basic/utilities/tasks)

### 6.6 감사 도구

기존 `codex_audit`, `glm_audit`, `audit_multi`는 task tool과 분리해 유지한다.

- `codex_task` 성공은 audit PASS가 아니다.
- 교차 회사 검토가 필요한 gate에서만 `audit_multi`를 사용한다.
- provider capacity나 인증 실패는 FAIL을 PASS로 완화하지 않고 GAP/NOT-RUN으로 남긴다.
- 작성 모델과 다른 회사/모델 검토라는 기존 cross-review 원칙은 gate에서만 강제한다.

---

## 7. Factory와 MCP의 조합

| 메인 세션 | native Agent() | 선택적 MCP 전문 호출 | 권장 사용 |
|---|---|---|---|
| `moai cc` | Claude 모델 상속 | `codex_task`, `glm_task`, 이미지 도구 | Claude가 주도하고 교차 검토·전문 작업만 외부 호출 |
| `moai gpt` | 선택한 동일 GPT 모델 상속 | `glm_task`; 필요 시 별도 GPT thread | GPT 메인 개발, 다른 GPT 모델이 꼭 필요할 때만 격리 호출 |
| `moai glm` | GLM 세션 정책 상속 | `codex_task`, 이미지 도구 | 경제적 주 작업 + 고난도 GPT 검토/실행 |
| Factory lead | lead 세션 모델 상속 | 원칙적으로 최소화 | 카드 분해·배정·증거 판정에 집중 |
| Factory lane | lane 시작 시 고른 모델 상속 | 경계가 작은 보조 작업만 | 카드 전체를 끝까지 소유 |

MCP를 모든 단계에 자동 호출하면 비용과 지연이 늘고, 실패 경계가 흐려진다. 다음 조건을 모두 만족할 때만 쓴다.

1. 작업 산출물과 입력이 작고 명시적이다.
2. 주 lane과 독립적으로 실패·재시도할 수 있다.
3. 다른 모델의 장점이 호출·컨텍스트 비용보다 크다.
4. 결과를 원래 lane이 acceptance 기준으로 검증한다.

용량 부족 시 fallback 체인은 자동 실행하지 않는다. `CAPACITY_UNAVAILABLE`, requested model, retryable을 반환하고 lead 또는 사용자가 새 lane/모델을 선택한다.

---

## 8. 이미지 생성 설계

텍스트 task에 이미지 옵션을 섞지 않고 별도 도구를 둔다.

```json
{
  "tool": "codex_image_generate",
  "provider_route": "codex-images-subscription",
  "model": "gpt-image-2.5",
  "prompt": "required string",
  "size": "provider-supported value",
  "quality": "provider-supported value",
  "output_format": "png"
}
```

필수 정책:

- exact model은 `gpt-image-2.5`다.
- 개발 중 공식 App Server/provider capability에서 이 ID와 원본 provenance가 확인되기 전에는 tool을 default OFF로 둔다.
- 현재 조사한 `claude-code-proxy` 경로는 `gpt-image-2`로 제한되어 있으므로 `gpt-image-2.5` 성공 근거가 아니다.
- 검증되지 않은 상태에서는 `MODEL_ID_UNVERIFIED`로 fail closed한다.
- `gpt-image-2`, Flare, Sunburst 또는 다른 모델로 조용히 대체하지 않는다.
- 결과는 MCP image content와 구조화 metadata를 함께 반환한다.
- base64 본문, prompt 원문, 인증 정보는 로그에 남기지 않는다.
- 이미지 생성은 코드 turn과 분리해 timeout·취소·usage를 독립 관리한다.

---

## 9. 구현 순서

### M0 — 문서와 제품 표면 정합성

- Kanban 제품 폐기를 canonical product/workflow 문서에 반영한다.
- `-k` CLI·규칙·테스트·도움말 제거를 별도 카드로 추적한다.
- Factory를 “lane이 카드 전체 생명주기 소유”로 단일 정의한다.
- 이미 merge된 t654/t840/t841 범위를 다시 구현하지 않는다.

### M1 — App Server production turn cutover

- 현행 `OpenAIAdapter + AuthPKCE` turn 경로를 공식 App Server authority/adapter로 교체한다.
- 설치·로그인·model/list·thread/start·turn/start의 readiness gate를 만든다.
- 정확한 네 모델과 live effort 교집합을 단일 resolver로 만든다.
- t842 model catalog SSOT와 중복 상수를 만들지 않는다.

### M2 — MCP V2 계약

- `codex_task`에 model, effort, project_root, explicit continuation/thread_id를 추가한다.
- generic task의 auditor 기본 profile 결합을 제거한다.
- 결과 envelope와 error taxonomy를 통일한다.
- 기존 필드는 호환하고 deprecated 동작은 경고한다.

### M3 — GLM 의미 정합성

- provider-native thinking mode를 추가한다.
- Factory에서 explicit model을 조용히 무시하는 분기를 제거한다.
- Coding Plan entitlement와 일반 API 사용 조건을 문서·설정으로 분리한다.

### M4 — job·lane 안전성

- job에 session/lane/project identity를 넣는다.
- restart reconciliation과 orphaned 상태를 구현한다.
- Factory에서 `resume_last`를 거부하고 explicit thread만 허용한다.
- worktree 밖 path와 교차 lane write를 차단한다.

### M5 — 이미지 capability gate

- 별도 이미지 tool과 redaction을 구현한다.
- `gpt-image-2.5` exact request·응답 provenance가 확인될 때만 enable한다.
- 미확인·다른 모델 응답·모델 자동 치환을 모두 실패 처리한다.

### M6 — Factory 관측성과 운영 acceptance

- lead/lane별 requested/resolved model, effort, provider, usage source를 receipt에 남긴다.
- t844 실세션, t848 Astra effort, t850/t851 400 이슈와 결과를 연결한다.
- 내부 카드 유형별 benchmark로 Terra/Luna 확대 여부를 판단한다.

---

## 10. 개발 시 acceptance와 테스트 지시

이번 문서 작업에서는 아래 테스트를 실행하지 않는다. 구현 카드의 acceptance로 사용한다.

### 10.1 Gateway 계약

- Claude Code의 multi-turn text, tool call, tool result, cancellation이 App Server thread/turn과 왕복한다.
- 네 모델 외 ID는 listener가 provider 호출 전에 거부한다.
- requested model과 resolved model이 다르면 실패한다.
- live `model/list`에 없는 effort는 실패한다.
- 알 수 없는 App Server event는 `PROTOCOL_DRIFT`로 실패한다.
- auth token이 argv, env dump, log, receipt에 나타나지 않는다.

### 10.2 세션·sub-agent

- `moai gpt --model gpt-5.6-sol` 메인 세션이 Sol로 확인된다.
- 그 세션의 native Agent가 모델을 따로 선택하지 않고 Sol을 상속한다.
- effort 변경은 모델 capability 범위에서만 전달된다.
- sub-agent UI에 가짜 역할별 GPT model preset이 노출되지 않는다.

### 10.3 Factory

- `-f`, `-f N`, `-f lane-N`의 documented 동작이 일치한다.
- lane 하나가 한 카드의 `plan → run → sync`를 소유한다.
- 서로 다른 lane을 서로 다른 main backend/model로 시작할 수 있다.
- 동일 lane 안에서 phase가 바뀌어도 main model이 자동 변경되지 않는다.
- `-k`는 제거 카드 완료 후 지원 중단 오류를 내고 도움말·규칙에 정상 기능으로 남지 않는다.

### 10.4 MCP

- `codex_task(model, effort)`가 exact value를 App Server에 전달한다.
- invalid model/effort가 fallback 없이 구조화 오류를 반환한다.
- explicit `thread_id`만 정확한 대화를 resume한다.
- Factory에서 `resume_last`가 명시적 모호성 오류를 반환한다.
- `project_root`의 symlink·`..`·다른 worktree escape가 거부된다.
- `write=true`는 프로젝트 opt-in 없이는 거부된다.
- MCP 재시작 후 이전 running job은 orphaned가 된다.
- provider usage가 없을 때 `usage.source=unavailable`이다.
- task 성공이 audit PASS로 집계되지 않는다.

### 10.5 이미지

- provider가 정확히 `gpt-image-2.5`를 선언하고 응답 provenance에 남길 때만 성공한다.
- `gpt-image-2`나 alias 응답은 성공 처리하지 않는다.
- capability 미확인 상태에서는 네트워크 호출 전 fail closed한다.
- image content와 metadata가 반환되고 base64/prompt/credential이 로그에 없다.

### 10.6 실환경 acceptance

- 신규 사용자 로그인, 기존 사용자 토큰 갱신, 로그아웃 후 재로그인을 관찰한다.
- 장시간 multi-turn, context compaction, 도구 호출, 취소, capacity failure를 관찰한다.
- Claude main, GPT main, GLM main, 혼합 Factory lane 조합을 각각 실행한다.
- API key가 아닌 실제 구독 계정 경로라는 provenance를 receipt로 확인한다.
- 실행한 명령과 원문 출력, 기준 commit, NOT-RUN 항목, 잔여 위험을 보고한다.

---

## 11. 개발 금지 사항

- 역할·단계별 모델 고정표를 코드화하지 않는다.
- native sub-agent와 MCP 외부 task를 같은 개념으로 부르지 않는다.
- Kanban 잔존 코드를 Factory 기능으로 재해석하지 않는다.
- unofficial proxy의 OAuth와 private endpoint를 생산 authority로 복사하지 않는다.
- 모델·effort·provider를 조용히 fallback하지 않는다.
- GLM thinking을 GPT effort 이름으로 위장하지 않는다.
- `gpt-image-2.5`가 확인되기 전에 다른 이미지 모델을 같은 이름으로 내보내지 않는다.
- 외부 Rust 테스트를 MoAI Go 구현 acceptance로 계산하지 않는다.
- local fixture나 선택 테스트만으로 구독 계정 운영 완료를 선언하지 않는다.

---

## 12. 증거 보고

### Claim

- `moai gpt`는 별도 보조 명령이 아니라 GPT gateway를 사용하는 메인 Claude Code 세션 launcher로 현재 develop에 연결되어 있다.
- Factory의 올바른 단위는 단계별 모델이 아니라 카드 전체를 소유하는 lane이다.
- Kanban은 제품상 폐기됐으며 현재 `-k` 코드·규칙은 별도 제거가 필요한 레거시 GAP이다.
- 현재 production GPT turn 경로는 공식 App Server end-to-end가 아니라 `OpenAIAdapter + AuthPKCE`여서 목표 구조로의 cutover가 남아 있다.
- 기존 MCP는 cross-model 호출 기반을 갖췄지만 모델·effort·thread·worktree·job 내구성 계약이 부족하다.

### Evidence

이번 조사에서 실행·관찰한 핵심 명령:

```text
$ git -C .claude/worktrees/develop rev-parse --short HEAD
4056f69e1

$ git -C .claude/worktrees/develop branch --show-current
develop

$ git -C .claude/worktrees/develop rev-list --count --left-right origin/main...HEAD
0 3937

$ git -C .claude/worktrees/develop rev-list --count --left-right origin/develop...HEAD
0 55

$ git merge-base --is-ancestor 15f3eacd7 HEAD; echo $?
0
$ git merge-base --is-ancestor 5f6632ef5 HEAD; echo $?
0
$ git merge-base --is-ancestor 6918d9db9 HEAD; echo $?
0
$ git merge-base --is-ancestor 2b7d9b2ad HEAD; echo $?
0
$ git merge-base --is-ancestor c044a911f HEAD; echo $?
0
```

코드 관찰 범위:

- `internal/cli/gpt.go`
- `internal/cli/root.go`
- `internal/cli/gateway_prepare.go`
- `internal/cli/gateway_product_binding.go`
- `internal/cli/gateway_factory.go`
- `internal/gateway/auth/broker.go`
- `internal/cli/factory.go`
- `internal/cli/cc.go`
- `internal/cli/mcp_server.go`
- `internal/cli/codex_task.go`
- `internal/cli/glm_task.go`
- `internal/cli/mcp_codex.go`
- `.moai/reports/t654/as5-verdict.md`
- t840/t841 관련 verdict와 현재 todo JSON readback

### Baseline-attribution

모든 구현 상태 판단은 2026-09-14에 읽은 `.claude/worktrees/develop`의 `develop@4056f69e1` 기준이다. 웹 사실은 같은 날 열어 본 위 공식 문서와 공개 프로젝트 문서를 기준으로 했다. DeepSWE 수치는 2026-09-03 갱신으로 표시된 live JSON의 해당 행을 사용했다.

### Gaps

- 이번 작업은 설계 문서 갱신이며 제품 코드를 수정하거나 제품 테스트를 실행하지 않았다.
- 실제 ChatGPT/Codex 구독 계정으로 `moai gpt` 장시간 세션을 실행하지 않았다.
- production App Server turn cutover는 현재 tracked develop에서 완료됐다고 확인되지 않았다.
- `gpt-image-2.5` subscription 모델 ID와 원본 provenance는 확인되지 않았다.
- GLM Coding Plan이 MoAI MCP 직접 호출을 공식 허용하는지 계약상 확정하지 못했다.
- Kanban 잔존 코드·규칙은 이번 문서 작업에서 제거하지 않았다.
- DeepSWE는 MoAI Factory와 동일 workload·과금 구조가 아니다.
- 다른 세션이 수정 중인 canonical SPEC와 untracked test는 이 문서의 완료 증거로 사용하지 않았다.

### Residual-risk

- Anthropic은 non-Claude gateway routing을 지원하지 않으므로 Claude Code 업데이트 때 호환성이 깨질 수 있다.
- 공식 Codex App Server도 protocol version과 event schema가 바뀔 수 있어 capability negotiation과 fail-closed 처리가 필요하다.
- 구독 allowance와 capacity는 계정·요금제·작업 길이에 따라 달라 고정 비용표로 보장할 수 없다.
- 모델 ID는 live catalog에서 사라지거나 effort capability가 달라질 수 있다.
- process-local MCP background 작업을 내구성 있다고 오인하면 완료·취소 상태가 틀릴 수 있다.
- 레거시 `-k` 표면을 제거하기 전에는 사용자와 agent가 폐기된 모드를 다시 호출할 수 있다.

---

## 13. 의사결정 요약

| 결정 | 최종 선택 |
|---|---|
| GPT 메인 사용 | `moai gpt` 전체 메인 세션 |
| GPT 인증·대화 authority | 공식 Codex App Server 목표 |
| Gateway 구현 언어 | 기존 Go 코드 확장 |
| 허용 GPT 모델 | Luna, Sol, Terra, Astra 정확히 네 개 |
| sub-agent 모델 | 메인 GPT 모델 inherit |
| Factory 모델 단위 | lane 세션 시작 시 한 번 선택 |
| Factory 작업 단위 | lane 하나가 카드 전체 생명주기 소유 |
| Kanban | 제품상 폐기, 잔존 `-k`는 별도 제거 GAP |
| 다른 모델 사용 | 별도 lane/세션 또는 제한된 MCP 전문 호출 |
| MCP API | provider별 `codex_task`/`glm_task` V2 |
| 이미지 | 별도 도구, exact `gpt-image-2.5`, 확인 전 fail closed |
| 자동 fallback | 금지 |
| 비용 전략 | Sol 기본, Astra high/xhigh 선별, max 비기본, Terra/Luna는 내부 benchmark 후 확대 |

이 설계의 핵심은 “모델을 많이 섞는 것”이 아니다. **긴 책임은 한 lane과 한 메인 모델에 맡기고, 다른 모델은 이득이 분명한 작은 경계에서만 호출하며, 모델·인증·비용·결과 provenance를 숨기지 않는 것**이다.

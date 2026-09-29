---
title: "moai gpt 팩토리 모드: Codex App Server 전환 설계"
date: "2026-09-14"
lang: ko
---

## 1. Claim — 명령과 팩토리 UX를 유지하는 설계가 가능하다

**권고: `moai gpt`, `moai gpt -f N`, `moai gpt -f lane-N`과 Claude Code의 도구·hooks·권한 체계를 유지하고, MoAI gateway의 GPT 실행부를 공식 Codex App Server로 교체한다.** 사용자에게 MCP 위임을 따로 요구하는 이전 권고는 이 요구를 충족하지 못한다. 이 보고서는 구현 가능성·설계안을 제시하며 제품 전체 성공이나 약관 무위험을 주장하지 않는다.

여기서 proxy는 Claude Code와 연결하는 **입구의 역할**, App Server는 GPT 인증·추론·세션을 담당하는 **출구의 구현**이다. 둘 중 하나를 없애는 선택이 아니다. 외형은 현재 proxy와 같고 내부는 상태를 가진 App Server bridge가 된다.

- Claude Code: 화면, Agent 호출, Bash/Read/Edit, ToolSearch, MCP, hooks, 사용자 승인과 팩토리 작업 지시.
- MoAI: Messages/SSE 호환, 모델 매핑, lane·agent 식별, 도구 결과 상관관계, 취소·재개·관측.
- Codex App Server: 공식 인증, GPT 모델 실행, native thread와 reasoning 이력.
- 초기 이관 대상: 새로운 팩토리 실행. 기존 직접 프록시의 진행 중 세션을 무손실 이전한다고 주장하지 않는다.

**정정:** Anthropic의 “비-Claude 모델 gateway 라우팅 비지원”은 실제 문서에 있다. 그러나 이 문장만으로 “구현 불가능” 또는 “곧바로 ToS 위반”이라고 판정할 수는 없다. 기술적 가능성, 공급사 지원, 계약상 허용을 나눠야 한다. 이 아키텍처는 OpenAI 측 공식 통합 경로를 사용하지만, Claude Code→GPT 조합에 대한 Anthropic의 지원 보증은 없다. [Anthropic gateway 문서](https://code.claude.com/docs/en/llm-gateway), [OpenAI App Server](https://developers.openai.com/codex/app-server/)

## 2. Baseline-attribution — 조사 기준과 범위

| 항목 | 이번 실행에서 확인한 값 |
|---|---|
| 코드 기준 | 기존 develop 워크트리, HEAD `c9ceff175` |
| 원격 비교 | fetch 후 `origin/develop...HEAD = 0 12` |
| Claude Code | `2.1.270 (Claude Code)` |
| Codex | `codex-cli 0.154.0` |
| 설치 MoAI | `v3.2.0-rc.10`, `643abfb8c-dirty`, build `2026-09-13T17:20:13Z` |
| 관련 진행 브랜치 | t654 `886959071`; t851 `54fcf2e1b` |
| 큐 스냅샷 | `moai todo list --json`: last_seq 853, items 147 |
| 현재 관련 카드 | t654·t837·t850·t851·t853 picked; t844·t848 queued |
| 작업 범위 | 코드/커밋/문서/로그 읽기, 비과금 선택 테스트, 스키마 생성, HTML·MD 보고서 |
| 미수행 | 로그인 변경, 모델 추론, 팩토리 기동, 바이너리 교체, 코드 수리, 큐 변경, commit/push |

큐의 147개는 반환된 현재 항목 수이며 전체 역사 카드 수가 아니다. 완료·보관 카드 상태는 이 스냅샷으로 추정하지 않았다. t650~653은 커밋 및 당시 보고서로 분석했다. 카드명과 예전 verdict가 충돌할 수 있으므로 카드 번호만 보지 않고 제목·브랜치·커밋을 대조했다. 예를 들어 t654의 일반 `verdict.md`는 과거 가드 오탐 보고이며 App Server의 AS5 증거는 `as5-*.md`에 있다.

공유 main에서 브랜치를 전환하지 않았다. 새 보고서는 develop 워크트리의 고유 디렉터리에 저장했다. 기존 picked 카드 작업물을 편집하지 않았다.

## 3. 현재 코드에서 확인한 두 가지 직접 원인

### 3.1 팩토리 진입점은 이미 공통화되어 있다

`internal/cli/gpt.go:56`은 `runClaudeEntry(..., kanban.BackendGPT, launch)`를 호출한다. `internal/cli/cc.go:163` 이후 공통 파서는 `-f`, `-f N`, `-f lane-N`을 lead/worker로 분기한다. 따라서 팩토리 제어면을 새로 만들 필요가 없다.

이번 선택 테스트는 공통 GPT 진입점, factory flag 파싱, lead/worker 환경을 각각 통과했다. **실제 GPT factory 전체 여정의 통과를 뜻하지는 않는다.** `-f N`은 공통 entry의 worker 수 설정이며, 모든 lane을 자동으로 켜는 명령이라고 확대 해석하지 않는다.

### 3.2 GPT의 네 모델 슬롯은 현재 하나로 합쳐져 있다

`internal/cli/gateway_prepare.go:104–116`의 GLM 경로는 High/Medium/Low/Fable을 각각 환경변수로 넣는다. GPT를 포함한 non-GLM 경로는 OPUS/SONNET/HAIKU/FABLE 모두 선택한 `model` 하나로 설정한다. `gateway_provider_contract_test.go`는 네 슬롯이 전부 astra가 되는 현재 계약을 검사한다.

따라서 외부 셸에서 `ANTHROPIC_DEFAULT_HAIKU_MODEL`만 바꿔도 launcher가 다시 덮어쓴다. **GLM식 GPT 매핑은 설정·생산 launch 조립·회귀 테스트를 함께 바꾸어야 한다.**

| 역할/별칭 | 제안 기본값 — 계정 capability 확인 후 채택 | 설정 의도 |
|---|---|---|
| Opus / high | gpt-6-astra | 고난도 설계·최종 판단 |
| Sonnet / medium | gpt-5.6-sol | 기본 실행 |
| Haiku / low | gpt-5.6-luna | 범위가 좁은 빠른 작업 |
| Fable | gpt-5.6-terra | 별도 코딩 작업 슬롯 |
| 명시 `--model` | 요청한 GPT 모델 | 해당 세션의 main 모델만 변경 |
| subagent `inherit` | 부모 모델 | 역할별 명시 override가 없는 경우 |

위 표는 성능 벤치마크 결론이 아니라 구성 제안이다. 프로젝트의 기존 역할 정책과 계정별 `model/list`를 대조해 확정한다. `--model` 때문에 전체 tier 매핑이 사라지지 않도록 main 선택과 tier 맵을 분리한다. 중간 turn의 모델 변경은 해당 thread가 idle일 때만 적용하고, pending 도구 결과는 원래 turn의 모델·effort로 완료한다.

Claude Code의 모델 별칭 환경변수, custom picker, allowlist 적용 범위는 공식 문서에 있다. gateway 모델 discovery는 ID에 `claude` 또는 `anthropic`이 없는 모델을 걸러내므로 `/v1/models`에 GPT ID를 반환하는 것만으로 picker가 완성되지 않는다. 이미 있는 명시 picker/모델 설정 경로를 검증하고 실제 GPT 이름을 표시한다. 모델명을 Claude처럼 속일 필요는 없다. [모델 설정](https://code.claude.com/docs/en/model-config), [gateway 프로토콜](https://code.claude.com/docs/en/llm-gateway-protocol)

## 4. 카드·커밋 분석 — 무엇이 끝났고 무엇이 남았는가

아래 “포함”은 `git merge-base --is-ancestor <sha> c9ceff175` exit 0을 뜻한다. 카드 완료·원격 배포·실계정 성공과 다르다.

| 카드 | 커밋/근거 | develop 포함 | 실제 내용과 남는 범위 |
|---|---|---|---|
| t649 | ce79ef7ca | 포함 | GPT launcher·인증·직접 Responses 변환·runtime 수리 통합. 현재 실제 경로의 출발점 |
| t650 / AS1 | dad5e9a9e | 포함 | App Server stdio·managed auth·profile·RPC 기반. 실계정 전부 성공은 아님 |
| t651 / AS2 | 28b9988d7 | 포함 | 초기 도구 + 후발 dispatcher hybrid registry, schema·권한 검증 |
| t652 / AS3 | 530bd7330 | 포함 | turn/tool continuation, pending 내구성, 취소/EOF 경계. mock/subprocess 자산 |
| t653 / AS4 | 4976e5a06, 14dba89c5 | 포함 | owned-thread resume, idle 모델 변경 |
| t653 / AS4 | e45f50a8d, 64885fa06 | 포함 | summary-turn rebase, 완료 경계 fork 기반. native Agent fork 실증은 별도 gap |
| t654 / AS5 | 5672f5029, aec09a5d6, 3e28b60e3, 2113c1e14 | 미포함 | auth/outputPolicy 표시, epoch 복원, capability 검사. 최종 생산 App Server cutover로 단정 불가 |
| t695 | .moai/reports/t695/verdict.md | 당시 보고 참조 | effort/max_tokens/숨겨진 변환 오류 개선. 57셀 결과는 당시 실행이며 이번 baseline 수치 아님 |
| t703 | 9bc624a39 | 포함 | 모델 전환 시 봉투 탈락을 CauseReasoning으로 분류. 거절 원인 표시이지 정상 진행 수리는 아님 |
| t707 | 1b4547e3a | 포함 | round-trip·tamper 대조군과 offline forensic. 실제 wire body 부재가 핵심 gap |
| t708 | b46d33271 | 포함 | launcher의 봉투 복구·byte-preserving splice. CauseChain 전체 수리가 아님 |
| t700 | a2f46eda3 | 포함 | 실패 후 단발성 client-side recovery. 의미가 달라진 과거 이력을 임의 허용하는 해결은 아님 |
| t697 | .moai/reports/t697/verdict.md | 당시 보고 참조 | upstream 5xx와 연결 오류 재시도. 400 receipt 검증 실패와 구분 |
| t838·850 | runbook + 현 큐 | 진행 관측 | 원본 요청 경계 계측. 기존 stderr 폐기로 진단이 가려진 문제 |
| t839·849 | 53c672eb3 / 1920216e3 | 범위별 참조 | retry-safe fixture, dial 대신 served counter. 테스트 수리 ≠ 실서비스 400 해결 |
| t841·848 | 2b7d9b2ad / queued | t841 포함 | 모델별 effort 허용표. 과거 astra max 거절과 현재 문서의 차이는 live 재측정 필요 |
| t851 | a5ccaca5a → 54fcf2e1b | 미포함 | CauseChain forensic. 최신 verdict는 삽입/에코 차이를 경쟁 가설로 남김 |
| t844 | queued | 해당 없음 | 실제 회상·turn 모델 일치·Claude 압축 양성 실증 이관 |
| t837·853 | picked | 해당 없음 | 전면 검수/ToS 비교 보고 범위와 겹침. 이번 보고서는 기존 카드 상태를 변경하지 않음 |

핵심 공백은 `internal/cli/gateway_product_binding.go:162` 생산 팩토리다. GPT는 여전히 Codex broker + `PolicyGPTNative` + HTTP transport로 조립된다. `gatewayGPTModels`는 `AuthPKCE`를 선택하고 `auth/store.go:378`은 `chatgpt.com/backend-api/codex/responses`를 직접 지정한다.

App Server 어댑터가 존재하고 테스트가 통과해도 생산 팩토리가 그 어댑터를 만들지 않으면 사용자의 `moai gpt`는 계속 예전 경로를 탄다. t654의 `authMethod`/`outputPolicy`는 표시 데이터이며 그 키가 들어갔다는 사실은 실제 transport 전환 증거가 아니다. nil service/binding 때의 “awaiting transport verification” 문자열도 정상 경로 전체가 차단됐다는 증거가 아니다.

## 5. 400·502·멈춤 문제의 분류

| 오류 계열 | 관측/증거 | 이번 판정 | 필요한 해결 |
|---|---|---|---|
| CauseChain / Edit 직후 반복 400 | 실제 family 2e343aa0 로그 932·1004·1090·1127행 | gateway 자체 receipt 검사 발생 확정 | 성공 응답에서 발행한 메시지와 다음 wire 요청의 구조·값 비교; 이벤트/신뢰 경계 재설계 |
| CauseReasoning / 모델 전환 | t703 + 현재 ModelSwitch 테스트 | 봉투 제거 시 거절 재현. sol→luna 자체는 정상 이력에서 통과 | reasoning을 Claude의 thinking 재전송에 의존하지 않고 Codex thread에 보존 |
| CauseLineage / 요약·자식 세션 | 오류 분류 코드, 과거 agent_summary 기록 | 새 agent와 부모 receipt의 소유 경계 문제 후보. 모든 fork를 동일 원인으로 묶을 수 없음 | session/agent 키, 명시 부모 경계, 새 thread 또는 검증된 native fork |
| is_error=true tool_result | request.go:401–423 | 현재 실패 텍스트 변환이 존재. 정상 도구 실패는 400 사유가 아님 | 실패를 정상 tool output으로 반환하고 `success:false`를 매핑 |
| reasoning effort / max_tokens / beta 필드 | t695·841·848 | 모델·경로별 capability 차이 및 변환 정책 | model/list + 설치 스키마 + 실제 계정 probe. 자동 effort 변경/과금 전환 금지 |
| upstream 5xx / 연결 실패 | t697 + 이번 retry 테스트 | 사유 있는 502와 제한 재시도 동작 확인 | 초기/중간 스트림 구분. 결과가 불명확한 turn 재실행 금지 |
| 실패 후 wedge | t700·708 + receipt 정책 | 클라이언트 기록과 발행 경계 불일치가 세션에 잔류 가능 | 완료된 경계에서만 복구. 부분 응답과 성공 receipt의 원자성 |
| 긴 무응답 | appserver.go:66, engine.go의 seg.Text 누적 | 새 App Server adapter는 tool/turn 경계까지 segment buffering | text delta 전달, SSE ping, backpressure, durable barrier 후 terminal event |
| profile/trust 및 stale binary | 설치 빌드와 source HEAD 다름 | 코드 통과와 설치 실행은 서로 다른 기준 | 실행 바이너리 경로/SHA·effective profile·child env 기록 |

### 5.1 이번 실행의 직접 대조

기존 t851 읽기 전용 도구를 실행해 실제 family의 manifest와 transcript 봉투를 다시 비교했다. **후보 7개, generation 8, reasoning 봉투 digest와 체인 이전 링크 7/7 일치**를 관측했다. 실제 로그에는 같은 CauseChain 400이 네 번 있다. 이것은 봉투 손상이 이 표본의 필수 원인이라는 설명에 반대되는 증거다.

추가로 develop 코드를 변경하지 않고 Go overlay 테스트를 실행했다. 동일한 발행 이력에서 다음 결과가 나왔다.

```text
=== RUN   TestResearchReminderPlacement
unchanged=accepted tail_reminder=accepted
mid_history_reminder=CauseChain; opaque envelopes unchanged
--- PASS: TestResearchReminderPlacement (0.06s)
ok github.com/modu-ai/moai-adk/internal/gateway/translate 0.462s
```

**이 실험은 “중간 삽입이 이 검증기를 실패시킬 수 있음”을 증명한다. 실제 실패 요청에서 그 삽입이 일어났다는 증거는 아니다.** t851 최신 verdict도 (I) 과거 이력 삽입과 (II) Edit 응답 에코의 값 변화 두 가설을 남긴다. 공식 prompt-caching 문서는 대화 도중 파일 변경 알림 등 system context를 추가한다고 설명하지만, 그것만으로 특정 과거 Read 위치를 나중에 수정했다고 확정할 수 없다. [Claude Code prompt caching](https://code.claude.com/docs/en/prompt-caching)

### 5.2 피해야 할 수리

`<system-reminder>` 태그를 찾고 그 내용을 전부 해시에서 제외하면 안 된다. 파일·도구 결과·사용자가 같은 태그 문자열을 만들 수 있기 때문이다. “태그 모양”은 신뢰된 Claude 주입물의 출처 증명이 아니다. 동일 길이 placeholder도 내용 바꿔치기를 놓칠 수 있으며 삽입 수·길이가 달라지면 여전히 실패할 수 있다.

먼저 실패 전후의 wire 경계와 도구 call ID·인수 digest·결과 digest를 포착해야 한다. 접두부 위치는 조사 단서이고 의미 차이의 종류까지 증명하지는 않는다. 본문 캡처가 필요하면 해당 세션의 로컬 비공개 파일로 한정한다. 모든 로그에 프롬프트 원문을 저장하는 방식은 채택하지 않는다.

## 6. 최신 자료·라이브러리 비교

| 후보 | 최신 1차 자료에서 확인한 기능 | 요구 적합성 | 선택 |
|---|---|---|---|
| 공식 Codex App Server | stdio, managed login, thread/turn, model discovery, dynamic tools, fork 경계 | Codex를 실행하면서 Claude UI·도구권을 유지하는 bridge 가능 | **주 권고** |
| OpenAI Responses API + API key | 명시적 conversation state 관리 | GPT 실행은 가능하지만 Codex runtime을 사용하는 것은 아님 | API 과금 허용 시 단순 대안·대조군 |
| LiteLLM | Claude Code용 Anthropic Messages 변환, 비-Anthropic 모델 공식 튜토리얼 | API key 기반 GPT proxy에 적합. native Codex thread 완성품은 아님 | API 경로 우선 비교 대상 |
| Vercel AI Gateway | Claude Code compatibility endpoint와 모델 선택 설정 | managed gateway 대안. 외부 중계·별도 과금/데이터 경로가 생김 | 외부 서비스 허용 시 대안 |
| Claude Code Router | 로컬 routing·provider 변환·agent 설정 | 제품별 라우팅 기능 참고 가능. MoAI receipt/factory 전체 보존은 미실증 | 구현 패턴 참고 |
| CLIProxyAPI | Codex OAuth·여러 CLI 호환 API | 구독을 API로 노출하는 기능은 있으나 공식 App Server managed 경로와 동일하지 않음 | ToS 불확실성을 줄이는 주 대안으로 삼지 않음 |
| Codex SDK | 코드에서 Codex task 실행·재개 | 백엔드 작업 위임에는 편리. Messages/SSE·Claude 도구 continuation은 별도 필요 | 현재 요구의 drop-in 대체 아님 |
| MCP codex_task | 기존 MoAI 도구 위임 | Claude가 주 모델로 남음. 사용자 요구와 다른 UX | 보조 리뷰용만 유지 |

자료: [App Server](https://developers.openai.com/codex/app-server/), [Responses state](https://developers.openai.com/api/docs/guides/conversation-state), [LiteLLM](https://docs.litellm.ai/docs/tutorials/claude_non_anthropic_models), [Vercel](https://vercel.com/docs/ai-gateway/coding-agents/claude-code), [Claude Code Router](https://github.com/musistudio/claude-code-router), [CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI), [Codex SDK](https://developers.openai.com/codex/sdk/).

각 제품의 최신 문서를 확인했지만 외부 라이브러리를 설치·실행 비교하지는 않았다. “최신 릴리스가 가장 안정적”이라는 주장은 하지 않는다. 현재 MoAI에 stdio/RPC/bridge/registry가 이미 있으므로 외부 프록시를 추가로 겹치기보다 기존 구성요소를 실제 생산 경로에 연결하는 편이 작업 범위를 줄인다.

## 7. 구체 설계 — Messages 입구 + 상태형 App Server 출구

### 7.1 팩토리 실행과 식별

사용자 명령은 유지한다.

```sh
moai gpt
moai gpt -f 4
moai gpt -f lane-2
moai gpt --model gpt-6-astra
```

아래는 **제안 설정 형상이며 현재 CLI 지원 문법이 아니다.**

```yaml
llm:
  gpt:
    transport: codex-app-server
    auth: chatgpt  # 또는 사용자가 명시한 api-key
    models:
      high: gpt-6-astra
      medium: gpt-5.6-sol
      low: gpt-5.6-luna
      fable: gpt-5.6-terra
    default_tier: medium
```

thread 키는 계정 scope + 프로젝트 canonical root + factory run ID + lane ID + Claude session ID + agent ID를 결합한다. 모델은 키의 영구 정체성에서 분리해 idle 전환을 허용한다. 각 lane은 독립 thread 집합과 pending map을 가진다. 첫 버전은 lane별 App Server 프로세스 경계로 장애 전파를 제한하되, 인증 profile lock과 공유 계정 동시성 정책은 실측 후 정한다. 여러 프로세스가 단일 private profile의 배타 lock을 동시에 취득할 수 있다고 가정하지 않는다.

최신 Claude Code 문서는 `x-claude-code-session-id`, `x-claude-code-agent-id`, `x-claude-code-parent-agent-id`를 공개한다. 이 헤더를 gateway 인증·소유 lease와 결합한다. 헤더 문자열만으로 다른 lane/thread 접근을 허용하지 않는다. 현재 `internal/gateway`/gateway CLI 범위 검색에서는 이 헤더 이름 소비를 찾지 못했으므로 별도 통합 항목이다. 실제 2.1.270이 모든 main/summary/nested 요청에 헤더를 보내는지는 무과금 mock ingress에서 포착해야 한다. [프로토콜 문서](https://code.claude.com/docs/en/llm-gateway-protocol)

### 7.2 도구 호출은 Claude에서 실행

1. Claude 요청의 완전한 초기 tool schema를 App Server `dynamicTools`로 등록한다.
2. GPT가 tool을 호출하면 App Server의 server request ID·thread·turn·call ID를 pending record에 먼저 저장한다.
3. gateway는 Claude에 같은 기능의 `tool_use`와 `stop_reason=tool_use`를 반환한다.
4. Claude가 기존 승인·PreToolUse·실제 도구·PostToolUse를 수행한다.
5. 다음 Messages 요청의 `tool_result`를 pending call에 정확히 한 번 결합한다.
6. App Server의 기다리는 `item/tool/call`에 응답하고 같은 native turn을 계속한다.

ToolSearch 뒤에 발견된 도구는 t651의 authenticated snapshot과 schema 검증을 거친 dispatcher를 이용한다. tool 이름만 받은 결과로 임의 실행 권한을 만들지 않는다. Codex의 shell·patch·MCP·native multi-agent는 초기 bridge에서 비활성화해 두 실행 엔진이 같은 파일을 바꾸지 않게 한다. 이 때문에 Codex 자체의 모든 도구 UX가 Claude Code 안에 그대로 나타나는 것은 아니다.

`dynamicTools`와 `item/tool/call`은 공식 문서상 experimental이다. 이를 숨기지 않고 Codex 버전·스키마 hash를 pin한다. 설치 0.154.0에서 생성한 JSON Schema에 `ThreadStartParams.dynamicTools`, `ThreadForkParams.lastTurnId`, `TurnStartParams.model/effort` 존재를 직접 확인했다. 스키마 존재는 실제 모델 tool 왕복 성공과 다르다.

### 7.3 reasoning과 이력의 소유권

Codex thread를 GPT 실행 이력의 기준으로 삼는다. Claude transcript는 UI·도구 실행 기록과 새 입력 전달을 담당한다. 매 요청마다 Claude가 되돌려 보낸 전체 이력을 OpenAI reasoning과 섞어 재구축하지 않는다.

단, “마지막 사용자 메시지만 보낸다”는 단순화도 틀리다. 정책·파일 변경 알림·압축·새 tool 결과를 놓칠 수 있다. 입력을 다음 세 부류로 조정하는 reconciliation 계층이 필요하다.

- 이미 수락한 tool call/result와 assistant 출력: ID·소유권·내용 digest를 확인하고 재실행하지 않는다.
- 새로운 사용자/도구 입력: 정확히 한 번 native turn 또는 pending RPC에 전달한다.
- Claude의 컨텍스트 변경: 검증된 변경 이벤트로 새 turn에 전달하거나, 의미가 달라진 이력은 명시적으로 새 thread로 분기한다. 출처를 식별할 수 없는 변경은 추측해 무시하지 않는다.

이 설계는 receipt 무결성을 버리는 것이 아니다. 임의로 재작성되는 전체 transcript prefix 대신 **소유권·도구 호출·결과·완료 경계와 변경 이벤트**를 검증한다. 기존 prefix 알고리즘을 App Server 앞에 그대로 남기면 동일한 400이 재발할 수 있다.

### 7.4 하위 에이전트와 fork

평범한 Claude Agent 작업은 자식의 새 prompt로 새 Codex thread를 만들 수 있다. 부모가 아직 tool 요청에서 대기 중이어도 자식을 부모 thread에 붙이지 않는다. 부모에는 Agent 결과를 tool_result로 돌려준다.

진짜 과거 이력 fork가 필요한 경우에만 공식 `thread/fork(lastTurnId)`를 사용한다. 설치 스키마는 완료되지 않은 turn 경계의 fork를 허용하지 않는다고 명시한다. 따라서 진행 중인 부모 turn을 그대로 복제하는 것을 필수 전제로 두지 않는다. 완료 경계 + 명시적 task context로 새 자식을 시작하는 경로를 기본으로 한다. t653 현재 테스트는 “새 thread + inherited prefix”를 검증하며 Codex native fork 양성 실증과 동일하지 않다.

### 7.5 압축·재개·취소

- resume: 서버가 소유한 thread ID와 저장한 binding을 재개한다. `thread/resume.history` 재구축으로 native reasoning을 복원했다고 주장하지 않는다.
- compact: Claude의 실제 압축 요청을 식별해 summary turn과 단 한 번의 rebase를 기록한다. 모든 요약 요청을 전체 대화 reset으로 처리하지 않는다.
- cancel: 해당 turn만 interrupt한다. 다른 lane의 App Server 프로세스를 죽이지 않는다.
- crash: RPC 응답 성공 여부가 불명확한 pending tool을 자동 재실행하지 않는다. 기존 job/call 기록을 조회한 후 복구하거나 명확한 recovery 상태를 반환한다.
- streaming: text delta는 즉시 중계하고 침묵 구간에는 ping을 보낸다. 성공 terminal event는 durable completion 뒤 보낸다. 현재 segment buffering 구현은 별도 개선 대상이다.

## 8. ToS 판단 — 가능한 범위와 남는 제약

OpenAI 공식 App Server의 ChatGPT managed login을 사용하면 Codex가 OAuth·저장·갱신을 소유한다. API 키 모드도 명시적으로 선택할 수 있다. 직접 토큰 복사·내부 endpoint 호출을 없애는 것이 현재 구독 프록시보다 근거가 분명한 선택이다. [인증 문서](https://developers.openai.com/codex/auth/)

그러나 App Server를 이용한다는 사실만으로 구독 한도 우회·계정 공유·재판매가 허용되지는 않는다. 팩토리 lane 수는 한도 확대가 아니다. 계정 단위 동시성 제한·429 backoff를 적용하고 계정 회전으로 제한을 피하지 않는다. API 과금으로 자동 전환하지 않는다. [OpenAI 이용약관](https://openai.com/policies/terms-of-use/)

Anthropic 문서의 non-Claude gateway 비지원은 유지된다. 공식 gateway 설정 표면을 이용하는 것과 GPT 조합을 Anthropic이 보증하는 것은 다르다. Claude 바이너리 patch, 숨겨진 인증 우회, Claude 구독 토큰을 외부 서비스에 사용하는 방식은 설계에 포함하지 않는다. 소비자/조직 계약에 따라 허용 범위가 달라지므로 이 보고서는 법률 보증을 제공하지 않는다. [Anthropic 소비자 약관](https://www.anthropic.com/legal/consumer-terms), [gateway 지원 경계](https://code.claude.com/docs/en/llm-gateway)

**사용자 요구에 맞는 기술 권고는 App Server bridge다. 양사 공식 지원까지 필수라면 현재 공개 문서로 그 조건을 충족하는 동일 UX를 확인하지 못했다.** 이 두 문장을 함께 제품 결정에 남겨야 한다.

## 9. 구현 순서와 인수 조건

### High — 기준 고정과 실제 400 wire 계측

t851·t850의 기존 소유권을 존중하고 성공 Publish와 실패 Check의 경계별 구조/digest를 수집한다. 계측 패치를 적용하기 전에 실행 바이너리 SHA와 source SHA를 맞춘다. synthetic 테스트 결과만으로 reminder stripping 수리를 출시하지 않는다.

### High — 팩토리 매핑과 실제 transport 배선

GPT tier 설정을 GLM과 같은 구조로 추가한다. main 모델 override와 네 슬롯을 분리한다. `productionGatewayHandlerFactory`가 실제로 managed App Server runtime·registry·engine·adapter를 생성하도록 연결한다. 표시 필드나 nil guard 테스트만으로 cutover를 판정하지 않는다. 기존 구성요소를 우선 재사용한다.

### High — 1 lane + 1 child 수직 검증

새 세션에서 main→Read→Edit→tool_result→다음 turn, main→Agent→자식 도구→부모 결과 경로를 검증한다. 과거 이력 변화와 tool 에코 차이를 모두 포함한 음성·양성 쌍대 테스트를 둔다. 계정·lane·agent·call ID를 바꾼 입력은 계속 거절되어야 한다.

### Medium — ToolSearch·resume·compact·모델 변경

후발 도구 schema·권한 회수, idle 모델 전환, pending 중 모델 변경 거절, 프로세스 재시작, 실제 회상, Claude 압축과 hook 후속을 검증한다. t844의 실계정 항목을 이 단계와 연결한다.

### Medium — 다중 lane 팩토리

2 lane에서 시작해 목표 N lane으로 확대한다. 한 lane 취소/EOF/429가 다른 lane을 손상시키지 않아야 한다. child/parent agent 요청의 thread 매핑과 계정 단위 backpressure를 검사한다. production 과정에서는 실제 task 수행 결과와 factory 상태 저장·재개까지 확인한다.

### 출시 판정표

| 인수 항목 | 완료 기준 |
|---|---|
| 사용자 명령 | gpt, -f N, -f lane-N, --model, --spawn, --continue 동일 진입 동작 |
| 모델 매핑 | main/Opus/Sonnet/Haiku/Fable 요청의 선택 모델과 native turn 실제 모델 일치 |
| 도구권 | Bash/Read/Edit/Agent는 Claude 쪽 승인·hooks를 거쳐 정확히 한 번 실행 |
| 이력 | Read→Edit 뒤 정상 계속, 실제 wire 비교로 원인 해소; 변조/교차계정은 거절 |
| 하위 에이전트 | main/child/nested/agent_summary별 구분, 형제 thread 오염 없음 |
| 회복 | resume·compact·취소·중간 EOF·프로세스 crash 각각 관측 |
| 스트리밍 | 첫 이벤트·첫 텍스트·완료 시각 구분; heartbeat 유지, 완료 전 성공 표시 없음 |
| 동시성 | N lane + 자식 실행에서 call result 오배달·중복 실행·공유 profile 경쟁 없음 |
| 인증 | 구독과 API 키 양쪽 실증, 자동 billing 전환 없음 |
| 운영 | 설치 binary provenance, macOS/Windows CI, 기능별 report/readback |

이 보고서는 설계안까지 제공한다. 제품 코드 변경은 별도 실행 작업이며 현재 카드들의 완료 상태를 바꾸지 않는다.

## 10. Evidence — 이번 명령과 출력

모든 Go 명령은 develop `c9ceff175`에서 실행했다. API 관련 환경변수는 같은 compound invocation에서 해제했다. 전체 스위트는 실행하지 않았다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY
# 실제 실행에서는 위 unset과 아래 go test를 &&로 한 호출에 결합했다.
go test ./internal/cli -run 'TestGPTLaunchPreservesCommonEntry|TestParseFactoryFlag|TestResolveFactoryBranch|TestEnterFactory.*Env|TestGatewayProviderContractPickerAndSlots' -count=1 -v -timeout 90s
go test ./internal/gateway/translate -run 'TestModelSwitch|TestSubscriptionReceiptModelSwitch|TestReceiptHistoryRejectsPublicLayoutMutation|TestReceiptHistorySpawnShapedFreshHistoryNeedsNoLineageSeeding' -count=1 -v -timeout 90s
```

원출력의 패키지 결과:

```text
ok   github.com/modu-ai/moai-adk/internal/cli 0.945s
ok   github.com/modu-ai/moai-adk/internal/gateway/translate 0.558s
```

bridge/HTTP adapter/registry의 선택 검증:

```text
TestCompactionIsANormalSummaryTurnWithZeroCompactRPCs PASS
TestToolSegmentRetainsRPCAndCompletesExactlyOnce PASS
TestRestartPendingNeverReplays PASS
TestForkChildStartsNewThreadAtInheritedPrefix PASS
TestResumeAttachesOwnedThreadWithoutRebuild PASS
ok github.com/modu-ai/moai-adk/internal/codexbridge 0.855s

TestAppServerSubprocessHTTPToolContinuation/chatgpt PASS
TestAppServerSubprocessHTTPToolContinuation/apiKey PASS
TestUpstreamRetryRecoversSingleFiveHundred PASS
TestUpstreamNoRetryAfterResponseStarted PASS
ok github.com/modu-ai/moai-adk/internal/gateway 2.137s

TestHybridPositiveAndNoThreadRecreation PASS
TestSixNegativeGroups PASS
TestNativeToolWireDiscriminator PASS
TestRevocationBeforeInvocation PASS
ok github.com/modu-ai/moai-adk/internal/codextools 0.241s
```

위 테스트 이름 PASS 표는 원출력의 축약이며, subprocess의 chatgpt/apiKey는 **fake App Server fixture**다. 실제 구독/API 인증 성공으로 해석하지 않는다. 최초 묶음에서 codextools는 `[no tests to run]`이었고 이를 PASS로 세지 않은 뒤 해당 실제 테스트 이름으로 따로 실행했다.

스키마 검증 명령:

```sh
codex app-server generate-json-schema --experimental --out /tmp/moai-gpt-factory-schema-5fWMMB
```

실행 exit 0. 생성물의 `ThreadForkParams.properties.lastTurnId`, `ThreadStartParams.properties.dynamicTools`, `TurnStartParams.properties.model/effort`를 jq로 판독했다. 선언된 프로토콜 확인이며 실제 모델 응답은 아니다.

별도 `evidence.md`에 실행 명령·출력·synthetic overlay 원본 위치를 남겼다.

## 11. Gaps와 Residual-risk

- 실제 실패 request wire body가 없어 mid-history 삽입의 정확한 위치·값을 확정하지 못했다.
- 실제 Codex account/model/turn을 호출하지 않았다. 설치 스키마·mock 통과로 실계정 성공을 주장하지 않는다.
- 실제 `moai gpt -f` N lane UI 여정, Windows CI, native fork, live compact/resume는 미실행이다.
- t654/t851의 진행 브랜치를 읽었지만 develop에 포함되지 않은 커밋을 포함됐다고 표시하지 않았다.
- 기존 receipt 경로와 새 App Server 경로는 한동안 공존할 수 있다. runtime 식별 로그 없이 어느 경로가 사용됐는지 추정하면 잘못된 완료 판정이 재발한다.
- App Server dynamic tool 계약, Claude custom model·헤더·기능은 버전별 차이가 있다. schema pin과 무과금 클라이언트 probe가 필요하다.
- 모델 전환·하위 작업의 실제 품질·비용·지연은 측정하지 않았다. tier 제안은 실측 성능 순위가 아니다.
- 계정 scope 격리, tool call/result의 변조 방지, 중복 실행 방지는 새 이력 설계에서도 유지해야 한다.
- 조사 중 활성 카드/커밋은 바뀔 수 있다. 원격 배포·계약상 허용·운영 인수는 이 보고서의 완료 범위 밖이다.

## 12. 근거 파일

- [현재 GPT 생산 팩토리](../../internal/cli/gateway_product_binding.go): 162, 236행
- [모델 매핑](../../internal/cli/gateway_prepare.go): 104–116행
- [팩토리 공통 진입점](../../internal/cli/cc.go): 163–205행
- [receipt 검사](../../internal/gateway/translate/receipt_history.go), [정준 prefix](../../internal/gateway/receipt/projection.go)
- [App Server HTTP adapter](../../internal/gateway/appserver.go): 59–69행
- [t703 모델 전환 판정](../../.moai/reports/t703/verdict.md)
- [t707 round-trip 판정](../../.moai/reports/t707/verdict.md)
- [t653 native fork gap](../../.moai/reports/t653/m1-native-fork-probe.md)
- [t654 실제 진행 브랜치의 AS5 보고](../../../t654/.moai/reports/t654/as5-launcher-integration.md)
- [t851 최신 종합 판정](../../../t851/.moai/reports/t851/verdict.md)
- [t851 초기 원인 보고](../../../t851/.moai/reports/t851/root-cause.md)

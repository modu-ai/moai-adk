# Claude Code 안의 다중 LLM Factory — 최적 통합안

조사 기준: 2026-09-14 · 설계 비교 및 공개 근거 조사 · 구현·설치·유료 호출 없음

## 1. 결론: moai-mcp 중심으로 조율하고, Gateway는 선택 경로로 둔다

**현재 요구에 가장 적합한 기본안은 Claude Code lead·lane + 기존 moai-mcp 도구 + 공식 Codex 실행부 + 공급자별 허용 API의 조합이다.** 각 lane의 Claude Code 에이전트가 GPT·GLM에 전문 작업을 맡기고, MoAI가 작업 소유권·모델·effort·비용·결과를 관리한다. GPT를 Claude Code의 주 모델로 직접 쓰는 Gateway 경로는 별도 선택 기능으로 유지한다. 이는 설계 권고이며 완성된 제품의 성능 평가가 아니다.

결정적 근거는 OpenAI의 공식 [Codex plugin for Claude Code](https://github.com/openai/codex-plugin-cc)다. Claude Code 안에서 작업과 리뷰를 Codex에 위임하고, ChatGPT 구독을 사용하며, 모델·effort를 지정하는 사례가 공급자 자신에게서 확인된다. 기존 기획안의 “공식 실행기로 가면 Claude Code 내부 조건을 포기해야 한다”는 구분은 **도구 위임까지 포함하면 성립하지 않는다.**

다만 “Claude Code 안”의 의미를 바꾸어 결론을 내리지는 않는다.

| 요구의 정확한 의미 | 판정 | 가장 적합한 경로 |
|---|---|---|
| 사용자는 Claude Code에서 지시·승인·결과 확인을 하고, GPT·GLM이 전문 작업 수행 | 충족 가능한 설계. 공식 선례 있음 | moai-mcp + 공식 실행기/API |
| 모든 lane의 Claude Code 자체 추론을 GPT·GLM이 담당 | MCP 위임만으로 충족하지 않음 | Messages Gateway 필요, Anthropic 비지원 경계 존재 |
| 모든 파일·셸 도구 호출이 반드시 Claude Code 개별 승인·hook을 통과 | 외부 실행기의 자체 도구 실행으로는 자동 충족 안 됨 | 읽기·제안 전용 MCP 또는 별도 도구 중개 검증 |
| 구독만으로 무인·무제한·상용 다중 사용자 Factory | 확인된 보편적 허용 경로 없음 | 자동화 허용 계약/API/기업용 인증을 별도로 선택 |

**실행 인터페이스, 모델 선택, 인증·과금, 파일 실행 권한은 서로 다른 축이다.** 하나의 gateway나 오픈소스 라이선스로 네 문제가 한꺼번에 해결되지는 않는다.

## 2. 가장 중요한 최신 근거와 이전 안의 수정점

**OpenAI 공식 플러그인과 Codex MCP 제거 공지를 기준으로 설계를 갱신해야 한다.** 과거 설치 글을 그대로 적용하면 이미 사라진 명령에 의존하게 된다.

| 확인 대상 | 이번 관찰 | 설계에 미치는 영향 |
|---|---|---|
| OpenAI 공식 Claude Code 플러그인 | 작업·리뷰·background job·상태·결과·취소 제공. 로컬 Codex App Server 사용 | GPT 도구 위임의 우선 참조 구현 |
| 플러그인 버전 | manifest 1.0.6; 태그 페이지에서 v1.0.6, 2026-07-08, db52e28 확인 | 날짜 없는 블로그보다 재현 기준이 명확함. 태그와 main을 동일 코드라고 단정하지 않음 |
| Codex MCP 서버 | 공식 문서가 codex mcp-server 및 standalone binary 제거를 명시 | MCP 외부 인터페이스는 MoAI가 소유하고 내부는 App Server로 연결 |
| App Server 안정성 | 제거 안내는 experimental 및 production 비지원 명시 | 공식 인터페이스라도 운영 SLA·안정 API 보증으로 확대 금지 |
| 사용자 지정 기획안 | Gateway로 GPT가 Claude Code 자체 모델을 맡고 dynamicTools를 역중개 | 이 강한 요구가 있을 때만 비용과 복잡성을 감수할 가치가 있음 |

근거: [공식 플러그인](https://github.com/openai/codex-plugin-cc), [manifest](https://raw.githubusercontent.com/openai/codex-plugin-cc/main/plugins/codex/.claude-plugin/plugin.json), [태그](https://github.com/openai/codex-plugin-cc/tags), [MCP 제거 공지](https://learn.chatgpt.com/docs/mcp-server).

공식 플러그인 README에는 모델·effort 생략 시 Codex 기본값 사용, 작업 재개, 과도한 리뷰 반복에 따른 구독 소모 경고가 있다. 따라서 “구독 연결 성공”보다 **재개·취소·반복 종료 조건까지 포함한 위임 계약**이 좋은 참조점이다. 이 보고서에서는 플러그인을 설치하거나 실계정 실행하지 않았다.

## 3. 약관·인증: 무엇이 허용되는지 경로별로 판단

**공식 문서가 안내하는 경로는 존재한다. 그러나 어떤 도구든 모든 구독 토큰을 가져다 써도 된다는 뜻은 아니다.** 아래는 공개 문서에 근거한 기술·정책 분류이며 개별 계약의 법률 자문이나 무위험 보증은 아니다.

| 경로 | 공개 근거에 따른 판단 | Factory 운영 조건 |
|---|---|---|
| 본인의 Claude 구독 → 수정하지 않은 Claude Code | Anthropic 문서에 명시된 사용 경로 | 개인의 정상 사용·한도 준수. 타인 대신 구독 중개 금지 |
| 본인의 ChatGPT 구독 → 공식 Codex → Claude Code 위임 | OpenAI 공식 플러그인이 직접 안내 | 사용자별 공식 로그인, 구독 한도 준수. 무인 상용 서비스로 확대 해석 금지 |
| Claude 구독 OAuth → OMP/OpenCode 자체 추론 엔진 | 기본안에서 제외 | 오픈소스 구현의 인증 성공은 Anthropic 허가가 아님 |
| Claude Code → 비-Claude Messages Gateway | Anthropic이 지원하지 않는다고 명시 | “위반 확정”과 “공식 비지원”을 구분. 양사 공식 지원이라고 광고 불가 |
| API 키 → 허용된 모델·공급자 API | 자동화에 적합한 별도 과금 경로 | 계약·한도·데이터 처리 조건 적용 |
| 구독 토큰 공유·클라이언트 사칭·한도 우회 | 채택하지 않음 | 계정별 한도를 늘리려는 우회 설계 금지 |

Anthropic은 자사 로그인 흐름과 수정하지 않은 바이너리를 통한 사용자 본인 로그인을 구분해 허용하면서, 개발자가 Claude.ai 자격증명·세션 토큰을 수집·저장·중개하는 행위와 사용자를 대신한 구독 요청을 제한한다. 제품 배포 시 바이너리 수정·인증 방식 제한·사용량 재판매 조건도 적용된다. [Anthropic 법률·인증 안내](https://code.claude.com/docs/en/legal-and-compliance).

Claude Code에 gateway 인증 변수를 넣으면 해당 세션에서 Claude 구독 대신 gateway 자격증명이 사용된다. 반대로 BASE_URL만 설정하면 저장된 Claude 로그인이 계속 사용될 수 있다. **Claude 구독 lane과 외부 공급자 lane의 유효 인증을 각각 확인해야 한다.** [Gateway 구독·인증](https://code.claude.com/docs/en/llm-gateway).

한국 개인 사용 검토에는 OpenAI의 ROW 약관을 사용한다. 일반 terms-of-use 주소가 이번 조회에서 유럽 약관을 반환했으므로 지역을 확인했다. 계정 공유 및 한도·보호 조치 우회를 제한한다. 일반 자동 추출 제한만 떼어 공식 Codex 통합까지 모두 금지라고 해석하지 않고, 구체적인 제품 문서와 함께 읽어야 한다. [OpenAI ROW 이용약관](https://openai.com/policies/row-terms-of-use/), [Anthropic 소비자 약관](https://www.anthropic.com/legal/consumer-terms).

OpenAI 인증 문서는 프로그램 방식의 CI/CD에 API 키를 권고하고, 기업용 비대화형 작업을 위한 Codex access token도 별도로 안내한다. **개인이 지시하고 감독하는 Factory와 서비스형 무인 작업자는 인증 정책을 분리한다.** [Codex 인증](https://learn.chatgpt.com/docs/auth).

## 4. 외부 코딩 플랜 비교 — GLM 외에 무엇을 등록할까

플랜 등록은 endpoint와 키를 저장하는 데서 끝나지 않는다. 지원 도구·대화형 사용 조건·동시성·초과 과금을 함께 기록해야 한다.

| 공급자/플랜 | Claude Code 연결 근거 | Factory 적합성 | 등록 시 필수 조건 |
|---|---|---|---|
| Z.AI GLM Coding Plan | 공급자 공식 Claude Code 가이드 | 개인 코딩·지원 도구 범위의 후보. Subagent 동시 호출 안내 있음 | 지원 도구 한정, 계정 공유 금지, 동적 동시성 한도. 사용자 정의 moai-mcp 직접 호출의 허용 범위는 별도 확인 |
| MiniMax Token Plan | 공식 Claude Code 가이드 | 개인 대화형 개발 후보. 무인 production의 기본 과금으로 권고하지 않음 | 구독 키와 일반 API 키 구분. 구독 소진 후 purchased Credits 자동 사용 가능 |
| Kimi Code | 공식 Claude Code 가이드 | 대화형 lane 후보. 비대화형 배치 Factory에서는 제외 | 가이드라인이 비대화형 자동화 금지. 도구 호환성과 자동화 허용을 혼동 금지 |
| OpenCode Go | 공급자 문서에 Claude Code 검증 클라이언트 표시 | 여러 모델을 하나의 플랜으로 평가하기 좋은 후보 | 실제 client identity·안정 session ID 유지, 모델별 한도 및 초과 과금 정책 확인 |
| OpenRouter/직접 API/기업용 계약 | 각 API의 계약에 따름 | 무인·상용 자동화용 별도 경로 | Claude/GPT 개인 구독을 이식하는 방식이 아님 |

GLM 가이드는 https://api.z.ai/api/anthropic 연결을 안내하고, 사용 정책은 지원 도구·개인 사용 및 등급별 동적 동시성 조건을 명시한다. MCP로 감쌌다는 이유만으로 지원 도구 범위가 넓어지지는 않는다. **GLM 직접 추론 도구에는 일반 API 자격증명을 기본 후보로 두고, Coding Plan은 MoAI 호출 형태의 허용 여부를 확인한 뒤 켠다.** [Z.AI 연결](https://docs.z.ai/devpack/tool/claude), [사용 정책](https://docs.z.ai/devpack/usage-policy).

MiniMax는 구독 키와 종량제 API 키가 다르며, 여러 도구가 한 구독 한도를 공유한다. 구매한 Credits가 있으면 구독 소진 후 자동 사용될 수 있어 “같은 키니까 구독 내 사용”이라는 판정은 틀릴 수 있다. 개인 대화형 개발용이며 production에는 종량제를 권고한다. [MiniMax 연결](https://platform.minimax.io/docs/token-plan/claude-code), [FAQ](https://platform.minimax.io/docs/token-plan/faq).

Kimi는 Claude Code 연결을 문서화하지만, 구독을 개인 대화형 용도로 한정하고 스크립트 배치 실행을 금지한다. 가이드에 포함된 로그인 생략·홈 설정 변경 스크립트를 MoAI가 그대로 실행하는 방식도 권고하지 않는다. [연결 가이드](https://www.kimi.com/code/docs/en/third-party-tools/claude-code.html), [사용 가이드라인](https://www.kimi.com/code/docs/en/kimi-code/community-guidelines.html).

OpenCode Go는 조회 시 월 $10으로 표시되고, Claude Code의 native session header를 인식한다고 안내한다. OMP/OpenCode 전체 엔진을 도입하지 않고 **공급자 플랜만 추가하는 대안**이다. 가격은 조회 시 표시값이며 품질·월 처리량을 MoAI에서 측정한 결과가 아니다. [OpenCode Go](https://opencode.ai/docs/go/).

## 5. OMP·OpenCode·공식 플러그인·MoAI MCP 비교

**OMP를 oh-my-pi로 해석하여 조사했다.** 다른 프로젝트를 의미했다면 해당 항목만 추가 조사하면 된다. 비교 대상의 오픈소스 라이선스는 코드 사용 조건이며 모델 서비스 이용 권한과 별개다.

| 대안 | Claude Code 유지 | 모델·effort 조율 | 구독 경로 | 판단 |
|---|---|---|---|---|
| 원안: Messages Gateway + App Server | GPT가 CC 자체 모델을 맡는 목표 | 네 슬롯 및 요청별 변환 필요 | Codex 공식 인증 사용 가능, CC 비-Claude 비지원 | 강한 동일 실행 루프 요구에 한해 선택 |
| OpenAI 공식 CC 플러그인 | CC에서 명령·에이전트로 위임 | 모델·effort·재개·background | ChatGPT 구독 명시 | 가장 명확한 참조 구현·비교 기준 |
| 기존 moai-mcp 확장 | CC 에이전트 도구로 통합 | Factory 정책과 일원화 가능. 현재 부족한 필드 있음 | GPT 공식 실행기 + 공급자별 인증 | **MoAI의 장기 기본안으로 권고** |
| OMP 전체 실행기 | 자체 TUI는 조건 불충족; RPC를 MCP로 감싸면 위임 가능 | 역할별 모델 및 RPC 제공 | OAuth·플랜을 광고하나 공급자 허용 별도 | 다중 공급자 구현 참고, 기본 의존성으로는 보류 |
| OpenCode 전체 실행기 | 자체 TUI는 조건 불충족; server를 도구로 감쌀 수 있음 | provider/model/agent 구조 활용 가능 | ChatGPT 로그인 안내; Claude 구독 플러그인 제외 | 필요할 때 추가 실행기로 검토 |
| API Gateway만 추가 | 호출 endpoint 통일 가능 | 요청 변환 가능 | 구독 권한을 만들어 주지는 않음 | 세션·도구 실행·재개의 대체물 아님 |

OMP README는 역할별 모델, SDK, RPC, ACP를 공개한다. Claude/GPT OAuth와 여러 coding plan 지원을 표시하지만 이번 조사에서 OMP용 Claude 구독 사용을 허용한다는 Anthropic 근거는 확인하지 못했다. OMP가 자신의 도구 루프를 실행하면 Claude Code hook이 그 내부 동작마다 실행되는 것도 아니다. [OMP 원 저장소](https://github.com/can1357/oh-my-pi).

OpenCode는 ChatGPT Plus/Pro 로그인을 안내한다. Claude Pro/Max 플러그인은 Anthropic이 금지하며 1.3.0부터 기본 포함하지 않는다고 적는다. 같은 페이지에 과거 Claude 로그인 설명이 남아 있어, 이를 허용 근거로 삼지 않는다. OpenCode의 구현 안내를 OpenAI가 모든 제3자 클라이언트에 부여한 포괄 허가로 확대하지 않는다. [OpenCode 공급자 문서](https://opencode.ai/docs/providers), [server 통합](https://opencode.ai/docs/server/).

## 6. moai-mcp는 이미 어디까지 구현돼 있나

이번 조사는 지정 기획안이 있는 develop 트리에서 진행했다. **GPT·GLM 도구는 이미 존재한다. 새 서버를 만드는 것보다 현재 계약을 보완하는 것이 우선이다.** 아래는 소스·선택 테스트 범위의 관찰이며 설치된 MCP 서버의 실계정 성공을 뜻하지 않는다.

| 현재 표면 | 확인한 동작/스키마 | 목표와의 차이 |
|---|---|---|
| codex_task | prompt, background, write, resume_last; 공식 app-server 세션 실행 | 도구 입력에 명시 model·effort·lane/thread selector 없음 |
| codex_job_status/result/cancel | 작업 조회·결과·취소 도구 등록 | background 실행은 MCP 프로세스 수명에 묶인다고 설명 |
| glm_task | prompt, background, model, system, max_tokens | 명시 effort 입력 없음. 현재 요청은 단일 user 메시지 중심의 텍스트 추론 |
| Factory 안의 glm_task | 호출자가 넘긴 model을 무시하고 세션 정책의 기본값 사용 | 임의 모델 변경 허용은 현재 테스트된 계약과 충돌 |
| codex_audit / glm_audit | 독립 리뷰 경로와 공통 결과 형태 | task와 audit의 effort 처리를 동일하다고 가정하면 안 됨 |
| resume_last | 프로젝트에 마지막 기록된 Codex thread 조회 | 다중 lane은 명시 owner/thread binding 계약을 별도로 갖춰야 함 |

직접 근거: [MCP 등록: mcp_server.go](../../internal/cli/mcp_server.go) 304–399행, [Codex task](../../internal/cli/codex_task.go) 200–253행, [GLM task](../../internal/cli/glm_task.go) 138–177·294–305행, [GLM Factory 테스트](../../internal/cli/glm_task_test.go) 560행 이후.

callGLMTask의 관찰한 요청 생성부는 Model·MaxTokens·System·Messages를 채운다. 이 경로에서 호출자별 effort를 지정하는 입력·대입은 확인되지 않았다. glm_audit의 reasoning_effort 지원을 glm_task의 지원으로 잘못 표시하면 안 된다. 이것은 **새 요구 대비 계약 차이**이며 별도 실계정 effort 비교 실험을 마친 결함 판정은 아니다.

현재 Gateway 선택 테스트는 GPT main을 Astra로 지정할 때 네 슬롯 모두 Astra가 되는 기존 동작을 검증한다. 기존 보고서의 사용자 지정 매핑 fable→Astra / opus→Sol / sonnet→Terra / haiku→Luna와 같은 의미가 아니다. 이번 테스트 PASS를 새 매핑 구현 완료로 읽으면 안 된다. [기존 슬롯 테스트](../../internal/cli/gateway_provider_contract_test.go) 25–71행.

## 7. 권고 아키텍처: 한 Factory, 두 실행 경로

모든 작업의 접점은 Claude Code에 둔다. **MCP는 전문 작업 위임, Gateway는 세션 자체 모델 교체**를 담당한다. 이 구분으로 기존 moai cc·gpt·glm의 목적을 보존하면서 필요한 부분만 확장한다.

1. Claude Code lead가 작업·완료 조건·소유 worktree·허용 공급자를 확정한다.
2. 각 Claude Code lane은 작업별 전문 에이전트를 실행하고 moai-mcp로 GPT·GLM 작업을 요청한다.
3. MoAI가 선택한 provider/account/model/effort와 사용 조건을 검증하고 실행을 접수한다.
4. GPT는 공식 Codex 실행부, GLM 등은 허용된 API 또는 검증한 세션 경로로 보낸다.
5. 작업은 job ID를 즉시 반환하고 상태·결과·취소는 별도로 조회한다.
6. 산출물과 검증 근거를 lane이 확인한 후 lead가 다음 작업을 배정한다.

GPT가 계획을 만들어도 Claude lead가 결정·배정을 집행할 수 있다. **GPT가 기술적 설계 역할을 맡는 것과 GPT가 Claude Code의 주 모델이 되는 것은 별개**다. 후자가 필수인 사용자는 기존 Gateway 경로를 선택한다.

MCP 도구 승인은 외부 실행기 전체를 호출하는 승인이다. 그 안에서 Codex가 수행하는 개별 shell/patch가 Claude Code의 PreToolUse·PostToolUse로 자동 감싸지는 것은 아니다. 따라서 두 권한 모드를 명확히 제안한다.

| 모드 | 실행 주체 | 적합한 작업 |
|---|---|---|
| 조언·패치 제안 | 외부 LLM은 읽기·분석/제안, 실제 변경은 Claude Code | 초기 기본값. 설계·검토·리팩터링 제안 |
| 격리 작업 위임 | Codex가 지정 worktree에서 자체 sandbox로 변경·검증 | 구현 lane. 별도 쓰기 허용 및 산출물 readback 필수 |

공식 MCP는 도구 연결 표준이며 모델 호출권이나 파일 sandbox를 자동 부여하지 않는다. [Claude Code의 local stdio MCP](https://code.claude.com/docs/en/mcp).

## 8. 모델·effort·능력 등록의 최소 계약

아래는 **제안 데이터 계약**이다. 현재 CLI에 그대로 붙여 넣어 동작하는 설정이 아니다. 새 중간 모델 등급을 만들지 않고 기존 네 별칭과 실제 모델 ID를 유지한다.

```yaml
execution:
  owner: {run_id: required, lane_id: required, task_id: required}
  provider: openai
  transport: codex-app-server
  auth_profile_ref: user-owned-profile
  model: gpt-5.6-sol
  effort: high
  thread_ref: explicit-owned-thread-or-new
  capability: reasoning
  permissions: read-only
  billing: subscription-only
```

현재 codex_task에 새 gpt_task를 중복 구현하기보다 기존 실행 함수를 재사용한다. 도구 이름 변경 여부는 호환 소비자 조사 후 결정한다. model·effort·owner·명시 thread를 확장하고, 결과에 요청값과 실제 적용값을 함께 남긴다. Factory 안의 GLM model override 무시는 무작정 제거하지 않는다. **lead가 서명·확정한 작업별 선택만 허용하는 새 계약과 테스트로 교체**한다.

모델 목록과 허용 effort는 공급자 catalog에서 확인한다. Codex App Server의 model/list는 supportedReasoningEfforts 등을 제공한다. 미지원 effort를 임의 하향하거나 다른 모델로 바꾸지 않는다. dynamicTools는 별도의 experimental 기능이므로 MCP 작업 위임과 동일시하지 않는다. [App Server 모델·도구 계약](https://learn.chatgpt.com/docs/app-server).

추천 배정은 성능 보증이 아니라 평가할 초기 가설이다.

| 작업 | 초기 배정 가설 | 실제 선정 기준 |
|---|---|---|
| 어려운 추론·설계 | GPT 상위 모델 또는 Claude 상위 모델 | 요구 충족률·독립 검증·재작업량 |
| 일반 구현 | GPT/Claude 중 해당 저장소 성과가 좋은 조합 | 테스트 통과와 수정 범위 |
| 반복 분석·단순 변환 | GLM 또는 허용 코딩 플랜 | 성공한 작업당 비용·실패 후 재시도 포함 |
| 이미지 제작 | Codex image generation 별도 능력 | 이미지 파일·크기·MIME·요구 충족 확인 |
| 최종 검토 | 작성자와 독립된 모델/세션 | 지적의 재현성과 실제 결함 탐지율 |

provider, model, effort, image_input, image_generation을 별도 필드로 둔다. 이미지 입력을 받는 모델이라는 이유만으로 이미지를 생성할 수 있다고 표시하지 않는다. 공식 문서는 내장 이미지 생성의 구독 사용을 설명하지만, 현재 moai-mcp가 해당 이미지 결과를 내보내는지는 미검증이다. [공식 이미지 생성 안내](https://learn.chatgpt.com/docs/image-generation).

## 9. 최신 사례: 성공 증거의 수준을 구분한 분석

검색 결과의 홍보 문구보다 원 저장소·공식 공지·실험 방법을 우선했다. 아래의 성공은 저자 또는 공급자가 공개한 관찰이며 이번 MoAI 환경에서 재현한 결과는 아니다.

| 사례와 시점 | 공개 증거 | 가져올 패턴 | 적용 한계 |
|---|---|---|---|
| OpenAI codex-plugin-cc · 2026-07-08 태그 확인 | 공식 저장소·1.0.6 manifest·작업 위임 문서 | 공식 실행기·background job·명시 모델/effort | MoAI lane/owner 계약과 통합된 제품은 아님 |
| codex-on-claude · 2026-05-20 실험 | Claude Max + ChatGPT auth, CC 2.1.144/Codex 0.131.0, T1/T2/T3/T5 성공 보고 | 새 세션의 에이전트/MCP 자동 발견·짧은 결과 전달 | T4 cross-process thread 재개 실패. 현재 제거된 native MCP 기반의 역사적 사례 |
| mcp-agents · 2026-09-14 현재 문서 조회 | 자체 MCP 위에 App Server, jobs·leases·재개 설계 공개 | MCP 계약 유지 + 실행 transport 교체 | README 기능 주장, 이번 실행 재현 없음. legacy 설명은 공식 제거 공지보다 약한 근거 |
| 교차 모델 리뷰 연구 · 2026-07-22 | 116개 LCB 과제의 통제 실험 | 모델 조합의 순서도 측정해야 함 | 알고리즘 과제·리뷰어 테스트 실행 불가 조건; MoAI 최신 모델 전체로 일반화 금지 |

근거: [공식 플러그인 태그](https://github.com/openai/codex-plugin-cc/tags), [codex-on-claude 실험 보고](https://github.com/pathcosmos/codex-on-claude/blob/main/docs/test-report-2026-05-20.md), [mcp-agents](https://github.com/thomaswitt/mcp-agents), [교차 리뷰 연구](https://arxiv.org/abs/2607.21656).

교차 리뷰 연구는 Codex 초안을 Claude가 검토했을 때 71.6%→89.7%, 반대 방향은 91.4%→82.8%라고 보고한다. “다른 모델을 붙이면 무조건 좋아진다”는 가정에 반하는 사례다. 이 수치는 연구자의 조건에서 나온 값이며 MoAI의 기대 성능이나 최신 모델 순위가 아니다. 논문 초록·방법 요약을 확인했으며 데이터셋 재현은 하지 않았다.

**최신성이 가장 중요한 교훈:** 오래된 codex mcp-server 성공 사례는 설계 아이디어의 증거이지 지금 사용할 명령의 근거가 아니다. 현재 공식 제거 안내와 설치된 CLI 0.154.0의 help를 함께 확인했다.

## 10. Factory 운영에 적용할 실무 원칙

다음은 위 사례와 현재 MoAI 계약에서 도출한 설계 권고다. 구현된 안전장치라고 주장하지 않는다.

- **작업 단위로 위임한다.** 매 메시지마다 모델을 돌려 가며 전체 이력을 보내기보다 입력·산출물·완료 조건이 있는 작업을 맡긴다. 이력 변환과 중복 컨텍스트 비용을 줄일 수 있다는 설계상 기대이며 절감률은 미측정이다.
- **명시 thread로 재개한다.** 프로젝트의 “마지막 thread” 대신 run/lane/task/account와 연결된 thread를 사용한다. 모델 변경은 작업 경계에서 처리한다.
- **한 작업에는 한 쓰기 주체를 둔다.** Claude와 Codex가 같은 파일을 동시에 바꾸지 않도록 worktree와 소유권을 부여한다.
- **start/status/result/cancel을 유지한다.** 장기 작업을 단일 MCP 응답 시간에 묶지 않는다. 프로세스 종료 후 상태 복구는 별도 기능으로 검증한다.
- **전송 성공과 업무 성공을 구분한다.** 구조화된 failed/inconclusive가 정상 MCP 응답으로 왔다고 Factory가 완료 처리하면 안 된다.
- **구독 한도는 계정별로 합산한다.** lane 수가 늘어도 구독 한도가 늘지는 않는다. 429에는 대기·동시성 축소·명시 실패를 적용한다.
- **provider 변경과 과금 변경을 구분한다.** 구독 소진 후 API·Credits로 전환하는 정책은 별도 설정과 실제 과금 상태 확인이 필요하다.
- **리뷰 반복의 종료 조건을 둔다.** 재현 가능한 지적과 수정 근거가 없으면 무한 재검토를 종료한다. 최대 반복·사용량 한도는 운영자가 정한다.
- **도구별 공개 범위를 좁힌다.** 설계 에이전트에는 읽기 분석 도구, 구현 에이전트에는 허용된 작업 도구만 연결한다. 환경의 비밀값은 prompt나 결과에 싣지 않는다.
- **스킬은 사용 방법, 실행기는 강제 조건을 담당한다.** LLM에게 “소유권을 확인하라”라고 쓰는 것만으로 lock·중복 방지·결과 저장을 대체하지 않는다.

## 11. 최소 구현 순서와 인수 기준

제품 코드는 이번 조사에서 변경하지 않았다. 아래 단계는 후속 구현 계획이며 시간 예측이나 전체 재설계 승인을 전제하지 않는다.

| 우선순위/단계 | 변경 범위 | 통과 기준 |
|---|---|---|
| High · 현재 연결 확인 | 기존 moai MCP tools/list, 설치된 바이너리 provenance, 자격증명 종류 | 도구 등록·현재 모델·과금 종류 확인; 비밀값 출력 없음 |
| High · 읽기 전용 수직 검증 | Claude agent → codex_task / glm_task → 결과 | 지정 model·effort의 실제 적용과 오류 전달 확인 |
| High · Factory 식별 확장 | 명시 owner/thread + 모델 선택 정책 | lane 간 잘못된 결과·재개 차단을 음성 테스트로 입증 |
| High · job 복구 | 저장 상태와 process liveness 재조정 | MCP 재접속·취소·중복 요청에서 거짓 완료 없음 |
| Medium · 공급자 등록 | GLM/API·MiniMax·OpenCode Go 등 | 허용 사용 방식·endpoint·한도·초과 과금 검증 |
| Medium · 격리 쓰기 | Codex 전용 worktree와 실행 권한 | 허용 파일만 변경, Claude hook과 Codex sandbox 경계 명시 |
| Medium · 이미지 능력 | 별도 결과 artifact 계약 | 생성 성공 event와 실제 이미지 파일 readback |
| Low · Gateway 확대 | GPT 자체 주 모델이 필요한 lane | tool round-trip·compact·resume·취소·effort 변환·한도 인수 |

평가는 MoAI 실제 작업 묶음으로 단일 모델, MCP 분업, Gateway 분업을 비교한다. 동일한 시작 commit과 완료 기준을 사용하고 성공률·재작업·사용량·응답 시간·한도 대기를 기록한다. **저렴한 토큰 가격보다 검증된 작업 하나를 완료하는 총비용**을 기준으로 배정한다.

## 12. 최종 선택과 적용 범위

**지금의 최선: 기존 moai-mcp를 공통 작업 접점으로 완성하고, GPT는 공식 Codex 실행부로 연결한다.** Claude 구독은 Claude Code가 직접 관리한다. GLM 등은 사용 형태가 허용된 공급자 자격증명으로 연결한다. 공식 OpenAI 플러그인은 계약과 동작을 비교하는 기준으로 사용하고, MoAI의 Go 실행부와 중복 job 관리자·중복 자동 리뷰를 동시에 켜지 않는다.

OMP/OpenCode 전체를 도입하는 결정을 먼저 내릴 이유는 현재 확인한 범위에서 약하다. MoAI에는 이미 Factory·MCP task·job·Codex App Server 코드가 있기 때문이다. 다양한 외부 플랜 추가가 목적이라면 **실행기 교체보다 공급자 등록 계약부터 확장**하는 편이 기존 목적에 가깝다.

Gateway는 버리지 않는다. “각 Claude Code 세션의 주 모델을 GPT·GLM으로 바꾼다”는 기존 목적을 담당하는 선택 경로다. 다만 이를 약관·공식 지원 문제가 모두 해소된 유일한 기본안으로 삼지는 않는다. MCP 위임을 허용한 이번 추가 요구 덕분에 더 명확한 공식 선례를 가진 기본 경로가 생겼다.

## 13. Claim · 이번에 주장하는 범위

- GPT·GLM 도구 위임 구조에 공식/공개 선례가 있으며, 현재 develop의 MoAI MCP에 관련 task/job 도구가 존재한다.
- 선택한 기존 계약 테스트가 이번 트리에서 통과했다. 특히 GLM Factory의 model override 무시 동작을 확인했다.
- 공식 구독 사용 경로와 비-Claude Gateway 지원 경계, 공급자별 자동화 제한을 비교했다.
- 새 아키텍처 구현·실계정 모델 조율·Factory 운영 성공·약관 무위험을 주장하지 않는다.

## 14. Evidence · 실행 명령과 관찰 출력

기준 확인 명령:

```text
git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop rev-parse HEAD
4056f69e1c20d942d4f9fc7363d3d79bffde899a

codex --version
WARNING: proceeding, even though we could not create PATH aliases: Operation not permitted (os error 1)
codex-cli 0.154.0
```

선택 테스트 A. 작업 디렉터리는 아래 Baseline-attribution의 develop 트리다.

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-factory-research-20260914-cache go test ./internal/cli -run '^(TestGatewayProviderContractPickerAndSlots|TestGatewayProviderContractRejectsForeignModel|TestParseFactoryFlag)$' -count=1 -timeout 90s
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	0.965s
```

선택 테스트 B:

```sh
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN OPENAI_API_KEY Z_AI_API_KEY && GOCACHE=/tmp/moai-factory-research-20260914-cache go test ./internal/cli -run '^(TestCodexTools_Registered|TestCodexTask_ForegroundReturnsOutput|TestCodexTask_WriteGateFailClosed|TestCodexTask_ResumeLastReusesRecordedThread|TestGLMTask_SyncReturnsOutput|TestGLMTask_ForwardsSystemModelAndMaxTokens|TestGLMTaskFactoryModeIgnoresModelOverride)$' -count=1 -v -timeout 90s
```

```text
=== RUN   TestCodexTask_ForegroundReturnsOutput
--- PASS: TestCodexTask_ForegroundReturnsOutput (0.00s)
=== RUN   TestCodexTask_WriteGateFailClosed
=== RUN   TestCodexTask_WriteGateFailClosed/absent-file
=== RUN   TestCodexTask_WriteGateFailClosed/absent-key
=== RUN   TestCodexTask_WriteGateFailClosed/malformed-yaml
=== RUN   TestCodexTask_WriteGateFailClosed/explicit-false
=== RUN   TestCodexTask_WriteGateFailClosed/explicit-true
--- PASS: TestCodexTask_WriteGateFailClosed (0.00s)
    --- PASS: TestCodexTask_WriteGateFailClosed/absent-file (0.00s)
    --- PASS: TestCodexTask_WriteGateFailClosed/absent-key (0.00s)
    --- PASS: TestCodexTask_WriteGateFailClosed/malformed-yaml (0.00s)
    --- PASS: TestCodexTask_WriteGateFailClosed/explicit-false (0.00s)
    --- PASS: TestCodexTask_WriteGateFailClosed/explicit-true (0.00s)
=== RUN   TestCodexTask_ResumeLastReusesRecordedThread
--- PASS: TestCodexTask_ResumeLastReusesRecordedThread (0.00s)
=== RUN   TestGLMTask_SyncReturnsOutput
--- PASS: TestGLMTask_SyncReturnsOutput (0.00s)
=== RUN   TestGLMTask_ForwardsSystemModelAndMaxTokens
--- PASS: TestGLMTask_ForwardsSystemModelAndMaxTokens (0.00s)
=== RUN   TestGLMTaskFactoryModeIgnoresModelOverride
=== RUN   TestGLMTaskFactoryModeIgnoresModelOverride/factory_mode_ignores_the_override
=== RUN   TestGLMTaskFactoryModeIgnoresModelOverride/outside_factory_mode_the_override_is_honored
--- PASS: TestGLMTaskFactoryModeIgnoresModelOverride (0.00s)
    --- PASS: TestGLMTaskFactoryModeIgnoresModelOverride/factory_mode_ignores_the_override (0.00s)
    --- PASS: TestGLMTaskFactoryModeIgnoresModelOverride/outside_factory_mode_the_override_is_honored (0.00s)
=== RUN   TestCodexTools_Registered
--- PASS: TestCodexTools_Registered (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.968s
```

웹 검증은 본문의 직접 링크를 web open/find로 조회했다. 핵심 공식 원문 발췌:

| 출처 | 조회한 짧은 원문 | 범위 |
|---|---|---|
| OpenAI 공식 CC 플러그인 | ChatGPT subscription (incl. Free) or OpenAI API key. | README Requirements |
| Codex MCP 제거 | The app-server command is experimental and isn't supported for production workloads. | 공식 제거 안내 |
| Kimi 가이드라인 | Kimi Code subscriptions are for interactive use only. | Scope of Use |
| Z.AI 정책 | GLM Coding Plan may only be used within officially supported tools and products. | Account Usage Policy |

## 15. Baseline-attribution · 측정 기준

소스 및 테스트: /Users/goos/MoAI/moai-adk-go/.claude/worktrees/develop, HEAD 4056f69e1c20d942d4f9fc7363d3d79bffde899a. 테스트 전후 HEAD를 확인했다. 기존 SPEC의 수정 및 여러 untracked 보고서가 존재하는 dirty working tree 기준이며, commit만의 깨끗한 checkout 검증으로 표현하지 않는다. 기존 제품 파일은 수정하지 않았다.

원 기획안: [gpt-final-design report](../gpt-final-design-20260914-b40ff9f4/report.html), 동반 report.md를 읽었다. 그 문서의 과거 테스트 1.000s와 명칭 전수조사 수치는 이번 측정값으로 재사용하지 않았다.

문서 조회: 2026-09-14. 공개 페이지의 조회 시점과 사례의 실행·발행 날짜를 구분했다. OpenAI 공식 플러그인 태그, 커뮤니티의 5월 실험, 7월 논문은 서로 다른 baseline이다. 본 세션 ID: 01a09dfb-5908-7e50-ac3a-bde14d9af79f.

## 16. Gaps · 확인하지 않은 범위

- Claude/GPT 구독 실계정 호출, 이미지 생성, GLM Coding Plan의 MoAI MCP 사용 승인, 실제 과금 readback은 수행하지 않았다.
- 설치된 moai MCP 서버의 현재 tools/list, 프로세스 바이너리 일치, live model/effort, 다중 lane 재개는 미확인이다.
- OMP·OpenCode·공식 플러그인을 설치하거나 각 저장소 전체 테스트를 실행하지 않았다. 커뮤니티 성공 보고는 저자 보고다.
- OMP OAuth 세부 소스의 시도한 raw 경로와 docs 사이트 본문을 확보하지 못했다. 인증 지원 평가는 README 광고 수준으로 한정했다.
- 교차 리뷰 논문은 초록과 공개 방법 요약을 사용했다. 전체 실험 재현·최신 모델 성능 비교는 하지 않았다.
- 일반적인 공급자 목록 전수조사는 아니다. 사용자 지정 OMP/OpenCode와 대표 coding plan, 직접 관련 공식/커뮤니티 사례를 조사했다.
- HTML 스킬이 참조하는 design-tokens.md는 현재 경로에 없었다. 스킬 본문과 제공 plan 템플릿의 토큰으로 렌더링했다.

## 17. Residual-risk · 남는 위험

공식 플러그인의 존재는 모든 무인 Factory·구독 중개·다중 사용자 서비스에 대한 허가가 아니다. 정책·모델·가격·실험 API는 바뀔 수 있다. 현재 task 도구의 fail-open·프로세스 수명·프로젝트 단위 재개 계약을 Factory 완료 판정과 그대로 결합하면 요구를 충족하지 못할 수 있다. 실제 결함과 성능은 후속 수직 검증에서 판단해야 한다.

**권고의 신뢰 수준:** 공식 인증·통합 선례와 현행 소스 계약은 근거 있음. MCP 중심 아키텍처의 우월성은 요구 적합성과 재사용 범위에 따른 설계 판단. 절감 비용·처리량·실서비스 안정성은 미측정이다.

# MoAI Hybrid Factory 재설계

2026-09-14 · 브레인스토밍을 반영한 설계 제안 · 제품 구현·설정 변경 없음

## 1. 결정 요약 — 역할은 유지하고 실행 모델을 선택한다

권고는 하나의 공통 실행 정책에 두 연결 방식을 지원하는 Hybrid Factory다. 사용자는 기존 moai cc, moai glm, moai gpt의 주 모델 선택을 유지하면서, 각 역할·lane의 작업 모델은 별도로 선택한다. Claude main + GPT·GLM 전문 작업을 기본 사용 경험으로 제공하되 Gateway 사용자를 강제로 전환하지 않는다.

- 메인 모델: 사용자가 대화하는 Claude Code 세션의 모델. native Claude 또는 Gateway 모델.
- 역할: plan, run, review, image처럼 수행할 일. 공급자 이름과 분리한다.
- 실행 경로: Claude Code 자체 실행, MCP를 통한 외부 실행기 위임, 별도 Gateway lane 중 선택한다.
- 실행 정책: provider, model, effort, 인증·과금, 쓰기 범위, 동시성, 재시도 상한을 함께 결정한다.

이 보고서는 개인이 감독하는 로컬 개발을 기본 사용 시나리오로 제안한다. 무인 운영·다중 사용자 서비스는 별도의 허용 API·계약을 요구한다. 품질 향상·절감률은 아직 측정하지 않은 목표다.

## 2. 3단계 sub-agent는 가능한가

가능하다. 최신 Claude Code 공식 문서는 메인 아래 기본 3단계 중첩과 CLAUDE_CODE_MAX_SUBAGENT_SPAWN_DEPTH 설정을 설명한다. v2.1.219부터 기본 3단계이며, 로컬 claude --version 결과는 2.1.270이다. 다만 실제 3단계 생성은 이번에 실행하지 않았다. 검색 결과에 남은 예전의 중첩 불가 설명보다 현재 본문을 기준으로 판단했다. [공식 중첩 설명](https://code.claude.com/docs/en/sub-agents#let-subagents-spawn-their-own-subagents).

중첩 가능성과 GPT 사용은 별개다. Claude subagent 이름을 gpt-agent-plan으로 정해도 실제 추론이 GPT로 바뀌지는 않는다. native Claude 세션에서는 해당 subagent가 Claude로 실행되어 MCP의 GPT 작업을 요청하는 중개 역할이 된다. Gateway 모델 매핑은 별도의 경로다.

현재 MoAI에는 일반 retained agent의 Agent 도구 제한과 manager-lead의 leaf-worker 예외가 있다. 기존 규칙에는 전면 flat 설명도 남아 있어, 새 중첩 예외 도입 전 정합성 검토가 필요하다. 이 보고서는 규칙을 임의 변경하지 않는다. [manager-lead 정의](../../../../../.claude/agents/moai/manager-lead.md), [에이전트 작성 규칙](../../../../../.claude/rules/moai/development/agent-authoring.md).

권고: 기본은 얕게, 필요할 때만 중첩한다. 메인 → 역할 담당 → 전문 실행은 논리적 3단계이며 마지막 MCP 작업은 Claude native subagent 깊이에 포함되는 Agent 호출이 아니다. 별도 전역 작업 수 제한이 필요하다. plan → run → review는 중첩 트리로 만들기보다 동일 소유자가 순서대로 실행하는 단계 흐름으로 만든다.

## 3. 비교한 다섯 가지 구성

| 구성 | 장점 | 비용·제약 | 권고 |
|---|---|---|---|
| 메인 모델이 모든 작업 수행 | 설정·문맥 전달이 단순 | 모든 작업에 같은 모델 비용 | 짧고 단순한 변경에 유지 |
| 메인에서 MCP 직접 호출 | Claude 중간 관리자 호출을 생략 | 메인이 외부 결과를 처리 | 짧은 질문·독립 리뷰·이미지 요청 기본 |
| Claude 역할 subagent → MCP | 문맥 격리, 역할별 권한·근거 검증 | Claude wrapper와 외부 실행 양쪽 비용 | 긴 설계·독립 구현의 기본 |
| 모든 역할을 3단계 중첩 | 복잡한 병렬 작업 분해 | 관리자 증가, 결과 중복, 동시성 팽창 | 기본값으로 채택하지 않음 |
| lane별 Gateway 주 모델 | lane의 전체 개발 루프를 해당 모델이 수행 | 프로토콜·인증·history 호환 검증 | 기존 모드 유지, 고급 사용자 선택 |

최적안은 두 번째·세 번째를 작업 크기에 따라 쓰고 다섯 번째를 함께 지원하는 것이다. 단순 lint 실패 수정마다 plan/run/review 전문 에이전트를 모두 만들지 않는다. 반대로 결제·인증 변경은 비용 최적화 설정이어도 검증을 줄이지 않는다.

## 4. 사용 경험 — 기존 진입점, 공통 프로필

아래는 제안 UX다. 실제 존재하는 명령 옵션이라고 주장하지 않는다. 기존 명령의 기본 동작은 유지하고 사용자가 명시적으로 프로필을 선택했을 때만 역할별 분업을 켠다.

| 진입 방식 | 메인 기본값 유지 | 추가 선택 가능한 작업 방식 |
|---|---|---|
| moai cc | Claude | GPT 설계·구현 위임, GLM 텍스트 작업, GPT 이미지 |
| moai glm | 기존 GLM 연결 | 복잡한 단계만 GPT 위임, 이미지 작업 별도 |
| moai gpt | 기존 GPT Gateway 연결 | GLM에 저위험 작업, GPT 이미지 별도 |
| Factory -f | 선택한 lead, 각 lane 세션 | lane 기본 경로 + 역할별 예외 정책 |

사용자 요청의 moai -f는 Factory 사용 경험으로 해석한다. 현재 문서에서 확인한 형태는 moai cc -f N이며, top-level moai -f 별칭 여부는 이번에 검증하지 않았다.

초기 화면은 네 가지만 묻는다: 메인 모델, 작업 프로필, 추가 API 과금 허용 여부, 동시 작업 상한. 상세 model·effort는 펼쳐서 수정한다. 프로필은 프로젝트에 저장하고 세션·lane별 임시 덮어쓰기를 제공한다. 같은 터미널 전역 환경을 덮어쓰지 않는다.

- 절약형: 짧은 작업은 직접 처리, 검증된 저비용 모델에 경계가 명확한 작업 위임, 실패 시 승격.
- 균형형: 복잡한 설계에 상위 추론, 구현은 중간 단계, 위험한 변경만 독립 검토. 권장 초기값.
- 품질형: 고위험 설계·구현을 상위 모델로, 독립 검증을 추가하되 무제한 토론은 금지.
- 수동형: 공급자·모델·effort·실행 경로를 역할별 지정, 자동 변경 없음.

실행 전 미리보기에는 요청 모델과 실제 모델, 인증 종류, 추가 과금 가능성, 쓰기 권한, 선택 이유를 표시한다. 구독 잔여량을 공식적으로 조회할 수 없으면 미확인으로 표기한다. 토큰 수를 구독 잔여량으로 환산해 꾸며내지 않는다.

## 5. gpt-agent-plan / run을 어떻게 만들 것인가

사용자에게는 gpt-agent-plan, gpt-agent-run, glm-agent-run 같은 친숙한 이름을 제공할 수 있다. 내부적으로는 역할 템플릿 하나와 실행 프로필을 합성해야 한다. 공급자 × 역할 × effort마다 독립 프롬프트 파일을 복제하지 않는다.

| 사용자 이름 | 내부 역할 | 실제 추론 및 실행 | 반환 계약 |
|---|---|---|---|
| gpt-agent-plan | plan | 필요 시 Claude wrapper → Codex | 설계안, 전제, 위험, 검증 조건 |
| gpt-agent-run | run | Codex의 제한된 독립 작업 | 기준 commit, diff, 실행 명령·출력, 미완료 |
| glm-agent-run | bounded implementation | 지원 API 또는 GLM Gateway lane | 변경안 또는 검증된 diff; 실행 근거 구분 |
| cross-review | review | 작성자와 독립된 검토 실행 | 구체적 결함·재현·잔여 위험 |
| gpt-image | image | 검증된 Codex 이미지 경로 또는 Images API | 이미지 파일과 메타데이터 |

기존 manager-spec·manager-develop·sync-auditor 등의 책임을 대체하는 병렬 조직을 만들지 않는다. 역할 소유자는 기존 에이전트이고 외부 모델은 그 역할이 선택한 실행기다. SPEC 작성 권한과 코드 작성 권한을 혼합하지 않는다. 새로운 agent 파일·frontmatter는 이 보고서에서 생성하지 않는다.

Claude wrapper는 짧게 요청을 정리하고 결과의 형식·근거를 확인한다. GPT가 이미 읽은 전체 저장소를 wrapper가 다시 읽거나, 받은 구현을 똑같이 다시 작성하지 않는다. 간단한 위임은 wrapper 자체를 생략한다.

## 6. 공통 실행 계약 — 한 번 정하고 모든 모드에서 적용

새 제어 서비스를 별도로 만드는 대신 기존 MoAI의 모델 해석기·작업 레지스트리·MCP를 확장한다. 아래 필드는 제안이며 현재 schema가 아니다.

```yaml
role: run
route: mcp
provider: openai
model: user_selected_model
effort: user_selected_supported_effort
auth_ref: local_official_codex_login
owner: {session_id: required, lane_id: optional, task_id: required}
workspace: {worktree: explicit_path, base_commit: required}
permissions: {write_scope: assigned_paths, network: policy}
limits: {max_attempts: 2, concurrent_jobs: 1, api_spend_cap: explicit}
fallback: ask_before_billing_or_provider_change
```

정책 우선순위는 강제 보안·과금 제약 → 사용자 명시 선택 → task → agent role → lane → session → project 기본값이다. 강제 제약과 충돌하면 조용히 다른 모델을 쓰지 않고 거부하거나 선택을 요청한다. 일반 편의 프로필 변경은 이미 실행 중인 작업에 소급 적용하지 않는다.

모델과 effort를 따로 저장한다. 공급자마다 지원 수준이 다르므로 동일한 high 문자열이 동일 비용·추론량을 뜻한다고 가정하지 않는다. GPT는 공식 model/list의 supportedReasoningEfforts 같은 능력 정보를 활용하는 방향으로 설계한다. [Codex App Server](https://learn.chatgpt.com/docs/app-server).

Claude main의 provider는 MCP worker에 자동 상속하지 않는다. 반대로 외부 worker의 키·endpoint가 main으로 새지 않도록 별도 실행 환경을 사용한다. 새 모델·공급자·권한·기준 commit으로 바뀌면 기존 대화 thread를 무조건 재사용하지 않고 산출물 요약에서 새로 시작한다.

MCP 작업에는 session/lane/task/attempt ID, actual model·effort, status, artifacts, 검증 출력, 사용량과 과금 근거를 반환한다. provider 변경이나 재시도는 요청 ID와 연결한다. 작업 완료 문자열만으로 카드 완료를 승인하지 않는다.

## 7. Factory — lead는 조율, lane은 작업 소유

Factory lead는 카드 배정과 결과 확인을 담당한다. lane은 자신의 worktree·카드·실행 정책을 소유하고 plan → run → review → sync를 진행한다. 일반 모드도 같은 계약을 쓰되 lane_id만 생략한다.

추천 출발점은 Claude lead + Claude lane의 MCP 분업이다. 저비용 모델의 도구 루프가 검증된 작업군은 GLM Gateway lane으로 선택할 수 있다. GPT 중심의 긴 구현은 GPT Gateway lane 또는 Codex 위임 중 실제 성공 작업당 비용이 낮은 경로를 고른다. 이미지 전용 lane은 요청량이 많을 때만 두고, 보통은 공용 image 작업 큐를 쓴다.

모든 lane에 plan/run/review용 별도 상주 세션을 만들지 않는다. 병렬성은 독립된 카드·파일 경계가 있을 때만 사용한다. 같은 파일을 여러 모델이 동시에 수정하지 않으며, 작성 후 검토를 수행한다. 외부 writer와 Claude wrapper 중 한쪽만 해당 변경의 쓰기 소유자가 된다.

동시성은 lane 수 × subagent 수 × 외부 작업 수로 증가할 수 있다. 예를 들어 4 lane이 각각 3 wrapper를 만들고 wrapper마다 외부 작업 2개를 요청하면 외부 작업은 최대 24개다. 이는 설계상 산술 예시이며 실측이 아니다. native Agent 제한만으로 MCP 외부 작업이 통제된다고 가정하지 않는다. provider/account 전역 semaphore와 실행 전 예산 예약으로 제한한다.

기존 Factory의 glm_task model override 무시 정책은 현재의 단일 기준을 보호한다. 이를 삭제하는 대신 승인된 실행 정책 ID를 받아 실제 모델을 명시적으로 결정하도록 재설계한다. lead의 정책 권한을 개별 worker가 임의로 넓히지 못하게 한다.

## 8. 경제성 — 가장 싼 모델이 아니라 성공한 작업의 총비용

최적화 목표는 합격한 변경 1건의 총비용이다. 구독은 이미 지불한 금액, 한도 소모, API 추가 요금, 검증·재작업 비용을 구분한다. 구독이라는 이유로 모든 호출을 무료로 표시하지 않는다.

```text
총비용 = 조율 + 문맥 전달 + 실행 + 검증 + 실패 재시도 + 이미지 비용
성공 작업당 비용 = 해당 작업군의 전체 비용 / 승인된 결과 수
```

출발 모델 배정은 가설이다. 설계는 Claude/GPT 상위 추론, 명확한 구현은 검증된 중간 모델, 반복 변환·테스트 후보는 저비용 GLM 후보로 두고 실제 과제로 조정한다. 특정 회사가 항상 설계나 구현에 우월하다고 고정하지 않는다.

- 문맥 절약: 전체 대화 대신 요구사항·관련 파일·기준 commit·제약·검증 명령의 작업 묶음을 전달한다. 비밀과 무관한 개인 자료는 제외한다.
- 캐시: 동일 작업의 따뜻한 세션은 이어가되, 공급자 간 캐시 공유는 가정하지 않는다. 필요한 역할 설명만 로드한다.
- 승격: 실패 원인이 모델 추론인지 도구 오류인지 구분한다. 인증·네트워크 실패를 상위 모델 호출로 해결하지 않는다.
- 재시도: 같은 원인 반복은 중단하고 증거와 함께 상위 역할에 전달한다. 예시 상한 2회는 제안값이지 측정된 최적값이 아니다.
- 검토: 일상 변경은 기계 검증부터, 고위험 변경은 독립 리뷰를 추가한다. 모든 작업을 여러 모델에게 중복 작성시키지 않는다.

완료 조건은 기존 기준 대비 비용 감소만이 아니다. 누락된 요구사항·회귀·보안 결함이 늘지 않아야 한다. 토큰 절감률을 이번 보고서에서 약속하지 않는다.

## 9. GPT 이미지 생성은 별도의 실행 능력으로 제공

이미지 생성은 코딩 모델 이름의 부수 기능으로 취급하지 않고 image.generate / image.edit 작업 유형으로 분리한다. 계획·구현 역할과 같은 권한·예산·작업 상태 체계를 쓰되 산출물은 파일이다. 메인 모델이 Claude든 GLM이든 동일한 요청 경험을 제공한다.

공식 문서는 Codex 내장 이미지 생성이 gpt-image-2를 사용하며 일반 Codex 사용 한도에 포함된다고 설명한다. 비슷한 비이미지 turn보다 한도 소모가 평균 3–5배 빠르다는 안내도 있으나 이를 MoAI 원가 측정치로 사용하지 않는다. 프로그램 방식 생성은 Images API 경로를 안내한다. [공식 이미지 생성 문서](https://learn.chatgpt.com/docs/image-generation).

| 경로 | 용도 | 도입 조건 |
|---|---|---|
| 공식 Codex 내장 이미지 생성 | 개인 감독하의 소량 제작·수정 | 현재 설치 버전·계정의 지원과 MCP 산출물 회수를 실제 검증 |
| OpenAI Images API | 명시적 자동화·정형 요청·배치 | 별도 API 키·과금 동의, 현재 API 규격 확인 |

구독 OAuth를 Images API용 키처럼 재사용하지 않는다. 현재 MoAI의 codex_task가 텍스트 작업을 수행한다는 사실만으로 이미지 반환도 동작한다고 주장하지 않는다. 같은 runtime을 재사용하더라도 이미지 전용 artifact 수집 계약이 필요하다.

입력은 목적·프롬프트·참고 이미지·크기·품질·변경 금지 조건·출력 경로·예산이다. 결과에는 검증된 파일 경로, MIME, 크기, hash, 사용한 backend, 참조 이미지 provenance를 기록한다. 참고 파일 전송은 사용자가 허용한 범위만 사용한다. 파일 경로 탈출·symlink·임의 덮어쓰기를 막고 브라우저 실행 가능한 형식은 별도 취급한다.

사용자는 초안 확인 → 수정 또는 채택 → 프로젝트 반영 순서로 진행한다. 생성한 모든 후보를 배포하지 않는다. 이미지 1개 요청 실패가 전체 개발 lane의 무제한 재시도로 이어지지 않게 별도 한도를 둔다. 이번 요청은 이미지 생성 기능 설계이며 실제 이미지 제작은 수행하지 않았다.

## 10. 인증·승인·실행 권한을 한 화면에서 숨기지 않는다

Claude Code의 MCP 호출 승인과 외부 Codex가 실행하는 모든 shell·patch의 승인은 동일하지 않다. 초기에는 읽기·제안 전용을 우선 검증하고, 독립 구현은 worktree와 파일 범위를 명시한 writer 모드로 분리한다. 모든 실행이 Claude hook을 통과해야 한다면 native 실행 또는 검증된 도구 중개가 필요하다.

비-Claude Gateway는 Anthropic 공식 지원 대상이 아니다. 저장된 Claude 구독과 Gateway 인증이 혼합되지 않도록 실행 직전 유효 인증을 확인한다. [Gateway 공식 경계](https://code.claude.com/docs/en/llm-gateway).

Claude 구독은 수정하지 않은 공식 도구와 본인 인증을 유지하며, MoAI가 사용자 토큰을 수집·중개하는 로그인 서비스로 변하지 않게 한다. 개인 구독 사용을 무인 상용 서비스 허가로 확대하지 않는다. [인증·이용 조건](https://code.claude.com/docs/en/legal-and-compliance).

GLM Coding Plan은 지원 도구 조건이 있으므로 사용자 정의 MCP worker에서의 사용 허용을 별도 확인한다. 지원이 불명확한 구독을 범용 API인 것처럼 자동 등록하지 않는다. [Z.AI 정책](https://docs.z.ai/devpack/usage-policy).

API 과금 전환, 새로운 공급자로 코드 전송, 쓰기 범위 확대, 이미지 추가 제작은 사전 설정한 승인 범위를 벗어나면 중단한다. 모델 오류를 이유로 사용자 동의 없이 유료 fallback을 실행하지 않는다.

## 11. 최소 도입 순서와 통과 기준

| 순서 | 변경 목표 | 완료 판단 |
|---|---|---|
| P0 계약 정리 | main/role/lane 정책과 인증·예산 경계, 중첩 예외 승인 | 기존 cc/glm/gpt 기본 동작 보존, 실행 미리보기 일치 |
| P1 단일 위임 | 기존 codex_task의 모델·effort·소유권·결과 확장 | 실제 모델 확인, 읽기/쓰기 거부, 취소, 오류 전달 |
| P2 이미지 수직 검증 | 생성·수정→파일→미리보기→채택 | 실제 이미지 열기, 경로 검증, 과금 종류 확인 |
| P3 역할 프로필 | plan/run/review 역할과 외부 실행 매핑 | 동일 역할 지침, 중복 작성 없음, 사용자 override 보존 |
| P4 Factory 통합 | lane별 정책, 전역 동시성, 재개·작업 소유권 | 일반/Factory 동일 의미, restart 후 상태 정합성 |
| P5 최적화 | 절약/균형/품질 프로필 평가 | 수락률·회귀·성공 작업당 비용 비교 후 기본값 결정 |

평가 과제는 간단한 수정, 다중 파일 기능, 까다로운 버그, 보안 민감 변경, 이미지 포함 UI로 나눈다. 같은 기준 commit·요구·검증 조건으로 단일 모델, 직접 MCP, wrapper MCP, Gateway lane을 비교한다. 반복 평가에서 모델 순서 효과와 캐시 상태를 기록한다. 모든 모델을 모든 작업에 호출하는 온라인 경연은 기본 제품 동작이 아니라 별도 평가다.

## 12. Claim · Evidence · Baseline-attribution

Claim: Claude Code의 현재 공식 문서가 3단계 중첩을 안내함을 확인했다. 로컬 버전과 기존 MoAI MCP·Factory 관련 소스를 읽었으며, 그 바탕으로 공통 실행 정책 설계를 제안했다. 새 기능이 구현되었다고 주장하지 않는다.

이번 실행 명령 및 출력:

```text
claude --version
2.1.270 (Claude Code)

git branch --show-current
main

git rev-parse HEAD
2213871afb7d655f411c46a34fd38bb8153278fc

git -C .claude/worktrees/develop rev-parse HEAD
4056f69e1c20d942d4f9fc7363d3d79bffde899a
```

소스 조회: rg로 codex_task/glm_task 등록을 찾고 mcp_server.go:304–320, glm_task.go:138–177, codex_task.go:1–24를 읽었다. 등록, Factory override 무시, 프로세스 내부 background job 계약을 확인했다. 명령 결과에는 codex_task.go의 다음 주석이 포함됐다.

```text
// job is lost when the server exits. No reattachment is attempted and none
// is recorded.
```

재설계에서는 기존 background 작업을 durable 작업으로 가장하지 않는다. 중단 후 완료 여부가 불명확하면 unknown 상태로 두고 산출물·기준 commit을 대조한 후 재개 여부를 결정한다.

Baseline-attribution: 규칙은 main 작업 트리, task 구현은 develop 작업 트리의 현재 파일을 읽었다. 두 트리를 같은 baseline이라고 부르지 않는다. 이전 보고서의 테스트 통과 수치를 이번 실행으로 재사용하지 않았다. 공식 웹 문서는 2026-09-14 조회 기준이다.

## 13. Gaps · Residual-risk · 결정할 사항

Gaps: 이번에는 모델 호출·native 3단계 생성·MCP live tools/list·이미지 생성·비용 측정·제품 테스트를 실행하지 않았다. 이미지 App Server 산출물 수집과 구독별 권한은 실제 수직 검증이 필요하다. HTML 스킬의 design-tokens.md는 경로에 없어서 본문 토큰·제공 템플릿을 사용한다.

Residual-risk: 깊은 중첩과 외부 작업이 합쳐지면 비용·동시성이 증폭될 수 있다. 모델을 바꾸면 결과 형식과 오류 양상이 달라진다. 공식 지원 여부·구독 조건은 변경될 수 있다. 계정 한도가 불명확하면 안전한 중단과 사용자 선택이 필요하다.

권고 결정: Hybrid Factory, 역할 중심 프로필, 균형형 초기값, 얕은 기본 계층, 외부 writer 명시 권한, 이미지 독립 작업, 과금 전환 사전 승인. 3단계 중첩을 전면 허용하는 정책 변경은 별도 승인 후 진행한다. 이 설계에 동의하더라도 바로 전면 구현하지 않고 단일 GPT 위임과 이미지 수직 검증부터 시작한다.

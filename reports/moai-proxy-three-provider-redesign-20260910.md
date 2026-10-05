---
title: "moai-proxy 재설계: cc·gpt·glm 세 명령과 공통 /model"
date: "2026-09-10"
lang: ko
---

## 1. 최종 사용자 경험

**명령은 시작 모델을 정한다. 명령 이름으로 세션의 공급자를 제한하지 않는다.** 공개 launcher는 `moai cc`, `moai gpt`, `moai glm` 세 개뿐이며 모두 Claude Code를 실행하고 동일한 3사 모델 목록을 제공한다. `moai gg`와 `moai cg`는 제거한다. Codex CLI를 실행하는 설계가 아니다. 이 문서는 이전 내부 설계의 단일 공급자 세션·네 번째 선택 launcher·CG 혼합 모드 유지 결정을 대체한다. 구현 완료 보고서는 아니다.

| 명령 | 시작 동작 | 세션 안의 `/model` | 인증 변경 |
|---|---|---|---|
| `moai cc` | 설정된 Claude 기본 모델로 시작 | Claude·GPT·GLM 공통 목록 | 없음 |
| `moai glm` | 설정된 GLM 기본 모델로 시작 | 같은 공통 목록 | 없음; 기존 GLM 설정 재사용 |
| `moai gpt` | 설정된 GPT 기본 모델로 시작 | 같은 공통 목록 | 없음 |
| `moai gpt login` | MoAI 전용 PKCE 로그인 | 세션 실행과 분리 | MoAI 소유 GPT 인증만 저장 |
| `moai gpt logout` | MoAI 소유 GPT 인증 제거 | 다음 GPT 요청부터 인증 없음 | Codex CLI 인증 보존 |

세 회사의 모델 목록이 있다는 말은 세 회사의 이용권을 제공한다는 뜻이 아니다. 해당 공급자의 인증·모델 권한·조직 정책을 충족해야 실제 요청할 수 있다. 초기 모델을 명시적으로 고르려면 세 launcher의 `--model`을 사용한다. 공급자 혼합은 별도 `cg` 모드가 아니라 각 세션의 `/model` 전환으로 처리한다.

## 2. GPT 모델과 인증을 분리한다

제품명, 모델명, 결제 경로를 각각 저장한다. 사용자 화면의 공급자는 **GPT / OpenAI**이며 내부 transport 이름을 모델 이름으로 쓰지 않는다.

| 항목 | 의미 | 설계 결정 |
|---|---|---|
| Claude Code | 화면·도구 실행·권한 확인을 담당하는 클라이언트 | 모든 명령이 실행하는 유일한 코딩 클라이언트 |
| GPT 모델 | 답변과 도구 호출을 생성하는 모델 | `gpt-6-astra` 등 정확한 ID 사용 |
| ChatGPT 구독 인증 | PKCE로 얻는 backend 접근 자격 | `moai gpt login` 기본 계약 유지; backend가 허용하는 GPT만 제공 |
| OpenAI API 인증 | 공개 API의 별도 자격과 과금 경로 | 선택적 `moai gpt login --method api-key`; 숨김 입력, 명시적 선택 |
| Codex CLI | 별도 실행 프로그램과 인증 저장소 | 실행·설치·로그아웃 강제 없음 |

**PKCE 토큰을 공개 API 키처럼 사용할 수 있다고 가정하지 않는다.** 구독 backend가 특정 GPT 모델을 지원하지 않으면 명확한 지원 불가 오류를 반환한다. 공개 API로 조용히 우회하여 비용을 발생시키지 않는다. 반대로 “Codex가 아닌 모델”이라는 요구를 “PKCE를 없애고 API 과금만 사용”으로 해석하지 않는다. 인증 방식은 기존 요구대로 유지하고, 정확한 GPT 모델의 지원 여부는 인증 경로별로 검증한다.

| 요청한 계열 | 이번 공식 문서 확인 | 기본 목록 정책 |
|---|---|---|
| GPT-6 | `gpt-6-astra`; streaming·function calling 지원 표기 | 우선 후보. 계정별 접근 및 Claude 도구 왕복 검증 후 활성화 |
| GPT-5.6 Sol | 공식 문서상 `gpt-5.6`은 Sol로 연결되는 alias; 명시 ID는 `gpt-5.6-sol` | `moai gpt` 기본 모델. registry에는 명시 ID 저장 |
| GPT-5.6 Terra | 공식 명시 ID는 `gpt-5.6-terra` | 공통 후보 목록에 포함; 계정·transport별 접근 검증 후 활성화 |
| GPT-5.6 Luna | 공식 명시 ID는 `gpt-5.6-luna` | 공통 후보 목록에 포함; 계정·transport별 접근 검증 후 활성화 |
| “GPT-6 astro” 표기 | 이번 공식 문서에서 확인한 명칭은 GPT-6 Astra, ID는 `gpt-6-astra` | 보고서와 실제 요청은 공식 표기 Astra로 통일; `gpt-6-astro`를 임의 생성하지 않음 |
| Codex 전용 모델 | 이번 요구 대상 아님 | 기본 목록 제외. GPT 모델 대신 자동 대체하지 않음 |

사용자의 정정에 따라 GPT-5.3 관련 검토는 대상에서 제외했다. GPT 계열의 기본값은 GPT-5.6 Sol이며 후보 목록은 Sol, Terra, Luna, GPT-6 Astra의 네 모델이다. 공식 모델 문서의 기능 표시는 MoAI 호환성 시험 결과가 아니다. 출처: [GPT-6 Astra](https://developers.openai.com/api/docs/models/gpt-6-astra), [GPT-5.6 Sol](https://developers.openai.com/api/docs/models/gpt-5.6-sol), [GPT-5.6 Terra](https://developers.openai.com/api/docs/models/gpt-5.6-terra), [GPT-5.6 Luna](https://developers.openai.com/api/docs/models/gpt-5.6-luna).

## 3. 공통 /model 목록과 요청별 라우팅

세 launcher가 동일한 `CatalogSnapshot`에서 picker 설정을 생성한다. snapshot에는 공급자, 정확한 upstream ID, 인증 방식, context·도구·입력 형식 지원, 검증 상태가 들어간다. 시작 모델만 launcher마다 다르다.

Claude Code 공식 문서는 여러 사용자 지정 항목에 `modelPicker`를 안내하고 단일 추가 항목에 `ANTHROPIC_CUSTOM_MODEL_OPTION`을 안내한다. 이번 설계는 이전 세션에서 관찰한 `modelPicker.options` 경로를 우선 재검증한다. `/v1/models` 자동 검색만으로 비-Claude 모델이 모두 표시된다고 가정하지 않는다. [모델 설정](https://code.claude.com/docs/en/model-config)

1. 인증 가능한 공급자와 조직 allowlist의 교집합을 계산한다.
2. 동일한 picker를 자식 프로세스 전용 settings overlay로 주입한다. 전역·프로젝트 settings를 MoAI가 수정하지 않는다.
3. `--model`에 시작 모델을 전달한다. 명시 인수는 명령 기본값보다 우선하지만 정책 제한을 넘지 못한다.
4. `/model`에서 선택된 ID가 다음 Messages 요청에 실리면 registry가 정확히 일치하는 경로를 결정한다.
5. 로컬 gateway 인증은 제거하고 선택한 공급자의 인증만 upstream에 넣는다.

`/model` 선택 순간 proxy에 별도 이벤트가 온다는 가정은 금지한다. **실제 요청의 `model`이 라우팅 기준**이다. 상태 표시는 “마지막 처리 모델”과 “시작 기본값”을 구분한다. picker에서 선택만 하고 요청하지 않은 상태를 실제 호출로 기록하지 않는다.

인증 없는 항목은 설명에 인증 필요를 표시할 수 있으면 표시하고, 선택 후 요청은 외부 송신 전에 거절한다. native picker의 disabled 행 지원은 확인되지 않았으므로 회색 비활성 UI를 확정 설계로 약속하지 않는다. 지원하지 않는 클라이언트 버전에서는 3사 전환 가능이라고 표시하지 않고 업그레이드 안내 또는 명시적 제한 모드를 제공한다.

### `moai cc`에서 GPT·GLM으로 바꿀 수 있는가

**판정은 “현재는 불가, 공통 proxy 적용 뒤에는 기술적으로 가능, 실제 동일 세션 E2E는 미검증”이다.**

| 근거층 | 이번 관찰 | 판정 |
|---|---|---|
| 현재 MoAI | `moai cc`는 `unifiedLaunch(profileName, "claude", ...)`를 호출하고 `applyCCMode`가 GLM 환경을 제거한 뒤 POSIX에서 `syscall.Exec`로 Claude Code를 실행 | 현재 `cc` 세션에는 GPT·GLM을 받을 공통 gateway가 없으므로 불가 |
| Claude Code 공식 계약 | `/model <name>`은 세션 모델을 즉시 바꾸며, custom `ANTHROPIC_BASE_URL`에서는 인식하지 못한 모델 문자열도 provider로 전달; 여러 항목은 `modelPicker`로 구성 가능 | 한 base URL 아래 request별 model routing 설계 가능 |
| 이번 loopback 시험 | Claude Code 2.1.267이 `gpt-5.6-sol`과 `glm-5.3-flash`를 각각 `POST /v1/messages?beta=true`의 `model`로 그대로 전송하고 모의 응답을 소비; 두 실행 exit code 0 | 비-Claude ID 전달과 Messages 응답 소비는 관찰함 |
| 동일 대화 `/model` 연속 전환 | 격리 TTY가 workspace trust 화면에서 멈추어 inference를 만들지 못함 | 아직 PASS 아님; 구현 단계 T03으로 유지 |

gateway의 `/v1/models` 자동 검색은 `claude` 또는 `anthropic`으로 시작하는 ID만 picker에 추가하므로 GPT·GLM 목록은 자동 검색에 맡길 수 없다. 세션 전용 `modelPicker` 구성이 필요하다. 또한 picker에서 `Enter`는 user settings에 기본 모델을 저장한다. 부모·전역 settings 불변을 지키려면 사용자에게 **`s`로 “이번 세션만” 선택**하도록 명확히 표시하고, 전환 E2E에서 `~/.claude/settings.json`의 실행 전후 hash가 같음을 검증한다. 직접 `/model <name>` 입력은 `Enter`와 같이 저장되므로 불변 계약을 만족하는 기본 안내로 사용하지 않는다. [Claude Code 모델 설정](https://code.claude.com/docs/en/model-config), [Claude Code gateway](https://code.claude.com/docs/en/llm-gateway)

## 4. 내부 package 구조

같은 Go module의 `internal/proxy`로 구현하여 기존 `moai` 바이너리에 포함한다. 별도 proxy 설치, CA 설치, hosts 수정, 관리자 권한은 요구하지 않는다. 다음 이름은 구현할 책임 경계이며 이미 존재한다는 뜻이 아니다.

| 파일 / 경계 | 책임 | 금지할 결합 |
|---|---|---|
| `internal/cli/gpt.go` | GPT 실행·login·logout 조립 | HTTP 변환 로직 직접 포함 |
| `internal/cli/proxy_launch.go` | cc/gpt/glm의 공통 launch plan | 기존 worktree·session 옵션 누락 |
| `internal/proxy/server.go` | loopback listener·인증·종료 | CLI/config package import |
| `internal/proxy/catalog.go` | 모델·계정 권한·capability snapshot | 접두사 추측 라우팅 |
| `internal/proxy/router.go` | 요청 모델 검증·adapter 선택 | 공유 `currentProvider` 변수 |
| `internal/proxy/anthropic.go` | Claude Messages·OAuth/API 인증 분기 | 다른 공급자의 인증 전달 |
| `internal/proxy/openai.go` | GPT transport 선택·Responses 호출 | API와 구독 backend 정책 혼합 |
| `internal/proxy/glm.go` | Z.AI Messages 경로 | OpenAI endpoint로 임의 대체 |
| `internal/proxy/translate/` | Messages↔Responses·tool·SSE | 전역 tool ID map |
| `internal/proxy/auth/` | PKCE·refresh·전용 저장·삭제 | Codex CLI 저장소 수정 |
| `internal/proxy/policy.go` | 전환·기능·본문 크기·외부 송신 정책 | 실패 시 무단 fallback |

핵심 타입은 `LaunchPlan{InitialModel, Catalog, ChildEnv, ChildSettings}`, `ModelEntry{RouteID, Provider, UpstreamID, AuthMethod, Capabilities}`, `RequestContext{RequestID, Entry, CredentialRef}`로 제한한다. transport가 두 가지일 뿐 공급자는 OpenAI 하나다. 필요 이상으로 plugin framework나 별도 daemon을 만들지 않는다.

## 5. ccmproxy에서 차용할 부분과 고칠 부분

상류 분석 기준은 `112588175eb2b3b693a5bd23d58d8c501e0ae406`이다. 소스 차용 기록·원본 고지를 남기되 GPL을 기술 분석의 중단 사유로 삼지 않는다. 이 문서는 배포의 법적 허가 판단이 아니다.

| 상류 로직 | 차용할 핵심 | moai-proxy 재설계 |
|---|---|---|
| `translate/claude_to_codex.go` | Messages를 Responses 입력으로 변환, tool 이름 복원 정보 | GPT용 순수 변환기로 분리. 구독 backend의 `store:false`, 강제 streaming 조건을 공개 API와 분리 |
| Responses→Messages streaming | 시작·delta·tool·종료 이벤트 대응 | 요청별 상태기계; 중간 EOF를 성공 종료로 위장하지 않음 |
| `SignatureMarker = "ccmp1:"` | 암호화 reasoning 출처 구분 | 공급자·계정·형식 버전별 provenance; 단순 접두사만으로 신뢰하지 않음 |
| Codex 인증 store | PKCE·refresh·저장 | MoAI 전용 owner namespace·원자적 교체·refresh 직렬화 |
| model registry | 표시 ID와 backend 모델 대응 | Claude/GPT/GLM 통합; 인증 방식과 모델을 별도 축으로 저장 |
| router | Messages 요청 수신과 upstream 전달 | unknown은 명시 오류. Claude를 무조건 fallback 대상으로 사용하지 않음 |
| TLS/CA/host interception | 원래 endpoint를 가로채는 배치 | 차용하지 않음. 자식 전용 base URL로 연결 |

상류 Codex adapter 전체의 이름만 GPT로 바꾸지 않는다. HTTP endpoint, 허용 헤더, request fields, quota와 오류 형식은 transport별 계약이다. API와 구독의 통합 테스트는 별도로 유지한다.

## 6. 세션 중 공급자를 바꾸는 계약

동일한 대화를 유지한다는 것은 다른 회사의 내부 reasoning까지 재사용한다는 뜻이 아니다. 사용자·assistant의 공개 text, 정상적으로 완료된 tool use/result를 다음 공급자 형식으로 재구성한다.

| 상황 | 처리 계약 |
|---|---|
| Claude → GPT → GLM → Claude | 원본 대화는 보존; 요청 사본에서 대상 모델에 맞게 변환 |
| 서명된 thinking / encrypted reasoning | 동일 계정·동일 backend 호환 범위에서만 재사용. 다른 공급자에는 전달하지 않음 |
| 진행 중 tool call | 기존 요청은 원래 공급자로 끝냄. 미완료 tool pair가 있는 새 전환 요청은 명시 오류 |
| 병렬 subagent 요청 | 각 요청의 model로 독립 라우팅. 부모의 마지막 provider로 덮어쓰지 않음 |
| 작은 context 모델로 전환 | 전체 history 기준 한도 검사; 명시적 compact 또는 새 세션 안내. 몰래 앞부분 삭제 금지 |
| 이미지·PDF·서버 도구 등 미지원 입력 | 요청 전 명시 거절 또는 사용자 승인된 별도 변환. 묵시적 내용 누락 금지 |
| quota·429·인증 실패 | 같은 경로에서 제한적 재시도만. 다른 회사로 자동 전송 금지 |

Claude 자체의 `/compact`·background model·별칭 요청은 별도 시험한다. MVP에서 보조 요청의 route는 명시적인 launch slot mapping으로 고정하여 숨은 업체 변경을 막는다. `/model`이 모든 background slot까지 따라 바꾼다고 주장하지 않는다. 추후 요청별 식별 근거가 확보되면 follow-selected 정책을 추가한다.

전환하면 기존 대화가 새 공급자에게 전달될 수 있음을 최초 통합 모델 설정 시 고지하고 동의를 받는다. 모델 선택은 단순 외형 변경이 아니라 데이터 전송처 변경이다. 고지 없는 자동 fallback은 금지한다.

## 7. Claude 인증과 GPT 로그인의 소유권

Anthropic 공식 문서는 base URL만 변경하면 기존 subscription 인증을 유지할 수 있지만 gateway credential을 설정하면 subscription 대신 해당 credential의 경로를 사용한다고 설명한다. 또한 비-Claude 모델 라우팅은 공식 지원 대상이 아니라고 명시한다. 따라서 기술 가능성과 공식 지원을 구분한다. [Gateway 인증과 지원 범위](https://code.claude.com/docs/en/llm-gateway)

Claude 구독을 보존하려면 실제 클라이언트에서 OAuth 전달·refresh·필수 beta 헤더를 검증해야 한다. 로컬 gateway token을 넣고도 Claude Max 과금이 그대로라고 주장하지 않는다. 아래 두 인증 모드를 분리한다.

| 모드 | 클라이언트 → proxy | proxy → Claude | 제한 |
|---|---|---|---|
| 구독 보존 후보 | 기존 Claude 자격 + 별도 검증된 로컬 접근 통제 | 허용된 OAuth 헤더만 원래 Anthropic endpoint로 전달 | 클라이언트가 추가 로컬 header를 지원하는지 포함한 실측 선행 |
| API gateway | 세션 전용 local token | MoAI에 설정한 Anthropic API 자격 | Claude 구독과 별도 과금임을 명시 |

구독 모드는 local token 충돌을 해결하지 못한 상태로 출시하지 않는다. loopback이라는 이유만으로 무인증 proxy를 허용하지 않는다. 첫 단계에서 양쪽 모드를 측정하여 구독 보존 가능 여부를 판정한다. API 모드가 된다고 구독 모드의 갭을 닫지 않는다.

GPT login은 PKCE state·verifier 검증, loopback callback timeout·취소, refresh 단일화, 제한 권한 저장을 구현한다. 외부 Codex auth를 복사해 독립 refresh token처럼 사용하지 않는다. `logout`은 MoAI 소유 파일 또는 keyring 항목만 제거하며 외부 인증을 지우지 않는다. 이미 실행 중인 서버는 credential generation을 요청 전 검사하여 logout 이후 새 GPT 요청을 거절한다. 진행 중 요청 종료·원격 token 철회 여부는 별도 고지한다. 단순 파일 삭제를 공급자 전체 로그아웃이라고 부르지 않는다.

## 8. 런처 수명과 설정 불변

현재 POSIX 런처는 `syscall.Exec`를 사용한다. 같은 프로세스 안에 시작한 proxy는 exec 순간 사라지므로 proxy-enabled 세 명령은 공통 spawn/wait 경로가 필요하다.

1. 현재 프로젝트·worktree·session 옵션을 기존 런처에서 해석한다.
2. catalog와 인증을 점검한다. 첫 모델이 실행 불가면 다른 회사로 시작하지 않는다.
3. `127.0.0.1:0` listener와 세션별 접근 통제를 준비한다.
4. 비밀이 없는 임시 settings 파일과 자식 env를 조립한다. credential은 argv·로그에 넣지 않는다.
5. Claude child를 실행하고 실제 child PID로 session attribution을 설정한다.
6. 종료 코드·signal·터미널 동작을 보존하며 child 종료 뒤 listener와 임시 파일을 정리한다.

POSIX의 PID 주입은 검증된 self-reexec wrapper 또는 기존 안전한 helper를 활용한다. Windows는 별도 child/job 종료 계약을 시험한다. Ctrl-C, Ctrl-Z, fg, 부모 강제 종료, background hooks와 session registry 회수는 각각 검증한다. 부모 shell env 불변은 자식 env 조립으로 보장할 수 있으나, Claude 자체가 저장하는 설정까지 불변인지는 파일 diff 시험으로 확인한다. 허용 규칙·hooks·MCP를 없애는 격리 설정으로 이 시험을 대신하지 않는다.

## 9. 프로토콜과 보안의 최소 계약

Messages 수신은 model 중복 key·압축 전후 크기·잘못된 JSON을 검사한다. 모델 ID는 registry exact match로 처리하며 접두사에 따라 외부 URL을 만들지 않는다. 공급자 endpoint는 허용 목록으로 제한하고 redirect에 인증을 전달하지 않는다.

Responses 변환은 system, text, image, tool use/result, tool choice, output limit를 capability에 맞게 처리한다. 도구 이름 축약은 충돌을 감지하고 요청별 역매핑한다. SSE는 `message_start → content blocks → message_delta → message_stop` 순서를 보장하고 여러 tool call의 index/ID를 구분한다. upstream 실패·중간 EOF·취소에 성공 terminal event를 추가하지 않는다. 이미 응답 bytes를 보낸 뒤 재시도하지 않는다.

`/v1/messages/count_tokens`는 공급자별 정확성 수준을 명시한다. 추정치를 실제 tokenizer 값으로 표시하지 않는다. `/v1/models`는 세션 catalog만 노출한다. bootstrap·telemetry 등 부가 경로는 path별 정책을 두며 모든 unknown path를 Anthropic으로 전달하지 않는다. 본문·API key·OAuth token·reasoning은 기본 로그에서 제외한다. statusline은 로컬 상태만 읽고 새 API 요청을 만들지 않는다.

## 10. 설정 및 설치 흐름

다음 YAML은 **제안 schema**이며 현재 parser가 지원하는 설정이 아니다.

```yaml
proxy:
  schema_version: 2
  catalog: unified
  launch_defaults:
    cc: claude-default
    gpt: gpt-default
    glm: glm-default
  providers:
    anthropic:
      auth_mode: subscription-passthrough # 실측 합격 전 활성화 금지
    openai:
      auth_mode: subscription-pkce
      default_model: gpt-5.6-sol
      candidate_models: [gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna, gpt-6-astra]
      api_fallback: false
    zai:
      auth_mode: existing-glm-credential
  switching:
    automatic_provider_fallback: false
    history_transfer_consent: required
  legacy_models:
    enabled: false
```

설치 프로그램은 proxy 코드를 함께 배포할 뿐 이용권·API 키를 만들지 않는다. 최초 실행에서 Claude Code 버전, 공급자 인증, 기본 모델, 조직 정책을 점검하고 해당 공급자만 설정하도록 안내한다. 기존 `moai glm`의 사용자 설정은 읽어서 adapter에 연결하며 소유권 없는 파일을 마이그레이션 명목으로 덮어쓰지 않는다.

`cg` 제거는 `internal/cli/cg.go` 하나를 지우는 작업이 아니다. 이번 `git grep`은 runtime·문서·과거 기록을 합쳐 341개 tracked file에서 관련 문자열을 찾았다. 과거 SPEC·release note·감사 기록은 역사 자료이므로 고쳐 쓰지 않는다. live CLI 등록, `claude_glm` 분기, `team_mode: cg` 판정, tmux 주입, 현재 README·docs-site·배포 template와 그 시험만 제거·재기준화한다. 기존 사용자 설정에서 `team_mode: cg`를 발견하면 자동으로 다른 provider를 선택하지 않고 `moai doctor`가 `cc` 또는 `glm`을 명시적으로 고르도록 안내한다. `moai cg`는 더 이상 help에 나타나지 않고 unknown command로 끝난다.

## 11. 구현 순서와 20개 합격 기준

| 순서 | 우선순위 | 작업 | 다음 단계 진입 조건 |
|---|---|---|---|
| A | High | Claude 구독/local 인증 충돌, picker, GPT backend별 정확한 모델 권한 측정 | 인증 방식별 가능/불가와 증거 확정 |
| B | High | common catalog·request router·3개 adapter·변환 golden | unknown·인증 누출·잘못된 fallback 0건 |
| C | High | gpt login/logout·공통 supervisor·cc/gpt/glm 연결·cg 제거 | 프로세스·설정·외부 auth 보존 시험 합격 |
| D | High | 세 회사 전환·tool 왕복·compact·hooks·MCP·subagents | 아래 실제 세션 기준 합격 |
| E | Medium | 설치 onboarding·doctor·설명·모델 catalog 갱신 | 새 사용자 설치 및 기존 사용자 회귀 확인 |

아래 항목은 **향후 검증 계약 20개이며 현재 PASS 수가 아니다.**

| ID | 시험 | 합격 판정 |
|---|---|---|
| T01 | 세 명령의 initial model | cc/Claude, gpt/GPT, glm/GLM이 첫 요청에 일치 |
| T02 | 세 명령 각각 `/model` | 동일한 3사 허용 목록, 관리 정책 우회 없음 |
| T03 | cc에서 Claude→GPT→GLM→Claude | 같은 세션 네 turn의 upstream 기록과 응답 확인; picker의 `s` 사용 |
| T04 | 세 공급자 tool 왕복 | Read·승인된 임시 Write·Bash 결과 재입력 정상 |
| T05 | PKCE | state 불일치·취소·timeout 거절, 정상 callback 저장 |
| T06 | GPT logout | MoAI auth만 삭제; Codex auth hash 불변; 실행 중 서버 신규 요청 거절 |
| T07 | GPT 모델 권한 | transport별 정확한 ID 호출; 미지원 ID를 다른 모델로 대체하지 않음 |
| T08 | 구독/API 과금 경계 | API fallback 0건, 선택한 endpoint·credential 종류 일치 |
| T09 | Claude subscription | OAuth refresh·beta 헤더·local 인증의 실제 공존 확인 |
| T10 | 외부 auth 보존 | Claude/Codex 기존 자격의 소유권 외 수정·삭제 없음 |
| T11 | 설정 보존 | 부모 env·전역/프로젝트 settings hash 불변; `/model` 선택은 `s` 경로 사용 |
| T12 | 프로세스 수명 | child PID·exit status·signal·PTY·부모 종료 cleanup |
| T13 | tool 동시성 | interleaved call ID/인수·이름 역매핑 오염 없음 |
| T14 | SSE 실패 | EOF·429·취소를 성공으로 위장하지 않고 post-commit retry 없음 |
| T15 | history 전환 | 타사 reasoning 미전달, 공개 tool pair 보존, 미완료 pair 거절 |
| T16 | context/입력 기능 | 작은 context·미지원 이미지/PDF 명시 오류, 무단 절단 없음 |
| T17 | subagents·보조 slot | 요청별 model 독립성, 고정 slot 표시와 실제 route 일치 |
| T18 | hooks·MCP·compact | 실제 활성 환경에서 전후 동작·승인 경계 보존 |
| T19 | 외부 송신 보안 | unknown model/path·redirect·압축 공격·local token 누출 차단 |
| T20 | 설치·회귀·제거 | 새 설치·기존 GLM·기존 cc/worktree 회귀와 `cg`/`gg` help·실행 부재 확인 |

## 12. Claim · Evidence · Baseline · Gaps · Residual-risk

**Claim:** 공개 launcher를 cc·gpt·glm 세 개로 줄이고 gg·cg를 제거하는 설계로 변경했다. 현재 `moai cc`에서는 GPT·GLM 전환이 불가능하지만, 공통 proxy를 거치는 목표 구조에서는 request model 기반 전환이 기술적으로 가능하다. 이 턴은 구현·유료 API 호출·동일 세션 3사 전환 완료를 주장하지 않는다.

**Evidence — 이번 실행의 명령과 실제 출력 발췌:**

```text
$ git rev-parse --short HEAD
2213871af
$ git branch --show-current
main
$ claude --version
2.1.267 (Claude Code)
$ moai --version
moai-adk v3.2.0-rc.5
$ node /tmp/moai-proxy-model-routing-probe.mjs | jq '{runs: [.runs[] | {model, exit, result}], post_requests: [.requests[] | select(.method == "POST")]}'
{
  "runs": [
    {
      "model": "gpt-5.6-sol",
      "exit": {
        "code": 0,
        "signal": null
      },
      "result": "MOCK_OK:gpt-5.6-sol"
    },
    {
      "model": "glm-5.3-flash",
      "exit": {
        "code": 0,
        "signal": null
      },
      "result": "MOCK_OK:glm-5.3-flash"
    }
  ],
  "post_requests": [
    {
      "method": "POST",
      "path": "/v1/messages?beta=true",
      "model": "gpt-5.6-sol"
    },
    {
      "method": "POST",
      "path": "/v1/messages?beta=true",
      "model": "glm-5.3-flash"
    }
  ]
}
$ git -C /tmp/ccmproxy-audit.B3eTfF rev-parse HEAD
112588175eb2b3b693a5bd23d58d8c501e0ae406
$ rg -n 'SignatureMarker|store|stream' /tmp/ccmproxy-audit.B3eTfF/internal/translate/claude_to_codex.go
21:const SignatureMarker = "ccmp1:"
86: body, _ = sjson.SetBytes(body, "store", false)
87: body, _ = sjson.SetBytes(body, "stream", true)
```

loopback 시험 출력은 결정 필드만 정리한 발췌이며 원문에는 각 HEAD `/api/hello`와 Claude 결과 JSON도 있었다. 마지막 두 소스 줄은 원본 tab을 공백으로 표시했다. 소스 읽기와 runtime 재현을 구분한다.

**Baseline-attribution:** 2026-09-10, `main@2213871af`의 현재 dirty checkout을 읽었다. 다른 변경 파일은 건드리지 않았다. 이전 보고서의 GPT·GLM 개별 실측은 [이전 실측 원장](moai-proxy-feasibility-plan-20260910.md)에 보존하며 이번 측정으로 재분류하지 않는다. 이전 상세 package 분석은 [앞선 내부 설계](moai-proxy-internal-redesign-20260910.md)를 참고하되 명령·인증·catalog·Claude 범위가 충돌하면 이 문서를 우선한다.

**Gaps:** 새 CLI와 internal/proxy 구현, 현재 계정별 GPT 권한, 같은 대화의 `/model` 3사 순차 전환, Claude 구독/local 인증 공존, Windows·PTY·hooks·MCP·subagent 통합, cg 제거 후 전체 회귀는 이번에 관찰하지 않았다. 격리 interactive probe는 workspace trust 화면에서 세 차례 inference에 도달하지 못했다. 따라서 두 개의 성공한 non-interactive model-ID 시험을 동일 세션 `/model` 성공으로 과장하지 않는다. 공식 settings-reference 상세 페이지는 웹 도구의 응답 크기 제한으로 열지 못했다.

**Residual-risk:** Claude Code의 비-Claude 모델 경로는 공식 지원 밖이다. 모델 폐기·구독 backend 변경·관리자 정책에 따라 picker 또는 요청이 달라질 수 있다. API 경로 성공만으로 구독 경로의 성공을 추론할 수 없다. 특히 Claude 구독 보존과 로컬 gateway 인증의 동시 충족을 확인하지 못하면 “기존 로그인 그대로 3사 전환” 출시 문구를 사용할 수 없다.

## 13. 결정 요약

최종 구조는 **MoAI 단일 바이너리 → 내부 proxy → Claude Code 공통 /model → Claude / GPT / GLM**이다. 공개 launcher는 cc·gpt·glm 세 개뿐이며 gg와 cg는 제거한다. `moai cc`도 공통 proxy를 거쳐 Claude로 시작해야 그 세션에서 GPT·GLM으로 바꿀 수 있다. **`moai gpt`는 GPT-5.6 Sol로 시작하고 `/model`에는 GPT-5.6 Sol·Terra·Luna와 GPT-6 Astra를 제공**한다. 부모·전역 설정을 보존하려면 picker에서 `s`를 사용한다. production 구현에 앞서 동일 세션 T03과 인증 경계를 먼저 실측한다.

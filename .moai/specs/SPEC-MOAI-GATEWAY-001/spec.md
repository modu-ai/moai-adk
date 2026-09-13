---
id: SPEC-MOAI-GATEWAY-001
title: "moai 공통 loopback gateway — cc·gpt·glm 세 launcher의 단일 ingress와 제공자별 모델 선택 경계"
version: "0.11.0"
status: draft
created: 2026-09-10
updated: 2026-09-12
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/gateway, internal/cli, internal/kanban, internal/hook"
lifecycle: spec-anchored
tags: "gateway, launcher, model-routing, anthropic-ingress, sse, loopback, kanban"
tier: L
---

# SPEC-MOAI-GATEWAY-001 — 공통 loopback gateway 코어

## HISTORY

- 0.11.0 (2026-09-12, t649 및 t650~t654) — 운영자가 공식 App Server 문서를 기준으로 전면 재설계,
  HTML 보고, 카드 발행, 구현을 지시했다. GPT 구독 인증·실행·reasoning 이력은 Codex App Server에 맡긴다.
  MoAI의 직접 구독 backend 호출·구독 토큰 읽기와 직접 opaque 복원 설계를 현재 구독 경로에서 대체한다.
  Claude Code UI, 제공자별 `/model`, Claude 도구 실행권과 명시 API 과금 지원은 유지한다.
  브라우저 구독·device-code 구독·API 키의 세 로그인 방식을 채택한다. t650~t654는 실제 발급된 실행 카드다.
  실험 API의 설치 버전 차이, 도구 전체 schema, native 실행 차단, pending call 복구는 완료 전 필수 게이트다.
  단일 dynamic tool의 별도 실증을 제품 전체 성공으로 승격하지 않는다.
  후속 운영자 결정: “ToolSearch 유지·발견한 도구만 공통 통로 사용”. 초기 native schema와 후발 dispatcher의
  hybrid를 채택하며 ToolSearch 비활성 전체 선등록은 채택하지 않는다. 과거 M0·receipt 승인과 관측은 당시 기록이다.

- 0.10.0 (2026-09-12, t649) — 운영자가 조사 결과의 구현·검증과 카드 발행을 승인했다.
  cc는 Claude, glm은 GLM, gpt는 GPT만 선택·요청할 수 있도록 세션 경계를 변경한다.
  기존 3사 교차 전환과 PICKER 이관·출시 대기 결정 중 충돌하는 부분은 이 판에서 대체한다.
  picker·기본값·모델 슬롯·재개·저장 격리를 코어에 포함하며 기존 REQ/AC의 하위 판정을 보강한다.
  REQ-MG-020/AC-MG-002 묘비는 유지한다. 과거 프로브·감사 기록은 당시의 관측이다.
  reasoning·receipt의 권한 및 유실 거절, 구독 서버 출력 정책, OAuth broker, Windows CI 검증과
  Windows API 기준 내구성 결정은 유지한다. 실제 제품 PTY·응답·tool·resume 판정은 아직 미완료다.
  같은 0.10.0의 t649 F1 보정으로 reasoning이 포함된 응답의 공개 message 경계·원래 phase·상대 순서 보존을 명시한다.
  multipart text와 별도 message를 구별하고 복원 metadata도 대화 receipt의 검증 대상에 포함한다.
  같은 판의 후속 보정으로 GPT의 명시 effort `medium`을 `high`와 함께 허용하고 원래 값을 보존한다.
  effort 누락은 계속 누락이며 다른 effort·수동 budget 변환·추론량 동등성으로 범위를 넓히지 않는다.

- 0.9.0 (2026-09-11) — 19시 이후 M0의 정상 응답·직접 OAuth refresh 관측으로 별도 세션 헤더를 확정했다.
  구독 SSE의 Content-Type 부재와 빈 terminal output을 엄격한 완료 item 검증으로 수용하는 계약을 보강했다.
  실제 합성 carrier·tool ID·정확한 resume의 보존 관측을 근거로 가역 tool ID binding과 대화별 hash-only receipt를
  선택된 설계 후보로 정했다. 독립 계획 감사·구현·실제 음성 시험 전 reasoning 제품 활성화는 금지한다.
  최초 보강 시 출력 상한과 Windows 의미는 답변 대기였다. 이후 같은 버전에서 아래 사용자 결정을 반영했다.
  기존 REQ/AC 번호와 상태를 유지하고 제품 통합 PASS를 선언하지 않는다.
  같은 0.9.0의 receipt 감사 보정으로 new/resume/continue/선택/fork의 사전 UUID와 retained native profile 수명을
  확정하고, 정상 병렬 A/B 및 required=false/혼재 후보의 수용 표와 lookup miss의 거절을 명시했다. 활성화 게이트는 유지한다.
  “구독 서버 출력 정책으로 진행” 승인으로 고정 구독 경로만 outgoing max_output_tokens를 생략한다. 입력 양의 정수
  검증과 API 키 매핑은 유지한다. Windows API 저장 기준의 별도 승인은 AUTH 0.2.0 소관이며 receipt 본문은 바꾸지 않았다.
  native-policy 후보의 독립 변경 준비 검토 PASS 뒤 같은 버전에서 native 입력/응답 thinking 보존과 GPT high/title/
  keep-all/로컬 metadata·계량 정책을 보강했다. shipped context 선언과 실제 계정 관측은 구별하고 receipt 본문은 보존했다.
  2026-09-12 실제 Claude Code 2.1.268 요청에서 `thinking.type=adaptive`와
  `thinking.display=omitted`, tool의 `defer_loading=true`, Responses 출력 message의
  `phase=final_answer`를 관측해 허용 문법을 좁게 보강했다. 고정 구독 endpoint가
  `stream=false`를 HTTP 400(`Stream must be set to true`)으로 거절한 사실에 따라
  비스트림 클라이언트 요청은 upstream에만 `stream=true`로 보내고 검증된 SSE를 JSON으로
  모은다. Responses reasoning item은 현재 opaque 보존 게이트가 열리기 전까지 명시
  오류로 남기며, 이 실측은 GPT 제품 완료나 resume 호환성 PASS를 뜻하지 않는다.

- 0.8.0 (2026-09-11) — 운영자가 M1 캡처 불일치 보정과 구현 계속을 승인했다. Claude Code 2.1.268의
  원본 `request-003.json`은 `stream` 키 없이 나머지 세 조건을 만족했다(`research.md` §17).
  `REQ-MG-023`과 `design.md` §4.1의 가정 A-VAL-2.1.268은 키 부재 또는 명시 JSON `false`만 허용하고,
  `null`·문자열·`true`는 제외한다. `AC-MG-003` (c)의 키 부재 음성 변형을 문자열 `"false"`로 바꾸고
  키 부재·명시 `false`를 각각 양성 대조군으로 둔다. G6-A1은 `AC-MG-018` (a)의 프로세스 env에
  정리 집합 14키와 `Z_AI_API_KEY`를 모두 오염시키는 것으로 반영했다. REQ/AC 번호와 계수는 유지한다.
  운영자 지시로 후속 Claude 실서비스 시험은 2026-09-11 19:00 Asia/Seoul 이후 Opus 5·Sonnet 5로 한다.
  과거 Sonnet 4.5 캡처는 역사적 근거이며 새 시험의 대체 근거가 아니다. M0는 429로 INCONCLUSIVE이고
  T09 양성 게이트는 그대로다. 최종 목표인 `moai gpt` 안의 GPT-6·GPT-5.6 사용에는 필요한 형제 카드의
  구현·시험도 포함되지만, 기존 AUTH·PICKER·TEAMMATE·CG 소관과 코어 출시 결합 조건을 유지한다.

- 0.1.0 (2026-09-10) — plan 단계 최초 작성. 이 항목은 **0.1.0 당시의 명칭과 식별자를
  그대로 둔다**(당시 이름은 "proxy", SPEC ID는 옛 식별자, 요구사항 접두사는 `REQ-MP`).
  작업 트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`,
  브랜치 `WT-moai-proxy-unified`, HEAD `d060e0d13`(= `origin/develop`,
  `git rev-list --count --left-right origin/develop...HEAD` → `0 0`, <!-- moving-ref-ok: origin/develop is the SUBJECT of this identity reading, not its anchor; the anchor is the pinned HEAD d060e0d13 stated on the preceding line, and the criterion is the measuring command with a dated 2026-09-10 reference -->
  `git status --short` 빈 출력)를 기준선으로 삼았다. 이 저장소의 카드 브랜치는
  `main`이 아니라 `develop`에서 분기한다.

  반영한 운영자 결정 세 가지:

  - **D1 — 범위는 3분할 Epic이고 이 SPEC은 코어만 맡는다.** GPT PKCE 인증은
    `SPEC-MOAI-GPT-AUTH-001`(제안 ID)로, `moai cg` 철거 스윕은
    `SPEC-MOAI-CG-RETIRE-001`(제안 ID)로 분리한다. 이 SPEC이 운반하는 수용 시험은
    설계 보고서 §11의 T01·T02·T03·T04·T11·T12·T13·T14·T15·T16·T17·T18·T19이다.
  - **D2 — 프로세스 모델은 분리된 detached child이며 `syscall.Exec`은 보존한다.**
    `internal/cli/launch_exec_posix.go:24-27`의
    `syscall.Exec(claudeBin, args, withSessionPID(env, os.Getpid()))`는 현재 프로세스를
    `claude`로 치환하므로 같은 프로세스 안의 goroutine listener는 POSIX에서 살아남지
    못한다. exec 이전에 별도 `moai` child로 띄우고 bind된 포트를 돌려받는다.
  - **D3 — `moai gg`는 존재하지 않으므로 제거 작업 항목을 폐기한다.** 근거는 이 SPEC
    산출물을 쓰기 전에 이 작업 트리에서 잰 값이다:
    `/usr/bin/grep -rn "moai gg" .` → `0`, 동일 범위 양성 대조군
    `/usr/bin/grep -rn "moai cg" .` → `745`. 제거 요구사항 대신 부재 회귀 가드만 둔다.

- 0.2.0 (2026-09-10) — plan-audit iter1 FAIL(0.70) 반영, 명칭 변경, clarification 6건 확정.

  **입력.** iter1 감사 보고서는 `.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md`에
  옛 경로·옛 식별자 그대로 남는다. 감사 증거이므로 이동·수정하지 않는다(핸드오프 §8).
  브랜치는 `WT-unified-gateway`로 바뀌었고 HEAD는 `d060e0d13` 그대로다. 작업 트리
  디렉터리 이름은 세션이 그 경로에 고정되어 있어 바뀌지 않았다.

  **D3 측정의 재현성.** 0.1.0의 `0`은 작성 전 값이다. 이 SPEC 문서와 iter1 감사 보고서가 그
  토큰을 담게 된 뒤로는 같은 명령으로 재현되지 않는다. 0.2.0의 재측정은 이 SPEC 디렉터리와
  `.moai/reports/`를 제외해 다시 잰 값이다(`research.md` §1.1).

  **명칭 변경: proxy → gateway.** 운영자 결정이며 근거는 셋이다.

  - `internal/config/envkeys.go:479`가 이미 이 자리를 "an LLM gateway"라 부른다 —
    "when ANTHROPIC_BASE_URL points at an LLM gateway".
  - `internal/harness/router`가 이미 `package router`를 쓰고 있다.
  - "proxy"는 CA 설치와 투명 가로채기를 연상시키는데, 설계 보고서 §5는 그 방식을
    명시적으로 기각한다.

  **번호 규칙.** 살아남은 항목은 번호를 유지하고, 통합된 항목의 번호는 폐기하며 재사용하지
  않는다. 새 항목은 현재 최댓값 뒤에 붙인다. 접두사만 `REQ-MP`→`REQ-MG`,
  `AC-MP`→`AC-MG`로 바뀐다.

  | 옛 ID | 새 ID | 상태 |
  |---|---|---|
  | `REQ-MP-001` ~ `REQ-MP-006` | `REQ-MG-001` ~ `REQ-MG-006` | 유지 — 본문 정정: 001, 003, 005(`REQ-MP-007` 흡수) |
  | `REQ-MP-007` | — | **폐기** — `REQ-MG-005`에 통합 |
  | `REQ-MP-008` ~ `REQ-MP-025` | `REQ-MG-008` ~ `REQ-MG-025` | 유지 — 본문 정정: 009, 016, 020, 021, 023, 024 |
  | — | `REQ-MG-026` | **신규** — `moai gpt`의 kanban/factory 진입과 `BackendGPT` |
  | `AC-MP-001` ~ `AC-MP-020` | `AC-MG-001` ~ `AC-MG-020` | 유지 — 본문 정정: 003, 005, 006, 014, 015, 018, 019 / 헤더 표기만: 002, 010, 013 |
  | — | `AC-MG-021` ~ `AC-MG-025` | **신규** — iter1 B1의 무연결 요구사항 해소 |

  (이 대응표는 0.3.0에서 iter2 감사의 diff 대조 결과에 맞춰 정정했다.)

  **감사 결함 처리.**

  - B1 — 무연결 요구사항 7건 해소: AC 5개 신설(`REQ-MG-017`·`022`·`003`·`014`·`018`
    대상), `REQ-MG-009`를 `AC-MG-006` 헤더로, `REQ-MG-016`을 `AC-MG-021` 헤더와 이관 표
    항목으로 승격.
  - B2 — `AC-MG-015`를 help 출력과 command tree 판정으로 재작성하고, D3 수치를 작성 전
    측정으로 정정.
  - B3 — `REQ-MG-021`을 판정 기구별로 재작성하고 비-GLM 세션 teardown 금지 불변식 추가.
  - B4 — `REQ-MG-020`을 측정 게이트로 전환하고, 전역 settings 불변은 `s`에 기대지 않고
    쓰기 대상 배제로 보장.
  - B5 — clarification 6건을 운영자 결정으로 확정(`plan.md` §H).
  - A1~A9 — 가드가 빌드가 아니라 시험임을 정정, 자기 무효화 부재 주장 정정, §A 가정 표시,
    launcher 이름 목록 세 곳, spawn 리터럴 위치, 교차 SPEC 상태 열, 요구사항 본문의 행 범위
    제거, `REQ-MG-014` 의무 분해, `s` 미측정 근거 귀속 정정.

  **요구사항 통합.** `BackendGPT` 요구사항(`REQ-MG-026`)의 자리는 `REQ-MP-007`을
  `REQ-MG-005`에 합쳐 만들었다. 두 요구사항은 같은 판정 표면(`AC-MG-006`)을 공유한다.
  `MOAI_SESSION_PID` 각인·signal·PTY·job control은 POSIX에서 `syscall.Exec` 보존으로부터
  나오는 보장이고, 종료 코드 전파는 POSIX에서는 같은 exec에서, Windows에서는 spawn-and-wait
  경로에서 나온다. (0.2.0 원문은 다섯 보장이 모두 `syscall.Exec`에서 파생된다고 적었으며,
  Windows에 대해 틀린 이 설명을 0.3.0에서 바로잡았다.)

- 0.3.0 (2026-09-10) — plan-audit iter2 FAIL(0.80) 반영.

  **입력.** iter2 감사 보고서: `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md`.
  감사 증거이므로 이동·수정하지 않는다.

  **차단 결함 처리.**

  - G2-MP1 — §D.2의 결번 자리에 `REQ-MG-007` 폐기 묘비를 넣었다.
  - G2-B1 — gateway launch의 `.claude/settings.local.json` 계약을 정했다. 세 launcher 모두
    exec 전에 GLM 라우팅 키 집합을 `removeGLMEnv`와 같은 의미로 정리하고, 어떤 gateway launch도
    base URL이나 GLM credential을 그 파일에 쓰지 않으며, SessionStart 재주입은
    `MOAI_LAUNCH_PROVIDER`로 억제한다. settings `env`와 프로세스 env의 우선순위는 미측정이라고
    명시했다(`REQ-MG-021`, `REQ-MG-022`, `design.md` §5·§6).
  - G2-B2 — `AC-MG-018`의 파일 전체 해시 단언을 GLM 라우팅 키 집합 투영 비교로 바꾸고,
    `teammateMode`를 명시 제외했다.
  - G2-B3 — `internal/tmux`의 `sessionEnvHasGLM`·`hasGLMEnv`를 이관 목록에서 빼 형제
    `SPEC-MOAI-CG-RETIRE-001`(제안)에 넘겼다. 두 술어는 토큰 존재가 먼저인 이중 판정이고,
    프로덕션 호출 경로가 없으며, 다른 SPEC의 시험이 고정한 동작을 담는다. 이것으로 0.2.0의
    교차 렌즈 모순 C1도 해소되었다.
  - G2-B4 — `REQ-MG-005`에 플랫폼 무관한 종료 코드 전파 의무를 되살렸다.

  **권고 13건.** A1(배지 근거 문장, `REQ-MG-026`), A2(`AC-MG-014` 네 분기 명시), A3(0.2.0
  대응표 정정), A5(passthrough 게이트 표현, `design.md` §2.1), A6(M0 spike 하네스), A7(감시
  기구의 빌드 태그와 PID 재사용 방어), A8(`AC-MG-016`의 `withSessionPID` 형태 단언),
  A9·A10(`research.md` 측정 정정), A11(`moai gpt` 출시 판단 입력과 사용자 출력의 내부 식별자
  노출 금지), A12(`REQ-MG-016`에서 절차 문구 제거), A13(0.1.0 항목에 덧붙였던 0.2.0 내용 이전)을
  반영했다. **A4는 보류했다** — Windows 검증 방식을 운영자가 다시 결정하는 중이므로
  `REQ-MG-009`, `AC-MG-006`의 Windows 절반, `plan.md` M4, 그리고 같은 주제의 `design.md` §3.5와
  `research.md` §10을 이 판에서 고치지 않았다.

  **본문을 고친 항목.** 요구사항: 003, 005, 016, 021, 022, 026(과 007 묘비). 수용 기준: 014,
  016, 018, 021, 022, 023. 예산은 요구사항 25 / 수용 기준 25로 그대로다.

- 0.3.1 (2026-09-10) — 정정 일곱 건(C1~C7) 적용.

  **경위.** 아래 정정은 iter3 개정 지시에 담겨 내려왔지만 0.3.0 산출물에 착지하지 않았다. 이 판이
  그 일곱 건을 적용한다. 0.3.0에서 받아들여진 결정 — `MOAI_LAUNCH_PROVIDER`의 **존재**로 모든 gateway
  launch의 훅 재주입과 teardown을 억제하는 판정, 세션 접근 토큰 운반 키의 M0 이월, 복원된 백업 토큰을
  드러나는 실패로 기록한 잔여 위험, `REQ-MG-007` 묘비 — 은 바꾸지 않는다. 코드 인용은 모두 HEAD
  `d060e0d13`에서 저자가 다시 열어 확인했다.

  - C1 — 거짓 전제 교체. 0.3.0은 `moai glm`이 오늘 `injectGLMEnv`로 `settings.local.json`에 쓴다고
    적었다. 실제로는 `injectGLMEnv`의 비테스트 참조가 주석과 정의뿐이고, `applyGLMMode`는 그 파일을
    의도적으로 쓰지 않으며, `moai glm setup`도 쓰지 않는다. 현재 프로덕션 launch 경로 중 GLM 키를 그
    파일에 쓰는 것은 없고, 살아 있는 쓰기 주체는 SessionStart 훅 `ensureGLMCredentials` 하나다
    (`design.md` §5.2·§6.2, `research.md` §1.5).
  - C2 — `ensureGLMCredentials`의 발동 조건과 두 분기의 쓰기 내용을 코드 기준으로 기록했다. stale 키
    위험의 출처는 현재 코드가 아니라 옛 바이너리가 쓰거나 사람이 고친 **남은 상태**이며, 업그레이드
    사용자가 그 대상이다. launch 단계 정리는 여전히 필요하다(`research.md` §6.5).
  - C3 — 정리 키 집합을 `removeGLMEnv`에 고정하던 서술을 두 층 계약으로 바꿨다. 층 (a) 라우팅 키,
    층 (b) 쓰기 주체에서 유도한 동작 영향 키, 그리고 `moai cc`보다 적게 지우지 않는다는 비회귀 규칙이다.
    `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 **포함**한다 — 0.3.0이 열어 둔 결정의 답이다. 최종 집합은
    14키이며 용어를 "GLM 정리 키 집합"으로 바꿨다(`REQ-MG-021`, `design.md` §6.1).
  - C4 — 저장소의 네 GLM 삭제 목록은 같은 키 집합이 아니다. 측정한 차이를 표로 기록했고 어느 목록도
    정답으로 삼지 않았다. 세 파일 정리 함수 어느 것도 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`를 지우지 않는
    기존 간극은 관측으로만 남긴다(`research.md` §6.5).
  - C5 — SessionStart 억제가 `ensureGLMCredentials`의 **두 분기 모두**가 하는 파일 쓰기를 덮어야 함을
    명시했다. 토큰 있음 분기도 context 창 키 두 개를 쓴다(`REQ-MG-021`, `design.md` §6.3,
    `AC-MG-018` (c)).
  - C6 — `AC-MG-018` (a)에 남은 키 경로의 결정적 픽스처를 흡수했다. 새 AC는 만들지 않았다.
  - C7 — Windows 검증 방식을 운영자가 **release PR 게이트**로 결정했다. 0.3.0이 보류한 A4를 이 결정으로
    닫는다. 카드·develop CI에서는 이 시험이 Windows에서 돌지 않는다는 사실은 잔여 위험으로 기록했다
    (`REQ-MG-009`, `AC-MG-006`, `plan.md` M4·결정 4, `design.md` §3.5, `research.md` §10·§14).

  **본문을 고친 항목.** 요구사항: 009, 021, 022. 수용 기준: 006, 018. 예산은 요구사항 25 / 수용 기준
  25로 그대로다.

- 0.3.2 (2026-09-10) — `AC-MG-018` (c) 토큰 없음 픽스처 정밀화 두 건. T1: `MOAI_HOME`을 절대 경로로 두고 훅
  실행 전 `paths.GlmEnvFile()`이 격리 디렉터리 아래로 풀리는지 단언한다(`paths.go:69`의 `filepath.IsAbs`
  조건, `research.md` §6.5). T2: `.env.glm`을 `loadGLMKeyFromEnvFile`이 받는 `GLM_API_KEY=<값>` 형태로
  쓴다(`session_start.go:1355-1370`). 요구사항 25 / 수용 기준 25는 그대로다.

- 0.4.0 (2026-09-10) — plan-audit iter3 FAIL(0.84) 반영. 운영자가 Tier L 3회 상한에 한 번의 예외를 두어 이
  개정과 변경분 한정 감사를 허용했다. 모순되는 문장은 옆에 주석을 달지 않고 바꿔 썼다.

  - G3-B1 — 운영자 결정: tmux 세션 env에는 gateway 주소만 싣는다. tmux 세션 안의 gateway launch는 tmux GLM
    정리 키 집합(15키)을 지운 뒤 loopback `ANTHROPIC_BASE_URL`과 `MOAI_LAUNCH_PROVIDER`만 주입하고(운반 키가
    tmux로 전달되면 그 키만 더한다), GLM credential과 Z.AI 주소는 쓰지 않는다. 두 tmux 정리 함수의
    `ANTHROPIC_AUTH_TOKEN` 불일치는 훅 쪽 근거를 택해 해소했고, 지우는지 덮어쓰는지는 M0 운반 키 결정에
    맡겼다. signal이 있으면 `ensureTmuxGLMEnv`의 tmux 쓰기와 SessionEnd tmux 정리를 억제한다. gateway child는
    주소를 받은 teammate pane이 남아 있는 동안 서빙하고, 생존을 판정할 수 없으면 드러나는 실패를 계약으로
    삼는다. teammate의 tmux 세션 env 상속과 teammate pane 생존 판정은 미측정으로 두고 측정 과제 둘을 더했다
    (`REQ-MG-008`, `REQ-MG-018`, `REQ-MG-021`, `REQ-MG-022`, §A·§B·§C·§E, `AC-MG-006`, `AC-MG-018`,
    `AC-MG-022`, `design.md` §3.2·§3.4·§5.2·§5.4·§6.2·§6.5·§6.6, `plan.md` M4·M7·결정 8, `research.md`
    §6.6·§14).
  - 판독으로 확인한 사실 하나를 계약에 반영했다: `ensureTmuxGLMEnv`는 gateway launch 뒤에 저절로 멈추지 않는다.
    launch 정리가 백업 토큰을 복원한 파일에서는 발동 조건이 모두 참이 될 수 있어, 그 경우 억제가 유일한
    방어다(`design.md` §5.4).
  - G3-B2 — `AC-MG-018` (c)의 `cleanupGLMSettingsLocal` 억제 판정을, SessionEnd 시점에 `ANTHROPIC_BASE_URL`과
    대표 GLM 키가 남아 있는 픽스처와 signal 없는 대조군으로 바꿨다.
  - 권고 반영: A1(`ensureTeammateMode`의 legacy 키 삭제), A2(Windows 판정의 주체·시점·기록 위치와 7일 보존),
    A3(`progress.md`와 `research.md` 귀속 목록에 0.3.2·0.4.0), A4(`MOAI_HOME` 우회의 `filepath.IsAbs` 조건),
    A6(`REQ-MG-009`의 검증 장소 문구를 §E로 이동), A7("어느 하나라도"), A8(빈 백업 키 삭제와 `AC-MG-018` (a)
    변형), A9(복원 토큰 잔여 위험의 운반 키 조건), A10(`AC-MG-018` 하위 판정 독립성). A5는 이전 HISTORY 항목을
    고치지 않는다는 원칙에 따라 여기서 정정한다 — 0.3.0 항목이 적은 A5 반영 위치 `design.md` §2.1은 §2.2가
    맞다.

  **본문을 고친 항목.** 요구사항: 008, 009, 018, 021, 022. 수용 기준: 006, 018, 022. 요구사항 25 / 수용 기준
  25는 그대로다.

- 0.5.0 (2026-09-10) — 클라이언트 수준 실측 두 건과 운영자 결정 두 건 반영. 새 요구사항·수용 기준 번호는 없다.

  **입력.** 실제 Claude Code 2.1.267 TUI를 격리 `CLAUDE_CONFIG_DIR`와 `env -i` 아래에서, Anthropic Messages
  형식으로 답하는 Python loopback mock에 붙인 프로브 두 건이다. 둘 다 upstream에 닿지 않았다. Go gateway나 실제
  provider가 아니라 **클라이언트 동작**을 잰 것이다.

  - 프로브 1: `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`
  - 프로브 2: `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md`

  **실측으로 바뀐 전제.** 같은 세션 안 `/model` 전환은 클라이언트 수준에서 동작한다(§A). `/model <id>`는 선택
  시점에 검증 요청(`stream: false`, `max_tokens: 1`, user 메시지 하나)을 보내고, 401·404를 받으면 전환을 거부하며,
  성공하면 확인 대화상자 뒤 선택을 기본값으로 저장한다. picker `s` 경로는 검증 요청을 보내지 않고 저장하지도
  않는다(`REQ-MG-020`). `messages` 안에 `role: "system"` 항목이 섞여 온다(`REQ-MG-015`). 상세는 `research.md` §15.

  **운영자 결정 두 건 — 기존 항목에 접어 넣음.** 두 결정은 새 요구사항을 요구했지만 요구사항과 수용 기준이 모두
  Tier L 상한 25/25에 닿아 있다. 운영자는 새 번호를 만드는 대신 기존 요구사항을 정밀화하는 쪽을 택했다. 운영자
  지시에서는 두 결정을 D1·D2라 불렀으나 0.1.0 항목의 D1~D3과 겹치므로 여기서는 `plan.md` §H의 번호를 쓴다.

  - 결정 9 — gateway는 선택 시점 검증 요청을 로컬 상태만으로 답하고 upstream에 넘기지 않는다. `REQ-MG-023`에
    접었고 `AC-MG-003`에 판정을 더했다(`design.md` §4.1).
  - 결정 10 — 세 launcher는 초기 모델을 언제나 명시 `--model`로 넘기고, 사용자가 준 `--model`이 그 값을 대체한다.
    `REQ-MG-019`에 접고 `REQ-MG-002`와 교차 참조했으며 `AC-MG-001`에 판정을 더했다. 명시 `--model`이 저장된
    `model` 설정보다 우선하는지는 미측정이며 `plan.md` M1의 게이트 측정이다.

  **본문을 고친 항목.** 요구사항: 002(교차 참조), 015, 019, 020, 023. 수용 기준: 001·003(판정 추가), 005,
  009(입력 픽스처와 잔여 항목). 판정이 늘어난 수용 기준은 `AC-MG-001`과 `AC-MG-003` 둘이다. 요구사항 25 / 수용
  기준 25는 그대로다(`plan.md` §I).

- 0.6.0 (2026-09-11) — plan-audit iter4 FAIL(0.80, STOP 신호) 반영. 운영자 결정 세 건(2026-09-11)으로 두 표면을 형제
  SPEC 제안으로 떼어 내고 남은 코어 결함을 고쳤다. 상태는 `draft` 그대로다.

  **입력과 기준선.** iter4 감사 보고서 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter4.md`. 작업 트리
  `.claude/worktrees/moai-proxy-unified`, 브랜치 `WT-unified-gateway`, HEAD `ed71054d3`(이번 세션에서 `d060e0d13`으로부터
  fast-forward). 새 HEAD의 divergence `0 0`은 오케스트레이터가 이 개정 착수 전에 쟀다. 코드 인용 파일의 불변은 저자가
  다시 쟀다 — `git diff --stat d060e0d13 ed71054d3 -- <launcher·glm·spawn·launch_exec_posix·settings·session_start·
  session_end·glm_tmux·glmcred·paths·envkeys·ci.yml·release-pr-multi-os.yml>` → 빈 출력, 대조군 `git diff --shortstat
  d060e0d13 ed71054d3` → `1158 files changed`(전체 명령과 출력은 `research.md` §0). 이 판에서 새로 인용한 코드 행은 저자가
  `ed71054d3`에서 직접 열었다.

  **운영자 결정 세 건.** 운영자 지시의 A·B·C는 0.1.0의 D1~D3과 겹치지 않게 `plan.md` §H 번호로 적는다.

  - 결정 11(A) — picker·모델 선택 표면을 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로, tmux pane teammate 표면을
    `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 분리한다. 기존 `SPEC-MOAI-GPT-AUTH-001`·`SPEC-MOAI-CG-RETIRE-001`과 같은 방식으로
    §G Out of Scope 항목과 §H 형제 목록에 올렸다. 0.1.0 항목의 D1 문단도 형제 SPEC을 나열하지만, 앞선 HISTORY 항목은 고치지
    않는다는 원칙에 따라 그 문단은 그대로 두고 이 항목에 기록한다.
  - 결정 9 유지(B) — 선택 시점 검증 요청의 로컬 응답(`REQ-MG-023`, `design.md` §4.1, `AC-MG-003` (a)~(d))은 코어에 남는다.
    이유: 이 결정 없이 코어만 출시하면 `/model <id>`의 `max_tokens: 1` 검증 요청이 upstream으로 전달되어 전환마다 최소한의
    유료 호출이 생기고(`design.md` §4.1 잔여 위험 "형태 변경"), 사용자가 예상하지 못한 유료 경로를 두지 않는다는 확정
    방침(`REQ-MG-022`)과 충돌한다.
  - 결정 12(C) — gateway launch는 in-process teammate만 허용한다. `REQ-MG-021`에 접었고 새 번호는 없다.

  **분리 지도.**

  | 항목 | 행선 |
  |---|---|
  | `REQ-MG-020` 전체 | picker 형제 — 폐기 묘비. `~/.claude/settings.json` 쓰기 배제만 `REQ-MG-006`으로 옮겨 코어에 남김(`AC-MG-005`가 의존) |
  | `REQ-MG-019`의 "언제나 명시 `--model`" 조항(옛 결정 10) | picker 형제. 코어 `REQ-MG-019`는 공유 catalog와 초기 모델만 정하고 전달 방식은 정하지 않음 |
  | `design.md` §6.7 picker 구성과 "Default" 행 위험 | picker 형제. 코어에는 한 줄 포인터 |
  | `AC-MG-002`(T02) | picker 형제 — 폐기 묘비(번호 유지) |
  | `AC-MG-001` 저장된 기본값 판정, `AC-MG-005` picker `s` 판정 | picker 형제. `AC-MG-001`의 기본 판정은 코어에 남음 |
  | `plan.md` M1의 overlay 구성·`s` 재확인·`--model` 우선순위 게이트, M3의 명시 `--model` 전달, §H 결정 6·10 | picker 형제(§H에 이관 표시) |
  | `REQ-MG-008`의 teammate pane 수명 계약 | teammate 형제. 코어는 lead 세션 수명만 |
  | `REQ-MG-021`의 tmux 세션 env 주입 계약(15키 정리 집합, `ANTHROPIC_AUTH_TOKEN` 선택, 운반 키 tmux 전달) | teammate 형제. 코어는 tmux 세션 env 무기록과 훅 억제 |
  | `REQ-MG-018`의 tmux pane teammate GLM 경험 | teammate 형제. 코어는 알려진 UX 변화 한 가지만 적음 |
  | `AC-MG-006` teammate 수명 판정, `AC-MG-018` (a) tmux 주입 판정 | teammate 형제. 코어 `AC-MG-018` (a)는 무기록과 in-process 판정으로 바뀜 |
  | `plan.md` M7 teammate pane 측정(tmux env 상속, pane 생존), §H 결정 8 | teammate 형제(결정 8은 결정 12로 대체 표시) |

  **차단 결함 처분.**

  - G4-B1 (tmux 세션 env 소유권) — teammate 형제로 이관. 코어에서는 결정 12로 소거했다 — gateway launch는 tmux 세션 env에
    어떤 키도 쓰거나 지우지 않는다(`REQ-MG-021`, `design.md` §6.6, `AC-MG-018` (a)).
  - G4-B2 (검증 요청 인식 근거) — 코어에서 해소, `plan.md` M1 진입 게이트에 묶었다. 프로브 조건(억제 플래그 넷, `stream`의
    `bool` 기록, 원본 본문 부재)을 Gap으로 기록했고(`research.md` §15.6, §E), M1 첫 작업을 억제 플래그 없는 원본 캡처로
    정했으며, 인식 기준을 네 조건(`stream` 키가 있고 JSON `false`, `max_tokens` 정수 `1`, 메시지 하나이고 `role`이 `user`,
    `tools` 없거나 빈 배열)으로 좁혔다. `AC-MG-003` (c)에 `stream` 키 없음, `role`이 `user`가 아닌 단일 메시지, 도구 정의가
    있는 요청의 세 변형을 기대 결과(대상 provider에 정확히 한 번 전달)와 함께 더했다.
  - G4-B3 (빈 기본 모델·재개 경로의 `--model`) — picker 형제로 이관.
  - G4-B4 (picker 구성 방식 미결) — picker 형제로 이관.

  **권고 처분.**

  | 권고 | 처분 |
  |---|---|
  | G4-A1 `AC-MG-001` When 불일치 | 코어에서 수정 — "최초 turn 요청"으로 통일, 제목 생성 요청과 구분, 두 판정의 독립성 문장 추가 |
  | G4-A2 `--model=X` 결합 형태 | picker 형제 |
  | G4-A3 `s` 측정의 항목 종류 범위 | picker 형제 |
  | G4-A4 안내 의무의 AC 부재와 판정 불가능한 Where | picker 형제(`REQ-MG-020` 폐기) |
  | G4-A5 tmux에 남는 `ANTHROPIC_AUTH_TOKEN` 값 | teammate 형제. 코어 판정은 tmux 호출 `0`건이라 이 여지가 없다 |
  | G4-A6 흐름도의 4xx가 검증 요청 판정보다 앞섬 | 코어에서 수정 — `design.md` §4에서 검증 요청 판정을 registry·credential 판정 앞으로 옮기고 이유를 적음 |
  | G4-A7 `REQ-MG-008`의 "조용한 우회 경로 없음" | 코어에서 수정 — 부재 주장을 지우고, 연결 오류와 upstream 계수 `0`이라는 판정 가능한 의무(`AC-MG-006`)와 미측정 문장으로 바꿈 |
  | G4-A8 `CLAUDE_CONFIG_DIR` tmux 정리 | teammate 형제 |
  | G4-A9 teammate 모델 ID 음성 기준 | teammate 형제 |
  | G4-A10 `-p` 프로필의 `settings.json` | picker 형제 |

  **결정 12 구체화와 판독.** 운영자 지시는 "`teammateMode`가 `tmux`가 아닐 것"이었다. 코드는 tmux 밖에서 `auto`를 쓰고
  (`internal/hook/session_start.go:1021-1023`), Claude Code 문서는 `auto`가 tmux 세션 안에서 split pane을 연다고 적는다. 그래서
  요구사항과 `AC-MG-018` (a)는 `in-process`를 요구한다. 호출부 주석(`:628-630`)과 함수 주석(`:989`)이 코드와 다르다는 사실은
  `research.md` §16과 `design.md` §6.6에 적었다. "tmux 세션 env에 쓰지 않는다"는 지우기까지 포함하므로, gateway `moai cc`는
  오늘 `moai cc`가 launch 때 하던 tmux GLM 키 정리(`internal/cli/launcher.go:224`)도 하지 않는다. 이 변화는 `design.md` §6.5에
  잔여 위험으로 적고 형제 SPEC에 넘겼다. in-process teammate의 gateway 주소 상속은 미측정이며 `plan.md` M7의 측정 과제다.

  **본문을 고친 항목.** 요구사항: 006(쓰기 배제 이동), 008, 018, 019, 020(묘비), 021, 022, 023. 수용 기준: 001, 002(묘비),
  003, 005, 006, 018, 022, §D·§E. 요구사항은 24(묘비 `REQ-MG-007`·`REQ-MG-020` 제외), 수용 기준은 24(묘비 `AC-MG-002` 제외)다
  (`plan.md` §I). 미해결 clarification 표지는 0건이다.

- 0.7.0 (2026-09-11) — plan-audit iter5 FAIL(0.83) 반영. 운영자는 한 번의 한정 수정과 G5-B1·G5-B2·G5-B3 변경분 한정 재감사를
  허용했다. 상태는 `draft` 그대로다.

  **입력과 기준선.** iter5 감사 보고서 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md`. 작업 트리
  `.claude/worktrees/moai-proxy-unified`, 브랜치 `WT-unified-gateway`, HEAD `81c1d58f9`(오케스트레이터가 `ed71054d3`에서
  fast-forward). 저자가 fetch 없이 잰 divergence는 `0 0`이다(`progress.md` §E.1). 인용 코드 파일 가운데 두 커밋 사이에 바뀐 것은
  `internal/hook/session_end.go`와 `.github/workflows/ci.yml` 둘이며 그 행 인용을 옮겼다(명령과 출력은 `research.md` §0). 이 판에서
  새로 인용한 코드 행은 저자가 `81c1d58f9`에서 직접 열었다.

  **운영자 결정과 차단 결함 처분 (2026-09-11).**

  - G5-B1 (상속 GLM 키가 Claude child에 닿음) — 수정 (i). 세 launcher는 Claude child env를 조립하기 전에 상속 env에서 GLM 정리
    키 집합 14키를 모두 지우고, 이 정리를 launcher가 더하는 키보다 먼저 한다. `Z_AI_API_KEY`는 `moai cc`·`moai gpt`에서 지우고
    gateway `moai glm`만 GLM credential 저장소 값을 MCP 도구 인증용으로 싣는다 — Claude child env의 GLM credential 금지 조항에
    이 좁은 예외와 이유를 적었다. `ANTHROPIC_AUTH_TOKEN`은 운반 키 결정의 두 갈래를 모두 적었다. 반영: `REQ-MG-021`, §C, §E,
    `design.md` §3.2·§6.1·§6.2·§6.5, `AC-MG-018` (a), `plan.md` M7·결정 12.
  - G5-B2 (빈 기본값의 코어 단독 동작과 출시 결합) — (a). 코어의 노출과 출시는 `SPEC-MOAI-GATEWAY-PICKER-001`(제안) 착지를
    기다린다. 빈 기본값의 코어 단독 동작은 `REQ-MG-019`에 두 문장으로 적었다. 반영: `REQ-MG-019`, §G, `plan.md` M3·§A·결정 11,
    `design.md` §7.4, `AC-MG-001`.
  - G5-B3 (`REQ-MG-022`의 in-process 완화 과장) — (a), 새 플래그 없음. in-process 한정은 경로를 좁힐 뿐 닫지 못한다고 고치고
    잔여 셋(공유 파일 재기록, 미측정 재판독 시점, 치우지 않는 stale tmux GLM 키)과 닫는 형제 SPEC을 적었다. `--teammate-mode`는
    채택하지 않았다. 반영: `REQ-MG-022`, §G, `design.md` §6.5, `plan.md` 결정 12.

  **권고 처분.** G5-A1 — `AC-MG-018` (a)에 tmux 밖 변형과 그 대조군. G5-A2 — `AC-MG-003` (c)에 `stream: null`, `max_tokens` 0,
  `role`이 `system`인 단일 메시지 세 변형을 더해 여섯에서 아홉으로(`plan.md` M1 계수 포함). G5-A3 — `research.md` §15.5의 게이트
  포인터를 picker 형제로. G5-A4 — `plan.md` M1에 원본 캡처 픽스처 경로와 게이트 대조 결과 기록 경로.

  **모델 슬롯 키 결정 (2026-09-11, 운영자, 감사 전 한정 수정).** 0.7.0 작성 직후 모순 하나가 남아 있었다. `REQ-MG-021`은 상속
  모델 슬롯을 지우면서 launcher에 슬롯 키를 요구하지 않았고 `REQ-MG-018`의 "gateway 안쪽 tier 매핑"에는 장치가 없었다. 그러면
  gateway `moai glm`의 별칭 tier 요청이 Claude 모델 ID를 싣고 Anthropic route로 해석될 수 있었다(추론, 실행하지 않음). 운영자는
  gateway `moai glm` launcher가 GLM tier 슬롯 키를 설정하도록 정했고, 버전은 0.7.0 그대로다. `REQ-MG-021` 모델 슬롯 조항을
  다시 썼다 — 상속 정리 뒤 네 슬롯 키를 `llm.glm.models`의 `high`·`medium`·`low`·`fable` 값으로 더하고, `moai cc`·`moai gpt`는
  더하지 않는다. `REQ-MG-018`의 tier 매핑을 슬롯 키와 registry exact match로 구체화했고, 그 전제인 GLM catalog 등록을
  `REQ-MG-019`에 적었다(이전 판에는 GLM catalog 항목의 출처가 없었다). §E에 슬롯 해석이 미측정 클라이언트 동작이라는 줄을,
  §G picker 항목에 범위 한정을 더했다. 반영: `design.md` §1·§3.2·§6.1·§6.2·§6.5·§6.7, `AC-MG-018` (a), `AC-MG-025`, `plan.md`
  M2·M6·M7·결정 11·결정 12·§I, `research.md` §15.3.

  **본문을 고친 항목.** 요구사항: 018, 019, 021, 022. 수용 기준: 001, 003, 018, 025. 요구사항 24 / 수용 기준 24는 그대로다
  (`plan.md` §I). 미해결 clarification 표지는 0건이다.

## 0. 명칭 대응

설계 원문(`reports/moai-proxy-three-provider-redesign-20260910.md`), 핸드오프
(`reports/moai-proxy-next-session-handoff-20260910.md`), iter1 감사 보고서는 이 구성 요소를
"proxy"라 부르고 `REQ-MP-nnn` / `AC-MP-nnn`을 쓴다. 이 SPEC은 같은 구성 요소를
"gateway"라 부르고 `REQ-MG-nnn` / `AC-MG-nnn`을 쓴다. 보고서 파일명은 실제 읽기 전용
경로이므로 바꾸지 않는다. 상류 프로젝트 `ccmproxy`, `httputil.ReverseProxy`,
`internal/harness/router`의 `ConfigProxy`, `HTTP_PROXY`는 다른 대상을 가리키므로 원래
이름을 유지한다.

## A. Context / 왜 이 SPEC이 필요한가

지금 provider 선택은 **exec 시점의 프로세스 환경변수**로 결정되고, 한 세션은 backend를
하나만 갖는다. `setGLMEnv`가 `ANTHROPIC_AUTH_TOKEN`, `ANTHROPIC_BASE_URL`, 네 개의
`ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL`, `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1`,
`API_TIMEOUT_MS`, `Z_AI_API_KEY`를 프로세스 env에 심는다. tmux 세션 안에서는 `applyGLMMode`가 teammate
pane용으로 GLM credential, Z.AI base URL, tier 모델 슬롯을 **tmux 세션 env**에도 싣는다. `applyCCMode`는
반대 방향으로 `settings.local.json`과 tmux 세션 env에 남은 GLM 키를 걷어내는데, 두 삭제 목록은 서로 같지
않다. provider를 바꾸려면 세션을 다시 띄워야 한다.

이 SPEC은 그 환경변수 주입 지점을 loopback gateway 하나로 바꾼다. `ANTHROPIC_BASE_URL`이
`http://127.0.0.1:<port>`를 가리키면, 세션 안에서 `/model`로 고른 모델 ID가 실제 요청의
`model` 필드에 실려 오고 gateway가 그 ID를 registry exact match로 해석해 대상 provider로
보낸다는 것이 이 설계의 전제다. launch funnel 자체(`unifiedLaunch` →
`unifiedLaunchDefault`)는 그대로 둔다. 교체하는 것은 mode switch의 env 주입 단계이며, 프로세스 env와
tmux 세션 env 두 표면이 모두 대상이다. gateway launch는 tmux 세션 env에 아무것도 쓰거나 지우지 않고, Agent
Teams teammate는 in-process로만 띄운다(`REQ-MG-021`).
launch 전에 `settings.local.json`에 남은 GLM 키를 정리하는 단계는 세 launcher 모두에 **남기고 넓힌다**. 오늘
`moai cc`가 `removeGLMEnv`로 지우는 키를 모두 포함하고 그 함수가 지우지 않는 `CLAUDE_CODE_MAX_CONTEXT_TOKENS`까지
지운다(`REQ-MG-021`). tmux 세션 env의 정리와 소유권은 형제 SPEC 소관이다(§G).

**이 전제는 클라이언트 수준에서만 관측되었다.** 설계 보고서 §3은 비대화형 단발 요청 두 건에서
비-Claude 모델 ID가 `model` 필드로 전달되는 것을 관측했다. 0.5.0에서 실제 Claude Code 2.1.267 TUI를 Python
loopback mock에 붙인 프로브가 한 세션 안의 연속 전환을 관측했다. 세션 ID 하나에서, 전환 뒤의 turn 요청은 새 모델
ID와 이전 대화 기록을 함께 실었다(프로브 1은 `/model <id>`로 claude → gpt-5.6-sol → glm-5.3-flash → claude, 프로브
2는 picker `s`로 claude → gpt-5.6-sol, `research.md` §15). 이 관측은 mock이 모든 모델을 Anthropic 형식으로 받아
준 결과다. Go gateway를 거친 전환과 실제 provider 왕복은 아직 관측하지 않았으며, `plan.md` M1이 gateway 수준 재현을
첫 구축 마일스톤으로 세운다.

## B. 범위 요약

이 SPEC은 **gateway 코어**만 정의한다. 세 launcher의 CLI 계약(kanban/factory 진입 포함),
공통 supervisor, Anthropic 형태 ingress 정규화, 세 provider adapter, 공유 model registry,
launch 시점 provider signal과 `settings.local.json`·teammate 표시 계약, 실패 의미론, 보안 경계가
여기 들어간다. provider 중립 credential 참조 인터페이스는 이 SPEC이 고정하고, GPT 연결은 코어의 App Server managed/API 경로로 구현한다. 제공자별 picker는 코어 소관이고,
MoAI 자체 PKCE, `moai cg` 철거 스윕과 tmux pane teammate 표면은 §G의 제외 범위다.

## C. 용어

| 용어 | 정의 |
|---|---|
| launcher | 사용자가 입력하는 공개 실행 명령. `moai cc`, `moai gpt`, `moai glm` 세 개 |
| supervisor | gateway child를 띄우고 포트를 넘겨받은 뒤 Claude child를 exec하는 launch 경로 |
| ingress | gateway가 Claude Code로부터 받는 Anthropic 형태 HTTP 표면 (`/v1/messages` 외) |
| adapter | 특정 provider(Anthropic / OpenAI / Z.AI)의 upstream 호출을 담당하는 경계 |
| registry | 표시 모델 ID와 upstream 경로·인증 방식·capability를 잇는 세션 catalog |
| launch provider signal | launch 시점의 초기 provider를 자식 env로 운반하는 표지. base URL 문자열에 의존하지 않는다 |
| 초기 provider | launcher가 세션을 시작한 provider. 세션의 제공자 경계이며 `/model` 선택도 이 경계를 벗어나지 않는다 |
| GLM 정리 키 집합 | gateway launch가 exec 전에 `settings.local.json`의 `env` 블록에서 복원하거나 지우고, Claude child env를 조립하기 전에 상속 env에서 지우는 14개 키. 층 (a) 라우팅 키, 층 (b) 동작 영향 키, 비회귀 키로 이루어지며 기존 정리 함수 어느 하나의 목록과도 같지 않다. 목록은 `design.md` §6.1 |
| teammate 표시 | Agent Teams teammate를 lead 터미널 안(in-process)에 띄울지 tmux·iTerm2 split pane에 띄울지를 정하는 Claude Code 설정 `teammateMode`. gateway launch는 `in-process`만 허용한다(`REQ-MG-021`, `design.md` §6.6) |

## D. 요구사항 (GEARS)

### D.1 CLI 표면

**REQ-MG-001** (Ubiquitous) — `moai` CLI는 `gpt`를 `launch` cobra 그룹의 launcher로
등록해야 한다. launcher 이름은 단일 원천에서 오지 않으므로 등록은 **최소 세 곳**에 반영해야
한다: 도움말 정렬표 `helpGroupFrequency`의 `launch` 항목, `rootHelpGroups()`의 Launchers 행,
그리고 각 launcher 실행 함수 안에서 `spawnLaunch`에 넘기는 launcher 이름 리터럴. 한 곳이라도
빠지면 도움말 표시나 `--spawn` 재발행이 어긋난다.

**REQ-MG-002** (Ubiquitous) — `moai cc`, `moai gpt`, `moai glm` 세 launcher는 동일한
공통 launch plan을 거쳐야 하며, 기존 `-p/--profile`, `-w/--worktree`, `--spawn`,
`--model`, `--` 뒤 pass-through 인수의 동작을 그대로 보존해야 한다. 사용자가 준 `--model`은 launcher의 초기
모델을 같은 제공자의 허용 ID 안에서 대체한다(`REQ-MG-019`).

**REQ-MG-003** (Ubiquitous) — The CLI shall `moai gpt login`·`logout`·`status`를 명시 처리기로
연결하고 REQ-MG-017의 세 로그인 방식을 제공해야 한다. 미등록 동사는 명시 오류이며 launch 인수와 혼동해서는 안 된다.
로그인 취소·실패·지원하지 않는 설치 버전은 사용자 용어로 알리고 내부 SPEC 식별자를 드러내지 않아야 한다.

**REQ-MG-004** (Unwanted) — `moai` CLI는 `gg` 이름의 launcher를 등록해서는 안 되며,
help 출력과 command tree 어디에도 `gg`가 나타나서는 안 된다.

### D.2 공통 supervisor와 프로세스 수명

**REQ-MG-005** (Ubiquitous) — supervisor는 gateway를 **별도의 `moai` child 프로세스**로
실행해야 하며, gateway를 launcher 프로세스 내부 goroutine으로 두어서는 안 된다. supervisor는
현재 launcher가 각 플랫폼에서 제공하는 다섯 가지 보장 — `MOAI_SESSION_PID` 각인, 종료 코드
전파, signal 전달(Ctrl-C·Ctrl-Z·fg), PTY 동작, job control — 을 그대로 보존해야 한다.
POSIX에서는 Claude child를 띄우는 `syscall.Exec` 호출을 보존하며, POSIX의 이 보장들은 그
현재 플랫폼 계약은 `internal/cli/launch_exec_posix.go`의 `//go:build !windows`에서만 syscall.Exec를 사용하고,
`internal/cli/launch_exec_windows.go`의 `//go:build windows`에서는 자식 실행·대기·종료 코드 전달을 사용하는 것이다.
공통 파일에 무조건 syscall.Exec를 넣지 않는다. Windows 실행 검증은 GitHub CI에서 수행한다.
호출이 프로세스를 치환하는 데서 나온다(`MOAI_SESSION_PID` 각인은 현재 POSIX 경로에서만
이루어진다). **종료 코드 전파는 플랫폼과 무관하게 보존해야 한다** — Windows에서는
`execOrSpawnClaude`의 Windows 구현이 `child.Wait()`로 자식을 기다린 뒤
`os.Exit(ee.ExitCode())`로 자식의 종료 코드를 launcher 프로세스의 종료 코드로 삼는 현재
동작이 그 근거이며, 그 경로는 `REQ-MG-009`의 spawn-and-wait 경로다.

**REQ-MG-006** (Event-driven + Ubiquitous) — **When** gateway child가 loopback listener bind를 마치면,
supervisor는 bind된 포트를 포트 인계 채널로 돌려받고 그 값으로
`ANTHROPIC_BASE_URL=http://127.0.0.1:<port>`를 **자식 env에만** 조립한 뒤 exec를
수행해야 한다. 부모 shell의 환경은 바뀌지 않아야 한다. 시스템은 `~/.claude/settings.json`을 쓰기 대상에서 완전히
제외해야 한다. 이 쓰기 배제는 0.5.0까지 `REQ-MG-020`에 있던 불변식을 옮긴 것이며(`AC-MG-005`), Claude Code 자신이
`/model <id>` 선택을 설정 디렉터리에 쓰는 동작은 이 요구사항의 대상이 아니다.

**REQ-MG-007** — [RETIRED] 0.2.0에서 `REQ-MG-005`에 통합되어 폐기되었다. 번호는 재사용하지
않는다. 이 줄은 결번 표시일 뿐 요구사항이 아니며, GEARS 패턴을 갖지 않고 요구사항 예산 계수에서
제외한다.

**REQ-MG-008** (Event-driven) — **When** gateway child가 부모 감시로 lead Claude 세션이 어떤 원인으로든 종료했음을
감지하면, gateway child는 스스로 종료하며 listener와 임시 settings 파일을 정리해야 한다. 그 뒤 옛 loopback 주소로 도착하는
요청은 연결 오류로 드러나게 실패해야 하며, gateway child의 종료 경로는 어떤 upstream에도 요청을 보내서는 안 된다. 이 두
의무는 `AC-MG-006`이 연결 오류와 upstream mock 요청 계수로 판정한다. lead 프로세스 밖에서 그 loopback 주소를 가진
프로세스가 있는지 — in-process teammate가 lead의 gateway 주소를 물려받는지, lead 종료 뒤 남는지를 포함해 — 는 측정하지
않았으며, 이 요구사항은 그런 프로세스의 부재도, 그 요청이 다른 경로로 새지 않는다는 것도 주장하지 않는다(§E, `plan.md`
M7). tmux pane teammate의 수명 계약은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이다.

**REQ-MG-009** (Where Windows) — **Where** 플랫폼이 Windows이면, supervisor는 부모가
살아 있는 spawn-and-wait 경로(`execOrSpawnClaude`의 Windows 구현)를 사용하며 gateway
child의 수명을 그 부모 또는 job object에 묶어야 한다. POSIX 계약을 그대로 가져다 쓰지
않는다.

**REQ-MG-010** (Event-driven) — **When** loopback bind가 실패하면, gateway는 명시 오류로
종료해야 하며 비-loopback 주소로 대체 bind해서는 안 된다. 포트는 `0`(임의 여유 포트)로
요청한다.

### D.3 Anthropic 형태 ingress 정규화

**REQ-MG-011** (Ubiquitous) — gateway는 `/v1/messages` 요청 본문의 `model` 값을 registry
exact match로 해석해야 하며, 접두사 추측으로 upstream URL을 만들어서는 안 된다.
`model` 키 중복, 압축 전후 본문 크기 초과, 잘못된 JSON은 외부 송신 전에 거절해야 한다.

**REQ-MG-012** (Ubiquitous) — SSE 응답은 `message_start` → content block들 →
`message_delta` → `message_stop` 순서를 보장해야 한다. **When** upstream이 중간 EOF,
429, 취소로 끊기면 gateway는 성공 terminal event를 합성해서는 안 되며, 응답 bytes를 이미
보낸 뒤에는 재시도해서는 안 된다.
Where 검증된 GPT 구독 streaming 경로인 경우, the gateway shall 응답 media type이 없어도 제한된 SSE 형식과 완료를
엄격히 검증해야 하며, 임의 JSON·HTML·깨진 데이터나 충돌하는 명시 media type을 성공으로 수용해서는 안 된다.
When 전체 완료의 출력 배열이 비어 있으면, the gateway shall 앞서 유효하게 완료된 출력만 원래 순서로 사용해야 한다.
중복·ID/index 충돌·모순된 최종 출력·미완료 item·전체 완료 없는 EOF는 성공 종료로 바꾸어서는 안 된다.

**REQ-MG-013** (Ubiquitous) — tool_use / tool_result는 요청별 상태로 왕복해야 한다.
도구 이름 축약은 충돌을 감지하고 **요청 단위로** 역매핑해야 하며, 전역 tool ID map을
두어서는 안 된다. 여러 tool call이 섞여 들어오면 index와 ID를 구분해야 한다.

**REQ-MG-014** (Ubiquitous) — `/v1/messages/count_tokens`는 provider별 정확성 수준을
응답 또는 문서에 명시해야 하고 추정치를 실제 tokenizer 값으로 표시해서는 안 된다.
`/v1/models`는 세션 catalog만 노출해야 한다.
입력 추정은 전달되는 history·system·도구·출력 schema를 선언한 범위에 포함해야 하고 모델·인증 경로별 cap의 출처와
기본값/최대 override/여유분을 구분해야 한다. 출처 선언을 실제 계정 한도나 보장 tokenizer 값으로 표시해서는 안 된다. 등록되지 않은 path는 path별 정책으로
처리하며 일괄로 Anthropic에 전달해서는 안 된다.

**REQ-MG-015** (Event-driven) — **When** 세션의 허용 provider와 다른 provider를 요구하는 요청이 도착하면,
the gateway shall 외부 송신 전에 거절해야 한다. 다른 provider의 서명된 thinking / encrypted reasoning을
전달해서는 안 되며 미완료 tool pair도 명시 오류로 거절해야 한다.
Where GPT 대화가 App Server로 이어지는 경우, the gateway shall Codex가 소유한 thread 이력을 유지하며
새 사용자 입력과 도구 결과만 검증된 대화 위치에 한 번 반영해야 한다. 인증 계정·대화·thread·turn·call의 귀속이
불명확하면 외부 실행 전에 명시 오류로 거절해야 한다. reasoning 암호문을 직접 읽거나 Claude history에서 복원해서는 안 된다.
When pending 도구 호출 중 프로세스가 끊기면, the gateway shall 실행 여부가 불명확한 도구를 자동 재실행하지 않고
안전하게 재연결할 수 없는 상태를 명시 실패로 알려야 한다. 이미 끝난 HTTP 응답을 App Server turn 종료로 간주해서는 안 된다.
The gateway shall 같은 제공자 모델 변경·정상 resume·Claude Agent의 독립 대화·native Agent fork·명시 세션 fork·compaction을 대화 소유권과 이력 정합성 아래
지원해야 한다. 미확인 family 호환성이나 압축 표면을 추정하지 않고, 지원 검증 전 해당 상태는 명시 오류로 알려야 한다.
검증 전 오류 처리는 완료 기준을 낮추지 않으며 t653에서 해당 경로의 양성·음성 증거를 확보해야 한다.
When Claude가 압축용 요약을 요청하면, the gateway shall 정상 대화 요청으로 처리하고 실제 모델이 반환한 공개 요약을
Claude에 전달해야 한다. 별도의 압축을 중복 실행해서는 안 된다. The gateway shall 인증된 압축 완료 통지와
이미 반환한 요약의 일치를 확인한 뒤 공개 이력 기준을 재설정하고 새 입력만 한 번 반영해야 한다.
When 압축 완료 통지·요약·대화 귀속이 일치하지 않으면, the gateway shall 과거 입력을 다시 보내거나 임의의 요약으로
대체하지 않고 명시 실패로 알려야 한다. 압축 뒤 사실·도구 결과 회상과 새 프로세스 재개를 유지해야 한다.
The gateway shall 일반 자식의 독립 대화와 명시적인 세션 분기를 구별하고, 명시 분기에서는 승인된 원본 대화의
완료된 위치까지 이력을 보존해야 한다. 요청 내용으로 부모를 추측하거나 독립 자식 성공을 분기 성공으로 표시해서는 안 된다.
The gateway shall native Agent fork의 부모 이력 상속·분기 위치 보존·자식 격리를 양성 기능 의무로 유지해야 한다.
설치본에서 기능을 찾지 못한 경우는 가용성 전제 실패에 따른 NOT-RUN이며 기능 삭제나 통과를 뜻하지 않는다.
해당 양성·음성 증거를 확보하기 전 AS4 및 전체 지원 완료를 보류해야 한다. 일반 자식이나 명시 세션 분기의 성공으로
이를 대체해서는 안 된다. 독립적으로 구현 가능한 다른 AS4 경로는 계속 진행한다.

재구성의 입력에는 클라이언트가 `messages` 배열 안에 넣은 `role: "system"` 항목도 들어온다. Claude Code 2.1.267이
주입한 알림(예: `<total_tokens>…`)이 그런 항목으로 실려 오는 것이 실측되었다(`research.md` §15). gateway는 이
항목을 공개 text·tool pair와 구분되는 입력으로 다뤄야 하며, 그 처리 규칙은 결정적이고 문서화되어야 한다. 규칙은
M5에서 정한 결정적 규칙은 `design.md` §4.2에 있다. 대상 provider가 이 항목을 어떻게 받아들이는지는 측정하지 않았으며, 이
문장은 provider 동작에 대한 주장이 아니다.

### D.4 provider adapter

**REQ-MG-016** (Where 측정 통과) — **Where** Claude 구독 OAuth와 로컬 gateway 인증의
공존이 실측으로 확인된 경우에만, Anthropic passthrough adapter는 허용된 OAuth 헤더를
원래 Anthropic endpoint로 전달해야 한다. 측정 전에는 이 경로를 활성화해서는 안 되고
구독 보존을 지원 사실로 문서화해서도 안 되며, 측정 결과가 음성이면 Anthropic passthrough
adapter는 출시하지 않는다.
Where 검증된 native Anthropic policy인 경우, the gateway shall thinking·effort·출력 schema·문맥 보존 정책과 같은
provider의 native reasoning/signature 응답을 의미·순서대로 보존해야 한다. 입력 허용만 늘리고 응답 block을 제거하거나
thinking을 끄는 방식으로 성공을 만들어서는 안 된다.

**REQ-MG-017** (Ubiquitous + When) — The GPT adapter shall Claude Code UI를 유지하며 Codex App Server를
통해 GPT 구독 및 명시 API 키 모드를 제공해야 한다. GPT 허용 ID와 설치 경로의 지원 모델을 교차 검증하고
`gpt-6-astra`를 정확히 사용해야 한다. 미지원 ID를 다른 모델로 대체해서는 안 된다.
The launcher shall 로그인 방식으로 ChatGPT 브라우저 managed, ChatGPT device-code managed, OpenAI API 키를
제공하고 앞의 둘은 구독, 마지막은 API 과금임을 표시해야 한다. 구독 OAuth·토큰 저장·갱신은 Codex가 소유하며
MoAI는 구독 토큰 파일을 읽거나 직접 구독 backend를 호출해서는 안 된다. 구독 실패를 API 과금으로 자동 전환해서는 안 된다.

The adapter shall Claude가 제공한 도구의 이름·schema·발견 상태를 보존하고 도구 실행과 승인 권한을 Claude에
남겨야 한다. Codex의 native shell·file·MCP·agent·hook 실행은 이 경로에서 발생해서는 안 된다.
When 등록되지 않은 도구·변조된 tool_reference·다른 대화의 결과 또는 지원되지 않는 schema 변경이 발견되면,
the adapter shall 실행 전에 명시 오류로 거절해야 한다. 필요한 도구를 빼고 성공을 주장해서는 안 된다.

The adapter shall 실제 설치 버전의 capability와 공식 문서 차이를 명시하고 도구 왕복·상태 유지·모델 전환·재개·
압축의 지원 여부를 실제 실행으로 판정해야 한다. 기본 ToolSearch 상태에서 첫 요청만으로 전체 deferred 정의를 확보할 수 있다고 가정해서는 안 된다.
Claude 2.1.269의 확인된 픽스처에서는 첫 요청에 placeholder만 있고 실제 schema는 검색 뒤 요청에 추가되었다.
The adapter shall 운영자가 승인한 hybrid 방식으로 ToolSearch를 유지해야 한다. 최초 요청에서 완전한 정의가 있는
도구는 도구별 native schema로 등록하고, 검색 뒤 처음 발견한 도구만 공통 dispatcher로 연결해야 한다.
초기 도구까지 dispatcher로 바꾸거나 ToolSearch를 비활성화해서는 안 된다. 후발 도구의 schema·발견·실행 권한을
대화별로 검증하고 이름 충돌·변조·잘못된 인수는 실행 전에 거절해야 한다. 등록 갱신을 위해 thread를 재시작해서는 안 된다.
When 이후 요청의 ToolSearch가 이미 알려진 도구를 동일 schema로 다시 발견하면, the adapter shall 기존 등록과
실행 경로를 유지하며 멱등 처리해야 한다. 한 요청 안의 중복 참조 거절을 정상 반복 검색의 거절로 확대해서는 안 된다.
후발 도구의 모델 측 schema 표현은 native 등록과 같다고 주장하지 않으며 양성·음성 제품 검증을 완료해야 한다.
운영자 승인에 따라 구독과 API 키 모드 모두 App Server의 출력 정책을 사용한다. The adapter shall 이 정책이
Claude의 생성 토큰 상한과 같다고 표시해서는 안 된다. MoAI의 입력·출력 바이트와 취소 제한은 유지하되 동일한
생성 토큰 수 제한으로 표시하지 않아야 한다. 명시한 API 과금 및 선택 인증 방식을 유지하며 자동 과금 전환은 금지한다.
명목 context와 현재 App Server 유효 한도는 구분해야 하며 direct endpoint의 수용 실험을 이 경로의 보장으로 쓰지 않는다.
Claude metadata는 로컬 귀속에 필요한 범위로만 사용하고 GPT에 통째로 전달해서는 안 된다.

**REQ-MG-018** (Ubiquitous) — Z.AI GLM adapter는 기존 `moai glm` 사용자 경험을 보존해야
한다. 현재 환경변수로 처리하던 두 동작은 다음과 같이 옮겨야 한다. Anthropic beta 헤더 제거는 gateway 안쪽 책임이다.
tier별 모델 매핑은 두 장치가 함께 맡는다 — gateway `moai glm` launcher가 Claude child env에 더하는 GLM tier 모델 슬롯
키(`REQ-MG-021`)가 별칭 tier 요청의 `model` 값을 설정된 GLM 모델 ID로 만들고, 세션 registry가 그 ID를 exact match로
Z.AI route에 잇는다(`REQ-MG-011`, `REQ-MG-019`의 GLM catalog 항목). 이 보존에는 알려진 예외가 하나 있다. `moai glm` gateway launch의 Agent Teams teammate는
tmux pane이 아니라 lead 터미널 안(in-process)에 나타난다(`REQ-MG-021`). tmux 세션 안의 `moai glm`이 오늘 teammate를
tmux pane으로 띄우던 경험은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)이 착지할 때까지 제공되지 않으며, tmux pane
teammate의 GLM tier 매핑도 그 SPEC 소관이다.

### D.5 model registry, launch provider signal, settings 계약

**REQ-MG-019** (Ubiquitous + When) — The launcher shall 세션별 모델 선택과 요청 허용 범위를
`moai cc`는 Claude/Anthropic, `moai glm`은 GLM/Z.AI, `moai gpt`는 GPT/OpenAI로 제한해야 한다.
공통 catalog 원천을 재사용하더라도 각 세션에 노출·허용하는 ID 집합에는 해당 제공자만 있어야 한다.
GPT 집합은 `gpt-6-astra`를 포함하며 `gpt-6`를 그 모델의 별칭으로 임의 해석해서는 안 된다.
GLM 집합은 설정된 high·medium·low·fable ID의 중복을 제거한 집합이다.

The launcher shall 시작 모델, picker의 Default·현재 모델·선택 행, 보조 모델·fallback·네 tier 슬롯을
같은 제공자의 허용 ID로 한정해야 한다. 새 세션의 기본값은 설정된 Claude 기본값(비어 있으면 허용된 Claude
기본 모델), GPT는 `gpt-5.6-sol`, GLM은 high이다. 명시한 허용 모델은 이 기본값보다 우선한다.
재개는 같은 제공자·대화 소유권이 확인된 기록에 한정하고 명시 모델, 재개 모델, 제공자 기본값 순으로 결정한다.
When 다른 제공자 또는 미등록 ID가 명시 인수·직접 입력·보조 요청·재개에서 발견되면, the gateway shall
외부 송신 전에 명시 오류로 거절해야 하며 다른 제공자로 자동 우회해서는 안 된다.

The launcher shall 모델 선택·발견 캐시·재개 상태를 제공자 및 대화 소유 범위에 격리해야 한다.
`/model` Enter·직접 입력으로 기본값을 저장하거나 `s`로 세션 선택을 바꾸어도 다른 launcher 또는 일반 Claude의
기본값을 바꾸어서는 안 된다. 공유 사용자 설정과 프로젝트 설정을 모델 선택 목적으로 덮어써서는 안 된다.
When 상위 managed 정책과 제공자 전용 목록이 충돌하거나 목록이 비면, the launcher shall 충돌을 알리고
잘못된 제공자 목록을 정상 지원으로 표시해서는 안 된다. UI 제한과 별개로 gateway 허용 목록은 항상 적용한다.

**REQ-MG-020** — [RETIRED] 0.6.0에서 형제 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로 이관되어 폐기되었다. picker 이관은 당시 결정이며 t649의 회수 범위는 REQ-MG-019에서 규정한다. 이 요구사항에 있던 `~/.claude/settings.json` 쓰기 배제 불변식만
`REQ-MG-006`으로 옮겼다. 번호는 재사용하지 않는다. 이 줄은 결번 표시일 뿐 요구사항이 아니며, GEARS 패턴을 갖지 않고
요구사항 예산 계수에서 제외한다.

**REQ-MG-021** (Ubiquitous + While) — gateway launcher는 launch 시점의 초기 provider를
`internal/config/envkeys.go`에 등록한 환경변수 `MOAI_LAUNCH_PROVIDER`(값: `claude` |
`gpt` | `glm`)로 자식 env에 실어야 한다.

`.claude/settings.local.json`에 대해서는 다음 계약을 지켜야 한다.

- 세 gateway launcher는 `moai glm`을 포함해 모두, exec 이전에 `settings.local.json`의 `env`
  블록에서 GLM 정리 키 집합을 정리해야 한다. 이 집합은 다음 셋의 합집합이다.
  - 층 (a) 라우팅 키(필수) — `ANTHROPIC_BASE_URL`, `ANTHROPIC_AUTH_TOKEN`, 네
    `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` 슬롯. `MOAI_BACKUP_AUTH_TOKEN`이 비어 있지 않으면
    그 값으로 `ANTHROPIC_AUTH_TOKEN`을 복원하고, 없거나 비어 있으면 `ANTHROPIC_AUTH_TOKEN`을 지운다. 어느
    경우든 `MOAI_BACKUP_AUTH_TOKEN` 키 자체는 값이 비어 있어도 지운다.
  - 층 (b) 동작 영향 키 — GLM 쓰기 주체가 그 파일에 쓰는 키에서 유도한
    `CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS`, `API_TIMEOUT_MS`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW`,
    `CLAUDE_CODE_MAX_CONTEXT_TOKENS`.
  - 비회귀 키 — 어떤 gateway launch도 오늘 `moai cc`가 지우는 키보다 적게 지워서는 안 되므로
    `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `CLAUDE_CODE_TEAMMATE_DISPLAY`,
    `MOAI_STATUSLINE_CONTEXT_SIZE`도 지운다.

  이 정리는 옛 바이너리가 쓰거나 사람이 고쳐 남은 키가 `moai cc`·`moai gpt` 세션의 gateway base URL을
  덮어쓰거나 세션 동작을 바꿀 여지를 없애기 위한 것이다. 현재 프로덕션 launch 경로 중 이 파일에 GLM
  키를 쓰는 것은 없으므로, 정리 대상은 그렇게 남은 상태다.
- 어떤 gateway launch도, `moai glm`까지 포함해, `ANTHROPIC_BASE_URL`이나 GLM credential을
  `settings.local.json`에 써서는 안 된다. GLM credential과 Z.AI base URL은 gateway child에 전달되며,
  `settings.local.json`과 tmux 세션 env에는 실려서는 안 된다. Claude child env에도 실려서는 안 되며 예외는 하나다 —
  gateway `moai glm`의 Claude child env에는 launcher가 GLM credential 저장소에서 읽어 설정한 `Z_AI_API_KEY`가 MCP 도구
  인증용으로 실린다(아래 Claude child env 계약). 그 밖의 GLM credential과 GLM 정리 키 집합의 상속 값은 세 launcher의
  Claude child env 어디에도 실려서는 안 된다. Claude child의 base URL은 언제나 loopback gateway다.
- settings 파일의 `env` 값이 프로세스 env보다 우선하는지는 측정하지 않은 Claude Code 동작이다.
  위 정리는 그 우선순위가 어느 쪽이든 요구된다 — 우선순위를 측정하지 않았기 때문에 요구된다.

Claude child env에 대해서는 다음 계약을 지켜야 한다(2026-09-11 운영자 결정). 오늘 launcher는 자신이 물려받은 프로세스 env를
그대로 자식 env의 기반으로 쓴다(`design.md` §3.2). tmux 세션 env에 남은 GLM 키는 그 세션에서 새로 연 pane의 프로세스 env로
들어올 수 있으므로, 정리하지 않으면 그 pane에서 실행한 gateway launch의 Claude child까지 닿는다.

- 상속 env 정리 — 세 gateway launcher는 Claude child env를 조립하기 전에 상속 env에서 GLM 정리 키 집합의 14키를 모두
  지워야 한다. 프로세스 env에는 복원할 백업이 없으므로 `ANTHROPIC_AUTH_TOKEN`과 `MOAI_BACKUP_AUTH_TOKEN`도 값과 무관하게
  지운다. 이 정리는 상속 기반에만 적용하며, launcher가 더하는 키 — loopback `ANTHROPIC_BASE_URL`, `MOAI_LAUNCH_PROVIDER`,
  세션 접근 토큰, effort 키, gateway `moai glm`의 모델 슬롯 키, 그 밖에 launcher가 설정하는 키 — 보다 먼저 일어나야 한다. 그래서 launcher가 더한 키는 정리
  뒤에도 남는다.
- `Z_AI_API_KEY` — `moai cc`와 `moai gpt`는 상속된 `Z_AI_API_KEY`를 지워야 한다. `moai glm tools`가 기본값으로 사용자 범위
  `~/.claude.json`에 등록하는 Z.AI MCP 서버는 인증 헤더에 리터럴 `Bearer ${Z_AI_API_KEY}`를 두고 Claude Code가 그 값을
  프로세스 env에서 채우므로(코드 주석 판독, `design.md` §6.5), 이 키가 남으면 사용자가 GLM을 고르지 않은 세션이 유료
  Z.AI를 부른다(`REQ-MG-022`). gateway `moai glm`은 MCP 도구 인증을 위해 이 키를 Claude child env에 두되, 그 값은 상속된
  값이 아니라 launcher가 GLM credential 저장소에서 읽어 설정한 값이어야 한다.
- `ANTHROPIC_AUTH_TOKEN` — 세 launcher 모두 상속 값은 위 정리로 지워진다. M0에서 별도 세션 헤더를 선택했으므로
  launcher는 이 키에 세션 토큰을 다시 넣지 않으며 Claude child env에 `ANTHROPIC_AUTH_TOKEN`이 없다(`design.md` §3.2).
- 모델 슬롯 (2026-09-11 운영자 결정) — 세 launcher 모두 상속된 `ANTHROPIC_DEFAULT_*_MODEL` 값을 Claude child env에 싣지
  않는다(위 상속 env 정리). gateway `moai glm` launcher는 그 정리 뒤에 네 슬롯 키를 설정된 GLM tier 모델 ID로 더해야
  한다 — `ANTHROPIC_DEFAULT_OPUS_MODEL`에 `llm.glm.models`의 `high`, `ANTHROPIC_DEFAULT_SONNET_MODEL`에 `medium`,
  `ANTHROPIC_DEFAULT_HAIKU_MODEL`에 `low`, `ANTHROPIC_DEFAULT_FABLE_MODEL`에 `fable`. 오늘 `setGLMEnv`의 대응과 같다
  (`internal/cli/glm.go:367-370`, HEAD `81c1d58f9`). 그래서 별칭 tier 요청(`opus`·`sonnet`·`haiku`·`fable`)이 GLM 모델 ID를
  싣고 registry exact match로 Z.AI route에 해석된다(`REQ-MG-018`, `REQ-MG-019`). 슬롯 키가 없으면 그런 요청은 Claude Code의
  기본 Claude 모델 ID를 싣고 exact match가 그 요청을 Anthropic route로 보낼 수 있다 — Claude Code 문서의 슬롯 변수 설명에서
  끌어낸 추론이며 실행하지 않았다(§E). 이 슬롯 키는 launcher가 더하는 키이므로 정리 뒤에도 남고, `settings.local.json`과
  tmux 세션 env에는 쓰지 않는다. `moai cc`와 `moai gpt`도 정리 뒤 네 슬롯을 해당 제공자의 허용 ID로
  설정한다. Default·fallback·서브에이전트와 picker의 모델 범위는 REQ-MG-019를 따른다.

teammate 표시와 tmux 세션 env에 대해서는 다음 계약을 지켜야 한다(`plan.md` 결정 12). gateway launch에서는 이 계약이
`applyGLMMode`가 하던 tmux 세션 env 주입과 `applyCCMode`가 하던 tmux 세션 env 정리를 대체한다.

- in-process 한정 — gateway launch로 시작한 세션에서는, tmux 세션 안이든 밖이든, SessionStart 체인이 끝난 시점의
  `settings.local.json` 최상위 `teammateMode`가 `in-process`여야 한다. 시스템은 그 세션에서 `teammateMode`를 `tmux`나
  `auto`로 남겨서는 안 된다. Claude Code 문서는 `auto`가 tmux 세션 안에서 split pane을 연다고 적는다(`research.md` §16).
- tmux 세션 env 무기록 — gateway launch는 tmux 세션 env에 어떤 키도 쓰거나 지워서는 안 된다. 따라서 GLM credential,
  Z.AI base URL, loopback gateway 주소, `MOAI_LAUNCH_PROVIDER` 어느 것도 gateway launch를 통해 tmux 세션 env에 실리지
  않는다.

`MOAI_LAUNCH_PROVIDER`가 없는 세션의 `ensureTeammateMode`는 오늘 동작(tmux 안 `tmux`, 밖 `auto`)을 유지한다. Claude Code가
`in-process` 설정을 따르는지, in-process teammate가 lead의 gateway 주소를 물려받는지는 측정하지 않았다(§E).

GLM 활성 판정의 이관 대상은 두 지점이다.

- `internal/hook`의 `hookProcessEnvHasGLM` — 훅 프로세스 env의 `ANTHROPIC_BASE_URL`에
  `"z.ai"`가 들어 있는지만 보는 단일 부분 문자열 술어. `MOAI_LAUNCH_PROVIDER`가 `glm`인지를
  읽도록 옮긴다.
- `internal/hook`의 `cleanupGLMSettingsLocal` — `settings.local.json`의 `env` 블록에
  `ANTHROPIC_BASE_URL` 키가 있는지로 GLM 활성을 판정하는 존재 술어.

`MOAI_LAUNCH_PROVIDER`가 없는 세션에서는 두 지점 모두 기존 판정을 유지한다. **While** 세션이
gateway launch로 시작되어 `MOAI_LAUNCH_PROVIDER`가 설정되어 있으면, 다음 네 훅 동작은 실행되어서는 안 된다.

- SessionEnd의 `settings.local.json` GLM teardown(`cleanupGLMSettingsLocal`)
- SessionEnd의 tmux 세션 env 정리(`internal/hook`의 `clearTmuxSessionEnv`)
- SessionStart의 `ensureGLMCredentials`
- SessionStart의 `ensureTmuxGLMEnv`가 하는 tmux 세션 env 쓰기

그런 세션의 훅은 GLM 정리 키 집합을 `settings.local.json`에 써서는 안 되고, tmux 세션 env에 쓰거나 그 env에서
지워서도 안 된다. gateway launch는 tmux 세션 env에 아무것도 쓰지 않았으므로 그 세션의 훅이 치울 것이 없고, 남아 있는
키는 같은 tmux 세션의 다른 launch가 쓴 것이다. `ensureTmuxGLMEnv`의 억제는 launch 정리가 복원한 사용자 토큰이 tmux
세션 env로 새는 경로를 막는다(`design.md` §5.4).
`ensureGLMCredentials`의 억제는 그 함수가 파일에 하는 **모든** 쓰기를 덮어야 한다 —
토큰이 없을 때의 credential 주입뿐 아니라, 토큰이 있을 때 context 창 키
(`CLAUDE_CODE_AUTO_COMPACT_WINDOW`, `CLAUDE_CODE_MAX_CONTEXT_TOKENS`)를 쓰는 분기도 포함한다.

`internal/tmux`의 `sessionEnvHasGLM`과 `hasGLMEnv`는 이 SPEC의 이관 대상이 아니며 형제
`SPEC-MOAI-CG-RETIRE-001`(제안)이 맡는다. 이 signal은 exec 시점에 고정되는 launch 시점 값만
담는다 — 요청별 provider는 gateway 내부에서만 알 수 있다.

### D.6 실패 의미론과 보안

**REQ-MG-022** (Unwanted) — 시스템은 provider 간 암묵적 fallback을 수행해서는 안 되며,
사용자가 선택하지 않은 유료 API 경로로 요청을 흘려보내서는 안 된다. 사용자가 선택하지 않은
경로에는 이전 방식의 세션이나 옛 바이너리가 `settings.local.json`에 남긴 라우팅 키 때문에 요청이 gateway를
거치지 않고 upstream으로 곧장 가는 우회 경로도 포함되며, 그 경로는 `REQ-MG-021`의 launch
단계 정리가 막는다. tmux 세션 env에 남은 GLM 키 때문에 teammate pane의 요청이 gateway를 거치지 않고
Z.AI로 곧장 가는 경로도 포함된다. `REQ-MG-021`이 gateway launch의 teammate를 in-process로 한정해 이 경로를 좁히지만
닫지는 못한다. 남는 여지는 셋이다 — `teammateMode`는 프로젝트가 함께 쓰는 `settings.local.json`의 키라서 나중에 시작한
gateway 아닌 세션의 SessionStart(`ensureTeammateMode`)가 tmux 안에서 `tmux`로 다시 쓸 수 있고(`design.md` §6.6), 이미
떠 있는 gateway 세션이 그 값을 언제 다시 읽는지는 측정하지 않았으며, gateway launch는 tmux 세션 env에 남은 GLM 키를
치우지 않는다(`design.md` §6.5). 이 경로를 닫는 일은 형제 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안) 소관이고, 이 SPEC이
판정하는 것은 좁히는 의무(`AC-MG-018` (a))까지다. Claude Code가 `in-process` 설정을 따르는지도 측정하지 않았다(§E).

**REQ-MG-023** (Event-driven) — **When** 인증이 없거나 registry에 없는 모델 ID가
요청되면, gateway는 외부 실행 이전에 명시 오류로 거절해야 한다. 직접 secret을 사용하는 API·GLM 등은
provider 중립 credential 참조 seam으로 검증한다. App Server 경로는 secret 대신 관리 세션 권한을 검증해야 한다.
관리 세션 권한은 선택 프로필·계정 귀속·인증 모드·현재 세대·허용 모델의 일치를 확인하며, 누락·오래된 세대·
다른 계정·허용하지 않은 모델은 거절해야 한다. 가짜 credential이나 토큰 읽기로 이를 대신해서는 안 된다.

**When** 선택 시점 검증 요청으로 인식되는 요청이 도착하면, gateway는 그 요청을 어떤 upstream에도 전달하지 않고
로컬 상태만으로 답해야 한다.

- 요청한 모델이 세션 catalog에 없으면 HTTP 404 `not_found_error`로 답한다.
- 직접 secret route가 credential 참조 seam에서 credential을 얻지 못하거나 App Server route의 관리 세션 권한
  검증이 실패하면 HTTP 401 `authentication_error`로 답한다. 프로필·계정·세대는 현재 로컬의 검증된 상태로 확인한다.
- 그 밖에는 최소한의 비스트리밍 성공 message로 답한다.

어느 경우에도 다른 provider로 넘기거나 유료 호출을 해서는 안 된다. 인식 기준은 Claude Code 2.1.268의 원본 캡처로 보정한
명명된 가정 A-VAL-2.1.268(`design.md` §4.1)이며, 네 조건을 모두 만족하는 요청만 인식한다 — `stream` 키가 없거나 값이 JSON
`false`, `max_tokens`가 정수 `1`, `messages`가 길이 1이고 그 항목의 `role`이 `user`, `tools`가 없거나 빈 배열. 네 조건
가운데 하나라도 어긋난 요청은 로컬로 답해서는 안 된다. `stream: null`, 문자열 `"false"`, JSON `true`는 인식하지 않는다. 이 기준의 근거는 0.5.0 프로브가
아니라 `plan.md` M1 진입 게이트가 억제 플래그 없이 캡처한 원본 요청 본문이며, 캡처한 형태가 기준과 다르면 구현 전에 이
요구사항을 먼저 고친다(프로브의 한계는 `research.md` §15.6). 이 응답은 직접 route의 credential 존재 또는 App Server route의 로컬 관리 세션 권한만 확인한다.
현재 서버의 실제 모델 수용·만료·폐기 상태를 보증하지 않으며 서버 오류는 첫 실제 turn에서 드러날 수 있다. picker `s` 경로는 검증 요청을 보내지 않으므로
(`research.md` §15 F5), 그 경로의 첫 turn에서 credential이 없으면 위의 외부 송신 전 거절이 fallback 없이 드러나야 한다.

**REQ-MG-024** (Ubiquitous) — gateway는 loopback에만 bind해야 하고, credential을 argv와
기본 로그에서 제외해야 하며, 본문·API key·OAuth token·reasoning을 기본 로그에 남기지
않아야 한다. 새로 도입하는 환경변수 이름은 `internal/config/envkeys.go`에 상수로
등록해야 한다. 이 저장소에는 production Go의 맨몸 `ANTHROPIC_*` 리터럴을 잡아내는 AST 기반
가드 시험 `TestNoBareAnthropicEnvVarLiteralsInProduction`이 있으며, 이 가드는 빌드가 아니라
변경 패키지 테스트에서 실패한다.

**REQ-MG-025** (Unwanted) — The MoAI adapter shall Codex credential 파일의 토큰을 읽거나 직접
쓰거나 지워서는 안 된다. 사용자가 선택한 managed 로그인·갱신·로그아웃은 Codex가 자신의 저장소에서 수행한다.
기존 일반 Codex 프로필과 다른 계정·API 프로필을 묵시 변경해서는 안 된다. 승인된 인증 작업에 따른 Codex 자체 쓰기와
MoAI의 직접 파일 변경을 구분하여 검증해야 한다.

### D.7 kanban / factory 진입

**REQ-MG-026** (Event-driven) — **When** `moai gpt`가 `-k`(kanban) 또는 `-f`(factory)
진입 형태로 실행되면, launcher는 `moai cc`·`moai glm`과 같은 조건으로 Kanban Mode 또는
Factory Mode에 진입해야 하며 kanban 기록의 backend 값으로 `gpt`를 남겨야 한다. gateway
아래에서는 모든 세션이 `/model`로 provider를 바꿀 수 있으므로, kanban 기록의 backend 값은
세 값(`claude`·`glm`·`gpt`) 모두 **launcher의 초기 provider**를 뜻해야 하고 세션의 provider를
뜻해서는 안 된다. `gpt` 값은 kanban backend 집합에만 추가하며, 같은 이름의 상수를 쓰는 감사
backend 집합(`claude`·`codex`·`glm`)에는 추가해서는 안 된다. 웹 콘솔은 `gpt` 기록에 대해
과금 방식을 단정해 표시해서는 안 된다 — 현재 backend 배지(`backendBadge`)는 `glm`이 아닌
모든 값을 정액제로 표시하므로, 손대지 않으면 `gpt` lane에 검증되지 않은 과금 주장이 붙는다.
GPT 구독 경로와 API key 경로는 과금 방식이 다르며 그 판정은 형제 SPEC의 소관이다.

## E. 제약

- POSIX `syscall.Exec` 보존은 협상 대상이 아니다. `MOAI_SESSION_PID` 각인이 여기에
  올라타 있다.
- 새 환경변수 이름은 `envkeys.go` SSOT를 거쳐야 한다. 이를 어기면 변경 패키지 테스트의
  AST 기반 가드 시험이 실패한다 — `go build`는 이를 잡지 않는다.
- `internal/web`의 CSRF 계층(`Sec-Fetch-Site: same-origin`)은 재사용 대상이 아니다.
  Claude Code는 그 헤더를 보내지 않고 `/v1/messages`는 POST 전용이다.
- upstream SSE **중계** 선례가 저장소에 없다. `internal/web/events.go`의 Hub는 계약상
  신호 전용이다.
- OpenAI/GPT HTTP 클라이언트가 저장소에 전혀 없다.
- `settings.local.json`의 `env` 값과 프로세스 env 사이의 우선순위는 측정하지 않은 Claude Code
  동작이다. 이 SPEC의 어떤 문장도 그 우선순위를 전제로 삼지 않는다.
- Windows supervisor 계약(`REQ-MG-009`)의 실행 판정은 운영자 결정에 따라 release PR 게이트의 Windows
  레그에서 내린다. 판정의 계기가 되는 시험은 그 레그에서 건너뛰거나 빌드 태그로 제외하지 않는다. 판정은
  release PR 시점에만 생기며, 카드와 develop CI는 그 시험을 Windows에서 실행하지 않으므로 Windows 회귀는
  release PR에서야 드러난다(판정 절차와 기록 위치: `acceptance.md` `AC-MG-006`, `plan.md` 결정 4).
- in-process teammate가 lead 프로세스의 env(loopback `ANTHROPIC_BASE_URL`, `MOAI_LAUNCH_PROVIDER`)를 요청 경로와 훅
  env로 물려받는지, lead 종료 뒤 남는지, Claude Code가 `teammateMode: in-process`를 따라 split pane을 열지 않는지는
  측정하지 않았다. 근거는 코드 판독과 Claude Code 문서뿐이다(`research.md` §16). 이 SPEC의 어떤 문장도 그 셋을 전제로
  삼지 않으며, 셋은 `plan.md` M7의 측정 과제다. 상속이 음성이면 운영자 결정으로 돌아간다.
- gateway `moai glm` 세션의 Claude 프로세스 env에는 GLM credential 저장소에서 읽은 Z.AI 키가 `Z_AI_API_KEY`로 있다.
  Z.AI MCP 도구 인증을 위한 좁은 예외이며(`REQ-MG-021`), 그 세션의 Claude Code와 그 자식 프로세스가 이 값을 물려받을 수
  있다. Claude Code가 MCP 헤더의 `${Z_AI_API_KEY}`를 프로세스 env에서 채운다는 것은 코드 주석 판독이며 실행으로 확인하지
  않았다(`design.md` §6.5).
- gateway `moai glm`의 tier 매핑(`REQ-MG-018`)은 Claude Code가 별칭 `opus`·`sonnet`·`haiku`·`fable`을
  `ANTHROPIC_DEFAULT_*_MODEL` 슬롯 값으로 해석해 그 ID를 요청의 `model`에 싣는다는 클라이언트 동작에 기댄다. 이는 Claude Code
  모델 설정 문서의 환경변수 표에 적힌 동작이며(문서는 `haiku` 슬롯이 백그라운드 기능에도 쓰인다고 적는다), 이 SPEC에서
  측정하지 않았다. subagent의 별칭 요청이 실제로 이 해석을 거쳐 슬롯 값을 싣는지도 측정하지 않았다.
- 0.5.0의 클라이언트 실측(`research.md` §15)은 Python loopback mock에 붙인 Claude Code 2.1.267에서만, 네 억제 플래그
  (`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`, `DISABLE_TELEMETRY`, `DISABLE_ERROR_REPORTING`, `DISABLE_AUTOUPDATER`)를 켠
  채로 이루어졌다. 요청은 원본 본문이 아니라 mock이 파생한 요약으로 기록되었고 `stream`은 Python `bool` 변환값이다
  (`research.md` §15.6). 다음은 측정하지 않았으며, 이 SPEC의 어떤 문장도 이를 지원 사실로 전제하지 않는다: 검증 요청의
  실제 OpenAI Responses 변환, Z.AI endpoint,
  실제 provider 인증, 도구 왕복(T04), provider 간 reasoning·서명 처리(T15). 원본 형태와 억제 플래그 없는 요청 목록은
  0.8.0에서 2.1.268의 다섯 요청에 한해 추가 관측했다(`research.md` §17). 그 밖의 버전·요청 형태는 여전히 미측정이다.
  picker·모델 선택 경로의 미측정 항목(picker Enter 경로, 명시 `--model`과 저장된 `model` 설정의 우선순위 등)은
  t649에서 코어 검증으로 회수했다. 직접 Claude UI 실험과 실제 MoAI 통합 검증을 구분한다.

## F. 성공 기준

- §D의 24개 요구사항(폐기 묘비 `REQ-MG-007`·`REQ-MG-020` 두 줄은 요구사항이 아니므로 제외) 각각이
  `acceptance.md`에서 하나 이상 AC의 헤더 추적 괄호에 등장한다.
- 새 코드 경로의 테스트 커버리지가 85% 이상이다.
- `go vet ./...`, `golangci-lint run`, 변경 패키지 테스트가 통과한다.
- 실측하지 못한 항목은 PASS가 아니라 Gap으로 보고된다.

## G. 제외 범위 (out of scope)

이 SPEC은 아래 항목을 **구현하지 않는다.** 각 항목은 형제 SPEC 또는 별도 측정 축이며,
여기서는 의존·경계로만 이름을 붙인다.

### Out of Scope — MoAI 자체 GPT 구독 OAuth 및 direct backend

- MoAI 전용 OAuth 등록, 구독 토큰 파싱·저장·갱신, 비공개 구독 endpoint 직접 호출.
- Codex UI로 Claude Code를 대체하거나 Codex native 도구에 Claude의 실행·승인 권한을 넘기는 기능.
- 기존 AUTH 형제의 자체 credential 방식은 역사적 설계다. App Server managed/API 연결은 현재 코어 t650 소관이다.


### Out of Scope — `moai cg` 철거 스윕 (`SPEC-MOAI-CG-RETIRE-001` 제안)

- live runtime 제거(`internal/cli/cg.go`, `applyCGMode`, `TeamModeCG`, 거부 경로)
- `internal/tmux`의 `sessionEnvHasGLM`·`hasGLMEnv`의 처리(삭제 또는 이관). 두 술어는 토큰 존재가
  먼저인 이중 판정이고, 비테스트 호출자는 `IsCGMode` 하나이며 `IsCGMode`에는 비테스트 호출이 없고,
  토큰 판정 쪽은 다른 SPEC의 시험이 고정한 동작이다
- 기존 프로젝트의 `team_mode: cg` 마이그레이션 정책
- docs-site 4-locale 스윕, 배포 template 스윕, README 스윕, 테스트 재기준화
- 설계 보고서의 T20
- 분리 근거: docs-site 114개 파일 303건이 같은 PR 4-locale 의무 아래 있고, template
  73건, 테스트 18개 파일, 교차 SPEC 계약 7건(그중 1건은 이미 superseded)이 딸린 별도 카드
  규모의 작업이다.

### Out of Scope — 제공자 간 대화 전환과 임의 모델 등록

- 한 launcher 세션에서 다른 제공자로 전환하는 기능과 미등록 모델을 임의로 등록하는 기능.
- PICKER 형제로 이관했던 제공자 전용 목록·Default·저장·재개·슬롯 범위는 t649에서 코어로 회수했다.
  이 기능의 출시는 형제 문서의 존재가 아니라 본 SPEC의 실제 제품 검증 완료에 달려 있다.

### Out of Scope — tmux pane teammate 표면 (`SPEC-MOAI-GATEWAY-TEAMMATE-001` 제안)

- gateway launch 아래에서 tmux pane teammate를 되살리는 연결 방식 — tmux 세션 env 주입, 15키 tmux GLM 정리 키 집합, 두
  tmux 정리 함수의 `ANTHROPIC_AUTH_TOKEN` 불일치 해소, 세션 접근 토큰 운반 키의 tmux 전달(옛 `REQ-MG-021` tmux 계약, 옛
  `plan.md` 결정 8)
- 같은 tmux 세션의 여러 gateway에 대한 tmux 세션 env 소유권 모델(iter4 G4-B1 — `--spawn`이 호출자의 tmux 세션에 새 창을
  연다), 마지막 gateway 뒤에 남는 signal, gateway 아닌 launch가 남긴 stale tmux GLM 키를 누가 치울지
- teammate pane 수명과 생존 판정(옛 `REQ-MG-008` teammate 계약, 옛 `AC-MG-006` teammate 판정), teammate pane의 env 상속 측정
- tmux pane teammate 요청의 모델 ID와 GLM tier 매핑(옛 `REQ-MG-018` tmux 경험 조항)
- iter4 권고 G4-A5(tmux에 남는 `ANTHROPIC_AUTH_TOKEN` 값 판정), G4-A8(`CLAUDE_CONFIG_DIR`을 tmux에서 지우는 영향),
  G4-A9(teammate 모델 ID의 음성 기준)
- `REQ-MG-022`가 좁히기만 한 경로를 닫는 일 — tmux pane teammate 요청이 tmux 세션 env에 남은 GLM 키로 Z.AI에 곧장 가는
  경로(iter5 G5-B3)
- 분리 근거: tmux 세션 하나에는 env가 하나뿐이어서 소유권 설계 없이는 계약이 성립하지 않는다(G4-B1). 코어는 그 설계를
  기다리지 않도록 gateway launch를 in-process teammate로 한정한다(`REQ-MG-021`).

### Out of Scope — kanban 기록 필드 통합

- 기존 `MOAI_KANBAN_BACKEND`와 새 `MOAI_LAUNCH_PROVIDER`를 하나로 합치는 작업
- 이유: 두 키는 같은 값 어휘를 쓰지만 설정 범위가 다르다(전자는 kanban/factory 경로에서만
  실린다). 통합은 이 SPEC의 판정 표면을 넓힌다.

### Out of Scope — 과거 기록 재작성

- 과거 SPEC, release note, 감사 증거의 일괄 치환
- 이유: 역사 자료이며 현재 상태를 서술하지 않는다.

### Out of Scope — `moai gg` 제거

- 대상이 존재하지 않는다(§HISTORY D3의 작성 전 측정). 제거 요구사항 대신
  `REQ-MG-004`의 부재 회귀 가드만 둔다.

## H. 교차 참조

- 설계 원문: `reports/moai-proxy-three-provider-redesign-20260910.md` (읽기 전용, 파일명 유지)
- 핸드오프: `reports/moai-proxy-next-session-handoff-20260910.md` (읽기 전용, 파일명 유지)
- iter1 감사: `.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md` (옛 경로·옛 식별자 보존)
- iter2 감사: `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md`
- iter3 감사: `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter3.md`
- iter4 감사: `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter4.md` (0.6.0 입력)
- iter5 감사: `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md` (0.7.0 입력)
- 코드베이스 조사: `research.md`
- 내부 경계 설계: `design.md`
- 클라이언트 실측(0.5.0): `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`,
  `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md` (기계 로컬 경로, 요약은 `research.md` §15)
- Claude Code agent teams 문서(`teammateMode`): `https://code.claude.com/docs/en/agent-teams` (2026-09-11 판독, 요약 `research.md` §16)
- 형제 SPEC(제안): `SPEC-MOAI-GPT-AUTH-001`, `SPEC-MOAI-CG-RETIRE-001`, `SPEC-MOAI-GATEWAY-PICKER-001`(0.6.0), `SPEC-MOAI-GATEWAY-TEAMMATE-001`(0.6.0)

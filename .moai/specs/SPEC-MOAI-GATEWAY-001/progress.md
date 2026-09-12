# SPEC-MOAI-GATEWAY-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- 0.8.0 M5 문서 보강: system 위치·schema·종료 매핑 결정은 design §4.2~4.3, opaque reasoning 운반은 PROBE ONLY다.
  전체 carrier 유실 탐지 계약과 실제 TUI/resume 게이트는 미충족이며 run 증거로 세지 않는다. 새 REQ/AC 번호는 없다.

- 산출물: `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md` (Tier L 5종)
- 개정: 0.8.0 — 사용자가 M1 불일치 보정과 구현 계속을 승인했다. A-VAL-2.1.268은 `stream` 키 부재 또는 JSON `false`와
  나머지 세 조건을 함께 요구한다. AC-MG-003의 음성 아홉 변형은 유지하고 키 부재·명시 `false`를 각각 양성 대조군으로 둔다.
  G6-A1 프로세스 env 픽스처는 14키 전부와 Z_AI 키를 오염시킨다. REQ/AC 번호·계수는 유지한다.
  이번 변경분 재감사와 제품 검증은 별도 게이트이며, 아래 §E.2의 과거 0.7.0 측정 기록을 바꾸지 않는다.
  M0 재시험은 2026-09-11 19:00 Asia/Seoul 이후 Opus 5·Sonnet 5 기준이다. T09는 여전히 미충족이다.
  0.7.0 — plan-audit iter5 FAIL(0.83) 반영. 운영자는 한 번의 한정 수정과 G5-B1·G5-B2·G5-B3 변경분 한정 재감사를
  허용했다. 운영자 결정(2026-09-11): G5-B1 수정 (i) — 세 launcher가 Claude child env를 조립하기 전에 상속 env에서 GLM 정리 키
  집합 14키를 지우고, `Z_AI_API_KEY`는 `moai cc`·`moai gpt`에서 지우며 gateway `moai glm`만 GLM credential 저장소 값을 MCP
  도구 인증용으로 싣는다(`REQ-MG-021`, `AC-MG-018` (a)). G5-B2 (a) — 코어의 노출과 출시를 `SPEC-MOAI-GATEWAY-PICKER-001`(제안)
  착지에 묶었다(`plan.md` M3, `design.md` §7.4, `REQ-MG-019`). G5-B3 (a) — `REQ-MG-022`가 in-process 한정이 tmux pane의 Z.AI
  경로를 좁힐 뿐 닫지 못한다고 적고 잔여 셋을 명시했다(`--teammate-mode`는 채택하지 않음). 권고 G5-A1~A4 반영.
  같은 날 감사 전에 운영자 결정으로 gateway `moai glm` launcher가 상속 정리 뒤 GLM tier 슬롯 키 넷을 `llm.glm.models` 값으로
  더하게 했다(`REQ-MG-021`·`REQ-MG-018`, GLM catalog 등록 `REQ-MG-019`, `AC-MG-018` (a)·`AC-MG-025`).
  0.6.0 — plan-audit iter4 FAIL(0.80, STOP 신호) 반영. 운영자 결정 세 건(2026-09-11): picker·모델 선택 표면을
  `SPEC-MOAI-GATEWAY-PICKER-001`(제안)로, tmux pane teammate 표면을 `SPEC-MOAI-GATEWAY-TEAMMATE-001`(제안)로 분리(`plan.md`
  결정 11), 결정 9를 M1 진입 게이트와 함께 코어에 유지, gateway launch는 in-process teammate만 허용(결정 12). 차단 4건 처분:
  G4-B1 teammate 형제로 이관하고 코어에서는 결정 12로 소거, G4-B2 코어에서 M1 진입 게이트로 해소, G4-B3·G4-B4 picker 형제로
  이관. 권고 10건(G4-A1~A10) 처분은 `spec.md` HISTORY 0.6.0.
  0.5.0 — 클라이언트 수준 실측 두 건(Claude Code 2.1.267 × Python loopback mock)과 운영자 결정 두 건
  (`plan.md` 결정 9·10) 반영. 두 결정은 운영자 결정에 따라 기존 요구사항에 접었다 — 결정 9는 `REQ-MG-023`·
  `AC-MG-003`, 결정 10은 `REQ-MG-019`(`REQ-MG-002` 교차 참조)·`AC-MG-001`. 새 요구사항·수용 기준 번호는 없다.
  0.4.0 — plan-audit iter3 FAIL(0.84) 반영. 운영자가 Tier L 3회 상한에 한 번의 예외를 두어 개정 1회와
  변경분 한정 감사를 허용. 차단 2건(G3-B1 tmux 세션 env 계약, G3-B2 `cleanupGLMSettingsLocal` 억제 판정의
  공허) 처리, 권고 10건(G3-A1~A10) 처분, 잔여 3건 반영.
  0.3.2 — `AC-MG-018` (c) 토큰 없음 픽스처 정밀화 두 건(`MOAI_HOME` 절대 경로 단언, `.env.glm` 형태).
  0.3.1 — iter3 개정 도중 내려졌으나 0.3.0에 착지하지 않았던 정정 일곱 건(C1~C7) 적용.
  0.3.0 — plan-audit iter2 FAIL(0.80) 반영. 차단 5건(G2-MP1, G2-B1~B4) 처리, 권고 13건 중 12건 반영.
  0.3.0에서 보류한 A4(Windows)는 0.3.1에서 운영자 결정(release PR 게이트)으로 닫힘
- 감사 보고서: iter1 `.moai/reports/SPEC-MOAI-PROXY-001/plan-audit-iter1.md` (옛 경로·옛 식별자 보존),
  iter2 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter2.md`,
  iter3 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter3.md`,
  iter4 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter4.md`,
  iter5 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter5.md`,
  iter6 `.moai/reports/SPEC-MOAI-GATEWAY-001/plan-audit-iter6.md` — PASS, 변경분 점수 0.92
- 0.5.0 클라이언트 실측 증거(기계 로컬, 버전 관리 제외 경로):
  `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe/README.md`,
  `.moai/state/verify/25b43a41-c0ac-4110-9dad-f3983da6a527/gwprobe2/README.md` — 요약 `research.md` §15
- 작성 기준선: 작업 트리 `.claude/worktrees/moai-proxy-unified`, 브랜치
  `WT-unified-gateway`, HEAD `81c1d58f9` (0.7.0 재기준, 오케스트레이터가 `ed71054d3`에서 fast-forward. 0.6.0은 `ed71054d3`,
  0.5.0까지는 `d060e0d13`. 저자가 fetch 없이 잰 `git rev-list --count --left-right origin/develop...HEAD` → `0 0`, 로컬 `origin/develop` = `81c1d58f9`) <!-- moving-ref-ok: origin/develop is the SUBJECT of this identity reading, not its anchor; the anchor is the pinned HEAD 81c1d58f9 on the preceding line, and the criterion is the measuring command with a dated 2026-09-11 reference -->
- 요구사항 24 (그 밖에 폐기 묘비 `REQ-MG-007`·`REQ-MG-020` 두 줄, 계수 제외) / 수용 기준 24 (그 밖에 폐기 묘비 `AC-MG-002` 한 줄, 계수 제외) (Tier L 상한 25 / 25, 독립 적용) — 0.6.0에서 각 축 하나씩 형제 SPEC으로 이관
- clarification 항목: 운영자 결정으로 전부 확정 — `plan.md` §H (결정 12건. 결정 6·10은 형제 SPEC으로 이관, 결정 8은 결정 12로 대체). 미해결 clarification 표지 0건
- 상태: 0.7.0의 G5-B1·G5-B2·G5-B3 변경분 한정 6차 감사 PASS(0.92). 제품 구현 통과를 뜻하지 않는다.
- 구현 진행 승인: 2026-09-11 사용자의 작업 재개·계획 완료 지시를
  [재개 기록](../../reports/SPEC-MOAI-GATEWAY-001/codex-resume-20260911.md)에 기록했다.
  M0·M1 진입 게이트를 실측한 뒤 M1 중단 조건이 관측되었다. 사용자가 보정을 승인하여 0.8.0에 반영했다.
  의존 구현은 변경분 감사와 추적 픽스처 재판정 뒤 재개한다. 아래 실행 기록은 0.7.0 당시의 판정이다.

## §Mode Selection

- 마일스톤 구현은 순차 진행하고, 서로 독립적인 읽기 전용 준비·검증만 병렬로 진행한다.
- 현재 단계는 M0 인증 측정과 M1 원본 요청 캡처 게이트다. 검증 요청 인식기와 제품 gateway 구현은 M1 게이트 해결 뒤에 착수한다.

## §E.2 Run-phase Evidence

측정 기준선: `WT-unified-gateway`, HEAD `81c1d58f9`, Claude Code `2.1.268`.
세션 근거 디렉터리는 `.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/`다.

| 항목 | Actual Output | Status | 근거·판정 범위 |
|---|---|---|---|
| M0 — 구독 OAuth와 세션 인증 공존 | `custom: timed_out=true`, upstream `429` 14건; `auth-token: timed_out=true`, upstream `401` 14건 | INCONCLUSIVE | `m0-result.json`; [M0 보고서](../../reports/SPEC-MOAI-GATEWAY-001/m0-auth-gate.md). OAuth 정상 응답·refresh 공존 양성으로 판정할 수 없다. |
| M1 — 실제 TUI 원본 요청 캡처 | `version: 2.1.268 (Claude Code)`, `suppression_flags_present: []`; `/v1/messages?beta=true` 5건 | BLOCKED | `m1-conditions.json`, `m1-summary.json`; [M1 보고서](../../reports/SPEC-MOAI-GATEWAY-001/m1-capture-gate.md). 아래의 명시적 중단 조건을 관측했다. |
| M1 — 선택 시점 검증 요청 | `request-003.json`: `model=gpt-5.6-sol`, `max_tokens=1`, user 메시지 하나, tools 키 없음, **stream 키 없음** | FAIL | `m1-raw/request-003.json` 원문 직접 확인. `design.md` §4.1의 A-VAL-2.1.267은 `stream` 키 존재와 JSON `false`를 요구한다. `plan.md` M1에 따라 인식 기준을 먼저 SPEC에서 고쳐야 한다. |
| M1 — 같은 세션의 첫 turn·후속 turn | `MOCK_OK:claude-sonnet-4-5` → `/model gpt-5.6-sol` → `MOCK_OK:gpt-5.6-sol` | OBSERVED | `m1-screen-final.txt`, `m1-summary.json`. 실제 TUI와 loopback mock의 관측이며 제품 gateway 검증은 아니다. |

- Gaps: OAuth refresh와 정상 응답, 네 번의 모델 전환, 제품 gateway 동작, 전체 AC·불변식의 실행 결과는 아직 확보하지 못했다. 로컬 원본 캡처를 추적 testdata로 승격한 것으로 간주하지 않는다.
- Residual-risk: `stream` 키 없는 선택 검증 요청을 현재 인식 기준으로 처리하면 통상 upstream 경로로 흘러갈 수 있다. 이는 제품에서 재현한 결함이 아니라 캡처와 SPEC 조건의 불일치다. 제품 인식기를 만들기 전에 SPEC 수정·재판정이 필요하다.

### 2026-09-11 — 0.8.0 M1 인식기·M2 기반

- Claim: 익명화 실제 요청 5건, 인식기, immutable catalog, provider 중립 CredentialRef 계약 구현.
- Evidence: `go test ./internal/gateway/...` → `ok .../internal/gateway 0.403s`; `go test -race ./internal/gateway/...` → `ok .../internal/gateway 1.427s`; `go vet ./internal/gateway/...` exit 0. 합산 coverage 96.2%. 명령 전문과 실제 출력·RED는 [foundation-verification.md](../../reports/SPEC-MOAI-GATEWAY-001/foundation-verification.md)에 보존했다.
- Baseline-attribution: 이번 WT `WT-unified-gateway`, HEAD `81c1d58f9`, SPEC 0.8.0.
- Gaps: 전체 AC·M1 네 번 모델 전환·M0 OAuth 정상 응답·refresh·HTTP 송신 판정·실제 credential 구현은 미완료. 커밋 전이므로 status 전이는 하지 않았다.
- Residual-risk: capability 지원 수치는 미선언이며 HTTP 정책과 credential provider 검증은 다음 구현의 책임이다. 제품 완료 판정은 발행하지 않는다.

### 2026-09-11 — HTTP ingress 범위 검증

- Claim: 본문 읽기 전 세션 인증, 경로 정책, 압축 전후 크기, JSON 중복 거절, 로컬 validation/count/models, 요청별 adapter 호출 경계를 구현했다.
- Evidence: 최종 `go test -coverpkg=./internal/gateway/... -coverprofile=/tmp/gateway-ingress-cover.out ./internal/gateway/...` → `ok .../internal/gateway 0.844s coverage: 96.5%`; race → `ok .../internal/gateway 1.525s`; vet와 gopls exit 0. [ingress-verification.md](../../reports/SPEC-MOAI-GATEWAY-001/ingress-verification.md)에 실제 RED/GREEN 출력과 API 계약을 보존했다.
- Baseline-attribution: WT-unified-gateway, HEAD 81c1d58f9, SPEC 0.8.0, 신규 foundation 위에서 측정했다.
- Gaps: 실제 socket·upstream, 변환 adapter, M0 인증·TUI 전환, 전체 AC는 미완료다.
- Residual-risk: count_tokens는 추정이며 실제 토크나이저가 아니다. 실제 송신 직전 Generation 재검사와 history 정규화는 adapter 구현에서 보장해야 한다.

### 2026-09-11 — M4 supervisor 기반

- Claim: 별도 child, private stdin config/stdout bound-address 인계, 부모 PID+homestate 지문 감시, 필수 수명 상한, Stop/Wait, child 소유 overlay 정리를 구현했다.
- Evidence: gateway/auth 한정 coverage 시험 `ok .../internal/gateway 7.333s coverage: 93.5%`; race `ok .../internal/gateway 9.619s`; vet·gopls·Windows amd64 시험 바이너리 컴파일 exit 0. 수정 파일별 coverage는 supervisor.go 93.5%, runner 86.7%, POSIX 100%. 전체 명령·RED/GREEN은 [supervisor-verification.md](../../reports/SPEC-MOAI-GATEWAY-001/supervisor-verification.md)에 기록했다.
- Baseline-attribution: WT-unified-gateway, HEAD 81c1d58f9. 동시 translate worker의 RED는 별개이며 해당 파일을 수정하지 않았다.
- Gaps: CLI/실제 Claude exec·PTY·signal, Windows 런타임, 전체 M4 AC는 미완료다.
- Residual-risk: 제품 수명 기본값은 아직 정하지 않았다. overlay는 child가 만든 private 디렉터리의 독점 소유 계약이며 악의적인 같은 UID의 동시 변경에 대한 원자적 삭제 보장은 없다. 강제 OS 종료 시 child cleanup 미실행 가능성은 남는다.

### 2026-09-11 — CLI·환경·훅 연결 기반

- Claim: gpt 공통 진입·login/logout/status 명령, 실제 private supervisor 연결, 14키/GLM 슬롯·초기 provider env, env를 분리한 settings overlay 후보, normal/continue/fallback 처리 및 gateway 훅 보호를 구현했다. 세 gateway 명령은 옛 team_mode와 무관하게 Claude Code profile effort를 유지한다. production gateway 활성화는 아직 닫혀 있다.
- Evidence: 관련 CLI 시험 `ok .../internal/cli 3.703s coverage: 10.7%`; 신규 파일별 87.0~100.0%. race `ok .../internal/cli 3.717s`, hook `(cached)`; vet·gopls·Windows CLI test compile exit 0. 실제 helper에서 bound localhost 204, Stop 후 port 폐쇄·owned overlay 삭제를 검사했다. RED/GREEN·전체 명령은 [cli-integration-verification.md](../../reports/SPEC-MOAI-GATEWAY-001/cli-integration-verification.md)에 보존했다.
- Baseline-attribution: WT-unified-gateway, HEAD 81c1d58f9, SPEC 0.8.0. auth/translate worker 파일은 수정하지 않았다. status 전이나 commit은 하지 않았다.
- Gaps: M0 인증 운반 키·요청별 OAuth·provider factory 제품 연결, PICKER/source precedence/fallback, 실제 Claude/Codex 계정·GLM effort wire·Windows runtime·전체 AC는 미검증이다. session record의 실제 초기 provider와 factory launch event의 command backend는 현재 구분한다.
- Residual-risk: settings.env 메모리 이동은 실제 Claude 우선순위 동등성이 검증되지 않은 후보이며 inherited ANTHROPIC_REASONING_EFFORT 영향은 남은 정책 게이트다. AUTH Store의 Windows 지원은 현재 명시 거절 상태다. 전체 저장소 판정은 통합 브랜치 CI로 PENDING이다.

## §E.3 Run-phase Audit-Ready Signal

감사 준비 미완료. M0는 INCONCLUSIVE다. 0.8.0 인식기와 실제 캡처 판정은 통과했으나 M1 전체 및 제품 AC는 미완료다.
`run_complete_at`, `run_commit_sha`, AC PASS 수치와 완료 신호는 발행하지 않는다.
제품 코드 구현 완료·검증 완료를 주장하지 않으며, 저장소 전체 시험 판정은 통합 브랜치 CI의 소관으로 PENDING이다.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

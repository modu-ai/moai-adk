# SPEC-MOAI-GATEWAY-001 — 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- 0.13.0 (2026-09-14, t654) — AS-5 plan-phase 개정: 세 launcher 생산 통합·구독/API 이중 경로 검증·
  실제 Claude PTY 표면(도구검색·서브에이전트·재개·모델전환) 실증·경로별 context 판정·Windows GitHub CI
  실행 증거·rc 로컬 배포 게이트. `REQ-MG-027`/`AC-MG-026` 신설(요구사항 25 / 수용 기준 25 — Tier L
  상한 도달, `plan.md` §I), 기존 AC 하위 시나리오 AS-017~AS-022 신설, §E 표에 T22 경계 정합 행 추가.
  t653 잔여 위험 흡수 — compaction `appliedEpoch` 생산 판독(`REQ-MG-027` (b), `design.md` §11.2)과
  fork 자식 inherited prefix 원장 대조((c), §11.3). T21(t844 이관)·AS-013 NOT-RUN은 유지. status
  implemented → in-progress(다중 카드 시리즈 재개; implemented→completed는 후속 sync 몫).
  마일스톤 A5-M1~M7(통합 → 검증 → 배포 게이트), 증거는 `.moai/reports/t654/`의 `as5-` 접두사
  (옛 t654 가드 보고서와 충돌 방지). `design.md` §11·`research.md` §20 신설, §E.2~§E.4 미접촉.
  open question: 카드 문구 "rc.8"은 발행 시점 표기 — 배포 시 다음 미사용 rc 번호 적용
  (`.moai/docs/version-management.md` Local RC Numbering, 보고서에 명시).
  plan-audit: CONDITIONAL PASS 0.90(`.moai/reports/t654/as5-plan-audit.md`) — 수선 3건(D1 런처 대기
  오류 제2 위치 `launcher.go:142`를 A5-M1 예상 변경에 추가, D2 `AC-MG-026` (d) 강제 전제 집합 명명,
  D3 `design.md` §11.1 진입 함수 비공유 정정) 반영으로 루프 없이 종결.
- t653 (AS-4: 소유 thread resume·model 변경·fork·compaction) plan-phase 기록 — 2026-09-13.
  기준선: worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`, HEAD `74d872aaf` (선행 t652 = `530bd7330`).
  범위: 계정·family·agent·thread와 완료 public prefix 원장 고정, resume 및 idle 모델 변경 연결, compaction의 정상 요약 turn 처리
  (PostCompact digest 대조 뒤 public history 재설정), 명시 세션 fork의 lastTurnId 분기, native Agent(fork)/subtask 가용성
  probe 기록. 마일스톤 M1~M6와 AC 대응(AS-010~AS-013), 위험·제약은 카드 실행 계획
  [.moai/reports/t653/plan.md](../../reports/t653/plan.md) 참조. 전제 관측: `NewAppServerAdapter`·`codexbridge.New`의
  생산 호출 지점 부재, `codexbridge`의 `thread/resume` 경로와 compaction 처리 부재, FileStore "재개 불가" AS3 한계의
  AS4 계약 해제 필요 — 상세는 계획서 §1.3 gap 목록.

- 0.12.0 (2026-09-14, t653) — 운영자 결정(리드 전달)으로 AS-010·011·012 실세션 양성 실증을 카드 t844로 이관
  (`acceptance.md` §E 표 T21). AS-013은 설계된 NOT-RUN 유지. status 변경·run_complete_at 발행 없음.
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

### 2026-09-13 — t653 AS-4 M1 probe + M2 소유 thread resume

- Claim: (M1) 설치본 Claude Code 2.1.270의 호출 표면에서 native Agent(fork)/subtask 파라미터를 확인하지 못해 전제 실패 NOT-RUN으로 기록했다 — native fork 양성 의무는 유지되고 AS-013 전체 완료는 보류다. (M2) Engine에 소유 thread resume 경로를 추가했다: idle 배리어에서만 재개, thread 신원은 store 레코드에서만, model·CWD·완료 public prefix 고정 대조, legacy AS3 레코드는 명시 거절(마이그레이션 없음), 비-idle 배리어는 재개 불가 유지(AS-008 보존).
- Evidence (baseline HEAD `3c34e90ee` → 커밋 `e573ad720`, `4976e5a06`):
  - M1 RED 조건 없음(실증 probe). probe 방법·출력 전문: `.moai/reports/t653/m1-native-fork-probe.md`. 핵심 관측: `claude --help`의 fork 계열은 launcher 세션 플래그 `--fork-session`뿐, Agent 도구 파라미터는 `prompt`/`subagent_type`/`run_in_background`/`name`/`isolation("worktree")`/`model`/`effort`이고 `fork` 파라미터 0건(바이너리 문자열 실측).
  - M2 RED: `go test ./internal/codexbridge/ -run 'TestResume' -count=1` → `internal/codexbridge/resume_test.go:38:4: q.Resume undefined (type Request has no field or method Resume)` / `FAIL` (`.moai/reports/t653/m2-resume-red.log`).
  - M2 GREEN: `go test ./internal/codexbridge/ -run 'TestResume' -count=1 -v` → PASS 4 (`TestResumeAttachesOwnedThreadWithoutRebuild`, `TestResumeRejectsLegacyBarrierRecord`, `TestResumeRejectsIncompleteBarrier`, `TestResumeRejectsForeignBindings` 6 서브케이스 전부) (`.moai/reports/t653/m2-resume-green.log`). 패키지 전체 `-count=1`: 21 PASS + 실패 4건(`TestAudit*`, `TestCanceledStart*` — lifecycle_subprocess, 아래 Gaps).
  - race: `go test ./internal/codexbridge/ -race -count=1` → DATA RACE 경고 0건, 실패는 동일 4건뿐.
  - Windows: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
  - lint: `golangci-lint run ./internal/codexbridge/...` → 18건, 전부 기존 코드(줄번호 이동 포함)이고 added-line과의 기계 대조(`/tmp/t653.diff` 기준)에서 신규 이슈 0건. gofmt 정리 완료.
- Store-contract decision: legacy AS3 레코드는 **거절**(마이그레이션 없음). 이유 — 구 레코드는 resume에 필요한 고정 model/CWD 필드를 의도적으로 운반하지 않았으므로 마이그레이션은 결핍된 귀속 사실을 조작해 만드는 것이고, 묵시적 스키마 확장은 금지됐다. 거절은 `ErrRecovery` 래핑 명시 오류("barrier predates the resumable schema")로 한다. 레코드 스키마는 `Schema=2` + `Model`/`CWD` 추가이며 save가 스키마를 기록한다.
- Baseline-attribution: worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`, HEAD `3c34e90ee` 기준 측정, 커밋 `e573ad720`(M1)·`4976e5a06`(M2). 4건의 lifecycle_subprocess 실패는 이 diff 없는 기준 커밋에서 동일 재현해 사전 존재 환경 의존으로 귀속했다(stash SHA `33a6ed820`으로 임시 분리 측정).
- Gaps: AS-010의 "새 MoAI process 실제 세션 회상"은 실증 항목으로 미수행(환경상 실제 app-server 서브프로세스 기동 실패 — lifecycle_subprocess 4건과 동일 원인으로 판단, 단 원인 규명은 범위 밖). idle model 변경(M3), compaction(M4), 경계 fork(M5)는 미착수. native fork 양성 실증은 NOT-RUN(전제 실패).
- Residual-risk: thread/resume의 실제 App Server 응답 형태(`{thread:{id}}` 검증)는 fake 기준이며 실제 서버와의 일치는 실증에서 확인해야 한다. resume 뒤 첫 Step의 attach 실패는 failed 배리어로 귀결되는데, 실제 서버가 일시 오류를 냈을 때의 재시도 정책은 명시적 recovery 가이드 몫이다(자동 재시도 없음 — 의도된 계약).

### 2026-09-14 — t653 AS-4 M3 idle 모델 변경 + M4 compaction + M5 경계 fork + M6 증거

- Claim: (M3) idle 배리어에서만 모델 전환을 허용하고 전환값이 다음 turn/start로 운반되며 배리어에 재고정된다 — waiting 전환은 거절돼 잘못된 turn에 적용되지 않는다. gateway 측 거절군(bare `gpt-6` ErrUnknownModel, 타 provider의 bridge 진입 ErrManagedAuthority)을 테스트로 고정했다. (M4) compaction을 정상 요약 turn으로 처리한다: 반환 summary를 인증된 PostCompact의 exact digest·scope·epoch와 대조해 한 번만 rebase하고, `Manifest.Rebase`·`Store.Rebase`가 public history를 재설정한다 — 중복·stale·gap(유실)·재시작 재생·substring-only는 명시 거절이고 bridge는 compact RPC를 0회 발행하며 RPC id 없는 서버 주도 compaction 알림은 무시된다. PreCompact-after-HTTP 분류와 추가 compact RPC는 도입하지 않았다. (M5) exact completedTurnID 경계 fork를 착지했다: `Manifest.ChainTo`/`ForkAt`이 경계까지의 완료 체인만 복사해 부모가 이후 turn을 완료해도 경계 이후 사실이 자식에 유입되지 않고, 미지·영 경계·사이클·불일치 링크·미지 원본은 자식 상태 생성 전에 거절된다. `Manager.ForkAt`이 family 인덱스에 경계 fork를 노출하고 `Manager.Fork`는 lastTurnId(tip) 분기로 정렬됐다. codexbridge `Request.Fork`는 자식이 부모 thread를 입양하지 않고 inherited prefix로 새 thread를 시작하게 하며, 자식 배리어는 새 프로세스에서 독립 resume된다. (M6) 커버리지·Windows 빌드·lint·gofmt를 실측해 취합했다.
- Evidence (baseline HEAD `92db2cfe7` → 커밋 `14dba89c5`(M3), `e45f50a8d`(M4), `64885fa06`(M5)):
  - M3 RED: `go test ./internal/codexbridge/ -run 'TestIdleModelChange|TestWaitingAndNewPhase'` → `model_test.go:28: idle model change rejected` FAIL (`.moai/reports/t653/m3-model-red.log`).
  - M3 GREEN: codexbridge 패키지 전체 `ok` (`.moai/reports/t653/m3-model-green.log`); gateway 거절군 `ok` (`.moai/reports/t653/m3-model-gateway.log`).
  - M4 RED: `CompactBase`/`NewRebaseLedger` 등 미정의 컴파일 실패 (`.moai/reports/t653/m4-compact-red.log`).
  - M4 GREEN: receipt 패키지 `ok` (`.moai/reports/t653/m4-compact-green-receipt.log`); codexbridge compaction 분류 2건 `ok` (`.moai/reports/t653/m4-compact-codexbridge.log`).
  - M5 RED: `ForkAt`/`ChainTo`/`Request.Fork` 미정의 컴파일 실패 (`.moai/reports/t653/m5-fork-red.log`).
  - M5 GREEN: receipt·conversation·codexbridge 3패키지 `ok` (`.moai/reports/t653/m5-fork-green.log`).
  - M6: 커버리지 codexbridge 83.1% / receipt 88.9% / conversation 80.3% / gateway 91.7% (`.moai/reports/t653/m6-coverage.log`); `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `golangci-lint run --timeout=5m` → `0 issues.` exit 0 (`.moai/reports/t653/m6-lint.log`); `gofmt -l` 빈 출력. 요약: `.moai/reports/t653/m6-evidence-summary.md`.
- Store-contract decisions: (1) ChainTo는 Previous 링크가 manifest에 없는 접두사를 가리키는 것을 체인 뿌리로 취급한다 — 이 원장에서는 그것이 정상이며(core_test 자체가 그렇게 구성), 거절은 미지·영 경계, 사이클, 같은 접두사의 불일치 링크에 국한한다. (2) compaction은 engine 코드 변경 0건으로 분류된다 — 정상 turn 경로가 요약 turn을 운반하고 알림 무시는 기존 fail-open 루프가 수행한다(테스트로 고정). (3) `Manager.Fork`(tip 분기)도 ChainTo로 정렬해 계층 간 fork 의미를 일치시켰다.
- Baseline-attribution: worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`. 커밋 진행 `92db2cfe7` → `14dba89c5` → `e45f50a8d` → `64885fa06`. 사전 존재 환경 실패 5건(codexbridge lifecycle_subprocess 4건 + gateway `TestAppServerSubprocessHTTPToolContinuation`)은 작업 변경이 없는 커밋 트리(`git archive HEAD` 추출본)에서 동일 재현해 이 diff 이전 환경 의존으로 귀속했다.
- Gaps: AS-010 새 프로세스 실세션 회상, AS-011 실제 turn model 일치 양성, AS-012 실제 Claude 압축 수집, AS-013 `--fork-session` 실분기·병렬/중첩 자식 격리는 전부 실증 미수행. M1 native fork 전제 실패 NOT-RUN 유지 — AS-013 전체 지원 완료는 계속 보류. 전체 스위트 판정은 CI 소관.
- Residual-risk: App Server 실제 응답 형태는 fake 기준. RebaseLedger의 재시작 복원은 생성자 `appliedEpoch` 주입에 의존하며 생산 배선(t654)이 그 값을 읽는 위치는 미설계. Fork 자식의 inherited prefix는 engine에서 caller-asserted — 원장 대조는 gateway 계층 책임으로 남는다. idle 전환의 배리어 model 재고정은 turn 성공 시점에 일어난다.

## §E.3 Run-phase Audit-Ready Signal

run-phase 완료. `run_complete_at: 2026-09-14`, `run_commit_sha: 9f3dc41e0`.
t653 run-phase의 자동화 가능 부분(M2 resume, M3 idle 모델, M4 compaction, M5 경계 fork)은 커밋 `14dba89c5`·`e45f50a8d`·`64885fa06`로 착지하고 패키지 테스트·커버리지(codexbridge 83.1% / receipt 88.9% / conversation 80.3% / gateway 91.7%)·Windows 빌드·lint 0·gofmt가 이번 실행에서 실측됐다.
실증 gap의 처분 근거: acceptance.md:469 **T21** — AS-010(실세션 회상)·AS-011(실제 turn model 일치)·AS-012(실제 Claude 압축 수집)의 실세션 양성 실증은 카드 **t844**(라이브 계측 세션)로 이관됐고, AS-013은 이관 대상이 아니며 설계된 전제 실패 NOT-RUN(M1 probe, `.moai/reports/t653/m1-native-fork-probe.md`)을 유지한다. t844의 실세션 양성이 도래하기 전까지 전체 기능 통과는 보류다.
저장소 전체 시험 판정은 통합 브랜치 CI의 소관이다. 상세 판정 초안: [run-verdict.md](../../reports/t653/run-verdict.md).

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready (card t653 sync — the umbrella SPEC does NOT close per card; t654 and later cards remain)
sync_complete_at: 2026-09-14
sync_commit_sha: "pending-backfill"
frontmatter_status_transitions:
  - transition: "in-progress → implemented (NOT completed — SPEC-MOAI-GATEWAY-001 is a multi-card Tier L series; the implemented → completed transition rides a later sync after t654+ cards land)"
    owner: manager-docs
    surfaces: spec.md only (plan.md / acceptance.md are frontmatter-stateless; progress.md carries no status axis)
card_verdict: .moai/reports/t653/verdict.md (5-section format; absorbs the run-verdict.md draft)
sync_audit: .moai/reports/t653/sync-audit.md (4-dimension score — Functionality/Security/Craft/Consistency)
sync_remeasurement: 변경 패키지 재측정(이 실행) — receipt 88.9% / conversation 80.3% / gateway 91.7%(-skip 사전존재 1건) ok; codexbridge는 skip 없이 사전 존재 lifecycle_subprocess 4건만 실패(신규 0), 4건 skip 시 82.5%(atomic 82.7%). run-phase 기록 83.1%와의 잔차는 skip 표현식 미기록 attribution gap으로 verdict Gaps에 기록.
changelog_entry: deferred (B12 pre-emission grep `grep -c 'SPEC-MOAI-GATEWAY-001' CHANGELOG.md` → 0; series convention documents at release/launcher-wiring time — card t653's surface is internal packages only, user-visible effect first exposed by t654 launcher wiring; decision recorded in .moai/reports/t653/verdict.md § CHANGELOG 결정)
b12_self_test_a: pass (pre-emission grep count 0 — no duplicate-emission risk; no halt condition)
b12_self_test_b: n/a (no CHANGELOG entry emitted — AC-count match test has no target)
b12_self_test_c: n/a (no CHANGELOG entry emitted — no file paths claimed in an entry)
mx_tag_validation: pass (@MX:WARN/REASON annotations observed in 11 files across internal/codexbridge·internal/gateway — sync sub-step scan 2026-09-14)
t21_transfer_consistency: acceptance.md:469 ↔ spec.md HISTORY 0.12.0 ↔ progress.md §E.3 — three surfaces carry the same t844 transfer content (verified by sync-audit)
as013_disposition: designed NOT-RUN maintained (M1 pre-condition failure probe, .moai/reports/t653/m1-native-fork-probe.md) — not a transfer target; AS4 overall-support completion remains deferred

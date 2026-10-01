# progress.md — SPEC-SESSION-ANCHOR-ATTR-001

## 카드 연계

- **카드**: t1339 — "Bash 워크트리 세션 앵커 교차 레인 오판독 수리" (등록 2026-09-29T07:40:40Z)
- **배차**: 리드 디스패치 — t1337+t1339 동일 근원 통합 수리, 저장소 내 수리 가능분 + 재현/측정 스위치 + 회피 규율 문서(부분 산출 허가) + 레인-병렬 재현 환경 심각도 판정
- **SPEC**: SPEC-SESSION-ANCHOR-ATTR-001 (Tier M)
- **병목 연계**: autonomy-bottleneck-proposal-20260929.md bottleneck #7 → P7 (추가 자율화의 전제로 시퀀싱)

## Phase 1 SKIP 근거

리드가 계획 단계 진입 전 조사 팬아웃(plan-research-fanout 3렌즈 + 종합)을 이 레인에서 선실행하여 산출물을 /tmp에 전달했다. 본 SPEC의 research.md가 그 종합본을 축어 보존하므로, 표준 Phase 1(계획 내 재조사)은 SKIP — 중복 조사는 컨텍스트 낭비이며 1차 기록은 이미 수집·교차검증됨. 렌즈 원본 3종의 경로는 research.md 헤더 인용.

## Decision Point 1 흡수

계획-검토 후보 게이트(Decision Point 1)는 lane 도크트린에 따라 factory Implementation Kickoff 게이트에 흡수된다 — 운영자가 레인 창(pane)에서 직접 답한다. 별도의 중간 승인 라운드는 두지 않는다.

---

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-30
- artifacts: spec.md, plan.md, acceptance.md, research.md, spec-compact.md, progress.md (Tier M + research + compact)
- plan_audit: PASS 1.00 (반복 2/2, Tier M 임계 0.80) — 1차 FAIL 0.71(D1~D4) → 수정 → 델타 재감사
- plan_audit_reports: .moai/reports/plan-audit/SPEC-SESSION-ANCHOR-ATTR-001-review-1.md, -review-2.md (gitignored 로컬 런타임 기록)
- plan_audit_model: glm-5.3-flash (GLM 프록시 레인 실사용 모델, 1차는 audit_multi로 codex gpt-6.1-sol/high 수렴 병행)
- FO-PLAN-2 렌즈 생략 근거: 당일 실측 429 레이트리밋 압박(워크플로 중 1에이전트 429·t1347 쿼터 기록) + plan-auditor --deep 자체 증거 수집으로 대체 — fallback 절(single plan-auditor path) 적용
- DP2/3/3.5 흡수: 개발환경(워크트리 이미 진입)·다음 행위(run, 배차 지시)·실행 모드(serial, 레인 기본)는 factory Kickoff 게이트에서 일괄 노출
- MX 계획: plan.md는 감사 PASS 해시 고정 유지 — MX 대상(신규 exported 함수 @MX:NOTE, RelocateSession @MX:ANCHOR 후보, 미테스트 public @MX:TODO)은 run 위임 프롬프트로 전달
- run 진입 시 Phase 1 skip 예상: PASS 1.00 ≥ 0.80 + 아티팩트 해시 불변 (skip 계약 3조건 중 2개, verdict+score+hash)

## §E.2 Run-phase Evidence

### M1 — W1 거부 귀속 가능화 (2026-09-30, HEAD 78df22755 기준 작업)

- REQ-SAA-001/002 구현: `harness.Event`에 `cwd`·`worktree_path` additive omitempty 필드 추가(internal/harness/types.go — 행 스키마의 기계적 소재라 M1 봉투 내 캐스케이드), failure_observer.go에 WorktreeGuardRefusal 행 전용 attribution(session_id/cwd/트리 경로, 미해상 시 `unknown` 마커, 행 유실 없음). 다른 카테고리 행은 기존 형태 유지(REQ 스코프).
- RED 증거(E8): 신규 3테스트 컴파일 실패 — `undefined: guardRefusalWorktreePath` / `undefined: guardRefusalAttribution` / `harness.Event has no field or method Cwd` (구현 전 실측).
- GREEN: `go test ./internal/hook/ -run '^(TestGuardRefusalWorktreePath|TestClassifyError_GuardRefusal_UnknownMarkers|TestRecordToolFailureEvent_GuardRefusalRowFields|…기존 guard군)$'` — 전부 PASS, ok.
- 기존 살생 변이 2건 테스트(TestClassifyError_GuardRefusal_AnchorOnly, _BeatsOOM) 무수정 GREEN — REQ-SAA-010.
- draft→in-progress 전환(spec.md status+updated만) 본 M1 커밋에 동승.

### M2 — W2 재배치 감사 + 소유 타당성 (2026-09-30)

- REQ-SAA-003: `Registry.RelocateSession`이 cwd 재작성 시 감사 행(session_id/from/to/trigger/timestamp)을 `<project-root>/.moai/logs/anchor-relocation-audit.jsonl`에 추가 — 내부 구현은 신설 `RelocateSessionWithOptions`(anchor_relocate_audit.go)로 위임, 모든 진입점이 감사. fail-open(기록 실패가 재배치를 막지 않음).
- REQ-SAA-004: 케이스표 4팔 구현(self/other-live/unreadable/registry-only + 잔여 dead-holder 기록). lock reason card-id/pid 토큰 파싱은 `parseLockPID`/신설 `ParseLockCardID` 재사용. other-live·registry-only는 advisory 플래그 + 진행; unreadable은 기존 fail-closed 준수(플래그 없음).
- REQ-SAA-005: `workflow.anchor_relocation_guard.enabled` opt-in 키(config types.go+defaults.go — BranchGuard/AgentStopGuard 선례 동형, 기본 false, 템플릿 중립). on + 플래그 → 거부 + 감사행 Refused:true, `ErrRelocationRefused`.
- REQ-SAA-006 보존: 두-패스 후보 순서(relocateRegistryCandidatesFrom)·fail-open 무수정 — 기존 8 RUN 테스트 GREEN.
- 훅 연결: cwd_changed_relocate.go에 git-context seam(`relocationGitContext` 패키지 변수 — sessionProcessLiveness 선례 동형), `relocationTargetContext`(git rev-parse --show-toplevel + worktree list --porcelain fail-open), `anchorRelocationGuardEnabled`(nil-safe). cwdChangedHandler cfg 필드 + NewCwdChangedHandlerWithConfig, deps.go 갱신, handleWorktreeMove에 cfg 전달.
- RED 증거(E8): (a) session측 — `undefined: guardRefusalAttribution` 아니라 `RelocationAudit` 관련 compile fail (audit 타입/메서드 미존재 시절); (b) hook측 — `undefined: relocationGitContext` ×3, `too many arguments in call to relocateSessionCwd` ×3, `undefined: anchorRelocationGuardEnabled` ×3 (구현 전 실측).
- GREEN: session 신규 5테스트 + 기존 RelocateSession 3테스트, hook 신규 4테스트 + 기존 two-pass 8 RUN — 전부 PASS. session·config 패키지 전체 ok.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop 소관>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_

## §F Phase 4 Mode Selection

- 입력 파라미터: tier M · scope 7~9파일 (internal/hook 3 + internal/session 2 + docs 1 + 테스트) · 도메인 수 2 (hook/session + rules doc) · 언어 혼합 Go 위주 · concurrency benefit LOW (coding-heavy)
- 모드 평가: direct 미선정(다중 파일·의미 변경) / fanout 미선정(coding-heavy — Anthropic 병렬화 주의사항) / sweep 미선정(기계적 대량 변형 아님) / **serial 선정**
- Decision: serial (단일 manager-develop 순차 위탁, per-milestone 커밋)
- 근거: Anthropic coding-task parallelism caveat — 코딩 과업은 연구와 달리 진병렬화가 드묾. 앵커 수리는 hook/session 패키지에 걸치는 의미 변경이라 단일 작성자 순차가 안전. 리서치는 이미 plan 단계에서 fanout으로 소진.
- Kickoff: 운영자 직답 승인(진입 승인·자율 진행) 2026-09-30 — 이 기록이 그 게이트의 결정 레코드
- Phase 1 skip 근거(위임 Section A에도 기재): verdict PASS 1.00(≥0.80) + 아티팩트 해시 불변(판정 후 spec/plan/acceptance/research 무변경 — progress.md는 해시 대상 아님)

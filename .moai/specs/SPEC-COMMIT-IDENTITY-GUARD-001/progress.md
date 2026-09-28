# SPEC-COMMIT-IDENTITY-GUARD-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- card t1289, Tier M. plan-phase 산출물 4 개(`spec.md`, `plan.md`, `acceptance.md`, `progress.md`),
  워크트리 `.claude/worktrees/t1289`, 브랜치 `WT-commit-identity-guard`, base `37dc766b9`.
- SPEC ID 사전 점검: Bash 정규식 검사 출력 `PASS`.
- 입력 근거: `.moai/reports/t1289/root-cause.md`. plan 페이즈의 새 측정은 픽스처 이메일 열거
  (51 개, `plan.md` §B.1)와 센티널 스윕(14 종)뿐이다.
- 요구사항 REQ-CIG-001..012, 인수조건 AC-CIG-001..014.
- `plan.md` §B.3 저장소 범위 판단: 레인 결정 (b) 로 해소(2026-09-28, spec 0.1.1). 미해소 `[NEEDS CLARIFICATION]` 0 건.
- status: `draft`.
- plan_status: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## 진행 기록

### 2026-09-28 — plan-audit iter2 D14/D15 수리 (감사 한도 2/2 소진, 리드 채널로 검증)

- **D14(blocking, MP-2 GEARS)** — REQ-CIG-001·008 이 상시 의무와 조건부 의무를 한 항목에 묶은 것을 **분할, 압축 금지** 방향으로 고쳤다. REQ-CIG-001 → 분류·비트리거 통과(Ubiquitous, REQ-CIG-001) + 트리거 동사 탐침(Event-driven, REQ-CIG-002)으로 분리. REQ-CIG-008 → 배선 순서·선행 거부 보존(Ubiquitous, REQ-CIG-009) + 분류 불가 PowerShell 간접 구문 통과·감사 1줄(Event-driven, REQ-CIG-010)으로 분리. 나머지는 001..012 연속 번호로 재배치했고(구 002→003 … 구 009→011, 구 010→012), AC 수는 14 개 그대로다 — 분할 REQ 모두 기존 AC 가 덮어서 신규 AC 는 불필요했고, AC-CIG-001..014 의 Covers 만 새 번호로 재귀속했다.
- **구 REQ-CIG-010 분할 여부**: 세미콜론 뒤 별도 의무가 없다 — "While <범위 불일치> … shall allow … and without an identity probe" 는 단일 State-driven 응답 안의 한 동작이라 판정했고, 분할 없이 REQ-CIG-012 로 번호만 옮겼다.
- **D15(optional)** — AC-CIG-012(c) 의 Given 은 선행된 수리 커밋(`5e8599426`)에서 이미 구체화돼 있었다(autonomy `contract` 모드 + `checkClosurePush` 가 `second_review_not_performed` 로 거부하는 push-readiness fixture). 근거 재확인: 모드 게이트는 `internal/hook/closure_push.go` L83(`s.Mode != "contract"` → 빈 결정), 거부 코드 `second_review_not_performed` 는 `internal/closure/readiness.go:22`. 이번 커밋에서 acceptance.md 는 Covers 재귀속만 했다.
- **교차 참조 갱신**: spec.md §6 제약(구 REQ-CIG-006→REQ-CIG-007), plan.md §B.1(007→008)·§B.3(010→012)·§F M2(006→007)·§H(001..010→001..012), progress.md §E.1 범위 표기.
- **검증 한계**: 정식 plan-auditor 재감사는 하지 않는다 — 감사 한도 2/2 소진. 레인이 `moai spec lint` + REQ↔AC 매핑 재검증 + 연속 번호 검사를 이 트리에서 실행했고, MP-2 해소 판정은 리드 채널이 읽는다.

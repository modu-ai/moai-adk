# Progress — SPEC-SERVED-MODEL-AUDIT-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-27
- card: t1282 (class C)
- tier: M — 산출물 spec.md, plan.md, acceptance.md, progress.md (편차 없음)
- REQ: 16건 (REQ-SMA-001..016) · AC: 16건 (AC-SMA-001..016) — Tier M 예산 안 (v0.2.0)
- plan-audit iter1: FAIL 0.75 (`.moai/reports/plan-audit/SPEC-SERVED-MODEL-AUDIT-001-review-1.md`). v0.2.0 반영:
  - D3 Tier 예산 → REQ 19→16, AC 25→16 통합(형제 부정 사례를 표 기반 AC 로). Tier L 승격 대신 통합을 택함 — 판단이 들어가는 독립 편집 단위는 4개 패키지이고 나머지 파일은 미러·기계 방출물(spec.md §D)
  - D2 공유 스코프 비확장 → 전용 술어(plan.md D3), REQ-SMA-013, AC-SMA-009
  - D4 해석 모델 출처 → 같은 `deps.Config` 주입(plan.md D4), REQ-SMA-003/004, AC-SMA-003
  - D1/D11 워크트리 slug → 열거 규칙 REQ-SMA-014, plan.md D5, AC-SMA-011, spec.md §A.2 측정 원문(primary 203 / 워크트리 slug 200)
  - D5 자기 보고 표면 → REQ-SMA-015(보고서 파일 + 최종 메시지), REQ-SMA-016(`last_assistant_message`), AC-SMA-014
  - D6 센티널·병합 사유 → REQ-SMA-010, AC-SMA-007
  - D7 AC 얕음 → AC-SMA-013 의무 문장 고정 grep + C2 중립성 부정 grep
  - D8 커버리지 → AC-SMA-004 `session_id`, AC-SMA-005 unknown 경고 행, AC-SMA-011 종료 코드
  - D9 빈 스캔 → REQ-SMA-014 info(ok 아님), AC-SMA-011 (b)
  - D10 레거시 kind → REQ-SMA-011, plan.md D2, AC-SMA-008; M7(규칙 문서 편집)은 sync 단계로 이관(spec.md §E)
- 근거: `.moai/reports/t1282/verdict.md`(착수 판정서) + spec.md §A.2 plan 단계 실측(M1-M4)
- 측정 트리: develop `b59a5d69c` 기준 워크트리 `WT-served-model-audit`
- 미확인 항목(spec.md §A.4): SubagentStop 실제 페이로드의 `agent_transcript_path` 채움 여부, 발화 시점 마지막 assistant 행 기록 여부, t1237·t1239·t1099 개별 대조, L2 워크트리 slug
- open_clarifications: 0 (리드 결정 (c) 로 설계 확정)
- run-phase 필수 단계: `make agents-emit` 후 `make agents-emit-check`, doctor 점검 이름의 `namesAddedAfterBaseline` 등록

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

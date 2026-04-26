---
spec_id: SPEC-V3R3-DEF-001
title: Task Decomposition — ORC Dependency Cycle Resolution
version: "1.0.0"
status: draft
created: 2026-04-25
related_plan: .moai/specs/SPEC-V3R3-DEF-001/plan.md
related_spec: .moai/specs/SPEC-V3R3-DEF-001/spec.md
---

# 작업 분해 — SPEC-V3R3-DEF-001

> **범례**:
> - **File owner**: 단독 소유 파일 경로
> - **Depends on**: 선행 task ID
> - **Wave**: A.1 / A.2 / A.3 / A.4
> - **Parallel OK**: 동일 Wave 내 병렬 가능 여부

---

## 전체 Task 개요

| Wave | Task 수 | Parallel 가능 | Sequential 필수 |
|------|---------|---------------|-----------------|
| A.1 ORC HISTORY 추가 | 5 | T-A1-1 ~ 5 (모두 다른 파일) parallel | — |
| A.2 MIG-001 update | 1 | — | T-A2-1 |
| A.3 WF-001 reaffirm | 1 | — | T-A3-1 |
| A.4 Verification | 2 | A4-1 단독, A4-2 sequential | T-A4-1 ~ 2 |

**총 task 수: 9**

---

## Wave A.1 — ORC HISTORY 추가 (5 tasks, parallel OK)

### T-A1-1: SPEC-V3R2-ORC-001 HISTORY 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-ORC-001/spec.md`
- **Depends on**: 없음
- **Parallel OK**: Yes
- **Action**: HISTORY 섹션 마지막에 row 추가:
  ```
  | 1.X.0 | 2026-04-25 | manager-spec | SPEC-V3R3-DEF-001 적용 — DAG invariant 명시 (D-CRIT-001 baseline). dependencies 값 무변경. |
  ```
- **Verification**: `grep "SPEC-V3R3-DEF-001" .moai/specs/SPEC-V3R2-ORC-001/spec.md` 매치
- **Rollback**: 해당 row 삭제

### T-A1-2: SPEC-V3R2-ORC-002 HISTORY 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-ORC-002/spec.md`
- **Parallel OK**: Yes
- 이하 동일 패턴

### T-A1-3: SPEC-V3R2-ORC-003 HISTORY 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-ORC-003/spec.md`
- **Parallel OK**: Yes

### T-A1-4: SPEC-V3R2-ORC-004 HISTORY 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-ORC-004/spec.md`
- **Parallel OK**: Yes

### T-A1-5: SPEC-V3R2-ORC-005 HISTORY 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-ORC-005/spec.md`
- **Parallel OK**: Yes

### Wave A.1 Checkpoint
- 5개 ORC spec.md 모두 HISTORY +1 row
- AC-DEF001-06 (ORC 부분) verification

---

## Wave A.2 — MIG-001 Update (1 task)

### T-A2-1: MIG-001 dependencies + HISTORY + version bump
- **File owner**: `.moai/specs/SPEC-V3R2-MIG-001/spec.md`
- **Depends on**: 없음
- **Parallel OK**: — (단일 task)
- **Action**:
  1. 현재 MIG-001 dependencies 읽기 (EXT-004, MIG-002, MIG-003)
  2. `- SPEC-V3R2-WF-001` 추가 (정렬 위치: EXT-004 이후, MIG-002 이전 또는 list 끝)
  3. version 필드 +0.1.0 bump (e.g., 1.0.0 → 1.1.0)
  4. HISTORY +1 row:
     ```
     | 1.1.0 | 2026-04-25 | manager-spec | SPEC-V3R3-DEF-001 적용 — WF-001 → MIG-001 단방향 contract 명시. dependencies 추가. |
     ```
- **Verification**:
  - `grep "SPEC-V3R2-WF-001" .moai/specs/SPEC-V3R2-MIG-001/spec.md` 매치
  - HISTORY row 1 추가 확인
  - 본문 무수정 확인 (`git diff` inspect)
- **Rollback**: dependencies 라인 + HISTORY row + version 모두 revert

### Wave A.2 Checkpoint
- AC-DEF001-02 verification

---

## Wave A.3 — WF-001 Reaffirm (1 task)

### T-A3-1: SPEC-V3R2-WF-001 HISTORY reaffirm 항목 추가
- **File owner**: `.moai/specs/SPEC-V3R2-WF-001/spec.md`
- **Depends on**: 없음
- **Parallel OK**: — (단일 task; A.2와 다른 파일이므로 사실상 A.2와 병렬 가능하지만 audit 명확성 위해 sequential)
- **Action**: HISTORY 섹션 마지막에 row 추가:
  ```
  | 1.2.0 | 2026-04-25 | manager-spec | SPEC-V3R3-DEF-001 적용 — MIG-001과의 단방향 contract reaffirm (WF-001은 MIG-001 비참조). |
  ```
  - version bump (1.1.0 → 1.2.0) 선택사항. dependencies 무변경이므로 HISTORY note만 추가하고 version 유지 권장.
- **Verification**:
  - `grep "SPEC-V3R3-DEF-001" .moai/specs/SPEC-V3R2-WF-001/spec.md` 매치
  - WF-001 dependencies가 MIG-001 미포함 확인
- **Rollback**: HISTORY row 삭제

### Wave A.3 Checkpoint
- AC-DEF001-03, AC-DEF001-06 (WF/MIG 부분) verification

---

## Wave A.4 — Verification (2 tasks)

### T-A4-1: ORC DAG invariant + MIG/WF contract 정적 검증
- **Depends on**: T-A1-* 모두, T-A2-1, T-A3-1 완료
- **Parallel OK**: Yes (단독 실행)
- **Action**:
  - acceptance.md AC-DEF001-01 ~ 04 verification 스크립트 실행
  - cycle 검사 Python script 실행
  - git diff inspect (본문 무수정 확인)
- **Verification**: 모든 AC 통과, 0 violations

### T-A4-2: plan-auditor 호출 및 D-CRIT-001 RESOLVED 확인
- **Depends on**: T-A4-1 완료
- **Parallel OK**: No
- **Action**:
  ```
  Agent(subagent_type: "plan-auditor",
        prompt: "Validate the V3R2 SPEC dependency graph. Specifically check:
                 1. ORC-001 ~ ORC-005 form a DAG (no higher-number references lower)
                 2. WF-001 → MIG-001 unidirectional (MIG-001 deps include WF-001, WF-001 deps exclude MIG-001)
                 3. Report D-CRIT-001 status (RESOLVED or open).
                 Reference: SPEC-V3R3-DEF-001 acceptance.md AC-DEF001-05.")
  ```
- **Verification**:
  - 보고서에 "D-CRIT-001: RESOLVED" 또는 equivalent verdict 포함
  - AC-DEF001-05 통과

### Wave A.4 Checkpoint
- AC-DEF001-01 ~ 06 모두 통과
- DoD 모두 충족

---

## Edge Case Tasks (조건부)

### T-EC-1: ORC dependencies invariant pre-existing violation
- **Trigger**: T-A1-* 실행 전 grep으로 위반 발견
- **Action**: STOP, operator 보고
- **Verification**: 사용자 명시 승인 후 진행

### T-EC-2: MIG-001 이미 WF-001 보유
- **Trigger**: T-A2-1 실행 전 grep
- **Action**: dependencies 추가 skip, HISTORY 항목만 추가
- **Verification**: 중복 없음 (`grep -c` = 1)

### T-EC-3: WF-001이 MIG-001 참조 발견
- **Trigger**: T-A3-1 실행 전 grep
- **Action**: STOP, cycle alert, operator 명시 승인 필요
- **Verification**: 결정 사항 진행 후 진행

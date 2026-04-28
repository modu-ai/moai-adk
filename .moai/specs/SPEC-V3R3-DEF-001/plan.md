---
spec_id: SPEC-V3R3-DEF-001
title: Implementation Plan — ORC Dependency Cycle Resolution
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-DEF-001/spec.md
---

# Plan — SPEC-V3R3-DEF-001

## 1. Objectives

- D-CRIT-001 (ORC Dependency Graph Cycle Risk) 해소
- ORC-001 ~ ORC-005 dependencies 단방향 DAG 명시적 검증 및 재정렬
- WF-001 → MIG-001 단방향 contract 확립
- plan-auditor GREEN verdict baseline 확보

## 2. Technical Approach

### 2.1 ORC SPEC dependencies 검증 및 정렬

각 ORC-001 ~ ORC-005의 frontmatter `dependencies` 필드 검사:

| SPEC | 현재 dependencies | 기대값 (DAG) |
|------|------------------|--------------|
| ORC-001 | CON-001 | CON-001 (변경 없음) |
| ORC-002 | CON-001, ORC-001 | CON-001, ORC-001 (변경 없음) |
| ORC-003 | CON-001, ORC-001, ORC-002 | 변경 없음 |
| ORC-004 | CON-001, ORC-001, ORC-002 | 변경 없음 |
| ORC-005 | CON-001, ORC-001, ORC-004 | 변경 없음 |

→ 현재 상태가 이미 DAG. 본 SPEC의 ORC 작업은 **정적 검증 + invariant 명시 + HISTORY 항목 추가**가 핵심 (실제 dependencies 값 변경 없음).

### 2.2 MIG-001 dependencies 추가

현재 MIG-001 dependencies: `EXT-004, MIG-002, MIG-003`
→ 추가: `SPEC-V3R2-WF-001`
→ 결과: `EXT-004, MIG-002, MIG-003, WF-001` (정렬: EXT < MIG < WF)

### 2.3 HISTORY 항목 추가

7개 affected SPEC 각각에 HISTORY row 추가:

```markdown
| 1.X.0   | 2026-04-25 | manager-spec | SPEC-V3R3-DEF-001 적용 — dependency graph DAG invariant 명시 (D-CRIT-001 해소) |
```

ORC-001 ~ ORC-005는 dependencies 값 변경 없으므로 minor version bump 불필요 (HISTORY note만 추가).
MIG-001은 dependencies 추가로 minor bump (e.g., 0.1.0 → 0.2.0 또는 1.0.0 → 1.1.0).
WF-001은 dependency 부재 reaffirm — HISTORY note만 추가.

## 3. Wave / Phase 설계

### Wave A.1 — ORC SPEC HISTORY 추가 (5 tasks, parallel)

ORC-001 ~ ORC-005 각각 HISTORY 항목 1개 추가. 본문/dependencies 무변경.

### Wave A.2 — MIG-001 dependency 추가 (1 task)

MIG-001 frontmatter dependencies에 WF-001 추가 + HISTORY 항목 + version bump.

### Wave A.3 — WF-001 reaffirm (1 task)

WF-001 HISTORY에 "MIG-001 단방향 contract reaffirmed" 항목 추가.

### Wave A.4 — Verification (2 tasks)

- plan-auditor 실행
- D-CRIT-001 verdict 확인

## 4. File 영향 요약

| File | Change Type |
|------|-------------|
| `.moai/specs/SPEC-V3R2-ORC-001/spec.md` | HISTORY +1 row |
| `.moai/specs/SPEC-V3R2-ORC-002/spec.md` | HISTORY +1 row |
| `.moai/specs/SPEC-V3R2-ORC-003/spec.md` | HISTORY +1 row |
| `.moai/specs/SPEC-V3R2-ORC-004/spec.md` | HISTORY +1 row |
| `.moai/specs/SPEC-V3R2-ORC-005/spec.md` | HISTORY +1 row |
| `.moai/specs/SPEC-V3R2-MIG-001/spec.md` | dependencies +WF-001, HISTORY +1, version bump |
| `.moai/specs/SPEC-V3R2-WF-001/spec.md` | HISTORY +1 row |

총 7개 spec.md 수정. Template/local 동기화 불필요 (.moai/specs/는 local-only).

## 5. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| MIG-001 dependency 추가가 implementation 순서 영향 | Low | MIG-001 미구현 상태, 순서만 고정 |
| ORC dependencies 의도치 않은 변경 | Medium | dependencies 필드 read-then-rewrite, diff 검증 |
| plan-auditor 실행 실패 | Medium | Agent invocation 시 retry (max 3) |
| HISTORY row format 불일치 | Low | 기존 HISTORY 패턴 모방 (V3R2-WF-001 spec.md line 28-31 참조) |

## 6. Open Questions

- OQ1: MIG-001 version bump 수치는? → **Decision**: 현재 version 확인 후 +0.1.0 (예: 1.0.0 → 1.1.0)
- OQ2: ORC SPEC들은 version bump 필요? → **Decision**: 불필요. dependencies 값 무변경, HISTORY note만.
- OQ3: plan-auditor가 invariant 자동 검증 가능한가? → **Decision**: 본 SPEC 시점에는 manual grep + agent. 자동화는 별도 SPEC.

## 7. Milestones

- M1: ORC-001 ~ ORC-005 HISTORY 항목 추가 완료
- M2: MIG-001 dependencies + HISTORY + version bump 완료
- M3: WF-001 HISTORY 항목 추가 완료
- M4: plan-auditor RESOLVED verdict 획득

## 8. Definition of Done

- [ ] 5개 ORC SPEC HISTORY +1 row 추가 (값 변경 없이 invariant 명시)
- [ ] MIG-001 dependencies에 WF-001 추가
- [ ] MIG-001 HISTORY +1 row + version bump
- [ ] WF-001 HISTORY +1 row (reaffirm)
- [ ] plan-auditor 실행 후 D-CRIT-001 RESOLVED
- [ ] AC-DEF001-01 ~ 06 모두 만족

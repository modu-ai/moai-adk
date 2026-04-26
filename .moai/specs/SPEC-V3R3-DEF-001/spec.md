---
id: SPEC-V3R3-DEF-001
title: ORC Dependency Cycle Resolution — D-CRIT-001 해소 + WF/MIG 단방향 정리
version: "1.0.0"
status: implemented
created: 2026-04-25
updated: 2026-04-26
author: manager-spec
priority: P0 Critical
phase: "v3.0.0 R3 — Phase A — Dependency Graph Sanitization"
module: ".moai/specs/SPEC-V3R2-ORC-*/, .moai/specs/SPEC-V3R2-WF-001/, .moai/specs/SPEC-V3R2-MIG-001/"
dependencies:
  - SPEC-V3R3-DEF-007
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "dependency-graph, cycle-break, orc, mig, wf, plan-audit, v3r3, phase-a"
related_theme: "Phase A — Defect Cleanup"
released_in: v2.15.0
---

# SPEC-V3R3-DEF-001: ORC Dependency Cycle Resolution

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 1.0.0   | 2026-04-25 | manager-spec | Initial draft. Phase A P0 — D-CRIT-001 (ORC graph cycle risk) 해소 + WF-001/MIG-001 단방향 reaffirm. |

---

## 1. Goal (목적)

V3R2 Plan-Audit에서 식별된 **D-CRIT-001 (ORC Dependency Graph Cycle Risk)** 결함을 해소한다. ORC-001 ~ ORC-005 5개 SPEC의 `dependencies` 필드를 명시적으로 검증·재정렬하여 단방향(DAG) 구조를 보장하고, WF-001 ↔ MIG-001 사이의 잠재적 cycle 위험을 단방향(WF-001 → MIG-001) 계약으로 고정한다. 이 작업은 plan-auditor가 v3.0.0 release 시점에 GREEN verdict를 발행하기 위한 필수 baseline이다.

### 1.1 배경

- 현재(2026-04-25) ORC SPEC들의 dependencies는 단방향처럼 보이지만 명시적 audit 없이 누적된 상태:
  - ORC-001 → CON-001
  - ORC-002 → CON-001, ORC-001
  - ORC-003 → CON-001, ORC-001, ORC-002
  - ORC-004 → CON-001, ORC-001, ORC-002
  - ORC-005 → CON-001, ORC-001, ORC-004
- D-CRIT-001 risk: ORC-002가 ORC-003/004/005를 역참조할 가능성 (예: 후속 revision에서 누군가 추가하면 cycle 발생). plan-auditor가 이를 정적 검증할 수 있는 계약 부재.
- WF-001 ↔ MIG-001: 현재 MIG-001 deps에 WF-001 미포함이지만, WF-001 변경이 MIG-001 영향 → 명시적 단방향 (WF-001 → MIG-001) 계약 필요.

### 1.2 비목표 (Non-Goals)

- ORC-001 ~ ORC-005의 본문 (Goal, Scope, Requirements) 수정 금지
- WF-001 / MIG-001 본문 수정 금지
- 다른 SPEC의 dependencies 정리 금지 (본 SPEC은 ORC + WF/MIG 한정)
- plan-auditor 자체 코드/로직 수정 금지

---

## 2. Scope (범위)

### 2.1 In Scope

- ORC-001 ~ ORC-005 5개 SPEC의 frontmatter `dependencies` 필드 검증 및 정렬:
  - 모두 SPEC-V3R2-CON-001 + 상위 ORC만 참조
  - 역방향(낮은 번호가 높은 번호 참조) 금지 invariant 명시
- WF-001 / MIG-001 단방향 계약:
  - MIG-001 frontmatter dependencies에 SPEC-V3R2-WF-001 추가 (WF-001 → MIG-001 단방향)
  - WF-001은 MIG-001을 참조하지 않음 (이미 의존성 없음, 유지)
- HISTORY 항목 추가 (각 수정된 SPEC에 변경 이력 기록)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- ORC SPEC 본문 변경
- WF-001 본문 변경 (이미 v1.1.0)
- MIG-001 본문 변경 (의존성 외 모든 부분 보존)
- 다른 V3R2 SPEC의 dependency graph 재검증
- plan-auditor cycle detection 로직 추가 (별도 SPEC 필요)
- 새로운 SPEC ID 생성 또는 supersession

---

## 3. Environment

- File system writable: `.moai/specs/SPEC-V3R2-ORC-*/spec.md`, `.moai/specs/SPEC-V3R2-WF-001/spec.md`, `.moai/specs/SPEC-V3R2-MIG-001/spec.md`
- plan-auditor 실행 환경 (`Agent` 호출 가능)

## 4. Assumptions

- 현재 ORC-001 ~ ORC-005 dependencies는 grep 결과 단방향이지만, audit-grade 명시적 invariant 부재
- MIG-001은 WF-001과 contract 관계이며 WF-001 변경 시 MIG-001 재실행 필요
- plan-auditor는 frontmatter dependencies를 정적 분석으로 cycle 검사 가능

## 5. Requirements (EARS)

### REQ-DEF001-001 (Ubiquitous)

The 5 ORC SPECs (SPEC-V3R2-ORC-001 ~ ORC-005) **shall** have their `dependencies` frontmatter field validated such that no ORC SPEC depends on a higher-numbered ORC SPEC (DAG topological ordering).

### REQ-DEF001-002 (Ubiquitous)

SPEC-V3R2-MIG-001 frontmatter `dependencies` field **shall** include `SPEC-V3R2-WF-001`, establishing the WF-001 → MIG-001 unidirectional contract.

### REQ-DEF001-003 (Ubiquitous)

SPEC-V3R2-WF-001 frontmatter `dependencies` field **shall** remain empty (or non-MIG-related), preserving the unidirectional flow.

### REQ-DEF001-004 (Event-Driven)

**When** any of the 5 ORC SPEC dependencies are modified, the change **shall** preserve the DAG invariant (lower-numbered SPECs may not depend on higher-numbered ones).

### REQ-DEF001-005 (Event-Driven)

**When** the cycle resolution is applied to MIG-001, a HISTORY entry **shall** be added recording the dependency addition with date and rationale.

### REQ-DEF001-006 (State-Driven)

**While** the dependencies are in their resolved state, plan-auditor **shall** be able to verify the DAG invariant via a single static pass (no runtime resolution required).

### REQ-DEF001-007 (Unwanted)

The cycle resolution **shall not** modify the body content (sections after frontmatter) of any of the 7 affected SPECs (ORC-001 ~ ORC-005, WF-001, MIG-001) except for adding HISTORY entries.

### REQ-DEF001-008 (Unwanted)

The cycle resolution **shall not** introduce new dependencies beyond the documented WF-001 → MIG-001 edge and the already-existing ORC chain.

---

## 6. Acceptance Criteria (요약)

전체 acceptance.md 참조. 핵심:

- AC-DEF001-01: ORC-001 ~ ORC-005 dependencies가 모두 단방향 DAG (낮은 번호만 참조)
- AC-DEF001-02: MIG-001 dependencies에 SPEC-V3R2-WF-001 포함
- AC-DEF001-03: WF-001 dependencies에 MIG-001 미포함
- AC-DEF001-04: 7개 SPEC 본문 무수정 (HISTORY 외)
- AC-DEF001-05: plan-auditor 실행 결과 D-CRIT-001 항목 RESOLVED
- AC-DEF001-06: 각 수정된 SPEC에 HISTORY 항목 1개씩 추가

---

## 7. Constraints

- **C1**: SPEC frontmatter 수정 시 YAML structure 무손상 (모든 기존 필드 유지)
- **C2**: HISTORY 항목은 markdown table row format으로 기존 위치 (HISTORY 섹션 마지막) 추가
- **C3**: dependencies 필드 정렬 — ID 알파벳/숫자 오름차순 (CON < ORC-001 < ORC-002 < ...)
- **C4**: WF-001 → MIG-001 contract는 MIG-001 spec.md HISTORY에 명시적 기록 (audit trail)
- **C5**: 본 SPEC 자체는 SPEC-V3R3-DEF-007 완료 후 진행 권장 (frontmatter 일관성 baseline 우선)

---

## 8. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| MIG-001에 WF-001 추가로 기존 implementation 영향 | Medium | MIG-001은 아직 implementation 미진행, dependencies만 명시 |
| ORC dependencies 정렬 시 yaml parse error | Low | Edit tool 정확한 들여쓰기 사용 |
| plan-auditor 정적 검사 false negative | Medium | Manual grep + plan-auditor cross-check |
| WF-001 / MIG-001 contract가 향후 backward dependency 유발 | Low | 단방향 invariant 본 SPEC §5 REQ-DEF001-001로 고정 |

---

## 9. Dependencies

- **선행**: SPEC-V3R3-DEF-007 (frontmatter convention baseline)

후속:
- 본 SPEC은 다른 V3R3 SPEC의 prerequisite 아님 (parallel 가능)

---

## 10. Traceability

| REQ ID | Acceptance Criteria | Source |
|--------|---------------------|--------|
| REQ-DEF001-001 | AC-DEF001-01 | V3R2 Plan-Audit D-CRIT-001 |
| REQ-DEF001-002 | AC-DEF001-02 | Master plan, MIG-001 contract |
| REQ-DEF001-003 | AC-DEF001-03 | DAG invariant |
| REQ-DEF001-004 | AC-DEF001-01 | DAG topology |
| REQ-DEF001-005 | AC-DEF001-06 | Audit trail policy |
| REQ-DEF001-006 | AC-DEF001-05 | plan-auditor static analysis |
| REQ-DEF001-007 | AC-DEF001-04 | §1.2 비목표 |
| REQ-DEF001-008 | AC-DEF001-02, AC-DEF001-03 | §1.2 |

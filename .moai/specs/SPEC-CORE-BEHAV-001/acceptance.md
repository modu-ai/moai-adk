---
spec_id: SPEC-CORE-BEHAV-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  Post-implementation SDD artifact. 6개 HARD 행동 규칙이 이미
  `.claude/rules/moai/core/moai-constitution.md`의 "Agent Core Behaviors"
  섹션에 추가되었고 템플릿(`internal/template/templates/.../moai-constitution.md`)
  에도 동일하게 반영됨. CLAUDE.md Section 7 및 template CLAUDE.md의
  cross-reference도 완료. spec.md의 R1/R2/R3를 실제 파일 내용과 대조해
  AC 역도출. plan-auditor 2026-04-24 감사 시 acceptance.md 부재 확인.
---

# Acceptance Criteria — SPEC-CORE-BEHAV-001

6개의 교차(cross-cutting) Agent Core Behaviors를 moai-constitution.md에 HARD 규칙으로 통합하고, 워크플로우 스킬에 HUMAN GATE 마커를 추가하며, CLAUDE.md에 cross-reference를 추가하는 구현의 관찰 가능한 인수 기준. 본 AC는 문서 내용의 존재성을 검증한다.

## Traceability

| REQ ID | AC ID | Test / Evidence Reference |
|--------|-------|---------------------------|
| R1 (Agent Core Behaviors Section) | AC-001 ~ AC-007 | `.claude/rules/moai/core/moai-constitution.md:175-266`, 템플릿 동기본 |
| R2 (HUMAN GATE Markers) | AC-008 | `.claude/skills/moai/workflows/{plan,run,sync}.md` 6개 마커 전수 검증 완료 (2026-04-24) |
| R3 (Cross-Reference Consistency) | AC-009 | `CLAUDE.md:20`, 템플릿 CLAUDE.md:20 |

## AC-001: moai-constitution.md에 "Agent Core Behaviors" 섹션이 존재한다

**Given** `.claude/rules/moai/core/moai-constitution.md` 파일에서,
**When** 문서를 검색하면,
**Then** `## Agent Core Behaviors` 레벨-2 헤딩이 존재해야 하며, 섹션 intro 문장이 "Six cross-cutting HARD behaviors that apply to all agents regardless of active skill or workflow phase"로 시작해야 한다.

**Verification**: `.claude/rules/moai/core/moai-constitution.md:176,178` — 정확한 헤딩과 intro 문구. 템플릿 카피본도 동일(`internal/template/templates/.claude/rules/moai/core/moai-constitution.md`).

## AC-002: Rule 1 "Surface Assumptions [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션 안에서,
**When** 번호 1번 HARD 규칙을 조회하면,
**Then** 헤딩 `### 1. Surface Assumptions [HARD]`가 존재하고, format 블록에 `ASSUMPTIONS I'M MAKING:` 템플릿이 포함되며, Cross-reference `CLAUDE.md Section 7 Rule 5`가 명시되어야 한다.

**Verification**: `moai-constitution.md:180-194` — 헤딩 line 180, format block line 185-190, cross-ref line 192.

## AC-003: Rule 2 "Manage Confusion Actively [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션에서,
**When** 번호 2번 HARD 규칙을 조회하면,
**Then** `### 2. Manage Confusion Actively [HARD]` 헤딩이 존재하고 4단계 절차(STOP → Name → Present → Wait)가 정의되어야 하며, anti-pattern 예시가 포함되어야 한다.

**Verification**: `moai-constitution.md:196-206` — 헤딩 line 196, Steps 1-4 line 200-204, Anti-pattern line 206.

## AC-004: Rule 3 "Push Back When Warranted [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션에서,
**When** 번호 3번 HARD 규칙을 조회하면,
**Then** `### 3. Push Back When Warranted [HARD]` 헤딩과 "Sycophancy is a failure mode" 문구가 존재하고, "When to push back" / "How to push back" 서브리스트가 포함되며, anti-pattern `"Of course!"` 예시가 명시되어야 한다.

**Verification**: `moai-constitution.md:208-223` — 헤딩 line 208, Sycophancy 문구 line 210, anti-pattern line 223.

## AC-005: Rule 4 "Enforce Simplicity [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션에서,
**When** 번호 4번 HARD 규칙을 조회하면,
**Then** `### 4. Enforce Simplicity [HARD]` 헤딩이 존재하고, "Can this be done in fewer lines without loss of clarity?"를 포함한 체크 질문이 있으며, Cross-reference `TRUST 5 Readable principle`이 명시되어야 한다.

**Verification**: `moai-constitution.md:225-236` — 헤딩 line 225, 체크 질문 line 229-232, cross-ref line 234.

## AC-006: Rule 5 "Maintain Scope Discipline [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션에서,
**When** 번호 5번 HARD 규칙을 조회하면,
**Then** `### 5. Maintain Scope Discipline [HARD]` 헤딩이 존재하고 "Do NOT" 목록(comment 제거, 정리 리팩토, 인접 시스템 수정, 미승인 삭제, 명세 외 기능 추가) 5개 항목이 정의되며, Cross-reference `CLAUDE.md Section 7 Rule 2`가 명시되어야 한다.

**Verification**: `moai-constitution.md:238-251` — 헤딩 line 238, Do NOT 목록 line 242-247, cross-ref line 249.

## AC-007: Rule 6 "Verify, Don't Assume [HARD]"가 존재한다

**Given** Agent Core Behaviors 섹션에서,
**When** 번호 6번 HARD 규칙을 조회하면,
**Then** `### 6. Verify, Don't Assume [HARD]` 헤딩이 존재하고 Evidence requirements 4가지(Tests passing, Build succeeding, File created, Behavior correct)가 정의되며, anti-pattern `"Claiming 'tests pass' without running them"`이 명시되어야 한다.

**Verification**: `moai-constitution.md:253-265` — 헤딩 line 253, evidence list line 257-261, anti-pattern line 265.

## AC-008: 워크플로우 스킬에 HUMAN GATE 마커가 추가된다 (PASS)

**Given** spec.md R2는 plan.md/run.md/sync.md에 6개 HUMAN GATE 마커 추가를 요구함,
**When** `.claude/skills/moai/workflows/{plan,run,sync}.md`를 검사하면,
**Then** 각 파일에 `### HUMAN GATE: [Gate Name]` 형식 섹션과 관련 필드들이 존재해야 한다.

**Verification (2026-04-24):** `grep -n "HUMAN GATE"` 전수 검사 결과 정확히 **6개** 마커 확인:
- `plan.md:308` — HUMAN GATE: Plan Review
- `plan.md:607` — HUMAN GATE: SPEC Quality Validation
- `run.md:260` — HUMAN GATE: Plan Approval
- `run.md:642` — HUMAN GATE: Implementation Complete
- `sync.md:128` — HUMAN GATE: Pre-Sync Quality
- `sync.md:584` — HUMAN GATE: Documentation Scope

Template(`internal/template/templates/.claude/skills/moai/workflows/`)과 로컬(`.claude/skills/moai/workflows/`) 양쪽에서 동일 6개 마커가 동일 라인 번호로 확인됨 → template/local drift 없음. **R2 requirement SATISFIED**.

## AC-009: CLAUDE.md가 moai-constitution.md Agent Core Behaviors를 cross-reference한다

**Given** `CLAUDE.md`와 템플릿 `internal/template/templates/CLAUDE.md` 양쪽에서,
**When** line 20을 조회하면,
**Then** `Core principles (1-4) and six Agent Core Behaviors (consolidated cross-cutting rules) are defined in .claude/rules/moai/core/moai-constitution.md. Development safeguards (5-9) are detailed in Section 7.` 문장이 존재해야 한다.

**Verification**: `CLAUDE.md:20` 및 `internal/template/templates/CLAUDE.md:20` — grep 결과로 두 파일 모두 정확히 동일 문장 확인됨.

## Edge Cases

- **EC-01**: 섹션 위치 — R1 수락 기준은 "after Lessons Protocol and before URL Verification" 이었으며, 실제 배치는 "Lessons Protocol (line ~175) 이후" + `<!-- moai:evolvable-start id="agent-core-behaviors" -->` 마커로 래핑되어 요구사항 충족.
- **EC-02**: 템플릿-로컬 싱크 — 템플릿 파일과 로컬 파일이 동일 6개 규칙을 포함해야 하며 `grep "Agent Core Behaviors"` 결과 두 파일 모두 매칭.
- **EC-03**: CLAUDE.md 40K 문자 제한 — cross-reference 단일 라인만 추가하여 제한 내 유지.

## Definition of Done

- [x] Agent Core Behaviors 섹션 추가 (AC-001) with evolvable-zone markers
- [x] 6개 HARD 규칙 모두 존재 (AC-002 ~ AC-007)
  - [x] Surface Assumptions
  - [x] Manage Confusion Actively
  - [x] Push Back When Warranted
  - [x] Enforce Simplicity
  - [x] Maintain Scope Discipline
  - [x] Verify, Don't Assume
- [x] 각 규칙이 name + rule + cross-reference + anti-pattern 구조를 따름
- [x] 템플릿과 로컬 파일 동기화 (AC-001 ~ AC-007 모두)
- [x] CLAUDE.md cross-reference (AC-009)
- [x] **PASS**: HUMAN GATE 마커 6개(plan×2 + run×2 + sync×2) 전수 검증 완료 (2026-04-24)

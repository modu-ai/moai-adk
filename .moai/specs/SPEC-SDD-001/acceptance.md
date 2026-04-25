---
spec_id: SPEC-SDD-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOCUMENTED SPEC: `status: completed` 표시 상태에서 acceptance.md 가 누락된
  SDD 갭을 메움 (plan-auditor 2026-04-24 감사). 3 개 REQ (harness.yaml / Delta Markers /
  spec-compact.md) 구현체는 `.moai/config/sections/harness.yaml`,
  `.claude/skills/moai/workflows/{plan,run,moai}.md` 에 반영되어 있음.
  Large-scope SPEC 이므로 AC 는 "관찰 가능한 behavior" (yaml 필드 존재, 워크플로우 스킬
  내 phase 구조, spec-compact 생성 규칙) 중심으로 구성.
---

# Acceptance Criteria — SPEC-SDD-001 (SDD Integration: Harness + Delta + Compact)

## Traceability

| REQ ID   | AC ID     | Test/Evidence Reference                                                                         |
|----------|-----------|-------------------------------------------------------------------------------------------------|
| REQ-001  | AC-001    | `internal/template/templates/.moai/config/sections/harness.yaml` (전체 파일)                     |
| REQ-001  | AC-002    | `harness.yaml:51-83` (3 레벨: minimal/standard/thorough)                                          |
| REQ-001  | AC-003    | `harness.yaml:10-14` (mode_defaults: solo/team/cg)                                                |
| REQ-001  | AC-004    | `harness.yaml:16-32` (auto_detection 규칙)                                                        |
| REQ-002  | AC-005    | `plan.md:398-402` (Delta Marker 4 타입 템플릿)                                                    |
| REQ-002  | AC-006    | `run.md:428-447` (ANALYZE 단계 Delta Marker 처리 로직)                                             |
| REQ-003  | AC-007    | `plan.md:407-419` (spec-compact.md 자동 생성 규칙)                                                |
| REQ-003  | AC-008    | `run.md:74, 92, 620` (Run 페이즈의 spec-compact 우선 로딩 + fallback)                              |
| Excl.    | AC-009    | `run.md:74` (fallback to spec.md when spec-compact.md 부재)                                       |

## AC-001: harness.yaml 파일 존재 및 구조

- **Given** MoAI-ADK 템플릿이 배포된 프로젝트
- **When** `.moai/config/sections/harness.yaml` 파일을 확인한다
- **Then** 파일이 존재하며 최상위에 `harness:` 키가 선언되어 있다
- **And** 하위에 `default_profile`, `mode_defaults`, `auto_detection`, `escalation`,
  `effort_mapping`, `levels`, `model_upgrade_review` 키가 모두 존재한다
- **Verification**: `internal/template/templates/.moai/config/sections/harness.yaml:1-132`

## AC-002: 3 레벨 harness 정의 (minimal / standard / thorough)

- **Given** `harness.yaml` 의 `levels:` 섹션
- **When** 3 개 레벨의 스키마를 확인한다
- **Then** 정확히 3 개 레벨 (`minimal`, `standard`, `thorough`) 이 정의된다
- **And** 각 레벨은 `description`, `evaluator`, `sprint_contract`, (optional) `skip_phases`,
  (optional) `evaluator_mode`, (optional) `plan_audit` 를 포함한다
- **And** 레벨별 속성 매핑:
  - minimal: `evaluator: false`, `sprint_contract: false`, `skip_phases` 목록 존재
  - standard: `evaluator: true`, `evaluator_mode: "final-pass"`, `sprint_contract: false`
  - thorough: `evaluator: true`, `evaluator_mode: "per-sprint"`, `sprint_contract: true`,
    `playwright_testing: true`
- **Verification**: `harness.yaml:51-83`

## AC-003: 실행 모드별 default harness 매핑

- **Given** `harness.yaml:mode_defaults` 섹션
- **When** solo/team/cg 모드의 default 값을 확인한다
- **Then** 다음 매핑이 선언된다:
  - `solo: auto` (auto-detect)
  - `team: auto` (auto-detect)
  - `cg: thorough` (CG 모드는 항상 thorough)
- **Verification**: `harness.yaml:10-14`

## AC-004: Complexity Estimator 자동 감지 규칙

- **Given** `harness.yaml:auto_detection` 섹션
- **When** `enabled: true` 이고 auto_detection 규칙을 조회한다
- **Then** 3 레벨별 조건이 선언된다:
  - minimal: `file_count <= 3 AND single_domain` OR `spec_type in [bugfix, docs, config]`
  - standard: `file_count > 3 OR multi_domain` OR `spec_type in [feature, refactor]`
  - thorough: `security_keywords OR payment_keywords present` OR `spec_priority == critical`
    OR `domain in [auth, payment, migration, public_api]`
- **And** `moai.md` 워크플로우에 Complexity Estimator 계산식이 구현되어 있다
  (`complexity_score = domain_count * 2 + file_count / 3`)
- **Verification**: `harness.yaml:16-32`, `.claude/skills/moai/workflows/moai.md:224-228`

## AC-005: Delta Marker 4 타입 템플릿 (plan.md)

- **Given** Plan 워크플로우 스킬 (`.claude/skills/moai/workflows/plan.md`)
- **When** `[DELTA]` 섹션 템플릿을 확인한다
- **Then** 정확히 4 가지 마커 타입이 템플릿에 문서화되어 있다:
  - `[EXISTING]` — 변경 없이 characterization test 만
  - `[MODIFY]` — 수정 전 characterization test 요구
  - `[NEW]` — 전체 구현 + 신규 test
  - `[REMOVE]` — dependency 분석 후 안전 삭제
- **Verification**: `plan.md:398-402`

## AC-006: Run 페이즈 Delta Marker 처리 로직

- **Given** Run 워크플로우 스킬 (`.claude/skills/moai/workflows/run.md`)
- **When** DDD ANALYZE 페이즈가 [DELTA] 마커가 있는 SPEC 을 처리한다
- **Then** 다음 순서로 routing 된다:
  1. `[EXISTING]` 항목 — characterization test 만 생성 (코드 수정 금지)
  2. `[MODIFY]` 항목 — characterization test 생성 후 수정, test 통과 재검증
  3. `[NEW]` 항목 — 전체 구현 사이클 (DDD ANALYZE-PRESERVE-IMPROVE 또는 TDD RED-GREEN-REFACTOR)
  4. `[REMOVE]` 항목 — 호출자/의존자 확인 후 안전 삭제
- **Verification**: `run.md:428-447`

## AC-007: spec-compact.md 자동 생성 규칙 (Plan 완료 시)

- **Given** Plan 페이즈가 성공적으로 완료된 시점
- **When** plan.md 워크플로우가 산출물 생성을 마친다
- **Then** `.moai/specs/SPEC-{ID}/spec-compact.md` 파일이 자동 생성된다
- **And** spec-compact.md 는 다음 섹션만 포함한다:
  - EARS 형식 REQ 항목 (REQ-XXX)
  - Acceptance criteria (Given/When/Then)
  - Files to modify 목록
  - Exclusions 목록
- **And** research notes / plan 상세 / discussion history / overview 는 제외한다
- **And** 원본 spec.md 는 삭제 또는 수정되지 않는다
- **Verification**: `plan.md:407-419`, `plan.md:719, 725`

## AC-008: Run 페이즈 spec-compact.md 우선 로딩

- **Given** Run 페이즈 진입 시점
- **When** SPEC 문서를 로딩한다
- **Then** `.moai/specs/SPEC-{ID}/spec-compact.md` 가 존재하면 우선 로딩한다 (~30% 토큰 절감)
- **And** spec-compact.md 가 없으면 full spec.md 로 fallback (backward compatible)
- **Verification**: `run.md:74` ("If `.moai/specs/SPEC-{ID}/spec-compact.md` exists: Load
  spec-compact.md (~30% token savings)"), `run.md:92, 620`

## AC-009: Backward Compatibility (legacy SPEC)

- **Given** spec-compact.md 가 존재하지 않는 레거시 SPEC (v2.9.0 이전)
- **When** Run 페이즈가 시작된다
- **Then** full spec.md 를 오류 없이 로딩한다
- **And** [DELTA] 마커가 없는 경우 delta marker 처리 로직은 실행되지 않는다
- **Verification**: `run.md:74` (conditional "if exists"), `run.md:428-431` ([DELTA] 스캔 후
  없으면 fall-through)

## Exclusions Verification

| Exclusion (spec.md) | Status | Verification |
|---------------------|--------|--------------|
| end user 에게 harness level CLI 노출 금지 | OK | `harness.yaml` 에 CLI flag 정의 부재, auto-detect 만 제공 |
| greenfield 프로젝트에 delta marker 요구 금지 | OK | `run.md:428-431` 는 marker 부재 시 조용히 fall-through |
| 원본 spec.md 삭제/수정 금지 | OK | `plan.md:407-419` 는 별도 파일 spec-compact.md 생성 |
| 마커 없을 시 delta 로직 미적용 | OK | `run.md:428-431` 는 `[EXISTING/MODIFY/NEW/REMOVE]` 패턴 매칭 후에만 실행 |

## Quality Gate Criteria

- [x] harness.yaml 전체 스키마 (7 최상위 키) 배포
- [x] 3 레벨 (minimal/standard/thorough) 속성 매핑 완전
- [x] mode_defaults 매핑 (solo/team/cg) 선언
- [x] auto_detection 규칙 3 레벨 조건 정의
- [x] Complexity Estimator 계산식 moai.md 에 구현
- [x] [DELTA] 4 타입 (EXISTING/MODIFY/NEW/REMOVE) 템플릿 존재
- [x] Delta Marker 처리 순서 run.md 에 문서화
- [x] spec-compact.md 자동 생성 규칙 plan.md 에 선언
- [x] Run 페이즈 spec-compact 우선 로딩 + fallback 구현
- [x] Legacy SPEC backward compatibility 보장

## Definition of Done

- [x] `.moai/config/sections/harness.yaml` 배포
- [x] `.claude/skills/moai/workflows/plan.md` 에 Delta Marker 템플릿 + spec-compact 생성 로직
- [x] `.claude/skills/moai/workflows/run.md` 에 Delta Marker routing + spec-compact 로딩 로직
- [x] `.claude/skills/moai/workflows/moai.md` 에 Complexity Estimator 구현
- [x] spec.md frontmatter `status: completed` 상태
- [x] backfill acceptance.md (본 문서)

## Observable Integration Evidence

### Scenario: Complex SPEC Auto-Detects as thorough

- **Given** SPEC 에 domain_count=3, file_count=15, security 키워드 포함
- **When** `/moai run SPEC-XXX` 실행 시 Complexity Estimator 가 동작
- **Then** `harness.yaml:auto_detection.rules.thorough` 조건 매치
- **And** `levels.thorough.evaluator=true`, `sprint_contract=true`, `evaluator_mode=per-sprint`,
  `playwright_testing=true` 활성
- **Verification**: `moai.md:224-228` (Estimator), `harness.yaml:71-83` (thorough 속성)

### Scenario: Brownfield SPEC with MODIFY Markers

- **Given** plan.md 에 `[MODIFY] existing-service.go - refactor auth flow` 포함된 SPEC
- **When** Run 페이즈 DDD ANALYZE 실행
- **Then** run.md Delta-aware routing 이 [MODIFY] 항목을 감지
- **And** characterization test 를 현재 코드에 대해 먼저 작성 / 통과 검증 후 수정 적용
- **Verification**: `run.md:444-447` (처리 순서)

### Scenario: Token Savings via spec-compact.md

- **Given** Plan 완료 시점에 full spec.md 가 8000 tokens
- **When** plan.md 워크플로우가 spec-compact 생성 단계 실행
- **Then** spec-compact.md 는 약 5600 tokens 이하 (~30% 감소)
- **And** 이후 `/moai run` 은 spec-compact.md 를 로딩하여 Run 페이즈 토큰 예산 절감
- **Verification**: `plan.md:419` ("Run phase loads spec-compact.md (~30% token savings)")

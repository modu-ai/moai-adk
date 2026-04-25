---
spec_id: SPEC-EVAL-001
version: 1.0.0
status: backfilled
created_at: 2026-04-24
author: manager-spec (backfill)
backfill_reason: |
  REVERSE-DOCUMENTED SPEC: 구현이 먼저 상용화되고 acceptance.md가 누락된 SDD 갭을 메움
  (plan-auditor 2026-04-24 감사). 기능은 `internal/template/templates/.claude/agents/moai/evaluator-active.md`
  및 `.../.claude/skills/moai/workflows/run.md` (Phase 2.0, 2.8a, 2.8b) 에 이미 반영되어 있음.
  본 문서는 관찰된 구현 동작을 EARS 5 REQ에 역매핑하여 AC를 확정한다.
---

# Acceptance Criteria — SPEC-EVAL-001 (evaluator-active Agent & Sprint Contract)

## Traceability

| REQ ID   | AC ID   | Test/Evidence Reference                                                                                      |
|----------|---------|--------------------------------------------------------------------------------------------------------------|
| REQ-001  | AC-001  | `internal/template/templates/.claude/agents/moai/evaluator-active.md:13-15` (model/effort/permissionMode)    |
| REQ-001  | AC-002  | `evaluator-active.md:38-43` (HARD RULES, 회의론적 평가 지시)                                                     |
| REQ-001  | AC-003  | `evaluator-active.md:48-55` (4 차원 rubric 및 Security HARD THRESHOLD)                                        |
| REQ-002  | AC-004  | `run.md:401-416` (Phase 2.0 Sprint Contract Negotiation)                                                     |
| REQ-002  | AC-005  | `run.md:416, 424` (contract.md 경로), `harness.yaml:71-77` (thorough 전용)                                      |
| REQ-003  | AC-006  | `run.md:613-635` (Phase 2.8a evaluator-active)                                                                |
| REQ-003  | AC-007  | `run.md:653` (Phase 2.8b manager-quality MANDATORY)                                                           |
| REQ-003  | AC-008  | `run.md:623-629` (retry/escalation 로직) — PARTIAL (max 3 retries 문구는 스펙대로지만 코드상 명시 확인 필요)     |
| REQ-004  | AC-009  | `evaluator-active.md:104-108` (Mode-Specific Deployment 섹션)                                                 |
| REQ-005  | AC-010  | `evaluator-active.md:99-102` + `harness.yaml:62-77` (intervention modes mapping)                              |

## AC-001: evaluator-active 에이전트 정의 존재 및 필수 frontmatter

- **Given** MoAI-ADK 템플릿이 배포된 프로젝트
- **When** `.claude/agents/moai/evaluator-active.md` 를 확인한다
- **Then** 파일이 존재하며 frontmatter 에 다음 값이 포함된다:
  - `model: sonnet`
  - `permissionMode: plan` (read-only)
  - `tools: Read, Grep, Glob, Bash, mcp__sequential-thinking__sequentialthinking`
  - `effort: high`
- **Verification**: `internal/template/templates/.claude/agents/moai/evaluator-active.md:12-15`

## AC-002: 회의론적 평가(Skeptical) HARD 규칙 강제

- **Given** evaluator-active 에이전트 정의 본문
- **When** agent body 의 "Skeptical Evaluation Mandate" 섹션을 읽는다
- **Then** 다음 HARD 규칙이 명시되어 있다:
  - "NEVER rationalize acceptance"
  - "Do NOT award PASS without concrete evidence"
  - "When in doubt, FAIL"
  - "PASS in one area does NOT offset a FAIL in another"
- **Verification**: `evaluator-active.md:38-44`

## AC-003: 4 차원 평가 가중치 및 Security HARD 임계값

- **Given** evaluator-active 본문의 Evaluation Dimensions 표
- **When** 차원/가중치/실패 조건을 확인한다
- **Then** 다음 가중치가 그대로 선언되어 있다:
  - Functionality 40%, Security 25%, Craft 20%, Consistency 15%
- **And** "Security dimension FAIL = Overall FAIL (regardless of other scores)" HARD THRESHOLD 라벨이 선언된다
- **Verification**: `evaluator-active.md:48-55`

## AC-004: Phase 2.0 Sprint Contract Negotiation 정의 (thorough 전용)

- **Given** harness level = thorough 인 SPEC
- **When** Run 워크플로우 스킬 `.claude/skills/moai/workflows/run.md` 를 로드한다
- **Then** Phase 2.0 "Sprint Contract Negotiation" 섹션이 존재하고 Phase 2.1 보다 먼저 실행되도록 순서가 정의되어 있다
- **And** Phase 2.0 에서 evaluator-active 가 manager-ddd/tdd 의 구현 계획을 검토한다
- **Verification**: `internal/template/templates/.claude/skills/moai/workflows/run.md:401-416`

## AC-005: contract.md 산출물 경로 및 thorough 게이팅

- **Given** Phase 2.0 가 실행되어 negotiation 이 완료된 상태
- **When** 산출물 저장 위치를 조회한다
- **Then** 결과가 `.moai/specs/SPEC-{ID}/contract.md` 에 저장된다
- **And** harness.yaml `levels.thorough.sprint_contract: true`, `levels.standard.sprint_contract: false` 로 thorough 에서만 활성화된다
- **Verification**: `run.md:416, 424`, `harness.yaml:64, 77`

## AC-006: Phase 2.8a 능동 평가 (evaluator-active)

- **Given** 구현 단계가 종료된 SPEC
- **When** Phase 2.8a 가 실행된다
- **Then** evaluator-active 가 SPEC acceptance criteria 및 (thorough 의 경우) contract.md 에 대해 평가를 수행한다
- **And** 4 차원 점수와 overall PASS/FAIL verdict 를 포함한 Evaluation Report 를 산출한다
- **Verification**: `run.md:613-629`

## AC-007: Phase 2.8b manager-quality 정적 검증 분리

- **Given** Phase 2.8a 가 PASS 또는 max retry 도달
- **When** Phase 2.8b 가 실행된다
- **Then** manager-quality 가 기존 TRUST 5 정적 검증을 수행한다 ([MANDATORY] 라벨)
- **And** Phase 2.8a 와 2.8b 는 책임 분리되어 있다 (2.8a = 능동 평가, 2.8b = 정적 검증)
- **Verification**: `run.md:653` ("Phase 2.8b: TRUST 5 Static Verification (manager-quality) [MANDATORY]")

## AC-008: 평가 실패 시 재시도 루프 (PARTIAL)

- **Given** Phase 2.8a 에서 evaluator-active 가 FAIL 판정
- **When** 재시도 사이클이 시작된다
- **Then** 구현 에이전트가 피드백을 반영하여 재시도한다
- **Gap**: spec.md REQ-003 의 "maximum of 3 retry cycles" 는 run.md 본문에서 재시도 루프 존재는 명시되나 "최대 3 cycles" 수치가 코드 상수가 아닌 가이드 텍스트로만 확인됨. 강제 enforcement 는 escalation 로직 (`harness.yaml:34-40 escalation.max_escalations: 2`) 으로 대체됨
- **Status**: PARTIAL — 재시도 루프 존재 OK, 3 회 하드 리밋 enforcement 는 보증되지 않음
- **Verification**: `run.md:623-629`, `harness.yaml:34-40`

## AC-009: 모드별 배포 (Sub-agent / Team / CG)

- **Given** evaluator-active 가 호출되는 시점의 실행 모드
- **When** run.md 가 모드를 판단한다
- **Then** 다음 매핑으로 배포된다:
  - Sub-agent 모드: `Agent(subagent_type="evaluator-active")` 호출
  - Team 모드: reviewer role teammate 가 SendMessage 로 평가 태스크 수신
  - CG 모드: Leader(Claude) 가 별도 agent spawn 없이 직접 평가 수행
- **Verification**: `evaluator-active.md:104-108`, `run.md:420, 635`

## AC-010: Intervention Modes (final-pass / per-sprint)

- **Given** harness level 설정
- **When** evaluator-active 개입 시점이 결정된다
- **Then** 다음 매핑을 따른다:
  - `standard` harness → `final-pass` mode (Phase 2.8a 에서만 1회 평가)
  - `thorough` harness → `per-sprint` mode (Phase 2.0 + Phase 2.8a 모두 평가)
  - `minimal` harness → evaluator-active 비활성 (`evaluator: false`)
- **Verification**: `harness.yaml:52-77` (`levels.minimal.evaluator: false`, `levels.standard.evaluator_mode: "final-pass"`, `levels.thorough.evaluator_mode: "per-sprint"`), `evaluator-active.md:99-102`

## Exclusions Verification

| Exclusion (spec.md) | Status | Verification |
|---------------------|--------|--------------|
| manager-quality 를 대체하지 않음 | OK | Phase 2.8a (evaluator-active) 와 Phase 2.8b (manager-quality) 가 분리 (`run.md:613, 653`) |
| 기존 test 파일 수정 금지 | OK | `permissionMode: plan` (read-only) (`evaluator-active.md:15`) |
| minimal harness 비활성 | OK | `harness.yaml:57` `levels.minimal.evaluator: false` |
| 코드 작성/수정 금지 | OK | `permissionMode: plan` + tools 화이트리스트에 Write/Edit 부재 (`evaluator-active.md:12-15`) |

## Quality Gate Criteria

- [x] evaluator-active 에이전트가 `permissionMode: plan` 으로 read-only 동작
- [x] 4 차원 가중치 (40/25/20/15) 합이 100%
- [x] Security FAIL hard threshold 문서화
- [x] Phase 2.0 (Sprint Contract) 정의 및 thorough 게이팅
- [x] Phase 2.8a/2.8b 분리 및 책임 경계 명확
- [x] 3 가지 모드 (sub-agent/team/cg) 배포 경로 정의
- [x] 2 가지 intervention mode (final-pass/per-sprint) 정의
- [ ] Max 3 retries 하드 enforcement — PARTIAL (run.md 가이드 텍스트로만 존재)

## Definition of Done

- [x] `.claude/agents/moai/evaluator-active.md` 배포
- [x] run.md 에 Phase 2.0, 2.8a, 2.8b 구현
- [x] harness.yaml 에 intervention mode 매핑 선언
- [x] 4 개 evaluator-profiles 파일 존재 (교차 검증: SPEC-EVALLIB-001)
- [x] backfill acceptance.md (본 문서)

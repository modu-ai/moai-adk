---
id: SPEC-V3R3-PROJECT-HARNESS-001
title: Project Harness Activation — 16Q Interview + 5-Layer
version: "0.1.0"
status: draft
created_at: 2026-04-26
updated_at: 2026-04-26
author: manager-spec
priority: P0
labels: [project, harness, interview, integration, v3r3, phase-c]
issue_number: null
phase: "v3.0.0 R3 — Phase C — Project Harness Activation"
module: ".claude/skills/moai/workflows/project.md, .moai/harness/, .claude/agents/my-harness/, .claude/skills/my-harness-*/, internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync,design}.md"
depends_on:
  - SPEC-V3R3-HARNESS-001
related_specs:
  - SPEC-V3R3-HARNESS-LEARNING-001
  - SPEC-V3R3-DESIGN-PIPELINE-001
breaking: false
bc_id: []
lifecycle: spec-anchored
target_release: v2.17.0
---

# SPEC-V3R3-PROJECT-HARNESS-001: Project Harness Activation

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 0.1.0   | 2026-04-26 | manager-spec | Initial draft. Phase C P0 — `/moai project` Phase 5+ 소크라테스 인터뷰 (16 질문 / 4 라운드) + 5-Layer 통합 장치 (사용자 통찰 반영). |

---

## 1. Goal (목적)

`/moai project` 명령에 Phase 5+를 추가하여 16 질문 / 4 라운드 소크라테스 인터뷰를 수행하고, 그 결과를 바탕으로 SPEC-V3R3-HARNESS-001의 `moai-meta-harness` 스킬을 호출해 사용자 영역 dynamic harness (`.claude/agents/my-harness/*`, `.claude/skills/my-harness-*/`, `.moai/harness/`)를 자동 생성한다. 동시에 5-Layer 통합 장치 (frontmatter triggers, workflow.yaml.harness 섹션, CLAUDE.md @import marker, workflow static import line, `.moai/harness/` user directory)를 활성화하여 새 세션 시작 시 harness가 자동으로 인식·활용되도록 한다.

### 1.1 Background

- SPEC-V3R3-HARNESS-001은 meta-harness skill을 도입하여 사용자 영역에 harness 산출물을 동적 생성할 수 있는 메커니즘을 제공한다.
- 그러나 사용자가 어떤 도메인 / 방법론 / 디자인툴 / 보안 요구사항을 가지고 있는지 인터뷰 없이는 meta-harness가 정확한 산출물을 생성할 수 없다.
- 또한 생성된 harness가 새 세션에서 자동으로 활용되려면 5-Layer 통합 장치가 필요하다 (사용자 통찰 ★ — handoff §0.2 verbatim).
- moai-managed workflow files (`.claude/skills/moai/workflows/{plan,run,sync,design}.md`)는 `moai update` 시 갱신되므로 사용자 customization을 직접 작성하면 보존되지 않는다. 따라서 template static import line + `.moai/harness/` user directory convention이 필요하다 (Layer 4 + Layer 5).

### 1.2 Non-Goals

- 본 SPEC은 `moai-meta-harness` skill 자체를 정의하지 않는다 (SPEC-V3R3-HARNESS-001 책임).
- 본 SPEC은 harness self-learning / auto-evolution을 다루지 않는다 (SPEC-V3R3-HARNESS-LEARNING-001 책임).
- 본 SPEC은 16개 정적 skills 제거 마이그레이션을 다루지 않는다 (SPEC-V3R3-HARNESS-001 BC-V3R3-007 범위).
- 본 SPEC은 design-pipeline (Path A / B1 / B2) 분기를 직접 구현하지 않는다 (SPEC-V3R3-DESIGN-PIPELINE-001 책임). 단, Q5/Q7/Q8 답변은 design-pipeline 실행 가능성을 결정한다.

---

## 2. Scope

### 2.1 In Scope

- `/moai project` Phase 5+ 신설 (`.claude/skills/moai/workflows/project.md` 확장).
- 16 질문 / 4 라운드 소크라테스 인터뷰 (`AskUserQuestion`, max 4 질문 per call).
- meta-harness 호출 → `.claude/agents/my-harness/*` + `.claude/skills/my-harness-*/` + `.moai/harness/*` 생성.
- 5-Layer 통합 장치 활성화:
  - L1: harness skill frontmatter `triggers` 자동 inject.
  - L2: `workflow.yaml.harness` 섹션 갱신.
  - L3: `CLAUDE.md` `<!-- moai:harness-start --> ~ <!-- moai:harness-end -->` marker inject.
  - L4: `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync,design}.md` 한 줄 정적 import line 추가.
  - L5: `.moai/harness/` user directory convention (main.md, *-extension.md, chaining-rules.yaml, interview-results.md, README.md).
- 인터뷰 결과 영구 기록 (`.moai/harness/interview-results.md`).
- `moai update` 안전성 (사용자 영역 보존).
- `moai doctor` 통합 (5-Layer 활성 상태 진단).

### 2.2 Out of Scope

- meta-harness skill 본체 (SPEC-V3R3-HARNESS-001).
- harness self-learning (SPEC-V3R3-HARNESS-LEARNING-001).
- design-pipeline Path A/B1/B2 분기 구현 (SPEC-V3R3-DESIGN-PIPELINE-001).
- 정적 16 skills 제거 (SPEC-V3R3-HARNESS-001 BC-V3R3-007).
- 다국어 인터뷰 (현재 conversation_language 기준만; multilingual prompt expansion은 별도).

---

## 3. Stakeholders

| Role | Interest |
|------|----------|
| User (developer) | 프로젝트 초기화 시 단 한 번의 인터뷰로 dynamic harness 자동 활성화. |
| MoAI orchestrator | `/moai project` 실행 시 5-Layer를 일관되게 적용. |
| manager-spec | 인터뷰 답변 기반으로 도메인 SPEC 작성 시 ios-architect 등 chain 사용. |
| manager-tdd / manager-ddd | `workflow.yaml.harness.chaining_rules`를 read하여 expert-* 호출 전후에 my-harness 에이전트 chain. |
| moai update 로직 | 사용자 영역 (`.moai/harness/`, `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/`)을 보존. |
| moai doctor | 5-Layer 활성 상태 + `my-harness-*` prefix 충돌 진단. |

---

## 4. EARS Requirements

### 4.1 Ubiquitous Requirements

- **REQ-PH-001** [Ubiquitous] `/moai project` 명령은 기존 Phase 1-4 (project doc generation) 완료 후 항상 Phase 5+ Harness Activation을 수행해야 한다.
- **REQ-PH-002** [Ubiquitous] 16 질문 인터뷰는 `AskUserQuestion` 도구를 통해 4 라운드 (Round 1: Q1-Q4, Round 2: Q5-Q8, Round 3: Q9-Q12, Round 4: Q13-Q16)로 분할 수행해야 한다.
- **REQ-PH-003** [Ubiquitous] 인터뷰 완료 시 시스템은 답변 16개를 `.moai/harness/interview-results.md`에 영구 기록해야 한다 (Q1-Q16 각각 question + answer + timestamp).

### 4.2 Event-Driven Requirements

- **REQ-PH-004** [Event-Driven] WHEN Round 4 Q16 (최종 확인) 응답이 "Confirm"인 경우, THEN 시스템은 `moai-meta-harness` skill을 호출하여 dynamic 산출물을 생성해야 한다.
- **REQ-PH-005** [Event-Driven] WHEN meta-harness 산출물 생성이 완료되면, THEN 시스템은 5-Layer 통합 장치 (L1, L2, L3, L4 verify, L5)를 순차 활성화해야 한다.
- **REQ-PH-006** [Event-Driven] WHEN 새 Claude Code 세션이 시작되면, THEN `CLAUDE.md`의 `<!-- moai:harness-start -->` marker가 follow-able @import로 동작하여 harness customization이 context에 자동 포함되어야 한다.
- **REQ-PH-007** [Event-Driven] WHEN `/moai run SPEC-XXX` 실행 시 manager-tdd가 Phase 2를 시작하면, THEN `workflow.yaml.harness.chaining_rules`를 read하여 `before_specialist` (예: ios-architect) → `expert-*` → `after_specialist` (예: swiftui-engineer) 순서로 chain을 적용해야 한다.

### 4.3 State-Driven Requirements

- **REQ-PH-008** [State-Driven] WHILE harness가 활성 상태인 동안 (`.moai/harness/main.md` 존재), `manager-tdd` / `manager-ddd` / `manager-spec`가 paths 매칭되는 파일 작업 시 `my-harness-*` skill이 자동 활성화되어야 한다 (L1 frontmatter triggers).
- **REQ-PH-009** [State-Driven] WHILE `moai update`가 실행되는 동안, 시스템은 `.moai/harness/`, `.claude/agents/my-harness/`, `.claude/skills/my-harness-*/` 디렉터리를 절대 수정·삭제하지 않아야 한다.

### 4.4 Unwanted Behavior Requirements

- **REQ-PH-010** [Unwanted] IF 사용자가 인터뷰 도중 abort하면, THEN 시스템은 부분 산출물을 생성해서는 안 되며, `.moai/harness/`에 어떠한 파일도 작성하지 않아야 한다.
- **REQ-PH-011** [Unwanted] 시스템은 `.claude/agents/moai/` 또는 `.claude/skills/moai-*/` 디렉터리 (moai-managed area)에 어떠한 파일도 작성·수정해서는 안 된다 (FROZEN).

### 4.5 Optional Requirements

- **REQ-PH-012** [Optional] WHERE Q13 (customization 범위) 답변이 "Advanced (full custom)"인 경우, 시스템은 `.moai/harness/design-extension.md`도 추가로 생성할 수 있어야 한다 (`/moai design` 워크플로우 확장 옵션).

---

## 5. Acceptance Criteria Summary

상세 시나리오는 `acceptance.md` 참조. 8개 핵심 AC:

- **AC-PH-01**: 16Q 인터뷰 시뮬레이션 (iOS 프로젝트 시나리오, Round 1-4 검증).
- **AC-PH-02**: 5-Layer 각각 독립 단위 테스트.
- **AC-PH-03**: meta-harness 호출 산출물 검증 (`.claude/agents/my-harness/*` + `my-harness-*` skills).
- **AC-PH-04**: 새 세션 시작 시 harness 자동 활성화 (CLAUDE.md @import follow).
- **AC-PH-05**: `moai update` 후 사용자 영역 보존 검증 (diff -rq).
- **AC-PH-06**: `moai doctor` 5-Layer 진단 + `my-harness-*` prefix 충돌 경고.
- **AC-PH-07**: `interview-results.md` 16개 답변 영구 기록 정상.
- **AC-PH-08**: 5-Layer 모두 활성 시 manager-tdd 자동 chain 적용 (handoff §5.2 iOS 시나리오 verbatim).

AC ↔ REQ traceability 100% 매트릭스는 `acceptance.md` §5.

---

## 6. Constraints

- **C-PH-001** [HARD] Template-First: `internal/template/templates/.claude/skills/moai/workflows/{plan,run,sync,design}.md`에 Layer 4 import line을 먼저 추가하고 `make build` 후 local 동기화.
- **C-PH-002** [HARD] FROZEN zone 보존: `.claude/rules/moai/design/constitution.md` §2 verbatim 준수. moai-managed area 절대 수정 금지.
- **C-PH-003** [HARD] AskUserQuestion 제약: max 4 질문 per call. 16 질문은 정확히 4 라운드로 분할. 각 옵션 첫 번째는 "(권장)" 마커 + 상세 설명 필수.
- **C-PH-004** [HARD] Conversation language 준수: 인터뷰 질문 / 옵션 / 라벨은 user의 conversation_language (현재 ko). 답변 기록 (interview-results.md)은 기록 시점 언어 그대로 보존.
- **C-PH-005** [HARD] 한국어 commit body 사용 (project rule).
- **C-PH-006** [HARD] EARS format 강제: 본 spec.md 모든 요구사항은 Ubiquitous / Event-Driven / State-Driven / Unwanted / Optional 5 패턴 중 하나.
- **C-PH-007** SPEC-V3R3-HARNESS-001 선행: meta-harness skill이 존재하지 않으면 본 SPEC 구현 불가. `depends_on` 필드 명시.

---

## 7. Exclusions (What NOT to Build)

- **EX-PH-001**: meta-harness skill 본체 (SPEC-V3R3-HARNESS-001).
- **EX-PH-002**: harness self-learning, auto-evolution (SPEC-V3R3-HARNESS-LEARNING-001).
- **EX-PH-003**: 정적 16 skills 제거 (BC-V3R3-007, HARNESS-001 범위).
- **EX-PH-004**: `/moai design` Path A / B1 / B2 분기 구현 (SPEC-V3R3-DESIGN-PIPELINE-001).
- **EX-PH-005**: 인터뷰 다국어 (en, ja, zh) — 본 SPEC은 conversation_language 단일.
- **EX-PH-006**: harness 산출물 cross-machine 동기화 (로컬 only).
- **EX-PH-007**: `moai cg` / GLM 모드 통합 (orthogonal, 별도 SPEC).

---

## 8. References

- handoff document: `.moai/release/v3r3-extreme-aggressive-handoff.md` §0.2, §0.3, §4.2, §5.2, §5.3, §6 (verbatim source for 5-Layer + 16Q).
- 디자인 시스템 헌법: `.claude/rules/moai/design/constitution.md` §2 (FROZEN zone), §11 (GAN Loop) — 미러링 참조만.
- Reference SPEC: `.moai/specs/SPEC-V3R3-HARNESS-LEARNING-001/` (구조 동형, depends 관계).
- 운영 정책: `CLAUDE.md` §8 (User Interaction Architecture), §10 (Worktree Isolation), `.claude/rules/moai/core/agent-common-protocol.md` (AskUserQuestion boundary).

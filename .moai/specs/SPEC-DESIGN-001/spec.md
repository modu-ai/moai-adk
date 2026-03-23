# SPEC-DESIGN-001: interface-design Plugin Integration

## Metadata
- **ID**: SPEC-DESIGN-001
- **Title**: interface-design 플러그인 개념의 MoAI ADK 통합
- **Status**: draft
- **Priority**: P1
- **Created**: 2026-03-23
- **Issue**: #550
- **Author**: MoAI Team (researcher + analyst + architect)

---

## Summary

interface-design 플러그인(https://github.com/Dammyjay93/interface-design)의 핵심 철학인 Intent-First 디자인 프로세스, 디자인 메모리, 도메인 탐색을 MoAI ADK의 템플릿 시스템에 통합한다. 새로운 `moai-design-craft` 스킬을 생성하고, 기존 에이전트(expert-frontend, team-designer)에 연결하여 모든 UI 작업에서 "동일함은 실패"(Sameness Is Failure) 원칙을 적용한다.

---

## Requirements (EARS Format)

### R1: Intent-First Design Process [P1]

**REQ-DESIGN-001**: When expert-frontend or team-designer agent is invoked for a UI task, IF `.moai/design/system.md` does NOT exist, the system SHALL execute the Intent-First process (Who? What? Feel?) before generating any design artifacts.

**REQ-DESIGN-002**: When expert-frontend or team-designer agent is invoked for a UI task, IF `.moai/design/system.md` DOES exist, the system SHALL load its contents as design context and apply established direction.

**REQ-DESIGN-003**: While executing the Intent-First process, the system SHALL produce: 5+ domain concepts, 5+ color world entries, 1 signature element, and a list of defaults to avoid.

### R2: Design Direction Sub-phase in /moai plan [P2]

**REQ-DESIGN-004**: When `/moai plan` is executed, IF the SPEC description contains UI/UX keywords (ui, frontend, interface, design, component, page, screen, layout, form, dashboard, button, modal, view), the plan workflow SHALL include a Design Direction sub-phase after Phase 0.5 (Deep Research).

**REQ-DESIGN-005**: Where a Design Direction sub-phase is executed, the system SHALL produce `.moai/specs/SPEC-{ID}/design-direction.md` containing intent statement, domain concepts, color world, signature element, and defaults to avoid.

### R3: Design Commands Integration [P3]

**REQ-DESIGN-006**: IF the user invokes `/moai review --design`, the system SHALL extract design patterns from existing code and create/update `.moai/design/system.md`.

**REQ-DESIGN-007**: IF the user invokes `/moai review --critique`, the system SHALL perform craft review focusing on subtle layering, surface elevation, token architecture, and typography hierarchy.

### R4: Design Memory System [P1]

**REQ-DESIGN-008**: When any design-related agent initializes, IF `.moai/design/system.md` exists, the system SHALL load its contents as part of the agent's initial context.

**REQ-DESIGN-009**: When the Intent-First process completes, the system SHALL persist design decisions to `.moai/design/system.md` with sections: intent, domain_concepts, color_world, signature_element, defaults_to_avoid, created, updated.

**REQ-DESIGN-010**: Where `.moai/design/system.md` is present, the system SHALL treat it as a project-level design contract that MUST NOT be silently overwritten.

### R5: Anti-Sameness Enforcement [P2]

**REQ-DESIGN-011**: When expert-frontend or team-designer generates UI components, IF no domain exploration has been completed (no system.md AND no design-direction.md), the system SHALL block component generation and trigger the Intent-First process first.

---

## Acceptance Criteria

### AC1: Intent-First triggered when no system.md
- **Given**: expert-frontend invoked with "Create a dashboard component" AND `.moai/design/system.md` does NOT exist
- **When**: The agent begins execution
- **Then**: Agent produces domain exploration (5+ concepts, 5+ colors, 1 signature) BEFORE generating code

### AC2: system.md applied when exists
- **Given**: expert-frontend invoked AND `.moai/design/system.md` exists
- **When**: Agent begins execution
- **Then**: Agent reads system.md and applies its design direction without re-triggering Intent-First

### AC3: Design sub-phase for UI SPEC
- **Given**: User runs `/moai plan "Create user analytics dashboard"`
- **When**: Plan workflow detects "dashboard" as UI keyword
- **Then**: Design Direction sub-phase executes, producing `design-direction.md`

### AC4: Design sub-phase skipped for non-UI SPEC
- **Given**: User runs `/moai plan "Fix database connection pool"`
- **When**: Plan workflow detects NO UI keywords
- **Then**: No Design Direction sub-phase; plan proceeds normally

### AC5: system.md not silently overwritten
- **Given**: `.moai/design/system.md` exists with content
- **When**: Intent-First would generate conflicting direction
- **Then**: User prompted: "Overwrite or merge?" before writing

### AC6: Design memory deployed by moai init
- **Given**: User runs `moai init myproject`
- **When**: Template deployment executes
- **Then**: `.moai/design/system.md` stub file is created

---

## Architecture

### Integration Strategy: New `moai-design-craft` skill

독립적인 새 스킬로 생성. 기존 moai-design-tools (도구 역학)와 moai-domain-uiux (토큰/WCAG)와 중복 없음.

### Design Memory Location: `.moai/design/system.md`

MoAI 네임스페이스 일관성 유지 (.moai/project/, .moai/specs/, .moai/design/). Git에 추적됨.

### Command Mapping: 기존 워크플로우 통합

| interface-design | MoAI Mapping |
|-----------------|--------------|
| /init | /moai plan Design Direction sub-phase |
| /extract | /moai review --design |
| /audit | /moai review (auto when system.md exists) |
| /critique | /moai review --critique |

---

## Implementation Plan

### Phase 1: Foundation (P1)

#### 1.1 Create `moai-design-craft` skill
- **NEW**: `internal/template/templates/.claude/skills/moai-design-craft/SKILL.md`
  - Frontmatter: name, description, triggers (intent-first, design craft, domain exploration)
  - Progressive disclosure: L1 ~100 tokens, L2 ~5000 tokens
  - Quick reference: Intent-First checklist, design memory protocol
- **NEW**: `internal/template/templates/.claude/skills/moai-design-craft/modules/intent-first.md`
  - Intent-First process (Who/What/Feel)
  - Product domain exploration (4 required outputs)
  - "Sameness Is Failure" mandate and checks
  - Craft foundations (subtle layering, surface elevation, token architecture)
- **NEW**: `internal/template/templates/.claude/skills/moai-design-craft/modules/design-memory.md`
  - system.md read/write protocol
  - When to update vs create new
  - Structure guide with 7 required sections
- **NEW**: `internal/template/templates/.claude/skills/moai-design-craft/modules/critique-workflow.md`
  - Post-build critique: composition, craft, content, structure
  - The swap test, squint test, signature test, token test

#### 1.2 Create design memory stub
- **NEW**: `internal/template/templates/.moai/design/system.md`
  - Placeholder with structure guide and instructions

#### 1.3 Agent integration
- **MODIFY**: `internal/template/templates/.claude/agents/moai/expert-frontend.md`
  - Add `- moai-design-craft` to skills list (after moai-design-tools)
- **MODIFY**: `internal/template/templates/.claude/agents/moai/team-designer.md`
  - Add `- moai-design-craft` to skills list

### Phase 2: Workflow Integration (P2)

#### 2.1 Plan workflow Design Direction sub-phase
- **MODIFY**: `internal/template/templates/.claude/skills/moai/workflows/plan.md`
  - Add Phase 1.25: Design Direction (conditional, after Phase 0.5)
  - UI keyword detection logic
  - design-direction.md template

### Phase 3: Review Integration (P3)

#### 3.1 Review workflow design flags
- **MODIFY**: `internal/template/templates/.claude/skills/moai/workflows/review.md`
  - Add --design flag: extract patterns → system.md
  - Add --critique flag: craft review workflow

### Post-Implementation

- Mirror all template changes to local `.claude/` and `.moai/` copies
- Run `make build` to regenerate embedded.go
- Run `go test ./...` for regression check

---

## Token Budget Impact

| Component | Tokens | Frequency |
|-----------|--------|-----------|
| moai-design-craft L1 (metadata) | ~100 | Always for design agents |
| moai-design-craft L2 (full) | ~5,000 | On-demand via triggers |
| Intent-First process output | ~2,000-3,000 | Once per project |
| Design Direction sub-phase | ~3,000-5,000 | Per UI SPEC |
| system.md loading | ~500-1,500 | Every UI agent invocation |

**Constraint**: Non-UI workflows add 0 tokens overhead.

---

## Non-Functional Requirements

- **NFR1**: Projects without `.moai/design/system.md` MUST work identically to current behavior
- **NFR2**: All design features are opt-in (keyword detection or --design flag)
- **NFR3**: `--prototype` flag bypasses all design enforcement
- **NFR4**: Configurable via `.moai/config/sections/quality.yaml` (`design.enabled: true/false`)

---

## Risks

| Risk | Mitigation |
|------|-----------|
| False positive UI keyword detection | Require 2+ keywords or context analysis |
| Token cost for large system.md | Progressive disclosure; L2 only when triggered |
| Generic names despite enforcement | Swap test and signature test in critique |
| Prototype friction | --prototype flag bypasses enforcement |

---

## File Change Map

### New Files (5)
| File | Est. Lines |
|------|-----------|
| `.claude/skills/moai-design-craft/SKILL.md` | ~120 |
| `.claude/skills/moai-design-craft/modules/intent-first.md` | ~80 |
| `.claude/skills/moai-design-craft/modules/design-memory.md` | ~60 |
| `.claude/skills/moai-design-craft/modules/critique-workflow.md` | ~80 |
| `.moai/design/system.md` | ~20 |

### Modified Files (2)
| File | Change |
|------|--------|
| `.claude/agents/moai/expert-frontend.md` | +1 line (skill) |
| `.claude/agents/moai/team-designer.md` | +1 line (skill) |

### Deferred (Phase 2-3)
| File | Change |
|------|--------|
| `.claude/skills/moai/workflows/plan.md` | +Design Direction sub-phase |
| `.claude/skills/moai/workflows/review.md` | +--design/--critique flags |

All template changes mirrored to local copies. `make build` required after changes.

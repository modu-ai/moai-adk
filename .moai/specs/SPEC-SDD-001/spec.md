---
spec_id: SPEC-SDD-001
title: "SDD Integration (Harness + Delta + Compact)"
created: "2026-04-01"
status: implemented
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# SDD Integration (Harness + Delta + Compact)

## Overview

MoAI-ADK의 SPEC 기반 개발 파이프라인에 3가지 핵심 메커니즘을 통합한다: (1) harness.yaml 기반 품질 수준 자동 선택, (2) 브라운필드 프로젝트를 위한 Delta Marker 시스템, (3) Run 단계의 토큰 효율을 위한 spec-compact.md 자동 생성. 이 3가지 메커니즘은 프로젝트 복잡도와 특성에 맞는 적응형 개발 워크플로우를 제공한다.

## Requirements (EARS Format)

### REQ-001: harness.yaml Configuration

**When** a MoAI-ADK project is initialized, the system **shall** include a `.moai/config/sections/harness.yaml` configuration file defining 3 harness levels with the following structure:

```
levels:
  minimal:  # Fast iteration, minimal checks
  standard: # Balanced quality and speed (default)
  thorough: # Maximum quality gates
mode_defaults:
  solo: auto
  team: auto
  cg: thorough
auto_detection:
  enabled: true
  thresholds:
    minimal: complexity < 3
    standard: 3 <= complexity < 7
    thorough: complexity >= 7
escalation:
  triggers:
    - quality_gate_fail
    - review_critical
    - test_coverage_low
```

**Acceptance**:
- **Given** a new MoAI-ADK project
- **When** template deployment completes
- **Then** `.moai/config/sections/harness.yaml` exists with all 3 levels defined
- **And** mode_defaults maps each execution mode to a default harness level
- **And** auto_detection thresholds are configured for complexity-based selection

### REQ-002: Delta Markers for Brownfield Projects

**When** a SPEC document describes modifications to an existing codebase, the spec.md template **shall** support [DELTA] section markers with 4 marker types:
- [EXISTING]: Unchanged code that provides context only
- [MODIFY]: Existing code to be changed (requires characterization tests)
- [NEW]: New code to be created (requires full implementation + tests)
- [REMOVE]: Code to be deleted (requires migration verification)

**When** the DDD ANALYZE phase processes a SPEC with delta markers, the system **shall**:
- Generate characterization tests for [EXISTING] and [MODIFY] items only
- Apply full implementation flow for [NEW] items
- Verify safe removal for [REMOVE] items

**Acceptance**:
- **Given** a SPEC with delta markers: 2 [EXISTING], 3 [MODIFY], 4 [NEW], 1 [REMOVE]
- **When** the DDD ANALYZE phase processes the SPEC
- **Then** [EXISTING] items receive characterization tests only (no modifications)
- **And** [MODIFY] items receive characterization tests before modification
- **And** [NEW] items receive full implementation with new tests
- **And** [REMOVE] items are verified for safe deletion with dependency check

### REQ-003: Token-Efficient spec-compact.md

**When** the Plan phase completes successfully, the system **shall** auto-generate a `.moai/specs/SPEC-{ID}/spec-compact.md` containing only:
- EARS format requirements (REQ-XXX entries)
- Acceptance criteria (Given/When/Then scenarios)
- Files to modify list
- Exclusions list

All research notes, plan details, discussion history, and overview text **shall** be excluded from spec-compact.md.

**When** the Run phase begins, the system **shall** load spec-compact.md instead of the full spec.md to achieve approximately 30% token savings.

**Acceptance**:
- **Given** a completed Plan phase with a full spec.md of 500 lines
- **When** spec-compact.md is auto-generated
- **Then** spec-compact.md contains only requirements, acceptance criteria, files, and exclusions
- **And** spec-compact.md is approximately 30% smaller than spec.md
- **And** the Run phase successfully loads and operates from spec-compact.md
- **Given** spec-compact.md does not exist (legacy SPEC)
- **When** the Run phase begins
- **Then** the system falls back to loading full spec.md (backward compatible)

## Exclusions (What NOT to Build)

- Shall NOT expose harness level selection to end users directly (auto-detect only, no CLI flag)
- Shall NOT require delta markers for greenfield projects (markers are optional)
- Shall NOT delete or modify the original spec.md when generating spec-compact.md
- Shall NOT apply delta marker logic when no markers are present in spec.md

## Acceptance Criteria

**Scenario 1: Auto-Detection of Harness Level**
- Given a SPEC with complexity score = 8 (3 domains, 15 files)
- When the Complexity Estimator runs in moai.md
- Then harness level = thorough is auto-selected
- And all thorough-level quality gates are activated

**Scenario 2: Brownfield Delta Processing**
- Given a SPEC with [MODIFY] markers on 3 existing service files
- When the DDD ANALYZE phase runs
- Then characterization tests are generated for all 3 [MODIFY] files before any changes
- And the characterization tests pass against the current codebase

**Scenario 3: Token Savings with spec-compact.md**
- Given a SPEC whose spec.md is 8000 tokens
- When the Plan phase completes and spec-compact.md is generated
- Then spec-compact.md is approximately 5600 tokens or less (30% reduction)
- And the Run phase loads spec-compact.md successfully

**Scenario 4: Backward Compatibility**
- Given an existing SPEC created before v2.9.0 (no spec-compact.md, no delta markers)
- When the Run phase starts
- Then the system loads full spec.md without errors
- And no delta marker processing occurs

## Technical Approach

1. Create harness.yaml template with 3-level configuration and auto-detection thresholds
2. Add Complexity Estimator to moai.md that calculates domain count, file count, and dependency depth
3. Extend spec.md format to recognize [DELTA] markers in plan.md
4. Implement DDD ANALYZE phase delta-aware routing in run.md
5. Add spec-compact.md generation step at Plan phase completion in plan.md
6. Modify Run phase entry point in run.md to prefer spec-compact.md over spec.md

## Files to Modify

- `internal/template/templates/.moai/config/sections/harness.yaml` (NEW)
- `internal/template/templates/.claude/skills/moai/workflows/plan.md` (delta markers, compact generation)
- `internal/template/templates/.claude/skills/moai/workflows/run.md` (harness routing, compact loading)
- `internal/template/templates/.claude/skills/moai/workflows/moai.md` (Complexity Estimator)

---
spec_id: SPEC-DRIFT-001
title: "Spec Drift Guard & Persistent TASKS.md"
created: "2026-04-01"
status: planned
priority: high
module: template
version: "1.0.0"
lifecycle: spec-anchored
---

# Spec Drift Guard & Persistent TASKS.md

## Overview

SPEC 기반 개발에서 구현 과정 중 발생하는 스코프 드리프트(scope drift)를 실시간으로 감지하고 경고하는 메커니즘을 도입한다. Phase 1.5에서 생성되는 persistent tasks.md를 통해 계획된 작업과 실제 구현을 추적하며, DDD/TDD 사이클 완료 시마다 계획 대비 실제 변경 파일을 비교하여 드리프트를 감지한다.

## Requirements (EARS Format)

### REQ-001: Persistent tasks.md Artifact

**When** the Run phase reaches Phase 1.5 (task decomposition), the system **shall** output a `.moai/specs/SPEC-{ID}/tasks.md` file containing the full task decomposition with the following fields per task:
- Task ID (sequential, e.g., T-001)
- Description
- Requirement mapping (REQ-XXX reference)
- Dependencies (list of prerequisite Task IDs)
- Status (pending, in-progress, completed, skipped)
- Planned files (list of files expected to be modified/created)

The tasks.md file **shall** be git-tracked for audit trail purposes.

**Acceptance**:
- **Given** a SPEC entering Phase 1.5 of the Run phase
- **When** task decomposition completes
- **Then** `.moai/specs/SPEC-{ID}/tasks.md` is created with all required fields
- **And** each task maps to at least one REQ-XXX requirement
- **And** the file is suitable for git tracking (deterministic output, no timestamps)

### REQ-002: Real-Time Drift Guard

**When** a DDD or TDD cycle completes, the system **shall** compare planned_files (from tasks.md) against actual_files (files actually modified/created) and:
- Calculate drift percentage: (actual_new_files - planned_files) / planned_files * 100
- Alert if new files exceed the plan by more than 20%
- Log the drift measurement to progress.md

The drift check step **shall** be added to run.md Phase 2A (DDD) and Phase 2B (TDD) cycle completion points.

**Acceptance**:
- **Given** a DDD cycle that completes with files matching the plan
- **When** drift guard runs
- **Then** no alert is generated and progress.md records "drift: 0%"
- **Given** a DDD cycle that creates 3 unplanned files out of 10 planned files
- **When** drift guard runs
- **Then** an alert is generated (30% > 20% threshold)
- **And** progress.md records the drift warning with file details

### REQ-003: Scope Alarm Integration

**When** drift is detected, the system **shall** append a warning entry to `.moai/specs/SPEC-{ID}/progress.md` with drift details.
**When** cumulative scope expansion exceeds 30%, the system **shall** trigger the existing Phase 2.7 re-planning gate for scope review.

**Acceptance**:
- **Given** cumulative drift reaches 25%
- **When** the next cycle completes
- **Then** a warning is appended to progress.md but re-planning is not triggered
- **Given** cumulative drift reaches 35%
- **When** the next cycle completes
- **Then** a warning is appended to progress.md
- **And** the Phase 2.7 re-planning gate is triggered for scope review

## Exclusions (What NOT to Build)

- Shall NOT block implementation on drift detection (warning only, not blocking gate)
- Shall NOT modify existing Phase 2.7 re-planning logic (extend trigger conditions, not replace)
- Shall NOT track drift for files in exclusion patterns (.gitignore, test fixtures, generated files)
- Shall NOT require manual drift acknowledgment (automated warning flow)

## Acceptance Criteria

**Scenario 1: Clean Implementation (No Drift)**
- Given a SPEC with 5 tasks and 12 planned files
- When all 5 tasks complete modifying exactly the 12 planned files
- Then tasks.md shows all tasks as "completed"
- And progress.md shows "drift: 0%" for each cycle
- And no alerts are generated

**Scenario 2: Minor Drift (Below Threshold)**
- Given a SPEC with 10 planned files
- When implementation adds 1 unplanned utility file (10% drift)
- Then progress.md records "drift: 10%" as informational
- And no alert is generated (below 20% threshold)

**Scenario 3: Significant Drift (Above Threshold)**
- Given a SPEC with 10 planned files
- When implementation adds 4 unplanned files (40% drift)
- Then progress.md records drift warning with details of unplanned files
- And the Phase 2.7 re-planning gate triggers (>30% cumulative)
- And the user is informed of scope expansion

## Technical Approach

1. Add Phase 1.5 tasks.md generation to run.md with structured task format
2. Implement drift calculation logic in run.md cycle completion hooks
3. Integrate drift warnings with existing progress.md tracking
4. Connect >30% cumulative drift to Phase 2.7 re-planning trigger
5. Add drift check step references in workflow-modes.md

## Files to Modify

- `internal/template/templates/.claude/skills/moai/workflows/run.md` (Phase 1.5 tasks.md output, drift guard in Phase 2A/2B)
- `internal/template/templates/.claude/rules/moai/workflow/workflow-modes.md` (drift check in cycle completion)

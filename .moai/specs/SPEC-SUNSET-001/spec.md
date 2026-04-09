---
id: SPEC-SUNSET-001
title: "Build to Delete Framework"
status: draft
priority: P2
created: "2026-04-07"
harness_pillar: "All - Evolutionary"
---

# SPEC-SUNSET-001: Build to Delete Framework

## Overview

최고의 하네스는 에이전트 성장에 따라 점진적으로 삭제된다 (Richard Sutton's Bitter Lesson).
각 quality gate에 sunset_condition을 명시하고 에이전트 성공률을 추적하여
삭제 가능한 게이트를 자동 식별하는 프레임워크.

## Requirements (EARS Format)

### REQ-SUN-001 (Ubiquitous)
`.moai/config/sections/sunset.yaml` SHALL define sunset conditions for quality gates.

### REQ-SUN-002 (Ubiquitous)
Each sunset condition SHALL specify: gate_name, metric, threshold, consecutive_passes, and action (advisor/remove).

### REQ-SUN-003 (Ubiquitous)
The trace system (SPEC-OBSERVE-001) SHALL track per-gate pass/fail counts in session summaries.

### REQ-SUN-004 (Event-Driven)
When a gate meets its sunset condition, the system SHALL log a recommendation to `.moai/reports/sunset-recommendations.md`.

### REQ-SUN-005 (Ubiquitous)
Sunset actions SHALL be advisory only — actual gate removal requires human approval.

## Implementation Scope

### New Files
- `internal/template/templates/.moai/config/sections/sunset.yaml` — Sunset conditions config
- `.moai/specs/SPEC-SUNSET-001/spec.md` — This file

### Configuration Schema
```yaml
sunset:
  enabled: true
  conditions:
    - gate: "go_vet"
      metric: "consecutive_passes"
      threshold: 50
      action: "advisor"
      description: "go vet passes 50 consecutive times -> switch to advisor mode"
    - gate: "golangci_lint"
      metric: "consecutive_passes"
      threshold: 30
      action: "advisor"
      description: "golangci-lint passes 30 consecutive times -> switch to advisor mode"
    - gate: "go_test"
      metric: "consecutive_passes"
      threshold: 20
      action: "advisor"
      description: "go test passes 20 consecutive times -> switch to advisor mode (never remove)"
```

## Non-Goals
- Automatic gate removal (advisory only for v1)
- Per-agent sunset tracking (gate-level only)
- UI dashboard for sunset metrics

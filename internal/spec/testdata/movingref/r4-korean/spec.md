---
id: SPEC-MRGA-001
title: "Moving-ref fixture — unpinned anchor"
version: "0.1.0"
status: draft
created: 2026-08-28
updated: 2026-08-28
author: Test Author
priority: P2 Medium
phase: "v3.2.0"
module: "internal/spec"
era: V3R6
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test, movingref"
breaking: false
related_rule: []
---

# SPEC-MRGA-001: Moving-ref fixture — unpinned anchor

The Korean-form R4 line (card t353). Both conjuncts in the Korean idiom: the
imperative measuring directive ("측정: `<cmd>`", the shape the live corpus line
`SPEC-SPECLINT-GITBLIND-001/progress.md:233` carries) and the demoted dated
reference in parentheses ("(참조 판독 2026-08-28: empty)"). Flaggable absent the
exclusion: claim marker "empty", no hex SHA, moving ref inside a git-command
context.

| AC | Command | Expected |
|---|---|---|
| AC-X | 측정: `git diff --name-only origin/develop -- internal/hook` — 읽는 시점에 재측정 (참조 판독 2026-08-28: empty) | 0 findings |

Nothing else in this fixture directory names a moving ref.

## 2. Scope

### 2.1 In Scope

- The single Korean R4-form row above

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the Korean R4-form row, When lint runs, Then zero findings. (maps REQ-MGA-001-001)

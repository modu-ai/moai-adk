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

CM-1 (positional bypass) of AC-MRG-013 (REQ-MRG-010, card t353). The row
mentions a git command and a parenthesis but states a RESULT, not an
instruction to measure — the dated reference conjunct is absent entirely. A
positional exclusion ("a git verb before the ref and a parenthesized value
after it") would exempt this row and AC-MRG-001's fixture with it; this
criterion keeps both red.

| AC | Command | Expected |
|---|---|---|
| AC-X | `git diff --name-only origin/main -- internal/ (unchanged)` | 1 finding |

Nothing else in this fixture directory names a moving ref.

## 2. Scope

### 2.1 In Scope

- The single CM-1 row above

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the CM-1 row, When lint runs, Then one finding. (maps REQ-MGA-001-001)

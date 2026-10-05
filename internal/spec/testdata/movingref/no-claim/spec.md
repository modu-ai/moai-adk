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

The flagged line lives in this fixture's `acceptance.md`. This body is deliberately
clean so the only findings the linter reports are `MovingRefUnpinned`.

## 2. Scope

### 2.1 In Scope

- The single flagged row in acceptance.md

### 2.2 Out of Scope

- Nothing else

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-MGA-001-001: The system SHALL flag an unpinned moving-ref invariant claim.

## 6. Acceptance Criteria

- AC-MGA-001-01: Given the flagged row, When lint runs, Then one finding. (maps REQ-MGA-001-001)

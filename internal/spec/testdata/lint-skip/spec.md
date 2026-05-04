---
id: SPEC-V3R2-TST-009
title: "Lint skip test SPEC"
version: "0.1.0"
status: draft
created: 2026-04-30
updated: 2026-04-30
author: Test Author
priority: P2 Medium
phase: "v3.0.0"
module: "internal/test"
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test"
breaking: false
related_rule:
  - CONST-V3R2-999
lint:
  skip:
    - DanglingRuleReference
---

# SPEC-V3R2-TST-009: Lint skip

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-TST-009-001: The system SHALL do thing A.

## 6. Acceptance Criteria

- AC-TST-009-01: Given A, When something, Then result. (maps REQ-TST-009-001)

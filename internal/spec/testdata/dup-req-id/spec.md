---
id: SPEC-V3R2-TST-003
title: "Duplicate REQ ID test SPEC"
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
related_rule: []
---

# SPEC-V3R2-TST-003: Duplicate REQ ID

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

### 5.1 Ubiquitous

- REQ-TST-003-001: The system SHALL do thing A.
- REQ-TST-003-005: The system SHALL do thing B.
- REQ-TST-003-005: The system SHALL do thing C (duplicate ID).

## 6. Acceptance Criteria

- AC-TST-003-01: Given A, When something, Then result. (maps REQ-TST-003-001)
- AC-TST-003-02: Given B, When something, Then result. (maps REQ-TST-003-005)

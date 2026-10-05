---
id: SPEC-DAI-001
title: "Duplicate inline AC id with a citing bullet above the declaration"
version: "0.1.0"
status: draft
created: 2026-09-10
updated: 2026-09-10
author: Test Author
priority: P2 Medium
phase: "v3.2.0"
module: "internal/spec"
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test"
breaking: false
related_rule: []
---

# SPEC-DAI-001: Duplicate inline AC id with a citing bullet above the declaration

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

- REQ-DAI-001-001: The system SHALL keep the mapping of the real AC declaration.

## 6. Acceptance Criteria

- AC-DAI-001-01: cited here only as a cross-reference note
- AC-DAI-001-01: Given a SPEC, When it is linted, Then the declaration counts (maps REQ-DAI-001-001)

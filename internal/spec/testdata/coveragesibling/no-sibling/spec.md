---
id: SPEC-CSB-003
title: "Tier S with no sibling acceptance artifact"
version: "0.1.0"
status: draft
created: 2026-08-31
updated: 2026-08-31
author: Test Author
priority: P2 Medium
phase: "v3.1.0"
module: "internal/spec"
dependencies: []
bc_id: []
lifecycle: spec-anchored
tags: "test"
breaking: false
related_rule: []
---

# SPEC-CSB-003: Tier S with no sibling acceptance artifact

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

- REQ-CSB-003-001: The system SHALL judge coverage from the inline AC section.
- REQ-CSB-003-002: The system SHALL report this REQ, which no inline AC maps.

## 6. Acceptance Criteria

- AC-CSB-003-01: Given a Tier S SPEC, When lint runs, Then inline AC decides coverage. (maps REQ-CSB-003-001)

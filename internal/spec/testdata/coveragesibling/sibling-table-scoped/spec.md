---
id: SPEC-CST-002
title: "Sibling acceptance table collects only REQ-headed columns within one cell"
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

# SPEC-CST-002: Sibling acceptance table collects only REQ-headed columns within one cell

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

- REQ-CST-002-001: The system SHALL count a full REQ id in a REQ-headed column as covered.
- REQ-CST-002-002: The system SHALL NOT expand a bare numeric tail that has no full REQ id before it in the same cell.
- REQ-CST-002-003: The system SHALL NOT count a REQ id that appears only in a column whose header does not name a requirement.

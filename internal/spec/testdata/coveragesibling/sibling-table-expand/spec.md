---
id: SPEC-CST-001
title: "Sibling acceptance table with header-identified REQ column"
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

# SPEC-CST-001: Sibling acceptance table with header-identified REQ column

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

- REQ-CST-001-001: The system SHALL read a mapping from the REQ column of a sibling table.
- REQ-CST-001-002: The system SHALL expand a numeric tail that follows a full REQ id in the same cell.
- REQ-CST-001-003: The system SHALL read the REQ column on every data row of the table.

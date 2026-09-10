---
id: SPEC-CST-004
title: "Sibling acceptance REQ-first table counts only rows that name an AC"
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

# SPEC-CST-004: Sibling acceptance REQ-first table counts only rows that name an AC

## 2. Scope

### 2.2 Out of Scope

- Nothing excluded

## 5. Requirements (EARS)

- REQ-CST-004-001: The system SHALL count a REQ-first row that names an AC as covered.
- REQ-CST-004-002: The system SHALL keep reporting a REQ-first row whose AC cell names no AC.

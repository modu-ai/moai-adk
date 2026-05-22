---
id: SPEC-FIXTURE-GEARS-001
title: "GEARS Canonical Lint Fixture"
version: "0.0.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: GEARS-MIGRATION-001
priority: P3
phase: "v3.0.0"
module: "test-fixtures"
lifecycle: spec-anchored
tags: "fixture, gears, canonical"
---

# SPEC-FIXTURE-GEARS-001: GEARS Canonical Sample Fixture

This fixture exercises GEARS canonical notation (WHEN/WHILE/WHERE/Ubiquitous). No IF/THEN.
Lint output MUST be empty (no findings) per AC-GM-003.

## 2. Scope

### 3.1 Out of Scope

- IF/THEN legacy patterns (covered by legacy-ears-sample.md)
- Real-world REQ semantics (synthetic test data only)

## 5. Requirements (GEARS)

### 5.1 Ubiquitous

- REQ-GRS-001-001: The system SHALL always preserve EARS modality compliance.

### 5.2 Event-driven (WHEN)

- REQ-GRS-001-002: WHEN a new SPEC is added, the system SHALL detect it during discovery.

### 5.3 State-driven (WHILE)

- REQ-GRS-001-003: WHILE the linter holds the rule registry, the system SHALL apply rules in declaration order.

### 5.4 Optional precondition (WHERE)

- REQ-GRS-001-004: WHERE strict mode is enabled, the system SHALL escalate warnings to errors via Report.HasErrors().

## 6. Acceptance Criteria

- AC-GRS-001-01: Given the GEARS fixture, When linted, Then zero findings. (maps REQ-GRS-001-001)
- AC-GRS-001-02: Given new SPEC, When linted, Then it is discovered. (maps REQ-GRS-001-002)
- AC-GRS-001-03: Given the fixture, When linted, Then rule order is deterministic. (maps REQ-GRS-001-003)
- AC-GRS-001-04: Given strict mode, When warnings present, Then exit 1. (maps REQ-GRS-001-004)
